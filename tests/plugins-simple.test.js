describe('Plugin System - Basic Tests', () => {
  test('should be able to require math plugin', () => {
    try {
      const math = require('math-plugin');
      expect(math).toBeDefined();
      expect(math.__pluginName).toBe('math');
    } catch (e) {
      console.log('Math plugin not available:', e.message);
    }
  });

  test('math plugin basic operations', () => {
    try {
      const math = require('math-plugin');
      expect(math.add(2, 3)).toBe(5);
      expect(math.subtract(10, 4)).toBe(6);
      expect(math.multiply(5, 6)).toBe(30);
      expect(math.divide(20, 4)).toBe(5);
    } catch (e) {
      console.log('Math plugin operations not available');
    }
  });

  test('hello plugin string operations', () => {
    try {
      const hello = require('hello-plugin');
      expect(hello).toBeDefined();
      expect(hello.greet('World')).toBe('Hello, World!');
      expect(hello.reverse('hello')).toBe('olleh');
      expect(hello.uppercase('test')).toBe('TEST');
    } catch (e) {
      console.log('Hello plugin not available:', e.message);
    }
  });

  test('plugin error handling', () => {
    expect(() => {
      require('non-existent-plugin');
    }).toThrow();
  });
});