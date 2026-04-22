// Custom CLI extension code. Not generated.
package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
	docs "github.com/urfave/cli-docs/v3"
	"github.com/urfave/cli/v3"

	"github.com/boltz-bio/boltz-compute-api-cli/internal/requestflag"
)

func TestApplyCustomizationsPreservesNativeIDFlags(t *testing.T) {
	t.Parallel()

	ApplyCustomizations(Command)

	for _, tc := range []struct {
		path []string
	}{
		{[]string{"small-molecule:design", "retrieve"}},
		{[]string{"small-molecule:library-screen", "list-results"}},
		{[]string{"protein:design", "stop"}},
		{[]string{"protein:library-screen", "delete-data"}},
	} {
		cmd := mustFindCommand(t, Command, tc.path...)
		mustFindFlag(t, cmd, "id")
		require.Nil(t, findFlag(cmd, "run-id"))
		require.Nil(t, findFlag(cmd, "screen-id"))
	}
}

func TestApplyCustomizationsIsIdempotent(t *testing.T) {
	t.Parallel()

	root := &cli.Command{
		Name: "boltz-api",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "transform"},
		},
	}

	ApplyCustomizations(root)
	ApplyCustomizations(root)

	require.Len(t, root.Commands, 2)
	require.Len(t, root.Flags, 12)
	require.Equal(t, transformUsage, usageForFlag(t, mustFindFlag(t, root, "transform")))
}

func TestApplyCustomizationsAnnotatesRepeatableArrayFlags(t *testing.T) {
	t.Parallel()

	ApplyCustomizations(Command)

	moleculeFlag := mustFindFlag(t, mustFindCommand(t, Command, "small-molecule:library-screen", "start"), "molecule")
	require.Contains(t, usageForFlag(t, moleculeFlag), "Repeat --molecule to add entries.")
	require.Contains(t, usageForFlag(t, moleculeFlag), "use the body field molecules")

	proteinFlag := mustFindFlag(t, mustFindCommand(t, Command, "protein:library-screen", "start"), "protein")
	require.Contains(t, usageForFlag(t, proteinFlag), "Repeat --protein to add entries.")
	require.Contains(t, usageForFlag(t, proteinFlag), "use the body field proteins")
}

func TestTransformUsageMentionsPerItemListBehavior(t *testing.T) {
	t.Parallel()

	ApplyCustomizations(Command)

	transformFlag := mustFindFlag(t, Command, "transform")
	require.Contains(t, usageForFlag(t, transformFlag), "runs on each emitted item")
	require.Contains(t, usageForFlag(t, transformFlag), "--format raw")
	require.Contains(t, usageForFlag(t, transformFlag), "jq")
}

func TestManpageIncludesRepeatableArrayFlagGuidance(t *testing.T) {
	t.Parallel()

	ApplyCustomizations(Command)

	manpage, err := docs.ToManWithSection(Command, 1)
	require.NoError(t, err)
	require.Contains(t, manpage, "Repeat --molecule to add entries")
	require.Contains(t, manpage, "use the body field molecules")
	require.Contains(t, manpage, "Repeat --protein to add entries")
	require.Contains(t, manpage, "use the body field proteins")
}

func TestRepeatedSingularFlagsPopulatePluralBodyField(t *testing.T) {
	t.Parallel()

	moleculeFlag := &requestflag.Flag[[]map[string]any]{
		Name:     "molecule",
		BodyPath: "molecules",
	}
	require.NoError(t, moleculeFlag.PreParse())
	require.NoError(t, moleculeFlag.Set("molecule", "{smiles: CCO}"))
	require.NoError(t, moleculeFlag.Set("molecule", "{smiles: CCN}"))

	cmd := &cli.Command{
		Name:  "start",
		Flags: []cli.Flag{moleculeFlag},
	}

	body, ok := requestflag.ExtractRequestContents(cmd).Body.(map[string]any)
	require.True(t, ok)

	molecules, ok := body["molecules"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, molecules, 2)
	require.Equal(t, "CCO", molecules[0]["smiles"])
	require.Equal(t, "CCN", molecules[1]["smiles"])
}

func mustFindCommand(t *testing.T, root *cli.Command, path ...string) *cli.Command {
	t.Helper()

	cmd := root
	for _, part := range path {
		var next *cli.Command
		for _, child := range cmd.Commands {
			if child != nil && child.Name == part {
				next = child
				break
			}
		}
		require.NotNilf(t, next, "command %q not found under %q", part, cmd.Name)
		cmd = next
	}
	return cmd
}

func mustFindFlag(t *testing.T, cmd *cli.Command, name string) cli.Flag {
	t.Helper()

	flag := findFlag(cmd, name)
	if flag != nil {
		return flag
	}
	t.Fatalf("flag %q not found on command %q", name, cmd.Name)
	return nil
}

func findFlag(cmd *cli.Command, name string) cli.Flag {
	for _, flag := range cmd.Flags {
		if len(flag.Names()) > 0 && flag.Names()[0] == name {
			return flag
		}
	}
	return nil
}

func usageForFlag(t *testing.T, flag cli.Flag) string {
	t.Helper()

	docFlag, ok := flag.(interface{ GetUsage() string })
	require.True(t, ok)
	return docFlag.GetUsage()
}
