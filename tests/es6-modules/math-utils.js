// Math utilities module using ES6 exports
export const PI = 3.14159;
export const E = 2.71828;

export function add(a, b) {
    return a + b;
}

export function multiply(a, b) {
    return a * b;
}

export function power(base, exponent) {
    return Math.pow(base, exponent);
}

export const operations = {
    add,
    multiply,
    power
};

export class Calculator {
    constructor() {
        this.history = [];
    }
    
    calculate(operation, a, b) {
        let result;
        switch(operation) {
            case 'add':
                result = add(a, b);
                break;
            case 'multiply':
                result = multiply(a, b);
                break;
            default:
                result = 0;
        }
        this.history.push({ operation, a, b, result });
        return result;
    }
}

console.log("Math utilities module loaded");