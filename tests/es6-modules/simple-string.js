// Simple string module with only const exports
export const GREETING = "Hello";
export const FAREWELL = "Goodbye";
export const SEPARATOR = " - ";
export const message = GREETING + SEPARATOR + "World";

console.log("Simple string module loaded");
console.log("Message:", message);