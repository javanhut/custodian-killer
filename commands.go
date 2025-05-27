package main

import (
	"custodian-killer/storage"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Policy command structure
var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage policies",
	Long:  "Create, edit, delete, and manage your AWS policies",
}

var createPolicyCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new policy interactively",
	Run: func(cmd *cobra.Command, args []string) {
		startPolicyCreation()
	},
}

var listPoliciesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all policies",
	Run: func(cmd *cobra.Command, args []string) {
		listPolicies()
	},
}

var editPolicyCmd = &cobra.Command{
	Use:   "edit [policy-name]",
	Short: "Edit an existing policy",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		editPolicy(args[0])
	},
}

var deletePolicyCmd = &cobra.Command{
	Use:   "delete [policy-name]",
	Short: "Delete a policy",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deletePolicy(args[0])
	},
}

var exportPolicyCmd = &cobra.Command{
	Use:   "export [policy-name] [output-file]",
	Short: "Export a policy to file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		exportPolicy(args[0], args[1])
	},
}

var importPolicyCmd = &cobra.Command{
	Use:   "import [input-file]",
	Short: "Import a policy from file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		importPolicy(args[0])
	},
}

// Scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan resources with policies (dry-run)",
	Long:  "Run your policies against AWS resources to see what would happen without making changes",
	Run: func(cmd *cobra.Command, args []string) {
		runScanCommand(cmd)
	},
}

// Execute command
var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute policies against AWS resources",
	Long:  "Actually run your policies and make changes to AWS resources",
	Run: func(cmd *cobra.Command, args []string) {
		runExecuteCommand(cmd)
	},
}

// Report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate compliance and cost reports",
	Long:  "Create detailed reports about your AWS compliance and potential cost savings",
}

var complianceReportCmd = &cobra.Command{
	Use:   "compliance",
	Short: "Generate compliance report",
	Run: func(cmd *cobra.Command, args []string) {
		generateComplianceReportCmd(cmd)
	},
}

var costReportCmd = &cobra.Command{
	Use:   "cost",
	Short: "Generate cost optimization report",
	Run: func(cmd *cobra.Command, args []string) {
		generateCostReportCmd(cmd)
	},
}

var inventoryReportCmd = &cobra.Command{
	Use:   "inventory",
	Short: "Generate resource inventory report",
	Run: func(cmd *cobra.Command, args []string) {
		generateInventoryReportCmd(cmd)
	},
}

// Config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure AWS credentials and settings",
	Long:  "Set up your AWS credentials, regions, and other preferences",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		showConfig()
	},
}

var configTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test AWS connection",
	Run: func(cmd *cobra.Command, args []string) {
		testAWSConnection()
	},
}

func init() {
	// Add subcommands to policy command
	policyCmd.AddCommand(createPolicyCmd)
	policyCmd.AddCommand(listPoliciesCmd)
	policyCmd.AddCommand(editPolicyCmd)
	policyCmd.AddCommand(deletePolicyCmd)
	policyCmd.AddCommand(exportPolicyCmd)
	policyCmd.AddCommand(importPolicyCmd)

	// Add subcommands to report command
	reportCmd.AddCommand(complianceReportCmd)
	reportCmd.AddCommand(costReportCmd)
	reportCmd.AddCommand(inventoryReportCmd)

	// Add subcommands to config command
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configTestCmd)

	// Add flags to scan command
	scanCmd.Flags().StringP("policy", "p", "", "Run specific policy only")
	scanCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	scanCmd.Flags().StringP("output", "o", "table", "Output format (table, json, csv)")
	scanCmd.Flags().StringP("region", "r", "", "AWS region to scan")

	// Add flags to execute command
	executeCmd.Flags().StringP("policy", "p", "", "Execute specific policy only")
	executeCmd.Flags().BoolP("force", "f", false, "Force execution without confirmation")
	executeCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode (same as scan)")
	executeCmd.Flags().StringP("region", "r", "", "AWS region to execute in")

	// Add flags to report commands
	complianceReportCmd.Flags().StringP("output", "o", "html", "Output format (html, json, csv)")
	complianceReportCmd.Flags().StringP("file", "f", "", "Output file path")
	complianceReportCmd.Flags().StringP("region", "r", "", "AWS region to analyze")

	costReportCmd.Flags().StringP("output", "o", "csv", "Output format (html, json, csv)")
	costReportCmd.Flags().StringP("file", "f", "", "Output file path")
	costReportCmd.Flags().StringP("region", "r", "", "AWS region to analyze")

	inventoryReportCmd.Flags().StringP("output", "o", "csv", "Output format (csv, json)")
	inventoryReportCmd.Flags().StringP("file", "f", "", "Output file path")
	inventoryReportCmd.Flags().StringP("region", "r", "", "AWS region to analyze")
}

// Command implementations
func editPolicy(policyName string) {
	fmt.Printf("✏️  Editing policy: %s\n", policyName)

	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized!")
		return
	}

	// Check if policy exists
	if !policyStorage.PolicyExists(policyName) {
		fmt.Printf("❌ Policy '%s' not found!\n", policyName)
		fmt.Println("💡 Use 'custodian-killer policy list' to see available policies")
		return
	}

	fmt.Println("📝 Policy editing via interactive mode - coming soon!")
	fmt.Println("💡 For now, you can:")
	fmt.Printf("   1. Export: custodian-killer policy export %s policy.json\n", policyName)
	fmt.Println("   2. Edit the JSON file manually")
	fmt.Printf("   3. Import: custodian-killer policy import policy.json\n")
}

func deletePolicy(policyName string) {
	fmt.Printf("🗑️  Deleting policy: %s\n", policyName)

	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized!")
		return
	}

	// Check if policy exists
	if !policyStorage.PolicyExists(policyName) {
		fmt.Printf("❌ Policy '%s' not found!\n", policyName)
		return
	}

	// Confirm deletion
	fmt.Printf("⚠️  Are you sure you want to delete policy '%s'? (y/N): ", policyName)
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("❌ Deletion cancelled")
		return
	}

	// Delete the policy
	if err := policyStorage.DeletePolicy(policyName); err != nil {
		fmt.Printf("❌ Failed to delete policy: %v\n", err)
		return
	}

	fmt.Printf("✅ Policy '%s' deleted successfully\n", policyName)
}

func exportPolicy(policyName, outputFile string) {
	fmt.Printf("📤 Exporting policy '%s' to: %s\n", policyName, outputFile)

	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized!")
		return
	}

	// Check if policy exists
	if !policyStorage.PolicyExists(policyName) {
		fmt.Printf("❌ Policy '%s' not found!\n", policyName)
		return
	}

	// Export the policy
	if fileStorage, ok := policyStorage.(*storage.FileStorage); ok {
		if err := fileStorage.ExportPolicy(policyName, outputFile); err != nil {
			fmt.Printf("❌ Failed to export policy: %v\n", err)
			return
		}
	} else {
		fmt.Println("❌ Export not supported with current storage type")
		return
	}

	fmt.Printf("✅ Policy exported successfully\n")
}

func importPolicy(inputFile string) {
	fmt.Printf("📥 Importing policy from: %s\n", inputFile)

	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized!")
		return
	}

	// Check if file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Printf("❌ File not found: %s\n", inputFile)
		return
	}

	// Import the policy
	if fileStorage, ok := policyStorage.(*storage.FileStorage); ok {
		if err := fileStorage.ImportPolicy(inputFile); err != nil {
			fmt.Printf("❌ Failed to import policy: %v\n", err)
			return
		}
	} else {
		fmt.Println("❌ Import not supported with current storage type")
		return
	}

	fmt.Printf("✅ Policy imported successfully\n")
}

func runScanCommand(cmd *cobra.Command) {
	fmt.Println("🔍 Running policy scan...")

	// Get flags
	specificPolicy, _ := cmd.Flags().GetString("policy")
	verbose, _ := cmd.Flags().GetBool("verbose")
	outputFormat, _ := cmd.Flags().GetString("output")
	region, _ := cmd.Flags().GetString("region")

	// Set region if provided
	if region != "" {
		os.Setenv("AWS_REGION", region)
	}

	if specificPolicy != "" {
		fmt.Printf("🎯 Scanning specific policy: %s\n", specificPolicy)
		runSpecificPolicyScan(specificPolicy, verbose, outputFormat)
	} else {
		fmt.Println("🚀 Scanning all active policies")
		runScan() // Use the interactive function from main.go
	}
}

func runExecuteCommand(cmd *cobra.Command) {
	fmt.Println("⚡ Running policy execution...")

	// Get flags
	specificPolicy, _ := cmd.Flags().GetString("policy")
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	region, _ := cmd.Flags().GetString("region")

	// Set region if provided
	if region != "" {
		os.Setenv("AWS_REGION", region)
	}

	if dryRun {
		fmt.Println("🧪 Dry-run mode enabled - no changes will be made")
		runScanCommand(cmd)
		return
	}

	if !force {
		fmt.Print("⚠️  This will make real changes to AWS resources. Continue? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)

		if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
			fmt.Println("❌ Execution cancelled")
			return
		}
	}

	if specificPolicy != "" {
		fmt.Printf("🎯 Executing specific policy: %s\n", specificPolicy)
		runSpecificPolicyExecution(specificPolicy)
	} else {
		fmt.Println("🚀 Executing policies")
		executePolicy() // Use the interactive function from main.go
	}
}

func generateComplianceReportCmd(cmd *cobra.Command) {
	fmt.Println("📊 Generating compliance report...")

	outputFormat, _ := cmd.Flags().GetString("output")
	outputFile, _ := cmd.Flags().GetString("file")
	region, _ := cmd.Flags().GetString("region")

	// Set region if provided
	if region != "" {
		os.Setenv("AWS_REGION", region)
	}

	// Initialize AWS client (returns our stub)
	awsClient, err := initializeAWSClient(true)
	if err != nil {
		fmt.Printf("❌ Failed to initialize AWS client: %v\n", err)
		return
	}
	defer awsClient.Close()

	// Get resources (stub returns empty slices)
	ec2InstancesRaw, _ := awsClient.GetEC2Instances(nil)
	s3BucketsRaw, _ := awsClient.GetS3Buckets(nil)

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	fmt.Printf("📊 Found %d EC2 instances and %d S3 buckets (stub data)\n",
		len(ec2InstancesRaw), len(s3BucketsRaw))

	switch outputFormat {
	case "html":
		if outputFile == "" {
			outputFile = fmt.Sprintf("compliance_report_%s.html", timestamp)
		}
		fmt.Printf("📄 HTML compliance report would be generated: ./reports/%s\n", outputFile)
	case "json":
		if outputFile == "" {
			outputFile = fmt.Sprintf("compliance_report_%s.json", timestamp)
		}
		fmt.Printf("📄 JSON compliance report would be generated: ./reports/%s\n", outputFile)
	case "csv":
		if outputFile == "" {
			outputFile = fmt.Sprintf("compliance_summary_%s.csv", timestamp)
		}
		fmt.Printf("📄 CSV compliance report would be generated: ./reports/%s\n", outputFile)
	default:
		fmt.Printf("❌ Unsupported output format: %s\n", outputFormat)
		return
	}

	fmt.Println("✅ Compliance report generation completed (stub mode)")
}

func generateCostReportCmd(cmd *cobra.Command) {
	fmt.Println("💰 Generating cost report...")

	outputFormat, _ := cmd.Flags().GetString("output")
	outputFile, _ := cmd.Flags().GetString("file")
	region, _ := cmd.Flags().GetString("region")

	// Set region if provided
	if region != "" {
		os.Setenv("AWS_REGION", region)
	}

	// Initialize AWS client (stub)
	awsClient, err := initializeAWSClient(true)
	if err != nil {
		fmt.Printf("❌ Failed to initialize AWS client: %v\n", err)
		return
	}
	defer awsClient.Close()

	// Get resources (stub)
	ec2InstancesRaw, _ := awsClient.GetEC2Instances(nil)
	s3BucketsRaw, _ := awsClient.GetS3Buckets(nil)

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	fmt.Printf("📊 Found %d EC2 instances and %d S3 buckets (stub data)\n",
		len(ec2InstancesRaw), len(s3BucketsRaw))

	if outputFile == "" {
		outputFile = fmt.Sprintf("cost_report_%s.%s", timestamp, outputFormat)
	}

	fmt.Printf("💰 Cost optimization report would be generated: ./reports/%s\n", outputFile)

	// Show stub cost summary
	fmt.Println("\n📈 Cost Summary (sample data):")
	fmt.Printf("   • Current Monthly Cost: $1,234.56\n")
	fmt.Printf("   • Potential Savings: $456.78/month\n")
	fmt.Printf("   • Annual Savings: $5,481.36\n")

	fmt.Println("✅ Cost report generation completed (stub mode)")
}

func generateInventoryReportCmd(cmd *cobra.Command) {
	fmt.Println("📋 Generating inventory report...")

	outputFormat, _ := cmd.Flags().GetString("output")
	outputFile, _ := cmd.Flags().GetString("file")
	region, _ := cmd.Flags().GetString("region")

	// Set region if provided
	if region != "" {
		os.Setenv("AWS_REGION", region)
	}

	// Initialize AWS client (stub)
	awsClient, err := initializeAWSClient(true)
	if err != nil {
		fmt.Printf("❌ Failed to initialize AWS client: %v\n", err)
		return
	}
	defer awsClient.Close()

	// Get resources (stub)
	ec2InstancesRaw, _ := awsClient.GetEC2Instances(nil)
	s3BucketsRaw, _ := awsClient.GetS3Buckets(nil)

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	fmt.Printf("📊 Found %d EC2 instances and %d S3 buckets (stub data)\n",
		len(ec2InstancesRaw), len(s3BucketsRaw))

	if outputFile == "" {
		outputFile = fmt.Sprintf("inventory_report_%s.%s", timestamp, outputFormat)
	}

	fmt.Printf("📋 Resource inventory report would be generated: ./reports/%s\n", outputFile)

	// Show stub inventory summary
	fmt.Println("\n📊 Resource Summary (sample data):")
	fmt.Printf("   • EC2 Instances: 15 (12 running, 3 stopped)\n")
	fmt.Printf("   • S3 Buckets: 8 (2 public, 6 private)\n")
	fmt.Printf("   • EBS Volumes: 25 (5 unattached)\n")
	fmt.Printf("   • Security Groups: 12\n")

	fmt.Println("✅ Inventory report generation completed (stub mode)")
}

func showConfig() {
	fmt.Println("⚙️  Current Configuration:")
	fmt.Println("═════════════════════════")

	// Show environment variables
	fmt.Println("🌍 Environment Variables:")
	fmt.Printf("   AWS_REGION: %s\n", getEnvOrDefault("AWS_REGION", "not set"))
	fmt.Printf("   AWS_PROFILE: %s\n", getEnvOrDefault("AWS_PROFILE", "not set"))
	fmt.Printf("   AWS_ACCESS_KEY_ID: %s\n", maskCredential(os.Getenv("AWS_ACCESS_KEY_ID")))
	fmt.Printf("   AWS_SECRET_ACCESS_KEY: %s\n", maskCredential(os.Getenv("AWS_SECRET_ACCESS_KEY")))

	// Show storage info
	if policyStorage != nil {
		if fileStorage, ok := policyStorage.(*storage.FileStorage); ok {
			info, err := fileStorage.GetStorageInfo()
			if err == nil {
				fmt.Println("\n📁 Storage Configuration:")
				fmt.Printf("   Type: %s\n", info["storage_type"])
				fmt.Printf("   Directory: %s\n", info["base_directory"])
				fmt.Printf("   Policies: %d\n", info["policies_count"])
				fmt.Printf("   Storage Size: %.2f MB\n", info["storage_size_mb"])
			}
		}
	}
}

func testAWSConnection() {
	fmt.Println("🧪 Testing AWS connection...")

	awsClient, err := initializeAWSClient(true)
	if err != nil {
		fmt.Printf("❌ Failed to initialize AWS client: %v\n", err)
		return
	}
	defer awsClient.Close()

	fmt.Println("✅ AWS connection successful (stub mode)!")

	// Show basic info
	fmt.Println("\n📊 Connection Details:")
	fmt.Printf("   Region: %s\n", awsClient.Region)
	fmt.Printf("   Profile: %s\n", awsClient.Profile)

	// Test basic API calls
	fmt.Println("\n🔍 Testing API access...")

	regions, err := awsClient.GetRegions()
	if err != nil {
		fmt.Printf("⚠️  Failed to list regions: %v\n", err)
	} else {
		fmt.Printf("✅ Can access %d regions\n", len(regions))
	}

	quotas := awsClient.GetServiceQuotas()
	fmt.Println("\n📈 Service Quotas (estimates):")
	for service, quota := range quotas {
		fmt.Printf("   %s: %v\n", service, quota)
	}
}

// Helper functions
func runSpecificPolicyScan(policyName string, verbose bool, outputFormat string) {
	fmt.Printf("🎯 Scanning policy: %s\n", policyName)

	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized!")
		return
	}

	// Check if policy exists
	if !policyStorage.PolicyExists(policyName) {
		fmt.Printf("❌ Policy '%s' not found!\n", policyName)
		return
	}

	fmt.Printf("🔍 Policy scan mode: %s (verbose: %t)\n", outputFormat, verbose)
	fmt.Println("💡 Policy scanning implementation coming soon!")
}

func runSpecificPolicyExecution(policyName string) {
	fmt.Printf("⚡ Executing policy: %s\n", policyName)

	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized!")
		return
	}

	// Check if policy exists
	if !policyStorage.PolicyExists(policyName) {
		fmt.Printf("❌ Policy '%s' not found!\n", policyName)
		return
	}

	fmt.Println("💡 Policy execution implementation coming soon!")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func maskCredential(credential string) string {
	if credential == "" {
		return "not set"
	}
	if len(credential) > 8 {
		return credential[:4] + "****" + credential[len(credential)-4:]
	}
	return "****"
}

// Commented out functions that require real AWS types - uncomment when implementing real AWS integration
/*
func generateCostAnalysisReport(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
	timestamp string,
) {
	fmt.Println("\n💰 Generating cost analysis report...")

	csvGen := reports.NewCSVReportGenerator("./reports")

	// Generate cost analysis CSV
	costFilename := fmt.Sprintf("cost_analysis_%s.csv", timestamp)
	if err := csvGen.GenerateCostAnalysisReport(ec2Instances, s3Buckets, costFilename); err != nil {
		fmt.Printf("❌ Failed to generate cost analysis report: %v\n", err)
		return
	}

	fmt.Printf("\n🎉 Cost analysis report generated successfully!\n")
	fmt.Printf("📁 File saved: ./reports/%s\n", costFilename)
	fmt.Printf("📊 Open the CSV file in Excel or Google Sheets\n")

	// Show quick summary
	totalMonthlyCost := 0.0
	potentialSavings := 0.0

	for _, instance := range ec2Instances {
		totalMonthlyCost += instance.MonthlyCost
		if instance.CPUUtilization < 5.0 && instance.RunningDays > 7 {
			potentialSavings += instance.MonthlyCost
		}
	}

	for _, bucket := range s3Buckets {
		totalMonthlyCost += bucket.MonthlyCostEstimate
	}

	fmt.Printf("\n📈 Cost Summary:\n")
	fmt.Printf("   • Current Monthly Cost: $%.2f\n", totalMonthlyCost)
	fmt.Printf("   • Potential Savings: $%.2f/month\n", potentialSavings)
	fmt.Printf("   • Annual Savings: $%.2f\n", potentialSavings*12)
}

func generateResourceInventoryReport(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
	timestamp string,
) {
	fmt.Println("\n📋 Generating resource inventory reports...")

	csvGen := reports.NewCSVReportGenerator("./reports")

	// Generate EC2 inventory
	ec2Filename := fmt.Sprintf("ec2_inventory_%s.csv", timestamp)
	if err := csvGen.GenerateEC2Report(ec2Instances, ec2Filename); err != nil {
		fmt.Printf("❌ Failed to generate EC2 report: %v\n", err)
	} else {
		fmt.Printf("✅ EC2 inventory saved: ./reports/%s\n", ec2Filename)
	}

	// Generate S3 inventory
	s3Filename := fmt.Sprintf("s3_inventory_%s.csv", timestamp)
	if err := csvGen.GenerateS3Report(s3Buckets, s3Filename); err != nil {
		fmt.Printf("❌ Failed to generate S3 report: %v\n", err)
	} else {
		fmt.Printf("✅ S3 inventory saved: ./reports/%s\n", s3Filename)
	}

	// Generate compliance summary
	summaryFilename := fmt.Sprintf("compliance_summary_%s.csv", timestamp)
	if err := csvGen.GenerateComplianceSummaryReport(ec2Instances, s3Buckets, summaryFilename); err != nil {
		fmt.Printf("❌ Failed to generate compliance summary: %v\n", err)
	} else {
		fmt.Printf("✅ Compliance summary saved: ./reports/%s\n", summaryFilename)
	}

	fmt.Printf("\n🎉 Resource inventory reports generated successfully!\n")
	fmt.Printf("📁 All files saved in: ./reports/\n")
}

func generateComplianceReportHTML(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
	filename string,
) {
	htmlGen := reports.NewHTMLReportGenerator("./reports")
	report, err := htmlGen.GenerateComplianceReport(ec2Instances, s3Buckets)
	if err != nil {
		fmt.Printf("❌ Failed to generate report: %v\n", err)
		return
	}

	if err := htmlGen.SaveHTMLReport(report, filename); err != nil {
		fmt.Printf("❌ Failed to save report: %v\n", err)
		return
	}

	fmt.Printf("✅ HTML compliance report saved: ./reports/%s\n", filename)
}

func generateComplianceReportJSON(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
	filename string,
) {
	jsonGen := reports.NewJSONReportGenerator("./reports")
	report, err := jsonGen.GenerateComplianceReportJSON(ec2Instances, s3Buckets)
	if err != nil {
		fmt.Printf("❌ Failed to generate report: %v\n", err)
		return
	}

	if err := jsonGen.SaveJSONReport(report, filename); err != nil {
		fmt.Printf("❌ Failed to save report: %v\n", err)
		return
	}

	fmt.Printf("✅ JSON compliance report saved: ./reports/%s\n", filename)
}

func generateComplianceReportCSV(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
	filename string,
) {
	csvGen := reports.NewCSVReportGenerator("./reports")

	if err := csvGen.GenerateComplianceSummaryReport(ec2Instances, s3Buckets, filename); err != nil {
		fmt.Printf("❌ Failed to generate report: %v\n", err)
		return
	}

	fmt.Printf("✅ CSV compliance report saved: ./reports/%s\n", filename)
}
*/
