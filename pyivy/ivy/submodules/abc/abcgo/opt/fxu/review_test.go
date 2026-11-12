package fxu

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
)

// Goldens are produced by the unchanged C algorithm, not the Go implementation.
func TestOriginalSourceGoldens(t *testing.T) {
	for mode, name := range []string{"single", "double", "complement"} {
		n := &Network{SOPs: []string{"", "", "", "", "", "111-- 1\n11-1- 1\n11--1 1\n"}, Fanins: [][]int32{nil, nil, nil, nil, nil, {0, 1, 2, 3, 4}}}
		if mode > 0 {
			n.SOPs = []string{"", "", "", "", "", "1-1-- 1\n-11-- 1\n", "1--1- 1\n-1-1- 1\n", "00--1 1\n"}
			n.Fanins = [][]int32{nil, nil, nil, nil, nil, {0, 1, 2, 3, 4}, {0, 1, 2, 3, 4}, {0, 1, 2, 3, 4}}
		}
		p := networkParams(n)
		p.OnlyS = mode == 0
		p.Use0 = mode > 0
		p.UseCompl = mode == 2
		count, err := Extract(p)
		if err != nil {
			t.Fatal(err)
		}
		var b strings.Builder
		fmt.Fprintf(&b, "COUNT %d\n", count)
		for i, s := range p.NewSOPs {
			if s == "" || i < len(p.SOPs) && s == p.SOPs[i] {
				continue
			}
			fmt.Fprintf(&b, "NODE %d", i)
			for _, f := range p.NewFanins[i] {
				fmt.Fprintf(&b, " %d", f)
			}
			fmt.Fprintf(&b, "\n%s", s)
		}
		want, err := os.ReadFile("test_vectors/" + name + ".txt")
		if err != nil {
			t.Fatal(err)
		}
		if b.String() != string(want) {
			t.Fatalf("%s\ngot %swant %s", name, b.String(), want)
		}
	}
}

func TestHeapMutationAndReinsertion(t *testing.T) {
	h := Fxu_HeapSingleStart()
	r := rand.New(rand.NewSource(71))
	items := make([]*Fxu_Single, 100)
	for i := range items {
		items[i] = &Fxu_Single{Weight: int32(r.Intn(1000))}
		Fxu_HeapSingleInsert(h, items[i])
	}
	for step := 0; step < 1000; step++ {
		x := items[r.Intn(len(items))]
		x.Weight = int32(r.Intn(1000))
		Fxu_HeapSingleUpdate(h, x)
		Fxu_HeapSingleCheck(h)
	}
	last := int32(1001)
	for len(h.items) > 0 {
		x := Fxu_HeapSingleGetMax(h)
		if x.Weight > last {
			t.Fatal("heap order")
		}
		last = x.Weight
		if x.heapIndex != -1 {
			t.Fatal("stale heap membership")
		}
	}
	Fxu_HeapSingleInsert(h, items[0])
	Fxu_HeapSingleDelete(h, items[0])
	Fxu_HeapSingleInsert(h, items[0])
	Fxu_HeapSingleCheck(h)
}

func TestZeroLimitAndRejectedCovers(t *testing.T) {
	n := &Network{SOPs: []string{"", "", "", "111 1\n110 1\n"}, Fanins: [][]int32{nil, nil, nil, {0, 1, 2}}}
	p := networkParams(n)
	p.NodesExt = 0
	if c, e := Extract(p); c != 0 || e != nil {
		t.Fatal(c, e)
	}
	if p.NewSOPs[3] != p.SOPs[3] {
		t.Fatal("zero budget changed cover")
	}
	for _, s := range []string{"11- 1\n111 1\n", "111 1\n111 1\n", "111 0\n110 1\n"} {
		p = networkParams(n)
		p.SOPs[3] = s
		if _, e := Extract(p); e == nil {
			t.Fatalf("accepted invalid cover %q", s)
		}
	}
	p = networkParams(n)
	p.Fanins[3] = []int32{0, 1, 3}
	if _, e := Extract(p); e == nil {
		t.Fatal("accepted cycle")
	}
}

func TestCombinedComplementRewrite(t *testing.T) {
	n := &Network{Inputs: []int32{0, 1, 2, 3, 4, 5, 6}, Outputs: []int32{7, 8, 9}, SOPs: []string{"", "", "", "", "", "", "", "1-1---- 1\n-11---- 1\n", "1--1--- 1\n-1-1--- 1\n", "00--1-- 1\n00---1- 1\n00----1 1\n"}, Fanins: make([][]int32, 10)}
	for i := 7; i < 10; i++ {
		n.Fanins[i] = []int32{0, 1, 2, 3, 4, 5, 6}
	}
	p := networkParams(n)
	p.UseCompl = true
	p.NodesExt = 1
	m, err := createMatrix(p)
	if err != nil {
		t.Fatal(err)
	}
	var s *Fxu_Single
	var d *Fxu_Double
	weight := Fxu_Select(m, &s, &d)
	if s == nil || d == nil || weight != 5 {
		t.Fatal("combined complement not selected", weight, s, d)
	}
	count, err := Extract(p)
	if err != nil || count != 1 {
		t.Fatal(count, err)
	}
	compareNetworkTruth(t, n, &Network{Inputs: n.Inputs, Outputs: n.Outputs, SOPs: p.NewSOPs, Fanins: p.NewFanins})
	if literalCost(n.SOPs, n.Fanins)-literalCost(p.NewSOPs, p.NewFanins) != int(weight) {
		t.Fatal("combined savings mismatch")
	}
}
