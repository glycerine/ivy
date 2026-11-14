package abci

import "testing"

func TestAbcNtkEcoMatchesEmptyCExperiment(t *testing.T) {
	// The C entry point is intentionally empty and does not open its file list.
	Abc_NtkEco([]string{"missing-network.blif", "missing-library.genlib", "missing-script"})
}
