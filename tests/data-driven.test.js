// Comprehensive test demonstrating data-driven testing patterns

describe('Data-Driven Testing Patterns', () => {
  describe('Parameterized Tests', () => {
    // Test data sets
    const mathTestCases = [
      { a: 2, b: 3, expected: 5, operation: 'addition' },
      { a: 10, b: 4, expected: 6, operation: 'subtraction' },
      { a: 6, b: 7, expected: 42, operation: 'multiplication' },
      { a: 15, b: 3, expected: 5, operation: 'division' }
    ];

    const calculator = {
      add: (a, b) => a + b,
      subtract: (a, b) => a - b,
      multiply: (a, b) => a * b,
      divide: (a, b) => a / b
    };

    // Generate tests for each data case
    mathTestCases.forEach(({ a, b, expected, operation }) => {
      test(`${operation}: ${a} and ${b} should equal ${expected}`, () => {
        let result;
        switch (operation) {
          case 'addition':
            result = calculator.add(a, b);
            break;
          case 'subtraction':
            result = calculator.subtract(a, b);
            break;
          case 'multiplication':
            result = calculator.multiply(a, b);
            break;
          case 'division':
            result = calculator.divide(a, b);
            break;
        }
        expect(result).toBe(expected);
      });
    });
  });

  describe('Validation Test Cases', () => {
    const validationTestCases = [
      // Email validation cases
      { input: 'test@example.com', type: 'email', expected: true, description: 'valid email' },
      { input: 'invalid-email', type: 'email', expected: false, description: 'invalid email format' },
      { input: 'user@domain', type: 'email', expected: false, description: 'email without TLD' },
      { input: '', type: 'email', expected: false, description: 'empty email' },
      
      // Phone number validation cases
      { input: '+1-555-123-4567', type: 'phone', expected: true, description: 'valid US phone' },
      { input: '555-123-4567', type: 'phone', expected: true, description: 'valid phone without country code' },
      { input: '123', type: 'phone', expected: false, description: 'too short phone' },
      { input: 'not-a-phone', type: 'phone', expected: false, description: 'invalid phone format' },
      
      // Age validation cases
      { input: 25, type: 'age', expected: true, description: 'valid adult age' },
      { input: 18, type: 'age', expected: true, description: 'minimum adult age' },
      { input: 17, type: 'age', expected: false, description: 'underage' },
      { input: -5, type: 'age', expected: false, description: 'negative age' },
      { input: 150, type: 'age', expected: false, description: 'unrealistic age' }
    ];

    const validators = {
      email: (input) => {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(input);
      },
      phone: (input) => {
        const phoneRegex = /^(\+\d{1,3}-?)?\d{3}-?\d{3}-?\d{4}$/;
        return phoneRegex.test(input);
      },
      age: (input) => {
        return typeof input === 'number' && input >= 18 && input <= 120;
      }
    };

    validationTestCases.forEach(({ input, type, expected, description }) => {
      test(`${type} validation: ${description}`, () => {
        const result = validators[type](input);
        expect(result).toBe(expected);
      });
    });
  });

  describe('String Transformation Tests', () => {
    const transformationCases = [
      { input: 'hello world', transform: 'uppercase', expected: 'HELLO WORLD' },
      { input: 'HELLO WORLD', transform: 'lowercase', expected: 'hello world' },
      { input: 'hello world', transform: 'capitalize', expected: 'Hello World' },
      { input: 'hello-world', transform: 'camelCase', expected: 'helloWorld' },
      { input: 'HelloWorld', transform: 'kebabCase', expected: 'hello-world' },
      { input: '  hello world  ', transform: 'trim', expected: 'hello world' },
      { input: 'hello world', transform: 'reverse', expected: 'dlrow olleh' },
      { input: 'hello', transform: 'repeat3', expected: 'hellohellohello' }
    ];

    const transformers = {
      uppercase: (str) => str.toUpperCase(),
      lowercase: (str) => str.toLowerCase(),
      capitalize: (str) => str.replace(/\b\w/g, l => l.toUpperCase()),
      camelCase: (str) => str.replace(/-([a-z])/g, (match, letter) => letter.toUpperCase()),
      kebabCase: (str) => str.replace(/([A-Z])/g, '-$1').toLowerCase().replace(/^-/, ''),
      trim: (str) => str.trim(),
      reverse: (str) => str.split('').reverse().join(''),
      repeat3: (str) => str.repeat(3)
    };

    transformationCases.forEach(({ input, transform, expected }) => {
      test(`${transform}: "${input}" should become "${expected}"`, () => {
        const result = transformers[transform](input);
        expect(result).toBe(expected);
      });
    });
  });

  describe('Array Operation Tests', () => {
    const arrayTestCases = [
      {
        name: 'sum of positive numbers',
        input: [1, 2, 3, 4, 5],
        operation: 'sum',
        expected: 15
      },
      {
        name: 'sum with negative numbers',
        input: [-2, -1, 0, 1, 2],
        operation: 'sum',
        expected: 0
      },
      {
        name: 'average of numbers',
        input: [2, 4, 6, 8],
        operation: 'average',
        expected: 5
      },
      {
        name: 'maximum value',
        input: [3, 1, 4, 1, 5, 9, 2, 6],
        operation: 'max',
        expected: 9
      },
      {
        name: 'minimum value',
        input: [3, 1, 4, 1, 5, 9, 2, 6],
        operation: 'min',
        expected: 1
      },
      {
        name: 'filter even numbers',
        input: [1, 2, 3, 4, 5, 6],
        operation: 'filterEven',
        expected: [2, 4, 6]
      },
      {
        name: 'double all values',
        input: [1, 2, 3],
        operation: 'double',
        expected: [2, 4, 6]
      }
    ];

    const arrayOperations = {
      sum: (arr) => arr.reduce((sum, val) => sum + val, 0),
      average: (arr) => arr.reduce((sum, val) => sum + val, 0) / arr.length,
      max: (arr) => Math.max(...arr),
      min: (arr) => Math.min(...arr),
      filterEven: (arr) => arr.filter(n => n % 2 === 0),
      double: (arr) => arr.map(n => n * 2)
    };

    arrayTestCases.forEach(({ name, input, operation, expected }) => {
      test(name, () => {
        const result = arrayOperations[operation](input);
        expect(result).toEqual(expected);
      });
    });
  });

  describe('Edge Case Matrix Testing', () => {
    const edgeCaseInputs = [
      null,
      undefined,
      '',
      0,
      -0,
      1,
      -1,
      Infinity,
      -Infinity,
      NaN,
      true,
      false,
      [],
      {},
      'string'
    ];

    const typeChecker = (input) => {
      if (input === null) return 'null';
      if (input === undefined) return 'undefined';
      if (typeof input === 'number') {
        if (isNaN(input)) return 'NaN';
        if (input === Infinity) return 'Infinity';
        if (input === -Infinity) return '-Infinity';
        return 'number';
      }
      if (typeof input === 'string') return 'string';
      if (typeof input === 'boolean') return 'boolean';
      if (Array.isArray(input)) return 'array';
      if (typeof input === 'object') return 'object';
      return 'unknown';
    };

    edgeCaseInputs.forEach((input, index) => {
      test(`edge case ${index}: ${JSON.stringify(input)} type detection`, () => {
        const result = typeChecker(input);
        expect(result).toBeTruthy(); // Should return some type
        expect(typeof result).toBe('string');
      });
    });
  });

  describe('Combinatorial Testing', () => {
    // Test combinations of different user roles and permissions
    const userRoles = ['admin', 'user', 'guest'];
    const permissions = ['read', 'write', 'delete'];
    const resources = ['user-data', 'system-config', 'logs'];

    const hasPermission = (role, permission, resource) => {
      const rules = {
        admin: { 'user-data': ['read', 'write', 'delete'], 'system-config': ['read', 'write'], 'logs': ['read'] },
        user: { 'user-data': ['read', 'write'], 'system-config': [], 'logs': [] },
        guest: { 'user-data': ['read'], 'system-config': [], 'logs': [] }
      };
      
      return rules[role] && rules[role][resource] && rules[role][resource].includes(permission);
    };

    // Generate tests for all combinations
    userRoles.forEach(role => {
      permissions.forEach(permission => {
        resources.forEach(resource => {
          test(`${role} ${permission} access to ${resource}`, () => {
            const result = hasPermission(role, permission, resource);
            expect(typeof result).toBe('boolean');
            
            // Specific assertions based on known rules
            if (role === 'admin' && resource === 'user-data') {
              expect(result).toBeTruthy(); // Admin can do anything with user-data
            } else if (role === 'guest' && permission === 'delete') {
              expect(result).toBeFalsy(); // Guest can never delete
            } else if (resource === 'system-config' && role !== 'admin') {
              expect(result).toBeFalsy(); // Only admin can access system-config
            }
          });
        });
      });
    });
  });

  describe('Configuration Testing', () => {
    const configurationTests = [
      {
        name: 'development environment',
        config: { env: 'development', debug: true, cache: false, timeout: 1000 },
        expectations: {
          shouldLog: true,
          shouldCache: false,
          timeoutValue: 1000
        }
      },
      {
        name: 'production environment',
        config: { env: 'production', debug: false, cache: true, timeout: 5000 },
        expectations: {
          shouldLog: false,
          shouldCache: true,
          timeoutValue: 5000
        }
      },
      {
        name: 'testing environment',
        config: { env: 'test', debug: true, cache: false, timeout: 500 },
        expectations: {
          shouldLog: true,
          shouldCache: false,
          timeoutValue: 500
        }
      }
    ];

    const processConfig = (config) => {
      return {
        shouldLog: config.debug === true,
        shouldCache: config.cache === true,
        timeoutValue: config.timeout || 3000
      };
    };

    configurationTests.forEach(({ name, config, expectations }) => {
      test(`configuration for ${name}`, () => {
        const result = processConfig(config);
        
        expect(result.shouldLog).toBe(expectations.shouldLog);
        expect(result.shouldCache).toBe(expectations.shouldCache);
        expect(result.timeoutValue).toBe(expectations.timeoutValue);
      });
    });
  });
});