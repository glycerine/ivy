package gia

import (
	"os/exec"
	"testing"
)

func TestMappingNativeHashLastCutAndCollapseBoundaries(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mapping_hash_policy_review.c")
	for _, tc := range []struct{ mode, want string }{{"0", "root 4 cut 1 2 3\n"}, {"1", "root 4 cut 4 4\n"}, {"2", "root 5 cut 1 2 6\n"}} {
		source, err := exec.Command(exe, tc.mode).CombinedOutput()
		if err != nil {
			t.Fatalf("source hash policy%s: %v %s", tc.mode, err, source)
		}
		if string(source) != tc.want {
			t.Fatalf("source last-cut/collapse behavior changed mode%s: %q", tc.mode, source)
		}
	}
}
