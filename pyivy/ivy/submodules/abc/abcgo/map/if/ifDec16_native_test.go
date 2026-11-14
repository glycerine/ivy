package if_

import (
	"encoding/json"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestClusteringNativeFixtures(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), "../../../test_vectors/map_if/clustering.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Vars, Leaf, Root int32
		Truth            []string
	}
	if err = json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		table := make([]uint64, len(fixture.Truth))
		for i, s := range fixture.Truth {
			table[i], err = strconv.ParseUint(s, 16, 64)
			if err != nil {
				t.Fatal(err)
			}
		}
		var r If_Grp_t
		var f0, f1 uint64
		g := If_CluCheck(nil, table, fixture.Vars, 0, 0, fixture.Leaf, fixture.Root, &r, &f0, &f1, nil, 0)
		if g.nVars != 0 {
			If_CluVerify(table, fixture.Vars, &g, &r, f1, []uint64{f0})
			If_CluCheckGroup(table, fixture.Vars, &g)
		}
	}
}
func TestClusteringNativeReconstruction(t *testing.T) {
	rng := rand.New(rand.NewSource(912))
	successes := 0
	for trial := 0; trial < 100; trial++ {
		n := int32(5 + trial%4)
		leaf := int32(3 + trial%3)
		root := n - leaf + 1
		bound := rng.Uint64()
		free := rng.Uint64()
		if root > 6 {
			continue
		}
		truth := make([]uint64, If_CluWordNum(n))
		for assignment := 0; assignment < 1<<uint(n); assignment++ {
			bv := bound >> uint(assignment&((1<<uint(leaf))-1)) & 1
			rv := assignment>>uint(leaf) | int(bv)<<uint(n-leaf)
			if free>>uint(rv)&1 != 0 {
				ifTruthSet(truth, assignment)
			}
		}
		if n < 6 {
			truth[0] = If_CluAdjust(truth[0], n)
		}
		var r If_Grp_t
		var f0, f1 uint64
		g := If_CluCheck(nil, truth, n, 0, 0, leaf, root, &r, &f0, &f1, nil, 0)
		if g.nVars != 0 {
			successes++
			If_CluVerify(truth, n, &g, &r, f1, []uint64{f0})
		}
	}
	if successes < 20 {
		t.Fatal("too few reconstructed decompositions", successes)
	}
}
func TestClusteringNativePermutationsAndCofactors(t *testing.T) {
	rng := rand.New(rand.NewSource(19))
	for n := int32(3); n <= 10; n++ {
		truth := make([]uint64, If_CluWordNum(n))
		for i := range truth {
			truth[i] = rng.Uint64()
		}
		if n < 6 {
			truth[0] = If_CluAdjust(truth[0], n)
		}
		original := append([]uint64(nil), truth...)
		a, b := cluIdentity(n)
		for i := 0; i < 20; i++ {
			If_CluSwapVars(truth, n, a, b, int32(rng.Intn(int(n))), int32(rng.Intn(int(n))))
		}
		If_CluCheckPerm(original, truth, n, a, b)
		permutation := make([]int32, n)
		phase := If_CluSemiCanonicize(truth, n, permutation)
		If_CluSemiCanonicizeVerify(truth, original, n, permutation, phase)
	}
}
func TestClusteringNativeThreeLUT(t *testing.T) {
	truth := []uint64{0x6dd9926d962d6996}
	var r, g2 If_Grp_t
	var f0, f1, f2 uint64
	g := If_CluCheck3(nil, truth, 6, 4, 4, 4, &r, &g2, &f0, &f1, &f2)
	if g.nVars == 0 {
		t.Fatal("source example did not decompose")
	}
	If_CluVerify3(truth, 6, &g, &g2, &r, f1, f2, f0)
}
func TestClusteringNativeThreeStageReconstruction(t *testing.T) {
	rng := rand.New(rand.NewSource(614))
	success := 0
	for trial := 0; trial < 30; trial++ {
		n := int32(9)
		a, b, c := rng.Uint64(), rng.Uint64(), rng.Uint64()
		truth := make([]uint64, If_CluWordNum(n))
		for assignment := 0; assignment < 512; assignment++ {
			first := a >> uint(assignment&7) & 1
			secondIndex := int(first) | (assignment>>3&7)<<1
			second := b >> uint(secondIndex) & 1
			rootIndex := int(second) | (assignment>>6&7)<<1
			if c>>uint(rootIndex)&1 != 0 {
				ifTruthSet(truth, assignment)
			}
		}
		var root, group2 If_Grp_t
		var f0, f1, f2 uint64
		group := If_CluCheck3(nil, truth, n, 4, 4, 4, &root, &group2, &f0, &f1, &f2)
		if group.nVars != 0 {
			success++
			If_CluVerify3(truth, n, &group, &group2, &root, f1, f2, f0)
		}
	}
	if success < 5 {
		t.Fatal("three-stage search never found composed functions", success)
	}
}
func TestClusteringNativeCacheAndStructDelay(t *testing.T) {
	p := &If_Man_t{pPars: &If_Par_t{NLutSize: 6, NCutsMax: 1, FCutMin: 1}}
	first := []uint64{0x6dd9926d962d6996}
	slot := If_CluHashLookup(p, first, 0)
	*slot = 36
	for i := 0; i < 10; i++ {
		if got := If_CluHashLookup(p, first, 0); *got != 36 {
			t.Fatal("cache lost result")
		}
	}
	for i := 0; i < 25; i++ {
		If_CluHashLookup(p, []uint64{uint64(i)}, 0)
	}
	if *If_CluHashLookup(p, first, 0) != 36 {
		t.Fatal("frequent cache entry was pruned")
	}
	if p.nTableEntries[0] > 2*p.nTableSize[0]+1 {
		t.Fatal("cache is not bounded")
	}
	If_CluHashTableCheck(p)
	p.vObjs = vec.Vec_PtrAlloc(5)
	for i := 0; i < 5; i++ {
		o := &If_Obj_t{Id: int32(i), Type: IF_CI}
		o.CutBest.Delay = float32(i)
		vec.Vec_PtrPush(p.vObjs, o)
	}
	cut := &If_Cut_t{nLimit: 6, nLeaves: 4, pLeaves: []int32{1, 2, 3, 4}}
	if delay := If_CutDelayLutStruct(p, cut, "44", 0); delay != 5 {
		t.Fatal("single LUT delay", delay)
	}
	if len(cut.permutation) != 4 {
		t.Fatal("delay profile not stored")
	}
	if If_CutPerformCheck16(p, []uint32{0x88888888, 0x88888888}, 6, 2, "33") != 1 {
		t.Fatal("single small LUT rejected")
	}
	if If_CutPerformCheck16(p, []uint32{0, 0}, 6, 6, "92") != 0 {
		t.Fatal("invalid LUT structure accepted")
	}
}
