// Comprehensive test demonstrating error handling and edge cases

describe('Error Handling and Edge Cases', () => {
  describe('Exception Testing', () => {
    test('basic error throwing', () => {
      const throwError = () => {
        throw new Error('Basic error');
      };
      
      expect(throwError).toThrow();
      expect(throwError).toThrow('Basic error');
      expect(throwError).toThrow('Basic');
    });

    test('different error types', () => {
      const throwTypeError = () => {
        throw new TypeError('Type error message');
      };
      
      const throwRangeError = () => {
        throw new RangeError('Range error message');
      };
      
      const throwCustomError = () => {
        const error = new Error('Custom error');
        error.code = 'CUSTOM_ERROR';
        throw error;
      };
      
      expect(throwTypeError).toThrow('Type error');
      expect(throwRangeError).toThrow('Range error');
      expect(throwCustomError).toThrow('Custom error');
    });

    test('conditional error throwing', () => {
      const conditionalError = (shouldThrow) => {
        if (shouldThrow) {
          throw new Error('Conditional error');
        }
        return 'success';
      };
      
      expect(() => conditionalError(true)).toThrow('Conditional error');
      expect(() => conditionalError(false)).not.toThrow();
      expect(conditionalError(false)).toBe('success');
    });

    test('error in nested function calls', () => {
      const level3 = () => {
        throw new Error('Error from level 3');
      };
      
      const level2 = () => {
        return level3();
      };
      
      const level1 = () => {
        return level2();
      };
      
      expect(level1).toThrow('Error from level 3');
    });
  });

  describe('Boundary Value Testing', () => {
    const divide = (a, b) => {
      if (b === 0) {
        throw new Error('Division by zero');
      }
      if (typeof a !== 'number' || typeof b !== 'number') {
        throw new TypeError('Arguments must be numbers');
      }
      return a / b;
    };

    test('normal division', () => {
      expect(divide(10, 2)).toBe(5);
      expect(divide(7, 2)).toBe(3.5);
      expect(divide(-10, 2)).toBe(-5);
    });

    test('division by zero', () => {
      expect(() => divide(10, 0)).toThrow('Division by zero');
      expect(() => divide(-5, 0)).toThrow('Division by zero');
    });

    test('invalid input types', () => {
      expect(() => divide('10', 2)).toThrow('Arguments must be numbers');
      expect(() => divide(10, '2')).toThrow('Arguments must be numbers');
      expect(() => divide('10', '2')).toThrow('Arguments must be numbers');
      expect(() => divide(null, 2)).toThrow('Arguments must be numbers');
      expect(() => divide(10, undefined)).toThrow('Arguments must be numbers');
    });

    test('extreme values', () => {
      expect(divide(Number.MAX_VALUE, 2)).toBeGreaterThan(0);
      expect(typeof divide(Number.MIN_VALUE, 2)).toBe('number'); // May underflow to 0
      expect(typeof divide(1, Number.MAX_VALUE)).toBe('number'); // Very small number, may be 0
    });
  });

  describe('Null and Undefined Handling', () => {
    const processValue = (value) => {
      if (value === null) {
        throw new Error('Null value not allowed');
      }
      if (value === undefined) {
        throw new Error('Undefined value not allowed');
      }
      return value.toString();
    };

    test('valid values', () => {
      expect(processValue('hello')).toBe('hello');
      expect(processValue(42)).toBe('42');
      expect(processValue(true)).toBe('true');
      expect(processValue(false)).toBe('false');
      expect(processValue(0)).toBe('0');
      expect(processValue('')).toBe('');
    });

    test('null handling', () => {
      expect(() => processValue(null)).toThrow('Null value not allowed');
    });

    test('undefined handling', () => {
      expect(() => processValue(undefined)).toThrow('Undefined value not allowed');
    });

    test('falsy but valid values', () => {
      expect(processValue(0)).toBe('0');
      expect(processValue(false)).toBe('false');
      expect(processValue('')).toBe('');
    });
  });

  describe('Array and Object Edge Cases', () => {
    test('empty arrays', () => {
      const arr = [];
      expect(arr).toEqual([]);
      expect(arr).toHaveLength(0);
      expect(arr[0]).toBe(undefined);
    });

    test('sparse arrays', () => {
      const sparse = new Array(5);
      sparse[2] = 'middle';
      
      expect(sparse).toHaveLength(5);
      expect(sparse[0]).toBe(undefined);
      expect(sparse[2]).toBe('middle');
      expect(sparse[4]).toBe(undefined);
    });

    test('empty objects', () => {
      const obj = {};
      expect(obj).toEqual({});
      expect(Object.keys(obj)).toHaveLength(0);
      expect(obj.nonExistent).toBe(undefined);
    });

    test('nested null/undefined in objects', () => {
      const obj = {
        a: null,
        b: undefined,
        c: {
          d: null,
          e: undefined
        }
      };
      
      expect(obj.a).toBeNull();
      expect(obj.b).toBe(undefined);
      expect(obj.c.d).toBeNull();
      expect(obj.c.e).toBe(undefined);
    });
  });

  describe('String Edge Cases', () => {
    test('empty strings', () => {
      const empty = '';
      expect(empty).toBe('');
      expect(empty).toBeFalsy();
      expect(empty).toHaveLength(0);
    });

    test('whitespace strings', () => {
      const space = ' ';
      const tab = '\t';
      const newline = '\n';
      const mixed = ' \t\n ';
      
      expect(space).toBeTruthy();
      expect(tab).toBeTruthy();
      expect(newline).toBeTruthy();
      expect(mixed).toBeTruthy();
      
      expect(space).toHaveLength(1);
      expect(mixed).toHaveLength(4);
    });

    test('special characters', () => {
      const unicode = '🚀';
      const escaped = 'Hello\nWorld';
      const quotes = "It's a \"test\"";
      
      expect(unicode).toBeTruthy();
      expect(escaped).toContain('\n');
      expect(quotes).toContain('"');
      expect(quotes).toContain("'");
    });
  });

  describe('Numeric Edge Cases', () => {
    test('special numeric values', () => {
      expect(Infinity).toBeTruthy();
      expect(-Infinity).toBeTruthy();
      expect(NaN).toBeFalsy(); // NaN is falsy and NaN !== NaN
      expect(Number.MAX_VALUE).toBeTruthy();
      expect(Number.MIN_VALUE).toBeTruthy();
    });

    test('NaN behavior', () => {
      const notANumber = NaN;
      expect(isNaN(notANumber)).toBeTruthy(); // Use isNaN to check for NaN
      expect(notANumber === notANumber).toBeFalsy(); // NaN !== NaN using direct comparison
    });

    test('floating point precision', () => {
      const result1 = 0.1 + 0.2;
      const result2 = 0.3;
      
      // This demonstrates floating point precision issues
      expect(result1).not.toBe(result2); // 0.30000000000000004 !== 0.3
    });

    test('integer overflow behavior', () => {
      const maxInt = Number.MAX_SAFE_INTEGER;
      const overflowResult = maxInt + 1;
      
      expect(overflowResult).toBeTruthy();
      expect(overflowResult).not.toBe(maxInt);
    });
  });

  describe('Error Recovery Patterns', () => {
    const safeOperation = (operation, fallback) => {
      try {
        return operation();
      } catch (error) {
        return fallback;
      }
    };

    test('successful operation', () => {
      const result = safeOperation(() => 2 + 2, 'fallback');
      expect(result).toBe(4);
    });

    test('operation with fallback', () => {
      const result = safeOperation(() => {
        throw new Error('Operation failed');
      }, 'fallback');
      expect(result).toBe('fallback');
    });

    test('complex error recovery', () => {
      const processArray = (arr) => {
        if (!Array.isArray(arr)) {
          throw new TypeError('Input must be an array');
        }
        
        return arr.map(item => {
          if (typeof item !== 'number') {
            throw new TypeError('Array items must be numbers');
          }
          return item * 2;
        });
      };

      const safeProcessArray = (arr) => {
        return safeOperation(() => processArray(arr), []);
      };

      expect(safeProcessArray([1, 2, 3])).toEqual([2, 4, 6]);
      expect(safeProcessArray('not array')).toEqual([]);
      expect(safeProcessArray([1, 'not number', 3])).toEqual([]);
    });
  });

  describe('State Corruption Prevention', () => {
    let sharedState;

    beforeEach(() => {
      sharedState = {
        counter: 0,
        items: [],
        flags: { initialized: true }
      };
    });

    test('state isolation test 1', () => {
      sharedState.counter = 5;
      sharedState.items.push('item1');
      sharedState.flags.initialized = false;
      
      expect(sharedState.counter).toBe(5);
      expect(sharedState.items).toEqual(['item1']);
      expect(sharedState.flags.initialized).toBeFalsy();
    });

    test('state isolation test 2', () => {
      // State should be fresh due to beforeEach
      expect(sharedState.counter).toBe(0);
      expect(sharedState.items).toEqual([]);
      expect(sharedState.flags.initialized).toBeTruthy();
    });

    test('error in state modification', () => {
      const corruptState = () => {
        sharedState.counter = 'invalid';
        throw new Error('State corruption');
      };
      
      expect(corruptState).toThrow('State corruption');
      // State should still be corrupted within this test
      expect(sharedState.counter).toBe('invalid');
    });

    test('state recovery after error', () => {
      // Fresh state again due to beforeEach
      expect(sharedState.counter).toBe(0);
    });
  });
});