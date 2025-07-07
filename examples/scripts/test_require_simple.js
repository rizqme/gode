console.log("Testing simple require...");

try {
    console.log("Attempting to require built-in stream...");
    const stream = require('stream');
    console.log("Stream module loaded:", typeof stream);
} catch (e) {
    console.error("Error loading stream:", e);
}

console.log("Simple require test complete");