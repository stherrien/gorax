package tenantctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTenantID(t *testing.T) {
	t.Run("sets tenant ID in context", func(t *testing.T) {
		ctx := context.Background()
		tenantID := "tenant-123"

		newCtx := WithTenantID(ctx, tenantID)

		result := GetTenantID(newCtx)
		assert.Equal(t, tenantID, result)
	})

	t.Run("overwrites existing tenant ID", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithTenantID(ctx, "tenant-1")
		ctx = WithTenantID(ctx, "tenant-2")

		result := GetTenantID(ctx)
		assert.Equal(t, "tenant-2", result)
	})
}

func TestGetTenantID(t *testing.T) {
	t.Run("returns empty string when not set", func(t *testing.T) {
		ctx := context.Background()
		result := GetTenantID(ctx)
		assert.Equal(t, "", result)
	})

	t.Run("returns tenant ID when set", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), "tenant-123")
		result := GetTenantID(ctx)
		assert.Equal(t, "tenant-123", result)
	})
}

func TestMustGetTenantID(t *testing.T) {
	t.Run("returns tenant ID when set", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), "tenant-123")
		result, err := MustGetTenantID(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "tenant-123", result)
	})

	t.Run("returns error when not set", func(t *testing.T) {
		ctx := context.Background()
		_, err := MustGetTenantID(ctx)
		assert.Error(t, err)
		assert.Equal(t, ErrNoTenant, err)
	})

	t.Run("returns error for empty tenant ID", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), "")
		_, err := MustGetTenantID(ctx)
		assert.Error(t, err)
		assert.Equal(t, ErrNoTenant, err)
	})
}

func TestWithSwitchedTenant(t *testing.T) {
	t.Run("switches tenant and preserves original", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), "original-tenant")
		newCtx := WithSwitchedTenant(ctx, "target-tenant")

		// Current tenant should be the target
		assert.Equal(t, "target-tenant", GetTenantID(newCtx))

		// Original tenant should be preserved
		assert.Equal(t, "original-tenant", GetOriginalTenantID(newCtx))

		// Should be marked as switched
		assert.True(t, IsTenantSwitched(newCtx))
	})

	t.Run("handles empty original tenant", func(t *testing.T) {
		ctx := context.Background()
		newCtx := WithSwitchedTenant(ctx, "target-tenant")

		assert.Equal(t, "target-tenant", GetTenantID(newCtx))
		assert.Equal(t, "", GetOriginalTenantID(newCtx))
		assert.True(t, IsTenantSwitched(newCtx))
	})
}

func TestGetOriginalTenantID(t *testing.T) {
	t.Run("returns empty when not switched", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), "tenant-123")
		result := GetOriginalTenantID(ctx)
		assert.Equal(t, "", result)
	})

	t.Run("returns original after switch", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), "original")
		ctx = WithSwitchedTenant(ctx, "target")
		result := GetOriginalTenantID(ctx)
		assert.Equal(t, "original", result)
	})
}

func TestIsTenantSwitched(t *testing.T) {
	t.Run("returns false when not switched", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), "tenant-123")
		assert.False(t, IsTenantSwitched(ctx))
	})

	t.Run("returns true when switched", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), "original")
		ctx = WithSwitchedTenant(ctx, "target")
		assert.True(t, IsTenantSwitched(ctx))
	})

	t.Run("returns false for empty context", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, IsTenantSwitched(ctx))
	})
}

func TestResetToOriginalTenant(t *testing.T) {
	t.Run("resets to original tenant", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), "original")
		ctx = WithSwitchedTenant(ctx, "target")

		assert.Equal(t, "target", GetTenantID(ctx))
		assert.True(t, IsTenantSwitched(ctx))

		resetCtx := ResetToOriginalTenant(ctx)

		assert.Equal(t, "original", GetTenantID(resetCtx))
		assert.False(t, IsTenantSwitched(resetCtx))
	})

	t.Run("no-op when not switched", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), "tenant-123")
		resetCtx := ResetToOriginalTenant(ctx)

		assert.Equal(t, "tenant-123", GetTenantID(resetCtx))
		assert.False(t, IsTenantSwitched(resetCtx))
	})

	t.Run("handles empty original", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithSwitchedTenant(ctx, "target")

		resetCtx := ResetToOriginalTenant(ctx)

		// Since original was empty, the reset returns unchanged context
		assert.Equal(t, "target", GetTenantID(resetCtx))
	})
}

func TestContextChaining(t *testing.T) {
	t.Run("multiple operations chain correctly", func(t *testing.T) {
		// Start with base context
		ctx := context.Background()
		assert.Equal(t, "", GetTenantID(ctx))

		// Set initial tenant
		ctx = WithTenantID(ctx, "tenant-1")
		assert.Equal(t, "tenant-1", GetTenantID(ctx))
		assert.False(t, IsTenantSwitched(ctx))

		// Switch to another tenant
		ctx = WithSwitchedTenant(ctx, "tenant-2")
		assert.Equal(t, "tenant-2", GetTenantID(ctx))
		assert.Equal(t, "tenant-1", GetOriginalTenantID(ctx))
		assert.True(t, IsTenantSwitched(ctx))

		// Reset to original
		ctx = ResetToOriginalTenant(ctx)
		assert.Equal(t, "tenant-1", GetTenantID(ctx))
		assert.False(t, IsTenantSwitched(ctx))
	})
}
