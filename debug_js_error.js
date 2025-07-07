// Debug script to see what JavaScript errors look like
console.log('=== JavaScript Error Debug ===');

function debugError() {
    try {
        // Create a nested call stack
        function a() { return b(); }
        function b() { return c(); }
        function c() { return d(); }
        function d() { return nonExistentVar.method(); }
        
        a();
    } catch (error) {
        console.log('Caught error in JavaScript:');
        console.log('Error type:', typeof error);
        console.log('Error constructor:', error.constructor.name);
        console.log('Error message:', error.message);
        console.log('Error name:', error.name);
        
        if (error.stack) {
            console.log('Error stack:', error.stack);
        } else {
            console.log('No stack property found');
        }
        
        // Try to get all properties
        console.log('All error properties:');
        for (let prop in error) {
            console.log(`  ${prop}:`, error[prop]);
        }
        
        // Re-throw to see what Go sees
        throw error;
    }
}

debugError();