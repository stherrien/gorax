# Contributing to Gorax

Thank you for your interest in contributing to Gorax! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Documentation](#documentation)
- [Issue Reporting](#issue-reporting)
- [Community](#community)

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive environment for all contributors. We pledge to:

- Be respectful and considerate in all interactions
- Welcome diverse perspectives and experiences
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards other community members

### Unacceptable Behavior

- Harassment, discrimination, or personal attacks
- Trolling, insulting comments, or derogatory remarks
- Publishing others' private information without permission
- Other conduct which could reasonably be considered inappropriate

### Enforcement

Instances of unacceptable behavior may be reported to the project maintainers. All complaints will be reviewed and investigated promptly and fairly.

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.21+** installed
- **Node.js 20+** and npm installed
- **PostgreSQL 15+** running locally
- **Git** configured with your name and email
- Read the [Developer Guide](docs/DEVELOPER_GUIDE.md)

### Local Development Setup

1. **Fork and Clone**
   ```bash
   # Fork the repository on GitHub
   git clone https://github.com/YOUR_USERNAME/gorax.git
   cd gorax
   ```

2. **Set Up Upstream**
   ```bash
   git remote add upstream https://github.com/gorax/gorax.git
   git fetch upstream
   ```

3. **Environment Setup**
   ```bash
   # Copy environment template
   cp .env.example .env

   # Configure your local environment
   # Edit .env with your database credentials and settings
   ```

4. **Install Dependencies**
   ```bash
   # Backend
   go mod download

   # Frontend
   cd web
   npm install
   cd ..
   ```

5. **Database Setup**
   ```bash
   # Create database
   createdb gorax_dev

   # Run migrations
   make migrate-up
   ```

6. **Verify Setup**
   ```bash
   # Run tests
   make test

   # Start development servers
   make dev
   ```

For detailed setup instructions, see [docs/getting-started.md](docs/getting-started.md).

## Development Workflow

### Git Flow Process

**IMPORTANT**: We follow Git Flow. **Never commit directly to `main` or `dev` branches.**

1. **Create a Feature Branch**
   ```bash
   # Always branch from dev
   git checkout dev
   git pull upstream dev

   # Create feature branch with Jira ticket number
   git checkout -b RFLOW-123-add-webhook-filtering
   ```

2. **Branch Naming Convention**
   ```
   <ticket>-<short-description>

   Examples:
   - RFLOW-123-add-webhook-auth
   - RFLOW-456-fix-execution-timeout
   - RFLOW-789-update-readme
   ```

3. **Make Changes**
   - Follow TDD: Write tests first, then implementation
   - Keep commits small and focused
   - Write meaningful commit messages

4. **Commit Guidelines**
   ```bash
   # Format: <type>: <description>
   #
   # Types: feat, fix, docs, test, refactor, style, chore

   git commit -m "feat: add webhook signature verification

   - Implement HMAC-SHA256 signature validation
   - Add secret rotation endpoint
   - Include comprehensive tests"
   ```

5. **Keep Branch Updated**
   ```bash
   # Regularly sync with upstream dev
   git fetch upstream
   git rebase upstream/dev
   ```

6. **Push Changes**
   ```bash
   git push origin RFLOW-123-add-webhook-filtering
   ```

### Test-Driven Development (TDD)

**MANDATORY**: All new code must follow TDD.

1. **Red**: Write a failing test
2. **Green**: Write minimal code to pass
3. **Refactor**: Clean up while keeping tests green

Example workflow:
```bash
# 1. Write test
# Create internal/webhook/service_test.go with failing test

# 2. Run test (should fail)
go test ./internal/webhook/

# 3. Implement feature
# Edit internal/webhook/service.go

# 4. Run test (should pass)
go test ./internal/webhook/

# 5. Refactor if needed
```

See [docs/DEVELOPER_GUIDE.md#test-driven-development](docs/DEVELOPER_GUIDE.md) for detailed patterns.

## Pull Request Process

### Before Submitting

1. **Run Full Test Suite**
   ```bash
   # Backend tests
   make test

   # Frontend tests
   cd web && npm test

   # Integration tests (optional but recommended)
   TEST_DATABASE_URL=postgres://... make test-integration
   ```

2. **Run Linters**
   ```bash
   # Go linting
   golangci-lint run ./...

   # Frontend linting
   cd web && npm run lint
   ```

3. **Check Code Coverage**
   ```bash
   # Aim for 80%+ coverage on business logic
   make coverage
   ```

4. **Update Documentation**
   - Update relevant docs if you changed APIs
   - Add examples for new features
   - Update CHANGELOG.md if applicable

### Submitting a Pull Request

1. **Push to Your Fork**
   ```bash
   git push origin RFLOW-123-add-webhook-filtering
   ```

2. **Create PR on GitHub**
   - Base branch: `dev` (not `main`)
   - Title: Clear, descriptive summary
   - Description: Use the PR template

3. **PR Template**
   ```markdown
   ## Summary
   Brief description of changes

   ## Changes
   - Added webhook signature verification
   - Implemented secret rotation
   - Added comprehensive tests

   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests added/updated
   - [ ] Manual testing completed

   ## Documentation
   - [ ] Updated relevant documentation
   - [ ] Added code comments where needed

   ## Checklist
   - [ ] Tests pass locally
   - [ ] Linter passes
   - [ ] No new security vulnerabilities
   - [ ] Follows coding standards
   ```

4. **CI/CD Checks**
   - All tests must pass
   - Linter must pass
   - Code coverage must not decrease
   - Security scan must pass

5. **Code Review**
   - Address reviewer feedback promptly
   - Make requested changes in new commits
   - Don't force-push during review
   - Be open to constructive criticism

6. **Merge**
   - Maintainers will squash and merge approved PRs
   - Delete your feature branch after merge

### PR Review Guidelines

**What Reviewers Look For:**

- ✅ Tests cover new functionality (TDD followed)
- ✅ Code follows SOLID principles
- ✅ Functions are small and focused (< 50 lines)
- ✅ Cognitive complexity < 15 per function
- ✅ No code duplication (DRY principle)
- ✅ Error handling is complete
- ✅ Security best practices followed
- ✅ Documentation is updated
- ✅ Commit messages are clear

**Review Turnaround:**
- Initial review: within 2-3 business days
- Follow-up reviews: within 1-2 business days

## Coding Standards

### General Principles

Gorax follows clean code principles:

1. **SOLID Principles**
   - Single Responsibility
   - Open/Closed
   - Liskov Substitution
   - Interface Segregation
   - Dependency Inversion

2. **Clean Code**
   - Meaningful names
   - Small functions (< 20 lines preferred)
   - No comments for bad code (rewrite instead)
   - DRY (Don't Repeat Yourself)
   - YAGNI (You Aren't Gonna Need It)

3. **Complexity Limits**
   - Cognitive complexity: < 15
   - Cyclomatic complexity: < 10
   - Max function length: 50 lines
   - Max file length: 400 lines
   - Max parameters: 4 (use objects for more)

See [CLAUDE.md](CLAUDE.md) for comprehensive coding standards.

### Go-Specific Standards

```go
// ✅ Good: Small, focused function
func (s *Service) ValidateWebhook(ctx context.Context, id string) error {
    webhook, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("get webhook: %w", err)
    }

    return webhook.Validate()
}

// ❌ Bad: Too complex, does too much
func (s *Service) ProcessWebhook(ctx context.Context, id string) error {
    // 100+ lines of complex logic
    // Multiple responsibilities
    // High cognitive complexity
}
```

**Go Guidelines:**
- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Error wrapping: `fmt.Errorf("operation failed: %w", err)`
- Table-driven tests
- Dependency injection via constructors
- Interfaces defined where used, not where implemented

### TypeScript/React Standards

```typescript
// ✅ Good: Functional component with hooks
function WorkflowList({ workflows }: Props) {
  const { data, loading } = useWorkflows()

  if (loading) return <Loading />

  return (
    <div>
      {workflows.map(w => <WorkflowCard key={w.id} workflow={w} />)}
    </div>
  )
}

// ❌ Bad: Class component (we don't use these)
class WorkflowList extends React.Component {
  // Don't use class components
}
```

**TypeScript Guidelines:**
- Functional components only
- Custom hooks for reusable logic
- Zustand for global state
- React Query for server state
- No `any` types (use `unknown` if needed)
- Explicit function return types

## Testing Requirements

### Coverage Requirements

- **Business Logic**: 80%+ coverage
- **API Handlers**: 70%+ coverage
- **Utilities**: 90%+ coverage
- **UI Components**: 60%+ coverage

### Go Testing

```go
// Table-driven test pattern (preferred)
func TestValidateWorkflow(t *testing.T) {
    tests := []struct {
        name    string
        input   *Workflow
        wantErr bool
    }{
        {
            name:    "valid workflow",
            input:   &Workflow{Name: "Test"},
            wantErr: false,
        },
        {
            name:    "empty name",
            input:   &Workflow{Name: ""},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateWorkflow(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got err=%v, want err=%v", err, tt.wantErr)
            }
        })
    }
}
```

### React Testing

```typescript
// Component test with React Testing Library
describe('WorkflowList', () => {
  it('should render workflows', () => {
    const workflows = [
      { id: '1', name: 'Workflow 1' },
      { id: '2', name: 'Workflow 2' },
    ]

    render(<WorkflowList workflows={workflows} />)

    expect(screen.getByText('Workflow 1')).toBeInTheDocument()
    expect(screen.getByText('Workflow 2')).toBeInTheDocument()
  })
})
```

## Documentation

### What to Document

1. **Code Comments**
   - Public APIs and exported functions
   - Complex algorithms or business logic
   - Non-obvious design decisions
   - Security-critical code

2. **API Documentation**
   - All new endpoints in `docs/API_REFERENCE.md`
   - Request/response examples
   - Error codes and meanings

3. **User Documentation**
   - New features in user guides
   - Configuration options
   - Examples and tutorials

4. **Developer Documentation**
   - Architecture changes in `docs/architecture.md`
   - New patterns in `docs/DEVELOPER_GUIDE.md`
   - Migration guides for breaking changes

### Documentation Standards

```go
// ✅ Good: Clear documentation
// ValidateWebhook checks if a webhook configuration is valid.
// It verifies the webhook URL format, authentication settings,
// and filter expressions. Returns ErrInvalidWebhook if validation fails.
func (s *Service) ValidateWebhook(ctx context.Context, webhook *Webhook) error {
    // Implementation
}

// ❌ Bad: Unnecessary comment
// Get webhook by ID
func (s *Service) GetWebhook(ctx context.Context, id string) (*Webhook, error) {
    // Function name already says what it does
}
```

## Issue Reporting

### Bug Reports

Use the bug report template:

```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce:
1. Go to '...'
2. Click on '...'
3. See error

**Expected behavior**
What you expected to happen.

**Actual behavior**
What actually happened.

**Environment**
- OS: [e.g., Ubuntu 22.04]
- Gorax version: [e.g., v1.2.0]
- Go version: [e.g., 1.21]
- Browser: [e.g., Chrome 120]

**Logs**
```
Paste relevant logs here
```
```

### Feature Requests

Use the feature request template:

```markdown
**Is your feature request related to a problem?**
A clear description of the problem.

**Describe the solution you'd like**
What you want to happen.

**Describe alternatives you've considered**
Other solutions or features you've considered.

**Additional context**
Any other context, screenshots, or examples.
```

### Security Issues

**DO NOT** create public issues for security vulnerabilities.

Instead:
1. Email security@gorax.dev
2. Include detailed description
3. Wait for acknowledgment
4. Coordinate disclosure timeline

See [SECURITY.md](SECURITY.md) for details.

## Community

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and ideas
- **Slack**: [Join our Slack](https://gorax.slack.com) for real-time chat
- **Email**: dev@gorax.dev for private inquiries

### Getting Help

1. **Check Documentation**
   - [Getting Started](docs/getting-started.md)
   - [Developer Guide](docs/DEVELOPER_GUIDE.md)
   - [API Reference](docs/API_REFERENCE.md)

2. **Search Existing Issues**
   - Someone may have already asked your question

3. **Ask in Discussions**
   - For questions, not bugs

4. **Join Slack**
   - Real-time help from the community

### Recognition

Contributors are recognized in:
- Release notes
- Contributors list in README
- Monthly contributor highlights

Thank you for contributing to Gorax! 🚀

---

**Questions?** Open a discussion or join our Slack.

**Found a bug?** Open an issue with details.

**Want to contribute but don't know where to start?** Check issues labeled `good-first-issue`.
