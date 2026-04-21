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

var proteinLibraryScreenRetrieve = cli.Command{
	Name:    "retrieve",
	Usage:   "Retrieve a library screen by ID, including progress and status",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:     "screen-id",
			Required: true,
		},
		&requestflag.Flag[string]{
			Name:      "workspace-id",
			Usage:     "Workspace ID. Only used with admin API keys. Ignored (or validated) for workspace-scoped keys.",
			QueryPath: "workspace_id",
		},
	},
	Action:          handleProteinLibraryScreenRetrieve,
	HideHelpCommand: true,
}

var proteinLibraryScreenList = cli.Command{
	Name:    "list",
	Usage:   "List protein library screens, optionally filtered by workspace",
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
			Usage:     "Max items to return. Defaults to 100.",
			Default:   100,
			QueryPath: "limit",
		},
		&requestflag.Flag[string]{
			Name:      "workspace-id",
			Usage:     "Filter by workspace ID. Only used with admin API keys. If not provided, defaults to the workspace associated with the API key, or the default workspace for admin keys.",
			QueryPath: "workspace_id",
		},
		&requestflag.Flag[int64]{
			Name:  "max-items",
			Usage: "The maximum number of items to return (use -1 for unlimited).",
		},
	},
	Action:          handleProteinLibraryScreenList,
	HideHelpCommand: true,
}

var proteinLibraryScreenDeleteData = cli.Command{
	Name:    "delete-data",
	Usage:   "Permanently delete the input, output, and result data associated with this\nlibrary screen. The library screen record itself is retained with a\n`data_deleted_at` timestamp. This action is irreversible.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:     "screen-id",
			Required: true,
		},
	},
	Action:          handleProteinLibraryScreenDeleteData,
	HideHelpCommand: true,
}

var proteinLibraryScreenEstimateCost = requestflag.WithInnerFlags(cli.Command{
	Name:    "estimate-cost",
	Usage:   "Estimate the cost of a protein library screen without creating any resource or\nconsuming GPU.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[[]map[string]any]{
			Name:     "protein",
			Usage:    "List of protein entries to screen.",
			Required: true,
			BodyPath: "proteins",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "target",
			Usage:    "Target specification (structure template or template-free)",
			Required: true,
			BodyPath: "target",
		},
		&requestflag.Flag[string]{
			Name:     "idempotency-key",
			Usage:    "Client-provided key to prevent duplicate submissions on retries",
			BodyPath: "idempotency_key",
		},
		&requestflag.Flag[string]{
			Name:     "workspace-id",
			Usage:    "Target workspace ID (admin keys only; ignored for workspace keys)",
			BodyPath: "workspace_id",
		},
	},
	Action:          handleProteinLibraryScreenEstimateCost,
	HideHelpCommand: true,
}, map[string][]requestflag.HasOuterFlag{
	"protein": {
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "protein.entities",
			Usage:      "Entities that make up this protein complex",
			InnerField: "entities",
		},
		&requestflag.InnerFlag[string]{
			Name:       "protein.id",
			Usage:      "Optional client-provided identifier for this entry",
			InnerField: "id",
		},
	},
})

var proteinLibraryScreenListResults = cli.Command{
	Name:    "list-results",
	Usage:   "Retrieve paginated results from a protein library screen",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:     "screen-id",
			Required: true,
		},
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
			Usage:     "Max results to return. Defaults to 100.",
			Default:   100,
			QueryPath: "limit",
		},
		&requestflag.Flag[string]{
			Name:      "workspace-id",
			Usage:     "Workspace ID. Only used with admin API keys. Ignored (or validated) for workspace-scoped keys.",
			QueryPath: "workspace_id",
		},
		&requestflag.Flag[int64]{
			Name:  "max-items",
			Usage: "The maximum number of items to return (use -1 for unlimited).",
		},
	},
	Action:          handleProteinLibraryScreenListResults,
	HideHelpCommand: true,
}

var proteinLibraryScreenStart = requestflag.WithInnerFlags(cli.Command{
	Name:    "start",
	Usage:   "Screen a set of protein candidates against a target",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[[]map[string]any]{
			Name:     "protein",
			Usage:    "List of protein entries to screen.",
			Required: true,
			BodyPath: "proteins",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "target",
			Usage:    "Target specification (structure template or template-free)",
			Required: true,
			BodyPath: "target",
		},
		&requestflag.Flag[string]{
			Name:     "idempotency-key",
			Usage:    "Client-provided key to prevent duplicate submissions on retries",
			BodyPath: "idempotency_key",
		},
		&requestflag.Flag[string]{
			Name:     "workspace-id",
			Usage:    "Target workspace ID (admin keys only; ignored for workspace keys)",
			BodyPath: "workspace_id",
		},
	},
	Action:          handleProteinLibraryScreenStart,
	HideHelpCommand: true,
}, map[string][]requestflag.HasOuterFlag{
	"protein": {
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "protein.entities",
			Usage:      "Entities that make up this protein complex",
			InnerField: "entities",
		},
		&requestflag.InnerFlag[string]{
			Name:       "protein.id",
			Usage:      "Optional client-provided identifier for this entry",
			InnerField: "id",
		},
	},
})

var proteinLibraryScreenStop = cli.Command{
	Name:    "stop",
	Usage:   "Stop an in-progress protein library screen early",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:     "screen-id",
			Required: true,
		},
	},
	Action:          handleProteinLibraryScreenStop,
	HideHelpCommand: true,
}

func handleProteinLibraryScreenRetrieve(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()
	if !cmd.IsSet("screen-id") && len(unusedArgs) > 0 {
		cmd.Set("screen-id", unusedArgs[0])
		unusedArgs = unusedArgs[1:]
	}
	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	params := boltzcompute.ProteinLibraryScreenGetParams{}

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
	_, err = client.Protein.LibraryScreen.Get(
		ctx,
		cmd.Value("screen-id").(string),
		params,
		options...,
	)
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
		Title:          "protein:library-screen retrieve",
		Transform:      transform,
	})
}

func handleProteinLibraryScreenList(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	params := boltzcompute.ProteinLibraryScreenListParams{}

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
		_, err = client.Protein.LibraryScreen.List(ctx, params, options...)
		if err != nil {
			return err
		}
		obj := gjson.ParseBytes(res)
		return ShowJSON(obj, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "protein:library-screen list",
			Transform:      transform,
		})
	} else {
		iter := client.Protein.LibraryScreen.ListAutoPaging(ctx, params, options...)
		maxItems := int64(-1)
		if cmd.IsSet("max-items") {
			maxItems = cmd.Value("max-items").(int64)
		}
		return ShowJSONIterator(iter, maxItems, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "protein:library-screen list",
			Transform:      transform,
		})
	}
}

func handleProteinLibraryScreenDeleteData(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()
	if !cmd.IsSet("screen-id") && len(unusedArgs) > 0 {
		cmd.Set("screen-id", unusedArgs[0])
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
	_, err = client.Protein.LibraryScreen.DeleteData(ctx, cmd.Value("screen-id").(string), options...)
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
		Title:          "protein:library-screen delete-data",
		Transform:      transform,
	})
}

func handleProteinLibraryScreenEstimateCost(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	params := boltzcompute.ProteinLibraryScreenEstimateCostParams{}

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
	_, err = client.Protein.LibraryScreen.EstimateCost(ctx, params, options...)
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
		Title:          "protein:library-screen estimate-cost",
		Transform:      transform,
	})
}

func handleProteinLibraryScreenListResults(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()
	if !cmd.IsSet("screen-id") && len(unusedArgs) > 0 {
		cmd.Set("screen-id", unusedArgs[0])
		unusedArgs = unusedArgs[1:]
	}
	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	params := boltzcompute.ProteinLibraryScreenListResultsParams{}

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
		_, err = client.Protein.LibraryScreen.ListResults(
			ctx,
			cmd.Value("screen-id").(string),
			params,
			options...,
		)
		if err != nil {
			return err
		}
		obj := gjson.ParseBytes(res)
		return ShowJSON(obj, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "protein:library-screen list-results",
			Transform:      transform,
		})
	} else {
		iter := client.Protein.LibraryScreen.ListResultsAutoPaging(
			ctx,
			cmd.Value("screen-id").(string),
			params,
			options...,
		)
		maxItems := int64(-1)
		if cmd.IsSet("max-items") {
			maxItems = cmd.Value("max-items").(int64)
		}
		return ShowJSONIterator(iter, maxItems, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "protein:library-screen list-results",
			Transform:      transform,
		})
	}
}

func handleProteinLibraryScreenStart(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	params := boltzcompute.ProteinLibraryScreenStartParams{}

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
	_, err = client.Protein.LibraryScreen.Start(ctx, params, options...)
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
		Title:          "protein:library-screen start",
		Transform:      transform,
	})
}

func handleProteinLibraryScreenStop(ctx context.Context, cmd *cli.Command) error {
	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()
	if !cmd.IsSet("screen-id") && len(unusedArgs) > 0 {
		cmd.Set("screen-id", unusedArgs[0])
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
	_, err = client.Protein.LibraryScreen.Stop(ctx, cmd.Value("screen-id").(string), options...)
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
		Title:          "protein:library-screen stop",
		Transform:      transform,
	})
}
