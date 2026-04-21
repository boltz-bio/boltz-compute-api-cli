package requestflag

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCLIArgStructuredIncludeMap(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "input.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"entities":[{"type":"protein","value":"SEQ"}]}`), 0644))

	value, err := parseCLIArg[map[string]any]("@json://" + path)
	require.NoError(t, err)

	entities, ok := value["entities"].([]any)
	require.True(t, ok)
	require.Len(t, entities, 1)
	entity, ok := entities[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "protein", entity["type"])
	require.Equal(t, "SEQ", entity["value"])
}

func TestParseCLIArgStructuredIncludeSlice(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "molecules.yaml")
	require.NoError(t, os.WriteFile(path, []byte("- smiles: CCO\n- smiles: CCN\n"), 0644))

	value, err := parseCLIArg[[]map[string]any]("@yaml://" + path)
	require.NoError(t, err)
	require.Len(t, value, 2)
	require.Equal(t, "CCO", value[0]["smiles"])
	require.Equal(t, "CCN", value[1]["smiles"])
}

func TestParseCLIArgStructuredIncludeParseError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "broken.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"entities":[}`), 0644))

	_, err := parseCLIArg[map[string]any]("@json://" + path)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse @json://")
	require.Contains(t, err.Error(), path)
}
