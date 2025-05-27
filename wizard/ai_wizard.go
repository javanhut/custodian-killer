// wizard/ai_wizard.go
package wizard

import (
	"bufio"
	"custodian-killer/storage"
	"fmt"
	"strings"
	"time"
)

// AIWizard handles AI-assisted policy creation
type AIWizard struct {
	reader  *bufio.Reader
	storage storage.PolicyStorage
}

// NewAIWizard creates a new AI wizard
func NewAIWizard(reader *bufio.Reader, storage storage.PolicyStorage) *AIWizard {
	return &AIWizard{
		reader:  reader,
		storage: storage,
	}
}

// CreatePolicyWithAI creates a policy using AI assistance
func (aw *AIWizard) CreatePolicyWithAI() {
	fmt.Println("🤖 AI-Assisted Policy Creation")
	fmt.Println("Just describe what you want in plain English, and I'll build the perfect policy!")
	fmt.Println()

	description := GetInput(aw.reader, "📝 Describe what you want this policy to do: ")

	fmt.Println("\n🧠 Analyzing your request...")
	fmt.Println("💭 Understanding your requirements...")

	// Enhanced AI policy suggestion
	suggestedPolicy, confidence, explanation := SmartPolicySuggestion(description)

	fmt.Printf("\n💡 AI Confidence: %d%%\n", confidence)
	fmt.Printf("🎯 Analysis: %s\n", explanation)
	fmt.Println("\n🤖 Based on your description, I suggest this policy:")

	ShowPolicySummary(suggestedPolicy)

	fmt.Println("\n🎛️  What would you like to do?")
	fmt.Println("1. ✅ Perfect! Save it as-is")
	fmt.Println("2. ✏️  Let me tweak some settings")
	fmt.Println("3. 🔄 Try describing it differently")
	fmt.Println("4. 🎯 Show me more details about this policy")

	choice := GetChoice(aw.reader, 1, 4, "Choose option: ")

	switch choice {
	case 1:
		if ConfirmSave(aw.reader) {
			SavePolicy(aw.storage, suggestedPolicy)
			fmt.Println("🎉 AI-generated policy saved! The machines are learning... 🤖")
		}
	case 2:
		policy := aw.tweakAIPolicy(suggestedPolicy)
		if ConfirmSave(aw.reader) {
			SavePolicy(aw.storage, policy)
			fmt.Println("🎉 Customized AI policy saved!")
		}
	case 3:
		aw.CreatePolicyWithAI() // Try again
	case 4:
		aw.explainPolicyInDetail(suggestedPolicy, explanation)
		// Ask again what to do
		fmt.Println("\nNow that you understand the policy better:")
		aw.CreatePolicyWithAI()
	}
}

// tweakAIPolicy allows users to customize AI-generated policies
func (aw *AIWizard) tweakAIPolicy(policy Policy) Policy {
	fmt.Println("\n🔧 Let's customize your AI-generated policy!")

	for {
		fmt.Println("\nWhat would you like to modify?")
		fmt.Println("1. 📝 Change name or description")
		fmt.Println("2. 🔍 Adjust filters")
		fmt.Println("3. ⚡ Modify actions")
		fmt.Println("4. 🎛️  Change execution mode")
		fmt.Println("5. ✅ Looks good, I'm done tweaking")

		choice := GetChoice(aw.reader, 1, 5, "Choose what to modify: ")

		switch choice {
		case 1:
			fmt.Printf("\nCurrent name: %s\n", policy.Name)
			newName := GetInput(aw.reader, "New name (or press Enter to keep): ")
			if newName != "" {
				policy.Name = newName
			}

			fmt.Printf("\nCurrent description: %s\n", policy.Description)
			newDesc := GetInput(aw.reader, "New description (or press Enter to keep): ")
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

			filterChoice := GetChoice(aw.reader, 1, 3, "Choose: ")
			if filterChoice == 1 {
				newFilter := aw.createSingleFilter("custom")
				policy.Filters = append(policy.Filters, newFilter)
			} else if filterChoice == 2 && len(policy.Filters) > 0 {
				removeIdx := GetChoice(aw.reader, 1, len(policy.Filters), "Remove which filter? ") - 1
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

			actionChoice := GetChoice(aw.reader, 1, 4, "Choose: ")
			switch actionChoice {
			case 1:
				newAction := aw.createSingleAction("custom")
				policy.Actions = append(policy.Actions, newAction)
			case 2:
				if len(policy.Actions) > 0 {
					toggleIdx := GetChoice(
						aw.reader,
						1,
						len(policy.Actions),
						"Toggle dry-run for which action? ",
					) - 1
					policy.Actions[toggleIdx].DryRun = !policy.Actions[toggleIdx].DryRun
				}
			case 3:
				if len(policy.Actions) > 0 {
					removeIdx := GetChoice(
						aw.reader,
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
			policy.Mode = CreatePolicyMode(aw.reader)

		case 5:
			fmt.Println("✅ Customization complete!")
			return policy
		}

		// Show updated summary after each change
		fmt.Println("\n📋 Updated Policy Summary:")
		ShowPolicySummary(policy)
	}
}

// explainPolicyInDetail provides detailed explanation of the AI policy
func (aw *AIWizard) explainPolicyInDetail(policy Policy, aiExplanation string) {
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

// createSingleFilter creates a single filter (reuse from policy wizard)
func (aw *AIWizard) createSingleFilter(filterType string) Filter {
	var filter Filter
	filter.Type = filterType

	switch filterType {
	case "tag":
		filter.Key = GetInput(aw.reader, "Tag key: ")
		filter.Value = GetInput(aw.reader, "Tag value (or leave empty to check for existence): ")
		if filter.Value == "" {
			filter.Op = "exists"
		} else {
			filter.Op = "eq"
		}
	case "instance-state", "state":
		fmt.Println("Common states: running, stopped, terminated, pending")
		filter.Value = GetInput(aw.reader, "State: ")
		filter.Op = "eq"
	case "creation-date", "launch-time":
		fmt.Println("Examples: '30 days ago', '2024-01-01', 'last week'")
		filter.Value = GetInput(aw.reader, "Date/time: ")
		filter.Op = "lt" // older than
	default:
		filter.Value = GetInput(aw.reader, fmt.Sprintf("Value for %s: ", filterType))
		filter.Op = "eq"
	}

	return filter
}

// createSingleAction creates a single action (reuse from policy wizard)
func (aw *AIWizard) createSingleAction(actionType string) Action {
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

			choice := GetChoice(aw.reader, 1, 2, "Choose mode: ")
			action.DryRun = choice == 1
			break
		}
	}

	// Action-specific settings
	switch actionType {
	case "tag":
		tagKey := GetInput(aw.reader, "Tag key to add: ")
		tagValue := GetInput(aw.reader, "Tag value: ")
		action.Settings["key"] = tagKey
		action.Settings["value"] = tagValue
	case "stop":
		fmt.Println("Should instances be force-stopped if graceful stop fails?")
		force := GetChoice(aw.reader, 1, 2, "1. Graceful only  2. Force if needed: ") == 2
		action.Settings["force"] = force
	}

	return action
}

// SmartPolicySuggestion generates an AI policy suggestion with enhanced intelligence
func SmartPolicySuggestion(description string) (Policy, int, string) {
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

// detectResourceType detects the AWS resource type from description
func detectResourceType(description string, words []string) (string, int) {
	// Resource type detection with confidence scoring
	resourceKeywords := map[string][]string{
		"ec2": {
			"ec2", "instance", "instances", "server", "servers", "vm", "virtual machine", "compute",
		},
		"s3":     {"s3", "bucket", "buckets", "storage", "object storage", "files"},
		"rds":    {"rds", "database", "db", "mysql", "postgres", "sql", "aurora"},
		"lambda": {"lambda", "function", "functions", "serverless", "code"},
		"iam": {
			"iam", "user", "users", "role", "roles", "permission", "permissions", "policy", "policies",
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

// detectIntent detects the intent and generates appropriate filters/actions
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
				"unused", "idle", "waste", "cost", "expensive", "cheap", "save", "money", "bill", "optimize",
			},
			confidence:  25,
			explanation: "from cost optimization keywords. ",
		},
		"security": {
			keywords: []string{
				"public", "secure", "security", "private", "encrypt", "encryption", "vulnerable", "exposed",
			},
			confidence:  25,
			explanation: "from security-related keywords. ",
		},
		"compliance": {
			keywords: []string{
				"tag", "tags", "untagged", "missing", "required", "comply", "compliance", "standard", "policy",
			},
			confidence:  25,
			explanation: "from compliance keywords. ",
		},
		"cleanup": {
			keywords: []string{
				"old", "delete", "remove", "clean", "cleanup", "unused", "orphaned", "stale",
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

// generateFiltersAndActions generates appropriate filters and actions
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

// Helper functions
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
