describe('Basic HTTP', () => {
    test('fetch should be available globally', () => {
        expect(typeof fetch).toBe('function');
    });

    test('fetch should return a Promise-like object', () => {
        const result = fetch('https://httpbin.org/get');
        expect(typeof result).toBe('object');
        expect(typeof result.then).toBe('function');
    });

    test('fetch should handle simple GET request', () => {
        // Test that fetch doesn't throw on simple call
        expect(() => {
            fetch('https://httpbin.org/get');
        }).not.toThrow();
    });

    test('fetch should handle options object', () => {
        // Test that fetch doesn't throw with options
        expect(() => {
            fetch('https://httpbin.org/get', {
                method: 'GET'
            });
        }).not.toThrow();
    });

    test('fetch should handle POST method', () => {
        // Test that fetch doesn't throw with POST
        expect(() => {
            fetch('https://httpbin.org/post', {
                method: 'POST'
            });
        }).not.toThrow();
    });
});