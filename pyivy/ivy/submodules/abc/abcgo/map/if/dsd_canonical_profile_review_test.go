package if_

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"slices"
	"testing"
)

func dsdReviewManager(t *testing.T, expression string, n int, arrivals []int32) (*If_Man_t, *If_Cut_t, []word) {
	t.Helper()
	p := If_ManStart(&If_Par_t{NLutSize: int32(n), NCutsMax: 4})
	t.Cleanup(func() { If_ManStop(p) })
	cut := &If_Cut_t{nLeaves: uint32(n), pLeaves: make([]int32, n), iCutFunc: 2}
	for i := range cut.pLeaves {
		node := If_ManCreateCi(p)
		node.CutBest.Delay = float32(arrivals[i])
		cut.pLeaves[i] = node.Id
	}
	p.pIfDsdMan = If_DsdManAlloc(int32(n), 0)
	t.Cleanup(func() { If_DsdManFree(p.pIfDsdMan, 0) })
	If_DsdManAllocIsops(p.pIfDsdMan, int32(n))
	perm := make([]byte, n)
	truth := dau.Dau_DsdToTruth(expression, int32(n))
	lit := If_DsdManCompute(p.pIfDsdMan, truth, int32(n), perm, "")
	p.vTtDsds[n] = vec.Vec_IntStart(2)
	vec.Vec_IntWriteEntry(p.vTtDsds[n], 1, lit)
	stride := max(n, 6)
	p.vTtPerms[n] = vec.Vec_StrAlloc(2 * stride)
	vec.Vec_StrFill(p.vTtPerms[n], 2*stride, 0)
	for i, pin := range perm {
		vec.Vec_StrWriteEntry(p.vTtPerms[n], stride+i, pin)
	}
	return p, cut, truth
}

func dsdReviewPinProfile(t *testing.T, expression string, n int, arrivals []int32) {
	t.Helper()
	p, cut, truth := dsdReviewManager(t, expression, n, arrivals)
	aig := vec.Vec_IntAlloc(0)
	defer vec.Vec_IntFree(aig)
	delay := If_CutDsdBalanceEval(p, cut, aig)
	if delay < 0 {
		t.Fatal("DSD balancing rejected supported expression")
	}
	// The source-shaped verifier clears the vector length after checking it.
	data := vec.Vec_IntArray(aig)
	if got := If_AigVerifyArray(aig, int32(n)); got != truth[0] {
		t.Fatalf("balanced truth%x want%x", got, truth[0])
	}
	paths := make([][]int, n)
	for i := range paths {
		paths[i] = make([]int, n)
		for j := range paths[i] {
			paths[i][j] = -1
		}
		paths[i][i] = 0
	}
	for i := 0; i < len(data)-1; i += 2 {
		a, b := paths[data[i]>>1], paths[data[i+1]>>1]
		row := make([]int, n)
		for pin := range row {
			row[pin] = -1
			if a[pin] >= 0 {
				row[pin] = a[pin] + 1
			}
			if b[pin] >= 0 {
				row[pin] = max(row[pin], b[pin]+1)
			}
		}
		paths = append(paths, row)
	}
	root := paths[len(paths)-1]
	expected := make([]byte, n)
	var expectedDelay int32
	for pin, depth := range root {
		expected[pin] = byte(depth)
		expectedDelay = max(expectedDelay, arrivals[pin]+int32(depth))
	}
	profile := make([]byte, n)
	profileDelay := If_CutDsdBalancePinDelays(p, cut, profile)
	if !slices.Equal(profile, expected) || int32(delay) != expectedDelay || profileDelay != expectedDelay {
		t.Fatalf("%s canonical permutation%v pin profile%v constructed depths%v delays(eval/profile/paths)%v/%d/%d", expression, If_CutDsdPerm(p, cut), profile, expected, delay, profileDelay, expectedDelay)
	}
}
func TestDsdCanonicalPinProfileMatchesConstructedPaths(t *testing.T) {
	dsdReviewPinProfile(t, "(abc)", 3, []int32{0, 0, 0})
}
func TestDsdNestedPrimePinProfileMatchesConstructedPaths(t *testing.T) {
	dsdReviewPinProfile(t, "E8{a(bc)d}", 4, []int32{2, 1, 4, 0})
}

func TestDsdGraphConstructionKeepsSourceCost(t *testing.T) {
	p, cut, _ := dsdReviewManager(t, "(abc)", 3, []int32{0, 0, 0})
	aig := vec.Vec_IntAlloc(0)
	defer vec.Vec_IntFree(aig)
	If_CutDsdBalanceEval(p, cut, aig)
	if cut.Cost != 0 {
		t.Fatalf("construction cost%d; original source reports zero (timing-only mode reports area)", cut.Cost)
	}
}

func TestDsdGraphConstructionKeepsSourceEmissionOrder(t *testing.T) {
	p, cut, _ := dsdReviewManager(t, "(ab[cd]<efg>)", 7, make([]int32, 7))
	aig := vec.Vec_IntAlloc(0)
	defer vec.Vec_IntFree(aig)
	If_CutDsdBalanceEval(p, cut, aig)
	// Fresh unmodified ifDsd.c, case2/mode0 in if_dsd_balance_review.c.
	expected := []int32{0, 2, 4, 7, 5, 6, 17, 19, 21, 14, 8, 10, 9, 12, 25, 27, 29, 22, 0}
	if got := vec.Vec_IntArray(aig); !slices.Equal(got, expected) {
		t.Fatalf("AIG emission order%v; original C%v", got, expected)
	}
}
