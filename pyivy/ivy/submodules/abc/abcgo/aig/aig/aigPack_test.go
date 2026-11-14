package aig

import (
	"bytes"
	"math/bits"
	"reflect"
	"strings"
	"testing"
)

func TestPackedSimulationMatchesAllLanes(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{a, InvalidEdge}})
	later := g.AND(b, c.Not())
	Aig_ObjConnect(g, buffer, later, InvalidEdge)
	for _, edge := range []Edge{g.AND(a, b), g.AND(a.Not(), b), g.AND(a, b.Not()), g.AND(a.Not(), b.Not()), g.NewXor(a, c), buffer, g.Const0(), g.Const1()} {
		g.AddOutput(edge)
	}
	p := Aig_ManPackAlloc(g)
	words := []uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0}
	for i, w := range words {
		p.SetInputWord(i, w)
	}
	Aig_ManPackSimulate(p)
	for lane := 0; lane < 64; lane++ {
		in := make([]bool, len(words))
		for i, w := range words {
			in[i] = w&(uint64(1)<<lane) != 0
		}
		want, err := g.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		for i, output := range g.outputs {
			got := p.vSigns[output.NodeID()]&(uint64(1)<<lane) != 0
			if got != want[i] {
				t.Fatalf("lane %d output %d", lane, i)
			}
		}
	}
	if p.vSigns[0] != ^uint64(0) {
		t.Fatal("constant signature")
	}
}
func TestPackedPatternsAtomicAndLiteralPolarity(t *testing.T) {
	g := NewGraph()
	g.NewInput()
	g.NewInput()
	p := Aig_ManPackAlloc(g)
	if Aig_ManPackAddPatternTry(p, 5, []int32{0, 3}) != 1 {
		t.Fatal("failed initial pattern")
	}
	mask := uint64(1) << 5
	if p.vPiPats[0]&mask == 0 || p.vPiPats[1]&mask != 0 || p.vPiCare[0]&mask == 0 || p.vPiCare[1]&mask == 0 {
		t.Fatal(p.vPiPats, p.vPiCare)
	}
	beforeWords, beforeCare := p.InputWords(), p.CareWords()
	if Aig_ManPackAddPatternTry(p, 5, []int32{1, 2}) != 0 {
		t.Fatal("accepted conflict")
	}
	if !reflect.DeepEqual(beforeWords, p.InputWords()) || !reflect.DeepEqual(beforeCare, p.CareWords()) {
		t.Fatal("partial conflicting write")
	}
	if Aig_ManPackAddPatternTry(p, 6, []int32{0, 1}) != 0 {
		t.Fatal("accepted contradictory cube")
	}
	if !reflect.DeepEqual(beforeWords, p.InputWords()) || !reflect.DeepEqual(beforeCare, p.CareWords()) {
		t.Fatal("partial contradictory cube write")
	}
	if Aig_ManPackCountCares(p) != 2 {
		t.Fatal(Aig_ManPackCountCares(p))
	}
}
func TestPackedPatternCapacityAndZeroLane(t *testing.T) {
	g := NewGraph()
	for i := 0; i < 7; i++ {
		g.NewInput()
	}
	p := Aig_ManPackAlloc(g)
	Aig_ManPackSetRandom(p)
	for assignment := 0; assignment < 128; assignment++ {
		cube := make([]int32, 7)
		for i := range cube {
			cube[i] = int32(2 * i)
			if assignment&(1<<i) == 0 {
				cube[i] |= 1
			}
		}
		Aig_ManPackAddPattern(p, cube)
	}
	total, skipped := p.PatternCounts()
	if total != 128 || skipped != 65 {
		t.Fatal(total, skipped)
	}
	for i, care := range p.vPiCare {
		if care != ^uint64(1) {
			t.Fatal(i, care)
		}
		if p.vPiPats[i]&1 != 0 {
			t.Fatal("reserved zero lane overwritten")
		}
	}
	if Aig_ManPackCountCares(p) != 7*63 {
		t.Fatal(Aig_ManPackCountCares(p))
	}
}
func TestPackedConstantCandidatesAndStatistics(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	nearZero := g.AND(a, b)
	nearOne := g.AND(a.Not(), b.Not())
	g.AddOutput(nearZero)
	p := Aig_ManPackAlloc(g)
	p.SetInputWord(0, 1)
	p.SetInputWord(1, 1)
	Aig_ManPackSimulate(p)
	candidates := Aig_ManPackConstNodes(p)
	if !reflect.DeepEqual(candidates, []int32{int32(nearZero.NodeID()), int32(nearOne.NodeID())}) {
		t.Fatal(candidates)
	}
	var out bytes.Buffer
	p.WriteStats(&out)
	if strings.Count(out.String(), "\n") != 33 || strings.Contains(out.String(), "NaN") {
		t.Fatal(out.String())
	}
	out.Reset()
	p.WriteCare(&out)
	if strings.Count(out.String(), "\n") != 2 {
		t.Fatal(out.String())
	}
}
func TestPackedEmptyGraphAndWordHelpers(t *testing.T) {
	p := Aig_ManPackAlloc(NewGraph())
	Aig_ManPackSetRandom(p)
	Aig_ManPackSimulate(p)
	var out bytes.Buffer
	p.WriteStats(&out)
	if strings.Contains(out.String(), "NaN") {
		t.Fatal(out.String())
	}
	for _, w := range []uint64{0, 1, 1 << 63, ^uint64(0), 0x800000001} {
		if Aig_Word6CountOnes(w) != int32(bits.OnesCount64(w)) {
			t.Fatal(w)
		}
		want := int32(0)
		if bits.OnesCount64(w) <= 1 {
			want = 1
		}
		if Aig_Word6HasOneBit(w) != want {
			t.Fatal(w)
		}
	}
}
func TestPackedInvalidInputs(t *testing.T) {
	for _, kind := range []string{"negative lane", "large lane", "negative literal", "large literal", "changed interface"} {
		t.Run(kind, func(t *testing.T) {
			g := NewGraph()
			g.NewInput()
			p := Aig_ManPackAlloc(g)
			defer func() {
				if recover() == nil {
					t.Fatal("accepted invalid input")
				}
			}()
			switch kind {
			case "negative lane":
				Aig_ManPackAddPatternTry(p, -1, nil)
			case "large lane":
				Aig_ManPackAddPatternTry(p, 64, nil)
			case "negative literal":
				Aig_ManPackAddPatternTry(p, 1, []int32{-1})
			case "large literal":
				Aig_ManPackAddPatternTry(p, 1, []int32{2})
			case "changed interface":
				g.NewInput()
				Aig_ManPackSimulate(p)
			}
		})
	}
}

func TestPackedManualSeedingPreservesCares(t *testing.T) {
	g := NewGraph()
	g.NewInput()
	p := Aig_ManPackAlloc(g)
	Aig_ManPackAddPatternTry(p, 63, []int32{0})
	Aig_ManPackAddPatternTry(p, 32, []int32{1})
	care, before := p.vPiCare[0], p.vPiPats[0]&p.vPiCare[0]
	for i := 0; i < 10; i++ {
		p.SetInputWord(0, ^uint64(0)>>uint(i))
		if p.vPiPats[0]&care != before || p.vPiCare[0] != care {
			t.Fatal("manual seeding changed packed cube")
		}
	}
	p.SetInputWord(0, ^uint64(0))
	if p.vPiPats[0]&care != before {
		t.Fatal("manual seeding changed packed cube")
	}
}
func TestPackedSnapshotsIndependent(t *testing.T) {
	g := NewGraph()
	g.NewInput()
	p := Aig_ManPackAlloc(g)
	p.SetInputWord(0, 123)
	words := p.InputWords()
	words[0] = 0
	if p.vPiPats[0] != 123 {
		t.Fatal("aliased input snapshot")
	}
	care := p.CareWords()
	care[0] = 123
	if p.vPiCare[0] != 0 {
		t.Fatal("aliased care snapshot")
	}
	Aig_ManPackSimulate(p)
	signatures := p.Signatures()
	signatures[0] = 0
	if p.vSigns[0] != ^uint64(0) {
		t.Fatal("aliased signature snapshot")
	}
}
