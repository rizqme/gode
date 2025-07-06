// 🧪 Gode HttpServer Example
// This demonstrates the Express-like API running on Goja + Gin

const srv = new HttpServer();

// Global middleware for logging (disabled for benchmarking)
// srv.use((req, res, next) => {
//   console.log(`[${new Date().toISOString()}] ${req.method} ${req.path}`);
//   next();
// });

// JSON endpoint for benchmarking
srv.handle("GET", "/ping", (req, res) => {
  res.writeHeader(200);
  res.header("Content-Type", "application/json");
  res.write(JSON.stringify({ pong: true, timestamp: Date.now() }));
  res.end();
});

// Streaming endpoint for testing chunked responses
srv.handle("GET", "/stream", async (req, res) => {
  res.writeHeader(200);
  res.header("Content-Type", "text/plain");
  res.header("Transfer-Encoding", "chunked");

  for (let i = 1; i <= 5; i++) {
    res.write(`Chunk ${i}\n`);
    res.flush();
    await delay(500);
  }

  res.end();
});

// JSON POST endpoint
srv.handle("POST", "/echo", (req, res) => {
  const body = req.json();
  res.writeHeader(200);
  res.header("Content-Type", "application/json");
  res.write(JSON.stringify({
    received: body,
    method: req.method,
    path: req.path,
    timestamp: Date.now()
  }));
  res.end();
});

// Health check endpoint
srv.handle("GET", "/health", (req, res) => {
  res.writeHeader(200);
  res.header("Content-Type", "application/json");
  res.write(JSON.stringify({
    status: "healthy",
    runtime: "goja",
    server: "gin",
    timestamp: Date.now()
  }));
  res.end();
});

// Start server
console.log("🚀 Starting Gode HttpServer...");
srv.listen(":8080"); 