package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "1.0.0"

func main() {
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

// startPolicyCreation is implemented in wizard.go

// Policy management functions are implemented in other files
// listPolicies is in wizard.go
// runScan, executePolicy, generateReport, configureSettings are in commands.go
