// Hello plugin demonstration
console.log("=== Hello Plugin Demo ===");

try {
    // Load the hello plugin using relative path
    const hello = require('./hello.so');
    
    console.log("Hello plugin loaded successfully!");
    console.log("Plugin name:", hello.__pluginName);
    console.log("Plugin version:", hello.__pluginVersion);
    console.log("Available functions:", Object.keys(hello).filter(k => !k.startsWith('__')));
    
    // Greeting operations
    console.log("\n--- Greeting Functions ---");
    console.log(hello.greet("World"));
    console.log(hello.greet("Gode Runtime"));
    console.log(hello.greet("JavaScript Developer"));
    
    // String manipulation
    console.log("\n--- String Operations ---");
    const testString = "Hello, Gode!";
    console.log(`Original: "${testString}"`);
    console.log(`Reversed: "${hello.reverse(testString)}"`);
    console.log(`Echo: "${hello.echo(testString)}"`);
    
    // More string tests
    console.log("\n--- More String Tests ---");
    const words = ["programming", "javascript", "golang", "runtime"];
    words.forEach(word => {
        console.log(`"${word}" reversed is "${hello.reverse(word)}"`);
    });
    
    // Current time from plugin
    console.log("\n--- Time Function ---");
    console.log("Current time from plugin:", hello.getTime());
    
    // Performance test
    console.log("\n--- Performance Test ---");
    const start = Date.now();
    let result = "";
    for (let i = 0; i < 1000; i++) {
        result = hello.echo(`test${i}`);
    }
    const elapsed = Date.now() - start;
    console.log(`Echo test 1000 times: last result "${result}" (${elapsed}ms)`);
    
} catch (error) {
    console.error("Error loading hello plugin:", error.message);
    console.log("Make sure to build the plugin first:");
    console.log("make build");
}

console.log("\nHello plugin demo complete!");