package if_

import (
	"fmt"
	"math/bits"
	"slices"
	"testing"
)

func cluSharedReviewValue(x, n, free, shared int) bool {
	boundParity, boundAnd, freeParity, freeAnd := false, true, false, true
	for pin := 0; pin < n; pin++ {
		bit := x>>pin&1 != 0
		if pin < free {
			freeParity = freeParity != bit
			freeAnd = freeAnd && bit
		} else if pin != shared {
			boundParity = boundParity != bit
			boundAnd = boundAnd && bit
		}
	}
	// One free input needs four distinct one-variable functions rather
	// than parity/conjunction, which coincide at that support size.
	if free == 1 {
		if x>>shared&1 != 0 {
			return !freeAnd || boundAnd
		}
		return freeAnd && boundParity
	}
	if x>>shared&1 != 0 {
		return freeAnd != boundAnd
	}
	return freeParity != boundParity
}

func TestCluSharedSelectionAndDerivation(t *testing.T) {
	for bound := 3; bound <= 6; bound++ {
		for _, free := range []int{1, 2, 4, 5, 6, 9, 10, 13} {
			n := bound + free
			if n > 16 {
				continue
			}
			for _, shared := range []int{free, free + bound/2, n - 1} {
				for classes := 3; classes <= 4; classes++ {
					for phase := 0; phase < 4; phase++ {
						t.Run(fmt.Sprintf("bound%d/free%d/shared%d/classes%d/phase%d", bound, free, shared, classes, phase), func(t *testing.T) {
							signedMask := 0
							if phase&2 != 0 {
								signedMask = ((1 << n) - 1) & 0x5555
							}
							value := func(x int) bool {
								logical := x ^ signedMask
								result := cluSharedReviewValue(logical, n, free, shared)
								if classes == 3 {
									result = logical>>shared&1 == 0 && bits.OnesCount(uint(logical&^(1<<shared)))%2 != 0
								}
								return result != (phase&1 != 0)
							}
							truth := make([]uint64, If_CluWordNum(int32(n)))
							for x := 0; x < 64*len(truth); x++ {
								if value(x & ((1 << n) - 1)) {
									truth[x>>6] |= 1 << uint(x&63)
								}
							}
							v2p, p2v := cluIdentity(int32(n))
							group := If_Grp_t{nVars: int32(bound), nMyu: int32(classes)}
							for pin := 0; pin < bound; pin++ {
								group.pVars[pin] = int32(free + pin)
							}
							if got := If_CluCheckNonDisjointGroup(truth, int32(n), v2p, p2v, &group); got != 1 {
								t.Fatal("known shared decomposition rejected")
							}
							var expected []int32
							for pin := free; pin < n; pin++ {
								if pin != shared {
									expected = append(expected, int32(pin))
								}
							}
							expected = append(expected, int32(shared))
							if group.nVars != int32(bound) || group.nMyu != int32(classes) || !slices.Equal(group.pVars[:bound], expected) {
								t.Fatalf("selected group%+v want pins%v", group, expected)
							}
							for pos, pin := range p2v[:n] {
								if v2p[pin] != int32(pos) {
									t.Fatalf("permutation inverse at%d", pos)
								}
							}
							for x := 0; x < 1<<n; x++ {
								original := 0
								for pos, pin := range p2v[:n] {
									original |= (x >> pos & 1) << pin
								}
								if got := truth[x>>6]>>uint(x&63)&1 != 0; got != value(original) {
									t.Fatalf("moved truth at%d", x)
								}
							}
							before := slices.Clone(truth)
							beforeGroup := group
							var root If_Grp_t
							function := If_CluDeriveNonDisjoint(truth, int32(n), v2p, p2v, &group, &root)
							if group != beforeGroup {
								t.Fatal("derivation changed bound group")
							}
							if root.nVars != int32(free+2) || root.pVars[free] != int32(n) || root.pVars[free+1] != int32(shared) {
								t.Fatalf("remaining group%+v", root)
							}
							for pin := 0; pin < free; pin++ {
								if root.pVars[pin] != int32(pin) {
									t.Fatalf("free pin%d", pin)
								}
							}
							if words := If_CluWordNum(root.nVars); !slices.Equal(truth[words:], before[words:]) {
								t.Fatal("derivation overwrote unused truth suffix")
							}
							for x := 0; x < 1<<n; x++ {
								index := 0
								for pin, v := range group.pVars[:bound] {
									index |= (x >> v & 1) << pin
								}
								boundBit := function >> index & 1
								rootIndex := (x & ((1 << free) - 1)) | int(boundBit)<<free | (x>>shared&1)<<(free+1)
								if got := truth[rootIndex>>6]>>uint(rootIndex&63)&1 != 0; got != value(x) {
									t.Fatalf("assignment%d got%t want%t", x, got, value(x))
								}
							}
						})
					}
				}
			}
		}
	}
}

func TestCluSharedRejectionPreservesState(t *testing.T) {
	for _, classes := range []int32{0, 1, 2, 3, 4, 5} {
		truth := []uint64{0x9896698e899ee966}
		before := slices.Clone(truth)
		v2p, p2v := cluIdentity(6)
		group := If_Grp_t{nVars: 4, nMyu: classes}
		copy(group.pVars[:], []int32{2, 3, 4, 5})
		beforeGroup := group
		if got := If_CluCheckNonDisjointGroup(truth, 6, v2p, p2v, &group); got != 0 {
			t.Fatalf("class%d accepted", classes)
		}
		expected, _ := cluIdentity(6)
		if group != beforeGroup || !slices.Equal(truth, before) || !slices.Equal(v2p, expected) || !slices.Equal(p2v, expected) {
			t.Fatalf("class%d rejection changed state", classes)
		}
	}
}

func BenchmarkCluSharedSelection(b *testing.B) {
	const n, free, bound = 16, 10, 6
	truth := make([]uint64, 1024)
	for x := 0; x < 1<<n; x++ {
		if cluSharedReviewValue(x, n, free, n-1) {
			truth[x>>6] |= 1 << uint(x&63)
		}
	}
	v2p, p2v := cluIdentity(n)
	group := If_Grp_t{nVars: bound, nMyu: 4}
	for i := 0; i < bound; i++ {
		group.pVars[i] = free + int32(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := If_CluCheckNonDisjointGroup(truth, n, v2p, p2v, &group); got != 1 {
			b.Fatal(got)
		}
	}
}

func BenchmarkCluSharedDerivation(b *testing.B) {
	const n, free, bound = 16, 10, 6
	truth, work := make([]uint64, 1024), make([]uint64, 1024)
	for x := 0; x < 1<<n; x++ {
		if cluSharedReviewValue(x, n, free, n-1) {
			truth[x>>6] |= 1 << uint(x&63)
		}
	}
	v2p, p2v := cluIdentity(n)
	group := If_Grp_t{nVars: bound, nMyu: 4}
	for i := 0; i < bound; i++ {
		group.pVars[i] = free + int32(i)
	}
	var root If_Grp_t
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(work, truth)
		If_CluDeriveNonDisjoint(work, n, v2p, p2v, &group, &root)
	}
}
