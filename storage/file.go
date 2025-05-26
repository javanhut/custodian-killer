package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PolicyStorage interface for different storage backends
type PolicyStorage interface {
	SavePolicy(policy StoredPolicy) error
	GetPolicy(name string) (*StoredPolicy, error)
	ListPolicies() ([]StoredPolicy, error)
	DeletePolicy(name string) error
	PolicyExists(name string) bool
	GetPolicyHistory(name string) ([]StoredPolicy, error)
}

// StoredPolicy represents a policy stored in the system
type StoredPolicy struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	ResourceType string                 `json:"resource_type"`
	Filters      []StoredFilter         `json:"filters"`
	Actions      []StoredAction         `json:"actions"`
	Mode         StoredPolicyMode       `json:"mode"`
	Tags         map[string]string      `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CreatedBy    string                 `json:"created_by"`
	Version      int                    `json:"version"`
	Status       string                 `json:"status"` // active, inactive, draft
	LastRun      *time.Time             `json:"last_run,omitempty"`
	RunCount     int                    `json:"run_count"`
	Source       string                 `json:"source"` // template, manual, import
	TemplateID   string                 `json:"template_id,omitempty"`
}

type StoredFilter struct {
	Type     string      `json:"type"`
	Key      string      `json:"key,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	Op       string      `json:"op,omitempty"`
	Required bool        `json:"required,omitempty"`
	Negate   bool        `json:"negate,omitempty"`
}

type StoredAction struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	DryRun   bool                   `json:"dry_run"`
}

type StoredPolicyMode struct {
	Type     string            `json:"type"`
	Schedule string            `json:"schedule,omitempty"`
	Settings map[string]string `json:"settings,omitempty"`
}

// FileStorage implements PolicyStorage using local filesystem
type FileStorage struct {
	baseDir string
}

// NewFileStorage creates a new file-based storage system
func NewFileStorage(baseDir string) (*FileStorage, error) {
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %v", err)
		}
		baseDir = filepath.Join(homeDir, ".custodian-killer")
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Create subdirectories
	policiesDir := filepath.Join(baseDir, "policies")
	historyDir := filepath.Join(baseDir, "history")

	if err := os.MkdirAll(policiesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create policies directory: %v", err)
	}

	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %v", err)
	}

	return &FileStorage{baseDir: baseDir}, nil
}

// SavePolicy saves a policy to the filesystem
func (fs *FileStorage) SavePolicy(policy StoredPolicy) error {
	// Set timestamps
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now()
	}
	policy.UpdatedAt = time.Now()

	// Set default values
	if policy.Status == "" {
		policy.Status = "active"
	}
	if policy.CreatedBy == "" {
		policy.CreatedBy = "custodian-killer-user"
	}

	// Increment version if policy exists
	if fs.PolicyExists(policy.Name) {
		existing, err := fs.GetPolicy(policy.Name)
		if err == nil {
			policy.Version = existing.Version + 1
			// Save to history before updating
			fs.saveToHistory(*existing)
		}
	} else {
		policy.Version = 1
	}

	// Convert to JSON
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %v", err)
	}

	// Save to file
	filename := filepath.Join(fs.baseDir, "policies", fmt.Sprintf("%s.json", policy.Name))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write policy file: %v", err)
	}

	fmt.Printf("💾 Policy '%s' saved to: %s\n", policy.Name, filename)
	return nil
}

// GetPolicy retrieves a policy by name
func (fs *FileStorage) GetPolicy(name string) (*StoredPolicy, error) {
	filename := filepath.Join(fs.baseDir, "policies", fmt.Sprintf("%s.json", name))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("policy '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read policy file: %v", err)
	}

	var policy StoredPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy file: %v", err)
	}

	return &policy, nil
}

// ListPolicies returns all stored policies
func (fs *FileStorage) ListPolicies() ([]StoredPolicy, error) {
	policiesDir := filepath.Join(fs.baseDir, "policies")

	files, err := os.ReadDir(policiesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read policies directory: %v", err)
	}

	var policies []StoredPolicy
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			policyName := file.Name()[:len(file.Name())-5] // Remove .json extension
			policy, err := fs.GetPolicy(policyName)
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to load policy '%s': %v\n", policyName, err)
				continue
			}
			policies = append(policies, *policy)
		}
	}

	return policies, nil
}

// DeletePolicy removes a policy
func (fs *FileStorage) DeletePolicy(name string) error {
	filename := filepath.Join(fs.baseDir, "policies", fmt.Sprintf("%s.json", name))

	// Check if policy exists
	if !fs.PolicyExists(name) {
		return fmt.Errorf("policy '%s' not found", name)
	}

	// Save to history before deleting
	policy, err := fs.GetPolicy(name)
	if err == nil {
		policy.Status = "deleted"
		fs.saveToHistory(*policy)
	}

	// Delete the file
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete policy file: %v", err)
	}

	fmt.Printf("🗑️  Policy '%s' deleted\n", name)
	return nil
}

// PolicyExists checks if a policy exists
func (fs *FileStorage) PolicyExists(name string) bool {
	filename := filepath.Join(fs.baseDir, "policies", fmt.Sprintf("%s.json", name))
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// GetPolicyHistory returns the version history of a policy
func (fs *FileStorage) GetPolicyHistory(name string) ([]StoredPolicy, error) {
	historyDir := filepath.Join(fs.baseDir, "history", name)

	files, err := os.ReadDir(historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []StoredPolicy{}, nil // No history
		}
		return nil, fmt.Errorf("failed to read history directory: %v", err)
	}

	var history []StoredPolicy
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(historyDir, file.Name()))
			if err != nil {
				continue
			}

			var policy StoredPolicy
			if err := json.Unmarshal(data, &policy); err != nil {
				continue
			}

			history = append(history, policy)
		}
	}

	return history, nil
}

// saveToHistory saves a policy version to history
func (fs *FileStorage) saveToHistory(policy StoredPolicy) error {
	historyDir := filepath.Join(fs.baseDir, "history", policy.Name)
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(
		historyDir,
		fmt.Sprintf("v%d_%d.json", policy.Version, time.Now().Unix()),
	)

	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GetStorageInfo returns information about the storage system
func (fs *FileStorage) GetStorageInfo() (map[string]interface{}, error) {
	policies, err := fs.ListPolicies()
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"storage_type":   "file",
		"base_directory": fs.baseDir,
		"policies_count": len(policies),
		"storage_path":   filepath.Join(fs.baseDir, "policies"),
		"history_path":   filepath.Join(fs.baseDir, "history"),
	}

	// Calculate storage size
	var totalSize int64
	filepath.Walk(fs.baseDir, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			totalSize += info.Size()
		}
		return nil
	})
	info["storage_size_bytes"] = totalSize
	info["storage_size_mb"] = float64(totalSize) / (1024 * 1024)

	return info, nil
}

// ExportPolicy exports a policy to a specific file
func (fs *FileStorage) ExportPolicy(name, outputPath string) error {
	policy, err := fs.GetPolicy(name)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %v", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %v", err)
	}

	fmt.Printf("📤 Policy '%s' exported to: %s\n", name, outputPath)
	return nil
}

// ImportPolicy imports a policy from a file
func (fs *FileStorage) ImportPolicy(inputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %v", err)
	}

	var policy StoredPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return fmt.Errorf("failed to parse import file: %v", err)
	}

	// Mark as imported
	policy.Source = "import"
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	if err := fs.SavePolicy(policy); err != nil {
		return fmt.Errorf("failed to save imported policy: %v", err)
	}

	fmt.Printf("📥 Policy '%s' imported from: %s\n", policy.Name, inputPath)
	return nil
}
