package reports

import (
	"time"
)

// ExecutionResult represents policy execution results
type ExecutionResult struct {
	PolicyName       string           `json:"policy_name"`
	StartTime        time.Time        `json:"start_time"`
	EndTime          time.Time        `json:"end_time"`
	Duration         time.Duration    `json:"duration"`
	ResourceType     string           `json:"resource_type"`
	DryRun           bool             `json:"dry_run"`
	Success          bool             `json:"success"`
	ResourcesFound   int              `json:"resources_found"`
	ResourcesMatched int              `json:"resources_matched"`
	ActionResults    []ActionResult   `json:"action_results"`
	Errors           []string         `json:"errors"`
	Summary          ExecutionSummary `json:"summary"`
}

// ActionResult represents individual action results
type ActionResult struct {
	Action        string                 `json:"action"`
	ResourceID    string                 `json:"resource_id"`
	ResourceType  string                 `json:"resource_type"`
	Success       bool                   `json:"success"`
	DryRun        bool                   `json:"dry_run"`
	Message       string                 `json:"message"`
	Details       map[string]interface{} `json:"details,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	ExecutionTime time.Duration          `json:"execution_time"`
}

// ExecutionSummary provides high-level execution statistics
type ExecutionSummary struct {
	TotalActions            int     `json:"total_actions"`
	SuccessfulActions       int     `json:"successful_actions"`
	FailedActions           int     `json:"failed_actions"`
	ResourcesModified       int     `json:"resources_modified"`
	EstimatedMonthlySavings float64 `json:"estimated_monthly_savings"`
	SecurityImprovements    int     `json:"security_improvements"`
}
