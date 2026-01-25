package onboarding

import (
	"fmt"
	"time"

	"durable-execution-engine/engine"

	"golang.org/x/sync/errgroup"
)

type Employee struct {
	ID    string
	Name  string
	Email string
}

func EmployeeOnboardingWorkflow(ctx *engine.Context) error {
	employee, err := engine.Step(ctx, "create_employee", func() (Employee, error) {
		fmt.Println("  📝 Creating employee record in database...")
		time.Sleep(1 * time.Second)
		return Employee{
			ID:    "EMP-001",
			Name:  "Manish Sharma",
			Email: "manish.sharma@zeotap.com",
		}, nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("  Employee created: %s (%s)\n\n", employee.Name, employee.ID)

	g := new(errgroup.Group)

	var laptop string
	g.Go(func() error {
		var stepErr error
		laptop, stepErr = engine.Step(ctx, "provision_laptop", func() (string, error) {
			fmt.Println("  💻 Provisioning laptop...")
			time.Sleep(2 * time.Second)
			return "MacBook Pro 16\" - Serial: MB12345", nil
		})
		return stepErr
	})

	var access string
	g.Go(func() error {
		var stepErr error
		access, stepErr = engine.Step(ctx, "provision_access", func() (string, error) {
			fmt.Println("  🔑 Provisioning system access...")
			time.Sleep(2 * time.Second)
			return "Access granted: Email, Slack, GitHub, AWS", nil
		})
		return stepErr
	})

	if err := g.Wait(); err != nil {
		return err
	}
	fmt.Printf("  Laptop: %s\n", laptop)
	fmt.Printf("  Access: %s\n\n", access)

	emailResult, err := engine.Step(ctx, "send_welcome_email", func() (string, error) {
		fmt.Println("  📧 Sending welcome email...")
		time.Sleep(1 * time.Second)
		return fmt.Sprintf("Welcome email sent to %s", employee.Email), nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("  %s\n", emailResult)

	return nil
}
