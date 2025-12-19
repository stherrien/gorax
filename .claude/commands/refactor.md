# Refactor Command

Analyze code and suggest refactoring opportunities.

## Instructions

Analyze the specified code: $ARGUMENTS

### Identify Refactoring Opportunities

#### 1. Extract Method
When to apply:
- Function is too long (> 20 lines)
- Code block has a comment explaining what it does
- Same code appears in multiple places

#### 2. Extract Variable
When to apply:
- Complex expression that's hard to understand
- Same expression used multiple times
- Magic numbers or strings

#### 3. Replace Conditional with Polymorphism
When to apply:
- Switch statement that selects behavior based on type
- Multiple if-else chains checking the same condition
- Type codes that determine behavior

#### 4. Introduce Parameter Object
When to apply:
- Function has more than 4 parameters
- Same group of parameters passed to multiple functions

#### 5. Replace Nested Conditionals with Guard Clauses
When to apply:
- Deeply nested if statements
- Complex condition checking

Before:
```go
func process(item Item) error {
    if item != nil {
        if item.IsValid() {
            if item.CanProcess() {
                return item.Process()
            }
        }
    }
    return errors.New("cannot process")
}
```

After:
```go
func process(item Item) error {
    if item == nil {
        return errors.New("item is nil")
    }
    if !item.IsValid() {
        return errors.New("item is invalid")
    }
    if !item.CanProcess() {
        return errors.New("item cannot be processed")
    }
    return item.Process()
}
```

#### 6. Decompose Conditional
When to apply:
- Complex boolean conditions
- Conditions that need comments to explain

#### 7. Replace Magic Numbers with Constants
When to apply:
- Literal numbers in code without clear meaning
- Same number used in multiple places

### Output Format

```
## Refactoring Analysis

### Current Issues
1. [Issue with code reference]

### Recommended Refactorings

#### Refactoring 1: [Name]
**Location**: file:line
**Reason**: [Why this improves the code]
**Before**:
```
[current code]
```
**After**:
```
[refactored code]
```

### Impact Assessment
- Complexity reduction: [Before] → [After]
- Lines of code: [Before] → [After]
- Test impact: [What tests need updating]
```

Analyze and suggest refactorings now.
