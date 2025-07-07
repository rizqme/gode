// Data processor using imports from other ES6 modules
import { add, multiply, Calculator } from "./math-utils.js";
import { capitalize, formatters, StringProcessor } from "./string-utils.js";

export class DataProcessor {
    constructor() {
        this.calc = new Calculator();
        this.stringProc = new StringProcessor();
        this.results = [];
    }
    
    processNumbers(numbers) {
        const sum = numbers.reduce((acc, num) => add(acc, num), 0);
        const product = numbers.reduce((acc, num) => multiply(acc, num), 1);
        
        const result = {
            sum,
            product,
            average: sum / numbers.length,
            count: numbers.length
        };
        
        this.results.push({ type: 'numbers', data: numbers, result });
        return result;
    }
    
    processStrings(strings) {
        const processed = strings.map(str => ({
            original: str,
            capitalized: capitalize(str),
            uppercase: formatters.uppercase(str),
            lowercase: formatters.lowercase(str)
        }));
        
        const result = {
            processed,
            count: strings.length,
            totalLength: strings.reduce((acc, str) => acc + str.length, 0)
        };
        
        this.results.push({ type: 'strings', data: strings, result });
        return result;
    }
    
    getStats() {
        return {
            totalProcessed: this.results.length,
            mathHistory: this.calc.history,
            stringHistory: this.stringProc.getHistory(),
            results: this.results
        };
    }
}

export function processData(data) {
    const processor = new DataProcessor();
    
    if (Array.isArray(data)) {
        if (typeof data[0] === 'number') {
            return processor.processNumbers(data);
        } else if (typeof data[0] === 'string') {
            return processor.processStrings(data);
        }
    }
    
    return { error: "Unsupported data type" };
}

console.log("Data processor module loaded");