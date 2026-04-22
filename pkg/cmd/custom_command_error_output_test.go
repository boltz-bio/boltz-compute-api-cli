// Custom CLI extension code. Not generated.
package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/boltz-bio/boltz-compute-api-cli/internal/autherror"
)

func TestWriteCommandErrorOutputFormatsAuthErrors(t *testing.T) {
	t.Parallel()

	root := parsedErrorTestRoot(t, "--format-error", "json", "call")
	stdoutR, stdoutW, err := os.Pipe()
	require.NoError(t, err)
	defer stdoutR.Close()

	var stderr bytes.Buffer
	handled := WriteCommandErrorOutput(root, autherror.New("auth_required", "Authentication required", "Run `boltz-api auth login`."), stdoutW, &stderr)
	require.True(t, handled)
	require.NoError(t, stdoutW.Close())

	output, err := io.ReadAll(stdoutR)
	require.NoError(t, err)
	require.Contains(t, string(output), `"type": "auth_error"`)
	require.Contains(t, string(output), `"code": "auth_required"`)
	require.Empty(t, stderr.String())
}

func TestWriteCommandErrorOutputFallsBackToBufferedCommandErrors(t *testing.T) {
	t.Parallel()

	root := parsedErrorTestRoot(t, "call")
	stdoutR, stdoutW, err := os.Pipe()
	require.NoError(t, err)
	defer stdoutR.Close()
	require.NoError(t, stdoutW.Close())

	CommandErrorBuffer.Reset()
	t.Cleanup(CommandErrorBuffer.Reset)
	_, err = CommandErrorBuffer.WriteString("Incorrect Usage: bad flag\n")
	require.NoError(t, err)

	var stderr bytes.Buffer
	handled := WriteCommandErrorOutput(root, cli.Exit("bad flag", 1), stdoutW, &stderr)
	require.True(t, handled)
	require.Contains(t, stderr.String(), "Incorrect Usage: bad flag")
}

func parsedErrorTestRoot(t *testing.T, args ...string) *cli.Command {
	t.Helper()

	var captured *cli.Command
	root := &cli.Command{
		Name: "boltz-api",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format-error", Value: "auto"},
			&cli.StringFlag{Name: "transform-error"},
		},
		Commands: []*cli.Command{
			{
				Name: "call",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					_ = ctx
					captured = cmd
					return nil
				},
			},
		},
	}

	runArgs := append([]string{"boltz-api"}, args...)
	require.NoError(t, root.Run(context.Background(), runArgs))
	require.NotNil(t, captured)
	return captured
}
