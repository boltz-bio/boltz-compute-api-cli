package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/boltz-bio/boltz-compute-api-go"
	"github.com/boltz-bio/boltz-compute-api-go/option"
	"github.com/boltz-bio/boltz-compute-api-go/packages/pagination"
	"github.com/boltz-bio/boltz-compute-api-go/packages/param"
	"github.com/urfave/cli/v3"
)

const (
	downloadResultsDefaultRootDir     = "boltz-experiments"
	downloadResultsMetadataFileName   = ".boltz-run.json"
	downloadResultsSchemaVersion      = 1
	downloadResultsPageLimit          = 20
	downloadResultsSummaryIntervalSec = 30 * time.Second
)

var (
	downloadResultsRunTypePrefixes = []struct {
		prefix  string
		runType downloadRunType
	}{
		{prefix: "pred_", runType: downloadRunTypePrediction},
		{prefix: "prot_des_", runType: downloadRunTypeProteinDesign},
		{prefix: "prot_scr_", runType: downloadRunTypeProteinLibraryScreen},
		{prefix: "sm_des_", runType: downloadRunTypeSmallMoleculeDesign},
		{prefix: "sm_scr_", runType: downloadRunTypeSmallMoleculeScreen},
	}
	downloadResultsUnsupportedPrefixes   = []string{"pres_", "psmp_"}
	downloadResultsAdjectives            = []string{"amber", "brisk", "calm", "clear", "keen", "lively", "lucid", "quiet", "rapid", "steady"}
	downloadResultsNouns                 = []string{"atom", "binder", "cluster", "enzyme", "helix", "ligand", "motif", "pocket", "sample", "target"}
	downloadResultsVerbs                 = []string{"aligns", "builds", "checks", "drifts", "folds", "guides", "mixes", "screens", "shifts", "tracks"}
	downloadResultsKnownMetadataFields   = map[string]struct{}{"schema_version": {}, "name": {}, "run_type": {}, "request_fingerprint": {}, "idempotency_key": {}, "remote": {}, "pending": {}, "cursor_after_id": {}}
	downloadResultsSupportedArchiveTypes = []string{".tar.gz", ".tgz", ".tar", ".zip"}
)

var downloadResultsCommand = &cli.Command{
	Name:            "download-results",
	Usage:           "Download and resume result archives for predictions and pipelines",
	Suggest:         true,
	HideHelpCommand: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "id",
			Usage: "Prediction or pipeline run ID",
		},
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
		&cli.StringFlag{
			Name:  "workspace-id",
			Usage: "Workspace ID. Only used with admin API keys. Ignored (or validated) for workspace-scoped keys.",
		},
		&cli.Float64Flag{
			Name:  "poll-interval-seconds",
			Usage: "Polling interval while waiting for remote results",
			Value: 5.0,
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Print progress logs to stderr while downloading",
		},
	},
	Action: handleDownloadResults,
}

type downloadRunType string

const (
	downloadRunTypePrediction            downloadRunType = "prediction"
	downloadRunTypeProteinDesign         downloadRunType = "protein_design"
	downloadRunTypeProteinLibraryScreen  downloadRunType = "protein_library_screen"
	downloadRunTypeSmallMoleculeDesign   downloadRunType = "small_molecule_design"
	downloadRunTypeSmallMoleculeScreen   downloadRunType = "small_molecule_library_screen"
	downloadPendingKindPredictionArchive string          = "prediction_archive"
	downloadPendingKindResultPage        string          = "result_page"
)

type downloadResultsSpec struct {
	ID                  *string
	Name                *string
	RunDir              *string
	RootDir             string
	WorkspaceID         *string
	PollIntervalSeconds float64
	Verbose             bool
}

type downloadRunMetadata struct {
	SchemaVersion      int                        `json:"schema_version"`
	Name               string                     `json:"name"`
	RunType            downloadRunType            `json:"run_type"`
	RequestFingerprint string                     `json:"request_fingerprint"`
	IdempotencyKey     string                     `json:"idempotency_key"`
	Remote             downloadRemoteState        `json:"remote"`
	Pending            *downloadPendingState      `json:"pending"`
	CursorAfterID      *string                    `json:"cursor_after_id"`
	ExtraFields        map[string]json.RawMessage `json:"-"`
}

type downloadRemoteState struct {
	RunID          *string `json:"run_id"`
	WorkspaceID    *string `json:"workspace_id"`
	Status         *string `json:"status"`
	StartedAt      *string `json:"started_at"`
	CompletedAt    *string `json:"completed_at"`
	StoppedAt      *string `json:"stopped_at"`
	LatestResultID *string `json:"latest_result_id"`
	ErrorCode      *string `json:"error_code"`
}

type downloadPendingState struct {
	Kind       string   `json:"kind"`
	AfterID    *string  `json:"after_id"`
	PageLastID *string  `json:"page_last_id"`
	ResultIDs  []string `json:"result_ids"`
}

type downloadResultsEngine struct {
	client *boltzcompute.Client
	sink   *downloadResultsSink
}

type downloadResultsSink struct {
	verbose              bool
	writer               io.Writer
	lastRunningSummaryAt time.Time
}

type downloadPredictionRunInfo struct {
	Status      string
	WorkspaceID *string
	StartedAt   *string
	CompletedAt *string
	ErrorCode   *string
	ArchiveURL  *string
}

type downloadPipelineRunInfo struct {
	Status         string
	WorkspaceID    *string
	StartedAt      *string
	CompletedAt    *string
	StoppedAt      *string
	LatestResultID *string
	ErrorCode      *string
}

type downloadPipelineResultInfo struct {
	ID         string
	ArchiveURL string
}

type downloadPipelinePage struct {
	Results []downloadPipelineResultInfo
	LastID  *string
}

type downloadPipelineAdapter struct {
	get         func(ctx context.Context, client *boltzcompute.Client, runID string, workspaceID *string) (downloadPipelineRunInfo, error)
	listResults func(ctx context.Context, client *boltzcompute.Client, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error)
}

var downloadPipelineAdapters = map[downloadRunType]downloadPipelineAdapter{
	downloadRunTypeProteinDesign: {
		get: func(ctx context.Context, client *boltzcompute.Client, runID string, workspaceID *string) (downloadPipelineRunInfo, error) {
			params := boltzcompute.ProteinDesignGetParams{}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			response, err := client.Protein.Design.Get(ctx, runID, params)
			if err != nil {
				return downloadPipelineRunInfo{}, err
			}
			return newDownloadPipelineRunInfo(
				string(response.Status),
				response.WorkspaceID,
				response.StartedAt,
				response.CompletedAt,
				response.StoppedAt,
				response.Progress.LatestResultID,
				response.Error.Code,
			), nil
		},
		listResults: func(ctx context.Context, client *boltzcompute.Client, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error) {
			params := boltzcompute.ProteinDesignListResultsParams{Limit: param.NewOpt(int64(downloadResultsPageLimit))}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			setOptionalStringOpt(&params.AfterID, afterID)
			response, err := client.Protein.Design.ListResults(ctx, runID, params)
			if err != nil {
				return downloadPipelinePage{}, err
			}
			return normalizePipelinePage(response, func(result boltzcompute.ProteinDesignListResultsResponse) downloadPipelineResultInfo {
				return downloadPipelineResultInfo{ID: result.ID, ArchiveURL: result.Artifacts.Archive.URL}
			}), nil
		},
	},
	downloadRunTypeProteinLibraryScreen: {
		get: func(ctx context.Context, client *boltzcompute.Client, runID string, workspaceID *string) (downloadPipelineRunInfo, error) {
			params := boltzcompute.ProteinLibraryScreenGetParams{}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			response, err := client.Protein.LibraryScreen.Get(ctx, runID, params)
			if err != nil {
				return downloadPipelineRunInfo{}, err
			}
			return newDownloadPipelineRunInfo(
				string(response.Status),
				response.WorkspaceID,
				response.StartedAt,
				response.CompletedAt,
				response.StoppedAt,
				response.Progress.LatestResultID,
				response.Error.Code,
			), nil
		},
		listResults: func(ctx context.Context, client *boltzcompute.Client, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error) {
			params := boltzcompute.ProteinLibraryScreenListResultsParams{Limit: param.NewOpt(int64(downloadResultsPageLimit))}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			setOptionalStringOpt(&params.AfterID, afterID)
			response, err := client.Protein.LibraryScreen.ListResults(ctx, runID, params)
			if err != nil {
				return downloadPipelinePage{}, err
			}
			return normalizePipelinePage(response, func(result boltzcompute.ProteinLibraryScreenListResultsResponse) downloadPipelineResultInfo {
				return downloadPipelineResultInfo{ID: result.ID, ArchiveURL: result.Artifacts.Archive.URL}
			}), nil
		},
	},
	downloadRunTypeSmallMoleculeDesign: {
		get: func(ctx context.Context, client *boltzcompute.Client, runID string, workspaceID *string) (downloadPipelineRunInfo, error) {
			params := boltzcompute.SmallMoleculeDesignGetParams{}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			response, err := client.SmallMolecule.Design.Get(ctx, runID, params)
			if err != nil {
				return downloadPipelineRunInfo{}, err
			}
			return newDownloadPipelineRunInfo(
				string(response.Status),
				response.WorkspaceID,
				response.StartedAt,
				response.CompletedAt,
				response.StoppedAt,
				response.Progress.LatestResultID,
				response.Error.Code,
			), nil
		},
		listResults: func(ctx context.Context, client *boltzcompute.Client, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error) {
			params := boltzcompute.SmallMoleculeDesignListResultsParams{Limit: param.NewOpt(int64(downloadResultsPageLimit))}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			setOptionalStringOpt(&params.AfterID, afterID)
			response, err := client.SmallMolecule.Design.ListResults(ctx, runID, params)
			if err != nil {
				return downloadPipelinePage{}, err
			}
			return normalizePipelinePage(response, func(result boltzcompute.SmallMoleculeDesignListResultsResponse) downloadPipelineResultInfo {
				return downloadPipelineResultInfo{ID: result.ID, ArchiveURL: result.Artifacts.Archive.URL}
			}), nil
		},
	},
	downloadRunTypeSmallMoleculeScreen: {
		get: func(ctx context.Context, client *boltzcompute.Client, runID string, workspaceID *string) (downloadPipelineRunInfo, error) {
			params := boltzcompute.SmallMoleculeLibraryScreenGetParams{}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			response, err := client.SmallMolecule.LibraryScreen.Get(ctx, runID, params)
			if err != nil {
				return downloadPipelineRunInfo{}, err
			}
			return newDownloadPipelineRunInfo(
				string(response.Status),
				response.WorkspaceID,
				response.StartedAt,
				response.CompletedAt,
				response.StoppedAt,
				response.Progress.LatestResultID,
				response.Error.Code,
			), nil
		},
		listResults: func(ctx context.Context, client *boltzcompute.Client, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error) {
			params := boltzcompute.SmallMoleculeLibraryScreenListResultsParams{Limit: param.NewOpt(int64(downloadResultsPageLimit))}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			setOptionalStringOpt(&params.AfterID, afterID)
			response, err := client.SmallMolecule.LibraryScreen.ListResults(ctx, runID, params)
			if err != nil {
				return downloadPipelinePage{}, err
			}
			return normalizePipelinePage(response, func(result boltzcompute.SmallMoleculeLibraryScreenListResultsResponse) downloadPipelineResultInfo {
				return downloadPipelineResultInfo{ID: result.ID, ArchiveURL: result.Artifacts.Archive.URL}
			}), nil
		},
	},
}

func handleDownloadResults(ctx context.Context, cmd *cli.Command) error {
	spec, err := parseDownloadResultsSpec(cmd)
	if err != nil {
		return err
	}

	client := boltzcompute.NewClient(getDefaultRequestOptions(cmd)...)
	engine := downloadResultsEngine{
		client: &client,
		sink: &downloadResultsSink{
			verbose: spec.Verbose,
			writer:  commandErrorWriter(cmd),
		},
	}

	runDir, err := engine.download(ctx, spec)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(commandWriter(cmd), runDir)
	return err
}

func parseDownloadResultsSpec(cmd *cli.Command) (downloadResultsSpec, error) {
	unusedArgs := cmd.Args().Slice()
	id := trimOptionalString(cmd.String("id"))
	if id == nil && len(unusedArgs) > 0 {
		id = trimOptionalString(unusedArgs[0])
		unusedArgs = unusedArgs[1:]
	}
	if len(unusedArgs) > 0 {
		return downloadResultsSpec{}, fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	name := trimOptionalString(cmd.String("name"))
	runDir := trimOptionalString(cmd.String("run-dir"))
	rootDir := strings.TrimSpace(cmd.String("root-dir"))
	workspaceID := trimOptionalString(cmd.String("workspace-id"))
	pollIntervalSeconds := cmd.Float64("poll-interval-seconds")
	if pollIntervalSeconds < 0 {
		return downloadResultsSpec{}, errors.New("--poll-interval-seconds must be non-negative")
	}
	if name != nil && runDir != nil {
		return downloadResultsSpec{}, errors.New("--name and --run-dir are mutually exclusive")
	}
	if runDir != nil && cmd.IsSet("root-dir") {
		return downloadResultsSpec{}, errors.New("--root-dir cannot be used with --run-dir")
	}
	if name != nil {
		if _, err := validateDownloadRunName(*name); err != nil {
			return downloadResultsSpec{}, err
		}
	}
	if runDir != nil && strings.TrimSpace(*runDir) == "" {
		return downloadResultsSpec{}, errors.New("--run-dir must not be empty")
	}
	if id != nil {
		if _, err := inferDownloadRunType(*id); err != nil {
			return downloadResultsSpec{}, err
		}
	}

	return downloadResultsSpec{
		ID:                  id,
		Name:                name,
		RunDir:              runDir,
		RootDir:             rootDir,
		WorkspaceID:         workspaceID,
		PollIntervalSeconds: pollIntervalSeconds,
		Verbose:             cmd.Bool("verbose"),
	}, nil
}

func (e *downloadResultsEngine) download(ctx context.Context, spec downloadResultsSpec) (string, error) {
	runDir, metadata, err := prepareDownloadRun(spec)
	if err != nil {
		return "", err
	}

	switch metadata.RunType {
	case downloadRunTypePrediction:
		if err := e.waitForPrediction(ctx, runDir, &metadata, spec.PollIntervalSeconds); err != nil {
			return "", err
		}
	default:
		if err := e.waitForPipeline(ctx, runDir, &metadata, spec.PollIntervalSeconds); err != nil {
			return "", err
		}
	}

	return runDir, nil
}

func prepareDownloadRun(spec downloadResultsSpec) (string, downloadRunMetadata, error) {
	runDir, err := resolveDownloadRunDir(spec)
	if err != nil {
		return "", downloadRunMetadata{}, err
	}

	metadataPath := downloadMetadataPath(runDir)
	if _, err := os.Stat(metadataPath); err == nil {
		metadata, err := loadDownloadMetadata(runDir)
		if err != nil {
			return "", downloadRunMetadata{}, err
		}
		if err := reconcileMetadataWithSpec(runDir, &metadata, spec); err != nil {
			return "", downloadRunMetadata{}, err
		}
		return runDir, metadata, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", downloadRunMetadata{}, err
	}

	if spec.ID == nil {
		return "", downloadRunMetadata{}, fmt.Errorf("No local metadata exists for %s. Provide --id to create a new download run.", runDir)
	}

	if err := ensureDownloadDirectoryReady(runDir); err != nil {
		return "", downloadRunMetadata{}, err
	}

	runType, err := inferDownloadRunType(*spec.ID)
	if err != nil {
		return "", downloadRunMetadata{}, err
	}

	metadata := newDownloadRunMetadata(filepath.Base(runDir), runType, *spec.ID, spec.WorkspaceID)
	if err := saveDownloadMetadata(runDir, metadata); err != nil {
		return "", downloadRunMetadata{}, err
	}
	return runDir, metadata, nil
}

func resolveDownloadRunDir(spec downloadResultsSpec) (string, error) {
	if spec.RunDir != nil {
		return resolveDownloadPath(*spec.RunDir)
	}

	rootDir, err := resolveDownloadRootDir(spec.RootDir)
	if err != nil {
		return "", err
	}

	if spec.Name != nil {
		name, err := validateDownloadRunName(*spec.Name)
		if err != nil {
			return "", err
		}
		return filepath.Join(rootDir, name), nil
	}

	if spec.ID == nil {
		return "", errors.New("Either local metadata must already exist or --id must be provided")
	}

	return resolveCreateDownloadRunDir(rootDir)
}

func resolveDownloadRootDir(rootDir string) (string, error) {
	if strings.TrimSpace(rootDir) == "" {
		rootDir = downloadResultsDefaultRootDir
	}
	return resolveDownloadPath(rootDir)
}

func resolveDownloadPath(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("path must not be empty")
	}
	if strings.HasPrefix(trimmed, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		switch trimmed {
		case "~":
			trimmed = home
		default:
			if strings.HasPrefix(trimmed, "~/") || strings.HasPrefix(trimmed, `~\`) {
				trimmed = filepath.Join(home, trimmed[2:])
			}
		}
	}
	if filepath.IsAbs(trimmed) {
		return filepath.Clean(trimmed), nil
	}
	return filepath.Abs(trimmed)
}

func resolveCreateDownloadRunDir(rootDir string) (string, error) {
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return "", err
	}
	for i := 0; i < 256; i++ {
		candidate := filepath.Join(rootDir, generateDownloadRunName())
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		} else if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("Unable to generate a unique run directory under %s", rootDir)
}

func validateDownloadRunName(name string) (string, error) {
	stripped := strings.TrimSpace(name)
	if stripped == "" {
		return "", errors.New("`name` must not be empty")
	}
	if stripped == "." || stripped == ".." {
		return "", errors.New("`name` must not be '.' or '..'")
	}
	if filepath.Base(stripped) != stripped {
		return "", errors.New("`name` must be a single directory name, not a path")
	}
	return stripped, nil
}

func generateDownloadRunName() string {
	return fmt.Sprintf(
		"%s-%s-%s-%s",
		randomChoice(downloadResultsAdjectives),
		randomChoice(downloadResultsNouns),
		randomChoice(downloadResultsVerbs),
		randomHex(3),
	)
}

func randomChoice(values []string) string {
	if len(values) == 0 {
		return "run"
	}
	index := 0
	if len(values) > 1 {
		index = int(randomUint32() % uint32(len(values)))
	}
	return values[index]
}

func randomUint32() uint32 {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return uint32(time.Now().UnixNano())
	}
	return uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
}

func randomHex(numBytes int) string {
	if numBytes <= 0 {
		return ""
	}
	buf := make([]byte, numBytes)
	if _, err := rand.Read(buf); err != nil {
		fallback := make([]byte, numBytes)
		for i := range fallback {
			fallback[i] = byte((time.Now().UnixNano() >> (i * 8)) & 0xff)
		}
		buf = fallback
	}
	return hex.EncodeToString(buf)
}

func ensureDownloadDirectoryReady(runDir string) error {
	info, err := os.Stat(runDir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("Run path is not a directory: %s", runDir)
		}
		if _, err := os.Stat(downloadMetadataPath(runDir)); err == nil {
			return nil
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		entries, err := os.ReadDir(runDir)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return fmt.Errorf("Run directory already exists without experiments metadata: %s. Choose a different `name`.", runDir)
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.MkdirAll(runDir, 0o755)
}

func reconcileMetadataWithSpec(runDir string, metadata *downloadRunMetadata, spec downloadResultsSpec) error {
	if spec.ID != nil {
		if metadata.Remote.RunID == nil || *metadata.Remote.RunID != *spec.ID {
			return fmt.Errorf("Run directory %s belongs to remote run %s, not %s", runDir, derefString(metadata.Remote.RunID), *spec.ID)
		}
		inferredType, err := inferDownloadRunType(*spec.ID)
		if err != nil {
			return err
		}
		if metadata.RunType != inferredType {
			return fmt.Errorf("Run directory %s belongs to %s, not %s", runDir, metadata.RunType, inferredType)
		}
	}
	if spec.WorkspaceID != nil {
		if metadata.Remote.WorkspaceID == nil {
			metadata.Remote.WorkspaceID = cloneString(spec.WorkspaceID)
			metadata.RequestFingerprint = fingerprintDownloadRun(metadata.RunType, derefString(metadata.Remote.RunID), metadata.Remote.WorkspaceID)
			return saveDownloadMetadata(runDir, *metadata)
		}
		if *metadata.Remote.WorkspaceID != *spec.WorkspaceID {
			return fmt.Errorf("Run directory %s belongs to workspace %s, not %s", runDir, *metadata.Remote.WorkspaceID, *spec.WorkspaceID)
		}
	}
	if metadata.SchemaVersion != downloadResultsSchemaVersion {
		return fmt.Errorf("Unsupported experiments metadata schema version: %d", metadata.SchemaVersion)
	}
	return nil
}

func newDownloadRunMetadata(name string, runType downloadRunType, runID string, workspaceID *string) downloadRunMetadata {
	return downloadRunMetadata{
		SchemaVersion:      downloadResultsSchemaVersion,
		Name:               name,
		RunType:            runType,
		RequestFingerprint: fingerprintDownloadRun(runType, runID, workspaceID),
		IdempotencyKey:     "download_" + randomHex(16),
		Remote: downloadRemoteState{
			RunID:       cloneString(&runID),
			WorkspaceID: cloneString(workspaceID),
		},
		Pending:       nil,
		CursorAfterID: nil,
	}
}

func fingerprintDownloadRun(runType downloadRunType, runID string, workspaceID *string) string {
	payload := struct {
		RunType     downloadRunType `json:"run_type"`
		RunID       string          `json:"run_id"`
		WorkspaceID *string         `json:"workspace_id"`
	}{
		RunType:     runType,
		RunID:       runID,
		WorkspaceID: cloneString(workspaceID),
	}
	encoded, _ := json.Marshal(payload)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func (m *downloadRunMetadata) UnmarshalJSON(data []byte) error {
	type alias downloadRunMetadata
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for key := range downloadResultsKnownMetadataFields {
		delete(raw, key)
	}

	*m = downloadRunMetadata(decoded)
	if len(raw) > 0 {
		m.ExtraFields = raw
	} else {
		m.ExtraFields = nil
	}
	return nil
}

func (m downloadRunMetadata) MarshalJSON() ([]byte, error) {
	type alias downloadRunMetadata
	known, err := json.Marshal(alias(m))
	if err != nil {
		return nil, err
	}

	var merged map[string]json.RawMessage
	if err := json.Unmarshal(known, &merged); err != nil {
		return nil, err
	}
	for key, value := range m.ExtraFields {
		if _, exists := merged[key]; !exists {
			merged[key] = value
		}
	}
	return json.Marshal(merged)
}

func loadDownloadMetadata(runDir string) (downloadRunMetadata, error) {
	raw, err := os.ReadFile(downloadMetadataPath(runDir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return downloadRunMetadata{}, fmt.Errorf("Run metadata does not exist: %s", downloadMetadataPath(runDir))
		}
		return downloadRunMetadata{}, err
	}
	var metadata downloadRunMetadata
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return downloadRunMetadata{}, fmt.Errorf("Invalid run metadata: %w", err)
	}
	if metadata.SchemaVersion != downloadResultsSchemaVersion {
		return downloadRunMetadata{}, fmt.Errorf("Unsupported experiments metadata schema version: %d", metadata.SchemaVersion)
	}
	if metadata.Name == "" {
		return downloadRunMetadata{}, errors.New("Invalid run metadata.name: required")
	}
	if !isSupportedDownloadRunType(metadata.RunType) {
		return downloadRunMetadata{}, fmt.Errorf("Invalid run metadata.run_type: %q", metadata.RunType)
	}
	if metadata.Pending != nil {
		if metadata.Pending.Kind != downloadPendingKindPredictionArchive && metadata.Pending.Kind != downloadPendingKindResultPage {
			return downloadRunMetadata{}, fmt.Errorf("Invalid run metadata.pending.kind: %q", metadata.Pending.Kind)
		}
	}
	return metadata, nil
}

func saveDownloadMetadata(runDir string, metadata downloadRunMetadata) error {
	payload, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	path := downloadMetadataPath(runDir)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func downloadMetadataPath(runDir string) string {
	return filepath.Join(runDir, downloadResultsMetadataFileName)
}

func isSupportedDownloadRunType(runType downloadRunType) bool {
	if runType == downloadRunTypePrediction {
		return true
	}
	_, ok := downloadPipelineAdapters[runType]
	return ok
}

func inferDownloadRunType(id string) (downloadRunType, error) {
	for _, spec := range downloadResultsRunTypePrefixes {
		if strings.HasPrefix(id, spec.prefix) {
			return spec.runType, nil
		}
	}
	for _, prefix := range downloadResultsUnsupportedPrefixes {
		if strings.HasPrefix(id, prefix) {
			return "", fmt.Errorf("Unsupported ID prefix %q. download-results only supports prediction and pipeline run IDs.", strings.TrimSuffix(prefix, "_"))
		}
	}
	prefix := id
	if idx := strings.Index(prefix, "_"); idx >= 0 {
		prefix = prefix[:idx]
	}
	return "", fmt.Errorf("Could not infer download-results run type from ID prefix %q", prefix)
}

func (e *downloadResultsEngine) waitForPrediction(ctx context.Context, runDir string, metadata *downloadRunMetadata, pollIntervalSeconds float64) error {
	runID := derefString(metadata.Remote.RunID)
	if runID == "" {
		return fmt.Errorf("Run %s has no confirmed remote run ID", runDir)
	}
	e.sink.info(fmt.Sprintf("Waiting for prediction in %s", runDir))

	for {
		response, err := e.getPrediction(ctx, runID, metadata.Remote.WorkspaceID)
		if err != nil {
			return err
		}
		previousStatus := derefString(metadata.Remote.Status)
		updateRunMetadata(metadata, response.Status, response.WorkspaceID, response.StartedAt, response.CompletedAt, nil, nil, response.ErrorCode)
		if err := saveDownloadMetadata(runDir, *metadata); err != nil {
			return err
		}
		if response.Status != previousStatus {
			e.sink.info(fmt.Sprintf("Prediction status: %s", response.Status))
		}

		switch response.Status {
		case "pending", "running":
			e.sink.maybeRunningSummary("Prediction still running...")
			if err := sleepWithContext(ctx, pollIntervalSeconds); err != nil {
				return err
			}
			continue
		case "failed":
			return failureForMetadata(*metadata)
		case "succeeded":
			if response.ArchiveURL == nil || strings.TrimSpace(*response.ArchiveURL) == "" {
				return errors.New("Prediction succeeded but did not return an archive URL")
			}
			archivePath, extractedDir, err := materializationPaths(filepath.Join(runDir, "outputs"), *response.ArchiveURL)
			if err != nil {
				return err
			}
			if !isDownloadMaterialized(archivePath, extractedDir) {
				metadata.Pending = &downloadPendingState{Kind: downloadPendingKindPredictionArchive}
				if err := saveDownloadMetadata(runDir, *metadata); err != nil {
					return err
				}
				if err := e.materializeArchive(ctx, *response.ArchiveURL, archivePath, extractedDir); err != nil {
					return err
				}
				metadata.Pending = nil
				if err := saveDownloadMetadata(runDir, *metadata); err != nil {
					return err
				}
			}
			e.sink.info(fmt.Sprintf("Prediction ready in %s", runDir))
			return nil
		default:
			return fmt.Errorf("Unsupported prediction status %q", response.Status)
		}
	}
}

func (e *downloadResultsEngine) waitForPipeline(ctx context.Context, runDir string, metadata *downloadRunMetadata, pollIntervalSeconds float64) error {
	runID := derefString(metadata.Remote.RunID)
	if runID == "" {
		return fmt.Errorf("Run %s has no confirmed remote run ID", runDir)
	}
	e.sink.info(fmt.Sprintf("Waiting for %s in %s", strings.ReplaceAll(string(metadata.RunType), "_", " "), runDir))

	for {
		response, err := e.getPipeline(ctx, metadata.RunType, runID, metadata.Remote.WorkspaceID)
		if err != nil {
			return err
		}
		previousStatus := derefString(metadata.Remote.Status)
		updateRunMetadata(metadata, response.Status, response.WorkspaceID, response.StartedAt, response.CompletedAt, response.StoppedAt, response.LatestResultID, response.ErrorCode)
		if err := saveDownloadMetadata(runDir, *metadata); err != nil {
			return err
		}
		if response.Status != previousStatus {
			e.sink.info(fmt.Sprintf("Pipeline status: %s", response.Status))
		}

		madeProgress := false
		for {
			if metadata.Pending != nil && metadata.Pending.Kind == downloadPendingKindResultPage {
				if err := e.drainPendingPipelinePage(ctx, runDir, metadata); err != nil {
					return err
				}
				madeProgress = true
				continue
			}
			discovered, err := e.discoverNextPipelinePage(ctx, metadata)
			if err != nil {
				return err
			}
			if discovered {
				if err := saveDownloadMetadata(runDir, *metadata); err != nil {
					return err
				}
				madeProgress = true
				continue
			}
			break
		}

		switch response.Status {
		case "pending", "running":
			if !madeProgress {
				e.sink.maybeRunningSummary(pipelineRunningSummary(metadata.Remote.LatestResultID))
				if err := sleepWithContext(ctx, pollIntervalSeconds); err != nil {
					return err
				}
			}
			continue
		case "failed":
			return failureForMetadata(*metadata)
		case "succeeded", "stopped":
			e.sink.info(fmt.Sprintf("Pipeline ready in %s", runDir))
			return nil
		default:
			return fmt.Errorf("Unsupported pipeline status %q", response.Status)
		}
	}
}

func (e *downloadResultsEngine) discoverNextPipelinePage(ctx context.Context, metadata *downloadRunMetadata) (bool, error) {
	runID := derefString(metadata.Remote.RunID)
	if runID == "" {
		return false, errors.New("Pipeline metadata is missing a remote run ID")
	}
	page, err := e.listPipelineResults(ctx, metadata.RunType, runID, metadata.Remote.WorkspaceID, metadata.CursorAfterID)
	if err != nil {
		return false, err
	}
	if len(page.Results) == 0 {
		return false, nil
	}

	pageLastID := cloneString(page.LastID)
	if pageLastID == nil && len(page.Results) > 0 {
		lastID := page.Results[len(page.Results)-1].ID
		pageLastID = &lastID
	}

	resultIDs := make([]string, 0, len(page.Results))
	for _, result := range page.Results {
		resultIDs = append(resultIDs, result.ID)
	}

	metadata.Pending = &downloadPendingState{
		Kind:       downloadPendingKindResultPage,
		AfterID:    cloneString(metadata.CursorAfterID),
		PageLastID: pageLastID,
		ResultIDs:  resultIDs,
	}
	return true, nil
}

func (e *downloadResultsEngine) drainPendingPipelinePage(ctx context.Context, runDir string, metadata *downloadRunMetadata) error {
	if metadata.Pending == nil || metadata.Pending.Kind != downloadPendingKindResultPage {
		return errors.New("Pipeline metadata does not contain a pending page")
	}
	runID := derefString(metadata.Remote.RunID)
	if runID == "" {
		return errors.New("Pipeline metadata is missing a remote run ID")
	}

	page, err := e.listPipelineResults(ctx, metadata.RunType, runID, metadata.Remote.WorkspaceID, metadata.Pending.AfterID)
	if err != nil {
		return err
	}
	results := make(map[string]downloadPipelineResultInfo, len(page.Results))
	for _, result := range page.Results {
		results[result.ID] = result
	}

	for len(metadata.Pending.ResultIDs) > 0 {
		resultID := metadata.Pending.ResultIDs[0]
		result, ok := results[resultID]
		if !ok {
			return fmt.Errorf("Unable to resume result page; result %s is no longer present", resultID)
		}

		archivePath, extractedDir, err := materializationPaths(filepath.Join(runDir, "results", result.ID), result.ArchiveURL)
		if err != nil {
			return err
		}
		if !isDownloadMaterialized(archivePath, extractedDir) {
			if err := e.materializeArchive(ctx, result.ArchiveURL, archivePath, extractedDir); err != nil {
				return err
			}
		}

		metadata.Pending.ResultIDs = metadata.Pending.ResultIDs[1:]
		if err := saveDownloadMetadata(runDir, *metadata); err != nil {
			return err
		}
	}

	metadata.CursorAfterID = cloneString(metadata.Pending.PageLastID)
	metadata.Pending = nil
	return saveDownloadMetadata(runDir, *metadata)
}

func (e *downloadResultsEngine) materializeArchive(ctx context.Context, archiveURL string, archivePath string, extractedDir string) error {
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(archivePath); errors.Is(err, os.ErrNotExist) {
		e.sink.info(fmt.Sprintf("Downloading archive to %s", archivePath))
		if err := e.downloadArchive(ctx, archiveURL, archivePath); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if info, err := os.Stat(extractedDir); err == nil && info.IsDir() {
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	e.sink.info(fmt.Sprintf("Extracting archive to %s", extractedDir))
	return extractArchive(archivePath, extractedDir)
}

func (e *downloadResultsEngine) downloadArchive(ctx context.Context, archiveURL string, destination string) error {
	partPath := destination + ".part"
	if err := os.Remove(partPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	file, err := os.Create(partPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var response *http.Response
	err = e.client.Get(
		ctx,
		archiveURL,
		nil,
		nil,
		option.WithResponseInto(&response),
		option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			r.Header.Del("Authorization")
			r.Header.Del("x-api-key")
			return next(r)
		}),
	)
	if err != nil {
		return err
	}
	if response == nil {
		return errors.New("archive download returned no response")
	}
	defer response.Body.Close()

	if _, err := io.Copy(file, response.Body); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(partPath, destination)
}

func extractArchive(archivePath string, destination string) error {
	if err := os.RemoveAll(destination); err != nil {
		return err
	}

	baseDir := filepath.Dir(destination)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp(baseDir, filepath.Base(destination)+".tmp-")
	if err != nil {
		return err
	}

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	switch suffix := archiveSuffixFromPath(archivePath); suffix {
	case ".tar.gz", ".tgz":
		err = extractTarGz(archivePath, tempDir)
	case ".tar":
		err = extractTar(archivePath, tempDir)
	case ".zip":
		err = extractZipArchive(archivePath, tempDir)
	default:
		err = fmt.Errorf("Unsupported archive format for %s", archivePath)
	}
	if err != nil {
		cleanup()
		return err
	}

	if err := os.RemoveAll(destination); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tempDir, destination); err != nil {
		cleanup()
		return err
	}
	return nil
}

func extractTarGz(archivePath string, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	return extractTarReader(tar.NewReader(reader), destination)
}

func extractTar(archivePath string, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return extractTarReader(tar.NewReader(file), destination)
}

func extractTarReader(reader *tar.Reader, destination string) error {
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		targetPath, err := joinArchivePath(destination, header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := writeArchiveFile(targetPath, os.FileMode(header.Mode), reader); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unsupported archive entry type %d in %s", header.Typeflag, header.Name)
		}
	}
}

func extractZipArchive(archivePath string, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath, err := joinArchivePath(destination, file.Name)
		if err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.Mode()); err != nil {
				return err
			}
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		if err := writeArchiveFile(targetPath, file.Mode(), rc); err != nil {
			rc.Close()
			return err
		}
		if err := rc.Close(); err != nil {
			return err
		}
	}

	return nil
}

func writeArchiveFile(targetPath string, mode os.FileMode, reader io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, reader); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func joinArchivePath(destination string, entryName string) (string, error) {
	cleaned := filepath.Clean(filepath.FromSlash(entryName))
	if cleaned == "." {
		return destination, nil
	}
	target := filepath.Join(destination, cleaned)
	rel, err := filepath.Rel(destination, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("Archive entry escapes destination: %s", entryName)
	}
	return target, nil
}

func materializationPaths(baseDir string, archiveURL string) (string, string, error) {
	suffix, err := archiveSuffixFromURL(archiveURL)
	if err != nil {
		return "", "", err
	}
	return filepath.Join(baseDir, "archive"+suffix), filepath.Join(baseDir, "files"), nil
}

func archiveSuffixFromURL(rawURL string) (string, error) {
	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return "", err
	}
	name := path.Base(parsed.Path)
	suffix := archiveSuffixFromPath(name)
	if suffix == "" {
		return "", fmt.Errorf("Archive URL does not include a supported file suffix: %s", rawURL)
	}
	return suffix, nil
}

func archiveSuffixFromPath(rawPath string) string {
	lower := strings.ToLower(rawPath)
	for _, suffix := range downloadResultsSupportedArchiveTypes {
		if strings.HasSuffix(lower, suffix) {
			return suffix
		}
	}
	return ""
}

func isDownloadMaterialized(archivePath string, extractedDir string) bool {
	if _, err := os.Stat(archivePath); err != nil {
		return false
	}
	info, err := os.Stat(extractedDir)
	return err == nil && info.IsDir()
}

func updateRunMetadata(metadata *downloadRunMetadata, status string, workspaceID *string, startedAt *string, completedAt *string, stoppedAt *string, latestResultID *string, errorCode *string) {
	if workspaceID != nil {
		metadata.Remote.WorkspaceID = cloneString(workspaceID)
	}
	metadata.Remote.Status = cloneString(&status)
	metadata.Remote.StartedAt = cloneString(startedAt)
	metadata.Remote.CompletedAt = cloneString(completedAt)
	metadata.Remote.StoppedAt = cloneString(stoppedAt)
	metadata.Remote.LatestResultID = cloneString(latestResultID)
	metadata.Remote.ErrorCode = cloneString(errorCode)
	if metadata.Remote.RunID != nil {
		metadata.RequestFingerprint = fingerprintDownloadRun(metadata.RunType, *metadata.Remote.RunID, metadata.Remote.WorkspaceID)
	}
}

func failureForMetadata(metadata downloadRunMetadata) error {
	runID := metadata.Name
	if metadata.Remote.RunID != nil && *metadata.Remote.RunID != "" {
		runID = *metadata.Remote.RunID
	}
	if metadata.Remote.ErrorCode != nil && *metadata.Remote.ErrorCode != "" {
		return fmt.Errorf("Run %s failed with error code %s", runID, *metadata.Remote.ErrorCode)
	}
	return fmt.Errorf("Run %s failed", runID)
}

func pipelineRunningSummary(latestResultID *string) string {
	if latestResultID == nil || *latestResultID == "" {
		return "Pipeline still running..."
	}
	return fmt.Sprintf("Pipeline still running... latest result: %s", *latestResultID)
}

func (s *downloadResultsSink) info(message string) {
	if !s.verbose {
		return
	}
	_, _ = fmt.Fprintln(s.writer, message)
}

func (s *downloadResultsSink) maybeRunningSummary(message string) {
	if !s.verbose {
		return
	}
	now := time.Now()
	if !s.lastRunningSummaryAt.IsZero() && now.Sub(s.lastRunningSummaryAt) < downloadResultsSummaryIntervalSec {
		return
	}
	s.lastRunningSummaryAt = now
	s.info(message)
}

func (e *downloadResultsEngine) getPrediction(ctx context.Context, runID string, workspaceID *string) (downloadPredictionRunInfo, error) {
	params := boltzcompute.PredictionStructureAndBindingGetParams{}
	setOptionalStringOpt(&params.WorkspaceID, workspaceID)
	response, err := e.client.Predictions.StructureAndBinding.Get(ctx, runID, params)
	if err != nil {
		return downloadPredictionRunInfo{}, err
	}
	return downloadPredictionRunInfo{
		Status:      string(response.Status),
		WorkspaceID: normalizedStringPointer(response.WorkspaceID),
		StartedAt:   normalizedTimePointer(response.StartedAt),
		CompletedAt: normalizedTimePointer(response.CompletedAt),
		ErrorCode:   normalizedStringPointer(response.Error.Code),
		ArchiveURL:  normalizedStringPointer(response.Output.Archive.URL),
	}, nil
}

func (e *downloadResultsEngine) getPipeline(ctx context.Context, runType downloadRunType, runID string, workspaceID *string) (downloadPipelineRunInfo, error) {
	adapter, err := pipelineAdapterFor(runType)
	if err != nil {
		return downloadPipelineRunInfo{}, err
	}
	return adapter.get(ctx, e.client, runID, workspaceID)
}

func (e *downloadResultsEngine) listPipelineResults(ctx context.Context, runType downloadRunType, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error) {
	adapter, err := pipelineAdapterFor(runType)
	if err != nil {
		return downloadPipelinePage{}, err
	}
	return adapter.listResults(ctx, e.client, runID, workspaceID, afterID)
}

func newDownloadPipelineRunInfo(status string, workspaceID string, startedAt time.Time, completedAt time.Time, stoppedAt time.Time, latestResultID string, errorCode string) downloadPipelineRunInfo {
	return downloadPipelineRunInfo{
		Status:         status,
		WorkspaceID:    normalizedStringPointer(workspaceID),
		StartedAt:      normalizedTimePointer(startedAt),
		CompletedAt:    normalizedTimePointer(completedAt),
		StoppedAt:      normalizedTimePointer(stoppedAt),
		LatestResultID: normalizedStringPointer(latestResultID),
		ErrorCode:      normalizedStringPointer(errorCode),
	}
}

func normalizePipelinePage[T any](page *pagination.CursorPage[T], normalize func(T) downloadPipelineResultInfo) downloadPipelinePage {
	results := make([]downloadPipelineResultInfo, 0, len(page.Data))
	for _, result := range page.Data {
		results = append(results, normalize(result))
	}
	return downloadPipelinePage{
		Results: results,
		LastID:  trimOptionalString(page.LastID),
	}
}

func pipelineAdapterFor(runType downloadRunType) (downloadPipelineAdapter, error) {
	adapter, ok := downloadPipelineAdapters[runType]
	if !ok {
		return downloadPipelineAdapter{}, fmt.Errorf("Unsupported run type %q", runType)
	}
	return adapter, nil
}

func setOptionalStringOpt(target *param.Opt[string], value *string) {
	if value == nil {
		return
	}
	*target = param.NewOpt(*value)
}

func normalizedStringPointer(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizedTimePointer(value time.Time) *string {
	if value.IsZero() {
		return nil
	}
	formatted := value.Format(time.RFC3339Nano)
	return &formatted
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func trimOptionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func sleepWithContext(ctx context.Context, seconds float64) error {
	if seconds <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(time.Duration(seconds * float64(time.Second)))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
