# Durable Execution Engine

## Overview
A native Go implementation of a durable execution engine that provides crash-resilience and automatic memoization for workflow-based applications.

## Project Structure
```
├── main.go                    # CLI entry point with interactive demo
├── go.mod                     # Go module definition
├── README.md                  # Project documentation
├── Prompts.txt               # AI prompts used
├── engine/
│   ├── context.go            # Durable Context with sequence tracking
│   ├── step.go               # Generic Step primitive with memoization
│   ├── persistence.go        # SQLite persistence layer
│   └── workflow.go           # Workflow runner orchestrator
├── examples/
│   └── onboarding/
│       └── workflow.go       # Employee onboarding example workflow
└── tests/
    ├── engine_test.go        # Core engine tests
    └── onboarding_test.go    # Workflow integration tests
```

## Key Features
- Automatic step memoization
- Crash recovery and resume
- Parallel execution support (goroutines + errgroup)
- Type-safe generics for step results
- SQLite persistence layer (pure Go, no CGO)
- Thread-safe concurrent operations
- Zombie step handling

## Running the Application
```bash
go run main.go
```

## Running Tests
```bash
go test ./tests/... -v
```

## Recent Changes
- 2026-01-24: Initial implementation with full feature set
  - Persistence layer with SQLite (modernc.org/sqlite)
  - Durable Context with atomic sequence tracking
  - Generic Step function with memoization
  - Employee onboarding workflow example
  - CLI with crash simulation options
  - Comprehensive test suite

## Architecture Decisions
1. **Pure Go SQLite**: Using modernc.org/sqlite instead of go-sqlite3 to avoid CGO dependencies
2. **Atomic Sequence Counter**: Thread-safe incrementing using sync/atomic
3. **Mutex Protection**: Database operations protected with sync.Mutex
4. **JSON Serialization**: Standard library encoding/json for step results
5. **errgroup**: Parallel step coordination for concurrent execution
