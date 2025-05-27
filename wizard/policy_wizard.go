// wizard/policy_wizard.go
package wizard

import (
	"bufio"
	"custodian-killer/filters"
	"custodian-killer/storage"
	"fmt"
	"os"
	"strings"
)

// PolicyWizard handles interactive policy creation
type PolicyWizard struct {
	reader  *bufio.Reader
	storage storage.PolicyStorage
}

// NewPolicyWizard creates a new policy wizard
func NewPolicyWizard(storage storage.PolicyStorage) *PolicyWizard {
	return &PolicyWizard{
		reader:  bufio.NewReader(os.Stdin),
		storage: storage,
	}
}

// StartPolicyCreation is the main entry point for policy creation
func (pw *PolicyWizard) StartPolicyCreation() {
	fmt.Println("🎯 Welcome to the Policy Creation Wizard!")
	fmt.Println("Let's build something awesome together. I'll guide you through every step.")
	fmt.Println()

	// Step 1: Choose creation method
	fmt.Println("How would you like to create your policy?")
	fmt.Println("1. 🏗️  Build from scratch (full control)")
	fmt.Println("2. 🚀 Use a template (quick start)")
	fmt.Println("3. 🤖 AI-assisted creation (describe what you want)")
	fmt.Println("4. 🔥 Advanced filter wizard (power users)")

	choice := GetChoice(pw.reader, 1, 4, "Choose option (1-4): ")
	fmt.Println()

	switch choice {
	case 1:
		pw.createPolicyFromScratch()
	case 2:
		pw.createPolicyFromTemplate()
	case 3:
		pw.createPolicyWithAI()
	case 4:
		pw.createPolicyWithAdvancedFilters()
	}
}

// createPolicyFromScratch creates a policy from scratch
func (pw *PolicyWizard) createPolicyFromScratch() {
	fmt.Println("🏗️  Building policy from scratch - You've got full control!")
	fmt.Println()

	var policy Policy

	// Get basic info
	policy.Name = GetInput(pw.reader, "📝 Policy name: ")
	policy.Description = GetInput(pw.reader, "📋 Description: ")

	// Choose resource type
	fmt.Println("\n🎯 What AWS resource type should this policy target?")
	resourceTypes := GetResourceTypes()
	for i, rt := range resourceTypes {
		resourceInfo := SupportedResources[rt]
		fmt.Printf("%d. %s - %s\n", i+1, strings.ToUpper(rt), resourceInfo.Description)
	}

	choice := GetChoice(pw.reader, 1, len(resourceTypes), "Choose resource type: ")
	policy.ResourceType = resourceTypes[choice-1]

	fmt.Printf("\n✅ Great! Working with %s resources.\n", strings.ToUpper(policy.ResourceType))

	// Ask about filter complexity
	fmt.Println("\n🔍 How complex should your filters be?")
	fmt.Println("1. 🎯 Simple filters (basic field matching)")
	fmt.Println("2. 🔥 Advanced filters (complex logic, relationships)")

	filterChoice := GetChoice(pw.reader, 1, 2, "Choose filter complexity: ")

	if filterChoice == 1 {
		// Use simple filters
		policy.Filters = pw.createSimpleFilters(policy.ResourceType)
	} else {
		// Use advanced filter wizard
		advancedFilters := pw.createAdvancedFilters(policy.ResourceType)
		policy.Filters = ConvertAdvancedFiltersToSimple(advancedFilters)
	}

	// Add actions
	policy.Actions = pw.createActions(policy.ResourceType)

	// Set execution mode
	policy.Mode = pw.createPolicyMode()

	// Show summary and save
	ShowPolicySummary(policy)

	if ConfirmSave(pw.reader) {
		SavePolicy(pw.storage, policy)
		fmt.Println("🎉 Policy saved successfully! You can now scan or execute it.")
	}
}

// createPolicyWithAdvancedFilters creates a policy using the advanced filter wizard
func (pw *PolicyWizard) createPolicyWithAdvancedFilters() {
	fmt.Println("🔥 Advanced Policy Creation with Complex Filters!")
	fmt.Println("═══════════════════════════════════════════════════")

	var policy Policy

	// Get basic info
	policy.Name = GetInput(pw.reader, "📝 Policy name: ")
	policy.Description = GetInput(pw.reader, "📋 Description: ")

	// Choose resource type
	fmt.Println("\n🎯 What AWS resource type should this policy target?")
	resourceTypes := GetResourceTypes()
	for i, rt := range resourceTypes {
		resourceInfo := SupportedResources[rt]
		fmt.Printf("%d. %s - %s\n", i+1, strings.ToUpper(rt), resourceInfo.Description)
	}

	choice := GetChoice(pw.reader, 1, len(resourceTypes), "Choose resource type: ")
	policy.ResourceType = resourceTypes[choice-1]

	fmt.Printf(
		"\n✅ Creating advanced filters for %s resources.\n",
		strings.ToUpper(policy.ResourceType),
	)

	// Create advanced filters
	advancedFilters := pw.createAdvancedFilters(policy.ResourceType)
	policy.Filters = ConvertAdvancedFiltersToSimple(advancedFilters)

	// Add actions
	policy.Actions = pw.createActions(policy.ResourceType)

	// Set execution mode
	policy.Mode = pw.createPolicyMode()

	// Show summary and save
	ShowPolicySummary(policy)

	if ConfirmSave(pw.reader) {
		SavePolicy(pw.storage, policy)
		fmt.Println("🎉 Advanced policy saved successfully!")
	}
}

// createAdvancedFilters creates advanced filters using the filter wizard
func (pw *PolicyWizard) createAdvancedFilters(resourceType string) []filters.AdvancedFilter {
	fmt.Println("\n🚀 Launching Advanced Filter Wizard...")

	var allFilters []filters.AdvancedFilter

	for {
		fmt.Printf("\n📋 Current filters: %d\n", len(allFilters))
		if len(allFilters) > 0 {
			for i, filter := range allFilters {
				summary := pw.summarizeFilter(filter)
				fmt.Printf("   %d. %s\n", i+1, summary)
			}
		}

		fmt.Println("\nWhat would you like to do?")
		fmt.Println("1. ➕ Add a new filter")
		if len(allFilters) > 0 {
			fmt.Println("2. ✅ Finish and use these filters")
			fmt.Println("3. 🗑️  Remove a filter")
			fmt.Println("4. 🔧 Edit a filter")
		}

		maxChoice := 1
		if len(allFilters) > 0 {
			maxChoice = 4
		}

		choice := GetChoice(pw.reader, 1, maxChoice, "Choose action: ")

		switch choice {
		case 1:
			// Create new advanced filter
			wizard := NewAdvancedFilterWizard(pw.reader, resourceType)
			newFilter, err := wizard.CreateAdvancedFilter()
			if err != nil {
				fmt.Printf("❌ Error creating filter: %v\n", err)
				continue
			}
			allFilters = append(allFilters, newFilter)

		case 2:
			if len(allFilters) == 0 {
				fmt.Println("❌ Need at least one filter!")
				continue
			}
			return allFilters

		case 3:
			if len(allFilters) == 0 {
				continue
			}
			removeIdx := GetChoice(pw.reader, 1, len(allFilters), "Remove which filter? ") - 1
			allFilters = append(allFilters[:removeIdx], allFilters[removeIdx+1:]...)

		case 4:
			if len(allFilters) == 0 {
				continue
			}
			editIdx := GetChoice(pw.reader, 1, len(allFilters), "Edit which filter? ") - 1

			fmt.Printf("Editing filter: %s\n", pw.summarizeFilter(allFilters[editIdx]))
			fmt.Println("For now, you can recreate it. Advanced editing coming soon!")

			wizard := NewAdvancedFilterWizard(pw.reader, resourceType)
			newFilter, err := wizard.CreateAdvancedFilter()
			if err != nil {
				fmt.Printf("❌ Error editing filter: %v\n", err)
				continue
			}
			allFilters[editIdx] = newFilter
		}
	}
}

// summarizeFilter creates a human-readable summary of a filter
func (pw *PolicyWizard) summarizeFilter(filter filters.AdvancedFilter) string {
	if filter.Field != "" {
		return fmt.Sprintf("%s %s %v", filter.Field, filter.Operator, filter.Value)
	}

	if len(filter.AND) > 0 {
		return fmt.Sprintf("AND with %d conditions", len(filter.AND))
	}

	if len(filter.OR) > 0 {
		return fmt.Sprintf("OR with %d conditions", len(filter.OR))
	}

	if filter.NOT != nil {
		return "NOT condition"
	}

	if filter.Collection != nil {
		return fmt.Sprintf("Collection filter (%s)", filter.Collection.Operation)
	}

	if filter.Relationship != nil {
		return fmt.Sprintf("Relationship filter (%s)", filter.Relationship.Type)
	}

	return "Complex filter"
}

// createSimpleFilters creates traditional simple filters
func (pw *PolicyWizard) createSimpleFilters(resourceType string) []Filter {
	fmt.Println("\n🔍 Let's add some filters to target the right resources...")

	resourceInfo := SupportedResources[resourceType]
	var filters []Filter

	for {
		fmt.Println("\nAvailable filters for", strings.ToUpper(resourceType), ":")
		for i, filterType := range resourceInfo.Filters {
			fmt.Printf("%d. %s\n", i+1, filterType)
		}
		fmt.Printf("%d. ✅ Done adding filters\n", len(resourceInfo.Filters)+1)

		choice := GetChoice(pw.reader, 1, len(resourceInfo.Filters)+1, "Choose filter to add: ")

		if choice == len(resourceInfo.Filters)+1 {
			break
		}

		filterType := resourceInfo.Filters[choice-1]
		filter := pw.createSingleFilter(filterType)
		filters = append(filters, filter)

		fmt.Printf("✅ Added filter: %s\n", filterType)
	}

	return filters
}

// createSingleFilter creates a single filter
func (pw *PolicyWizard) createSingleFilter(filterType string) Filter {
	var filter Filter
	filter.Type = filterType

	switch filterType {
	case "tag":
		filter.Key = GetInput(pw.reader, "Tag key: ")
		filter.Value = GetInput(pw.reader, "Tag value (or leave empty to check for existence): ")
		if filter.Value == "" {
			filter.Op = "exists"
		} else {
			filter.Op = "eq"
		}
	case "instance-state", "state":
		fmt.Println("Common states: running, stopped, terminated, pending")
		filter.Value = GetInput(pw.reader, "State: ")
		filter.Op = "eq"
	case "creation-date", "launch-time":
		fmt.Println("Examples: '30 days ago', '2024-01-01', 'last week'")
		filter.Value = GetInput(pw.reader, "Date/time: ")
		filter.Op = "lt" // older than
	default:
		filter.Value = GetInput(pw.reader, fmt.Sprintf("Value for %s: ", filterType))
		filter.Op = "eq"
	}

	return filter
}

// createActions creates actions for the policy
func (pw *PolicyWizard) createActions(resourceType string) []Action {
	fmt.Println("\n⚡ Now let's define what actions to take on matching resources...")

	resourceInfo := SupportedResources[resourceType]
	var actions []Action

	for {
		fmt.Println("\nAvailable actions for", strings.ToUpper(resourceType), ":")
		for i, actionType := range resourceInfo.Actions {
			fmt.Printf("%d. %s\n", i+1, actionType)
		}
		fmt.Printf("%d. ✅ Done adding actions\n", len(resourceInfo.Actions)+1)

		choice := GetChoice(pw.reader, 1, len(resourceInfo.Actions)+1, "Choose action to add: ")

		if choice == len(resourceInfo.Actions)+1 {
			break
		}

		actionType := resourceInfo.Actions[choice-1]
		action := pw.createSingleAction(actionType)
		actions = append(actions, action)

		fmt.Printf("✅ Added action: %s\n", actionType)
	}

	return actions
}

// createSingleAction creates a single action
func (pw *PolicyWizard) createSingleAction(actionType string) Action {
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

			choice := GetChoice(pw.reader, 1, 2, "Choose mode: ")
			action.DryRun = choice == 1
			break
		}
	}

	// Action-specific settings
	switch actionType {
	case "tag":
		tagKey := GetInput(pw.reader, "Tag key to add: ")
		tagValue := GetInput(pw.reader, "Tag value: ")
		action.Settings["key"] = tagKey
		action.Settings["value"] = tagValue
	case "stop":
		fmt.Println("Should instances be force-stopped if graceful stop fails?")
		force := GetChoice(pw.reader, 1, 2, "1. Graceful only  2. Force if needed: ") == 2
		action.Settings["force"] = force
	}

	return action
}

// createPolicyMode creates the policy execution mode
func (pw *PolicyWizard) createPolicyMode() PolicyMode {
	fmt.Println("\n⏱️  How should this policy run?")
	fmt.Println("1. 🔄 On-demand (run when you tell it to)")
	fmt.Println("2. ⏰ Scheduled (runs automatically)")
	fmt.Println("3. 📡 Event-driven (responds to AWS events)")

	choice := GetChoice(pw.reader, 1, 3, "Choose execution mode: ")

	var mode PolicyMode
	switch choice {
	case 1:
		mode.Type = "pull"
	case 2:
		mode.Type = "periodic"
		mode.Schedule = GetInput(
			pw.reader,
			"Schedule (cron format, e.g., '0 2 * * *' for daily at 2am): ",
		)
	case 3:
		mode.Type = "event"
		// TODO: Add event configuration
	}

	return mode
}

// createPolicyFromTemplate creates a policy from a template
func (pw *PolicyWizard) createPolicyFromTemplate() {
	templateWizard := NewTemplateWizard(pw.reader, pw.storage)
	templateWizard.CreatePolicyFromTemplate()
}

// createPolicyWithAI creates a policy using AI assistance
func (pw *PolicyWizard) createPolicyWithAI() {
	aiWizard := NewAIWizard(pw.reader, pw.storage)
	aiWizard.CreatePolicyWithAI()
}
