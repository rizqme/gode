// JSON operations example
console.log("=== JSON Operations Demo ===");

// Create sample data
const data = {
    project: "Gode",
    type: "JavaScript Runtime",
    features: ["ES modules", "TypeScript support", "Go integration"],
    config: {
        version: "0.1.0",
        experimental: true
    },
    stats: {
        tests: 267,
        plugins: 3,
        uptime: 42.5
    }
};

console.log("Original data:");
console.log(data);

// JSON stringify
const jsonString = JSON.stringify(data);
console.log("\nJSON stringified:");
console.log(jsonString);

// JSON parse
const parsed = JSON.parse(jsonString);
console.log("\nParsed back:");
console.log(parsed);

// Pretty formatting
const prettyJson = JSON.stringify(data, null, 2);
console.log("\nPretty formatted JSON:");
console.log(prettyJson);

// Test with arrays
const array = [1, "hello", true, null, {nested: "object"}];
console.log("\nArray to JSON:");
console.log(JSON.stringify(array));

console.log("\nJSON operations demo complete!");