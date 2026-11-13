package fraig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/sat/msat"
)

// This runs the real C compression function with overlapping source patterns
// and a partially full permanent word. It does not require the ABC executable.
func TestCompressionAgainstOriginalC(t *testing.T) {
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
	binary := filepath.Join(t.TempDir(), "compression-oracle")
	args := []string{"-DABC_USE_STDINT_H", "-I", filepath.Join(root, "src"), "-ffunction-sections", "-fdata-sections", filepath.Join("test_vectors", "compress_oracle.c")}
	for _, f := range []string{"fraigFeed.c", "fraigTable.c", "fraigMem.c", "fraigVec.c", "fraigNode.c", "fraigApi.c", "fraigPrime.c", "fraigUtil.c"} {
		args = append(args, filepath.Join(root, "src", "proof", "fraig", f))
	}
	args = append(args, filepath.Join(root, "src", "sat", "msat", "msatVec.c"), "-Wl,--gc-sections", "-o", binary)
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile original C: %v\n%s", err, out)
	}
	original, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("original C: %v\n%s", err, original)
	}
	a := &Fraig_Node_t{Num: 1, NumPi: 0, puSimD: []uint32{0x80000000, 1, 0, 0, 0, 0, 0, 0}}
	b := &Fraig_Node_t{Num: 2, NumPi: 1, puSimD: []uint32{0, 1, 0, 0, 0, 0, 0, 0}}
	c := &Fraig_Node_t{NumPi: -1}
	a.pNextD = b
	p := &Fraig_Man_t{vInputs: &Fraig_NodeVec_t{pArray: []*Fraig_Node_t{a, b}}, vNodes: &Fraig_NodeVec_t{pArray: []*Fraig_Node_t{c, a, b}}, vCones: Fraig_NodeVecAlloc(0), pTableF: &Fraig_HashTable_t{pBins: []*Fraig_Node_t{a}}, nWordsDyna: 8, iWordStart: 3, iWordPerm: 1, iPatsPerm: 31, pSimsReal: make([]uint32, 8), pSimsTemp: make([]uint32, 8), vPatsReal: msat.Msat_IntVecAlloc(4)}
	p.pSimsReal[0] = 0x7fffffff
	msat.Msat_IntVecPush(p.vPatsReal, 32)
	msat.Msat_IntVecPush(p.vPatsReal, 31)
	Fraig_FeedBackCompress(p)
	got := fmt.Sprintf("%d %d %d %08x %08x %08x %08x %08x %08x\n", p.iWordStart, p.iWordPerm, p.iPatsPerm, a.puSimD[0], a.puSimD[1], b.puSimD[0], b.puSimD[1], p.pSimsReal[0], p.pSimsReal[1])
	if got != string(original) {
		t.Fatalf("Go compression: %soriginal C: %s", got, original)
	}
}
