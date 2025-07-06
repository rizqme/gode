// JavaScript wrapper for Gode stream module
// This provides a thin layer over the Go implementation

// Simple EventEmitter implementation since we don't have Node.js events module
class EventEmitter {
  constructor() {
    this._events = {};
    this._maxListeners = 10;
  }

  setMaxListeners(n) {
    this._maxListeners = n;
    return this;
  }

  on(event, listener) {
    if (!this._events[event]) {
      this._events[event] = [];
    }
    this._events[event].push(listener);
    return this;
  }

  once(event, listener) {
    const onceWrapper = (...args) => {
      this.removeListener(event, onceWrapper);
      listener.apply(this, args);
    };
    return this.on(event, onceWrapper);
  }

  removeListener(event, listener) {
    if (!this._events[event]) return this;
    const index = this._events[event].indexOf(listener);
    if (index !== -1) {
      this._events[event].splice(index, 1);
    }
    return this;
  }

  off(event, listener) {
    return this.removeListener(event, listener);
  }

  emit(event, ...args) {
    if (!this._events[event]) return false;
    this._events[event].slice().forEach(listener => {
      try {
        listener.apply(this, args);
      } catch (err) {
        // In a real implementation, we'd handle this better
        console.error('EventEmitter error:', err);
      }
    });
    return true;
  }
}

const events = { EventEmitter };

// Base EventEmitter for all streams
class StreamEventEmitter extends events.EventEmitter {
  constructor() {
    super();
    this.setMaxListeners(100); // Allow many pipes
  }
}

// Readable stream wrapper
class Readable extends StreamEventEmitter {
  constructor(options = {}) {
    super();
    
    // Set default options
    this._readableState = {
      objectMode: options.objectMode || false,
      highWaterMark: options.highWaterMark || 16 * 1024,
      encoding: options.encoding || null,
      flowing: null,
      ended: false,
      destroyed: false,
      reading: false
    };

    this.readable = true;
    this.destroyed = false;

    // Initialize the Go stream through the bridge
    this._initGoStream(options);
  }

  _initGoStream(options) {
    // This will be set up by the Go bridge
    this.__stream = null;
  }

  read(size) {
    if (!this.readable) return null;
    
    // Delegate to Go implementation
    if (this.__stream) {
      return this.__stream.read(size);
    }
    
    return null;
  }

  push(chunk, encoding) {
    if (this.destroyed) return false;
    
    // Convert string to buffer if encoding provided
    if (typeof chunk === 'string' && encoding) {
      chunk = Buffer.from(chunk, encoding);
    }
    
    // Delegate to Go implementation
    if (this.__stream) {
      return this.__stream.push(chunk);
    }
    
    return false;
  }

  pause() {
    if (!this.readable) return this;
    
    this._readableState.flowing = false;
    
    // Delegate to Go implementation
    if (this.__stream) {
      this.__stream.pause();
    }
    
    this.emit('pause');
    return this;
  }

  resume() {
    if (!this.readable) return this;
    
    this._readableState.flowing = true;
    
    // Delegate to Go implementation
    if (this.__stream) {
      this.__stream.resume();
    }
    
    this.emit('resume');
    return this;
  }

  isPaused() {
    if (this.__stream) {
      return this.__stream.isPaused();
    }
    return this._readableState.flowing === false;
  }

  pipe(dest, options = {}) {
    if (!dest || typeof dest.write !== 'function') {
      throw new TypeError('dest must be a writable stream');
    }

    const src = this;
    const opts = {
      end: options.end !== false
    };

    // Set up event handlers
    function onData(chunk) {
      const ret = dest.write(chunk);
      if (!ret) {
        src.pause();
      }
    }

    function onEnd() {
      if (opts.end) {
        dest.end();
      }
    }

    function onError(err) {
      dest.destroy(err);
    }

    function onDrain() {
      src.resume();
    }

    function onClose() {
      cleanup();
    }

    function cleanup() {
      src.removeListener('data', onData);
      src.removeListener('end', onEnd);
      src.removeListener('error', onError);
      dest.removeListener('drain', onDrain);
      dest.removeListener('close', onClose);
    }

    // Wire up events
    src.on('data', onData);
    src.on('end', onEnd);
    src.on('error', onError);
    dest.on('drain', onDrain);
    dest.on('close', onClose);

    // Handle unpipe
    dest.on('unpipe', (stream) => {
      if (stream === src) {
        cleanup();
      }
    });

    // Start flowing if not already
    if (this._readableState.flowing !== true) {
      this.resume();
    }

    return dest;
  }

  unpipe(dest) {
    if (dest) {
      dest.emit('unpipe', this);
    } else {
      // Unpipe all destinations
      this.emit('unpipe', this);
    }
    return this;
  }

  destroy(error) {
    if (this.destroyed) return this;
    
    this.destroyed = true;
    this.readable = false;
    this._readableState.destroyed = true;

    // Delegate to Go implementation
    if (this.__stream) {
      this.__stream.destroy(error);
    }

    if (error) {
      this.emit('error', error);
    }
    
    this.emit('close');
    return this;
  }

  // Static method to create readable from iterable
  static from(iterable, options) {
    const readable = new Readable(options);
    
    if (Array.isArray(iterable)) {
      // Handle arrays
      let index = 0;
      readable._read = function() {
        if (index < iterable.length) {
          this.push(iterable[index++]);
        } else {
          this.push(null); // End stream
        }
      };
    } else if (typeof iterable[Symbol.iterator] === 'function') {
      // Handle iterables
      const iterator = iterable[Symbol.iterator]();
      readable._read = function() {
        const { value, done } = iterator.next();
        if (done) {
          this.push(null);
        } else {
          this.push(value);
        }
      };
    } else if (typeof iterable[Symbol.asyncIterator] === 'function') {
      // Handle async iterables
      const iterator = iterable[Symbol.asyncIterator]();
      readable._read = async function() {
        try {
          const { value, done } = await iterator.next();
          if (done) {
            this.push(null);
          } else {
            this.push(value);
          }
        } catch (err) {
          this.destroy(err);
        }
      };
    }
    
    return readable;
  }
}

// Writable stream wrapper
class Writable extends StreamEventEmitter {
  constructor(options = {}) {
    super();
    
    this._writableState = {
      objectMode: options.objectMode || false,
      highWaterMark: options.highWaterMark || 16 * 1024,
      decoding: options.decoding || null,
      ended: false,
      destroyed: false,
      corked: 0,
      needDrain: false
    };

    this.writable = true;
    this.destroyed = false;

    // Store user-provided write function
    if (options.write && typeof options.write === 'function') {
      this._write = options.write;
    }

    if (options.writev && typeof options.writev === 'function') {
      this._writev = options.writev;
    }

    if (options.final && typeof options.final === 'function') {
      this._final = options.final;
    }

    // Initialize the Go stream through the bridge
    this._initGoStream(options);
  }

  _initGoStream(options) {
    // This will be set up by the Go bridge
    this.__stream = null;
  }

  write(chunk, encoding, callback) {
    if (typeof encoding === 'function') {
      callback = encoding;
      encoding = null;
    }

    if (this.destroyed) {
      const err = new Error('write after destroy');
      if (callback) {
        process.nextTick(callback, err);
      } else {
        this.emit('error', err);
      }
      return false;
    }

    if (this._writableState.ended) {
      const err = new Error('write after end');
      if (callback) {
        process.nextTick(callback, err);
      } else {
        this.emit('error', err);
      }
      return false;
    }

    // Convert string to buffer if needed
    if (typeof chunk === 'string') {
      chunk = Buffer.from(chunk, encoding || 'utf8');
    }

    // Delegate to Go implementation or user function
    let result = true;
    
    if (this.__stream) {
      result = this.__stream.write(chunk, encoding, callback);
    } else if (this._write) {
      this._write(chunk, encoding, callback || (() => {}));
    }

    if (!result) {
      this._writableState.needDrain = true;
    }

    return result;
  }

  end(chunk, encoding, callback) {
    if (typeof chunk === 'function') {
      callback = chunk;
      chunk = null;
    } else if (typeof encoding === 'function') {
      callback = encoding;
      encoding = null;
    }

    if (this._writableState.ended) {
      if (callback) {
        process.nextTick(callback);
      }
      return this;
    }

    if (chunk != null) {
      this.write(chunk, encoding);
    }

    this._writableState.ended = true;
    this.writable = false;

    // Delegate to Go implementation
    if (this.__stream) {
      this.__stream.end(chunk, encoding, callback);
    } else if (this._final) {
      this._final(callback || (() => {}));
    }

    this.emit('finish');
    
    if (callback) {
      process.nextTick(callback);
    }

    return this;
  }

  cork() {
    this._writableState.corked++;
    
    // Delegate to Go implementation
    if (this.__stream) {
      this.__stream.cork();
    }
  }

  uncork() {
    if (this._writableState.corked > 0) {
      this._writableState.corked--;
    }
    
    // Delegate to Go implementation
    if (this.__stream) {
      this.__stream.uncork();
    }

    if (this._writableState.corked === 0 && this._writableState.needDrain) {
      this._writableState.needDrain = false;
      this.emit('drain');
    }
  }

  destroy(error) {
    if (this.destroyed) return this;
    
    this.destroyed = true;
    this.writable = false;
    this._writableState.destroyed = true;

    // Delegate to Go implementation
    if (this.__stream) {
      this.__stream.destroy(error);
    }

    if (error) {
      this.emit('error', error);
    }
    
    this.emit('close');
    return this;
  }
}

// Duplex stream (both readable and writable)
class Duplex extends Readable {
  constructor(options = {}) {
    super(options);
    
    // Add writable properties
    this._writableState = {
      objectMode: options.objectMode || false,
      highWaterMark: options.writableHighWaterMark || options.highWaterMark || 16 * 1024,
      decoding: options.decoding || null,
      ended: false,
      destroyed: false,
      corked: 0,
      needDrain: false
    };

    this.writable = true;

    // Store user-provided functions
    if (options.write && typeof options.write === 'function') {
      this._write = options.write;
    }

    if (options.read && typeof options.read === 'function') {
      this._read = options.read;
    }
  }

  // Include all writable methods
  write(chunk, encoding, callback) {
    return Writable.prototype.write.call(this, chunk, encoding, callback);
  }

  end(chunk, encoding, callback) {
    return Writable.prototype.end.call(this, chunk, encoding, callback);
  }

  cork() {
    return Writable.prototype.cork.call(this);
  }

  uncork() {
    return Writable.prototype.uncork.call(this);
  }
}

// Transform stream (duplex with transformation)
class Transform extends Duplex {
  constructor(options = {}) {
    super(options);

    // Store transform functions
    if (options.transform && typeof options.transform === 'function') {
      this._transform = options.transform;
    }

    if (options.flush && typeof options.flush === 'function') {
      this._flush = options.flush;
    }
  }

  _transform(chunk, encoding, callback) {
    // Default implementation just passes through
    callback(null, chunk);
  }

  _flush(callback) {
    // Default implementation does nothing
    callback();
  }
}

// PassThrough stream (transform with no transformation)
class PassThrough extends Transform {
  constructor(options) {
    super(options);
  }

  _transform(chunk, encoding, callback) {
    callback(null, chunk);
  }
}

// Utility functions

function pipeline(...streams) {
  return new Promise((resolve, reject) => {
    let callback;
    
    // Check if last argument is a callback
    if (typeof streams[streams.length - 1] === 'function') {
      callback = streams.pop();
    }

    if (streams.length < 2) {
      const err = new Error('pipeline requires at least 2 streams');
      if (callback) {
        process.nextTick(callback, err);
        return;
      }
      return Promise.reject(err);
    }

    let currentStream = streams[0];
    const destroyStreams = [];

    function cleanup() {
      destroyStreams.forEach(stream => {
        if (typeof stream.destroy === 'function') {
          stream.destroy();
        }
      });
    }

    function onError(err) {
      cleanup();
      if (callback) {
        callback(err);
      } else {
        reject(err);
      }
    }

    function onFinish() {
      if (callback) {
        callback();
      } else {
        resolve();
      }
    }

    try {
      // Connect streams
      for (let i = 1; i < streams.length; i++) {
        const nextStream = streams[i];
        
        currentStream.on('error', onError);
        nextStream.on('error', onError);
        
        destroyStreams.push(currentStream, nextStream);
        
        currentStream.pipe(nextStream);
        currentStream = nextStream;
      }

      // Listen for completion on the last stream
      currentStream.on('finish', onFinish);
      currentStream.on('end', onFinish);
      
    } catch (err) {
      onError(err);
    }
  });
}

function finished(stream, options = {}) {
  return new Promise((resolve, reject) => {
    const checkError = options.error !== false;
    
    function onEnd() {
      resolve();
    }
    
    function onFinish() {
      resolve();
    }
    
    function onError(err) {
      if (checkError) {
        reject(err);
      }
    }
    
    function onClose() {
      resolve();
    }

    // Set up event listeners based on stream type
    if (typeof stream.readable === 'boolean') {
      stream.once('end', onEnd);
    }
    
    if (typeof stream.writable === 'boolean') {
      stream.once('finish', onFinish);
    }
    
    stream.once('close', onClose);
    
    if (checkError) {
      stream.once('error', onError);
    }
  });
}

// Export the module
module.exports = {
  Readable,
  Writable,
  Duplex,
  Transform,
  PassThrough,
  pipeline,
  finished
};