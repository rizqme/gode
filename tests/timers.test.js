describe('Timers Module', () => {
    describe('setTimeout', () => {
        test('setTimeout should be available globally', () => {
            expect(typeof setTimeout).toBe('function');
        });

        test('setTimeout should execute callback after delay', (done) => {
            let executed = false;
            const startTime = Date.now();

            setTimeout(() => {
                executed = true;
                const elapsed = Date.now() - startTime;
                expect(elapsed).toBeGreaterThan(90); // Allow some margin
                expect(elapsed).toBeLessThan(200);
                done();
            }, 100);

            expect(executed).toBe(false);
        });

        test('setTimeout should return a timer ID', () => {
            const timerId = setTimeout(() => {}, 100);
            expect(typeof timerId).toBe('number');
            expect(timerId).toBeGreaterThan(0);
            clearTimeout(timerId);
        });

        test('setTimeout should execute immediately with 0 delay', (done) => {
            const startTime = Date.now();

            setTimeout(() => {
                const elapsed = Date.now() - startTime;
                expect(elapsed).toBeLessThan(50); // Should be very fast
                done();
            }, 0);
        });

        test('setTimeout should pass arguments to callback', (done) => {
            setTimeout((arg1, arg2, arg3) => {
                expect(arg1).toBe('hello');
                expect(arg2).toBe(42);
                expect(arg3).toBe(true);
                done();
            }, 10, 'hello', 42, true);
        });

        test('setTimeout should handle multiple concurrent timers', (done) => {
            let count = 0;
            const results = [];

            for (let i = 0; i < 5; i++) {
                setTimeout((index) => {
                    results.push(index);
                    count++;
                    if (count === 5) {
                        expect(results).toHaveLength(5);
                        expect(results).toContain(0);
                        expect(results).toContain(1);
                        expect(results).toContain(2);
                        expect(results).toContain(3);
                        expect(results).toContain(4);
                        done();
                    }
                }, 10 + i * 5, i);
            }
        });
    });

    describe('clearTimeout', () => {
        test('clearTimeout should be available globally', () => {
            expect(typeof clearTimeout).toBe('function');
        });

        test('clearTimeout should cancel a timeout', (done) => {
            let executed = false;

            const timerId = setTimeout(() => {
                executed = true;
            }, 100);

            clearTimeout(timerId);

            setTimeout(() => {
                expect(executed).toBe(false);
                done();
            }, 150);
        });

        test('clearTimeout should handle invalid timer IDs gracefully', () => {
            expect(() => {
                clearTimeout(-1);
                clearTimeout(99999);
                clearTimeout(0);
            }).not.toThrow();
        });
    });

    describe('setInterval', () => {
        test('setInterval should be available globally', () => {
            expect(typeof setInterval).toBe('function');
        });

        test('setInterval should execute callback repeatedly', (done) => {
            let count = 0;
            const startTime = Date.now();

            const intervalId = setInterval(() => {
                count++;
                if (count === 3) {
                    clearInterval(intervalId);
                    const elapsed = Date.now() - startTime;
                    expect(elapsed).toBeGreaterThan(250); // 3 intervals of ~100ms
                    expect(elapsed).toBeLessThan(400);
                    done();
                }
            }, 100);
        });

        test('setInterval should return a timer ID', () => {
            const intervalId = setInterval(() => {}, 100);
            expect(typeof intervalId).toBe('number');
            expect(intervalId).toBeGreaterThan(0);
            clearInterval(intervalId);
        });

        test('setInterval should pass arguments to callback', (done) => {
            let callCount = 0;

            const intervalId = setInterval((arg1, arg2) => {
                expect(arg1).toBe('test');
                expect(arg2).toBe(123);
                callCount++;
                if (callCount === 2) {
                    clearInterval(intervalId);
                    done();
                }
            }, 50, 'test', 123);
        });

        test('setInterval should handle minimum interval correctly', (done) => {
            let count = 0;
            const startTime = Date.now();

            const intervalId = setInterval(() => {
                count++;
                if (count === 5) {
                    clearInterval(intervalId);
                    const elapsed = Date.now() - startTime;
                    expect(elapsed).toBeGreaterThan(40); // 5 intervals of ~10ms
                    expect(elapsed).toBeLessThan(100);
                    done();
                }
            }, 10);
        });
    });

    describe('clearInterval', () => {
        test('clearInterval should be available globally', () => {
            expect(typeof clearInterval).toBe('function');
        });

        test('clearInterval should stop interval execution', (done) => {
            let count = 0;

            const intervalId = setInterval(() => {
                count++;
                if (count === 2) {
                    clearInterval(intervalId);
                }
            }, 50);

            setTimeout(() => {
                expect(count).toBe(2);
                done();
            }, 200);
        });

        test('clearInterval should handle invalid timer IDs gracefully', () => {
            expect(() => {
                clearInterval(-1);
                clearInterval(99999);
                clearInterval(0);
            }).not.toThrow();
        });
    });

    describe('Timer interactions', () => {
        test('setTimeout and setInterval should have different ID spaces', () => {
            const timeoutId = setTimeout(() => {}, 100);
            const intervalId = setInterval(() => {}, 100);

            expect(timeoutId).not.toBe(intervalId);

            clearTimeout(timeoutId);
            clearInterval(intervalId);
        });

        test('clearing wrong timer type should not affect other timers', (done) => {
            let timeoutExecuted = false;
            let intervalCount = 0;

            const timeoutId = setTimeout(() => {
                timeoutExecuted = true;
            }, 100);

            const intervalId = setInterval(() => {
                intervalCount++;
                if (intervalCount === 3) {
                    clearInterval(intervalId);
                    expect(timeoutExecuted).toBe(true);
                    done();
                }
            }, 50);

            // Try to clear timeout with clearInterval and vice versa
            clearInterval(timeoutId); // Should not affect timeout
            clearTimeout(intervalId); // Should not affect interval
        });

        test('multiple timers should execute in correct order', (done) => {
            const results = [];

            setTimeout(() => results.push('timeout-100'), 100);
            setTimeout(() => results.push('timeout-50'), 50);
            setTimeout(() => results.push('timeout-150'), 150);
            setTimeout(() => results.push('timeout-25'), 25);

            setTimeout(() => {
                expect(results).toEqual([
                    'timeout-25',
                    'timeout-50', 
                    'timeout-100',
                    'timeout-150'
                ]);
                done();
            }, 200);
        });
    });

    describe('Error handling', () => {
        test('setTimeout should handle non-function callbacks gracefully', () => {
            expect(() => {
                setTimeout(null, 100);
            }).not.toThrow();

            expect(() => {
                setTimeout(undefined, 100);
            }).not.toThrow();

            expect(() => {
                setTimeout('not a function', 100);
            }).not.toThrow();
        });

        test('setInterval should handle non-function callbacks gracefully', () => {
            expect(() => {
                const id = setInterval(null, 100);
                clearInterval(id);
            }).not.toThrow();
        });

        test('setTimeout should handle negative delays', (done) => {
            const startTime = Date.now();

            setTimeout(() => {
                const elapsed = Date.now() - startTime;
                expect(elapsed).toBeLessThan(50); // Should execute quickly
                done();
            }, -100);
        });

        test('setInterval should handle very small intervals', (done) => {
            let count = 0;

            const intervalId = setInterval(() => {
                count++;
                if (count === 3) {
                    clearInterval(intervalId);
                    expect(count).toBe(3);
                    done();
                }
            }, 1); // 1ms interval
        });
    });
});