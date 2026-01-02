import Chart from 'chart.js/auto';

/**
 * Chart Manager
 * Handles all chart rendering and updates
 */
export class ChartManager {
  constructor() {
    this.charts = {};
    this.chartColors = {
      put: {
        bg: 'rgba(99, 102, 241, 0.2)',
        border: 'rgba(99, 102, 241, 1)',
      },
      get: {
        bg: 'rgba(16, 185, 129, 0.2)',
        border: 'rgba(16, 185, 129, 1)',
      },
      delete: {
        bg: 'rgba(239, 68, 68, 0.2)',
        border: 'rgba(239, 68, 68, 1)',
      },
      throughput: {
        bg: 'rgba(245, 158, 11, 0.2)',
        border: 'rgba(245, 158, 11, 1)',
      },
    };
    this.defaultOptions = {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'top',
          labels: {
            color: '#94a3b8',
            font: { size: 11 },
            padding: 15,
            usePointStyle: true,
          },
        },
      },
      scales: {
        x: {
          grid: {
            color: 'rgba(100, 100, 150, 0.1)',
          },
          ticks: {
            color: '#64748b',
            font: { size: 10 },
            maxTicksLimit: 10,
          },
        },
        y: {
          grid: {
            color: 'rgba(100, 100, 150, 0.1)',
          },
          ticks: {
            color: '#64748b',
            font: { size: 10 },
          },
        },
      },
    };
  }

  /**
   * Initialize all charts
   */
  init() {
    this.initLatencyChart();
    this.initSuccessChart();
    this.initThroughputChart();
    this.initDistributionChart();
  }

  /**
   * Initialize latency over time chart
   */
  initLatencyChart() {
    const ctx = document.getElementById('latencyChart');
    if (!ctx) return;

    this.charts.latency = new Chart(ctx, {
      type: 'line',
      data: {
        labels: [],
        datasets: [
          {
            label: 'PUT',
            data: [],
            borderColor: this.chartColors.put.border,
            backgroundColor: this.chartColors.put.bg,
            tension: 0.3,
            fill: true,
          },
          {
            label: 'GET',
            data: [],
            borderColor: this.chartColors.get.border,
            backgroundColor: this.chartColors.get.bg,
            tension: 0.3,
            fill: true,
          },
          {
            label: 'DELETE',
            data: [],
            borderColor: this.chartColors.delete.border,
            backgroundColor: this.chartColors.delete.bg,
            tension: 0.3,
            fill: true,
          },
        ],
      },
      options: {
        ...this.defaultOptions,
        scales: {
          ...this.defaultOptions.scales,
          y: {
            ...this.defaultOptions.scales.y,
            title: {
              display: true,
              text: 'Latency (ms)',
              color: '#64748b',
            },
          },
        },
      },
    });
  }

  /**
   * Initialize success rate chart
   */
  initSuccessChart() {
    const ctx = document.getElementById('successChart');
    if (!ctx) return;

    this.charts.success = new Chart(ctx, {
      type: 'doughnut',
      data: {
        labels: ['PUT Success', 'PUT Failed', 'GET Success', 'GET Failed', 'DELETE Success', 'DELETE Failed'],
        datasets: [{
          data: [0, 0, 0, 0, 0, 0],
          backgroundColor: [
            this.chartColors.put.border,
            'rgba(99, 102, 241, 0.3)',
            this.chartColors.get.border,
            'rgba(16, 185, 129, 0.3)',
            this.chartColors.delete.border,
            'rgba(239, 68, 68, 0.3)',
          ],
          borderWidth: 0,
        }],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: 'right',
            labels: {
              color: '#94a3b8',
              font: { size: 10 },
              padding: 10,
              usePointStyle: true,
            },
          },
        },
      },
    });
  }

  /**
   * Initialize throughput chart
   */
  initThroughputChart() {
    const ctx = document.getElementById('throughputChart');
    if (!ctx) return;

    this.charts.throughput = new Chart(ctx, {
      type: 'line',
      data: {
        labels: [],
        datasets: [{
          label: 'Throughput',
          data: [],
          borderColor: this.chartColors.throughput.border,
          backgroundColor: this.chartColors.throughput.bg,
          tension: 0.3,
          fill: true,
        }],
      },
      options: {
        ...this.defaultOptions,
        scales: {
          ...this.defaultOptions.scales,
          y: {
            ...this.defaultOptions.scales.y,
            title: {
              display: true,
              text: 'Operations/sec',
              color: '#64748b',
            },
          },
        },
      },
    });
  }

  /**
   * Initialize latency distribution chart
   */
  initDistributionChart() {
    const ctx = document.getElementById('distributionChart');
    if (!ctx) return;

    this.charts.distribution = new Chart(ctx, {
      type: 'bar',
      data: {
        labels: ['0-10ms', '10-50ms', '50-100ms', '100-500ms', '500ms+'],
        datasets: [
          {
            label: 'PUT',
            data: [0, 0, 0, 0, 0],
            backgroundColor: this.chartColors.put.border,
          },
          {
            label: 'GET',
            data: [0, 0, 0, 0, 0],
            backgroundColor: this.chartColors.get.border,
          },
          {
            label: 'DELETE',
            data: [0, 0, 0, 0, 0],
            backgroundColor: this.chartColors.delete.border,
          },
        ],
      },
      options: {
        ...this.defaultOptions,
        scales: {
          ...this.defaultOptions.scales,
          y: {
            ...this.defaultOptions.scales.y,
            title: {
              display: true,
              text: 'Count',
              color: '#64748b',
            },
          },
        },
      },
    });
  }

  /**
   * Calculate latency distribution buckets
   */
  calculateDistribution(latencies) {
    const buckets = [0, 0, 0, 0, 0];
    for (const latency of latencies) {
      if (latency < 10) buckets[0]++;
      else if (latency < 50) buckets[1]++;
      else if (latency < 100) buckets[2]++;
      else if (latency < 500) buckets[3]++;
      else buckets[4]++;
    }
    return buckets;
  }

  /**
   * Update all charts with new data
   */
  update(data) {
    const { results, timeSeriesData } = data;

    // Update latency chart
    if (this.charts.latency) {
      // Keep only last 50 data points for readability
      const maxPoints = 50;
      const timestamps = timeSeriesData.timestamps.slice(-maxPoints);
      
      this.charts.latency.data.labels = timestamps;
      this.charts.latency.data.datasets[0].data = timeSeriesData.putLatencies.slice(-maxPoints);
      this.charts.latency.data.datasets[1].data = timeSeriesData.getLatencies.slice(-maxPoints);
      this.charts.latency.data.datasets[2].data = timeSeriesData.deleteLatencies.slice(-maxPoints);
      this.charts.latency.update('none');
    }

    // Update success chart
    if (this.charts.success) {
      this.charts.success.data.datasets[0].data = [
        results.put.success,
        results.put.failed,
        results.get.success,
        results.get.failed,
        results.delete.success,
        results.delete.failed,
      ];
      this.charts.success.update('none');
    }

    // Update throughput chart
    if (this.charts.throughput) {
      const maxPoints = 50;
      this.charts.throughput.data.labels = timeSeriesData.timestamps.slice(-maxPoints);
      this.charts.throughput.data.datasets[0].data = timeSeriesData.throughput.slice(-maxPoints);
      this.charts.throughput.update('none');
    }

    // Update distribution chart
    if (this.charts.distribution) {
      this.charts.distribution.data.datasets[0].data = this.calculateDistribution(results.put.latencies);
      this.charts.distribution.data.datasets[1].data = this.calculateDistribution(results.get.latencies);
      this.charts.distribution.data.datasets[2].data = this.calculateDistribution(results.delete.latencies);
      this.charts.distribution.update('none');
    }
  }

  /**
   * Reset all charts
   */
  reset() {
    if (this.charts.latency) {
      this.charts.latency.data.labels = [];
      this.charts.latency.data.datasets.forEach(ds => ds.data = []);
      this.charts.latency.update();
    }

    if (this.charts.success) {
      this.charts.success.data.datasets[0].data = [0, 0, 0, 0, 0, 0];
      this.charts.success.update();
    }

    if (this.charts.throughput) {
      this.charts.throughput.data.labels = [];
      this.charts.throughput.data.datasets[0].data = [];
      this.charts.throughput.update();
    }

    if (this.charts.distribution) {
      this.charts.distribution.data.datasets.forEach(ds => ds.data = [0, 0, 0, 0, 0]);
      this.charts.distribution.update();
    }
  }

  /**
   * Destroy all charts
   */
  destroy() {
    Object.values(this.charts).forEach(chart => {
      if (chart) chart.destroy();
    });
    this.charts = {};
  }
}
