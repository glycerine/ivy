package webui

import (
	"os/exec"
	"testing"
)

func TestFrontendFileStateSaveRestoresHandle(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not found")
	}
	cmd := exec.Command(node, "static/js/ivyweb_file_state_test.mjs")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("frontend file-state test failed: %v\n%s", err, out)
	}
}
