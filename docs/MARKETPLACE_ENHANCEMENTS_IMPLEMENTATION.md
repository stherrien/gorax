# Marketplace Enhancements Implementation Guide

## Overview
This document tracks the implementation of marketplace enhancements including categories, featured templates, and improved UI.

## Implementation Status

### ✅ Completed
1. **Database Migration** (`migrations/026_marketplace_enhancements.sql`)
   - Categories table with hierarchy support
   - Template-categories junction table
   - Featured template columns
   - Full-text search support
   - Seed data for 10 popular categories
   - Automatic template count maintenance

2. **Backend Models**
   - `category_model.go`: Category, CreateCategoryInput, UpdateCategoryInput, FeatureTemplateInput
   - `category_model_test.go`: Comprehensive model validation tests
   - `enhanced_model.go`: EnhancedSearchFilter, MarketplaceTemplateWithCategories
   - Enhanced MarketplaceTemplate with `IsFeatured`, `FeaturedAt`, `FeaturedBy`, `Categories[]`

3. **Category Repository** (`category_repository.go`)
   - Full CRUD operations for categories
   - Hierarchy support (parent-child relationships)
   - Template-category associations
   - Comprehensive test coverage (`category_repository_test.go`)

### 🔄 In Progress

4. **Enhanced Marketplace Repository**
   - Need to add methods for:
     - `SearchWithCategories()` - Search with category filtering
     - `GetFeatured()` - Get featured templates
     - `FeatureTemplate()` - Feature/unfeature a template
     - `GetByCategory()` - Get templates by category
     - Full-text search support

5. **Category Service**
   - Business logic for category management
   - Validation and authorization
   - Hierarchy operations
   - Template categorization

6. **Marketplace Service Enhancements**
   - Enhanced search with categories
   - Featured template management
   - Category-based filtering

### 📋 TODO

7. **API Handlers**
   - `GET /api/v1/marketplace/categories` - List all categories
   - `POST /api/v1/marketplace/categories` - Create category (admin)
   - `GET /api/v1/marketplace/categories/:id` - Get category details
   - `PUT /api/v1/marketplace/categories/:id` - Update category (admin)
   - `DELETE /api/v1/marketplace/categories/:id` - Delete category (admin)
   - `GET /api/v1/marketplace/templates?category_ids[]=X&featured=true` - Enhanced search
   - `PUT /api/v1/marketplace/templates/:id/featured` - Feature template (admin)
   - `PUT /api/v1/marketplace/templates/:id/categories` - Set template categories

8. **Frontend Types** (`web/src/types/marketplace.ts`)
   ```typescript
   interface Category {
     id: string;
     name: string;
     slug: string;
     description: string;
     icon: string;
     parent_id?: string;
     display_order: number;
     template_count: number;
     created_at: string;
     updated_at: string;
     children?: Category[];
   }

   interface EnhancedMarketplaceTemplate extends MarketplaceTemplate {
     is_featured: boolean;
     featured_at?: string;
     categories: Category[];
   }

   interface EnhancedSearchFilter {
     category_ids?: string[];
     tags?: string[];
     search_query?: string;
     min_rating?: number;
     is_verified?: boolean;
     is_featured?: boolean;
     sort_by?: 'popular' | 'recent' | 'rating' | 'name' | 'relevance';
     page?: number;
     limit?: number;
   }
   ```

9. **Frontend API Client** (`web/src/api/marketplace.ts`)
   - `getCategories()` - Fetch all categories
   - `getCategoryBySlug(slug)` - Get category by slug
   - `searchTemplatesEnhanced(filter)` - Enhanced search
   - `getFeaturedTemplates(limit?)` - Get featured templates
   - `featureTemplate(id, featured)` - Feature/unfeature template (admin)
   - `setTemplateCategories(id, categoryIds)` - Set template categories

10. **React Components**

### CategoryCard Component
```tsx
// web/src/components/marketplace/CategoryCard.tsx
interface CategoryCardProps {
  category: Category;
  onClick?: () => void;
}

export const CategoryCard: React.FC<CategoryCardProps> = ({ category, onClick }) => {
  return (
    <div className="category-card" onClick={onClick}>
      <div className="icon">{category.icon}</div>
      <h3>{category.name}</h3>
      <p>{category.description}</p>
      <span className="count">{category.template_count} templates</span>
    </div>
  );
};
```

### CategoryList Component
```tsx
// web/src/components/marketplace/CategoryList.tsx
interface CategoryListProps {
  categories: Category[];
  selectedCategoryId?: string;
  onSelectCategory: (category: Category) => void;
}

export const CategoryList: React.FC<CategoryListProps> = ({
  categories,
  selectedCategoryId,
  onSelectCategory,
}) => {
  return (
    <div className="category-list">
      {categories.map((category) => (
        <CategoryCard
          key={category.id}
          category={category}
          onClick={() => onSelectCategory(category)}
          isSelected={category.id === selectedCategoryId}
        />
      ))}
    </div>
  );
};
```

### FeaturedTemplates Component
```tsx
// web/src/components/marketplace/FeaturedTemplates.tsx
export const FeaturedTemplates: React.FC = () => {
  const { data: templates, isLoading } = useFeaturedTemplates();

  if (isLoading) return <Skeleton />;
  if (!templates?.length) return null;

  return (
    <section className="featured-templates">
      <h2>Featured Templates</h2>
      <div className="template-carousel">
        {templates.map((template) => (
          <EnhancedTemplateCard
            key={template.id}
            template={template}
            featured
          />
        ))}
      </div>
    </section>
  );
};
```

### Enhanced TemplateCard
```tsx
// web/src/components/marketplace/EnhancedTemplateCard.tsx
interface EnhancedTemplateCardProps {
  template: EnhancedMarketplaceTemplate;
  onInstall?: () => void;
  onPreview?: () => void;
}

export const EnhancedTemplateCard: React.FC<EnhancedTemplateCardProps> = ({
  template,
  onInstall,
  onPreview,
}) => {
  return (
    <div className="template-card">
      {template.is_featured && <Badge variant="featured">Featured</Badge>}
      {template.is_verified && <Badge variant="verified">Verified</Badge>}

      <h3>{template.name}</h3>
      <p>{template.description}</p>

      <div className="categories">
        {template.categories.map((cat) => (
          <Tag key={cat.id} icon={cat.icon}>
            {cat.name}
          </Tag>
        ))}
      </div>

      <div className="stats">
        <Star rating={template.average_rating} />
        <Download count={template.download_count} />
      </div>

      <div className="actions">
        <Button onClick={onPreview}>Preview</Button>
        <Button variant="primary" onClick={onInstall}>
          Use Template
        </Button>
      </div>
    </div>
  );
};
```

11. **Redesigned Marketplace Page**
```tsx
// web/src/pages/Marketplace.tsx
export const Marketplace: React.FC = () => {
  const [selectedCategory, setSelectedCategory] = useState<string>();
  const [searchQuery, setSearchQuery] = useState('');
  const [filter, setFilter] = useState<EnhancedSearchFilter>({});

  const { data: categories } = useCategories();
  const { data: featured } = useFeaturedTemplates(6);
  const { data: templates } = useMarketplaceSearch(filter);
  const { data: recent } = useRecentTemplates(10);
  const { data: popular } = usePopularTemplates(10);

  return (
    <div className="marketplace">
      {/* Hero Section */}
      <section className="hero">
        <h1>Workflow Template Marketplace</h1>
        <SearchBar
          value={searchQuery}
          onChange={setSearchQuery}
          onSearch={() => setFilter({ ...filter, search_query: searchQuery })}
        />
      </section>

      {/* Featured Templates */}
      <FeaturedTemplates templates={featured} />

      {/* Browse by Category */}
      <section className="categories">
        <h2>Browse by Category</h2>
        <CategoryList
          categories={categories}
          selectedCategoryId={selectedCategory}
          onSelectCategory={(cat) => {
            setSelectedCategory(cat.id);
            setFilter({ ...filter, category_ids: [cat.id] });
          }}
        />
      </section>

      {/* Main Content */}
      <div className="content">
        {/* Filters Sidebar */}
        <aside className="filters">
          <FilterPanel
            filter={filter}
            onFilterChange={setFilter}
            categories={categories}
          />
        </aside>

        {/* Template Grid */}
        <main className="templates">
          {selectedCategory && (
            <CategoryBreadcrumb
              category={categories?.find((c) => c.id === selectedCategory)}
              onClear={() => setSelectedCategory(undefined)}
            />
          )}

          <TemplateGrid
            templates={templates}
            loading={isLoading}
            emptyMessage="No templates found"
          />
        </main>
      </div>

      {/* Recently Added */}
      <section className="recent">
        <h2>Recently Added</h2>
        <TemplateRow templates={recent} />
      </section>

      {/* Most Popular */}
      <section className="popular">
        <h2>Most Popular</h2>
        <TemplateRow templates={popular} />
      </section>
    </div>
  );
};
```

12. **Custom Hooks**
```typescript
// web/src/hooks/useMarketplaceEnhanced.ts
export const useCategories = () => {
  return useQuery({
    queryKey: ['marketplace', 'categories'],
    queryFn: () => marketplaceApi.getCategories(),
  });
};

export const useFeaturedTemplates = (limit = 10) => {
  return useQuery({
    queryKey: ['marketplace', 'featured', limit],
    queryFn: () => marketplaceApi.getFeaturedTemplates(limit),
  });
};

export const useMarketplaceSearch = (filter: EnhancedSearchFilter) => {
  return useQuery({
    queryKey: ['marketplace', 'search', filter],
    queryFn: () => marketplaceApi.searchTemplatesEnhanced(filter),
    enabled: Object.keys(filter).length > 0,
  });
};

export const useFeatureTemplate = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, featured }: { id: string; featured: boolean }) =>
      marketplaceApi.featureTemplate(id, featured),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace'] });
    },
  });
};
```

## Testing Strategy

### Backend Tests
- ✅ Unit tests for models and validation
- ✅ Repository tests with sqlmock
- 🔄 Service tests with business logic
- ⏳ Integration tests with real database
- ⏳ API handler tests

### Frontend Tests
- ⏳ Component tests (CategoryCard, CategoryList, FeaturedTemplates)
- ⏳ Hook tests (useCategories, useFeaturedTemplates)
- ⏳ Integration tests (Marketplace page)
- ⏳ E2E tests (full marketplace flow)

## Migration Steps

1. **Run Migration**
   ```bash
   # Apply the new migration
   make migrate-up
   # or
   goose -dir migrations postgres "connection-string" up
   ```

2. **Verify Seed Data**
   ```sql
   SELECT * FROM marketplace_categories ORDER BY display_order;
   -- Should show 10 categories
   ```

3. **Migrate Existing Templates** (if any)
   The migration automatically maps old string categories to new category table.

4. **Deploy Backend**
   - Deploy new models and repositories
   - Deploy category service
   - Deploy enhanced marketplace service
   - Deploy API handlers

5. **Deploy Frontend**
   - Deploy new components
   - Update marketplace page
   - Test thoroughly

## API Documentation

### Category Endpoints

#### List Categories
```
GET /api/v1/marketplace/categories
Query params:
  - parent_id (optional): Filter by parent category
Response: Category[]
```

#### Create Category (Admin Only)
```
POST /api/v1/marketplace/categories
Headers: Authorization: Bearer <admin-token>
Body: CreateCategoryInput
Response: Category
```

#### Get Category
```
GET /api/v1/marketplace/categories/:id
Response: Category with children
```

#### Update Category (Admin Only)
```
PUT /api/v1/marketplace/categories/:id
Headers: Authorization: Bearer <admin-token>
Body: UpdateCategoryInput
Response: Category
```

#### Delete Category (Admin Only)
```
DELETE /api/v1/marketplace/categories/:id
Headers: Authorization: Bearer <admin-token>
Response: 204 No Content
```

### Enhanced Template Endpoints

#### Search with Categories
```
GET /api/v1/marketplace/templates
Query params:
  - category_ids[]: string[] (multi-select)
  - tags[]: string[]
  - search_query: string
  - min_rating: number
  - is_verified: boolean
  - is_featured: boolean
  - sort_by: 'popular'|'recent'|'rating'|'name'|'relevance'
  - page: number
  - limit: number (max 100)
Response: { templates: EnhancedMarketplaceTemplate[], total: number }
```

#### Feature Template (Admin Only)
```
PUT /api/v1/marketplace/templates/:id/featured
Headers: Authorization: Bearer <admin-token>
Body: { is_featured: boolean }
Response: EnhancedMarketplaceTemplate
```

#### Set Template Categories
```
PUT /api/v1/marketplace/templates/:id/categories
Body: { category_ids: string[] }
Response: EnhancedMarketplaceTemplate with categories
```

## Performance Considerations

1. **Database Indexes** (already created in migration)
   - Category slug index for fast lookups
   - Template featured index for homepage queries
   - Full-text search index for search queries
   - Junction table indexes for category filtering

2. **Caching Strategy**
   - Cache category list (rarely changes)
   - Cache featured templates (5-10 min TTL)
   - Cache popular templates (15 min TTL)

3. **Query Optimization**
   - Use JOINs to fetch categories with templates in single query
   - Limit results with pagination
   - Use full-text search for better performance than LIKE queries

## Security Considerations

1. **Authorization**
   - Only admins can create/update/delete categories
   - Only admins can feature/unfeature templates
   - All users can browse and search

2. **Validation**
   - Validate slug format (lowercase, alphanumeric, hyphens only)
   - Prevent circular parent-child relationships
   - Validate category IDs before association

3. **SQL Injection Prevention**
   - All queries use parameterized statements
   - Input validation on all user inputs

## Next Steps

1. Complete marketplace repository enhancements
2. Implement category service with tests
3. Update marketplace service for enhanced search
4. Implement API handlers with tests
5. Create frontend types and API client
6. Implement React components with tests
7. Redesign marketplace page
8. Run full integration tests
9. Update API documentation
10. Deploy to staging for QA

## Files Created

### Backend
- ✅ `/migrations/026_marketplace_enhancements.sql`
- ✅ `/internal/marketplace/category_model.go`
- ✅ `/internal/marketplace/category_model_test.go`
- ✅ `/internal/marketplace/category_repository.go`
- ✅ `/internal/marketplace/category_repository_test.go`
- ✅ `/internal/marketplace/enhanced_model.go`
- ⏳ `/internal/marketplace/category_service.go`
- ⏳ `/internal/marketplace/category_service_test.go`
- ⏳ `/internal/api/handlers/marketplace_category_handler.go`
- ⏳ `/internal/api/handlers/marketplace_category_handler_test.go`

### Frontend
- ⏳ `/web/src/types/marketplace-enhanced.ts`
- ⏳ `/web/src/api/marketplace-enhanced.ts`
- ⏳ `/web/src/components/marketplace/CategoryCard.tsx`
- ⏳ `/web/src/components/marketplace/CategoryCard.test.tsx`
- ⏳ `/web/src/components/marketplace/CategoryList.tsx`
- ⏳ `/web/src/components/marketplace/CategoryList.test.tsx`
- ⏳ `/web/src/components/marketplace/FeaturedTemplates.tsx`
- ⏳ `/web/src/components/marketplace/FeaturedTemplates.test.tsx`
- ⏳ `/web/src/components/marketplace/EnhancedTemplateCard.tsx`
- ⏳ `/web/src/components/marketplace/EnhancedTemplateCard.test.tsx`
- ⏳ `/web/src/hooks/useMarketplaceEnhanced.ts`
- ⏳ `/web/src/hooks/useMarketplaceEnhanced.test.ts`
- ⏳ `/web/src/pages/MarketplaceEnhanced.tsx` (or update existing)

### Documentation
- ✅ `/docs/MARKETPLACE_ENHANCEMENTS_IMPLEMENTATION.md` (this file)

## Legend
- ✅ Completed
- 🔄 In Progress
- ⏳ TODO
