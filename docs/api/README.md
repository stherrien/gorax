# Gorax API Documentation

This directory contains the OpenAPI/Swagger documentation for the Gorax REST API.

## Files

- **swagger.json** - OpenAPI 3.0 specification in JSON format
- **swagger.yaml** - OpenAPI 3.0 specification in YAML format
- **docs.go** - Generated Go documentation (auto-imported by main.go)
- **gorax.postman_collection.json** - Postman collection with all API endpoints
- **gorax.postman_environment.json** - Postman environment template

## Viewing Documentation

### Swagger UI (Interactive)

Start the API server and navigate to:

```
http://localhost:8080/api/docs/
```

The Swagger UI provides:
- Interactive API exploration
- Try-it-out functionality for all endpoints
- Request/response examples
- Schema definitions
- Authentication configuration

### Static Documentation

For detailed examples and use cases, see:

```
/Users/shawntherrien/Projects/gorax/docs/API_REFERENCE.md
```

## Using Postman

### Import Collection

1. Open Postman
2. Click **Import**
3. Select `gorax.postman_collection.json`
4. Import `gorax.postman_environment.json` for pre-configured variables

### Collection Features

- **Auto-extraction of IDs**: Test scripts automatically save workflow IDs, webhook IDs, and execution IDs to collection variables
- **Pre-configured requests**: All endpoints include sample request bodies
- **Environment variables**: Easy switching between dev/staging/prod environments

### Collection Variables

The collection uses these variables (set in environment or collection):

- `baseUrl` - API base URL (default: http://localhost:8080)
- `apiBasePath` - API version path (default: /api/v1)
- `tenantId` - Tenant identifier for multi-tenant isolation
- `userId` - User identifier (dev mode only)
- `workflowId` - Auto-populated when creating workflows
- `webhookId` - Auto-populated when creating webhooks
- `executionId` - Auto-populated when executing workflows

## Regenerating Documentation

After adding or modifying API endpoints, regenerate the Swagger docs:

```bash
cd /Users/shawntherrien/Projects/gorax

# Generate Swagger documentation
swag init -g cmd/api/main.go -o docs/api --parseDependency --parseInternal

# Verify build
go build ./cmd/api
```

## Adding Swagger Annotations

When creating new API endpoints, add Swagger annotations above the handler:

```go
// CreateResource creates a new resource
// @Summary Create resource
// @Description Creates a new resource with the provided data
// @Tags Resources
// @Accept json
// @Produce json
// @Param resource body ResourceInput true "Resource data"
// @Security TenantID
// @Security UserID
// @Success 201 {object} map[string]interface{} "Created resource"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /resources [post]
func (h *Handler) CreateResource(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

### Annotation Reference

- `@Summary` - Short description (appears in endpoint list)
- `@Description` - Detailed description
- `@Tags` - Group endpoints by tag (Resources, Workflows, etc.)
- `@Accept` - Request content type
- `@Produce` - Response content type
- `@Param` - Parameter definition
  - Format: `name location type required "description"`
  - Locations: `path`, `query`, `body`, `header`
- `@Security` - Authentication requirements
- `@Success` - Success response
  - Format: `code {type} ResponseType "description"`
- `@Failure` - Error response
- `@Router` - API route path and HTTP method

## OpenAPI Specification

The generated OpenAPI 3.0 specification can be used to:

- Generate client SDKs in multiple languages
- Import into API gateways (Kong, Apigee, etc.)
- Generate mock servers for testing
- Validate API responses in integration tests
- Create API documentation portals

### SDK Generation Examples

**TypeScript/JavaScript:**
```bash
npx @openapitools/openapi-generator-cli generate \
  -i docs/api/swagger.yaml \
  -g typescript-axios \
  -o clients/typescript
```

**Python:**
```bash
openapi-generator-cli generate \
  -i docs/api/swagger.yaml \
  -g python \
  -o clients/python
```

**Go:**
```bash
oapi-codegen -package goraxclient docs/api/swagger.yaml > clients/go/client.go
```

## CI/CD Integration

### Validate Documentation

Add to your CI pipeline:

```bash
# Verify Swagger annotations are valid
swag init -g cmd/api/main.go -o /tmp/swagger --parseDependency --parseInternal

# Validate OpenAPI spec
npx @apidevtools/swagger-cli validate docs/api/swagger.yaml
```

### Auto-update on PR

Consider using GitHub Actions to auto-regenerate docs:

```yaml
name: Update API Docs
on:
  pull_request:
    paths:
      - 'internal/api/**'
      - 'cmd/api/**'

jobs:
  update-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v3
      - run: go install github.com/swaggo/swag/cmd/swag@latest
      - run: swag init -g cmd/api/main.go -o docs/api
      - uses: stefanzweifel/git-auto-commit-action@v4
        with:
          commit_message: "docs: auto-update API documentation"
```

## Troubleshooting

### Missing Annotations

If endpoints don't appear in Swagger UI:

1. Ensure handler has `@Router` annotation
2. Check that handler is registered in router
3. Verify `@Tags` matches expected tag name
4. Regenerate docs with `swag init`

### Type Resolution Issues

If types don't resolve correctly:

1. Use fully-qualified package paths in `@Param` body types
2. Add `--parseDependency` flag when generating
3. Ensure struct has exported fields (capitalized)
4. Add JSON tags to struct fields

### Build Errors

If build fails after adding Swagger:

1. Ensure `docs/api` package is imported in `main.go`:
   ```go
   import _ "github.com/gorax/gorax/docs/api"
   ```
2. Run `go mod tidy`
3. Verify all Swagger packages are in `go.mod`

## Resources

- [Swag Documentation](https://github.com/swaggo/swag)
- [OpenAPI Specification](https://spec.openapis.org/oas/v3.0.0)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)
- [Postman Documentation](https://learning.postman.com/docs/getting-started/introduction/)

## Support

For API documentation issues:
- Create an issue: https://github.com/gorax/gorax/issues
- Email: support@gorax.io

---

Last Updated: 2024-01-20
