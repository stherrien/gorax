package testing

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gorax/gorax/internal/integration"
)

// mockTestIntegration is a mock integration for testing the test suite
type mockTestIntegration struct {
	*integration.BaseIntegration
	executeFunc func(ctx context.Context, config *integration.Config, params integration.JSONMap) (*integration.Result, error)
}

func newMockTestIntegration() *mockTestIntegration {
	return &mockTestIntegration{
		BaseIntegration: integration.NewBaseIntegration("mock", integration.TypeHTTP),
	}
}

func (m *mockTestIntegration) Execute(ctx context.Context, config *integration.Config, params integration.JSONMap) (*integration.Result, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, config, params)
	}
	return integration.NewSuccessResult(nil, 0), nil
}

func (m *mockTestIntegration) Validate(config *integration.Config) error {
	return m.BaseIntegration.ValidateConfig(config)
}

func (m *mockTestIntegration) HealthCheck(ctx context.Context) error {
	return nil
}

func TestNewTestSuite(t *testing.T) {
	integ := newMockTestIntegration()
	suite := NewTestSuite(t, integ)
	defer suite.Close()

	assert.NotNil(t, suite)
	assert.NotNil(t, suite.mockServer)
	assert.NotEmpty(t, suite.ServerURL())
}

func TestTestSuite_WithConfig(t *testing.T) {
	integ := newMockTestIntegration()
	suite := NewTestSuite(t, integ)
	defer suite.Close()

	config := &integration.Config{
		Name: "test-config",
		Type: integration.TypeHTTP,
	}

	result := suite.WithConfig(config)
	assert.Same(t, suite, result)
	assert.Equal(t, config, suite.config)
}

func TestTestSuite_WithTimeout(t *testing.T) {
	integ := newMockTestIntegration()
	suite := NewTestSuite(t, integ)
	defer suite.Close()

	result := suite.WithTimeout(60 * time.Second)
	assert.Same(t, suite, result)
	assert.Equal(t, 60*time.Second, suite.timeout)
}

func TestTestSuite_MockServer(t *testing.T) {
	integ := newMockTestIntegration()
	suite := NewTestSuite(t, integ)
	defer suite.Close()

	mockServer := suite.MockServer()
	assert.NotNil(t, mockServer)

	// Can configure mock responses
	mockServer.OnGet("/test", JSONResponse(http.StatusOK, nil))
}

func TestTestSuite_Run(t *testing.T) {
	integ := newMockTestIntegration()
	suite := NewTestSuite(t, integ)
	defer suite.Close()

	executed := false
	suite.Run("test case", func(ctx context.Context) {
		executed = true
		assert.NotNil(t, ctx)
	})

	assert.True(t, executed)
}

func TestTestSuite_Execute(t *testing.T) {
	integ := newMockTestIntegration()
	integ.executeFunc = func(ctx context.Context, config *integration.Config, params integration.JSONMap) (*integration.Result, error) {
		return integration.NewSuccessResult(integration.JSONMap{"result": "ok"}, 100), nil
	}

	suite := NewTestSuite(t, integ)
	defer suite.Close()

	suite.WithConfig(&integration.Config{
		Name: "test",
		Type: integration.TypeHTTP,
	})

	result, err := suite.Execute(context.Background(), nil)
	assert.NoError(t, err)
	assert.True(t, result.Success)
}

func TestTestSuite_AssertSuccess(t *testing.T) {
	// Create a sub-test to catch test failures
	tt := &testing.T{}
	integ := newMockTestIntegration()
	suite := NewTestSuite(tt, integ)
	defer suite.Close()

	t.Run("success case", func(t *testing.T) {
		result := integration.NewSuccessResult(nil, 0)
		suite.AssertSuccess(result, nil)
		// No assertion here - it should pass silently
	})
}

func TestTestSuite_AssertError(t *testing.T) {
	// Create a sub-test to catch test failures
	tt := &testing.T{}
	integ := newMockTestIntegration()
	suite := NewTestSuite(tt, integ)
	defer suite.Close()

	t.Run("error case", func(t *testing.T) {
		err := integration.ErrNotFound
		suite.AssertError(err)
		// No assertion here - it should pass silently
	})
}

func TestTestSuite_AssertStatusCode(t *testing.T) {
	tt := &testing.T{}
	integ := newMockTestIntegration()
	suite := NewTestSuite(tt, integ)
	defer suite.Close()

	result := &integration.Result{
		StatusCode: http.StatusOK,
	}

	suite.AssertStatusCode(result, http.StatusOK)
}

func TestTestSuite_AssertRequestCount(t *testing.T) {
	tt := &testing.T{}
	integ := newMockTestIntegration()
	suite := NewTestSuite(tt, integ)
	defer suite.Close()

	suite.MockServer().OnGet("/test", JSONResponse(http.StatusOK, nil))

	http.Get(suite.ServerURL() + "/test")
	http.Get(suite.ServerURL() + "/test")

	suite.AssertRequestCount(2)
}

func TestTestSuite_AssertLastRequestMethod(t *testing.T) {
	tt := &testing.T{}
	integ := newMockTestIntegration()
	suite := NewTestSuite(tt, integ)
	defer suite.Close()

	suite.MockServer().OnPost("/test", JSONResponse(http.StatusOK, nil))

	http.Post(suite.ServerURL()+"/test", "application/json", nil)

	suite.AssertLastRequestMethod("POST")
}

func TestTestSuite_AssertLastRequestPath(t *testing.T) {
	tt := &testing.T{}
	integ := newMockTestIntegration()
	suite := NewTestSuite(tt, integ)
	defer suite.Close()

	suite.MockServer().OnGet("/users/123", JSONResponse(http.StatusOK, nil))

	http.Get(suite.ServerURL() + "/users/123")

	suite.AssertLastRequestPath("/users/123")
}

func TestTestSuite_AssertLastRequestHeader(t *testing.T) {
	tt := &testing.T{}
	integ := newMockTestIntegration()
	suite := NewTestSuite(tt, integ)
	defer suite.Close()

	suite.MockServer().OnGet("/test", JSONResponse(http.StatusOK, nil))

	req, _ := http.NewRequest("GET", suite.ServerURL()+"/test", nil)
	req.Header.Set("X-Custom-Header", "custom-value")
	http.DefaultClient.Do(req)

	suite.AssertLastRequestHeader("X-Custom-Header", "custom-value")
}

func TestTestCase(t *testing.T) {
	tc := TestCase{
		Name: "test case",
		Config: &integration.Config{
			Name: "test",
			Type: integration.TypeHTTP,
		},
		Params: integration.JSONMap{
			"key": "value",
		},
		MockResponses: map[string]*MockResponse{
			"GET:/test": JSONResponse(http.StatusOK, nil),
		},
		ExpectSuccess: true,
		ExpectError:   false,
		ExpectStatus:  http.StatusOK,
	}

	assert.Equal(t, "test case", tc.Name)
	assert.NotNil(t, tc.Config)
	assert.NotNil(t, tc.Params)
}

func TestTestSuite_RunTestCases(t *testing.T) {
	integ := newMockTestIntegration()
	integ.executeFunc = func(ctx context.Context, config *integration.Config, params integration.JSONMap) (*integration.Result, error) {
		return &integration.Result{
			Success:    true,
			StatusCode: http.StatusOK,
		}, nil
	}

	suite := NewTestSuite(t, integ)
	defer suite.Close()

	suite.WithConfig(&integration.Config{
		Name: "default",
		Type: integration.TypeHTTP,
	})

	cases := []TestCase{
		{
			Name:          "success case",
			ExpectSuccess: true,
			ExpectStatus:  http.StatusOK,
		},
	}

	suite.RunTestCases(cases)
}

func TestNewIntegrationTester(t *testing.T) {
	tester := NewIntegrationTester(t)
	defer tester.Close()

	assert.NotNil(t, tester)
	assert.NotNil(t, tester.Server())
}

func TestIntegrationTester_TestHealthCheck(t *testing.T) {
	tester := NewIntegrationTester(t)
	defer tester.Close()

	integ := newMockTestIntegration()
	tester.TestHealthCheck(context.Background(), integ)
}

func TestIntegrationTester_TestValidation(t *testing.T) {
	tester := NewIntegrationTester(t)
	defer tester.Close()

	integ := newMockTestIntegration()

	validConfig := &integration.Config{
		Name: "valid",
		Type: integration.TypeHTTP,
	}

	invalidConfig := &integration.Config{
		Name: "",
		Type: integration.TypeHTTP,
	}

	tester.TestValidation(integ, validConfig, invalidConfig)
}

func TestIntegrationTester_TestSchema(t *testing.T) {
	tester := NewIntegrationTester(t)
	defer tester.Close()

	integ := newMockTestIntegration()
	tester.TestSchema(integ)
}

func TestIntegrationTester_TestMetadata(t *testing.T) {
	tester := NewIntegrationTester(t)
	defer tester.Close()

	integ := newMockTestIntegration()
	tester.TestMetadata(integ)
}

func TestBenchmarkConfig(t *testing.T) {
	config := BenchmarkConfig{
		Config: &integration.Config{
			Name: "test",
			Type: integration.TypeHTTP,
		},
		Params: integration.JSONMap{
			"key": "value",
		},
		Iterations: 100,
	}

	assert.NotNil(t, config.Config)
	assert.Equal(t, 100, config.Iterations)
}

func TestBenchmarkResult_String(t *testing.T) {
	result := &BenchmarkResult{
		TotalDuration:   time.Second,
		AverageDuration: 10 * time.Millisecond,
		MinDuration:     5 * time.Millisecond,
		MaxDuration:     20 * time.Millisecond,
		SuccessCount:    95,
		ErrorCount:      5,
	}

	str := result.String()
	assert.Contains(t, str, "Benchmark")
	assert.Contains(t, str, "95")
	assert.Contains(t, str, "5")
}

func TestIntegrationTester_RunBenchmark(t *testing.T) {
	tester := NewIntegrationTester(t)
	defer tester.Close()

	integ := newMockTestIntegration()
	integ.executeFunc = func(ctx context.Context, config *integration.Config, params integration.JSONMap) (*integration.Result, error) {
		return integration.NewSuccessResult(nil, 0), nil
	}

	config := BenchmarkConfig{
		Config: &integration.Config{
			Name: "test",
			Type: integration.TypeHTTP,
		},
		Iterations: 10,
	}

	result := tester.RunBenchmark(context.Background(), integ, config)

	assert.NotNil(t, result)
	assert.Equal(t, 10, result.SuccessCount)
	assert.Equal(t, 0, result.ErrorCount)
	assert.Greater(t, result.TotalDuration, time.Duration(0))
}

func TestIntegrationTester_RunBenchmark_DefaultIterations(t *testing.T) {
	tester := NewIntegrationTester(t)
	defer tester.Close()

	integ := newMockTestIntegration()

	config := BenchmarkConfig{
		Config: &integration.Config{
			Name: "test",
			Type: integration.TypeHTTP,
		},
		Iterations: 0, // Should default to 100
	}

	result := tester.RunBenchmark(context.Background(), integ, config)
	assert.Equal(t, 100, result.SuccessCount+result.ErrorCount)
}
