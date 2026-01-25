package tests

import (
        "fmt"
        "os"
        "testing"

        "durable-execution-engine/engine"
)

func TestStepMemoization(t *testing.T) {
        dbPath := "./test_workflows.db"
        defer os.Remove(dbPath)

        runner, err := engine.NewWorkflowRunner(dbPath)
        if err != nil {
                t.Fatalf("Failed to create runner: %v", err)
        }
        defer runner.Close()

        executionCount := 0
        workflowFunc := func(ctx *engine.Context) error {
                _, err := engine.Step(ctx, "test_step", func() (string, error) {
                        executionCount++
                        return "result", nil
                })
                return err
        }

        if err := runner.Run("test-workflow", workflowFunc); err != nil {
                t.Fatalf("First run failed: %v", err)
        }

        if executionCount != 1 {
                t.Errorf("Expected 1 execution, got %d", executionCount)
        }

        runner2, err := engine.NewWorkflowRunner(dbPath)
        if err != nil {
                t.Fatalf("Failed to create second runner: %v", err)
        }
        defer runner2.Close()

        if err := runner2.Run("test-workflow", workflowFunc); err != nil {
                t.Fatalf("Second run failed: %v", err)
        }

        if executionCount != 1 {
                t.Errorf("Expected step to be memoized, but it executed again. Count: %d", executionCount)
        }
}

func TestSequenceTracking(t *testing.T) {
        dbPath := "./test_sequence.db"
        defer os.Remove(dbPath)

        runner, err := engine.NewWorkflowRunner(dbPath)
        if err != nil {
                t.Fatalf("Failed to create runner: %v", err)
        }
        defer runner.Close()

        workflowFunc := func(ctx *engine.Context) error {
                for i := 0; i < 3; i++ {
                        _, err := engine.Step(ctx, "loop_step", func() (int, error) {
                                return i, nil
                        })
                        if err != nil {
                                return err
                        }
                }
                return nil
        }

        if err := runner.Run("sequence-test", workflowFunc); err != nil {
                t.Fatalf("Workflow failed: %v", err)
        }
}

func TestStepFailure(t *testing.T) {
        dbPath := "./test_failure.db"
        defer os.Remove(dbPath)

        runner, err := engine.NewWorkflowRunner(dbPath)
        if err != nil {
                t.Fatalf("Failed to create runner: %v", err)
        }
        defer runner.Close()

        workflowFunc := func(ctx *engine.Context) error {
                _, err := engine.Step(ctx, "failing_step", func() (string, error) {
                        return "", fmt.Errorf("intentional failure")
                })
                return err
        }

        err = runner.Run("failure-test", workflowFunc)
        if err == nil {
                t.Error("Expected workflow to fail, but it succeeded")
        }
}
