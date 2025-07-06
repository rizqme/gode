// Script for testing built-in module functionality
console.log("Testing built-in modules...");

// Test gode:core module
try {
    const core = require("gode:core");
    console.log("gode:core module loaded successfully");
    console.log("Platform:", core.platform);
    console.log("Version:", core.version);
} catch (e) {
    console.error("Failed to load gode:core:", e.message);
}

// Test console functionality
console.log("Testing console.log with different types:");
console.log("String:", "test string");
console.log("Number:", 42);
console.log("Boolean:", true);
console.log("Array:", [1, 2, 3]);
console.log("Object:", {key: "value"});

// Test JSON functionality
try {
    var obj = {name: "test", value: 123, active: true};
    var jsonString = JSON.stringify(obj);
    console.log("JSON.stringify result:", jsonString);
} catch (e) {
    console.error("JSON.stringify error:", e.message);
}

console.log("Built-in modules test completed!");