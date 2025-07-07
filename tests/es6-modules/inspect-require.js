// Inspect what require() returns for ES6 modules
console.log("=== Require Inspection Test ===");

console.log("Testing require() directly:");
const result = require('./exports-module.js');
console.log("require() result:", result);
console.log("typeof result:", typeof result);
console.log("result properties:", Object.keys(result || {}));

if (result) {
    console.log("result.namedConst:", result.namedConst);
    console.log("result.default:", result.default);
}

console.log("\n=== Inspection Complete ===");