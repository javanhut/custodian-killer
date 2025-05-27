// wizard/utils.go
package wizard

import (
	"bufio"
	"custodian-killer/filters"
	"custodian-killer/storage"
	"custodian-killer/templates"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Policy represents a policy structure (should match the one in main package)
type Policy struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	ResourceType string                 `json:"resource_type"`
	Filters      []Filter               `json:"filters"`
	Actions      []Action               `json:"actions"`
	Mode         PolicyMode             `json:"mode"`
	Tags         map[string]string      `json:"tags,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Filter represents a filter structure
type Filter struct {
	Type     string      `json:"type"`
	Key      string      `json:"key,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	Op       string      `json:"op,omitempty"`
	Required bool        `json:"required,omitempty"`
	Negate   bool        `json:"negate,omitempty"`
}

// Action represents an action structure
type Action struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	DryRun   bool                   `json:"dry_run"`
}

// PolicyMode represents policy execution mode
type PolicyMode struct {
	Type     string            `json:"type"`
	Schedule string            `json:"schedule,omitempty"`
	Settings map[string]string `json:"settings,omitempty"`
}

// ResourceType represents supported resource types
type ResourceType struct {
	Name        string   `json:"name"`
	Service     string   `json:"service"`
	Description string   `json:"description"`
	Filters     []string `json:"available_filters"`
	Actions     []string `json:"available_actions"`
}

// SupportedResources contains all supported AWS resource types
var SupportedResources = map[string]ResourceType{
	"ec2": {
		Name:        "ec2",
		Service:     "EC2",
		Description: "EC2 Instances",
		Filters: []string{
			"instance-state",
			"tag",
			"instance-type",
			"launch-time",
			"vpc-id",
			"subnet-id",
		},
		Actions: []string{"stop", "terminate", "tag", "detach-volume", "create-snapshot"},
	},
	"s3": {
		Name:        "s3",
		Service:     "S3",
		Description: "S3 Buckets",
		Filters: []string{
			"bucket-name",
			"tag",
			"creation-date",
			"encryption",
			"public-access",
			"versioning",
		},
		Actions: []string{
			"delete",
			"tag",
			"encrypt",
			"block-public-access",
			"enable-versioning",
		},
	},
	"rds": {
		Name:        "rds",
		Service:     "RDS",
		Description: "RDS Instances",
		Filters: []string{
			"engine",
			"instance-class",
			"tag",
			"backup-retention",
			"multi-az",
			"encryption",
		},
		Actions: []string{
			"stop",
			"delete",
			"tag",
			"create-snapshot",
			"modify-backup-retention",
		},
	},
	"lambda": {
		Name:        "lambda",
		Service:     "Lambda",
		Description: "Lambda Functions",
		Filters: []string{
			"runtime",
			"last-modified",
			"tag",
			"memory-size",
			"timeout",
			"environment",
		},
		Actions: []string{"delete", "tag", "update-configuration", "update-environment"},
	},
	"iam": {
		Name:        "iam",
		Service:     "IAM",
		Description: "IAM Users, Roles, and Policies",
		Filters:     []string{"creation-date", "last-used", "attached-policies", "tag", "path"},
		Actions:     []string{"delete", "tag", "detach-policy", "add-to-group"},
	},
	"vpc": {
		Name:        "vpc",
		Service:     "VPC",
		Description: "VPC Resources",
		Filters:     []string{"vpc-id", "tag", "state", "cidr-block", "is-default"},
		Actions:     []string{"delete", "tag", "modify-attribute"},
	},
	"ebs": {
		Name:        "ebs",
		Service:     "EC2",
		Description: "EBS Volumes",
		Filters: []string{
			"volume-type",
			"state",
			"tag",
			"creation-time",
			"attachment-state",
			"encrypted",
		},
		Actions: []string{"delete", "tag", "create-snapshot", "detach", "encrypt"},
	},
	"elb": {
		Name:        "elb",
		Service:     "ELB",
		Description: "Load Balancers",
		Filters:     []string{"load-balancer-name", "tag", "scheme", "vpc-id", "state"},
		Actions:     []string{"delete", "tag", "modify-attributes"},
	},
}

// Utility functions for user input
func GetInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func GetChoice(reader *bufio.Reader, min, max int, prompt string) int {
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		choice, err := strconv.Atoi(strings.TrimSpace(input))

		if err != nil || choice < min || choice > max {
			fmt.Printf("Please enter a number between %d and %d.\n", min, max)
			continue
		}

		return choice
	}
}

func ConfirmSave(reader *bufio.Reader) bool {
	fmt.Print("\n💾 Save this policy? (y/n): ")
	input, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(input)) == "y"
}

// ShowPolicySummary displays a summary of the policy
func ShowPolicySummary(policy Policy) {
	fmt.Println("\n📋 Policy Summary:")
	fmt.Println("═══════════════════")
	fmt.Printf("Name: %s\n", policy.Name)
	fmt.Printf("Description: %s\n", policy.Description)
	fmt.Printf("Resource Type: %s\n", strings.ToUpper(policy.ResourceType))

	fmt.Println("\nFilters:")
	for _, filter := range policy.Filters {
		fmt.Printf("  • %s %s %v\n", filter.Type, filter.Op, filter.Value)
	}

	fmt.Println("\nActions:")
	for _, action := range policy.Actions {
		dryRunStatus := ""
		if action.DryRun {
			dryRunStatus = " (DRY RUN)"
		}
		fmt.Printf("  • %s%s\n", action.Type, dryRunStatus)
	}

	fmt.Printf("\nExecution Mode: %s\n", policy.Mode.Type)
	if policy.Mode.Schedule != "" {
		fmt.Printf("Schedule: %s\n", policy.Mode.Schedule)
	}
}

func SavePolicy(policyStorage storage.PolicyStorage, policy Policy) {
	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized! Policy not saved.")
		return
	}

	// Convert to StoredPolicy (the types DO exist in storage package)
	storedPolicy := storage.StoredPolicy{
		Name:         policy.Name,
		Description:  policy.Description,
		ResourceType: policy.ResourceType,
		Tags:         policy.Tags,
		Metadata:     policy.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    "custodian-killer-user",
		Version:      1,
		Status:       "active",
		Source:       "wizard",
	}

	// Convert filters
	for _, filter := range policy.Filters {
		storedFilter := storage.StoredFilter{
			Type:     filter.Type,
			Key:      filter.Key,
			Value:    filter.Value,
			Op:       filter.Op,
			Required: filter.Required,
			Negate:   filter.Negate,
		}
		storedPolicy.Filters = append(storedPolicy.Filters, storedFilter)
	}

	// Convert actions
	for _, action := range policy.Actions {
		storedAction := storage.StoredAction{
			Type:     action.Type,
			Settings: action.Settings,
			DryRun:   action.DryRun,
		}
		storedPolicy.Actions = append(storedPolicy.Actions, storedAction)
	}

	// Convert mode
	storedPolicy.Mode = storage.StoredPolicyMode{
		Type:     policy.Mode.Type,
		Schedule: policy.Mode.Schedule,
		Settings: policy.Mode.Settings,
	}

	// Save to storage
	if err := policyStorage.SavePolicy(storedPolicy); err != nil {
		fmt.Printf("❌ Failed to save policy: %v\n", err)
		return
	}

	fmt.Printf("✅ Policy '%s' saved successfully!\n", policy.Name)
	fmt.Printf("📁 You can find it in your policies directory\n")
}

// GetResourceTypes returns all supported resource types
func GetResourceTypes() []string {
	var types []string
	for key := range SupportedResources {
		types = append(types, key)
	}
	return types
}

// ConvertTemplatePolicyToPolicy converts template policy definition to main Policy struct
func ConvertTemplatePolicyToPolicy(policyDef templates.PolicyDefinition) Policy {
	var policy Policy
	policy.Name = policyDef.Name
	policy.Description = policyDef.Description
	policy.ResourceType = policyDef.ResourceType
	policy.Tags = policyDef.Tags
	policy.Metadata = policyDef.Metadata

	// Convert filters
	for _, filterDef := range policyDef.Filters {
		filter := Filter{
			Type:     filterDef.Type,
			Key:      filterDef.Key,
			Value:    filterDef.Value,
			Op:       filterDef.Op,
			Required: filterDef.Required,
			Negate:   filterDef.Negate,
		}
		policy.Filters = append(policy.Filters, filter)
	}

	// Convert actions
	for _, actionDef := range policyDef.Actions {
		action := Action{
			Type:     actionDef.Type,
			Settings: actionDef.Settings,
			DryRun:   actionDef.DryRun,
		}
		policy.Actions = append(policy.Actions, action)
	}

	// Convert mode
	policy.Mode = PolicyMode{
		Type:     policyDef.Mode.Type,
		Schedule: policyDef.Mode.Schedule,
		Settings: policyDef.Mode.Settings,
	}

	return policy
}

// ConvertAdvancedFiltersToSimple converts advanced filters to simple filters
func ConvertAdvancedFiltersToSimple(advancedFilters []filters.AdvancedFilter) []Filter {
	var simpleFilters []Filter

	for _, advFilter := range advancedFilters {
		// Convert simple field-based filters
		if advFilter.Field != "" && advFilter.Operator != "" {
			filter := Filter{
				Type:  advFilter.Field,
				Op:    advFilter.Operator,
				Value: advFilter.Value,
			}
			simpleFilters = append(simpleFilters, filter)
		}

		// Handle AND conditions - flatten them
		for _, andFilter := range advFilter.AND {
			if andFilter.Field != "" && andFilter.Operator != "" {
				filter := Filter{
					Type:  andFilter.Field,
					Op:    andFilter.Operator,
					Value: andFilter.Value,
				}
				simpleFilters = append(simpleFilters, filter)
			}
		}

		// Handle OR conditions - for now, just take the first one
		// In a full implementation, you'd need to support OR logic in the filter system
		if len(advFilter.OR) > 0 {
			firstOr := advFilter.OR[0]
			if firstOr.Field != "" && firstOr.Operator != "" {
				filter := Filter{
					Type:  firstOr.Field,
					Op:    firstOr.Operator,
					Value: firstOr.Value,
				}
				simpleFilters = append(simpleFilters, filter)
			}
		}

		// Handle NOT conditions - mark as negated
		if advFilter.NOT != nil && advFilter.NOT.Field != "" {
			filter := Filter{
				Type:   advFilter.NOT.Field,
				Op:     advFilter.NOT.Operator,
				Value:  advFilter.NOT.Value,
				Negate: true,
			}
			simpleFilters = append(simpleFilters, filter)
		}
	}

	return simpleFilters
}

// CreatePolicyMode creates a policy execution mode (shared function)
func CreatePolicyMode(reader *bufio.Reader) PolicyMode {
	fmt.Println("\n⏱️  How should this policy run?")
	fmt.Println("1. 🔄 On-demand (run when you tell it to)")
	fmt.Println("2. ⏰ Scheduled (runs automatically)")
	fmt.Println("3. 📡 Event-driven (responds to AWS events)")

	choice := GetChoice(reader, 1, 3, "Choose execution mode: ")

	var mode PolicyMode
	switch choice {
	case 1:
		mode.Type = "pull"
	case 2:
		mode.Type = "periodic"
		mode.Schedule = GetInput(
			reader,
			"Schedule (cron format, e.g., '0 2 * * *' for daily at 2am): ",
		)
	case 3:
		mode.Type = "event"
		// TODO: Add event configuration
	}

	return mode
}

// ValidatePolicy performs basic policy validation
func ValidatePolicy(policy Policy) []string {
	var errors []string

	if policy.Name == "" {
		errors = append(errors, "policy name cannot be empty")
	}

	if policy.ResourceType == "" {
		errors = append(errors, "resource type cannot be empty")
	}

	// Check if resource type is supported
	if _, exists := SupportedResources[policy.ResourceType]; !exists {
		errors = append(errors, fmt.Sprintf("unsupported resource type: %s", policy.ResourceType))
	}

	// Validate filters
	for i, filter := range policy.Filters {
		if filter.Type == "" {
			errors = append(errors, fmt.Sprintf("filter %d: type cannot be empty", i+1))
		}
	}

	// Validate actions
	for i, action := range policy.Actions {
		if action.Type == "" {
			errors = append(errors, fmt.Sprintf("action %d: type cannot be empty", i+1))
		}
	}

	return errors
}

// ShowValidationErrors displays validation errors
func ShowValidationErrors(errors []string) {
	if len(errors) == 0 {
		return
	}

	fmt.Println("\n⚠️  Policy Validation Errors:")
	for i, err := range errors {
		fmt.Printf("%d. %s\n", i+1, err)
	}
}

// PolicyStats provides statistics about a policy
type PolicyStats struct {
	FilterCount   int
	ActionCount   int
	DryRunActions int
	LiveActions   int
	Complexity    string
	RiskLevel     string
}

// CalculatePolicyStats calculates statistics for a policy
func CalculatePolicyStats(policy Policy) PolicyStats {
	stats := PolicyStats{
		FilterCount: len(policy.Filters),
		ActionCount: len(policy.Actions),
	}

	// Count dry-run vs live actions
	for _, action := range policy.Actions {
		if action.DryRun {
			stats.DryRunActions++
		} else {
			stats.LiveActions++
		}
	}

	// Calculate complexity
	complexity := stats.FilterCount + stats.ActionCount
	switch {
	case complexity <= 3:
		stats.Complexity = "Simple"
	case complexity <= 6:
		stats.Complexity = "Moderate"
	default:
		stats.Complexity = "Complex"
	}

	// Calculate risk level
	if stats.LiveActions == 0 {
		stats.RiskLevel = "Safe (Dry-run only)"
	} else if stats.LiveActions <= 2 {
		stats.RiskLevel = "Low"
	} else if stats.LiveActions <= 4 {
		stats.RiskLevel = "Medium"
	} else {
		stats.RiskLevel = "High"
	}

	return stats
}

// ShowPolicyStats displays policy statistics
func ShowPolicyStats(policy Policy) {
	stats := CalculatePolicyStats(policy)

	fmt.Println("\n📊 Policy Statistics:")
	fmt.Printf("   Filters: %d | Actions: %d\n", stats.FilterCount, stats.ActionCount)
	fmt.Printf("   Dry-run: %d | Live: %d\n", stats.DryRunActions, stats.LiveActions)
	fmt.Printf("   Complexity: %s\n", stats.Complexity)
	fmt.Printf("   Risk Level: %s\n", stats.RiskLevel)
}
