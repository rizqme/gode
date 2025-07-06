// 🚨 Event Loop Overload Example
// This simulates multiple external API calls that saturate Node.js event loop

const srv = new HttpServer();

// Simulate external API delay
function simulateApiCall(service, delay = 100) {
  return new Promise(resolve => {
    setTimeout(() => {
      resolve({
        service: service,
        timestamp: Date.now(),
        data: Math.random().toString(36).substring(7)
      });
    }, delay);
  });
}

// Complex business logic endpoint that makes MANY external calls
srv.handle("GET", "/api/user-profile/:userId", async (req, res) => {
  const userId = req.path.split('/').pop();
  
  try {
    // Simulate 15+ external API calls per request (event loop killer!)
    const apiCalls = [
      simulateApiCall("user-service", 50),
      simulateApiCall("auth-service", 75),
      simulateApiCall("profile-service", 60),
      simulateApiCall("permissions-service", 90),
      simulateApiCall("billing-service", 45),
      simulateApiCall("analytics-service", 80),
      simulateApiCall("notification-service", 55),
      simulateApiCall("recommendation-service", 120),
      simulateApiCall("social-service", 70),
      simulateApiCall("content-service", 65),
      simulateApiCall("search-service", 85),
      simulateApiCall("logging-service", 40),
      simulateApiCall("metrics-service", 95),
      simulateApiCall("cache-service", 35),
      simulateApiCall("backup-service", 110)
    ];

    // All APIs must complete before response
    const results = await Promise.all(apiCalls);
    
    // Process the results (more CPU work)
    const userProfile = {
      id: userId,
      timestamp: Date.now(),
      services: results.length,
      aggregated_data: results.reduce((acc, result) => {
        acc[result.service] = result.data;
        return acc;
      }, {}),
      computed_score: results.reduce((sum, r) => sum + r.data.length, 0),
      status: "success"
    };

    res.writeHeader(200);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify(userProfile));
    res.end();
    
  } catch (error) {
    res.writeHeader(500);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({ error: "Service overload", message: error.message }));
    res.end();
  }
});

// Heavy data processing endpoint
srv.handle("POST", "/api/process-data", async (req, res) => {
  const data = req.json();
  const batchSize = data.batch_size || 1000;
  
  try {
    // Simulate processing large dataset with external validations
    const validationCalls = [];
    for (let i = 0; i < batchSize; i++) {
      // Each record needs validation from external service
      validationCalls.push(simulateApiCall(`validator-${i % 10}`, 20 + Math.random() * 50));
    }
    
    console.log(`Processing ${validationCalls.length} validation calls...`);
    
    // This creates THOUSANDS of pending promises (event loop nightmare!)
    const validations = await Promise.all(validationCalls);
    
    // CPU-intensive processing
    const processed = validations.map((v, index) => ({
      id: index,
      validated: true,
      checksum: v.data.split('').reduce((a, b) => a + b.charCodeAt(0), 0),
      timestamp: v.timestamp
    }));

    res.writeHeader(200);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({
      processed_count: processed.length,
      total_time: Date.now() - (data.start_time || Date.now()),
      status: "completed"
    }));
    res.end();
    
  } catch (error) {
    res.writeHeader(500);
    res.header("Content-Type", "application/json");
    res.write(JSON.stringify({ error: "Processing failed", message: error.message }));
    res.end();
  }
});

// WebSocket-style simulation with many concurrent connections
srv.handle("GET", "/api/realtime-data", async (req, res) => {
  res.writeHeader(200);
  res.header("Content-Type", "application/json");
  res.header("Transfer-Encoding", "chunked");
  
  // Simulate 50 concurrent data sources
  const dataSources = Array.from({length: 50}, (_, i) => `source-${i}`);
  
  try {
    for (let round = 0; round < 5; round++) {
      // Each round creates 50 pending API calls
      const sourceData = await Promise.all(
        dataSources.map(source => simulateApiCall(source, 30 + Math.random() * 70))
      );
      
      const roundData = {
        round: round + 1,
        timestamp: Date.now(),
        sources: sourceData.length,
        data: sourceData.map(d => ({ source: d.service, value: d.data }))
      };
      
      res.write(JSON.stringify(roundData) + '\n');
      res.flush();
    }
    
    res.end();
    
  } catch (error) {
    res.writeHeader(500);
    res.write(JSON.stringify({ error: "Realtime stream failed" }));
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
    server: "gin"
  }));
  res.end();
});

console.log("🚨 Starting Event Loop Overload Test Server...");
console.log("Endpoints:");
console.log("  GET  /api/user-profile/:userId  - 15 external API calls per request");
console.log("  POST /api/process-data          - Batch processing with validations");
console.log("  GET  /api/realtime-data         - 250 concurrent API calls");
console.log("  GET  /health                    - Simple health check");
console.log("");

srv.listen(":8080"); 