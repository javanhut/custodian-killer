package main

import (
	"bufio"
	"custodian-killer/aws"
	"custodian-killer/reports"
	"custodian-killer/scanner"
	"custodian-killer/storage"
	"custodian-killer/templates"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Global storage instance
var policyStorage storage.PolicyStorage

func init() {
	var err error
	policyStorage, err = storage.NewFileStorage("")
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to initialize storage: %v\n", err)
	}
}

func startPolicyCreation() {
	fmt.Println("🎯 Welcome to the Policy Creation Wizard!")
	fmt.Println("Let's build something awesome together. I'll guide you through every step.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Step 1: Choose creation method
	fmt.Println("How would you like to create your policy?")
	fmt.Println("1. 🏗️  Build from scratch (full control)")
	fmt.Println("2. 🚀 Use a template (quick start)")
	fmt.Println("3. 🤖 AI-assisted creation (describe what you want)")

	choice := getChoice(reader, 1, 3, "Choose option (1-3): ")
	fmt.Println()

	switch choice {
	case 1:
		createPolicyFromScratch(reader)
	case 2:
		createPolicyFromTemplate(reader)
	case 3:
		createPolicyWithAI(reader)
	}
}

func createPolicyFromScratch(reader *bufio.Reader) {
	fmt.Println("🏗️  Building policy from scratch - You've got full control!")
	fmt.Println()

	var policy Policy

	// Get basic info
	policy.Name = getInput(reader, "📝 Policy name: ")
	policy.Description = getInput(reader, "📋 Description: ")

	// Choose resource type
	fmt.Println("\n🎯 What AWS resource type should this policy target?")
	resourceTypes := GetResourceTypes()
	for i, rt := range resourceTypes {
		resourceInfo := SupportedResources[rt]
		fmt.Printf("%d. %s - %s\n", i+1, strings.ToUpper(rt), resourceInfo.Description)
	}

	choice := getChoice(reader, 1, len(resourceTypes), "Choose resource type: ")
	policy.ResourceType = resourceTypes[choice-1]

	fmt.Printf("\n✅ Great! Working with %s resources.\n", strings.ToUpper(policy.ResourceType))

	// Add filters
	policy.Filters = createFilters(reader, policy.ResourceType)

	// Add actions
	policy.Actions = createActions(reader, policy.ResourceType)

	// Set execution mode
	policy.Mode = createPolicyMode(reader)

	// Show summary and save
	showPolicySummary(policy)

	if confirmSave(reader) {
		savePolicy(policy)
		fmt.Println("🎉 Policy saved successfully! You can now scan or execute it.")
	}
}

func createPolicyFromTemplate(reader *bufio.Reader) {
	fmt.Println("🚀 Using templates for quick policy creation!")
	fmt.Println()

	templateManager := templates.NewTemplateManager()

	// Show template categories
	fmt.Println("What kind of policy do you need?")
	categories := templateManager.GetCategories()
	for i, category := range categories {
		fmt.Printf("%d. %s\n", i+1, strings.Title(strings.ReplaceAll(category, "-", " ")))
	}
	fmt.Printf("%d. 🔍 Search templates\n", len(categories)+1)
	fmt.Printf("%d. 📊 Show all templates\n", len(categories)+2)

	categoryChoice := getChoice(reader, 1, len(categories)+2, "Choose category: ")

	var availableTemplates []templates.PolicyTemplate

	if categoryChoice == len(categories)+1 {
		// Search templates
		query := getInput(reader, "Search for templates (keywords): ")
		availableTemplates = templateManager.SearchTemplates(query)
	} else if categoryChoice == len(categories)+2 {
		// Show all templates
		availableTemplates = templateManager.GetAllTemplates()
	} else {
		// Filter by category
		selectedCategory := categories[categoryChoice-1]
		availableTemplates = templateManager.GetTemplatesByCategory(selectedCategory)
	}

	if len(availableTemplates) == 0 {
		fmt.Println("❌ No templates found matching your criteria!")
		return
	}

	fmt.Printf("\n📋 Found %d templates:\n", len(availableTemplates))
	for i, template := range availableTemplates {
		fmt.Printf("%d. 🎯 %s (%s impact)\n", i+1, template.Name, template.Impact)
		fmt.Printf("   📝 %s\n", template.Description)
		fmt.Printf("   🏷️  %s | %s\n", template.ResourceType, template.Difficulty)
		fmt.Println()
	}

	templateChoice := getChoice(reader, 1, len(availableTemplates), "Choose template: ")
	selectedTemplate := availableTemplates[templateChoice-1]

	fmt.Printf("\n✅ Selected: %s\n", selectedTemplate.Name)
	fmt.Printf("📋 Description: %s\n", selectedTemplate.Description)
	fmt.Printf("🎯 Resource Type: %s\n", selectedTemplate.ResourceType)
	fmt.Printf("⚠️  Impact: %s\n", selectedTemplate.Impact)

	// Show examples
	if len(selectedTemplate.Examples) > 0 {
		fmt.Println("\n💡 Examples of what this template does:")
		for _, example := range selectedTemplate.Examples {
			fmt.Printf("   • %s\n", example)
		}
	}

	// Customize template variables
	variables := make(map[string]interface{})

	// Get policy name first
	variables["policy_name"] = getInput(reader, "\n📝 Policy name: ")

	// Collect template variables
	if len(selectedTemplate.Variables) > 0 {
		fmt.Println("\n⚙️  Template Configuration:")
		for _, variable := range selectedTemplate.Variables {
			fmt.Printf("\n🔧 %s\n", variable.Description)
			if variable.DefaultValue != nil {
				fmt.Printf("   Default: %v\n", variable.DefaultValue)
			}
			if len(variable.Options) > 0 {
				fmt.Printf("   Options: %s\n", strings.Join(variable.Options, ", "))
			}

			prompt := fmt.Sprintf("   %s", variable.Name)
			if !variable.Required {
				prompt += " (optional)"
			}
			prompt += ": "

			input := getInput(reader, prompt)

			// Use default if empty and not required
			if input == "" && !variable.Required && variable.DefaultValue != nil {
				variables[variable.Name] = variable.DefaultValue
			} else if input != "" {
				// Type conversion based on variable type
				switch variable.Type {
				case "int":
					if intVal, err := strconv.Atoi(input); err == nil {
						variables[variable.Name] = intVal
					} else {
						fmt.Printf("   ⚠️  Invalid integer, using default: %v\n", variable.DefaultValue)
						variables[variable.Name] = variable.DefaultValue
					}
				case "bool":
					variables[variable.Name] = strings.ToLower(input) == "true" || input == "1"
				case "list":
					variables[variable.Name] = strings.Split(input, ",")
				default:
					variables[variable.Name] = input
				}
			} else if variable.Required {
				fmt.Printf("   ⚠️  Required field, using default: %v\n", variable.DefaultValue)
				variables[variable.Name] = variable.DefaultValue
			}
		}
	}

	// Generate policy from template
	policyDef, err := templateManager.InstantiateTemplate(selectedTemplate.ID, variables)
	if err != nil {
		fmt.Printf("❌ Error creating policy from template: %v\n", err)
		return
	}

	// Convert template policy to our main Policy struct
	policy := convertTemplatePolicyToPolicy(policyDef)

	// Show summary and save
	showPolicySummary(policy)

	fmt.Println("\n🎉 Template customization complete!")
	if confirmSave(reader) {
		savePolicy(policy)
		fmt.Println("🚀 Policy created from template! Ready to use.")
	}
}

func createPolicyWithAI(reader *bufio.Reader) {
	fmt.Println("🤖 AI-Assisted Policy Creation")
	fmt.Println("Just describe what you want in plain English, and I'll build the perfect policy!")
	fmt.Println()

	description := getInput(reader, "📝 Describe what you want this policy to do: ")

	fmt.Println("\n🧠 Analyzing your request...")
	fmt.Println("💭 Understanding your requirements...")

	// Enhanced AI policy suggestion
	suggestedPolicy, confidence, explanation := smartPolicySuggestion(description)

	fmt.Printf("\n💡 AI Confidence: %d%%\n", confidence)
	fmt.Printf("🎯 Analysis: %s\n", explanation)
	fmt.Println("\n🤖 Based on your description, I suggest this policy:")

	showPolicySummary(suggestedPolicy)

	fmt.Println("\n🎛️  What would you like to do?")
	fmt.Println("1. ✅ Perfect! Save it as-is")
	fmt.Println("2. ✏️  Let me tweak some settings")
	fmt.Println("3. 🔄 Try describing it differently")
	fmt.Println("4. 🎯 Show me more details about this policy")

	choice := getChoice(reader, 1, 4, "Choose option: ")

	switch choice {
	case 1:
		if confirmSave(reader) {
			savePolicy(suggestedPolicy)
			fmt.Println("🎉 AI-generated policy saved! The machines are learning... 🤖")
		}
	case 2:
		policy := tweakAIPolicy(reader, suggestedPolicy)
		if confirmSave(reader) {
			savePolicy(policy)
			fmt.Println("🎉 Customized AI policy saved!")
		}
	case 3:
		createPolicyWithAI(reader) // Try again
	case 4:
		explainPolicyInDetail(suggestedPolicy, explanation)
		// Ask again what to do
		fmt.Println("\nNow that you understand the policy better:")
		createPolicyWithAI(reader)
	}
}

// Enhanced AI policy suggestion with better intelligence
func smartPolicySuggestion(description string) (Policy, int, string) {
	lower := strings.ToLower(description)
	words := strings.Fields(lower)

	var policy Policy
	policy.Mode = PolicyMode{Type: "pull"}
	confidence := 50 // Start with base confidence
	var explanation strings.Builder

	// Detect resource type
	resourceType, resourceConfidence := detectResourceType(lower, words)
	policy.ResourceType = resourceType
	confidence += resourceConfidence
	explanation.WriteString(fmt.Sprintf("Detected resource type '%s' ", resourceType))

	// Detect intent/action
	intent, filters, actions, intentConfidence, intentExplanation := detectIntent(
		lower,
		words,
		resourceType,
	)
	policy.Filters = filters
	policy.Actions = actions
	confidence += intentConfidence
	explanation.WriteString(intentExplanation)

	// Generate policy name
	policy.Name = generatePolicyName(intent, resourceType)
	policy.Description = generatePolicyDescription(description, intent, resourceType)

	// Cap confidence at 95% (never be too sure!)
	if confidence > 95 {
		confidence = 95
	}

	return policy, confidence, explanation.String()
}

func detectResourceType(description string, words []string) (string, int) {
	// Resource type detection with confidence scoring
	resourceKeywords := map[string][]string{
		"ec2": {
			"ec2",
			"instance",
			"instances",
			"server",
			"servers",
			"vm",
			"virtual machine",
			"compute",
		},
		"s3":     {"s3", "bucket", "buckets", "storage", "object storage", "files"},
		"rds":    {"rds", "database", "db", "mysql", "postgres", "sql", "aurora"},
		"lambda": {"lambda", "function", "functions", "serverless", "code"},
		"iam": {
			"iam",
			"user",
			"users",
			"role",
			"roles",
			"permission",
			"permissions",
			"policy",
			"policies",
		},
		"vpc": {"vpc", "network", "networking", "subnet", "subnets", "security group"},
		"ebs": {"ebs", "volume", "volumes", "disk", "disks", "storage"},
		"elb": {"elb", "load balancer", "loadbalancer", "alb", "nlb"},
	}

	scores := make(map[string]int)

	for resourceType, keywords := range resourceKeywords {
		for _, keyword := range keywords {
			if strings.Contains(description, keyword) {
				scores[resourceType] += 20
				// Bonus for exact matches
				for _, word := range words {
					if word == keyword {
						scores[resourceType] += 10
					}
				}
			}
		}
	}

	// Find highest scoring resource type
	maxScore := 0
	bestResource := "ec2" // Default fallback

	for resource, score := range scores {
		if score > maxScore {
			maxScore = score
			bestResource = resource
		}
	}

	return bestResource, maxScore
}

func detectIntent(
	description string,
	words []string,
	resourceType string,
) (string, []Filter, []Action, int, string) {
	// Intent detection patterns
	intentPatterns := map[string]struct {
		keywords    []string
		confidence  int
		explanation string
	}{
		"cost-optimization": {
			keywords: []string{
				"unused",
				"idle",
				"waste",
				"cost",
				"expensive",
				"cheap",
				"save",
				"money",
				"bill",
				"optimize",
			},
			confidence:  25,
			explanation: "from cost optimization keywords. ",
		},
		"security": {
			keywords: []string{
				"public",
				"secure",
				"security",
				"private",
				"encrypt",
				"encryption",
				"vulnerable",
				"exposed",
			},
			confidence:  25,
			explanation: "from security-related keywords. ",
		},
		"compliance": {
			keywords: []string{
				"tag",
				"tags",
				"untagged",
				"missing",
				"required",
				"comply",
				"compliance",
				"standard",
				"policy",
			},
			confidence:  25,
			explanation: "from compliance keywords. ",
		},
		"cleanup": {
			keywords: []string{
				"old",
				"delete",
				"remove",
				"clean",
				"cleanup",
				"unused",
				"orphaned",
				"stale",
			},
			confidence:  25,
			explanation: "from cleanup-related keywords. ",
		},
	}

	// Score each intent
	intentScores := make(map[string]int)
	var explanations []string

	for intent, pattern := range intentPatterns {
		for _, keyword := range pattern.keywords {
			if strings.Contains(description, keyword) {
				intentScores[intent] += pattern.confidence
				explanations = append(explanations, pattern.explanation)
			}
		}
	}

	// Find best intent
	maxScore := 0
	bestIntent := "general"

	for intent, score := range intentScores {
		if score > maxScore {
			maxScore = score
			bestIntent = intent
		}
	}

	// Generate filters and actions based on intent and resource type
	filters, actions := generateFiltersAndActions(bestIntent, resourceType, description, words)

	explanation := fmt.Sprintf("and detected '%s' intent ", bestIntent)
	if len(explanations) > 0 {
		explanation += explanations[0]
	}

	return bestIntent, filters, actions, maxScore, explanation
}

func generateFiltersAndActions(
	intent, resourceType, description string,
	words []string,
) ([]Filter, []Action) {
	var filters []Filter
	var actions []Action

	switch intent {
	case "cost-optimization":
		switch resourceType {
		case "ec2":
			filters = append(filters, Filter{Type: "instance-state", Value: "running", Op: "eq"})
			filters = append(filters, Filter{Type: "cpu-utilization-avg", Value: 5, Op: "lt"})
			if containsAny(words, []string{"old", "days", "week", "month"}) {
				filters = append(filters, Filter{Type: "running-days", Value: 7, Op: "gte"})
			}

			if containsAny(words, []string{"stop", "shutdown"}) {
				actions = append(actions, Action{Type: "stop", DryRun: true})
			} else if containsAny(words, []string{"terminate", "delete"}) {
				actions = append(actions, Action{Type: "terminate", DryRun: true})
			} else {
				actions = append(actions, Action{Type: "stop", DryRun: true}) // Default
			}

		case "ebs":
			filters = append(
				filters,
				Filter{Type: "state", Value: "available", Op: "eq"},
			) // Unattached volumes
			actions = append(actions, Action{Type: "delete", DryRun: true})
		}

	case "security":
		switch resourceType {
		case "s3":
			filters = append(filters, Filter{Type: "public-read", Value: true, Op: "eq"})
			actions = append(actions, Action{Type: "block-public-access", DryRun: true})

		case "ec2":
			if containsAny(words, []string{"security", "group", "port", "open"}) {
				filters = append(filters, Filter{Type: "security-group", Value: "open", Op: "eq"})
				actions = append(actions, Action{Type: "modify-security-group", DryRun: true})
			}
		}

	case "compliance":
		// Tag compliance is common across all resource types
		requiredTags := extractTags(description, words)
		if len(requiredTags) == 0 {
			requiredTags = []string{"Environment", "Owner"} // Default required tags
		}

		for _, tag := range requiredTags {
			filters = append(filters, Filter{Type: "tag-missing", Key: tag, Op: "missing"})
		}

		tagSettings := make(map[string]interface{})
		for _, tag := range requiredTags {
			tagSettings[tag] = "auto-tagged"
		}
		tagSettings["AutoTagged"] = "true"
		tagSettings["TaggedDate"] = "{{.current_date}}"

		actions = append(actions, Action{
			Type:     "tag",
			Settings: tagSettings,
			DryRun:   true,
		})

	case "cleanup":
		switch resourceType {
		case "ebs":
			if containsAny(words, []string{"snapshot", "snapshots"}) {
				filters = append(filters, Filter{Type: "age", Value: 30, Op: "gt"})
				actions = append(actions, Action{Type: "delete", DryRun: true})
			}
		case "ec2":
			if containsAny(words, []string{"terminated", "stopped"}) {
				filters = append(
					filters,
					Filter{Type: "instance-state", Value: "stopped", Op: "eq"},
				)
				filters = append(filters, Filter{Type: "stopped-days", Value: 30, Op: "gt"})
				actions = append(actions, Action{Type: "terminate", DryRun: true})
			}
		}
	}

	// If no specific filters/actions were generated, create generic ones
	if len(filters) == 0 {
		filters = append(filters, Filter{Type: "tag-missing", Key: "Environment", Op: "missing"})
	}
	if len(actions) == 0 {
		actions = append(
			actions,
			Action{
				Type:     "tag",
				Settings: map[string]interface{}{"AutoGenerated": "true"},
				DryRun:   true,
			},
		)
	}

	return filters, actions
}

func containsAny(words []string, targets []string) bool {
	for _, word := range words {
		for _, target := range targets {
			if strings.Contains(word, target) || strings.Contains(target, word) {
				return true
			}
		}
	}
	return false
}

func extractTags(description string, words []string) []string {
	var tags []string

	// Look for common tag patterns
	commonTags := []string{"environment", "owner", "project", "team", "cost-center", "department"}

	for _, tag := range commonTags {
		if strings.Contains(description, tag) {
			tags = append(tags, strings.Title(tag))
		}
	}

	return tags
}

func generatePolicyName(intent, resourceType string) string {
	timestamp := time.Now().Format("0102") // MMDD format
	switch intent {
	case "cost-optimization":
		return fmt.Sprintf("cost-optimizer-%s-%s", resourceType, timestamp)
	case "security":
		return fmt.Sprintf("security-enforcer-%s-%s", resourceType, timestamp)
	case "compliance":
		return fmt.Sprintf("compliance-checker-%s-%s", resourceType, timestamp)
	case "cleanup":
		return fmt.Sprintf("cleanup-agent-%s-%s", resourceType, timestamp)
	default:
		return fmt.Sprintf("ai-generated-%s-%s", resourceType, timestamp)
	}
}

func generatePolicyDescription(originalDescription, intent, resourceType string) string {
	base := fmt.Sprintf("AI-generated policy for %s resources", resourceType)

	switch intent {
	case "cost-optimization":
		return fmt.Sprintf(
			"%s to optimize costs by managing unused resources. Original request: %s",
			base,
			originalDescription,
		)
	case "security":
		return fmt.Sprintf(
			"%s to enforce security best practices. Original request: %s",
			base,
			originalDescription,
		)
	case "compliance":
		return fmt.Sprintf(
			"%s to ensure compliance with organizational standards. Original request: %s",
			base,
			originalDescription,
		)
	case "cleanup":
		return fmt.Sprintf(
			"%s to clean up and organize resources. Original request: %s",
			base,
			originalDescription,
		)
	default:
		return fmt.Sprintf("%s based on: %s", base, originalDescription)
	}
}

func tweakAIPolicy(reader *bufio.Reader, policy Policy) Policy {
	fmt.Println("\n🔧 Let's customize your AI-generated policy!")

	for {
		fmt.Println("\nWhat would you like to modify?")
		fmt.Println("1. 📝 Change name or description")
		fmt.Println("2. 🔍 Adjust filters")
		fmt.Println("3. ⚡ Modify actions")
		fmt.Println("4. 🎛️  Change execution mode")
		fmt.Println("5. ✅ Looks good, I'm done tweaking")

		choice := getChoice(reader, 1, 5, "Choose what to modify: ")

		switch choice {
		case 1:
			fmt.Printf("\nCurrent name: %s\n", policy.Name)
			newName := getInput(reader, "New name (or press Enter to keep): ")
			if newName != "" {
				policy.Name = newName
			}

			fmt.Printf("\nCurrent description: %s\n", policy.Description)
			newDesc := getInput(reader, "New description (or press Enter to keep): ")
			if newDesc != "" {
				policy.Description = newDesc
			}

		case 2:
			fmt.Println("\n🔍 Current filters:")
			for i, filter := range policy.Filters {
				fmt.Printf("%d. %s %s %v\n", i+1, filter.Type, filter.Op, filter.Value)
			}

			fmt.Println("\nFilter options:")
			fmt.Println("1. Add a new filter")
			fmt.Println("2. Remove a filter")
			fmt.Println("3. Keep filters as-is")

			filterChoice := getChoice(reader, 1, 3, "Choose: ")
			if filterChoice == 1 {
				newFilter := createSingleFilter(reader, "custom")
				policy.Filters = append(policy.Filters, newFilter)
			} else if filterChoice == 2 && len(policy.Filters) > 0 {
				removeIdx := getChoice(reader, 1, len(policy.Filters), "Remove which filter? ") - 1
				policy.Filters = append(policy.Filters[:removeIdx], policy.Filters[removeIdx+1:]...)
			}

		case 3:
			fmt.Println("\n⚡ Current actions:")
			for i, action := range policy.Actions {
				dryRun := ""
				if action.DryRun {
					dryRun = " (DRY RUN)"
				}
				fmt.Printf("%d. %s%s\n", i+1, action.Type, dryRun)
			}

			fmt.Println("\nAction options:")
			fmt.Println("1. Add a new action")
			fmt.Println("2. Toggle dry-run mode")
			fmt.Println("3. Remove an action")
			fmt.Println("4. Keep actions as-is")

			actionChoice := getChoice(reader, 1, 4, "Choose: ")
			switch actionChoice {
			case 1:
				newAction := createSingleAction(reader, "custom")
				policy.Actions = append(policy.Actions, newAction)
			case 2:
				if len(policy.Actions) > 0 {
					toggleIdx := getChoice(
						reader,
						1,
						len(policy.Actions),
						"Toggle dry-run for which action? ",
					) - 1
					policy.Actions[toggleIdx].DryRun = !policy.Actions[toggleIdx].DryRun
				}
			case 3:
				if len(policy.Actions) > 0 {
					removeIdx := getChoice(
						reader,
						1,
						len(policy.Actions),
						"Remove which action? ",
					) - 1
					policy.Actions = append(
						policy.Actions[:removeIdx],
						policy.Actions[removeIdx+1:]...)
				}
			}

		case 4:
			policy.Mode = createPolicyMode(reader)

		case 5:
			fmt.Println("✅ Customization complete!")
			return policy
		}

		// Show updated summary after each change
		fmt.Println("\n📋 Updated Policy Summary:")
		showPolicySummary(policy)
	}
}

func explainPolicyInDetail(policy Policy, aiExplanation string) {
	fmt.Println("\n🎓 Detailed Policy Explanation")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Printf("🤖 AI Analysis: %s\n\n", aiExplanation)

	fmt.Printf("📋 Policy: %s\n", policy.Name)
	fmt.Printf("📝 Description: %s\n", policy.Description)
	fmt.Printf("🎯 Resource Type: %s\n\n", strings.ToUpper(policy.ResourceType))

	fmt.Println("🔍 What the filters do:")
	for i, filter := range policy.Filters {
		fmt.Printf("%d. ", i+1)
		switch filter.Type {
		case "instance-state":
			fmt.Printf(
				"Only look at instances that are '%s' (not stopped or terminated)\n",
				filter.Value,
			)
		case "cpu-utilization-avg":
			fmt.Printf(
				"Find instances with average CPU usage below %v%% (likely unused)\n",
				filter.Value,
			)
		case "running-days":
			fmt.Printf(
				"Focus on instances running for %v+ days (long-running unused resources)\n",
				filter.Value,
			)
		case "tag-missing":
			fmt.Printf("Find resources missing the '%s' tag (compliance issue)\n", filter.Key)
		case "public-read":
			fmt.Printf("Identify S3 buckets that allow public read access (security risk)\n")
		default:
			fmt.Printf("Check for %s %s %v\n", filter.Type, filter.Op, filter.Value)
		}
	}

	fmt.Println("\n⚡ What the actions will do:")
	for i, action := range policy.Actions {
		fmt.Printf("%d. ", i+1)
		dryRunNote := ""
		if action.DryRun {
			dryRunNote = " (SAFE: Dry-run mode - won't actually make changes)"
		}

		switch action.Type {
		case "stop":
			fmt.Printf("Stop the matching instances to save money%s\n", dryRunNote)
		case "terminate":
			fmt.Printf("Permanently delete the matching instances%s\n", dryRunNote)
		case "tag":
			fmt.Printf("Add identifying tags to resources for better organization%s\n", dryRunNote)
		case "block-public-access":
			fmt.Printf("Remove public access from S3 buckets to improve security%s\n", dryRunNote)
		default:
			fmt.Printf("Execute '%s' action%s\n", action.Type, dryRunNote)
		}
	}

	fmt.Println("\n💡 Why this policy makes sense:")
	fmt.Println("   • The AI detected your intent and chose appropriate filters")
	fmt.Println("   • Actions are set to dry-run mode for safety")
	fmt.Println("   • You can test it with 'scan' before executing for real")
	fmt.Println("   • All actions can be customized if needed")
}

func createFilters(reader *bufio.Reader, resourceType string) []Filter {
	fmt.Println("\n🔍 Let's add some filters to target the right resources...")

	resourceInfo := SupportedResources[resourceType]
	var filters []Filter

	for {
		fmt.Println("\nAvailable filters for", strings.ToUpper(resourceType), ":")
		for i, filterType := range resourceInfo.Filters {
			fmt.Printf("%d. %s\n", i+1, filterType)
		}
		fmt.Printf("%d. ✅ Done adding filters\n", len(resourceInfo.Filters)+1)

		choice := getChoice(reader, 1, len(resourceInfo.Filters)+1, "Choose filter to add: ")

		if choice == len(resourceInfo.Filters)+1 {
			break
		}

		filterType := resourceInfo.Filters[choice-1]
		filter := createSingleFilter(reader, filterType)
		filters = append(filters, filter)

		fmt.Printf("✅ Added filter: %s\n", filterType)
	}

	return filters
}

func createSingleFilter(reader *bufio.Reader, filterType string) Filter {
	var filter Filter
	filter.Type = filterType

	switch filterType {
	case "tag":
		filter.Key = getInput(reader, "Tag key: ")
		filter.Value = getInput(reader, "Tag value (or leave empty to check for existence): ")
		if filter.Value == "" {
			filter.Op = "exists"
		} else {
			filter.Op = "eq"
		}
	case "instance-state", "state":
		fmt.Println("Common states: running, stopped, terminated, pending")
		filter.Value = getInput(reader, "State: ")
		filter.Op = "eq"
	case "creation-date", "launch-time":
		fmt.Println("Examples: '30 days ago', '2024-01-01', 'last week'")
		filter.Value = getInput(reader, "Date/time: ")
		filter.Op = "lt" // older than
	default:
		filter.Value = getInput(reader, fmt.Sprintf("Value for %s: ", filterType))
		filter.Op = "eq"
	}

	return filter
}

func createActions(reader *bufio.Reader, resourceType string) []Action {
	fmt.Println("\n⚡ Now let's define what actions to take on matching resources...")

	resourceInfo := SupportedResources[resourceType]
	var actions []Action

	for {
		fmt.Println("\nAvailable actions for", strings.ToUpper(resourceType), ":")
		for i, actionType := range resourceInfo.Actions {
			fmt.Printf("%d. %s\n", i+1, actionType)
		}
		fmt.Printf("%d. ✅ Done adding actions\n", len(resourceInfo.Actions)+1)

		choice := getChoice(reader, 1, len(resourceInfo.Actions)+1, "Choose action to add: ")

		if choice == len(resourceInfo.Actions)+1 {
			break
		}

		actionType := resourceInfo.Actions[choice-1]
		action := createSingleAction(reader, actionType)
		actions = append(actions, action)

		fmt.Printf("✅ Added action: %s\n", actionType)
	}

	return actions
}

func createSingleAction(reader *bufio.Reader, actionType string) Action {
	var action Action
	action.Type = actionType
	action.Settings = make(map[string]interface{})

	// Always ask about dry run for destructive actions
	destructive := []string{"delete", "terminate", "stop"}
	for _, d := range destructive {
		if actionType == d {
			fmt.Println("⚠️  This is a destructive action!")
			fmt.Println("1. 🧪 Dry run (recommended - see what would happen)")
			fmt.Println("2. 💥 Live execution (actually make changes)")

			choice := getChoice(reader, 1, 2, "Choose mode: ")
			action.DryRun = choice == 1
			break
		}
	}

	// Action-specific settings
	switch actionType {
	case "tag":
		tagKey := getInput(reader, "Tag key to add: ")
		tagValue := getInput(reader, "Tag value: ")
		action.Settings["key"] = tagKey
		action.Settings["value"] = tagValue
	case "stop":
		fmt.Println("Should instances be force-stopped if graceful stop fails?")
		force := getChoice(reader, 1, 2, "1. Graceful only  2. Force if needed: ") == 2
		action.Settings["force"] = force
	}

	return action
}

func createPolicyMode(reader *bufio.Reader) PolicyMode {
	fmt.Println("\n⏱️  How should this policy run?")
	fmt.Println("1. 🔄 On-demand (run when you tell it to)")
	fmt.Println("2. ⏰ Scheduled (runs automatically)")
	fmt.Println("3. 📡 Event-driven (responds to AWS events)")

	choice := getChoice(reader, 1, 3, "Choose execution mode: ")

	var mode PolicyMode
	switch choice {
	case 1:
		mode.Type = "pull"
	case 2:
		mode.Type = "periodic"
		mode.Schedule = getInput(
			reader,
			"Schedule (cron format, e.g., '0 2 * * *' for daily at 2am): ",
		)
	case 3:
		mode.Type = "event"
		// TODO: Add event configuration
	}

	return mode
}

// Initialize AWS client with configuration
func initializeAWSClient(forceDryRun bool) (*aws.CustodianClient, error) {
	fmt.Println("🔧 Initializing AWS connection...")

	// Get AWS configuration from environment or use defaults
	config := aws.ClientConfig{
		Region:          os.Getenv("AWS_REGION"),
		Profile:         os.Getenv("AWS_PROFILE"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
		DryRun:          forceDryRun,
		Timeout:         30 * time.Second,
	}

	// Set defaults
	if config.Region == "" {
		config.Region = "us-east-1"
	}
	if config.Profile == "" {
		config.Profile = "default"
	}

	return aws.NewCustodianClient(config)
}

// Display execution result in a nice format
func displayExecutionResult(result *ExecutionResult) {
	fmt.Printf("\n📊 Execution Results for: %s\n", result.PolicyName)
	fmt.Printf("⏰ Duration: %v\n", result.Duration.Round(time.Second))
	fmt.Printf("🎯 Resource Type: %s\n", strings.ToUpper(result.ResourceType))

	if result.DryRun {
		fmt.Println("🧪 DRY RUN MODE - No actual changes made")
	} else {
		fmt.Println("💥 LIVE MODE - Changes were applied")
	}

	// Summary
	fmt.Println("\n📈 Summary:")
	fmt.Printf("   • Resources Found: %d\n", result.ResourcesFound)
	fmt.Printf("   • Resources Matched: %d\n", result.ResourcesMatched)
	fmt.Printf("   • Actions Executed: %d\n", result.Summary.TotalActions)
	fmt.Printf("   • Successful Actions: %d\n", result.Summary.SuccessfulActions)

	if result.Summary.FailedActions > 0 {
		fmt.Printf("   • ❌ Failed Actions: %d\n", result.Summary.FailedActions)
	}

	if result.Summary.ResourcesModified > 0 {
		fmt.Printf("   • 🔧 Resources Modified: %d\n", result.Summary.ResourcesModified)
	}

	if result.Summary.EstimatedMonthlySavings > 0 {
		fmt.Printf(
			"   • 💰 Estimated Monthly Savings: $%.2f\n",
			result.Summary.EstimatedMonthlySavings,
		)
	}

	if result.Summary.SecurityImprovements > 0 {
		fmt.Printf("   • 🔒 Security Improvements: %d\n", result.Summary.SecurityImprovements)
	}

	// Cost Impact
	if result.CostImpact.MonthlySavings > 0 {
		fmt.Println("\n💰 Cost Impact:")
		fmt.Printf("   • Previous Monthly Cost: $%.2f\n", result.CostImpact.PreviousMonthlyCost)
		fmt.Printf("   • New Monthly Cost: $%.2f\n", result.CostImpact.NewMonthlyCost)
		fmt.Printf("   • Monthly Savings: $%.2f\n", result.CostImpact.MonthlySavings)
		fmt.Printf("   • Annual Savings: $%.2f\n", result.CostImpact.AnnualSavings)
	}

	// Errors
	if len(result.Errors) > 0 {
		fmt.Println("\n⚠️  Errors:")
		for _, err := range result.Errors {
			fmt.Printf("   • %s\n", err)
		}
	}

	// Action Details
	if len(result.ActionResults) > 0 {
		fmt.Println("\n⚡ Action Details:")
		actionGroups := make(map[string][]ActionResult)

		// Group actions by type
		for _, actionResult := range result.ActionResults {
			actionGroups[actionResult.Action] = append(
				actionGroups[actionResult.Action],
				actionResult,
			)
		}

		// Display each action group
		for actionType, actions := range actionGroups {
			successful := 0
			failed := 0

			for _, action := range actions {
				if action.Success {
					successful++
				} else {
					failed++
				}
			}

			fmt.Printf("\n   🔧 %s:\n", strings.ToUpper(actionType))
			fmt.Printf("      ✅ Successful: %d | ❌ Failed: %d\n", successful, failed)

			// Show details for first few actions
			displayCount := len(actions)
			if displayCount > 5 {
				displayCount = 5
			}

			for _, action := range actions[:displayCount] {
				icon := "✅"
				if !action.Success {
					icon = "❌"
				}

				dryRunStr := ""
				if action.DryRun {
					dryRunStr = " (DRY RUN)"
				}

				fmt.Printf(
					"      %s %s: %s%s\n",
					icon,
					action.ResourceID,
					action.Message,
					dryRunStr,
				)
			}

			if len(actions) > displayCount {
				fmt.Printf("      ... and %d more\n", len(actions)-displayCount)
			}
		}
	}

	// Overall status
	if result.Success {
		fmt.Println("\n🎉 Execution completed successfully!")
	} else {
		fmt.Println("\n⚠️  Execution completed with errors")
	}
}

func runScan() {
	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized!")
		return
	}

	fmt.Println("🔍 Policy Scanner - See what your policies would do!")
	fmt.Println("═══════════════════════════════════════════════════")

	// Initialize scanner
	scannerConfig := scanner.ScannerConfig{
		AWSRegion:     os.Getenv("AWS_REGION"),
		AWSProfile:    os.Getenv("AWS_PROFILE"),
		DryRunDefault: true,
		MaxResources:  1000,
		Timeout:       300,
	}

	policyScanner := scanner.NewPolicyScanner(policyStorage, scannerConfig)

	// List available policies
	policies, err := policyStorage.ListPolicies()
	if err != nil {
		fmt.Printf("❌ Failed to list policies: %v\n", err)
		return
	}

	if len(policies) == 0 {
		fmt.Println("📋 No policies found!")
		fmt.Println("💡 Create your first policy with 'make policy'")
		return
	}

	// Show active policies
	var activePolicies []storage.StoredPolicy
	for _, policy := range policies {
		if policy.Status == "active" {
			activePolicies = append(activePolicies, policy)
		}
	}

	if len(activePolicies) == 0 {
		fmt.Println("📋 No active policies found!")
		return
	}

	fmt.Printf("📋 Active Policies (%d available):\n", len(activePolicies))
	for i, policy := range activePolicies {
		fmt.Printf("%d. 🎯 %s (%s)\n", i+1, policy.Name, strings.ToUpper(policy.ResourceType))
	}
	fmt.Printf("%d. 🚀 Scan ALL policies\n", len(activePolicies)+1)

	reader := bufio.NewReader(os.Stdin)
	choice := getChoice(reader, 1, len(activePolicies)+1, "\nChoose policy to scan: ")

	if choice == len(activePolicies)+1 {
		// Scan all policies
		fmt.Println("\n🚀 Scanning ALL active policies...")
		fmt.Println("═══════════════════════════════════════")

		results, err := policyScanner.ScanAllPolicies()
		if err != nil {
			fmt.Printf("❌ Failed to scan policies: %v\n", err)
			return
		}

		for _, result := range results {
			displayScanResult(&result)
		}
	} else {
		// Scan single policy
		selectedPolicy := activePolicies[choice-1]
		fmt.Printf("\n🎯 Scanning policy: %s\n", selectedPolicy.Name)
		fmt.Println("═══════════════════════════════════════")

		result, err := policyScanner.ScanPolicy(selectedPolicy.Name)
		if err != nil {
			fmt.Printf("❌ Failed to scan policy: %v\n", err)
			return
		}

		displayScanResult(result)
	}

	fmt.Println("\n💡 Next steps:")
	fmt.Println("   • Review the results above")
	fmt.Println("   • Use 'execute' to run policies for real")
	fmt.Println("   • Modify policies if needed")
}

func displayScanResult(result *scanner.ScanResult) {
	fmt.Printf("\n📊 Scan Results for: %s\n", result.PolicyName)
	fmt.Printf("🎯 Resource Type: %s\n", strings.ToUpper(result.ResourceType))
	fmt.Printf("⏰ Scan Time: %s\n", result.ScanTime.Format("2006-01-02 15:04:05"))

	if result.DryRun {
		fmt.Println("🧪 DRY RUN MODE - This is just a preview")
	}

	// Summary
	fmt.Println("\n📈 Summary:")
	fmt.Printf("   • Total Scanned: %d\n", result.Summary.TotalScanned)
	fmt.Printf("   • Matched Resources: %d\n", result.Summary.MatchedResources)
	fmt.Printf("   • Actions Planned: %d\n", result.Summary.ActionsPlanned)

	if result.Summary.HighRiskActions > 0 {
		fmt.Printf("   • ⚠️  High Risk Actions: %d\n", result.Summary.HighRiskActions)
	}

	if result.Summary.CostSavings > 0 {
		fmt.Printf("   • 💰 Estimated Savings: $%.2f/month\n", result.Summary.CostSavings)
	}

	// Show matched resources
	if len(result.MatchedResources) > 0 {
		fmt.Println("\n🎯 Matched Resources:")

		displayCount := len(result.MatchedResources)
		if displayCount > 10 {
			displayCount = 10
		}

		for i, resource := range result.MatchedResources[:displayCount] {
			fmt.Printf("\n%d. 📋 %s (%s)\n", i+1, resource.Name, resource.ID)
			fmt.Printf("   🏷️  Type: %s | State: %s | Region: %s\n",
				resource.Type, resource.State, resource.Region)

			if resource.RiskLevel != "low" {
				fmt.Printf("   ⚠️  Risk Level: %s\n", strings.ToUpper(resource.RiskLevel))
			}

			if !resource.Compliance.Compliant {
				fmt.Printf("   ❌ Compliance Issues: %s\n",
					strings.Join(resource.Compliance.Issues, ", "))
			}

			if len(resource.Actions) > 0 {
				fmt.Printf("   ⚡ Planned Actions: ")
				var actionNames []string
				for _, action := range resource.Actions {
					actionName := action.Type
					if action.DryRun {
						actionName += " (dry-run)"
					}
					actionNames = append(actionNames, actionName)
				}
				fmt.Printf("%s\n", strings.Join(actionNames, ", "))
			}
		}

		if len(result.MatchedResources) > displayCount {
			fmt.Printf("\n... and %d more resources\n", len(result.MatchedResources)-displayCount)
		}
	}

	// Show errors
	if len(result.Errors) > 0 {
		fmt.Println("\n⚠️  Errors:")
		for _, err := range result.Errors {
			fmt.Printf("   • %s\n", err)
		}
	}

	// Cost estimate
	if result.EstimatedCost != nil {
		fmt.Println("\n💰 Cost Impact:")
		fmt.Printf("   • Current Monthly Cost: $%.2f\n", result.EstimatedCost.CurrentMonthlyCost)
		fmt.Printf("   • Projected Savings: $%.2f\n", result.EstimatedCost.ProjectedSavings)
		fmt.Printf("   • Currency: %s\n", result.EstimatedCost.Currency)
	}
}

func executePolicy() {
	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized!")
		return
	}

	fmt.Println("⚡ Policy Executor - Make real changes to AWS resources!")
	fmt.Println("═══════════════════════════════════════════════════")

	// Initialize AWS client in live mode
	awsClient, err := initializeAWSClient(false) // Live mode
	if err != nil {
		fmt.Printf("❌ Failed to initialize AWS client: %v\n", err)
		return
	}
	defer awsClient.Close()

	// List available policies
	policies, err := policyStorage.ListPolicies()
	if err != nil {
		fmt.Printf("❌ Failed to list policies: %v\n", err)
		return
	}

	if len(policies) == 0 {
		fmt.Println("📋 No policies found!")
		fmt.Println("💡 Create your first policy with 'make policy'")
		return
	}

	// Show active policies
	var activePolicies []storage.StoredPolicy
	for _, policy := range policies {
		if policy.Status == "active" {
			activePolicies = append(activePolicies, policy)
		}
	}

	if len(activePolicies) == 0 {
		fmt.Println("📋 No active policies found!")
		return
	}

	fmt.Printf("📋 Active Policies (%d available):\n", len(activePolicies))
	for i, policy := range activePolicies {
		fmt.Printf("%d. 🎯 %s (%s)\n", i+1, policy.Name, strings.ToUpper(policy.ResourceType))
		fmt.Printf("   📝 %s\n", policy.Description)

		// Show dry-run status
		dryRunActions := 0
		for _, action := range policy.Actions {
			if action.DryRun {
				dryRunActions++
			}
		}

		if dryRunActions == len(policy.Actions) {
			fmt.Printf("   🧪 All actions in DRY-RUN mode (safe)\n")
		} else if dryRunActions > 0 {
			fmt.Printf("   ⚠️  %d actions in DRY-RUN, %d in LIVE mode\n", dryRunActions, len(policy.Actions)-dryRunActions)
		} else {
			fmt.Printf("   💥 All actions in LIVE mode - WILL MAKE CHANGES!\n")
		}
		fmt.Println()
	}

	reader := bufio.NewReader(os.Stdin)
	choice := getChoice(reader, 1, len(activePolicies), "\nChoose policy to execute: ")

	selectedPolicy := activePolicies[choice-1]

	// Show policy details and confirm
	fmt.Printf("\n🎯 Selected Policy: %s\n", selectedPolicy.Name)
	fmt.Printf("📝 Description: %s\n", selectedPolicy.Description)
	fmt.Printf("🏷️  Resource Type: %s\n", strings.ToUpper(selectedPolicy.ResourceType))
	fmt.Printf(
		"🔍 Filters: %d | ⚡ Actions: %d\n",
		len(selectedPolicy.Filters),
		len(selectedPolicy.Actions),
	)

	// Count dry-run vs live actions
	dryRunActions := 0
	liveActions := 0
	for _, action := range selectedPolicy.Actions {
		if action.DryRun {
			dryRunActions++
		} else {
			liveActions++
		}
	}

	if liveActions > 0 {
		fmt.Printf(
			"\n⚠️  WARNING: This policy has %d LIVE actions that will make actual changes!\n",
			liveActions,
		)
		fmt.Println("💡 Consider running 'scan' first to see what would happen")
	}

	if dryRunActions > 0 {
		fmt.Printf("🧪 %d actions are in dry-run mode (safe)\n", dryRunActions)
	}

	// Final confirmation
	fmt.Printf("\n🤔 Are you sure you want to execute this policy? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("❌ Execution cancelled")
		return
	}

	// Create executor and run policy
	executor := NewPolicyExecutor(awsClient, policyStorage)

	fmt.Printf("\n🚀 Executing policy: %s\n", selectedPolicy.Name)
	fmt.Println("═══════════════════════════════════")

	result, err := executor.ExecutePolicy(selectedPolicy.Name)
	if err != nil {
		fmt.Printf("❌ Failed to execute policy: %v\n", err)
		return
	}

	displayExecutionResult(result)

	fmt.Println("\n🎉 Policy execution completed!")
	fmt.Println("💡 Check your AWS console to verify the changes")
}

func generateReport() {
	fmt.Println("📊 Report Generator - Choose your adventure!")
	fmt.Println("═══════════════════════════════════════════════")

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("What kind of report do you want?")
	fmt.Println("1. 📋 Compliance Report (HTML)")
	fmt.Println("2. 💰 Cost Analysis Report (CSV)")
	fmt.Println("3. 📊 Resource Inventory (CSV)")
	fmt.Println("4. 📈 Executive Summary (All formats)")

	choice := getChoice(reader, 1, 4, "Choose report type: ")

	// Initialize AWS client
	awsClient, err := initializeAWSClient(true)
	if err != nil {
		fmt.Printf("❌ Failed to initialize AWS client: %v\n", err)
		return
	}
	defer awsClient.Close()

	// Get resources
	fmt.Println("\n🔍 Gathering AWS resource data...")
	ec2Instances, _ := awsClient.GetEC2Instances(aws.EC2Filter{})
	s3Buckets, _ := awsClient.GetS3Buckets(aws.S3Filter{})

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	switch choice {
	case 1:
		generateComplianceReportInteractive(ec2Instances, s3Buckets, timestamp)
	case 2:
		generateCostAnalysisReport(ec2Instances, s3Buckets, timestamp)
	case 3:
		generateResourceInventoryReport(ec2Instances, s3Buckets, timestamp)
	case 4:
		generateExecutiveSummaryReport(ec2Instances, s3Buckets, timestamp)
	}
}

func generateComplianceReportInteractive(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
	timestamp string,
) {
	fmt.Println("\n📋 Generating Compliance Report...")

	htmlGen := reports.NewHTMLReportGenerator("./reports")
	report, err := htmlGen.GenerateComplianceReport(ec2Instances, s3Buckets)
	if err != nil {
		fmt.Printf("❌ Failed to generate report: %v\n", err)
		return
	}

	filename := fmt.Sprintf("compliance_report_%s.html", timestamp)
	if err := htmlGen.SaveHTMLReport(report, filename); err != nil {
		fmt.Printf("❌ Failed to save report: %v\n", err)
		return
	}

	fmt.Printf("✅ HTML compliance report saved: ./reports/%s\n", filename)
	fmt.Println("\n📊 Quick Summary:")
	fmt.Printf("   • Critical Issues: %d\n", report.Summary.CriticalIssues)
	fmt.Printf("   • Compliance Score: %.0f%%\n", report.Summary.CompliancePercentage)
	fmt.Printf("   • Security Score: %d/100\n", report.SecurityScore)
	fmt.Printf("   • Potential Savings: $%.0f/month\n", report.CostImpact.PotentialSavings)
	fmt.Printf("   • EC2 Issues: %d\n", len(report.EC2Findings))
	fmt.Printf("   • S3 Issues: %d\n", len(report.S3Findings))
}

func generateExecutiveSummaryReport(
	ec2Instances []aws.EC2Instance,
	s3Buckets []aws.S3Bucket,
	timestamp string,
) {
	fmt.Println("\n📈 Generating Executive Summary (All Formats)...")

	// Generate HTML compliance report
	htmlGen := reports.NewHTMLReportGenerator("./reports")
	complianceReport, err := htmlGen.GenerateComplianceReport(ec2Instances, s3Buckets)
	if err == nil {
		htmlFilename := fmt.Sprintf("executive_summary_%s.html", timestamp)
		htmlGen.SaveHTMLReport(complianceReport, htmlFilename)
		fmt.Printf("✅ HTML report saved: ./reports/%s\n", htmlFilename)
	}

	// Generate JSON summary
	jsonGen := reports.NewJSONReportGenerator("./reports")
	jsonReport, err := jsonGen.GenerateComplianceReportJSON(ec2Instances, s3Buckets)
	if err == nil {
		jsonFilename := fmt.Sprintf("executive_summary_%s.json", timestamp)
		jsonGen.SaveJSONReport(jsonReport, jsonFilename)
		fmt.Printf("✅ JSON report saved: ./reports/%s\n", jsonFilename)
	}

	// Generate CSV summaries
	csvGen := reports.NewCSVReportGenerator("./reports")

	// Compliance summary
	complianceCsvFilename := fmt.Sprintf("compliance_summary_%s.csv", timestamp)
	csvGen.GenerateComplianceSummaryReport(ec2Instances, s3Buckets, complianceCsvFilename)
	fmt.Printf("✅ Compliance CSV saved: ./reports/%s\n", complianceCsvFilename)

	// Cost analysis
	costCsvFilename := fmt.Sprintf("cost_analysis_%s.csv", timestamp)
	csvGen.GenerateCostAnalysisReport(ec2Instances, s3Buckets, costCsvFilename)
	fmt.Printf("✅ Cost analysis CSV saved: ./reports/%s\n", costCsvFilename)

	fmt.Println("\n🎉 Executive summary package generated!")
	fmt.Println("📁 All files saved in: ./reports/")
	fmt.Println("💡 Open the HTML file in your browser for the best experience")
}

func configureSettings() {
	fmt.Println("⚙️  Configuration Settings")
	fmt.Println("═══════════════════════════")

	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized!")
		return
	}

	// Show storage info
	if fileStorage, ok := policyStorage.(*storage.FileStorage); ok {
		info, err := fileStorage.GetStorageInfo()
		if err != nil {
			fmt.Printf("❌ Failed to get storage info: %v\n", err)
			return
		}

		fmt.Printf("📁 Storage Type: %s\n", info["storage_type"])
		fmt.Printf("📂 Base Directory: %s\n", info["base_directory"])
		fmt.Printf("📊 Policies Stored: %d\n", info["policies_count"])
		fmt.Printf("💾 Storage Size: %.2f MB\n", info["storage_size_mb"])
		fmt.Printf("📋 Policies Path: %s\n", info["storage_path"])
		fmt.Printf("📜 History Path: %s\n", info["history_path"])

		fmt.Println("\n💡 Tip: You can backup your policies by copying the base directory!")
	}

	fmt.Println("\n⚙️  Available Actions:")
	fmt.Println("   • Export policy: custodian-killer policy export <name> <file>")
	fmt.Println("   • Import policy: custodian-killer policy import <file>")
	fmt.Println("   • Test AWS connection: custodian-killer config test")
	fmt.Println("   • View AWS config: custodian-killer config show")
}

// Additional functions called from main interactive loop
func listPolicies() {
	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized!")
		return
	}

	policies, err := policyStorage.ListPolicies()
	if err != nil {
		fmt.Printf("❌ Failed to list policies: %v\n", err)
		return
	}

	if len(policies) == 0 {
		fmt.Println("📋 No policies found!")
		fmt.Println("💡 Create your first policy with 'make policy'")
		return
	}

	fmt.Printf("📋 Your Policies (%d total):\n", len(policies))
	fmt.Println("═══════════════════════════════════════════════════════════")

	for i, policy := range policies {
		fmt.Printf("\n%d. 🎯 %s\n", i+1, policy.Name)
		fmt.Printf("   📝 %s\n", policy.Description)
		fmt.Printf("   🏷️  Resource: %s | Status: %s | Version: v%d\n",
			strings.ToUpper(policy.ResourceType), policy.Status, policy.Version)
		fmt.Printf("   📅 Created: %s | Updated: %s\n",
			policy.CreatedAt.Format("2006-01-02 15:04"),
			policy.UpdatedAt.Format("2006-01-02 15:04"))

		if policy.Source != "" {
			fmt.Printf("   🔧 Source: %s", policy.Source)
			if policy.TemplateID != "" {
				fmt.Printf(" (template: %s)", policy.TemplateID)
			}
			fmt.Println()
		}

		if policy.LastRun != nil {
			fmt.Printf("   ⚡ Last run: %s | Runs: %d\n",
				policy.LastRun.Format("2006-01-02 15:04"), policy.RunCount)
		}

		// Show filters and actions summary
		fmt.Printf("   🔍 Filters: %d | ⚡ Actions: %d",
			len(policy.Filters), len(policy.Actions))

		// Show if any actions are dry-run
		dryRunCount := 0
		for _, action := range policy.Actions {
			if action.DryRun {
				dryRunCount++
			}
		}
		if dryRunCount > 0 {
			fmt.Printf(" (%d dry-run)", dryRunCount)
		}
		fmt.Println()

		if len(policy.Tags) > 0 {
			fmt.Printf("   🏷️  Tags: ")
			var tagStrs []string
			for key, value := range policy.Tags {
				tagStrs = append(tagStrs, fmt.Sprintf("%s=%s", key, value))
			}
			fmt.Println(strings.Join(tagStrs, ", "))
		}
	}

	fmt.Println("\n💡 Commands:")
	fmt.Println("   • Type 'scan' to test your policies")
	fmt.Println("   • Type 'execute' to run them for real")
	fmt.Println("   • Use 'custodian-killer policy edit <name>' to modify")
}

// Helper functions
func getInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func getChoice(reader *bufio.Reader, min, max int, prompt string) int {
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		choice, err := strconv.Atoi(strings.TrimSpace(input))

		if err != nil || choice < min || choice > max {
			fmt.Printf("Please enter a number between %d and %d.\n", min, max)
			continue
		}

		return choice
	}
}

func confirmSave(reader *bufio.Reader) bool {
	fmt.Print("\n💾 Save this policy? (y/n): ")
	input, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(input)) == "y"
}

func showPolicySummary(policy Policy) {
	fmt.Println("\n📋 Policy Summary:")
	fmt.Println("═══════════════════")
	fmt.Printf("Name: %s\n", policy.Name)
	fmt.Printf("Description: %s\n", policy.Description)
	fmt.Printf("Resource Type: %s\n", strings.ToUpper(policy.ResourceType))

	fmt.Println("\nFilters:")
	for _, filter := range policy.Filters {
		fmt.Printf("  • %s %s %v\n", filter.Type, filter.Op, filter.Value)
	}

	fmt.Println("\nActions:")
	for _, action := range policy.Actions {
		dryRunStatus := ""
		if action.DryRun {
			dryRunStatus = " (DRY RUN)"
		}
		fmt.Printf("  • %s%s\n", action.Type, dryRunStatus)
	}

	fmt.Printf("\nExecution Mode: %s\n", policy.Mode.Type)
	if policy.Mode.Schedule != "" {
		fmt.Printf("Schedule: %s\n", policy.Mode.Schedule)
	}
}

func savePolicy(policy Policy) {
	if policyStorage == nil {
		fmt.Println("❌ Storage not initialized! Policy not saved.")
		return
	}

	// Convert to StoredPolicy
	storedPolicy := storage.StoredPolicy{
		Name:         policy.Name,
		Description:  policy.Description,
		ResourceType: policy.ResourceType,
		Tags:         policy.Tags,
		Metadata:     policy.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    "custodian-killer-user",
		Version:      1,
		Status:       "active",
		Source:       "wizard",
	}

	// Convert filters
	for _, filter := range policy.Filters {
		storedFilter := storage.StoredFilter{
			Type:     filter.Type,
			Key:      filter.Key,
			Value:    filter.Value,
			Op:       filter.Op,
			Required: filter.Required,
			Negate:   filter.Negate,
		}
		storedPolicy.Filters = append(storedPolicy.Filters, storedFilter)
	}

	// Convert actions
	for _, action := range policy.Actions {
		storedAction := storage.StoredAction{
			Type:     action.Type,
			Settings: action.Settings,
			DryRun:   action.DryRun,
		}
		storedPolicy.Actions = append(storedPolicy.Actions, storedAction)
	}

	// Convert mode
	storedPolicy.Mode = storage.StoredPolicyMode{
		Type:     policy.Mode.Type,
		Schedule: policy.Mode.Schedule,
		Settings: policy.Mode.Settings,
	}

	// Save to storage
	if err := policyStorage.SavePolicy(storedPolicy); err != nil {
		fmt.Printf("❌ Failed to save policy: %v\n", err)
		return
	}

	fmt.Printf("✅ Policy '%s' saved successfully!\n", policy.Name)
	fmt.Printf("📁 You can find it in your policies directory\n")
}

// Convert template policy definition to main Policy struct
func convertTemplatePolicyToPolicy(policyDef templates.PolicyDefinition) Policy {
	var policy Policy
	policy.Name = policyDef.Name
	policy.Description = policyDef.Description
	policy.ResourceType = policyDef.ResourceType
	policy.Tags = policyDef.Tags
	policy.Metadata = policyDef.Metadata

	// Convert filters
	for _, filterDef := range policyDef.Filters {
		filter := Filter{
			Type:     filterDef.Type,
			Key:      filterDef.Key,
			Value:    filterDef.Value,
			Op:       filterDef.Op,
			Required: filterDef.Required,
			Negate:   filterDef.Negate,
		}
		policy.Filters = append(policy.Filters, filter)
	}

	// Convert actions
	for _, actionDef := range policyDef.Actions {
		action := Action{
			Type:     actionDef.Type,
			Settings: actionDef.Settings,
			DryRun:   actionDef.DryRun,
		}
		policy.Actions = append(policy.Actions, action)
	}

	// Convert mode
	policy.Mode = PolicyMode{
		Type:     policyDef.Mode.Type,
		Schedule: policyDef.Mode.Schedule,
		Settings: policyDef.Mode.Settings,
	}

	return policy
}
