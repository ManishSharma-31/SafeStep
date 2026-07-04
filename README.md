# Durable Execution Engine

A native Go implementation of a durable execution engine that provides crash-resilience and automatic memoization for workflow-based applications.

## Features

- Automatic step memoization
- Crash recovery and resume
- Parallel execution support (goroutines + errgroup)
- Type-safe generics for step results
- SQLite persistence layer
- Thread-safe concurrent operations
- Zombie step handling

## How It Works

### Sequence Tracking
The engine tracks invocation counts per step ID to generate deterministic sequence suffixes. This ensures that:
- Steps in loops are uniquely identified
- Conditional branches don't cause conflicts
- Parallel steps replay correctly even if goroutine scheduling differs across runs

Step keys format: `<step_id>#<sequence_number>`

### Thread Safety
Parallel execution is safe through:
1. Mutex-protected per-step sequence tracking
2. Mutex-protected database operations
3. Transaction-based state updates

## Running the Demo

```bash
# Install dependencies
go mod download

# Run normally
go run main.go

# Run tests
go test ./tests/... -v
```

## Crash Recovery Demo

1. Choose option 2 or 3 to simulate a crash
2. Run the program again
3. Observe that completed steps are skipped
4. Workflow resumes from the point of failure

## Architecture

- `engine/context.go` - Durable context with sequence tracking
- `engine/step.go` - Generic step primitive with memoization
- `engine/persistence.go` - SQLite storage layer
- `engine/workflow.go` - Workflow runner orchestrator
- `examples/onboarding/` - Employee onboarding demo workflow

## Design Decisions

### Zombie Step Problem
If a crash occurs between step execution and database commit, the step is marked as RUNNING. On restart:
- The engine detects the RUNNING status
- Re-executes the step (at-least-once semantics)
- Updates status to COMPLETED on success

### Concurrency Model
- errgroup for parallel step coordination
- Mutex-protected DB writes prevent race conditions
- Atomic sequence counter ensures deterministic replay

## Project Structure

```
durable-execution-engine/
├── main.go                    # CLI entry point
├── go.mod                     # Go module definition
├── README.md                  # Project documentation
├── Prompts.txt               # AI prompts used
├── engine/
│   ├── context.go            # Durable Context implementation
│   ├── step.go               # Step primitive with generics
│   ├── persistence.go        # SQLite database layer
│   └── workflow.go           # Workflow runner
├── examples/
│   └── onboarding/
│       └── workflow.go       # Employee onboarding example
└── tests/
    ├── engine_test.go        # Core engine tests
    └── onboarding_test.go    # Workflow integration tests
```

## Key Implementation Notes

1. **Deterministic Step Sequencing**: Tracks invocation counts per step ID for replay-safe key generation
2. **Mutex Protection**: Database operations are protected with `sync.Mutex`
3. **Generic Step Function**: `Step[T any]()` supports any return type
4. **JSON Serialization**: Standard library encoding/json for step results
5. **Error Handling**: Proper error propagation and step failure recording
6. **Idempotency**: Completed steps return cached results on replay


---

## 👤 Author

**Manish Sharma**
- GitHub: [manishsharma31](https://www.linkedin.com/in/manishsharma31/)
- Email: manishsharmadota@gmail.com

---
