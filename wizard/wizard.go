// wizard/wizard.go
package wizard

import (
	"bufio"
	"custodian-killer/filters"
	"custodian-killer/storage"
	"fmt"
	"os"
	"strings"
)

// Wizard is the main orchestrator that coordinates all the specialized wizards
type Wizard struct {
	reader  *bufio.Reader
	storage storage.PolicyStorage
}

// NewWizard creates a new wizard orchestrator
func NewWizard(storage storage.PolicyStorage) *Wizard {
	return &Wizard{
		reader:  bufio.NewReader(os.Stdin),
		storage: storage,
	}
}

// Start is the main entry point for the wizard system
func (w *Wizard) Start() {
	fmt.Println("🎯 Welcome to the Policy Creation Wizard!")
	fmt.Println("Let's build something awesome together. I'll guide you through every step.")
	fmt.Println()

	// Show creation options
	fmt.Println("How would you like to create your policy?")
	fmt.Println("1. 🏗️  Build from scratch (full control)")
	fmt.Println("2. 🚀 Use a template (quick start)")
	fmt.Println("3. 🤖 AI-assisted creation (describe what you want)")
	fmt.Println("4. 🔥 Advanced filter wizard (power users)")

	choice := GetChoice(w.reader, 1, 4, "Choose option (1-4): ")
	fmt.Println()

	// Delegate to appropriate specialized wizard
	switch choice {
	case 1:
		w.buildFromScratch()
	case 2:
		w.useTemplate()
	case 3:
		w.useAI()
	case 4:
		w.useAdvancedFilters()
	}
}

// buildFromScratch delegates to the policy wizard for scratch creation
func (w *Wizard) buildFromScratch() {
	policyWizard := NewPolicyWizard(w.storage)
	policyWizard.StartPolicyCreation()
}

// useTemplate delegates to the template wizard
func (w *Wizard) useTemplate() {
	templateWizard := NewTemplateWizard(w.reader, w.storage)
	templateWizard.CreatePolicyFromTemplate()
}

// useAI delegates to the AI wizard
func (w *Wizard) useAI() {
	aiWizard := NewAIWizard(w.reader, w.storage)
	aiWizard.CreatePolicyWithAI()
}

// useAdvancedFilters delegates to advanced filter creation
func (w *Wizard) useAdvancedFilters() {
	fmt.Println("🔥 Advanced Policy Creation with Complex Filters!")
	fmt.Println("═══════════════════════════════════════════════════")

	var policy Policy

	// Get basic info
	policy.Name = GetInput(w.reader, "📝 Policy name: ")
	policy.Description = GetInput(w.reader, "📋 Description: ")

	// Choose resource type
	fmt.Println("\n🎯 What AWS resource type should this policy target?")
	resourceTypes := GetResourceTypes()
	for i, rt := range resourceTypes {
		resourceInfo := SupportedResources[rt]
		fmt.Printf("%d. %s - %s\n", i+1, strings.ToUpper(rt), resourceInfo.Description)
	}

	choice := GetChoice(w.reader, 1, len(resourceTypes), "Choose resource type: ")
	policy.ResourceType = resourceTypes[choice-1]

	fmt.Printf(
		"\n✅ Creating advanced filters for %s resources.\n",
		strings.ToUpper(policy.ResourceType),
	)

	// Use the advanced filter wizard to create filters
	advancedFilterWizard := NewAdvancedFilterWizard(w.reader, policy.ResourceType)

	var allFilters []filters.AdvancedFilter
	for {
		fmt.Printf("\n📋 Current filters: %d\n", len(allFilters))
		if len(allFilters) > 0 {
			for i, filter := range allFilters {
				summary := summarizeAdvancedFilter(filter)
				fmt.Printf("   %d. %s\n", i+1, summary)
			}
		}

		fmt.Println("\nWhat would you like to do?")
		fmt.Println("1. ➕ Add a new advanced filter")
		if len(allFilters) > 0 {
			fmt.Println("2. ✅ Finish and continue with policy")
			fmt.Println("3. 🗑️  Remove a filter")
		}

		maxChoice := 1
		if len(allFilters) > 0 {
			maxChoice = 3
		}

		filterChoice := GetChoice(w.reader, 1, maxChoice, "Choose action: ")

		switch filterChoice {
		case 1:
			// Create new advanced filter using the wizard
			newFilter, err := advancedFilterWizard.CreateAdvancedFilter()
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
			// Convert advanced filters and continue with policy creation
			policy.Filters = ConvertAdvancedFiltersToSimple(allFilters)
			goto createActions

		case 3:
			if len(allFilters) == 0 {
				continue
			}
			removeIdx := GetChoice(w.reader, 1, len(allFilters), "Remove which filter? ") - 1
			allFilters = append(allFilters[:removeIdx], allFilters[removeIdx+1:]...)
		}
	}

createActions:
	// Create actions using the policy wizard's action creation
	policyWizard := NewPolicyWizard(w.storage)
	policy.Actions = policyWizard.createActions(policy.ResourceType)

	// Set execution mode
	policy.Mode = CreatePolicyMode(w.reader)

	// Show summary and save
	ShowPolicySummary(policy)

	if ConfirmSave(w.reader) {
		SavePolicy(w.storage, policy)
		fmt.Println("🎉 Advanced policy saved successfully!")
	}
}

// summarizeAdvancedFilter creates a simple summary of an advanced filter
func summarizeAdvancedFilter(filter filters.AdvancedFilter) string {
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
