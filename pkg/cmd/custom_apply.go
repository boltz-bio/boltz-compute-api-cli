// Custom CLI extension code. Not generated.
package cmd

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/boltz-bio/boltz-compute-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-compute-api-cli/internal/requestflag"
	"github.com/urfave/cli/v3"
)

const (
	customizationsAppliedKey = "command_customizations_applied"
	transformUsage           = "The GJSON transformation for data output. For paginated or streamed list commands, this runs on each emitted item except in --format raw, where it runs on the full response page; use jq for whole-list reshaping."
)

func ApplyCustomizations(app *cli.Command) {
	if app == nil {
		return
	}
	if app.Metadata == nil {
		app.Metadata = map[string]any{}
	}
	if applied, _ := app.Metadata[customizationsAppliedKey].(bool); applied {
		return
	}
	app.Metadata[customizationsAppliedKey] = true

	for _, flag := range authFlags() {
		if hasFlag(app.Flags, flag) {
			continue
		}
		app.Flags = append(app.Flags, flag)
	}

	if !hasCommand(app.Commands, authCommand.Name) {
		app.Commands = append(app.Commands, authCommand)
	}

	if !hasCommand(app.Commands, downloadResultsCommand.Name) {
		app.Commands = append(app.Commands, downloadResultsCommand)
	}

	if !hasCommand(app.Commands, downloadStatusCommand.Name) {
		app.Commands = append(app.Commands, downloadStatusCommand)
	}

	customizeCommandTree(app)
}

func authFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "auth-issuer-url",
			Value:   authconfig.DefaultIssuerURL,
			Usage:   "OIDC issuer URL for OAuth login and bearer-token refresh",
			Sources: cli.EnvVars(authconfig.EnvAuthIssuerURL),
		},
		&cli.StringFlag{
			Name:    "auth-client-id",
			Value:   authconfig.DefaultClientID,
			Usage:   "OAuth client ID for the public client",
			Sources: cli.EnvVars(authconfig.EnvAuthClientID),
		},
		&cli.StringSliceFlag{
			Name:    "auth-scope",
			Usage:   "OAuth scope to request. Repeat the flag or use a comma-separated env value.",
			Sources: cli.EnvVars(authconfig.EnvAuthScope),
		},
		&cli.StringFlag{
			Name:    "auth-audience",
			Value:   authconfig.DefaultAudience,
			Usage:   "OAuth audience/resource to request during login",
			Sources: cli.EnvVars(authconfig.EnvAuthAudience),
		},
		&cli.StringFlag{
			Name:    "auth-authorization-url",
			Usage:   "Override the discovered authorization endpoint",
			Sources: cli.EnvVars(authconfig.EnvAuthAuthorizationURL),
		},
		&cli.StringFlag{
			Name:    "auth-token-url",
			Usage:   "Override the discovered token endpoint",
			Sources: cli.EnvVars(authconfig.EnvAuthTokenURL),
		},
		&cli.StringFlag{
			Name:    "auth-userinfo-url",
			Usage:   "Override the discovered userinfo endpoint",
			Sources: cli.EnvVars(authconfig.EnvAuthUserInfoURL),
		},
		&cli.StringFlag{
			Name:    "auth-revocation-url",
			Usage:   "Override the discovered revocation endpoint",
			Sources: cli.EnvVars(authconfig.EnvAuthRevocationURL),
		},
		&cli.StringFlag{
			Name:    "org",
			Usage:   "Local organization selection used by auth commands and future org context",
			Sources: cli.EnvVars(authconfig.EnvOrg),
		},
		&cli.BoolFlag{
			Name:    "no-browser",
			Usage:   "Print the OAuth URL without trying to open a browser",
			Sources: cli.EnvVars(authconfig.EnvNoBrowser),
		},
		&cli.IntFlag{
			Name:    "listen-port",
			Value:   authconfig.DefaultListenPort,
			Usage:   "Bind the OAuth loopback callback listener to this port",
			Sources: cli.EnvVars(authconfig.EnvListenPort),
		},
	}
}

func hasCommand(commands []*cli.Command, name string) bool {
	for _, command := range commands {
		if command != nil && command.Name == name {
			return true
		}
	}
	return false
}

func hasFlag(flags []cli.Flag, candidate cli.Flag) bool {
	for _, flag := range flags {
		if slices.Equal(flag.Names(), candidate.Names()) {
			return true
		}
	}
	return false
}

func customizeCommandTree(cmd *cli.Command) {
	if cmd == nil {
		return
	}

	for _, flag := range cmd.Flags {
		maybeCustomizeRootFlag(flag)
		maybeAnnotateRepeatableArrayFlag(flag)
	}

	for _, child := range cmd.Commands {
		customizeCommandTree(child)
	}
}

func maybeCustomizeRootFlag(flag cli.Flag) {
	if canonicalFlagName(flag) != "transform" {
		return
	}
	setFlagStringField(flag, "Usage", transformUsage)
}

func maybeAnnotateRepeatableArrayFlag(flag cli.Flag) {
	inReq, ok := flag.(requestflag.InRequest)
	if !ok || inReq.GetBodyPath() == "" || !flagDefaultKindIs(flag, reflect.Slice) {
		return
	}

	flagName := canonicalFlagName(flag)
	bodyField := inReq.GetBodyPath()
	if !isSingularPluralRename(flagName, bodyField) {
		return
	}

	usage, ok := flagStringField(flag, "Usage")
	if !ok {
		return
	}

	note := fmt.Sprintf("Repeat --%s to add entries. When piping JSON or YAML, use the body field %s.", flagName, bodyField)
	if strings.Contains(usage, note) {
		return
	}
	if usage == "" {
		setFlagStringField(flag, "Usage", note)
		return
	}
	setFlagStringField(flag, "Usage", usage+" "+note)
}

func canonicalFlagName(flag cli.Flag) string {
	names := flag.Names()
	if len(names) == 0 {
		return ""
	}
	return names[0]
}

func flagDefaultKindIs(flag cli.Flag, want reflect.Kind) bool {
	field, ok := flagField(flag, "Default")
	return ok && field.Kind() == want
}

func flagStringField(flag cli.Flag, name string) (string, bool) {
	field, ok := flagField(flag, name)
	if !ok || field.Kind() != reflect.String {
		return "", false
	}
	return field.String(), true
}

func setFlagStringField(flag cli.Flag, name, value string) bool {
	field, ok := flagField(flag, name)
	if !ok || field.Kind() != reflect.String || !field.CanSet() {
		return false
	}
	field.SetString(value)
	return true
}

func flagField(flag cli.Flag, name string) (reflect.Value, bool) {
	v := reflect.ValueOf(flag)
	if !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() {
		return reflect.Value{}, false
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	field := v.FieldByName(name)
	if !field.IsValid() {
		return reflect.Value{}, false
	}
	return field, true
}

func isSingularPluralRename(flagName, bodyField string) bool {
	normalizedFlag := strings.ReplaceAll(flagName, "-", "_")
	return pluralize(normalizedFlag) == bodyField
}

func pluralize(s string) string {
	switch {
	case len(s) >= 2 && strings.HasSuffix(s, "y") && !isASCIIvowel(s[len(s)-2]):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(s, "s"), strings.HasSuffix(s, "x"), strings.HasSuffix(s, "z"),
		strings.HasSuffix(s, "ch"), strings.HasSuffix(s, "sh"):
		return s + "es"
	default:
		return s + "s"
	}
}

func isASCIIvowel(b byte) bool {
	switch b {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}
