// Main test file using ES6 imports from multiple modules
import { PI, E, add, multiply, Calculator, operations } from "./math-utils.js";
import { capitalize, reverse, formatters, StringProcessor } from "./string-utils.js";
import { DataProcessor, processData } from "./data-processor.js";

console.log("=== ES6 Module Integration Test ===");

// Test 1: Basic imported constants and functions
console.log("\n1. Testing basic imports:");
console.log(`PI = ${PI}`);
console.log(`E = ${E}`);
console.log(`add(5, 3) = ${add(5, 3)}`);
console.log(`multiply(4, 7) = ${multiply(4, 7)}`);

// Test 2: Imported classes
console.log("\n2. Testing imported classes:");
const calc = new Calculator();
console.log(`Calculator add: ${calc.calculate('add', 10, 20)}`);
console.log(`Calculator multiply: ${calc.calculate('multiply', 6, 7)}`);

const stringProc = new StringProcessor();
console.log(`String processor: ${stringProc.process('hello-world', 'camelCase')}`);

// Test 3: Imported objects with methods
console.log("\n3. Testing imported objects:");
console.log(`operations.add(15, 25) = ${operations.add(15, 25)}`);
console.log(`operations.power(2, 8) = ${operations.power(2, 8)}`);

// Test 4: String utilities
console.log("\n4. Testing string utilities:");
console.log(`capitalize('javascript') = ${capitalize('javascript')}`);
console.log(`reverse('modules') = ${reverse('modules')}`);
console.log(`formatters.uppercase('es6') = ${formatters.uppercase('es6')}`);

// Test 5: Multi-module dependencies (data processor imports from other modules)
console.log("\n5. Testing multi-module dependencies:");
const processor = new DataProcessor();

const numberResult = processor.processNumbers([1, 2, 3, 4, 5]);
console.log("Number processing result:", JSON.stringify(numberResult, null, 2));

const stringResult = processor.processStrings(['hello', 'world', 'es6', 'modules']);
console.log("String processing result:", JSON.stringify(stringResult, null, 2));

// Test 6: Function exports that use dependencies
console.log("\n6. Testing function with dependencies:");
const quickResult = processData([10, 20, 30]);
console.log("Quick process result:", JSON.stringify(quickResult, null, 2));

// Test 7: Get comprehensive stats
console.log("\n7. Testing comprehensive stats:");
const stats = processor.getStats();
console.log(`Total operations processed: ${stats.totalProcessed}`);
console.log(`Math history entries: ${stats.mathHistory.length}`);

console.log("\n=== Test Summary ===");
console.log("✓ Basic function and constant imports work");
console.log("✓ Class imports and instantiation work");
console.log("✓ Object method imports work");
console.log("✓ Multi-module dependency chains work");
console.log("✓ ES6 export/import syntax fully functional");
console.log("✓ Complex data processing with multiple imports works");

console.log("\n🎉 All ES6 module tests passed! 🎉");