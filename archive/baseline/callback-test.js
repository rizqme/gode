// Callback-based test for Node.js/Express - matches Go approach
const express = require('express');
const app = express();
app.use(express.json());

// Simulate external API delay with callback
function simulateApiCall(service, delay, callback) {
  setTimeout(() => {
    // Simulate some CPU work (like processing API response)
    const data = Math.random().toString(36).substring(7);
    
    const response = {
      service: service,
      timestamp: Date.now(),
      data: data,
      latency: delay,
      source: "node-settimeout"
    };
    
    // Simulate occasional failures (5% chance)
    if (Math.random() < 0.05) {
      callback(new Error(`API call failed for service: ${service}`), null);
    } else {
      callback(null, response);
    }
  }, delay);
}

// Batch API call with callbacks
function batchApiCall(services, baseDelay, callback) {
  const results = new Array(services.length);
  let completed = 0;
  let errorCount = 0;
  
  services.forEach((service, index) => {
    const delay = baseDelay + Math.floor(Math.random() * 50);
    
    simulateApiCall(service, delay, (err, result) => {
      if (err) {
        errorCount++;
        results[index] = {
          service: service,
          error: "API timeout",
          timestamp: Date.now()
        };
      } else {
        results[index] = {
          service: result.service,
          timestamp: result.timestamp,
          data: result.data,
          latency: result.latency,
          source: "node-batch"
        };
      }
      
      completed++;
      if (completed === services.length) {
        if (errorCount > services.length / 2) {
          callback(new Error(`Too many API failures: ${errorCount}/${services.length}`), null);
        } else {
          callback(null, results);
        }
      }
    });
  });
}

// Simple sync endpoint
app.get('/test-sync', (req, res) => {
  console.log("Sync handler called");
  res.json({ message: "sync works", timestamp: Date.now() });
});

// Test single API call with callback
app.get('/test-go-api', (req, res) => {
  console.log("API handler called");
  
  simulateApiCall("test-service", 100, (err, result) => {
    console.log("Callback called with:", err, result);
    
    if (err) {
      res.status(500).json({ error: err.message });
    } else {
      res.json({ 
        message: "api works", 
        result: result, 
        timestamp: Date.now() 
      });
    }
  });
});

// Test batch API call with callback
app.get('/test-go-batch', (req, res) => {
  console.log("Batch API handler called");
  
  batchApiCall(["service1", "service2", "service3"], 50, (err, results) => {
    console.log("Batch callback called with:", err, results);
    
    if (err) {
      res.status(500).json({ error: err.message });
    } else {
      res.json({ 
        message: "batch works", 
        results: results, 
        timestamp: Date.now() 
      });
    }
  });
});

// Complex endpoint with multiple callback-based API calls
app.get('/test-user-profile/:userId', (req, res) => {
  const userId = req.params.userId;
  console.log("User profile handler called for:", userId);
  
  const services = [
    "user-service",
    "auth-service", 
    "profile-service",
    "permissions-service",
    "billing-service"
  ];

  // Use callback-based batch API call
  batchApiCall(services, 60, (err, results) => {
    if (err) {
      res.status(500).json({ 
        error: "Service overload", 
        message: err.message,
        source: "node-callback-error"
      });
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
        node_source: results.every(r => r.source === "node-batch"),
        status: "success"
      };

      res.json(userProfile);
    }
  });
});

// Health check
app.get('/health', (req, res) => {
  res.json({ 
    status: "healthy", 
    timestamp: Date.now(),
    api_backend: "node-callbacks"
  });
});

const port = process.env.PORT || 8080;
app.listen(port, () => {
  console.log("🧪 Callback Test Server (Node.js)");
  console.log("External API calls using callbacks with setTimeout");
  console.log("Endpoints:");
  console.log("  GET /test-sync             - Simple sync endpoint");
  console.log("  GET /test-go-api           - Single API call (callback)");
  console.log("  GET /test-go-batch         - Batch API calls (callback)");
  console.log("  GET /test-user-profile/:id - User profile with 5 API calls");
  console.log("  GET /health                - Health check");
  console.log("");
  console.log(`🚀 Express server listening on :${port}`);
}); 