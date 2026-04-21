// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"context"
	"fmt"

	"github.com/boltz-bio/boltz-compute-api-cli/internal/apiquery"
	"github.com/boltz-bio/boltz-compute-api-cli/internal/requestflag"
	"github.com/boltz-bio/boltz-compute-api-go"
	"github.com/boltz-bio/boltz-compute-api-go/option"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v3"
)

var adminAPIKeysCreate = cli.Command{
	Name:    "create",
	Usage:   "Create a workspace API key",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:     "name",
			Usage:    "API key name",
			Required: true,
			BodyPath: "name",
		},
		&requestflag.Flag[[]string]{
			Name:     "allowed-ip",
			Usage:    "IP addresses allowed to use this key (IPv4 or IPv6). An empty array (the default) means all IPs are allowed.",
			Default:  []string{},
			BodyPath: "allowed_ips",
		},
		&requestflag.Flag[int64]{
			Name:     "expires-in-days",
			Usage:    "Days until the key expires. Omit for a key that does not expire.",
			BodyPath: "expires_in_days",
		},
		&requestflag.Flag[string]{
			Name:     "mode",
			Usage:    "Key mode. Test keys create test-mode resources with synthetic data.",
			Default:  "live",
			BodyPath: "mode",
		},
		&requestflag.Flag[string]{
			Name:     "workspace-id",
			Usage:    "Workspace ID to scope this key to. Omit for default workspace.",
			BodyPath: "workspace_id",
		},
	},
	Action:          handleAdminAPIKeysCreate,
	HideHelpCommand: true,
}

var adminAPIKeysList = cli.Command{
	Name:    "list",
	Usage:   "List API keys",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "after-id",
			Usage:     "Return results after this ID",
			QueryPath: "after_id",
		},
		&requestflag.Flag[string]{
			Name:      "before-id",
			Usage:     "Return results before this ID",
			QueryPath: "before_id",
		},
		&requestflag.Flag[int64]{
			Name:      "limit",
			Usage:     "Max items to return",
			Default:   100,
			QueryPath: "limit",
		},
		&requestflag.Flag[string]{
			Name:      "workspace-id",
			Usage:     "Filter by workspace ID. If not provided, returns keys across all workspaces.",
			QueryPath: "workspace_id",
		},
		&requestflag.Flag[int64]{
			Name:  "max-items",
			Usage: "The maximum number of items to return (use -1 for unlimited).",
		},
	},
	Action:          handleAdminAPIKeysList,
	HideHelpCommand: true,
}

var adminAPIKeysRevoke = cli.Command{
	Name:    "revoke",
	Usage:   "Revoke an API key",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:     "api-key-id",
			Required: true,
		},
	},
	Action:          handleAdminAPIKeysRevoke,
	HideHelpCommand: true,
}

func handleAdminAPIKeysCreate(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	params := boltzcompute.AdminAPIKeyNewParams{}

	options, err := flagOptions(
		cmd,
		apiquery.NestedQueryFormatBrackets,
		apiquery.ArrayQueryFormatComma,
		ApplicationJSON,
		false,
	)
	if err != nil {
		return err
	}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Admin.APIKeys.New(ctx, params, options...)
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
		Title:          "admin:api-keys create",
		Transform:      transform,
	})
}

func handleAdminAPIKeysList(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	params := boltzcompute.AdminAPIKeyListParams{}

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

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.Admin.APIKeys.List(ctx, params, options...)
		if err != nil {
			return err
		}
		obj := gjson.ParseBytes(res)
		return ShowJSON(obj, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "admin:api-keys list",
			Transform:      transform,
		})
	} else {
		iter := client.Admin.APIKeys.ListAutoPaging(ctx, params, options...)
		maxItems := int64(-1)
		if cmd.IsSet("max-items") {
			maxItems = cmd.Value("max-items").(int64)
		}
		return ShowJSONIterator(iter, maxItems, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "admin:api-keys list",
			Transform:      transform,
		})
	}
}

func handleAdminAPIKeysRevoke(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()
	if !cmd.IsSet("api-key-id") && len(unusedArgs) > 0 {
		cmd.Set("api-key-id", unusedArgs[0])
		unusedArgs = unusedArgs[1:]
	}
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
	_, err = client.Admin.APIKeys.Revoke(ctx, cmd.Value("api-key-id").(string), options...)
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
		Title:          "admin:api-keys revoke",
		Transform:      transform,
	})
}
