// Complete test for all stream functionality

console.log('Complete Stream Module Test');
console.log('==========================');

try {
  if (typeof __gode_stream !== 'undefined') {
    const { Readable, Writable, Transform, PassThrough, pipeline } = __gode_stream;
    console.log('✓ Stream module imported successfully');
    
    // Test 1: Readable.from with events
    console.log('\nTest 1: Readable.from with event handling');
    const readable = Readable.from(['Hello', ' ', 'World', '!']);
    
    let receivedData = '';
    
    readable.on('data', (chunk) => {
      receivedData += chunk.toString();
      console.log('  Data event:', chunk.toString());
    });
    
    readable.on('end', () => {
      console.log('  End event - Final data:', receivedData);
      console.log('✓ Readable.from with events working');
      
      // Test 2: Writable with events
      console.log('\nTest 2: Writable with events');
      
      const chunks = [];
      const writable = new Writable({
        write(chunk, encoding, callback) {
          chunks.push(chunk.toString());
          console.log('  Wrote:', chunk.toString());
          if (callback) callback();
        }
      });
      
      writable.on('finish', () => {
        console.log('  Finish event - All chunks:', chunks.join(''));
        console.log('✓ Writable with events working');
        
        // Test 3: Transform stream
        console.log('\nTest 3: Transform stream');
        
        const upperTransform = new Transform();
        upperTransform._transform = function(chunk, encoding, callback) {
          const upper = chunk.toString().toUpperCase();
          console.log('  Transforming:', chunk.toString(), '->', upper);
          this.push(upper);
          callback();
        };
        
        upperTransform.on('data', (chunk) => {
          console.log('  Transform output:', chunk.toString());
        });
        
        upperTransform.on('end', () => {
          console.log('✓ Transform stream working');
          
          // Test 4: PassThrough stream
          console.log('\nTest 4: PassThrough stream');
          
          const passThrough = new PassThrough();
          
          passThrough.on('data', (chunk) => {
            console.log('  PassThrough data:', chunk.toString());
          });
          
          passThrough.on('end', () => {
            console.log('✓ PassThrough stream working');
            
            // Test 5: Basic piping
            console.log('\nTest 5: Basic piping');
            
            const source = Readable.from(['pipe', 'test']);
            const dest = new Writable({
              write(chunk, encoding, callback) {
                console.log('  Piped data:', chunk.toString());
                if (callback) callback();
              }
            });
            
            dest.on('finish', () => {
              console.log('✓ Basic piping working');
              console.log('\n🎉 ALL STREAM TESTS PASSED! 🎉');
            });
            
            source.pipe(dest);
          });
          
          passThrough.write('PassThrough ');
          passThrough.write('Test');
          passThrough.end();
        });
        
        upperTransform.write('transform ');
        upperTransform.write('test');
        upperTransform.end();
      });
      
      writable.write('Hello ');
      writable.write('Writable ');
      writable.write('Stream!');
      writable.end();
    });
    
  } else {
    console.error('✗ Stream module not found in __gode_stream');
  }
  
} catch (error) {
  console.error('✗ Complete stream test failed:', error.message);
  console.error(error.stack);
}