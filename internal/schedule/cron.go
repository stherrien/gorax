package schedule

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// CronExpression represents a parsed cron expression with extended fields
type CronExpression struct {
	Second      string `json:"second,omitempty"`
	Minute      string `json:"minute"`
	Hour        string `json:"hour"`
	DayOfMonth  string `json:"day_of_month"`
	Month       string `json:"month"`
	DayOfWeek   string `json:"day_of_week"`
	Timezone    string `json:"timezone"`
	Description string `json:"description,omitempty"`
	Original    string `json:"original"`
}

// CronParser provides enhanced cron expression parsing with support for
// standard and extended cron syntax including L, W, and # characters.
type CronParser struct {
	baseParser cron.Parser
}

// NewCronParser creates a new enhanced cron parser
func NewCronParser() *CronParser {
	// Create parser that supports standard cron format with optional seconds
	parser := cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour |
			cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	return &CronParser{
		baseParser: parser,
	}
}

// ParseCron parses a cron expression and returns a CronExpression struct.
// Supports both 5-field (standard) and 6-field (with seconds) formats.
// Special characters supported: *, -, /, ?, L, W, #
func (p *CronParser) ParseCron(expression string) (*CronExpression, error) {
	if expression == "" {
		return nil, &ValidationError{Message: "cron expression cannot be empty"}
	}

	// Handle descriptors (@yearly, @monthly, etc.)
	if strings.HasPrefix(expression, "@") {
		return p.parseDescriptor(expression)
	}

	fields := strings.Fields(expression)
	if len(fields) < 5 || len(fields) > 6 {
		return nil, &ValidationError{
			Message: fmt.Sprintf("invalid cron expression: expected 5 or 6 fields, got %d", len(fields)),
		}
	}

	cronExpr := &CronExpression{
		Original: expression,
		Timezone: "UTC",
	}

	if len(fields) == 6 {
		// 6-field format: second minute hour day-of-month month day-of-week
		cronExpr.Second = fields[0]
		cronExpr.Minute = fields[1]
		cronExpr.Hour = fields[2]
		cronExpr.DayOfMonth = fields[3]
		cronExpr.Month = fields[4]
		cronExpr.DayOfWeek = fields[5]
	} else {
		// 5-field format: minute hour day-of-month month day-of-week
		cronExpr.Second = "0"
		cronExpr.Minute = fields[0]
		cronExpr.Hour = fields[1]
		cronExpr.DayOfMonth = fields[2]
		cronExpr.Month = fields[3]
		cronExpr.DayOfWeek = fields[4]
	}

	// Validate each field
	if err := p.validateFields(cronExpr); err != nil {
		return nil, err
	}

	// Generate human-readable description
	cronExpr.Description = p.generateDescription(cronExpr)

	return cronExpr, nil
}

// ValidateCron validates a cron expression and returns detailed validation errors
func (p *CronParser) ValidateCron(expression string) *CronValidationResult {
	result := &CronValidationResult{
		Valid:      true,
		Expression: expression,
		Errors:     []string{},
		Warnings:   []string{},
	}

	if expression == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "cron expression cannot be empty")
		return result
	}

	// Handle descriptors
	if strings.HasPrefix(expression, "@") {
		return p.validateDescriptor(expression, result)
	}

	fields := strings.Fields(expression)
	if len(fields) < 5 || len(fields) > 6 {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid field count: expected 5 or 6 fields, got %d", len(fields)))
		return result
	}

	// Validate each field
	if len(fields) == 6 {
		p.validateField("second", fields[0], 0, 59, result, false, false)
		p.validateField("minute", fields[1], 0, 59, result, false, false)
		p.validateField("hour", fields[2], 0, 23, result, false, false)
		p.validateField("day-of-month", fields[3], 1, 31, result, true, true)
		p.validateField("month", fields[4], 1, 12, result, false, false)
		p.validateField("day-of-week", fields[5], 0, 6, result, true, false)
	} else {
		p.validateField("minute", fields[0], 0, 59, result, false, false)
		p.validateField("hour", fields[1], 0, 23, result, false, false)
		p.validateField("day-of-month", fields[2], 1, 31, result, true, true)
		p.validateField("month", fields[3], 1, 12, result, false, false)
		p.validateField("day-of-week", fields[4], 0, 6, result, true, false)
	}

	// Add warnings for potentially problematic patterns
	p.checkWarnings(fields, result)

	// Try to parse with base parser for final validation
	_, err := p.baseParser.Parse(p.normalizeExpression(expression))
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("cron parse error: %v", err))
	}

	return result
}

// CronValidationResult contains detailed validation results
type CronValidationResult struct {
	Valid         bool           `json:"valid"`
	Expression    string         `json:"expression"`
	Normalized    string         `json:"normalized,omitempty"`
	Errors        []string       `json:"errors,omitempty"`
	Warnings      []string       `json:"warnings,omitempty"`
	NextRun       string         `json:"next_run,omitempty"`
	FieldAnalysis *FieldAnalysis `json:"field_analysis,omitempty"`
}

// FieldAnalysis provides detailed field-by-field analysis
type FieldAnalysis struct {
	Second     *FieldInfo `json:"second,omitempty"`
	Minute     *FieldInfo `json:"minute"`
	Hour       *FieldInfo `json:"hour"`
	DayOfMonth *FieldInfo `json:"day_of_month"`
	Month      *FieldInfo `json:"month"`
	DayOfWeek  *FieldInfo `json:"day_of_week"`
}

// FieldInfo contains information about a single cron field
type FieldInfo struct {
	Raw         string `json:"raw"`
	Valid       bool   `json:"valid"`
	Description string `json:"description"`
	Values      []int  `json:"values,omitempty"`
	Error       string `json:"error,omitempty"`
}

// validateFields validates all fields of a cron expression
func (p *CronParser) validateFields(expr *CronExpression) error {
	// Validate second field
	if expr.Second != "" && expr.Second != "0" {
		if err := p.validateFieldSyntax(expr.Second, 0, 59, false, false); err != nil {
			return &ValidationError{Message: fmt.Sprintf("invalid second field: %s", err.Error())}
		}
	}

	// Validate minute field
	if err := p.validateFieldSyntax(expr.Minute, 0, 59, false, false); err != nil {
		return &ValidationError{Message: fmt.Sprintf("invalid minute field: %s", err.Error())}
	}

	// Validate hour field
	if err := p.validateFieldSyntax(expr.Hour, 0, 23, false, false); err != nil {
		return &ValidationError{Message: fmt.Sprintf("invalid hour field: %s", err.Error())}
	}

	// Validate day-of-month field (supports L, W, ?)
	if err := p.validateFieldSyntax(expr.DayOfMonth, 1, 31, true, true); err != nil {
		return &ValidationError{Message: fmt.Sprintf("invalid day-of-month field: %s", err.Error())}
	}

	// Validate month field
	if err := p.validateFieldSyntax(expr.Month, 1, 12, false, false); err != nil {
		return &ValidationError{Message: fmt.Sprintf("invalid month field: %s", err.Error())}
	}

	// Validate day-of-week field (supports L, #, ?)
	if err := p.validateFieldSyntax(expr.DayOfWeek, 0, 6, true, false); err != nil {
		return &ValidationError{Message: fmt.Sprintf("invalid day-of-week field: %s", err.Error())}
	}

	return nil
}

// validateFieldSyntax validates the syntax of a single cron field
func (p *CronParser) validateFieldSyntax(field string, min, max int, allowL bool, allowW bool) error {
	if field == "" {
		return fmt.Errorf("field cannot be empty")
	}

	// Handle special characters
	if field == "*" || field == "?" {
		return nil
	}

	// Handle L (last)
	if strings.Contains(field, "L") {
		if !allowL {
			return fmt.Errorf("'L' not allowed in this field")
		}
		return p.validateLField(field, min, max)
	}

	// Handle W (nearest weekday)
	if strings.Contains(field, "W") {
		if !allowW {
			return fmt.Errorf("'W' not allowed in this field")
		}
		return p.validateWField(field, min, max)
	}

	// Handle # (nth occurrence)
	if strings.Contains(field, "#") {
		return p.validateHashField(field)
	}

	// Handle comma-separated values
	parts := strings.Split(field, ",")
	for _, part := range parts {
		if err := p.validateRangePart(part, min, max); err != nil {
			return err
		}
	}

	return nil
}

// validateLField validates fields containing 'L' (last)
func (p *CronParser) validateLField(field string, min, max int) error {
	if field == "L" {
		return nil
	}

	// L-n format (e.g., L-3 for 3 days before the last day)
	if strings.HasPrefix(field, "L-") {
		numStr := strings.TrimPrefix(field, "L-")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return fmt.Errorf("invalid L offset: %s", field)
		}
		if num < 1 || num > max-min {
			return fmt.Errorf("L offset out of range: %d", num)
		}
		return nil
	}

	// nL format for day-of-week (e.g., 5L for last Friday)
	if strings.HasSuffix(field, "L") {
		numStr := strings.TrimSuffix(field, "L")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return fmt.Errorf("invalid day-of-week with L: %s", field)
		}
		if num < 0 || num > 6 {
			return fmt.Errorf("day-of-week out of range: %d", num)
		}
		return nil
	}

	return fmt.Errorf("invalid L syntax: %s", field)
}

// validateWField validates fields containing 'W' (nearest weekday)
func (p *CronParser) validateWField(field string, min, max int) error {
	// LW format (last weekday of month)
	if field == "LW" || field == "WL" {
		return nil
	}

	// nW format (nearest weekday to day n)
	if strings.HasSuffix(field, "W") {
		numStr := strings.TrimSuffix(field, "W")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return fmt.Errorf("invalid W syntax: %s", field)
		}
		if num < min || num > max {
			return fmt.Errorf("day value out of range: %d (must be %d-%d)", num, min, max)
		}
		return nil
	}

	return fmt.Errorf("invalid W syntax: %s", field)
}

// validateHashField validates fields containing '#' (nth occurrence)
func (p *CronParser) validateHashField(field string) error {
	parts := strings.Split(field, "#")
	if len(parts) != 2 {
		return fmt.Errorf("invalid # syntax: %s", field)
	}

	// Validate day-of-week
	dow, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid day-of-week in #: %s", parts[0])
	}
	if dow < 0 || dow > 6 {
		return fmt.Errorf("day-of-week out of range: %d (must be 0-6)", dow)
	}

	// Validate occurrence
	occurrence, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid occurrence in #: %s", parts[1])
	}
	if occurrence < 1 || occurrence > 5 {
		return fmt.Errorf("occurrence out of range: %d (must be 1-5)", occurrence)
	}

	return nil
}

// validateRangePart validates a single part of a cron field (value, range, or step)
func (p *CronParser) validateRangePart(part string, min, max int) error {
	// Handle step values (e.g., */5, 0-30/5)
	if strings.Contains(part, "/") {
		stepParts := strings.Split(part, "/")
		if len(stepParts) != 2 {
			return fmt.Errorf("invalid step syntax: %s", part)
		}

		// Validate step value
		step, err := strconv.Atoi(stepParts[1])
		if err != nil {
			return fmt.Errorf("invalid step value: %s", stepParts[1])
		}
		if step < 1 || step > max-min {
			return fmt.Errorf("step value out of range: %d", step)
		}

		// Validate range part
		if stepParts[0] != "*" {
			return p.validateRangePart(stepParts[0], min, max)
		}
		return nil
	}

	// Handle range values (e.g., 1-5)
	if strings.Contains(part, "-") {
		rangeParts := strings.Split(part, "-")
		if len(rangeParts) != 2 {
			return fmt.Errorf("invalid range syntax: %s", part)
		}

		start, err := strconv.Atoi(rangeParts[0])
		if err != nil {
			return fmt.Errorf("invalid range start: %s", rangeParts[0])
		}
		end, err := strconv.Atoi(rangeParts[1])
		if err != nil {
			return fmt.Errorf("invalid range end: %s", rangeParts[1])
		}

		if start < min || start > max {
			return fmt.Errorf("range start out of bounds: %d (must be %d-%d)", start, min, max)
		}
		if end < min || end > max {
			return fmt.Errorf("range end out of bounds: %d (must be %d-%d)", end, min, max)
		}
		if start > end {
			return fmt.Errorf("range start (%d) must be <= end (%d)", start, end)
		}
		return nil
	}

	// Handle single values
	val, err := strconv.Atoi(part)
	if err != nil {
		return fmt.Errorf("invalid value: %s", part)
	}
	if val < min || val > max {
		return fmt.Errorf("value out of range: %d (must be %d-%d)", val, min, max)
	}

	return nil
}

// parseDescriptor parses cron descriptor expressions (@yearly, @monthly, etc.)
func (p *CronParser) parseDescriptor(expression string) (*CronExpression, error) {
	descriptors := map[string]*CronExpression{
		"@yearly":   {Second: "0", Minute: "0", Hour: "0", DayOfMonth: "1", Month: "1", DayOfWeek: "*", Description: "Once a year at midnight on January 1st"},
		"@annually": {Second: "0", Minute: "0", Hour: "0", DayOfMonth: "1", Month: "1", DayOfWeek: "*", Description: "Once a year at midnight on January 1st"},
		"@monthly":  {Second: "0", Minute: "0", Hour: "0", DayOfMonth: "1", Month: "*", DayOfWeek: "*", Description: "Once a month at midnight on the 1st"},
		"@weekly":   {Second: "0", Minute: "0", Hour: "0", DayOfMonth: "*", Month: "*", DayOfWeek: "0", Description: "Once a week at midnight on Sunday"},
		"@daily":    {Second: "0", Minute: "0", Hour: "0", DayOfMonth: "*", Month: "*", DayOfWeek: "*", Description: "Once a day at midnight"},
		"@midnight": {Second: "0", Minute: "0", Hour: "0", DayOfMonth: "*", Month: "*", DayOfWeek: "*", Description: "Once a day at midnight"},
		"@hourly":   {Second: "0", Minute: "0", Hour: "*", DayOfMonth: "*", Month: "*", DayOfWeek: "*", Description: "Once an hour at the beginning of the hour"},
	}

	expr, ok := descriptors[strings.ToLower(expression)]
	if !ok {
		// Handle @every duration
		if strings.HasPrefix(strings.ToLower(expression), "@every ") {
			return nil, &ValidationError{Message: "@every expressions are not supported"}
		}
		return nil, &ValidationError{Message: fmt.Sprintf("unknown descriptor: %s", expression)}
	}

	result := &CronExpression{
		Second:      expr.Second,
		Minute:      expr.Minute,
		Hour:        expr.Hour,
		DayOfMonth:  expr.DayOfMonth,
		Month:       expr.Month,
		DayOfWeek:   expr.DayOfWeek,
		Timezone:    "UTC",
		Description: expr.Description,
		Original:    expression,
	}

	return result, nil
}

// validateDescriptor validates a descriptor expression
func (p *CronParser) validateDescriptor(expression string, result *CronValidationResult) *CronValidationResult {
	validDescriptors := []string{
		"@yearly", "@annually", "@monthly", "@weekly",
		"@daily", "@midnight", "@hourly",
	}

	lower := strings.ToLower(expression)
	for _, desc := range validDescriptors {
		if lower == desc {
			result.Normalized = expression
			return result
		}
	}

	if strings.HasPrefix(lower, "@every ") {
		result.Valid = false
		result.Errors = append(result.Errors, "@every expressions are not supported")
		return result
	}

	result.Valid = false
	result.Errors = append(result.Errors, fmt.Sprintf("unknown descriptor: %s", expression))
	return result
}

// validateField validates a single field and adds errors/warnings to result
func (p *CronParser) validateField(name, value string, min, max int, result *CronValidationResult, allowL, allowW bool) {
	err := p.validateFieldSyntax(value, min, max, allowL, allowW)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s field: %s", name, err.Error()))
	}
}

// checkWarnings adds warnings for potentially problematic patterns
func (p *CronParser) checkWarnings(fields []string, result *CronValidationResult) {
	// Check for very frequent execution
	offset := 0
	if len(fields) == 6 {
		offset = 1
	}

	// Every minute warning
	if fields[offset] == "*" && fields[offset+1] == "*" {
		result.Warnings = append(result.Warnings, "this schedule runs every minute, which may be resource-intensive")
	}

	// Both day-of-month and day-of-week specified
	domIndex := offset + 2
	dowIndex := offset + 4
	if fields[domIndex] != "*" && fields[domIndex] != "?" &&
		fields[dowIndex] != "*" && fields[dowIndex] != "?" {
		result.Warnings = append(result.Warnings, "both day-of-month and day-of-week are specified; the schedule will run when EITHER matches")
	}
}

// normalizeExpression normalizes a cron expression for the base parser
func (p *CronParser) normalizeExpression(expression string) string {
	// The base parser handles most cases, but we need to handle special characters
	// that it doesn't support by converting them to standard syntax

	fields := strings.Fields(expression)
	normalized := make([]string, len(fields))

	for i, field := range fields {
		normalized[i] = p.normalizeField(field, i, len(fields))
	}

	return strings.Join(normalized, " ")
}

// normalizeField normalizes a single field for the base parser
func (p *CronParser) normalizeField(field string, index, totalFields int) string {
	// Replace ? with *
	if field == "?" {
		return "*"
	}

	// Handle L in day-of-month (convert to last day approximation)
	if strings.Contains(field, "L") {
		if field == "L" {
			return "28-31" // Approximate last day
		}
		if strings.HasPrefix(field, "L-") {
			// L-n: approximate by using a range
			return "*"
		}
		// nL in day-of-week: keep as-is for now
	}

	// Handle W (nearest weekday)
	if strings.Contains(field, "W") {
		// Convert to approximate
		return "*"
	}

	// Handle # (nth occurrence)
	if strings.Contains(field, "#") {
		// Convert to just the day of week
		parts := strings.Split(field, "#")
		if len(parts) == 2 {
			return parts[0]
		}
	}

	return field
}

// generateDescription generates a human-readable description of the cron expression
func (p *CronParser) generateDescription(expr *CronExpression) string {
	var parts []string

	// Time description
	if expr.Minute == "0" && expr.Hour == "0" {
		parts = append(parts, "At midnight")
	} else if expr.Minute == "0" && expr.Hour == "12" {
		parts = append(parts, "At noon")
	} else if expr.Minute == "*" && expr.Hour == "*" {
		parts = append(parts, "Every minute")
	} else if expr.Minute == "0" && expr.Hour == "*" {
		parts = append(parts, "Every hour")
	} else if expr.Minute != "*" && expr.Hour != "*" {
		parts = append(parts, fmt.Sprintf("At %s:%s", p.padTime(expr.Hour), p.padTime(expr.Minute)))
	} else if expr.Minute != "*" {
		parts = append(parts, fmt.Sprintf("At minute %s", expr.Minute))
	}

	// Day description
	if expr.DayOfMonth == "1" && expr.Month == "*" && expr.DayOfWeek == "*" {
		parts = append(parts, "on the 1st of every month")
	} else if expr.DayOfMonth == "L" {
		parts = append(parts, "on the last day of the month")
	} else if expr.DayOfMonth != "*" && expr.DayOfMonth != "?" {
		parts = append(parts, fmt.Sprintf("on day %s", expr.DayOfMonth))
	}

	// Day of week description
	if expr.DayOfWeek != "*" && expr.DayOfWeek != "?" {
		dayName := p.dayOfWeekName(expr.DayOfWeek)
		parts = append(parts, fmt.Sprintf("on %s", dayName))
	}

	// Month description
	if expr.Month != "*" {
		monthName := p.monthName(expr.Month)
		parts = append(parts, fmt.Sprintf("in %s", monthName))
	}

	if len(parts) == 0 {
		return "Custom schedule"
	}

	return strings.Join(parts, " ")
}

// padTime pads a time value with leading zero if needed
func (p *CronParser) padTime(value string) string {
	if len(value) == 1 {
		return "0" + value
	}
	return value
}

// dayOfWeekName converts day-of-week value to name
func (p *CronParser) dayOfWeekName(value string) string {
	days := map[string]string{
		"0": "Sunday", "1": "Monday", "2": "Tuesday", "3": "Wednesday",
		"4": "Thursday", "5": "Friday", "6": "Saturday", "7": "Sunday",
	}
	if name, ok := days[value]; ok {
		return name
	}
	return value
}

// monthName converts month value to name
func (p *CronParser) monthName(value string) string {
	months := map[string]string{
		"1": "January", "2": "February", "3": "March", "4": "April",
		"5": "May", "6": "June", "7": "July", "8": "August",
		"9": "September", "10": "October", "11": "November", "12": "December",
	}
	if name, ok := months[value]; ok {
		return name
	}
	return value
}

// CalculateNextRuns calculates the next N execution times for a cron expression
func (p *CronParser) CalculateNextRuns(expression, timezone string, count int) ([]time.Time, error) {
	if count <= 0 {
		return []time.Time{}, nil
	}

	// Cap count
	const maxCount = 1000
	if count > maxCount {
		count = maxCount
	}

	// Parse expression
	normalized := p.normalizeExpression(expression)
	sched, err := p.baseParser.Parse(normalized)
	if err != nil {
		return nil, &ValidationError{Message: fmt.Sprintf("invalid cron expression: %v", err)}
	}

	// Load timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	// Calculate next N run times
	times := make([]time.Time, 0, count)
	current := time.Now().In(loc)

	for i := 0; i < count; i++ {
		nextRun := sched.Next(current)
		times = append(times, nextRun)
		current = nextRun
	}

	return times, nil
}

// IsValidTimezone checks if a timezone string is valid
func IsValidTimezone(tz string) bool {
	if tz == "" {
		return true // Empty means UTC
	}
	_, err := time.LoadLocation(tz)
	return err == nil
}

// GetTimezoneOffset returns the UTC offset for a timezone at a given time
func GetTimezoneOffset(tz string, t time.Time) (string, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", fmt.Errorf("invalid timezone: %s", tz)
	}

	_, offset := t.In(loc).Zone()
	hours := offset / 3600
	minutes := (offset % 3600) / 60

	sign := "+"
	if hours < 0 {
		sign = "-"
		hours = -hours
	}

	return fmt.Sprintf("UTC%s%02d:%02d", sign, hours, minutes), nil
}

// CommonCronPatterns provides common cron expression patterns
var CommonCronPatterns = map[string]string{
	"every_minute":     "* * * * *",
	"every_5_minutes":  "*/5 * * * *",
	"every_15_minutes": "*/15 * * * *",
	"every_30_minutes": "*/30 * * * *",
	"every_hour":       "0 * * * *",
	"every_2_hours":    "0 */2 * * *",
	"every_6_hours":    "0 */6 * * *",
	"daily_midnight":   "0 0 * * *",
	"daily_noon":       "0 12 * * *",
	"weekly_sunday":    "0 0 * * 0",
	"weekly_monday":    "0 0 * * 1",
	"monthly_first":    "0 0 1 * *",
	"monthly_last":     "0 0 L * *",
	"weekdays":         "0 9 * * 1-5",
	"weekends":         "0 9 * * 0,6",
}

// CronExpressionRegex is a regex pattern for validating cron expressions
var CronExpressionRegex = regexp.MustCompile(`^(@(yearly|annually|monthly|weekly|daily|midnight|hourly))$|^((\*|[0-9,\-\/]+)\s+){4,5}(\*|[0-9,\-\/LW#]+)$`)
