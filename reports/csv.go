package reports

import (
	"custodian-killer/aws"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// CSVReportGenerator creates CSV reports for spreadsheet analysis
type CSVReportGenerator struct {
	outputDir string
}

// NewCSVReportGenerator creates new CSV report maker
func NewCSVReportGenerator(outputDir string) *CSVReportGenerator {
	if outputDir == "" {
		outputDir = "./reports"
	}

	// Make sure output directory exists
	os.MkdirAll(outputDir, 0755)

	return &CSVReportGenerator{
		outputDir: outputDir,
	}
}

// GenerateEC2Report creates CSV report for EC2 instances
func (c *CSVReportGenerator) GenerateEC2Report(instances []aws.EC2Instance, filename string) error {
	fmt.Printf("📝 Generating EC2 CSV report: %s\n", filename)

	fullPath := filepath.Join(c.outputDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Instance ID",
		"Name",
		"Instance Type",
		"State",
		"Launch Time",
		"Running Days",
		"CPU Utilization %",
		"Monthly Cost $",
		"Public IP",
		"Private IP",
		"VPC ID",
		"Environment Tag",
		"Owner Tag",
		"Project Tag",
		"Compliance Issues",
		"Risk Level",
		"Recommendations",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data rows
	for _, instance := range instances {
		// Determine compliance issues
		var issues []string
		riskLevel := "Low"
		var recommendations []string

		// Check for missing tags
		requiredTags := []string{"Environment", "Owner", "Project"}
		for _, tag := range requiredTags {
			if _, exists := instance.Tags[tag]; !exists {
				issues = append(issues, fmt.Sprintf("Missing %s tag", tag))
				recommendations = append(recommendations, fmt.Sprintf("Add %s tag", tag))
			}
		}

		// Check utilization
		if instance.CPUUtilization < 5.0 && instance.RunningDays > 7 {
			issues = append(issues, "Low CPU utilization")
			riskLevel = "High"
			if instance.RunningDays > 30 {
				recommendations = append(recommendations, "Consider terminating")
			} else {
				recommendations = append(recommendations, "Consider stopping")
			}
		} else if instance.CPUUtilization < 20.0 {
			issues = append(issues, "Below average CPU utilization")
			riskLevel = "Medium"
			recommendations = append(recommendations, "Monitor usage and consider downsizing")
		}

		// Check for expensive instances with low utilization
		if instance.MonthlyCost > 100 && instance.CPUUtilization < 30 {
			issues = append(issues, "High cost with low utilization")
			riskLevel = "High"
			recommendations = append(recommendations, "Downsize instance type")
		}

		// Get tag values
		envTag := instance.Tags["Environment"]
		ownerTag := instance.Tags["Owner"]
		projectTag := instance.Tags["Project"]

		row := []string{
			instance.InstanceID,
			instance.Name,
			instance.InstanceType,
			instance.State,
			instance.LaunchTime.Format("2006-01-02 15:04:05"),
			strconv.Itoa(instance.RunningDays),
			fmt.Sprintf("%.1f", instance.CPUUtilization),
			fmt.Sprintf("%.2f", instance.MonthlyCost),
			instance.PublicIP,
			instance.PrivateIP,
			instance.VpcID,
			envTag,
			ownerTag,
			projectTag,
			joinStrings(issues, "; "),
			riskLevel,
			joinStrings(recommendations, "; "),
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	fmt.Printf("✅ EC2 CSV report saved: %s (%d instances)\n", fullPath, len(instances))
	return nil
}

// GenerateS3Report creates CSV report for S3 buckets
func (c *CSVReportGenerator) GenerateS3Report(buckets []aws.S3Bucket, filename string) error {
	fmt.Printf("📝 Generating S3 CSV report: %s\n", filename)

	fullPath := filepath.Join(c.outputDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Bucket Name",
		"Region",
		"Creation Date",
		"Size GB",
		"Object Count",
		"Monthly Cost $",
		"Public Read",
		"Public Write",
		"Encrypted",
		"Encryption Type",
		"Versioning",
		"Security Score",
		"Environment Tag",
		"Owner Tag",
		"Compliance Issues",
		"Risk Level",
		"Recommendations",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data rows
	for _, bucket := range buckets {
		// Determine compliance issues
		var issues []string
		riskLevel := "Low"
		var recommendations []string

		// Public access issues
		if bucket.PublicReadACL || bucket.PublicReadPolicy {
			issues = append(issues, "Public read access")
			riskLevel = "Critical"
			recommendations = append(recommendations, "Block public read access")
		}

		if bucket.PublicWriteACL || bucket.PublicWritePolicy {
			issues = append(issues, "Public write access")
			riskLevel = "Critical"
			recommendations = append(recommendations, "Block public write access")
		}

		// Encryption issues
		if !bucket.Encryption.Enabled {
			issues = append(issues, "No encryption")
			if riskLevel == "Low" {
				riskLevel = "High"
			}
			recommendations = append(recommendations, "Enable encryption")
		}

		// Versioning issues
		if bucket.Versioning == "Disabled" {
			issues = append(issues, "Versioning disabled")
			if riskLevel == "Low" {
				riskLevel = "Medium"
			}
			recommendations = append(recommendations, "Enable versioning")
		}

		// Missing tags
		requiredTags := []string{"Environment", "Owner"}
		for _, tag := range requiredTags {
			if _, exists := bucket.Tags[tag]; !exists {
				issues = append(issues, fmt.Sprintf("Missing %s tag", tag))
				recommendations = append(recommendations, fmt.Sprintf("Add %s tag", tag))
			}
		}

		// Get tag values
		envTag := bucket.Tags["Environment"]
		ownerTag := bucket.Tags["Owner"]

		// Calculate size in GB
		sizeGB := float64(bucket.SizeBytes) / (1024 * 1024 * 1024)
		_ = sizeGB // Use the variable to avoid compiler error

		row := []string{
			bucket.Name,
			bucket.Region,
			bucket.CreationDate.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.2f", sizeGB),
			strconv.FormatInt(bucket.ObjectCount, 10),
			fmt.Sprintf("%.2f", bucket.MonthlyCostEstimate),
			boolToYesNo(bucket.PublicReadACL || bucket.PublicReadPolicy),
			boolToYesNo(bucket.PublicWriteACL || bucket.PublicWritePolicy),
			boolToYesNo(bucket.Encryption.Enabled),
			bucket.Encryption.Algorithm,
			bucket.Versioning,
			strconv.Itoa(bucket.SecurityScore),
			envTag,
			ownerTag,
			joinStrings(issues, "; "),
			riskLevel,
			joinStrings(recommendations, "; "),
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	fmt.Printf("✅ S3 CSV report saved: %s (%d buckets)\n", fullPath, len(buckets))
	return nil
}

// GeneratePolicyExecutionReport creates CSV report for policy execution results
func (c *CSVReportGenerator) GeneratePolicyExecutionReport(
	results []ExecutionResult,
	filename string,
) error {
	fmt.Printf("📝 Generating policy execution CSV report: %s\n", filename)

	fullPath := filepath.Join(c.outputDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Policy Name",
		"Resource Type",
		"Execution Time",
		"Start Time",
		"End Time",
		"Duration (seconds)",
		"Dry Run",
		"Success",
		"Resources Found",
		"Resources Matched",
		"Actions Executed",
		"Successful Actions",
		"Failed Actions",
		"Resources Modified",
		"Monthly Savings $",
		"Security Improvements",
		"Errors",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data rows
	for _, result := range results {
		row := []string{
			result.PolicyName,
			result.ResourceType,
			result.StartTime.Format("2006-01-02 15:04:05"),
			result.StartTime.Format("2006-01-02 15:04:05"),
			result.EndTime.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.0f", result.Duration.Seconds()),
			boolToYesNo(result.DryRun),
			boolToYesNo(result.Success),
			strconv.Itoa(result.ResourcesFound),
			strconv.Itoa(result.ResourcesMatched),
			strconv.Itoa(result.Summary.TotalActions),
			strconv.Itoa(result.Summary.SuccessfulActions),
			strconv.Itoa(result.Summary.FailedActions),
			strconv.Itoa(result.Summary.ResourcesModified),
			fmt.Sprintf("%.2f", result.Summary.EstimatedMonthlySavings),
			strconv.Itoa(result.Summary.SecurityImprovements),
			joinStrings(result.Errors, "; "),
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	fmt.Printf("✅ Policy execution CSV report saved: %s (%d results)\n", fullPath, len(results))
	return nil
}

// GenerateActionDetailsReport creates detailed CSV report for individual actions
func (c *CSVReportGenerator) GenerateActionDetailsReport(
	results []ExecutionResult,
	filename string,
) error {
	fmt.Printf("📝 Generating action details CSV report: %s\n", filename)

	fullPath := filepath.Join(c.outputDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Policy Name",
		"Resource Type",
		"Resource ID",
		"Action Type",
		"Success",
		"Dry Run",
		"Message",
		"Timestamp",
		"Execution Time (ms)",
		"Details",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data rows for each action
	for _, result := range results {
		for _, actionResult := range result.ActionResults {
			// Convert details to string
			detailsStr := ""
			if actionResult.Details != nil {
				var detailParts []string
				for key, value := range actionResult.Details {
					detailParts = append(detailParts, fmt.Sprintf("%s=%v", key, value))
				}
				detailsStr = joinStrings(detailParts, ", ")
			}

			row := []string{
				result.PolicyName,
				actionResult.ResourceType,
				actionResult.ResourceID,
				actionResult.Action,
				boolToYesNo(actionResult.Success),
				boolToYesNo(actionResult.DryRun),
				actionResult.Message,
				actionResult.Timestamp.Format("2006-01-02 15:04:05"),
				fmt.Sprintf("%.0f", actionResult.ExecutionTime.Seconds()*1000),
				detailsStr,
			}

			if err := writer.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %v", err)
			}
		}
	}

	fmt.Printf("✅ Action details CSV report saved: %s\n", fullPath)
	return nil
}

// GenerateCostAnalysisReport creates CSV report focused on cost analysis
func (c *CSVReportGenerator) GenerateCostAnalysisReport(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
	filename string,
) error {
	fmt.Printf("📝 Generating cost analysis CSV report: %s\n", filename)

	fullPath := filepath.Join(c.outputDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Resource Type",
		"Resource ID",
		"Resource Name",
		"Current Monthly Cost $",
		"Utilization %",
		"Running Days",
		"Potential Monthly Savings $",
		"Annual Savings $",
		"Recommendation",
		"Priority",
		"Risk Level",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Process EC2 instances
	for _, instance := range ec2Instances {
		potentialSavings := 0.0
		recommendation := "Monitor usage"
		priority := "Low"
		riskLevel := "Low"

		// Calculate potential savings
		if instance.CPUUtilization < 5.0 && instance.RunningDays > 7 {
			potentialSavings = instance.MonthlyCost
			recommendation = "Stop or terminate instance"
			priority = "High"
			riskLevel = "High"
		} else if instance.CPUUtilization < 20.0 && instance.MonthlyCost > 50 {
			potentialSavings = instance.MonthlyCost * 0.5 // Assume downsizing saves 50%
			recommendation = "Consider downsizing"
			priority = "Medium"
			riskLevel = "Medium"
		}

		row := []string{
			"EC2",
			instance.InstanceID,
			instance.Name,
			fmt.Sprintf("%.2f", instance.MonthlyCost),
			fmt.Sprintf("%.1f", instance.CPUUtilization),
			strconv.Itoa(instance.RunningDays),
			fmt.Sprintf("%.2f", potentialSavings),
			fmt.Sprintf("%.2f", potentialSavings*12),
			recommendation,
			priority,
			riskLevel,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	// Process S3 buckets
	for _, bucket := range s3Buckets {
		potentialSavings := 0.0
		recommendation := "Monitor usage"
		priority := "Low"
		riskLevel := "Low"

		// Calculate potential savings from storage class optimization
		if standardSize, exists := bucket.StorageClass["STANDARD"]; exists {
			standardSizeGB := float64(standardSize) / (1024 * 1024 * 1024)
			if standardSizeGB > 100 {
				// Assume 30% could move to IA
				potentialSavings = standardSizeGB * 0.3 * (0.023 - 0.0125)
				recommendation = "Implement lifecycle policies"
				priority = "Medium"
				riskLevel = "Low"
			}
		}

		// Security cost considerations
		if bucket.PublicReadACL || bucket.PublicWriteACL {
			riskLevel = "Critical"
			priority = "Critical"
			recommendation = "Secure bucket immediately"
		}

		sizeGB := float64(bucket.SizeBytes) / (1024 * 1024 * 1024)

		row := []string{
			"S3",
			bucket.Name,
			bucket.Name,
			fmt.Sprintf("%.2f", bucket.MonthlyCostEstimate),
			"N/A",
			fmt.Sprintf("%.0f", time.Since(bucket.CreationDate).Hours()/24),
			fmt.Sprintf("%.2f", potentialSavings),
			fmt.Sprintf("%.2f", potentialSavings*12),
			fmt.Sprintf("%.2f", sizeGB),
			recommendation,
			priority,
			riskLevel,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	fmt.Printf("✅ Cost analysis CSV report saved: %s\n", fullPath)
	return nil
}

// GenerateComplianceSummaryReport creates high-level compliance summary CSV
func (c *CSVReportGenerator) GenerateComplianceSummaryReport(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
	filename string,
) error {
	fmt.Printf("📝 Generating compliance summary CSV report: %s\n", filename)

	fullPath := filepath.Join(c.outputDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Compliance Category",
		"Total Resources",
		"Compliant Resources",
		"Non-Compliant Resources",
		"Compliance %",
		"Critical Issues",
		"High Risk Issues",
		"Medium Risk Issues",
		"Potential Monthly Savings $",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// EC2 Compliance Summary
	ec2Total := len(ec2Instances)
	ec2NonCompliant := 0
	ec2Critical := 0
	ec2High := 0
	ec2Medium := 0
	ec2Savings := 0.0

	for _, instance := range ec2Instances {
		hasIssues := false
		severity := "low"

		// Check for missing tags
		requiredTags := []string{"Environment", "Owner"}
		missingTags := 0
		for _, tag := range requiredTags {
			if _, exists := instance.Tags[tag]; !exists {
				missingTags++
				hasIssues = true
			}
		}

		if missingTags > 0 {
			severity = "medium"
		}

		// Check utilization
		if instance.CPUUtilization < 5.0 && instance.RunningDays > 7 {
			hasIssues = true
			severity = "critical"
			ec2Savings += instance.MonthlyCost
		} else if instance.CPUUtilization < 20.0 {
			hasIssues = true
			if severity == "low" {
				severity = "high"
			}
		}

		if hasIssues {
			ec2NonCompliant++
			switch severity {
			case "critical":
				ec2Critical++
			case "high":
				ec2High++
			case "medium":
				ec2Medium++
			}
		}
	}

	ec2Compliant := ec2Total - ec2NonCompliant
	ec2CompliancePercent := 0.0
	if ec2Total > 0 {
		ec2CompliancePercent = float64(ec2Compliant) / float64(ec2Total) * 100
	}

	// Write EC2 row
	ec2Row := []string{
		"EC2 Instances",
		strconv.Itoa(ec2Total),
		strconv.Itoa(ec2Compliant),
		strconv.Itoa(ec2NonCompliant),
		fmt.Sprintf("%.1f", ec2CompliancePercent),
		strconv.Itoa(ec2Critical),
		strconv.Itoa(ec2High),
		strconv.Itoa(ec2Medium),
		fmt.Sprintf("%.2f", ec2Savings),
	}

	if err := writer.Write(ec2Row); err != nil {
		return fmt.Errorf("failed to write CSV row: %v", err)
	}

	// S3 Compliance Summary
	s3Total := len(s3Buckets)
	s3NonCompliant := 0
	s3Critical := 0
	s3High := 0
	s3Medium := 0
	s3Savings := 0.0

	for _, bucket := range s3Buckets {
		hasIssues := false
		severity := "low"

		// Public access issues
		if bucket.PublicReadACL || bucket.PublicWriteACL {
			hasIssues = true
			severity = "critical"
		}

		// Encryption issues
		if !bucket.Encryption.Enabled {
			hasIssues = true
			if severity == "low" {
				severity = "high"
			}
		}

		// Versioning issues
		if bucket.Versioning == "Disabled" {
			hasIssues = true
			if severity == "low" {
				severity = "medium"
			}
		}

		if hasIssues {
			s3NonCompliant++
			switch severity {
			case "critical":
				s3Critical++
			case "high":
				s3High++
			case "medium":
				s3Medium++
			}
		}
	}

	s3Compliant := s3Total - s3NonCompliant
	s3CompliancePercent := 0.0
	if s3Total > 0 {
		s3CompliancePercent = float64(s3Compliant) / float64(s3Total) * 100
	}

	// Write S3 row
	s3Row := []string{
		"S3 Buckets",
		strconv.Itoa(s3Total),
		strconv.Itoa(s3Compliant),
		strconv.Itoa(s3NonCompliant),
		fmt.Sprintf("%.1f", s3CompliancePercent),
		strconv.Itoa(s3Critical),
		strconv.Itoa(s3High),
		strconv.Itoa(s3Medium),
		fmt.Sprintf("%.2f", s3Savings),
	}

	if err := writer.Write(s3Row); err != nil {
		return fmt.Errorf("failed to write CSV row: %v", err)
	}

	// Overall summary row
	totalResources := ec2Total + s3Total
	totalCompliant := ec2Compliant + s3Compliant
	totalNonCompliant := ec2NonCompliant + s3NonCompliant
	overallCompliancePercent := 0.0
	if totalResources > 0 {
		overallCompliancePercent = float64(totalCompliant) / float64(totalResources) * 100
	}

	overallRow := []string{
		"OVERALL",
		strconv.Itoa(totalResources),
		strconv.Itoa(totalCompliant),
		strconv.Itoa(totalNonCompliant),
		fmt.Sprintf("%.1f", overallCompliancePercent),
		strconv.Itoa(ec2Critical + s3Critical),
		strconv.Itoa(ec2High + s3High),
		strconv.Itoa(ec2Medium + s3Medium),
		fmt.Sprintf("%.2f", ec2Savings+s3Savings),
	}

	if err := writer.Write(overallRow); err != nil {
		return fmt.Errorf("failed to write CSV row: %v", err)
	}

	fmt.Printf("✅ Compliance summary CSV report saved: %s\n", fullPath)
	return nil
}

// Helper functions
func joinStrings(strs []string, separator string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += separator + strs[i]
	}
	return result
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
