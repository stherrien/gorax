package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gorax/gorax/internal/tenant"
)

// MockQuotaTenantService implements QuotaTenantService for testing
type MockQuotaTenantService struct {
	mock.Mock
}

func (m *MockQuotaTenantService) GetWorkflowCount(ctx context.Context, tenantID string) (int, error) {
	args := m.Called(ctx, tenantID)
	return args.Int(0), args.Error(1)
}

func (m *MockQuotaTenantService) GetConcurrentExecutions(ctx context.Context, tenantID string) (int, error) {
	args := m.Called(ctx, tenantID)
	return args.Int(0), args.Error(1)
}

// MockQuotaRedisClient implements QuotaRedisClient for testing
type MockQuotaRedisClient struct {
	mock.Mock
}

func (m *MockQuotaRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockQuotaRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockQuotaRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockQuotaRedisClient) Pipeline() redis.Pipeliner {
	args := m.Called()
	return args.Get(0).(redis.Pipeliner)
}

// MockRedisPipeline implements redis.Pipeliner for testing
type MockRedisPipeline struct {
	mock.Mock
}

func (m *MockRedisPipeline) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisPipeline) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisPipeline) ZRemRangeByScore(ctx context.Context, key, min, max string) *redis.IntCmd {
	args := m.Called(ctx, key, min, max)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisPipeline) ZCard(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisPipeline) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisPipeline) Exec(ctx context.Context) ([]redis.Cmder, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]redis.Cmder), args.Error(1)
}

// Implement other redis.Pipeliner methods as no-ops for the interface
func (m *MockRedisPipeline) Process(ctx context.Context, cmd redis.Cmder) error           { return nil }
func (m *MockRedisPipeline) Discard()                                                     {}
func (m *MockRedisPipeline) Len() int                                                     { return 0 }
func (m *MockRedisPipeline) Do(ctx context.Context, args ...interface{}) *redis.Cmd       { return nil }
func (m *MockRedisPipeline) Command(ctx context.Context) *redis.CommandsInfoCmd           { return nil }
func (m *MockRedisPipeline) CommandList(ctx context.Context, filter *redis.FilterBy) *redis.StringSliceCmd {
	return nil
}
func (m *MockRedisPipeline) CommandGetKeys(ctx context.Context, commands ...interface{}) *redis.StringSliceCmd {
	return nil
}
func (m *MockRedisPipeline) CommandDocs(ctx context.Context, commands ...string) *redis.MapMapStringInterfaceCmd {
	return nil
}
func (m *MockRedisPipeline) ClientGetName(ctx context.Context) *redis.StringCmd           { return nil }
func (m *MockRedisPipeline) Echo(ctx context.Context, message interface{}) *redis.StringCmd {
	return nil
}
func (m *MockRedisPipeline) Ping(ctx context.Context) *redis.StatusCmd                    { return nil }
func (m *MockRedisPipeline) Quit(ctx context.Context) *redis.StatusCmd                    { return nil }
func (m *MockRedisPipeline) Unlink(ctx context.Context, keys ...string) *redis.IntCmd     { return nil }
func (m *MockRedisPipeline) BgRewriteAOF(ctx context.Context) *redis.StatusCmd            { return nil }
func (m *MockRedisPipeline) BgSave(ctx context.Context) *redis.StatusCmd                  { return nil }
func (m *MockRedisPipeline) ClientKill(ctx context.Context, ipPort string) *redis.StatusCmd {
	return nil
}
func (m *MockRedisPipeline) ClientKillByFilter(ctx context.Context, keys ...string) *redis.IntCmd {
	return nil
}
func (m *MockRedisPipeline) ClientList(ctx context.Context) *redis.StringCmd              { return nil }
func (m *MockRedisPipeline) ClientPause(ctx context.Context, dur time.Duration) *redis.BoolCmd {
	return nil
}
func (m *MockRedisPipeline) ClientUnpause(ctx context.Context) *redis.BoolCmd             { return nil }
func (m *MockRedisPipeline) ClientID(ctx context.Context) *redis.IntCmd                   { return nil }
func (m *MockRedisPipeline) ClientUnblock(ctx context.Context, id int64) *redis.IntCmd    { return nil }
func (m *MockRedisPipeline) ClientUnblockWithError(ctx context.Context, id int64) *redis.IntCmd {
	return nil
}
func (m *MockRedisPipeline) ConfigGet(ctx context.Context, parameter string) *redis.MapStringStringCmd {
	return nil
}
func (m *MockRedisPipeline) ConfigResetStat(ctx context.Context) *redis.StatusCmd         { return nil }
func (m *MockRedisPipeline) ConfigSet(ctx context.Context, parameter, value string) *redis.StatusCmd {
	return nil
}
func (m *MockRedisPipeline) ConfigRewrite(ctx context.Context) *redis.StatusCmd           { return nil }
func (m *MockRedisPipeline) DBSize(ctx context.Context) *redis.IntCmd                     { return nil }
func (m *MockRedisPipeline) FlushAll(ctx context.Context) *redis.StatusCmd                { return nil }
func (m *MockRedisPipeline) FlushAllAsync(ctx context.Context) *redis.StatusCmd           { return nil }
func (m *MockRedisPipeline) FlushDB(ctx context.Context) *redis.StatusCmd                 { return nil }
func (m *MockRedisPipeline) FlushDBAsync(ctx context.Context) *redis.StatusCmd            { return nil }
func (m *MockRedisPipeline) Info(ctx context.Context, section ...string) *redis.StringCmd { return nil }
func (m *MockRedisPipeline) LastSave(ctx context.Context) *redis.IntCmd                   { return nil }
func (m *MockRedisPipeline) Save(ctx context.Context) *redis.StatusCmd                    { return nil }
func (m *MockRedisPipeline) Shutdown(ctx context.Context) *redis.StatusCmd                { return nil }
func (m *MockRedisPipeline) ShutdownSave(ctx context.Context) *redis.StatusCmd            { return nil }
func (m *MockRedisPipeline) ShutdownNoSave(ctx context.Context) *redis.StatusCmd          { return nil }
func (m *MockRedisPipeline) Time(ctx context.Context) *redis.TimeCmd                      { return nil }
func (m *MockRedisPipeline) DebugObject(ctx context.Context, key string) *redis.StringCmd { return nil }
func (m *MockRedisPipeline) ReadOnly(ctx context.Context) *redis.StatusCmd                { return nil }
func (m *MockRedisPipeline) ReadWrite(ctx context.Context) *redis.StatusCmd               { return nil }
func (m *MockRedisPipeline) MemoryUsage(ctx context.Context, key string, samples ...int) *redis.IntCmd {
	return nil
}

func createTestQuotas(maxWorkflows, maxExecPerDay, maxConcurrent, maxAPIPerMin int) json.RawMessage {
	quotas := tenant.TenantQuotas{
		MaxWorkflows:            maxWorkflows,
		MaxExecutionsPerDay:     maxExecPerDay,
		MaxConcurrentExecutions: maxConcurrent,
		MaxAPICallsPerMinute:    maxAPIPerMin,
	}
	data, _ := json.Marshal(quotas)
	return data
}

func addTenantToContext(req *http.Request, t *tenant.Tenant) *http.Request {
	ctx := context.WithValue(req.Context(), TenantContextKey, t)
	return req.WithContext(ctx)
}

func TestNewQuotaChecker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	qc := NewQuotaChecker(nil, nil, logger)

	assert.NotNil(t, qc)
	assert.Nil(t, qc.tenantService)
	assert.Nil(t, qc.redis)
	assert.Equal(t, logger, qc.logger)
}

func TestQuotaChecker_CheckQuotas_NoTenant(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	qc := &QuotaChecker{
		logger: logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := qc.CheckQuotas()
	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "tenant not found")
}

func TestQuotaChecker_CheckQuotas_InvalidQuotasJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	qc := &QuotaChecker{
		logger: logger,
	}

	testTenant := &tenant.Tenant{
		ID:     "tenant-123",
		Quotas: json.RawMessage(`{invalid json}`),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req = addTenantToContext(req, testTenant)
	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := qc.CheckQuotas()
	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "failed to parse quotas")
}

func TestQuotaChecker_DetectOperation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	qc := &QuotaChecker{logger: logger}

	tests := []struct {
		name      string
		method    string
		path      string
		expected  string
	}{
		{
			name:     "POST to workflows creates workflow",
			method:   http.MethodPost,
			path:     "/api/v1/workflows",
			expected: "create_workflow",
		},
		{
			name:     "POST to execute triggers execution",
			method:   http.MethodPost,
			path:     "/api/v1/workflows/123/execute",
			expected: "execute_workflow",
		},
		{
			name:     "GET request is API call",
			method:   http.MethodGet,
			path:     "/api/v1/workflows",
			expected: "api_call",
		},
		{
			name:     "PUT request is API call",
			method:   http.MethodPut,
			path:     "/api/v1/workflows/123",
			expected: "api_call",
		},
		{
			name:     "DELETE request is API call",
			method:   http.MethodDelete,
			path:     "/api/v1/workflows/123",
			expected: "api_call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			result := qc.detectOperation(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuotaChecker_HandleQuotaExceeded(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	qc := &QuotaChecker{logger: logger}

	rr := httptest.NewRecorder()
	err := errors.New("quota exceeded: 10/10 workflows used")

	qc.handleQuotaExceeded(rr, err)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, "3600", rr.Header().Get("Retry-After"))

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "quota_exceeded", response["error"])
	assert.Contains(t, response["message"], "quota exceeded")
}

func TestQuotaExempt(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	var capturedCtx context.Context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	middleware := QuotaExempt()
	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, true, capturedCtx.Value("quota_exempt"))
}

func TestQuotaChecker_CheckWorkflowQuota_Unlimited(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockTenantSvc := new(MockQuotaTenantService)

	qc := &QuotaChecker{
		tenantService: mockTenantSvc,
		logger:        logger,
	}

	quotas := tenant.TenantQuotas{MaxWorkflows: -1}
	err := qc.checkWorkflowQuota(context.Background(), "tenant-123", quotas)

	assert.NoError(t, err)
	mockTenantSvc.AssertNotCalled(t, "GetWorkflowCount")
}

func TestQuotaChecker_CheckWorkflowQuota_UnderLimit(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockTenantSvc := new(MockQuotaTenantService)
	mockTenantSvc.On("GetWorkflowCount", mock.Anything, "tenant-123").Return(5, nil)

	qc := &QuotaChecker{
		tenantService: mockTenantSvc,
		logger:        logger,
	}

	quotas := tenant.TenantQuotas{MaxWorkflows: 10}
	err := qc.checkWorkflowQuota(context.Background(), "tenant-123", quotas)

	assert.NoError(t, err)
	mockTenantSvc.AssertExpectations(t)
}

func TestQuotaChecker_CheckWorkflowQuota_AtLimit(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockTenantSvc := new(MockQuotaTenantService)
	mockTenantSvc.On("GetWorkflowCount", mock.Anything, "tenant-123").Return(10, nil)

	qc := &QuotaChecker{
		tenantService: mockTenantSvc,
		logger:        logger,
	}

	quotas := tenant.TenantQuotas{MaxWorkflows: 10}
	err := qc.checkWorkflowQuota(context.Background(), "tenant-123", quotas)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workflow quota exceeded")
	mockTenantSvc.AssertExpectations(t)
}

func TestQuotaChecker_CheckWorkflowQuota_ServiceError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockTenantSvc := new(MockQuotaTenantService)
	mockTenantSvc.On("GetWorkflowCount", mock.Anything, "tenant-123").Return(0, errors.New("db error"))

	qc := &QuotaChecker{
		tenantService: mockTenantSvc,
		logger:        logger,
	}

	quotas := tenant.TenantQuotas{MaxWorkflows: 10}
	err := qc.checkWorkflowQuota(context.Background(), "tenant-123", quotas)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check quota")
	mockTenantSvc.AssertExpectations(t)
}

func TestQuotaChecker_CheckExecutionQuota_Unlimited(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	qc := &QuotaChecker{
		logger: logger,
	}

	quotas := tenant.TenantQuotas{MaxExecutionsPerDay: -1}
	err := qc.checkExecutionQuota(context.Background(), "tenant-123", quotas)

	assert.NoError(t, err)
}

func TestQuotaChecker_GetConcurrentExecutions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockTenantSvc := new(MockQuotaTenantService)
	mockTenantSvc.On("GetConcurrentExecutions", mock.Anything, "tenant-123").Return(3, nil)

	qc := &QuotaChecker{
		tenantService: mockTenantSvc,
		logger:        logger,
	}

	count, err := qc.getConcurrentExecutions(context.Background(), "tenant-123")

	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	mockTenantSvc.AssertExpectations(t)
}

func TestQuotaChecker_CheckAPIRateLimit_Unlimited(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	qc := &QuotaChecker{
		logger: logger,
	}

	quotas := tenant.TenantQuotas{MaxAPICallsPerMinute: -1}
	err := qc.checkAPIRateLimit(context.Background(), "tenant-123", quotas)

	assert.NoError(t, err)
}
