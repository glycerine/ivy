package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestTernaryRandomHelpersConsumeSourceStream(t *testing.T) {
	Aig_ManRandom(1)
	defer Aig_ManRandom(1)
	r := NewAigRandomSource()
	r.Uint32() // Reset returns and consumes the first source word.
	for i := 0; i < 32; i++ {
		want2 := 1 + int32(r.Uint32()&1)
		if got := Aig_XsimRand2(); got != want2 {
			t.Errorf("binary draw %d = %d, want %d", i, got, want2)
		}
		var want3 int32
		for want3 == 0 {
			want3 = int32(r.Uint32() & 3)
		}
		if got := Aig_XsimRand3(); got != want3 {
			t.Errorf("ternary draw %d = %d, want %d", i, got, want3)
		}
	}
	if got, want := Aig_ManRandom(0), r.Uint32(); got != want {
		t.Errorf("shared stream advanced to %d, want %d", got, want)
	}
}

func TestTernaryRandomHelpersAgainstCompiledSource(t *testing.T) {
	out := runAigSourceFixture(t, "ternary_random_compare.c", "aigUtil.c")
	Aig_ManRandom(1)
	defer Aig_ManRandom(1)
	var got strings.Builder
	for i := 0; i < 32; i++ {
		binary, ternary := Aig_XsimRand2(), Aig_XsimRand3()
		fmt.Fprintf(&got, "%d %d\n", binary, ternary)
	}
	fmt.Fprintf(&got, "next %d\n", Aig_ManRandom(0))
	if got.String() != out {
		t.Fatalf("Go stream:\n%sC stream:\n%s", got.String(), out)
	}
}

func TestTernaryMappingAcrossPackedWordsAgainstSource(t *testing.T) {
	out := runAigSourceFixture(t, "ternary_mapping_compare.c", "aigMem.c")
	p := NewGraph()
	input := p.NewInput()
	los := make([]Edge, 18)
	for i := range los {
		los[i] = p.NewInput()
	}
	for i, e := range los {
		switch i % 4 {
		case 0:
			p.AddOutput(e)
		case 1:
			p.AddOutput(e.Not())
		case 2:
			p.AddOutput(input)
		case 3:
			p.AddOutput(p.Const1())
		}
	}
	p.nRegs = 18
	var got strings.Builder
	got.WriteString("mapping:")
	for _, e := range Aig_ManTernarySimulate(p, 0, 0) {
		fmt.Fprintf(&got, " %d", e)
	}
	got.WriteString("\nvalues:")
	for _, n := range p.nodes {
		fmt.Fprintf(&got, " %d", n.Xsim)
	}
	got.WriteByte('\n')
	if got.String() != out {
		t.Fatalf("Go simulation:\n%sC simulation:\n%s", got.String(), out)
	}
}
