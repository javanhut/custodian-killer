package reports

import (
	"custodian-killer/aws"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// JSONReportGenerator creates JSON reports for APIs and automation
type JSONReportGenerator struct {
	outputDir string
}

// NewJSONReportGenerator creates new JSON report maker
func NewJSONReportGenerator(outputDir string) *JSONReportGenerator {
	if outputDir == "" {
		outputDir = "./reports"
	}

	// Make sure output directory exists
	os.MkdirAll(outputDir, 0755)

	return &JSONReportGenerator{
		outputDir: outputDir,
	}
}

// PolicyExecutionReport represents policy execution results in JSON
type PolicyExecutionReport struct {
	GeneratedAt     time.Time               `json:"generated_at"`
	ReportType      string                  `json:"report_type"`
	Summary         ExecutionSummaryJSON    `json:"summary"`
	PolicyResults   []PolicyExecutionResult `json:"policy_results"`
	ResourceChanges []ResourceChange        `json:"resource_changes"`
	CostImpact      DetailedCostImpact      `json:"cost_impact"`
	Errors          []ErrorDetail           `json:"errors"`
	Metadata        map[string]interface{}  `json:"metadata"`
}

// ExecutionSummaryJSON provides high-level execution statistics for JSON reports
type ExecutionSummaryJSON struct {
	TotalPolicies         int       `json:"total_policies"`
	SuccessfulPolicies    int       `json:"successful_policies"`
	FailedPolicies        int       `json:"failed_policies"`
	TotalResourcesScanned int       `json:"total_resources_scanned"`
	TotalResourcesChanged int       `json:"total_resources_changed"`
	TotalActionsExecuted  int       `json:"total_actions_executed"`
	ExecutionDuration     string    `json:"execution_duration"`
	StartTime             time.Time `json:"start_time"`
	EndTime               time.Time `json:"end_time"`
}

// PolicyExecutionResult represents individual policy execution result
type PolicyExecutionResult struct {
	PolicyName        string             `json:"policy_name"`
	ResourceType      string             `json:"resource_type"`
	Status            string             `json:"status"` // success, failed, partial
	ExecutionTime     string             `json:"execution_time"`
	ResourcesFound    int                `json:"resources_found"`
	ResourcesMatched  int                `json:"resources_matched"`
	ActionsExecuted   int                `json:"actions_executed"`
	ActionsSuccessful int                `json:"actions_successful"`
	ActionsFailed     int                `json:"actions_failed"`
	CostSavings       float64            `json:"cost_savings"`
	Actions           []ActionResultJSON `json:"actions"`
	Errors            []string           `json:"errors"`
	DryRun            bool               `json:"dry_run"`
}

// ActionResultJSON represents action results for JSON reports
type ActionResultJSON struct {
	Action        string                 `json:"action"`
	ResourceID    string                 `json:"resource_id"`
	ResourceType  string                 `json:"resource_type"`
	Success       bool                   `json:"success"`
	DryRun        bool                   `json:"dry_run"`
	Message       string                 `json:"message"`
	Details       map[string]interface{} `json:"details,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	ExecutionTime string                 `json:"execution_time"`
	CostImpact    float64                `json:"cost_impact"`
}

// ResourceChange represents a change made to a resource
type ResourceChange struct {
	ResourceID    string                 `json:"resource_id"`
	ResourceType  string                 `json:"resource_type"`
	ChangeType    string                 `json:"change_type"` // created, modified, deleted
	PolicyName    string                 `json:"policy_name"`
	Action        string                 `json:"action"`
	PreviousState map[string]interface{} `json:"previous_state"`
	NewState      map[string]interface{} `json:"new_state"`
	Timestamp     time.Time              `json:"timestamp"`
	CostImpact    float64                `json:"cost_impact"`
	DryRun        bool                   `json:"dry_run"`
}

// DetailedCostImpact provides detailed cost analysis
type DetailedCostImpact struct {
	BeforeExecution CostBreakdown        `json:"before_execution"`
	AfterExecution  CostBreakdown        `json:"after_execution"`
	Savings         CostBreakdown        `json:"savings"`
	ProjectedAnnual CostBreakdown        `json:"projected_annual"`
	CostByPolicy    map[string]float64   `json:"cost_by_policy"`
	CostByAction    map[string]float64   `json:"cost_by_action"`
	TopSavings      []SavingsOpportunity `json:"top_savings"`
}

// CostBreakdown breaks down costs by resource type
type CostBreakdown struct {
	Total      float64            `json:"total"`
	EC2        float64            `json:"ec2"`
	S3         float64            `json:"s3"`
	RDS        float64            `json:"rds"`
	Lambda     float64            `json:"lambda"`
	Other      float64            `json:"other"`
	ByResource map[string]float64 `json:"by_resource"`
}

// SavingsOpportunity represents a cost saving opportunity
type SavingsOpportunity struct {
	ResourceID      string  `json:"resource_id"`
	ResourceType    string  `json:"resource_type"`
	CurrentCost     float64 `json:"current_cost"`
	PotentialSaving float64 `json:"potential_saving"`
	Recommendation  string  `json:"recommendation"`
	Confidence      string  `json:"confidence"` // high, medium, low
}

// ErrorDetail provides detailed error information
type ErrorDetail struct {
	Timestamp  time.Time `json:"timestamp"`
	PolicyName string    `json:"policy_name,omitempty"`
	ResourceID string    `json:"resource_id,omitempty"`
	Action     string    `json:"action,omitempty"`
	ErrorType  string    `json:"error_type"`
	Message    string    `json:"message"`
	Severity   string    `json:"severity"` // critical, high, medium, low
}

// GenerateExecutionReport creates a detailed execution report in JSON
func (j *JSONReportGenerator) GenerateExecutionReport(
	results []ExecutionResult,
) (*PolicyExecutionReport, error) {
	fmt.Println("📊 Generating JSON execution report...")

	report := &PolicyExecutionReport{
		GeneratedAt:     time.Now(),
		ReportType:      "policy_execution",
		PolicyResults:   []PolicyExecutionResult{},
		ResourceChanges: []ResourceChange{},
		Errors:          []ErrorDetail{},
		Metadata: map[string]interface{}{
			"version":   "1.0.0",
			"generator": "custodian-killer",
		},
	}

	// Calculate summary
	summary := ExecutionSummaryJSON{}
	var startTimes []time.Time
	var endTimes []time.Time

	// Process each policy result
	for _, result := range results {
		policyResult := PolicyExecutionResult{
			PolicyName:        result.PolicyName,
			ResourceType:      result.ResourceType,
			ExecutionTime:     result.Duration.String(),
			ResourcesFound:    result.ResourcesFound,
			ResourcesMatched:  result.ResourcesMatched,
			ActionsExecuted:   result.Summary.TotalActions,
			ActionsSuccessful: result.Summary.SuccessfulActions,
			ActionsFailed:     result.Summary.FailedActions,
			CostSavings:       result.Summary.EstimatedMonthlySavings,
			DryRun:            result.DryRun,
			Actions:           []ActionResultJSON{},
			Errors:            result.Errors,
		}

		// Determine status
		if result.Success {
			policyResult.Status = "success"
			summary.SuccessfulPolicies++
		} else {
			policyResult.Status = "failed"
			summary.FailedPolicies++
		}

		// Convert action results
		for _, actionResult := range result.ActionResults {
			jsonActionResult := ActionResultJSON{
				Action:        actionResult.Action,
				ResourceID:    actionResult.ResourceID,
				ResourceType:  actionResult.ResourceType,
				Success:       actionResult.Success,
				DryRun:        actionResult.DryRun,
				Message:       actionResult.Message,
				Details:       actionResult.Details,
				Timestamp:     actionResult.Timestamp,
				ExecutionTime: actionResult.ExecutionTime.String(),
				CostImpact:    0.0, // Could calculate from action type
			}
			policyResult.Actions = append(policyResult.Actions, jsonActionResult)

			// Create resource change entries
			if actionResult.Success && !actionResult.DryRun {
				change := ResourceChange{
					ResourceID:   actionResult.ResourceID,
					ResourceType: actionResult.ResourceType,
					ChangeType:   "modified",
					PolicyName:   result.PolicyName,
					Action:       actionResult.Action,
					Timestamp:    actionResult.Timestamp,
					DryRun:       actionResult.DryRun,
					CostImpact:   jsonActionResult.CostImpact,
					PreviousState: map[string]interface{}{
						"status": "before_" + actionResult.Action,
					},
					NewState: map[string]interface{}{
						"status": "after_" + actionResult.Action,
					},
				}
				report.ResourceChanges = append(report.ResourceChanges, change)
			}
		}

		// Convert errors
		for _, errMsg := range result.Errors {
			errorDetail := ErrorDetail{
				Timestamp:  result.StartTime,
				PolicyName: result.PolicyName,
				ErrorType:  "execution_error",
				Message:    errMsg,
				Severity:   "medium",
			}
			report.Errors = append(report.Errors, errorDetail)
		}

		report.PolicyResults = append(report.PolicyResults, policyResult)

		// Track times for summary
		startTimes = append(startTimes, result.StartTime)
		endTimes = append(endTimes, result.EndTime)

		// Update summary counters
		summary.TotalResourcesScanned += result.ResourcesFound
		summary.TotalResourcesChanged += result.Summary.ResourcesModified
		summary.TotalActionsExecuted += result.Summary.TotalActions
	}

	// Finalize summary
	summary.TotalPolicies = len(results)
	if len(startTimes) > 0 {
		summary.StartTime = findEarliestTime(startTimes)
		summary.EndTime = findLatestTime(endTimes)
		summary.ExecutionDuration = summary.EndTime.Sub(summary.StartTime).String()
	}

	report.Summary = summary

	// Generate cost impact analysis
	report.CostImpact = j.generateCostImpact(results)

	fmt.Printf(
		"✅ JSON execution report generated with %d policy results\n",
		len(report.PolicyResults),
	)

	return report, nil
}

// generateCostImpact creates detailed cost impact analysis
func (j *JSONReportGenerator) generateCostImpact(results []ExecutionResult) DetailedCostImpact {
	impact := DetailedCostImpact{
		CostByPolicy: make(map[string]float64),
		CostByAction: make(map[string]float64),
		TopSavings:   []SavingsOpportunity{},
	}

	// Initialize cost breakdowns
	impact.BeforeExecution.ByResource = make(map[string]float64)
	impact.AfterExecution.ByResource = make(map[string]float64)
	impact.Savings.ByResource = make(map[string]float64)
	impact.ProjectedAnnual.ByResource = make(map[string]float64)

	for _, result := range results {
		savings := result.Summary.EstimatedMonthlySavings
		impact.CostByPolicy[result.PolicyName] = savings

		// Breakdown by resource type
		switch result.ResourceType {
		case "ec2":
			impact.Savings.EC2 += savings
		case "s3":
			impact.Savings.S3 += savings
		case "rds":
			impact.Savings.RDS += savings
		case "lambda":
			impact.Savings.Lambda += savings
		default:
			impact.Savings.Other += savings
		}

		// Track by action type
		for _, actionResult := range result.ActionResults {
			if actionResult.Success {
				actionSavings := 0.0

				// Estimate savings by action type
				switch actionResult.Action {
				case "stop":
					actionSavings = 50.0 // Rough estimate
				case "terminate":
					actionSavings = 75.0
				case "delete":
					actionSavings = 25.0
				}

				impact.CostByAction[actionResult.Action] += actionSavings

				// Add to top savings opportunities
				if actionSavings > 20 {
					opportunity := SavingsOpportunity{
						ResourceID:      actionResult.ResourceID,
						ResourceType:    actionResult.ResourceType,
						CurrentCost:     actionSavings * 1.5, // Estimate current cost
						PotentialSaving: actionSavings,
						Recommendation:  fmt.Sprintf("Execute %s action", actionResult.Action),
						Confidence:      "high",
					}
					impact.TopSavings = append(impact.TopSavings, opportunity)
				}
			}
		}
	}

	// Calculate totals
	impact.Savings.Total = impact.Savings.EC2 + impact.Savings.S3 + impact.Savings.RDS + impact.Savings.Lambda + impact.Savings.Other
	impact.ProjectedAnnual.Total = impact.Savings.Total * 12
	impact.ProjectedAnnual.EC2 = impact.Savings.EC2 * 12
	impact.ProjectedAnnual.S3 = impact.Savings.S3 * 12
	impact.ProjectedAnnual.RDS = impact.Savings.RDS * 12
	impact.ProjectedAnnual.Lambda = impact.Savings.Lambda * 12
	impact.ProjectedAnnual.Other = impact.Savings.Other * 12

	return impact
}

// SaveJSONReport saves the report as JSON
func (j *JSONReportGenerator) SaveJSONReport(report interface{}, filename string) error {
	fmt.Printf("💾 Saving JSON report: %s\n", filename)

	// Create full path
	fullPath := filepath.Join(j.outputDir, filename)

	// Create output file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Create JSON encoder with pretty printing
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	// Encode report to JSON
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	fmt.Printf("✅ JSON report saved: %s\n", fullPath)
	return nil
}

// GenerateComplianceReportJSON creates compliance report in JSON format
func (j *JSONReportGenerator) GenerateComplianceReportJSON(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
) (map[string]interface{}, error) {
	fmt.Println("📊 Generating JSON compliance report...")

	report := map[string]interface{}{
		"generated_at": time.Now(),
		"report_type":  "compliance",
		"version":      "1.0.0",
		"summary": map[string]interface{}{
			"total_resources":   len(ec2Instances) + len(s3Buckets),
			"ec2_instances":     len(ec2Instances),
			"s3_buckets":        len(s3Buckets),
			"compliance_issues": 0,
			"security_score":    0,
			"estimated_savings": 0.0,
		},
		"ec2_analysis":    j.analyzeEC2JSON(ec2Instances),
		"s3_analysis":     j.analyzeS3JSON(s3Buckets),
		"cost_analysis":   j.analyzeCostJSON(ec2Instances, s3Buckets),
		"recommendations": []string{},
	}

	// Update summary with analysis results
	ec2Analysis := report["ec2_analysis"].(map[string]interface{})
	s3Analysis := report["s3_analysis"].(map[string]interface{})
	costAnalysis := report["cost_analysis"].(map[string]interface{})

	summary := report["summary"].(map[string]interface{})
	summary["compliance_issues"] = ec2Analysis["issues_found"].(int) + s3Analysis["issues_found"].(int)
	summary["estimated_savings"] = costAnalysis["potential_monthly_savings"].(float64)

	fmt.Println("✅ JSON compliance report generated")

	return report, nil
}

// analyzeEC2JSON analyzes EC2 instances for JSON report
func (j *JSONReportGenerator) analyzeEC2JSON(instances []aws.EC2Instance) map[string]interface{} {
	analysis := map[string]interface{}{
		"total_instances": len(instances),
		"issues_found":    0,
		"by_state":        make(map[string]int),
		"by_type":         make(map[string]int),
		"cost_analysis": map[string]interface{}{
			"total_monthly_cost": 0.0,
			"unused_cost":        0.0,
		},
		"compliance_issues": []map[string]interface{}{},
	}

	totalCost := 0.0
	unusedCost := 0.0
	issuesFound := 0

	stateCount := make(map[string]int)
	typeCount := make(map[string]int)

	for _, instance := range instances {
		// Count by state and type
		stateCount[instance.State]++
		typeCount[instance.InstanceType]++

		totalCost += instance.MonthlyCost

		// Check for issues
		issues := []string{}

		// Missing tags
		requiredTags := []string{"Environment", "Owner"}
		for _, tag := range requiredTags {
			if _, exists := instance.Tags[tag]; !exists {
				issues = append(issues, fmt.Sprintf("Missing %s tag", tag))
			}
		}

		// Low utilization
		if instance.CPUUtilization < 5.0 && instance.RunningDays > 7 {
			issues = append(issues, "Low CPU utilization")
			unusedCost += instance.MonthlyCost
		}

		if len(issues) > 0 {
			issuesFound++
			issue := map[string]interface{}{
				"instance_id":     instance.InstanceID,
				"name":            instance.Name,
				"instance_type":   instance.InstanceType,
				"state":           instance.State,
				"issues":          issues,
				"monthly_cost":    instance.MonthlyCost,
				"cpu_utilization": instance.CPUUtilization,
				"running_days":    instance.RunningDays,
			}
			analysis["compliance_issues"] = append(
				analysis["compliance_issues"].([]map[string]interface{}),
				issue,
			)
		}
	}

	analysis["issues_found"] = issuesFound
	analysis["by_state"] = stateCount
	analysis["by_type"] = typeCount

	costAnalysis := analysis["cost_analysis"].(map[string]interface{})
	costAnalysis["total_monthly_cost"] = totalCost
	costAnalysis["unused_cost"] = unusedCost

	return analysis
}

// analyzeS3JSON analyzes S3 buckets for JSON report
func (j *JSONReportGenerator) analyzeS3JSON(buckets []aws.S3Bucket) map[string]interface{} {
	analysis := map[string]interface{}{
		"total_buckets": len(buckets),
		"issues_found":  0,
		"security_analysis": map[string]interface{}{
			"public_buckets":         0,
			"unencrypted_buckets":    0,
			"average_security_score": 0.0,
		},
		"compliance_issues": []map[string]interface{}{},
	}

	publicBuckets := 0
	unencryptedBuckets := 0
	totalSecurityScore := 0
	issuesFound := 0

	for _, bucket := range buckets {
		issues := []string{}
		severity := "low"

		// Public access
		if bucket.PublicReadACL || bucket.PublicWriteACL {
			issues = append(issues, "Public access enabled")
			publicBuckets++
			severity = "critical"
		}

		// Encryption
		if !bucket.Encryption.Enabled {
			issues = append(issues, "Encryption disabled")
			unencryptedBuckets++
			if severity != "critical" {
				severity = "high"
			}
		}

		// Versioning
		if bucket.Versioning == "Disabled" {
			issues = append(issues, "Versioning disabled")
			if severity == "low" {
				severity = "medium"
			}
		}

		totalSecurityScore += bucket.SecurityScore

		if len(issues) > 0 {
			issuesFound++
			issue := map[string]interface{}{
				"bucket_name":    bucket.Name,
				"issues":         issues,
				"severity":       severity,
				"security_score": bucket.SecurityScore,
				"public_access":  bucket.PublicReadACL || bucket.PublicWriteACL,
				"encrypted":      bucket.Encryption.Enabled,
				"versioning":     bucket.Versioning,
				"monthly_cost":   bucket.MonthlyCostEstimate,
			}
			analysis["compliance_issues"] = append(
				analysis["compliance_issues"].([]map[string]interface{}),
				issue,
			)
		}
	}

	analysis["issues_found"] = issuesFound

	securityAnalysis := analysis["security_analysis"].(map[string]interface{})
	securityAnalysis["public_buckets"] = publicBuckets
	securityAnalysis["unencrypted_buckets"] = unencryptedBuckets
	if len(buckets) > 0 {
		securityAnalysis["average_security_score"] = float64(
			totalSecurityScore,
		) / float64(
			len(buckets),
		)
	}

	return analysis
}

// analyzeCostJSON analyzes cost opportunities for JSON report
func (j *JSONReportGenerator) analyzeCostJSON(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
) map[string]interface{} {
	analysis := map[string]interface{}{
		"current_monthly_cost":      0.0,
		"potential_monthly_savings": 0.0,
		"annual_savings":            0.0,
		"savings_percentage":        0.0,
		"cost_by_service": map[string]float64{
			"ec2": 0.0,
			"s3":  0.0,
		},
		"savings_opportunities": []map[string]interface{}{},
	}

	totalCost := 0.0
	totalSavings := 0.0

	// EC2 analysis
	ec2Cost := 0.0
	for _, instance := range ec2Instances {
		ec2Cost += instance.MonthlyCost

		// Identify savings opportunities
		if instance.CPUUtilization < 5.0 && instance.RunningDays > 7 {
			totalSavings += instance.MonthlyCost

			opportunity := map[string]interface{}{
				"resource_id":      instance.InstanceID,
				"resource_type":    "ec2",
				"current_cost":     instance.MonthlyCost,
				"potential_saving": instance.MonthlyCost,
				"recommendation":   "Stop or terminate unused instance",
				"confidence":       "high",
			}
			analysis["savings_opportunities"] = append(
				analysis["savings_opportunities"].([]map[string]interface{}),
				opportunity,
			)
		}
	}

	// S3 analysis
	s3Cost := 0.0
	for _, bucket := range s3Buckets {
		s3Cost += bucket.MonthlyCostEstimate
	}

	totalCost = ec2Cost + s3Cost

	analysis["current_monthly_cost"] = totalCost
	analysis["potential_monthly_savings"] = totalSavings
	analysis["annual_savings"] = totalSavings * 12

	if totalCost > 0 {
		analysis["savings_percentage"] = (totalSavings / totalCost) * 100
	}

	costByService := analysis["cost_by_service"].(map[string]float64)
	costByService["ec2"] = ec2Cost
	costByService["s3"] = s3Cost

	return analysis
}

// Helper functions
func findEarliestTime(times []time.Time) time.Time {
	if len(times) == 0 {
		return time.Now()
	}

	earliest := times[0]
	for _, t := range times[1:] {
		if t.Before(earliest) {
			earliest = t
		}
	}
	return earliest
}

func findLatestTime(times []time.Time) time.Time {
	if len(times) == 0 {
		return time.Now()
	}

	latest := times[0]
	for _, t := range times[1:] {
		if t.After(latest) {
			latest = t
		}
	}
	return latest
}
