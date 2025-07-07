// Minimal test to isolate the issue
try {
    const math = require('./plugins/examples/math/math.so');
    
    // Test 1
    console.log("Test 1: math.add(1, 2) =", math.add(1, 2));
    if (math.add(1, 2) !== 3) {
        throw new Error("Test 1 failed: expected 3, got " + math.add(1, 2));
    }
    
    // Test 2 
    console.log("Test 2: math.add(1, 3) =", math.add(1, 3));
    if (math.add(1, 3) !== 4) {
        throw new Error("Test 2 failed: expected 4, got " + math.add(1, 3));
    }
    
    console.log("All tests passed!");
    
} catch (e) {
    console.error("Error:", e.message);
}