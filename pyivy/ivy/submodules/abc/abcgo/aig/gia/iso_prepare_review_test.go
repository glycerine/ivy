package gia

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestIsoPrepareIncludesDeepDanglingLevels(t *testing.T) {
	for _, depth := range []int{1, 2, 5, 32} {
		for _, output := range []string{"constant", "input", "firstAnd"} {
			for _, sequential := range []bool{false, true} {
				t.Run(fmt.Sprintf("depth%d/%s/sequential%t", depth, output, sequential), func(t *testing.T) {
					graph := Gia_ManStart(int32(depth + 6))
					a, b := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
					wantLevels := []int32{0, 0, 0}
					root, first := a, int32(0)
					for level := 1; level <= depth; level++ {
						root = Gia_ManAppendAnd(graph, root, b)
						if level == 1 {
							first = root
						}
						wantLevels = append(wantLevels, int32(level))
					}
					literal, coLevel := int32(0), int32(1)
					if output == "input" {
						literal = a
					} else if output == "firstAnd" {
						literal, coLevel = first, 2
					}
					Gia_ManAppendCo(graph, literal)
					wantLevels = append(wantLevels, coLevel)
					if sequential {
						Gia_ManAppendCo(graph, b)
						Gia_ManSetRegNum(graph, 1)
						wantLevels = append(wantLevels, 1)
					}
					for id := int32(0); id < Gia_ManObjNum(graph); id++ {
						Gia_ManObj(graph, id).Value = 0x100 + uint32(id)
					}
					p := Gia_IsoManStart(graph)
					defer Gia_IsoManStop(p)
					for i := range p.pStoreW {
						p.pStoreW[i] = word(0xa5000000 + uint32(i))
					}
					Gia_IsoPrepare(p)
					if !slices.Equal(p.pLevels, wantLevels) {
						t.Fatalf("levels%v want%v", p.pLevels, wantLevels)
					}
					// Independent stable buckets from the construction's known
					// chain levels, including shallow COs and the isolated const.
					wantClasses, wantItems := []int32{0, 1}, []uint32{0}
					for level := int32(0); level <= max(int32(depth), coLevel); level++ {
						begin := len(wantItems)
						for id := 1; id < len(wantLevels); id++ {
							if wantLevels[id] == level {
								wantItems = append(wantItems, uint32(id))
							}
						}
						wantClasses = append(wantClasses, int32(begin), int32(len(wantItems)-begin))
					}
					if !slices.Equal(vec.Vec_IntArray(p.vClasses), wantClasses) {
						t.Fatalf("classes%v want%v", vec.Vec_IntArray(p.vClasses), wantClasses)
					}
					for i, want := range wantItems {
						if Gia_IsoGetItem(p, int32(i)) != want || Gia_IsoGetValue(p, int32(i)) != 0xa5000000+uint32(i) {
							t.Fatalf("packed bucket%d changed item or low-word signature", i)
						}
						if Gia_ManObj(graph, int32(i)).Value != 0x100+uint32(i) {
							t.Fatal("preparation changed graph node signatures")
						}
					}
				})
			}
		}
	}
}

func TestIsoPrepareResetsSourceTerminalLevels(t *testing.T) {
	graph := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	root := Gia_ManAppendAnd(graph, a, b)
	Gia_ManAppendCo(graph, root)
	p := Gia_IsoManStart(graph)
	defer Gia_IsoManStop(p)
	for i := range p.pLevels {
		p.pLevels[i] = -7
	}
	Gia_IsoPrepare(p)
	if !slices.Equal(p.pLevels, []int32{0, 0, 0, 1, 2}) {
		t.Fatalf("source terminal reinitialization lost: %v", p.pLevels)
	}
}
