package gia

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestWindowExtractionRetainsSimpleAnds(t *testing.T) {
	for _, poOnly := range []int32{0, 1} {
		src := Gia_ManStart(8)
		src.fGiaSimple = 1
		a := Gia_ManAppendCi(src)
		root := Gia_ManAppendAnd(src, a, a^1)
		Gia_ManAppendCo(src, root)
		selected := Vec_IntAlloc(1)
		Vec_IntPush(selected, 0)
		window := Gia_ManExtractWin(src, selected, poOnly)
		if window.fGiaSimple != 1 {
			t.Fatal("extraction lost permission for repeated-fanin ANDs")
		}
		for _, input := range []bool{false, true} {
			checkQuantificationOutputs(t, window, []bool{input}, []bool{false})
		}
	}
}

// Extraction exposes regular object boundaries, not signed/ordered selected
// COs. Insertion restores the original CO phases and logical interface order.
func TestWindowTypedLayoutsAndBoundaryTruth(t *testing.T) {
	cases := 0
	for signs := 0; signs < 8; signs++ {
		for typed := 0; typed < 2; typed++ {
			for layout := 0; layout < 4; layout++ {
				for _, mask := range []int{0, 1, 2, 4, 8, 15, 16, 32, 64, 128, 48, 96, 129, 170, 255} {
					for po := int32(0); po < 2; po++ {
						g := windowReviewGraph(signs, typed, layout)
						Gia_ManAppendCo(g, Gia_ObjToLit(g, Gia_ManCi(g, 0))^1)
						Gia_ManAppendCo(g, 0)
						Gia_ManAppendCo(g, 1)
						Gia_ManAppendCo(g, Gia_ManAppendBuf(g, Gia_ObjToLit(g, Gia_ManCi(g, 3))^1))
						Gia_ManSetRegNum(g, 1)
						if err := g.SetAigerProperties(1, 1); err != nil {
							t.Fatal(err)
						}
						g.SetIdentity("window-review", "window-review.aig")
						g.aigerLatchInits = []int32{1}
						g.aigerComment = []byte("window resets")
						selected := Vec_IntAlloc(8)
						for i := 0; i < 8; i++ {
							if mask>>i&1 != 0 {
								Vec_IntPush(selected, int32(i))
							}
						}
						var inputs, boundary, gates *Vec_Int_t
						Gia_ManPrepareWin(g, selected, &inputs, &boundary, &gates, po)
						if po != 0 {
							expected := make(map[int32]bool)
							for _, index := range Vec_IntArray(selected) {
								expected[Gia_ObjFaninId0p(g, Gia_ManCo(g, index))] = true
							}
							if len(expected) != Vec_IntSize(boundary) {
								t.Fatalf("signs%d typed%d layout%d mask%d: dropped or duplicated selected boundary", signs, typed, layout, mask)
							}
							for _, id := range Vec_IntArray(boundary) {
								if !expected[id] {
									t.Fatalf("unexpected boundary%d", id)
								}
							}
						}
						w := Gia_ManExtractWin(g, selected, po)
						if w.nRegs != 0 || w.nConstrs != 0 || w.aigerBadOutputs != 0 {
							t.Fatal("a combinational boundary window inherited source property roles")
						}
						for assignment := 0; assignment < 16; assignment++ {
							ci := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
							values, _, err := g.evaluateObjects(ci)
							if err != nil {
								t.Fatal(err)
							}
							windowInputs := make([]bool, Vec_IntSize(inputs))
							for i, id := range Vec_IntArray(inputs) {
								windowInputs[i] = values[id]
							}
							want := make([]bool, Vec_IntSize(boundary))
							for i, id := range Vec_IntArray(boundary) {
								want[i] = values[id]
							}
							checkQuantificationOutputs(t, w, windowInputs, want)
						}
						if po == 0 {
							r := Gia_ManInsertWin(g, selected, w)
							if r.nRegs != 1 || r.nConstrs != 1 || r.aigerBadOutputs != 1 || !slices.Equal(r.aigerLatchInits, []int32{1}) || string(r.aigerComment) != "window resets" {
								t.Fatal("insertion lost the source property partition or reset state")
							}
							if name, spec := r.Identity(); name != "window-review" || spec != "window-review.aig" {
								t.Fatal("insertion lost identity")
							}
							restored := partitionReviewRoundTrip(t, r, 1, 1)
							for assignment := 0; assignment < 16; assignment++ {
								ci := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
								want, err := g.Evaluate(ci)
								if err != nil {
									t.Fatal(err)
								}
								checkQuantificationOutputs(t, r, ci, want)
								checkQuantificationOutputs(t, restored, ci, want)
							}
						}
						cases++
					}
				}
			}
		}
	}
	t.Logf("%d typed/layout/subset boundary cases, %d reconstruction checks and %d source assignments", cases, cases/2, cases*16)
}

func BenchmarkWindowPrepare(b *testing.B) {
	for _, typed := range []bool{false, true} {
		name := "AND"
		if typed {
			name = "MUX"
		}
		b.Run(name, func(b *testing.B) {
			g := Gia_ManStart(2048)
			a, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			root := a
			if typed {
				g.EnableMuxes()
			}
			for i := 0; i < 1024; i++ {
				if typed && i%4 == 3 {
					root = Gia_ManAppendMuxReal(g, d, root, c)
				} else {
					root = Gia_ManAppendAnd(g, root, c)
				}
			}
			Gia_ManAppendCo(g, root)
			selected := windowReviewSelection(1)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var inputs, outputs, gates *Vec_Int_t
				Gia_ManPrepareWin(g, selected, &inputs, &outputs, &gates, 0)
			}
		})
	}
}

func TestWindowWrappersEmptySelectionAndArgumentChecks(t *testing.T) {
	for _, choice := range [][2]int32{{0, 0}, {1, 1}} {
		for _, call := range []func(){
			func() { Gia_ManExtractWindow(Gia_ManStart(1), choice[0], choice[1], 0) },
			func() {
				giaBalanceWindow(Gia_ManStart(1), choice[0], choice[1], 0, func(*Gia_Man_t) *Gia_Man_t { t.Fatal("invalid arguments reached balance engine"); return nil })
			},
		} {
			func() {
				defer func() {
					if recover() == nil {
						t.Error("expected mutually exclusive target/window assertion")
					}
				}()
				call()
			}()
		}
	}
	for _, driver := range []bool{false, true} {
		src := windowReviewGraph(3, 0, 3)
		Gia_ManSetRegNum(src, 1)
		if err := src.SetAigerProperties(1, 1); err != nil {
			t.Fatal(err)
		}
		src.aigerLatchInits = []int32{1}
		src.aigerComment = []byte("fallback")
		var out *Gia_Man_t
		text := hashReviewStdout(t, func() {
			if driver {
				out = giaBalanceWindow(src, 100, 0, 1, func(*Gia_Man_t) *Gia_Man_t { t.Fatal("empty selection called balance engine"); return nil })
			} else {
				out = Gia_ManExtractWindow(src, 100, 0, 1)
			}
		})
		if !strings.Contains(text, "less than the target level (100)") || !strings.Contains(text, "Collected 0 outputs to extract.") {
			t.Fatalf("missing source fallback diagnostics: %q", text)
		}
		if out == src || out.nRegs != 1 || out.nConstrs != 1 || out.aigerBadOutputs != 1 || !slices.Equal(out.aigerLatchInits, src.aigerLatchInits) || string(out.aigerComment) != "fallback" {
			t.Fatal("empty window did not return independent full-interface duplicate")
		}
		for mask := 0; mask < 16; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			want, err := src.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			checkQuantificationOutputs(t, out, inputs, want)
		}
	}
	for _, choice := range [][2]int32{{1, 0}, {0, 50}, {0, 100}} {
		src := windowReviewGraph(2, 1, 1)
		selected := Gia_ManFindLatest(src, choice[0], choice[1])
		expected := Gia_ManExtractWin(src, selected, 1)
		var out *Gia_Man_t
		text := hashReviewStdout(t, func() { out = Gia_ManExtractWindow(src, choice[0], choice[1], 1) })
		if !strings.Contains(text, fmt.Sprintf("Collected %d outputs to extract.", Vec_IntSize(selected))) {
			t.Fatal("missing selected output count")
		}
		for mask := 0; mask < 1<<Gia_ManCiNum(out); mask++ {
			inputs := make([]bool, Gia_ManCiNum(out))
			for i := range inputs {
				inputs[i] = mask>>i&1 != 0
			}
			want, err := expected.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			checkQuantificationOutputs(t, out, inputs, want)
		}
	}
}
