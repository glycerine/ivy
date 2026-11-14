package gia

import (
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

func TestGiarfClassesRefineCollisionsAndSimulationHistory(t *testing.T) {
	const count = 40
	g := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	ids := make([]int32, count)
	for i := range ids {
		ids[i] = Gia_ManAppendAnd(g, a, b) >> 1
	}
	p := Gia_ManEquivStart(g, 10, 0)
	Hcd_ManClassesCreate(p)
	// Force distinct nonconstant signatures into the same hash bucket. Hash
	// equality must never substitute for full simulation equality.
	size := util.Abc_PrimeCudd(100 + count/5)
	var words []uint32
	key := Hcd_ManHashKey([]uint32{2}, 1, size)
	for word := uint32(2); len(words) < 4; word += 2 {
		if Hcd_ManHashKey([]uint32{word}, 1, size) == key {
			words = append(words, word)
		}
	}
	history := make([]string, count)
	for round := 0; round < 4; round++ {
		groups := make(map[string][]int32)
		for i, id := range ids {
			var word uint32
			switch round {
			case 0:
				if i >= 8 {
					word = words[i%4]
				}
			case 1:
				word = words[(i/4)%4]
			case 2:
				word = words[(i/16)%4]
			case 3:
				// Making every current signature equal must not merge classes
				// already distinguished by earlier simulation rounds.
			}
			if word != 0 {
				history[i] += fmt.Sprintf("%d:%08x;", round, word)
			}
			groups[history[i]] = append(groups[history[i]], id)
			if i%2 != 0 {
				word = ^word
			}
			Hcd_ObjSetSim(p, id, word)
		}
		Hcd_ManClassesRefine(p)
		for signature, members := range groups {
			for i, id := range members {
				repr, next := int32(GIA_VOID), int32(0)
				if signature == "" {
					repr = 0 // still a candidate constant, not a linked class
				} else {
					if i != 0 {
						repr = members[0]
					}
					if i+1 < len(members) {
						next = members[i+1]
					}
				}
				if Gia_ObjRepr(g, id) != repr || Gia_ObjNext(g, id) != next {
					t.Fatalf("round %d object %d: repr/next %d/%d, want %d/%d", round, id,
						Gia_ObjRepr(g, id), Gia_ObjNext(g, id), repr, next)
				}
			}
		}
	}
}
