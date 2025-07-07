// Async plugin demonstration
console.log("=== Async Plugin Demo ===");

try {
    // Load the async plugin
    const async = require('./async.so');
    
    console.log("Async plugin loaded successfully!");
    console.log("Plugin name:", async.__pluginName);
    console.log("Plugin version:", async.__pluginVersion);
    
    let completedOperations = 0;
    const totalOperations = 6;
    
    function checkComplete() {
        completedOperations++;
        if (completedOperations === totalOperations) {
            console.log("\n🎉 All async operations completed!");
        }
    }
    
    // Test 1: Basic delayed addition
    console.log("\n--- Test 1: Delayed Addition ---");
    const start1 = Date.now();
    async.delayedAdd(10, 20, 200, (error, result) => {
        const elapsed = Date.now() - start1;
        console.log(`✓ delayedAdd(10, 20, 200ms) = ${result} (took ${elapsed}ms)`);
        checkComplete();
    });
    
    // Test 2: Delayed multiplication with success
    console.log("\n--- Test 2: Delayed Multiplication (Success) ---");
    async.delayedMultiply(7, 8, 150, (error, result) => {
        if (error) {
            console.log(`✗ delayedMultiply error: ${error}`);
        } else {
            console.log(`✓ delayedMultiply(7, 8, 150ms) = ${result}`);
        }
        checkComplete();
    });
    
    // Test 3: Delayed multiplication with error
    console.log("\n--- Test 3: Delayed Multiplication (Error) ---");
    async.delayedMultiply(-5, 3, 100, (error, result) => {
        if (error) {
            console.log(`✓ delayedMultiply(-5, 3) correctly failed: ${error}`);
        } else {
            console.log(`✗ delayedMultiply should have failed but got: ${result}`);
        }
        checkComplete();
    });
    
    // Test 4: Fetch data
    console.log("\n--- Test 4: Fetch Data ---");
    async.fetchData('user123', (error, data) => {
        if (error) {
            console.log(`✗ fetchData error: ${error}`);
        } else {
            console.log(`✓ fetchData('user123') =`, data);
        }
        checkComplete();
    });
    
    // Test 5: Process array
    console.log("\n--- Test 5: Process Array ---");
    const numbers = [5, 10, 15, 20, 25];
    async.processArray(numbers, (error, result) => {
        if (error) {
            console.log(`✗ processArray error: ${error}`);
        } else {
            console.log(`✓ processArray([5,10,15,20,25]) =`, result);
        }
        checkComplete();
    });
    
    // Test 6: Promise example
    console.log("\n--- Test 6: Promise Example ---");
    const start2 = Date.now();
    const promise = async.promiseAdd(25, 35, 150);
    promise.then((result) => {
        const elapsed = Date.now() - start2;
        console.log(`✓ promiseAdd(25, 35, 150ms) = ${result} (took ${elapsed}ms)`);
        checkComplete();
    });
    
    console.log("\n⏳ Waiting for all async operations to complete...");
    
    // Keep the script running
    setTimeout(() => {
        console.log("\nDemo timeout reached.");
    }, 2000);
    
} catch (error) {
    console.error("Error loading async plugin:", error.message);
    console.log("Make sure to build the plugin first:");
    console.log("make build");
}

console.log("\nAsync plugin demo started!");