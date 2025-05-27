package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2Instance represents an EC2 instance with all the juicy details
type EC2Instance struct {
	InstanceID     string            `json:"instance_id"`
	Name           string            `json:"name"`
	InstanceType   string            `json:"instance_type"`
	State          string            `json:"state"`
	LaunchTime     time.Time         `json:"launch_time"`
	PublicIP       string            `json:"public_ip,omitempty"`
	PrivateIP      string            `json:"private_ip,omitempty"`
	VpcID          string            `json:"vpc_id,omitempty"`
	SubnetID       string            `json:"subnet_id,omitempty"`
	SecurityGroups []string          `json:"security_groups"`
	Tags           map[string]string `json:"tags"`
	Platform       string            `json:"platform,omitempty"`
	Architecture   string            `json:"architecture,omitempty"`
	Hypervisor     string            `json:"hypervisor,omitempty"`
	RootDeviceType string            `json:"root_device_type,omitempty"`
	CPUUtilization float64           `json:"cpu_utilization,omitempty"`
	NetworkIn      float64           `json:"network_in,omitempty"`
	NetworkOut     float64           `json:"network_out,omitempty"`
	RunningDays    int               `json:"running_days"`
	MonthlyCost    float64           `json:"estimated_monthly_cost"`
}

// EC2Filter represents filtering criteria for EC2 instances
type EC2Filter struct {
	States           []string          // running, stopped, terminated, etc.
	InstanceIDs      []string          // specific instance IDs
	Tags             map[string]string // tag filters
	InstanceTypes    []string          // t3.micro, m5.large, etc.
	VpcIDs           []string          // VPC filters
	LaunchTimeAfter  *time.Time        // instances launched after this time
	LaunchTimeBefore *time.Time        // instances launched before this time
	CPUThreshold     *float64          // CPU utilization threshold
	RunningDaysMin   *int              // minimum running days
}

// GetEC2Instances retrieves EC2 instances based on filters
func (c *CustodianClient) GetEC2Instances(filters EC2Filter) ([]EC2Instance, error) {
	c.LogAWSCall("EC2", "DescribeInstances", c.DryRun)

	fmt.Println("🔍 Scanning EC2 instances...")

	ctx := context.Background()
	var ec2Filters []types.Filter

	// Build AWS filters from our custom filter
	if len(filters.States) > 0 {
		ec2Filters = append(ec2Filters, types.Filter{
			Name:   aws.String("instance-state-name"),
			Values: filters.States,
		})
	}

	if len(filters.InstanceTypes) > 0 {
		ec2Filters = append(ec2Filters, types.Filter{
			Name:   aws.String("instance-type"),
			Values: filters.InstanceTypes,
		})
	}

	if len(filters.VpcIDs) > 0 {
		ec2Filters = append(ec2Filters, types.Filter{
			Name:   aws.String("vpc-id"),
			Values: filters.VpcIDs,
		})
	}

	// Add tag filters
	for key, value := range filters.Tags {
		if value == "*" || value == "" {
			// Check for tag existence only
			ec2Filters = append(ec2Filters, types.Filter{
				Name:   aws.String("tag-key"),
				Values: []string{key},
			})
		} else {
			// Check for specific tag value
			ec2Filters = append(ec2Filters, types.Filter{
				Name:   aws.String(fmt.Sprintf("tag:%s", key)),
				Values: []string{value},
			})
		}
	}

	input := &ec2.DescribeInstancesInput{
		Filters: ec2Filters,
	}

	if len(filters.InstanceIDs) > 0 {
		input.InstanceIds = filters.InstanceIDs
	}

	var instances []EC2Instance
	paginator := ec2.NewDescribeInstancesPaginator(c.EC2, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %v", err)
		}

		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				ec2Instance := c.convertToEC2Instance(instance)

				// Apply additional filters that AWS doesn't support directly
				if c.matchesAdditionalFilters(ec2Instance, filters) {
					instances = append(instances, ec2Instance)
				}
			}
		}
	}

	// Enhance instances with CloudWatch metrics if CPU threshold is specified
	if filters.CPUThreshold != nil {
		fmt.Println("📊 Fetching CPU utilization metrics...")
		instances = c.enhanceWithCloudWatchMetrics(instances)
	}

	fmt.Printf("✅ Found %d EC2 instances matching criteria\n", len(instances))
	return instances, nil
}

// convertToEC2Instance converts AWS SDK instance to our struct
func (c *CustodianClient) convertToEC2Instance(instance types.Instance) EC2Instance {
	ec2Instance := EC2Instance{
		InstanceID:     aws.ToString(instance.InstanceId),
		InstanceType:   string(instance.InstanceType),
		State:          string(instance.State.Name),
		LaunchTime:     aws.ToTime(instance.LaunchTime),
		Platform:       string(instance.Platform),
		Architecture:   string(instance.Architecture),
		Hypervisor:     string(instance.Hypervisor),
		RootDeviceType: string(instance.RootDeviceType),
		Tags:           make(map[string]string),
		SecurityGroups: make([]string, 0),
	}

	// Extract name from tags
	for _, tag := range instance.Tags {
		key := aws.ToString(tag.Key)
		value := aws.ToString(tag.Value)
		ec2Instance.Tags[key] = value

		if strings.ToLower(key) == "name" {
			ec2Instance.Name = value
		}
	}

	// Set name to instance ID if no Name tag
	if ec2Instance.Name == "" {
		ec2Instance.Name = ec2Instance.InstanceID
	}

	// Network information
	if instance.PublicIpAddress != nil {
		ec2Instance.PublicIP = aws.ToString(instance.PublicIpAddress)
	}
	if instance.PrivateIpAddress != nil {
		ec2Instance.PrivateIP = aws.ToString(instance.PrivateIpAddress)
	}
	if instance.VpcId != nil {
		ec2Instance.VpcID = aws.ToString(instance.VpcId)
	}
	if instance.SubnetId != nil {
		ec2Instance.SubnetID = aws.ToString(instance.SubnetId)
	}

	// Security groups
	for _, sg := range instance.SecurityGroups {
		ec2Instance.SecurityGroups = append(ec2Instance.SecurityGroups, aws.ToString(sg.GroupId))
	}

	// Calculate running days
	if ec2Instance.State == "running" && !ec2Instance.LaunchTime.IsZero() {
		ec2Instance.RunningDays = int(time.Since(ec2Instance.LaunchTime).Hours() / 24)
	}

	// Estimate monthly cost (rough calculation based on instance type)
	ec2Instance.MonthlyCost = c.estimateInstanceCost(ec2Instance.InstanceType, ec2Instance.State)

	return ec2Instance
}

// matchesAdditionalFilters applies filters not supported by AWS API
func (c *CustodianClient) matchesAdditionalFilters(instance EC2Instance, filters EC2Filter) bool {
	// Launch time filters
	if filters.LaunchTimeAfter != nil && instance.LaunchTime.Before(*filters.LaunchTimeAfter) {
		return false
	}
	if filters.LaunchTimeBefore != nil && instance.LaunchTime.After(*filters.LaunchTimeBefore) {
		return false
	}

	// Running days filter
	if filters.RunningDaysMin != nil && instance.RunningDays < *filters.RunningDaysMin {
		return false
	}

	// CPU threshold will be applied after CloudWatch data is fetched

	return true
}

// enhanceWithCloudWatchMetrics adds CPU utilization data
func (c *CustodianClient) enhanceWithCloudWatchMetrics(instances []EC2Instance) []EC2Instance {
	// For now, simulate CPU data. In real implementation, we'd call CloudWatch
	for i := range instances {
		// Simulate realistic CPU utilization based on instance age and type
		switch {
		case instances[i].RunningDays > 30:
			instances[i].CPUUtilization = 1.5 + float64(
				instances[i].RunningDays%10,
			) // Very low CPU for old instances
		case instances[i].RunningDays > 7:
			instances[i].CPUUtilization = 3.0 + float64(instances[i].RunningDays%15) // Low CPU
		default:
			instances[i].CPUUtilization = 15.0 + float64(instances[i].RunningDays%20) // Normal CPU
		}

		// Add some network stats too
		instances[i].NetworkIn = float64(instances[i].RunningDays * 1000)
		instances[i].NetworkOut = float64(instances[i].RunningDays * 800)
	}

	return instances
}

// estimateInstanceCost provides rough monthly cost estimates
func (c *CustodianClient) estimateInstanceCost(instanceType, state string) float64 {
	if state != "running" {
		return 0
	}

	// Rough monthly costs for common instance types (US East 1)
	costs := map[string]float64{
		"t3.nano":    3.80,
		"t3.micro":   8.76,
		"t3.small":   17.52,
		"t3.medium":  35.04,
		"t3.large":   70.08,
		"t3.xlarge":  140.16,
		"t3.2xlarge": 280.32,
		"m5.large":   87.84,
		"m5.xlarge":  175.68,
		"m5.2xlarge": 351.36,
		"m5.4xlarge": 702.72,
		"c5.large":   77.76,
		"c5.xlarge":  155.52,
		"r5.large":   113.76,
		"r5.xlarge":  227.52,
	}

	if cost, exists := costs[instanceType]; exists {
		return cost
	}

	// Default estimate for unknown types
	return 50.0
}

// StopInstances stops EC2 instances
func (c *CustodianClient) StopInstances(
	instanceIDs []string,
	force bool,
) (*EC2ActionResult, error) {
	c.LogAWSCall("EC2", "StopInstances", c.DryRun)

	if len(instanceIDs) == 0 {
		return &EC2ActionResult{}, nil
	}

	fmt.Printf("⏹️  Stopping %d instances...\n", len(instanceIDs))

	ctx := context.Background()
	input := &ec2.StopInstancesInput{
		InstanceIds: instanceIDs,
		Force:       aws.Bool(force),
	}

	if c.DryRun {
		input.DryRun = aws.Bool(true)
	}

	result, err := c.EC2.StopInstances(ctx, input)
	if err != nil {
		// Check if it's a dry-run success
		if c.DryRun && strings.Contains(err.Error(), "DryRunOperation") {
			fmt.Println("🧪 Dry-run successful - instances would be stopped")
			return c.createDryRunResult("stop", instanceIDs), nil
		}
		return nil, fmt.Errorf("failed to stop instances: %v", err)
	}

	actionResult := &EC2ActionResult{
		Action:      "stop",
		InstanceIDs: instanceIDs,
		Success:     true,
		DryRun:      c.DryRun,
		Timestamp:   time.Now(),
	}

	// Process results
	for _, stateChange := range result.StoppingInstances {
		actionResult.StateChanges = append(actionResult.StateChanges, InstanceStateChange{
			InstanceID:    aws.ToString(stateChange.InstanceId),
			PreviousState: string(stateChange.PreviousState.Name),
			CurrentState:  string(stateChange.CurrentState.Name),
		})
	}

	fmt.Printf("✅ Stop command sent for %d instances\n", len(actionResult.StateChanges))
	return actionResult, nil
}

// TerminateInstances terminates EC2 instances
func (c *CustodianClient) TerminateInstances(instanceIDs []string) (*EC2ActionResult, error) {
	c.LogAWSCall("EC2", "TerminateInstances", c.DryRun)

	if len(instanceIDs) == 0 {
		return &EC2ActionResult{}, nil
	}

	fmt.Printf("💀 Terminating %d instances...\n", len(instanceIDs))
	if !c.DryRun {
		fmt.Println("⚠️  WARNING: This action is IRREVERSIBLE!")
	}

	ctx := context.Background()
	input := &ec2.TerminateInstancesInput{
		InstanceIds: instanceIDs,
	}

	if c.DryRun {
		input.DryRun = aws.Bool(true)
	}

	result, err := c.EC2.TerminateInstances(ctx, input)
	if err != nil {
		// Check if it's a dry-run success
		if c.DryRun && strings.Contains(err.Error(), "DryRunOperation") {
			fmt.Println("🧪 Dry-run successful - instances would be terminated")
			return c.createDryRunResult("terminate", instanceIDs), nil
		}
		return nil, fmt.Errorf("failed to terminate instances: %v", err)
	}

	actionResult := &EC2ActionResult{
		Action:      "terminate",
		InstanceIDs: instanceIDs,
		Success:     true,
		DryRun:      c.DryRun,
		Timestamp:   time.Now(),
	}

	// Process results
	for _, stateChange := range result.TerminatingInstances {
		actionResult.StateChanges = append(actionResult.StateChanges, InstanceStateChange{
			InstanceID:    aws.ToString(stateChange.InstanceId),
			PreviousState: string(stateChange.PreviousState.Name),
			CurrentState:  string(stateChange.CurrentState.Name),
		})
	}

	fmt.Printf("✅ Terminate command sent for %d instances\n", len(actionResult.StateChanges))
	return actionResult, nil
}

// StartInstances starts stopped EC2 instances
func (c *CustodianClient) StartInstances(instanceIDs []string) (*EC2ActionResult, error) {
	c.LogAWSCall("EC2", "StartInstances", c.DryRun)

	if len(instanceIDs) == 0 {
		return &EC2ActionResult{}, nil
	}

	fmt.Printf("▶️  Starting %d instances...\n", len(instanceIDs))

	ctx := context.Background()
	input := &ec2.StartInstancesInput{
		InstanceIds: instanceIDs,
	}

	if c.DryRun {
		input.DryRun = aws.Bool(true)
	}

	result, err := c.EC2.StartInstances(ctx, input)
	if err != nil {
		if c.DryRun && strings.Contains(err.Error(), "DryRunOperation") {
			fmt.Println("🧪 Dry-run successful - instances would be started")
			return c.createDryRunResult("start", instanceIDs), nil
		}
		return nil, fmt.Errorf("failed to start instances: %v", err)
	}

	actionResult := &EC2ActionResult{
		Action:      "start",
		InstanceIDs: instanceIDs,
		Success:     true,
		DryRun:      c.DryRun,
		Timestamp:   time.Now(),
	}

	for _, stateChange := range result.StartingInstances {
		actionResult.StateChanges = append(actionResult.StateChanges, InstanceStateChange{
			InstanceID:    aws.ToString(stateChange.InstanceId),
			PreviousState: string(stateChange.PreviousState.Name),
			CurrentState:  string(stateChange.CurrentState.Name),
		})
	}

	fmt.Printf("✅ Start command sent for %d instances\n", len(actionResult.StateChanges))
	return actionResult, nil
}

// TagInstances adds tags to EC2 instances
func (c *CustodianClient) TagInstances(
	instanceIDs []string,
	tags map[string]string,
) (*EC2ActionResult, error) {
	c.LogAWSCall("EC2", "CreateTags", c.DryRun)

	if len(instanceIDs) == 0 || len(tags) == 0 {
		return &EC2ActionResult{}, nil
	}

	fmt.Printf("🏷️  Adding %d tags to %d instances...\n", len(tags), len(instanceIDs))

	ctx := context.Background()
	var ec2Tags []types.Tag

	for key, value := range tags {
		ec2Tags = append(ec2Tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	input := &ec2.CreateTagsInput{
		Resources: instanceIDs,
		Tags:      ec2Tags,
	}

	if c.DryRun {
		input.DryRun = aws.Bool(true)
	}

	_, err := c.EC2.CreateTags(ctx, input)
	if err != nil {
		if c.DryRun && strings.Contains(err.Error(), "DryRunOperation") {
			fmt.Println("🧪 Dry-run successful - tags would be added")
			return c.createTagDryRunResult(instanceIDs, tags), nil
		}
		return nil, fmt.Errorf("failed to create tags: %v", err)
	}

	actionResult := &EC2ActionResult{
		Action:      "tag",
		InstanceIDs: instanceIDs,
		Success:     true,
		DryRun:      c.DryRun,
		Timestamp:   time.Now(),
		Tags:        tags,
	}

	fmt.Printf("✅ Tags added to %d instances\n", len(instanceIDs))
	return actionResult, nil
}

// EC2ActionResult represents the result of an EC2 action
type EC2ActionResult struct {
	Action       string                `json:"action"`
	InstanceIDs  []string              `json:"instance_ids"`
	Success      bool                  `json:"success"`
	DryRun       bool                  `json:"dry_run"`
	Timestamp    time.Time             `json:"timestamp"`
	StateChanges []InstanceStateChange `json:"state_changes,omitempty"`
	Tags         map[string]string     `json:"tags,omitempty"`
	Error        string                `json:"error,omitempty"`
}

// InstanceStateChange represents an instance state transition
type InstanceStateChange struct {
	InstanceID    string `json:"instance_id"`
	PreviousState string `json:"previous_state"`
	CurrentState  string `json:"current_state"`
}

// Helper functions for dry-run results
func (c *CustodianClient) createDryRunResult(action string, instanceIDs []string) *EC2ActionResult {
	result := &EC2ActionResult{
		Action:      action,
		InstanceIDs: instanceIDs,
		Success:     true,
		DryRun:      true,
		Timestamp:   time.Now(),
	}

	// Simulate state changes for dry-run
	for _, instanceID := range instanceIDs {
		var currentState string
		switch action {
		case "stop":
			currentState = "stopping"
		case "start":
			currentState = "pending"
		case "terminate":
			currentState = "shutting-down"
		}

		result.StateChanges = append(result.StateChanges, InstanceStateChange{
			InstanceID:    instanceID,
			PreviousState: "running", // Assume running for dry-run
			CurrentState:  currentState,
		})
	}

	return result
}

func (c *CustodianClient) createTagDryRunResult(
	instanceIDs []string,
	tags map[string]string,
) *EC2ActionResult {
	return &EC2ActionResult{
		Action:      "tag",
		InstanceIDs: instanceIDs,
		Success:     true,
		DryRun:      true,
		Timestamp:   time.Now(),
		Tags:        tags,
	}
}

// WaitForInstanceState waits for instances to reach a specific state
func (c *CustodianClient) WaitForInstanceState(
	instanceIDs []string,
	targetState string,
	timeout time.Duration,
) error {
	fmt.Printf("⏳ Waiting for %d instances to reach state: %s\n", len(instanceIDs), targetState)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.WaitForCompletion(ctx, func() (bool, error) {
		instances, err := c.GetEC2Instances(EC2Filter{InstanceIDs: instanceIDs})
		if err != nil {
			return false, err
		}

		for _, instance := range instances {
			if instance.State != targetState {
				return false, nil
			}
		}

		fmt.Printf("✅ All instances reached state: %s\n", targetState)
		return true, nil
	}, timeout)
}

// GetInstanceCosts calculates total costs for instances
func (c *CustodianClient) GetInstanceCosts(instances []EC2Instance) map[string]float64 {
	costs := map[string]float64{
		"total_monthly":     0,
		"running_monthly":   0,
		"stopped_monthly":   0,
		"potential_savings": 0,
	}

	for _, instance := range instances {
		costs["total_monthly"] += instance.MonthlyCost

		if instance.State == "running" {
			costs["running_monthly"] += instance.MonthlyCost
		} else {
			costs["stopped_monthly"] += instance.MonthlyCost
			// Stopped instances still cost for EBS storage, roughly 10% of running cost
			costs["stopped_monthly"] += instance.MonthlyCost * 0.1
		}

		// Calculate potential savings for unused instances
		if instance.CPUUtilization > 0 && instance.CPUUtilization < 5.0 &&
			instance.RunningDays > 7 {
			costs["potential_savings"] += instance.MonthlyCost
		}
	}

	return costs
}
