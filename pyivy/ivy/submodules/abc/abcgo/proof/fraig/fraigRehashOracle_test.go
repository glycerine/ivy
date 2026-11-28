package fraig

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"path/filepath"
	"strings"
	"testing"
)

// FRAIG_C_DIFFERENTIAL=1 compiles only the original rehash function and a
// small fixture; normal package tests do not require a C compiler.
func TestSparseRehashAgainstOriginalC(t *testing.T) {
	if testoracle.Gate("FRAIG_C_DIFFERENTIAL") != "1" {
		t.Skip("set FRAIG_C_DIFFERENTIAL=1 for the original C oracle")
	}
	cc, err := testoracle.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "rehash-oracle")
	cmd := testoracle.Command(t, cc, "-DABC_USE_STDINT_H", "-I", filepath.Join(root, "src"), "-ffunction-sections", "-fdata-sections", filepath.Join("test_vectors", "rehash_oracle.c"), filepath.Join(root, "src", "proof", "fraig", "fraigTable.c"), "-Wl,--gc-sections", "-o", binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile original C: %v\n%s", err, out)
	}
	original, err := testoracle.Command(t, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("original C: %v\n%s", err, original)
	}
	table := &Fraig_HashTable_t{nBins: 3, nEntries: 3, pBins: make([]*Fraig_Node_t, 3)}
	a, b, c := &Fraig_Node_t{uHashD: 6}, &Fraig_Node_t{uHashD: 6}, &Fraig_Node_t{uHashD: 0}
	a.pNextF = b
	b.pNextF = c
	table.pBins[0] = a
	m := &Fraig_Man_t{pTableF0: table}
	var got strings.Builder
	for i := 0; i < 3; i++ {
		b.uHashD = 6
		linked := Fraig_TableRehashF0(m, 1)
		grouped := fraigBool(a.pNextD == b || b.pNextD == a)
		b.uHashD = 5
		Fraig_TableRehashF0(m, 0)
		separate := fraigBool(a.pNextD == nil && b.pNextD == nil && c.pNextD == nil)
		fmt.Fprintf(&got, "%d %d %d %d\n", linked, grouped, separate, table.nEntries)
	}
	if got.String() != string(original) {
		t.Fatalf("Go rehash:\n%soriginal C:\n%s", got.String(), original)
	}
}
