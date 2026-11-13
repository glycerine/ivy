package ssw

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestPairClassesCompiledCOracle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("fixture uses GNU linker section garbage collection")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	binary := filepath.Join(t.TempDir(), "pairs")
	fixture := filepath.Join(filepath.Dir(file), "test_vectors/ssw_pairs_compare.c")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src, fixture, "-Wl,--gc-sections", "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile C fixture: %v: %s", err, out)
	}
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"1", "1"}, "1:1,1;"},
		{[]string{"3", "5", "1", "5"}, "1:1,3,5;"},
		{[]string{"1", "2", "3", "4", "2", "4"}, "1:1,2,3,4;"},
		{[]string{"3", "5", "1", "2", "2", "5"}, "1:1,2,3,5;"},
	}
	for _, tc := range cases {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		out, err := exec.CommandContext(ctx, binary, tc.args...).CombinedOutput()
		cancel()
		if err != nil || strings.TrimSpace(string(out)) != tc.want {
			t.Fatalf("C %v: %v, %q, want %q", tc.args, err, out, tc.want)
		}
	}
	args := strings.Fields("6 8 3 8 5 8 2 6 1 5 1 5 3 8 2 7 4 9 1 5 1 8 2 6")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, binary, args...).CombinedOutput()
	if err == nil || ctx.Err() != nil || !strings.Contains(string(out), "Vec_IntEntry") {
		t.Fatalf("C invalid offset: %v %q", err, out)
	}
}
