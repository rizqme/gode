// Test file to verify stacktrace system
console.log("Testing stacktrace system...");

// Test 1: Basic JavaScript error
function testBasicError() {
    console.log("Test 1: Basic JavaScript error");
    try {
        throw new Error("This is a basic error");
    } catch (e) {
        console.log("Caught error:", e.message);
    }
}

// Test 2: Module loading error
function testModuleError() {
    console.log("Test 2: Module loading error");
    try {
        require("./nonexistent-module.js");
    } catch (e) {
        console.log("Caught module error:", e.message);
    }
}

// Test 3: Plugin loading error
function testPluginError() {
    console.log("Test 3: Plugin loading error");
    try {
        require("./nonexistent-plugin.so");
    } catch (e) {
        console.log("Caught plugin error:", e.message);
    }
}

// Test 4: Reference error
function testReferenceError() {
    console.log("Test 4: Reference error");
    try {
        console.log(undefinedVariable);
    } catch (e) {
        console.log("Caught reference error:", e.message);
    }
}

// Test 5: Type error
function testTypeError() {
    console.log("Test 5: Type error");
    try {
        null.someMethod();
    } catch (e) {
        console.log("Caught type error:", e.message);
    }
}

// Test 6: Nested function error
function level1() {
    level2();
}

function level2() {
    level3();
}

function level3() {
    throw new Error("Deep stack error");
}

function testNestedError() {
    console.log("Test 6: Nested function error");
    try {
        level1();
    } catch (e) {
        console.log("Caught nested error:", e.message);
    }
}

// Run all tests
testBasicError();
testModuleError();
testPluginError();
testReferenceError();
testTypeError();
testNestedError();

console.log("All error tests completed successfully!");