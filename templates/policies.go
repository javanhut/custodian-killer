package templates

import (
	"fmt"
	"strings"
	"time"
)

// PolicyTemplate represents a reusable policy template
type PolicyTemplate struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Category     string           `json:"category"`
	ResourceType string           `json:"resource_type"`
	Difficulty   string           `json:"difficulty"` // beginner, intermediate, advanced
	Impact       string           `json:"impact"`     // low, medium, high
	Variables    []TemplateVar    `json:"variables"`
	Template     PolicyDefinition `json:"template"`
	Examples     []string         `json:"examples"`
	Tags         []string         `json:"tags"`
	CreatedBy    string           `json:"created_by"`
}

// TemplateVar represents a customizable variable in a template
type TemplateVar struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Type         string      `json:"type"` // string, int, bool, duration, list
	DefaultValue interface{} `json:"default_value"`
	Required     bool        `json:"required"`
	Options      []string    `json:"options,omitempty"`    // for enum-like variables
	Validation   string      `json:"validation,omitempty"` // regex or validation rule
}

// PolicyDefinition matches the main Policy struct but for templates
type PolicyDefinition struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	ResourceType string                 `json:"resource_type"`
	Filters      []FilterDefinition     `json:"filters"`
	Actions      []ActionDefinition     `json:"actions"`
	Mode         PolicyModeDefinition   `json:"mode"`
	Tags         map[string]string      `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type FilterDefinition struct {
	Type     string      `json:"type"`
	Key      string      `json:"key,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	Op       string      `json:"op,omitempty"`
	Required bool        `json:"required,omitempty"`
	Negate   bool        `json:"negate,omitempty"`
}

type ActionDefinition struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	DryRun   bool                   `json:"dry_run"`
}

type PolicyModeDefinition struct {
	Type     string            `json:"type"`
	Schedule string            `json:"schedule,omitempty"`
	Settings map[string]string `json:"settings,omitempty"`
}

// Built-in policy templates - The good stuff! 🔥
var BuiltInTemplates = []PolicyTemplate{
	{
		ID:           "unused-ec2-killer",
		Name:         "Unused EC2 Instance Killer",
		Description:  "Find and terminate EC2 instances with low CPU utilization",
		Category:     "cost-optimization",
		ResourceType: "ec2",
		Difficulty:   "beginner",
		Impact:       "high",
		Variables: []TemplateVar{
			{
				Name:         "cpu_threshold",
				Description:  "CPU utilization threshold percentage",
				Type:         "int",
				DefaultValue: 5,
				Required:     true,
			},
			{
				Name:         "days_unused",
				Description:  "Number of days of low utilization",
				Type:         "int",
				DefaultValue: 7,
				Required:     true,
			},
			{
				Name:         "action_type",
				Description:  "What to do with unused instances",
				Type:         "string",
				DefaultValue: "stop",
				Options:      []string{"stop", "terminate", "tag-only"},
				Required:     true,
			},
		},
		Template: PolicyDefinition{
			Name:         "{{.policy_name}}",
			Description:  "Automatically handle EC2 instances with CPU < {{.cpu_threshold}}% for {{.days_unused}} days",
			ResourceType: "ec2",
			Filters: []FilterDefinition{
				{Type: "instance-state", Value: "running", Op: "eq"},
				{Type: "cpu-utilization-avg", Value: "{{.cpu_threshold}}", Op: "lt"},
				{Type: "running-days", Value: "{{.days_unused}}", Op: "gte"},
			},
			Actions: []ActionDefinition{
				{Type: "{{.action_type}}", DryRun: true},
			},
			Mode: PolicyModeDefinition{Type: "pull"},
		},
		Examples: []string{
			"Find instances with <5% CPU for 7+ days and stop them",
			"Terminate instances unused for 30+ days",
		},
		Tags:      []string{"cost", "ec2", "optimization", "popular"},
		CreatedBy: "custodian-killer-team",
	},
	{
		ID:           "untagged-resources-tagger",
		Name:         "Untagged Resources Auto-Tagger",
		Description:  "Find and tag resources missing required tags",
		Category:     "compliance",
		ResourceType: "ec2",
		Difficulty:   "beginner",
		Impact:       "medium",
		Variables: []TemplateVar{
			{
				Name:         "required_tags",
				Description:  "Comma-separated list of required tag keys",
				Type:         "list",
				DefaultValue: "Environment,Owner,Project",
				Required:     true,
			},
			{
				Name:         "default_environment",
				Description:  "Default value for Environment tag",
				Type:         "string",
				DefaultValue: "untagged",
				Required:     false,
			},
			{
				Name:         "default_owner",
				Description:  "Default value for Owner tag",
				Type:         "string",
				DefaultValue: "unknown",
				Required:     false,
			},
		},
		Template: PolicyDefinition{
			Name:         "{{.policy_name}}",
			Description:  "Auto-tag resources missing required tags: {{.required_tags}}",
			ResourceType: "ec2",
			Filters: []FilterDefinition{
				{Type: "tag-missing", Key: "Environment", Op: "missing"},
			},
			Actions: []ActionDefinition{
				{
					Type: "tag",
					Settings: map[string]interface{}{
						"Environment": "{{.default_environment}}",
						"Owner":       "{{.default_owner}}",
						"AutoTagged":  "true",
						"TaggedDate":  "{{.current_date}}",
					},
					DryRun: true,
				},
			},
			Mode: PolicyModeDefinition{Type: "pull"},
		},
		Examples: []string{
			"Tag all EC2 instances missing Environment tag",
			"Bulk tag resources with compliance tags",
		},
		Tags:      []string{"compliance", "tagging", "governance"},
		CreatedBy: "custodian-killer-team",
	},
	{
		ID:           "public-s3-buckets-locker",
		Name:         "Public S3 Bucket Security Locker",
		Description:  "Find and secure publicly accessible S3 buckets",
		Category:     "security",
		ResourceType: "s3",
		Difficulty:   "intermediate",
		Impact:       "high",
		Variables: []TemplateVar{
			{
				Name:         "action_type",
				Description:  "How to secure the buckets",
				Type:         "string",
				DefaultValue: "block-public-access",
				Options:      []string{"block-public-access", "tag-only", "notify-only"},
				Required:     true,
			},
			{
				Name:         "notification_email",
				Description:  "Email to notify about public buckets",
				Type:         "string",
				DefaultValue: "",
				Required:     false,
			},
		},
		Template: PolicyDefinition{
			Name:         "{{.policy_name}}",
			Description:  "Secure S3 buckets with public access via {{.action_type}}",
			ResourceType: "s3",
			Filters: []FilterDefinition{
				{Type: "public-read", Value: true, Op: "eq"},
				{Type: "public-write", Value: true, Op: "eq"},
			},
			Actions: []ActionDefinition{
				{Type: "{{.action_type}}", DryRun: true},
				{
					Type: "tag",
					Settings: map[string]interface{}{
						"SecurityRisk":   "public-access",
						"RemediatedDate": "{{.current_date}}",
						"RemediatedBy":   "custodian-killer",
					},
				},
			},
			Mode: PolicyModeDefinition{Type: "pull"},
		},
		Examples: []string{
			"Block public access on all public S3 buckets",
			"Tag public buckets for manual review",
		},
		Tags:      []string{"security", "s3", "public-access", "critical"},
		CreatedBy: "custodian-killer-team",
	},
	{
		ID:           "old-ebs-snapshots-cleaner",
		Name:         "Old EBS Snapshots Cleaner",
		Description:  "Clean up old EBS snapshots to reduce storage costs",
		Category:     "cost-optimization",
		ResourceType: "ebs-snapshot",
		Difficulty:   "beginner",
		Impact:       "medium",
		Variables: []TemplateVar{
			{
				Name:         "retention_days",
				Description:  "Keep snapshots newer than this many days",
				Type:         "int",
				DefaultValue: 30,
				Required:     true,
			},
			{
				Name:         "keep_count",
				Description:  "Always keep this many recent snapshots per volume",
				Type:         "int",
				DefaultValue: 3,
				Required:     true,
			},
		},
		Template: PolicyDefinition{
			Name:         "{{.policy_name}}",
			Description:  "Delete EBS snapshots older than {{.retention_days}} days, keeping {{.keep_count}} recent ones",
			ResourceType: "ebs-snapshot",
			Filters: []FilterDefinition{
				{Type: "age", Value: "{{.retention_days}}", Op: "gt"},
				{Type: "state", Value: "completed", Op: "eq"},
			},
			Actions: []ActionDefinition{
				{Type: "delete", DryRun: true},
			},
			Mode: PolicyModeDefinition{Type: "pull"},
		},
		Examples: []string{
			"Delete snapshots older than 30 days, keep 3 recent per volume",
			"Clean up old snapshots to save storage costs",
		},
		Tags:      []string{"cost", "ebs", "snapshots", "cleanup"},
		CreatedBy: "custodian-killer-team",
	},
	{
		ID:           "rds-backup-enforcer",
		Name:         "RDS Backup Policy Enforcer",
		Description:  "Ensure RDS instances have proper backup configuration",
		Category:     "compliance",
		ResourceType: "rds",
		Difficulty:   "intermediate",
		Impact:       "high",
		Variables: []TemplateVar{
			{
				Name:         "min_backup_retention",
				Description:  "Minimum backup retention period in days",
				Type:         "int",
				DefaultValue: 7,
				Required:     true,
			},
			{
				Name:         "require_multi_az",
				Description:  "Require Multi-AZ deployment for production",
				Type:         "bool",
				DefaultValue: true,
				Required:     false,
			},
		},
		Template: PolicyDefinition{
			Name:         "{{.policy_name}}",
			Description:  "Enforce RDS backup retention >= {{.min_backup_retention}} days",
			ResourceType: "rds",
			Filters: []FilterDefinition{
				{Type: "backup-retention-period", Value: "{{.min_backup_retention}}", Op: "lt"},
			},
			Actions: []ActionDefinition{
				{
					Type: "modify-backup-retention",
					Settings: map[string]interface{}{
						"backup_retention_period": "{{.min_backup_retention}}",
					},
					DryRun: true,
				},
				{
					Type: "tag",
					Settings: map[string]interface{}{
						"BackupEnforced":  "true",
						"EnforcementDate": "{{.current_date}}",
					},
				},
			},
			Mode: PolicyModeDefinition{Type: "pull"},
		},
		Examples: []string{
			"Set minimum 7-day backup retention for all RDS instances",
			"Enforce backup policies for compliance",
		},
		Tags:      []string{"compliance", "rds", "backup", "data-protection"},
		CreatedBy: "custodian-killer-team",
	},
}

// Template management functions
type TemplateManager struct {
	templates []PolicyTemplate
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: BuiltInTemplates,
	}
}

// GetAllTemplates returns all available templates
func (tm *TemplateManager) GetAllTemplates() []PolicyTemplate {
	return tm.templates
}

// GetTemplatesByCategory returns templates filtered by category
func (tm *TemplateManager) GetTemplatesByCategory(category string) []PolicyTemplate {
	var filtered []PolicyTemplate
	for _, template := range tm.templates {
		if strings.EqualFold(template.Category, category) {
			filtered = append(filtered, template)
		}
	}
	return filtered
}

// GetTemplatesByResourceType returns templates for a specific resource type
func (tm *TemplateManager) GetTemplatesByResourceType(resourceType string) []PolicyTemplate {
	var filtered []PolicyTemplate
	for _, template := range tm.templates {
		if strings.EqualFold(template.ResourceType, resourceType) {
			filtered = append(filtered, template)
		}
	}
	return filtered
}

// GetTemplateByID returns a specific template by ID
func (tm *TemplateManager) GetTemplateByID(id string) (*PolicyTemplate, error) {
	for _, template := range tm.templates {
		if template.ID == id {
			return &template, nil
		}
	}
	return nil, fmt.Errorf("template with ID '%s' not found", id)
}

// SearchTemplates searches templates by name, description, or tags
func (tm *TemplateManager) SearchTemplates(query string) []PolicyTemplate {
	var results []PolicyTemplate
	query = strings.ToLower(query)

	for _, template := range tm.templates {
		// Search in name, description, and tags
		if strings.Contains(strings.ToLower(template.Name), query) ||
			strings.Contains(strings.ToLower(template.Description), query) {
			results = append(results, template)
			continue
		}

		// Search in tags
		for _, tag := range template.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, template)
				break
			}
		}
	}

	return results
}

// GetCategories returns all available template categories
func (tm *TemplateManager) GetCategories() []string {
	categoryMap := make(map[string]bool)
	for _, template := range tm.templates {
		categoryMap[template.Category] = true
	}

	var categories []string
	for category := range categoryMap {
		categories = append(categories, category)
	}

	return categories
}

// InstantiateTemplate creates a policy from a template with variable substitution
func (tm *TemplateManager) InstantiateTemplate(
	templateID string,
	variables map[string]interface{},
) (PolicyDefinition, error) {
	template, err := tm.GetTemplateByID(templateID)
	if err != nil {
		return PolicyDefinition{}, err
	}

	// Add current date for templates that need it
	variables["current_date"] = time.Now().Format("2006-01-02")

	// Start with the template
	policy := template.Template

	// Substitute variables in name and description
	policy.Name = substituteVariables(policy.Name, variables)
	policy.Description = substituteVariables(policy.Description, variables)

	// Substitute variables in filters
	for i, filter := range policy.Filters {
		if filter.Value != nil {
			if valueStr, ok := filter.Value.(string); ok {
				policy.Filters[i].Value = substituteVariables(valueStr, variables)
			}
		}
	}

	// Substitute variables in actions
	for i, action := range policy.Actions {
		for key, value := range action.Settings {
			if valueStr, ok := value.(string); ok {
				policy.Actions[i].Settings[key] = substituteVariables(valueStr, variables)
			}
		}

		// Substitute action type
		policy.Actions[i].Type = substituteVariables(action.Type, variables)
	}

	return policy, nil
}

// ValidateTemplateVariables checks if all required variables are provided
func (tm *TemplateManager) ValidateTemplateVariables(
	templateID string,
	variables map[string]interface{},
) error {
	template, err := tm.GetTemplateByID(templateID)
	if err != nil {
		return err
	}

	for _, variable := range template.Variables {
		if variable.Required {
			if _, exists := variables[variable.Name]; !exists {
				return fmt.Errorf("required variable '%s' is missing", variable.Name)
			}
		}
	}

	return nil
}

// Simple variable substitution (in a real implementation, you'd use a proper template engine)
func substituteVariables(text string, variables map[string]interface{}) string {
	result := text
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

// GetPopularTemplates returns the most commonly used templates
func (tm *TemplateManager) GetPopularTemplates() []PolicyTemplate {
	var popular []PolicyTemplate
	for _, template := range tm.templates {
		// Check if template has "popular" tag
		for _, tag := range template.Tags {
			if tag == "popular" {
				popular = append(popular, template)
				break
			}
		}
	}
	return popular
}

// AddCustomTemplate allows adding user-defined templates
func (tm *TemplateManager) AddCustomTemplate(template PolicyTemplate) error {
	// Validate template
	if template.ID == "" || template.Name == "" {
		return fmt.Errorf("template must have ID and name")
	}

	// Check for duplicate ID
	for _, existing := range tm.templates {
		if existing.ID == template.ID {
			return fmt.Errorf("template with ID '%s' already exists", template.ID)
		}
	}

	tm.templates = append(tm.templates, template)
	return nil
}
