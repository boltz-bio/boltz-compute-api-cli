// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

var downloadStatusCommand = &cli.Command{
	Name:            "download-status",
	Usage:           "Show local download status from .boltz-run.json",
	Suggest:         true,
	HideHelpCommand: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "Local run directory name under --root-dir",
		},
		&cli.StringFlag{
			Name:  "run-dir",
			Usage: "Explicit local run directory path",
		},
		&cli.StringFlag{
			Name:  "root-dir",
			Usage: "Root directory for generated local run directories",
			Value: downloadResultsDefaultRootDir,
		},
	},
	Action: handleDownloadStatus,
}

type downloadStatusSpec struct {
	Name    *string
	RunDir  *string
	RootDir string
}

func handleDownloadStatus(ctx context.Context, cmd *cli.Command) error {
	_ = ctx

	spec, err := parseDownloadStatusSpec(cmd)
	if err != nil {
		return err
	}

	runDir, err := resolveDownloadStatusRunDir(spec)
	if err != nil {
		return err
	}

	metadata, err := loadDownloadMetadata(runDir)
	if err != nil {
		return err
	}

	return showJSONValue(cmd, buildDownloadStatusResponse(runDir, metadata), "download-status")
}

func parseDownloadStatusSpec(cmd *cli.Command) (downloadStatusSpec, error) {
	unusedArgs := cmd.Args().Slice()
	if len(unusedArgs) > 0 {
		return downloadStatusSpec{}, fmt.Errorf("Unexpected extra arguments: %v. Use --name or --run-dir to select a local run.", unusedArgs)
	}

	name := trimOptionalString(cmd.String("name"))
	runDir := trimOptionalString(cmd.String("run-dir"))
	rootDir := strings.TrimSpace(cmd.String("root-dir"))

	if name == nil && runDir == nil {
		return downloadStatusSpec{}, errors.New("--name or --run-dir is required")
	}
	if name != nil && runDir != nil {
		return downloadStatusSpec{}, errors.New("--name and --run-dir are mutually exclusive")
	}
	if runDir != nil && cmd.IsSet("root-dir") {
		return downloadStatusSpec{}, errors.New("--root-dir cannot be used with --run-dir")
	}
	if name != nil {
		if _, err := validateDownloadRunName(*name); err != nil {
			return downloadStatusSpec{}, err
		}
	}
	if runDir != nil && strings.TrimSpace(*runDir) == "" {
		return downloadStatusSpec{}, errors.New("--run-dir must not be empty")
	}

	return downloadStatusSpec{
		Name:    name,
		RunDir:  runDir,
		RootDir: rootDir,
	}, nil
}

func resolveDownloadStatusRunDir(spec downloadStatusSpec) (string, error) {
	if spec.RunDir != nil {
		return resolveDownloadPath(*spec.RunDir)
	}

	rootDir, err := resolveDownloadRootDir(spec.RootDir)
	if err != nil {
		return "", err
	}

	name, err := validateDownloadRunName(derefString(spec.Name))
	if err != nil {
		return "", err
	}
	return filepath.Join(rootDir, name), nil
}
