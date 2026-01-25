package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"durable-execution-engine/engine"
	"durable-execution-engine/examples/onboarding"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════╗")
	fmt.Println("║   Durable Execution Engine - Demo         ║")
	fmt.Println("╚═══════════════════════════════════════════╝")
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

	// Handle reset before initializing runner
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
	fmt.Println("\n⚠️  Will simulate crash after Step 1")
	fmt.Println("Run the program again to see durability in action!")

	wrapper := func(ctx *engine.Context) error {
		_, err := engine.Step(ctx, "create_employee", func() (interface{}, error) {
			fmt.Println("  📝 Creating employee record...")
			return "Employee created", nil
		})
		if err != nil {
			return err
		}

		fmt.Println("\n💥 SIMULATED CRASH - Process terminated")
		os.Exit(0)
		return nil
	}

	runner.Run(workflowID, wrapper)
}

func runWithCrashAfterParallel(runner *engine.WorkflowRunner, workflowID string) {
	fmt.Println("\n⚠️  Will simulate crash after parallel steps")
	fmt.Println("Run the program again to see durability in action!")

	wrapper := func(ctx *engine.Context) error {
		_, err := engine.Step(ctx, "create_employee", func() (interface{}, error) {
			fmt.Println("  📝 Creating employee record...")
			return "Employee created", nil
		})
		if err != nil {
			return err
		}

		_, err = engine.Step(ctx, "provision_laptop", func() (string, error) {
			fmt.Println("  💻 Provisioning laptop...")
			return "MacBook Pro 16\"", nil
		})
		if err != nil {
			return err
		}

		_, err = engine.Step(ctx, "provision_access", func() (string, error) {
			fmt.Println("  🔑 Provisioning system access...")
			return "Access granted", nil
		})
		if err != nil {
			return err
		}

		fmt.Println("\n💥 SIMULATED CRASH - Process terminated")
		os.Exit(0)
		return nil
	}

	runner.Run(workflowID, wrapper)
}

func resetWorkflow() {
	dbPath := "./workflows.db"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("ℹ️  No workflow database found to reset.")
		return
	}

	if err := os.Remove(dbPath); err != nil {
		fmt.Printf("❌ Failed to delete database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Workflow database reset successfully")
	fmt.Println("Run the program again to start fresh")
}
