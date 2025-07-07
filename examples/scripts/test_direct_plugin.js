console.log("Testing direct plugin loading...");

try {
    console.log("Loading math plugin directly...");
    const math = require('./plugins/examples/math/math.so');
    console.log("Math plugin loaded:", typeof math);
    console.log("Math plugin methods:", Object.keys(math));
    
    if (math.add) {
        const result = math.add(5, 3);
        console.log("5 + 3 =", result);
    }
} catch (e) {
    console.error("Error loading math plugin:", e.message);
}

try {
    console.log("\nLoading hello plugin directly...");
    const hello = require('./plugins/examples/hello/hello.so');
    console.log("Hello plugin loaded:", typeof hello);
    console.log("Hello plugin methods:", Object.keys(hello));
    
    if (hello.greet) {
        const result = hello.greet("World");
        console.log("Greeting:", result);
    }
} catch (e) {
    console.error("Error loading hello plugin:", e.message);
}

console.log("\nDirect plugin loading test complete");