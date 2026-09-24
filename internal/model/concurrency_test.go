package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConcurrencyUnmarshalScalarAndMapping(t *testing.T) {
	var scalar Concurrency
	require.NoError(t, yaml.Unmarshal([]byte("my-group"), &scalar))
	assert.Equal(t, "my-group", scalar.Group)
	assert.False(t, scalar.CancelInProgress)

	var mapping Concurrency
	require.NoError(t, yaml.Unmarshal([]byte("group: g1\ncancel-in-progress: true\nqueue: single\n"), &mapping))
	assert.Equal(t, "g1", mapping.Group)
	assert.True(t, mapping.CancelInProgress)
	assert.Equal(t, "single", mapping.Queue)

	var wf Workflow
	require.NoError(t, yaml.Unmarshal([]byte("name: t\non: push\nconcurrency: ci-group\njobs: {}\n"), &wf))
	require.NotNil(t, wf.Concurrency)
	assert.Equal(t, "ci-group", wf.Concurrency.Group)
}
