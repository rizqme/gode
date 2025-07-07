// String utilities module using ES6 exports  
export function capitalize(str) {
    return str.charAt(0).toUpperCase() + str.slice(1);
}

export function reverse(str) {
    return str.split('').reverse().join('');
}

export function repeat(str, times) {
    return str.repeat(times);
}

export const formatters = {
    uppercase: str => str.toUpperCase(),
    lowercase: str => str.toLowerCase(),
    camelCase: str => str.replace(/[-_](.)/g, (_, char) => char.toUpperCase())
};

export class StringProcessor {
    constructor() {
        this.processed = [];
    }
    
    process(str, formatter) {
        const result = formatters[formatter] ? formatters[formatter](str) : str;
        this.processed.push({ original: str, formatted: result, formatter });
        return result;
    }
    
    getHistory() {
        return this.processed;
    }
}

console.log("String utilities module loaded");