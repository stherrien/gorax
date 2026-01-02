# GitHub Actions Workflow Badges

Add these badges to your main `README.md` to display workflow status.

## 📊 Recommended Badge Layout

### Full Badge Set
```markdown
## CI/CD Status

![CI](https://github.com/stherrien/gorax/workflows/CI/badge.svg)
![Security Scanning](https://github.com/stherrien/gorax/workflows/Security%20Scanning/badge.svg)
![CodeQL](https://github.com/stherrien/gorax/workflows/CodeQL%20Analysis/badge.svg)
![Secrets Scanning](https://github.com/stherrien/gorax/workflows/Secrets%20Scanning/badge.svg)
![codecov](https://codecov.io/gh/stherrien/gorax/branch/main/graph/badge.svg)](https://codecov.io/gh/stherrien/gorax)
```

### Compact Layout (Recommended)
```markdown
[![CI](https://github.com/stherrien/gorax/workflows/CI/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/ci.yml)
[![Security](https://github.com/stherrien/gorax/workflows/Security%20Scanning/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/stherrien/gorax/branch/main/graph/badge.svg)](https://codecov.io/gh/stherrien/gorax)
```

## 🎨 Individual Badges

### CI Pipeline
```markdown
![CI](https://github.com/stherrien/gorax/workflows/CI/badge.svg)
```
**Preview**: ![CI](https://github.com/stherrien/gorax/workflows/CI/badge.svg)

### Security Scanning
```markdown
![Security Scanning](https://github.com/stherrien/gorax/workflows/Security%20Scanning/badge.svg)
```
**Preview**: ![Security Scanning](https://github.com/stherrien/gorax/workflows/Security%20Scanning/badge.svg)

### CodeQL Analysis
```markdown
![CodeQL](https://github.com/stherrien/gorax/workflows/CodeQL%20Analysis/badge.svg)
```
**Preview**: ![CodeQL](https://github.com/stherrien/gorax/workflows/CodeQL%20Analysis/badge.svg)

### Secrets Scanning
```markdown
![Secrets Scanning](https://github.com/stherrien/gorax/workflows/Secrets%20Scanning/badge.svg)
```
**Preview**: ![Secrets Scanning](https://github.com/stherrien/gorax/workflows/Secrets%20Scanning/badge.svg)

### Staging Deploy
```markdown
![Deploy Staging](https://github.com/stherrien/gorax/workflows/Deploy%20to%20Staging/badge.svg)
```
**Preview**: ![Deploy Staging](https://github.com/stherrien/gorax/workflows/Deploy%20to%20Staging/badge.svg)

### Production Deploy
```markdown
![Deploy Production](https://github.com/stherrien/gorax/workflows/Deploy%20to%20Production/badge.svg)
```
**Preview**: ![Deploy Production](https://github.com/stherrien/gorax/workflows/Deploy%20to%20Production/badge.svg)

### Release
```markdown
![Release](https://github.com/stherrien/gorax/workflows/Release/badge.svg)
```
**Preview**: ![Release](https://github.com/stherrien/gorax/workflows/Release/badge.svg)

### Codecov
```markdown
[![codecov](https://codecov.io/gh/stherrien/gorax/branch/main/graph/badge.svg)](https://codecov.io/gh/stherrien/gorax)
```
**Preview**: [![codecov](https://codecov.io/gh/stherrien/gorax/branch/main/graph/badge.svg)](https://codecov.io/gh/stherrien/gorax)

## 🔗 Clickable Badges (With Links)

### CI with Link
```markdown
[![CI](https://github.com/stherrien/gorax/workflows/CI/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/ci.yml)
```

### Security with Link
```markdown
[![Security](https://github.com/stherrien/gorax/workflows/Security%20Scanning/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/security.yml)
```

### CodeQL with Link
```markdown
[![CodeQL](https://github.com/stherrien/gorax/workflows/CodeQL%20Analysis/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/codeql.yml)
```

## 📋 Branch-Specific Badges

### Main Branch
```markdown
![CI - Main](https://github.com/stherrien/gorax/workflows/CI/badge.svg?branch=main)
```

### Dev Branch
```markdown
![CI - Dev](https://github.com/stherrien/gorax/workflows/CI/badge.svg?branch=dev)
```

## 🎯 Custom Badge Styles

### Flat Style
```markdown
![CI](https://img.shields.io/github/actions/workflow/status/stherrien/gorax/ci.yml?style=flat&label=CI)
```

### Flat Square Style
```markdown
![CI](https://img.shields.io/github/actions/workflow/status/stherrien/gorax/ci.yml?style=flat-square&label=CI)
```

### For the Badge Style
```markdown
![CI](https://img.shields.io/github/actions/workflow/status/stherrien/gorax/ci.yml?style=for-the-badge&label=CI)
```

## 📊 Coverage Badges

### Line Coverage
```markdown
![Coverage](https://codecov.io/gh/stherrien/gorax/branch/main/graph/badge.svg?token=YOUR_TOKEN)
```

### Custom Coverage Badge (shields.io)
```markdown
![Coverage](https://img.shields.io/codecov/c/github/stherrien/gorax/main?label=coverage&logo=codecov)
```

## 🏷️ Additional Badges

### Go Version
```markdown
![Go Version](https://img.shields.io/github/go-mod/go-version/stherrien/gorax)
```

### License
```markdown
![License](https://img.shields.io/github/license/stherrien/gorax)
```

### Release Version
```markdown
![Release](https://img.shields.io/github/v/release/stherrien/gorax)
```

### Last Commit
```markdown
![Last Commit](https://img.shields.io/github/last-commit/stherrien/gorax)
```

### Contributors
```markdown
![Contributors](https://img.shields.io/github/contributors/stherrien/gorax)
```

### Issues
```markdown
![Issues](https://img.shields.io/github/issues/stherrien/gorax)
```

### Pull Requests
```markdown
![PRs](https://img.shields.io/github/issues-pr/stherrien/gorax)
```

## 🎨 Complete README Example

```markdown
# Gorax

> Modern Workflow Automation Platform

[![CI](https://github.com/stherrien/gorax/workflows/CI/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/ci.yml)
[![Security](https://github.com/stherrien/gorax/workflows/Security%20Scanning/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/security.yml)
[![CodeQL](https://github.com/stherrien/gorax/workflows/CodeQL%20Analysis/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/stherrien/gorax/branch/main/graph/badge.svg)](https://codecov.io/gh/stherrien/gorax)
[![Go Version](https://img.shields.io/github/go-mod/go-version/stherrien/gorax)](https://go.dev/)
[![License](https://img.shields.io/github/license/stherrien/gorax)](LICENSE)

## Overview

[Your project description here]

## Quick Start

[Your quick start guide here]

## Documentation

- [Developer Guide](docs/DEVELOPER_GUIDE.md)
- [CI/CD Guide](.github/workflows/CICD_GUIDE.md)
- [Security Practices](docs/WEBSOCKET_SECURITY.md)

## Status

| Environment | Status | URL |
|-------------|--------|-----|
| Production | ![Production](https://img.shields.io/badge/status-live-success) | https://gorax.dev |
| Staging | ![Staging](https://img.shields.io/badge/status-live-success) | https://staging.gorax.dev |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.
```

## 🔄 Dynamic Status Page

Create a status page with live workflow statuses:

```markdown
# Workflow Status Dashboard

## 🟢 Core Pipelines
| Workflow | Status | Last Run | Duration |
|----------|--------|----------|----------|
| CI | ![CI](https://github.com/stherrien/gorax/workflows/CI/badge.svg) | [View Runs](https://github.com/stherrien/gorax/actions/workflows/ci.yml) | ~10 min |
| Security | ![Security](https://github.com/stherrien/gorax/workflows/Security%20Scanning/badge.svg) | [View Runs](https://github.com/stherrien/gorax/actions/workflows/security.yml) | ~6 min |
| CodeQL | ![CodeQL](https://github.com/stherrien/gorax/workflows/CodeQL%20Analysis/badge.svg) | [View Runs](https://github.com/stherrien/gorax/actions/workflows/codeql.yml) | ~15 min |

## 🚀 Deployment Pipelines
| Environment | Status | Last Deploy | Link |
|-------------|--------|-------------|------|
| Staging | ![Staging](https://github.com/stherrien/gorax/workflows/Deploy%20to%20Staging/badge.svg) | [View Runs](https://github.com/stherrien/gorax/actions/workflows/deploy-staging.yml) | [staging.gorax.dev](https://staging.gorax.dev) |
| Production | ![Production](https://github.com/stherrien/gorax/workflows/Deploy%20to%20Production/badge.svg) | [View Runs](https://github.com/stherrien/gorax/actions/workflows/deploy-production.yml) | [gorax.dev](https://gorax.dev) |

## 📦 Release Pipeline
| Workflow | Status | Latest Release |
|----------|--------|----------------|
| Release | ![Release](https://github.com/stherrien/gorax/workflows/Release/badge.svg) | ![Latest](https://img.shields.io/github/v/release/stherrien/gorax) |

## 📊 Quality Metrics
| Metric | Status |
|--------|--------|
| Code Coverage | [![codecov](https://codecov.io/gh/stherrien/gorax/branch/main/graph/badge.svg)](https://codecov.io/gh/stherrien/gorax) |
| Go Report | [![Go Report Card](https://goreportcard.com/badge/github.com/stherrien/gorax)](https://goreportcard.com/report/github.com/stherrien/gorax) |
| Dependencies | ![Dependencies](https://img.shields.io/librariesio/github/stherrien/gorax) |
```

## 🛠️ Customization

### Update Repository Owner/Name

Replace `stherrien/gorax` with your repository in all badge URLs:

```markdown
# From
https://github.com/stherrien/gorax/workflows/CI/badge.svg

# To
https://github.com/YOUR_ORG/YOUR_REPO/workflows/CI/badge.svg
```

### Update Branch

For badges specific to a branch:

```markdown
# Main branch
?branch=main

# Dev branch
?branch=dev

# Feature branch
?branch=feature/new-feature
```

### Add Custom Labels

```markdown
# Custom label
![CI](https://img.shields.io/github/actions/workflow/status/stherrien/gorax/ci.yml?label=Build%20Status)

# With emoji
![CI](https://img.shields.io/github/actions/workflow/status/stherrien/gorax/ci.yml?label=🚀%20Build)
```

## 📝 Markdown Table Example

Create a status table in your README:

```markdown
## Pipeline Status

| Pipeline | Status | Purpose |
|----------|--------|---------|
| CI | ![CI](https://github.com/stherrien/gorax/workflows/CI/badge.svg) | Continuous Integration |
| Security | ![Security](https://github.com/stherrien/gorax/workflows/Security%20Scanning/badge.svg) | Security Scanning |
| CodeQL | ![CodeQL](https://github.com/stherrien/gorax/workflows/CodeQL%20Analysis/badge.svg) | Static Analysis |
| Coverage | [![codecov](https://codecov.io/gh/stherrien/gorax/branch/main/graph/badge.svg)](https://codecov.io/gh/stherrien/gorax) | Test Coverage |
```

## 🎯 Best Practices

1. **Keep It Simple**: Don't overwhelm with too many badges
2. **Most Important First**: Lead with CI and security badges
3. **Make Them Clickable**: Link badges to workflow pages
4. **Branch Awareness**: Show main branch status by default
5. **Update Regularly**: Keep badge URLs current when renaming workflows
6. **Consistent Styling**: Use the same style across all badges

## 📚 Resources

- [GitHub Actions Badge](https://docs.github.com/en/actions/monitoring-and-troubleshooting-workflows/adding-a-workflow-status-badge)
- [Shields.io](https://shields.io/) - Custom badge generator
- [Simple Icons](https://simpleicons.org/) - Icons for badges

---

**Note**: Replace `stherrien/gorax` with your actual GitHub organization/username and repository name in all badge URLs.
