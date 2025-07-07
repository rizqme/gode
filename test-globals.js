// Simple test of the new globals
console.log("Testing new globals...");

// Test process
console.log("Process PID:", process.PID);
console.log("Process platform:", process.platform);
console.log("Process argv:", process.Argv);

// Test __dirname and __filename
console.log("__dirname:", __dirname);
console.log("__filename:", __filename);

// Test Buffer
const buf = Buffer.from("Hello World");
console.log("Buffer:", buf.toString());
console.log("Buffer length:", buf.length());

// Test console methods
console.info("This is an info message");
console.warn("This is a warning message");
console.time("test-timer");
console.timeEnd("test-timer");

// Test URL
try {
    const url = new URL("https://example.com:8080/path?query=value#hash");
    console.log("URL href:", url.href());
    console.log("URL hostname:", url.hostname());
} catch (e) {
    console.error("URL test failed:", e.message);
}

// Test base64
try {
    const encoded = btoa("Hello World");
    console.log("Base64 encoded:", encoded);
    const decoded = atob(encoded);
    console.log("Base64 decoded:", decoded);
} catch (e) {
    console.error("Base64 test failed:", e.message);
}

// Test structuredClone
try {
    const original = { name: "test", nested: { value: 42 } };
    const cloned = structuredClone(original);
    console.log("Original:", original);
    console.log("Cloned:", cloned);
    console.log("Are different objects:", original !== cloned);
} catch (e) {
    console.error("structuredClone test failed:", e.message);
}

console.log("All globals test completed!");