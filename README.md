# Game of Life

A concurrent implementation of Conway's Game of Life in Go, developed as coursework for COMS20008 (Concurrent Computing) at the University of Bristol.

**Authors**: Jingxiang Zhang, Lingyi Lu

## Overview

This project contains two implementations of Conway's Game of Life:

- **Parallel**: A multi-threaded implementation using Go's goroutines and channels for parallel processing on a single machine.
- **Distributed**: A distributed implementation using a broker-server architecture for processing across multiple machines (AWS).

Both implementations support SDL live visualization and feature optimized algorithms for efficient cell state computation.

## Key Highlights

### Parallel Implementation Features

Our parallel implementation employs an innovative **flip-cell tracking algorithm** that significantly reduces computation overhead:

- **Dynamic Flip-Cell Detection**: Instead of checking every cell in the grid each turn, the algorithm tracks only cells that changed in the previous turn and their neighbors
- **Neighbor Propagation Optimization**: The `getNeighbor` function builds a candidate list of cells that might change, achieving up to **88% relative improvement** over the naive approach
- **Efficient Memory Usage**: Pre-allocated slices for worker results minimize garbage collection overhead
- **Parallel Worker Distribution**: Workload is evenly distributed across goroutines with efficient channel-based communication

#### How It Works

1. **Turn 1**: Workers scan the full grid in parallel, each handling a horizontal slice
2. **Subsequent Turns**: Only cells neighboring the previously flipped cells are evaluated
3. **Cell Flip Detection**: Each worker returns cells that need to flip, which are aggregated by the distributor
4. **Neighbor List Generation**: Flipped cells generate the candidate list for the next turn

### Distributed Implementation Features

Our distributed system uses a **broker-worker architecture** designed for fault tolerance and scalability:

- **Single Broker Process**: Manages all client connections and worker coordination via Go RPC
- **Dynamic Worker Pool**: Workers can join or leave the pool without disrupting ongoing computations
- **Persistent TCP Streaming**: Long-lived connections for turn data and alive cell counts eliminate per-turn handshakes

#### Architecture Highlights

```
┌─────────────┐        ┌─────────────┐        ┌─────────────┐
│   Client    │◄──────►│   Broker    │◄──────►│  Worker 1   │
│ (Distributor)│  RPC   │  (AWS EC2)  │  RPC   │  (AWS EC2)  │
└─────────────┘        └──────┬──────┘        └─────────────┘
                              │
                              │ RPC
                              ▼
                       ┌─────────────┐
                       │  Worker N   │
                       │  (AWS EC2)  │
                       └─────────────┘
```

### Special Design Advantages

#### 1. High Fault Tolerance
- **Automatic Worker Recovery**: Dead workers are detected and removed from the pool automatically
- **Turn Retry Mechanism**: Failed turns are automatically re-executed with remaining workers
- **Graceful Worker Addition**: New workers joining during execution are utilized in subsequent turns without crashes

#### 2. Minimized Communication Overhead
- **Full World Transfer Only Once**: The complete board is sent to the broker only at the start
- **Differential Updates**: Subsequent turns transfer only the flip-cell list and minimal halo rows
- **Efficient Halo Exchange**: Workers receive only the rows they need for boundary neighbor lookups

#### 3. Persistent TCP Stream
- **Long-lived Connections**: Avoids connection setup overhead for each turn
- **Real-time Updates**: Alive cell counts sent every 2 seconds over the same connection
- **Unified Event Stream**: Keypress responses and turn completions share the reliable stream

#### 4. Scalable Worker Pool
- **Open-ended Registration**: Any server that dials `Broker.Register` joins the pool
- **Dynamic Repartitioning**: Work is redistributed when workers join or leave
- **Thread Distribution**: Threads are allocated evenly across available workers

#### 5. Privacy & Decoupling
- **Address Isolation**: Client only knows broker's address; workers are hidden
- **Proactive Worker Connection**: Workers dial the broker, requiring no exposed ports
- **Zero Knowledge Between Workers**: Servers don't know about other servers or clients

## Development Evolution

The parallel implementation went through several iterations:

| Version | Approach | Key Improvement |
|---------|----------|-----------------|
| V1 | Basic parallel workers with neighbor counting | Baseline implementation |
| V2 | Enhanced I/O operations | Improved file handling |
| V3 | Flip-cell neighbor tracking | **88% performance improvement** |
| V4 | Efficient slice usage | Reduced memory allocations |
| Final | Optimized goroutine coordination | Mutex locks replaced with I/O channels |

## Performance Insights

Through extensive testing on both local machines and AWS t3.micro instances:

- **Parallel Performance**: Strong results on larger grids (512x512 and above) where computation outweighs coordination overhead
- **Distributed Trade-offs**: Network latency dominates for smaller workloads; best suited for massive grids or when local resources are limited
- **Key Learning**: Distributed computing doesn't inherently guarantee faster performance—success depends on workload granularity and network conditions

## Project Structure

```
.
├── Parallel/           # Parallel implementation
│   ├── gol/           # Core game logic (distributor, workers)
│   ├── sdl/           # SDL visualization
│   ├── images/        # Input PGM images
│   ├── out/           # Output directory
│   └── tests/         # Test files
├── Distributed/        # Distributed implementation
│   ├── gol/           # Core game logic (client-side distributor)
│   ├── broker/        # Broker component (coordination)
│   ├── server/        # Server component (worker logic)
│   ├── stubs/         # RPC stubs and data structures
│   ├── sdl/           # SDL visualization
│   ├── images/        # Input PGM images
│   ├── out/           # Output directory
│   └── tests/         # Test files
├── GOL.pptx           # Presentation slides
├── report.pdf         # Coursework report
└── LICENSE            # MIT License
```

## Requirements

- Go 1.17 or later
- SDL2 library (for visualization)

## Usage

### Parallel Version

```bash
cd Parallel
go run . [flags]
```

### Distributed Version

```bash
cd Distributed

# Start the broker (on one AWS node)
go run broker/broker.go

# Start server(s) on each AWS worker node
# Each server automatically connects to the broker
go run server/server.go

# Run the client (connects to broker)
go run . [flags]
```

> **Note**: For multi-node deployment, run `go run server/server.go` on each AWS node. Servers automatically register with the broker, and work is dynamically distributed across all connected workers.

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-t` | Number of worker threads | 8 |
| `-w` | Width of the image | 512 |
| `-h` | Height of the image | 512 |
| `-turns` | Number of turns to process | 100000000 (Parallel) / 10000000000 (Distributed) |
| `-headless` | Disable SDL window | false |

### Keyboard Controls

| Key | Action | Parallel | Distributed |
|-----|--------|----------|-------------|
| `s` | Save current state as PGM image | ✓ | ✓ |
| `q` | Save current state and quit | ✓ | ✓ |
| `p` | Pause/resume execution | ✓ | ✓ |
| `k` | Quit without saving (terminates workers) | ✗ | ✓ |

## Running Tests

All tests pass with the race detector enabled:

```bash
cd Parallel  # or Distributed
go test -v -race ./...
```

## Conclusion

This coursework demonstrates the practical challenges and trade-offs in concurrent and distributed systems:

- **Parallelism isn't free**: Coordination overhead can negate benefits for small workloads
- **Communication costs matter**: Network latency in distributed systems requires careful architectural decisions
- **Fault tolerance is achievable**: With proper design, distributed systems can gracefully handle node failures
- **Algorithm optimization compounds**: The flip-cell tracking optimization provides significant speedups across all configurations

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Documentation

Additional documentation is available at: https://uob-csa.github.io/gol-docs/
