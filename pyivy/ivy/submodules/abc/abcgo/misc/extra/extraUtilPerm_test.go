package extra

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/bits"
	"slices"
	"testing"
)

func zddFamily(p *Abc_ZddMan, root int32) map[uint32]bool {
	result := map[uint32]bool{}
	var visit func(int32, uint32)
	visit = func(id int32, mask uint32) {
		if id == 0 {
			return
		}
		if id == 1 {
			result[mask] = true
			return
		}
		node := p.pObjs[id]
		visit(node.False, mask)
		visit(node.True, mask|1<<uint(node.Var))
	}
	visit(root, 0)
	return result
}
func zddBuildFamily(p *Abc_ZddMan, mask uint32, variables int32) int32 {
	var root int32
	for set := int32(0); set < 1<<uint(variables); set++ {
		if mask&(1<<uint(set)) == 0 {
			continue
		}
		var vars []int32
		for v := int32(0); v < variables; v++ {
			if set&(1<<uint(v)) != 0 {
				vars = append(vars, v)
			}
		}
		root = Abc_ZddUnion(p, root, Abc_ZddBuildSet(p, vars, int32(len(vars))))
	}
	return root
}
func zddFamilyMask(p *Abc_ZddMan, root int32) uint32 {
	var mask uint32
	for set := range zddFamily(p, root) {
		mask |= 1 << set
	}
	return mask
}
func TestZddExhaustiveFamilies(t *testing.T) {
	p := Abc_ZddManAlloc(3, 100000)
	roots := make([]int32, 256)
	for i := range roots {
		roots[i] = zddBuildFamily(p, uint32(i), 3)
		if zddFamilyMask(p, roots[i]) != uint32(i) || Abc_ZddCountPaths(p, roots[i]) != uint64(bits.OnesCount(uint(i))) {
			t.Fatal("build/count", i)
		}
	}
	for a := 0; a < 256; a++ {
		for b := 0; b < 256; b++ {
			for _, tc := range []struct {
				got  int32
				want uint32
			}{{Abc_ZddUnion(p, roots[a], roots[b]), uint32(a | b)}, {Abc_ZddDiff(p, roots[a], roots[b]), uint32(a &^ b)}, {Abc_ZddIntersect(p, roots[a], roots[b]), uint32(a & b)}} {
				if zddFamilyMask(p, tc.got) != tc.want {
					t.Fatal("family operation", a, b, tc)
				}
			}
		}
	}
}
func TestZddDotProductAndThreshold(t *testing.T) {
	p := Abc_ZddManAlloc(3, 100000)
	for a := uint32(0); a < 256; a += 7 {
		for b := uint32(0); b < 256; b += 11 {
			A, B := zddBuildFamily(p, a, 3), zddBuildFamily(p, b, 3)
			var expected uint32
			for x := uint32(0); x < 8; x++ {
				if a&(1<<x) == 0 {
					continue
				}
				for y := uint32(0); y < 8; y++ {
					if b&(1<<y) != 0 {
						expected |= 1 << (x | y)
					}
				}
			}
			got := Abc_ZddDotProduct(p, A, B)
			if zddFamilyMask(p, got) != expected {
				t.Fatal(a, b)
			}
		}
	}
	all := zddBuildFamily(p, 255, 3)
	for threshold := int32(0); threshold <= 4; threshold++ {
		got := zddFamily(p, Abc_ZddThresh(p, all, threshold))
		for set := uint32(0); set < 8; set++ {
			if got[set] != (bits.OnesCount32(set) <= int(threshold)) {
				t.Fatal("threshold", threshold, set)
			}
		}
	}
}
func zddPermutation(p *Abc_ZddMan, perm []int32) int32 {
	work := slices.Clone(perm)
	comb := make([]int32, len(perm))
	n := Abc_ZddPerm2Comb(work, int32(len(work)), comb)
	for i := int32(0); i < n; i++ {
		v := uint32(comb[i])
		comb[i] = Abc_ZddVarIJ(p, int32(v>>16), int32(v&65535))
	}
	return Abc_ZddBuildSet(p, comb, n)
}
func zddDecodePermutation(p *Abc_ZddMan, root int32) []int32 {
	var path []int32
	for root >= 2 {
		node := p.pObjs[root]
		if node.False != 0 {
			panic("not a singleton permutation")
		}
		path = append(path, int32(uint32(p.pV2TI[node.Var])<<16|uint32(p.pV2TJ[node.Var])))
		root = node.True
	}
	if root != 1 {
		panic("empty permutation family")
	}
	perm := make([]int32, p.nPermSize)
	Abc_ZddComb2Perm(path, int32(len(path)), perm, p.nPermSize)
	return perm
}
func TestZddPermutationProducts(t *testing.T) {
	p := Abc_ZddManAlloc(3, 10000)
	Abc_ZddManCreatePerms(p, 3)
	perms := [][]int32{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	var all int32
	for _, a := range perms {
		A := zddPermutation(p, a)
		if !slices.Equal(zddDecodePermutation(p, A), a) {
			t.Fatal("conversion", a)
		}
		all = Abc_ZddUnion(p, all, A)
		for _, b := range perms {
			B := zddPermutation(p, b)
			expected := []int32{a[b[0]], a[b[1]], a[b[2]]}
			got := zddDecodePermutation(p, Abc_ZddPermProduct(p, A, B))
			if !slices.Equal(got, expected) {
				t.Fatal(a, b, got, expected)
			}
		}
		for v := int32(0); v < 3; v++ {
			got := Abc_ZddPerm(p, Abc_ZddPerm(p, A, v), v)
			if got != A {
				t.Fatal("transposition involution")
			}
		}
	}
	if Abc_ZddPermProduct(p, all, all) != all || Abc_ZddCountPaths(p, all) != 6 {
		t.Fatal("group closure")
	}
}
func TestZddCountsAndCubeStates(t *testing.T) {
	p := Abc_ZddManAlloc(60, 10000)
	root := int32(1)
	for v := int32(59); v >= 0; v-- {
		root = Abc_ZddUniqueCreate(p, v, root, root)
	}
	if Abc_ZddCountPaths(p, root) != 1<<60 || Abc_ZddCountNodes(p, root) != 60 || Abc_ZddCountNodesArray(p, []int32{root, root}) != 60 {
		t.Fatal("counts")
	}
	for round, want := range []uint64{1, 10, 64} {
		q, all := zddCubeStates(int32(round))
		if got := Abc_ZddCountPaths(q, all); got != want {
			t.Fatal("cube round", round, got, want)
		}
		Abc_ZddManFree(q)
	}
}

// Original-C oracle over all pairs of the 256 families on three variables,
// recording union, difference, dot product and minimal-union family masks.
func TestZddOriginalCFamilyOracle(t *testing.T) {
	p := Abc_ZddManAlloc(3, 16384)
	roots := make([]int32, 256)
	for i := range roots {
		roots[i] = zddBuildFamily(p, uint32(i), 3)
	}
	hash := sha256.New()
	record := make([]byte, 16)
	for _, a := range roots {
		for _, b := range roots {
			results := []int32{Abc_ZddUnion(p, a, b), Abc_ZddDiff(p, a, b), Abc_ZddDotProduct(p, a, b), Abc_ZddMinUnion(p, a, b)}
			for i, result := range results {
				binary.LittleEndian.PutUint32(record[4*i:], zddFamilyMask(p, result))
			}
			hash.Write(record)
		}
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "22239df603f3d00181debf4a2df6167ddce479dbfe94226a3364f1eeb760b434" {
		t.Fatal(got)
	}
}

type zddTestGraph struct{}

func (zddTestGraph) ObjectCount() int                        { return 4 }
func (zddTestGraph) InputObjectIDs() []int32                 { return []int32{1, 2} }
func (zddTestGraph) AndFanins(id int32) (int32, int32, bool) { return 1, 2, id == 3 }
func TestZddGiaDependencyBoundary(t *testing.T) {
	p, roots := zddGiaCuts(zddTestGraph{})
	defer Abc_ZddManFree(p)
	if len(roots) != 1 || Abc_ZddCountPaths(p, roots[0]) != 2 {
		t.Fatal(roots)
	}
	family := zddFamily(p, roots[0])
	if !family[(1<<1)|(1<<2)] || !family[1<<3] {
		t.Fatal(family)
	}
}
func TestZddInvalidInputsAndPathOverflow(t *testing.T) {
	for _, run := range []func(){
		func() { Abc_ZddManAlloc(3, 4) },
		func() { p := Abc_ZddManAlloc(3, 5); Abc_ZddBuildSet(p, []int32{0, 1}, 2) },
		func() { p := Abc_ZddManAlloc(3, 100); Abc_ZddManCreatePerms(p, 4) },
		func() { Abc_ZddPerm2Comb([]int32{1, 1}, 2, make([]int32, 2)) },
		func() { Abc_ZddComb2Perm([]int32{5}, 1, make([]int32, 2), 2) },
		func() {
			p := Abc_ZddManAlloc(64, 1000)
			root := int32(1)
			for v := int32(63); v >= 0; v-- {
				root = Abc_ZddUniqueCreate(p, v, root, root)
			}
			Abc_ZddCountPaths(p, root)
		},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid operation accepted")
				}
			}()
			run()
		}()
	}
	p := Abc_ZddManAlloc(1, 10)
	for i := int32(0); i < 100; i++ {
		Abc_ZddCacheInsert(p, i, 0, 0, 1)
	}
	if len(p.cache) > 10 {
		t.Fatal("unbounded memo cache")
	}
	Abc_ZddManFree(p)
	Abc_ZddManFree(p)
}
