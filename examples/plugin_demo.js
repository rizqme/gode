// Plugin Demo - Demonstrates loading and using Go plugins in Gode

console.log("Loading plugins...");

// Load math plugin
const math = require("./plugins/examples/math/math.so");

// Load hello plugin
const hello = require("./plugins/examples/hello/hello.so");

async function demoMathPlugin() {
    console.log("\n=== Math Plugin Demo ===");
    
    try {
        // Test addition
        const sum = await math.add(5, 3);
        console.log(`5 + 3 = ${sum}`);
        
        // Test multiplication
        const product = await math.multiply(4, 7);
        console.log(`4 * 7 = ${product}`);
        
        // Test fibonacci
        const fib = await math.fibonacci(10);
        console.log(`fibonacci(10) = ${fib}`);
        
        // Test prime checking
        const isPrime17 = await math.isPrime(17);
        const isPrime18 = await math.isPrime(18);
        console.log(`isPrime(17) = ${isPrime17}`);
        console.log(`isPrime(18) = ${isPrime18}`);
        
    } catch (error) {
        console.error("Math plugin error:", error.message);
    }
}

async function demoHelloPlugin() {
    console.log("\n=== Hello Plugin Demo ===");
    
    try {
        // Test greeting
        const greeting = await hello.greet("Gode");
        console.log(greeting);
        
        // Test time
        const currentTime = await hello.getTime();
        console.log(`Current time: ${currentTime}`);
        
        // Test echo
        const echo = await hello.echo("Hello from Go plugin!");
        console.log(`Echo: ${echo}`);
        
        // Test reverse
        const reversed = await hello.reverse("Gode Runtime");
        console.log(`Reversed: ${reversed}`);
        
    } catch (error) {
        console.error("Hello plugin error:", error.message);
    }
}

async function main() {
    console.log("Gode Plugin System Demo");
    console.log("=======================");
    
    await demoMathPlugin();
    await demoHelloPlugin();
    
    console.log("\n=== Plugin Info ===");
    console.log(`Math Plugin: ${math.__pluginName} v${math.__pluginVersion}`);
    console.log(`Hello Plugin: ${hello.__pluginName} v${hello.__pluginVersion}`);
    
    console.log("\nDemo complete!");
}

// Run the demo
main().catch(console.error);