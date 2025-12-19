# TDD Workflow Command

Start a Test-Driven Development workflow for implementing a feature.

## Instructions

Follow the strict TDD cycle for: $ARGUMENTS

### Phase 1: RED - Write Failing Tests

1. **Understand the requirement** - What behavior needs to be implemented?
2. **Identify test cases**:
   - Happy path (normal operation)
   - Edge cases (empty inputs, boundaries)
   - Error cases (invalid inputs, failures)
3. **Write test code FIRST**:
   - Use descriptive test names: `Test<Function>_<Scenario>_<ExpectedResult>`
   - One assertion per test (when possible)
   - Arrange-Act-Assert pattern
4. **Run tests** - Verify they FAIL (proves they test something)

### Phase 2: GREEN - Minimal Implementation

1. Write the **minimum code** to make tests pass
2. Don't optimize or refactor yet
3. Don't add features not covered by tests
4. Run tests - Verify they PASS

### Phase 3: REFACTOR - Clean Up

1. Remove duplication
2. Improve naming
3. Extract methods if needed
4. Keep tests passing after each change

## Test Structure Templates

### Go Test Template
```go
func TestFunctionName_Scenario_ExpectedBehavior(t *testing.T) {
    // Arrange
    input := setupTestData()
    expected := expectedResult()

    // Act
    result, err := FunctionUnderTest(input)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### TypeScript Test Template
```typescript
describe('FunctionName', () => {
  describe('when scenario', () => {
    it('should expected behavior', () => {
      // Arrange
      const input = setupTestData();
      const expected = expectedResult();

      // Act
      const result = functionUnderTest(input);

      // Assert
      expect(result).toEqual(expected);
    });
  });
});
```

## Output Format

For the requested feature, I will:

1. **List test cases** to be written
2. **Write the tests** (RED phase)
3. **Verify tests fail**
4. **Implement minimal code** (GREEN phase)
5. **Refactor** if needed
6. **Verify all tests pass**

Begin TDD workflow now.
