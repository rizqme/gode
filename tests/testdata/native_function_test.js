// Test script to validate native function integration works properly
console.log("Testing native function integration...");

// This tests the built-in module which uses native functions internally
const core = require("gode:core");
console.log("Core module platform:", core.platform);
console.log("Core module version:", core.version);

// Test console.log which is a native function
console.log("String test:", "Hello World");
console.log("Number test:", 42);
console.log("Boolean test:", true);
console.log("Array test:", [1, 2, 3]);
console.log("Object test:", {key: "value", number: 123});

// Test JSON which uses native functions
var testObj = {
    name: "Test Object",
    values: [1, 2, 3, 4],
    nested: {
        inner: "value"
    }
};

var jsonString = JSON.stringify(testObj);
console.log("JSON stringify result:", jsonString);

console.log("Native function integration test completed successfully!");