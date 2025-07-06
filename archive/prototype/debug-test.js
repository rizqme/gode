// Simple debug test to verify async handling works
const srv = new HttpServer();

// Simple sync endpoint (should work)
srv.handle("GET", "/test-sync", (req, res) => {
  console.log("Sync handler called");
  res.writeHeader(200);
  res.header("Content-Type", "application/json");
  res.write(JSON.stringify({ message: "sync works", timestamp: Date.now() }));
  res.end();
});

// Simple async endpoint using delay (should work)
srv.handle("GET", "/test-async", async (req, res) => {
  console.log("Async handler called");
  await delay(100);
  res.writeHeader(200);
  res.header("Content-Type", "application/json");
  res.write(JSON.stringify({ message: "async works", timestamp: Date.now() }));
  res.end();
});

// Test single Go API call
srv.handle("GET", "/test-go-api", async (req, res) => {
  console.log("Go API handler called");
  try {
    const result = await simulateApiCall("test-service", 100);
    console.log("Got result:", result);
    res.writeHeader(200);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({ message: "go api works", result: result, timestamp: Date.now() }));
    res.end();
  } catch (error) {
    console.log("Error in Go API:", error);
    res.writeHeader(500);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({ error: error.message }));
    res.end();
  }
});

// Test batch Go API call
srv.handle("GET", "/test-go-batch", async (req, res) => {
  console.log("Go batch API handler called");
  try {
    const results = await batchApiCall(["service1", "service2", "service3"], 50);
    console.log("Got batch results:", results);
    res.writeHeader(200);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({ message: "go batch works", results: results, timestamp: Date.now() }));
    res.end();
  } catch (error) {
    console.log("Error in Go batch API:", error);
    res.writeHeader(500);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({ error: error.message }));
    res.end();
  }
});

// Health check
srv.handle("GET", "/health", (req, res) => {
  res.writeHeader(200);
  res.header("Content-Type", "application/json");
  res.write(JSON.stringify({ status: "healthy", timestamp: Date.now() }));
  res.end();
});

console.log("🧪 Debug Test Server");
console.log("Endpoints:");
console.log("  GET /test-sync      - Simple sync endpoint");
console.log("  GET /test-async     - Simple async endpoint (delay)");
console.log("  GET /test-go-api    - Single Go API call");
console.log("  GET /test-go-batch  - Batch Go API calls");
console.log("  GET /health         - Health check");
console.log("");

srv.listen(":8080"); 