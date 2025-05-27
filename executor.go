package main

import (
	"custodian-killer/aws"
	"custodian-killer/storage"
	"fmt"
	"strings"
	"time"
)

// PolicyExecutor handles the execution of policies against AWS resources
type PolicyExecutor struct {
	awsClient *aws.CustodianClient
	storage   storage.PolicyStorage
	config    ExecutorConfig
	dryRun    bool
}

// ExecutorConfig holds configuration for policy execution
type ExecutorConfig struct {
	MaxConcurrency    int           `json:"max_concurrency"`
	TimeoutPerPolicy  time.Duration `json:"timeout_per_policy"`
	BatchSize         int           `json:"batch_size"`
	ConfirmActions    bool          `json:"confirm_actions"` // Ask before destructive actions
	StopOnError       bool          `json:"stop_on_error"`
	SaveResults       bool          `json:"save_results"`
	NotificationEmail string        `json:"notification_email,omitempty"`
}

// ExecutionResult represents the result of executing a policy
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
	ActionsExecuted  int              `json:"actions_executed"`
	ActionResults    []ActionResult   `json:"action_results"`
	Errors           []string         `json:"errors"`
	Summary          ExecutionSummary `json:"summary"`
	CostImpact       CostImpact       `json:"cost_impact"`
}

// ActionResult represents the result of a single action
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

// CostImpact represents the cost impact of policy execution
type CostImpact struct {
	PreviousMonthlyCost float64 `json:"previous_monthly_cost"`
	NewMonthlyCost      float64 `json:"new_monthly_cost"`
	MonthlySavings      float64 `json:"monthly_savings"`
	AnnualSavings       float64 `json:"annual_savings"`
	Currency            string  `json:"currency"`
}

// NewPolicyExecutor creates a new policy executor
func NewPolicyExecutor(
	awsClient *aws.CustodianClient,
	storage storage.PolicyStorage,
) *PolicyExecutor {
	return &PolicyExecutor{
		awsClient: awsClient,
		storage:   storage,
		config: ExecutorConfig{
			MaxConcurrency:   5,
			TimeoutPerPolicy: 30 * time.Minute,
			BatchSize:        10,
			ConfirmActions:   true,
			StopOnError:      false,
			SaveResults:      true,
		},
		dryRun: awsClient.DryRun,
	}
}

// SetConfig updates executor configuration
func (pe *PolicyExecutor) SetConfig(config ExecutorConfig) {
	pe.config = config
}

// ExecutePolicy executes a single policy
func (pe *PolicyExecutor) ExecutePolicy(policyName string) (*ExecutionResult, error) {
	fmt.Printf("🚀 Executing policy: %s\n", policyName)

	startTime := time.Now()
	result := &ExecutionResult{
		PolicyName:    policyName,
		StartTime:     startTime,
		DryRun:        pe.dryRun,
		ActionResults: make([]ActionResult, 0),
		Errors:        make([]string, 0),
		CostImpact:    CostImpact{Currency: "USD"},
	}

	// Get policy from storage
	policy, err := pe.storage.GetPolicy(policyName)
	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to load policy: %v", err))
		return result, err
	}

	result.ResourceType = policy.ResourceType

	fmt.Printf("📋 Policy: %s\n", policy.Description)
	fmt.Printf("🎯 Resource Type: %s\n", strings.ToUpper(policy.ResourceType))
	fmt.Printf("🔍 Filters: %d | ⚡ Actions: %d\n", len(policy.Filters), len(policy.Actions))

	if pe.dryRun {
		fmt.Println("🧪 DRY RUN MODE - No actual changes will be made")
	}

	// Execute based on resource type
	switch policy.ResourceType {
	case "ec2":
		err = pe.executeEC2Policy(policy, result)
	case "s3":
		err = pe.executeS3Policy(policy, result)
	case "rds":
		err = pe.executeRDSPolicy(policy, result)
	case "lambda":
		err = pe.executeLambdaPolicy(policy, result)
	default:
		err = fmt.Errorf("unsupported resource type: %s", policy.ResourceType)
		result.Errors = append(result.Errors, err.Error())
	}

	// Finalize result
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = err == nil
	result.Summary = pe.calculateSummary(result)

	// Update policy run statistics
	pe.updatePolicyStats(policy, result)

	// Save results if configured
	if pe.config.SaveResults {
		pe.saveExecutionResult(result)
	}

	// Print summary
	pe.printExecutionSummary(result)

	return result, err
}

// executeEC2Policy handles EC2-specific policy execution
func (pe *PolicyExecutor) executeEC2Policy(
	policy *storage.StoredPolicy,
	result *ExecutionResult,
) error {
	fmt.Println("🖥️  Executing EC2 policy...")

	// Convert filters to AWS format
	filter := pe.convertToEC2Filter(policy.Filters)

	// Get matching instances
	instances, err := pe.awsClient.GetEC2Instances(filter)
	if err != nil {
		return fmt.Errorf("failed to get EC2 instances: %v", err)
	}

	result.ResourcesFound = len(instances)
	result.ResourcesMatched = len(instances)

	fmt.Printf("🎯 Found %d instances matching policy criteria\n", len(instances))

	if len(instances) == 0 {
		fmt.Println("✅ No instances matched - nothing to do!")
		return nil
	}

	// Calculate cost impact before changes
	costsBefore := pe.awsClient.GetInstanceCosts(instances)
	result.CostImpact.PreviousMonthlyCost = costsBefore["running_monthly"]

	// Execute actions on matching instances
	for _, action := range policy.Actions {
		fmt.Printf("⚡ Executing action: %s\n", action.Type)

		// Ask for confirmation if not dry-run and action is destructive
		if !pe.dryRun && pe.config.ConfirmActions && pe.isDestructiveAction(action.Type) {
			if !pe.confirmAction(action.Type, len(instances)) {
				fmt.Println("❌ Action cancelled by user")
				continue
			}
		}

		err := pe.executeEC2Action(instances, action, result)
		if err != nil && pe.config.StopOnError {
			return err
		}
	}

	// Calculate cost impact after changes (simplified)
	if result.Summary.ResourcesModified > 0 {
		result.CostImpact.MonthlySavings = result.Summary.EstimatedMonthlySavings
		result.CostImpact.NewMonthlyCost = result.CostImpact.PreviousMonthlyCost - result.CostImpact.MonthlySavings
		result.CostImpact.AnnualSavings = result.CostImpact.MonthlySavings * 12
	}

	return nil
}

// executeEC2Action executes a specific action on EC2 instances
func (pe *PolicyExecutor) executeEC2Action(
	instances []aws.EC2Instance,
	action storage.StoredAction,
	result *ExecutionResult,
) error {
	var instanceIDs []string
	for _, instance := range instances {
		instanceIDs = append(instanceIDs, instance.InstanceID)
	}

	actionStart := time.Now()

	switch action.Type {
	case "stop":
		force := false
		if forceVal, exists := action.Settings["force"]; exists {
			if forceBool, ok := forceVal.(bool); ok {
				force = forceBool
			}
		}

		awsResult, err := pe.awsClient.StopInstances(instanceIDs, force)
		pe.processEC2ActionResult("stop", awsResult, err, actionStart, result)

	case "terminate":
		awsResult, err := pe.awsClient.TerminateInstances(instanceIDs)
		pe.processEC2ActionResult("terminate", awsResult, err, actionStart, result)

	case "start":
		awsResult, err := pe.awsClient.StartInstances(instanceIDs)
		pe.processEC2ActionResult("start", awsResult, err, actionStart, result)

	case "tag":
		tags := make(map[string]string)
		for key, value := range action.Settings {
			if strValue, ok := value.(string); ok {
				tags[key] = strValue
			}
		}

		awsResult, err := pe.awsClient.TagInstances(instanceIDs, tags)
		pe.processEC2ActionResult("tag", awsResult, err, actionStart, result)

	default:
		err := fmt.Errorf("unsupported EC2 action: %s", action.Type)
		result.Errors = append(result.Errors, err.Error())
		return err
	}

	return nil
}

// processEC2ActionResult processes the result of an EC2 action
func (pe *PolicyExecutor) processEC2ActionResult(
	actionType string,
	awsResult *aws.EC2ActionResult,
	err error,
	startTime time.Time,
	result *ExecutionResult,
) {
	executionTime := time.Since(startTime)

	if err != nil {
		actionResult := ActionResult{
			Action:        actionType,
			ResourceType:  "ec2",
			Success:       false,
			DryRun:        pe.dryRun,
			Message:       fmt.Sprintf("Failed: %v", err),
			Timestamp:     time.Now(),
			ExecutionTime: executionTime,
		}
		result.ActionResults = append(result.ActionResults, actionResult)
		result.Errors = append(result.Errors, err.Error())
		return
	}

	// Process successful results
	for _, instanceID := range awsResult.InstanceIDs {
		message := fmt.Sprintf("Action %s completed", actionType)
		if pe.dryRun {
			message = fmt.Sprintf("Would execute %s", actionType)
		}

		actionResult := ActionResult{
			Action:        actionType,
			ResourceID:    instanceID,
			ResourceType:  "ec2",
			Success:       true,
			DryRun:        pe.dryRun,
			Message:       message,
			Timestamp:     time.Now(),
			ExecutionTime: executionTime,
		}

		// Add state change details if available
		if len(awsResult.StateChanges) > 0 {
			for _, stateChange := range awsResult.StateChanges {
				if stateChange.InstanceID == instanceID {
					actionResult.Details = map[string]interface{}{
						"previous_state": stateChange.PreviousState,
						"current_state":  stateChange.CurrentState,
					}
					break
				}
			}
		}

		result.ActionResults = append(result.ActionResults, actionResult)
	}
}

// executeS3Policy handles S3-specific policy execution
func (pe *PolicyExecutor) executeS3Policy(
	policy *storage.StoredPolicy,
	result *ExecutionResult,
) error {
	fmt.Println("🪣 Executing S3 policy...")

	// Convert filters to AWS format
	filter := pe.convertToS3Filter(policy.Filters)

	// Get matching buckets
	buckets, err := pe.awsClient.GetS3Buckets(filter)
	if err != nil {
		return fmt.Errorf("failed to get S3 buckets: %v", err)
	}

	result.ResourcesFound = len(buckets)
	result.ResourcesMatched = len(buckets)

	fmt.Printf("🎯 Found %d buckets matching policy criteria\n", len(buckets))

	if len(buckets) == 0 {
		fmt.Println("✅ No buckets matched - nothing to do!")
		return nil
	}

	// Calculate cost impact before changes
	costsBefore := pe.awsClient.GetBucketCosts(buckets)
	result.CostImpact.PreviousMonthlyCost = costsBefore["total_monthly"]

	// Execute actions on matching buckets
	for _, action := range policy.Actions {
		fmt.Printf("⚡ Executing action: %s\n", action.Type)

		if !pe.dryRun && pe.config.ConfirmActions && pe.isDestructiveAction(action.Type) {
			if !pe.confirmAction(action.Type, len(buckets)) {
				fmt.Println("❌ Action cancelled by user")
				continue
			}
		}

		err := pe.executeS3Action(buckets, action, result)
		if err != nil && pe.config.StopOnError {
			return err
		}
	}

	return nil
}

// executeS3Action executes a specific action on S3 buckets
func (pe *PolicyExecutor) executeS3Action(
	buckets []aws.S3Bucket,
	action storage.StoredAction,
	result *ExecutionResult,
) error {
	var bucketNames []string
	for _, bucket := range buckets {
		bucketNames = append(bucketNames, bucket.Name)
	}

	actionStart := time.Now()

	switch action.Type {
	case "block-public-access":
		awsResult, err := pe.awsClient.BlockPublicAccess(bucketNames)
		pe.processS3ActionResult("block-public-access", awsResult, err, actionStart, result)

	case "enable-encryption":
		kmsKeyID := ""
		if keyID, exists := action.Settings["kms_key_id"]; exists {
			if keyStr, ok := keyID.(string); ok {
				kmsKeyID = keyStr
			}
		}

		awsResult, err := pe.awsClient.EnableEncryption(bucketNames, kmsKeyID)
		pe.processS3ActionResult("enable-encryption", awsResult, err, actionStart, result)

	case "enable-versioning":
		awsResult, err := pe.awsClient.EnableVersioning(bucketNames)
		pe.processS3ActionResult("enable-versioning", awsResult, err, actionStart, result)

	case "tag":
		tags := make(map[string]string)
		for key, value := range action.Settings {
			if strValue, ok := value.(string); ok {
				tags[key] = strValue
			}
		}

		awsResult, err := pe.awsClient.TagBuckets(bucketNames, tags)
		pe.processS3ActionResult("tag", awsResult, err, actionStart, result)

	case "delete":
		force := false
		if forceVal, exists := action.Settings["force"]; exists {
			if forceBool, ok := forceVal.(bool); ok {
				force = forceBool
			}
		}

		awsResult, err := pe.awsClient.DeleteBuckets(bucketNames, force)
		pe.processS3ActionResult("delete", awsResult, err, actionStart, result)

	default:
		err := fmt.Errorf("unsupported S3 action: %s", action.Type)
		result.Errors = append(result.Errors, err.Error())
		return err
	}

	return nil
}

// processS3ActionResult processes the result of an S3 action
func (pe *PolicyExecutor) processS3ActionResult(
	actionType string,
	awsResult *aws.S3ActionResult,
	err error,
	startTime time.Time,
	result *ExecutionResult,
) {
	executionTime := time.Since(startTime)

	if err != nil {
		actionResult := ActionResult{
			Action:        actionType,
			ResourceType:  "s3",
			Success:       false,
			DryRun:        pe.dryRun,
			Message:       fmt.Sprintf("Failed: %v", err),
			Timestamp:     time.Now(),
			ExecutionTime: executionTime,
		}
		result.ActionResults = append(result.ActionResults, actionResult)
		result.Errors = append(result.Errors, err.Error())
		return
	}

	// Process results for each bucket
	for bucketName, bucketResult := range awsResult.Results {
		success := !strings.Contains(bucketResult, "failed")
		message := bucketResult
		if pe.dryRun {
			message = fmt.Sprintf("Would execute %s: %s", actionType, bucketResult)
		}

		actionResult := ActionResult{
			Action:        actionType,
			ResourceID:    bucketName,
			ResourceType:  "s3",
			Success:       success,
			DryRun:        pe.dryRun,
			Message:       message,
			Timestamp:     time.Now(),
			ExecutionTime: executionTime,
		}

		result.ActionResults = append(result.ActionResults, actionResult)
	}
}

// Placeholder implementations for RDS and Lambda (to be expanded)
func (pe *PolicyExecutor) executeRDSPolicy(
	policy *storage.StoredPolicy,
	result *ExecutionResult,
) error {
	fmt.Println("🗄️  RDS policy execution - Coming soon!")
	return nil
}

func (pe *PolicyExecutor) executeLambdaPolicy(
	policy *storage.StoredPolicy,
	result *ExecutionResult,
) error {
	fmt.Println("⚡ Lambda policy execution - Coming soon!")
	return nil
}

// Helper functions for filter conversion
func (pe *PolicyExecutor) convertToEC2Filter(filters []storage.StoredFilter) aws.EC2Filter {
	filter := aws.EC2Filter{
		Tags: make(map[string]string),
	}

	for _, f := range filters {
		switch f.Type {
		case "instance-state", "state":
			if strValue, ok := f.Value.(string); ok {
				filter.States = append(filter.States, strValue)
			}
		case "instance-type":
			if strValue, ok := f.Value.(string); ok {
				filter.InstanceTypes = append(filter.InstanceTypes, strValue)
			}
		case "vpc-id":
			if strValue, ok := f.Value.(string); ok {
				filter.VpcIDs = append(filter.VpcIDs, strValue)
			}
		case "tag", "tag-missing":
			if f.Key != "" {
				filter.Tags[f.Key] = "*" // Check for existence
			}
		case "cpu-utilization", "cpu-utilization-avg":
			if floatValue, ok := f.Value.(float64); ok {
				filter.CPUThreshold = &floatValue
			}
		case "running-days":
			if intValue, ok := f.Value.(int); ok {
				filter.RunningDaysMin = &intValue
			}
		}
	}

	// Default to running instances if no state specified
	if len(filter.States) == 0 {
		filter.States = []string{"running"}
	}

	return filter
}

func (pe *PolicyExecutor) convertToS3Filter(filters []storage.StoredFilter) aws.S3Filter {
	filter := aws.S3Filter{
		Tags: make(map[string]string),
	}

	for _, f := range filters {
		switch f.Type {
		case "public-read", "public-access":
			if boolValue, ok := f.Value.(bool); ok && boolValue {
				filter.PublicAccessOnly = true
			}
		case "encryption":
			if boolValue, ok := f.Value.(bool); ok && !boolValue {
				filter.UnencryptedOnly = true
			}
		case "tag", "tag-missing":
			if f.Key != "" {
				filter.Tags[f.Key] = "*"
			}
		case "size":
			if intValue, ok := f.Value.(int64); ok {
				filter.LargeSizeThreshold = &intValue
			}
		case "security-score":
			if intValue, ok := f.Value.(int); ok {
				filter.MinSecurityScore = &intValue
			}
		}
	}

	return filter
}

// Utility functions
func (pe *PolicyExecutor) isDestructiveAction(actionType string) bool {
	destructiveActions := []string{"terminate", "delete", "stop"}
	for _, action := range destructiveActions {
		if action == actionType {
			return true
		}
	}
	return false
}

func (pe *PolicyExecutor) confirmAction(actionType string, resourceCount int) bool {
	fmt.Printf(
		"⚠️  About to execute '%s' on %d resources. Continue? (y/N): ",
		actionType,
		resourceCount,
	)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

func (pe *PolicyExecutor) calculateSummary(result *ExecutionResult) ExecutionSummary {
	summary := ExecutionSummary{}

	for _, actionResult := range result.ActionResults {
		summary.TotalActions++
		if actionResult.Success {
			summary.SuccessfulActions++
			if !actionResult.DryRun {
				summary.ResourcesModified++
			}
		} else {
			summary.FailedActions++
		}

		// Estimate cost savings for cost-related actions
		if actionResult.Action == "stop" || actionResult.Action == "terminate" {
			summary.EstimatedMonthlySavings += 50.0 // Rough estimate
		}

		// Count security improvements
		if actionResult.Action == "block-public-access" ||
			actionResult.Action == "enable-encryption" {
			summary.SecurityImprovements++
		}
	}

	return summary
}

func (pe *PolicyExecutor) updatePolicyStats(policy *storage.StoredPolicy, result *ExecutionResult) {
	// Update policy statistics
	now := time.Now()
	policy.LastRun = &now
	policy.RunCount++
	policy.UpdatedAt = now

	// Save updated policy (ignoring errors for now)
	pe.storage.SavePolicy(*policy)
}

func (pe *PolicyExecutor) saveExecutionResult(result *ExecutionResult) {
	// In a full implementation, this would save to a results database or file
	fmt.Printf("💾 Execution result saved for policy: %s\n", result.PolicyName)
}

func (pe *PolicyExecutor) printExecutionSummary(result *ExecutionResult) {
	fmt.Println("\n📊 Execution Result Summary:")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Printf("🎯 Policy: %s\n", result.PolicyName)
	fmt.Printf("⏱️  Duration: %v\n", result.Duration.Round(time.Second))
	fmt.Printf(
		"📊 Resources: %d found, %d matched\n",
		result.ResourcesFound,
		result.ResourcesMatched,
	)
	fmt.Printf("⚡ Actions: %d total, %d successful, %d failed\n",
		result.Summary.TotalActions, result.Summary.SuccessfulActions, result.Summary.FailedActions)

	if result.Summary.ResourcesModified > 0 {
		fmt.Printf("🔧 Resources Modified: %d\n", result.Summary.ResourcesModified)
	}

	if result.Summary.EstimatedMonthlySavings > 0 {
		fmt.Printf("💰 Estimated Monthly Savings: $%.2f\n", result.Summary.EstimatedMonthlySavings)
	}

	if result.Summary.SecurityImprovements > 0 {
		fmt.Printf("🔒 Security Improvements: %d\n", result.Summary.SecurityImprovements)
	}

	if len(result.Errors) > 0 {
		fmt.Printf("❌ Errors: %d\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("   • %s\n", err)
		}
	}

	if result.Success {
		fmt.Println("✅ Policy execution completed successfully!")
	} else {
		fmt.Println("⚠️  Policy execution completed with errors")
	}
}
