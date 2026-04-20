// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-compute-api-cli/internal/mocktest"
)

func TestAdminAPIKeysCreate(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"admin:api-keys", "create",
			"--name", "x",
			"--allowed-ip", "string",
			"--expires-in-days", "1",
			"--mode", "live",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"name: x\n" +
			"allowed_ips:\n" +
			"  - string\n" +
			"expires_in_days: 1\n" +
			"mode: live\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"admin:api-keys", "create",
		)
	})
}

func TestAdminAPIKeysList(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"admin:api-keys", "list",
			"--max-items", "10",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--limit", "1",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestAdminAPIKeysRevoke(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"admin:api-keys", "revoke",
			"--api-key-id", "api_key_id",
		)
	})
}
