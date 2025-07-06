// Script that intentionally throws an error for testing error handling
console.log("About to test error handling...");

// Throw a standard JavaScript error
throw new Error("This is a test error for error handling validation");

// This line should never be reached
console.log("This should not be printed!");