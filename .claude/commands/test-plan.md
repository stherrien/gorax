# Test Plan Command

Generate a comprehensive test plan for a feature or component.

## Instructions

Create a test plan for: $ARGUMENTS

### Test Categories

#### 1. Unit Tests
- Test individual functions/methods in isolation
- Mock all dependencies
- Focus on business logic

#### 2. Integration Tests
- Test component interactions
- Use real database (test container)
- Test API endpoints end-to-end

#### 3. Edge Cases
- Empty inputs
- Null/undefined values
- Boundary conditions (max/min values)
- Large inputs

#### 4. Error Cases
- Invalid inputs
- Network failures
- Database errors
- Timeout scenarios

#### 5. Security Tests
- Authentication required
- Authorization checks
- Input validation
- SQL injection prevention

### Output Format

```
## Test Plan: [Feature Name]

### Overview
[Brief description of what's being tested]

### Unit Tests

#### [Component/Function Name]
| Test Case | Input | Expected Output | Priority |
|-----------|-------|-----------------|----------|
| Happy path - normal operation | valid input | expected result | High |
| Edge case - empty input | [] | empty result | Medium |
| Error case - invalid input | null | throws Error | High |

### Integration Tests

| Test Case | Setup | Action | Verification |
|-----------|-------|--------|--------------|
| Create workflow via API | Auth token | POST /workflows | 201, workflow in DB |

### Test Data Requirements
- [What test fixtures are needed]
- [What mocks are required]

### Coverage Goals
- Unit test coverage: 80%+
- Critical paths: 100%

### Test File Locations
- Unit tests: `path/to/tests`
- Integration tests: `path/to/integration`
```

Generate test plan now.
