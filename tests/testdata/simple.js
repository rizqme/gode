// Simple test script for basic JavaScript functionality
console.log("Hello from test script!");

// Test variables and basic operations
var a = 10;
var b = 20;
var result = a + b;
console.log("10 + 20 =", result);

// Test string operations
var greeting = "Hello";
var name = "Gode";
console.log(greeting + ", " + name + "!");

// Test array operations
var numbers = [1, 2, 3, 4, 5];
console.log("Array:", numbers);
console.log("Array length:", numbers.length);

// Test object operations
var person = {
    name: "Test User",
    age: 30,
    city: "Test City"
};
console.log("Person:", person.name, "is", person.age, "years old");

// Test function
function multiply(x, y) {
    return x * y;
}
console.log("5 * 6 =", multiply(5, 6));

console.log("Simple test completed successfully!");