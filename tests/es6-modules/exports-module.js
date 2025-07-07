// Module with various export patterns for testing complex imports
console.log("Exports module loaded");

// Named exports
export const namedConst = "I am a named constant";
export let namedLet = "I am a named let";  
export var namedVar = "I am a named var";

// Multiple exports in one statement would need a different approach
// For now, individual exports work fine

console.log("Named exports created:", { namedConst, namedLet, namedVar });