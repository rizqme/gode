// Comprehensive test demonstrating nested describe blocks and organization

describe('Top Level Suite', () => {
  let topLevelData;

  beforeAll(() => {
    topLevelData = { initialized: true, level: 'top' };
  });

  test('top level test', () => {
    expect(topLevelData.initialized).toBeTruthy();
    expect(topLevelData.level).toBe('top');
  });

  describe('Level 1 Nested Suite', () => {
    let level1Data;

    beforeAll(() => {
      level1Data = { ...topLevelData, level: 'level1', count: 0 };
    });

    beforeEach(() => {
      level1Data.count++;
    });

    test('level 1 test A', () => {
      expect(level1Data.level).toBe('level1');
      expect(level1Data.count).toBe(1);
      expect(level1Data.initialized).toBeTruthy(); // inherited from parent
    });

    test('level 1 test B', () => {
      expect(level1Data.count).toBe(2); // incremented by beforeEach
    });

    describe('Level 2 Nested Suite A', () => {
      let level2AData;

      beforeAll(() => {
        level2AData = { ...level1Data, level: 'level2A', subCount: 0 };
      });

      beforeEach(() => {
        level2AData.subCount++;
      });

      test('level 2A test 1', () => {
        expect(level2AData.level).toBe('level2A');
        expect(level2AData.subCount).toBe(1);
        expect(level2AData.initialized).toBeTruthy(); // inherited from top level
      });

      test('level 2A test 2', () => {
        expect(level2AData.subCount).toBe(2);
      });

      describe('Level 3 Deep Nested Suite', () => {
        let level3Data;

        beforeAll(() => {
          level3Data = { 
            ...level2AData, 
            level: 'level3', 
            deepValue: 'deep',
            operations: []
          };
        });

        beforeEach(() => {
          level3Data.operations.push('beforeEach');
        });

        afterEach(() => {
          level3Data.operations.push('afterEach');
        });

        test('deeply nested test 1', () => {
          expect(level3Data.level).toBe('level3');
          expect(level3Data.deepValue).toBe('deep');
          expect(level3Data.operations).toEqual(['beforeEach']);
          expect(level3Data.initialized).toBeTruthy(); // still inherited
        });

        test('deeply nested test 2', () => {
          // operations from previous test should include afterEach + new beforeEach
          expect(level3Data.operations).toEqual(['beforeEach', 'afterEach', 'beforeEach']);
        });
      });
    });

    describe('Level 2 Nested Suite B', () => {
      let level2BData;

      beforeAll(() => {
        level2BData = { 
          ...level1Data, 
          level: 'level2B', 
          altValue: 'alternative',
          items: []
        };
      });

      beforeEach(() => {
        level2BData.items.push(`item-${level2BData.items.length + 1}`);
      });

      test('level 2B test 1', () => {
        expect(level2BData.level).toBe('level2B');
        expect(level2BData.altValue).toBe('alternative');
        expect(level2BData.items).toEqual(['item-1']);
      });

      test('level 2B test 2', () => {
        expect(level2BData.items).toEqual(['item-1', 'item-2']);
      });

      test('level 2B test 3', () => {
        expect(level2BData.items).toEqual(['item-1', 'item-2', 'item-3']);
      });
    });
  });

  describe('Level 1 Parallel Suite', () => {
    let parallelData;

    beforeAll(() => {
      parallelData = { 
        ...topLevelData, 
        level: 'parallel',
        counter: 100 
      };
    });

    test('parallel suite test 1', () => {
      expect(parallelData.level).toBe('parallel');
      expect(parallelData.counter).toBe(100);
      expect(parallelData.initialized).toBeTruthy();
    });

    test('parallel suite test 2', () => {
      // This suite runs independently of the other level 1 suite
      expect(parallelData.counter).toBe(100); // unchanged
    });

    describe('Parallel Nested Suite', () => {
      let nestedParallelData;

      beforeEach(() => {
        nestedParallelData = { 
          ...parallelData, 
          timestamp: Date.now(),
          random: Math.floor(Math.random() * 100)
        };
      });

      test('parallel nested test 1', () => {
        expect(nestedParallelData.level).toBe('parallel');
        expect(nestedParallelData.timestamp).toBeTruthy();
        expect(nestedParallelData.random).toBeTruthy();
      });

      test('parallel nested test 2', () => {
        // Each test gets fresh data from beforeEach
        expect(nestedParallelData.timestamp).toBeTruthy();
        expect(nestedParallelData.random).toBeTruthy();
      });
    });
  });
});

// Separate top-level suite to demonstrate isolation
describe('Isolated Suite', () => {
  let isolatedData;

  beforeAll(() => {
    isolatedData = { 
      isolated: true, 
      value: 'separate',
      suite: 'isolated'
    };
  });

  test('isolated test', () => {
    expect(isolatedData.isolated).toBeTruthy();
    expect(isolatedData.value).toBe('separate');
    expect(isolatedData.suite).toBe('isolated');
  });

  describe('Nested in Isolated', () => {
    test('nested isolated test', () => {
      expect(isolatedData.isolated).toBeTruthy();
      // This suite is completely separate from the first top-level suite
    });
  });
});

// Test suite demonstrating different testing patterns
describe('Testing Patterns Demo', () => {
  describe('Setup and Teardown Patterns', () => {
    let resource;

    beforeEach(() => {
      // Simulate resource creation
      resource = {
        id: Math.floor(Math.random() * 1000),
        status: 'active',
        data: []
      };
    });

    afterEach(() => {
      // Simulate resource cleanup
      resource = null;
    });

    test('resource creation test', () => {
      expect(resource).toBeTruthy();
      expect(resource.status).toBe('active');
      expect(resource.id).toBeTruthy();
      expect(resource.data).toEqual([]);
    });

    test('resource modification test', () => {
      resource.data.push('test-data');
      resource.status = 'modified';
      
      expect(resource.data).toEqual(['test-data']);
      expect(resource.status).toBe('modified');
    });

    test('resource independence test', () => {
      // This test should get a fresh resource due to beforeEach/afterEach
      expect(resource.status).toBe('active');
      expect(resource.data).toEqual([]);
    });
  });

  describe('Data Transformation Patterns', () => {
    const inputData = [
      { id: 1, name: 'Alice', age: 25 },
      { id: 2, name: 'Bob', age: 30 },
      { id: 3, name: 'Charlie', age: 35 }
    ];

    test('filtering patterns', () => {
      const adults = inputData.filter(person => person.age >= 30);
      expect(adults).toHaveLength(2);
      expect(adults[0].name).toBe('Bob');
      expect(adults[1].name).toBe('Charlie');
    });

    test('mapping patterns', () => {
      const names = inputData.map(person => person.name);
      expect(names).toEqual(['Alice', 'Bob', 'Charlie']);
    });

    test('reduction patterns', () => {
      const totalAge = inputData.reduce((sum, person) => sum + person.age, 0);
      expect(totalAge).toBe(90);
    });
  });
});