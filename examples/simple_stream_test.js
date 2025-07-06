// Simple test for stream module integration

console.log('Testing stream module...');

try {
  // Check if stream module is available globally (registered via our bridge)
  if (typeof __gode_stream !== 'undefined') {
    const { Readable, Writable, Transform, pipeline } = __gode_stream;
    console.log('✓ Stream module imported successfully');
  
  // Test 1: Basic Readable
  console.log('\nTest 1: Basic Readable Stream');
  const readable = Readable.from(['hello', ' ', 'world']);
  console.log('✓ Readable stream created');
  
  // Test 2: Basic Writable  
  console.log('\nTest 2: Basic Writable Stream');
  const chunks = [];
  const writable = new Writable({
    write(chunk, encoding, callback) {
      chunks.push(chunk.toString());
      console.log('Received:', chunk.toString());
      callback();
    }
  });
  console.log('✓ Writable stream created');
  
  // Test 3: Simple pipe
  console.log('\nTest 3: Pipe operation');
  readable.on('end', () => {
    console.log('✓ All chunks received:', chunks.join(''));
    console.log('✓ Stream test completed successfully!');
  });
  
  readable.pipe(writable);
  
  } else {
    console.error('✗ Stream module not found in __gode_stream');
    console.log('Available:', typeof __gode_stream !== 'undefined' ? Object.keys(__gode_stream) : 'none');
  }
  
} catch (error) {
  console.error('✗ Stream test failed:', error.message);
  console.error(error.stack);
}