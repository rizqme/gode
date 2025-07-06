// 🚨 Event Loop Overload Example - Node.js/Express Version
// This will demonstrate event loop saturation under load

const express = require('express');
const app = express();
app.use(express.json());

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
app.get('/api/user-profile/:userId', async (req, res) => {
  const userId = req.params.userId;
  
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

    res.json(userProfile);
    
  } catch (error) {
    res.status(500).json({ error: "Service overload", message: error.message });
  }
});

// Heavy data processing endpoint
app.post('/api/process-data', async (req, res) => {
  const data = req.body;
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

    res.json({
      processed_count: processed.length,
      total_time: Date.now() - (data.start_time || Date.now()),
      status: "completed"
    });
    
  } catch (error) {
    res.status(500).json({ error: "Processing failed", message: error.message });
  }
});

// WebSocket-style simulation with many concurrent connections
app.get('/api/realtime-data', async (req, res) => {
  res.writeHead(200, {
    'Content-Type': 'application/json',
    'Transfer-Encoding': 'chunked'
  });
  
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
    }
    
    res.end();
    
  } catch (error) {
    res.status(500).json({ error: "Realtime stream failed" });
  }
});

// Simple health check (should stay fast even under load)
app.get('/health', (req, res) => {
  res.json({
    status: "healthy",
    timestamp: Date.now(),
    runtime: "node",
    server: "express"
  });
});

const port = process.env.PORT || 8080;
app.listen(port, () => {
  console.log("🚨 Starting Event Loop Overload Test Server (Node.js)...");
  console.log("Endpoints:");
  console.log("  GET  /api/user-profile/:userId  - 15 external API calls per request");
  console.log("  POST /api/process-data          - Batch processing with validations");
  console.log("  GET  /api/realtime-data         - 250 concurrent API calls");
  console.log("  GET  /health                    - Simple health check");
  console.log("");
  console.log(`🚀 Express server listening on :${port}`);
}); 