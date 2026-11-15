package gia

import (
	"fmt"
	"slices"
	"testing"
)

// Scalar evaluation is independent of cofactor scratch mappings and hashing.
func cofReviewValues(g *Gia_Man_t, bits int) []bool {
	values := make([]bool, g.nObjs)
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		values[Gia_ManCi(g, i).id] = bits&(1<<i) != 0
	}
	literal := func(lit int32) bool { return values[lit>>1] != (lit&1 != 0) }
	for id := int32(1); id < g.nObjs; id++ {
		a, b, c, kind := g.Gate(id)
		switch kind {
		case 1:
		case 2:
			values[id] = literal(a) && literal(b)
		case 3:
			values[id] = literal(a) != literal(b)
		case 4:
			if literal(a) {
				values[id] = literal(b)
			} else {
				values[id] = literal(c)
			}
		case 5, 6:
			values[id] = literal(a)
		default:
			panic("invalid test graph")
		}
	}
	return values
}

func cofReviewCheckRebuild(t *testing.T, src, out *Gia_Man_t) {
	t.Helper()
	if out == nil {
		t.Fatal("cofactor rebuild failed")
	}
	if Gia_ManCiNum(src) != Gia_ManCiNum(out) || Gia_ManCoNum(src) != Gia_ManCoNum(out) || src.nRegs != out.nRegs {
		t.Fatal("cofactor changed interface dimensions")
	}
	for bits := 0; bits < 1<<Gia_ManCiNum(src); bits++ {
		before, after := cofReviewValues(src, bits), cofReviewValues(out, bits)
		for i := int32(0); i < Gia_ManCoNum(src); i++ {
			if before[Gia_ManCo(src, i).id] != after[Gia_ManCo(out, i).id] {
				t.Fatalf("input%d output%d differs", bits, i)
			}
		}
	}
}

func TestCofactorRebuildTruthMatrix(t *testing.T) {
	rebuilds := 0
	for seed := uint32(0); seed < 64; seed++ {
		src := cofReviewGraph(seed)
		Gia_ManHashStop(src)
		Gia_ManSetRegNum(src, int32(seed%3))
		for _, obj := range src.pObjs {
			if Gia_ObjIsCand(obj) == 0 {
				continue
			}
			for _, clean := range []bool{false, true} {
				var out *Gia_Man_t
				if clean {
					out = Gia_ManDupCof(src, obj.id)
				} else {
					out = Gia_ManDupCofInt(src, obj.id)
				}
				cofReviewCheckRebuild(t, src, out)
				Gia_ManStop(out)
				rebuilds++
			}
		}
		for _, limit := range []int32{0, 1, 2, 4, 1000} {
			out := Gia_ManDupCofAll(src, limit, 0)
			cofReviewCheckRebuild(t, src, out)
			Gia_ManStop(out)
			rebuilds++
		}
		Gia_ManStop(src)
	}
	t.Logf("%d rebuilt networks checked exhaustively", rebuilds)
}

func TestCofactorRebuildInterfaceMetadata(t *testing.T) {
	for _, op := range []string{"raw", "clean", "all"} {
		for _, layout := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/reverse%t", op, layout), func(t *testing.T) {
				src := quantificationReviewGraph(5, true, false)
				if layout {
					src = Gia_ManDupOrderDfsReverse(src)
				}
				if err := src.SetAigerProperties(1, 1); err != nil {
					t.Fatal(err)
				}
				src.aigerLatchInits = []int32{1}
				src.aigerComment = []byte("cofactor ownership")
				var out *Gia_Man_t
				switch op {
				case "raw":
					out = Gia_ManDupCofInt(src, Gia_ManCi(src, 1).id)
				case "clean":
					out = Gia_ManDupCof(src, Gia_ManCi(src, 1).id)
				case "all":
					out = Gia_ManDupCofAll(src, 2, 0)
				}
				cofReviewCheckRebuild(t, src, out)
				if out.nConstrs != src.nConstrs || out.aigerBadOutputs != src.aigerBadOutputs || !slices.Equal(out.aigerLatchInits, src.aigerLatchInits) || !slices.Equal(out.aigerComment, src.aigerComment) {
					t.Fatalf("lost interface metadata: bad%d constraints%d resets%v comment%q", out.aigerBadOutputs, out.nConstrs, out.aigerLatchInits, out.aigerComment)
				}
				out.aigerLatchInits[0] = 0
				out.aigerComment[0] = 'X'
				if src.aigerLatchInits[0] != 1 || string(src.aigerComment) != "cofactor ownership" {
					t.Fatal("aliased interface metadata")
				}
				Gia_ManStop(out)
				Gia_ManStop(src)
			})
		}
	}
}
