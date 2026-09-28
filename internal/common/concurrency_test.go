package common

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrencyController_CancelInProgress(t *testing.T) {
	cc := NewConcurrencyController()
	ctx := context.Background()

	ctx1, rel1 := cc.Acquire(ctx, "deploy", true, "single")
	defer rel1()

	assert.NoError(t, ctx1.Err())

	ctx2, rel2 := cc.Acquire(ctx, "deploy", true, "single")
	defer rel2()

	assert.ErrorIs(t, ctx1.Err(), context.Canceled)
	assert.NoError(t, ctx2.Err())
}

func TestConcurrencyController_SingleWaits(t *testing.T) {
	cc := NewConcurrencyController()
	ctx := context.Background()

	_, rel1 := cc.Acquire(ctx, "deploy", false, "single")

	var order []int
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, rel2 := cc.Acquire(ctx, "deploy", false, "")
		defer rel2()
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
	}()

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	order = append(order, 1)
	mu.Unlock()

	rel1()
	wg.Wait()

	require.Equal(t, []int{1, 2}, order)
}

func TestConcurrencyController_MaxOverlaps(t *testing.T) {
	cc := NewConcurrencyController()
	ctx := context.Background()

	ctx1, rel1 := cc.Acquire(ctx, "deploy", false, "max")
	defer rel1()
	ctx2, rel2 := cc.Acquire(ctx, "deploy", false, "max")
	defer rel2()

	assert.NoError(t, ctx1.Err())
	assert.NoError(t, ctx2.Err())
}

func TestConcurrencyController_UnknownQueueWaits(t *testing.T) {
	cc := NewConcurrencyController()
	ctx := context.Background()

	_, rel1 := cc.Acquire(ctx, "deploy", false, "bogus")
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		close(started)
		_, rel2 := cc.Acquire(ctx, "deploy", false, "nope")
		rel2()
		close(done)
	}()
	<-started
	select {
	case <-done:
		t.Fatal("unknown queue should wait like single")
	case <-time.After(40 * time.Millisecond):
	}
	rel1()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("waiter did not resume")
	}
}

func TestConcurrencyController_IndependentGroups(t *testing.T) {
	cc := NewConcurrencyController()
	ctx := context.Background()

	ctxA, relA := cc.Acquire(ctx, "groupA", false, "single")
	defer relA()
	ctxB, relB := cc.Acquire(ctx, "groupB", false, "single")
	defer relB()

	assert.NoError(t, ctxA.Err())
	assert.NoError(t, ctxB.Err())
}

func TestConcurrencyController_ContextCancelledWhileWaiting(t *testing.T) {
	cc := NewConcurrencyController()
	ctx := context.Background()

	_, rel1 := cc.Acquire(ctx, "group1", false, "single")
	defer rel1()

	waitCtx, waitCancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer waitCancel()

	acquiredCtx, rel2 := cc.Acquire(waitCtx, "group1", false, "single")
	defer rel2()

	assert.ErrorIs(t, acquiredCtx.Err(), context.DeadlineExceeded)
}
