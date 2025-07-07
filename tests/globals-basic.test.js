// Basic test suite for new global objects and functions
describe("Basic Global Objects", () => {
    
    describe("Global Availability", () => {
        test("__dirname should be available", () => {
            expect(typeof __dirname).toBe("string");
        });
        
        test("__filename should be available", () => {
            expect(typeof __filename).toBe("string");
        });
        
        test("process should be available", () => {
            expect(typeof process).toBe("object");
            expect(process).not.toBeNull();
        });
        
        test("Buffer should be available", () => {
            expect(typeof Buffer).toBe("function");
        });
        
        test("setImmediate should be available", () => {
            expect(typeof setImmediate).toBe("function");
        });
        
        test("clearImmediate should be available", () => {
            expect(typeof clearImmediate).toBe("function");
        });
        
        test("queueMicrotask should be available", () => {
            expect(typeof queueMicrotask).toBe("function");
        });
        
        test("URL should be available", () => {
            expect(typeof URL).toBe("function");
        });
        
        test("URLSearchParams should be available", () => {
            expect(typeof URLSearchParams).toBe("function");
        });
        
        test("TextEncoder should be available", () => {
            expect(typeof TextEncoder).toBe("function");
        });
        
        test("TextDecoder should be available", () => {
            expect(typeof TextDecoder).toBe("function");
        });
        
        test("btoa should be available", () => {
            expect(typeof btoa).toBe("function");
        });
        
        test("atob should be available", () => {
            expect(typeof atob).toBe("function");
        });
        
        test("structuredClone should be available", () => {
            expect(typeof structuredClone).toBe("function");
        });
        
        test("global should be available", () => {
            expect(typeof global).toBe("object");
        });
        
        test("module should be available", () => {
            expect(typeof module).toBe("object");
        });
        
        test("exports should be available", () => {
            expect(typeof exports).toBe("object");
        });
    });
    
    describe("Enhanced Console", () => {
        test("console should have basic methods", () => {
            expect(typeof console.log).toBe("function");
            expect(typeof console.error).toBe("function");
        });
        
        test("console should have extended methods", () => {
            expect(typeof console.info).toBe("function");
            expect(typeof console.warn).toBe("function");
            expect(typeof console.debug).toBe("function");
            expect(typeof console.time).toBe("function");
            expect(typeof console.timeEnd).toBe("function");
            expect(typeof console.group).toBe("function");
            expect(typeof console.groupEnd).toBe("function");
            expect(typeof console.assert).toBe("function");
            expect(typeof console.count).toBe("function");
            expect(typeof console.dir).toBe("function");
            expect(typeof console.trace).toBe("function");
            expect(typeof console.clear).toBe("function");
        });
        
        test("console methods should not throw", () => {
            expect(() => {
                console.info("test");
                console.warn("test");
                console.debug("test");
                console.time("timer");
                console.timeEnd("timer");
                console.group("group");
                console.groupEnd();
                console.assert(true);
                console.count("counter");
                console.dir({test: true});
                console.clear();
                console.trace("trace");
            }).not.toThrow();
        });
    });
    
    describe("Base64 Functions", () => {
        test("btoa should encode strings", () => {
            const result = btoa("Hello");
            expect(typeof result).toBe("string");
            expect(result.length).toBeGreaterThan(0);
        });
        
        test("atob should decode strings", () => {
            const encoded = btoa("Hello");
            const decoded = atob(encoded);
            expect(decoded).toBe("Hello");
        });
        
        test("btoa/atob round trip", () => {
            const original = "Hello World";
            const encoded = btoa(original);
            const decoded = atob(encoded);
            expect(decoded).toBe(original);
        });
    });
    
    describe("Directory and File Globals", () => {
        test("__dirname should be a non-empty string", () => {
            expect(__dirname.length).toBeGreaterThan(0);
            expect(__dirname).toContain("/");
        });
        
        test("__filename should be a non-empty string", () => {
            expect(__filename.length).toBeGreaterThan(0);
            expect(__filename).toContain("/");
            expect(__filename).toContain(".js");
        });
    });
    
    describe("Timer Functions", () => {
        test("setImmediate should return ID", () => {
            const id = setImmediate(() => {});
            expect(typeof id).toBe("number");
            expect(id).toBeGreaterThan(0);
            clearImmediate(id);
        });
        
        test("queueMicrotask should accept function", () => {
            expect(() => {
                queueMicrotask(() => {});
            }).not.toThrow();
        });
    });
    
    describe("Module System", () => {
        test("module should have exports property", () => {
            expect(typeof module.exports).toBe("object");
        });
        
        test("exports should be object", () => {
            expect(typeof exports).toBe("object");
        });
        
        test("global should reference global object", () => {
            expect(typeof global).toBe("object");
            expect(global).not.toBeNull();
        });
    });
    
    describe("Process Object Basics", () => {
        test("process should be an object", () => {
            expect(typeof process).toBe("object");
            expect(process).not.toBeNull();
        });
        
        test("process should have some basic properties", () => {
            // Test properties that should exist (using capital case as that's what Goja exposes)
            expect(process.PID).toBeDefined();
            expect(typeof process.PID).toBe("number");
            expect(process.PID).toBeGreaterThan(0);
        });
    });
});