package google

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestExtractString(t *testing.T) {
	tests := []struct {
		name      string
		data      map[string]interface{}
		path      string
		want      string
		wantErr   bool
		errString string
	}{
		{
			name: "simple key",
			data: map[string]interface{}{
				"key": "value",
			},
			path:    "key",
			want:    "value",
			wantErr: false,
		},
		{
			name: "nested key",
			data: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
			},
			path:    "env.tenant_id",
			want:    "tenant-123",
			wantErr: false,
		},
		{
			name: "deeply nested key",
			data: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": "deep-value",
					},
				},
			},
			path:    "level1.level2.level3",
			want:    "deep-value",
			wantErr: false,
		},
		{
			name: "key not found",
			data: map[string]interface{}{
				"key": "value",
			},
			path:      "missing",
			wantErr:   true,
			errString: "not found",
		},
		{
			name: "not a string",
			data: map[string]interface{}{
				"key": 123,
			},
			path:      "key",
			wantErr:   true,
			errString: "not a string",
		},
		{
			name: "not a map",
			data: map[string]interface{}{
				"key": "value",
			},
			path:      "key.nested",
			wantErr:   true,
			errString: "not a map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractString(tt.data, tt.path)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "single key",
			path: "key",
			want: []string{"key"},
		},
		{
			name: "two keys",
			path: "key1.key2",
			want: []string{"key1", "key2"},
		},
		{
			name: "three keys",
			path: "key1.key2.key3",
			want: []string{"key1", "key2", "key3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCreateOAuth2Token(t *testing.T) {
	tests := []struct {
		name           string
		credentialData map[string]interface{}
		wantErr        bool
		errString      string
		validateToken  func(t *testing.T, token *oauth2.Token)
	}{
		{
			name: "valid OAuth2 token with all fields",
			credentialData: map[string]interface{}{
				"access_token":  "access-123",
				"refresh_token": "refresh-456",
				"token_type":    "Bearer",
			},
			wantErr: false,
			validateToken: func(t *testing.T, token *oauth2.Token) {
				assert.Equal(t, "access-123", token.AccessToken)
				assert.Equal(t, "refresh-456", token.RefreshToken)
				assert.Equal(t, "Bearer", token.TokenType)
			},
		},
		{
			name: "valid OAuth2 token without refresh token",
			credentialData: map[string]interface{}{
				"access_token": "access-123",
			},
			wantErr: false,
			validateToken: func(t *testing.T, token *oauth2.Token) {
				assert.Equal(t, "access-123", token.AccessToken)
				assert.Empty(t, token.RefreshToken)
				assert.Equal(t, "Bearer", token.TokenType)
			},
		},
		{
			name: "missing access token",
			credentialData: map[string]interface{}{
				"refresh_token": "refresh-456",
			},
			wantErr:   true,
			errString: "access_token not found",
		},
		{
			name: "empty access token",
			credentialData: map[string]interface{}{
				"access_token": "",
			},
			wantErr:   true,
			errString: "access_token not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := createOAuth2Token(tt.credentialData)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, token)
				if tt.validateToken != nil {
					tt.validateToken(t, token)
				}
			}
		})
	}
}

func TestCreateServiceAccountClient(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		credentialData map[string]interface{}
		scopes         []string
		wantErr        bool
		errString      string
	}{
		{
			name: "invalid service account JSON",
			credentialData: map[string]interface{}{
				"credential_json": "invalid-json",
			},
			scopes:    []string{"https://www.googleapis.com/auth/gmail.send"},
			wantErr:   true,
			errString: "failed to create credentials",
		},
		{
			name: "missing service account data",
			credentialData: map[string]interface{}{
				"some_other_field": "value",
			},
			scopes:    []string{"https://www.googleapis.com/auth/gmail.send"},
			wantErr:   true,
			errString: "service account credentials not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createServiceAccountClient(ctx, tt.credentialData, tt.scopes)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateOAuth2Client(t *testing.T) {
	ctx := context.Background()
	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
	}

	clientOption := createOAuth2Client(ctx, token)
	assert.NotNil(t, clientOption)
}
