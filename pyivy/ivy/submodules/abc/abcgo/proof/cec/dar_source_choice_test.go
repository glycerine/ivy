package cec

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
	"github.com/glycerine/abcgo/proof/dch"
)

func TestDarSourceChoiceEngineAgainstABC(t *testing.T) {
	abc, err := filepath.Abs(filepath.Join("..", "..", "..", "abc"))
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(abc); err != nil || info.IsDir() || info.Mode()&0111 == 0 {
		t.Skip("local ABC executable unavailable")
	}
	blif := ".model c\n.inputs a b c d\n.outputs y\n.names a b n\n11 1\n.names a c m\n11 1\n.names n m t\n1- 1\n-1 1\n.names t d y\n11 1\n.end\n"
	path := filepath.Join(t.TempDir(), "choice.blif")
	if err := os.WriteFile(path, []byte(blif), 0600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, flag string
		gia, newer bool
	}{
		{"dch", "", false, false}, {"gia", " -g", true, false}, {"cec4", " -x", false, true}, {"cec4_precedes_gia", " -g -x", true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			out, err := exec.Command(abc, "-c", "read_blif "+path+"; strash; &get; &dch"+test.flag+"; &ps").CombinedOutput()
			if err != nil {
				t.Fatalf("C choice path: %v\n%s", err, out)
			}
			read := func(name string) int {
				match := regexp.MustCompile(name + `\s*=\s*(\d+)`).FindSubmatch(out)
				if len(match) != 2 {
					t.Fatalf("missing %s: %s", name, out)
				}
				n, err := strconv.Atoi(string(match[1]))
				if err != nil {
					t.Fatal(err)
				}
				return n
			}
			g := aig.NewGraph()
			a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
			g.AddOutput(g.AND(g.OR(g.AND(a, b), g.AND(a, c)), d))
			params := dch.DefaultParams()
			params.UseGia, params.UseNew = test.gia, test.newer
			options := dar.DefaultChoiceOptions()
			result, err := Dar_ManChoiceNew(g, &options, &params)
			if err != nil {
				t.Fatal(err)
			}
			if options.SynthesisDuration <= 0 || params.SynthDuration != options.SynthesisDuration {
				t.Fatal("source synthesis timing was not returned")
			}
			if nodes, choices := read("and"), read("ch"); nodes != int(aig.Aig_ManNodeNum(result)) || choices != int(aig.Aig_ManChoiceNum(result)) {
				t.Fatalf("C nodes=%d choices=%d; Go nodes=%d choices=%d", nodes, choices, aig.Aig_ManNodeNum(result), aig.Aig_ManChoiceNum(result))
			}
			for mask := 0; mask < 16; mask++ {
				input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				got, err := result.Evaluate(input)
				want := input[0] && (input[1] || input[2]) && input[3]
				if err != nil || want != got[0] {
					t.Fatalf("input %v got %v want %v err %v", input, got, want, err)
				}
			}
		})
	}
}

func TestDarSourceChoiceEngineRejectsInvalidSnapshots(t *testing.T) {
	engine := DarSourceChoiceEngine{}
	if _, err := engine.ComputeChoices(nil); err == nil {
		t.Fatal("accepted missing snapshots")
	}
	g := aig.NewGraph()
	g.AddOutput(g.NewInput())
	if _, err := engine.ComputeChoices([]*aig.Graph{g, nil, g}); err == nil {
		t.Fatal("accepted nil snapshot")
	}
	incompatible := aig.NewGraph()
	incompatible.AddOutput(incompatible.NewInput())
	incompatible.AddOutput(incompatible.NewInput())
	if _, err := engine.ComputeChoices([]*aig.Graph{g, incompatible, g}); err == nil {
		t.Fatal("accepted mismatched interfaces")
	}
}

func TestDarSourceChoiceEnginePreservesSnapshots(t *testing.T) {
	for _, mode := range []struct{ gia, newer bool }{{}, {gia: true}, {newer: true}} {
		versions := make([]*aig.Graph, 3)
		for i := range versions {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			g.AddOutput(g.AND(a, b))
			versions[i] = g
		}
		params := dch.DefaultParams()
		params.UseGia, params.UseNew = mode.gia, mode.newer
		result, err := (DarSourceChoiceEngine{Params: &params}).ComputeChoices(versions)
		if err != nil || result == nil {
			t.Fatalf("mode=%+v: result=%v error=%v", mode, result, err)
		}
		for i, graph := range versions {
			if err := graph.Check(); err != nil {
				t.Fatalf("mode=%+v snapshot %d: %v", mode, i, err)
			}
			got, err := graph.Evaluate([]bool{true, true})
			if err != nil || !got[0] {
				t.Fatalf("mode=%+v snapshot %d: %v, %v", mode, i, got, err)
			}
		}
	}
}
