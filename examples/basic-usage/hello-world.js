// Basic Hello World example for Gode runtime
console.log("Hello, World from Gode!");

// Demonstrate basic JavaScript features
const message = "Gode Runtime";
const version = "0.1.0";

console.log(`Welcome to ${message} v${version}`);

// Test basic arithmetic
const sum = 5 + 3;
const product = 4 * 6;
console.log(`Math: 5 + 3 = ${sum}, 4 * 6 = ${product}`);

// Test object creation
const user = {
    name: "Developer",
    language: "JavaScript",
    runtime: "Gode"
};

console.log("User object:", JSON.stringify(user));

// Test array operations
const numbers = [1, 2, 3, 4, 5];
const doubled = numbers.map(n => n * 2);
console.log("Original:", numbers);
console.log("Doubled:", doubled);

console.log("Basic usage demo complete!");