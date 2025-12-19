# Architecture Review Command

Review code architecture and suggest improvements.

## Instructions

Analyze the architecture of: $ARGUMENTS

### Check for Architectural Issues

#### 1. Layer Violations
- Is business logic leaking into handlers/controllers?
- Is data access logic mixed with business logic?
- Are there circular dependencies between packages?

#### 2. Dependency Direction
```
Correct flow:
  Handler → Service → Repository → Database

Wrong:
  Repository → Service (repository shouldn't know about service)
  Handler → Database (skipping layers)
```

#### 3. Package Cohesion
- Does each package have a single, clear purpose?
- Are related types grouped together?
- Are there packages that are too large?

#### 4. Interface Design
- Are interfaces defined where they're used (not where implemented)?
- Are interfaces small and focused?
- Do interfaces represent behavior, not data?

#### 5. Error Handling Strategy
- Are errors wrapped with context?
- Is there a consistent error handling pattern?
- Are domain errors distinguished from infrastructure errors?

#### 6. Configuration Management
- Is configuration centralized?
- Are secrets handled securely?
- Is there environment-specific configuration?

### Architecture Patterns to Verify

#### Clean Architecture Layers
```
┌─────────────────────────────────────┐
│         Handlers/Controllers        │  ← HTTP, gRPC, CLI
├─────────────────────────────────────┤
│         Application Services        │  ← Use cases, orchestration
├─────────────────────────────────────┤
│           Domain/Business           │  ← Core business logic
├─────────────────────────────────────┤
│         Infrastructure              │  ← DB, external APIs
└─────────────────────────────────────┘
```

#### Dependency Injection
- Are dependencies injected via constructors?
- Are there any hard-coded dependencies?
- Can components be tested in isolation?

### Output Format

```
## Architecture Review

### Current Structure
[Diagram or description of current architecture]

### Issues Found

#### Critical
- [Architectural violation with impact]

#### Recommendations
1. [Specific improvement with rationale]

### Suggested Refactoring
[Code or structural changes to improve architecture]
```

Analyze architecture now.
