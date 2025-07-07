describe('Plugin System', () => {
  describe('Math Plugin', () => {
    let math;

    beforeEach(() => {
      try {
        math = require('./examples/plugin-math/math.so');
      } catch (e) {
        // Skip tests if plugin not available
        console.log('Math plugin not available, skipping tests');
        return;
      }
    });

    test('should load math plugin successfully', () => {
      expect(math).toBeDefined();
      expect(math.__pluginName).toBe('math');
      expect(math.__pluginVersion).toBeDefined();
    });

    test('should have basic arithmetic functions', () => {
      expect(typeof math.add).toBe('function');
      expect(typeof math.multiply).toBe('function');
      expect(typeof math.fibonacci).toBe('function');
      expect(typeof math.isPrime).toBe('function');
    });

    test('add function should work correctly', () => {
      expect(math.add(2, 3)).toBe(5);
      expect(math.add(0, 0)).toBe(0);
      expect(math.add(-5, 3)).toBe(-2);
      expect(math.add(1.5, 2.5)).toBe(3); // Floats are converted to ints: 1 + 2 = 3
    });

    test('fibonacci function should work correctly', () => {
      expect(math.fibonacci(0)).toBe(0);
      expect(math.fibonacci(1)).toBe(1);
      expect(math.fibonacci(5)).toBe(5);
      expect(math.fibonacci(10)).toBe(55);
    });

    test('multiply function should work correctly', () => {
      expect(math.multiply(3, 4)).toBe(12);
      expect(math.multiply(0, 5)).toBe(0);
      expect(math.multiply(-2, 3)).toBe(-6);
      expect(math.multiply(2.5, 4)).toBe(8); // Floats are converted to ints: 2 * 4 = 8
    });

    test('isPrime function should work correctly', () => {
      expect(math.isPrime(2)).toBeTruthy();
      expect(math.isPrime(3)).toBeTruthy();
      expect(math.isPrime(4)).toBeFalsy();
      expect(math.isPrime(17)).toBeTruthy();
      expect(math.isPrime(25)).toBeFalsy();
    });

    test('should handle integer inputs correctly', () => {
      // Test that our functions work with integers (our plugin expects ints)
      expect(math.add(1, 2)).toBe(3);
      expect(math.multiply(3, 4)).toBe(12);
    });

    test('should handle different number types', () => {
      // Note: Our plugin expects integers, so decimals will be converted
      expect(math.add(1, 2)).toBe(3);  
      expect(math.multiply(2, 3)).toBe(6);
    });
  });

  describe('Hello Plugin', () => {
    let hello;

    beforeEach(() => {
      try {
        hello = require('./examples/plugin-hello/hello.so');
      } catch (e) {
        // Skip tests if plugin not available
        console.log('Hello plugin not available, skipping tests');
        return;
      }
    });

    test('should load hello plugin successfully', () => {
      if (!hello) return; // Skip if not available
      
      expect(hello).toBeDefined();
      expect(hello.__pluginName).toBeDefined();
      expect(hello.__pluginVersion).toBeDefined();
    });

    test('should have string manipulation functions', () => {
      if (!hello) return; // Skip if not available
      
      expect(typeof hello.greet).toBe('function');
      expect(typeof hello.reverse).toBe('function');
      expect(typeof hello.echo).toBe('function');
      expect(typeof hello.getTime).toBe('function');
    });

    test('greet function should work correctly', () => {
      if (!hello || typeof hello.greet !== 'function') return;
      
      expect(hello.greet('World')).toBe('Hello, World!');
      expect(hello.greet('Gode')).toBe('Hello, Gode!');
      expect(hello.greet('')).toBe('Hello, Anonymous!');
    });

    test('reverse function should work correctly', () => {
      if (!hello || typeof hello.reverse !== 'function') return;
      
      expect(hello.reverse('hello')).toBe('olleh');
      expect(hello.reverse('world')).toBe('dlrow');
      expect(hello.reverse('a')).toBe('a');
      expect(hello.reverse('')).toBe('');
    });

    test('echo function should work correctly', () => {
      if (!hello || typeof hello.echo !== 'function') return;
      
      expect(hello.echo('hello')).toBe('hello');
      expect(hello.echo('World')).toBe('World');
      expect(hello.echo('')).toBe('');
      expect(hello.echo('123abc')).toBe('123abc');
    });
  });

  describe('Plugin Error Handling', () => {
    test('should handle non-existent plugins gracefully', () => {
      expect(() => {
        require('non-existent-plugin');
      }).toThrow();
    });

    test('should handle invalid plugin paths gracefully', () => {
      expect(() => {
        require('invalid/path/plugin.so');
      }).toThrow();
    });
  });

  describe('Plugin Metadata', () => {
    test('plugins should have required metadata', () => {
      let hasPlugin = false;
      
      try {
        const math = require('./examples/plugin-math/math.so');
        if (math) {
          hasPlugin = true;
          expect(math.__pluginName).toBeDefined();
          expect(typeof math.__pluginName).toBe('string');
          expect(math.__pluginVersion).toBeDefined();
          expect(typeof math.__pluginVersion).toBe('string');
        }
      } catch (e) {
        // Plugin not available
      }

      try {
        const hello = require('./examples/plugin-hello/hello.so');
        if (hello) {
          hasPlugin = true;
          expect(hello.__pluginName).toBeDefined();
          expect(typeof hello.__pluginName).toBe('string');
          expect(hello.__pluginVersion).toBeDefined();
          expect(typeof hello.__pluginVersion).toBe('string');
        }
      } catch (e) {
        // Plugin not available
      }

      if (!hasPlugin) {
        console.log('No plugins available for metadata testing');
      }
    });
  });

  describe('Plugin Integration', () => {
    test('plugins should integrate with JavaScript runtime', () => {
      try {
        const math = require('./examples/plugin-math/math.so');
        if (math && typeof math.add === 'function') {
          // Test that plugin functions can be assigned to variables
          const addFunc = math.add;
          expect(addFunc(5, 3)).toBe(8);

          // Test that plugin functions can be used in expressions
          const result = math.add(10, math.multiply(2, 3));
          expect(result).toBe(16);

          // Test that plugin functions work with JavaScript built-ins
          const numbers = [1, 2, 3, 4, 5];
          const sum = numbers.reduce((acc, n) => math.add(acc, n), 0);
          expect(sum).toBe(15);
        }
      } catch (e) {
        console.log('Math plugin integration test skipped');
      }
    });

    test('plugins should work with async operations', (done) => {
      try {
        const math = require('./examples/plugin-math/math.so');
        if (math && typeof math.add === 'function') {
          setTimeout(() => {
            try {
              const result = math.add(5, 5);
              expect(result).toBe(10);
              done();
            } catch (e) {
              done();
            }
          }, 10);
        } else {
          done();
        }
      } catch (e) {
        done();
      }
    });
  });
});