package template

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreate verifies template creation
func TestCreate(t *testing.T) {
	repo := &MockRepository{}
	ctx := context.Background()
	tenantID := "test-tenant-123"
	userID := "user-123"

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
	template := &Template{
		TenantID:    &tenantID,
		Name:        "Test Template",
		Description: "Test description",
		Category:    "security",
		Definition:  definition,
		Tags:        []string{"test", "security"},
		IsPublic:    false,
		CreatedBy:   userID,
	}

	err := repo.Create(ctx, tenantID, template)
	require.NoError(t, err)
	assert.NotEmpty(t, template.ID)
	assert.Equal(t, "Test Template", template.Name)
}

// TestCreate_DuplicateName verifies duplicate name validation
func TestCreate_DuplicateName(t *testing.T) {
	repo := &MockRepository{}
	ctx := context.Background()
	tenantID := "test-tenant-123"

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
	template1 := &Template{
		TenantID:    &tenantID,
		Name:        "Duplicate Name",
		Category:    "security",
		Definition:  definition,
		CreatedBy:   "user-123",
	}

	err := repo.Create(ctx, tenantID, template1)
	require.NoError(t, err)

	template2 := &Template{
		TenantID:    &tenantID,
		Name:        "Duplicate Name",
		Category:    "monitoring",
		Definition:  definition,
		CreatedBy:   "user-123",
	}

	err = repo.Create(ctx, tenantID, template2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// TestGetByID verifies template retrieval
func TestGetByID(t *testing.T) {
	repo := &MockRepository{}
	ctx := context.Background()
	tenantID := "test-tenant-123"

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
	template := &Template{
		TenantID:    &tenantID,
		Name:        "Test Template",
		Category:    "security",
		Definition:  definition,
		CreatedBy:   "user-123",
	}

	err := repo.Create(ctx, tenantID, template)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, tenantID, template.ID)
	require.NoError(t, err)
	assert.Equal(t, template.ID, found.ID)
	assert.Equal(t, template.Name, found.Name)
	assert.Equal(t, template.Category, found.Category)
}

// TestGetByID_NotFound verifies not found error
func TestGetByID_NotFound(t *testing.T) {
	repo := &MockRepository{}
	ctx := context.Background()
	tenantID := "test-tenant-123"

	_, err := repo.GetByID(ctx, tenantID, "nonexistent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestList_NoFilter verifies listing all templates
func TestList_NoFilter(t *testing.T) {
	repo := &MockRepository{}
	ctx := context.Background()
	tenantID := "test-tenant-123"

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)

	templates := []*Template{
		{
			TenantID:    &tenantID,
			Name:        "Template 1",
			Category:    "security",
			Definition:  definition,
			Tags:        []string{"security", "scan"},
			CreatedBy:   "user-123",
		},
		{
			TenantID:    &tenantID,
			Name:        "Template 2",
			Category:    "monitoring",
			Definition:  definition,
			Tags:        []string{"monitoring", "alert"},
			CreatedBy:   "user-123",
		},
	}

	for _, tmpl := range templates {
		err := repo.Create(ctx, tenantID, tmpl)
		require.NoError(t, err)
	}

	result, err := repo.List(ctx, tenantID, TemplateFilter{})
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

// TestList_WithCategoryFilter verifies category filtering
func TestList_WithCategoryFilter(t *testing.T) {
	repo := &MockRepository{}
	ctx := context.Background()
	tenantID := "test-tenant-123"

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)

	templates := []*Template{
		{
			TenantID:    &tenantID,
			Name:        "Security Template",
			Category:    "security",
			Definition:  definition,
			CreatedBy:   "user-123",
		},
		{
			TenantID:    &tenantID,
			Name:        "Monitoring Template",
			Category:    "monitoring",
			Definition:  definition,
			CreatedBy:   "user-123",
		},
	}

	for _, tmpl := range templates {
		err := repo.Create(ctx, tenantID, tmpl)
		require.NoError(t, err)
	}

	filter := TemplateFilter{Category: "security"}
	result, err := repo.List(ctx, tenantID, filter)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "security", result[0].Category)
}

// TestList_WithTagsFilter verifies tag filtering
func TestList_WithTagsFilter(t *testing.T) {
	repo := &MockRepository{}
	ctx := context.Background()
	tenantID := "test-tenant-123"

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)

	templates := []*Template{
		{
			TenantID:    &tenantID,
			Name:        "Template 1",
			Category:    "security",
			Definition:  definition,
			Tags:        []string{"security", "scan"},
			CreatedBy:   "user-123",
		},
		{
			TenantID:    &tenantID,
			Name:        "Template 2",
			Category:    "security",
			Definition:  definition,
			Tags:        []string{"security", "compliance"},
			CreatedBy:   "user-123",
		},
		{
			TenantID:    &tenantID,
			Name:        "Template 3",
			Category:    "monitoring",
			Definition:  definition,
			Tags:        []string{"monitoring"},
			CreatedBy:   "user-123",
		},
	}

	for _, tmpl := range templates {
		err := repo.Create(ctx, tenantID, tmpl)
		require.NoError(t, err)
	}

	filter := TemplateFilter{Tags: []string{"scan"}}
	result, err := repo.List(ctx, tenantID, filter)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Template 1", result[0].Name)
}

// TestList_WithSearchQuery verifies search functionality
func TestList_WithSearchQuery(t *testing.T) {
	repo := &MockRepository{}
	ctx := context.Background()
	tenantID := "test-tenant-123"

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)

	templates := []*Template{
		{
			TenantID:    &tenantID,
			Name:        "Security Scan Pipeline",
			Description: "Automated security scanning",
			Category:    "security",
			Definition:  definition,
			CreatedBy:   "user-123",
		},
		{
			TenantID:    &tenantID,
			Name:        "Monitoring Alert",
			Description: "System monitoring",
			Category:    "monitoring",
			Definition:  definition,
			CreatedBy:   "user-123",
		},
	}

	for _, tmpl := range templates {
		err := repo.Create(ctx, tenantID, tmpl)
		require.NoError(t, err)
	}

	filter := TemplateFilter{SearchQuery: "security"}
	result, err := repo.List(ctx, tenantID, filter)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result[0].Name, "Security")
}

// TestList_PublicTemplates verifies public template access
func TestList_PublicTemplates(t *testing.T) {
	repo := &MockRepository{}
	ctx := context.Background()
	tenantID := "test-tenant-123"
	otherTenantID := "other-tenant-456"

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)

	publicTemplate := &Template{
		TenantID:    &otherTenantID,
		Name:        "Public Template",
		Category:    "integration",
		Definition:  definition,
		IsPublic:    true,
		CreatedBy:   "user-456",
	}

	err := repo.Create(ctx, otherTenantID, publicTemplate)
	require.NoError(t, err)

	privateTemplate := &Template{
		TenantID:    &tenantID,
		Name:        "Private Template",
		Category:    "security",
		Definition:  definition,
		IsPublic:    false,
		CreatedBy:   "user-123",
	}

	err = repo.Create(ctx, tenantID, privateTemplate)
	require.NoError(t, err)

	result, err := repo.List(ctx, tenantID, TemplateFilter{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 2)
}

// TestUpdate verifies template update
func TestUpdate(t *testing.T) {
	repo := &MockRepository{}
	ctx := context.Background()
	tenantID := "test-tenant-123"

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
	template := &Template{
		TenantID:    &tenantID,
		Name:        "Original Name",
		Description: "Original description",
		Category:    "security",
		Definition:  definition,
		CreatedBy:   "user-123",
	}

	err := repo.Create(ctx, tenantID, template)
	require.NoError(t, err)

	update := UpdateTemplateInput{
		Name:        "Updated Name",
		Description: "Updated description",
	}

	err = repo.Update(ctx, tenantID, template.ID, update)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, tenantID, template.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated description", updated.Description)
}

// TestDelete verifies template deletion
func TestDelete(t *testing.T) {
	repo := &MockRepository{}
	ctx := context.Background()
	tenantID := "test-tenant-123"

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
	template := &Template{
		TenantID:    &tenantID,
		Name:        "To Delete",
		Category:    "security",
		Definition:  definition,
		CreatedBy:   "user-123",
	}

	err := repo.Create(ctx, tenantID, template)
	require.NoError(t, err)

	err = repo.Delete(ctx, tenantID, template.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, tenantID, template.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// MockRepository is a mock implementation for testing
type MockRepository struct {
	templates map[string]*Template
	counter   int
}

func (m *MockRepository) Create(ctx context.Context, tenantID string, template *Template) error {
	if m.templates == nil {
		m.templates = make(map[string]*Template)
	}

	// Check for duplicate name
	for _, t := range m.templates {
		if t.TenantID != nil && *t.TenantID == tenantID && t.Name == template.Name {
			return errors.New("template already exists with this name")
		}
	}

	m.counter++
	template.ID = fmt.Sprintf("template-%s-%d", time.Now().Format("20060102150405"), m.counter)
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	m.templates[template.ID] = template
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, tenantID, id string) (*Template, error) {
	if m.templates == nil {
		return nil, errors.New("template not found")
	}

	template, ok := m.templates[id]
	if !ok {
		return nil, errors.New("template not found")
	}

	return template, nil
}

func (m *MockRepository) List(ctx context.Context, tenantID string, filter TemplateFilter) ([]*Template, error) {
	if m.templates == nil {
		return []*Template{}, nil
	}

	var result []*Template
	for _, template := range m.templates {
		// Check tenant access
		if template.TenantID != nil && *template.TenantID != tenantID && !template.IsPublic {
			continue
		}

		// Apply filters
		if filter.Category != "" && template.Category != filter.Category {
			continue
		}

		if len(filter.Tags) > 0 {
			hasTag := false
			for _, filterTag := range filter.Tags {
				for _, tag := range template.Tags {
					if tag == filterTag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		if filter.SearchQuery != "" {
			// Simple case-sensitive search
			found := false
			if contains(template.Name, filter.SearchQuery) {
				found = true
			}
			if contains(template.Description, filter.SearchQuery) {
				found = true
			}
			if !found {
				continue
			}
		}

		result = append(result, template)
	}

	return result, nil
}

func (m *MockRepository) Update(ctx context.Context, tenantID, id string, input UpdateTemplateInput) error {
	if m.templates == nil {
		return errors.New("template not found")
	}

	template, ok := m.templates[id]
	if !ok {
		return errors.New("template not found")
	}

	if input.Name != "" {
		template.Name = input.Name
	}
	if input.Description != "" {
		template.Description = input.Description
	}
	if input.Category != "" {
		template.Category = input.Category
	}
	if input.Definition != nil {
		template.Definition = input.Definition
	}
	if input.Tags != nil {
		template.Tags = input.Tags
	}
	if input.IsPublic != nil {
		template.IsPublic = *input.IsPublic
	}

	template.UpdatedAt = time.Now()
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, tenantID, id string) error {
	if m.templates == nil {
		return errors.New("template not found")
	}

	if _, ok := m.templates[id]; !ok {
		return errors.New("template not found")
	}

	delete(m.templates, id)
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
