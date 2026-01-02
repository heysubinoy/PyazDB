# PyazDB Benchmark Suite

A modern, interactive benchmark tool for testing PyazDB performance.

## Features

- 📊 **Real-time visualization** of benchmark results with interactive charts
- ⚡ **Automatic operations**: PUT, GET, DELETE benchmarks run sequentially
- 📈 **Comprehensive metrics**: Latency, throughput, success rates, percentiles
- 🎨 **Modern UI**: Dark theme with glassmorphism design
- 🔧 **Configurable**: Adjust operations count, concurrency, key/value sizes

## Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn
- PyazDB running (default: `http://localhost:8080`)

### Installation

```bash
cd benchmark
npm install
```

### Development

```bash
npm run dev
```

This starts the Vite development server at `http://localhost:3000`.

### Production Build

```bash
npm run build
npm run preview
```

## Usage

1. **Configure the endpoint**: Enter your PyazDB leader node URL (default: `http://localhost:8080`)
2. **Set benchmark parameters**:
   - **Operations per Test**: Number of PUT/GET/DELETE operations to run
   - **Concurrency**: Number of parallel requests
   - **Key Size**: Length of generated keys in bytes
   - **Value Size**: Length of generated values in bytes
3. **Run the benchmark**: Click "Run Benchmark" to start
4. **Monitor progress**: Watch real-time charts and statistics update
5. **Analyze results**: Review detailed latency percentiles and success rates

## Metrics Collected

| Metric | Description |
|--------|-------------|
| Total Operations | Number of operations performed per type |
| Success Rate | Percentage of successful operations |
| Avg Latency | Average response time in milliseconds |
| Min/Max Latency | Range of response times |
| P95/P99 Latency | 95th and 99th percentile latencies |
| Throughput | Operations per second |

## Charts

1. **Latency Over Time**: Line chart showing PUT/GET/DELETE latencies
2. **Success Rates**: Doughnut chart showing success/failure distribution
3. **Throughput Over Time**: Line chart of operations per second
4. **Latency Distribution**: Bar chart showing latency buckets (0-10ms, 10-50ms, etc.)

## API Reference

The benchmark uses PyazDB's HTTP API:

- `POST /set` - Set a key-value pair: `{"key": "foo", "value": "bar"}`
- `GET /get?key=foo` - Retrieve a value by key
- `POST /delete` - Delete a key: `{"key": "foo"}`

## Development

### Project Structure

```
benchmark/
├── index.html          # Main HTML page
├── package.json        # Dependencies
├── vite.config.js      # Vite configuration
└── src/
    ├── main.js         # Application entry point
    ├── benchmark.js    # Benchmark runner logic
    ├── client.js       # PyazDB API client
    ├── charts.js       # Chart.js manager
    └── styles.css      # Modern CSS styling
```

### Tech Stack

- **Vite**: Fast build tool and dev server
- **Chart.js**: Interactive chart library
- **Vanilla JS**: No framework dependencies for simplicity

## License

MIT
