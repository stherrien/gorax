# Code Review Command

Review the specified code for quality issues based on project standards.

## Instructions

Analyze the code provided or the files mentioned and check for:

### 1. SOLID Principles Violations
- Single Responsibility: Does each class/function do one thing?
- Open/Closed: Can it be extended without modification?
- Liskov Substitution: Are subtypes properly substitutable?
- Interface Segregation: Are interfaces focused?
- Dependency Inversion: Are dependencies abstracted?

### 2. Clean Code Issues
- Are names meaningful and intention-revealing?
- Are functions small (< 20 lines preferred)?
- Is there code duplication (DRY violations)?
- Are there unnecessary comments explaining bad code?
- Is there dead/commented-out code?

### 3. Complexity Analysis
- Calculate cognitive complexity (flag if > 15)
- Check cyclomatic complexity (flag if > 10)
- Identify deeply nested code (> 3 levels)
- Flag functions longer than 50 lines

### 4. Testing Gaps
- Are there tests for this code?
- What test cases are missing?
- Are edge cases covered?

### 5. Error Handling
- Are errors properly handled?
- Is error context preserved?
- Are errors logged appropriately?

### 6. Security Concerns
- Input validation
- SQL injection risks
- XSS vulnerabilities
- Hardcoded secrets

## Output Format

Provide a structured review:

```
## Code Review Summary

**Overall Quality**: [Good/Needs Work/Major Issues]

### Issues Found

#### Critical
- [Issue description with file:line reference]

#### Major
- [Issue description]

#### Minor
- [Issue description]

### Recommendations
1. [Specific actionable recommendation]
2. [...]

### Positive Aspects
- [What's done well]
```

Review the code now: $ARGUMENTS
