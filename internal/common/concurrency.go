package common

import (
	"context"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

type concurrencyKey struct{}

type concurrencyHolder struct {
	cancel context.CancelFunc
}

type concurrencyGroup struct {
	holders []*concurrencyHolder
}

// ConcurrencyController coordinates cancel-in-progress and queue groups for one act run.
type ConcurrencyController struct {
	mu     sync.Mutex
	cond   *sync.Cond
	groups map[string]*concurrencyGroup
}

// NewConcurrencyController creates an empty controller.
func NewConcurrencyController() *ConcurrencyController {
	c := &ConcurrencyController{groups: map[string]*concurrencyGroup{}}
	c.cond = sync.NewCond(&c.mu)
	return c
}

func normalizeQueue(queue string) string {
	switch strings.ToLower(strings.TrimSpace(queue)) {
	case "", "single":
		return "single"
	case "max":
		return "max"
	default:
		log.Warnf("unknown concurrency queue %q; using single", queue)
		return "single"
	}
}

// Acquire registers this run in group.
// cancelInProgress cancels whoever currently holds the group.
// queue "max" lets jobs in the group run together; "single" (and any unknown value) waits until the group is free.
// Call the returned release when the job finishes.
func (c *ConcurrencyController) Acquire(ctx context.Context, group string, cancelInProgress bool, queue string) (context.Context, context.CancelFunc) {
	if c == nil || group == "" {
		return ctx, func() {}
	}
	mode := normalizeQueue(queue)

	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-ctx.Done():
			c.mu.Lock()
			c.cond.Broadcast()
			c.mu.Unlock()
		case <-stop:
		}
	}()

	c.mu.Lock()
	for {
		if err := ctx.Err(); err != nil {
			c.mu.Unlock()
			return ctx, func() {}
		}
		g := c.groups[group]
		if g == nil {
			g = &concurrencyGroup{}
			c.groups[group] = g
		}
		if cancelInProgress {
			for _, h := range g.holders {
				h.cancel()
			}
			g.holders = nil
		}
		if mode == "max" || len(g.holders) == 0 {
			runCtx, cancel := context.WithCancel(ctx)
			h := &concurrencyHolder{cancel: cancel}
			g.holders = append(g.holders, h)
			c.mu.Unlock()
			return runCtx, func() {
				c.release(group, h)
			}
		}
		c.cond.Wait()
	}
}

func (c *ConcurrencyController) release(group string, h *concurrencyHolder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	g := c.groups[group]
	h.cancel()
	if g == nil {
		c.cond.Broadcast()
		return
	}
	next := g.holders[:0]
	for _, cur := range g.holders {
		if cur != h {
			next = append(next, cur)
		}
	}
	g.holders = next
	if len(g.holders) == 0 {
		delete(c.groups, group)
	}
	c.cond.Broadcast()
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
