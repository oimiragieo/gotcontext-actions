package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/oimiragieo/gotcontext-actions/internal/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintStepSummariesWritesFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "summary.md")
	rc := &RunContext{
		Config: &Config{StepSummaryFile: out},
		stepSummaries: map[string]string{
			"step-a": "hello summary",
		},
	}
	ctx := common.WithLogger(context.Background(), logrus.New())
	printStepSummaries(ctx, rc)
	body, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Contains(t, string(body), "step-a")
	assert.Contains(t, string(body), "hello summary")
}
