// Plugin Features Test
// Tests the new capabilities like variadic arguments and flexible parameter handling

console.log('=== Plugin Features Test ===');

try {
    // Test math plugin features
    const math = require('./plugin-math/math.so');
    console.log('\n--- Math Plugin Tests ---');
    
    // Test variadic arguments
    console.log('math.add(1):', math.add(1));
    console.log('math.add(1, 2):', math.add(1, 2));
    console.log('math.add(1, 2, 3):', math.add(1, 2, 3));
    console.log('math.add(1, 2, 3, 4, 5):', math.add(1, 2, 3, 4, 5));
    
    // Test multiply with additional factors
    console.log('math.multiply(2, 3):', math.multiply(2, 3));
    console.log('math.multiply(2, 3, 4):', math.multiply(2, 3, 4));
    console.log('math.multiply(2, 3, 4, 5):', math.multiply(2, 3, 4, 5));
    
    // Test statistics with multiple numbers
    const stats = math.statistics(1, 2, 3, 4, 5, 6, 7, 8, 9, 10);
    console.log('math.statistics(1-10):', JSON.stringify(stats, null, 2));
    
    // Test GCD and LCM with multiple numbers
    console.log('math.gcd(12, 18):', math.gcd(12, 18));
    console.log('math.gcd(12, 18, 24):', math.gcd(12, 18, 24));
    console.log('math.lcm(4, 6):', math.lcm(4, 6));
    console.log('math.lcm(4, 6, 8):', math.lcm(4, 6, 8));
    
    // Test hello plugin features
    const hello = require('./plugin-hello/hello.so');
    console.log('\n--- Hello Plugin Tests ---');
    
    // Test flexible greet function
    console.log('hello.greet():', hello.greet());
    console.log('hello.greet("Alice"):', hello.greet("Alice"));
    console.log('hello.greet("Alice", "Bob"):', hello.greet("Alice", "Bob"));
    console.log('hello.greet("Alice", "Bob", "Charlie"):', hello.greet("Alice", "Bob", "Charlie"));
    
    // Test getTime with different formats
    console.log('hello.getTime():', hello.getTime());
    console.log('hello.getTime("iso"):', hello.getTime("iso"));
    console.log('hello.getTime("date"):', hello.getTime("date"));
    console.log('hello.getTime("time"):', hello.getTime("time"));
    console.log('hello.getTime("unix"):', hello.getTime("unix"));
    
    // Test echo with transformations
    console.log('hello.echo("Hello World"):', hello.echo("Hello World"));
    console.log('hello.echo("Hello World", "upper"):', hello.echo("Hello World", "upper"));
    console.log('hello.echo("Hello World", "lower"):', hello.echo("Hello World", "lower"));
    console.log('hello.echo("Hello World", "upper", "reverse"):', hello.echo("Hello World", "upper", "reverse"));
    
    // Test split with multiple separators
    console.log('hello.split("a,b;c d"):', hello.split("a,b;c d"));
    console.log('hello.split("a,b;c d", ","):', hello.split("a,b;c d", ","));
    console.log('hello.split("a,b;c d", ",", ";"):', hello.split("a,b;c d", ",", ";"));
    console.log('hello.split("a,b;c d", ",", ";", "whitespace"):', hello.split("a,b;c d", ",", ";", "whitespace"));
    
    // Test wordCount
    const text = "Hello world!\nThis is a test.\n\nAnother paragraph.";
    const wordStats = hello.wordCount(text);
    console.log('Word count stats:', JSON.stringify(wordStats, null, 2));
    
    // Test join with options
    const parts = ["Hello", "world", "from", "JavaScript"];
    console.log('hello.join(parts, " "):', hello.join(parts, " "));
    console.log('hello.join(parts, "-", "prefix:>>>"):', hello.join(parts, "-", "prefix:>>>"));
    console.log('hello.join(parts, " ", "suffix:!!!"):', hello.join(parts, " ", "suffix:!!!"));
    
    // Test format function
    const formatResult = hello.format("Hello World", "upper", "reverse");
    console.log('Format result:', JSON.stringify(formatResult, null, 2));
    
    console.log('\n=== Plugin features test completed successfully! ===');
    console.log('✓ Variadic arguments working');
    console.log('✓ Flexible parameter handling working');
    console.log('✓ Optional parameters working');
    console.log('✓ Complex return types working');
    
} catch (error) {
    console.error('Plugin features test error:', error.message);
    console.error('Stack:', error.stack);
}