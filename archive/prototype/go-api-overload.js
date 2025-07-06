// 🚨 Event Loop Overload Example - Go-Based External API Calls
// This uses Go goroutines and HTTP client instead of JavaScript setTimeout

const srv = new HttpServer();

// Complex business logic endpoint that makes MANY external calls via Go
srv.handle("GET", "/api/user-profile/:userId", async (req, res) => {
  const userId = req.path.split('/').pop();
  
  try {
    // Use Go-based external API calls (handled by goroutines!)
    const services = [
      "user-service",
      "auth-service", 
      "profile-service",
      "permissions-service",
      "billing-service",
      "analytics-service",
      "notification-service",
      "recommendation-service",
      "social-service",
      "content-service",
      "search-service",
      "logging-service",
      "metrics-service",
      "cache-service",
      "backup-service"
    ];

    // This calls Go code that spawns 15 goroutines concurrently!
    const results = await batchApiCall(services, 60);
    
    // Process the results (CPU work in JavaScript)
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
    res.end();
    
  } catch (error) {
    res.writeHeader(500);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({ 
      error: "Service overload", 
      message: error.message,
      source: "go-goroutine-error"
    }));
    res.end();
  }
});

// Heavy data processing endpoint with Go-based validations
srv.handle("POST", "/api/process-data", async (req, res) => {
  const data = req.json();
  const batchSize = data.batch_size || 1000;
  
  try {
    // Create service list for validation
    const validatorServices = [];
    for (let i = 0; i < batchSize; i++) {
      validatorServices.push(`validator-${i % 10}`);
    }
    
    console.log(`Processing ${validatorServices.length} validation calls via Go...`);
    
    // This creates Go goroutines instead of JavaScript promises!
    const validations = await batchApiCall(validatorServices, 30);
    
    // CPU-intensive processing in JavaScript
    const processed = validations.map((v, index) => ({
      id: index,
      validated: !v.error,
      checksum: v.data ? v.data.split('').reduce((a, b) => a + b.charCodeAt(0), 0) : 0,
      timestamp: v.timestamp,
      source: v.source
    }));

    res.writeHeader(200);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({
      processed_count: processed.length,
      successful_validations: processed.filter(p => p.validated).length,
      go_powered: true,
      total_time: Date.now() - (data.start_time || Date.now()),
      status: "completed"
    }));
    res.end();
    
  } catch (error) {
    res.writeHeader(500);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({ 
      error: "Processing failed", 
      message: error.message,
      source: "go-batch-error"
    }));
    res.end();
  }
});

// Realtime data simulation using Go-based concurrent data sources
srv.handle("GET", "/api/realtime-data", async (req, res) => {
  res.writeHeader(200);
  res.header("Content-Type", "application/json");
  res.header("Transfer-Encoding", "chunked");
  
  // Create 50 concurrent data sources
  const dataSources = Array.from({length: 50}, (_, i) => `source-${i}`);
  
  try {
    for (let round = 0; round < 5; round++) {
      // Each round uses Go goroutines for all 50 sources
      const sourceData = await batchApiCall(dataSources, 40);
      
      const roundData = {
        round: round + 1,
        timestamp: Date.now(),
        sources: sourceData.length,
        successful_sources: sourceData.filter(d => !d.error).length,
        go_powered: true,
        data: sourceData.map(d => ({ 
          source: d.service, 
          value: d.data, 
          success: !d.error,
          latency: d.latency
        }))
      };
      
      res.write(JSON.stringify(roundData) + '\n');
      res.flush();
    }
    
    res.end();
    
  } catch (error) {
    res.writeHeader(500);
    res.write(JSON.stringify({ 
      error: "Realtime stream failed",
      source: "go-stream-error"
    }));
    res.end();
  }
});

// CPU-intensive endpoint that mixes Go API calls with JavaScript computation
srv.handle("POST", "/api/heavy-computation", async (req, res) => {
  const data = req.json();
  const iterations = data.iterations || 100;
  
  try {
    const results = [];
    
    // Mix Go API calls with JavaScript CPU work
    for (let i = 0; i < iterations; i++) {
      // Go-based API call
      const apiResult = await simulateApiCall(`compute-service-${i}`, 20 + (i % 50));
      
      // JavaScript CPU work
      let fibonacci = 0;
      let a = 0, b = 1;
      for (let j = 0; j < 1000; j++) {
        fibonacci = a + b;
        a = b;
        b = fibonacci;
      }
      
      results.push({
        iteration: i,
        api_data: apiResult.data,
        fibonacci_result: fibonacci,
        timestamp: apiResult.timestamp,
        source: apiResult.source
      });
    }

    res.writeHeader(200);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({
      total_iterations: iterations,
      results_count: results.length,
      go_api_calls: results.length,
      js_computations: results.length,
      mixed_workload: true,
      status: "completed"
    }));
    res.end();
    
  } catch (error) {
    res.writeHeader(500);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({ 
      error: "Heavy computation failed",
      source: "mixed-workload-error"
    }));
    res.end();
  }
});

// Simple health check (should stay fast even under load)
srv.handle("GET", "/health", (req, res) => {
  res.writeHeader(200);
  res.header("Content-Type", "application/json");
  res.write(JSON.stringify({
    status: "healthy",
    timestamp: Date.now(),
    runtime: "goja",
    server: "gin",
    api_backend: "go-goroutines"
  }));
  res.end();
});

console.log("🚨 Starting Go-Powered Event Loop Test Server...");
console.log("External API calls handled by Go goroutines, not JavaScript setTimeout!");
console.log("Endpoints:");
console.log("  GET  /api/user-profile/:userId    - 15 Go goroutine API calls");
console.log("  POST /api/process-data            - Batch processing via Go");
console.log("  GET  /api/realtime-data           - 250 concurrent Go API calls");
console.log("  POST /api/heavy-computation       - Mixed Go/JS workload");
console.log("  GET  /health                      - Simple health check");
console.log("");

srv.listen(":8080"); 