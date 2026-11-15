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

func TestCofactorTypedPivotMatrix(t *testing.T) {
	rebuilds := 0
	for signs := int32(0); signs < 8; signs++ {
		for _, layout := range []bool{false, true} {
			for _, regs := range []bool{false, true} {
				src := quantificationReviewGraph(signs, regs, false)
				if layout {
					src = Gia_ManDupOrderDfsReverse(src)
				}
				for _, obj := range src.pObjs {
					if Gia_ObjIsCand(obj) == 0 {
						continue
					}
					raw := Gia_ManDupCofInt(src, obj.id)
					cofReviewCheckRebuild(t, src, raw)
					clean := Gia_ManDupCof(src, obj.id)
					cofReviewCheckRebuild(t, src, clean)
					Gia_ManStop(raw)
					Gia_ManStop(clean)
					rebuilds += 2
				}
				for _, limit := range []int32{0, 1, 2, 100} {
					out := Gia_ManDupCofAll(src, limit, 0)
					cofReviewCheckRebuild(t, src, out)
					Gia_ManStop(out)
					rebuilds++
				}
				Gia_ManStop(src)
			}
		}
	}
	t.Logf("%d typed/reordered rebuilds checked exhaustively", rebuilds)
}

func TestCofactorTransferDropsMissingAndDeduplicates(t *testing.T) {
	original, cofactored, next := Gia_ManStart(16), Gia_ManStart(16), Gia_ManStart(16)
	for i := 0; i < 5; i++ {
		Gia_ManAppendCi(original)
		Gia_ManAppendCi(cofactored)
		Gia_ManAppendCi(next)
	}
	for i, lit := range []uint32{2, 4, 6, 8, 10} {
		original.pObjs[i+1].Value = lit
	}
	// Polarity is irrelevant to transferring a signal ID; preserve first order.
	cofactored.pObjs[1].Value = 7
	cofactored.pObjs[2].Value = 6
	cofactored.pObjs[3].Value = 0
	cofactored.pObjs[4].Value = ^uint32(0)
	cofactored.pObjs[5].Value = 2
	signals := Vec_IntAlloc(5)
	for i := int32(1); i <= 5; i++ {
		Vec_IntPush(signals, i)
	}
	got := Gia_ManTransfer(original, cofactored, next, signals)
	if !slices.Equal(Vec_IntArray(got), []int32{3, 1}) {
		t.Fatalf("transferred %v", Vec_IntArray(got))
	}
	if !slices.Equal(Vec_IntArray(signals), []int32{1, 2, 3, 4, 5}) {
		t.Fatal("mutated input signals")
	}
	Gia_ManStop(original)
	Gia_ManStop(cofactored)
	Gia_ManStop(next)
}

func TestCofactorInvalidPivotsAndLimit(t *testing.T) {
	g := cofReviewGraph(7)
	for _, id := range []int32{-1, 0, g.nObjs, g.nObjs + 1, Gia_ManCo(g, 0).id} {
		if Gia_ManDupCofInt(g, id) != nil || Gia_ManDupCof(g, id) != nil {
			t.Fatalf("invalid pivot%d accepted", id)
		}
	}
	signals := Vec_IntAlloc(201)
	for i := 0; i < 201; i++ {
		Vec_IntPush(signals, Gia_ManCi(g, 0).id)
	}
	text := hashReviewStdout(t, func() {
		if Gia_ManDupCofAllInt(g, signals, 0) != nil {
			t.Fatal("accepted more than 200 signals")
		}
	})
	if text != "Too many signals to cofactor.\n" {
		t.Fatalf("diagnostic %q", text)
	}
	Gia_ManStop(g)
}
