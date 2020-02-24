package caller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCallerFunctionName(t *testing.T) {
	names := firstFunction()
	require.Equal(t, len(names), 3)
	assert.Equal(t, "reserve-data/lib/caller.firstFunction", names[0])
	assert.Equal(t, "reserve-data/lib/caller.secondFunction", names[1])
	assert.Equal(t, "reserve-data/lib/caller.firstFunction", names[2])
}

func firstFunction() []string {
	return append([]string{
		GetCurrentFunctionName(),
	}, secondFunction()...)
}

func secondFunction() []string {
	return []string{
		GetCurrentFunctionName(),
		GetCallerFunctionName(),
	}
}
