// Custom CLI extension code. Not generated.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	boltzcompute "github.com/boltz-bio/boltz-api-go"
	"github.com/boltz-bio/boltz-compute-api-cli/internal/autherror"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v3"
)

// WriteCommandErrorOutput is a temporary Stainless seam for cross-cutting custom
// error formatting. Keep the generated entrypoint limited to calling this helper.
func WriteCommandErrorOutput(app *cli.Command, err error, stdout *os.File, stderr io.Writer) bool {
	if writeStructuredCommandError(app, err, stdout, stderr) {
		return true
	}

	if CommandErrorBuffer.Len() > 0 {
		_, _ = stderr.Write(CommandErrorBuffer.Bytes())
		return true
	}

	return false
}

func writeStructuredCommandError(app *cli.Command, err error, stdout *os.File, stderr io.Writer) bool {
	var apierr *boltzcompute.Error
	if errors.As(err, &apierr) {
		fmt.Fprintf(stderr, "%s %q: %d %s\n", apierr.Request.Method, apierr.Request.URL, apierr.Response.StatusCode, http.StatusText(apierr.Response.StatusCode))
		return showStructuredCommandError(app, apierr.RawJSON(), err, stdout, stderr)
	}

	var authErr *autherror.Error
	if errors.As(err, &authErr) {
		return showStructuredCommandError(app, authErr.RawJSON(), err, stdout, stderr)
	}

	return false
}

func showStructuredCommandError(app *cli.Command, rawJSON string, fallback error, stdout *os.File, stderr io.Writer) bool {
	opts := ShowJSONOpts{
		ExplicitFormat: app != nil && app.IsSet("format-error"),
		Format:         commandStringFlag(app, "format-error"),
		Stderr:         stderr,
		Stdout:         stdout,
		Title:          "Error",
		Transform:      commandStringFlag(app, "transform-error"),
	}

	if showErr := ShowJSON(gjson.Parse(rawJSON), opts); showErr != nil {
		if fallback != nil && fallback.Error() != "" {
			fmt.Fprintf(stderr, "%s\n", fallback.Error())
		}
	}

	return true
}

func commandStringFlag(cmd *cli.Command, name string) string {
	if cmd == nil {
		return ""
	}
	return cmd.String(name)
}
