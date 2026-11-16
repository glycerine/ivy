package gia_test

import (
	"github.com/glycerine/abcgo/aig/gia"
	"os/exec"
	"strings"
	"testing"
)

func TestMappingDSDOriginalBalancingContracts(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_dsd_balance_review.c")
	output, err := exec.Command(executable).CombinedOutput()
	if err != nil {
		t.Fatalf("original balancing: %v %s", err, output)
	}
	if !strings.HasPrefix(string(output), "balance 0 0 2 0 2 2 2 ") {
		t.Fatalf("original graph/timing cost contract changed: %s", output)
	}
	t.Logf("original balance contracts:\n%s", output)
}
