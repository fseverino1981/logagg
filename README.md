# logagg

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-success)](.)
[![Coverage](https://img.shields.io/badge/Coverage-82.4%25-brightgreen)](.)\

**Real-time log aggregator CLI** that monitors multiple log files simultaneously using Go concurrency patterns.

## Overview

`logagg` is a command-line tool designed to monitor and aggregate logs from multiple files in real-time. Built with Go's powerful concurrency primitives, it demonstrates production-ready patterns like Generator, Fan-In, and Pipeline while providing a practical solution for log monitoring.

### Key Features

- **Multi-file monitoring**: Track multiple log files simultaneously
- **Real-time filtering**: Filter logs by pattern as they arrive
- **Tail mode**: Continuously watch for new log entries (like `tail -f`)
- **Graceful shutdown**: Clean termination with `Ctrl+C`
- **Concurrent processing**: Efficient handling using Go channels and goroutines
- **Prefix labeling**: Each log line is tagged with its source file

## Installation

### Prerequisites

- Go 1.21 or higher

### Build from source

```bash
# Clone the repository
git clone https://github.com/fseverino1981/logagg.git
cd logagg

# Build the binary
make build

# Or use Go directly
go build -o logagg .
```

## Usage

### Basic Examples

```bash
# Monitor a single log file
./logagg --files app.log

# Monitor multiple log files
./logagg --files app.log,error.log,access.log

# Filter logs containing "ERROR"
./logagg --files app.log --filter "ERROR"

# Tail mode: continuously watch for new entries
./logagg --files app.log --tail

# Combine filtering and tail mode
./logagg --files app.log,error.log --filter "ERROR" --tail
```

### Command-line Flags

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--files` | `-f` | Comma-separated list of log files to monitor | `-f app.log,error.log` |
| `--filter` | `-F` | Filter logs by pattern (case-sensitive) | `-F "ERROR"` |
| `--tail` | `-t` | Continuously watch for new log entries | `-t` |

### Output Format

Each log line is prefixed with its source file:

```
[app.log] - 2024-01-15 10:23:45 INFO Application started
[error.log] - 2024-01-15 10:23:46 ERROR Database connection failed
[app.log] - 2024-01-15 10:23:47 INFO Retrying connection...
```

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────────┐
│                         logagg CLI                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │    Validator      │
                    │  (Check files)    │
                    └─────────┬─────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
┌───────▼────────┐    ┌───────▼────────┐    ┌──────▼─────────┐
│  Generator     │    │  Generator     │    │  Generator     │
│  (file1.log)   │    │  (file2.log)   │    │  (file3.log)   │
│                │    │                │    │                │
│  chan string───┤    │  chan string───┤    │  chan string───┤
└────────────────┘    └────────────────┘    └────────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │     Fan-In        │
                    │   (Aggregator)    │
                    │                   │
                    │   chan string─────┤
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │     Pipeline      │
                    │     (Filter)      │
                    │                   │
                    │   chan string─────┤
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │      Output       │
                    │   (fmt.Println)   │
                    └───────────────────┘
```

### Concurrency Patterns

#### 1. Generator Pattern

Each log file spawns a goroutine that reads lines and sends them to a channel:

```go
func ReadLines(ctx context.Context, file string, tail bool) <-chan string {
    out := make(chan string)
    go func() {
        defer close(out)
        // Read file line by line
        for scanner.Scan() {
            select {
            case out <- fmt.Sprintf("[%s] - %s", filename, scanner.Text()):
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

**Benefits:**
- Non-blocking I/O operations
- Each file processed independently
- Natural backpressure through channel blocking

#### 2. Fan-In Pattern

Multiplexes multiple channels into a single output channel:

```go
func Aggregate(ctx context.Context, channels ...<-chan string) chan string {
    out := make(chan string)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(ch <-chan string) {
            defer wg.Done()
            for msg := range ch {
                select {
                case out <- msg:
                case <-ctx.Done():
                    return
                }
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

**Benefits:**
- Combines multiple data sources seamlessly
- Maintains message ordering per source
- Automatic cleanup when all sources complete

#### 3. Pipeline Pattern

Transforms data by filtering through a channel:

```go
func Filter(ch <-chan string, filter string) <-chan string {
    out := make(chan string)
    go func() {
        defer close(out)
        for line := range ch {
            if strings.Contains(line, filter) {
                out <- line
            }
        }
    }()
    return out
}
```

**Benefits:**
- Composable data transformations
- Memory efficient (streams data)
- Easy to add more pipeline stages

#### 4. Graceful Shutdown

Uses `context.Context` for coordinated cancellation:

```go
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()

// All goroutines respect context cancellation
select {
case out <- msg:
case <-ctx.Done():
    return
}
```

**Benefits:**
- Clean resource cleanup
- No goroutine leaks
- Immediate shutdown on signal

## Technical Decisions

### Why Channels Over Shared Memory?

**Decision:** Use channels for communication between goroutines.

**Trade-offs:**
- ✅ **Pros:** Eliminates race conditions, enforces sequential access, clear ownership
- ⚠️ **Cons:** Slight overhead vs. shared memory, potential for deadlocks if misused

**Rationale:** Follows Go's philosophy "Don't communicate by sharing memory; share memory by communicating." This makes the code more maintainable and less prone to concurrency bugs.

### Buffered vs. Unbuffered Channels

**Decision:** Use unbuffered channels for backpressure.

**Trade-offs:**
- ✅ **Pros:** Natural flow control, prevents memory bloat from fast producers
- ⚠️ **Cons:** Blocking can slow down fast readers

**Rationale:** Log processing should respect consumer speed. If output can't keep up, we want readers to naturally slow down rather than buffer indefinitely.

### Tail Mode Implementation

**Decision:** Poll files every 500ms in tail mode.

**Trade-offs:**
- ✅ **Pros:** Simple, works across platforms, low resource usage
- ⚠️ **Cons:** 500ms latency, less efficient than inotify/fsnotify

**Rationale:** Simplicity and portability over marginal performance gains. For production use, consider integrating `fsnotify` for event-driven file watching.

### Error Handling Strategy

**Decision:** Log errors and continue processing remaining files.

**Trade-offs:**
- ✅ **Pros:** Resilient to individual file issues, better user experience
- ⚠️ **Cons:** Silent failures if user doesn't check output

**Rationale:** One bad file shouldn't break monitoring of others. Users see errors but processing continues.

## Project Structure

```
logagg/
├── cmd/
│   └── root.go              # Cobra CLI setup and command logic
├── internal/
│   ├── reader/
│   │   ├── reader.go        # File reading (Generator pattern)
│   │   ├── reader_test.go   # Reader tests
│   │   ├── validator.go     # File validation
│   │   └── validator_test.go
│   ├── aggregator/
│   │   ├── aggregator.go    # Channel multiplexing (Fan-In)
│   │   └── aggregator_test.go
│   └── filter/
│       ├── filter.go        # Log filtering (Pipeline)
│       └── filter_test.go
├── main.go                  # Application entry point
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
├── Makefile                 # Build automation
└── README.md                # This file
```

## Testing

The project includes comprehensive unit tests with **82.4% code coverage**.

### Run Tests

```bash
# Run all tests
make test

# Or use Go directly
go test ./... -v

# Run tests with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Test Coverage by Package

| Package | Coverage |
|---------|----------|
| `internal/aggregator` | 100% |
| `internal/filter` | 100% |
| `internal/reader` | 91.7% |
| **Total** | **82.4%** |

### Test Categories

- **Unit tests**: Each package has isolated tests
- **Concurrency tests**: Verify correct channel behavior and context cancellation
- **Edge cases**: Empty files, non-existent files, directories, large volumes
- **Table-driven tests**: Comprehensive validation scenarios

## Development

### Build Commands

```bash
# Build binary
make build

# Run tests
make test

# Clean build artifacts
make clean

# Show help
make help
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - Modern CLI framework

## Future Enhancements

- [ ] Add regex pattern matching for filters
- [ ] Implement file watching with `fsnotify` for better tail performance
- [ ] Add JSON output format option
- [ ] Support for compressed log files (gzip)
- [ ] Add timestamp-based filtering
- [ ] Colorized output for different log levels
- [ ] Configuration file support

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

**Flávio Antonio Severino**
- GitHub: [@fseverino1981](https://github.com/fseverino1981)

## Acknowledgments

- Built as part of a Go concurrency patterns study
- Inspired by Unix tools like `tail` and `multitail`
- Uses Go concurrency patterns from [Go Concurrency Patterns](https://go.dev/blog/pipelines)

---

**Made with ❤️ and Go**
