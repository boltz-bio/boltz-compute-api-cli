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

var adminUsageList = cli.Command{
	Name:    "list",
	Usage:   "Retrieve aggregated usage data across the organization, optionally grouped by\nworkspace and/or application.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[any]{
			Name:      "ending-at",
			Usage:     "End of the time range as an ISO 8601 date-time with timezone, for example 2026-04-08T18:56:46Z",
			Required:  true,
			QueryPath: "ending_at",
		},
		&requestflag.Flag[any]{
			Name:      "starting-at",
			Usage:     "Start of the time range as an ISO 8601 date-time with timezone, for example 2026-04-08T18:56:46Z",
			Required:  true,
			QueryPath: "starting_at",
		},
		&requestflag.Flag[string]{
			Name:      "window-size",
			Usage:     "Time window size. HOUR supports up to 31 days per query; DAY supports up to 365 days per query.",
			Required:  true,
			QueryPath: "window_size",
		},
		&requestflag.Flag[any]{
			Name:      "applications",
			Usage:     "Filter to specific applications",
			QueryPath: "applications[]",
		},
		&requestflag.Flag[any]{
			Name:      "group-by",
			Usage:     "Group results by workspace_id and/or application",
			QueryPath: "group_by[]",
		},
		&requestflag.Flag[int64]{
			Name:      "limit",
			Usage:     "Maximum number of buckets to return",
			Default:   100,
			QueryPath: "limit",
		},
		&requestflag.Flag[string]{
			Name:      "page",
			Usage:     "Cursor for pagination",
			QueryPath: "page",
		},
		&requestflag.Flag[any]{
			Name:      "workspace-ids",
			Usage:     "Filter to specific workspace IDs",
			QueryPath: "workspace_ids[]",
		},
		&requestflag.Flag[int64]{
			Name:  "max-items",
			Usage: "The maximum number of items to return (use -1 for unlimited).",
		},
	},
	Action:          handleAdminUsageList,
	HideHelpCommand: true,
}

func handleAdminUsageList(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	params := boltzcompute.AdminUsageListParams{}

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
		_, err = client.Admin.Usage.List(ctx, params, options...)
		if err != nil {
			return err
		}
		obj := gjson.ParseBytes(res)
		return ShowJSON(obj, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "admin:usage list",
			Transform:      transform,
		})
	} else {
		iter := client.Admin.Usage.ListAutoPaging(ctx, params, options...)
		maxItems := int64(-1)
		if cmd.IsSet("max-items") {
			maxItems = cmd.Value("max-items").(int64)
		}
		return ShowJSONIterator(iter, maxItems, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "admin:usage list",
			Transform:      transform,
		})
	}
}
