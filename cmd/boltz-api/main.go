// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package main

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/boltz-bio/boltz-api-cli/pkg/cmd"
	"github.com/urfave/cli/v3"
)

func main() {
	app := cmd.Command

	if slices.Contains(os.Args, "__complete") {
		prepareForAutocomplete(app)
	}

	if baseURL, ok := os.LookupEnv("BOLTZ_BASE_URL"); ok {
		if err := cmd.ValidateBaseURL(baseURL, "BOLTZ_BASE_URL"); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		exitCode := 1

		// Check if error has a custom exit code
		if exitErr, ok := err.(cli.ExitCoder); ok {
			exitCode = exitErr.ExitCode()
		}

		// Temporary Stainless seam: keep custom top-level error formatting behind a
		// single helper instead of embedding custom logic in the generated entrypoint.
		if !cmd.WriteCommandErrorOutput(app, err, os.Stdout, os.Stderr) && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
		os.Exit(exitCode)
	}
}

func prepareForAutocomplete(cmd *cli.Command) {
	// urfave/cli does not handle flag completions and will print an error if we inspect a command with invalid flags.
	// This skips that sort of validation
	cmd.SkipFlagParsing = true
	for _, child := range cmd.Commands {
		prepareForAutocomplete(child)
	}
}
