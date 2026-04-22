// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/boltz-bio/boltz-compute-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-compute-api-cli/internal/authstore"
	"github.com/boltz-bio/boltz-compute-api-cli/internal/requestflag"
	githubcomboltzbioboltzcomputeapigo "github.com/boltz-bio/boltz-compute-api-go"
	"github.com/boltz-bio/boltz-compute-api-go/option"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestRequestOptionsKeepAPIKeyMode(t *testing.T) {
	var gotAPIKey string
	var gotAuthorization string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("x-api-key")
		gotAuthorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cmd := parsedTestCommand(t,
		"--base-url", server.URL,
		"--api-key", "api-key-123",
		"call",
	)

	client := githubcomboltzbioboltzcomputeapigo.NewClient(getDefaultRequestOptions(cmd)...)
	var result map[string]any
	require.NoError(t, client.Get(context.Background(), "/check", nil, &result))
	require.Equal(t, "api-key-123", gotAPIKey)
	require.Empty(t, gotAuthorization)
}

func TestRequestOptionsInjectBearerTokenAndRemoveInheritedAPIKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-123",
		Audience:  authconfig.DefaultAudience,
		Scopes:    []string{"openid", "profile"},
	}))
	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "client-123",
		Audience:    authconfig.DefaultAudience,
		Scopes:      []string{"openid", "profile"},
		AccessToken: "oauth-access",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(10 * time.Minute),
	}))

	var gotAPIKey string
	var gotAuthorization string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("x-api-key")
		gotAuthorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cmd := parsedTestCommand(t,
		"--base-url", server.URL,
		"call",
	)

	options := append(getDefaultRequestOptions(cmd), option.WithAPIKey("sdk-default-key"))
	client := githubcomboltzbioboltzcomputeapigo.NewClient(options...)
	var result map[string]any
	require.NoError(t, client.Get(context.Background(), "/check", nil, &result))
	require.Empty(t, gotAPIKey)
	require.Equal(t, "Bearer oauth-access", gotAuthorization)
}

func parsedTestCommand(t *testing.T, args ...string) *cli.Command {
	t.Helper()

	var captured *cli.Command
	root := &cli.Command{
		Name: "boltz-api",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "base-url"},
			&requestflag.Flag[string]{
				Name:    "api-key",
				Sources: cli.EnvVars("BOLTZ_COMPUTE_API_KEY"),
			},
		},
		Commands: []*cli.Command{
			{
				Name: "call",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					_ = ctx
					captured = cmd
					return nil
				},
			},
		},
	}
	root.Flags = append(root.Flags, authFlags()...)

	runArgs := append([]string{"boltz-api"}, args...)
	require.NoError(t, root.Run(context.Background(), runArgs))
	require.NotNil(t, captured)
	return captured
}
