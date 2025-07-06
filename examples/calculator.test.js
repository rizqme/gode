// Example test file demonstrating the gode:test module

// Simple calculator to test
class Calculator {
  add(a, b) {
    return a + b;
  }
  
  subtract(a, b) {
    return a - b;
  }
  
  multiply(a, b) {
    return a * b;
  }
  
  divide(a, b) {
    if (b === 0) {
      throw new Error('Division by zero');
    }
    return a / b;
  }
  
  reset() {
    // Reset any internal state
  }
}

describe('Calculator', () => {
  let calculator;
  
  beforeEach(() => {
    calculator = new Calculator();
  });
  
  afterEach(() => {
    calculator.reset();
  });
  
  describe('addition', () => {
    test('should add two positive numbers', () => {
      const result = calculator.add(2, 3);
      expect(result).toBe(5);
    });
    
    test('should handle negative numbers', () => {
      const result = calculator.add(-5, 3);
      expect(result).toBe(-2);
    });
    
    test('should handle zero', () => {
      const result = calculator.add(0, 5);
      expect(result).toBe(5);
    });
  });
  
  describe('subtraction', () => {
    test('should subtract two numbers', () => {
      const result = calculator.subtract(10, 3);
      expect(result).toBe(7);
    });
    
    test('should handle negative results', () => {
      const result = calculator.subtract(3, 10);
      expect(result).toBe(-7);
    });
  });
  
  describe('multiplication', () => {
    test('should multiply two numbers', () => {
      const result = calculator.multiply(4, 5);
      expect(result).toBe(20);
    });
    
    test('should handle zero multiplication', () => {
      const result = calculator.multiply(5, 0);
      expect(result).toBe(0);
    });
  });
  
  describe('division', () => {
    test('should divide two numbers', () => {
      const result = calculator.divide(10, 2);
      expect(result).toBe(5);
    });
    
    test('should handle decimal results', () => {
      const result = calculator.divide(5, 2);
      expect(result).toBe(2.5);
    });
    
    test('should throw on division by zero', () => {
      expect(() => calculator.divide(10, 0)).toThrow('Division by zero');
    });
  });
  
  describe('edge cases', () => {
    test('should handle very large numbers', () => {
      const result = calculator.add(Number.MAX_SAFE_INTEGER, 1);
      expect(result).toBeGreaterThan(Number.MAX_SAFE_INTEGER);
    });
    
    test('should handle floating point precision', () => {
      const result = calculator.add(0.1, 0.2);
      // Note: This will demonstrate floating point precision issues
      expect(result).toBeGreaterThan(0.3);
    });
  });
});

// Test outside of describe block
test('standalone test', () => {
  const calc = new Calculator();
  expect(calc.add(1, 1)).toBe(2);
});

// Skipped test
test.skip('this test is skipped', () => {
  throw new Error('This should not run');
});

// Test with timeout
test('async-like test with timeout', () => {
  // Simulate work
  const start = Date.now();
  while (Date.now() - start < 10) {
    // busy wait for 10ms
  }
  expect(true).toBeTruthy();
}, { timeout: 1000 });