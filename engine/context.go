package engine

import (
	"fmt"
	"sync/atomic"
)

type Context struct {
	workflowID  string
	persistence *PersistenceLayer
	sequence    *atomic.Int64
}

func NewContext(workflowID string, persistence *PersistenceLayer) *Context {
	ctx := &Context{
		workflowID:  workflowID,
		persistence: persistence,
		sequence:    &atomic.Int64{},
	}
	return ctx
}

func (c *Context) WorkflowID() string {
	return c.workflowID
}

func (c *Context) nextSequence() int64 {
	return c.sequence.Add(1)
}

func (c *Context) generateStepKey(stepID string) string {
	seq := c.nextSequence()
	return fmt.Sprintf("%s#%d", stepID, seq)
}
