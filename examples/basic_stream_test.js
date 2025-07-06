// Basic test for stream module using Go constructors directly

console.log('Testing basic stream functionality...');

try {
  if (typeof __gode_stream !== 'undefined') {
    const { Readable, Writable, Transform, PassThrough, pipeline } = __gode_stream;
    console.log('✓ Stream module imported successfully');
    
    // Test 1: Create instances
    console.log('\nTest 1: Creating stream instances');
    
    const readable = new Readable();
    console.log('✓ Readable stream created');
    
    const writable = new Writable();
    console.log('✓ Writable stream created');
    
    const transform = new Transform();
    console.log('✓ Transform stream created');
    
    const passThrough = new PassThrough();
    console.log('✓ PassThrough stream created');
    
    // Test 2: Check methods exist
    console.log('\nTest 2: Checking stream methods');
    
    if (typeof readable.read === 'function') {
      console.log('✓ Readable.read method exists');
    }
    
    if (typeof writable.write === 'function') {
      console.log('✓ Writable.write method exists');
    }
    
    if (typeof transform.read === 'function' && typeof transform.write === 'function') {
      console.log('✓ Transform has both read and write methods');
    }
    
    if (typeof pipeline === 'function') {
      console.log('✓ pipeline function exists');
    }
    
    console.log('\n✓ Basic stream test completed successfully!');
    
  } else {
    console.error('✗ Stream module not found in __gode_stream');
  }
  
} catch (error) {
  console.error('✗ Stream test failed:', error.message);
  console.error(error.stack);
}