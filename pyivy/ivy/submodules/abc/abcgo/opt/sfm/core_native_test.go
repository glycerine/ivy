package sfm

import (
	"math/rand"
	"reflect"
	"testing"
)

func testNetwork(t *testing.T, fs [][]int, pis, pos int, ts []uint64) *Sfm_Ntk_t {
	t.Helper()
	n, err := Sfm_NtkConstruct(fs, pis, pos, nil, nil, ts, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return n
}
func coreTable(t *testing.T, p *Sfm_Ntk_t) [][]bool {
	t.Helper()
	r := make([][]bool, 1<<uint(p.Pis))
	for a := range r {
		in := make([]bool, p.Pis)
		for i := range in {
			in[i] = a&(1<<uint(i)) != 0
		}
		var err error
		r[a], err = p.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
	}
	return r
}
func TestCoreOneDivisorAndCleanup(t *testing.T) {
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	want := coreTable(t, p)
	changed := Sfm_NtkPerform(p, nil)
	if changed < 1 || p.Stats.Resubs < 1 || !p.Deleted[2] {
		t.Fatal(changed, p.Stats, p.Deleted)
	}
	if !reflect.DeepEqual(coreTable(t, p), want) {
		t.Fatal("resubstitution changed output")
	}
}
func TestCoreFixedAndEmptyObjects(t *testing.T) {
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	p.Fixed[2] = true
	p.Empty[2] = true
	if err := p.refresh(); err != nil {
		t.Fatal(err)
	}
	if p.Levels[2] != p.Levels[0] {
		t.Fatal("transparent object added level")
	}
	Sfm_NtkPerform(p, nil)
	if p.Deleted[2] || !reflect.DeepEqual(p.Fanins[2], []int{0, 1}) {
		t.Fatal("fixed object removed")
	}
}
func TestCoreAllThreeInputFunctions(t *testing.T) {
	for f := uint64(0); f < 256; f++ {
		p := testNetwork(t, [][]int{nil, nil, nil, {0, 1, 2}, {3}}, 3, 1, []uint64{0, 0, 0, f, 0})
		want := coreTable(t, p)
		Sfm_NtkPerform(p, nil)
		if !reflect.DeepEqual(coreTable(t, p), want) {
			t.Fatalf("function%x changed", f)
		}
	}
}
func TestCoreRandomTruthPreservation(t *testing.T) {
	rng := rand.New(rand.NewSource(714))
	for trial := 0; trial < 90; trial++ {
		fs := [][]int{nil, nil, nil, nil}
		ts := []uint64{0, 0, 0, 0}
		for k := 0; k < 7; k++ {
			row := []int{rng.Intn(len(fs)), rng.Intn(len(fs))}
			fs = append(fs, row)
			ts = append(ts, rng.Uint64())
		}
		fs = append(fs, []int{len(fs) - 1})
		ts = append(ts, 0)
		p := testNetwork(t, fs, 4, 1, ts)
		want := coreTable(t, p)
		pars := DefaultParams()
		pars.UseDcs = trial%2 != 0
		Sfm_NtkPerform(p, &pars)
		if !reflect.DeepEqual(coreTable(t, p), want) {
			t.Fatalf("trial%d output changed", trial)
		}
	}
}
func TestCNFMatchesEveryTruthAssignment(t *testing.T) {
	rng := rand.New(rand.NewSource(53))
	for vars := 0; vars <= 7; vars++ {
		for trial := 0; trial < 20; trial++ {
			words := make([]uint64, truthWords(vars))
			for i := range words {
				words[i] = rng.Uint64()
			}
			if vars < 6 {
				words[0] = stretch(words[0], vars)
			}
			clauses := Sfm_TruthToCnf(words[0], words, vars)
			for a := 0; a < 1<<uint(vars+1); a++ {
				satisfied := true
				for _, c := range clauses {
					ok := false
					for _, l := range c {
						value := a&(1<<uint(l>>1)) != 0
						ok = ok || value != (l&1 != 0)
					}
					satisfied = satisfied && ok
				}
				wanted := ((a>>uint(vars))&1 != 0) == bit(words, a&((1<<uint(vars))-1))
				if satisfied != wanted {
					t.Fatalf("CNF vars%d minterm%d", vars, a)
				}
			}
		}
	}
}
