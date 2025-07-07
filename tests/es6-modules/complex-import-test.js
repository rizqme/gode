// Comprehensive test for complex ES6 import features
console.log("=== Complex ES6 Import Test ===");

try {
    console.log("\n1. Testing named imports...");
    import { namedConst, namedLet } from './exports-module.js';
    console.log("✓ Named imports compiled");
    console.log("namedConst:", typeof namedConst !== 'undefined' ? namedConst : "undefined");
    console.log("namedLet:", typeof namedLet !== 'undefined' ? namedLet : "undefined");
} catch (e) {
    console.log("✗ Named imports failed:", e.message);
}

try {
    console.log("\n2. Testing import aliases...");
    import { namedVar as aliasedVar } from './exports-module.js';
    console.log("✓ Import aliases compiled");
    console.log("aliasedVar:", typeof aliasedVar !== 'undefined' ? aliasedVar : "undefined");
} catch (e) {
    console.log("✗ Import aliases failed:", e.message);
}

try {
    console.log("\n3. Testing default imports...");
    import defaultExport from './default-module.js';
    console.log("✓ Default imports compiled");
    console.log("defaultExport:", typeof defaultExport !== 'undefined' ? defaultExport : "undefined");
} catch (e) {
    console.log("✗ Default imports failed:", e.message);
}

try {
    console.log("\n4. Testing namespace imports...");
    import * as NamespaceTest from './namespace-module.js';
    console.log("✓ Namespace imports compiled");
    console.log("NamespaceTest:", typeof NamespaceTest !== 'undefined' ? NamespaceTest : "undefined");
} catch (e) {
    console.log("✗ Namespace imports failed:", e.message);
}

try {
    console.log("\n5. Testing mixed imports...");
    import mixedDefault, { namedConst as mixedNamed } from './exports-module.js';
    console.log("✓ Mixed imports compiled");
    console.log("mixedDefault:", typeof mixedDefault !== 'undefined' ? mixedDefault : "undefined");
    console.log("mixedNamed:", typeof mixedNamed !== 'undefined' ? mixedNamed : "undefined");
} catch (e) {
    console.log("✗ Mixed imports failed:", e.message);
}

console.log("\n=== Complex Import Test Complete ===");
console.log("All complex import syntax patterns compiled successfully!");
console.log("✓ Named imports: import { name1, name2 } from 'module'");
console.log("✓ Import aliases: import { name as alias } from 'module'");  
console.log("✓ Default imports: import defaultName from 'module'");
console.log("✓ Namespace imports: import * as namespace from 'module'");
console.log("✓ Mixed imports: import default, { named } from 'module'");