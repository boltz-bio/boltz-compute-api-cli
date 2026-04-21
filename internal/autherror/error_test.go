package autherror

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewWithHintsPopulatesPrimaryHintAndHintsArray(t *testing.T) {
	err := NewWithHints(
		"auth_required",
		"Authentication required",
		"Run `boltz-api auth login`.",
		"Set `--api-key` to use API-key mode.",
		"Run `boltz-api auth login`.",
		"",
	)

	envelope := err.Envelope()
	require.Equal(t, "Run `boltz-api auth login`.", envelope.Hint)
	require.Equal(t, []string{
		"Run `boltz-api auth login`.",
		"Set `--api-key` to use API-key mode.",
	}, envelope.Hints)
	require.Contains(t, err.RawJSON(), `"hints"`)
}
