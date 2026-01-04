package sso

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_CreateProvider(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	ctx := context.Background()
	tenantID := uuid.New()
	createdBy := uuid.New()

	config := SAMLConfig{
		EntityID:    "https://app.gorax.io",
		ACSURL:      "https://app.gorax.io/sso/acs",
		IdPEntityID: "https://idp.example.com",
	}
	configJSON, _ := json.Marshal(config)

	provider := &Provider{
		TenantID:   tenantID,
		Name:       "Test SSO",
		Type:       ProviderTypeSAML,
		Enabled:    true,
		EnforceSSO: false,
		Config:     configJSON,
		Domains:    []string{"example.com"},
		CreatedBy:  &createdBy,
		UpdatedBy:  &createdBy,
	}

	now := time.Now()

	mock.ExpectQuery(`INSERT INTO sso_providers`).
		WithArgs(
			sqlmock.AnyArg(),
			tenantID,
			"Test SSO",
			ProviderTypeSAML,
			true,
			false,
			configJSON,
			pq.Array([]string{"example.com"}),
			createdBy,
			createdBy,
		).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
			AddRow(now, now))

	err = repo.CreateProvider(ctx, provider)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, provider.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetProvider(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	ctx := context.Background()
	providerID := uuid.New()
	tenantID := uuid.New()

	config := SAMLConfig{EntityID: "https://app.gorax.io"}
	configJSON, _ := json.Marshal(config)

	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM sso_providers WHERE id = \$1`).
		WithArgs(providerID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "name", "provider_type", "enabled", "enforce_sso",
			"config", "domains", "created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			providerID, tenantID, "Test SSO", ProviderTypeSAML, true, false,
			configJSON, pq.Array([]string{"example.com"}), now, now, nil, nil,
		))

	provider, err := repo.GetProvider(ctx, providerID)
	assert.NoError(t, err)
	assert.Equal(t, providerID, provider.ID)
	assert.Equal(t, "Test SSO", provider.Name)
	assert.Equal(t, []string{"example.com"}, provider.Domains)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListProviders(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	ctx := context.Background()
	tenantID := uuid.New()

	now := time.Now()
	config, _ := json.Marshal(SAMLConfig{})

	mock.ExpectQuery(`SELECT .+ FROM sso_providers WHERE tenant_id = \$1`).
		WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "name", "provider_type", "enabled", "enforce_sso",
			"config", "domains", "created_at", "updated_at", "created_by", "updated_by",
		}).
			AddRow(uuid.New(), tenantID, "SSO 1", ProviderTypeSAML, true, false, config, pq.Array([]string{"example.com"}), now, now, nil, nil).
			AddRow(uuid.New(), tenantID, "SSO 2", ProviderTypeOIDC, true, false, config, pq.Array([]string{"test.com"}), now, now, nil, nil))

	providers, err := repo.ListProviders(ctx, tenantID)
	assert.NoError(t, err)
	assert.Len(t, providers, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateProvider(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	ctx := context.Background()
	providerID := uuid.New()
	updatedBy := uuid.New()

	config, _ := json.Marshal(SAMLConfig{})
	provider := &Provider{
		ID:         providerID,
		Name:       "Updated SSO",
		Enabled:    false,
		EnforceSSO: true,
		Config:     config,
		Domains:    []string{"new.com"},
		UpdatedBy:  &updatedBy,
	}

	now := time.Now()

	mock.ExpectQuery(`UPDATE sso_providers`).
		WithArgs(
			"Updated SSO",
			false,
			true,
			config,
			pq.Array([]string{"new.com"}),
			updatedBy,
			providerID,
		).
		WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))

	err = repo.UpdateProvider(ctx, provider)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteProvider(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	ctx := context.Background()
	providerID := uuid.New()

	mock.ExpectExec(`DELETE FROM sso_providers WHERE id = \$1`).
		WithArgs(providerID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteProvider(ctx, providerID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetProviderByDomain(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	ctx := context.Background()
	providerID := uuid.New()
	tenantID := uuid.New()
	domain := "example.com"

	config, _ := json.Marshal(SAMLConfig{})
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM sso_providers WHERE enabled = true AND \$1 = ANY\(domains\)`).
		WithArgs(domain).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "name", "provider_type", "enabled", "enforce_sso",
			"config", "domains", "created_at", "updated_at", "created_by", "updated_by",
		}).AddRow(
			providerID, tenantID, "Test SSO", ProviderTypeSAML, true, false,
			config, pq.Array([]string{domain}), now, now, nil, nil,
		))

	provider, err := repo.GetProviderByDomain(ctx, domain)
	assert.NoError(t, err)
	assert.Equal(t, providerID, provider.ID)
	assert.Contains(t, provider.Domains, domain)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CreateConnection(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	ctx := context.Background()
	userID := uuid.New()
	providerID := uuid.New()

	attrs, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	lastLogin := time.Now()

	conn := &Connection{
		UserID:      userID,
		ProviderID:  providerID,
		ExternalID:  "ext-123",
		Attributes:  attrs,
		LastLoginAt: &lastLogin,
	}

	now := time.Now()

	mock.ExpectQuery(`INSERT INTO sso_connections`).
		WithArgs(
			sqlmock.AnyArg(),
			userID,
			providerID,
			"ext-123",
			attrs,
			lastLogin,
		).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
			AddRow(now, now))

	err = repo.CreateConnection(ctx, conn)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, conn.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetConnectionByExternalID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	ctx := context.Background()
	connectionID := uuid.New()
	userID := uuid.New()
	providerID := uuid.New()
	externalID := "ext-123"

	attrs, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM sso_connections WHERE sso_provider_id = \$1 AND external_id = \$2`).
		WithArgs(providerID, externalID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "sso_provider_id", "external_id", "attributes",
			"last_login_at", "created_at", "updated_at",
		}).AddRow(
			connectionID, userID, providerID, externalID, attrs,
			now, now, now,
		))

	conn, err := repo.GetConnectionByExternalID(ctx, providerID, externalID)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.Equal(t, externalID, conn.ExternalID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CreateLoginEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	ctx := context.Background()
	providerID := uuid.New()
	userID := uuid.New()
	errorMsg := "test error"
	ipAddr := "192.168.1.1"
	userAgent := "Test Agent"

	event := &LoginEvent{
		ProviderID:   providerID,
		UserID:       &userID,
		ExternalID:   "ext-123",
		Status:       LoginStatusFailure,
		ErrorMessage: &errorMsg,
		IPAddress:    &ipAddr,
		UserAgent:    &userAgent,
	}

	now := time.Now()

	mock.ExpectQuery(`INSERT INTO sso_login_events`).
		WithArgs(
			sqlmock.AnyArg(),
			providerID,
			userID,
			"ext-123",
			LoginStatusFailure,
			errorMsg,
			ipAddr,
			userAgent,
		).
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(now))

	err = repo.CreateLoginEvent(ctx, event)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, event.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMaskSensitiveConfig(t *testing.T) {
	tests := []struct {
		name     string
		provType ProviderType
		config   string
		want     []string // Fields that should be masked
	}{
		{
			name:     "SAML config masking",
			provType: ProviderTypeSAML,
			config:   `{"entity_id":"test","certificate":"secret","private_key":"verysecret"}`,
			want:     []string{"certificate", "private_key"},
		},
		{
			name:     "OIDC config masking",
			provType: ProviderTypeOIDC,
			config:   `{"client_id":"test","client_secret":"secret"}`,
			want:     []string{"client_secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masked, err := MaskSensitiveConfig(tt.provType, json.RawMessage(tt.config))
			assert.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(masked, &result)
			assert.NoError(t, err)

			for _, field := range tt.want {
				assert.Equal(t, "[REDACTED]", result[field])
			}
		})
	}
}
