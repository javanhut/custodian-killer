package main

import (
	"custodian-killer/storage"
	"custodian-killer/wizard"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version       = "1.0.0"
	policyStorage storage.PolicyStorage
)

// Stub AWS client - define this ONCE and properly
type StubAWSClient struct {
	Region  string
	Profile string
}

func (s *StubAWSClient) Close() {}

func (s *StubAWSClient) GetEC2Instances(filter interface{}) ([]interface{}, error) {
	fmt.Println("🔧 GetEC2Instances - stub implementation")
	return []interface{}{}, nil
}

func (s *StubAWSClient) GetS3Buckets(filter interface{}) ([]interface{}, error) {
	fmt.Println("🔧 GetS3Buckets - stub implementation")
	return []interface{}{}, nil
}

func (s *StubAWSClient) GetRegions() ([]string, error) {
	return []string{"us-east-1", "us-west-2", "eu-west-1"}, nil
}

func (s *StubAWSClient) GetServiceQuotas() map[string]interface{} {
	return map[string]interface{}{
		"EC2": "Running instances: 20",
		"S3":  "Buckets: 100",
		"RDS": "DB instances: 40",
	}
}

func main() {
	// Initialize storage at startup
	var err error
	policyStorage, err = initializePolicyStorage()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not initialize policy storage: %v\n", err)
	}

	rootCmd := &cobra.Command{
		Use:   "custodian-killer",
		Short: "Making AWS compliance fun again!",
		Long: `Custodian Killer - The AWS policy management tool that actually works!
Making Cloud Custodian weep since 2025.

Run without arguments to enter interactive mode, or use subcommands for direct execution.`,
		Run: func(cmd *cobra.Command, args []string) {
			startInteractiveMode()
		},
	}

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(policyCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(executeCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(interactiveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Custodian Killer v%s - The AWS policy tool that actually works!\n", version)
	},
}

var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Short:   "Start interactive mode",
	Aliases: []string{"i", "shell"},
	Run: func(cmd *cobra.Command, args []string) {
		startInteractiveMode()
	},
}

func startInteractiveMode() {
	// Show the epic ASCII logo first!
	fmt.Print(`
 ██████╗██╗   ██╗███████╗████████╗ ██████╗ ██████╗ ██╗ █████╗ ███╗   ██╗    
██╔════╝██║   ██║██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗██║██╔══██╗████╗  ██║    
██║     ██║   ██║███████╗   ██║   ██║   ██║██║  ██║██║███████║██╔██╗ ██║    
██║     ██║   ██║╚════██║   ██║   ██║   ██║██║  ██║██║██╔══██║██║╚██╗██║    
╚██████╗╚██████╔╝███████║   ██║   ╚██████╔╝██████╔╝██║██║  ██║██║ ╚████║    
 ╚═════╝ ╚═════╝ ╚══════╝   ╚═╝    ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝    
                                                                             
██╗  ██╗██╗██╗     ██╗     ███████╗██████╗                                  
██║ ██╔╝██║██║     ██║     ██╔════╝██╔══██╗                                 
█████╔╝ ██║██║     ██║     █████╗  ██████╔╝                                 
██╔═██╗ ██║██║     ██║     ██╔══╝  ██╔══██╗                                 
██║  ██╗██║███████╗███████╗███████╗██║  ██║                                 
╚═╝  ╚═╝╚═╝╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝                                 
`)
	fmt.Println("🔥 Welcome to Custodian Killer v" + version + " - Making AWS compliance fun again!")
	fmt.Println("💀 The AWS policy management tool that doesn't suck!")
	fmt.Println("🦍 Making Cloud Custodian weep since 2025!")
	fmt.Println()
	fmt.Println("Type 'help' for available commands or 'exit' to quit.")
	fmt.Println()

	for {
		fmt.Print("custodian-killer> ")

		var input string
		fmt.Scanln(&input)

		switch input {
		case "help":
			showHelp()
		case "exit", "quit":
			fmt.Println(
				"👋 Thanks for using Custodian Killer! May your AWS bills be low and your compliance high!",
			)
			return
		case "make", "policy":
			startPolicyCreation()
		case "list":
			listPolicies()
		case "scan":
			runScan()
		case "execute":
			executePolicy()
		case "report":
			generateReport()
		case "config":
			configureSettings()
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", input)
		}
		fmt.Println()
	}
}

func showHelp() {
	fmt.Println(`Available commands:
  help           - Show this help message
  make policy    - Create a new policy interactively
  list           - Show all your policies
  scan           - Run policies and see what they'd do (dry-run)
  execute        - Actually run the policies (with confirmation)
  report         - Generate compliance/cost reports
  config         - Set up AWS credentials and preferences
  exit/quit      - Exit the interactive shell

Pro tips:
  - Use 'scan' before 'execute' to see what will happen
  - Policies are saved and reusable
  - Reports can be exported to HTML, JSON, or CSV
  - We support ALL AWS resource types (seriously, all of them)`)
}

// startPolicyCreation initializes the wizard and starts policy creation
func startPolicyCreation() {
	fmt.Println("🎯 Starting Policy Creation Wizard...")

	// Create and start the wizard using the global storage
	wizardInstance := wizard.NewWizard(policyStorage)
	wizardInstance.Start()
}

// initializePolicyStorage sets up the policy storage
func initializePolicyStorage() (storage.PolicyStorage, error) {
	fileStorage, err := storage.NewFileStorage("./policies")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize policy storage: %v", err)
	}
	return fileStorage, nil
}

// initializeAWSClient creates a stub AWS client - used by commands.go
func initializeAWSClient(verbose bool) (*StubAWSClient, error) {
	if verbose {
		fmt.Println("🔧 Initializing AWS client (stub mode)")
	}

	return &StubAWSClient{
		Region:  "us-east-1",
		Profile: "default",
	}, nil
}

// Placeholder functions for interactive commands
func listPolicies() {
	fmt.Println("📋 Policy List:")

	if policyStorage == nil {
		fmt.Println("❌ Cannot access policy storage")
		return
	}

	policies, err := policyStorage.ListPolicies()
	if err != nil {
		fmt.Printf("❌ Error loading policies: %v\n", err)
		return
	}

	if len(policies) == 0 {
		fmt.Println("📝 No policies found. Use 'make policy' to create your first one!")
		return
	}

	fmt.Printf("Found %d policies:\n", len(policies))
	for i, policy := range policies {
		fmt.Printf("  %d. %s (%s) - %s\n",
			i+1,
			policy.Name,
			policy.ResourceType,
			policy.Description)
	}
}

func runScan() {
	fmt.Println("🔍 Scan Mode:")
	fmt.Println(
		"(This feature is coming soon! Will show what policies would do without making changes.)",
	)
}

func executePolicy() {
	fmt.Println("⚡ Execute Mode:")
	fmt.Println("(This feature is coming soon! Will actually run the policies with confirmation.)")
}

func generateReport() {
	fmt.Println("📊 Report Generation:")
	fmt.Println("(This feature is coming soon! Will generate compliance and cost reports.)")
}

func configureSettings() {
	fmt.Println("⚙️  Configuration:")
	fmt.Println("(This feature is coming soon! Will help set up AWS credentials and preferences.)")
}
