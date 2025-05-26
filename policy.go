package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Policy represents a custodian policy
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

// PolicyMode defines how the policy should run
type PolicyMode struct {
	Type     string            `json:"type"`               // pull, push, periodic, event
	Schedule string            `json:"schedule,omitempty"` // for periodic mode
	Settings map[string]string `json:"settings,omitempty"`
}

// Filter defines resource filtering criteria
type Filter struct {
	Type     string      `json:"type"`
	Key      string      `json:"key,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	Op       string      `json:"op,omitempty"`       // eq, ne, in, not-in, gt, lt, gte, lte, contains, etc.
	Required bool        `json:"required,omitempty"` // whether this filter is required
	Negate   bool        `json:"negate,omitempty"`   // negate the filter result
}

// Action defines what to do with filtered resources
type Action struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	DryRun   bool                   `json:"dry_run"`
}

// ResourceType defines supported AWS resource types
type ResourceType struct {
	Name        string   `json:"name"`
	Service     string   `json:"service"`
	Description string   `json:"description"`
	Filters     []string `json:"available_filters"`
	Actions     []string `json:"available_actions"`
}

// PolicyTemplate for common use cases
type PolicyTemplate struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     string   `json:"category"` // security, cost, compliance, cleanup
	ResourceType string   `json:"resource_type"`
	Template     Policy   `json:"template"`
	Variables    []string `json:"variables"` // variables that can be customized
}

// Supported AWS resource types
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

// Common policy templates
var PolicyTemplates = []PolicyTemplate{
	{
		Name:         "unused-ec2-instances",
		Description:  "Find and stop/terminate unused EC2 instances",
		Category:     "cost",
		ResourceType: "ec2",
		Variables:    []string{"days_unused", "action_type"},
		Template: Policy{
			ResourceType: "ec2",
			Filters: []Filter{
				{Type: "instance-state", Value: "running", Op: "eq"},
				{Type: "cpu-utilization", Value: 5, Op: "lt"},
			},
			Actions: []Action{
				{Type: "stop", DryRun: true},
			},
		},
	},
	{
		Name:         "untagged-resources",
		Description:  "Find resources missing required tags",
		Category:     "compliance",
		ResourceType: "ec2",
		Variables:    []string{"required_tags"},
		Template: Policy{
			ResourceType: "ec2",
			Filters: []Filter{
				{Type: "tag-missing", Key: "Environment", Required: true},
				{Type: "tag-missing", Key: "Owner", Required: true},
			},
			Actions: []Action{
				{
					Type: "tag",
					Settings: map[string]interface{}{
						"Environment": "unknown",
						"Owner":       "unassigned",
					},
				},
			},
		},
	},
	{
		Name:         "public-s3-buckets",
		Description:  "Find and secure public S3 buckets",
		Category:     "security",
		ResourceType: "s3",
		Variables:    []string{"action_type"},
		Template: Policy{
			ResourceType: "s3",
			Filters: []Filter{
				{Type: "public-access", Value: true, Op: "eq"},
			},
			Actions: []Action{
				{Type: "block-public-access", DryRun: true},
			},
		},
	},
}

// PolicyEngine handles policy execution
type PolicyEngine struct {
	policies []Policy
	config   Config
}

// Config holds configuration settings
type Config struct {
	AWSRegion    string            `json:"aws_region"`
	AWSProfile   string            `json:"aws_profile"`
	DryRun       bool              `json:"dry_run"`
	OutputFormat string            `json:"output_format"`
	Settings     map[string]string `json:"settings"`
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine(config Config) *PolicyEngine {
	return &PolicyEngine{
		policies: []Policy{},
		config:   config,
	}
}

// AddPolicy adds a policy to the engine
func (pe *PolicyEngine) AddPolicy(policy Policy) {
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()
	pe.policies = append(pe.policies, policy)
}

// GetPolicy retrieves a policy by name
func (pe *PolicyEngine) GetPolicy(name string) (*Policy, error) {
	for i, policy := range pe.policies {
		if policy.Name == name {
			return &pe.policies[i], nil
		}
	}
	return nil, fmt.Errorf("policy '%s' not found", name)
}

// ListPolicies returns all policies
func (pe *PolicyEngine) ListPolicies() []Policy {
	return pe.policies
}

// ValidatePolicy validates a policy structure
func (pe *PolicyEngine) ValidatePolicy(policy Policy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}

	if policy.ResourceType == "" {
		return fmt.Errorf("resource type cannot be empty")
	}

	// Check if resource type is supported
	if _, exists := SupportedResources[policy.ResourceType]; !exists {
		return fmt.Errorf("unsupported resource type: %s", policy.ResourceType)
	}

	// Validate filters
	for _, filter := range policy.Filters {
		if filter.Type == "" {
			return fmt.Errorf("filter type cannot be empty")
		}
	}

	// Validate actions
	for _, action := range policy.Actions {
		if action.Type == "" {
			return fmt.Errorf("action type cannot be empty")
		}
	}

	return nil
}

// ExecutePolicy runs a policy against AWS resources
func (pe *PolicyEngine) ExecutePolicy(policyName string, dryRun bool) error {
	policy, err := pe.GetPolicy(policyName)
	if err != nil {
		return err
	}

	fmt.Printf("🚀 Executing policy: %s\n", policy.Name)
	fmt.Printf("📝 Description: %s\n", policy.Description)
	fmt.Printf("🎯 Resource Type: %s\n", policy.ResourceType)
	fmt.Printf("🔍 Dry Run: %t\n", dryRun)

	// TODO: Implement actual AWS API calls
	fmt.Println("(AWS API integration coming up...)")

	return nil
}

// PolicyToJSON converts a policy to JSON
func (pe *PolicyEngine) PolicyToJSON(policy Policy) (string, error) {
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// PolicyFromJSON creates a policy from JSON
func (pe *PolicyEngine) PolicyFromJSON(jsonData string) (Policy, error) {
	var policy Policy
	err := json.Unmarshal([]byte(jsonData), &policy)
	return policy, err
}

// GetResourceTypes returns all supported resource types
func GetResourceTypes() []string {
	var types []string
	for key := range SupportedResources {
		types = append(types, key)
	}
	return types
}

// GetResourceType returns details for a specific resource type
func GetResourceType(resourceType string) (ResourceType, error) {
	if rt, exists := SupportedResources[resourceType]; exists {
		return rt, nil
	}
	return ResourceType{}, fmt.Errorf("unsupported resource type: %s", resourceType)
}

// GetPolicyTemplates returns available policy templates
func GetPolicyTemplates(category string) []PolicyTemplate {
	if category == "" {
		return PolicyTemplates
	}

	var filtered []PolicyTemplate
	for _, template := range PolicyTemplates {
		if strings.EqualFold(template.Category, category) {
			filtered = append(filtered, template)
		}
	}
	return filtered
}
