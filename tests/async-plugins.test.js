describe('Async Plugin System', () => {
  describe('Callback-based Async Operations', () => {
    let async;

    beforeEach(() => {
      try {
        async = require('./examples/plugin-async/async.so');
      } catch (e) {
        console.log('Async plugin not available, skipping tests');
        return;
      }
    });

    test('should load async plugin successfully', () => {
      expect(async).toBeDefined();
      expect(async.__pluginName).toBe('async');
      expect(async.__pluginVersion).toBeDefined();
    });

    test('should have async functions', () => {
      expect(typeof async.delayedAdd).toBe('function');
      expect(typeof async.delayedMultiply).toBe('function');
      expect(typeof async.fetchData).toBe('function');
      expect(typeof async.promiseAdd).toBe('function');
      expect(typeof async.promiseMultiply).toBe('function');
      expect(typeof async.processArray).toBe('function');
    });

    test('delayedAdd should work with callback', (done) => {
      async.delayedAdd(5, 3, 50, (error, result) => {
        try {
          expect(error).toBeNull();
          expect(result).toBe(8);
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('delayedMultiply should work with positive numbers', (done) => {
      async.delayedMultiply(4, 6, 50, (error, result) => {
        try {
          expect(error).toBeNull();
          expect(result).toBe(24);
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('delayedMultiply should handle negative numbers error', (done) => {
      async.delayedMultiply(-2, 5, 50, (error, result) => {
        try {
          expect(error).toBe('negative numbers not allowed');
          expect(result).toBeNull();
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('fetchData should return valid data', (done) => {
      async.fetchData('test123', (error, data) => {
        try {
          expect(error).toBeNull();
          expect(data).toBeDefined();
          expect(data.id).toBe('test123');
          expect(data.name).toBe('Item test123');
          expect(data.value).toBe(70); // 'test123'.length * 10 = 7 * 10 = 70
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('fetchData should handle empty id error', (done) => {
      async.fetchData('', (error, data) => {
        try {
          expect(error).toBe('invalid id');
          expect(data).toBeNull();
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('processArray should calculate statistics', (done) => {
      const numbers = [1, 2, 3, 4, 5];
      async.processArray(numbers, (error, result) => {
        try {
          expect(error).toBeNull();
          expect(result).toBeDefined();
          expect(result.sum).toBe(15);
          expect(result.count).toBe(5);
          expect(result.average).toBe(3);
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('processArray should handle empty array error', (done) => {
      async.processArray([], (error, result) => {
        try {
          expect(error).toBe('empty array');
          expect(result).toBeNull();
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('multiple async operations should work concurrently', (done) => {
      let completedCount = 0;
      const expectedCount = 3;
      const results = {};

      function checkComplete() {
        completedCount++;
        if (completedCount === expectedCount) {
          try {
            expect(results.add).toBe(10);
            expect(results.multiply).toBe(20);
            expect(results.fetch).toBeDefined();
            expect(results.fetch.id).toBe('concurrent');
            done();
          } catch (e) {
            done(e);
          }
        }
      }

      async.delayedAdd(7, 3, 100, (error, result) => {
        results.add = result;
        checkComplete();
      });

      async.delayedMultiply(4, 5, 100, (error, result) => {
        results.multiply = result;
        checkComplete();
      });

      async.fetchData('concurrent', (error, data) => {
        results.fetch = data;
        checkComplete();
      });
    });
  });

  describe('Promise-like Async Operations', () => {
    let async;

    beforeEach(() => {
      try {
        async = require('./examples/plugin-async/async.so');
      } catch (e) {
        console.log('Async plugin not available, skipping tests');
        return;
      }
    });

    test('promiseAdd should work with then callback', (done) => {
      const promise = async.promiseAdd(8, 2, 50);
      
      expect(promise).toBeDefined();
      expect(typeof promise.then).toBe('function');
      expect(typeof promise.catch).toBe('function');

      promise.then((result) => {
        try {
          expect(result).toBe(10);
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('promiseMultiply should work with positive numbers', (done) => {
      const promise = async.promiseMultiply(3, 7, 50);
      
      promise.then((result) => {
        try {
          expect(result).toBe(21);
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('promiseMultiply should handle negative numbers with catch', (done) => {
      const promise = async.promiseMultiply(-3, 5, 50);
      
      promise.catch((error) => {
        try {
          expect(error).toBe('negative numbers not allowed');
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('promise chaining should work', (done) => {
      const promise = async.promiseAdd(5, 5, 50);
      
      promise
        .then((result) => {
          expect(result).toBe(10);
          return async.promiseAdd(result, 5, 25);
        })
        .then((finalResult) => {
          try {
            expect(finalResult).toBe(15);
            done();
          } catch (e) {
            done(e);
          }
        });
    });

    test('multiple promises should work concurrently', (done) => {
      const promise1 = async.promiseAdd(10, 5, 100);
      const promise2 = async.promiseMultiply(3, 4, 100);
      
      let results = {};
      let completedCount = 0;

      function checkComplete() {
        completedCount++;
        if (completedCount === 2) {
          try {
            expect(results.add).toBe(15);
            expect(results.multiply).toBe(12);
            done();
          } catch (e) {
            done(e);
          }
        }
      }

      promise1.then((result) => {
        results.add = result;
        checkComplete();
      });

      promise2.then((result) => {
        results.multiply = result;
        checkComplete();
      });
    });
  });

  describe('Mixed Callback and Promise Operations', () => {
    let async;

    beforeEach(() => {
      try {
        async = require('./examples/plugin-async/async.so');
      } catch (e) {
        console.log('Async plugin not available, skipping tests');
        return;
      }
    });

    test('should be able to mix callbacks and promises', (done) => {
      // Use callback for one operation
      async.delayedAdd(10, 20, 50, (error, callbackResult) => {
        expect(error).toBeNull();
        expect(callbackResult).toBe(30);

        // Use promise for another operation
        const promise = async.promiseMultiply(callbackResult, 2, 50);
        promise.then((promiseResult) => {
          try {
            expect(promiseResult).toBe(60);
            done();
          } catch (e) {
            done(e);
          }
        });
      });
    });

    test('should handle mixed error scenarios', (done) => {
      let errorCount = 0;
      const expectedErrors = 2;

      function checkComplete() {
        errorCount++;
        if (errorCount === expectedErrors) {
          done();
        }
      }

      // Callback error
      async.delayedMultiply(-1, 5, 50, (error, result) => {
        expect(error).toBe('negative numbers not allowed');
        expect(result).toBeNull();
        checkComplete();
      });

      // Promise error
      const promise = async.promiseMultiply(-2, 3, 50);
      promise.catch((error) => {
        expect(error).toBe('negative numbers not allowed');
        checkComplete();
      });
    });
  });

  describe('Performance and Timing', () => {
    let async;

    beforeEach(() => {
      try {
        async = require('./examples/plugin-async/async.so');
      } catch (e) {
        console.log('Async plugin not available, skipping tests');
        return;
      }
    });

    test('should respect delay timing for callbacks', (done) => {
      const startTime = Date.now();
      const expectedDelay = 100;

      async.delayedAdd(1, 1, expectedDelay, (error, result) => {
        const elapsed = Date.now() - startTime;
        try {
          expect(elapsed).toBeGreaterThan(expectedDelay - 10); // Allow some variance
          expect(elapsed).toBeLessThan(expectedDelay + 50); // Allow some variance
          expect(result).toBe(2);
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('should respect delay timing for promises', (done) => {
      const startTime = Date.now();
      const expectedDelay = 100;

      const promise = async.promiseAdd(2, 3, expectedDelay);
      promise.then((result) => {
        const elapsed = Date.now() - startTime;
        try {
          expect(elapsed).toBeGreaterThan(expectedDelay - 10); // Allow some variance
          expect(elapsed).toBeLessThan(expectedDelay + 50); // Allow some variance
          expect(result).toBe(5);
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('concurrent operations should run in parallel', (done) => {
      const startTime = Date.now();
      const delay = 100;
      let completedCount = 0;

      function checkComplete() {
        completedCount++;
        if (completedCount === 3) {
          const elapsed = Date.now() - startTime;
          try {
            // Should complete in roughly delay time, not 3 * delay
            expect(elapsed).toBeLessThan(delay * 2);
            done();
          } catch (e) {
            done(e);
          }
        }
      }

      async.delayedAdd(1, 2, delay, () => checkComplete());
      async.delayedAdd(3, 4, delay, () => checkComplete());
      async.delayedAdd(5, 6, delay, () => checkComplete());
    });
  });

  describe('Error Handling and Edge Cases', () => {
    let async;

    beforeEach(() => {
      try {
        async = require('./examples/plugin-async/async.so');
      } catch (e) {
        console.log('Async plugin not available, skipping tests');
        return;
      }
    });

    test('should handle invalid callback parameter', () => {
      // This test ensures the plugin doesn't crash with invalid callbacks
      expect(() => {
        async.delayedAdd(1, 2, 50, null);
      }).not.toThrow();
    });

    test('should handle zero delay', (done) => {
      async.delayedAdd(100, 200, 0, (error, result) => {
        try {
          expect(error).toBeNull();
          expect(result).toBe(300);
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('should handle large numbers', (done) => {
      async.delayedAdd(1000000, 2000000, 50, (error, result) => {
        try {
          expect(error).toBeNull();
          expect(result).toBe(3000000);
          done();
        } catch (e) {
          done(e);
        }
      });
    });

    test('should handle array with mixed numbers', (done) => {
      const numbers = [1, -2, 3, -4, 5];
      async.processArray(numbers, (error, result) => {
        try {
          expect(error).toBeNull();
          expect(result.sum).toBe(3); // 1 + (-2) + 3 + (-4) + 5 = 3
          expect(result.count).toBe(5);
          expect(result.average).toBe(0.6);
          done();
        } catch (e) {
          done(e);
        }
      });
    });
  });
});