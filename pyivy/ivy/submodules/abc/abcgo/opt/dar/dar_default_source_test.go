package dar_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/proof/dch"
)

func TestDefaultNewChoiceUsesSourcePipeline(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(g.OR(g.AND(a, b), g.AND(a, c)), d))
	options := dar.DefaultChoiceOptions()
	result, err := dar.Dar_ManChoiceNew(g, &options)
	if err != nil {
		t.Fatal(err)
	}
	if got := aig.Aig_ManChoiceNum(result); got != 1 {
		t.Fatalf("source fixture has one choice, got %d", got)
	}
	for mask := 0; mask < 16; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		got, err := result.Evaluate(in)
		want := in[0] && (in[1] || in[2]) && in[3]
		if err != nil || got[0] != want {
			t.Fatalf("input %v: got %v, want %v, err %v", in, got, want, err)
		}
	}
}

func TestDefaultNewChoiceOptionMatrixAgainstABC(t *testing.T) {
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
		name, flag     string
		useGia, useNew bool
	}{
		{"dch", "", false, false}, {"cec", " -g", true, false}, {"cec4", " -x", false, true}, {"cec4_precedes_cec", " -g -x", true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			out, err := exec.Command(abc, "-c", "read_blif "+path+"; strash; &get; &dch"+test.flag+"; &ps").CombinedOutput()
			if err != nil {
				t.Fatalf("C choice: %v\n%s", err, out)
			}
			stat := func(name string) int {
				m := regexp.MustCompile(name + `\s*=\s*(\d+)`).FindSubmatch(out)
				if len(m) != 2 {
					t.Fatalf("missing %s: %s", name, out)
				}
				n, err := strconv.Atoi(string(m[1]))
				if err != nil {
					t.Fatal(err)
				}
				return n
			}
			g := aig.NewGraph()
			a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
			g.AddOutput(g.AND(g.OR(g.AND(a, b), g.AND(a, c)), d))
			params := dch.DefaultParams()
			params.UseGia, params.UseNew = test.useGia, test.useNew
			options := dar.DefaultChoiceOptions()
			options.DchParams = &params
			result, err := dar.Dar_ManChoiceNew(g, &options)
			if err != nil {
				t.Fatal(err)
			}
			if nodes, choices := stat("and"), stat("ch"); nodes != int(aig.Aig_ManNodeNum(result)) || choices != int(aig.Aig_ManChoiceNum(result)) {
				t.Fatalf("C nodes=%d choices=%d; Go nodes=%d choices=%d", nodes, choices, aig.Aig_ManNodeNum(result), aig.Aig_ManChoiceNum(result))
			}
			if g.NodeCount() != 0 || params.SynthDuration <= 0 {
				t.Fatal("input ownership or synthesis timing not preserved")
			}
			for mask := 0; mask < 16; mask++ {
				input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				got, err := result.Evaluate(input)
				want := input[0] && (input[1] || input[2]) && input[3]
				if err != nil || got[0] != want {
					t.Fatalf("input %v got %v want %v err %v", input, got, want, err)
				}
			}
		})
	}
}
