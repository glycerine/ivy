package fxu

import (
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func networkParams(n *Network) *Fxu_Data_t {
	p := NewParams()
	p.SOPs = append([]string(nil), n.SOPs...)
	p.Fanins = cloneFanins(n.Fanins)
	p.NodesExt = 20
	p.Log = io.Discard
	return p
}
func compareNetworkTruth(t *testing.T, before, after *Network) {
	t.Helper()
	for bits := 0; bits < 1<<uint(len(before.Inputs)); bits++ {
		input := make([]bool, len(before.Inputs))
		for i := range input {
			input[i] = bits&(1<<uint(i)) != 0
		}
		want, err := before.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		got, err := after.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("bits%d got%v want%v\nold%v\nnew%v\nfans%v", bits, got, want, before.SOPs, after.SOPs, after.Fanins)
		}
	}
}
func literalCost(sops []string, fans [][]int32) int {
	n := 0
	for i, sop := range sops {
		if sop == "" {
			continue
		}
		cover, err := parseSOP(sop, len(fans[i]))
		if err != nil {
			panic(err)
		}
		for _, row := range cover.patterns {
			for _, c := range row {
				if c != '-' {
					n++
				}
			}
		}
	}
	return n
}
func TestSingleExtractionAndCoSingletonOutput(t *testing.T) {
	n := &Network{Inputs: []int32{0, 1, 2, 3, 4}, Outputs: []int32{5}, SOPs: []string{"", "", "", "", "", "111-- 1\n11-1- 1\n11--1 1\n"}, Fanins: [][]int32{nil, nil, nil, nil, nil, {0, 1, 2, 3, 4}}}
	p := networkParams(n)
	p.OnlyS = true
	count, err := Extract(p)
	if err != nil || count != 1 {
		t.Fatal(count, err)
	}
	after := &Network{Inputs: n.Inputs, Outputs: n.Outputs, SOPs: p.NewSOPs, Fanins: p.NewFanins}
	compareNetworkTruth(t, n, after)
	if cost := literalCost(p.NewSOPs, p.NewFanins); cost != 8 {
		t.Fatal("single savings", cost, p.NewSOPs)
	}
}
func TestDoubleAndComplementExtraction(t *testing.T) {
	for _, complement := range []bool{false, true} {
		n := &Network{Inputs: []int32{0, 1, 2, 3, 4}, Outputs: []int32{5, 6, 7}, SOPs: []string{"", "", "", "", "", "1-1-- 1\n-11-- 1\n", "1--1- 1\n-1-1- 1\n", "00--1 1\n"}, Fanins: [][]int32{nil, nil, nil, nil, nil, {0, 1, 2, 3, 4}, {0, 1, 2, 3, 4}, {0, 1, 2, 3, 4}}}
		p := networkParams(n)
		p.UseCompl = complement
		p.Use0 = true
		count, err := Extract(p)
		if err != nil || count < 1 {
			t.Fatal(count, err)
		}
		after := &Network{Inputs: n.Inputs, Outputs: n.Outputs, SOPs: p.NewSOPs, Fanins: p.NewFanins}
		compareNetworkTruth(t, n, after)
		if cost := literalCost(p.NewSOPs, p.NewFanins); cost > literalCost(n.SOPs, n.Fanins) {
			t.Fatal("extraction grew cost", cost, p.NewSOPs)
		}
	}
}
func TestRandomNetworkFunctionPreservation(t *testing.T) {
	rng := rand.New(rand.NewSource(831))
	for trial := 0; trial < 200; trial++ {
		n := &Network{Inputs: []int32{0, 1, 2, 3, 4}, SOPs: make([]string, 8), Fanins: make([][]int32, 8)}
		for id := 5; id < 8; id++ {
			n.Outputs = append(n.Outputs, int32(id))
			n.Fanins[id] = []int32{0, 1, 2, 3, 4}
			var b strings.Builder
			for cube := 0; cube < 8; cube++ {
				for v := 0; v < 5; v++ {
					b.WriteByte("-01"[rng.Intn(3)])
				}
				phase := '1'
				if trial%3 == 0 {
					phase = '0'
				}
				fmt.Fprintf(&b, " %c\n", phase)
			}
			sop, err := NormalizeSOP(b.String(), 5)
			if err != nil {
				t.Fatal(err)
			}
			n.SOPs[id] = sop
		}
		p := networkParams(n)
		p.UseCompl = trial%2 == 0
		p.Use0 = trial%4 == 0
		if trial%5 == 0 {
			p.OnlyD = true
		}
		count, err := Extract(p)
		if err != nil {
			t.Fatalf("trial%d %v", trial, err)
		}
		after := &Network{Inputs: n.Inputs, Outputs: n.Outputs, SOPs: p.NewSOPs, Fanins: p.NewFanins}
		compareNetworkTruth(t, n, after)
		if count > 20 || literalCost(p.NewSOPs, p.NewFanins) > literalCost(n.SOPs, n.Fanins) {
			t.Fatal("resource/cost", trial, count)
		}
	}
}
