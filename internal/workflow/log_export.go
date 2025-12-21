package workflow

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"
)

// LogExporter handles exporting execution logs in various formats
type LogExporter struct{}

// NewLogExporter creates a new log exporter
func NewLogExporter() *LogExporter {
	return &LogExporter{}
}

// ExportTXT exports logs in human-readable text format
func (e *LogExporter) ExportTXT(execution *Execution, steps []*StepExecution) []byte {
	var buf bytes.Buffer

	writeHeader(&buf, execution)
	writeStepsTXT(&buf, steps)

	return buf.Bytes()
}

// ExportJSON exports logs in JSON format
func (e *LogExporter) ExportJSON(execution *Execution, steps []*StepExecution) []byte {
	output := createJSONOutput(execution, steps)

	result, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return []byte(`{"error":"failed to marshal JSON"}`)
	}

	return result
}

// ExportCSV exports logs in CSV format
func (e *LogExporter) ExportCSV(execution *Execution, steps []*StepExecution) []byte {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	writeCSVHeader(writer)
	writeStepsCSV(writer, steps)

	writer.Flush()
	return buf.Bytes()
}

// writeHeader writes execution header information
func writeHeader(buf *bytes.Buffer, execution *Execution) {
	buf.WriteString("=" + repeatString("=", 70) + "\n")
	buf.WriteString("WORKFLOW EXECUTION LOG\n")
	buf.WriteString("=" + repeatString("=", 70) + "\n\n")

	buf.WriteString(fmt.Sprintf("Execution ID: %s\n", execution.ID))
	buf.WriteString(fmt.Sprintf("Workflow ID: %s\n", execution.WorkflowID))
	buf.WriteString(fmt.Sprintf("Status: %s\n", execution.Status))

	if execution.StartedAt != nil {
		buf.WriteString(fmt.Sprintf("Started At: %s\n", execution.StartedAt.Format(time.RFC3339)))
	}
	if execution.CompletedAt != nil {
		buf.WriteString(fmt.Sprintf("Completed At: %s\n", execution.CompletedAt.Format(time.RFC3339)))
	}

	buf.WriteString("\n")
}

// writeStepsTXT writes steps in text format
func writeStepsTXT(buf *bytes.Buffer, steps []*StepExecution) {
	if len(steps) == 0 {
		buf.WriteString("No steps executed\n")
		return
	}

	buf.WriteString("STEPS:\n")
	buf.WriteString("-" + repeatString("-", 70) + "\n\n")

	for i, step := range steps {
		writeStepTXT(buf, step, i+1)
	}
}

// writeStepTXT writes a single step in text format
func writeStepTXT(buf *bytes.Buffer, step *StepExecution, index int) {
	buf.WriteString(fmt.Sprintf("Step %d: %s\n", index, step.ID))
	buf.WriteString(fmt.Sprintf("  Node ID: %s\n", step.NodeID))
	buf.WriteString(fmt.Sprintf("  Type: %s\n", step.NodeType))
	buf.WriteString(fmt.Sprintf("  Status: %s\n", step.Status))

	if step.StartedAt != nil {
		buf.WriteString(fmt.Sprintf("  Started: %s\n", step.StartedAt.Format(time.RFC3339)))
	}
	if step.CompletedAt != nil {
		buf.WriteString(fmt.Sprintf("  Completed: %s\n", step.CompletedAt.Format(time.RFC3339)))
	}
	if step.DurationMs != nil {
		buf.WriteString(fmt.Sprintf("  Duration: %dms\n", *step.DurationMs))
	}

	if step.InputData != nil && len(*step.InputData) > 0 {
		buf.WriteString("  Input:\n")
		writeIndentedJSON(buf, *step.InputData, 4)
	}

	if step.OutputData != nil && len(*step.OutputData) > 0 {
		buf.WriteString("  Output:\n")
		writeIndentedJSON(buf, *step.OutputData, 4)
	}

	if step.ErrorMessage != nil {
		buf.WriteString(fmt.Sprintf("  Error: %s\n", *step.ErrorMessage))
	}

	buf.WriteString("\n")
}

// writeIndentedJSON writes JSON with indentation
func writeIndentedJSON(buf *bytes.Buffer, data json.RawMessage, indent int) {
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		buf.WriteString(repeatString(" ", indent) + string(data) + "\n")
		return
	}

	formatted, err := json.MarshalIndent(parsed, repeatString(" ", indent), "  ")
	if err != nil {
		buf.WriteString(repeatString(" ", indent) + string(data) + "\n")
		return
	}

	buf.WriteString(string(formatted) + "\n")
}

// createJSONOutput creates the JSON output structure
func createJSONOutput(execution *Execution, steps []*StepExecution) map[string]interface{} {
	output := map[string]interface{}{
		"execution_id":     execution.ID,
		"workflow_id":      execution.WorkflowID,
		"workflow_version": execution.WorkflowVersion,
		"status":           execution.Status,
		"trigger_type":     execution.TriggerType,
		"steps":            convertStepsToJSON(steps),
	}

	if execution.StartedAt != nil {
		output["started_at"] = execution.StartedAt.Format(time.RFC3339)
	}
	if execution.CompletedAt != nil {
		output["completed_at"] = execution.CompletedAt.Format(time.RFC3339)
	}
	if execution.ErrorMessage != nil {
		output["error_message"] = *execution.ErrorMessage
	}

	return output
}

// convertStepsToJSON converts steps to JSON format
func convertStepsToJSON(steps []*StepExecution) []map[string]interface{} {
	result := make([]map[string]interface{}, len(steps))

	for i, step := range steps {
		result[i] = convertStepToJSON(step)
	}

	return result
}

// convertStepToJSON converts a single step to JSON format
func convertStepToJSON(step *StepExecution) map[string]interface{} {
	stepData := map[string]interface{}{
		"step_id":   step.ID,
		"node_id":   step.NodeID,
		"node_type": step.NodeType,
		"status":    step.Status,
	}

	if step.StartedAt != nil {
		stepData["started_at"] = step.StartedAt.Format(time.RFC3339)
	}
	if step.CompletedAt != nil {
		stepData["completed_at"] = step.CompletedAt.Format(time.RFC3339)
	}
	if step.DurationMs != nil {
		stepData["duration_ms"] = *step.DurationMs
	}
	if step.InputData != nil {
		stepData["input_data"] = json.RawMessage(*step.InputData)
	}
	if step.OutputData != nil {
		stepData["output_data"] = json.RawMessage(*step.OutputData)
	}
	if step.ErrorMessage != nil {
		stepData["error_message"] = *step.ErrorMessage
	}

	return stepData
}

// writeCSVHeader writes CSV header row
func writeCSVHeader(writer *csv.Writer) {
	writer.Write([]string{
		"step_id",
		"node_id",
		"node_type",
		"status",
		"started_at",
		"completed_at",
		"duration_ms",
		"error_message",
	})
}

// writeStepsCSV writes steps in CSV format
func writeStepsCSV(writer *csv.Writer, steps []*StepExecution) {
	for _, step := range steps {
		writeStepCSV(writer, step)
	}
}

// writeStepCSV writes a single step in CSV format
func writeStepCSV(writer *csv.Writer, step *StepExecution) {
	row := []string{
		step.ID,
		step.NodeID,
		step.NodeType,
		step.Status,
		formatTimeCSV(step.StartedAt),
		formatTimeCSV(step.CompletedAt),
		formatIntCSV(step.DurationMs),
		formatStringCSV(step.ErrorMessage),
	}

	writer.Write(row)
}

// formatTimeCSV formats time for CSV
func formatTimeCSV(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// formatIntCSV formats int pointer for CSV
func formatIntCSV(i *int) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%d", *i)
}

// formatStringCSV formats string pointer for CSV
func formatStringCSV(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// repeatString repeats a string n times
func repeatString(s string, n int) string {
	var result bytes.Buffer
	for i := 0; i < n; i++ {
		result.WriteString(s)
	}
	return result.String()
}
