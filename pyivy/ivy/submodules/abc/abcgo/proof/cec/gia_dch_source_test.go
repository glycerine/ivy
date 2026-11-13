package cec

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/dch"
)

func TestGiaPerformDchAgainstABC(t *testing.T) {
	abc, err := filepath.Abs(filepath.Join("..", "..", "..", "abc"))
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(abc); err != nil || info.IsDir() || info.Mode()&0111 == 0 {
		t.Skip("local ABC executable unavailable")
	}
	path := filepath.Join(t.TempDir(), "choice.blif")
	blif := ".model c\n.inputs a b c d\n.outputs y\n.names a b n\n11 1\n.names a c m\n11 1\n.names n m t\n1- 1\n-1 1\n.names t d y\n11 1\n.end\n"
	if err := os.WriteFile(path, []byte(blif), 0600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, flag     string
		useGia, useNew bool
	}{
		{"gia_default", "", false, false}, {"cec", " -g", true, false}, {"cec4", " -x", false, true}, {"cec4_precedes_cec", " -g -x", true, true},
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
			input := gia.Gia_ManFromAig(g)
			var result *gia.Gia_Man_t
			if test.name == "gia_default" {
				result = gia.Gia_ManPerformDch(input, nil)
			} else {
				params := dch.DefaultParams()
				params.UseGia, params.UseNew = test.useGia, test.useNew
				result, err = Gia_ManPerformDch(input, &params)
				if err != nil {
					t.Fatal(err)
				}
			}
			if result == nil {
				t.Fatal("nil result")
			}
			if nodes, choices := stat("and"), stat("ch"); nodes != int(gia.Gia_ManAndNum(result)) || choices != int(gia.Gia_ManChoiceNum(result)) {
				t.Fatalf("C AND/choices=%d/%d; Go=%d/%d", nodes, choices, gia.Gia_ManAndNum(result), gia.Gia_ManChoiceNum(result))
			}
			for mask := 0; mask < 16; mask++ {
				v := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				got, err := result.Evaluate(v)
				want := v[0] && (v[1] || v[2]) && v[3]
				if err != nil || got[0] != want {
					t.Fatalf("input %v result %v want %v err %v", v, got, want, err)
				}
			}
		})
	}
}
