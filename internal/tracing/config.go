package tracing

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ExporterType defines the type of trace exporter to use
type ExporterType string

const (
	// ExporterTypeOTLP exports traces via OTLP protocol (default)
	ExporterTypeOTLP ExporterType = "otlp"
	// ExporterTypeJaeger exports traces to Jaeger (deprecated, use OTLP with Jaeger collector)
	ExporterTypeJaeger ExporterType = "jaeger"
	// ExporterTypeConsole outputs traces to console (for debugging)
	ExporterTypeConsole ExporterType = "console"
	// ExporterTypeNone disables trace export
	ExporterTypeNone ExporterType = "none"
)

// TracingConfig holds OpenTelemetry tracing configuration
type TracingConfig struct {
	// Enabled indicates whether distributed tracing is enabled
	Enabled bool
	// ServiceName is the name of this service in traces
	ServiceName string
	// ServiceVersion is the version of this service
	ServiceVersion string
	// ExporterType determines which exporter to use (otlp, jaeger, console)
	ExporterType ExporterType
	// ExporterEndpoint is the endpoint for the trace collector
	ExporterEndpoint string
	// SamplingRate is the probability of sampling a trace (0.0 to 1.0)
	SamplingRate float64
	// ResourceAttributes are additional attributes to add to all spans
	ResourceAttributes map[string]string
	// BatchConfig holds batch span processor configuration
	BatchConfig BatchConfig
	// Insecure indicates whether to use insecure connection (no TLS)
	Insecure bool
	// Headers are additional headers to send with OTLP requests
	Headers map[string]string
}

// BatchConfig holds configuration for batch span processing
type BatchConfig struct {
	// MaxQueueSize is the maximum queue size for pending spans
	MaxQueueSize int
	// BatchTimeout is the maximum time to wait before exporting a batch
	BatchTimeoutMs int
	// ExportTimeoutMs is the timeout for exporting spans
	ExportTimeoutMs int
	// MaxExportBatchSize is the maximum number of spans per export batch
	MaxExportBatchSize int
}

// DefaultBatchConfig returns default batch configuration
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		MaxQueueSize:       2048,
		BatchTimeoutMs:     5000,  // 5 seconds
		ExportTimeoutMs:    30000, // 30 seconds
		MaxExportBatchSize: 512,
	}
}

// LoadTracingConfig loads tracing configuration from environment variables
func LoadTracingConfig() *TracingConfig {
	cfg := &TracingConfig{
		Enabled:            getEnvAsBool("TRACING_ENABLED", false),
		ServiceName:        getEnv("TRACING_SERVICE_NAME", "gorax"),
		ServiceVersion:     getEnv("TRACING_SERVICE_VERSION", "1.0.0"),
		ExporterType:       ExporterType(getEnv("TRACING_EXPORTER_TYPE", string(ExporterTypeOTLP))),
		ExporterEndpoint:   getEnv("TRACING_ENDPOINT", "localhost:4317"),
		SamplingRate:       getEnvAsFloat("TRACING_SAMPLING_RATE", 1.0),
		ResourceAttributes: parseResourceAttributes(getEnv("TRACING_RESOURCE_ATTRIBUTES", "")),
		Insecure:           getEnvAsBool("TRACING_INSECURE", true),
		Headers:            parseHeaders(getEnv("TRACING_HEADERS", "")),
		BatchConfig: BatchConfig{
			MaxQueueSize:       getEnvAsInt("TRACING_BATCH_MAX_QUEUE_SIZE", 2048),
			BatchTimeoutMs:     getEnvAsInt("TRACING_BATCH_TIMEOUT_MS", 5000),
			ExportTimeoutMs:    getEnvAsInt("TRACING_BATCH_EXPORT_TIMEOUT_MS", 30000),
			MaxExportBatchSize: getEnvAsInt("TRACING_BATCH_MAX_EXPORT_SIZE", 512),
		},
	}

	// Set default resource attributes
	if cfg.ResourceAttributes == nil {
		cfg.ResourceAttributes = make(map[string]string)
	}
	if _, ok := cfg.ResourceAttributes["deployment.environment"]; !ok {
		cfg.ResourceAttributes["deployment.environment"] = getEnv("APP_ENV", "development")
	}

	return cfg
}

// ValidateConfig validates the tracing configuration
func (c *TracingConfig) ValidateConfig() error {
	if !c.Enabled {
		return nil // No validation needed if tracing is disabled
	}

	var errs []error

	// Validate service name
	if c.ServiceName == "" {
		errs = append(errs, errors.New("tracing service name cannot be empty"))
	}

	// Validate exporter type
	switch c.ExporterType {
	case ExporterTypeOTLP, ExporterTypeJaeger, ExporterTypeConsole, ExporterTypeNone:
		// Valid exporter types
	default:
		errs = append(errs, fmt.Errorf("invalid exporter type: %s (must be otlp, jaeger, console, or none)", c.ExporterType))
	}

	// Validate endpoint for non-console exporters
	if c.ExporterType != ExporterTypeConsole && c.ExporterType != ExporterTypeNone {
		if c.ExporterEndpoint == "" {
			errs = append(errs, errors.New("tracing endpoint cannot be empty for OTLP/Jaeger exporters"))
		}
	}

	// Validate sampling rate
	if c.SamplingRate < 0.0 || c.SamplingRate > 1.0 {
		errs = append(errs, fmt.Errorf("sampling rate must be between 0.0 and 1.0, got: %f", c.SamplingRate))
	}

	// Validate batch config
	if c.BatchConfig.MaxQueueSize <= 0 {
		errs = append(errs, errors.New("batch max queue size must be positive"))
	}
	if c.BatchConfig.BatchTimeoutMs <= 0 {
		errs = append(errs, errors.New("batch timeout must be positive"))
	}
	if c.BatchConfig.MaxExportBatchSize <= 0 {
		errs = append(errs, errors.New("max export batch size must be positive"))
	}

	if len(errs) > 0 {
		return combineErrors(errs)
	}
	return nil
}

// IsEnabled returns true if tracing is enabled
func (c *TracingConfig) IsEnabled() bool {
	return c.Enabled
}

// parseResourceAttributes parses a comma-separated list of key=value pairs
func parseResourceAttributes(s string) map[string]string {
	if s == "" {
		return make(map[string]string)
	}

	attrs := make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				attrs[key] = value
			}
		}
	}
	return attrs
}

// parseHeaders parses HTTP headers from environment variable
func parseHeaders(s string) map[string]string {
	return parseResourceAttributes(s) // Same format: key=value,key2=value2
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// combineErrors combines multiple errors into a single error
func combineErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}

	var sb strings.Builder
	sb.WriteString("tracing configuration errors: ")
	for i, err := range errs {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return errors.New(sb.String())
}
