package gia

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
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

func TestMiniAigGenerateMatchesOriginalFiles(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mini_generate.c")
	for name, text := range map[string]string{
		"independent outputs":    "#  2=0+1\n#  3=0+0\n",
		"shared unordered cones": "#  5=2+3\n#  4=0+1\n#  2=4+0\n#  6=4+1\n",
		"repeated fanins":        "#  2=0+1\n#  3=2+2\n#  4=0+1\n",
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(exe)
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("original C generation: %v: %s", err, out)
			}
			want, err := os.ReadFile(filepath.Join(dir, "test.miniaig"))
			if err != nil {
				t.Fatal(err)
			}
			triples, _, _, err := Gia_MiniAigReadTriples(strings.NewReader(text))
			if err != nil {
				t.Fatal(err)
			}
			p, err := Gia_MiniAigGenerate(triples)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "go.miniaig")
			miniaig.Mini_AigDump(p, path)
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("MiniAIG file differs:\nGo %x\nC  %x", got, want)
			}
		})
	}
}

func TestMiniAigGeneratePlacesOutputsAfterLogic(t *testing.T) {
	triples, _, _, err := Gia_MiniAigReadTriples(strings.NewReader("#  2=0+1\n#  3=0+0\n"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := Gia_MiniAigGenerate(triples)
	if err != nil {
		t.Fatal(err)
	}
	seenOutput := false
	for id := 1; id < miniaig.Mini_AigNodeNum(p); id++ {
		if miniaig.Mini_AigNodeIsPo(p, id) {
			seenOutput = true
		} else if seenOutput {
			t.Fatalf("logic/input node %d follows a primary output; C emits all logic before outputs", id)
		}
	}
	if miniaig.Mini_AigPoNum(p) != 2 {
		t.Fatal("missing output")
	}
}
