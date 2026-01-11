package oauth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOAuthConnection_IsExpired(t *testing.T) {
	tests := []struct {
		name        string
		tokenExpiry *time.Time
		want        bool
	}{
		{
			name:        "nil expiry - not expired",
			tokenExpiry: nil,
			want:        false,
		},
		{
			name:        "future expiry - not expired",
			tokenExpiry: timePtr(time.Now().Add(1 * time.Hour)),
			want:        false,
		},
		{
			name:        "past expiry - expired",
			tokenExpiry: timePtr(time.Now().Add(-1 * time.Hour)),
			want:        true,
		},
		{
			name:        "just expired",
			tokenExpiry: timePtr(time.Now().Add(-1 * time.Second)),
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &OAuthConnection{
				TokenExpiry: tt.tokenExpiry,
			}
			got := conn.IsExpired()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOAuthConnection_NeedsRefresh(t *testing.T) {
	tests := []struct {
		name        string
		tokenExpiry *time.Time
		want        bool
	}{
		{
			name:        "nil expiry - no refresh needed",
			tokenExpiry: nil,
			want:        false,
		},
		{
			name:        "expires in 10 minutes - no refresh needed",
			tokenExpiry: timePtr(time.Now().Add(10 * time.Minute)),
			want:        false,
		},
		{
			name:        "expires in 3 minutes - needs refresh",
			tokenExpiry: timePtr(time.Now().Add(3 * time.Minute)),
			want:        true,
		},
		{
			name:        "already expired - needs refresh",
			tokenExpiry: timePtr(time.Now().Add(-1 * time.Hour)),
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &OAuthConnection{
				TokenExpiry: tt.tokenExpiry,
			}
			got := conn.NeedsRefresh()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOAuthState_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "future expiry - not expired",
			expiresAt: time.Now().Add(10 * time.Minute),
			want:      false,
		},
		{
			name:      "past expiry - expired",
			expiresAt: time.Now().Add(-10 * time.Minute),
			want:      true,
		},
		{
			name:      "just expired",
			expiresAt: time.Now().Add(-1 * time.Second),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &OAuthState{
				ExpiresAt: tt.expiresAt,
			}
			got := state.IsExpired()
			assert.Equal(t, tt.want, got)
		})
	}
}

// Helper function
func timePtr(t time.Time) *time.Time {
	return &t
}
