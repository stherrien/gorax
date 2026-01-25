package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gorax/gorax/internal/integration"
)

// TestSuite provides a framework for testing integrations.
type TestSuite struct {
	t           *testing.T
	integration integration.Integration
	mockServer  *MockServer
	config      *integration.Config
	timeout     time.Duration
}

// NewTestSuite creates a new test suite for an integration.
func NewTestSuite(t *testing.T, integ integration.Integration) *TestSuite {
	return &TestSuite{
		t:           t,
		integration: integ,
		mockServer:  NewMockServer(),
		timeout:     30 * time.Second,
	}
}

// WithConfig sets the integration configuration.
func (ts *TestSuite) WithConfig(config *integration.Config) *TestSuite {
	ts.config = config
	return ts
}

// WithTimeout sets the test timeout.
func (ts *TestSuite) WithTimeout(timeout time.Duration) *TestSuite {
	ts.timeout = timeout
	return ts
}

// MockServer returns the mock server for configuring responses.
func (ts *TestSuite) MockServer() *MockServer {
	return ts.mockServer
}

// ServerURL returns the mock server URL.
func (ts *TestSuite) ServerURL() string {
	return ts.mockServer.URL()
}

// Close cleans up the test suite.
func (ts *TestSuite) Close() {
	ts.mockServer.Close()
}

// Run runs a test case.
func (ts *TestSuite) Run(name string, fn func(ctx context.Context)) {
	ts.t.Run(name, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), ts.timeout)
		defer cancel()
		fn(ctx)
	})
}

// Execute executes the integration and returns the result.
func (ts *TestSuite) Execute(ctx context.Context, params integration.JSONMap) (*integration.Result, error) {
	if ts.config == nil {
		ts.t.Fatal("config not set")
	}
	return ts.integration.Execute(ctx, ts.config, params)
}

// AssertSuccess asserts that the result is successful.
func (ts *TestSuite) AssertSuccess(result *integration.Result, err error) {
	ts.t.Helper()
	if err != nil {
		ts.t.Errorf("expected success, got error: %v", err)
	}
	if result == nil {
		ts.t.Error("expected result, got nil")
	} else if !result.Success {
		ts.t.Errorf("expected success, got failure: %s", result.Error)
	}
}

// AssertError asserts that an error occurred.
func (ts *TestSuite) AssertError(err error) {
	ts.t.Helper()
	if err == nil {
		ts.t.Error("expected error, got nil")
	}
}

// AssertStatusCode asserts that the result has the expected status code.
func (ts *TestSuite) AssertStatusCode(result *integration.Result, expected int) {
	ts.t.Helper()
	if result == nil {
		ts.t.Error("expected result, got nil")
		return
	}
	if result.StatusCode != expected {
		ts.t.Errorf("expected status code %d, got %d", expected, result.StatusCode)
	}
}

// AssertRequestCount asserts the number of requests made to the mock server.
func (ts *TestSuite) AssertRequestCount(expected int) {
	ts.t.Helper()
	actual := ts.mockServer.GetRequestCount()
	if actual != expected {
		ts.t.Errorf("expected %d requests, got %d", expected, actual)
	}
}

// AssertLastRequestMethod asserts the method of the last request.
func (ts *TestSuite) AssertLastRequestMethod(expected string) {
	ts.t.Helper()
	req := ts.mockServer.GetLastRequest()
	if req == nil {
		ts.t.Error("no requests recorded")
		return
	}
	if req.Method != expected {
		ts.t.Errorf("expected method %s, got %s", expected, req.Method)
	}
}

// AssertLastRequestPath asserts the path of the last request.
func (ts *TestSuite) AssertLastRequestPath(expected string) {
	ts.t.Helper()
	req := ts.mockServer.GetLastRequest()
	if req == nil {
		ts.t.Error("no requests recorded")
		return
	}
	if req.URL != expected {
		ts.t.Errorf("expected path %s, got %s", expected, req.URL)
	}
}

// AssertLastRequestHeader asserts a header value of the last request.
func (ts *TestSuite) AssertLastRequestHeader(key, expected string) {
	ts.t.Helper()
	req := ts.mockServer.GetLastRequest()
	if req == nil {
		ts.t.Error("no requests recorded")
		return
	}
	actual := req.Headers.Get(key)
	if actual != expected {
		ts.t.Errorf("expected header %s=%s, got %s", key, expected, actual)
	}
}

// TestCase represents a single test case.
type TestCase struct {
	Name          string
	Config        *integration.Config
	Params        integration.JSONMap
	MockResponses map[string]*MockResponse
	ExpectSuccess bool
	ExpectError   bool
	ExpectStatus  int
	Validate      func(*testing.T, *integration.Result, error)
}

// RunTestCases runs multiple test cases.
func (ts *TestSuite) RunTestCases(cases []TestCase) {
	for _, tc := range cases {
		ts.t.Run(tc.Name, func(t *testing.T) {
			// Reset mock server
			ts.mockServer.Reset()

			// Configure mock responses
			for endpoint, response := range tc.MockResponses {
				// Parse endpoint as "METHOD:path"
				method := "GET"
				path := endpoint
				for i, c := range endpoint {
					if c == ':' {
						method = endpoint[:i]
						path = endpoint[i+1:]
						break
					}
				}
				ts.mockServer.OnRequest(method, path, response)
			}

			// Execute
			ctx, cancel := context.WithTimeout(context.Background(), ts.timeout)
			defer cancel()

			config := tc.Config
			if config == nil {
				config = ts.config
			}

			result, err := ts.integration.Execute(ctx, config, tc.Params)

			// Validate results
			if tc.ExpectSuccess && (err != nil || result == nil || !result.Success) {
				t.Errorf("expected success, got error: %v, result: %+v", err, result)
			}

			if tc.ExpectError && err == nil {
				t.Error("expected error, got nil")
			}

			if tc.ExpectStatus > 0 && result != nil && result.StatusCode != tc.ExpectStatus {
				t.Errorf("expected status %d, got %d", tc.ExpectStatus, result.StatusCode)
			}

			if tc.Validate != nil {
				tc.Validate(t, result, err)
			}
		})
	}
}

// IntegrationTester provides utilities for testing specific integration behavior.
type IntegrationTester struct {
	t          *testing.T
	mockServer *MockServer
}

// NewIntegrationTester creates a new integration tester.
func NewIntegrationTester(t *testing.T) *IntegrationTester {
	return &IntegrationTester{
		t:          t,
		mockServer: NewMockServer(),
	}
}

// Server returns the mock server.
func (it *IntegrationTester) Server() *MockServer {
	return it.mockServer
}

// Close cleans up the tester.
func (it *IntegrationTester) Close() {
	it.mockServer.Close()
}

// TestHealthCheck tests the health check functionality of an integration.
func (it *IntegrationTester) TestHealthCheck(ctx context.Context, integ integration.Integration) {
	it.t.Helper()

	healthCheckable, ok := integ.(integration.HealthCheckable)
	if !ok {
		it.t.Skip("integration does not implement HealthCheckable")
		return
	}

	err := healthCheckable.HealthCheck(ctx)
	if err != nil {
		it.t.Errorf("health check failed: %v", err)
	}
}

// TestValidation tests the validation functionality of an integration.
func (it *IntegrationTester) TestValidation(integ integration.Integration, validConfig, invalidConfig *integration.Config) {
	it.t.Helper()

	// Test valid config
	if validConfig != nil {
		if err := integ.Validate(validConfig); err != nil {
			it.t.Errorf("validation failed for valid config: %v", err)
		}
	}

	// Test invalid config
	if invalidConfig != nil {
		if err := integ.Validate(invalidConfig); err == nil {
			it.t.Error("expected validation error for invalid config")
		}
	}
}

// TestSchema tests that an integration has a valid schema.
func (it *IntegrationTester) TestSchema(integ integration.Integration) {
	it.t.Helper()

	schema := integ.GetSchema()
	if schema == nil {
		it.t.Error("integration has nil schema")
		return
	}

	// Basic schema validation
	if len(schema.ConfigSpec) == 0 {
		it.t.Log("warning: integration has empty config spec")
	}
}

// TestMetadata tests that an integration has valid metadata.
func (it *IntegrationTester) TestMetadata(integ integration.Integration) {
	it.t.Helper()

	metadata := integ.GetMetadata()
	if metadata == nil {
		it.t.Error("integration has nil metadata")
		return
	}

	if metadata.Name == "" {
		it.t.Error("integration metadata has empty name")
	}

	if metadata.Version == "" {
		it.t.Log("warning: integration metadata has empty version")
	}
}

// BenchmarkConfig holds configuration for benchmarking an integration.
type BenchmarkConfig struct {
	Config     *integration.Config
	Params     integration.JSONMap
	Iterations int
}

// BenchmarkResult holds the results of a benchmark.
type BenchmarkResult struct {
	TotalDuration   time.Duration
	AverageDuration time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration
	SuccessCount    int
	ErrorCount      int
}

// RunBenchmark benchmarks an integration.
func (it *IntegrationTester) RunBenchmark(ctx context.Context, integ integration.Integration, config BenchmarkConfig) *BenchmarkResult {
	iterations := config.Iterations
	if iterations <= 0 {
		iterations = 100
	}

	result := &BenchmarkResult{
		MinDuration: time.Hour, // Start with a large value
	}

	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := integ.Execute(ctx, config.Config, config.Params)
		duration := time.Since(start)

		result.TotalDuration += duration

		if duration < result.MinDuration {
			result.MinDuration = duration
		}
		if duration > result.MaxDuration {
			result.MaxDuration = duration
		}

		if err != nil {
			result.ErrorCount++
		} else {
			result.SuccessCount++
		}
	}

	result.AverageDuration = result.TotalDuration / time.Duration(iterations)
	return result
}

// String returns a string representation of the benchmark result.
func (br *BenchmarkResult) String() string {
	return fmt.Sprintf(
		"Benchmark: total=%v, avg=%v, min=%v, max=%v, success=%d, errors=%d",
		br.TotalDuration, br.AverageDuration, br.MinDuration, br.MaxDuration,
		br.SuccessCount, br.ErrorCount,
	)
}
