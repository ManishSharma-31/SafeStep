package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"durable-execution-engine/engine"
	"durable-execution-engine/examples/onboarding"
	"golang.org/x/sync/errgroup"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  Durable Execution Engine - Demo")
	fmt.Println("===========================================")
	fmt.Println()

	fmt.Println("Options:")
	fmt.Println("  1. Run workflow normally")
	fmt.Println("  2. Simulate crash after Step 1")
	fmt.Println("  3. Simulate crash after parallel steps")
	fmt.Println("  4. Reset workflow (delete database)")
	fmt.Println()
	fmt.Print("Enter choice (1-4): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "4" {
		resetWorkflow()
		return
	}

	runner, err := engine.NewWorkflowRunner("./workflows.db")
	if err != nil {
		fmt.Printf("Failed to initialize runner: %v\n", err)
		os.Exit(1)
	}
	defer runner.Close()

	workflowID := "onboarding-workflow-001"

	switch choice {
	case "1":
		runNormally(runner, workflowID)
	case "2":
		runWithCrashAfterStep1(runner, workflowID)
	case "3":
		runWithCrashAfterParallel(runner, workflowID)
	default:
		fmt.Println("Invalid choice. Running normally.")
		runNormally(runner, workflowID)
	}
}

func runNormally(runner *engine.WorkflowRunner, workflowID string) {
	if err := runner.Run(workflowID, onboarding.EmployeeOnboardingWorkflow); err != nil {
		fmt.Printf("Workflow failed: %v\n", err)
		os.Exit(1)
	}
}

func runWithCrashAfterStep1(runner *engine.WorkflowRunner, workflowID string) {
	fmt.Println("\nWill simulate crash after Step 1")
	fmt.Println("Run the program again to see durability in action!")

	wrapper := func(ctx *engine.Context) error {
		if _, err := onboarding.CreateEmployeeStep(ctx); err != nil {
			return err
		}

		fmt.Println("\nSIMULATED CRASH - Process terminated")
		os.Exit(0)
		return nil
	}

	runner.Run(workflowID, wrapper)
}

func runWithCrashAfterParallel(runner *engine.WorkflowRunner, workflowID string) {
	fmt.Println("\nWill simulate crash after parallel steps")
	fmt.Println("Run the program again to see durability in action!")

	wrapper := func(ctx *engine.Context) error {
		if _, err := onboarding.CreateEmployeeStep(ctx); err != nil {
			return err
		}

		g := new(errgroup.Group)
		g.Go(func() error {
			_, stepErr := onboarding.ProvisionLaptopStep(ctx)
			return stepErr
		})
		g.Go(func() error {
			_, stepErr := onboarding.ProvisionAccessStep(ctx)
			return stepErr
		})
		if err := g.Wait(); err != nil {
			return err
		}

		fmt.Println("\nSIMULATED CRASH - Process terminated")
		os.Exit(0)
		return nil
	}

	runner.Run(workflowID, wrapper)
}

func resetWorkflow() {
	dbPath := "./workflows.db"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("No workflow database found to reset.")
		return
	}

	if err := os.Remove(dbPath); err != nil {
		fmt.Printf("Failed to delete database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Workflow database reset successfully")
	fmt.Println("Run the program again to start fresh")
}
