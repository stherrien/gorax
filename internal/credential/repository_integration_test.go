package credential

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests require a test database. They will be skipped if TEST_DATABASE_URL is not set.
// To run: TEST_DATABASE_URL="postgres://user:pass@localhost/test_db?sslmode=disable" go test -v

// Test UUIDs for user IDs (created_by column requires UUID format)
const (
	testUserID  = "00000000-0000-0000-0000-000000000001"
	testAdminID = "00000000-0000-0000-0000-000000000002"
)

func getTestDatabaseURL() string {
	return os.Getenv("TEST_DATABASE_URL")
}

func setupTestDB(t *testing.T) *sqlx.DB {
	dbURL := getTestDatabaseURL()
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sqlx.Connect("postgres", dbURL)
	require.NoError(t, err)

	// Clean up credential tables
	_, err = db.Exec("DELETE FROM credential_access_log")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM credential_rotations")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM credentials")
	require.NoError(t, err)

	return db
}

func createTestTenant(t *testing.T, db *sqlx.DB) string {
	tenantID := uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO tenants (id, name, subdomain, status, tier, settings, quotas, created_at, updated_at)
		VALUES ($1, $2, $3, 'active', 'free', '{}', '{}', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, tenantID, "Test Tenant "+tenantID[:8], "test-"+tenantID[:8])
	require.NoError(t, err)
	return tenantID
}

func createTestCredential(t *testing.T, repo *Repository, ctx context.Context, tenantID string, credType CredentialType, name string) *Credential {
	cred := &Credential{
		Name:         name,
		Type:         credType,
		Description:  "Test credential of type " + string(credType),
		Status:       StatusActive,
		EncryptedDEK: []byte("test-encrypted-dek"),
		Ciphertext:   []byte("test-ciphertext"),
		Nonce:        []byte("test-nonce12"),
		AuthTag:      []byte("test-auth-tag123"),
		KMSKeyID:     "test-kms-key",
		Metadata:     map[string]interface{}{"test": true},
	}

	created, err := repo.Create(ctx, tenantID, testUserID, cred)
	require.NoError(t, err)
	return created
}

// =============================================================================
// List Tests
// =============================================================================

func TestRepository_List_EmptyResult(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	t.Run("list returns empty array when no credentials exist", func(t *testing.T) {
		credentials, err := repo.List(ctx, tenantID, CredentialListFilter{})
		require.NoError(t, err)
		assert.NotNil(t, credentials)
		assert.Len(t, credentials, 0)
		assert.Equal(t, []*Credential{}, credentials)
	})

	t.Run("list with filters returns empty array", func(t *testing.T) {
		credentials, err := repo.List(ctx, tenantID, CredentialListFilter{
			Type:   TypeAPIKey,
			Status: StatusActive,
			Search: "nonexistent",
		})
		require.NoError(t, err)
		assert.NotNil(t, credentials)
		assert.Len(t, credentials, 0)
	})
}

func TestRepository_List_WithCredentials(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	// Create test credentials
	cred1 := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "api-key-1")
	cred2 := createTestCredential(t, repo, ctx, tenantID, TypeOAuth2, "oauth-app")
	cred3 := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "api-key-2")

	t.Run("list all credentials", func(t *testing.T) {
		credentials, err := repo.List(ctx, tenantID, CredentialListFilter{})
		require.NoError(t, err)
		assert.Len(t, credentials, 3)
	})

	t.Run("filter by type", func(t *testing.T) {
		credentials, err := repo.List(ctx, tenantID, CredentialListFilter{
			Type: TypeAPIKey,
		})
		require.NoError(t, err)
		assert.Len(t, credentials, 2)
		for _, c := range credentials {
			assert.Equal(t, TypeAPIKey, c.Type)
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		credentials, err := repo.List(ctx, tenantID, CredentialListFilter{
			Status: StatusActive,
		})
		require.NoError(t, err)
		assert.Len(t, credentials, 3)
	})

	t.Run("filter by search term", func(t *testing.T) {
		credentials, err := repo.List(ctx, tenantID, CredentialListFilter{
			Search: "oauth",
		})
		require.NoError(t, err)
		assert.Len(t, credentials, 1)
		assert.Equal(t, cred2.ID, credentials[0].ID)
	})

	t.Run("combined filters", func(t *testing.T) {
		credentials, err := repo.List(ctx, tenantID, CredentialListFilter{
			Type:   TypeAPIKey,
			Search: "api-key-1",
		})
		require.NoError(t, err)
		assert.Len(t, credentials, 1)
		assert.Equal(t, cred1.ID, credentials[0].ID)
	})

	t.Run("ordered by created_at DESC", func(t *testing.T) {
		credentials, err := repo.List(ctx, tenantID, CredentialListFilter{})
		require.NoError(t, err)
		assert.Len(t, credentials, 3)
		// cred3 was created last, should be first
		assert.Equal(t, cred3.ID, credentials[0].ID)
	})
}

func TestRepository_List_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	tenant1 := createTestTenant(t, db)
	tenant2 := createTestTenant(t, db)

	// Create credentials for tenant1
	createTestCredential(t, repo, ctx, tenant1, TypeAPIKey, "tenant1-cred")

	// Create credentials for tenant2
	createTestCredential(t, repo, ctx, tenant2, TypeAPIKey, "tenant2-cred")

	t.Run("tenant1 only sees own credentials", func(t *testing.T) {
		credentials, err := repo.List(ctx, tenant1, CredentialListFilter{})
		require.NoError(t, err)
		assert.Len(t, credentials, 1)
		assert.Equal(t, "tenant1-cred", credentials[0].Name)
	})

	t.Run("tenant2 only sees own credentials", func(t *testing.T) {
		credentials, err := repo.List(ctx, tenant2, CredentialListFilter{})
		require.NoError(t, err)
		assert.Len(t, credentials, 1)
		assert.Equal(t, "tenant2-cred", credentials[0].Name)
	})
}

// =============================================================================
// Create Tests
// =============================================================================

func TestRepository_Create_AllTypes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	credentialTypes := []struct {
		name     string
		credType CredentialType
	}{
		{"api_key credential", TypeAPIKey},
		{"oauth2 credential", TypeOAuth2},
		{"basic_auth credential", TypeBasicAuth},
		{"bearer_token credential", TypeBearerToken},
		{"custom credential", TypeCustom},
	}

	for _, tt := range credentialTypes {
		t.Run(tt.name, func(t *testing.T) {
			cred := &Credential{
				Name:         tt.name,
				Type:         tt.credType,
				Description:  "Test " + tt.name,
				Status:       StatusActive,
				EncryptedDEK: []byte("encrypted-dek-" + string(tt.credType)),
				Ciphertext:   []byte("ciphertext-" + string(tt.credType)),
				Nonce:        []byte("nonce123456"),
				AuthTag:      []byte("authtag1234567"),
				KMSKeyID:     "test-kms-key",
				Metadata:     map[string]interface{}{"type": string(tt.credType)},
			}

			created, err := repo.Create(ctx, tenantID, testUserID, cred)
			require.NoError(t, err)
			assert.NotEmpty(t, created.ID)
			assert.Equal(t, tenantID, created.TenantID)
			assert.Equal(t, tt.name, created.Name)
			assert.Equal(t, tt.credType, created.Type)
			assert.Equal(t, StatusActive, created.Status)
			assert.NotZero(t, created.CreatedAt)
			assert.NotZero(t, created.UpdatedAt)
		})
	}
}

func TestRepository_Create_Validation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	t.Run("empty tenant_id returns error", func(t *testing.T) {
		cred := &Credential{
			Name:         "test",
			Type:         TypeAPIKey,
			EncryptedDEK: []byte("test"),
			Ciphertext:   []byte("test"),
			Nonce:        []byte("test"),
			AuthTag:      []byte("test"),
			KMSKeyID:     "test",
		}
		_, err := repo.Create(ctx, "", testUserID, cred)
		assert.ErrorIs(t, err, ErrInvalidTenantID)
	})

	t.Run("empty created_by returns error", func(t *testing.T) {
		cred := &Credential{
			Name:         "test",
			Type:         TypeAPIKey,
			EncryptedDEK: []byte("test"),
			Ciphertext:   []byte("test"),
			Nonce:        []byte("test"),
			AuthTag:      []byte("test"),
			KMSKeyID:     "test",
		}
		_, err := repo.Create(ctx, tenantID, "", cred)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "created_by cannot be empty")
	})

	t.Run("nil credential returns error", func(t *testing.T) {
		_, err := repo.Create(ctx, tenantID, testUserID, nil)
		assert.ErrorIs(t, err, ErrEmptyCredentialData)
	})

	t.Run("empty name returns error", func(t *testing.T) {
		cred := &Credential{
			Name:         "",
			Type:         TypeAPIKey,
			EncryptedDEK: []byte("test"),
			Ciphertext:   []byte("test"),
			Nonce:        []byte("test"),
			AuthTag:      []byte("test"),
			KMSKeyID:     "test",
		}
		_, err := repo.Create(ctx, tenantID, testUserID, cred)
		assert.ErrorIs(t, err, ErrInvalidCredentialName)
	})

	t.Run("invalid type returns error", func(t *testing.T) {
		cred := &Credential{
			Name:         "test",
			Type:         CredentialType("invalid_type"),
			EncryptedDEK: []byte("test"),
			Ciphertext:   []byte("test"),
			Nonce:        []byte("test"),
			AuthTag:      []byte("test"),
			KMSKeyID:     "test",
		}
		_, err := repo.Create(ctx, tenantID, testUserID, cred)
		assert.ErrorIs(t, err, ErrInvalidCredentialType)
	})

	t.Run("duplicate name returns error", func(t *testing.T) {
		cred := &Credential{
			Name:         "duplicate-name",
			Type:         TypeAPIKey,
			EncryptedDEK: []byte("test"),
			Ciphertext:   []byte("test"),
			Nonce:        []byte("test"),
			AuthTag:      []byte("test"),
			KMSKeyID:     "test",
		}
		_, err := repo.Create(ctx, tenantID, testUserID, cred)
		require.NoError(t, err)

		// Try to create another with same name
		_, err = repo.Create(ctx, tenantID, testUserID, cred)
		assert.ErrorIs(t, err, ErrDuplicateCredential)
	})
}

func TestRepository_Create_GeneratesID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	cred := &Credential{
		Name:         "test-id-gen",
		Type:         TypeAPIKey,
		EncryptedDEK: []byte("test"),
		Ciphertext:   []byte("test"),
		Nonce:        []byte("test"),
		AuthTag:      []byte("test"),
		KMSKeyID:     "test",
	}

	created, err := repo.Create(ctx, tenantID, testUserID, cred)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)

	// Verify UUID format
	_, err = uuid.Parse(created.ID)
	assert.NoError(t, err)
}

// =============================================================================
// GetByID Tests
// =============================================================================

func TestRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	created := createTestCredential(t, repo, ctx, tenantID, TypeBearerToken, "bearer-token-cred")

	t.Run("get existing credential", func(t *testing.T) {
		cred, err := repo.GetByID(ctx, tenantID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, cred.ID)
		assert.Equal(t, created.Name, cred.Name)
		assert.Equal(t, TypeBearerToken, cred.Type)
	})

	t.Run("get non-existent credential", func(t *testing.T) {
		_, err := repo.GetByID(ctx, tenantID, uuid.New().String())
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("empty tenant_id returns error", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "", created.ID)
		assert.ErrorIs(t, err, ErrInvalidTenantID)
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := repo.GetByID(ctx, tenantID, "")
		assert.ErrorIs(t, err, ErrInvalidCredentialID)
	})

	t.Run("tenant isolation - cannot get other tenant credentials", func(t *testing.T) {
		otherTenant := createTestTenant(t, db)
		_, err := repo.GetByID(ctx, otherTenant, created.ID)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

// =============================================================================
// GetByName Tests
// =============================================================================

func TestRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	created := createTestCredential(t, repo, ctx, tenantID, TypeOAuth2, "my-oauth-credential")

	t.Run("get existing credential by name", func(t *testing.T) {
		cred, err := repo.GetByName(ctx, tenantID, "my-oauth-credential")
		require.NoError(t, err)
		assert.Equal(t, created.ID, cred.ID)
		assert.Equal(t, TypeOAuth2, cred.Type)
	})

	t.Run("get non-existent credential", func(t *testing.T) {
		_, err := repo.GetByName(ctx, tenantID, "nonexistent")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("empty tenant_id returns error", func(t *testing.T) {
		_, err := repo.GetByName(ctx, "", "my-oauth-credential")
		assert.ErrorIs(t, err, ErrInvalidTenantID)
	})

	t.Run("empty name returns error", func(t *testing.T) {
		_, err := repo.GetByName(ctx, tenantID, "")
		assert.ErrorIs(t, err, ErrInvalidCredentialName)
	})

	t.Run("tenant isolation - cannot get other tenant credentials", func(t *testing.T) {
		otherTenant := createTestTenant(t, db)
		_, err := repo.GetByName(ctx, otherTenant, "my-oauth-credential")
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

// =============================================================================
// Update Tests
// =============================================================================

func TestRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	t.Run("update name and description", func(t *testing.T) {
		cred := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "original-name")

		newName := "updated-name"
		newDesc := "updated description"
		input := &UpdateCredentialInput{
			Name:        &newName,
			Description: &newDesc,
		}

		updated, err := repo.Update(ctx, tenantID, cred.ID, input)
		require.NoError(t, err)
		assert.Equal(t, "updated-name", updated.Name)
		assert.Equal(t, "updated description", updated.Description)
		assert.True(t, updated.UpdatedAt.After(cred.UpdatedAt))
	})

	t.Run("update status", func(t *testing.T) {
		cred := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "status-test")

		newStatus := StatusInactive
		input := &UpdateCredentialInput{
			Status: &newStatus,
		}

		updated, err := repo.Update(ctx, tenantID, cred.ID, input)
		require.NoError(t, err)
		assert.Equal(t, StatusInactive, updated.Status)
	})

	t.Run("update metadata", func(t *testing.T) {
		cred := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "metadata-test")

		input := &UpdateCredentialInput{
			Metadata: map[string]interface{}{"key": "value", "nested": map[string]interface{}{"foo": "bar"}},
		}

		updated, err := repo.Update(ctx, tenantID, cred.ID, input)
		require.NoError(t, err)
		assert.Equal(t, "value", updated.Metadata["key"])
	})

	t.Run("update non-existent credential", func(t *testing.T) {
		newName := "test"
		input := &UpdateCredentialInput{Name: &newName}
		_, err := repo.Update(ctx, tenantID, uuid.New().String(), input)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("empty update input", func(t *testing.T) {
		cred := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "empty-update-test")
		input := &UpdateCredentialInput{}
		_, err := repo.Update(ctx, tenantID, cred.ID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no fields to update")
	})

	t.Run("tenant isolation", func(t *testing.T) {
		cred := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "isolation-test")
		otherTenant := createTestTenant(t, db)

		newName := "hacked"
		input := &UpdateCredentialInput{Name: &newName}
		_, err := repo.Update(ctx, otherTenant, cred.ID, input)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	t.Run("delete existing credential", func(t *testing.T) {
		cred := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "to-delete")

		err := repo.Delete(ctx, tenantID, cred.ID)
		require.NoError(t, err)

		// Verify deleted
		_, err = repo.GetByID(ctx, tenantID, cred.ID)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("delete non-existent credential", func(t *testing.T) {
		err := repo.Delete(ctx, tenantID, uuid.New().String())
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("empty tenant_id returns error", func(t *testing.T) {
		err := repo.Delete(ctx, "", uuid.New().String())
		assert.ErrorIs(t, err, ErrInvalidTenantID)
	})

	t.Run("empty id returns error", func(t *testing.T) {
		err := repo.Delete(ctx, tenantID, "")
		assert.ErrorIs(t, err, ErrInvalidCredentialID)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		cred := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "no-delete")
		otherTenant := createTestTenant(t, db)

		err := repo.Delete(ctx, otherTenant, cred.ID)
		assert.ErrorIs(t, err, ErrNotFound)

		// Verify still exists
		_, err = repo.GetByID(ctx, tenantID, cred.ID)
		require.NoError(t, err)
	})
}

// =============================================================================
// UpdateLastUsedAt Tests
// =============================================================================

func TestRepository_UpdateLastUsedAt(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	t.Run("update last_used_at", func(t *testing.T) {
		cred := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "last-used-test")
		assert.Nil(t, cred.LastUsedAt)

		err := repo.UpdateLastUsedAt(ctx, tenantID, cred.ID)
		require.NoError(t, err)

		// Verify updated
		updated, err := repo.GetByID(ctx, tenantID, cred.ID)
		require.NoError(t, err)
		assert.NotNil(t, updated.LastUsedAt)
		assert.WithinDuration(t, time.Now(), *updated.LastUsedAt, 5*time.Second)
	})

	t.Run("update non-existent credential", func(t *testing.T) {
		err := repo.UpdateLastUsedAt(ctx, tenantID, uuid.New().String())
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("empty tenant_id returns error", func(t *testing.T) {
		err := repo.UpdateLastUsedAt(ctx, "", uuid.New().String())
		assert.ErrorIs(t, err, ErrInvalidTenantID)
	})

	t.Run("empty id returns error", func(t *testing.T) {
		err := repo.UpdateLastUsedAt(ctx, tenantID, "")
		assert.ErrorIs(t, err, ErrInvalidCredentialID)
	})
}

// =============================================================================
// LogAccess Tests
// =============================================================================

func TestRepository_LogAccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	cred := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "access-log-test")

	t.Run("log successful access", func(t *testing.T) {
		log := &AccessLog{
			CredentialID: cred.ID,
			TenantID:     tenantID,
			AccessedBy:   testUserID,
			AccessType:   AccessTypeRead,
			IPAddress:    "192.168.1.1",
			UserAgent:    "TestClient/1.0",
			Success:      true,
		}

		err := repo.LogAccess(ctx, log)
		require.NoError(t, err)
		assert.NotEmpty(t, log.ID)
	})

	t.Run("log failed access", func(t *testing.T) {
		log := &AccessLog{
			CredentialID: cred.ID,
			TenantID:     tenantID,
			AccessedBy:   testUserID,
			AccessType:   AccessTypeRead,
			Success:      false,
			ErrorMessage: "permission denied",
		}

		err := repo.LogAccess(ctx, log)
		require.NoError(t, err)
	})

	t.Run("log rotate access", func(t *testing.T) {
		log := &AccessLog{
			CredentialID: cred.ID,
			TenantID:     tenantID,
			AccessedBy:   testAdminID,
			AccessType:   AccessTypeRotate,
			Success:      true,
		}

		err := repo.LogAccess(ctx, log)
		require.NoError(t, err)
	})

	t.Run("nil log returns error", func(t *testing.T) {
		err := repo.LogAccess(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("empty credential_id returns error", func(t *testing.T) {
		log := &AccessLog{
			CredentialID: "",
			TenantID:     tenantID,
			AccessedBy:   testUserID,
			AccessType:   AccessTypeRead,
			Success:      true,
		}
		err := repo.LogAccess(ctx, log)
		assert.ErrorIs(t, err, ErrInvalidCredentialID)
	})

	t.Run("empty tenant_id returns error", func(t *testing.T) {
		log := &AccessLog{
			CredentialID: cred.ID,
			TenantID:     "",
			AccessedBy:   testUserID,
			AccessType:   AccessTypeRead,
			Success:      true,
		}
		err := repo.LogAccess(ctx, log)
		assert.ErrorIs(t, err, ErrInvalidTenantID)
	})

	t.Run("empty accessed_by returns error", func(t *testing.T) {
		log := &AccessLog{
			CredentialID: cred.ID,
			TenantID:     tenantID,
			AccessedBy:   "",
			AccessType:   AccessTypeRead,
			Success:      true,
		}
		err := repo.LogAccess(ctx, log)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "accessed_by cannot be empty")
	})
}

// =============================================================================
// GetAccessLogs Tests
// =============================================================================

func TestRepository_GetAccessLogs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	cred := createTestCredential(t, repo, ctx, tenantID, TypeAPIKey, "access-logs-get-test")

	// Create some access logs
	for i := 0; i < 5; i++ {
		log := &AccessLog{
			CredentialID: cred.ID,
			TenantID:     tenantID,
			AccessedBy:   testUserID,
			AccessType:   AccessTypeRead,
			Success:      true,
		}
		err := repo.LogAccess(ctx, log)
		require.NoError(t, err)
	}

	t.Run("get access logs", func(t *testing.T) {
		logs, err := repo.GetAccessLogs(ctx, cred.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, logs, 5)
	})

	t.Run("get access logs with limit", func(t *testing.T) {
		logs, err := repo.GetAccessLogs(ctx, cred.ID, 2, 0)
		require.NoError(t, err)
		assert.Len(t, logs, 2)
	})

	t.Run("get access logs with offset", func(t *testing.T) {
		logs, err := repo.GetAccessLogs(ctx, cred.ID, 10, 3)
		require.NoError(t, err)
		assert.Len(t, logs, 2)
	})

	t.Run("empty credential_id returns error", func(t *testing.T) {
		_, err := repo.GetAccessLogs(ctx, "", 10, 0)
		assert.ErrorIs(t, err, ErrInvalidCredentialID)
	})

	t.Run("default limit applied", func(t *testing.T) {
		logs, err := repo.GetAccessLogs(ctx, cred.ID, 0, 0)
		require.NoError(t, err)
		assert.Len(t, logs, 5) // Default limit is 50, but we only have 5
	})

	t.Run("max limit enforced", func(t *testing.T) {
		logs, err := repo.GetAccessLogs(ctx, cred.ID, 1000, 0)
		require.NoError(t, err)
		// Max limit is 100, but we only have 5
		assert.Len(t, logs, 5)
	})

	t.Run("ordered by accessed_at DESC", func(t *testing.T) {
		logs, err := repo.GetAccessLogs(ctx, cred.ID, 10, 0)
		require.NoError(t, err)

		for i := 1; i < len(logs); i++ {
			assert.True(t, logs[i-1].AccessedAt.After(logs[i].AccessedAt) ||
				logs[i-1].AccessedAt.Equal(logs[i].AccessedAt))
		}
	})
}

// =============================================================================
// Full Workflow Integration Tests
// =============================================================================

func TestRepository_FullWorkflow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	t.Run("complete credential lifecycle", func(t *testing.T) {
		// 1. Create
		cred := &Credential{
			Name:         "lifecycle-test",
			Type:         TypeBearerToken,
			Description:  "Test bearer token",
			Status:       StatusActive,
			EncryptedDEK: []byte("encrypted-dek"),
			Ciphertext:   []byte("ciphertext"),
			Nonce:        []byte("nonce12bytes"),
			AuthTag:      []byte("authtag16bytes!"),
			KMSKeyID:     "test-kms-key",
			Metadata:     map[string]interface{}{"env": "test"},
		}

		created, err := repo.Create(ctx, tenantID, testAdminID, cred)
		require.NoError(t, err)
		t.Logf("✓ Created credential: %s", created.ID)

		// 2. Get by ID
		fetched, err := repo.GetByID(ctx, tenantID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, fetched.ID)
		t.Log("✓ Retrieved credential by ID")

		// 3. Get by name
		fetchedByName, err := repo.GetByName(ctx, tenantID, "lifecycle-test")
		require.NoError(t, err)
		assert.Equal(t, created.ID, fetchedByName.ID)
		t.Log("✓ Retrieved credential by name")

		// 4. List
		credentials, err := repo.List(ctx, tenantID, CredentialListFilter{})
		require.NoError(t, err)
		assert.Len(t, credentials, 1)
		t.Log("✓ Listed credentials")

		// 5. Update
		newDesc := "Updated description"
		updated, err := repo.Update(ctx, tenantID, created.ID, &UpdateCredentialInput{
			Description: &newDesc,
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated description", updated.Description)
		t.Log("✓ Updated credential")

		// 6. Update last used
		err = repo.UpdateLastUsedAt(ctx, tenantID, created.ID)
		require.NoError(t, err)
		t.Log("✓ Updated last_used_at")

		// 7. Log access
		err = repo.LogAccess(ctx, &AccessLog{
			CredentialID: created.ID,
			TenantID:     tenantID,
			AccessedBy:   testUserID,
			AccessType:   AccessTypeRead,
			Success:      true,
			IPAddress:    "10.0.0.1",
		})
		require.NoError(t, err)
		t.Log("✓ Logged access")

		// 8. Get access logs
		logs, err := repo.GetAccessLogs(ctx, created.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, logs, 1)
		t.Log("✓ Retrieved access logs")

		// 9. Delete
		err = repo.Delete(ctx, tenantID, created.ID)
		require.NoError(t, err)
		t.Log("✓ Deleted credential")

		// 10. Verify deleted
		_, err = repo.GetByID(ctx, tenantID, created.ID)
		assert.ErrorIs(t, err, ErrNotFound)
		t.Log("✓ Verified credential deleted")

		t.Log("\n✅ Complete lifecycle test passed!")
	})
}

// =============================================================================
// Credential Type Specific Tests
// =============================================================================

func TestRepository_CredentialTypes_APIKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	cred := &Credential{
		Name:         "openai-api-key",
		Type:         TypeAPIKey,
		Description:  "OpenAI API Key for GPT-4",
		Status:       StatusActive,
		EncryptedDEK: []byte("enc-dek-apikey"),
		Ciphertext:   []byte("enc-apikey-value"),
		Nonce:        []byte("apikey-nonce"),
		AuthTag:      []byte("apikey-authtag!"),
		KMSKeyID:     "aws-kms-key-id",
		Metadata: map[string]interface{}{
			"provider": "openai",
			"model":    "gpt-4",
		},
	}

	created, err := repo.Create(ctx, tenantID, testUserID, cred)
	require.NoError(t, err)
	assert.Equal(t, TypeAPIKey, created.Type)
	assert.Equal(t, "openai", created.Metadata["provider"])
}

func TestRepository_CredentialTypes_OAuth2(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	cred := &Credential{
		Name:         "google-oauth",
		Type:         TypeOAuth2,
		Description:  "Google OAuth2 credentials",
		Status:       StatusActive,
		EncryptedDEK: []byte("enc-dek-oauth"),
		Ciphertext:   []byte("enc-oauth-value"),
		Nonce:        []byte("oauth-nonce!"),
		AuthTag:      []byte("oauth-authtag!!"),
		KMSKeyID:     "aws-kms-key-id",
		Metadata: map[string]interface{}{
			"provider":  "google",
			"scopes":    []string{"email", "profile"},
			"token_url": "https://oauth2.googleapis.com/token",
		},
	}

	created, err := repo.Create(ctx, tenantID, testUserID, cred)
	require.NoError(t, err)
	assert.Equal(t, TypeOAuth2, created.Type)
}

func TestRepository_CredentialTypes_BasicAuth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	cred := &Credential{
		Name:         "jenkins-basic-auth",
		Type:         TypeBasicAuth,
		Description:  "Jenkins CI server credentials",
		Status:       StatusActive,
		EncryptedDEK: []byte("enc-dek-basic"),
		Ciphertext:   []byte("enc-basic-value"),
		Nonce:        []byte("basic-nonce!"),
		AuthTag:      []byte("basic-authtag!!"),
		KMSKeyID:     "aws-kms-key-id",
		Metadata: map[string]interface{}{
			"server": "https://jenkins.example.com",
		},
	}

	created, err := repo.Create(ctx, tenantID, testUserID, cred)
	require.NoError(t, err)
	assert.Equal(t, TypeBasicAuth, created.Type)
}

func TestRepository_CredentialTypes_BearerToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	cred := &Credential{
		Name:         "github-pat",
		Type:         TypeBearerToken,
		Description:  "GitHub Personal Access Token",
		Status:       StatusActive,
		EncryptedDEK: []byte("enc-dek-bearer"),
		Ciphertext:   []byte("enc-bearer-value"),
		Nonce:        []byte("bearer-nonce"),
		AuthTag:      []byte("bearer-authtag!"),
		KMSKeyID:     "aws-kms-key-id",
		Metadata: map[string]interface{}{
			"scopes":     []string{"repo", "user"},
			"expires_at": "2025-12-31",
		},
	}

	created, err := repo.Create(ctx, tenantID, testUserID, cred)
	require.NoError(t, err)
	assert.Equal(t, TypeBearerToken, created.Type)
}

func TestRepository_CredentialTypes_Custom(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := createTestTenant(t, db)

	cred := &Credential{
		Name:         "custom-ssh-key",
		Type:         TypeCustom,
		Description:  "SSH private key for deployment",
		Status:       StatusActive,
		EncryptedDEK: []byte("enc-dek-custom"),
		Ciphertext:   []byte("enc-custom-value"),
		Nonce:        []byte("custom-nonce"),
		AuthTag:      []byte("custom-authtag!"),
		KMSKeyID:     "aws-kms-key-id",
		Metadata: map[string]interface{}{
			"key_type":    "ssh-rsa",
			"fingerprint": "SHA256:abc123",
		},
	}

	created, err := repo.Create(ctx, tenantID, testUserID, cred)
	require.NoError(t, err)
	assert.Equal(t, TypeCustom, created.Type)
}
