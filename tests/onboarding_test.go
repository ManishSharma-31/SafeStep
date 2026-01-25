package tests

import (
	"fmt"
	"os"
	"testing"

	"durable-execution-engine/engine"
	"durable-execution-engine/examples/onboarding"
)

func TestEmployeeOnboardingWorkflow(t *testing.T) {
	dbPath := "./test_onboarding.db"
	defer os.Remove(dbPath)

	runner, err := engine.NewWorkflowRunner(dbPath)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}
	defer runner.Close()

	if err := runner.Run("test-onboarding", onboarding.EmployeeOnboardingWorkflow); err != nil {
		t.Fatalf("Onboarding workflow failed: %v", err)
	}
}

func TestOnboardingResume(t *testing.T) {
	dbPath := "./test_onboarding_resume.db"
	defer os.Remove(dbPath)

	runner, err := engine.NewWorkflowRunner(dbPath)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	stepCount := 0
	partialWorkflow := func(ctx *engine.Context) error {
		_, err := engine.Step(ctx, "create_employee", func() (string, error) {
			stepCount++
			return "Employee created", nil
		})
		return err
	}

	if err := runner.Run("resume-test", partialWorkflow); err != nil {
		t.Fatalf("First run failed: %v", err)
	}
	runner.Close()

	if stepCount != 1 {
		t.Errorf("Expected 1 step execution, got %d", stepCount)
	}

	runner2, err := engine.NewWorkflowRunner(dbPath)
	if err != nil {
		t.Fatalf("Failed to create second runner: %v", err)
	}
	defer runner2.Close()

	fullWorkflow := func(ctx *engine.Context) error {
		_, err := engine.Step(ctx, "create_employee", func() (string, error) {
			stepCount++
			return "Employee created", nil
		})
		if err != nil {
			return err
		}

		_, err = engine.Step(ctx, "provision_laptop", func() (string, error) {
			stepCount++
			return "Laptop provisioned", nil
		})
		return err
	}

	if err := runner2.Run("resume-test", fullWorkflow); err != nil {
		t.Fatalf("Resume run failed: %v", err)
	}

	if stepCount != 2 {
		t.Errorf("Expected 2 step executions (1 cached + 1 new), got %d", stepCount)
	}
}

func init() {
	_ = fmt.Sprintf
}
