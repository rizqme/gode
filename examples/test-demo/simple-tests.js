// Simple test demonstration using Gode's built-in test framework
console.log("=== Test Framework Demo ===");

// This file demonstrates the Jest-like testing API available in Gode

describe('Basic Math Operations', () => {
    test('addition works correctly', () => {
        expect(2 + 2).toBe(4);
        expect(5 + 3).toBe(8);
        expect(-1 + 1).toBe(0);
    });
    
    test('multiplication works correctly', () => {
        expect(3 * 4).toBe(12);
        expect(0 * 100).toBe(0);
        expect(-2 * 3).toBe(-6);
    });
    
    test('division works correctly', () => {
        expect(12 / 3).toBe(4);
        expect(100 / 10).toBe(10);
        expect(-8 / 2).toBe(-4);
    });
});

describe('String Operations', () => {
    test('string concatenation', () => {
        expect('Hello' + ' World').toBe('Hello World');
        expect('Gode' + ' Runtime').toBe('Gode Runtime');
    });
    
    test('string methods', () => {
        expect('hello'.toUpperCase()).toBe('HELLO');
        expect('WORLD'.toLowerCase()).toBe('world');
        expect('test'.length).toBe(4);
    });
    
    test('string contains', () => {
        expect('Hello World').toContain('World');
        expect('JavaScript').toContain('Script');
        expect('Gode Runtime').toContain('Runtime');
    });
});

describe('Array Operations', () => {
    test('array creation and access', () => {
        const arr = [1, 2, 3, 4, 5];
        expect(arr.length).toBe(5);
        expect(arr[0]).toBe(1);
        expect(arr[4]).toBe(5);
    });
    
    test('array methods', () => {
        const numbers = [1, 2, 3];
        expect(numbers.includes(2)).toBeTruthy();
        expect(numbers.includes(5)).toBeFalsy();
        
        const doubled = numbers.map(x => x * 2);
        expect(doubled).toEqual([2, 4, 6]);
    });
});

describe('Object Operations', () => {
    test('object creation and access', () => {
        const user = {
            name: 'John',
            age: 30,
            active: true
        };
        
        expect(user.name).toBe('John');
        expect(user.age).toBe(30);
        expect(user.active).toBeTruthy();
    });
    
    test('object methods', () => {
        const obj = { a: 1, b: 2, c: 3 };
        const keys = Object.keys(obj);
        expect(keys).toContain('a');
        expect(keys).toContain('b');
        expect(keys).toContain('c');
        expect(keys).toHaveLength(3);
    });
});

describe('Type Checking', () => {
    test('typeof operator', () => {
        expect(typeof 42).toBe('number');
        expect(typeof 'hello').toBe('string');
        expect(typeof true).toBe('boolean');
        expect(typeof {}).toBe('object');
        expect(typeof []).toBe('object');
        expect(typeof undefined).toBe('undefined');
    });
    
    test('null and undefined', () => {
        expect(null).toBeNull();
        expect(undefined).toBeUndefined();
        expect('').toBeDefined();
        expect(0).toBeDefined();
    });
});

describe('Error Handling', () => {
    test('throwing errors', () => {
        expect(() => {
            throw new Error('Test error');
        }).toThrow('Test error');
        
        expect(() => {
            throw new Error('Different error');
        }).toThrow();
    });
    
    test('not throwing errors', () => {
        expect(() => {
            return 'no error';
        }).not.toThrow();
    });
});

console.log("Test definitions complete! Run with: ./gode test examples/test-demo/simple-tests.js");