// Simple ES6 module for testing
console.log("Simple ES6 module loading...");

export const testExport = "I am an ES6 export!";

// Debug: manually check if __gode_exports is being created
console.log("After first export:");
try {
    console.log("__gode_exports exists:", typeof __gode_exports, __gode_exports);
} catch (e) {
    console.log("__gode_exports not accessible:", e.message);
}

export const numberExport = 42;

console.log("After second export:");
try {
    console.log("__gode_exports exists:", typeof __gode_exports, __gode_exports);
} catch (e) {
    console.log("__gode_exports not accessible:", e.message);
}

console.log("Module loading complete");