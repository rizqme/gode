// Basic export test - current implementation supports this
export const message = "Hello from ES6 module!";
export const version = "1.0.0";
export let counter = 0;
export var flag = true;

console.log("Basic export module loaded successfully");
console.log("Message:", message);
console.log("Version:", version);
console.log("Counter:", counter);
console.log("Flag:", flag);