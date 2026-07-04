package engine

import (
	"fmt"
	"sync"
)

type Context struct {
	workflowID  string
	persistence *PersistenceLayer
	mu          sync.Mutex
	sequences   map[string]int64
}

func NewContext(workflowID string, persistence *PersistenceLayer) *Context {
	ctx := &Context{
		workflowID:  workflowID,
		persistence: persistence,
		sequences:   make(map[string]int64),
	}
	return ctx
}

func (c *Context) WorkflowID() string {
	return c.workflowID
}

func (c *Context) nextSequence(stepID string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sequences[stepID]++
	return c.sequences[stepID]
}

func (c *Context) generateStepKey(stepID string) string {
	seq := c.nextSequence(stepID)
	return fmt.Sprintf("%s#%d", stepID, seq)
}
