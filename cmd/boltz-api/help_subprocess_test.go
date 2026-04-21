package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelpOutputIncludesRepeatableArrayGuidance(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	smallMolecule := runCLI(t, binary, env, "small-molecule:library-screen", "start", "--help")
	require.Equal(t, 0, smallMolecule.ExitCode, smallMolecule.Stderr)
	require.Contains(t, smallMolecule.Stdout, "Repeat --molecule to add entries.")
	require.Contains(t, smallMolecule.Stdout, "use the body field molecules")

	protein := runCLI(t, binary, env, "protein:library-screen", "start", "--help")
	require.Equal(t, 0, protein.ExitCode, protein.Stderr)
	require.Contains(t, protein.Stdout, "Repeat --protein to add entries.")
	require.Contains(t, protein.Stdout, "use the body field proteins")
}

func TestStructuredIncludeObjectFlagSubprocess(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.json")
	require.NoError(t, os.WriteFile(inputPath, []byte(`{"entities":[{"type":"protein","value":"SEQ","modifications":[]}]}`), 0644))

	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		capturedBody, err = io.ReadAll(r.Body)
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	result := runCLI(
		t,
		binary,
		env,
		"--api-key", "test-key",
		"--base-url", server.URL,
		"--format", "raw",
		"predictions:structure-and-binding", "estimate-cost",
		"--input", "@json://"+inputPath,
		"--model", "boltz-2.1",
	)
	require.Equal(t, 0, result.ExitCode, result.Stderr)

	var body map[string]any
	require.NoError(t, json.Unmarshal(capturedBody, &body))

	input, ok := body["input"].(map[string]any)
	require.True(t, ok)
	entities, ok := input["entities"].([]any)
	require.True(t, ok)
	require.Len(t, entities, 1)
	entity, ok := entities[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "protein", entity["type"])
	require.Equal(t, "SEQ", entity["value"])
}

func TestUniversalIDAliasParsesOnRetrieveCommands(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	var requestPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	result := runCLI(
		t,
		binary,
		env,
		"--api-key", "test-key",
		"--base-url", server.URL,
		"--format", "raw",
		"small-molecule:design", "retrieve",
		"--id", "run_123",
	)
	require.Equal(t, 0, result.ExitCode, result.Stderr)
	require.NotEmpty(t, requestPath)
	require.Contains(t, requestPath, "run_123")
	require.NotContains(t, result.Stderr, "flag provided but not defined")
}
