# Durable Execution Engine

A Go-based durable workflow engine that provides crash recovery, step memoization, and replay-safe execution for multi-step business processes.

## Features

- Automatic step memoization
- Crash recovery and resume
- Parallel step execution with `errgroup`
- Generic `Step[T any]` API
- SQLite-backed persistence
- Thread-safe step key generation
- Zombie `RUNNING` step recovery
- Replay-safe behavior for parallel workflows

## How It Works

Each workflow runs with a durable `Context`. Every call to `Step(...)`:

1. Builds a deterministic step key
2. Checks SQLite for a cached completed result
3. Replays the cached result if it already exists
4. Marks the step as `RUNNING` before execution
5. Stores the result as `COMPLETED` after success

If the process crashes after a step starts but before it completes, that step remains `RUNNING`. On the next run, the engine detects it and re-executes it.

## Step Keys and Replay

Step keys use this format:

```text
<step_id>#<sequence_number>
```

Sequence numbers are tracked per `stepID`, not globally. That matters because:

- Repeated steps in loops get unique keys
- Conditional branches stay stable across reruns
- Parallel goroutines can replay safely even if scheduling order changes

Example:

```text
create_employee#1
provision_laptop#1
provision_access#1
send_welcome_email#1
```

## Project Structure

```text
SafeStep/
|-- main.go
|-- go.mod
|-- go.sum
|-- README.md
|-- Prompts.txt
|-- engine/
|   |-- context.go
|   |-- persistence.go
|   |-- step.go
|   `-- workflow.go
|-- examples/
|   `-- onboarding/
|       `-- workflow.go
`-- tests/
    |-- engine_test.go
    `-- onboarding_test.go
```

## Core Components

- `engine/context.go`
  Handles workflow-scoped state and deterministic per-step sequencing.

- `engine/step.go`
  Implements the generic durable step primitive with memoization and replay.

- `engine/persistence.go`
  Stores workflow and step state in SQLite.

- `engine/workflow.go`
  Runs a workflow function with a durable context.

- `examples/onboarding/workflow.go`
  Demo workflow showing sequential and parallel steps plus reusable step helpers.

## Running the Project

From the project root:

```bash
go mod download
go run .
```

The CLI provides these options:

1. Run workflow normally
2. Simulate crash after Step 1
3. Simulate crash after parallel steps
4. Reset workflow database

The demo uses `workflows.db` in the project root.

## Running Tests

Run all tests:

```bash
go test ./... -v
```

Run a clean test pass without cache:

```bash
go test ./... -count=1 -timeout 3m -v
```

Run static checks:

```bash
go vet ./...
```

Optional race check:

```bash
go test ./... -race -v
```

## Crash Recovery Demo

To verify durability manually:

1. Run `go run .`
2. Choose `4` to reset the database
3. Run `go run .` again
4. Choose `3` to simulate a crash after the parallel steps
5. Run `go run .` again
6. Choose `1` to resume normally

Expected behavior:

- `create_employee`, `provision_laptop`, and `provision_access` replay from cache
- Only `send_welcome_email` executes on the final run

## Test Coverage

Current automated coverage includes:

- Step memoization across reruns
- Sequence tracking for repeated loop steps
- Step failure propagation
- Resume after partial workflow completion
- Replay correctness when parallel scheduling changes
- Zombie `RUNNING` step recovery
- Retry after JSON serialization failure
- Repeated use of the same `stepID` in parallel
- Invalid database path handling
- End-to-end onboarding workflow execution

## Current Design Notes

### Zombie Step Recovery

If a crash happens after `SaveStepStart(...)` but before `SaveStepComplete(...)`, the step remains `RUNNING`. On replay, the engine treats that as an interrupted step and executes it again.

This gives at-least-once semantics for interrupted steps, so step functions should be idempotent when they perform external side effects.

### Concurrency Model

- Per-step sequence counters are protected by a mutex
- SQLite writes are protected by a mutex
- Parallel workflow branches are coordinated with `errgroup`

### Persistence Model

Each step record stores:

- `workflow_id`
- `step_key`
- `status`
- `output`

Statuses used by the engine:

- `RUNNING`
- `COMPLETED`
- `FAILED`

## Example Integration Pattern

This engine is a good fit for real product workflows such as onboarding, billing setup, provisioning, notifications, and webhook handling.

Example:

```go
runner, err := engine.NewWorkflowRunner("./workflows.db")
if err != nil {
    return err
}
defer runner.Close()

err = runner.Run("signup-user-123", func(ctx *engine.Context) error {
    _, err := engine.Step(ctx, "create_user", func() (string, error) {
        return "user-created", nil
    })
    if err != nil {
        return err
    }

    _, err = engine.Step(ctx, "send_email", func() (string, error) {
        return "email-sent", nil
    })
    return err
})
```

## Limitations

This project works well for local durable workflow execution, but it is not yet a full production orchestration system. It does not yet include:

- Built-in retries with backoff
- Workflow input persistence
- Timeout or cancellation handling
- Multi-process worker coordination
- Schema migrations
- Dead-letter queues
- Observability or admin tooling
- PostgreSQL/MySQL backends

## Author

**Manish Sharma**
<<<<<<< HEAD
=======
- LinkedIn: [manishsharma31](https://www.linkedin.com/in/manishsharma31/)
- Email: manishsharmadota@gmail.com
>>>>>>> fa3dcb8d49a0fb32ef2105f9a618a34aa58010d6

- LinkedIn: [manishsharma31](https://www.linkedin.com/in/manishsharma31/)
- Email: `manishsharmadota@gmail.com`
