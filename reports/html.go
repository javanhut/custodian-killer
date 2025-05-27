package reports

import (
	"custodian-killer/aws"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HTMLReportGenerator creates fancy HTML reports
type HTMLReportGenerator struct {
	outputDir string
}

// NewHTMLReportGenerator creates new HTML report maker
func NewHTMLReportGenerator(outputDir string) *HTMLReportGenerator {
	if outputDir == "" {
		outputDir = "./reports"
	}

	// Make sure output directory exists
	os.MkdirAll(outputDir, 0755)

	return &HTMLReportGenerator{
		outputDir: outputDir,
	}
}

// ComplianceReport represents compliance report data
type ComplianceReport struct {
	GeneratedAt     time.Time                `json:"generated_at"`
	Title           string                   `json:"title"`
	Summary         ComplianceSummary        `json:"summary"`
	EC2Findings     []EC2ComplianceFinding   `json:"ec2_findings"`
	S3Findings      []S3ComplianceFinding    `json:"s3_findings"`
	PolicyResults   []PolicyComplianceResult `json:"policy_results"`
	CostImpact      CostImpactSummary        `json:"cost_impact"`
	Recommendations []string                 `json:"recommendations"`
	SecurityScore   int                      `json:"security_score"`
}

// ComplianceSummary provides high-level compliance stats
type ComplianceSummary struct {
	TotalResources        int     `json:"total_resources"`
	CompliantResources    int     `json:"compliant_resources"`
	NonCompliantResources int     `json:"non_compliant_resources"`
	CompliancePercentage  float64 `json:"compliance_percentage"`
	CriticalIssues        int     `json:"critical_issues"`
	HighRiskResources     int     `json:"high_risk_resources"`
	EstimatedSavings      float64 `json:"estimated_savings"`
}

// EC2ComplianceFinding represents EC2 compliance issues
type EC2ComplianceFinding struct {
	InstanceID     string            `json:"instance_id"`
	Name           string            `json:"name"`
	InstanceType   string            `json:"instance_type"`
	State          string            `json:"state"`
	Issues         []string          `json:"issues"`
	Severity       string            `json:"severity"`
	Tags           map[string]string `json:"tags"`
	EstimatedCost  float64           `json:"estimated_cost"`
	RunningDays    int               `json:"running_days"`
	CPUUtilization float64           `json:"cpu_utilization"`
}

// S3ComplianceFinding represents S3 compliance issues
type S3ComplianceFinding struct {
	BucketName    string   `json:"bucket_name"`
	Issues        []string `json:"issues"`
	Severity      string   `json:"severity"`
	PublicAccess  bool     `json:"public_access"`
	Encrypted     bool     `json:"encrypted"`
	Versioning    string   `json:"versioning"`
	SecurityScore int      `json:"security_score"`
	EstimatedCost float64  `json:"estimated_cost"`
	SizeGB        float64  `json:"size_gb"`
}

// PolicyComplianceResult represents policy execution results
type PolicyComplianceResult struct {
	PolicyName       string    `json:"policy_name"`
	ResourceType     string    `json:"resource_type"`
	ResourcesFound   int       `json:"resources_found"`
	ResourcesMatched int       `json:"resources_matched"`
	ActionsExecuted  int       `json:"actions_executed"`
	IssuesFixed      int       `json:"issues_fixed"`
	EstimatedSavings float64   `json:"estimated_savings"`
	LastRun          time.Time `json:"last_run"`
}

// CostImpactSummary represents cost impact analysis
type CostImpactSummary struct {
	CurrentMonthlyCost float64            `json:"current_monthly_cost"`
	PotentialSavings   float64            `json:"potential_savings"`
	SavingsPercentage  float64            `json:"savings_percentage"`
	AnnualSavings      float64            `json:"annual_savings"`
	CostByResourceType map[string]float64 `json:"cost_by_resource_type"`
	TopCostlyResources []CostlyResource   `json:"top_costly_resources"`
}

// CostlyResource represents expensive resources
type CostlyResource struct {
	ResourceID     string  `json:"resource_id"`
	ResourceType   string  `json:"resource_type"`
	MonthlyCost    float64 `json:"monthly_cost"`
	Utilization    string  `json:"utilization"`
	Recommendation string  `json:"recommendation"`
}

// GenerateComplianceReport creates a comprehensive compliance report
func (h *HTMLReportGenerator) GenerateComplianceReport(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
) (*ComplianceReport, error) {
	fmt.Println("📊 Generating compliance report...")

	report := &ComplianceReport{
		GeneratedAt: time.Now(),
		Title:       "Custodian Killer Compliance Report",
		Summary:     ComplianceSummary{},
		CostImpact: CostImpactSummary{
			CostByResourceType: make(map[string]float64),
		},
	}

	// Analyze EC2 instances
	report.EC2Findings = h.analyzeEC2Compliance(ec2Instances)

	// Analyze S3 buckets
	report.S3Findings = h.analyzeS3Compliance(s3Buckets)

	// Calculate summary statistics
	report.Summary = h.calculateComplianceSummary(report)

	// Calculate cost impact
	report.CostImpact = h.calculateCostImpact(ec2Instances, s3Buckets)

	// Generate recommendations
	report.Recommendations = h.generateRecommendations(report)

	// Calculate overall security score
	report.SecurityScore = h.calculateSecurityScore(report)

	fmt.Printf("✅ Compliance report generated with %d findings\n",
		len(report.EC2Findings)+len(report.S3Findings))

	return report, nil
}

// analyzeEC2Compliance analyzes EC2 instances for compliance issues
func (h *HTMLReportGenerator) analyzeEC2Compliance(
	instances []aws.EC2Instance,
) []EC2ComplianceFinding {
	var findings []EC2ComplianceFinding

	for _, instance := range instances {
		finding := EC2ComplianceFinding{
			InstanceID:     instance.InstanceID,
			Name:           instance.Name,
			InstanceType:   instance.InstanceType,
			State:          instance.State,
			Tags:           instance.Tags,
			EstimatedCost:  instance.MonthlyCost,
			RunningDays:    instance.RunningDays,
			CPUUtilization: instance.CPUUtilization,
			Issues:         []string{},
		}

		// Check for compliance issues
		severity := "low"

		// Missing required tags
		requiredTags := []string{"Environment", "Owner", "Project"}
		for _, tag := range requiredTags {
			if _, exists := instance.Tags[tag]; !exists {
				finding.Issues = append(
					finding.Issues,
					fmt.Sprintf("Missing required tag: %s", tag),
				)
				severity = "medium"
			}
		}

		// Unused instances (low CPU + long running)
		if instance.CPUUtilization < 5.0 && instance.RunningDays > 7 {
			finding.Issues = append(
				finding.Issues,
				fmt.Sprintf(
					"Low CPU utilization (%.1f%%) for %d days",
					instance.CPUUtilization,
					instance.RunningDays,
				),
			)
			severity = "high"
		}

		// Long-running instances without proper tagging
		if instance.RunningDays > 30 && len(instance.Tags) < 3 {
			finding.Issues = append(finding.Issues, "Long-running instance with insufficient tags")
			severity = "medium"
		}

		// Expensive instance types
		if instance.MonthlyCost > 100 {
			finding.Issues = append(finding.Issues,
				fmt.Sprintf("High monthly cost: $%.2f", instance.MonthlyCost))
			if instance.CPUUtilization < 20 {
				severity = "high"
			}
		}

		finding.Severity = severity

		// Only include instances with issues
		if len(finding.Issues) > 0 {
			findings = append(findings, finding)
		}
	}

	return findings
}

// analyzeS3Compliance analyzes S3 buckets for compliance issues
func (h *HTMLReportGenerator) analyzeS3Compliance(buckets []aws.S3Bucket) []S3ComplianceFinding {
	var findings []S3ComplianceFinding

	for _, bucket := range buckets {
		finding := S3ComplianceFinding{
			BucketName:    bucket.Name,
			PublicAccess:  bucket.PublicReadACL || bucket.PublicWriteACL,
			Encrypted:     bucket.Encryption.Enabled,
			Versioning:    bucket.Versioning,
			SecurityScore: bucket.SecurityScore,
			EstimatedCost: bucket.MonthlyCostEstimate,
			SizeGB:        float64(bucket.SizeBytes) / (1024 * 1024 * 1024),
			Issues:        []string{},
		}

		severity := "low"

		// Public access issues
		if bucket.PublicReadACL {
			finding.Issues = append(finding.Issues, "Bucket allows public read access")
			severity = "critical"
		}
		if bucket.PublicWriteACL {
			finding.Issues = append(finding.Issues, "Bucket allows public write access")
			severity = "critical"
		}

		// Encryption issues
		if !bucket.Encryption.Enabled {
			finding.Issues = append(finding.Issues, "Bucket encryption is disabled")
			if severity != "critical" {
				severity = "high"
			}
		}

		// Versioning issues
		if bucket.Versioning == "Disabled" {
			finding.Issues = append(finding.Issues, "Bucket versioning is disabled")
			if severity == "low" {
				severity = "medium"
			}
		}

		// Public access block configuration
		if !bucket.BlockPublicACLs || !bucket.BlockPublicPolicy {
			finding.Issues = append(finding.Issues, "Public access block not fully configured")
			if severity == "low" {
				severity = "medium"
			}
		}

		// Missing tags
		requiredTags := []string{"Environment", "Owner"}
		for _, tag := range requiredTags {
			if _, exists := bucket.Tags[tag]; !exists {
				finding.Issues = append(
					finding.Issues,
					fmt.Sprintf("Missing required tag: %s", tag),
				)
			}
		}

		finding.Severity = severity

		// Only include buckets with issues
		if len(finding.Issues) > 0 {
			findings = append(findings, finding)
		}
	}

	return findings
}

// calculateComplianceSummary calculates overall compliance statistics
func (h *HTMLReportGenerator) calculateComplianceSummary(
	report *ComplianceReport,
) ComplianceSummary {
	summary := ComplianceSummary{}

	// Count total resources and issues
	summary.TotalResources = len(report.EC2Findings) + len(report.S3Findings)

	for _, finding := range report.EC2Findings {
		if finding.Severity == "critical" || finding.Severity == "high" {
			summary.CriticalIssues++
		}
		if finding.Severity == "high" {
			summary.HighRiskResources++
		}
		summary.EstimatedSavings += finding.EstimatedCost
	}

	for _, finding := range report.S3Findings {
		if finding.Severity == "critical" || finding.Severity == "high" {
			summary.CriticalIssues++
		}
		if finding.Severity == "high" || finding.Severity == "critical" {
			summary.HighRiskResources++
		}
	}

	// Calculate compliance percentage (simplified)
	if summary.TotalResources > 0 {
		summary.NonCompliantResources = summary.TotalResources
		summary.CompliancePercentage = 0.0 // All resources in findings are non-compliant
	}

	return summary
}

// calculateCostImpact analyzes cost impact and savings opportunities
func (h *HTMLReportGenerator) calculateCostImpact(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
) CostImpactSummary {
	summary := CostImpactSummary{
		CostByResourceType: make(map[string]float64),
		TopCostlyResources: []CostlyResource{},
	}

	// Calculate EC2 costs
	ec2Cost := 0.0
	for _, instance := range ec2Instances {
		ec2Cost += instance.MonthlyCost

		// Identify potential savings from unused instances
		if instance.CPUUtilization < 5.0 && instance.RunningDays > 7 {
			summary.PotentialSavings += instance.MonthlyCost
		}

		// Add to top costly resources if expensive
		if instance.MonthlyCost > 50 {
			utilization := "Normal"
			recommendation := "Monitor usage"

			if instance.CPUUtilization < 5 {
				utilization = "Very Low"
				recommendation = "Consider stopping or terminating"
			} else if instance.CPUUtilization < 20 {
				utilization = "Low"
				recommendation = "Consider downsizing"
			}

			summary.TopCostlyResources = append(summary.TopCostlyResources, CostlyResource{
				ResourceID:     instance.InstanceID,
				ResourceType:   "EC2",
				MonthlyCost:    instance.MonthlyCost,
				Utilization:    utilization,
				Recommendation: recommendation,
			})
		}
	}
	summary.CostByResourceType["EC2"] = ec2Cost

	// Calculate S3 costs
	s3Cost := 0.0
	for _, bucket := range s3Buckets {
		s3Cost += bucket.MonthlyCostEstimate

		// Identify potential savings from storage optimization
		if standardSize, exists := bucket.StorageClass["STANDARD"]; exists {
			standardSizeGB := float64(standardSize) / (1024 * 1024 * 1024)
			if standardSizeGB > 100 { // Large buckets
				potentialSavings := standardSizeGB * 0.3 * (0.023 - 0.0125) // 30% to IA
				summary.PotentialSavings += potentialSavings
			}
		}
	}
	summary.CostByResourceType["S3"] = s3Cost

	summary.CurrentMonthlyCost = ec2Cost + s3Cost
	summary.AnnualSavings = summary.PotentialSavings * 12

	if summary.CurrentMonthlyCost > 0 {
		summary.SavingsPercentage = (summary.PotentialSavings / summary.CurrentMonthlyCost) * 100
	}

	return summary
}

// generateRecommendations creates actionable recommendations
func (h *HTMLReportGenerator) generateRecommendations(report *ComplianceReport) []string {
	var recommendations []string

	if report.Summary.CriticalIssues > 0 {
		recommendations = append(
			recommendations,
			fmt.Sprintf(
				"🚨 Address %d critical security issues immediately",
				report.Summary.CriticalIssues,
			),
		)
	}

	if report.CostImpact.PotentialSavings > 100 {
		recommendations = append(
			recommendations,
			fmt.Sprintf(
				"💰 Potential monthly savings of $%.2f by optimizing unused resources",
				report.CostImpact.PotentialSavings,
			),
		)
	}

	// Count untagged resources
	untaggedCount := 0
	for _, finding := range report.EC2Findings {
		for _, issue := range finding.Issues {
			if strings.Contains(issue, "Missing required tag") {
				untaggedCount++
				break
			}
		}
	}

	if untaggedCount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("🏷️  Implement tagging strategy for %d untagged resources", untaggedCount))
	}

	// S3 security recommendations
	publicBuckets := 0
	unencryptedBuckets := 0
	for _, finding := range report.S3Findings {
		if finding.PublicAccess {
			publicBuckets++
		}
		if !finding.Encrypted {
			unencryptedBuckets++
		}
	}

	if publicBuckets > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("🔒 Secure %d publicly accessible S3 buckets", publicBuckets))
	}

	if unencryptedBuckets > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("🔐 Enable encryption on %d S3 buckets", unencryptedBuckets))
	}

	// General recommendations
	recommendations = append(recommendations,
		"📊 Set up regular compliance monitoring with Custodian Killer policies")
	recommendations = append(recommendations,
		"🤖 Automate remediation for common compliance issues")

	return recommendations
}

// calculateSecurityScore calculates overall security score
func (h *HTMLReportGenerator) calculateSecurityScore(report *ComplianceReport) int {
	if len(report.S3Findings) == 0 {
		return 100 // No S3 buckets to evaluate
	}

	totalScore := 0
	for _, finding := range report.S3Findings {
		totalScore += finding.SecurityScore
	}

	return totalScore / len(report.S3Findings)
}

// SaveHTMLReport saves the compliance report as HTML
func (h *HTMLReportGenerator) SaveHTMLReport(report *ComplianceReport, filename string) error {
	fmt.Printf("💾 Saving HTML report: %s\n", filename)

	// Create full path
	fullPath := filepath.Join(h.outputDir, filename)

	// Create HTML template
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; border-bottom: 2px solid #007acc; padding-bottom: 20px; margin-bottom: 30px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .summary-card { background: #f8f9fa; padding: 20px; border-radius: 8px; text-align: center; border-left: 4px solid #007acc; }
        .summary-card h3 { margin: 0 0 10px 0; color: #333; }
        .summary-card .number { font-size: 2em; font-weight: bold; color: #007acc; }
        .critical { border-left-color: #dc3545; }
        .critical .number { color: #dc3545; }
        .warning { border-left-color: #ffc107; }
        .warning .number { color: #ffc107; }
        .success { border-left-color: #28a745; }
        .success .number { color: #28a745; }
        .section { margin-bottom: 30px; }
        .section h2 { color: #333; border-bottom: 1px solid #ddd; padding-bottom: 10px; }
        .findings { display: grid; gap: 15px; }
        .finding { background: #fff; border: 1px solid #ddd; border-radius: 8px; padding: 15px; }
        .finding.critical { border-left: 4px solid #dc3545; }
        .finding.high { border-left: 4px solid #fd7e14; }
        .finding.medium { border-left: 4px solid #ffc107; }
        .finding.low { border-left: 4px solid #6c757d; }
        .finding h4 { margin: 0 0 10px 0; color: #333; }
        .issues { list-style: none; padding: 0; }
        .issues li { background: #f8f9fa; margin: 5px 0; padding: 5px 10px; border-radius: 4px; }
        .recommendations { background: #e7f3ff; padding: 20px; border-radius: 8px; border-left: 4px solid #007acc; }
        .recommendations ul { margin: 0; padding-left: 20px; }
        .cost-table { width: 100%; border-collapse: collapse; margin-top: 15px; }
        .cost-table th, .cost-table td { border: 1px solid #ddd; padding: 10px; text-align: left; }
        .cost-table th { background: #f8f9fa; }
        .footer { text-align: center; margin-top: 40px; padding-top: 20px; border-top: 1px solid #ddd; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🦍 {{.Title}}</h1>
            <p>Generated on {{.GeneratedAt.Format "January 2, 2006 at 3:04 PM"}}</p>
        </div>

        <div class="summary">
            <div class="summary-card {{if gt .Summary.CriticalIssues 0}}critical{{else}}success{{end}}">
                <h3>Critical Issues</h3>
                <div class="number">{{.Summary.CriticalIssues}}</div>
            </div>
            <div class="summary-card {{if lt .Summary.CompliancePercentage 80}}warning{{else}}success{{end}}">
                <h3>Compliance Score</h3>
                <div class="number">{{printf "%.0f" .Summary.CompliancePercentage}}%</div>
            </div>
            <div class="summary-card success">
                <h3>Security Score</h3>
                <div class="number">{{.SecurityScore}}/100</div>
            </div>
            <div class="summary-card">
                <h3>Potential Savings</h3>
                <div class="number">${{printf "%.0f" .CostImpact.PotentialSavings}}</div>
                <small>per month</small>
            </div>
        </div>

        {{if .EC2Findings}}
        <div class="section">
            <h2>🖥️ EC2 Compliance Issues ({{len .EC2Findings}})</h2>
            <div class="findings">
                {{range .EC2Findings}}
                <div class="finding {{.Severity}}">
                    <h4>{{.Name}} ({{.InstanceID}})</h4>
                    <p><strong>Type:</strong> {{.InstanceType}} | <strong>State:</strong> {{.State}} | <strong>Cost:</strong> ${{printf "%.2f" .EstimatedCost}}/month</p>
                    {{if .Issues}}
                    <ul class="issues">
                        {{range .Issues}}<li>{{.}}</li>{{end}}
                    </ul>
                    {{end}}
                </div>
                {{end}}
            </div>
        </div>
        {{end}}

        {{if .S3Findings}}
        <div class="section">
            <h2>🪣 S3 Compliance Issues ({{len .S3Findings}})</h2>
            <div class="findings">
                {{range .S3Findings}}
                <div class="finding {{.Severity}}">
                    <h4>{{.BucketName}}</h4>
                    <p><strong>Security Score:</strong> {{.SecurityScore}}/100 | <strong>Size:</strong> {{printf "%.1f" .SizeGB}} GB | <strong>Cost:</strong> ${{printf "%.2f" .EstimatedCost}}/month</p>
                    {{if .Issues}}
                    <ul class="issues">
                        {{range .Issues}}<li>{{.}}</li>{{end}}
                    </ul>
                    {{end}}
                </div>
                {{end}}
            </div>
        </div>
        {{end}}

        <div class="section">
            <h2>💰 Cost Impact Analysis</h2>
            <table class="cost-table">
                <tr>
                    <th>Metric</th>
                    <th>Value</th>
                </tr>
                <tr>
                    <td>Current Monthly Cost</td>
                    <td>${{printf "%.2f" .CostImpact.CurrentMonthlyCost}}</td>
                </tr>
                <tr>
                    <td>Potential Monthly Savings</td>
                    <td>${{printf "%.2f" .CostImpact.PotentialSavings}}</td>
                </tr>
                <tr>
                    <td>Annual Savings</td>
                    <td>${{printf "%.2f" .CostImpact.AnnualSavings}}</td>
                </tr>
                <tr>
                    <td>Savings Percentage</td>
                    <td>{{printf "%.1f" .CostImpact.SavingsPercentage}}%</td>
                </tr>
            </table>
        </div>

        {{if .Recommendations}}
        <div class="section">
            <h2>💡 Recommendations</h2>
            <div class="recommendations">
                <ul>
                    {{range .Recommendations}}<li>{{.}}</li>{{end}}
                </ul>
            </div>
        </div>
        {{end}}

        <div class="footer">
            <p>🦍 Generated by Custodian Killer - Making AWS compliance fun again!</p>
        </div>
    </div>
</body>
</html>`

	// Parse and execute template
	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Create output file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Execute template
	if err := t.Execute(file, report); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	fmt.Printf("✅ HTML report saved: %s\n", fullPath)
	return nil
}
