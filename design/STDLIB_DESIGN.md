# Gode Standard Library Design

## Core Philosophy

Unlike Deno's approach with a global `Deno` namespace, Gode uses proper module imports for all standard library functionality. This provides better code organization, tree-shaking opportunities, and aligns with modern JavaScript module practices.

## Module Import System

All standard library modules use the `gode:` prefix to distinguish them from user modules:

```javascript
// ES module imports (preferred)
import { serve } from 'gode:http';
import { readFile } from 'gode:fs';

// CommonJS support
const { serve } = require('gode:http');
const { readFile } = require('gode:fs');
```

## Standard Library Overview

### Core Modules
- **`gode:http`** - HTTP server and client functionality
- **`gode:net`** - TCP and UDP networking
- **`gode:ws`** - WebSocket server and client
- **`gode:fs`** - File system operations
- **`gode:stream`** - Stream primitives and utilities
- **`gode:path`** - Path manipulation utilities
- **`gode:os`** - Operating system information
- **`gode:process`** - Process control and environment
- **`gode:crypto`** - Cryptographic operations
- **`gode:permissions`** - Permission management
- **`gode:console`** - Enhanced console logging
- **`gode:test`** - Built-in testing framework
- **`gode:encoding`** - Text encoding/decoding utilities

### Future Modules
- **`gode:worker`** - Worker threads
- **`gode:sqlite`** - Built-in SQLite database
- **`gode:compress`** - Compression utilities
- **`gode:email`** - Email sending
- **`gode:ssh`** - SSH client/server
- **`gode:grpc`** - gRPC support

## Module Definitions

### HTTP Module (`gode:http`)

High-level HTTP server with routing and middleware support, inspired by Gin.

#### Interfaces
```typescript
interface Context {
  request: Request;
  response: ResponseWriter;
  params: { [key: string]: string };
  query: URLSearchParams;
  get(key: string): string | undefined;
  set(key: string, value: any): void;
  json(data: any, status?: number): void;
  text(data: string, status?: number): void;
  html(data: string, status?: number): void;
  redirect(url: string, status?: number): void;
  abort(status: number, message?: string): void;
  next(): void;
}

interface ResponseWriter {
  status(code: number): ResponseWriter;
  header(key: string, value: string): ResponseWriter;
  write(data: string | Buffer | Uint8Array): void;
  end(): void;
}

type HandlerFunc = (ctx: Context) => void | Promise<void>;
type MiddlewareFunc = (ctx: Context) => void | Promise<void>;

interface RouterGroup {
  use(...middleware: MiddlewareFunc[]): RouterGroup;
  get(path: string, ...handlers: HandlerFunc[]): RouterGroup;
  post(path: string, ...handlers: HandlerFunc[]): RouterGroup;
  put(path: string, ...handlers: HandlerFunc[]): RouterGroup;
  delete(path: string, ...handlers: HandlerFunc[]): RouterGroup;
  patch(path: string, ...handlers: HandlerFunc[]): RouterGroup;
  head(path: string, ...handlers: HandlerFunc[]): RouterGroup;
  options(path: string, ...handlers: HandlerFunc[]): RouterGroup;
  any(path: string, ...handlers: HandlerFunc[]): RouterGroup;
  group(prefix: string): RouterGroup;
  static(prefix: string, root: string): RouterGroup;
}

interface Server extends RouterGroup {
  listen(port: number, hostname?: string): Promise<void>;
  close(): Promise<void>;
  ref(): void;
  unref(): void;
}

interface CreateServerOptions {
  http2?: boolean;
  http3?: boolean;
  cert?: string;
  key?: string;
  maxBodySize?: number;
  trustProxy?: boolean;
}

function createServer(options?: CreateServerOptions): Server;

// Middleware helpers
const middleware = {
  cors(options?: CorsOptions): MiddlewareFunc;
  compress(options?: CompressOptions): MiddlewareFunc;
  logger(format?: string): MiddlewareFunc;
  recovery(): MiddlewareFunc;
  timeout(duration: number): MiddlewareFunc;
  rateLimit(options: RateLimitOptions): MiddlewareFunc;
};
```

#### Example
```javascript
import { createServer, middleware } from 'gode:http';

// Create server with routing
const app = createServer({
  http2: true,
  trustProxy: true
});

// Global middleware
app.use(middleware.logger());
app.use(middleware.cors({ origin: '*' }));
app.use(middleware.recovery());

// Custom middleware
app.use(async (ctx) => {
  const start = Date.now();
  ctx.set('requestId', crypto.randomUUID());
  
  await ctx.next();
  
  const ms = Date.now() - start;
  ctx.response.header('X-Response-Time', `${ms}ms`);
});

// Routes
app.get('/health', (ctx) => {
  ctx.json({ status: 'ok' });
});

// Route with URL parameters
app.get('/users/:id', async (ctx) => {
  const userId = ctx.params.id;
  const user = await db.getUser(userId);
  
  if (!user) {
    ctx.abort(404, 'User not found');
    return;
  }
  
  ctx.json(user);
});

// Route group with prefix
const api = app.group('/api/v1');

api.use(middleware.rateLimit({ 
  windowMs: 15 * 60 * 1000,
  max: 100 
}));

api.get('/posts', async (ctx) => {
  const page = ctx.query.get('page') || '1';
  const posts = await db.getPosts(parseInt(page));
  ctx.json(posts);
});

api.post('/posts', async (ctx) => {
  const body = await ctx.request.json();
  const post = await db.createPost(body);
  ctx.json(post, 201);
});

// Multiple handlers
api.put('/posts/:id', 
  requireAuth,
  validatePost,
  async (ctx) => {
    const postId = ctx.params.id;
    const body = await ctx.request.json();
    const updated = await db.updatePost(postId, body);
    ctx.json(updated);
  }
);

// Static files
app.static('/public', './static');

// Start server
await app.listen(3000, 'localhost');
console.log('Server running on http://localhost:3000');

// Middleware functions
async function requireAuth(ctx) {
  const token = ctx.get('Authorization');
  if (!token) {
    ctx.abort(401, 'Authorization required');
    return;
  }
  
  const user = await validateToken(token);
  if (!user) {
    ctx.abort(401, 'Invalid token');
    return;
  }
  
  ctx.set('user', user);
  await ctx.next();
}

function validatePost(ctx) {
  // Validation logic
  ctx.next();
}
```

### Network Module (`gode:net`)

Low-level TCP and UDP networking operations.

#### Interfaces
```typescript
// TCP
interface NetConnectOptions {
  port: number;
  host?: string;
  localAddress?: string;
  localPort?: number;
  family?: 4 | 6;
  timeout?: number;
}

interface Socket extends stream.Duplex {
  connect(options: NetConnectOptions): void;
  write(data: string | Buffer): boolean;
  end(data?: string | Buffer): void;
  destroy(error?: Error): void;
  setTimeout(timeout: number): void;
  setKeepAlive(enable: boolean, delay?: number): void;
  address(): { address: string; family: string; port: number };
}

interface Server {
  listen(port: number, hostname?: string, backlog?: number): void;
  close(callback?: () => void): void;
  address(): { address: string; family: string; port: number };
  on(event: 'connection', listener: (socket: Socket) => void): void;
  on(event: 'error', listener: (error: Error) => void): void;
}

function createServer(connectionListener?: (socket: Socket) => void): Server;
function createConnection(options: NetConnectOptions): Socket;
function connect(options: NetConnectOptions): Socket;

// UDP
interface SocketOptions {
  type: 'udp4' | 'udp6';
  reuseAddr?: boolean;
}

interface UDPSocket extends EventEmitter {
  bind(port?: number, address?: string): void;
  send(msg: Buffer | string, port: number, address: string): void;
  close(): void;
  address(): { address: string; family: string; port: number };
  on(event: 'message', listener: (msg: Buffer, rinfo: RemoteInfo) => void): void;
  on(event: 'error', listener: (error: Error) => void): void;
}

function createSocket(options: SocketOptions): UDPSocket;
```

#### Example
```javascript
import { createServer, createConnection, createSocket } from 'gode:net';

// TCP Server
const tcpServer = createServer((socket) => {
  console.log('Client connected');
  
  socket.on('data', (data) => {
    socket.write(`Echo: ${data}`);
  });
  
  socket.on('end', () => {
    console.log('Client disconnected');
  });
});

tcpServer.listen(8080, 'localhost');

// TCP Client
const client = createConnection({ port: 8080, host: 'localhost' });
client.on('connect', () => {
  client.write('Hello Server');
});

// UDP Socket
const udpSocket = createSocket({ type: 'udp4' });
udpSocket.bind(41234);

udpSocket.on('message', (msg, rinfo) => {
  console.log(`UDP message from ${rinfo.address}:${rinfo.port}: ${msg}`);
});
```

### WebSocket Module (`gode:ws`)

WebSocket server and client implementation.

#### Interfaces
```typescript
interface WebSocketServerOptions {
  port: number;
  host?: string;
  backlog?: number;
  server?: Server;
  verifyClient?: (info: { origin: string; secure: boolean; req: Request }) => boolean;
  perMessageDeflate?: boolean | object;
  maxPayload?: number;
}

class WebSocketServer extends EventEmitter {
  constructor(options: WebSocketServerOptions);
  on(event: 'connection', listener: (ws: WebSocket, request: Request) => void): this;
  on(event: 'error', listener: (error: Error) => void): this;
  close(callback?: () => void): void;
}

class WebSocket extends EventEmitter {
  constructor(url: string, protocols?: string | string[]);
  send(data: string | Buffer | ArrayBuffer): void;
  close(code?: number, reason?: string): void;
  ping(data?: any): void;
  pong(data?: any): void;
  on(event: 'open', listener: () => void): this;
  on(event: 'message', listener: (data: string | Buffer) => void): this;
  on(event: 'close', listener: (code: number, reason: string) => void): this;
  on(event: 'error', listener: (error: Error) => void): this;
  readyState: number;
  static CONNECTING: 0;
  static OPEN: 1;
  static CLOSING: 2;
  static CLOSED: 3;
}
```

#### Example
```javascript
import { WebSocketServer, WebSocket } from 'gode:ws';

// WebSocket Server
const wss = new WebSocketServer({
  port: 8080,
  perMessageDeflate: {
    zlibDeflateOptions: {
      level: 9
    }
  }
});

wss.on('connection', (ws, request) => {
  console.log(`New connection from ${request.headers.get('origin')}`);
  
  ws.send('Welcome to the WebSocket server!');
  
  ws.on('message', (data) => {
    // Broadcast to all clients
    wss.clients.forEach((client) => {
      if (client.readyState === WebSocket.OPEN) {
        client.send(`Broadcast: ${data}`);
      }
    });
  });
  
  ws.on('close', (code, reason) => {
    console.log(`Connection closed: ${code} - ${reason}`);
  });
});

// WebSocket Client
const ws = new WebSocket('ws://localhost:8080');

ws.on('open', () => {
  console.log('Connected to server');
  ws.send('Hello from client');
});

ws.on('message', (data) => {
  console.log(`Received: ${data}`);
});
```

### File System Module (`gode:fs`)

File system operations with Promise-based API.

#### Interfaces
```typescript
interface FileHandle {
  fd: number;
  close(): Promise<void>;
  read(buffer: Buffer, offset: number, length: number, position: number): Promise<{ bytesRead: number; buffer: Buffer }>;
  write(buffer: Buffer, offset?: number, length?: number, position?: number): Promise<{ bytesWritten: number; buffer: Buffer }>;
}

interface Stats {
  isFile(): boolean;
  isDirectory(): boolean;
  isSymbolicLink(): boolean;
  size: number;
  mtime: Date;
  atime: Date;
  ctime: Date;
  mode: number;
}

interface Dirent {
  name: string;
  isFile(): boolean;
  isDirectory(): boolean;
  isSymbolicLink(): boolean;
}

// Main functions
function readFile(path: string, encoding?: string): Promise<string | Buffer>;
function writeFile(path: string, data: string | Buffer, options?: { encoding?: string; mode?: number }): Promise<void>;
function appendFile(path: string, data: string | Buffer, options?: { encoding?: string; mode?: number }): Promise<void>;
function readdir(path: string, options?: { withFileTypes?: boolean }): Promise<string[] | Dirent[]>;
function mkdir(path: string, options?: { recursive?: boolean; mode?: number }): Promise<void>;
function rmdir(path: string, options?: { recursive?: boolean }): Promise<void>;
function unlink(path: string): Promise<void>;
function rename(oldPath: string, newPath: string): Promise<void>;
function stat(path: string): Promise<Stats>;
function lstat(path: string): Promise<Stats>;
function chmod(path: string, mode: number): Promise<void>;
function chown(path: string, uid: number, gid: number): Promise<void>;
function open(path: string, flags: string, mode?: number): Promise<FileHandle>;

// From gode:fs/streams
function createReadStream(path: string, options?: { encoding?: string; start?: number; end?: number }): ReadStream;
function createWriteStream(path: string, options?: { encoding?: string; flags?: string; mode?: number }): WriteStream;
```

#### Example
```javascript
import { readFile, writeFile, readdir, mkdir, stat } from 'gode:fs';
import { createReadStream, createWriteStream } from 'gode:fs/streams';

// Read file
const content = await readFile('./config.json', 'utf8');
const config = JSON.parse(content);

// Write file
await writeFile('./output.txt', 'Hello World\n');

// Directory operations
await mkdir('./temp', { recursive: true });
const files = await readdir('./src', { withFileTypes: true });

for (const file of files) {
  if (file.isFile()) {
    const stats = await stat(`./src/${file.name}`);
    console.log(`${file.name}: ${stats.size} bytes`);
  }
}

// Streaming
const readStream = createReadStream('./large-file.dat');
const writeStream = createWriteStream('./copy.dat');

readStream.pipe(writeStream);

readStream.on('end', () => {
  console.log('File copied successfully');
});
```

### Stream Module (`gode:stream`)

Stream primitives and utilities for working with streaming data.

#### Interfaces
```typescript
interface Readable extends EventEmitter {
  read(size?: number): any;
  pipe<T extends Writable>(destination: T, options?: { end?: boolean }): T;
  unpipe(destination?: Writable): this;
  pause(): this;
  resume(): this;
  isPaused(): boolean;
  destroy(error?: Error): void;
  static from(iterable: Iterable<any> | AsyncIterable<any>): Readable;
}

interface Writable extends EventEmitter {
  write(chunk: any, encoding?: string, callback?: (error?: Error) => void): boolean;
  end(chunk?: any, encoding?: string, callback?: () => void): void;
  destroy(error?: Error): void;
  cork(): void;
  uncork(): void;
}

interface Duplex extends Readable, Writable {}

interface Transform extends Duplex {
  _transform(chunk: any, encoding: string, callback: (error?: Error, data?: any) => void): void;
  _flush(callback: (error?: Error, data?: any) => void): void;
}

class PassThrough extends Transform {}

function pipeline(...streams: Array<Readable | Writable | Duplex | Transform>): Promise<void>;
function finished(stream: Readable | Writable | Duplex, options?: { error?: boolean }): Promise<void>;
```

#### Example
```javascript
import { Readable, Writable, Transform, pipeline } from 'gode:stream';

// Create a readable stream
const readable = Readable.from(['Hello', ' ', 'World', '!']);

// Create a transform stream
class UpperCaseTransform extends Transform {
  _transform(chunk, encoding, callback) {
    this.push(chunk.toString().toUpperCase());
    callback();
  }
}

// Create a writable stream
const chunks = [];
const writable = new Writable({
  write(chunk, encoding, callback) {
    chunks.push(chunk);
    callback();
  }
});

// Pipeline streams together
await pipeline(
  readable,
  new UpperCaseTransform(),
  writable
);

console.log(chunks.join('')); // HELLO WORLD!
```

### Path Module (`gode:path`)

Path manipulation utilities.

#### Interfaces
```typescript
function basename(path: string, ext?: string): string;
function dirname(path: string): string;
function extname(path: string): string;
function format(pathObject: { dir?: string; root?: string; base?: string; name?: string; ext?: string }): string;
function isAbsolute(path: string): boolean;
function join(...paths: string[]): string;
function normalize(path: string): string;
function parse(path: string): { dir: string; root: string; base: string; name: string; ext: string };
function relative(from: string, to: string): string;
function resolve(...paths: string[]): string;

const sep: string;
const delimiter: string;
```

#### Example
```javascript
import { join, resolve, dirname, basename, extname } from 'gode:path';

const filePath = '/home/user/documents/file.txt';

console.log(dirname(filePath));   // /home/user/documents
console.log(basename(filePath));  // file.txt
console.log(extname(filePath));   // .txt

const fullPath = resolve('src', '..', 'lib', 'utils.js');
console.log(fullPath); // /current/working/directory/lib/utils.js

const joined = join('/users', 'john', 'documents', 'file.txt');
console.log(joined); // /users/john/documents/file.txt
```

### OS Module (`gode:os`)

Operating system information and utilities.

#### Interfaces
```typescript
interface CPU {
  model: string;
  speed: number;
  times: {
    user: number;
    nice: number;
    sys: number;
    idle: number;
    irq: number;
  };
}

interface NetworkInterface {
  address: string;
  netmask: string;
  family: 'IPv4' | 'IPv6';
  mac: string;
  internal: boolean;
  cidr: string;
}

function arch(): string;
function cpus(): CPU[];
function endianness(): 'BE' | 'LE';
function freemem(): number;
function homedir(): string;
function hostname(): string;
function loadavg(): [number, number, number];
function networkInterfaces(): { [name: string]: NetworkInterface[] };
function platform(): string;
function release(): string;
function tmpdir(): string;
function totalmem(): number;
function type(): string;
function uptime(): number;
function userInfo(): { username: string; uid: number; gid: number; shell: string; homedir: string };

const EOL: string;
```

#### Example
```javascript
import { platform, arch, cpus, networkInterfaces, homedir, totalmem, freemem } from 'gode:os';

console.log(`Platform: ${platform()}`);  // darwin, linux, win32
console.log(`Architecture: ${arch()}`);  // x64, arm64
console.log(`CPUs: ${cpus().length}`);
console.log(`Home directory: ${homedir()}`);

// Memory information
const totalMemory = totalmem();
const freeMemory = freemem();
console.log(`Memory: ${freeMemory / 1024 / 1024}MB free of ${totalMemory / 1024 / 1024}MB`);

// Network interfaces
const interfaces = networkInterfaces();
Object.entries(interfaces).forEach(([name, addresses]) => {
  addresses.forEach(addr => {
    if (addr.family === 'IPv4' && !addr.internal) {
      console.log(`${name}: ${addr.address}`);
    }
  });
});
```

### Process Module (`gode:process`)

Process control, environment variables, and runtime information.

#### Interfaces
```typescript
interface ProcessEnv {
  get(key: string): string | undefined;
  set(key: string, value: string): void;
  delete(key: string): void;
  has(key: string): boolean;
  toObject(): { [key: string]: string };
}

const argv: string[];
const env: ProcessEnv;
const pid: number;
const ppid: number;
const version: string;
const versions: { [key: string]: string };

function abort(): never;
function chdir(directory: string): void;
function cwd(): string;
function exit(code?: number): never;
function kill(pid: number, signal?: string | number): void;
function memoryUsage(): {
  rss: number;
  heapTotal: number;
  heapUsed: number;
  external: number;
  arrayBuffers: number;
};
function nextTick(callback: Function, ...args: any[]): void;
function uptime(): number;

// Events
on(event: 'exit', listener: (code: number) => void): void;
on(event: 'uncaughtException', listener: (error: Error) => void): void;
on(event: 'unhandledRejection', listener: (reason: any, promise: Promise<any>) => void): void;
```

#### Example
```javascript
import { env, exit, cwd, chdir, argv, memoryUsage, on } from 'gode:process';

// Environment variables
const apiKey = env.get('API_KEY');
if (!apiKey) {
  console.error('API_KEY environment variable is required');
  exit(1);
}

env.set('NODE_ENV', 'production');

// Process information
console.log(`Current directory: ${cwd()}`);
console.log(`Arguments: ${argv.join(' ')}`);

// Memory usage
const mem = memoryUsage();
console.log(`Memory usage: ${Math.round(mem.heapUsed / 1024 / 1024)}MB`);

// Process events
on('uncaughtException', (error) => {
  console.error('Uncaught exception:', error);
  exit(1);
});

on('exit', (code) => {
  console.log(`Process exiting with code: ${code}`);
});
```

### Crypto Module (`gode:crypto`)

Cryptographic operations including hashing, random data, and Web Crypto API.

#### Interfaces
```typescript
interface Hash {
  update(data: string | Buffer, encoding?: string): Hash;
  digest(): Buffer;
  digest(encoding: string): string;
}

interface Hmac {
  update(data: string | Buffer, encoding?: string): Hmac;
  digest(): Buffer;
  digest(encoding: string): string;
}

interface Cipher {
  update(data: string | Buffer, inputEncoding?: string, outputEncoding?: string): string | Buffer;
  final(outputEncoding?: string): string | Buffer;
}

interface Decipher {
  update(data: string | Buffer, inputEncoding?: string, outputEncoding?: string): string | Buffer;
  final(outputEncoding?: string): string | Buffer;
}

function createHash(algorithm: string): Hash;
function createHmac(algorithm: string, key: string | Buffer): Hmac;
function createCipher(algorithm: string, password: string | Buffer): Cipher;
function createDecipher(algorithm: string, password: string | Buffer): Decipher;
function randomBytes(size: number): Promise<Buffer>;
function randomInt(min: number, max: number): Promise<number>;
function randomUUID(): string;

// Web Crypto API subset
const subtle: {
  generateKey(algorithm: object, extractable: boolean, keyUsages: string[]): Promise<CryptoKey>;
  encrypt(algorithm: object, key: CryptoKey, data: BufferSource): Promise<ArrayBuffer>;
  decrypt(algorithm: object, key: CryptoKey, data: BufferSource): Promise<ArrayBuffer>;
  sign(algorithm: object, key: CryptoKey, data: BufferSource): Promise<ArrayBuffer>;
  verify(algorithm: object, key: CryptoKey, signature: BufferSource, data: BufferSource): Promise<boolean>;
};
```

#### Example
```javascript
import { createHash, randomBytes, randomUUID, subtle } from 'gode:crypto';

// Hashing
const hash = createHash('sha256');
hash.update('Hello World');
console.log(hash.digest('hex'));

// Random data
const randomData = await randomBytes(32);
console.log(`Random bytes: ${randomData.toString('hex')}`);

// UUID
const uuid = randomUUID();
console.log(`UUID: ${uuid}`);

// Web Crypto API
const key = await subtle.generateKey(
  {
    name: 'AES-GCM',
    length: 256
  },
  true,
  ['encrypt', 'decrypt']
);

const encoder = new TextEncoder();
const data = encoder.encode('Secret message');
const iv = await randomBytes(12);

const encrypted = await subtle.encrypt(
  {
    name: 'AES-GCM',
    iv: iv
  },
  key,
  data
);

console.log('Encrypted data:', new Uint8Array(encrypted));
```

### Permissions Module (`gode:permissions`)

Runtime permission management system.

#### Interfaces
```typescript
type PermissionName = 'read' | 'write' | 'net' | 'env' | 'run' | 'plugin' | 'hrtime';
type PermissionState = 'granted' | 'denied' | 'prompt';

interface PermissionDescriptor {
  name: PermissionName;
  path?: string;        // for read/write
  host?: string;        // for net
  port?: number;        // for net
  variable?: string;    // for env
  command?: string;     // for run
}

interface PermissionStatus {
  state: PermissionState;
  onchange: ((this: PermissionStatus) => void) | null;
}

interface PermissionRequestOptions extends PermissionDescriptor {
  reason?: string;      // Optional reason shown to user
}

function request(name: PermissionName, options?: PermissionRequestOptions): Promise<PermissionStatus>;
function query(descriptor: PermissionDescriptor): Promise<PermissionStatus>;
function revoke(descriptor: PermissionDescriptor): Promise<PermissionStatus>;
```

#### Example
```javascript
import { request, query, revoke } from 'gode:permissions';

// Request network permission
const netPermission = await request('net', {
  host: 'api.example.com',
  port: 443,
  reason: 'To fetch data from the API'
});

if (netPermission.state === 'granted') {
  const response = await fetch('https://api.example.com/data');
  // Process response
} else {
  console.error('Network permission denied');
}

// Check file read permission
const readPermission = await query({
  name: 'read',
  path: '/home/user/documents'
});

if (readPermission.state !== 'granted') {
  // Request permission if not granted
  const result = await request('read', {
    path: '/home/user/documents',
    reason: 'To read configuration files'
  });
}

// Revoke permission
await revoke({
  name: 'net',
  host: 'api.example.com'
});
```

### Console Module (`gode:console`)

Enhanced console with formatting and log levels.

#### Interfaces
```typescript
interface ConsoleOptions {
  colors?: boolean;
  timestamps?: boolean;
  logLevel?: 'debug' | 'info' | 'warn' | 'error';
  prefix?: string;
}

class Console {
  constructor(options?: ConsoleOptions);
  assert(condition: any, ...data: any[]): void;
  clear(): void;
  count(label?: string): void;
  countReset(label?: string): void;
  debug(...data: any[]): void;
  dir(item: any, options?: { showHidden?: boolean; depth?: number; colors?: boolean }): void;
  dirxml(...data: any[]): void;
  error(...data: any[]): void;
  group(...data: any[]): void;
  groupCollapsed(...data: any[]): void;
  groupEnd(): void;
  info(...data: any[]): void;
  log(...data: any[]): void;
  table(tabularData: any, properties?: string[]): void;
  time(label?: string): void;
  timeEnd(label?: string): void;
  timeLog(label?: string, ...data: any[]): void;
  trace(...data: any[]): void;
  warn(...data: any[]): void;
}
```

#### Example
```javascript
import { Console } from 'gode:console';

// Create custom console
const console = new Console({
  colors: true,
  timestamps: true,
  logLevel: 'debug',
  prefix: '[MyApp]'
});

// Various log levels
console.debug('Debug information');
console.info('Application started');
console.warn('Low memory warning');
console.error('Failed to connect to database');

// Table output
const users = [
  { id: 1, name: 'Alice', age: 30 },
  { id: 2, name: 'Bob', age: 25 },
  { id: 3, name: 'Charlie', age: 35 }
];
console.table(users);

// Timing
console.time('dataProcessing');
// ... do some work
console.timeEnd('dataProcessing');

// Grouping
console.group('User Details');
console.log('Name: Alice');
console.log('Age: 30');
console.groupEnd();
```

### Test Module (`gode:test`)

Built-in testing framework with familiar API.

#### Interfaces
```typescript
interface TestOptions {
  only?: boolean;
  skip?: boolean;
  timeout?: number;
}

interface ExpectMatchers<T> {
  toBe(expected: T): void;
  toEqual(expected: T): void;
  toBeNull(): void;
  toBeUndefined(): void;
  toBeDefined(): void;
  toBeTruthy(): void;
  toBeFalsy(): void;
  toBeGreaterThan(expected: number): void;
  toBeGreaterThanOrEqual(expected: number): void;
  toBeLessThan(expected: number): void;
  toBeLessThanOrEqual(expected: number): void;
  toContain(expected: any): void;
  toMatch(expected: string | RegExp): void;
  toThrow(expected?: string | RegExp | Error): void;
  toHaveProperty(property: string, value?: any): void;
  toHaveLength(expected: number): void;
  not: ExpectMatchers<T>;
}

function describe(name: string, fn: () => void): void;
function test(name: string, fn: () => void | Promise<void>, options?: TestOptions): void;
function it(name: string, fn: () => void | Promise<void>, options?: TestOptions): void;
function expect<T>(actual: T): ExpectMatchers<T>;
function beforeAll(fn: () => void | Promise<void>): void;
function afterAll(fn: () => void | Promise<void>): void;
function beforeEach(fn: () => void | Promise<void>): void;
function afterEach(fn: () => void | Promise<void>): void;
```

#### Example
```javascript
import { describe, test, expect, beforeEach, afterEach } from 'gode:test';

describe('Calculator', () => {
  let calculator;
  
  beforeEach(() => {
    calculator = new Calculator();
  });
  
  afterEach(() => {
    calculator.reset();
  });
  
  describe('addition', () => {
    test('should add two positive numbers', () => {
      const result = calculator.add(2, 3);
      expect(result).toBe(5);
    });
    
    test('should handle negative numbers', () => {
      const result = calculator.add(-5, 3);
      expect(result).toBe(-2);
    });
  });
  
  test('should throw on division by zero', () => {
    expect(() => calculator.divide(10, 0)).toThrow('Division by zero');
  });
  
  test.skip('pending test', () => {
    // This test will be skipped
  });
  
  test('async operations', async () => {
    const data = await calculator.fetchData();
    expect(data).toHaveProperty('result');
    expect(data.result).toBeGreaterThan(0);
  });
});
```

### Encoding Module (`gode:encoding`)

Text encoding and decoding utilities.

#### Interfaces
```typescript
class TextEncoder {
  readonly encoding: string;
  encode(input?: string): Uint8Array;
  encodeInto(source: string, destination: Uint8Array): { read: number; written: number };
}

class TextDecoder {
  readonly encoding: string;
  readonly fatal: boolean;
  readonly ignoreBOM: boolean;
  constructor(label?: string, options?: { fatal?: boolean; ignoreBOM?: boolean });
  decode(input?: BufferSource, options?: { stream?: boolean }): string;
}
```

#### Example
```javascript
import { TextEncoder, TextDecoder } from 'gode:encoding';

// Encoding
const encoder = new TextEncoder();
const encoded = encoder.encode('Hello, 世界!');
console.log(encoded); // Uint8Array

// Decoding
const decoder = new TextDecoder('utf-8');
const decoded = decoder.decode(encoded);
console.log(decoded); // Hello, 世界!

// Different encodings
const latin1Decoder = new TextDecoder('latin1');
const shiftJisDecoder = new TextDecoder('shift_jis');

// Encoding into existing buffer
const buffer = new Uint8Array(100);
const result = encoder.encodeInto('Hello World', buffer);
console.log(`Read ${result.read} characters, wrote ${result.written} bytes`);
```

## Global Objects

Only web-standard globals are available without import:

```javascript
// Available globally (web standards)
fetch()                              // HTTP client
Request, Response, Headers           // Fetch API
URL, URLSearchParams                 // URL handling
TextEncoder, TextDecoder            // Text encoding
setTimeout, setInterval             // Timers
clearTimeout, clearInterval         
console                             // Basic console
WebSocket                           // WebSocket client
AbortController, AbortSignal        // Abort operations
Event, EventTarget, CustomEvent     // Event system
Promise, queueMicrotask            // Async primitives
```

## Implementation Notes

### Module Loading
- Built-in modules are loaded from Go code, not from disk
- Zero filesystem overhead for standard library
- Tree-shaking friendly - only import what you use
- Lazy loading - modules loaded on first import

### Performance
- Direct Go bindings without intermediate layer
- Minimal JavaScript wrapper overhead
- Native performance for I/O operations
- Efficient memory usage through Go's GC

### Compatibility
- Can coexist with npm packages
- Support for both ES modules and CommonJS
- Compatible with existing JavaScript tooling
- Works with bundlers (webpack, rollup, etc.)

### TypeScript Support
All modules come with built-in TypeScript definitions, automatically available when using TypeScript.

## Module Aliases

Custom aliases can be configured for convenience:

```json
// In package.json
{
  "gode": {
    "moduleAliases": {
      "@http": "gode:http",
      "@fs": "gode:fs",
      "@test": "gode:test"
    }
  }
}
```

Usage:
```javascript
import { serve } from '@http';
import { readFile } from '@fs';
```