# Test Demo

This example demonstrates Gode's built-in Jest-like testing framework.

## Examples

### simple-tests.js
Comprehensive test suite showing all available testing features.
```bash
./gode test examples/test-demo/simple-tests.js
```

## Features Demonstrated

- **Test Organization**: `describe()` blocks for grouping tests
- **Test Cases**: `test()` functions for individual tests
- **Assertions**: Comprehensive `expect()` API with 15+ matchers
- **Error Testing**: Testing functions that throw errors
- **Type Checking**: Testing with `typeof` and type matchers

## Available Matchers

- `toBe(expected)` - Strict equality (===)
- `toEqual(expected)` - Deep equality
- `toContain(item)` - Array/string contains
- `toHaveLength(length)` - Array/string length
- `toBeTruthy()` / `toBeFalsy()` - Truthiness
- `toBeNull()` / `toBeUndefined()` / `toBeDefined()` - Null checks
- `toBeNaN()` - NaN check
- `toBeGreaterThan()` / `toBeLessThan()` - Numeric comparisons
- `toBeCloseTo(number, precision)` - Floating point comparison
- `toMatch(pattern)` - Regex matching
- `toThrow(expected?)` - Function throws error
- `.not` - Negates any matcher

## Running Tests

From the project root:
```bash
./gode test examples/test-demo/simple-tests.js
```

## Test Structure

```javascript
describe('Test Group', () => {
    test('individual test', () => {
        expect(2 + 2).toBe(4);
    });
});
```

## Notes

- Tests run in JavaScript with Go backend for performance
- Error messages are clear and helpful
- Supports nested describe blocks
- Compatible with Jest testing patterns