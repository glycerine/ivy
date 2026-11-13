package fraig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/sat/msat"
)

// This fixture compares internal-node resimulation and signatures after the
// C and Go compression paths both resize their simulation storage.
func TestInternalCompressionAgainstOriginalC(t *testing.T) {
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
	binary := filepath.Join(t.TempDir(), "internal-compression-oracle")
	args := []string{"-DABC_USE_STDINT_H", "-I", filepath.Join(root, "src"), "-ffunction-sections", "-fdata-sections", filepath.Join("test_vectors", "compress_internal_oracle.c")}
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
	a := &Fraig_Node_t{Num: 1, NumPi: 0, puSimR: []uint32{0x12345678}, puSimD: []uint32{0x80000000, 1, 0, 0, 0}}
	b := &Fraig_Node_t{Num: 2, NumPi: 1, puSimR: []uint32{0xa5a5a5a5}, puSimD: []uint32{0, 1, 0, 0, 0}}
	constant := &Fraig_Node_t{NumPi: -1}
	gate := &Fraig_Node_t{Num: 3, NumPi: -1, p1: a, p2: Fraig_Not(b), puSimR: make([]uint32, 1), puSimD: make([]uint32, 5)}
	Fraig_NodeSimulate(gate, 0, 1, 1)
	p := &Fraig_Man_t{pConst1: constant, vInputs: &Fraig_NodeVec_t{pArray: []*Fraig_Node_t{a, b}}, vNodes: &Fraig_NodeVec_t{pArray: []*Fraig_Node_t{constant, a, b, gate}}, vCones: Fraig_NodeVecAlloc(8), mmSims: Fraig_MemFixedStart(24), pTableF: &Fraig_HashTable_t{}, nWordsRand: 1, nWordsDyna: 5, iWordStart: 3, iWordPerm: 1, iPatsPerm: 31, pSimsReal: []uint32{0x7fffffff, 0, 0, 0, 0}, pSimsTemp: make([]uint32, 5), vPatsReal: msat.Msat_IntVecAlloc(4)}
	msat.Msat_IntVecPush(p.vPatsReal, 32)
	Fraig_FeedBackCompress(p)
	got := fmt.Sprintf("%d %d %d %d %08x %08x %08x %08x %08x %08x %08x\n", p.iWordStart, p.iWordPerm, p.iPatsPerm, p.nWordsDyna, a.puSimD[0], b.puSimD[0], gate.puSimR[0], gate.puSimD[0], gate.uHashR, gate.uHashD, p.pSimsReal[0])
	Fraig_MemFixedStop(p.mmSims, 0)
	if got != string(original) {
		t.Fatalf("Go compression: %soriginal C: %s", got, original)
	}
}
