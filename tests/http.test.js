describe('HTTP Module (fetch)', () => {
    test('fetch should be available globally', () => {
        expect(typeof fetch).toBe('function');
    });

    test('fetch should return a Promise', () => {
        const result = fetch('https://httpbin.org/get');
        expect(typeof result).toBe('object');
        expect(typeof result.then).toBe('function');
    });

    test('fetch should make GET request successfully', (done) => {
        fetch('https://httpbin.org/get')
            .then(response => {
                expect(response.status).toBe(200);
                expect(response.ok).toBe(true);
                expect(typeof response.body).toBe('string');
                done();
            })
            .catch(error => {
                done(error);
            });
        
        // Keep test alive
        setTimeout(() => done(new Error('Test timeout')), 10000);
    });

    test('fetch should handle POST requests with JSON', (done) => {
        const testData = {
            title: 'Test Post',
            body: 'This is a test',
            userId: 1
        };

        fetch('https://httpbin.org/post', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(testData)
        })
        .then(response => {
            expect(response.status).toBe(200);
            expect(response.ok).toBe(true);
            const responseData = JSON.parse(response.body);
            expect(responseData.json).toEqual(testData);
            done();
        })
        .catch(error => {
            done(error);
        });

        // Keep test alive
        setTimeout(() => done(new Error('Test timeout')), 10000);
    });

    test('fetch should handle custom headers', (done) => {
        fetch('https://httpbin.org/get', {
            headers: {
                'X-Custom-Header': 'test-value',
                'User-Agent': 'Gode-Test-Client'
            }
        })
        .then(response => {
            expect(response.status).toBe(200);
            const responseData = JSON.parse(response.body);
            expect(responseData.headers['X-Custom-Header']).toBe('test-value');
            expect(responseData.headers['User-Agent']).toBe('Gode-Test-Client');
            done();
        })
        .catch(error => {
            done(error);
        });

        // Keep test alive
        setTimeout(() => done(new Error('Test timeout')), 10000);
    });

    test('fetch should handle HTTP errors', (done) => {
        fetch('https://httpbin.org/status/404')
            .then(response => {
                expect(response.status).toBe(404);
                expect(response.ok).toBe(false);
                done();
            })
            .catch(error => {
                done(error);
            });

        // Keep test alive
        setTimeout(() => done(new Error('Test timeout')), 10000);
    });

    test('fetch should handle network errors', (done) => {
        fetch('https://invalid-domain-that-does-not-exist.com')
            .then(response => {
                done(new Error('Should have failed with network error'));
            })
            .catch(error => {
                expect(typeof error).toBe('object');
                done();
            });

        // Keep test alive
        setTimeout(() => done(new Error('Test timeout')), 10000);
    });

    test('fetch should handle timeout option', (done) => {
        fetch('https://httpbin.org/delay/5', {
            timeout: 1000 // 1 second timeout
        })
        .then(response => {
            done(new Error('Should have timed out'));
        })
        .catch(error => {
            expect(typeof error).toBe('object');
            done();
        });

        // Keep test alive
        setTimeout(() => done(new Error('Test timeout')), 10000);
    });

    test('fetch should handle different HTTP methods', (done) => {
        let completed = 0;
        const methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'];
        
        methods.forEach(method => {
            fetch(`https://httpbin.org/${method.toLowerCase()}`, {
                method: method
            })
            .then(response => {
                expect(response.status).toBe(200);
                completed++;
                if (completed === methods.length) {
                    done();
                }
            })
            .catch(error => {
                done(error);
            });
        });

        // Keep test alive
        setTimeout(() => done(new Error('Test timeout')), 15000);
    });
});