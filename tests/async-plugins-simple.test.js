describe('Async Plugin System - Goroutines', () => {
  describe('Basic Async Operations', () => {
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
      const startTime = Date.now();
      async.delayedAdd(5, 3, 100, (error, result) => {
        try {
          const elapsed = Date.now() - startTime;
          expect(error).toBeNull();
          expect(result).toBe(8);
          expect(elapsed).toBeGreaterThan(90); // Should take at least 100ms
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

    test('promiseAdd should work with then callback', (done) => {
      const startTime = Date.now();
      const promise = async.promiseAdd(8, 2, 100);
      
      expect(promise).toBeDefined();
      expect(typeof promise.then).toBe('function');
      expect(typeof promise.catch).toBe('function');

      promise.then((result) => {
        try {
          const elapsed = Date.now() - startTime;
          expect(result).toBe(10);
          expect(elapsed).toBeGreaterThan(90); // Should take at least 100ms
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

    test('multiple async operations should work concurrently', (done) => {
      const startTime = Date.now();
      let completedCount = 0;
      const expectedCount = 3;
      const results = {};

      function checkComplete() {
        completedCount++;
        if (completedCount === expectedCount) {
          try {
            const elapsed = Date.now() - startTime;
            expect(results.add).toBe(10);
            expect(results.multiply).toBe(20);
            expect(results.fetch).toBeDefined();
            expect(results.fetch.id).toBe('concurrent');
            // Should complete in roughly 100ms (max delay), not 300ms (sum of delays)
            expect(elapsed).toBeLessThan(200);
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

    test('should respect timing for different delays', (done) => {
      const startTime = Date.now();
      let fastCallbackTime, slowCallbackTime;
      let completedCount = 0;

      function checkComplete() {
        completedCount++;
        if (completedCount === 2) {
          try {
            // Fast callback should complete before slow callback
            expect(fastCallbackTime).toBeLessThan(slowCallbackTime);
            // Time difference should be roughly the delay difference
            expect(slowCallbackTime - fastCallbackTime).toBeGreaterThan(100);
            done();
          } catch (e) {
            done(e);
          }
        }
      }

      async.delayedAdd(1, 1, 50, (error, result) => {
        fastCallbackTime = Date.now() - startTime;
        checkComplete();
      });

      async.delayedAdd(2, 2, 200, (error, result) => {
        slowCallbackTime = Date.now() - startTime;
        checkComplete();
      });
    });
  });
});