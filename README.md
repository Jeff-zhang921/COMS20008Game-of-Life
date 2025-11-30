# Game of Life

A concurrent implementation of Conway's Game of Life in Go, developed as coursework for COMS20008 (Concurrent Computing) at the University of Bristol.

## Overview

This project contains two implementations of Conway's Game of Life:

- **Parallel**: A multi-threaded implementation using Go's goroutines and channels for parallel processing on a single machine.
- **Distributed**: A distributed implementation using a broker-server architecture for processing across multiple machines.

## Project Structure

```
.
├── Parallel/           # Parallel implementation
│   ├── gol/           # Core game logic
│   ├── sdl/           # SDL visualization
│   ├── images/        # Input PGM images
│   ├── out/           # Output directory
│   └── tests/         # Test files
├── Distributed/        # Distributed implementation
│   ├── gol/           # Core game logic (client-side)
│   ├── broker/        # Broker component
│   ├── server/        # Server component
│   ├── stubs/         # RPC stubs
│   ├── sdl/           # SDL visualization
│   ├── images/        # Input PGM images
│   ├── out/           # Output directory
│   └── tests/         # Test files
├── report.pdf          # Coursework report
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

# Start the server
go run server/server.go

# Start the broker
go run broker/broker.go

# Run the client
go run . [flags]
```

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-t` | Number of worker threads | 8 |
| `-w` | Width of the image | 512 |
| `-h` | Height of the image | 512 |
| `-turns` | Number of turns to process | 100000000 (Parallel) / 10000000000 (Distributed) |
| `-headless` | Disable SDL window | false |

### Keyboard Controls

- `s` - Save current state as PGM image
- `q` - Save current state and quit
- `p` - Pause/resume execution
- `k` - Quit without saving (distributed only)

## Running Tests

```bash
cd Parallel  # or Distributed
go test -v ./...
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Documentation

Additional documentation is available at: https://uob-csa.github.io/gol-docs/
