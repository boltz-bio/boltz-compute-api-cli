package cmd

import (
	"slices"

	"github.com/boltz-bio/boltz-compute-api-cli/internal/authconfig"
	"github.com/urfave/cli/v3"
)

const customizationsAppliedKey = "auth_customizations_applied"

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
}

func authFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "auth-issuer-url",
			Usage:   "OIDC issuer URL for OAuth login and bearer-token refresh",
			Sources: cli.EnvVars(authconfig.EnvAuthIssuerURL),
		},
		&cli.StringFlag{
			Name:    "auth-client-id",
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
			Usage:   "Optional OAuth audience to request during login",
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
