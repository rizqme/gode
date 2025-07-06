// Example showing how to use the gode:stream module

import { Readable, Writable, Transform, PassThrough, pipeline } from 'gode:stream';

console.log('Stream Module Example');

// Example 1: Basic Readable Stream
console.log('\n1. Basic Readable Stream');

const readable = Readable.from(['Hello', ' ', 'World', '!']);

readable.on('data', (chunk) => {
  console.log('Received chunk:', chunk.toString());
});

readable.on('end', () => {
  console.log('Readable stream ended');
});

// Example 2: Basic Writable Stream
console.log('\n2. Basic Writable Stream');

const chunks = [];
const writable = new Writable({
  write(chunk, encoding, callback) {
    chunks.push(chunk);
    console.log('Wrote chunk:', chunk.toString());
    callback();
  }
});

writable.on('finish', () => {
  console.log('Writable stream finished');
  console.log('All chunks:', Buffer.concat(chunks).toString());
});

writable.write('Hello ');
writable.write('Writable ');
writable.write('Stream!');
writable.end();

// Example 3: Transform Stream (uppercase)
console.log('\n3. Transform Stream');

class UpperCaseTransform extends Transform {
  _transform(chunk, encoding, callback) {
    const uppercased = chunk.toString().toUpperCase();
    callback(null, Buffer.from(uppercased));
  }
}

const upperTransform = new UpperCaseTransform();

upperTransform.on('data', (chunk) => {
  console.log('Transformed:', chunk.toString());
});

upperTransform.write('hello ');
upperTransform.write('transform ');
upperTransform.write('stream!');
upperTransform.end();

// Example 4: PassThrough Stream
console.log('\n4. PassThrough Stream');

const passThrough = new PassThrough();

passThrough.on('data', (chunk) => {
  console.log('Passed through:', chunk.toString());
});

passThrough.write('PassThrough ');
passThrough.write('works!');
passThrough.end();

// Example 5: Piping Streams
console.log('\n5. Piping Streams');

const source = Readable.from(['pipe ', 'this ', 'data']);
const dest = new Writable({
  write(chunk, encoding, callback) {
    console.log('Piped data:', chunk.toString());
    callback();
  }
});

dest.on('finish', () => {
  console.log('Pipe completed');
});

source.pipe(dest);

// Example 6: Pipeline with Multiple Transforms
console.log('\n6. Pipeline Example');

const sourceData = Readable.from(['transform ', 'me ', 'please']);

const addPrefix = new Transform({
  transform(chunk, encoding, callback) {
    callback(null, Buffer.from('[PREFIX] ' + chunk.toString()));
  }
});

const addSuffix = new Transform({
  transform(chunk, encoding, callback) {
    callback(null, Buffer.from(chunk.toString() + ' [SUFFIX]'));
  }
});

const finalDest = new Writable({
  write(chunk, encoding, callback) {
    console.log('Final result:', chunk.toString());
    callback();
  }
});

// Use pipeline to connect all streams
pipeline(sourceData, addPrefix, addSuffix, finalDest)
  .then(() => {
    console.log('Pipeline completed successfully');
  })
  .catch((err) => {
    console.error('Pipeline error:', err);
  });

// Example 7: Error Handling
console.log('\n7. Error Handling');

const errorStream = new Transform({
  transform(chunk, encoding, callback) {
    if (chunk.toString().includes('error')) {
      callback(new Error('Transform error!'));
    } else {
      callback(null, chunk);
    }
  }
});

errorStream.on('error', (err) => {
  console.log('Caught error:', err.message);
});

errorStream.on('data', (chunk) => {
  console.log('Success:', chunk.toString());
});

errorStream.write('good data');
errorStream.write('this will cause an error');
errorStream.write('more good data');
errorStream.end();

// Example 8: Stream Control (pause/resume)
console.log('\n8. Stream Control');

const controlledStream = Readable.from(['chunk1', 'chunk2', 'chunk3', 'chunk4']);

controlledStream.on('data', (chunk) => {
  console.log('Received:', chunk.toString());
  
  if (chunk.toString() === 'chunk2') {
    console.log('Pausing stream...');
    controlledStream.pause();
    
    setTimeout(() => {
      console.log('Resuming stream...');
      controlledStream.resume();
    }, 1000);
  }
});

controlledStream.on('end', () => {
  console.log('Controlled stream ended');
});

// Example 9: Object Mode
console.log('\n9. Object Mode Stream');

const objectStream = new Transform({
  objectMode: true,
  transform(obj, encoding, callback) {
    // Transform object by adding timestamp
    const transformed = {
      ...obj,
      timestamp: new Date().toISOString()
    };
    callback(null, transformed);
  }
});

objectStream.on('data', (obj) => {
  console.log('Object received:', JSON.stringify(obj));
});

objectStream.write({ id: 1, name: 'Alice' });
objectStream.write({ id: 2, name: 'Bob' });
objectStream.end();

// Example 10: Large Data Processing
console.log('\n10. Large Data Processing');

// Simulate processing large data with backpressure
const largeDataSource = new Readable({
  read() {
    for (let i = 0; i < 1000; i++) {
      if (!this.push(`data-${i}-`)) {
        // If push returns false, stop until drained
        break;
      }
    }
    this.push(null); // End stream
  }
});

const processor = new Transform({
  transform(chunk, encoding, callback) {
    // Simulate some processing time
    setTimeout(() => {
      callback(null, chunk.toString().toUpperCase());
    }, 1);
  }
});

const counter = new Writable({
  write(chunk, encoding, callback) {
    // Just count chunks
    this.count = (this.count || 0) + 1;
    if (this.count % 100 === 0) {
      console.log(`Processed ${this.count} chunks`);
    }
    callback();
  }
});

counter.on('finish', function() {
  console.log(`Total chunks processed: ${this.count}`);
});

largeDataSource.pipe(processor).pipe(counter);

console.log('\nAll stream examples started. Check output above.');