console.log("Testing basic functionality...");

// Test basic globals
console.log("typeof process:", typeof process);
console.log("typeof Buffer:", typeof Buffer);
console.log("typeof console:", typeof console);

// Test process properties
if (typeof process === 'object') {
    console.log("process.platform:", process.platform);
    console.log("process.Platform:", process.Platform);
    console.log("process.version:", process.version);
}

// Test Buffer constructor
if (typeof Buffer === 'function') {
    console.log("Buffer.alloc:", typeof Buffer.alloc);
    console.log("Buffer.from:", typeof Buffer.from);
    
    try {
        const buf1 = Buffer.alloc(5);
        console.log("Buffer.alloc worked, length:", buf1 ? buf1.length() : 'undefined');
    } catch (e) {
        console.error("Buffer.alloc failed:", e.message);
    }
}

console.log("Basic test completed!");