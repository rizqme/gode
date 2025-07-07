// Module with default export for testing default imports  
console.log("Default module loaded");

// Create a default export (this might not work perfectly yet in strict mode)
export const defaultExport = {
    name: "Default Export",
    value: 42,
    type: "object"
};

console.log("Default export created:", defaultExport);