// Simple import test
console.log("=== Simple Import Test ===");

// Test 1: Direct require
console.log("1. Direct require:");
const result1 = require('./basic-export.js');
console.log("Direct require result:", result1);

// Test 2: ES6 import compiled to require
console.log("\n2. ES6 import (compiled to require):");
try {
    import './basic-export.js';
    console.log("ES6 import succeeded");
} catch (e) {
    console.log("ES6 import failed:", e.message);
    console.log("Error type:", typeof e);
    console.log("Error constructor:", e.constructor.name);
}

console.log("\nDone with simple import test");