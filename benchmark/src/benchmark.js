import { PyazDBClient } from './client.js';

/**
 * Benchmark Runner
 * Executes benchmark tests and collects metrics
 */
export class BenchmarkRunner {
  constructor(config) {
    this.config = config;
    this.client = new PyazDBClient(config.endpoint);
    this.isRunning = false;
    this.shouldStop = false;
    this.results = {
      put: { total: 0, success: 0, failed: 0, latencies: [] },
      get: { total: 0, success: 0, failed: 0, latencies: [] },
      delete: { total: 0, success: 0, failed: 0, latencies: [] },
    };
    this.timeSeriesData = {
      timestamps: [],
      putLatencies: [],
      getLatencies: [],
      deleteLatencies: [],
      throughput: [],
    };
    this.callbacks = {
      onProgress: null,
      onLog: null,
      onStatsUpdate: null,
      onComplete: null,
    };
    this.startTime = null;
    this.operationCount = 0;
  }

  /**
   * Generate a random string of specified length
   */
  generateRandomString(length) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
  }

  /**
   * Set callback handlers
   */
  on(event, callback) {
    if (this.callbacks.hasOwnProperty(event)) {
      this.callbacks[event] = callback;
    }
  }

  /**
   * Log a message
   */
  log(message, type = 'info') {
    if (this.callbacks.onLog) {
      this.callbacks.onLog(message, type);
    }
  }

  /**
   * Update progress
   */
  updateProgress(current, total, phase) {
    if (this.callbacks.onProgress) {
      this.callbacks.onProgress(current, total, phase);
    }
  }

  /**
   * Update stats
   */
  updateStats() {
    if (this.callbacks.onStatsUpdate) {
      const elapsed = (Date.now() - this.startTime) / 1000;
      const throughput = elapsed > 0 ? Math.round(this.operationCount / elapsed) : 0;
      
      // Record time series data
      this.timeSeriesData.timestamps.push(new Date().toLocaleTimeString());
      
      // Calculate recent latencies (last batch average)
      const recentPut = this.results.put.latencies.slice(-10);
      const recentGet = this.results.get.latencies.slice(-10);
      const recentDelete = this.results.delete.latencies.slice(-10);
      
      this.timeSeriesData.putLatencies.push(
        recentPut.length ? recentPut.reduce((a, b) => a + b, 0) / recentPut.length : 0
      );
      this.timeSeriesData.getLatencies.push(
        recentGet.length ? recentGet.reduce((a, b) => a + b, 0) / recentGet.length : 0
      );
      this.timeSeriesData.deleteLatencies.push(
        recentDelete.length ? recentDelete.reduce((a, b) => a + b, 0) / recentDelete.length : 0
      );
      this.timeSeriesData.throughput.push(throughput);
      
      this.callbacks.onStatsUpdate({
        results: this.results,
        throughput,
        timeSeriesData: this.timeSeriesData,
      });
    }
  }

  /**
   * Sleep for specified milliseconds
   */
  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Calculate delay needed to achieve target TPS
   */
  calculateDelay(batchSize) {
    const { targetTps } = this.config;
    if (!targetTps || targetTps <= 0) return 0;
    // Time for batch at target TPS: (batchSize / targetTps) * 1000 ms
    return (batchSize / targetTps) * 1000;
  }

  /**
   * Run a batch of operations concurrently with rate limiting
   */
  async runBatch(operations) {
    const batchStart = performance.now();
    const results = await Promise.all(operations);
    
    // Apply rate limiting if target TPS is set
    const targetDelay = this.calculateDelay(operations.length);
    if (targetDelay > 0) {
      const elapsed = performance.now() - batchStart;
      const remainingDelay = targetDelay - elapsed;
      if (remainingDelay > 0) {
        await this.sleep(remainingDelay);
      }
    }
    
    return results;
  }

  /**
   * Execute the PUT phase
   */
  async runPutPhase(keys) {
    this.log(`Starting PUT phase with ${keys.length} operations...`, 'info');
    const { concurrency, valueSize } = this.config;
    
    for (let i = 0; i < keys.length && !this.shouldStop; i += concurrency) {
      const batch = keys.slice(i, i + concurrency);
      const operations = batch.map(key => {
        const value = this.generateRandomString(valueSize);
        return this.client.put(key, value);
      });
      
      const results = await this.runBatch(operations);
      
      for (const result of results) {
        this.results.put.total++;
        this.operationCount++;
        if (result.success) {
          this.results.put.success++;
          this.results.put.latencies.push(result.latency);
        } else {
          this.results.put.failed++;
          this.log(`PUT failed: ${result.error}`, 'error');
        }
      }
      
      this.updateProgress(i + batch.length, keys.length * 3, 'PUT');
      this.updateStats();
    }
    
    this.log(`PUT phase complete: ${this.results.put.success}/${this.results.put.total} successful`, 
      this.results.put.failed > 0 ? 'warning' : 'success');
  }

  /**
   * Execute the GET phase
   */
  async runGetPhase(keys) {
    this.log(`Starting GET phase with ${keys.length} operations...`, 'info');
    const { concurrency } = this.config;
    const baseProgress = this.config.operations;
    
    for (let i = 0; i < keys.length && !this.shouldStop; i += concurrency) {
      const batch = keys.slice(i, i + concurrency);
      const operations = batch.map(key => this.client.get(key));
      
      const results = await this.runBatch(operations);
      
      for (const result of results) {
        this.results.get.total++;
        this.operationCount++;
        if (result.success) {
          this.results.get.success++;
          this.results.get.latencies.push(result.latency);
        } else {
          this.results.get.failed++;
          if (!result.error.includes('not found')) {
            this.log(`GET failed: ${result.error}`, 'error');
          }
        }
      }
      
      this.updateProgress(baseProgress + i + batch.length, keys.length * 3, 'GET');
      this.updateStats();
    }
    
    this.log(`GET phase complete: ${this.results.get.success}/${this.results.get.total} successful`,
      this.results.get.failed > 0 ? 'warning' : 'success');
  }

  /**
   * Execute the DELETE phase
   */
  async runDeletePhase(keys) {
    this.log(`Starting DELETE phase with ${keys.length} operations...`, 'info');
    const { concurrency } = this.config;
    const baseProgress = this.config.operations * 2;
    
    for (let i = 0; i < keys.length && !this.shouldStop; i += concurrency) {
      const batch = keys.slice(i, i + concurrency);
      const operations = batch.map(key => this.client.delete(key));
      
      const results = await this.runBatch(operations);
      
      for (const result of results) {
        this.results.delete.total++;
        this.operationCount++;
        if (result.success) {
          this.results.delete.success++;
          this.results.delete.latencies.push(result.latency);
        } else {
          this.results.delete.failed++;
          this.log(`DELETE failed: ${result.error}`, 'error');
        }
      }
      
      this.updateProgress(baseProgress + i + batch.length, keys.length * 3, 'DELETE');
      this.updateStats();
    }
    
    this.log(`DELETE phase complete: ${this.results.delete.success}/${this.results.delete.total} successful`,
      this.results.delete.failed > 0 ? 'warning' : 'success');
  }

  /**
   * Calculate percentile from sorted array
   */
  percentile(arr, p) {
    if (arr.length === 0) return 0;
    const sorted = [...arr].sort((a, b) => a - b);
    const index = Math.ceil((p / 100) * sorted.length) - 1;
    return sorted[Math.max(0, index)];
  }

  /**
   * Calculate final statistics
   */
  calculateFinalStats() {
    const calcStats = (latencies) => {
      if (latencies.length === 0) {
        return { avg: 0, min: 0, max: 0, p95: 0, p99: 0 };
      }
      const sorted = [...latencies].sort((a, b) => a - b);
      return {
        avg: latencies.reduce((a, b) => a + b, 0) / latencies.length,
        min: sorted[0],
        max: sorted[sorted.length - 1],
        p95: this.percentile(latencies, 95),
        p99: this.percentile(latencies, 99),
      };
    };

    return {
      put: { ...this.results.put, stats: calcStats(this.results.put.latencies) },
      get: { ...this.results.get, stats: calcStats(this.results.get.latencies) },
      delete: { ...this.results.delete, stats: calcStats(this.results.delete.latencies) },
      timeSeriesData: this.timeSeriesData,
      duration: (Date.now() - this.startTime) / 1000,
      totalOperations: this.operationCount,
    };
  }

  /**
   * Run the full benchmark
   */
  async run() {
    if (this.isRunning) {
      this.log('Benchmark already running!', 'warning');
      return;
    }

    this.isRunning = true;
    this.shouldStop = false;
    this.startTime = Date.now();
    this.operationCount = 0;
    
    // Reset results
    this.results = {
      put: { total: 0, success: 0, failed: 0, latencies: [] },
      get: { total: 0, success: 0, failed: 0, latencies: [] },
      delete: { total: 0, success: 0, failed: 0, latencies: [] },
    };
    this.timeSeriesData = {
      timestamps: [],
      putLatencies: [],
      getLatencies: [],
      deleteLatencies: [],
      throughput: [],
    };

    this.client.setEndpoint(this.config.endpoint);
    
    this.log(`Starting benchmark against ${this.config.endpoint}`, 'info');
    this.log(`Config: ${this.config.operations} operations, concurrency: ${this.config.concurrency}`, 'info');

    // Check connection
    this.log('Checking connection to PyazDB...', 'info');
    const healthy = await this.client.healthCheck();
    if (!healthy) {
      this.log('Failed to connect to PyazDB endpoint!', 'error');
      this.isRunning = false;
      if (this.callbacks.onComplete) {
        this.callbacks.onComplete(null, 'Connection failed');
      }
      return;
    }
    this.log('Connection successful!', 'success');

    // Generate keys
    const { operations, keySize } = this.config;
    const keys = [];
    for (let i = 0; i < operations; i++) {
      keys.push(`bench_${this.generateRandomString(keySize)}_${i}`);
    }

    try {
      // Run phases
      await this.runPutPhase(keys);
      if (!this.shouldStop) await this.runGetPhase(keys);
      if (!this.shouldStop) await this.runDeletePhase(keys);

      const finalStats = this.calculateFinalStats();
      
      if (this.shouldStop) {
        this.log('Benchmark stopped by user', 'warning');
      } else {
        this.log(`Benchmark complete! Total time: ${finalStats.duration.toFixed(2)}s`, 'success');
        this.log(`Overall throughput: ${Math.round(finalStats.totalOperations / finalStats.duration)} ops/s`, 'success');
      }

      if (this.callbacks.onComplete) {
        this.callbacks.onComplete(finalStats);
      }
    } catch (error) {
      this.log(`Benchmark error: ${error.message}`, 'error');
      if (this.callbacks.onComplete) {
        this.callbacks.onComplete(null, error.message);
      }
    }

    this.isRunning = false;
  }

  /**
   * Stop the running benchmark
   */
  stop() {
    this.shouldStop = true;
    this.log('Stopping benchmark...', 'warning');
  }
}
