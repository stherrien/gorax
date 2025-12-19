# Generate Tests Command

Generate comprehensive tests for the specified code.

## Instructions

Generate tests for: $ARGUMENTS

### Test Generation Process

1. **Analyze the code** - Understand what it does
2. **Identify test cases** - Happy path, edge cases, error cases
3. **Generate test code** - Following project conventions
4. **Include mocks** - For external dependencies

### Test Case Categories

#### Happy Path Tests
- Normal operation with valid inputs
- Expected outputs are produced
- Side effects occur correctly

#### Edge Case Tests
- Empty inputs (nil, empty string, empty array)
- Boundary values (0, -1, max int, etc.)
- Single element collections
- Unicode/special characters in strings

#### Error Case Tests
- Invalid inputs
- Missing required fields
- Unauthorized access
- External service failures
- Timeout scenarios

#### Integration Points
- Database operations
- HTTP calls
- Message queue interactions
- File system operations

### Go Test Template

```go
package foo_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

// Mock for dependencies
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*Entity, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Entity), args.Error(1)
}

func TestService_GetEntity(t *testing.T) {
    tests := []struct {
        name      string
        id        string
        mockSetup func(*MockRepository)
        want      *Entity
        wantErr   bool
        errMsg    string
    }{
        {
            name: "success - returns entity",
            id:   "123",
            mockSetup: func(m *MockRepository) {
                m.On("GetByID", mock.Anything, "123").Return(&Entity{ID: "123"}, nil)
            },
            want:    &Entity{ID: "123"},
            wantErr: false,
        },
        {
            name: "error - entity not found",
            id:   "404",
            mockSetup: func(m *MockRepository) {
                m.On("GetByID", mock.Anything, "404").Return(nil, ErrNotFound)
            },
            want:    nil,
            wantErr: true,
            errMsg:  "not found",
        },
        {
            name: "error - empty id",
            id:   "",
            mockSetup: func(m *MockRepository) {},
            want:    nil,
            wantErr: true,
            errMsg:  "id is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            mockRepo := new(MockRepository)
            tt.mockSetup(mockRepo)
            svc := NewService(mockRepo)

            // Act
            got, err := svc.GetEntity(context.Background(), tt.id)

            // Assert
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### TypeScript/React Test Template

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

// Mock API
vi.mock('../api/client', () => ({
  api: {
    workflows: {
      get: vi.fn(),
      create: vi.fn(),
    },
  },
}));

import { api } from '../api/client';
import { WorkflowEditor } from './WorkflowEditor';

describe('WorkflowEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('when loading workflow', () => {
    it('should display loading state initially', () => {
      vi.mocked(api.workflows.get).mockImplementation(() => new Promise(() => {}));

      render(<WorkflowEditor id="123" />);

      expect(screen.getByText(/loading/i)).toBeInTheDocument();
    });

    it('should display workflow after loading', async () => {
      vi.mocked(api.workflows.get).mockResolvedValue({
        id: '123',
        name: 'Test Workflow',
      });

      render(<WorkflowEditor id="123" />);

      await waitFor(() => {
        expect(screen.getByText('Test Workflow')).toBeInTheDocument();
      });
    });

    it('should display error on failure', async () => {
      vi.mocked(api.workflows.get).mockRejectedValue(new Error('Network error'));

      render(<WorkflowEditor id="123" />);

      await waitFor(() => {
        expect(screen.getByText(/error/i)).toBeInTheDocument();
      });
    });
  });

  describe('when saving workflow', () => {
    it('should call API with workflow data', async () => {
      const user = userEvent.setup();
      vi.mocked(api.workflows.get).mockResolvedValue({ id: '123', name: 'Test' });
      vi.mocked(api.workflows.create).mockResolvedValue({ id: '123' });

      render(<WorkflowEditor id="123" />);

      await waitFor(() => screen.getByText('Test'));
      await user.click(screen.getByRole('button', { name: /save/i }));

      expect(api.workflows.create).toHaveBeenCalled();
    });
  });
});
```

### Output Format

```
## Generated Tests for [Component/Function]

### Test File: [path/to/test_file]

[Complete test code]

### Test Coverage
- Happy path: [x] cases
- Edge cases: [x] cases
- Error cases: [x] cases

### Mock Requirements
- [List of dependencies that need mocking]

### Notes
- [Any special considerations]
```

Generate tests now.
