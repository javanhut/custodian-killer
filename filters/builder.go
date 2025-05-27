// filters/builder.go
package filters

import (
	"custodian-killer/resources"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// FilterBuilder helps construct complex filters with validation
type FilterBuilder struct {
	resourceType string
	filter       AdvancedFilter
	errors       []string
}

// NewFilterBuilder creates a new filter builder for a resource type
func NewFilterBuilder(resourceType string) *FilterBuilder {
	return &FilterBuilder{
		resourceType: resourceType,
		filter:       AdvancedFilter{},
		errors:       make([]string, 0),
	}
}

// Field starts building a field-based filter
func (fb *FilterBuilder) Field(fieldPath string) *FieldFilterBuilder {
	return &FieldFilterBuilder{
		parent:        fb,
		fieldPath:     fieldPath,
		filterBuilder: fb,
	}
}

// AND combines filters with AND logic
func (fb *FilterBuilder) AND(filters ...AdvancedFilter) *FilterBuilder {
	fb.filter.AND = append(fb.filter.AND, filters...)
	return fb
}

// OR combines filters with OR logic
func (fb *FilterBuilder) OR(filters ...AdvancedFilter) *FilterBuilder {
	fb.filter.OR = append(fb.filter.OR, filters...)
	return fb
}

// NOT negates a filter
func (fb *FilterBuilder) NOT(filter AdvancedFilter) *FilterBuilder {
	fb.filter.NOT = &filter
	return fb
}

// Collection creates a collection filter
func (fb *FilterBuilder) Collection(fieldPath string) *CollectionFilterBuilder {
	return &CollectionFilterBuilder{
		parent:    fb,
		fieldPath: fieldPath,
	}
}

// Relationship creates a relationship filter
func (fb *FilterBuilder) Relationship(
	relationshipType, targetType string,
) *RelationshipFilterBuilder {
	return &RelationshipFilterBuilder{
		parent:           fb,
		relationshipType: relationshipType,
		targetType:       targetType,
	}
}

// Build constructs and validates the final filter
func (fb *FilterBuilder) Build() (AdvancedFilter, error) {
	if len(fb.errors) > 0 {
		return AdvancedFilter{}, fmt.Errorf(
			"filter validation errors: %s",
			strings.Join(fb.errors, "; "),
		)
	}
	return fb.filter, nil
}

// Validate checks the filter against resource definitions
func (fb *FilterBuilder) Validate() error {
	fb.errors = make([]string, 0)
	fb.validateFilter(fb.filter, fb.resourceType)

	if len(fb.errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(fb.errors, "; "))
	}
	return nil
}

func (fb *FilterBuilder) validateFilter(filter AdvancedFilter, resourceType string) {
	// Validate AND conditions
	for _, andFilter := range filter.AND {
		fb.validateFilter(andFilter, resourceType)
	}

	// Validate OR conditions
	for _, orFilter := range filter.OR {
		fb.validateFilter(orFilter, resourceType)
	}

	// Validate NOT condition
	if filter.NOT != nil {
		fb.validateFilter(*filter.NOT, resourceType)
	}

	// Validate leaf condition
	if filter.Field != "" {
		fb.validateLeafFilter(filter, resourceType)
	}
}

func (fb *FilterBuilder) validateLeafFilter(filter AdvancedFilter, resourceType string) {
	// Check if resource type exists
	resourceDef, exists := resources.GetResourceDefinition(resourceType)
	if !exists {
		fb.errors = append(fb.errors, fmt.Sprintf("unknown resource type: %s", resourceType))
		return
	}

	// Parse field path
	fieldParts := strings.Split(filter.Field, ".")
	currentField := fieldParts[0]

	// Check if field exists
	fieldDef, fieldExists := resourceDef.Fields[currentField]
	if !fieldExists {
		fb.errors = append(
			fb.errors,
			fmt.Sprintf("unknown field '%s' for resource type '%s'", currentField, resourceType),
		)
		return
	}

	// Validate operator for field type
	if !fb.isValidOperator(fieldDef, filter.Operator) {
		fb.errors = append(
			fb.errors,
			fmt.Sprintf(
				"operator '%s' not supported for field '%s' of type '%s'",
				filter.Operator,
				filter.Field,
				fieldDef.Type,
			),
		)
	}

	// Validate value type compatibility
	if !fb.isValidValue(fieldDef, filter.Value) {
		fb.errors = append(
			fb.errors,
			fmt.Sprintf(
				"value type incompatible with field '%s' of type '%s'",
				filter.Field,
				fieldDef.Type,
			),
		)
	}
}

func (fb *FilterBuilder) isValidOperator(fieldDef resources.FieldDefinition, operator string) bool {
	for _, validOp := range fieldDef.Operators {
		if validOp == operator {
			return true
		}
	}
	return false
}

func (fb *FilterBuilder) isValidValue(fieldDef resources.FieldDefinition, value interface{}) bool {
	if value == nil {
		return true // Null values are generally acceptable
	}

	switch fieldDef.Type {
	case "string":
		_, ok := value.(string)
		return ok
	case "int":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return true
		}
		return false
	case "float":
		switch value.(type) {
		case float32, float64, int, int8, int16, int32, int64:
			return true
		}
		return false
	case "bool":
		_, ok := value.(bool)
		return ok
	case "time":
		switch value.(type) {
		case time.Time, string:
			return true
		}
		return false
	case "array":
		// Arrays can contain various types
		return true
	case "map":
		// Maps can contain various types
		return true
	default:
		return true
	}
}

// FieldFilterBuilder for building field-specific filters
type FieldFilterBuilder struct {
	parent        *FilterBuilder
	fieldPath     string
	filterBuilder *FilterBuilder
}

// Equals creates an equality filter
func (ffb *FieldFilterBuilder) Equals(value interface{}) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "eq",
		Value:    value,
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// NotEquals creates a not-equals filter
func (ffb *FieldFilterBuilder) NotEquals(value interface{}) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "ne",
		Value:    value,
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// GreaterThan creates a greater-than filter
func (ffb *FieldFilterBuilder) GreaterThan(value interface{}) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "gt",
		Value:    value,
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// LessThan creates a less-than filter
func (ffb *FieldFilterBuilder) LessThan(value interface{}) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "lt",
		Value:    value,
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// Between creates a between filter
func (ffb *FieldFilterBuilder) Between(min, max interface{}) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "between",
		Value:    []interface{}{min, max},
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// In creates an "in" filter
func (ffb *FieldFilterBuilder) In(values ...interface{}) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "in",
		Value:    values,
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// Contains creates a contains filter
func (ffb *FieldFilterBuilder) Contains(value interface{}) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "contains",
		Value:    value,
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// StartsWith creates a starts-with filter
func (ffb *FieldFilterBuilder) StartsWith(value string) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "starts-with",
		Value:    value,
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// EndsWith creates an ends-with filter
func (ffb *FieldFilterBuilder) EndsWith(value string) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "ends-with",
		Value:    value,
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// Regex creates a regex filter
func (ffb *FieldFilterBuilder) Regex(pattern string) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "regex",
		Value:    pattern,
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// Exists creates an exists filter
func (ffb *FieldFilterBuilder) Exists() *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "exists",
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// NotExists creates a not-exists filter
func (ffb *FieldFilterBuilder) NotExists() *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "not-exists",
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// Empty creates an empty filter
func (ffb *FieldFilterBuilder) Empty() *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "empty",
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// AgeGreaterThan creates an age-based filter
func (ffb *FieldFilterBuilder) AgeGreaterThan(duration string) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "age-gt",
		Value:    duration,
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// AgeLessThan creates an age-based filter
func (ffb *FieldFilterBuilder) AgeLessThan(duration string) *FilterBuilder {
	filter := AdvancedFilter{
		Field:    ffb.fieldPath,
		Operator: "age-lt",
		Value:    duration,
	}
	ffb.parent.filter = filter
	return ffb.parent
}

// CollectionFilterBuilder for building collection filters
type CollectionFilterBuilder struct {
	parent    *FilterBuilder
	fieldPath string
}

// Any creates a collection filter that matches if any item matches
func (cfb *CollectionFilterBuilder) Any(itemFilter AdvancedFilter) *FilterBuilder {
	filter := AdvancedFilter{
		Field: cfb.fieldPath,
		Collection: &CollectionFilter{
			Operation: "any",
			Filter:    &itemFilter,
		},
	}
	cfb.parent.filter = filter
	return cfb.parent
}

// All creates a collection filter that matches if all items match
func (cfb *CollectionFilterBuilder) All(itemFilter AdvancedFilter) *FilterBuilder {
	filter := AdvancedFilter{
		Field: cfb.fieldPath,
		Collection: &CollectionFilter{
			Operation: "all",
			Filter:    &itemFilter,
		},
	}
	cfb.parent.filter = filter
	return cfb.parent
}

// None creates a collection filter that matches if no items match
func (cfb *CollectionFilterBuilder) None(itemFilter AdvancedFilter) *FilterBuilder {
	filter := AdvancedFilter{
		Field: cfb.fieldPath,
		Collection: &CollectionFilter{
			Operation: "none",
			Filter:    &itemFilter,
		},
	}
	cfb.parent.filter = filter
	return cfb.parent
}

// Count creates a collection filter based on count
func (cfb *CollectionFilterBuilder) Count(operator string, value int) *FilterBuilder {
	filter := AdvancedFilter{
		Field: cfb.fieldPath,
		Collection: &CollectionFilter{
			Operation: "count",
			Count: &CountFilter{
				Operator: operator,
				Value:    value,
			},
		},
	}
	cfb.parent.filter = filter
	return cfb.parent
}

// RelationshipFilterBuilder for building relationship filters
type RelationshipFilterBuilder struct {
	parent           *FilterBuilder
	relationshipType string
	targetType       string
}

// Where creates a relationship filter with target conditions
func (rfb *RelationshipFilterBuilder) Where(targetFilter AdvancedFilter) *FilterBuilder {
	filter := AdvancedFilter{
		Relationship: &RelationshipFilter{
			Type:         rfb.relationshipType,
			TargetType:   rfb.targetType,
			TargetFilter: &targetFilter,
		},
	}
	rfb.parent.filter = filter
	return rfb.parent
}

// PrebuiltFilters contains common filter patterns
type PrebuiltFilters struct{}

// NewPrebuiltFilters creates a new prebuilt filters instance
func NewPrebuiltFilters() *PrebuiltFilters {
	return &PrebuiltFilters{}
}

// UnusedEC2Instances creates a filter for unused EC2 instances
func (pf *PrebuiltFilters) UnusedEC2Instances(cpuThreshold float64, minDays int) AdvancedFilter {
	return AdvancedFilter{
		AND: []AdvancedFilter{
			{Field: "state", Operator: "eq", Value: "running"},
			{Field: "cpu_utilization", Operator: "lt", Value: cpuThreshold},
			{Field: "running_days", Operator: "gte", Value: minDays},
		},
	}
}

// PublicS3Buckets creates a filter for publicly accessible S3 buckets
func (pf *PrebuiltFilters) PublicS3Buckets() AdvancedFilter {
	return AdvancedFilter{
		OR: []AdvancedFilter{
			{Field: "public_read_acl", Operator: "eq", Value: true},
			{Field: "public_write_acl", Operator: "eq", Value: true},
			{Field: "public_read_policy", Operator: "eq", Value: true},
			{Field: "public_write_policy", Operator: "eq", Value: true},
		},
	}
}

// UnencryptedResources creates a filter for unencrypted resources
func (pf *PrebuiltFilters) UnencryptedResources() AdvancedFilter {
	return AdvancedFilter{
		OR: []AdvancedFilter{
			{Field: "encrypted", Operator: "eq", Value: false},
			{Field: "encryption.enabled", Operator: "eq", Value: false},
			{Field: "storage_encrypted", Operator: "eq", Value: false},
		},
	}
}

// MissingRequiredTags creates a filter for resources missing required tags
func (pf *PrebuiltFilters) MissingRequiredTags(requiredTags ...string) AdvancedFilter {
	var orConditions []AdvancedFilter

	for _, tag := range requiredTags {
		orConditions = append(orConditions, AdvancedFilter{
			Field:    fmt.Sprintf("tags.%s", tag),
			Operator: "not-exists",
		})
	}

	return AdvancedFilter{OR: orConditions}
}

// HighCostResources creates a filter for expensive resources
func (pf *PrebuiltFilters) HighCostResources(threshold float64) AdvancedFilter {
	return AdvancedFilter{
		Field:    "monthly_cost",
		Operator: "gt",
		Value:    threshold,
	}
}

// OldResources creates a filter for old resources
func (pf *PrebuiltFilters) OldResources(ageInDays int) AdvancedFilter {
	return AdvancedFilter{
		Field:    "age_days",
		Operator: "gt",
		Value:    ageInDays,
	}
}

// UnusedForDays creates a filter for resources unused for specified days
func (pf *PrebuiltFilters) UnusedForDays(days int) AdvancedFilter {
	return AdvancedFilter{
		OR: []AdvancedFilter{
			{Field: "days_since_last_used", Operator: "gt", Value: days},
			{Field: "days_since_last_invocation", Operator: "gt", Value: days},
			{Field: "unused_days", Operator: "gt", Value: days},
		},
	}
}

// DevelopmentResources creates a filter for development resources
func (pf *PrebuiltFilters) DevelopmentResources() AdvancedFilter {
	return AdvancedFilter{
		OR: []AdvancedFilter{
			{
				Field:    "tags.Environment",
				Operator: "in",
				Value:    []string{"dev", "development", "test", "testing", "staging"},
			},
			{Field: "name", Operator: "regex", Value: "(?i)(dev|test|stage|sandbox)"},
		},
	}
}

// ProductionResources creates a filter for production resources
func (pf *PrebuiltFilters) ProductionResources() AdvancedFilter {
	return AdvancedFilter{
		OR: []AdvancedFilter{
			{
				Field:    "tags.Environment",
				Operator: "in",
				Value:    []string{"prod", "production", "live"},
			},
			{Field: "name", Operator: "regex", Value: "(?i)(prod|production)"},
		},
	}
}

// FilterExamples contains example filters for different use cases
type FilterExamples struct{}

// GetExampleFilters returns example filters for a resource type
func GetExampleFilters(resourceType string) map[string]AdvancedFilter {
	switch resourceType {
	case "ec2":
		return getEC2Examples()
	case "s3":
		return getS3Examples()
	case "rds":
		return getRDSExamples()
	case "lambda":
		return getLambdaExamples()
	case "ebs":
		return getEBSExamples()
	case "iam-role":
		return getIAMRoleExamples()
	default:
		return make(map[string]AdvancedFilter)
	}
}

func getEC2Examples() map[string]AdvancedFilter {
	return map[string]AdvancedFilter{
		"Unused instances (CPU < 5% for 7+ days)": {
			AND: []AdvancedFilter{
				{Field: "state", Operator: "eq", Value: "running"},
				{Field: "cpu_utilization", Operator: "lt", Value: 5.0},
				{Field: "running_days", Operator: "gte", Value: 7},
			},
		},
		"Expensive instances (>$100/month)": {
			Field:    "monthly_cost",
			Operator: "gt",
			Value:    100.0,
		},
		"Development instances": {
			OR: []AdvancedFilter{
				{
					Field:    "tags.Environment",
					Operator: "in",
					Value:    []string{"dev", "test", "staging"},
				},
				{Field: "name", Operator: "regex", Value: "(?i)(dev|test|stage)"},
			},
		},
		"Instances missing Owner tag": {
			Field:    "tags.Owner",
			Operator: "not-exists",
		},
		"Old instances (>90 days)": {
			Field:    "running_days",
			Operator: "gt",
			Value:    90,
		},
		"Small instances with high cost": {
			AND: []AdvancedFilter{
				{
					Field:    "instance_type",
					Operator: "in",
					Value:    []string{"t3.nano", "t3.micro", "t3.small"},
				},
				{Field: "monthly_cost", Operator: "gt", Value: 50.0},
			},
		},
		"Instances with public IP in private VPC": {
			AND: []AdvancedFilter{
				{Field: "public_ip", Operator: "exists"},
				{Field: "vpc_id", Operator: "starts-with", Value: "vpc-"},
				{Field: "tags.Environment", Operator: "eq", Value: "prod"},
			},
		},
		"Windows instances older than 1 year": {
			AND: []AdvancedFilter{
				{Field: "platform", Operator: "eq", Value: "windows"},
				{Field: "launch_time", Operator: "age-gt", Value: "365 days"},
			},
		},
	}
}

func getS3Examples() map[string]AdvancedFilter {
	return map[string]AdvancedFilter{
		"Public buckets": {
			OR: []AdvancedFilter{
				{Field: "public_read_acl", Operator: "eq", Value: true},
				{Field: "public_write_acl", Operator: "eq", Value: true},
			},
		},
		"Unencrypted buckets": {
			Field:    "encrypted",
			Operator: "eq",
			Value:    false,
		},
		"Large buckets (>100GB)": {
			Field:    "size_gb",
			Operator: "gt",
			Value:    100.0,
		},
		"Buckets without versioning": {
			Field:    "versioning",
			Operator: "eq",
			Value:    "Disabled",
		},
		"Old empty buckets": {
			AND: []AdvancedFilter{
				{Field: "object_count", Operator: "eq", Value: 0},
				{Field: "age_days", Operator: "gt", Value: 30},
			},
		},
		"High-cost buckets": {
			Field:    "monthly_cost",
			Operator: "gt",
			Value:    100.0,
		},
		"Buckets with low security score": {
			Field:    "security_score",
			Operator: "lt",
			Value:    50,
		},
		"Buckets missing data classification": {
			Field:    "tags.DataClassification",
			Operator: "not-exists",
		},
	}
}

func getRDSExamples() map[string]AdvancedFilter {
	return map[string]AdvancedFilter{
		"Publicly accessible databases": {
			Field:    "publicly_accessible",
			Operator: "eq",
			Value:    true,
		},
		"Unencrypted databases": {
			Field:    "storage_encrypted",
			Operator: "eq",
			Value:    false,
		},
		"Databases without backup": {
			Field:    "backup_retention_period",
			Operator: "eq",
			Value:    0,
		},
		"Single-AZ production databases": {
			AND: []AdvancedFilter{
				{Field: "multi_az", Operator: "eq", Value: false},
				{Field: "tags.Environment", Operator: "eq", Value: "prod"},
			},
		},
		"Underutilized databases": {
			AND: []AdvancedFilter{
				{Field: "cpu_utilization", Operator: "lt", Value: 20.0},
				{Field: "database_connections", Operator: "lt", Value: 10},
				{Field: "age_days", Operator: "gt", Value: 7},
			},
		},
		"Expensive databases": {
			Field:    "monthly_cost",
			Operator: "gt",
			Value:    200.0,
		},
		"Old databases": {
			Field:    "age_days",
			Operator: "gt",
			Value:    365,
		},
	}
}

func getLambdaExamples() map[string]AdvancedFilter {
	return map[string]AdvancedFilter{
		"Unused functions (no invocations in 30 days)": {
			Field:    "days_since_last_invocation",
			Operator: "gt",
			Value:    30,
		},
		"High error rate functions": {
			Field:    "error_rate",
			Operator: "gt",
			Value:    10.0,
		},
		"Over-provisioned functions": {
			AND: []AdvancedFilter{
				{Field: "memory_size", Operator: "gt", Value: 1024},
				{Field: "duration_avg", Operator: "lt", Value: 1000},
			},
		},
		"Expensive functions": {
			Field:    "monthly_cost",
			Operator: "gt",
			Value:    50.0,
		},
		"Functions with VPC config": {
			Field:    "vpc_config",
			Operator: "eq",
			Value:    true,
		},
		"Old runtime functions": {
			Field:    "runtime",
			Operator: "in",
			Value:    []string{"python3.7", "nodejs12.x", "java8"},
		},
		"Functions with many environment variables": {
			Field:    "env_var_count",
			Operator: "gt",
			Value:    20,
		},
	}
}

func getEBSExamples() map[string]AdvancedFilter {
	return map[string]AdvancedFilter{
		"Unattached volumes": {
			Field:    "state",
			Operator: "eq",
			Value:    "available",
		},
		"Old unattached volumes": {
			AND: []AdvancedFilter{
				{Field: "state", Operator: "eq", Value: "available"},
				{Field: "unused_days", Operator: "gt", Value: 7},
			},
		},
		"Unencrypted volumes": {
			Field:    "encrypted",
			Operator: "eq",
			Value:    false,
		},
		"Large volumes": {
			Field:    "size",
			Operator: "gt",
			Value:    1000,
		},
		"Underutilized volumes": {
			AND: []AdvancedFilter{
				{Field: "state", Operator: "eq", Value: "in-use"},
				{Field: "utilization_percentage", Operator: "lt", Value: 10.0},
			},
		},
		"Expensive volumes": {
			Field:    "monthly_cost",
			Operator: "gt",
			Value:    100.0,
		},
		"GP2 volumes that could be GP3": {
			AND: []AdvancedFilter{
				{Field: "volume_type", Operator: "eq", Value: "gp2"},
				{Field: "size", Operator: "gt", Value: 100},
			},
		},
	}
}

func getIAMRoleExamples() map[string]AdvancedFilter {
	return map[string]AdvancedFilter{
		"Unused roles (never used)": {
			Field:    "never_used",
			Operator: "eq",
			Value:    true,
		},
		"Roles not used in 90 days": {
			Field:    "days_since_last_used",
			Operator: "gt",
			Value:    90,
		},
		"Roles with admin access": {
			Field:    "has_admin_access",
			Operator: "eq",
			Value:    true,
		},
		"Roles allowing cross-account access": {
			Field:    "allows_cross_account",
			Operator: "eq",
			Value:    true,
		},
		"Roles with many policies": {
			Field:    "attached_policy_count",
			Operator: "gt",
			Value:    10,
		},
		"Old unused roles": {
			AND: []AdvancedFilter{
				{Field: "age_days", Operator: "gt", Value: 90},
				{Field: "days_since_last_used", Operator: "gt", Value: 30},
			},
		},
		"Service roles for EC2": {
			Field:    "trusted_services",
			Operator: "contains",
			Value:    "ec2.amazonaws.com",
		},
	}
}

// FilterValidator provides validation utilities
type FilterValidator struct{}

// ValidateFilter validates a filter against resource definitions
func (fv *FilterValidator) ValidateFilter(filter AdvancedFilter, resourceType string) []string {
	var errors []string

	// Get resource definition
	resourceDef, exists := resources.GetResourceDefinition(resourceType)
	if !exists {
		return []string{fmt.Sprintf("unknown resource type: %s", resourceType)}
	}

	errors = append(errors, fv.validateFilterRecursive(filter, resourceDef)...)
	return errors
}

func (fv *FilterValidator) validateFilterRecursive(
	filter AdvancedFilter,
	resourceDef resources.ResourceDefinition,
) []string {
	var errors []string

	// Validate AND conditions
	for _, andFilter := range filter.AND {
		errors = append(errors, fv.validateFilterRecursive(andFilter, resourceDef)...)
	}

	// Validate OR conditions
	for _, orFilter := range filter.OR {
		errors = append(errors, fv.validateFilterRecursive(orFilter, resourceDef)...)
	}

	// Validate NOT condition
	if filter.NOT != nil {
		errors = append(errors, fv.validateFilterRecursive(*filter.NOT, resourceDef)...)
	}

	// Validate leaf condition
	if filter.Field != "" {
		errors = append(errors, fv.validateLeafCondition(filter, resourceDef)...)
	}

	return errors
}

func (fv *FilterValidator) validateLeafCondition(
	filter AdvancedFilter,
	resourceDef resources.ResourceDefinition,
) []string {
	var errors []string

	// Parse field path (handle nested fields like "tags.Environment")
	fieldParts := strings.Split(filter.Field, ".")
	baseField := fieldParts[0]

	// Check if base field exists
	fieldDef, exists := resourceDef.Fields[baseField]
	if !exists {
		errors = append(
			errors,
			fmt.Sprintf(
				"field '%s' does not exist for resource type '%s'",
				baseField,
				resourceDef.Name,
			),
		)
		return errors
	}

	// Validate operator
	validOperator := false
	for _, op := range fieldDef.Operators {
		if op == filter.Operator {
			validOperator = true
			break
		}
	}
	if !validOperator {
		errors = append(
			errors,
			fmt.Sprintf("operator '%s' is not valid for field '%s'", filter.Operator, filter.Field),
		)
	}

	// Validate value type
	if err := fv.validateValueType(filter.Value, fieldDef, filter.Operator); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate enum values
	if len(fieldDef.EnumValues) > 0 && filter.Operator == "eq" {
		if strValue, ok := filter.Value.(string); ok {
			validEnum := false
			for _, enumVal := range fieldDef.EnumValues {
				if enumVal == strValue {
					validEnum = true
					break
				}
			}
			if !validEnum {
				errors = append(
					errors,
					fmt.Sprintf(
						"value '%s' is not a valid enum value for field '%s'. Valid values: %v",
						strValue,
						filter.Field,
						fieldDef.EnumValues,
					),
				)
			}
		}
	}

	return errors
}

func (fv *FilterValidator) validateValueType(
	value interface{},
	fieldDef resources.FieldDefinition,
	operator string,
) error {
	if value == nil {
		// Null values are acceptable for existence checks
		if operator == "exists" || operator == "not-exists" {
			return nil
		}
		return fmt.Errorf("null value not allowed for operator '%s'", operator)
	}

	// Special handling for array operators
	if operator == "in" || operator == "not-in" {
		// Value should be an array
		switch reflect.TypeOf(value).Kind() {
		case reflect.Slice, reflect.Array:
			return nil
		default:
			return fmt.Errorf("operator '%s' requires array value", operator)
		}
	}

	if operator == "between" {
		// Value should be an array with exactly 2 elements
		val := reflect.ValueOf(value)
		if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
			return fmt.Errorf("operator 'between' requires array value")
		}
		if val.Len() != 2 {
			return fmt.Errorf("operator 'between' requires exactly 2 values")
		}
		return nil
	}

	// Type-specific validation
	switch fieldDef.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' expects string value", fieldDef.Type)
		}
	case "int":
		if !fv.isIntegerType(value) {
			return fmt.Errorf("field '%s' expects integer value", fieldDef.Type)
		}
	case "float":
		if !fv.isNumericType(value) {
			return fmt.Errorf("field '%s' expects numeric value", fieldDef.Type)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' expects boolean value", fieldDef.Type)
		}
	case "time":
		switch value.(type) {
		case time.Time, string:
			// Both time.Time and string representations are acceptable
		default:
			return fmt.Errorf("field '%s' expects time value (time.Time or string)", fieldDef.Type)
		}
	}

	return nil
}

func (fv *FilterValidator) isIntegerType(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}

func (fv *FilterValidator) isNumericType(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		return false
	}
}

// FilterHelper provides utility functions for common filter operations
type FilterHelper struct{}

// CombineWithAND combines multiple filters with AND logic
func (fh *FilterHelper) CombineWithAND(filters ...AdvancedFilter) AdvancedFilter {
	if len(filters) == 0 {
		return AdvancedFilter{}
	}
	if len(filters) == 1 {
		return filters[0]
	}

	return AdvancedFilter{AND: filters}
}

// CombineWithOR combines multiple filters with OR logic
func (fh *FilterHelper) CombineWithOR(filters ...AdvancedFilter) AdvancedFilter {
	if len(filters) == 0 {
		return AdvancedFilter{}
	}
	if len(filters) == 1 {
		return filters[0]
	}

	return AdvancedFilter{OR: filters}
}

// Negate creates a NOT filter
func (fh *FilterHelper) Negate(filter AdvancedFilter) AdvancedFilter {
	return AdvancedFilter{NOT: &filter}
}

// CreateTagFilter creates a filter for tag existence or value
func (fh *FilterHelper) CreateTagFilter(tagKey string, tagValue interface{}) AdvancedFilter {
	fieldPath := fmt.Sprintf("tags.%s", tagKey)

	if tagValue == nil {
		// Check for tag existence
		return AdvancedFilter{
			Field:    fieldPath,
			Operator: "exists",
		}
	}

	// Check for specific tag value
	return AdvancedFilter{
		Field:    fieldPath,
		Operator: "eq",
		Value:    tagValue,
	}
}

// CreateCostFilter creates a filter based on cost thresholds
func (fh *FilterHelper) CreateCostFilter(operator string, threshold float64) AdvancedFilter {
	return AdvancedFilter{
		Field:    "monthly_cost",
		Operator: operator,
		Value:    threshold,
	}
}

// CreateAgeFilter creates a filter based on resource age
func (fh *FilterHelper) CreateAgeFilter(operator string, ageInDays int) AdvancedFilter {
	return AdvancedFilter{
		Field:    "age_days",
		Operator: operator,
		Value:    ageInDays,
	}
}

// CreateUtilizationFilter creates a filter based on resource utilization
func (fh *FilterHelper) CreateUtilizationFilter(
	utilizationField string,
	operator string,
	threshold float64,
) AdvancedFilter {
	return AdvancedFilter{
		Field:    utilizationField,
		Operator: operator,
		Value:    threshold,
	}
}

// FilterUsageExamples demonstrates how to use the advanced filtering system
func FilterUsageExamples() {
	fmt.Println("=== Advanced Filter Usage Examples ===")

	// Example 1: Simple field filter
	fmt.Println("1. Simple field filter - EC2 instances in running state:")
	filter1 := NewFilterBuilder("ec2").
		Field("state").Equals("running")
	result1, _ := filter1.Build()
	fmt.Printf("   Filter: %+v\n\n", result1)

	// Example 2: Complex AND filter
	fmt.Println("2. Complex AND filter - Unused EC2 instances:")
	filter2 := NewFilterBuilder("ec2").AND(
		AdvancedFilter{Field: "state", Operator: "eq", Value: "running"},
		AdvancedFilter{Field: "cpu_utilization", Operator: "lt", Value: 5.0},
		AdvancedFilter{Field: "running_days", Operator: "gte", Value: 7},
	)
	result2, _ := filter2.Build()
	fmt.Printf("   Filter: %+v\n\n", result2)

	// Example 3: OR filter with tag matching
	fmt.Println("3. OR filter - Development or test instances:")
	filter3 := NewFilterBuilder("ec2").OR(
		AdvancedFilter{
			Field:    "tags.Environment",
			Operator: "in",
			Value:    []string{"dev", "test", "staging"},
		},
		AdvancedFilter{Field: "name", Operator: "regex", Value: "(?i)(dev|test|stage)"},
	)
	result3, _ := filter3.Build()
	fmt.Printf("   Filter: %+v\n\n", result3)

	// Example 4: Collection filter
	fmt.Println("4. Collection filter - Instances with any security group containing 'web':")
	filter4 := NewFilterBuilder("ec2").
		Collection("security_groups").Any(
		AdvancedFilter{Field: ".", Operator: "contains", Value: "web"},
	)
	result4, _ := filter4.Build()
	fmt.Printf("   Filter: %+v\n\n", result4)

	// Example 5: Age-based filter
	fmt.Println("5. Age-based filter - Resources older than 90 days:")
	filter5 := NewFilterBuilder("ec2").
		Field("launch_time").AgeGreaterThan("90 days")
	result5, _ := filter5.Build()
	fmt.Printf("   Filter: %+v\n\n", result5)

	// Example 6: Between filter for cost range
	fmt.Println("6. Between filter - Resources costing between $50-$200/month:")
	filter6 := NewFilterBuilder("ec2").
		Field("monthly_cost").Between(50.0, 200.0)
	result6, _ := filter6.Build()
	fmt.Printf("   Filter: %+v\n\n", result6)

	// Example 7: NOT filter
	fmt.Println("7. NOT filter - Non-production instances:")
	productionFilter := AdvancedFilter{Field: "tags.Environment", Operator: "eq", Value: "prod"}
	filter7 := NewFilterBuilder("ec2").NOT(productionFilter)
	result7, _ := filter7.Build()
	fmt.Printf("   Filter: %+v\n\n", result7)

	// Example 8: Using prebuilt filters
	fmt.Println("8. Using prebuilt filters:")
	prebuilt := NewPrebuiltFilters()
	unusedEC2 := prebuilt.UnusedEC2Instances(5.0, 7)
	publicS3 := prebuilt.PublicS3Buckets()
	missingTags := prebuilt.MissingRequiredTags("Environment", "Owner")

	fmt.Printf("   Unused EC2: %+v\n", unusedEC2)
	fmt.Printf("   Public S3: %+v\n", publicS3)
	fmt.Printf("   Missing tags: %+v\n\n", missingTags)

	// Example 9: Filter validation
	fmt.Println("9. Filter validation:")
	validator := &FilterValidator{}

	// Valid filter
	validFilter := AdvancedFilter{Field: "state", Operator: "eq", Value: "running"}
	validErrors := validator.ValidateFilter(validFilter, "ec2")
	fmt.Printf("   Valid filter errors: %v\n", validErrors)

	// Invalid filter
	invalidFilter := AdvancedFilter{Field: "nonexistent_field", Operator: "eq", Value: "test"}
	invalidErrors := validator.ValidateFilter(invalidFilter, "ec2")
	fmt.Printf("   Invalid filter errors: %v\n\n", invalidErrors)

	// Example 10: Getting example filters
	fmt.Println("10. Example filters for EC2:")
	ec2Examples := GetExampleFilters("ec2")
	for name, filter := range ec2Examples {
		fmt.Printf("   %s: %+v\n", name, filter)
	}
}
