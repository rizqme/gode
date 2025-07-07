# Hello Plugin Example

This example demonstrates the hello plugin for Gode runtime.

## Features

- **Greeting Functions**: Personalized greetings
- **String Manipulation**: Reverse, echo operations
- **Time Functions**: Get current timestamp
- **Unicode Support**: Works with international characters

## Building

```bash
make build
```

## Running

```bash
# From the project root
./gode run examples/plugin-hello/demo.js

# Or from this directory
cd examples/plugin-hello
../../gode run demo.js
```

## Available Functions

### `greet(name)`
Creates a personalized greeting.
```javascript
hello.greet("World") // returns "Hello, World!"
```

### `reverse(text)`
Reverses a string with proper Unicode support.
```javascript
hello.reverse("hello") // returns "olleh"
```

### `echo(text)`
Returns the input text unchanged.
```javascript
hello.echo("test") // returns "test"
```

### `getTime()`
Returns the current timestamp.
```javascript
hello.getTime() // returns "2025-01-07 15:30:45"
```

## Notes

- All string functions support Unicode characters
- Time format is "YYYY-MM-DD HH:MM:SS"
- Implemented in Go for reliable string handling