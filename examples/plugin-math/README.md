# Math Plugin Example

This example demonstrates the math plugin for Gode runtime.

## Features

- **Basic Operations**: `add()`, `multiply()`
- **Advanced Functions**: `fibonacci()`, `isPrime()`
- **High Performance**: Implemented in Go for fast computation

## Building

```bash
make build
```

## Running

```bash
# From the project root
./gode run examples/plugin-math/demo.js

# Or from this directory
cd examples/plugin-math
../../gode run demo.js
```

## Available Functions

### `add(a, b)`
Adds two integers.
```javascript
math.add(5, 3) // returns 8
```

### `multiply(a, b)`
Multiplies two integers.
```javascript
math.multiply(4, 6) // returns 24
```

### `fibonacci(n)`
Calculates the nth Fibonacci number.
```javascript
math.fibonacci(10) // returns 55
```

### `isPrime(n)`
Checks if a number is prime.
```javascript
math.isPrime(17) // returns true
math.isPrime(25) // returns false
```

## Notes

- All functions expect integer inputs
- Floating point numbers will be converted to integers
- The plugin is written in Go for optimal performance