// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-compute-api-cli/internal/mocktest"
	"github.com/boltz-bio/boltz-compute-api-cli/internal/requestflag"
)

func TestSmallMoleculeDesignRetrieve(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "retrieve",
			"--run-id", "run_id",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestSmallMoleculeDesignList(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "list",
			"--max-items", "10",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--limit", "1",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestSmallMoleculeDesignDeleteData(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "delete-data",
			"--run-id", "run_id",
		)
	})
}

func TestSmallMoleculeDesignEstimateCost(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "estimate-cost",
			"--num-molecules", "10",
			"--target", "{entities: [{chain_ids: [string], modifications: [{residue_index: 0, type: ccd, value: value}], type: protein, value: value, cyclic: true}], pocket_residues: {A: [42, 43, 44, 67, 68, 69]}, reference_ligands: [string]}",
			"--chemical-space", "enamine_real",
			"--idempotency-key", "idempotency_key",
			"--molecule-filters", "{boltz_smarts_catalog_filter_level: recommended, custom_filters: [{max_hba: 0, max_hbd: 0, max_logp: 0, max_mw: 0, type: lipinski_filter, allow_single_violation: true}]}",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(smallMoleculeDesignEstimateCost)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "estimate-cost",
			"--num-molecules", "10",
			"--target.entities", "[{chain_ids: [string], modifications: [{residue_index: 0, type: ccd, value: value}], type: protein, value: value, cyclic: true}]",
			"--target.pocket-residues", "{A: [42, 43, 44, 67, 68, 69]}",
			"--target.reference-ligands", "[string]",
			"--chemical-space", "enamine_real",
			"--idempotency-key", "idempotency_key",
			"--molecule-filters.boltz-smarts-catalog-filter-level", "recommended",
			"--molecule-filters.custom-filters", "[{max_hba: 0, max_hbd: 0, max_logp: 0, max_mw: 0, type: lipinski_filter, allow_single_violation: true}]",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"num_molecules: 10\n" +
			"target:\n" +
			"  entities:\n" +
			"    - chain_ids:\n" +
			"        - string\n" +
			"      modifications:\n" +
			"        - residue_index: 0\n" +
			"          type: ccd\n" +
			"          value: value\n" +
			"      type: protein\n" +
			"      value: value\n" +
			"      cyclic: true\n" +
			"  pocket_residues:\n" +
			"    A:\n" +
			"      - 42\n" +
			"      - 43\n" +
			"      - 44\n" +
			"      - 67\n" +
			"      - 68\n" +
			"      - 69\n" +
			"  reference_ligands:\n" +
			"    - string\n" +
			"chemical_space: enamine_real\n" +
			"idempotency_key: idempotency_key\n" +
			"molecule_filters:\n" +
			"  boltz_smarts_catalog_filter_level: recommended\n" +
			"  custom_filters:\n" +
			"    - max_hba: 0\n" +
			"      max_hbd: 0\n" +
			"      max_logp: 0\n" +
			"      max_mw: 0\n" +
			"      type: lipinski_filter\n" +
			"      allow_single_violation: true\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"small-molecule:design", "estimate-cost",
		)
	})
}

func TestSmallMoleculeDesignListResults(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "list-results",
			"--max-items", "10",
			"--run-id", "run_id",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--limit", "1",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestSmallMoleculeDesignStart(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "start",
			"--num-molecules", "10",
			"--target", "{entities: [{chain_ids: [string], modifications: [{residue_index: 0, type: ccd, value: value}], type: protein, value: value, cyclic: true}], pocket_residues: {A: [42, 43, 44, 67, 68, 69]}, reference_ligands: [string]}",
			"--chemical-space", "enamine_real",
			"--idempotency-key", "idempotency_key",
			"--molecule-filters", "{boltz_smarts_catalog_filter_level: recommended, custom_filters: [{max_hba: 0, max_hbd: 0, max_logp: 0, max_mw: 0, type: lipinski_filter, allow_single_violation: true}]}",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(smallMoleculeDesignStart)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "start",
			"--num-molecules", "10",
			"--target.entities", "[{chain_ids: [string], modifications: [{residue_index: 0, type: ccd, value: value}], type: protein, value: value, cyclic: true}]",
			"--target.pocket-residues", "{A: [42, 43, 44, 67, 68, 69]}",
			"--target.reference-ligands", "[string]",
			"--chemical-space", "enamine_real",
			"--idempotency-key", "idempotency_key",
			"--molecule-filters.boltz-smarts-catalog-filter-level", "recommended",
			"--molecule-filters.custom-filters", "[{max_hba: 0, max_hbd: 0, max_logp: 0, max_mw: 0, type: lipinski_filter, allow_single_violation: true}]",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"num_molecules: 10\n" +
			"target:\n" +
			"  entities:\n" +
			"    - chain_ids:\n" +
			"        - string\n" +
			"      modifications:\n" +
			"        - residue_index: 0\n" +
			"          type: ccd\n" +
			"          value: value\n" +
			"      type: protein\n" +
			"      value: value\n" +
			"      cyclic: true\n" +
			"  pocket_residues:\n" +
			"    A:\n" +
			"      - 42\n" +
			"      - 43\n" +
			"      - 44\n" +
			"      - 67\n" +
			"      - 68\n" +
			"      - 69\n" +
			"  reference_ligands:\n" +
			"    - string\n" +
			"chemical_space: enamine_real\n" +
			"idempotency_key: idempotency_key\n" +
			"molecule_filters:\n" +
			"  boltz_smarts_catalog_filter_level: recommended\n" +
			"  custom_filters:\n" +
			"    - max_hba: 0\n" +
			"      max_hbd: 0\n" +
			"      max_logp: 0\n" +
			"      max_mw: 0\n" +
			"      type: lipinski_filter\n" +
			"      allow_single_violation: true\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"small-molecule:design", "start",
		)
	})
}

func TestSmallMoleculeDesignStop(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "stop",
			"--run-id", "run_id",
		)
	})
}
