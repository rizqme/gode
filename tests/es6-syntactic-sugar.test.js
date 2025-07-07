// ES6 Syntactic Sugar Test Suite
// Tests modern JavaScript features like arrow functions, spread operators, destructuring, etc.

describe("ES6 Syntactic Sugar", () => {
    describe("Arrow Functions (Lambda)", () => {
        test("basic arrow function with single parameter", () => {
            const square = x => x * x;
            expect(square(5)).toBe(25);
        });

        test("arrow function with multiple parameters", () => {
            const add = (a, b) => a + b;
            expect(add(3, 4)).toBe(7);
        });

        test("arrow function with block body", () => {
            const multiply = (a, b) => {
                const result = a * b;
                return result;
            };
            expect(multiply(6, 7)).toBe(42);
        });

        test("arrow function with no parameters", () => {
            const getValue = () => 42;
            expect(getValue()).toBe(42);
        });

        test("arrow function in array methods", () => {
            const numbers = [1, 2, 3, 4, 5];
            const doubled = numbers.map(x => x * 2);
            expect(doubled).toEqual([2, 4, 6, 8, 10]);
        });

        test("arrow function with implicit return", () => {
            const createObject = name => ({ name: name, type: "user" });
            const obj = createObject("John");
            expect(obj.name).toBe("John");
            expect(obj.type).toBe("user");
        });

        test("nested arrow functions", () => {
            const outer = x => y => x + y;
            const add5 = outer(5);
            expect(add5(3)).toBe(8);
        });

        test("arrow function with destructured parameters", () => {
            const getFullName = ({ first, last }) => `${first} ${last}`;
            expect(getFullName({ first: "John", last: "Doe" })).toBe("John Doe");
        });
    });

    describe("Spread Operator", () => {
        test("spread operator with arrays", () => {
            const arr1 = [1, 2, 3];
            const arr2 = [4, 5, 6];
            const combined = [...arr1, ...arr2];
            expect(combined).toEqual([1, 2, 3, 4, 5, 6]);
        });

        test("spread operator with function calls", () => {
            function sum(a, b, c) {
                return a + b + c;
            }
            const numbers = [1, 2, 3];
            expect(sum(...numbers)).toBe(6);
        });

        test("spread operator with objects", () => {
            const obj1 = { a: 1, b: 2 };
            const obj2 = { c: 3, d: 4 };
            const combined = { ...obj1, ...obj2 };
            expect(combined).toEqual({ a: 1, b: 2, c: 3, d: 4 });
        });

        test("spread operator for object cloning", () => {
            const original = { x: 1, y: 2 };
            const clone = { ...original };
            clone.x = 10;
            expect(original.x).toBe(1);
            expect(clone.x).toBe(10);
        });

        test("spread operator with overriding properties", () => {
            const defaults = { color: "red", size: "medium" };
            const config = { ...defaults, color: "blue" };
            expect(config).toEqual({ color: "blue", size: "medium" });
        });

        test("spread operator in array literals", () => {
            const start = [1, 2];
            const middle = [3, 4];
            const end = [5, 6];
            const result = [...start, ...middle, ...end];
            expect(result).toEqual([1, 2, 3, 4, 5, 6]);
        });
    });

    describe("Destructuring Assignment", () => {
        test("array destructuring basic", () => {
            const [a, b] = [1, 2];
            expect(a).toBe(1);
            expect(b).toBe(2);
        });

        test("array destructuring with skipping", () => {
            const [first, , third] = [1, 2, 3];
            expect(first).toBe(1);
            expect(third).toBe(3);
        });

        test("array destructuring with rest", () => {
            const [first, ...rest] = [1, 2, 3, 4, 5];
            expect(first).toBe(1);
            expect(rest).toEqual([2, 3, 4, 5]);
        });

        test("object destructuring basic", () => {
            const { name, age } = { name: "John", age: 30 };
            expect(name).toBe("John");
            expect(age).toBe(30);
        });

        test("object destructuring with aliases", () => {
            const { name: fullName, age: years } = { name: "John", age: 30 };
            expect(fullName).toBe("John");
            expect(years).toBe(30);
        });

        test("object destructuring with defaults", () => {
            const { name = "Anonymous", age = 0 } = { name: "John" };
            expect(name).toBe("John");
            expect(age).toBe(0);
        });

        test("nested destructuring", () => {
            const user = {
                name: "John",
                address: {
                    city: "New York",
                    zip: "10001"
                }
            };
            const { name, address: { city } } = user;
            expect(name).toBe("John");
            expect(city).toBe("New York");
        });

        test("destructuring in function parameters", () => {
            function greet({ name, age }) {
                return `Hello ${name}, you are ${age} years old`;
            }
            expect(greet({ name: "Alice", age: 25 })).toBe("Hello Alice, you are 25 years old");
        });

        test("destructuring with computed property names", () => {
            const key = "dynamicKey";
            const obj = { [key]: "value" };
            const { [key]: value } = obj;
            expect(value).toBe("value");
        });
    });

    describe("Template Literals", () => {
        test("basic template literal", () => {
            const name = "World";
            const message = `Hello, ${name}!`;
            expect(message).toBe("Hello, World!");
        });

        test("template literal with expressions", () => {
            const a = 5;
            const b = 10;
            const result = `The sum of ${a} and ${b} is ${a + b}`;
            expect(result).toBe("The sum of 5 and 10 is 15");
        });

        test("multiline template literal", () => {
            const multiline = `Line 1
Line 2
Line 3`;
            expect(multiline).toContain("Line 1");
            expect(multiline).toContain("Line 2");
            expect(multiline).toContain("Line 3");
        });

        test("template literal with function calls", () => {
            function formatName(first, last) {
                return `${first} ${last}`;
            }
            const greeting = `Hello, ${formatName("John", "Doe")}!`;
            expect(greeting).toBe("Hello, John Doe!");
        });

        test("template literal with object properties", () => {
            const user = { name: "Alice", age: 30 };
            const info = `User: ${user.name}, Age: ${user.age}`;
            expect(info).toBe("User: Alice, Age: 30");
        });

        test("nested template literals", () => {
            const inner = "inner";
            const outer = `outer ${`nested ${inner}`} template`;
            expect(outer).toBe("outer nested inner template");
        });
    });

    describe("Enhanced Object Literals", () => {
        test("shorthand property names", () => {
            const name = "John";
            const age = 30;
            const person = { name, age };
            expect(person.name).toBe("John");
            expect(person.age).toBe(30);
        });

        test("computed property names", () => {
            const propName = "dynamicProperty";
            const obj = {
                [propName]: "value",
                [`${propName}2`]: "value2"
            };
            expect(obj.dynamicProperty).toBe("value");
            expect(obj.dynamicProperty2).toBe("value2");
        });

        test("method shorthand", () => {
            const obj = {
                greet() {
                    return "Hello!";
                },
                add(a, b) {
                    return a + b;
                }
            };
            expect(obj.greet()).toBe("Hello!");
            expect(obj.add(2, 3)).toBe(5);
        });

        test("getter and setter shorthand", () => {
            const obj = {
                _value: 0,
                get value() {
                    return this._value;
                },
                set value(val) {
                    this._value = val;
                }
            };
            obj.value = 42;
            expect(obj.value).toBe(42);
        });
    });

    describe("Default Parameters", () => {
        test("function with default parameter", () => {
            function greet(name = "World") {
                return `Hello, ${name}!`;
            }
            expect(greet()).toBe("Hello, World!");
            expect(greet("Alice")).toBe("Hello, Alice!");
        });

        test("multiple default parameters", () => {
            function createUser(name = "Anonymous", age = 0, active = true) {
                return { name, age, active };
            }
            const user = createUser();
            expect(user.name).toBe("Anonymous");
            expect(user.age).toBe(0);
            expect(user.active).toBe(true);
        });

        test("default parameters with expressions", () => {
            function multiply(a, b = a * 2) {
                return a * b;
            }
            expect(multiply(3)).toBe(18); // 3 * (3 * 2) = 18
            expect(multiply(3, 4)).toBe(12); // 3 * 4 = 12
        });

        test("default parameters with destructuring", () => {
            function processData({ name = "Unknown", count = 1 } = {}) {
                return `${name}: ${count}`;
            }
            expect(processData()).toBe("Unknown: 1");
            expect(processData({ name: "Test" })).toBe("Test: 1");
            expect(processData({ count: 5 })).toBe("Unknown: 5");
        });
    });

    describe("Rest Parameters", () => {
        test("rest parameters basic", () => {
            function sum(...numbers) {
                return numbers.reduce((a, b) => a + b, 0);
            }
            expect(sum(1, 2, 3, 4)).toBe(10);
        });

        test("rest parameters with other parameters", () => {
            function logMessage(level, ...messages) {
                return `[${level}] ${messages.join(" ")}`;
            }
            expect(logMessage("INFO", "Hello", "World")).toBe("[INFO] Hello World");
        });

        test("rest parameters in arrow functions", () => {
            const multiply = (...numbers) => numbers.reduce((a, b) => a * b, 1);
            expect(multiply(2, 3, 4)).toBe(24);
        });
    });

    describe("Let and Const", () => {
        test("let block scope", () => {
            let x = 1;
            {
                let x = 2;
                expect(x).toBe(2);
            }
            expect(x).toBe(1);
        });

        test("const immutability", () => {
            const obj = { value: 1 };
            obj.value = 2; // This should work (object is mutable)
            expect(obj.value).toBe(2);
        });

        test("const with arrays", () => {
            const arr = [1, 2, 3];
            arr.push(4);
            expect(arr).toEqual([1, 2, 3, 4]);
        });

        test("temporal dead zone with let", () => {
            expect(() => {
                console.log(x); // Should throw ReferenceError
                let x = 1;
            }).toThrow();
        });
    });

    describe("For...of and For...in", () => {
        test("for...of with arrays", () => {
            const arr = [1, 2, 3];
            const result = [];
            for (const item of arr) {
                result.push(item * 2);
            }
            expect(result).toEqual([2, 4, 6]);
        });

        test("for...of with strings", () => {
            const result = [];
            for (const char of "abc") {
                result.push(char);
            }
            expect(result).toEqual(["a", "b", "c"]);
        });

        test("for...in with objects", () => {
            const obj = { a: 1, b: 2, c: 3 };
            const keys = [];
            for (const key in obj) {
                keys.push(key);
            }
            expect(keys).toEqual(["a", "b", "c"]);
        });
    });

    describe("Complex Combinations", () => {
        test("arrow functions with destructuring and defaults", () => {
            const createUser = ({ name = "Anonymous", age = 0 } = {}) => ({ name, age });
            expect(createUser()).toEqual({ name: "Anonymous", age: 0 });
            expect(createUser({ name: "John" })).toEqual({ name: "John", age: 0 });
        });

        test("template literals with destructuring", () => {
            const user = { name: "Alice", details: { age: 30, city: "NYC" } };
            const { name, details: { age, city } } = user;
            const message = `${name} is ${age} years old and lives in ${city}`;
            expect(message).toBe("Alice is 30 years old and lives in NYC");
        });

        test("spread operator with arrow functions and destructuring", () => {
            const numbers = [1, 2, 3, 4, 5];
            const [first, second, ...rest] = numbers;
            const process = (...args) => args.map(x => x * 2);
            expect(process(...rest)).toEqual([6, 8, 10]);
        });

        test("complex object manipulation", () => {
            const defaults = { theme: "light", language: "en" };
            const userPrefs = { theme: "dark" };
            const config = {
                ...defaults,
                ...userPrefs,
                features: ["feature1", "feature2"]
            };
            
            const { theme, language, features } = config;
            expect(theme).toBe("dark");
            expect(language).toBe("en");
            expect(features).toEqual(["feature1", "feature2"]);
        });
    });

    describe("Performance", () => {
        test("arrow functions perform well", () => {
            const start = Date.now();
            const numbers = Array.from({ length: 1000 }, (_, i) => i);
            const result = numbers.map(x => x * 2).filter(x => x > 500);
            const end = Date.now();
            
            expect(result.length).toBeGreaterThan(0);
            expect(end - start).toBeLessThan(1000); // Should complete in under 1 second
        });

        test("destructuring assignment performs well", () => {
            const start = Date.now();
            for (let i = 0; i < 1000; i++) {
                const obj = { a: i, b: i * 2, c: i * 3 };
                const { a, b, c } = obj;
                // Simple operations to prevent optimization
                const sum = a + b + c;
            }
            const end = Date.now();
            
            expect(end - start).toBeLessThan(1000); // Should complete in under 1 second
        });
    });
});