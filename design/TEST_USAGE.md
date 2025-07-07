# Gode Test System Usage Guide

## Overview

Gode's test system provides a Jest-like testing API for JavaScript/TypeScript code running in the Gode runtime. It features JavaScript-based expectation matching for optimal performance and natural semantics.

## Basic Test Structure

### Test Definition

```javascript
// Simple test
test('basic arithmetic', () => {
  expect(1 + 1).toBe(2);
});

// Test with describe block
describe('Calculator', () => {
  test('addition', () => {
    expect(add(2, 3)).toBe(5);
  });
  
  test('subtraction', () => {
    expect(subtract(5, 3)).toBe(2);
  });
});
```

### Test Configuration

```javascript
// Test with timeout
test('async operation', () => {
  // test code
}, { timeout: 10000 });

// Skip test
test.skip('not ready yet', () => {
  // this won't run
});

// Run only this test
test.only('focus on this', () => {
  // only this test will run
});
```

## Available Matchers

### Equality Matchers

```javascript
// Strict equality (===)
expect(actual).toBe(expected);

// Deep equality (for objects/arrays)
expect(actual).toEqual(expected);

// Negation
expect(actual).not.toBe(unexpected);
expect(actual).not.toEqual(unexpected);
```

### Truthiness Matchers

```javascript
// Truthy/falsy checks
expect(value).toBeTruthy();
expect(value).toBeFalsy();

// Null/undefined checks
expect(value).toBeNull();
expect(value).toBeUndefined();
expect(value).toBeDefined();

// NaN check
expect(NaN).toBeNaN();
expect(42).not.toBeNaN();
```

### Numeric Matchers

```javascript
// Comparison
expect(5).toBeGreaterThan(3);
expect(3).toBeLessThan(5);
expect(5).toBeGreaterThanOrEqual(5);
expect(3).toBeLessThanOrEqual(3);

// Floating point precision
expect(0.1 + 0.2).toBeCloseTo(0.3);
expect(Math.PI).toBeCloseTo(3.14, 2); // precision: 2 decimal places
```

### String/Array Matchers

```javascript
// Containment
expect(['a', 'b', 'c']).toContain('b');
expect('hello world').toContain('world');

// Length
expect([1, 2, 3]).toHaveLength(3);
expect('hello').toHaveLength(5);

// Pattern matching
expect('hello world').toMatch(/world/);
expect('test123').toMatch(/\\d+/);
expect('testing').toMatch('test'); // substring
```

### Function/Error Matchers

```javascript
// Function throws
expect(() => {
  throw new Error('Something went wrong');
}).toThrow();

// Specific error message (partial match)
expect(() => {
  throw new Error('Something went wrong');
}).toThrow('went wrong');

// Function doesn't throw
expect(() => {
  return 'safe';
}).not.toThrow();
```

## Test Hooks

### Setup and Teardown

```javascript
describe('Database Tests', () => {
  // Run once before all tests in this describe block
  beforeAll(() => {
    setupDatabase();
  });

  // Run once after all tests in this describe block
  afterAll(() => {
    teardownDatabase();
  });

  // Run before each test
  beforeEach(() => {
    clearDatabase();
    seedTestData();
  });

  // Run after each test
  afterEach(() => {
    cleanupTestData();
  });

  test('user creation', () => {
    // test code
  });

  test('user deletion', () => {
    // test code
  });
});
```

### Hook Execution Order

For nested describe blocks:
1. Outer `beforeAll`
2. Inner `beforeAll`
3. Outer `beforeEach`
4. Inner `beforeEach`
5. **Test execution**
6. Inner `afterEach`
7. Outer `afterEach`
8. Inner `afterAll`
9. Outer `afterAll`

## Running Tests

### Command Line

```bash
# Run all tests
gode test

# Run specific test file
gode test tests/calculator.test.js

# Run all tests in directory
gode test tests/

# Run with pattern
gode test tests/**/unit/*.test.js
```

### Test File Discovery

Gode automatically discovers test files with these patterns:
- `*.test.js`
- `*.spec.js`
- Files in `tests/` directories

## Advanced Patterns

### Data-Driven Tests

```javascript
describe('Validation Tests', () => {
  const testCases = [
    { input: 'valid@email.com', expected: true },
    { input: 'invalid-email', expected: false },
    { input: '', expected: false }
  ];

  testCases.forEach(({ input, expected }) => {
    test(`email validation: ${input}`, () => {
      expect(validateEmail(input)).toBe(expected);
    });
  });
});
```

### Async Testing

```javascript
test('async operation', async () => {
  const result = await fetchData();
  expect(result).toBeDefined();
});

test('promise rejection', async () => {
  await expect(async () => {
    await failingOperation();
  }).toThrow();
});
```

### State Management Testing

```javascript
describe('Counter State', () => {
  let counter;

  beforeEach(() => {
    counter = new Counter(0);
  });

  test('increment', () => {
    counter.increment();
    expect(counter.value).toBe(1);
  });

  test('decrement', () => {
    counter.decrement();
    expect(counter.value).toBe(-1);
  });
});
```

## Best Practices

### Test Organization

1. **Group related tests** with `describe` blocks
2. **Use descriptive test names** that explain the behavior
3. **Keep tests focused** - one assertion per concept
4. **Use setup/teardown hooks** to avoid repetition

### Writing Good Assertions

```javascript
// Good: Specific and descriptive
expect(user.name).toBe('John Doe');
expect(errors).toHaveLength(0);
expect(response.status).toBe(200);

// Avoid: Vague or unclear
expect(result).toBeTruthy();
expect(data).toBeDefined();
```

### Error Testing

```javascript
// Test specific error conditions
test('validates required fields', () => {
  expect(() => {
    createUser({ name: '' });
  }).toThrow('Name is required');
});

// Test error recovery
test('handles network failures gracefully', () => {
  mockNetworkFailure();
  const result = fetchWithRetry();
  expect(result.attempts).toBeGreaterThan(1);
});
```

### Performance Testing

```javascript
test('operation completes within time limit', () => {
  const start = Date.now();
  performExpensiveOperation();
  const duration = Date.now() - start;
  
  expect(duration).toBeLessThan(1000); // Under 1 second
});
```

## Error Messages

The test system provides clear, contextual error messages:

```javascript
// Example failure messages:
// expected 2 to be 3
// expected "hello world" to contain "xyz"
// expected [1, 2, 3] to have length 5 but got 3
// expected function to throw "specific error" but got "different error"
```

## Integration with Gode Runtime

### Module Testing

```javascript
// Import Gode modules
import { stream } from 'gode:stream';

test('stream functionality', () => {
  const readable = stream.Readable.from(['a', 'b', 'c']);
  expect(readable).toBeDefined();
});
```

### Plugin Testing

```javascript
test('math plugin integration', () => {
  const result = add(5, 3); // From loaded plugin
  expect(result).toBe(8);
});
```

## Configuration

Tests can be configured via `package.json`:

```json
{
  "gode": {
    "test": {
      "timeout": 5000,
      "testMatch": ["tests/**/*.test.js"],
      "verbose": true
    }
  }
}
```

## Common Patterns

### Resource Cleanup

```javascript
describe('File Operations', () => {
  let tempFiles = [];

  afterEach(() => {
    // Clean up created files
    tempFiles.forEach(file => fs.unlinkSync(file));
    tempFiles = [];
  });

  test('creates temporary file', () => {
    const file = createTempFile();
    tempFiles.push(file);
    expect(fs.existsSync(file)).toBe(true);
  });
});
```

### Mock Data

```javascript
describe('API Client', () => {
  let mockData;

  beforeEach(() => {
    mockData = {
      users: [
        { id: 1, name: 'Alice' },
        { id: 2, name: 'Bob' }
      ]
    };
  });

  test('fetches users', () => {
    const client = new APIClient(mockData);
    const users = client.getUsers();
    expect(users).toHaveLength(2);
    expect(users[0].name).toBe('Alice');
  });
});
```

This test system provides a robust, Jest-compatible testing environment optimized for the Gode runtime with excellent performance and clear error reporting.