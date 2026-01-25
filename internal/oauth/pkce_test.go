package oauth

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateState(t *testing.T) {
	// Generate multiple states to ensure uniqueness
	states := make(map[string]bool)
	for range 100 {
		state, err := GenerateState()
		require.NoError(t, err)
		assert.NotEmpty(t, state)

		// Check it's valid base64
		_, err = base64.RawURLEncoding.DecodeString(state)
		require.NoError(t, err)

		// Check uniqueness
		assert.False(t, states[state], "duplicate state generated")
		states[state] = true
	}
}

func TestGeneratePKCEVerifier(t *testing.T) {
	// Generate multiple verifiers to ensure uniqueness
	verifiers := make(map[string]bool)
	for range 100 {
		verifier, err := GeneratePKCEVerifier()
		require.NoError(t, err)
		assert.NotEmpty(t, verifier)

		// Check length (43-128 characters as per RFC 7636)
		assert.GreaterOrEqual(t, len(verifier), 43)
		assert.LessOrEqual(t, len(verifier), 128)

		// Check it's valid base64
		_, err = base64.RawURLEncoding.DecodeString(verifier)
		require.NoError(t, err)

		// Check uniqueness
		assert.False(t, verifiers[verifier], "duplicate verifier generated")
		verifiers[verifier] = true
	}
}

func TestGeneratePKCEChallenge(t *testing.T) {
	tests := []struct {
		name     string
		verifier string
		want     string // Expected challenge for deterministic input
	}{
		{
			name:     "standard verifier",
			verifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			want:     "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		},
		{
			name:     "another verifier",
			verifier: "test-verifier-123",
			want:     "zSNMHiBdtxNs8L9onqEnW-Xq5fuLcM7EksItq1aKBDY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challenge := GeneratePKCEChallenge(tt.verifier)
			assert.Equal(t, tt.want, challenge)

			// Verify it's valid base64
			_, err := base64.RawURLEncoding.DecodeString(challenge)
			require.NoError(t, err)
		})
	}
}

func TestPKCEChallengeConsistency(t *testing.T) {
	verifier, err := GeneratePKCEVerifier()
	require.NoError(t, err)

	// Same verifier should produce same challenge
	challenge1 := GeneratePKCEChallenge(verifier)
	challenge2 := GeneratePKCEChallenge(verifier)
	assert.Equal(t, challenge1, challenge2)
}

func TestPKCEChallengeUniqueness(t *testing.T) {
	// Different verifiers should produce different challenges
	verifier1, err := GeneratePKCEVerifier()
	require.NoError(t, err)

	verifier2, err := GeneratePKCEVerifier()
	require.NoError(t, err)

	challenge1 := GeneratePKCEChallenge(verifier1)
	challenge2 := GeneratePKCEChallenge(verifier2)

	assert.NotEqual(t, challenge1, challenge2)
}
