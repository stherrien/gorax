package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/integration"
)

func TestAPIKeyAuth(t *testing.T) {
	t.Run("authenticate header", func(t *testing.T) {
		auth := NewAPIKeyAuth("test-api-key", "X-API-Key", APIKeyLocationHeader)
		req, _ := http.NewRequest(http.MethodGet, "https://api.example.com", nil)

		err := auth.Authenticate(req)
		require.NoError(t, err)

		assert.Equal(t, "test-api-key", req.Header.Get("X-API-Key"))
	})

	t.Run("authenticate query", func(t *testing.T) {
		auth := NewAPIKeyAuth("test-api-key", "api_key", APIKeyLocationQuery)
		req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/resource", nil)

		err := auth.Authenticate(req)
		require.NoError(t, err)

		assert.Equal(t, "test-api-key", req.URL.Query().Get("api_key"))
	})

	t.Run("from credentials", func(t *testing.T) {
		creds := &integration.Credentials{
			Type: integration.CredTypeAPIKey,
			Data: integration.JSONMap{
				"key":      "my-key",
				"name":     "X-Custom-Key",
				"location": "header",
			},
		}

		auth, err := NewAPIKeyAuthFromCredentials(creds)
		require.NoError(t, err)

		assert.Equal(t, "my-key", auth.Key())
		assert.Equal(t, "X-Custom-Key", auth.Name())
		assert.Equal(t, APIKeyLocationHeader, auth.Location())
	})

	t.Run("validation error empty key", func(t *testing.T) {
		auth := NewAPIKeyAuth("", "X-API-Key", APIKeyLocationHeader)
		err := auth.Validate()
		assert.Error(t, err)
	})

	t.Run("validation error empty name", func(t *testing.T) {
		auth := &APIKeyAuth{key: "test", name: "", location: APIKeyLocationHeader}
		err := auth.Validate()
		assert.Error(t, err)
	})

	t.Run("validation error invalid location", func(t *testing.T) {
		auth := &APIKeyAuth{key: "test", name: "X-API-Key", location: "invalid"}
		err := auth.Validate()
		assert.Error(t, err)
	})

	t.Run("type returns correct type", func(t *testing.T) {
		auth := NewAPIKeyAuth("test", "X-API-Key", APIKeyLocationHeader)
		assert.Equal(t, integration.CredTypeAPIKey, auth.Type())
	})
}

func TestBearerTokenAuth(t *testing.T) {
	t.Run("authenticate", func(t *testing.T) {
		auth := NewBearerTokenAuth("test-token")
		req, _ := http.NewRequest(http.MethodGet, "https://api.example.com", nil)

		err := auth.Authenticate(req)
		require.NoError(t, err)

		assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
	})

	t.Run("authenticate with custom scheme", func(t *testing.T) {
		auth := NewBearerTokenAuthWithScheme("test-token", "Token")
		req, _ := http.NewRequest(http.MethodGet, "https://api.example.com", nil)

		err := auth.Authenticate(req)
		require.NoError(t, err)

		assert.Equal(t, "Token test-token", req.Header.Get("Authorization"))
	})

	t.Run("from credentials", func(t *testing.T) {
		creds := &integration.Credentials{
			Type: integration.CredTypeBearerToken,
			Data: integration.JSONMap{
				"token":  "my-token",
				"scheme": "JWT",
			},
		}

		auth, err := NewBearerTokenAuthFromCredentials(creds)
		require.NoError(t, err)

		assert.Equal(t, "my-token", auth.Token())
		assert.Equal(t, "JWT", auth.Scheme())
	})

	t.Run("validation error empty token", func(t *testing.T) {
		auth := NewBearerTokenAuth("")
		err := auth.Validate()
		assert.Error(t, err)
	})

	t.Run("type returns correct type", func(t *testing.T) {
		auth := NewBearerTokenAuth("test")
		assert.Equal(t, integration.CredTypeBearerToken, auth.Type())
	})
}

func TestBasicAuth(t *testing.T) {
	t.Run("authenticate", func(t *testing.T) {
		auth := NewBasicAuth("user", "pass")
		req, _ := http.NewRequest(http.MethodGet, "https://api.example.com", nil)

		err := auth.Authenticate(req)
		require.NoError(t, err)

		user, pass, ok := req.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "user", user)
		assert.Equal(t, "pass", pass)
	})

	t.Run("from credentials", func(t *testing.T) {
		creds := &integration.Credentials{
			Type: integration.CredTypeBasicAuth,
			Data: integration.JSONMap{
				"username": "testuser",
				"password": "testpass",
			},
		}

		auth, err := NewBasicAuthFromCredentials(creds)
		require.NoError(t, err)

		assert.Equal(t, "testuser", auth.Username())
	})

	t.Run("validation error empty username", func(t *testing.T) {
		auth := NewBasicAuth("", "pass")
		err := auth.Validate()
		assert.Error(t, err)
	})

	t.Run("empty password is allowed", func(t *testing.T) {
		auth := NewBasicAuth("user", "")
		err := auth.Validate()
		assert.NoError(t, err)
	})

	t.Run("type returns correct type", func(t *testing.T) {
		auth := NewBasicAuth("user", "pass")
		assert.Equal(t, integration.CredTypeBasicAuth, auth.Type())
	})
}

func TestOAuth2Auth(t *testing.T) {
	t.Run("with existing token", func(t *testing.T) {
		config := &OAuth2AuthConfig{
			AccessToken: "existing-token",
		}

		auth, err := NewOAuth2Auth(config)
		require.NoError(t, err)

		req, _ := http.NewRequest(http.MethodGet, "https://api.example.com", nil)
		err = auth.Authenticate(req)
		require.NoError(t, err)

		assert.Equal(t, "Bearer existing-token", req.Header.Get("Authorization"))
	})

	t.Run("validation with token", func(t *testing.T) {
		auth := &OAuth2Auth{accessToken: "token"}
		err := auth.Validate()
		assert.NoError(t, err)
	})

	t.Run("validation without token needs credentials", func(t *testing.T) {
		auth := &OAuth2Auth{
			grantType: GrantTypeClientCredentials,
		}
		err := auth.Validate()
		assert.Error(t, err)
	})

	t.Run("needs refresh", func(t *testing.T) {
		auth := &OAuth2Auth{}
		assert.True(t, auth.NeedsRefresh())

		auth.SetToken("token", 3600)
		assert.False(t, auth.NeedsRefresh())
	})

	t.Run("set refresh token", func(t *testing.T) {
		auth := &OAuth2Auth{}
		auth.SetRefreshToken("refresh-token")
		assert.True(t, auth.HasRefreshToken())
	})

	t.Run("from credentials", func(t *testing.T) {
		creds := &integration.Credentials{
			Type: integration.CredTypeOAuth2,
			Data: integration.JSONMap{
				"client_id":     "client123",
				"client_secret": "secret456",
				"token_url":     "https://auth.example.com/token",
				"grant_type":    "client_credentials",
				"access_token":  "existing-token",
			},
		}

		auth, err := NewOAuth2AuthFromCredentials(creds)
		require.NoError(t, err)
		assert.NotNil(t, auth)
	})

	t.Run("type returns correct type", func(t *testing.T) {
		auth := &OAuth2Auth{}
		assert.Equal(t, integration.CredTypeOAuth2, auth.Type())
	})
}

func TestAuthenticatorFactory(t *testing.T) {
	factory := NewAuthenticatorFactory()

	t.Run("create API key auth", func(t *testing.T) {
		creds := &integration.Credentials{
			Type: integration.CredTypeAPIKey,
			Data: integration.JSONMap{
				"key": "test-key",
			},
		}

		auth, err := factory.Create(creds)
		require.NoError(t, err)
		assert.NotNil(t, auth)
		assert.Equal(t, integration.CredTypeAPIKey, auth.Type())
	})

	t.Run("create bearer token auth", func(t *testing.T) {
		creds := &integration.Credentials{
			Type: integration.CredTypeBearerToken,
			Data: integration.JSONMap{
				"token": "test-token",
			},
		}

		auth, err := factory.Create(creds)
		require.NoError(t, err)
		assert.NotNil(t, auth)
	})

	t.Run("create basic auth", func(t *testing.T) {
		creds := &integration.Credentials{
			Type: integration.CredTypeBasicAuth,
			Data: integration.JSONMap{
				"username": "user",
				"password": "pass",
			},
		}

		auth, err := factory.Create(creds)
		require.NoError(t, err)
		assert.NotNil(t, auth)
	})

	t.Run("create oauth2 auth", func(t *testing.T) {
		creds := &integration.Credentials{
			Type: integration.CredTypeOAuth2,
			Data: integration.JSONMap{
				"access_token": "token",
			},
		}

		auth, err := factory.Create(creds)
		require.NoError(t, err)
		assert.NotNil(t, auth)
	})

	t.Run("nil credentials error", func(t *testing.T) {
		_, err := factory.Create(nil)
		assert.Error(t, err)
	})

	t.Run("unsupported type error", func(t *testing.T) {
		creds := &integration.Credentials{
			Type: integration.CredentialType("unsupported"),
			Data: integration.JSONMap{},
		}

		_, err := factory.Create(creds)
		assert.Error(t, err)
	})

	t.Run("create from config with no credentials", func(t *testing.T) {
		config := &integration.Config{}
		auth, err := factory.CreateFromConfig(config)
		require.NoError(t, err)
		assert.Nil(t, auth)
	})
}

func TestNoAuth(t *testing.T) {
	auth := NewNoAuth()
	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com", nil)

	err := auth.Authenticate(req)
	assert.NoError(t, err)

	assert.Empty(t, req.Header.Get("Authorization"))
	assert.NoError(t, auth.Validate())
	assert.Equal(t, integration.CredTypeCustom, auth.Type())
}
