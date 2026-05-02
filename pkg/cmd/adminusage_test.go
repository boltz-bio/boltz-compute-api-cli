// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/mocktest"
)

func TestAdminUsageList(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"admin:usage", "list",
			"--max-items", "10",
			"--ending-at", "'2019-12-27T18:11:19.117Z'",
			"--starting-at", "'2019-12-27T18:11:19.117Z'",
			"--window-size", "HOUR",
			"--applications", "structure_and_binding",
			"--group-by", "workspace_id",
			"--limit", "1",
			"--page", "page",
			"--workspace-ids", "string",
		)
	})
}
