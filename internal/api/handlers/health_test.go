package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDBPinger implements DBPinger for testing
type MockDBPinger struct {
	mock.Mock
}

func (m *MockDBPinger) PingContext(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockRedisPinger implements RedisPinger for testing
type MockRedisPinger struct {
	mock.Mock
}

func (m *MockRedisPinger) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func TestHealthHandler_Health(t *testing.T) {
	handler := &HealthHandler{
		db:    nil, // Not used in Health endpoint
		redis: nil, // Not used in Health endpoint
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.Health(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response HealthResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response.Status)
	assert.NotEmpty(t, response.Timestamp)
}

func TestHealthHandler_Ready(t *testing.T) {
	tests := []struct {
		name           string
		mockDBSetup    func(*MockDBPinger)
		mockRedisSetup func(*MockRedisPinger)
		expectedStatus int
		expectedHealth string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "all healthy",
			mockDBSetup: func(m *MockDBPinger) {
				m.On("PingContext", mock.Anything).Return(nil)
			},
			mockRedisSetup: func(m *MockRedisPinger) {
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetVal("PONG")
				m.On("Ping", mock.Anything).Return(cmd)
			},
			expectedStatus: http.StatusOK,
			expectedHealth: "ok",
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response HealthResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "healthy", response.Checks["database"])
				assert.Equal(t, "healthy", response.Checks["redis"])
			},
		},
		{
			name: "database unhealthy",
			mockDBSetup: func(m *MockDBPinger) {
				m.On("PingContext", mock.Anything).Return(errors.New("connection refused"))
			},
			mockRedisSetup: func(m *MockRedisPinger) {
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetVal("PONG")
				m.On("Ping", mock.Anything).Return(cmd)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "degraded",
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response HealthResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Checks["database"], "unhealthy")
				assert.Contains(t, response.Checks["database"], "connection refused")
				assert.Equal(t, "healthy", response.Checks["redis"])
			},
		},
		{
			name: "redis unhealthy",
			mockDBSetup: func(m *MockDBPinger) {
				m.On("PingContext", mock.Anything).Return(nil)
			},
			mockRedisSetup: func(m *MockRedisPinger) {
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetErr(errors.New("redis connection refused"))
				m.On("Ping", mock.Anything).Return(cmd)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "degraded",
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response HealthResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "healthy", response.Checks["database"])
				assert.Contains(t, response.Checks["redis"], "unhealthy")
				assert.Contains(t, response.Checks["redis"], "redis connection refused")
			},
		},
		{
			name: "both unhealthy",
			mockDBSetup: func(m *MockDBPinger) {
				m.On("PingContext", mock.Anything).Return(errors.New("db error"))
			},
			mockRedisSetup: func(m *MockRedisPinger) {
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetErr(errors.New("redis error"))
				m.On("Ping", mock.Anything).Return(cmd)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "degraded",
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response HealthResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Checks["database"], "unhealthy")
				assert.Contains(t, response.Checks["redis"], "unhealthy")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDBPinger)
			mockRedis := new(MockRedisPinger)

			tt.mockDBSetup(mockDB)
			tt.mockRedisSetup(mockRedis)

			handler := &HealthHandler{
				db:    mockDB,
				redis: mockRedis,
			}

			req := httptest.NewRequest(http.MethodGet, "/ready", nil)
			rr := httptest.NewRecorder()

			handler.Ready(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var response HealthResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedHealth, response.Status)
			assert.NotEmpty(t, response.Timestamp)

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}

			mockDB.AssertExpectations(t)
			mockRedis.AssertExpectations(t)
		})
	}
}

func TestNewHealthHandler(t *testing.T) {
	// NewHealthHandler expects concrete types, but our handler stores interfaces
	// Since we can't easily create nil *sqlx.DB and *redis.Client for testing,
	// we verify the handler is created without panicking
	handler := NewHealthHandler(nil, nil)
	assert.NotNil(t, handler)
	assert.Nil(t, handler.db)
	assert.Nil(t, handler.redis)
}
