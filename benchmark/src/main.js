import { BenchmarkRunner } from './benchmark.js';
import { ChartManager } from './charts.js';
import { PyazDBClient } from './client.js';

/**
 * PyazDB Benchmark Suite - Main Application
 */
class BenchmarkApp {
  constructor() {
    this.runner = null;
    this.chartManager = new ChartManager();
    this.client = new PyazDBClient('http://localhost:8080');
    this.connectionCheckInterval = null;
    
    // DOM Elements
    this.elements = {
      // Config inputs
      endpoint: document.getElementById('endpoint'),
      operations: document.getElementById('operations'),
      concurrency: document.getElementById('concurrency'),
      keySize: document.getElementById('keySize'),
      valueSize: document.getElementById('valueSize'),
      targetTps: document.getElementById('targetTps'),
      
      // Buttons
      runBenchmark: document.getElementById('runBenchmark'),
      stopBenchmark: document.getElementById('stopBenchmark'),
      clearResults: document.getElementById('clearResults'),
      clearLog: document.getElementById('clearLog'),
      
      // Progress
      progressSection: document.getElementById('progressSection'),
      progressText: document.getElementById('progressText'),
      progressPercent: document.getElementById('progressPercent'),
      progressFill: document.getElementById('progressFill'),
      
      // Stats
      putCount: document.getElementById('putCount'),
      putSuccess: document.getElementById('putSuccess'),
      getCount: document.getElementById('getCount'),
      getSuccess: document.getElementById('getSuccess'),
      deleteCount: document.getElementById('deleteCount'),
      deleteSuccess: document.getElementById('deleteSuccess'),
      avgLatency: document.getElementById('avgLatency'),
      throughput: document.getElementById('throughput'),
      
      // Connection status
      connectionStatus: document.getElementById('connectionStatus'),
      
      // Results & Log
      resultsBody: document.getElementById('resultsBody'),
      logContainer: document.getElementById('logContainer'),
    };

    this.init();
  }

  /**
   * Initialize the application
   */
  init() {
    this.chartManager.init();
    this.bindEvents();
    this.startConnectionCheck();
    this.log('Benchmark suite initialized. Configure settings and click "Run Benchmark" to start.', 'info');
  }

  /**
   * Bind event handlers
   */
  bindEvents() {
    this.elements.runBenchmark.addEventListener('click', () => this.startBenchmark());
    this.elements.stopBenchmark.addEventListener('click', () => this.stopBenchmark());
    this.elements.clearResults.addEventListener('click', () => this.clearResults());
    this.elements.clearLog.addEventListener('click', () => this.clearLog());
    
    // Update client endpoint when input changes
    this.elements.endpoint.addEventListener('change', () => {
      this.client.setEndpoint(this.elements.endpoint.value);
      this.checkConnection();
    });
  }

  /**
   * Start connection status checking
   */
  startConnectionCheck() {
    this.checkConnection();
    this.connectionCheckInterval = setInterval(() => this.checkConnection(), 5000);
  }

  /**
   * Check connection to PyazDB
   */
  async checkConnection() {
    this.client.setEndpoint(this.elements.endpoint.value);
    const isConnected = await this.client.healthCheck();
    
    if (isConnected) {
      this.elements.connectionStatus.classList.add('connected');
      this.elements.connectionStatus.querySelector('.status-text').textContent = 'Connected';
    } else {
      this.elements.connectionStatus.classList.remove('connected');
      this.elements.connectionStatus.querySelector('.status-text').textContent = 'Disconnected';
    }
  }

  /**
   * Get current configuration
   */
  getConfig() {
    return {
      endpoint: this.elements.endpoint.value,
      operations: parseInt(this.elements.operations.value) || 100,
      concurrency: parseInt(this.elements.concurrency.value) || 10,
      keySize: parseInt(this.elements.keySize.value) || 16,
      valueSize: parseInt(this.elements.valueSize.value) || 128,
      targetTps: parseInt(this.elements.targetTps.value) || 0,
    };
  }

  /**
   * Log a message to the activity log
   */
  log(message, type = 'info') {
    const timestamp = new Date().toLocaleTimeString();
    const entry = document.createElement('div');
    entry.className = `log-entry ${type}`;
    entry.textContent = `[${timestamp}] ${message}`;
    
    this.elements.logContainer.appendChild(entry);
    this.elements.logContainer.scrollTop = this.elements.logContainer.scrollHeight;
  }

  /**
   * Clear the activity log
   */
  clearLog() {
    this.elements.logContainer.innerHTML = '';
    this.log('Log cleared', 'info');
  }

  /**
   * Start the benchmark
   */
  async startBenchmark() {
    const config = this.getConfig();
    
    // Validate config
    if (!config.endpoint) {
      this.log('Please enter a valid endpoint', 'error');
      return;
    }
    
    // Update UI
    this.elements.runBenchmark.disabled = true;
    this.elements.stopBenchmark.disabled = false;
    this.elements.progressSection.style.display = 'block';
    
    // Reset stats
    this.resetStats();
    this.chartManager.reset();
    
    // Create runner
    this.runner = new BenchmarkRunner(config);
    
    // Set callbacks
    this.runner.on('onProgress', (current, total, phase) => {
      const percent = Math.round((current / total) * 100);
      this.elements.progressFill.style.width = `${percent}%`;
      this.elements.progressPercent.textContent = `${percent}%`;
      this.elements.progressText.textContent = `Running ${phase} phase...`;
    });
    
    this.runner.on('onLog', (message, type) => {
      this.log(message, type);
    });
    
    this.runner.on('onStatsUpdate', (data) => {
      this.updateStats(data);
      this.chartManager.update(data);
    });
    
    this.runner.on('onComplete', (results, error) => {
      this.onBenchmarkComplete(results, error);
    });
    
    // Run benchmark
    await this.runner.run();
  }

  /**
   * Stop the benchmark
   */
  stopBenchmark() {
    if (this.runner) {
      this.runner.stop();
    }
  }

  /**
   * Handle benchmark completion
   */
  onBenchmarkComplete(results, error) {
    this.elements.runBenchmark.disabled = false;
    this.elements.stopBenchmark.disabled = true;
    
    if (error) {
      this.elements.progressText.textContent = `Benchmark failed: ${error}`;
      return;
    }
    
    this.elements.progressText.textContent = 'Benchmark complete!';
    this.elements.progressFill.style.width = '100%';
    this.elements.progressPercent.textContent = '100%';
    
    if (results) {
      this.updateResultsTable(results);
    }
  }

  /**
   * Reset stats display
   */
  resetStats() {
    this.elements.putCount.textContent = '0';
    this.elements.putSuccess.textContent = '0%';
    this.elements.getCount.textContent = '0';
    this.elements.getSuccess.textContent = '0%';
    this.elements.deleteCount.textContent = '0';
    this.elements.deleteSuccess.textContent = '0%';
    this.elements.avgLatency.textContent = '0 ms';
    this.elements.throughput.textContent = '0';
  }

  /**
   * Update stats display
   */
  updateStats(data) {
    const { results, throughput } = data;
    
    // PUT stats
    this.elements.putCount.textContent = results.put.total;
    const putSuccessRate = results.put.total > 0 
      ? Math.round((results.put.success / results.put.total) * 100) 
      : 0;
    this.elements.putSuccess.textContent = `${putSuccessRate}%`;
    
    // GET stats
    this.elements.getCount.textContent = results.get.total;
    const getSuccessRate = results.get.total > 0 
      ? Math.round((results.get.success / results.get.total) * 100) 
      : 0;
    this.elements.getSuccess.textContent = `${getSuccessRate}%`;
    
    // DELETE stats
    this.elements.deleteCount.textContent = results.delete.total;
    const deleteSuccessRate = results.delete.total > 0 
      ? Math.round((results.delete.success / results.delete.total) * 100) 
      : 0;
    this.elements.deleteSuccess.textContent = `${deleteSuccessRate}%`;
    
    // Average latency
    const allLatencies = [
      ...results.put.latencies,
      ...results.get.latencies,
      ...results.delete.latencies,
    ];
    const avgLatency = allLatencies.length > 0 
      ? allLatencies.reduce((a, b) => a + b, 0) / allLatencies.length 
      : 0;
    this.elements.avgLatency.textContent = `${avgLatency.toFixed(2)} ms`;
    
    // Throughput
    this.elements.throughput.textContent = throughput;
  }

  /**
   * Update results table
   */
  updateResultsTable(results) {
    const formatLatency = (ms) => `${ms.toFixed(2)} ms`;
    
    const rows = ['put', 'get', 'delete'].map(op => {
      const data = results[op];
      const stats = data.stats;
      const successRate = data.total > 0 
        ? `${Math.round((data.success / data.total) * 100)}%` 
        : '0%';
      
      return `
        <tr>
          <td>${op.toUpperCase()}</td>
          <td>${data.total}</td>
          <td>${data.success} (${successRate})</td>
          <td>${data.failed}</td>
          <td>${formatLatency(stats.avg)}</td>
          <td>${formatLatency(stats.min)}</td>
          <td>${formatLatency(stats.max)}</td>
          <td>${formatLatency(stats.p95)}</td>
          <td>${formatLatency(stats.p99)}</td>
        </tr>
      `;
    }).join('');
    
    // Add summary row
    const totalOps = results.put.total + results.get.total + results.delete.total;
    const totalSuccess = results.put.success + results.get.success + results.delete.success;
    const totalFailed = results.put.failed + results.get.failed + results.delete.failed;
    const allLatencies = [
      ...results.put.latencies,
      ...results.get.latencies,
      ...results.delete.latencies,
    ].sort((a, b) => a - b);
    
    const summaryStats = {
      avg: allLatencies.length > 0 ? allLatencies.reduce((a, b) => a + b, 0) / allLatencies.length : 0,
      min: allLatencies[0] || 0,
      max: allLatencies[allLatencies.length - 1] || 0,
      p95: this.percentile(allLatencies, 95),
      p99: this.percentile(allLatencies, 99),
    };
    
    const summaryRow = `
      <tr style="background: rgba(99, 102, 241, 0.1); font-weight: 600;">
        <td>TOTAL</td>
        <td>${totalOps}</td>
        <td>${totalSuccess} (${totalOps > 0 ? Math.round((totalSuccess / totalOps) * 100) : 0}%)</td>
        <td>${totalFailed}</td>
        <td>${formatLatency(summaryStats.avg)}</td>
        <td>${formatLatency(summaryStats.min)}</td>
        <td>${formatLatency(summaryStats.max)}</td>
        <td>${formatLatency(summaryStats.p95)}</td>
        <td>${formatLatency(summaryStats.p99)}</td>
      </tr>
    `;
    
    this.elements.resultsBody.innerHTML = rows + summaryRow;
  }

  /**
   * Calculate percentile
   */
  percentile(arr, p) {
    if (arr.length === 0) return 0;
    const index = Math.ceil((p / 100) * arr.length) - 1;
    return arr[Math.max(0, index)];
  }

  /**
   * Clear all results
   */
  clearResults() {
    this.resetStats();
    this.chartManager.reset();
    this.elements.progressSection.style.display = 'none';
    this.elements.resultsBody.innerHTML = `
      <tr>
        <td colspan="9" class="empty-state">Run a benchmark to see results</td>
      </tr>
    `;
    this.log('Results cleared', 'info');
  }
}

// Initialize application when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  new BenchmarkApp();
});
