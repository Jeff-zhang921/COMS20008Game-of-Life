<div align="center">

![Header](https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=2,3,4,6,24,30&height=200&section=header&text=Conway%27s%20Game%20of%20Life&fontSize=45&fontColor=fff&animation=twinkling&fontAlignY=35&desc=A%20High-Performance%20Concurrent%20%26%20Distributed%20Implementation%20in%20Go&descSize=18&descAlignY=55)

<br>

![Visitor Count](https://komarev.com/ghpvc/?username=Jeff-zhang921&repo=COMS20008Game-of-Life&color=blueviolet&style=for-the-badge&label=VISITORS)

<br>

![Go Version](https://img.shields.io/badge/Go-1.17+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Platform](https://img.shields.io/badge/Platform-AWS%20%7C%20Local-orange?style=for-the-badge&logo=amazon-aws)
![University](https://img.shields.io/badge/University%20of%20Bristol-COMS20008-red?style=for-the-badge)
![Stars](https://img.shields.io/github/stars/Jeff-zhang921/COMS20008Game-of-Life?style=for-the-badge&logo=github&color=yellow)

<br>

<img alt="Game of Life Animation" src="https://upload.wikimedia.org/wikipedia/commons/e/e5/Gospers_glider_gun.gif" width="450">

<br>

### ✨ *Experience the mesmerizing patterns of cellular automaton with lightning-fast parallel processing* ✨

<br>

**👨‍💻 Authors**: [Jingxiang Zhang](https://github.com/Jeff-zhang921) • Lingyi Lu

<br>

[![Readme Card](https://github-readme-stats.vercel.app/api/pin/?username=Jeff-zhang921&repo=COMS20008Game-of-Life&theme=radical)](https://github.com/Jeff-zhang921/COMS20008Game-of-Life)

---

[Features](#-features) • [Architecture](#-architecture) • [Performance](#-performance-benchmarks) • [Quick Start](#-quick-start) • [Documentation](#-documentation)

</div>

<br>

## ✨ Features

> [!TIP]
> 🎯 Our implementation achieves **9.5× speedup** over baseline through innovative flip-cell algorithms!

<table>
<tr>
<td width="50%">

### 🚀 Parallel Implementation
- **Innovative Flip-Cell Algorithm** — 88% performance boost
- **Dynamic Workload Distribution** across goroutines
- **Real-time SDL Visualization**
- **Optimized Memory Management** with pre-allocated slices

</td>
<td width="50%">

### 🌐 Distributed Implementation
- **Broker-Worker Architecture** on AWS EC2
- **Fault-Tolerant Design** with auto-recovery
- **Persistent TCP Streaming** for minimal overhead
- **Dynamic Worker Scaling** — add/remove nodes live

</td>
</tr>
</table>

<div align="center">

![Gosper Glider Gun](https://upload.wikimedia.org/wikipedia/commons/e/e5/Gospers_glider_gun.gif)

*Gosper's Glider Gun — a classic pattern that produces gliders indefinitely*

</div>

<br>

---

## 🏗️ Architecture

> [!NOTE]
> Our implementation features a sophisticated **multi-layered architecture** designed for maximum performance and reliability.

### System Overview

<br>

<div align="center">

### 📊 Parallel Implementation Flow

<img src="https://github.com/user-attachments/assets/f13edfc4-be45-4f8b-b94c-a13975fcc667" alt="Parallel Implementation Flow Diagram" width="90%">

<br>

*The distributor coordinates I/O operations and worker goroutines, efficiently managing cell state updates across turns*

</div>

<br>

### 🔄 How the Parallel Algorithm Works

> [!IMPORTANT]
> The flip-cell algorithm is our key innovation — instead of checking all cells every turn, we only check neighbors of cells that changed!

```
┌─────────────────────────────────────────────────────────────────────┐
│                         TURN PROCESSING                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   Turn 1:  Full grid scan → Identify flipped cells                  │
│            ┌──────┬──────┬──────┬──────┐                            │
│            │ W1   │ W2   │ W3   │ W4   │  ← Workers scan slices     │
│            └──────┴──────┴──────┴──────┘                            │
│                        ↓                                             │
│   Turn 2+: Check ONLY neighbors of flipped cells                    │
│            ┌──────────────────────────┐                             │
│            │  ○ ○ ● ○ ○ ○ ○ ○ ○ ○ ○  │  ● = Check this cell        │
│            │  ○ ● ● ● ○ ○ ○ ○ ○ ○ ○  │  ○ = Skip (unchanged)       │
│            │  ○ ○ ● ○ ○ ○ ○ ○ ○ ○ ○  │                              │
│            └──────────────────────────┘                             │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 🌍 Distributed Architecture

```
                    ┌─────────────────────────────────────┐
                    │           CLIENT MACHINE            │
                    │  ┌─────────────────────────────┐   │
                    │  │       Distributor           │   │
                    │  │   • Manages game state      │   │
                    │  │   • Handles SDL display     │   │
                    │  │   • Processes key events    │   │
                    │  └──────────────┬──────────────┘   │
                    └─────────────────┼──────────────────┘
                                      │ RPC
                    ┌─────────────────▼──────────────────┐
                    │            AWS BROKER              │
                    │  ┌─────────────────────────────┐   │
                    │  │     Coordination Layer       │   │
                    │  │   • Worker registration      │   │
                    │  │   • Task distribution        │   │
                    │  │   • Fault detection          │   │
                    │  └──────────────┬──────────────┘   │
                    └─────────────────┼──────────────────┘
                                      │ RPC
          ┌───────────────────────────┼───────────────────────────┐
          │                           │                           │
          ▼                           ▼                           ▼
   ┌─────────────┐            ┌─────────────┐            ┌─────────────┐
   │  Worker 1   │            │  Worker 2   │            │  Worker N   │
   │  (AWS EC2)  │            │  (AWS EC2)  │            │  (AWS EC2)  │
   │             │            │             │            │             │
   │ ┌─────────┐ │            │ ┌─────────┐ │            │ ┌─────────┐ │
   │ │Thread 1 │ │            │ │Thread 1 │ │            │ │Thread 1 │ │
   │ │Thread 2 │ │            │ │Thread 2 │ │            │ │Thread 2 │ │
   │ │  ...    │ │            │ │  ...    │ │            │ │  ...    │ │
   │ └─────────┘ │            │ └─────────┘ │            │ └─────────┘ │
   └─────────────┘            └─────────────┘            └─────────────┘
```

<br>

---

## 📈 Performance Benchmarks

> [!NOTE]
> 📊 **Benchmark Environment**: Intel Core i7, 16GB RAM, Ubuntu 22.04

Our optimizations deliver **exceptional performance gains** across different workloads and configurations.

<br>

<div align="center">

### 🏆 Version Evolution Performance (512×512 Grid)

<img src="https://github.com/user-attachments/assets/10034665-b506-4710-b47a-9a2eada88977" alt="Version Performance Comparison" width="85%">

<br>

*Version 3's flip-cell algorithm dramatically outperforms earlier implementations*

</div>

<br>

### 📊 Detailed Performance Analysis

<div align="center">

<table>
<tr>
<td align="center" width="50%">

#### Thread Scaling Performance

<img src="https://github.com/user-attachments/assets/4da2da5a-a8e6-4085-8615-cd2999a6513b" alt="Thread Scaling" width="100%">

*Performance across different grid sizes and thread counts*

</td>
<td align="center" width="50%">

#### Early Version Comparison

<img src="https://github.com/user-attachments/assets/d525d320-5a7a-40b1-ae75-868b27c16912" alt="Version Comparison" width="100%">

*V1 vs V2 baseline performance analysis*

</td>
</tr>
</table>

</div>

<br>

### 🔬 Profiling Insights

<div align="center">

<img src="https://github.com/user-attachments/assets/769520b2-53fc-43e7-8c8c-4968f6d555cf" alt="Go Profiler Analysis" width="85%">

<br>

*Go pprof analysis showing CPU time distribution between workers and distributor*

</div>

<br>

### 📋 Development Evolution

| Version | Approach | Key Improvement | Speedup |
|:-------:|:---------|:----------------|:-------:|
| **V1** | Basic parallel workers with neighbor counting | Baseline implementation | 1.0× |
| **V2** | Enhanced I/O operations | Improved file handling | 1.2× |
| **V3** | Flip-cell neighbor tracking | Reduced computation scope | **8.8×** 🔥 |
| **V4** | Efficient slice usage | Reduced GC pressure | 9.1× |
| **Final** | Optimized goroutine coordination | Channel-based sync | **9.5×** 🚀 |

<br>

---

## 🎯 Key Design Advantages

<img align="right" src="https://upload.wikimedia.org/wikipedia/commons/1/1c/Game_of_life_loaf.svg" width="100" alt="Loaf Pattern">

<table>
<tr>
<td width="33%" valign="top">

### 🛡️ High Fault Tolerance

- **Auto Worker Recovery**
  - Dead workers detected automatically
  - Seamless removal from pool
  
- **Turn Retry Mechanism**
  - Failed turns re-executed
  - No data loss guaranteed

- **Graceful Scaling**
  - Add workers during execution
  - Zero downtime operations

</td>
<td width="33%" valign="top">

### ⚡ Minimal Overhead

- **Single World Transfer**
  - Full board sent only once
  - Subsequent: flip-cells only

- **Efficient Halo Exchange**
  - Workers get only needed rows
  - Boundary lookups optimized

- **Persistent TCP Stream**
  - No per-turn handshakes
  - Real-time cell counts

</td>
<td width="33%" valign="top">

### 🔒 Privacy & Security

- **Address Isolation**
  - Client knows only broker
  - Workers completely hidden

- **Proactive Connections**
  - Workers dial broker
  - No exposed ports needed

- **Zero Knowledge**
  - Servers isolated from each other
  - Complete decoupling

</td>
</tr>
</table>

<br>

---

## 🚀 Quick Start

> [!CAUTION]
> Make sure you have **Go 1.17+** and **SDL2** installed before running the project!

### Prerequisites

```bash
# Required
- Go 1.17 or later
- SDL2 library (for visualization)
```

### Installation

```bash
# Clone the repository
git clone https://github.com/Jeff-zhang921/COMS20008Game-of-Life.git
cd COMS20008Game-of-Life
```

### Running the Parallel Version

```bash
cd Parallel
go run . -t 8 -w 512 -h 512
```

### Running the Distributed Version

> [!WARNING]
> For distributed mode, ensure your AWS security groups allow the required ports and all nodes can communicate!

```bash
cd Distributed

# Terminal 1: Start the broker (on AWS or local)
go run broker/broker.go

# Terminal 2-N: Start workers (on each AWS node)
go run server/server.go

# Terminal N+1: Run the client
go run . -t 8 -w 512 -h 512
```

<br>

### ⌨️ Keyboard Controls

<div align="center">

| Key | Action | Parallel | Distributed |
|:---:|:-------|:--------:|:-----------:|
| <kbd>S</kbd> | Save current state as PGM | ✅ | ✅ |
| <kbd>Q</kbd> | Save and quit gracefully | ✅ | ✅ |
| <kbd>P</kbd> | Pause/Resume execution | ✅ | ✅ |
| <kbd>K</kbd> | Terminate all workers | ❌ | ✅ |

</div>

<br>

### 🎛️ Command Line Flags

| Flag | Description | Default |
|:----:|:------------|:-------:|
| `-t` | Number of worker threads | `8` |
| `-w` | Width of the grid | `512` |
| `-h` | Height of the grid | `512` |
| `-turns` | Number of turns to simulate | `100000000` |
| `-headless` | Disable SDL visualization | `false` |

<br>

---

## 📁 Project Structure

```
🎮 COMS20008Game-of-Life/
│
├── 🔄 Parallel/                    # Single-machine parallel implementation
│   ├── gol/                        # Core game logic
│   │   ├── distributor.go          # Main coordinator
│   │   └── gol.go                  # Game rules & worker management
│   ├── sdl/                        # SDL2 visualization
│   ├── images/                     # Input PGM files
│   ├── out/                        # Output directory
│   └── tests/                      # Test suite
│
├── 🌐 Distributed/                 # Multi-machine distributed implementation
│   ├── gol/                        # Client-side distributor
│   ├── broker/                     # Central coordinator (AWS)
│   ├── server/                     # Worker nodes (AWS)
│   ├── stubs/                      # RPC definitions
│   ├── sdl/                        # SDL2 visualization
│   ├── images/                     # Input PGM files
│   ├── out/                        # Output directory
│   └── tests/                      # Test suite
│
├── 📄 docs/                        # Documentation
│   ├── GOL.pptx                    # Presentation slides
│   └── report.pdf                  # Technical report
│
└── 📜 LICENSE                      # MIT License
```

<br>

---

## 🧪 Testing

> [!TIP]
> Always run tests with `-race` flag to detect potential race conditions in concurrent code!

Run the complete test suite with race detection:

```bash
# Parallel tests
cd Parallel
go test -v -race ./...

# Distributed tests
cd Distributed
go test -v -race ./...
```

<br>

---

## 💡 Key Learnings

<div align="center">

> *"The flip-cell optimization taught us that clever algorithms can outperform brute-force parallelism"*

| Insight | Description |
|:--------|:------------|
| 🎯 **Parallelism Has Costs** | Coordination overhead can negate benefits for small workloads |
| 📡 **Network Latency Matters** | Distributed systems require careful architectural decisions |
| 🛡️ **Fault Tolerance Works** | With proper design, systems gracefully handle node failures |
| ⚡ **Algorithm > Hardware** | The flip-cell optimization outperforms adding more threads |

</div>

<div align="center">

![Pulsar](https://upload.wikimedia.org/wikipedia/commons/0/07/Game_of_life_pulsar.gif)

*The Pulsar — a period-3 oscillator, one of the most common oscillators in Game of Life*

</div>

<br>

---

## 📚 Documentation

> [!NOTE]
> 📖 Check out our detailed technical report for in-depth analysis of the implementation!

- 📖 [Official GOL Documentation](https://uob-csa.github.io/gol-docs/)
- 📊 [Project Presentation](docs/GOL.pptx)
- 📝 [Technical Report](docs/report.pdf)

<br>

---

<div align="center">

## 📄 License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

<br>

### 🤝 Contributing

We welcome contributions! Feel free to:
- ⭐ Star this repository
- 🐛 Report bugs
- 💡 Suggest new features
- 🔀 Submit pull requests

<br>

---

<br>

**⭐ Star this repo if you found it helpful!**

<br>

![Typing SVG](https://readme-typing-svg.herokuapp.com?font=Fira+Code&size=18&duration=3000&pause=1000&color=FF6B6B&center=true&vCenter=true&width=500&lines=Thanks+for+visiting!+%F0%9F%8E%AE;Made+with+%E2%9D%A4%EF%B8%8F+at+University+of+Bristol;Go+is+awesome+for+concurrency!;Star+%E2%AD%90+if+you+like+it!)



![Footer](https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=2,3,4,6,24,30&height=120&section=footer)

</div>
