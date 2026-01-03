package marketplace

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCategoryRepository(sqlxDB)

	ctx := context.Background()
	now := time.Now()

	input := CreateCategoryInput{
		Name:         "DevOps",
		Slug:         "devops",
		Description:  "DevOps automation templates",
		Icon:         "server",
		DisplayOrder: 5,
	}

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "description", "icon", "parent_id", "display_order", "template_count", "created_at", "updated_at"}).
		AddRow("cat-123", "DevOps", "devops", "DevOps automation templates", "server", nil, 5, 0, now, now)

	mock.ExpectQuery("INSERT INTO marketplace_categories").
		WithArgs(input.Name, input.Slug, input.Description, input.Icon, input.ParentID, input.DisplayOrder).
		WillReturnRows(rows)

	category, err := repo.Create(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, "cat-123", category.ID)
	assert.Equal(t, "DevOps", category.Name)
	assert.Equal(t, "devops", category.Slug)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCategoryRepository(sqlxDB)

	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "description", "icon", "parent_id", "display_order", "template_count", "created_at", "updated_at"}).
		AddRow("cat-123", "Integration", "integration", "Integration templates", "link", nil, 1, 10, now, now)

	mock.ExpectQuery("SELECT (.+) FROM marketplace_categories WHERE id = ?").
		WithArgs("cat-123").
		WillReturnRows(rows)

	category, err := repo.GetByID(ctx, "cat-123")
	require.NoError(t, err)
	assert.Equal(t, "cat-123", category.ID)
	assert.Equal(t, "Integration", category.Name)
	assert.Equal(t, 10, category.TemplateCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_GetBySlug(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCategoryRepository(sqlxDB)

	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "description", "icon", "parent_id", "display_order", "template_count", "created_at", "updated_at"}).
		AddRow("cat-123", "Security", "security", "Security templates", "shield", nil, 6, 5, now, now)

	mock.ExpectQuery("SELECT (.+) FROM marketplace_categories WHERE slug = ?").
		WithArgs("security").
		WillReturnRows(rows)

	category, err := repo.GetBySlug(ctx, "security")
	require.NoError(t, err)
	assert.Equal(t, "cat-123", category.ID)
	assert.Equal(t, "security", category.Slug)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCategoryRepository(sqlxDB)

	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "description", "icon", "parent_id", "display_order", "template_count", "created_at", "updated_at"}).
		AddRow("cat-1", "Integration", "integration", "Integration templates", "link", nil, 1, 10, now, now).
		AddRow("cat-2", "Automation", "automation", "Automation templates", "zap", nil, 2, 15, now, now)

	mock.ExpectQuery("SELECT (.+) FROM marketplace_categories WHERE parent_id IS NULL ORDER BY display_order, name").
		WillReturnRows(rows)

	categories, err := repo.List(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, categories, 2)
	assert.Equal(t, "Integration", categories[0].Name)
	assert.Equal(t, "Automation", categories[1].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCategoryRepository(sqlxDB)

	ctx := context.Background()
	now := time.Now()

	input := UpdateCategoryInput{
		Name:        "Updated Name",
		Description: "Updated description",
	}

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "description", "icon", "parent_id", "display_order", "template_count", "created_at", "updated_at"}).
		AddRow("cat-123", "Updated Name", "devops", "Updated description", "server", nil, 5, 0, now, now)

	mock.ExpectQuery("UPDATE marketplace_categories SET").
		WithArgs("Updated Name", "Updated description", "cat-123").
		WillReturnRows(rows)

	category, err := repo.Update(ctx, "cat-123", input)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", category.Name)
	assert.Equal(t, "Updated description", category.Description)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCategoryRepository(sqlxDB)

	ctx := context.Background()

	mock.ExpectExec("DELETE FROM marketplace_categories WHERE id = ?").
		WithArgs("cat-123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(ctx, "cat-123")
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_GetTemplateCategories(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCategoryRepository(sqlxDB)

	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "description", "icon", "parent_id", "display_order", "template_count", "created_at", "updated_at"}).
		AddRow("cat-1", "Integration", "integration", "Integration templates", "link", nil, 1, 10, now, now).
		AddRow("cat-2", "DevOps", "devops", "DevOps templates", "server", nil, 5, 5, now, now)

	mock.ExpectQuery("SELECT c.(.+) FROM marketplace_categories c").
		WithArgs("template-123").
		WillReturnRows(rows)

	categories, err := repo.GetTemplateCategories(ctx, "template-123")
	require.NoError(t, err)
	assert.Len(t, categories, 2)
	assert.Equal(t, "Integration", categories[0].Name)
	assert.Equal(t, "DevOps", categories[1].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_AddTemplateCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCategoryRepository(sqlxDB)

	ctx := context.Background()

	mock.ExpectExec("INSERT INTO marketplace_template_categories").
		WithArgs("template-123", "cat-123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AddTemplateCategory(ctx, "template-123", "cat-123")
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_RemoveTemplateCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCategoryRepository(sqlxDB)

	ctx := context.Background()

	mock.ExpectExec("DELETE FROM marketplace_template_categories").
		WithArgs("template-123", "cat-123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.RemoveTemplateCategory(ctx, "template-123", "cat-123")
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
