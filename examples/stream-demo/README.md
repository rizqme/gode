# Stream Demo

This example demonstrates Node.js-compatible stream operations in Gode.

## Examples

### basic-streams.js
Demonstrates readable, writable, transform, and passthrough streams.
```bash
./gode run examples/stream-demo/basic-streams.js
```

## Features Demonstrated

- **Readable Streams**: Creating and reading data
- **Writable Streams**: Writing and processing data
- **Transform Streams**: Modifying data as it flows
- **PassThrough Streams**: Simple data passthrough
- **Stream Piping**: Connecting streams together
- **Event Handling**: Stream events like 'data', 'end', 'finish'

## Stream Types Available

- `stream.Readable` - For reading data
- `stream.Writable` - For writing data
- `stream.Transform` - For transforming data
- `stream.PassThrough` - For passing data through
- `stream.Duplex` - For bidirectional streams

## Running Examples

From the project root:
```bash
./gode run examples/stream-demo/basic-streams.js
```

## Notes

- Streams are compatible with Node.js stream API
- All stream operations are implemented in Go for performance
- Event emitter functionality is fully supported