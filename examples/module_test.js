// Basic ES6 import/export syntax test

// Test import statement parsing
import "./test-module.js";

// Test export statement parsing
export const message = "Hello from ES6 modules!";
export function greet(name) {
    return `Hello, ${name}!`;
}

console.log("ES6 module syntax parsing test completed successfully!");
console.log("This script demonstrates that import/export statements are now parsed correctly.");