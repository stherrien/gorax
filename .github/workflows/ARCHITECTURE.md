# CI/CD Pipeline Architecture

## Workflow Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         GitHub Events                            │
│  Push • Pull Request • Tag • Schedule • Manual Trigger          │
└────────────────────────┬────────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┬────────────────┐
         │               │               │                │
         ▼               ▼               ▼                ▼
┌────────────────┐ ┌────────────┐ ┌────────────┐ ┌──────────────┐
│   CI Workflow  │ │  Security  │ │   CodeQL   │ │   Release    │
│                │ │  Workflow  │ │  Workflow  │ │   Workflow   │
└────────┬───────┘ └──────┬─────┘ └──────┬─────┘ └──────┬───────┘
         │                │               │               │
         │                │               │               │
    ┌────▼────┐      ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
    │ Backend │      │  Vuln   │    │   Go    │    │  Build  │
    │  Tests  │      │  Scan   │    │ Analysis│    │ Binaries│
    └────┬────┘      └────┬────┘    └────┬────┘    └────┬────┘
         │                │               │               │
    ┌────▼────┐      ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
    │Frontend │      │Container│    │TypeScript│   │  Build  │
    │  Tests  │      │  Scan   │    │ Analysis│    │ Docker  │
    └────┬────┘      └────┬────┘    └────┬────┘    └────┬────┘
         │                │               │               │
    ┌────▼────┐      ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
    │   Go    │      │   Dep   │    │Security │    │ Create  │
    │  Lint   │      │  Review │    │  Report │    │ Release │
    └────┬────┘      └─────────┘    └─────────┘    └─────────┘
         │
    ┌────▼────┐
    │Frontend │
    │  Lint   │
    └────┬────┘
         │
    ┌────▼────┐
    │  Build  │
    │ Verify  │
    └─────────┘
```

## Data Flow

### CI Pipeline Flow
```
Developer Push
      │
      ▼
┌─────────────┐
│   Checkout  │
└──────┬──────┘
       │
       ├──────────────┬──────────────┬──────────────┐
       │              │              │              │
       ▼              ▼              ▼              ▼
  ┌────────┐    ┌─────────┐   ┌─────────┐   ┌─────────┐
  │  Go    │    │Frontend │   │   Go    │   │Frontend │
  │ Tests  │    │  Tests  │   │  Lint   │   │  Lint   │
  │        │    │         │   │         │   │         │
  │  ✅✅   │    │   ✅✅   │   │   ✅✅   │   │   ✅✅   │
  └────┬───┘    └────┬────┘   └────┬────┘   └────┬────┘
       │             │             │              │
       └─────────────┴─────────────┴──────────────┘
                     │
                     ▼
              ┌────────────┐
              │   Build    │
              │  Verify    │
              └─────┬──────┘
                    │
                    ▼
              ┌────────────┐
              │  Artifacts │
              │   Upload   │
              └────────────┘
```

### Security Pipeline Flow
```
PR or Schedule
      │
      ▼
┌─────────────────────────────────────┐
│        Parallel Security Scans       │
└─────────────────────────────────────┘
      │
      ├────┬────┬────┬────┬────┐
      │    │    │    │    │    │
      ▼    ▼    ▼    ▼    ▼    ▼
    ┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐
    │gov││gos││npm││Tri││Tri││Dep│
    │uln││ec ││aud││vy ││vy ││Rev│
    │chk││   ││it ││Repo││Dkr││iew│
    └─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘
      │    │    │    │    │    │
      └────┴────┴────┴────┴────┘
                │
                ▼
          ┌───────────┐
          │  SARIF    │
          │  Upload   │
          └─────┬─────┘
                │
                ▼
          ┌───────────┐
          │  Security │
          │    Tab    │
          └───────────┘
```

### Release Pipeline Flow
```
Tag Push (v*)
      │
      ▼
┌─────────────────────────────────────┐
│     Parallel Platform Builds         │
└─────────────────────────────────────┘
      │
      ├──────────────┬──────────────┬──────────────┐
      │              │              │              │
      ▼              ▼              ▼              ▼
  ┌────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
  │ Linux  │   │ macOS   │   │Windows  │   │Frontend │
  │amd64   │   │ amd64   │   │ amd64   │   │  Build  │
  │arm64   │   │ arm64   │   │         │   │         │
  └────┬───┘   └────┬────┘   └────┬────┘   └────┬────┘
       │            │             │              │
       └────────────┴─────────────┴──────────────┘
                    │
                    ▼
              ┌────────────┐
              │   Docker   │
              │  Multi-Arch│
              │   Build    │
              └─────┬──────┘
                    │
                    ▼
              ┌────────────┐
              │  Generate  │
              │ Changelog  │
              └─────┬──────┘
                    │
                    ▼
              ┌────────────┐
              │   Create   │
              │  Release   │
              └─────┬──────┘
                    │
                    ▼
              ┌────────────┐
              │   Publish  │
              │  Artifacts │
              └────────────┘
```

### Staging Deployment Flow
```
Push to dev
      │
      ▼
┌──────────────┐
│Build Docker  │
│    Image     │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│Push to       │
│  Registry    │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  Deploy to   │
│   Staging    │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  Run DB      │
│ Migrations   │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Health     │
│   Checks     │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Smoke      │
│   Tests      │
└──────────────┘
```

## Component Relationships

### Services and Dependencies
```
┌─────────────────────────────────────────────────────────────┐
│                     CI Workflow                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────┐        ┌──────────────────────┐  │
│  │   PostgreSQL:14      │◄───────│   Go Test Runner     │  │
│  │   Port: 5432         │        │   Database Tests     │  │
│  └──────────────────────┘        └──────────────────────┘  │
│                                                              │
│  ┌──────────────────────┐        ┌──────────────────────┐  │
│  │   Redis:6            │◄───────│   Go Test Runner     │  │
│  │   Port: 6379         │        │   Cache Tests        │  │
│  └──────────────────────┘        └──────────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Cache Strategy
```
┌─────────────────────────────────────────┐
│           Caching Layers                 │
├─────────────────────────────────────────┤
│                                          │
│  Go Modules Cache                        │
│  ├─ Key: go.sum hash                     │
│  ├─ Path: ~/go/pkg/mod                   │
│  └─ Path: ~/.cache/go-build              │
│                                          │
│  npm Cache                               │
│  ├─ Key: package-lock.json hash          │
│  └─ Path: ~/.npm                         │
│                                          │
│  Docker Layer Cache                      │
│  ├─ Type: gha                            │
│  └─ Mode: max                            │
│                                          │
└─────────────────────────────────────────┘
```

## Artifact Flow

```
┌──────────────────────────────────────────────────────────────┐
│                    Artifact Generation                        │
└──────────────────────────────────────────────────────────────┘
        │
        ├─────────────────┬─────────────────┬─────────────────┐
        │                 │                 │                 │
        ▼                 ▼                 ▼                 ▼
   ┌─────────┐      ┌──────────┐     ┌──────────┐     ┌──────────┐
   │Coverage │      │  Build   │     │Security  │     │ Release  │
   │ Reports │      │Artifacts │     │ Reports  │     │ Binaries │
   └────┬────┘      └────┬─────┘     └────┬─────┘     └────┬─────┘
        │                │                 │                 │
        └────────────────┴─────────────────┴─────────────────┘
                         │
                         ▼
                ┌─────────────────┐
                │  GitHub Actions │
                │  Artifact Store │
                └────────┬────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
    ┌─────────┐    ┌─────────┐    ┌─────────┐
    │Codecov  │    │ GitHub  │    │Container│
    │         │    │ Release │    │Registry │
    └─────────┘    └─────────┘    └─────────┘
```

## Security Scanning Matrix

```
┌────────────────────────────────────────────────────────────────┐
│              Security Scanning Coverage                         │
├────────────┬───────────┬──────────┬──────────────────────────┤
│   Tool     │  Target   │  Scope   │      Detects             │
├────────────┼───────────┼──────────┼──────────────────────────┤
│govulncheck │ Go deps   │ CVE DB   │ Known vulnerabilities    │
│gosec       │ Go code   │ SAST     │ Security anti-patterns   │
│npm audit   │ npm deps  │ CVE DB   │ Package vulnerabilities  │
│Trivy       │ Repo      │ SAST     │ Config issues, secrets   │
│Trivy       │ Container │ CVE DB   │ OS packages, libs        │
│CodeQL      │ Go code   │ SAST     │ Code vulnerabilities     │
│CodeQL      │ TS code   │ SAST     │ Frontend vulnerabilities │
│Dep Review  │ Changes   │ Diff     │ New risky dependencies   │
└────────────┴───────────┴──────────┴──────────────────────────┘
```

## Workflow Trigger Matrix

```
┌──────────────────────────────────────────────────────────────┐
│                   Workflow Triggers                           │
├─────────────┬──────┬────┬──────┬──────┬────────┬──────────┤
│  Workflow   │ Push │ PR │ Tag  │Sched │ Manual │ Comments │
├─────────────┼──────┼────┼──────┼──────┼────────┼──────────┤
│CI           │  ✅  │ ✅ │  -   │  -   │   -    │Every push│
│Security     │main  │ ✅ │  -   │  ✅  │   ✅   │Weekly    │
│             │dev   │    │      │      │        │          │
│CodeQL       │main  │ ✅ │  -   │  ✅  │   ✅   │Weekly    │
│             │dev   │    │      │      │        │          │
│Release      │  -   │ -  │  ✅  │  -   │   ✅   │On v* tag │
│Deploy Stage │dev   │ -  │  -   │  -   │   ✅   │Auto dev  │
└─────────────┴──────┴────┴──────┴──────┴────────┴──────────┘
```

## Performance Characteristics

### Average Execution Times
```
┌────────────────────────────────────────────┐
│        Workflow Duration Estimates          │
├──────────────────┬─────────────────────────┤
│ CI Workflow      │  5-10 minutes           │
│ ├─ Go Tests      │  2-3 minutes            │
│ ├─ Go Lint       │  1-2 minutes            │
│ ├─ Frontend Tests│  1-2 minutes            │
│ ├─ Frontend Lint │  1 minute               │
│ └─ Build         │  2-3 minutes            │
├──────────────────┼─────────────────────────┤
│ Security         │  8-12 minutes           │
│ ├─ govulncheck   │  1-2 minutes            │
│ ├─ gosec         │  1-2 minutes            │
│ ├─ npm audit     │  1 minute               │
│ ├─ Trivy Repo    │  2-3 minutes            │
│ ├─ Trivy Docker  │  2-3 minutes            │
│ └─ Dep Review    │  1 minute               │
├──────────────────┼─────────────────────────┤
│ CodeQL           │  10-15 minutes          │
│ ├─ Go Analysis   │  5-7 minutes            │
│ └─ TS Analysis   │  5-8 minutes            │
├──────────────────┼─────────────────────────┤
│ Release          │  20-30 minutes          │
│ ├─ Binaries      │  10-15 minutes (matrix) │
│ ├─ Frontend      │  2-3 minutes            │
│ ├─ Docker        │  5-8 minutes            │
│ └─ Release       │  2-3 minutes            │
├──────────────────┼─────────────────────────┤
│ Deploy Staging   │  5-10 minutes           │
│ ├─ Build         │  3-5 minutes            │
│ ├─ Deploy        │  1-2 minutes            │
│ └─ Smoke Tests   │  1-3 minutes            │
└──────────────────┴─────────────────────────┘
```

## Concurrency Control

```
┌─────────────────────────────────────────────────────────┐
│           Concurrency Group Strategy                     │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Each workflow has its own concurrency group:            │
│                                                          │
│  CI:       ${{ github.workflow }}-${{ github.ref }}     │
│  Security: ${{ github.workflow }}-${{ github.ref }}     │
│  CodeQL:   ${{ github.workflow }}-${{ github.ref }}     │
│                                                          │
│  Effect: New push cancels previous run on same branch   │
│                                                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │  Branch: feature/xyz                            │    │
│  │                                                  │    │
│  │  Push 1 ──► [CI Running] ──► Cancelled          │    │
│  │  Push 2 ──► [CI Running] ──► Cancelled          │    │
│  │  Push 3 ──► [CI Running] ──► Complete ✅         │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

## Integration Points

```
┌────────────────────────────────────────────────────────────┐
│              External Service Integrations                  │
├────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐       ┌────────────────────────────┐    │
│  │   Codecov    │◄──────│  Coverage Reports          │    │
│  └──────────────┘       └────────────────────────────┘    │
│                                                             │
│  ┌──────────────┐       ┌────────────────────────────┐    │
│  │   GHCR       │◄──────│  Docker Images             │    │
│  │  (Registry)  │       └────────────────────────────┘    │
│  └──────────────┘                                          │
│                                                             │
│  ┌──────────────┐       ┌────────────────────────────┐    │
│  │   GitHub     │◄──────│  SARIF Reports             │    │
│  │  Security    │       │  (gosec, Trivy, CodeQL)    │    │
│  └──────────────┘       └────────────────────────────┘    │
│                                                             │
│  ┌──────────────┐       ┌────────────────────────────┐    │
│  │   GitHub     │◄──────│  Release Artifacts         │    │
│  │  Releases    │       │  Binaries, Checksums       │    │
│  └──────────────┘       └────────────────────────────┘    │
│                                                             │
└────────────────────────────────────────────────────────────┘
```

## Error Handling and Notifications

```
┌─────────────────────────────────────────────────────────┐
│            Failure Handling Strategy                     │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Critical Failures (Block Merge):                        │
│  ├─ Go Tests Fail                                        │
│  ├─ Frontend Tests Fail                                  │
│  ├─ Build Fails                                          │
│  └─ Go/Frontend Lint Fails                               │
│                                                          │
│  Non-Critical Failures (Warning Only):                   │
│  ├─ Codecov Upload Fails    (continue-on-error: true)   │
│  ├─ Security Scan Fails     (report but don't block)    │
│  └─ Artifact Upload Fails   (continue-on-error: true)   │
│                                                          │
│  Notification Channels:                                  │
│  ├─ GitHub PR Checks        (automatic)                  │
│  ├─ GitHub Actions UI       (automatic)                  │
│  └─ Email to Committer      (GitHub setting)            │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## Key Design Decisions

1. **Parallel Execution**: Jobs run in parallel when possible to minimize total time
2. **Caching Strategy**: Aggressive caching of dependencies to speed up builds
3. **Matrix Builds**: Used for cross-platform binary compilation
4. **Security-First**: Multiple overlapping security scans for defense in depth
5. **Fail-Fast**: Critical tests fail early to save CI minutes
6. **Artifact Retention**: 7 days for CI, 90 days for releases
7. **Concurrency Control**: Auto-cancel outdated runs to save resources
8. **Non-Blocking Features**: Optional integrations don't block core workflows

---

**Maintained By**: Gorax DevOps Team
**Last Updated**: 2025-12-20
