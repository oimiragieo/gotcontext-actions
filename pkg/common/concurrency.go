package common

import (
	"context"
	"sync"
)

type concurrencyKey struct{}

type concurrencySlot struct {
	cancel context.CancelFunc
}

// ConcurrencyController coordinates cancel-in-progress groups for one act run.
type ConcurrencyController struct {
	mu     sync.Mutex
	groups map[string]*concurrencySlot
}

// NewConcurrencyController creates an empty controller.
func NewConcurrencyController() *ConcurrencyController {
	return &ConcurrencyController{groups: map[string]*concurrencySlot{}}
}

// Acquire cancels any prior run in the same group when cancelInProgress is true,
// then registers cancel for the current context. Call the returned release when done.
func (c *ConcurrencyController) Acquire(ctx context.Context, group string, cancelInProgress bool) (context.Context, context.CancelFunc) {
	if c == nil || group == "" {
		return ctx, func() {}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if cancelInProgress {
		if prev := c.groups[group]; prev != nil {
			prev.cancel()
		}
	}
	runCtx, cancel := context.WithCancel(ctx)
	slot := &concurrencySlot{cancel: cancel}
	c.groups[group] = slot
	return runCtx, func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		// Only clear if we still own the slot (a newer Acquire may have replaced us).
		if c.groups[group] == slot {
			delete(c.groups, group)
		}
		cancel()
	}
}

// WithConcurrencyController stores the controller on the context.
func WithConcurrencyController(ctx context.Context, cc *ConcurrencyController) context.Context {
	return context.WithValue(ctx, concurrencyKey{}, cc)
}

// GetConcurrencyController retrieves the controller if present.
func GetConcurrencyController(ctx context.Context) *ConcurrencyController {
	if v, ok := ctx.Value(concurrencyKey{}).(*ConcurrencyController); ok {
		return v
	}
	return nil
}
