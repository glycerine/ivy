package aig

import (
	"os"
	"strings"
	"testing"
)

func TestProcedureDemoReportsDumpFailure(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("aig_test_file.blif", 0700); err != nil {
		t.Fatal(err)
	}
	out := captureAigOutput(t, Aig_ProcedureTest)
	if !strings.Contains(out, "aig_test_file.blif") {
		t.Fatalf("BLIF failure was silent: %q", out)
	}
}
