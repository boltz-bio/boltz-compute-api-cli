// Custom CLI extension code. Not generated.
package cmd

import (
	"net/http"
	"strings"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/authmode"
	"github.com/boltz-bio/boltz-api-go/option"
	"github.com/urfave/cli/v3"
)

func additionalRequestOptions(cmd *cli.Command) []option.RequestOption {
	root := cmd
	if root != nil {
		if resolvedRoot := root.Root(); resolvedRoot != nil {
			root = resolvedRoot
		}
	}
	if root == nil {
		return nil
	}

	return []option.RequestOption{
		option.WithMiddleware(func(r *http.Request, mn option.MiddlewareNext) (*http.Response, error) {
			resolved, err := authconfig.Resolve(root)
			if err != nil {
				return nil, authmode.WrapConfigError(err)
			}

			auth, err := authmode.Resolve(r.Context(), resolved)
			if err != nil {
				return nil, err
			}

			if auth.Mode == authmode.ModeOAuth {
				r.Header.Del("x-api-key")
				r.Header.Set("Authorization", "Bearer "+auth.AccessToken)
				if selectedOrg := strings.TrimSpace(resolved.SelectedOrg); selectedOrg != "" {
					r.Header.Set("X-Boltz-Organization-Id", selectedOrg)
				}
			}

			return mn(r)
		}),
	}
}
