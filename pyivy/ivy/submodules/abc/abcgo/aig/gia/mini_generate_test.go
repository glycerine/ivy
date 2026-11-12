package gia

import (
	"strings"
	"testing"
)

func TestMiniAigGenerateTriples(t *testing.T) {
	triples, selected, lines, err := Gia_MiniAigReadTriples(strings.NewReader("ignored\n#  2=0+1\n#  3=2+1\n"))
	if err != nil || selected != 2 || lines != 3 {
		t.Fatal(selected, lines, err)
	}
	p, err := Gia_MiniAigGenerate(triples)
	if err != nil {
		t.Fatal(err)
	}
	g := Gia_ManFromMiniAig(p, nil)
	for i := 0; i < 4; i++ {
		out, err := g.Evaluate([]bool{i&1 != 0, i&2 != 0})
		if err != nil || len(out) != 1 || out[0] != (i == 3) {
			t.Fatalf("input %d: %v %v", i, out, err)
		}
	}
	cycle, _, _, err := Gia_MiniAigReadTriples(strings.NewReader("#  0=1+2\n#  1=0+2\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Gia_MiniAigGenerate(cycle); err == nil {
		t.Fatal("cycle accepted")
	}
}
