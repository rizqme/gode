# Async Plugin Example

This example demonstrates asynchronous operations using the async plugin for Gode runtime.

## Features

- **Callback-based Async**: Traditional callback patterns
- **Promise-like Objects**: `.then()` and `.catch()` methods
- **Real Concurrency**: Uses Go routines for true async execution
- **Error Handling**: Proper error propagation in both patterns

## Building

```bash
make build
```

## Running

```bash
# From the project root
./gode run examples/plugin-async/demo.js

# Or from this directory
cd examples/plugin-async
../../gode run demo.js
```

## Available Functions

### Callback-based Functions

#### `delayedAdd(a, b, delayMs, callback)`
Performs addition after a delay.
```javascript
async.delayedAdd(5, 3, 100, (error, result) => {
    console.log(result); // 8 (after 100ms)
});
```

#### `delayedMultiply(a, b, delayMs, callback)`
Performs multiplication with error handling for negative numbers.
```javascript
async.delayedMultiply(4, 6, 100, (error, result) => {
    if (error) {
        console.log("Error:", error);
    } else {
        console.log(result); // 24
    }
});
```

#### `fetchData(id, callback)`
Simulates fetching data asynchronously.
```javascript
async.fetchData('user123', (error, data) => {
    console.log(data); // {id: 'user123', name: 'Item user123', value: 70}
});
```

#### `processArray(numbers, callback)`
Processes an array and returns statistics.
```javascript
async.processArray([1, 2, 3, 4, 5], (error, result) => {
    console.log(result); // {sum: 15, count: 5, average: 3}
});
```

### Promise-like Functions

#### `promiseAdd(a, b, delayMs)`
Returns a promise-like object for addition.
```javascript
const promise = async.promiseAdd(10, 5, 100);
promise.then((result) => {
    console.log(result); // 15
});
```

#### `promiseMultiply(a, b, delayMs)`
Returns a promise-like object with error handling.
```javascript
const promise = async.promiseMultiply(-3, 5, 100);
promise.catch((error) => {
    console.log("Error:", error); // "negative numbers not allowed"
});
```

## Notes

- All async operations use real Go routines
- Delays are specified in milliseconds
- Negative numbers cause errors in multiply operations
- Promise-like objects support basic `.then()` and `.catch()` patterns
- All operations run concurrently when called together