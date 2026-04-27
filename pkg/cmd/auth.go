// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"context"
	"fmt"

	"github.com/boltz-bio/boltz-compute-api-cli/internal/apiquery"
	"github.com/boltz-bio/boltz-compute-api-go"
	"github.com/boltz-bio/boltz-compute-api-go/option"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v3"
)

var authMe = cli.Command{
	Name:            "me",
	Usage:           "Returns the organization context available to the current API key or OAuth\nbearer token. OAuth callers can use X-Boltz-Organization-Id to select one of\ntheir organizations.",
	Suggest:         true,
	Flags:           []cli.Flag{},
	Action:          handleAuthMe,
	HideHelpCommand: true,
}

func handleAuthMe(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	options, err := flagOptions(
		cmd,
		apiquery.NestedQueryFormatBrackets,
		apiquery.ArrayQueryFormatComma,
		EmptyBody,
		false,
	)
	if err != nil {
		return err
	}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Auth.Me(ctx, options...)
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
		Title:          "auth me",
		Transform:      transform,
	})
}
