// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"context"
	"fmt"

	boltzcompute "github.com/boltz-bio/boltz-api-go"
	"github.com/boltz-bio/boltz-api-go/option"
	"github.com/boltz-bio/boltz-api-cli/internal/apiquery"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v3"
)

var cliVersion = cli.Command{
	Name:    "version",
	Usage:   "Returns public boltz-api CLI version metadata for lightweight update checks.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "current",
			QueryPath: "current",
		},
		&requestflag.Flag[string]{
			Name:      "platform",
			QueryPath: "platform",
		},
	},
	Action:          handleCliVersion,
	HideHelpCommand: true,
}

func handleCliVersion(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	options, err := flagOptions(
		cmd,
		apiquery.NestedQueryFormatBrackets,
		apiquery.ArrayQueryFormatRepeat,
		EmptyBody,
		false,
	)
	if err != nil {
		return err
	}

	params := boltzcompute.CliVersionParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Cli.Version(ctx, params, options...)
	if err != nil {
		return err
	}

	obj := gjson.ParseBytes(res)
	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	return ShowJSON(obj, ShowJSONOpts{
		ExplicitFormat: explicitFormat,
		Format:         format,
		RawOutput:      cmd.Root().Bool("raw-output"),
		Title:          "cli version",
		Transform:      transform,
	})
}
