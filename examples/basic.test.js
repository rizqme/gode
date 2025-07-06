// Basic test examples demonstrating different features

describe('Basic Tests', () => {
  test('simple equality', () => {
    expect(2 + 2).toBe(4);
  });
  
  test('string operations', () => {
    const str = 'Hello World';
    expect(str).toContain('World');
    expect(str).toHaveLength(11);
  });
  
  test('array operations', () => {
    const arr = [1, 2, 3, 4, 5];
    expect(arr).toContain(3);
    expect(arr).toHaveLength(5);
  });
  
  test('boolean checks', () => {
    expect(true).toBeTruthy();
    expect(false).toBeFalsy();
    expect(1).toBeTruthy();
    expect(0).toBeFalsy();
    expect('').toBeFalsy();
    expect('hello').toBeTruthy();
  });
  
  test('null and undefined', () => {
    expect(null).toBeNull();
    expect(undefined).toBeUndefined();
    expect('defined').toBeDefined();
  });
  
  test('numeric comparisons', () => {
    expect(5).toBeGreaterThan(3);
    expect(5).toBeGreaterThanOrEqual(5);
    expect(3).toBeLessThan(5);
    expect(3).toBeLessThanOrEqual(3);
  });
});

describe('Expectation Negation', () => {
  test('not matchers', () => {
    expect(2 + 2).not.toBe(5);
    expect('hello').not.toContain('world');
    expect(true).not.toBeFalsy();
    expect(null).not.toBeDefined();
  });
});

describe('Error Handling', () => {
  test('function that throws', () => {
    function throwError() {
      throw new Error('Something went wrong');
    }
    
    expect(throwError).toThrow();
    expect(throwError).toThrow('Something went wrong');
  });
  
  test('function that does not throw', () => {
    function safeFunction() {
      return 'safe';
    }
    
    expect(safeFunction).not.toThrow();
  });
});

describe('Hooks Demo', () => {
  let counter = 0;
  
  beforeAll(() => {
    console.log('Setting up test suite');
    counter = 0;
  });
  
  afterAll(() => {
    console.log('Cleaning up test suite');
  });
  
  beforeEach(() => {
    console.log('Before each test');
    counter++;
  });
  
  afterEach(() => {
    console.log('After each test');
  });
  
  test('first test', () => {
    expect(counter).toBe(1);
  });
  
  test('second test', () => {
    expect(counter).toBe(2);
  });
});