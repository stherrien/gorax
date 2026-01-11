package marketplace

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCategoryRepository is a mock implementation of CategoryRepository
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(ctx context.Context, input CreateCategoryInput) (*Category, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id string) (*Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Category), args.Error(1)
}

func (m *MockCategoryRepository) GetBySlug(ctx context.Context, slug string) (*Category, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Category), args.Error(1)
}

func (m *MockCategoryRepository) List(ctx context.Context, parentID *string) ([]Category, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Category), args.Error(1)
}

func (m *MockCategoryRepository) GetWithChildren(ctx context.Context) ([]Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(ctx context.Context, id string, input UpdateCategoryInput) (*Category, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Category), args.Error(1)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetTemplateCategories(ctx context.Context, templateID string) ([]Category, error) {
	args := m.Called(ctx, templateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Category), args.Error(1)
}

func (m *MockCategoryRepository) AddTemplateCategory(ctx context.Context, templateID, categoryID string) error {
	args := m.Called(ctx, templateID, categoryID)
	return args.Error(0)
}

func (m *MockCategoryRepository) RemoveTemplateCategory(ctx context.Context, templateID, categoryID string) error {
	args := m.Called(ctx, templateID, categoryID)
	return args.Error(0)
}

func (m *MockCategoryRepository) SetTemplateCategories(ctx context.Context, templateID string, categoryIDs []string) error {
	args := m.Called(ctx, templateID, categoryIDs)
	return args.Error(0)
}

func TestCategoryService_CreateCategory(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name    string
		input   CreateCategoryInput
		setup   func(*MockCategoryRepository)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful creation",
			input: CreateCategoryInput{
				Name:        "DevOps",
				Slug:        "devops",
				Description: "DevOps templates",
				Icon:        "server",
			},
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetBySlug", ctx, "devops").Return(nil, errors.New("not found"))
				repo.On("Create", ctx, mock.AnythingOfType("CreateCategoryInput")).Return(&Category{
					ID:          "cat-123",
					Name:        "DevOps",
					Slug:        "devops",
					Description: "DevOps templates",
					Icon:        "server",
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "duplicate slug",
			input: CreateCategoryInput{
				Name: "DevOps",
				Slug: "devops",
			},
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetBySlug", ctx, "devops").Return(&Category{ID: "existing"}, nil)
			},
			wantErr: true,
			errMsg:  "already exists",
		},
		{
			name: "invalid parent",
			input: CreateCategoryInput{
				Name:     "CI/CD",
				Slug:     "cicd",
				ParentID: catStringPtr("invalid-parent"),
			},
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetBySlug", ctx, "cicd").Return(nil, errors.New("not found"))
				repo.On("GetByID", ctx, "invalid-parent").Return(nil, errors.New("not found"))
			},
			wantErr: true,
			errMsg:  "parent category not found",
		},
		{
			name: "too deep nesting",
			input: CreateCategoryInput{
				Name:     "SubCategory",
				Slug:     "subcategory",
				ParentID: catStringPtr("parent-with-parent"),
			},
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetBySlug", ctx, "subcategory").Return(nil, errors.New("not found"))
				grandparentID := "grandparent"
				repo.On("GetByID", ctx, "parent-with-parent").Return(&Category{
					ID:       "parent-with-parent",
					ParentID: &grandparentID,
				}, nil)
			},
			wantErr: true,
			errMsg:  "2 levels deep",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockCategoryRepository)
			tt.setup(repo)

			service := NewCategoryService(repo)
			category, err := service.CreateCategory(ctx, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, category)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, category)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestCategoryService_UpdateCategory(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name    string
		id      string
		input   UpdateCategoryInput
		setup   func(*MockCategoryRepository)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful update",
			id:   "cat-123",
			input: UpdateCategoryInput{
				Name:        "Updated Name",
				Description: "Updated description",
			},
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetByID", ctx, "cat-123").Return(&Category{
					ID:   "cat-123",
					Name: "Old Name",
					Slug: "old-slug",
				}, nil).Times(1)
				repo.On("Update", ctx, "cat-123", mock.AnythingOfType("UpdateCategoryInput")).Return(&Category{
					ID:          "cat-123",
					Name:        "Updated Name",
					Description: "Updated description",
					UpdatedAt:   now,
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "slug already exists",
			id:   "cat-123",
			input: UpdateCategoryInput{
				Slug: "existing-slug",
			},
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetByID", ctx, "cat-123").Return(&Category{
					ID:   "cat-123",
					Slug: "old-slug",
				}, nil)
				repo.On("GetBySlug", ctx, "existing-slug").Return(&Category{
					ID:   "other-cat",
					Slug: "existing-slug",
				}, nil)
			},
			wantErr: true,
			errMsg:  "already exists",
		},
		{
			name: "self-reference parent",
			id:   "cat-123",
			input: UpdateCategoryInput{
				ParentID: catStringPtr("cat-123"),
			},
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetByID", ctx, "cat-123").Return(&Category{
					ID: "cat-123",
				}, nil)
			},
			wantErr: true,
			errMsg:  "cannot be its own parent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockCategoryRepository)
			tt.setup(repo)

			service := NewCategoryService(repo)
			category, err := service.UpdateCategory(ctx, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, category)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestCategoryService_DeleteCategory(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		id      string
		setup   func(*MockCategoryRepository)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful deletion",
			id:   "cat-123",
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetByID", ctx, "cat-123").Return(&Category{ID: "cat-123"}, nil)
				repo.On("List", ctx, catStringPtr("cat-123")).Return([]Category{}, nil)
				repo.On("Delete", ctx, "cat-123").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "category has children",
			id:   "cat-123",
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetByID", ctx, "cat-123").Return(&Category{ID: "cat-123"}, nil)
				repo.On("List", ctx, catStringPtr("cat-123")).Return([]Category{
					{ID: "child-1"},
				}, nil)
			},
			wantErr: true,
			errMsg:  "with child categories",
		},
		{
			name: "category not found",
			id:   "invalid",
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetByID", ctx, "invalid").Return(nil, errors.New("not found"))
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockCategoryRepository)
			tt.setup(repo)

			service := NewCategoryService(repo)
			err := service.DeleteCategory(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestCategoryService_SetTemplateCategories(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		templateID  string
		categoryIDs []string
		setup       func(*MockCategoryRepository)
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "successful assignment",
			templateID:  "template-123",
			categoryIDs: []string{"cat-1", "cat-2"},
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetByID", ctx, "cat-1").Return(&Category{ID: "cat-1"}, nil)
				repo.On("GetByID", ctx, "cat-2").Return(&Category{ID: "cat-2"}, nil)
				repo.On("SetTemplateCategories", ctx, "template-123", []string{"cat-1", "cat-2"}).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "too many categories",
			templateID:  "template-123",
			categoryIDs: []string{"cat-1", "cat-2", "cat-3", "cat-4", "cat-5", "cat-6"},
			setup:       func(repo *MockCategoryRepository) {},
			wantErr:     true,
			errMsg:      "maximum of 5",
		},
		{
			name:        "invalid category",
			templateID:  "template-123",
			categoryIDs: []string{"cat-1", "invalid"},
			setup: func(repo *MockCategoryRepository) {
				repo.On("GetByID", ctx, "cat-1").Return(&Category{ID: "cat-1"}, nil)
				repo.On("GetByID", ctx, "invalid").Return(nil, errors.New("not found"))
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockCategoryRepository)
			tt.setup(repo)

			service := NewCategoryService(repo)
			err := service.SetTemplateCategories(ctx, tt.templateID, tt.categoryIDs)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestCategoryService_GetCategory(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	repo := new(MockCategoryRepository)
	repo.On("GetByID", ctx, "cat-123").Return(&Category{
		ID:        "cat-123",
		Name:      "DevOps",
		CreatedAt: now,
	}, nil)

	service := NewCategoryService(repo)
	category, err := service.GetCategory(ctx, "cat-123")

	require.NoError(t, err)
	assert.Equal(t, "cat-123", category.ID)
	assert.Equal(t, "DevOps", category.Name)

	repo.AssertExpectations(t)
}

func TestCategoryService_ListCategories(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	categories := []Category{
		{ID: "cat-1", Name: "Integration", CreatedAt: now},
		{ID: "cat-2", Name: "Automation", CreatedAt: now},
	}

	repo := new(MockCategoryRepository)
	repo.On("List", ctx, (*string)(nil)).Return(categories, nil)

	service := NewCategoryService(repo)
	result, err := service.ListCategories(ctx, nil)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Integration", result[0].Name)

	repo.AssertExpectations(t)
}
