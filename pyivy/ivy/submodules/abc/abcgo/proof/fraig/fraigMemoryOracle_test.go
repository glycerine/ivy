package fraig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The standalone pool has the same chunk policy, recycle order, and restart
// behavior as fraigMem.c, even though Go does not use intrusive free links.
func TestFixedMemoryAgainstOriginalC(t *testing.T) {
	if os.Getenv("FRAIG_C_DIFFERENTIAL") != "1" {
		t.Skip("set FRAIG_C_DIFFERENTIAL=1 for the original C oracle")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "memory-oracle")
	cmd := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", filepath.Join(root, "src"), "-ffunction-sections", "-fdata-sections", filepath.Join("test_vectors", "memory_oracle.c"), filepath.Join(root, "src", "proof", "fraig", "fraigMem.c"), "-Wl,--gc-sections", "-o", binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile C: %v\n%s", err, out)
	}
	original, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run C: %v\n%s", err, original)
	}
	var got strings.Builder
	for _, tc := range []struct{ size, count int32 }{{8, 1024}, {64, 1024}, {1000, 65}, {9000, 8}} {
		p := Fraig_MemFixedStart(tc.size)
		entries := make([][]byte, tc.count+1)
		for i := range entries {
			entries[i] = Fraig_MemFixedEntryFetch(p)
		}
		before := Fraig_MemFixedReadMemUsage(p)
		Fraig_MemFixedEntryRecycle(p, entries[1])
		Fraig_MemFixedEntryRecycle(p, entries[2])
		first := Fraig_MemFixedEntryFetch(p)
		second := Fraig_MemFixedEntryFetch(p)
		f, s := fraigBool(&first[0] == &entries[2][0]), fraigBool(&second[0] == &entries[1][0])
		Fraig_MemFixedRestart(p)
		after := Fraig_MemFixedReadMemUsage(p)
		restarted := Fraig_MemFixedEntryFetch(p)
		fmt.Fprintf(&got, "%d %d %d %d %d %d\n", tc.size, before, after, f, s, fraigBool(&restarted[0] == &entries[0][0]))
		Fraig_MemFixedStop(p, 0)
	}
	if got.String() != string(original) {
		t.Fatalf("Go memory:\n%sC memory:\n%s", got.String(), original)
	}
}
