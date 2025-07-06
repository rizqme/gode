// Callback-based test for Go-powered external API calls
const srv = new HttpServer();

// Simple sync endpoint
srv.handle("GET", "/test-sync", (req, res) => {
  console.log("Sync handler called");
  res.writeHeader(200);
  res.header("Content-Type", "application/json");
  res.write(JSON.stringify({ message: "sync works", timestamp: Date.now() }));
  res.end();
});

// Test single Go API call with callback
srv.handle("GET", "/test-go-api", (req, res) => {
  console.log("Go API handler called");
  
  simulateApiCall("test-service", 100, function(err, result) {
    console.log("Callback called with:", err, result);
    
    if (err) {
      res.writeHeader(500);
      res.header("Content-Type", "application/json");
      res.write(JSON.stringify({ error: err.message }));
    } else {
      res.writeHeader(200);
      res.header("Content-Type", "application/json");
      res.write(JSON.stringify({ 
        message: "go api works", 
        result: result, 
        timestamp: Date.now() 
      }));
    }
    res.end();
  });
});

// Test batch Go API call with callback
srv.handle("GET", "/test-go-batch", (req, res) => {
  console.log("Go batch API handler called");
  
  batchApiCall(["service1", "service2", "service3"], 50, function(err, results) {
    console.log("Batch callback called with:", err, results);
    
    if (err) {
      res.writeHeader(500);
      res.header("Content-Type", "application/json");
      res.write(JSON.stringify({ error: err.message }));
    } else {
      res.writeHeader(200);
      res.header("Content-Type", "application/json");
      res.write(JSON.stringify({ 
        message: "go batch works", 
        results: results, 
        timestamp: Date.now() 
      }));
    }
    res.end();
  });
});

// Complex endpoint with multiple callback-based API calls
srv.handle("GET", "/test-user-profile/:userId", (req, res) => {
  const userId = req.path.split('/').pop();
  console.log("User profile handler called for:", userId);
  
  const services = [
    "user-service",
    "auth-service", 
    "profile-service",
    "permissions-service",
    "billing-service"
  ];

  // Use Go-based batch API call with callback
  batchApiCall(services, 60, function(err, results) {
    if (err) {
      res.writeHeader(500);
      res.header("Content-Type", "application/json");
      res.write(JSON.stringify({ 
        error: "Service overload", 
        message: err.message,
        source: "go-goroutine-error"
      }));
    } else {
      // Process the results
      const userProfile = {
        id: userId,
        timestamp: Date.now(),
        services: results.length,
        aggregated_data: results.reduce((acc, result) => {
          acc[result.service] = result.data;
          return acc;
        }, {}),
        computed_score: results.reduce((sum, r) => sum + (r.data ? r.data.length : 0), 0),
        go_source: results.every(r => r.source === "go-batch"),
        status: "success"
      };

      res.writeHeader(200);
      res.header("Content-Type", "application/json");
      res.write(JSON.stringify(userProfile));
    }
    res.end();
  });
});

// Health check
srv.handle("GET", "/health", (req, res) => {
  res.writeHeader(200);
  res.header("Content-Type", "application/json");
  res.write(JSON.stringify({ 
    status: "healthy", 
    timestamp: Date.now(),
    api_backend: "go-callbacks"
  }));
  res.end();
});

console.log("🧪 Callback Test Server");
console.log("Go external API calls using callbacks instead of Promises");
console.log("Endpoints:");
console.log("  GET /test-sync             - Simple sync endpoint");
console.log("  GET /test-go-api           - Single Go API call (callback)");
console.log("  GET /test-go-batch         - Batch Go API calls (callback)");
console.log("  GET /test-user-profile/:id - User profile with 5 Go API calls");
console.log("  GET /health                - Health check");
console.log("");

srv.listen(":8080"); 