package bmc

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"os"
	"testing"
)

func TestZyxSynthesis(t *testing.T) {
	for _, tc := range []struct {
		name string
		p    Bmc_EsPar_t
		want bsat.Status
	}{
		{"and", Bmc_EsPar_t{nVars: 2, nNodes: 1, nLutSize: 2, pTtStr: "8"}, bsat.Satisfiable},
		{"xor", Bmc_EsPar_t{nVars: 2, nNodes: 1, nLutSize: 2, pTtStr: "6"}, bsat.Satisfiable},
		{"xor only and", Bmc_EsPar_t{nVars: 2, nNodes: 1, nLutSize: 2, pTtStr: "6", fOnlyAnd: 1}, bsat.Unsatisfiable},
		{"majority", Bmc_EsPar_t{nVars: 3, nNodes: 1, nLutSize: 3, fMajority: 1}, bsat.Satisfiable},
		{"and3 impossible", Bmc_EsPar_t{nVars: 3, nNodes: 1, nLutSize: 2, pTtStr: "80"}, bsat.Unsatisfiable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := newZyxSynthesis(&tc.p)
			if err != nil {
				t.Fatal(err)
			}
			net, status := p.solve()
			if status != tc.want {
				t.Fatalf("status %v want %v", status, tc.want)
			}
			if net != nil && p.mismatch(*net) != -1 {
				t.Fatal("mismatched solution")
			}
		})
	}
}

func TestZyxSolutionFile(t *testing.T) {
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	p, err := newZyxSynthesis(&Bmc_EsPar_t{nVars: 2, nNodes: 1, nLutSize: 2, pTtStr: "7"})
	if err != nil {
		t.Fatal(err)
	}
	net, status := p.solve()
	if status != bsat.Satisfiable {
		t.Fatal(status)
	}
	if err = p.writeSolution(net, true); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile("7-2-1.bool")
	if err != nil || string(contents) != "C0111ab\n\n" {
		t.Fatalf("solution=%q err=%v", contents, err)
	}
}

func TestZyxNodeOrdering(t *testing.T) {
	pars := Bmc_EsPar_t{nVars: 3, nNodes: 2, nLutSize: 2, pTtStr: "80", fOrderNodes: 1}
	p, err := newZyxSynthesis(&pars)
	if err != nil {
		t.Fatal(err)
	}
	net, status := p.solve()
	if status != bsat.Satisfiable || !p.ordered(*net) || p.mismatch(*net) != -1 {
		t.Fatalf("ordered synthesis: %v %v", status, net)
	}
	bad := zyxNetwork{fanins: [][]int{{1, 2}, {0, 1}}, funcs: [][]bool{{false, false, false, true}, {false, false, false, true}}}
	if p.ordered(bad) {
		t.Fatal("accepted reversed fanin order")
	}
}
