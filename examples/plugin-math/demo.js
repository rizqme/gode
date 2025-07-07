// Math plugin demonstration
console.log("=== Math Plugin Demo ===");

try {
    // Load the math plugin using relative path
    const math = require('./math.so');
    
    console.log("Math plugin loaded successfully!");
    console.log("Plugin name:", math.__pluginName);
    console.log("Plugin version:", math.__pluginVersion);
    console.log("Available functions:", Object.keys(math).filter(k => !k.startsWith('__')));
    
    // Basic arithmetic operations
    console.log("\n--- Basic Operations ---");
    console.log("math.add(15, 25) =", math.add(15, 25));
    console.log("math.multiply(7, 8) =", math.multiply(7, 8));
    
    // Fibonacci sequence
    console.log("\n--- Fibonacci Sequence ---");
    for (let i = 0; i <= 10; i++) {
        console.log(`fib(${i}) = ${math.fibonacci(i)}`);
    }
    
    // Prime number testing
    console.log("\n--- Prime Number Testing ---");
    const testNumbers = [2, 3, 4, 17, 25, 29, 100, 101];
    testNumbers.forEach(num => {
        const isPrime = math.isPrime(num);
        console.log(`${num} is ${isPrime ? 'prime' : 'not prime'}`);
    });
    
    // Performance test
    console.log("\n--- Performance Test ---");
    const start = Date.now();
    let result = 0;
    for (let i = 0; i < 10000; i++) {
        result = math.add(result, 1);
    }
    const elapsed = Date.now() - start;
    console.log(`Added 1 ten thousand times: ${result} (${elapsed}ms)`);
    
} catch (error) {
    console.error("Error loading math plugin:", error.message);
    console.log("Make sure to build the plugin first:");
    console.log("make build");
}

console.log("\nMath plugin demo complete!");