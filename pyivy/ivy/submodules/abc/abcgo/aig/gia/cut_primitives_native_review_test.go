package gia

import (
	"fmt"
	"math/bits"
	"os/exec"
	"strings"
	"testing"
)

func TestCutPrimitiveMergesAndDominanceMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cut_primitives_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native cut primitives: %v\n%s", err, native)
	}
	ids := [6]int32{0, 1, 63, 64, 65, 127}
	makeCut := func(mask int) Gia_Cut_t {
		var c Gia_Cut_t
		for i, id := range ids {
			if mask&(1<<uint(i)) != 0 {
				c.pLeaves[c.nLeaves] = id
				c.nLeaves++
			}
		}
		c.Sign = Gia_CutGetSign(&c)
		return c
	}
	var got strings.Builder
	merges := 0
	for ma := 0; ma < 64; ma++ {
		for mb := 0; mb < 64; mb++ {
			a, b := makeCut(ma), makeCut(mb)
			contained := Gia_CutCheck(&a, &b)
			if contained != giaBool(mb & ^ma == 0) {
				t.Fatal("incorrect subset result")
			}
			ordered := int32(-1)
			if a.nLeaves >= b.nLeaves {
				ordered = Gia_CutSetCutIsContainedOrder(&a, &b)
				if ordered != contained {
					t.Fatal("ordered subset disagrees")
				}
			}
			fmt.Fprintf(&got, "check %d %d %d %d %d\n", ma, mb, contained, ordered, Gia_CutCountBits(a.Sign|b.Sign))
			for limit := int32(0); limit <= 8; limit++ {
				for method := 0; method < 2; method++ {
					out := Gia_Cut_t{nLeaves: GIA_CUT_NO_LEAF, iFunc: 123}
					var status int32
					if method == 0 {
						status = Gia_CutMergeOrder(&a, &b, &out, limit)
					} else {
						status = Gia_CutMergeOrder2(&a, &b, &out, limit)
					}
					if status != giaBool(bits.OnesCount8(uint8(ma|mb)) <= int(limit)) {
						t.Fatal("incorrect merge bound")
					}
					fmt.Fprintf(&got, "merge %d %d %d %d %d", ma, mb, limit, method, status)
					if status != 0 {
						expected := makeCut(ma | mb)
						if out.nLeaves != expected.nLeaves || out.Sign != expected.Sign || out.iFunc != -1 {
							t.Fatal("merge metadata differs")
						}
						fmt.Fprintf(&got, " %d %016x %d", out.nLeaves, out.Sign, out.iFunc)
						for i, id := range out.pLeaves[:out.nLeaves] {
							if id != expected.pLeaves[i] {
								t.Fatal("merge leaves differ")
							}
							fmt.Fprintf(&got, " %d", id)
						}
					}
					got.WriteByte('\n')
					merges++
				}
			}
		}
	}
	for _, limit := range []int32{2, 5, 50} {
		var storage [GIA_MAX_CUTNUM]Gia_Cut_t
		var cuts [GIA_MAX_CUTNUM]*Gia_Cut_t
		slots := make(map[*Gia_Cut_t]int)
		for i := range cuts {
			cuts[i] = &storage[i]
			slots[cuts[i]] = i
		}
		n := int32(0)
		random := uint32(limit + 409)
		for step := 0; step < 64; step++ {
			random = random*1664525 + 1013904223
			mask := 1 + int(random%63)
			if step == 63 {
				mask = 0
			}
			*cuts[n] = makeCut(mask)
			if mask != 0 {
				cuts[n].nTreeLeaves = (random >> 19) & 3
			}
			if step%7 == 0 {
				cuts[n].nTreeLeaves += 1 << 20
			}
			accepted := Gia_CutSetLastCutIsContained(cuts[:], n) == 0
			if accepted {
				n = Gia_CutSetAddCut(cuts[:], n, limit)
			}
			if n < 1 || n >= limit {
				t.Fatal("priority set exceeded bound")
			}
			fmt.Fprintf(&got, "set %d %d %d %d", limit, step, giaBool(accepted), n)
			seen := map[*Gia_Cut_t]bool{}
			for _, c := range cuts[:n] {
				if seen[c] {
					t.Fatal("candidate storage aliases an active cut")
				}
				seen[c] = true
				fmt.Fprintf(&got, " %d:%d:%d:%016x", slots[c], c.nLeaves, c.nTreeLeaves, c.Sign)
			}
			got.WriteByte('\n')
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(got.String(), "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("cut primitive line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatal("primitive snapshot lengths differ")
	}
	t.Logf("%d exact native merges, 4096 subset checks and 192 dominance/tie/priority updates with colliding signatures", merges)
}
