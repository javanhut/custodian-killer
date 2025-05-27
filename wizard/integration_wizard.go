// wizard/integration_wizard.go
package wizard

import (
	"bufio"
	"custodian-killer/filters"
	"custodian-killer/resources"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// AdvancedFilterWizard handles the interactive creation of advanced filters
type AdvancedFilterWizard struct {
	reader       *bufio.Reader
	resourceType string
	validator    *filters.FilterValidator
}

// NewAdvancedFilterWizard creates a new advanced filter wizard
func NewAdvancedFilterWizard(reader *bufio.Reader, resourceType string) *AdvancedFilterWizard {
	return &AdvancedFilterWizard{
		reader:       reader,
		resourceType: resourceType,
		validator:    &filters.FilterValidator{},
	}
}

// CreateAdvancedFilter is the main entry point for advanced filter creation
func (afw *AdvancedFilterWizard) CreateAdvancedFilter() (filters.AdvancedFilter, error) {
	fmt.Println("\n🔥 Advanced Filter Creation - The Power Unleashed!")
	fmt.Println("═══════════════════════════════════════════════════════════")

	fmt.Printf("📊 Creating filter for: %s\n", strings.ToUpper(afw.resourceType))

	// Show available creation methods
	fmt.Println("\n🎯 How would you like to create your filter?")
	fmt.Println("1. 🚀 Quick Prebuilt Patterns (recommended)")
	fmt.Println("2. 🎨 Simple Field-Based Filter")
	fmt.Println("3. 🧠 Complex Logic Builder")
	fmt.Println("4. 💡 Browse Examples & Customize")
	fmt.Println("5. 🔧 JSON Expert Mode")

	choice := GetChoice(afw.reader, 1, 5, "Choose creation method: ")

	var filter filters.AdvancedFilter
	var err error

	switch choice {
	case 1:
		filter, err = afw.createPrebuiltFilter()
	case 2:
		filter, err = afw.createSimpleFilter()
	case 3:
		filter, err = afw.createComplexFilter()
	case 4:
		filter, err = afw.browseAndCustomizeExamples()
	case 5:
		filter, err = afw.createJSONFilter()
	}

	if err != nil {
		return filters.AdvancedFilter{}, err
	}

	// Validate the filter
	if err := afw.validateFilter(filter); err != nil {
		fmt.Printf("⚠️  Filter validation warnings: %v\n", err)
		fmt.Print("Continue anyway? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)

		if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
			return filters.AdvancedFilter{}, fmt.Errorf("filter creation cancelled")
		}
	}

	// Show filter summary
	afw.showFilterSummary(filter)

	return filter, nil
}

// createPrebuiltFilter creates a filter using prebuilt patterns
func (afw *AdvancedFilterWizard) createPrebuiltFilter() (filters.AdvancedFilter, error) {
	fmt.Println("\n🚀 Prebuilt Filter Patterns")
	fmt.Println("These are battle-tested patterns used by AWS experts!")

	prebuilt := filters.NewPrebuiltFilters()

	switch afw.resourceType {
	case "ec2":
		return afw.createEC2PrebuiltFilter(prebuilt)
	case "s3":
		return afw.createS3PrebuiltFilter(prebuilt)
	case "rds":
		return afw.createRDSPrebuiltFilter(prebuilt)
	case "lambda":
		return afw.createLambdaPrebuiltFilter(prebuilt)
	case "ebs":
		return afw.createEBSPrebuiltFilter(prebuilt)
	case "iam-role":
		return afw.createIAMPrebuiltFilter(prebuilt)
	default:
		return afw.createGenericPrebuiltFilter(prebuilt)
	}
}

func (afw *AdvancedFilterWizard) createEC2PrebuiltFilter(
	prebuilt *filters.PrebuiltFilters,
) (filters.AdvancedFilter, error) {
	fmt.Println("\n💻 EC2 Prebuilt Patterns:")
	fmt.Println("1. 💰 Unused Instances (Low CPU, Long Running)")
	fmt.Println("2. 🏷️  Missing Required Tags")
	fmt.Println("3. 💸 High Cost Resources")
	fmt.Println("4. 🕰️  Old Resources")
	fmt.Println("5. 🧪 Development/Test Resources")
	fmt.Println("6. 🏭 Production Resources")

	choice := GetChoice(afw.reader, 1, 6, "Choose EC2 pattern: ")

	switch choice {
	case 1:
		cpuThreshold := afw.getFloatInput("CPU threshold % (default 5.0): ", 5.0)
		minDays := afw.getIntInput("Minimum running days (default 7): ", 7)
		return prebuilt.UnusedEC2Instances(cpuThreshold, minDays), nil

	case 2:
		tags := afw.getTagsInput(
			"Required tags (comma-separated, default: Environment,Owner): ",
			[]string{"Environment", "Owner"},
		)
		return prebuilt.MissingRequiredTags(tags...), nil

	case 3:
		threshold := afw.getFloatInput("Cost threshold $ (default 100.0): ", 100.0)
		return prebuilt.HighCostResources(threshold), nil

	case 4:
		days := afw.getIntInput("Age in days (default 90): ", 90)
		return prebuilt.OldResources(days), nil

	case 5:
		return prebuilt.DevelopmentResources(), nil

	case 6:
		return prebuilt.ProductionResources(), nil
	}

	return filters.AdvancedFilter{}, fmt.Errorf("invalid choice")
}

func (afw *AdvancedFilterWizard) createS3PrebuiltFilter(
	prebuilt *filters.PrebuiltFilters,
) (filters.AdvancedFilter, error) {
	fmt.Println("\n🪣 S3 Prebuilt Patterns:")
	fmt.Println("1. 🌍 Public Buckets (Security Risk)")
	fmt.Println("2. 🔓 Unencrypted Buckets")
	fmt.Println("3. 🏷️  Missing Required Tags")
	fmt.Println("4. 💸 High Cost Buckets")
	fmt.Println("5. 🕰️  Old Buckets")

	choice := GetChoice(afw.reader, 1, 5, "Choose S3 pattern: ")

	switch choice {
	case 1:
		return prebuilt.PublicS3Buckets(), nil
	case 2:
		return prebuilt.UnencryptedResources(), nil
	case 3:
		tags := afw.getTagsInput(
			"Required tags (default: Environment,Owner): ",
			[]string{"Environment", "Owner"},
		)
		return prebuilt.MissingRequiredTags(tags...), nil
	case 4:
		threshold := afw.getFloatInput("Cost threshold $ (default 100.0): ", 100.0)
		return prebuilt.HighCostResources(threshold), nil
	case 5:
		days := afw.getIntInput("Age in days (default 90): ", 90)
		return prebuilt.OldResources(days), nil
	}

	return filters.AdvancedFilter{}, fmt.Errorf("invalid choice")
}

func (afw *AdvancedFilterWizard) createRDSPrebuiltFilter(
	prebuilt *filters.PrebuiltFilters,
) (filters.AdvancedFilter, error) {
	fmt.Println("\n🗄️  RDS Prebuilt Patterns:")
	fmt.Println("1. 🔓 Unencrypted Databases")
	fmt.Println("2. 🏷️  Missing Required Tags")
	fmt.Println("3. 💸 High Cost Databases")
	fmt.Println("4. 🕰️  Old Databases")
	fmt.Println("5. 💤 Unused Databases")

	choice := GetChoice(afw.reader, 1, 5, "Choose RDS pattern: ")

	switch choice {
	case 1:
		return prebuilt.UnencryptedResources(), nil
	case 2:
		tags := afw.getTagsInput(
			"Required tags (default: Environment,Owner): ",
			[]string{"Environment", "Owner"},
		)
		return prebuilt.MissingRequiredTags(tags...), nil
	case 3:
		threshold := afw.getFloatInput("Cost threshold $ (default 200.0): ", 200.0)
		return prebuilt.HighCostResources(threshold), nil
	case 4:
		days := afw.getIntInput("Age in days (default 180): ", 180)
		return prebuilt.OldResources(days), nil
	case 5:
		days := afw.getIntInput("Unused for days (default 30): ", 30)
		return prebuilt.UnusedForDays(days), nil
	}

	return filters.AdvancedFilter{}, fmt.Errorf("invalid choice")
}

func (afw *AdvancedFilterWizard) createLambdaPrebuiltFilter(
	prebuilt *filters.PrebuiltFilters,
) (filters.AdvancedFilter, error) {
	fmt.Println("\n⚡ Lambda Prebuilt Patterns:")
	fmt.Println("1. 💤 Unused Functions")
	fmt.Println("2. 🏷️  Missing Required Tags")
	fmt.Println("3. 💸 High Cost Functions")
	fmt.Println("4. 🕰️  Old Functions")

	choice := GetChoice(afw.reader, 1, 4, "Choose Lambda pattern: ")

	switch choice {
	case 1:
		days := afw.getIntInput("Unused for days (default 30): ", 30)
		return prebuilt.UnusedForDays(days), nil
	case 2:
		tags := afw.getTagsInput(
			"Required tags (default: Environment,Owner): ",
			[]string{"Environment", "Owner"},
		)
		return prebuilt.MissingRequiredTags(tags...), nil
	case 3:
		threshold := afw.getFloatInput("Cost threshold $ (default 50.0): ", 50.0)
		return prebuilt.HighCostResources(threshold), nil
	case 4:
		days := afw.getIntInput("Age in days (default 90): ", 90)
		return prebuilt.OldResources(days), nil
	}

	return filters.AdvancedFilter{}, fmt.Errorf("invalid choice")
}

func (afw *AdvancedFilterWizard) createEBSPrebuiltFilter(
	prebuilt *filters.PrebuiltFilters,
) (filters.AdvancedFilter, error) {
	fmt.Println("\n💾 EBS Prebuilt Patterns:")
	fmt.Println("1. 🔓 Unencrypted Volumes")
	fmt.Println("2. 🏷️  Missing Required Tags")
	fmt.Println("3. 💸 High Cost Volumes")
	fmt.Println("4. 🕰️  Old Volumes")

	choice := GetChoice(afw.reader, 1, 4, "Choose EBS pattern: ")

	switch choice {
	case 1:
		return prebuilt.UnencryptedResources(), nil
	case 2:
		tags := afw.getTagsInput(
			"Required tags (default: Environment,Owner): ",
			[]string{"Environment", "Owner"},
		)
		return prebuilt.MissingRequiredTags(tags...), nil
	case 3:
		threshold := afw.getFloatInput("Cost threshold $ (default 100.0): ", 100.0)
		return prebuilt.HighCostResources(threshold), nil
	case 4:
		days := afw.getIntInput("Age in days (default 90): ", 90)
		return prebuilt.OldResources(days), nil
	}

	return filters.AdvancedFilter{}, fmt.Errorf("invalid choice")
}

func (afw *AdvancedFilterWizard) createIAMPrebuiltFilter(
	prebuilt *filters.PrebuiltFilters,
) (filters.AdvancedFilter, error) {
	fmt.Println("\n🔐 IAM Role Prebuilt Patterns:")
	fmt.Println("1. 💤 Unused Roles")
	fmt.Println("2. 🏷️  Missing Required Tags")
	fmt.Println("3. 🕰️  Old Roles")

	choice := GetChoice(afw.reader, 1, 3, "Choose IAM pattern: ")

	switch choice {
	case 1:
		days := afw.getIntInput("Unused for days (default 90): ", 90)
		return prebuilt.UnusedForDays(days), nil
	case 2:
		tags := afw.getTagsInput(
			"Required tags (default: Environment,Owner): ",
			[]string{"Environment", "Owner"},
		)
		return prebuilt.MissingRequiredTags(tags...), nil
	case 3:
		days := afw.getIntInput("Age in days (default 180): ", 180)
		return prebuilt.OldResources(days), nil
	}

	return filters.AdvancedFilter{}, fmt.Errorf("invalid choice")
}

func (afw *AdvancedFilterWizard) createGenericPrebuiltFilter(
	prebuilt *filters.PrebuiltFilters,
) (filters.AdvancedFilter, error) {
	fmt.Println("\n🔧 Generic Prebuilt Patterns:")
	fmt.Println("1. 🏷️  Missing Required Tags")
	fmt.Println("2. 💸 High Cost Resources")
	fmt.Println("3. 🕰️  Old Resources")

	choice := GetChoice(afw.reader, 1, 3, "Choose generic pattern: ")

	switch choice {
	case 1:
		tags := afw.getTagsInput(
			"Required tags (default: Environment,Owner): ",
			[]string{"Environment", "Owner"},
		)
		return prebuilt.MissingRequiredTags(tags...), nil
	case 2:
		threshold := afw.getFloatInput("Cost threshold $ (default 100.0): ", 100.0)
		return prebuilt.HighCostResources(threshold), nil
	case 3:
		days := afw.getIntInput("Age in days (default 90): ", 90)
		return prebuilt.OldResources(days), nil
	}

	return filters.AdvancedFilter{}, fmt.Errorf("invalid choice")
}

// createSimpleFilter creates a simple field-based filter
func (afw *AdvancedFilterWizard) createSimpleFilter() (filters.AdvancedFilter, error) {
	fmt.Println("\n🎨 Simple Field-Based Filter")
	fmt.Println("Perfect for straightforward conditions!")

	resourceDef, exists := resources.GetResourceDefinition(afw.resourceType)
	if !exists {
		return filters.AdvancedFilter{}, fmt.Errorf("unknown resource type: %s", afw.resourceType)
	}

	// Show available fields grouped by category
	afw.showFieldsGrouped(resourceDef)

	fieldName := afw.selectField(resourceDef)
	fieldDef := resourceDef.Fields[fieldName]

	operator := afw.selectOperator(fieldDef)
	value := afw.getValue(fieldDef, operator)

	filter := filters.AdvancedFilter{
		Field:    fieldName,
		Operator: operator,
		Value:    value,
	}

	return filter, nil
}

// createComplexFilter creates a complex filter with multiple conditions
func (afw *AdvancedFilterWizard) createComplexFilter() (filters.AdvancedFilter, error) {
	fmt.Println("\n🧠 Complex Logic Builder")
	fmt.Println("Build sophisticated filters with AND/OR/NOT logic!")

	fmt.Println("\n🔧 How would you like to combine conditions?")
	fmt.Println("1. 🤝 AND - All conditions must be true")
	fmt.Println("2. 🎯 OR - Any condition can be true")
	fmt.Println("3. 🚫 NOT - Negate a condition")
	fmt.Println("4. 🏗️  Mixed Logic - Combine different operators")

	choice := GetChoice(afw.reader, 1, 4, "Choose logic type: ")

	switch choice {
	case 1:
		return afw.createANDFilter()
	case 2:
		return afw.createORFilter()
	case 3:
		return afw.createNOTFilter()
	case 4:
		return afw.createMixedLogicFilter()
	}

	return filters.AdvancedFilter{}, fmt.Errorf("invalid choice")
}

func (afw *AdvancedFilterWizard) createANDFilter() (filters.AdvancedFilter, error) {
	fmt.Println("\n🤝 Creating AND Filter (All conditions must match)")

	var conditions []filters.AdvancedFilter

	for {
		fmt.Printf("\n📋 Current conditions: %d\n", len(conditions))
		if len(conditions) > 0 {
			for i, condition := range conditions {
				fmt.Printf(
					"   %d. %s %s %v\n",
					i+1,
					condition.Field,
					condition.Operator,
					condition.Value,
				)
			}
		}

		fmt.Println("\nWhat would you like to do?")
		fmt.Println("1. ➕ Add a condition")
		if len(conditions) > 0 {
			fmt.Println("2. ✅ Finish building filter")
		}
		if len(conditions) > 1 {
			fmt.Println("3. 🗑️  Remove a condition")
		}

		maxChoices := 1
		if len(conditions) > 0 {
			maxChoices = 2
		}
		if len(conditions) > 1 {
			maxChoices = 3
		}

		choice := GetChoice(afw.reader, 1, maxChoices, "Choose action: ")

		switch choice {
		case 1:
			condition, err := afw.createSimpleFilter()
			if err != nil {
				fmt.Printf("❌ Error creating condition: %v\n", err)
				continue
			}
			conditions = append(conditions, condition)

		case 2:
			if len(conditions) == 0 {
				fmt.Println("❌ Need at least one condition!")
				continue
			}
			return filters.AdvancedFilter{AND: conditions}, nil

		case 3:
			if len(conditions) <= 1 {
				continue
			}
			removeIdx := GetChoice(afw.reader, 1, len(conditions), "Remove which condition? ") - 1
			conditions = append(conditions[:removeIdx], conditions[removeIdx+1:]...)
		}
	}
}

func (afw *AdvancedFilterWizard) createORFilter() (filters.AdvancedFilter, error) {
	fmt.Println("\n🎯 Creating OR Filter (Any condition can match)")

	var conditions []filters.AdvancedFilter

	for {
		fmt.Printf("\n📋 Current conditions: %d\n", len(conditions))
		if len(conditions) > 0 {
			for i, condition := range conditions {
				fmt.Printf(
					"   %d. %s %s %v\n",
					i+1,
					condition.Field,
					condition.Operator,
					condition.Value,
				)
			}
		}

		fmt.Println("\nWhat would you like to do?")
		fmt.Println("1. ➕ Add a condition")
		if len(conditions) > 0 {
			fmt.Println("2. ✅ Finish building filter")
		}
		if len(conditions) > 1 {
			fmt.Println("3. 🗑️  Remove a condition")
		}

		maxChoices := 1
		if len(conditions) > 0 {
			maxChoices = 2
		}
		if len(conditions) > 1 {
			maxChoices = 3
		}

		choice := GetChoice(afw.reader, 1, maxChoices, "Choose action: ")

		switch choice {
		case 1:
			condition, err := afw.createSimpleFilter()
			if err != nil {
				fmt.Printf("❌ Error creating condition: %v\n", err)
				continue
			}
			conditions = append(conditions, condition)

		case 2:
			if len(conditions) == 0 {
				fmt.Println("❌ Need at least one condition!")
				continue
			}
			return filters.AdvancedFilter{OR: conditions}, nil

		case 3:
			if len(conditions) <= 1 {
				continue
			}
			removeIdx := GetChoice(afw.reader, 1, len(conditions), "Remove which condition? ") - 1
			conditions = append(conditions[:removeIdx], conditions[removeIdx+1:]...)
		}
	}
}

func (afw *AdvancedFilterWizard) createNOTFilter() (filters.AdvancedFilter, error) {
	fmt.Println("\n🚫 Creating NOT Filter (Negate a condition)")

	fmt.Println("First, let's create the condition to negate:")
	condition, err := afw.createSimpleFilter()
	if err != nil {
		return filters.AdvancedFilter{}, err
	}

	return filters.AdvancedFilter{NOT: &condition}, nil
}

func (afw *AdvancedFilterWizard) createMixedLogicFilter() (filters.AdvancedFilter, error) {
	fmt.Println("\n🏗️  Mixed Logic Filter")
	fmt.Println("For now, let's create an AND filter. Advanced mixed logic coming soon!")
	return afw.createANDFilter()
}

// browseAndCustomizeExamples shows examples and allows customization
func (afw *AdvancedFilterWizard) browseAndCustomizeExamples() (filters.AdvancedFilter, error) {
	fmt.Println("\n💡 Browse Examples & Customize")
	fmt.Println("Learn from real-world examples!")

	examples := filters.GetExampleFilters(afw.resourceType)
	if len(examples) == 0 {
		fmt.Println("❌ No examples available for this resource type")
		return afw.createSimpleFilter()
	}

	fmt.Printf("\n📚 Available examples for %s:\n", strings.ToUpper(afw.resourceType))

	exampleNames := make([]string, 0, len(examples))
	for name := range examples {
		exampleNames = append(exampleNames, name)
	}

	for i, name := range exampleNames {
		fmt.Printf("%d. %s\n", i+1, name)
	}

	choice := GetChoice(afw.reader, 1, len(exampleNames), "Choose example to customize: ")
	selectedName := exampleNames[choice-1]
	selectedFilter := examples[selectedName]

	fmt.Printf("\n🔍 Selected Example: %s\n", selectedName)
	afw.showFilterSummary(selectedFilter)

	fmt.Println("\nWhat would you like to do?")
	fmt.Println("1. ✅ Use as-is")
	fmt.Println("2. 🔧 Customize values")
	fmt.Println("3. 📚 See another example")

	actionChoice := GetChoice(afw.reader, 1, 3, "Choose action: ")

	switch actionChoice {
	case 1:
		return selectedFilter, nil
	case 2:
		return afw.customizeFilter(selectedFilter)
	case 3:
		return afw.browseAndCustomizeExamples()
	}

	return selectedFilter, nil
}

// customizeFilter allows users to customize filter values
func (afw *AdvancedFilterWizard) customizeFilter(
	filter filters.AdvancedFilter,
) (filters.AdvancedFilter, error) {
	fmt.Println("\n🔧 Customizing Filter")
	fmt.Println("You can modify the values in this filter:")

	// For now, we'll handle simple field-based filters
	if filter.Field != "" {
		fmt.Printf("Field: %s\n", filter.Field)
		fmt.Printf("Operator: %s\n", filter.Operator)
		fmt.Printf("Current Value: %v\n", filter.Value)

		fmt.Print("Enter new value (or press Enter to keep current): ")
		newValueStr, _ := afw.reader.ReadString('\n')
		newValueStr = strings.TrimSpace(newValueStr)

		if newValueStr != "" {
			// Parse new value based on current value type
			if filter.Value != nil {
				switch filter.Value.(type) {
				case string:
					filter.Value = newValueStr
				case int:
					if intVal, err := strconv.Atoi(newValueStr); err == nil {
						filter.Value = intVal
					}
				case float64:
					if floatVal, err := strconv.ParseFloat(newValueStr, 64); err == nil {
						filter.Value = floatVal
					}
				case bool:
					if boolVal, err := strconv.ParseBool(newValueStr); err == nil {
						filter.Value = boolVal
					}
				default:
					filter.Value = newValueStr
				}
			} else {
				filter.Value = newValueStr
			}
		}
	}

	return filter, nil
}

// createJSONFilter allows expert users to enter JSON directly
func (afw *AdvancedFilterWizard) createJSONFilter() (filters.AdvancedFilter, error) {
	fmt.Println("\n🔧 JSON Expert Mode")
	fmt.Println("For power users who know exactly what they want!")

	fmt.Println("\n📝 Example JSON filter:")
	example := map[string]interface{}{
		"AND": []map[string]interface{}{
			{"field": "state", "operator": "eq", "value": "running"},
			{"field": "cpu_utilization", "operator": "lt", "value": 5.0},
		},
	}
	exampleJSON, _ := json.MarshalIndent(example, "", "  ")
	fmt.Printf("%s\n", exampleJSON)

	fmt.Println("\n✏️  Enter your JSON filter:")
	fmt.Print("> ")

	jsonStr, _ := afw.reader.ReadString('\n')
	jsonStr = strings.TrimSpace(jsonStr)

	if jsonStr == "" {
		return filters.AdvancedFilter{}, fmt.Errorf("empty JSON input")
	}

	var filter filters.AdvancedFilter
	if err := json.Unmarshal([]byte(jsonStr), &filter); err != nil {
		return filters.AdvancedFilter{}, fmt.Errorf("invalid JSON: %v", err)
	}

	return filter, nil
}

// Helper methods for getting user input
func (afw *AdvancedFilterWizard) getFloatInput(prompt string, defaultValue float64) float64 {
	fmt.Print(prompt)
	input, _ := afw.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}

	if value, err := strconv.ParseFloat(input, 64); err == nil {
		return value
	}

	return defaultValue
}

func (afw *AdvancedFilterWizard) getIntInput(prompt string, defaultValue int) int {
	fmt.Print(prompt)
	input, _ := afw.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}

	if value, err := strconv.Atoi(input); err == nil {
		return value
	}

	return defaultValue
}

func (afw *AdvancedFilterWizard) getTagsInput(prompt string, defaultTags []string) []string {
	fmt.Print(prompt)
	input, _ := afw.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultTags
	}

	tags := strings.Split(input, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	return tags
}

// showFieldsGrouped displays fields organized by category
func (afw *AdvancedFilterWizard) showFieldsGrouped(resourceDef resources.ResourceDefinition) {
	fmt.Printf("\n📊 Available fields for %s:\n", strings.ToUpper(afw.resourceType))

	// Group fields by type/category
	basicFields := make([]string, 0)
	computedFields := make([]string, 0)
	tagFields := make([]string, 0)

	for fieldName, fieldDef := range resourceDef.Fields {
		if strings.Contains(fieldName, "tag") {
			tagFields = append(tagFields, fieldName)
		} else if fieldDef.Computed {
			computedFields = append(computedFields, fieldName)
		} else {
			basicFields = append(basicFields, fieldName)
		}
	}

	if len(basicFields) > 0 {
		fmt.Println("\n🔷 Basic Fields:")
		for _, fieldName := range basicFields {
			fieldDef := resourceDef.Fields[fieldName]
			fmt.Printf("   • %s (%s) - %s\n", fieldName, fieldDef.Type, fieldDef.Description)
		}
	}

	if len(computedFields) > 0 {
		fmt.Println("\n📈 Computed Fields (calculated values):")
		for _, fieldName := range computedFields {
			fieldDef := resourceDef.Fields[fieldName]
			fmt.Printf("   • %s (%s) - %s\n", fieldName, fieldDef.Type, fieldDef.Description)
		}
	}

	if len(tagFields) > 0 {
		fmt.Println("\n🏷️  Tag Fields:")
		for _, fieldName := range tagFields {
			fieldDef := resourceDef.Fields[fieldName]
			fmt.Printf("   • %s (%s) - %s\n", fieldName, fieldDef.Type, fieldDef.Description)
		}
	}
}

// selectField helps user select a field
func (afw *AdvancedFilterWizard) selectField(resourceDef resources.ResourceDefinition) string {
	fieldNames := make([]string, 0, len(resourceDef.Fields))
	for fieldName := range resourceDef.Fields {
		fieldNames = append(fieldNames, fieldName)
	}

	fmt.Printf("\n🎯 Choose a field (1-%d): ", len(fieldNames))
	for i, fieldName := range fieldNames {
		fmt.Printf("\n%d. %s", i+1, fieldName)
	}
	fmt.Println()

	choice := GetChoice(afw.reader, 1, len(fieldNames), "Enter field number: ")
	return fieldNames[choice-1]
}

// selectOperator helps user select an operator
func (afw *AdvancedFilterWizard) selectOperator(fieldDef resources.FieldDefinition) string {
	fmt.Printf("\n🔧 Available operators for %s:\n", fieldDef.Type)
	for i, op := range fieldDef.Operators {
		description := afw.getOperatorDescription(op)
		fmt.Printf("%d. %s - %s\n", i+1, op, description)
	}

	choice := GetChoice(afw.reader, 1, len(fieldDef.Operators), "Choose operator: ")
	return fieldDef.Operators[choice-1]
}

// getValue helps user enter a value
func (afw *AdvancedFilterWizard) getValue(
	fieldDef resources.FieldDefinition,
	operator string,
) interface{} {
	// Special operators that don't need values
	if operator == "exists" || operator == "not-exists" || operator == "empty" ||
		operator == "not-empty" {
		return nil
	}

	fmt.Printf("\n💡 Enter value for %s %s:\n", fieldDef.Type, operator)

	// Show examples if available
	if len(fieldDef.Examples) > 0 {
		fmt.Printf("Examples: %s\n", strings.Join(fieldDef.Examples, ", "))
	}

	// Show enum values if available
	if len(fieldDef.EnumValues) > 0 {
		fmt.Printf("Valid values: %s\n", strings.Join(fieldDef.EnumValues, ", "))
	}

	fmt.Print("Value: ")
	valueStr, _ := afw.reader.ReadString('\n')
	valueStr = strings.TrimSpace(valueStr)

	return afw.parseValueByType(valueStr, fieldDef.Type, operator)
}

// parseValueByType converts string input to appropriate type
func (afw *AdvancedFilterWizard) parseValueByType(
	valueStr, fieldType, operator string,
) interface{} {
	// Handle array operators
	if operator == "in" || operator == "not-in" {
		return strings.Split(valueStr, ",")
	}

	if operator == "between" {
		parts := strings.Split(valueStr, ",")
		if len(parts) == 2 {
			val1 := afw.parseValueByType(strings.TrimSpace(parts[0]), fieldType, "eq")
			val2 := afw.parseValueByType(strings.TrimSpace(parts[1]), fieldType, "eq")
			return []interface{}{val1, val2}
		}
	}

	// Type-specific parsing
	switch fieldType {
	case "int":
		if val, err := strconv.Atoi(valueStr); err == nil {
			return val
		}
	case "float":
		if val, err := strconv.ParseFloat(valueStr, 64); err == nil {
			return val
		}
	case "bool":
		if val, err := strconv.ParseBool(valueStr); err == nil {
			return val
		}
	case "time":
		if val, err := time.Parse(time.RFC3339, valueStr); err == nil {
			return val
		}
		if val, err := time.Parse("2006-01-02", valueStr); err == nil {
			return val
		}
	}

	return valueStr // Default to string
}

// getOperatorDescription returns a human-readable description of an operator
func (afw *AdvancedFilterWizard) getOperatorDescription(operator string) string {
	descriptions := map[string]string{
		"eq":           "equals",
		"ne":           "not equals",
		"gt":           "greater than",
		"gte":          "greater than or equal",
		"lt":           "less than",
		"lte":          "less than or equal",
		"in":           "is one of (comma-separated list)",
		"not-in":       "is not one of (comma-separated list)",
		"contains":     "contains text",
		"not-contains": "does not contain text",
		"starts-with":  "starts with text",
		"ends-with":    "ends with text",
		"regex":        "matches regex pattern",
		"exists":       "field exists",
		"not-exists":   "field does not exist",
		"empty":        "field is empty",
		"not-empty":    "field is not empty",
		"age-gt":       "older than (e.g., '30 days')",
		"age-lt":       "newer than (e.g., '7 days')",
		"between":      "between two values (comma-separated)",
	}

	if desc, exists := descriptions[operator]; exists {
		return desc
	}
	return operator
}

// validateFilter validates the created filter
func (afw *AdvancedFilterWizard) validateFilter(filter filters.AdvancedFilter) error {
	errors := afw.validator.ValidateFilter(filter, afw.resourceType)
	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}
	return nil
}

// showFilterSummary displays a summary of the created filter
func (afw *AdvancedFilterWizard) showFilterSummary(filter filters.AdvancedFilter) {
	fmt.Println("\n📋 Filter Summary:")
	fmt.Println("═══════════════════")

	filterJSON, err := json.MarshalIndent(filter, "", "  ")
	if err != nil {
		fmt.Printf("Filter: %+v\n", filter)
	} else {
		fmt.Printf("Filter: %s\n", filterJSON)
	}

	fmt.Printf("Resource Type: %s\n", strings.ToUpper(afw.resourceType))
	fmt.Printf("Complexity: %s\n", afw.calculateFilterComplexity(filter))
}

// calculateFilterComplexity estimates filter complexity
func (afw *AdvancedFilterWizard) calculateFilterComplexity(filter filters.AdvancedFilter) string {
	complexity := 0

	if len(filter.AND) > 0 {
		complexity += len(filter.AND)
	}
	if len(filter.OR) > 0 {
		complexity += len(filter.OR)
	}
	if filter.NOT != nil {
		complexity += 1
	}
	if filter.Collection != nil {
		complexity += 2
	}
	if filter.Relationship != nil {
		complexity += 3
	}
	if filter.Field != "" {
		complexity += 1
	}

	switch {
	case complexity <= 2:
		return "Simple"
	case complexity <= 5:
		return "Moderate"
	default:
		return "Complex"
	}
}
