// Custom CLI extension code. Not generated.
package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratedFilesOnlyUseApprovedCustomizationSeams(t *testing.T) {
	t.Parallel()

	allowedSeams := map[string]struct{}{
		filepath.Join("cmd", "boltz-api", "main.go"): {},
		filepath.Join("pkg", "cmd", "cmd.go"):        {},
		filepath.Join("pkg", "cmd", "cmdutil.go"):    {},
	}
	disallowedSnippets := []string{
		"internal/authconfig",
		"internal/authmode",
		"internal/autherror",
		"internal/authstore",
		"internal/oauthclient",
		"ApplyCustomizations(",
		"additionalRequestOptions(",
		"WriteCommandErrorOutput(",
	}

	for _, relDir := range []string{filepath.Join("cmd", "boltz-api"), filepath.Join("pkg", "cmd")} {
		dir := filepath.Join(repoRoot(t), relDir)
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
				continue
			}

			relPath := filepath.Join(relDir, entry.Name())
			body, err := os.ReadFile(filepath.Join(repoRoot(t), relPath))
			require.NoError(t, err)

			if !strings.HasPrefix(string(body), "// File generated from our OpenAPI spec by Stainless.") {
				continue
			}
			if _, ok := allowedSeams[relPath]; ok {
				continue
			}

			for _, snippet := range disallowedSnippets {
				require.NotContainsf(t, string(body), snippet, "generated file %s should not contain custom seam %q", relPath, snippet)
			}
		}
	}
}

func TestPkgCmdUsesPredictableCustomFileNames(t *testing.T) {
	t.Parallel()

	allowedNonGenerated := map[string]struct{}{
		"cmdutil.go":          {},
		"cmdutil_test.go":     {},
		"cmdutil_unix.go":     {},
		"cmdutil_windows.go":  {},
		"flagoptions.go":      {},
		"flagoptions_test.go": {},
		"suggest.go":          {},
	}

	dir := filepath.Join(repoRoot(t), "pkg", "cmd")
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}

		body, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		require.NoError(t, err)

		if strings.HasPrefix(string(body), "// File generated from our OpenAPI spec by Stainless.") {
			continue
		}
		if strings.HasPrefix(entry.Name(), "custom_") {
			require.Truef(t, strings.HasPrefix(string(body), "// Custom CLI extension code. Not generated."), "custom file %s should carry the standard marker comment", entry.Name())
			continue
		}
		if _, ok := allowedNonGenerated[entry.Name()]; ok {
			continue
		}

		t.Fatalf("non-generated pkg/cmd file %s should either use the custom_ prefix or be in the explicit runtime allowlist", entry.Name())
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}
