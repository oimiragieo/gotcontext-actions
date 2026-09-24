package common

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThenErrorOnlyOnFailure(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	called := false
	ok := Executor(func(_ context.Context) error { return nil }).ThenError(func(_ context.Context, err error) error {
		called = true
		return err
	})
	assert.NoError(ok(ctx))
	assert.False(called, "ThenError must not invoke then on success")

	boom := fmt.Errorf("boom")
	fail := Executor(func(_ context.Context) error { return boom }).ThenError(func(_ context.Context, err error) error {
		called = true
		return fmt.Errorf("wrapped: %w", err)
	})
	err := fail(ctx)
	assert.True(called)
	assert.ErrorContains(err, "wrapped")
}
