// Simple Plugin Test for Current System
// This tests the basic functionality that should work with the current implementation

console.log('=== Simple Plugin Test ===');

try {
    // Load the basic plugins (old format should still work)
    console.log('Loading plugins...');
    
    // Test basic math plugin (if available)
    try {
        const math = require('./plugin-math/math.so');
        console.log('✓ Math plugin loaded');
        
        // Test basic functions
        const sum = math.add(5, 3);
        console.log('math.add(5, 3):', sum);
        
        const product = math.multiply(4, 6);
        console.log('math.multiply(4, 6):', product);
        
        const fib = math.fibonacci(8);
        console.log('math.fibonacci(8):', fib);
        
        const isPrime = math.isPrime(17);
        console.log('math.isPrime(17):', isPrime);
        
    } catch (error) {
        console.log('Math plugin error:', error.message);
    }
    
    // Test basic hello plugin (if available)
    try {
        const hello = require('./plugin-hello/hello.so');
        console.log('✓ Hello plugin loaded');
        
        // Test basic functions
        const greeting = hello.greet('World');
        console.log('hello.greet("World"):', greeting);
        
        const echo = hello.echo('Hello JavaScript!');
        console.log('hello.echo("Hello JavaScript!"):', echo);
        
        const reversed = hello.reverse('JavaScript');
        console.log('hello.reverse("JavaScript"):', reversed);
        
        const time = hello.getTime();
        console.log('hello.getTime():', time);
        
    } catch (error) {
        console.log('Hello plugin error:', error.message);
    }
    
    // Test basic async plugin (if available)
    try {
        const async = require('./plugin-async/async.so');
        console.log('✓ Async plugin loaded');
        
        // Test callback-based functions
        async.delayedAdd(10, 20, 50, function(error, result) {
            if (error) {
                console.log('DelayedAdd error:', error);
            } else {
                console.log('DelayedAdd result:', result);
            }
        });
        
        async.fetchData('test123', function(error, data) {
            if (error) {
                console.log('FetchData error:', error);
            } else {
                console.log('FetchData result:', data);
            }
        });
        
        // Test promise-like functions (if they work)
        try {
            const promiseResult = async.promiseAdd(5, 7, 100);
            if (promiseResult && promiseResult.then) {
                promiseResult.then(function(result) {
                    console.log('PromiseAdd result:', result);
                });
            } else {
                console.log('PromiseAdd returned:', promiseResult);
            }
        } catch (promiseError) {
            console.log('PromiseAdd error:', promiseError.message);
        }
        
    } catch (error) {
        console.log('Async plugin error:', error.message);
    }
    
    console.log('\n=== Basic plugin test completed ===');
    console.log('Note: Some async operations may complete in the background');
    
} catch (error) {
    console.error('Test error:', error.message);
    console.error('Stack:', error.stack);
}