// Comprehensive test demonstrating all test hooks and their execution order

describe('Test Hooks Comprehensive Demo', () => {
  const executionLog = [];

  // Global hooks for this suite
  beforeAll(() => {
    executionLog.push('Main Suite: beforeAll');
  });

  afterAll(() => {
    executionLog.push('Main Suite: afterAll');
  });

  beforeEach(() => {
    executionLog.push('Main Suite: beforeEach');
  });

  afterEach(() => {
    executionLog.push('Main Suite: afterEach');
  });

  test('first test in main suite', () => {
    executionLog.push('Main Suite: Test 1');
    expect(executionLog).toContain('Main Suite: beforeAll');
    expect(executionLog).toContain('Main Suite: beforeEach');
  });

  test('second test in main suite', () => {
    executionLog.push('Main Suite: Test 2');
    // Should see afterEach from previous test and new beforeEach
    expect(executionLog).toContain('Main Suite: afterEach');
  });

  describe('Nested Suite A', () => {
    beforeAll(() => {
      executionLog.push('Nested A: beforeAll');
    });

    afterAll(() => {
      executionLog.push('Nested A: afterAll');
    });

    beforeEach(() => {
      executionLog.push('Nested A: beforeEach');
    });

    afterEach(() => {
      executionLog.push('Nested A: afterEach');
    });

    test('test in nested suite A', () => {
      executionLog.push('Nested A: Test 1');
      // Should have both parent and child hooks
      expect(executionLog).toContain('Main Suite: beforeAll');
      expect(executionLog).toContain('Nested A: beforeAll');
      expect(executionLog).toContain('Main Suite: beforeEach');
      expect(executionLog).toContain('Nested A: beforeEach');
    });

    describe('Deeply Nested Suite', () => {
      beforeAll(() => {
        executionLog.push('Deep Nested: beforeAll');
      });

      afterAll(() => {
        executionLog.push('Deep Nested: afterAll');
      });

      beforeEach(() => {
        executionLog.push('Deep Nested: beforeEach');
      });

      afterEach(() => {
        executionLog.push('Deep Nested: afterEach');
      });

      test('deeply nested test', () => {
        executionLog.push('Deep Nested: Test 1');
        // Should have all three levels of hooks
        expect(executionLog).toContain('Main Suite: beforeAll');
        expect(executionLog).toContain('Nested A: beforeAll');
        expect(executionLog).toContain('Deep Nested: beforeAll');
      });
    });
  });

  describe('Nested Suite B', () => {
    beforeAll(() => {
      executionLog.push('Nested B: beforeAll');
    });

    afterAll(() => {
      executionLog.push('Nested B: afterAll');
    });

    beforeEach(() => {
      executionLog.push('Nested B: beforeEach');
    });

    afterEach(() => {
      executionLog.push('Nested B: afterEach');
    });

    test('test in nested suite B', () => {
      executionLog.push('Nested B: Test 1');
      // Should have parent hooks and own hooks
      expect(executionLog).toContain('Main Suite: beforeAll');
      expect(executionLog).toContain('Nested B: beforeAll');
      // Should NOT have Nested A hooks
      expect(executionLog).toContain('Nested A: afterAll'); // A finished before B started
    });
  });
});

describe('Hook Error Handling', () => {
  let setupState;

  describe('BeforeAll Error Handling', () => {
    beforeAll(() => {
      setupState = { initialized: true };
      // This will succeed
    });

    test('test after successful beforeAll', () => {
      expect(setupState.initialized).toBeTruthy();
    });
  });

  describe('BeforeEach Error Recovery Demo', () => {
    let testCounter = 0;
    let errors = [];

    beforeEach(() => {
      testCounter++;
      // Don't throw here - instead simulate hook error handling
    });

    test('simulate beforeEach error handling', () => {
      // Simulate what happens when a beforeEach hook fails
      const simulateBeforeEachError = () => {
        throw new Error('BeforeEach failed for test 2');
      };

      // Test how we would handle a beforeEach error
      try {
        simulateBeforeEachError();
        // This shouldn't execute
        expect(false).toBeTruthy();
      } catch (error) {
        // Catch and verify the beforeEach-style error
        expect(error.message).toBe('BeforeEach failed for test 2');
        errors.push(error);
      }

      expect(errors).toHaveLength(1);
      expect(testCounter).toBe(1);
    });

    test('recovery after simulated error', () => {
      expect(testCounter).toBe(2);
      expect(errors).toHaveLength(1); // Error from previous test
    });

    test('demonstrate hook error patterns', () => {
      // Show different hook error handling patterns
      const hookErrorHandler = (hookName, error) => {
        return {
          hookName,
          error: error.message,
          timestamp: Date.now(),
          recovered: true
        };
      };

      try {
        throw new Error('Hook execution failed');
      } catch (error) {
        const result = hookErrorHandler('beforeEach', error);
        expect(result.hookName).toBe('beforeEach');
        expect(result.error).toBe('Hook execution failed');
        expect(result.recovered).toBeTruthy();
      }
    });
  });

  describe('AfterEach Error Handling', () => {
    let cleanupCounter = 0;

    afterEach(() => {
      cleanupCounter++;
      // Don't throw error in first test, just track execution
    });

    test('test with normal afterEach', () => {
      expect(true).toBeTruthy();
      expect(cleanupCounter).toBe(0); // Before afterEach runs
    });

    test('test after normal afterEach', () => {
      expect(cleanupCounter).toBe(1); // afterEach ran once so far (will run again after this test)
    });
  });
});

describe('Hook Setup Patterns', () => {
  describe('Database-like Setup Pattern', () => {
    let database;

    beforeAll(() => {
      // Simulate database connection
      database = {
        connected: true,
        collections: {},
        operations: []
      };
    });

    afterAll(() => {
      // Simulate database disconnection
      database.connected = false;
      database.collections = null;
    });

    beforeEach(() => {
      // Clear data before each test
      database.collections = {};
      database.operations = [];
    });

    afterEach(() => {
      // Log operations after each test
      database.operations.push('test-completed');
    });

    test('database insert operation', () => {
      expect(database.connected).toBeTruthy();
      
      database.collections.users = [
        { id: 1, name: 'John' },
        { id: 2, name: 'Jane' }
      ];
      database.operations.push('insert');
      
      expect(database.collections.users).toHaveLength(2);
      expect(database.operations).toEqual(['insert']);
    });

    test('database query operation', () => {
      expect(database.connected).toBeTruthy();
      expect(database.collections).toEqual({}); // fresh state from beforeEach
      
      database.collections.products = [
        { id: 1, name: 'Product A' }
      ];
      database.operations.push('query');
      
      expect(database.collections.products).toHaveLength(1);
    });

    test('database isolation verification', () => {
      // Should not see data from previous tests
      expect(database.collections).toEqual({});
      expect(database.operations).toEqual([]);
    });
  });

  describe('Resource Management Pattern', () => {
    let resources;

    beforeAll(() => {
      resources = {
        pool: [],
        allocated: [],
        maxSize: 5
      };
      
      // Initialize resource pool
      for (let i = 0; i < resources.maxSize; i++) {
        resources.pool.push({ id: i, available: true });
      }
    });

    beforeEach(() => {
      // Reset allocations before each test
      resources.allocated = [];
      resources.pool.forEach(resource => {
        resource.available = true;
      });
    });

    afterEach(() => {
      // Cleanup allocations after each test
      resources.allocated.forEach(resource => {
        resource.available = true;
      });
      resources.allocated = [];
    });

    test('resource allocation', () => {
      const allocateResource = () => {
        const available = resources.pool.find(r => r.available);
        if (available) {
          available.available = false;
          resources.allocated.push(available);
          return available;
        }
        return null;
      };

      const resource1 = allocateResource();
      const resource2 = allocateResource();
      
      expect(resource1).toBeTruthy();
      expect(resource2).toBeTruthy();
      expect(resource1.id).not.toBe(resource2.id);
      expect(resources.allocated).toHaveLength(2);
    });

    test('resource pool exhaustion', () => {
      // Allocate all resources
      const allocated = [];
      for (let i = 0; i < resources.maxSize; i++) {
        const resource = resources.pool[i];
        resource.available = false;
        allocated.push(resource);
      }
      resources.allocated = allocated;

      const noMoreResources = resources.pool.find(r => r.available);
      expect(noMoreResources).toBeFalsy();
      expect(resources.allocated).toHaveLength(resources.maxSize);
    });

    test('resource cleanup verification', () => {
      // All resources should be available again due to afterEach
      const availableCount = resources.pool.filter(r => r.available).length;
      expect(availableCount).toBe(resources.maxSize);
      expect(resources.allocated).toHaveLength(0);
    });
  });

  describe('Stateful Counter Pattern', () => {
    let counter;

    beforeAll(() => {
      counter = { value: 0, history: [] };
    });

    beforeEach(() => {
      counter.history.push(`beforeEach: ${counter.value}`);
    });

    afterEach(() => {
      counter.history.push(`afterEach: ${counter.value}`);
    });

    test('increment counter', () => {
      counter.value++;
      expect(counter.value).toBe(1);
      expect(counter.history).toContain('beforeEach: 0');
    });

    test('increment counter again', () => {
      counter.value++;
      expect(counter.value).toBe(2); // Continues from previous test
      expect(counter.history).toContain('afterEach: 1'); // From previous test
      expect(counter.history).toContain('beforeEach: 1'); // Current test
    });

    test('verify counter history', () => {
      expect(counter.history.length).toBeGreaterThan(0);
      expect(counter.value).toBe(2); // State persists across tests in same suite
    });
  });
});