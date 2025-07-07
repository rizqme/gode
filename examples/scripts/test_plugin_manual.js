console.log("Testing plugin loading manually...");

try {
    console.log("Attempting to load math-plugin...");
    const math = require('math-plugin');
    console.log("Math plugin loaded:", math.__pluginName);
    console.log("2 + 3 =", math.add(2, 3));
} catch (e) {
    console.error("Error loading math plugin:", e);
}

console.log("Manual test complete");