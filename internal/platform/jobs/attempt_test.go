package jobs_test

import (
	"context"
	"testing"

	"go-echo-server-template/internal/platform/jobs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttemptInfoContext(t *testing.T) {
	ctx := context.Background()
	_, ok := jobs.AttemptFromContext(ctx)
	assert.False(t, ok)

	ctx = jobs.ContextWithAttempt(ctx, jobs.AttemptInfo{Attempt: 3, MaxAttempts: 5})
	info, ok := jobs.AttemptFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, 3, info.Attempt)
	assert.Equal(t, 5, info.MaxAttempts)
}
