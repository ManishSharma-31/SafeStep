package engine

import (
	"encoding/json"
	"fmt"
)

func Step[T any](ctx *Context, stepID string, fn func() (T, error)) (T, error) {
	var zero T

	stepKey := ctx.generateStepKey(stepID)

	record, err := ctx.persistence.GetStep(ctx.workflowID, stepKey)
	if err != nil {
		return zero, fmt.Errorf("failed to check step: %w", err)
	}

	if record != nil && record.Status == StatusCompleted {
		fmt.Printf("✓ Step '%s' already completed (replaying cached result)\n", stepKey)
		var result T
		if err := json.Unmarshal([]byte(record.Output), &result); err != nil {
			return zero, fmt.Errorf("failed to deserialize cached result: %w", err)
		}
		return result, nil
	}

	if record != nil && record.Status == StatusRunning {
		fmt.Printf("⚠ Step '%s' was running but didn't complete - re-executing\n", stepKey)
	}

	if err := ctx.persistence.SaveStepStart(ctx.workflowID, stepKey); err != nil {
		return zero, fmt.Errorf("failed to save step start: %w", err)
	}

	fmt.Printf("→ Executing step '%s'...\n", stepKey)
	result, err := fn()
	if err != nil {
		ctx.persistence.SaveStepFailed(ctx.workflowID, stepKey, err.Error())
		return zero, fmt.Errorf("step '%s' failed: %w", stepKey, err)
	}

	if err := ctx.persistence.SaveStepComplete(ctx.workflowID, stepKey, result); err != nil {
		return zero, fmt.Errorf("failed to save step result: %w", err)
	}

	fmt.Printf("✓ Step '%s' completed\n", stepKey)
	return result, nil
}
