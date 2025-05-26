package scanner

import (
	"custodian-killer/storage"
	"fmt"
	"strings"
	"time"
)

// ScanResult represents the result of scanning a policy
type ScanResult struct {
	PolicyName       string            `json:"policy_name"`
	ResourceType     string            `json:"resource_type"`
	ScanTime         time.Time         `json:"scan_time"`
	MatchedResources []MatchedResource `json:"matched_resources"`
	Summary          ScanSummary       `json:"summary"`
	Errors           []string          `json:"errors,omitempty"`
	DryRun           bool              `json:"dry_run"`
	EstimatedCost    *CostEstimate     `json:"estimated_cost,omitempty"`
}

// MatchedResource represents a resource that matched the policy filters
type MatchedResource struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name,omitempty"`
	Type       string                 `json:"type"`
	Region     string                 `json:"region"`
	State      string                 `json:"state,omitempty"`
	Tags       map[string]string      `json:"tags,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Actions    []PlannedAction        `json:"planned_actions"`
	RiskLevel  string                 `json:"risk_level"` // low, medium, high
	Compliance ComplianceStatus       `json:"compliance"`
}

// PlannedAction represents an action that would be taken on a resource
type PlannedAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	DryRun      bool                   `json:"dry_run"`
	Impact      string                 `json:"impact"` // low, medium, high
	Reversible  bool                   `json:"reversible"`
}

// ScanSummary provides high-level statistics about the scan
type ScanSummary struct {
	TotalScanned     int     `json:"total_scanned"`
	MatchedResources int     `json:"matched_resources"`
	ActionsPlanned   int     `json:"actions_planned"`
	HighRiskActions  int     `json:"high_risk_actions"`
	CostSavings      float64 `json:"estimated_cost_savings"`
}

// ComplianceStatus represents compliance information
type ComplianceStatus struct {
	Compliant bool     `json:"compliant"`
	Issues    []string `json:"issues,omitempty"`
	Severity  string   `json:"severity,omitempty"`
}

// CostEstimate represents estimated cost impact
type CostEstimate struct {
	CurrentMonthlyCost float64 `json:"current_monthly_cost"`
	ProjectedSavings   float64 `json:"projected_savings"`
	Currency           string  `json:"currency"`
}

// PolicyScanner handles policy scanning operations
type PolicyScanner struct {
	storage storage.PolicyStorage
	config  ScannerConfig
}

// ScannerConfig holds scanner configuration
type ScannerConfig struct {
	AWSRegion     string `json:"aws_region"`
	AWSProfile    string `json:"aws_profile"`
	DryRunDefault bool   `json:"dry_run_default"`
	MaxResources  int    `json:"max_resources"`
	Timeout       int    `json:"timeout_seconds"`
}

// NewPolicyScanner creates a new policy scanner
func NewPolicyScanner(storage storage.PolicyStorage, config ScannerConfig) *PolicyScanner {
	if config.MaxResources == 0 {
		config.MaxResources = 1000
	}
	if config.Timeout == 0 {
		config.Timeout = 300 // 5 minutes
	}
	if config.AWSRegion == "" {
		config.AWSRegion = "us-east-1"
	}

	return &PolicyScanner{
		storage: storage,
		config:  config,
	}
}

// ScanPolicy scans a specific policy and returns results
func (ps *PolicyScanner) ScanPolicy(policyName string) (*ScanResult, error) {
	fmt.Printf("🔍 Scanning policy: %s\n", policyName)

	// Get policy from storage
	policy, err := ps.storage.GetPolicy(policyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %v", err)
	}

	result := &ScanResult{
		PolicyName:   policy.Name,
		ResourceType: policy.ResourceType,
		ScanTime:     time.Now(),
		DryRun:       true, // Always dry run for scan
		Summary:      ScanSummary{},
	}

	fmt.Printf("📊 Resource Type: %s\n", strings.ToUpper(policy.ResourceType))
	fmt.Printf("🎯 Filters: %d | Actions: %d\n", len(policy.Filters), len(policy.Actions))

	// Mock scanning based on resource type
	switch policy.ResourceType {
	case "ec2":
		err = ps.scanEC2Resources(policy, result)
	case "s3":
		err = ps.scanS3Resources(policy, result)
	case "rds":
		err = ps.scanRDSResources(policy, result)
	case "lambda":
		err = ps.scanLambdaResources(policy, result)
	case "ebs":
		err = ps.scanEBSResources(policy, result)
	default:
		err = ps.scanGenericResources(policy, result)
	}

	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	// Calculate summary
	result.Summary = ps.calculateSummary(result)

	return result, nil
}

// ScanAllPolicies scans all stored policies
func (ps *PolicyScanner) ScanAllPolicies() ([]ScanResult, error) {
	policies, err := ps.storage.ListPolicies()
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %v", err)
	}

	var results []ScanResult
	for _, policy := range policies {
		if policy.Status != "active" {
			continue // Skip inactive policies
		}

		result, err := ps.ScanPolicy(policy.Name)
		if err != nil {
			fmt.Printf("⚠️  Failed to scan policy '%s': %v\n", policy.Name, err)
			continue
		}

		results = append(results, *result)
	}

	return results, nil
}

// Mock resource scanning functions (replace with real AWS SDK calls later)
func (ps *PolicyScanner) scanEC2Resources(policy *storage.StoredPolicy, result *ScanResult) error {
	fmt.Println("🖥️  Scanning EC2 instances...")

	// Mock EC2 instances
	mockInstances := []MatchedResource{
		{
			ID:     "i-1234567890abcdef0",
			Name:   "web-server-1",
			Type:   "t3.micro",
			Region: ps.config.AWSRegion,
			State:  "running",
			Tags: map[string]string{
				"Environment": "dev",
				"Owner":       "john@company.com",
			},
			Properties: map[string]interface{}{
				"instance_type":   "t3.micro",
				"launch_time":     "2024-01-15T10:30:00Z",
				"cpu_utilization": 2.5,
				"running_days":    15,
			},
			RiskLevel: "medium",
			Compliance: ComplianceStatus{
				Compliant: false,
				Issues:    []string{"CPU utilization below threshold"},
				Severity:  "medium",
			},
		},
		{
			ID:     "i-0987654321fedcba0",
			Name:   "database-server",
			Type:   "t3.small",
			Region: ps.config.AWSRegion,
			State:  "running",
			Tags: map[string]string{
				"Environment": "prod",
				"Owner":       "alice@company.com",
			},
			Properties: map[string]interface{}{
				"instance_type":   "t3.small",
				"launch_time":     "2024-01-10T08:15:00Z",
				"cpu_utilization": 1.2,
				"running_days":    20,
			},
			RiskLevel: "high",
			Compliance: ComplianceStatus{
				Compliant: false,
				Issues:    []string{"Very low CPU utilization", "Long running unused instance"},
				Severity:  "high",
			},
		},
	}

	// Apply filters and determine matches
	for _, instance := range mockInstances {
		matched := ps.applyFilters(policy.Filters, instance)
		if matched {
			// Add planned actions
			instance.Actions = ps.planActions(policy.Actions, instance)
			result.MatchedResources = append(result.MatchedResources, instance)
		}
	}

	result.Summary.TotalScanned = len(mockInstances)

	fmt.Printf("✅ Scanned %d EC2 instances, found %d matches\n",
		len(mockInstances), len(result.MatchedResources))

	return nil
}

func (ps *PolicyScanner) scanS3Resources(policy *storage.StoredPolicy, result *ScanResult) error {
	fmt.Println("🪣 Scanning S3 buckets...")

	// Mock S3 buckets
	mockBuckets := []MatchedResource{
		{
			ID:     "my-public-bucket",
			Name:   "my-public-bucket",
			Type:   "s3-bucket",
			Region: ps.config.AWSRegion,
			State:  "active",
			Properties: map[string]interface{}{
				"public_read":   true,
				"public_write":  false,
				"versioning":    false,
				"encryption":    false,
				"creation_date": "2024-01-01T00:00:00Z",
			},
			RiskLevel: "high",
			Compliance: ComplianceStatus{
				Compliant: false,
				Issues:    []string{"Bucket allows public read access", "No encryption enabled"},
				Severity:  "high",
			},
		},
	}

	for _, bucket := range mockBuckets {
		matched := ps.applyFilters(policy.Filters, bucket)
		if matched {
			bucket.Actions = ps.planActions(policy.Actions, bucket)
			result.MatchedResources = append(result.MatchedResources, bucket)
		}
	}

	result.Summary.TotalScanned = len(mockBuckets)

	fmt.Printf("✅ Scanned %d S3 buckets, found %d matches\n",
		len(mockBuckets), len(result.MatchedResources))

	return nil
}

func (ps *PolicyScanner) scanRDSResources(policy *storage.StoredPolicy, result *ScanResult) error {
	fmt.Println("🗄️  Scanning RDS instances...")

	// Mock RDS instances
	mockRDS := []MatchedResource{
		{
			ID:     "db-instance-1",
			Name:   "production-db",
			Type:   "db.t3.micro",
			Region: ps.config.AWSRegion,
			State:  "available",
			Properties: map[string]interface{}{
				"engine":                  "mysql",
				"backup_retention_period": 3,
				"multi_az":                false,
				"encrypted":               false,
			},
			RiskLevel: "medium",
			Compliance: ComplianceStatus{
				Compliant: false,
				Issues:    []string{"Backup retention below policy minimum"},
				Severity:  "medium",
			},
		},
	}

	for _, rds := range mockRDS {
		matched := ps.applyFilters(policy.Filters, rds)
		if matched {
			rds.Actions = ps.planActions(policy.Actions, rds)
			result.MatchedResources = append(result.MatchedResources, rds)
		}
	}

	result.Summary.TotalScanned = len(mockRDS)

	fmt.Printf("✅ Scanned %d RDS instances, found %d matches\n",
		len(mockRDS), len(result.MatchedResources))

	return nil
}

func (ps *PolicyScanner) scanLambdaResources(
	policy *storage.StoredPolicy,
	result *ScanResult,
) error {
	fmt.Println("⚡ Scanning Lambda functions...")
	// Mock Lambda scanning
	result.Summary.TotalScanned = 5
	fmt.Println("✅ Scanned 5 Lambda functions, found 0 matches")
	return nil
}

func (ps *PolicyScanner) scanEBSResources(policy *storage.StoredPolicy, result *ScanResult) error {
	fmt.Println("💾 Scanning EBS volumes...")
	// Mock EBS scanning
	result.Summary.TotalScanned = 8
	fmt.Println("✅ Scanned 8 EBS volumes, found 0 matches")
	return nil
}

func (ps *PolicyScanner) scanGenericResources(
	policy *storage.StoredPolicy,
	result *ScanResult,
) error {
	fmt.Printf("🔍 Scanning %s resources...\n", policy.ResourceType)
	result.Summary.TotalScanned = 3
	fmt.Printf("✅ Scanned 3 %s resources, found 0 matches\n", policy.ResourceType)
	return nil
}

// applyFilters checks if a resource matches the policy filters
func (ps *PolicyScanner) applyFilters(
	filters []storage.StoredFilter,
	resource MatchedResource,
) bool {
	for _, filter := range filters {
		if !ps.evaluateFilter(filter, resource) {
			return false
		}
	}
	return true
}

// evaluateFilter evaluates a single filter against a resource
func (ps *PolicyScanner) evaluateFilter(
	filter storage.StoredFilter,
	resource MatchedResource,
) bool {
	switch filter.Type {
	case "instance-state", "state":
		return resource.State == filter.Value
	case "cpu-utilization", "cpu-utilization-avg":
		if cpuUtil, ok := resource.Properties["cpu_utilization"].(float64); ok {
			threshold := 5.0
			if thresholdVal, ok := filter.Value.(float64); ok {
				threshold = thresholdVal
			}
			return cpuUtil < threshold
		}
	case "running-days":
		if days, ok := resource.Properties["running_days"].(int); ok {
			threshold := 7
			if thresholdVal, ok := filter.Value.(int); ok {
				threshold = thresholdVal
			}
			return days >= threshold
		}
	case "public-read", "public-access":
		if publicRead, ok := resource.Properties["public_read"].(bool); ok {
			return publicRead == filter.Value
		}
	case "backup-retention-period":
		if retention, ok := resource.Properties["backup_retention_period"].(int); ok {
			threshold := 7
			if thresholdVal, ok := filter.Value.(int); ok {
				threshold = thresholdVal
			}
			return retention < threshold
		}
	case "tag-missing":
		if filter.Key != "" {
			_, exists := resource.Tags[filter.Key]
			return !exists
		}
	}

	return true // Default to match if filter not implemented
}

// planActions determines what actions would be taken on a resource
func (ps *PolicyScanner) planActions(
	actions []storage.StoredAction,
	resource MatchedResource,
) []PlannedAction {
	var plannedActions []PlannedAction

	for _, action := range actions {
		planned := PlannedAction{
			Type:     action.Type,
			DryRun:   action.DryRun,
			Settings: action.Settings,
		}

		switch action.Type {
		case "stop":
			planned.Description = fmt.Sprintf("Stop EC2 instance %s", resource.ID)
			planned.Impact = "medium"
			planned.Reversible = true
		case "terminate":
			planned.Description = fmt.Sprintf("Terminate EC2 instance %s", resource.ID)
			planned.Impact = "high"
			planned.Reversible = false
		case "tag":
			planned.Description = fmt.Sprintf("Add tags to resource %s", resource.ID)
			planned.Impact = "low"
			planned.Reversible = true
		case "block-public-access":
			planned.Description = fmt.Sprintf("Block public access on S3 bucket %s", resource.ID)
			planned.Impact = "high"
			planned.Reversible = true
		case "modify-backup-retention":
			planned.Description = fmt.Sprintf("Modify backup retention for RDS %s", resource.ID)
			planned.Impact = "medium"
			planned.Reversible = true
		default:
			planned.Description = fmt.Sprintf("Execute %s on resource %s", action.Type, resource.ID)
			planned.Impact = "medium"
			planned.Reversible = true
		}

		plannedActions = append(plannedActions, planned)
	}

	return plannedActions
}

// calculateSummary calculates summary statistics for the scan result
func (ps *PolicyScanner) calculateSummary(result *ScanResult) ScanSummary {
	summary := ScanSummary{
		MatchedResources: len(result.MatchedResources),
	}

	for _, resource := range result.MatchedResources {
		summary.ActionsPlanned += len(resource.Actions)

		for _, action := range resource.Actions {
			if action.Impact == "high" {
				summary.HighRiskActions++
			}
		}

		// Mock cost savings calculation
		if resource.Type == "t3.micro" {
			summary.CostSavings += 8.76 // ~$8.76/month for t3.micro
		} else if resource.Type == "t3.small" {
			summary.CostSavings += 17.52 // ~$17.52/month for t3.small
		}
	}

	return summary
}
