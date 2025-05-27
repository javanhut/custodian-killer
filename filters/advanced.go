// filters/advanced.go
package filters

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// AdvancedFilter represents a sophisticated filtering system
type AdvancedFilter struct {
	// Logical operators
	AND []AdvancedFilter `json:"and,omitempty"`
	OR  []AdvancedFilter `json:"or,omitempty"`
	NOT *AdvancedFilter  `json:"not,omitempty"`

	// Leaf condition
	Field    string      `json:"field,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Value    interface{} `json:"value,omitempty"`

	// Advanced options
	CaseSensitive bool   `json:"case_sensitive,omitempty"`
	TimeZone      string `json:"timezone,omitempty"`

	// Resource relationship filters
	Relationship *RelationshipFilter `json:"relationship,omitempty"`

	// Collection filters
	Collection *CollectionFilter `json:"collection,omitempty"`
}

// RelationshipFilter for filtering based on resource relationships
type RelationshipFilter struct {
	Type         string          `json:"type"`          // attached-to, depends-on, child-of, etc.
	TargetType   string          `json:"target_type"`   // ec2, s3, rds, etc.
	TargetFilter *AdvancedFilter `json:"target_filter"` // Filter for the related resource
	Direction    string          `json:"direction"`     // inbound, outbound, bidirectional
}

// CollectionFilter for filtering collections (arrays/slices)
type CollectionFilter struct {
	Operation string          `json:"operation"` // any, all, count, none
	Filter    *AdvancedFilter `json:"filter"`
	Count     *CountFilter    `json:"count,omitempty"`
}

// CountFilter for count-based filtering
type CountFilter struct {
	Operator string `json:"operator"` // eq, ne, gt, lt, gte, lte
	Value    int    `json:"value"`
}

// FilterEvaluator handles the evaluation of advanced filters
type FilterEvaluator struct {
	resourceCache map[string]interface{}
	timeLocation  *time.Location
}

// NewFilterEvaluator creates a new filter evaluator
func NewFilterEvaluator() *FilterEvaluator {
	return &FilterEvaluator{
		resourceCache: make(map[string]interface{}),
		timeLocation:  time.UTC,
	}
}

// Evaluate evaluates an advanced filter against a resource
func (fe *FilterEvaluator) Evaluate(filter AdvancedFilter, resource interface{}) (bool, error) {
	// Handle logical operators
	if len(filter.AND) > 0 {
		return fe.evaluateAND(filter.AND, resource)
	}

	if len(filter.OR) > 0 {
		return fe.evaluateOR(filter.OR, resource)
	}

	if filter.NOT != nil {
		result, err := fe.Evaluate(*filter.NOT, resource)
		return !result, err
	}

	// Handle relationship filters
	if filter.Relationship != nil {
		return fe.evaluateRelationship(filter, resource)
	}

	// Handle collection filters
	if filter.Collection != nil {
		return fe.evaluateCollection(filter, resource)
	}

	// Handle leaf condition
	return fe.evaluateLeafCondition(filter, resource)
}

// evaluateAND evaluates AND conditions
func (fe *FilterEvaluator) evaluateAND(
	filters []AdvancedFilter,
	resource interface{},
) (bool, error) {
	for _, subFilter := range filters {
		result, err := fe.Evaluate(subFilter, resource)
		if err != nil {
			return false, err
		}
		if !result {
			return false, nil
		}
	}
	return true, nil
}

// evaluateOR evaluates OR conditions
func (fe *FilterEvaluator) evaluateOR(
	filters []AdvancedFilter,
	resource interface{},
) (bool, error) {
	for _, subFilter := range filters {
		result, err := fe.Evaluate(subFilter, resource)
		if err != nil {
			return false, err
		}
		if result {
			return true, nil
		}
	}
	return false, nil
}

// evaluateLeafCondition evaluates a leaf condition
func (fe *FilterEvaluator) evaluateLeafCondition(
	filter AdvancedFilter,
	resource interface{},
) (bool, error) {
	// Extract field value from resource
	fieldValue, err := fe.getFieldValue(resource, filter.Field)
	if err != nil {
		return false, err
	}

	// Handle null/missing values
	if fieldValue == nil {
		return fe.handleNullValue(filter.Operator, filter.Value)
	}

	// Evaluate based on operator
	switch strings.ToLower(filter.Operator) {
	case "eq", "equals", "==":
		return fe.evaluateEquals(fieldValue, filter.Value, filter.CaseSensitive)
	case "ne", "not-equals", "!=":
		result, err := fe.evaluateEquals(fieldValue, filter.Value, filter.CaseSensitive)
		return !result, err
	case "gt", ">":
		return fe.evaluateGreaterThan(fieldValue, filter.Value)
	case "gte", ">=":
		return fe.evaluateGreaterThanOrEqual(fieldValue, filter.Value)
	case "lt", "<":
		return fe.evaluateLessThan(fieldValue, filter.Value)
	case "lte", "<=":
		return fe.evaluateLessThanOrEqual(fieldValue, filter.Value)
	case "in":
		return fe.evaluateIn(fieldValue, filter.Value)
	case "not-in":
		result, err := fe.evaluateIn(fieldValue, filter.Value)
		return !result, err
	case "contains":
		return fe.evaluateContains(fieldValue, filter.Value, filter.CaseSensitive)
	case "not-contains":
		result, err := fe.evaluateContains(fieldValue, filter.Value, filter.CaseSensitive)
		return !result, err
	case "starts-with":
		return fe.evaluateStartsWith(fieldValue, filter.Value, filter.CaseSensitive)
	case "ends-with":
		return fe.evaluateEndsWith(fieldValue, filter.Value, filter.CaseSensitive)
	case "regex", "matches":
		return fe.evaluateRegex(fieldValue, filter.Value)
	case "exists":
		return true, nil // Field exists if we got here
	case "not-exists", "missing":
		return false, nil // Field exists if we got here
	case "empty":
		return fe.evaluateEmpty(fieldValue)
	case "not-empty":
		result, err := fe.evaluateEmpty(fieldValue)
		return !result, err
	case "age-gt":
		return fe.evaluateAgeGreaterThan(fieldValue, filter.Value)
	case "age-lt":
		return fe.evaluateAgeLessThan(fieldValue, filter.Value)
	case "between":
		return fe.evaluateBetween(fieldValue, filter.Value)
	default:
		return false, fmt.Errorf("unsupported operator: %s", filter.Operator)
	}
}

// getFieldValue extracts a field value from a resource using dot notation
func (fe *FilterEvaluator) getFieldValue(
	resource interface{},
	fieldPath string,
) (interface{}, error) {
	if fieldPath == "" {
		return resource, nil
	}

	parts := strings.Split(fieldPath, ".")
	current := resource

	for _, part := range parts {
		// Handle array/slice indexing
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			current = fe.handleArrayAccess(current, part)
			if current == nil {
				return nil, nil
			}
			continue
		}

		// Use reflection to access field
		val := reflect.ValueOf(current)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() != reflect.Struct {
			// Try map access
			if val.Kind() == reflect.Map {
				mapVal := val.MapIndex(reflect.ValueOf(part))
				if !mapVal.IsValid() {
					return nil, nil
				}
				current = mapVal.Interface()
				continue
			}
			return nil, fmt.Errorf("cannot access field %s on non-struct type %T", part, current)
		}

		// Find field (case-insensitive)
		fieldVal := fe.findStructField(val, part)
		if !fieldVal.IsValid() {
			return nil, nil
		}

		current = fieldVal.Interface()
	}

	return current, nil
}

// handleArrayAccess handles array/slice access like "tags[0]" or "security_groups[*]"
func (fe *FilterEvaluator) handleArrayAccess(current interface{}, part string) interface{} {
	openBracket := strings.Index(part, "[")
	closeBracket := strings.Index(part, "]")

	fieldName := part[:openBracket]
	indexStr := part[openBracket+1 : closeBracket]

	// Get the array field
	val := reflect.ValueOf(current)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	fieldVal := fe.findStructField(val, fieldName)
	if !fieldVal.IsValid() {
		return nil
	}

	// Handle different index types
	switch indexStr {
	case "*":
		// Return the entire array for collection operations
		return fieldVal.Interface()
	case "length", "len", "count":
		// Return the length of the array
		return fieldVal.Len()
	default:
		// Parse numeric index
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return nil
		}

		if index < 0 || index >= fieldVal.Len() {
			return nil
		}

		return fieldVal.Index(index).Interface()
	}
}

// findStructField finds a struct field case-insensitively
func (fe *FilterEvaluator) findStructField(val reflect.Value, fieldName string) reflect.Value {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)

		// Check direct name match
		if strings.EqualFold(field.Name, fieldName) {
			return val.Field(i)
		}

		// Check JSON tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			jsonName := strings.Split(jsonTag, ",")[0]
			if strings.EqualFold(jsonName, fieldName) {
				return val.Field(i)
			}
		}
	}

	return reflect.Value{}
}

// Evaluation methods for different operators
func (fe *FilterEvaluator) evaluateEquals(
	fieldValue, filterValue interface{},
	caseSensitive bool,
) (bool, error) {
	// Handle string comparison
	if fStr, ok := fieldValue.(string); ok {
		if vStr, ok := filterValue.(string); ok {
			if caseSensitive {
				return fStr == vStr, nil
			}
			return strings.EqualFold(fStr, vStr), nil
		}
	}

	// Handle numeric comparison
	if fe.isNumeric(fieldValue) && fe.isNumeric(filterValue) {
		fNum, _ := fe.toFloat64(fieldValue)
		vNum, _ := fe.toFloat64(filterValue)
		return fNum == vNum, nil
	}

	// Handle boolean comparison
	if fBool, ok := fieldValue.(bool); ok {
		if vBool, ok := filterValue.(bool); ok {
			return fBool == vBool, nil
		}
	}

	// Handle time comparison
	if fTime, ok := fieldValue.(time.Time); ok {
		if vTime, ok := filterValue.(time.Time); ok {
			return fTime.Equal(vTime), nil
		}
		if vStr, ok := filterValue.(string); ok {
			vTime, err := time.Parse(time.RFC3339, vStr)
			if err != nil {
				return false, err
			}
			return fTime.Equal(vTime), nil
		}
	}

	// Default reflection-based comparison
	return reflect.DeepEqual(fieldValue, filterValue), nil
}

func (fe *FilterEvaluator) evaluateGreaterThan(fieldValue, filterValue interface{}) (bool, error) {
	// Numeric comparison
	if fe.isNumeric(fieldValue) && fe.isNumeric(filterValue) {
		fNum, _ := fe.toFloat64(fieldValue)
		vNum, _ := fe.toFloat64(filterValue)
		return fNum > vNum, nil
	}

	// Time comparison
	if fTime, ok := fieldValue.(time.Time); ok {
		if vTime, ok := filterValue.(time.Time); ok {
			return fTime.After(vTime), nil
		}
		if vStr, ok := filterValue.(string); ok {
			vTime, err := time.Parse(time.RFC3339, vStr)
			if err != nil {
				return false, err
			}
			return fTime.After(vTime), nil
		}
	}

	// String comparison (lexicographic)
	if fStr, ok := fieldValue.(string); ok {
		if vStr, ok := filterValue.(string); ok {
			return fStr > vStr, nil
		}
	}

	return false, fmt.Errorf("cannot compare %T and %T with > operator", fieldValue, filterValue)
}

func (fe *FilterEvaluator) evaluateGreaterThanOrEqual(
	fieldValue, filterValue interface{},
) (bool, error) {
	gt, err := fe.evaluateGreaterThan(fieldValue, filterValue)
	if err != nil {
		return false, err
	}
	if gt {
		return true, nil
	}

	eq, err := fe.evaluateEquals(fieldValue, filterValue, true)
	return eq, err
}

func (fe *FilterEvaluator) evaluateLessThan(fieldValue, filterValue interface{}) (bool, error) {
	gt, err := fe.evaluateGreaterThan(fieldValue, filterValue)
	if err != nil {
		return false, err
	}

	eq, err := fe.evaluateEquals(fieldValue, filterValue, true)
	if err != nil {
		return false, err
	}

	return !gt && !eq, nil
}

func (fe *FilterEvaluator) evaluateLessThanOrEqual(
	fieldValue, filterValue interface{},
) (bool, error) {
	gt, err := fe.evaluateGreaterThan(fieldValue, filterValue)
	return !gt, err
}

func (fe *FilterEvaluator) evaluateIn(fieldValue, filterValue interface{}) (bool, error) {
	// Convert filterValue to slice
	val := reflect.ValueOf(filterValue)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return false, fmt.Errorf("'in' operator requires array/slice value")
	}

	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()
		equal, err := fe.evaluateEquals(fieldValue, item, true)
		if err != nil {
			continue
		}
		if equal {
			return true, nil
		}
	}

	return false, nil
}

func (fe *FilterEvaluator) evaluateContains(
	fieldValue, filterValue interface{},
	caseSensitive bool,
) (bool, error) {
	// String contains
	if fStr, ok := fieldValue.(string); ok {
		if vStr, ok := filterValue.(string); ok {
			if caseSensitive {
				return strings.Contains(fStr, vStr), nil
			}
			return strings.Contains(strings.ToLower(fStr), strings.ToLower(vStr)), nil
		}
	}

	// Array/slice contains
	val := reflect.ValueOf(fieldValue)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Interface()
			equal, err := fe.evaluateEquals(item, filterValue, caseSensitive)
			if err != nil {
				continue
			}
			if equal {
				return true, nil
			}
		}
		return false, nil
	}

	return false, fmt.Errorf("'contains' operator not supported for type %T", fieldValue)
}

func (fe *FilterEvaluator) evaluateStartsWith(
	fieldValue, filterValue interface{},
	caseSensitive bool,
) (bool, error) {
	fStr, ok1 := fieldValue.(string)
	vStr, ok2 := filterValue.(string)

	if !ok1 || !ok2 {
		return false, fmt.Errorf("'starts-with' operator requires string values")
	}

	if caseSensitive {
		return strings.HasPrefix(fStr, vStr), nil
	}
	return strings.HasPrefix(strings.ToLower(fStr), strings.ToLower(vStr)), nil
}

func (fe *FilterEvaluator) evaluateEndsWith(
	fieldValue, filterValue interface{},
	caseSensitive bool,
) (bool, error) {
	fStr, ok1 := fieldValue.(string)
	vStr, ok2 := filterValue.(string)

	if !ok1 || !ok2 {
		return false, fmt.Errorf("'ends-with' operator requires string values")
	}

	if caseSensitive {
		return strings.HasSuffix(fStr, vStr), nil
	}
	return strings.HasSuffix(strings.ToLower(fStr), strings.ToLower(vStr)), nil
}

func (fe *FilterEvaluator) evaluateRegex(fieldValue, filterValue interface{}) (bool, error) {
	fStr, ok1 := fieldValue.(string)
	pattern, ok2 := filterValue.(string)

	if !ok1 || !ok2 {
		return false, fmt.Errorf("'regex' operator requires string values")
	}

	matched, err := regexp.MatchString(pattern, fStr)
	return matched, err
}

func (fe *FilterEvaluator) evaluateEmpty(fieldValue interface{}) (bool, error) {
	if fieldValue == nil {
		return true, nil
	}

	val := reflect.ValueOf(fieldValue)
	switch val.Kind() {
	case reflect.String:
		return val.String() == "", nil
	case reflect.Slice, reflect.Array, reflect.Map:
		return val.Len() == 0, nil
	case reflect.Chan:
		return val.IsNil(), nil
	case reflect.Ptr, reflect.Interface:
		return val.IsNil(), nil
	default:
		return false, nil
	}
}

func (fe *FilterEvaluator) evaluateAgeGreaterThan(
	fieldValue, filterValue interface{},
) (bool, error) {
	timestamp, ok := fieldValue.(time.Time)
	if !ok {
		return false, fmt.Errorf("'age-gt' operator requires time.Time field")
	}

	duration, err := fe.parseDuration(filterValue)
	if err != nil {
		return false, err
	}

	age := time.Since(timestamp)
	return age > duration, nil
}

func (fe *FilterEvaluator) evaluateAgeLessThan(fieldValue, filterValue interface{}) (bool, error) {
	timestamp, ok := fieldValue.(time.Time)
	if !ok {
		return false, fmt.Errorf("'age-lt' operator requires time.Time field")
	}

	duration, err := fe.parseDuration(filterValue)
	if err != nil {
		return false, err
	}

	age := time.Since(timestamp)
	return age < duration, nil
}

func (fe *FilterEvaluator) evaluateBetween(fieldValue, filterValue interface{}) (bool, error) {
	// filterValue should be an array with two elements
	val := reflect.ValueOf(filterValue)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return false, fmt.Errorf("'between' operator requires array value")
	}

	if val.Len() != 2 {
		return false, fmt.Errorf("'between' operator requires exactly 2 values")
	}

	min := val.Index(0).Interface()
	max := val.Index(1).Interface()

	// Check if fieldValue is between min and max
	gte, err := fe.evaluateGreaterThanOrEqual(fieldValue, min)
	if err != nil {
		return false, err
	}

	lte, err := fe.evaluateLessThanOrEqual(fieldValue, max)
	if err != nil {
		return false, err
	}

	return gte && lte, nil
}

// Helper functions
func (fe *FilterEvaluator) handleNullValue(operator string, filterValue interface{}) (bool, error) {
	switch strings.ToLower(operator) {
	case "exists":
		return false, nil
	case "not-exists", "missing":
		return true, nil
	case "eq", "equals":
		return filterValue == nil, nil
	case "ne", "not-equals":
		return filterValue != nil, nil
	default:
		return false, nil
	}
}

func (fe *FilterEvaluator) isNumeric(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		return false
	}
}

func (fe *FilterEvaluator) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func (fe *FilterEvaluator) parseDuration(value interface{}) (time.Duration, error) {
	switch v := value.(type) {
	case string:
		// Try standard duration format first
		if duration, err := time.ParseDuration(v); err == nil {
			return duration, nil
		}

		// Try custom formats like "30 days", "2 weeks", etc.
		return fe.parseCustomDuration(v)
	case int:
		// Assume seconds
		return time.Duration(v) * time.Second, nil
	case float64:
		// Assume seconds
		return time.Duration(v * float64(time.Second)), nil
	default:
		return 0, fmt.Errorf("cannot parse duration from %T", value)
	}
}

func (fe *FilterEvaluator) parseCustomDuration(s string) (time.Duration, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	// Map of time units to their durations
	units := map[string]time.Duration{
		"second":  time.Second,
		"seconds": time.Second,
		"sec":     time.Second,
		"minute":  time.Minute,
		"minutes": time.Minute,
		"min":     time.Minute,
		"hour":    time.Hour,
		"hours":   time.Hour,
		"hr":      time.Hour,
		"day":     24 * time.Hour,
		"days":    24 * time.Hour,
		"week":    7 * 24 * time.Hour,
		"weeks":   7 * 24 * time.Hour,
		"month":   30 * 24 * time.Hour, // Approximate
		"months":  30 * 24 * time.Hour,
		"year":    365 * 24 * time.Hour, // Approximate
		"years":   365 * 24 * time.Hour,
	}

	parts := strings.Fields(s)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %s", parts[0])
	}

	unit, exists := units[parts[1]]
	if !exists {
		return 0, fmt.Errorf("unknown time unit: %s", parts[1])
	}

	return time.Duration(value * float64(unit)), nil
}

// Collection filter evaluation
func (fe *FilterEvaluator) evaluateCollection(
	filter AdvancedFilter,
	resource interface{},
) (bool, error) {
	collection := filter.Collection
	fieldValue, err := fe.getFieldValue(resource, filter.Field)
	if err != nil {
		return false, err
	}

	val := reflect.ValueOf(fieldValue)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return false, fmt.Errorf("collection filter requires array/slice field")
	}

	switch strings.ToLower(collection.Operation) {
	case "any":
		return fe.evaluateCollectionAny(val, collection.Filter)
	case "all":
		return fe.evaluateCollectionAll(val, collection.Filter)
	case "none":
		result, err := fe.evaluateCollectionAny(val, collection.Filter)
		return !result, err
	case "count":
		return fe.evaluateCollectionCount(val, collection.Filter, collection.Count)
	default:
		return false, fmt.Errorf("unsupported collection operation: %s", collection.Operation)
	}
}

func (fe *FilterEvaluator) evaluateCollectionAny(
	collection reflect.Value,
	filter *AdvancedFilter,
) (bool, error) {
	for i := 0; i < collection.Len(); i++ {
		item := collection.Index(i).Interface()
		result, err := fe.Evaluate(*filter, item)
		if err != nil {
			continue
		}
		if result {
			return true, nil
		}
	}
	return false, nil
}

func (fe *FilterEvaluator) evaluateCollectionAll(
	collection reflect.Value,
	filter *AdvancedFilter,
) (bool, error) {
	for i := 0; i < collection.Len(); i++ {
		item := collection.Index(i).Interface()
		result, err := fe.Evaluate(*filter, item)
		if err != nil {
			return false, err
		}
		if !result {
			return false, nil
		}
	}
	return true, nil
}

func (fe *FilterEvaluator) evaluateCollectionCount(
	collection reflect.Value,
	filter *AdvancedFilter,
	countFilter *CountFilter,
) (bool, error) {
	matchCount := 0

	for i := 0; i < collection.Len(); i++ {
		item := collection.Index(i).Interface()
		result, err := fe.Evaluate(*filter, item)
		if err != nil {
			continue
		}
		if result {
			matchCount++
		}
	}

	// Evaluate count condition
	switch strings.ToLower(countFilter.Operator) {
	case "eq", "equals":
		return matchCount == countFilter.Value, nil
	case "ne", "not-equals":
		return matchCount != countFilter.Value, nil
	case "gt":
		return matchCount > countFilter.Value, nil
	case "gte":
		return matchCount >= countFilter.Value, nil
	case "lt":
		return matchCount < countFilter.Value, nil
	case "lte":
		return matchCount <= countFilter.Value, nil
	default:
		return false, fmt.Errorf("unsupported count operator: %s", countFilter.Operator)
	}
}

// Relationship filter evaluation (placeholder - would need AWS client integration)
func (fe *FilterEvaluator) evaluateRelationship(
	filter AdvancedFilter,
	resource interface{},
) (bool, error) {
	// This would require AWS client integration to fetch related resources
	// For now, return true as a placeholder
	return true, fmt.Errorf(
		"relationship filters not yet implemented - requires AWS client integration",
	)
}
