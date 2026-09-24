package oidc

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockServerToken(t *testing.T) {
	s, err := Start(map[string]any{"repository": "nektos/act"})
	require.NoError(t, err)
	defer s.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(s.RequestURL() + "?audience=sts.amazonaws.com")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
