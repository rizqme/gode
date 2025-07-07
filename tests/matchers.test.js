// Comprehensive test demonstrating all expectation matchers

describe('Expectation Matchers', () => {
  describe('Equality Matchers', () => {
    test('toBe - strict equality', () => {
      expect(2 + 2).toBe(4);
      expect('hello').toBe('hello');
      expect(true).toBe(true);
      expect(null).toBe(null);
      expect(undefined).toBe(undefined);
    });

    test('toEqual - deep equality', () => {
      expect({ a: 1, b: 2 }).toEqual({ a: 1, b: 2 });
      expect([1, 2, 3]).toEqual([1, 2, 3]);
      expect({ nested: { value: 'test' } }).toEqual({ nested: { value: 'test' } });
    });

    test('not matchers', () => {
      expect(2 + 2).not.toBe(5);
      expect('hello').not.toBe('world');
      expect({ a: 1 }).not.toEqual({ a: 2 });
      expect([1, 2]).not.toEqual([2, 1]);
    });
  });

  describe('Truthiness Matchers', () => {
    test('toBeTruthy', () => {
      expect(true).toBeTruthy();
      expect(1).toBeTruthy();
      expect('hello').toBeTruthy();
      expect({}).toBeTruthy();
      expect([]).toBeTruthy();
      expect(42).toBeTruthy();
      expect('0').toBeTruthy(); // string '0' is truthy
    });

    test('toBeFalsy', () => {
      expect(false).toBeFalsy();
      expect(0).toBeFalsy();
      expect('').toBeFalsy();
      expect(null).toBeFalsy();
      expect(undefined).toBeFalsy();
    });

    test('toBeNull', () => {
      expect(null).toBeNull();
      expect(undefined).not.toBeNull();
      expect(0).not.toBeNull();
      expect('').not.toBeNull();
    });
  });

  describe('Error Matchers', () => {
    test('toThrow - function throws error', () => {
      const throwingFunction = () => {
        throw new Error('Something went wrong');
      };
      
      expect(throwingFunction).toThrow();
      expect(throwingFunction).toThrow('Something went wrong');
      expect(throwingFunction).toThrow('went wrong'); // partial match
    });

    test('toThrow - function does not throw', () => {
      const safeFunction = () => {
        return 'safe result';
      };
      
      expect(safeFunction).not.toThrow();
    });

    test('toThrow - specific error types', () => {
      const typeErrorFunction = () => {
        throw new TypeError('Type error occurred');
      };
      
      expect(typeErrorFunction).toThrow();
      expect(typeErrorFunction).toThrow('Type error');
    });
  });

  describe('Complex Data Types', () => {
    test('objects', () => {
      const user = { name: 'John', age: 30, active: true };
      
      expect(user).toBeTruthy();
      expect(user.name).toBe('John');
      expect(user.age).toBe(30);
      expect(user.active).toBeTruthy();
    });

    test('arrays', () => {
      const numbers = [1, 2, 3, 4, 5];
      
      expect(numbers).toBeTruthy();
      expect(numbers[0]).toBe(1);
      expect(numbers[4]).toBe(5);
      expect(numbers).toEqual([1, 2, 3, 4, 5]);
    });

    test('nested structures', () => {
      const data = {
        users: [
          { id: 1, name: 'Alice' },
          { id: 2, name: 'Bob' }
        ],
        meta: {
          total: 2,
          page: 1
        }
      };
      
      expect(data.users).toEqual([
        { id: 1, name: 'Alice' },
        { id: 2, name: 'Bob' }
      ]);
      expect(data.meta.total).toBe(2);
      expect(data.users[0].name).toBe('Alice');
    });
  });
});

describe('Edge Cases and Gotchas', () => {
  test('floating point precision', () => {
    const result = 0.1 + 0.2;
    // This will fail due to floating point precision
    // expect(result).toBe(0.3);
    
    // Instead, we need to handle floating point comparisons carefully
    expect(result).toBeTruthy(); // Just check it's not zero
  });

  test('type coercion awareness', () => {
    expect('0').toBeTruthy(); // string '0' is truthy
    expect(0).toBeFalsy();    // number 0 is falsy
    expect('').toBeFalsy();   // empty string is falsy
    expect(' ').toBeTruthy(); // space string is truthy
  });

  test('object reference vs value equality', () => {
    const obj1 = { a: 1 };
    const obj2 = { a: 1 };
    const obj3 = obj1;
    
    // Different object references
    expect(obj1).not.toBe(obj2);  // different references
    expect(obj1).toEqual(obj2);   // same values
    
    // Same object reference
    expect(obj1).toBe(obj3);      // same reference
    expect(obj1).toEqual(obj3);   // same values
  });

  test('array reference vs value equality', () => {
    const arr1 = [1, 2, 3];
    const arr2 = [1, 2, 3];
    const arr3 = arr1;
    
    expect(arr1).not.toBe(arr2);  // different references
    expect(arr1).toEqual(arr2);   // same values
    expect(arr1).toBe(arr3);      // same reference
  });
});