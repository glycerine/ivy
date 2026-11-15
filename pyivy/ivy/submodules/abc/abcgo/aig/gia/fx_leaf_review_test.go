package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/bool/dec"
)

func TestFxRootLeafOnlyRequiresSelectedSignal(t *testing.T) {
	for phase := int32(0); phase < 2; phase++ {
		g := Gia_ManStart(8)
		a := Gia_ManAppendCi(g)
		graph := dec.Dec_GraphCreateLeaf(2, 3, phase)
		dec.Dec_GraphSetLeafFunc(graph, 2, a)
		if got := Gia_ManGraphToAig(g, graph); got != a^phase {
			t.Fatalf("phase%d got%d want%d", phase, got, a^phase)
		}
		if g.nObjs != 2 {
			t.Fatal("leaf conversion allocated gates")
		}
		if dec.Dec_GraphLeafFunc(graph, 0) != nil || dec.Dec_GraphLeafFunc(graph, 1) != nil {
			t.Fatal("unused leaf signals changed")
		}
		dec.Dec_GraphFree(graph)
		Gia_ManStop(g)
	}
}

func TestFxInternalGraphIgnoresUnusedLeafSignals(t *testing.T) {
	g := Gia_ManStart(16)
	defer Gia_ManStop(g)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	graph := dec.Dec_GraphCreate(3)
	defer dec.Dec_GraphFree(graph)
	dec.Dec_GraphSetLeafFunc(graph, 0, a)
	dec.Dec_GraphSetLeafFunc(graph, 1, b)
	root := dec.Dec_GraphAddNodeAnd(graph, dec.Dec_EdgeCreate(0, 0), dec.Dec_EdgeCreate(1, 1))
	dec.Dec_GraphSetRoot(graph, root)
	Gia_ManHashAlloc(g)
	lit := Gia_ManGraphToAig(g, graph)
	for input := 0; input < 4; input++ {
		values := cofReviewValues(g, input)
		if got, want := values[lit>>1] != (lit&1 != 0), input&1 != 0 && input&2 == 0; got != want {
			t.Fatalf("input%d got%t want%t", input, got, want)
		}
	}
	if dec.Dec_GraphLeafFunc(graph, 2) != nil {
		t.Fatal("unused leaf changed")
	}
}

func TestFxNativeGraphConversion(t *testing.T) {
	exe := compileGiaSourceOracle(t, "fx_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C graph conversion: %v\n%s", err, native)
	}
	var got strings.Builder
	for signal := 0; signal < 2; signal++ {
		g := Gia_ManStart(8)
		graph := dec.Dec_GraphCreateConst0()
		if signal != 0 {
			dec.Dec_GraphFree(graph)
			graph = dec.Dec_GraphCreateConst1()
		}
		fmt.Fprintf(&got, "constant %d lit%d objs%d\n", signal, Gia_ManGraphToAig(g, graph), g.nObjs)
		dec.Dec_GraphFree(graph)
		Gia_ManStop(g)
	}
	for n := int32(1); n <= 4; n++ {
		for leaf := int32(0); leaf < n; leaf++ {
			for phase := int32(0); phase < 2; phase++ {
				for signal := int32(0); signal < 3; signal++ {
					for unused := 0; unused < 2; unused++ {
						if unused != 0 && n < 2 {
							continue
						}
						g := Gia_ManStart(32)
						vars := make([]int32, n)
						for i := range vars {
							vars[i] = Gia_ManAppendCi(g)
						}
						graph := dec.Dec_GraphCreateLeaf(leaf, n, phase)
						if unused != 0 {
							dec.Dec_GraphAddNodeAnd(graph, dec.Dec_EdgeCreate(0, 0), dec.Dec_EdgeCreate(1, 0))
							dec.Dec_GraphSetRoot(graph, dec.Dec_EdgeCreate(leaf, phase))
						}
						value := signal
						if signal == 2 {
							value = vars[leaf]
						}
						dec.Dec_GraphSetLeafFunc(graph, leaf, value)
						lit := Gia_ManGraphToAig(g, graph)
						fmt.Fprintf(&got, "var n%d leaf%d phase%d signal%d unused%d lit%d objs%d\n", n, leaf, phase, signal, unused, lit, g.nObjs)
						for input := 0; input < 1<<n; input++ {
							want := signal == 1
							if signal == 2 {
								want = input>>leaf&1 != 0
							}
							want = want != (phase != 0)
							values := cofReviewValues(g, input)
							if actual := values[lit>>1] != (lit&1 != 0); actual != want {
								t.Fatal("leaf truth changed")
							}
						}
						for i := int32(0); i < n; i++ {
							if i != leaf && dec.Dec_GraphLeafFunc(graph, i) != nil {
								t.Fatal("unused signal changed")
							}
						}
						if unused != 0 && dec.Dec_GraphNodeFunc(graph, n) != nil {
							t.Fatal("unused internal node evaluated")
						}
						dec.Dec_GraphFree(graph)
						Gia_ManStop(g)
					}
				}
			}
		}
	}
	for n := int32(2); n <= 5; n++ {
		for a := int32(0); a < 2; a++ {
			for b := int32(0); b < 2; b++ {
				for phase := int32(0); phase < 2; phase++ {
					g := Gia_ManStart(32)
					vars := make([]int32, n)
					for i := range vars {
						vars[i] = Gia_ManAppendCi(g)
					}
					graph := dec.Dec_GraphCreate(n)
					dec.Dec_GraphSetLeafFunc(graph, 0, vars[0])
					dec.Dec_GraphSetLeafFunc(graph, 1, vars[1])
					root := dec.Dec_GraphAddNodeAnd(graph, dec.Dec_EdgeCreate(0, a), dec.Dec_EdgeCreate(1, b))
					if phase != 0 {
						root.Complement()
					}
					dec.Dec_GraphSetRoot(graph, root)
					Gia_ManHashAlloc(g)
					lit := Gia_ManGraphToAig(g, graph)
					fmt.Fprintf(&got, "and n%d a%d b%d phase%d lit%d node%d objs%d\n", n, a, b, phase, lit, dec.Dec_GraphNodeFunc(graph, int32(root.Node)).(int32), g.nObjs)
					for input := 0; input < 1<<n; input++ {
						want := ((input&1 != 0) != (a != 0)) && ((input&2 != 0) != (b != 0))
						want = want != (phase != 0)
						values := cofReviewValues(g, input)
						if actual := values[lit>>1] != (lit&1 != 0); actual != want {
							t.Fatal("internal graph truth changed")
						}
					}
					dec.Dec_GraphFree(graph)
					Gia_ManStop(g)
				}
			}
		}
	}
	esopReviewCompare(t, native, got.String())
}

func TestFxSopAllowsUnusedTrailingLeaves(t *testing.T) {
	g := Gia_ManStart(16)
	defer Gia_ManStop(g)
	leaves := Vec_IntAlloc(3)
	defer Vec_IntFree(leaves)
	for i := 0; i < 3; i++ {
		Vec_IntPush(leaves, Gia_ManAppendCi(g))
	}
	Gia_ManHashStart(g)
	lit := Gia_ManSopToAig(g, "11 1\n", leaves)
	for input := 0; input < 8; input++ {
		values := cofReviewValues(g, input)
		if actual, want := values[lit>>1] != (lit&1 != 0), input&3 == 3; actual != want {
			t.Fatalf("input%d got%t want%t", input, actual, want)
		}
	}
}

func TestFxNativeSopConversion(t *testing.T) {
	covers := []string{
		" 0\n", " 1\n", "0 1\n", "1 0\n", "- 1\n",
		"11 1\n", "10 1\n01 1\n", "00 0\n11 0\n",
		"11- 1\n-11 1\n", "11- 0\n-11 0\n",
		"1--1 1\n-11- 1\n0-10 1\n", "---- 0\n",
	}
	exe := compileGiaSourceOracle(t, "fx_sop_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C SOP conversion: %v\n%s", err, native)
	}
	var got strings.Builder
	for c, sop := range covers {
		n := strings.IndexByte(sop, ' ')
		for extra := 0; extra <= 2; extra++ {
			for mask := 0; mask < 1<<n; mask++ {
				g := Gia_ManStart(64)
				leaves := Vec_IntAlloc(n + extra)
				for i := 0; i < n+extra; i++ {
					Vec_IntPush(leaves, Gia_ManAppendCi(g)^int32(mask>>i&1))
				}
				Gia_ManHashAlloc(g)
				// The C API stops at the terminating NUL as well.
				lit := Gia_ManSopToAig(g, sop+"\x00ignored", leaves)
				fmt.Fprintf(&got, "sop%d extra%d mask%d lit%d objs%d\n", c, extra, mask, lit, g.nObjs)
				for input := 0; input < 1<<(n+extra); input++ {
					want := false
					for _, row := range strings.Split(strings.TrimSuffix(sop, "\n"), "\n") {
						match := true
						for i := 0; i < n; i++ {
							bit := (input ^ mask) >> i & 1
							if row[i] != '-' && int(row[i]-'0') != bit {
								match = false
							}
						}
						want = want || match
					}
					want = want != (sop[n+1] == '0')
					values := cofReviewValues(g, input)
					if actual := values[lit>>1] != (lit&1 != 0); actual != want {
						t.Fatalf("cover%d extra%d mask%d input%d got%t want%t", c, extra, mask, input, actual, want)
					}
				}
				Vec_IntFree(leaves)
				Gia_ManStop(g)
			}
		}
	}
	esopReviewCompare(t, native, got.String())
}
