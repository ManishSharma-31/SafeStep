package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"durable-execution-engine/engine"
	"golang.org/x/sync/errgroup"
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

func TestParallelReplayUsesCachedResultsAcrossSchedulingChanges(t *testing.T) {
	dbPath := "./test_parallel_replay.db"
	defer os.Remove(dbPath)

	runner, err := engine.NewWorkflowRunner(dbPath)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	firstRun := func(ctx *engine.Context) error {
		if _, err := engine.Step(ctx, "create_employee", func() (string, error) {
			return "Employee created", nil
		}); err != nil {
			return err
		}

		if _, err := engine.Step(ctx, "provision_laptop", func() (string, error) {
			return "Laptop provisioned", nil
		}); err != nil {
			return err
		}

		if _, err := engine.Step(ctx, "provision_access", func() (string, error) {
			return "Access granted", nil
		}); err != nil {
			return err
		}

		return nil
	}

	if err := runner.Run("parallel-replay", firstRun); err != nil {
		t.Fatalf("First run failed: %v", err)
	}
	runner.Close()

	executedAgain := make(map[string]int)
	runner2, err := engine.NewWorkflowRunner(dbPath)
	if err != nil {
		t.Fatalf("Failed to create second runner: %v", err)
	}
	defer runner2.Close()

	secondRun := func(ctx *engine.Context) error {
		if _, err := engine.Step(ctx, "create_employee", func() (string, error) {
			executedAgain["create_employee"]++
			return "Employee created", nil
		}); err != nil {
			return err
		}

		g := new(errgroup.Group)
		g.Go(func() error {
			_, err := engine.Step(ctx, "provision_access", func() (string, error) {
				executedAgain["provision_access"]++
				return "Access granted", nil
			})
			return err
		})
		g.Go(func() error {
			_, err := engine.Step(ctx, "provision_laptop", func() (string, error) {
				executedAgain["provision_laptop"]++
				return "Laptop provisioned", nil
			})
			return err
		})

		return g.Wait()
	}

	if err := runner2.Run("parallel-replay", secondRun); err != nil {
		t.Fatalf("Second run failed: %v", err)
	}

	for stepID, count := range executedAgain {
		if count != 0 {
			t.Fatalf("Expected cached replay for %s, but step executed %d additional time(s)", stepID, count)
		}
	}
}

func TestZombieRunningStepIsReExecuted(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "zombie.db")

	persistence, err := engine.NewPersistence(dbPath)
	if err != nil {
		t.Fatalf("Failed to create persistence: %v", err)
	}

	workflowID := "zombie-recovery"
	if err := persistence.CreateWorkflow(workflowID); err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}
	if err := persistence.SaveStepStart(workflowID, "stuck_step#1"); err != nil {
		t.Fatalf("Failed to seed running step: %v", err)
	}
	if err := persistence.Close(); err != nil {
		t.Fatalf("Failed to close seeded persistence: %v", err)
	}

	runner, err := engine.NewWorkflowRunner(dbPath)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}
	defer runner.Close()

	executionCount := 0
	workflowFunc := func(ctx *engine.Context) error {
		_, err := engine.Step(ctx, "stuck_step", func() (string, error) {
			executionCount++
			return "recovered", nil
		})
		return err
	}

	if err := runner.Run(workflowID, workflowFunc); err != nil {
		t.Fatalf("Recovery run failed: %v", err)
	}

	if executionCount != 1 {
		t.Fatalf("Expected zombie step to be re-executed once, got %d", executionCount)
	}
}

func TestSerializationFailureCanBeRetried(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "serialization_retry.db")

	runner, err := engine.NewWorkflowRunner(dbPath)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}
	defer runner.Close()

	executionCount := 0
	workflowFunc := func(ctx *engine.Context) error {
		_, err := engine.Step(ctx, "serialize_once", func() (map[string]any, error) {
			executionCount++
			if executionCount == 1 {
				return map[string]any{"bad": make(chan int)}, nil
			}
			return map[string]any{"status": "ok"}, nil
		})
		return err
	}

	if err := runner.Run("serialization-retry", workflowFunc); err == nil {
		t.Fatal("Expected first run to fail because the step output is not JSON serializable")
	}

	if err := runner.Run("serialization-retry", workflowFunc); err != nil {
		t.Fatalf("Expected retry to succeed, got: %v", err)
	}

	if executionCount != 2 {
		t.Fatalf("Expected step to execute twice across failure and retry, got %d", executionCount)
	}
}

func TestConcurrentRepeatedStepIDUsesUniqueKeys(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "same_step_parallel.db")

	runner, err := engine.NewWorkflowRunner(dbPath)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}
	defer runner.Close()

	executionCount := 0
	workflowFunc := func(ctx *engine.Context) error {
		g := new(errgroup.Group)
		for i := 0; i < 8; i++ {
			g.Go(func() error {
				_, err := engine.Step(ctx, "parallel_same_step", func() (string, error) {
					executionCount++
					return "ok", nil
				})
				return err
			})
		}
		return g.Wait()
	}

	if err := runner.Run("same-step-parallel", workflowFunc); err != nil {
		t.Fatalf("Parallel same-step workflow failed: %v", err)
	}

	if executionCount != 8 {
		t.Fatalf("Expected all parallel repeated steps to execute once, got %d", executionCount)
	}
}

func TestNewWorkflowRunnerInvalidPath(t *testing.T) {
	invalidPath := filepath.Join(t.TempDir(), "nested", "missing", "workflows.db")

	runner, err := engine.NewWorkflowRunner(invalidPath)
	if err == nil {
		runner.Close()
		t.Fatalf("Expected invalid database path to fail: %s", invalidPath)
	}
}
