// Functional test for stream module with actual data flow

console.log('Testing functional stream operations...');

try {
  if (typeof __gode_stream !== 'undefined') {
    const { Readable, Writable, Transform, PassThrough } = __gode_stream;
    console.log('✓ Stream module imported successfully');
    
    // Test 1: Manual readable with push
    console.log('\nTest 1: Manual Readable Stream');
    
    const readable = new Readable();
    
    // Set up data handler
    readable.on('data', (chunk) => {
      console.log('Received data:', chunk.toString());
    });
    
    readable.on('end', () => {
      console.log('✓ Readable stream ended');
    });
    
    // Push some data
    console.log('Pushing data to readable stream...');
    readable.push('Hello ');
    readable.push('World!');
    readable.push(null); // End the stream
    
    // Test 2: Simple writable
    console.log('\nTest 2: Writable Stream');
    
    const chunks = [];
    const writable = new Writable();
    
    writable.on('finish', () => {
      console.log('✓ Writable finished. Collected:', chunks.join(''));
    });
    
    // Simulate writing data
    console.log('Writing data to writable stream...');
    writable.write('Hello ');
    writable.write('Stream ');
    writable.write('World!');
    writable.end();
    
    // Test 3: PassThrough
    console.log('\nTest 3: PassThrough Stream');
    
    const passThrough = new PassThrough();
    
    passThrough.on('data', (chunk) => {
      console.log('PassThrough data:', chunk.toString());
    });
    
    passThrough.on('end', () => {
      console.log('✓ PassThrough ended');
    });
    
    console.log('Writing to PassThrough...');
    passThrough.write('PassThrough ');
    passThrough.write('Test!');
    passThrough.end();
    
    console.log('\n✓ Functional stream test completed!');
    
  } else {
    console.error('✗ Stream module not found');
  }
  
} catch (error) {
  console.error('✗ Functional stream test failed:', error.message);
  console.error(error.stack);
}