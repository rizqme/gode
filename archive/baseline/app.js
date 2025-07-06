// 🧪 Node.js Express Baseline
// This mirrors the Gode HttpServer functionality for benchmarking

const express = require("express");
const app = express();

// Parse JSON bodies
app.use(express.json());

// Global middleware for logging (disabled for benchmarking)
// app.use((req, res, next) => {
//   console.log(`[${new Date().toISOString()}] ${req.method} ${req.path}`);
//   next();
// });

// JSON endpoint for benchmarking
app.get("/ping", (req, res) => {
  res.json({ pong: true, timestamp: Date.now() });
});

// Streaming endpoint for testing chunked responses
app.get("/stream", async (req, res) => {
  res.setHeader("Content-Type", "text/plain");
  res.setHeader("Transfer-Encoding", "chunked");

  for (let i = 1; i <= 5; i++) {
    res.write(`Chunk ${i}\n`);
    await new Promise(resolve => setTimeout(resolve, 500));
  }

  res.end();
});

// JSON POST endpoint
app.post("/echo", (req, res) => {
  res.json({
    received: req.body,
    method: req.method,
    path: req.path,
    timestamp: Date.now()
  });
});

// Health check endpoint
app.get("/health", (req, res) => {
  res.json({
    status: "healthy",
    runtime: "nodejs",
    server: "express",
    timestamp: Date.now()
  });
});

// Start server
const PORT = process.env.PORT || 8080;
app.listen(PORT, () => {
  console.log(`🚀 Express server listening on :${PORT}`);
}); 