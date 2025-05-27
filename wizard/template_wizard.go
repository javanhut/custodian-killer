// wizard/template_wizard.go
package wizard

import (
	"bufio"
	"custodian-killer/storage"
	"custodian-killer/templates"
	"fmt"
	"strconv"
	"strings"
)

// TemplateWizard handles template-based policy creation
type TemplateWizard struct {
	reader  *bufio.Reader
	storage storage.PolicyStorage
}

// NewTemplateWizard creates a new template wizard
func NewTemplateWizard(reader *bufio.Reader, storage storage.PolicyStorage) *TemplateWizard {
	return &TemplateWizard{
		reader:  reader,
		storage: storage,
	}
}

// CreatePolicyFromTemplate creates a policy from a template
func (tw *TemplateWizard) CreatePolicyFromTemplate() {
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

	categoryChoice := GetChoice(tw.reader, 1, len(categories)+2, "Choose category: ")

	var availableTemplates []templates.PolicyTemplate

	if categoryChoice == len(categories)+1 {
		// Search templates
		query := GetInput(tw.reader, "Search for templates (keywords): ")
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

	templateChoice := GetChoice(tw.reader, 1, len(availableTemplates), "Choose template: ")
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
	variables["policy_name"] = GetInput(tw.reader, "\n📝 Policy name: ")

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

			input := GetInput(tw.reader, prompt)

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
	policy := ConvertTemplatePolicyToPolicy(policyDef)

	// Show summary and save
	ShowPolicySummary(policy)

	fmt.Println("\n🎉 Template customization complete!")
	if ConfirmSave(tw.reader) {
		SavePolicy(tw.storage, policy)
		fmt.Println("🚀 Policy created from template! Ready to use.")
	}
}

// ShowTemplateDetails shows detailed information about a template
func (tw *TemplateWizard) ShowTemplateDetails(template templates.PolicyTemplate) {
	fmt.Printf("\n📋 Template Details: %s\n", template.Name)
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("📝 Description: %s\n", template.Description)
	fmt.Printf("🎯 Resource Type: %s\n", template.ResourceType)
	fmt.Printf("📊 Difficulty: %s\n", template.Difficulty)
	fmt.Printf("⚠️  Impact: %s\n", template.Impact)
	fmt.Printf("👤 Created by: %s\n", template.CreatedBy)

	if len(template.Tags) > 0 {
		fmt.Printf("🏷️  Tags: %s\n", strings.Join(template.Tags, ", "))
	}

	if len(template.Examples) > 0 {
		fmt.Println("\n💡 Examples:")
		for _, example := range template.Examples {
			fmt.Printf("   • %s\n", example)
		}
	}

	if len(template.Variables) > 0 {
		fmt.Println("\n⚙️  Configurable Variables:")
		for _, variable := range template.Variables {
			required := ""
			if variable.Required {
				required = " (required)"
			}
			fmt.Printf("   • %s (%s)%s - %s\n",
				variable.Name, variable.Type, required, variable.Description)

			if variable.DefaultValue != nil {
				fmt.Printf("     Default: %v\n", variable.DefaultValue)
			}

			if len(variable.Options) > 0 {
				fmt.Printf("     Options: %s\n", strings.Join(variable.Options, ", "))
			}
		}
	}
}

// BrowseTemplates provides an interactive template browser
func (tw *TemplateWizard) BrowseTemplates() {
	templateManager := templates.NewTemplateManager()
	allTemplates := templateManager.GetAllTemplates()

	if len(allTemplates) == 0 {
		fmt.Println("❌ No templates available!")
		return
	}

	fmt.Println("🚀 Template Browser")
	fmt.Println("═══════════════════")

	currentPage := 0
	templatesPerPage := 5

	for {
		start := currentPage * templatesPerPage
		end := start + templatesPerPage
		if end > len(allTemplates) {
			end = len(allTemplates)
		}

		fmt.Printf("\nPage %d of %d (showing %d-%d of %d templates)\n",
			currentPage+1,
			(len(allTemplates)+templatesPerPage-1)/templatesPerPage,
			start+1, end, len(allTemplates))

		// Show templates on current page
		for i := start; i < end; i++ {
			template := allTemplates[i]
			fmt.Printf("%d. 🎯 %s (%s)\n", i+1, template.Name, template.Impact)
			fmt.Printf("   📝 %s\n", template.Description)
			fmt.Printf("   🏷️  %s | %s | %s\n",
				template.ResourceType, template.Category, template.Difficulty)
		}

		fmt.Println("\nActions:")
		fmt.Println("1. 👀 View template details")
		fmt.Println("2. 🚀 Use template")
		if currentPage > 0 {
			fmt.Println("3. ⬅️  Previous page")
		}
		if end < len(allTemplates) {
			nextOption := 3
			if currentPage > 0 {
				nextOption = 4
			}
			fmt.Printf("%d. ➡️  Next page\n", nextOption)
		}
		fmt.Println("0. 🚪 Exit browser")

		choice := GetInput(tw.reader, "Choose action: ")

		switch choice {
		case "0":
			return
		case "1":
			templateNum := GetChoice(tw.reader, 1, len(allTemplates), "View which template? ")
			tw.ShowTemplateDetails(allTemplates[templateNum-1])
		case "2":
			templateNum := GetChoice(tw.reader, 1, len(allTemplates), "Use which template? ")
			tw.useSpecificTemplate(allTemplates[templateNum-1])
			return
		case "3":
			if currentPage > 0 {
				currentPage--
			}
		default:
			// Handle next page (4 or 3 depending on previous page availability)
			if (choice == "4" && currentPage > 0) || (choice == "3" && currentPage == 0) {
				if end < len(allTemplates) {
					currentPage++
				}
			}
		}
	}
}

// useSpecificTemplate uses a specific template to create a policy
func (tw *TemplateWizard) useSpecificTemplate(template templates.PolicyTemplate) {
	fmt.Printf("\n🚀 Using template: %s\n", template.Name)

	// Show template info
	tw.ShowTemplateDetails(template)

	fmt.Print("\nProceed with this template? (y/N): ")
	confirm := GetInput(tw.reader, "")

	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		return
	}

	// Continue with template customization (reuse existing logic)
	templateManager := templates.NewTemplateManager()

	// Get policy name
	variables := make(map[string]interface{})
	variables["policy_name"] = GetInput(tw.reader, "\n📝 Policy name: ")

	// Collect template variables (same as in CreatePolicyFromTemplate)
	if len(template.Variables) > 0 {
		fmt.Println("\n⚙️  Template Configuration:")
		for _, variable := range template.Variables {
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

			input := GetInput(tw.reader, prompt)

			// Process input (same logic as CreatePolicyFromTemplate)
			if input == "" && !variable.Required && variable.DefaultValue != nil {
				variables[variable.Name] = variable.DefaultValue
			} else if input != "" {
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
	policyDef, err := templateManager.InstantiateTemplate(template.ID, variables)
	if err != nil {
		fmt.Printf("❌ Error creating policy from template: %v\n", err)
		return
	}

	// Convert and save
	policy := ConvertTemplatePolicyToPolicy(policyDef)
	ShowPolicySummary(policy)

	if ConfirmSave(tw.reader) {
		SavePolicy(tw.storage, policy)
		fmt.Println("🚀 Policy created from template successfully!")
	}
}
