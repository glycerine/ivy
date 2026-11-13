package fraig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/sat/msat"
)

// Compare the original C covering order, overlapping pattern compaction,
// repeated compression, simulation masks and internal-node hashes.
func TestLargeRepeatedCompressionAgainstC(t *testing.T) {
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
	binary := filepath.Join(t.TempDir(), "compression-large")
	args := []string{"-DABC_USE_STDINT_H", "-I", filepath.Join(root, "src"), "-ffunction-sections", "-fdata-sections", filepath.Join("test_vectors", "compress_large_oracle.c")}
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
	const count = 6
	const words = 16
	nodes := make([]*Fraig_Node_t, count+3)
	for i := range nodes {
		nodes[i] = &Fraig_Node_t{Num: int32(i), NumPi: -1, puSimR: make([]uint32, 1), puSimD: make([]uint32, words)}
	}
	for i := 0; i < count; i++ {
		n := nodes[i+1]
		n.NumPi = int32(i)
		for w := 0; w < 4; w++ {
			n.puSimD[w] = uint32(uint64(0x9e3779b9)*uint64(i+1)) ^ uint32(uint64(0x6a09e667)*uint64(w+1))
		}
		n.puSimD[0] &= ^uint32(0x7fffffff)
		if i+1 < count {
			n.pNextD = nodes[i+2]
		}
	}
	nodes[count+1].p1 = nodes[1]
	nodes[count+1].p2 = Fraig_Not(nodes[2])
	nodes[count+2].p1 = nodes[count+1]
	nodes[count+2].p2 = nodes[3]
	p := &Fraig_Man_t{pConst1: nodes[0], vInputs: &Fraig_NodeVec_t{pArray: nodes[1 : count+1]}, vNodes: &Fraig_NodeVec_t{pArray: nodes}, vCones: Fraig_NodeVecAlloc(8), mmSims: Fraig_MemFixedStart(4 * (words + 1)), pTableF: &Fraig_HashTable_t{nBins: 1, pBins: []*Fraig_Node_t{nodes[1]}}, nWordsRand: 1, nWordsDyna: words, iWordStart: 4, iWordPerm: 1, iPatsPerm: 31, pSimsReal: make([]uint32, words), pSimsTemp: make([]uint32, words), vPatsReal: msat.Msat_IntVecAlloc(16)}
	p.pSimsReal[0] = 0x7fffffff
	var got strings.Builder
	for round := 0; round < 2; round++ {
		if round != 0 {
			start := p.iWordStart
			for i := 0; i < count; i++ {
				for w := 0; w < 2; w++ {
					nodes[i+1].puSimD[int(start)+w] = uint32(uint64(0x85ebca6b)*uint64(i+1)) ^ uint32(uint64(0xc2b2ae35)*uint64(w+1))
				}
			}
			p.iWordStart += 2
			msat.Msat_IntVecPush(p.vPatsReal, int(32*start+3))
			msat.Msat_IntVecPush(p.vPatsReal, int(32*start+35))
		} else {
			msat.Msat_IntVecPush(p.vPatsReal, 35)
			msat.Msat_IntVecPush(p.vPatsReal, 64)
		}
		Fraig_FeedBackCompress(p)
		fmt.Fprintf(&got, "round %d %d %d %d %d\n", round, p.iPatsPerm, p.iWordPerm, p.iWordStart, p.nWordsDyna)
		for i := 1; i < count+3; i++ {
			fmt.Fprintf(&got, "node %d %08x", i, nodes[i].uHashD)
			for w := int32(0); w < p.iWordPerm; w++ {
				fmt.Fprintf(&got, " %08x", nodes[i].puSimD[w])
			}
			got.WriteByte('\n')
		}
		got.WriteString("mask")
		for w := int32(0); w < p.iWordPerm; w++ {
			fmt.Fprintf(&got, " %08x", p.pSimsReal[w])
		}
		got.WriteByte('\n')
	}
	Fraig_MemFixedStop(p.mmSims, 0)
	if got.String() != string(original) {
		t.Fatalf("Go compression:\n%soriginal C:\n%s", got.String(), original)
	}
}
