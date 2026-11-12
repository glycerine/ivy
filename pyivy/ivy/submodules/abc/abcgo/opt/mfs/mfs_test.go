package mfs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"reflect"
	"testing"
)

func table(t *testing.T, n *Network) [][]bool {
	t.Helper()
	r := make([][]bool, 1<<uint(len(n.Inputs)))
	for a := range r {
		in := make([]bool, len(n.Inputs))
		for i := range in {
			in[i] = a&(1<<uint(i)) != 0
		}
		var err error
		r[a], err = n.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
	}
	return r
}
func equalTable(t *testing.T, n *Network, want [][]bool) {
	t.Helper()
	got := table(t, n)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("output truth changed\ngot%v\nwant%v", got, want)
	}
}
func setup(t *testing.T, n *Network, pars Mfs_Par_t, target int) *Mfs_Man_t {
	t.Helper()
	p, err := initialize(&pars, n)
	if err != nil {
		t.Fatal(err)
	}
	if !Abc_WinNode(p, target) {
		t.Fatal("window skipped")
	}
	return p
}
func TestSourceDefaults(t *testing.T) {
	p := DefaultParams()
	if p.WinTfoLevs != 2 || p.FanoutsMax != 30 || p.DepthMax != 20 || p.WinMax != 300 || p.BTLimit != 5000 || !p.Resub || p.Power || p.Area {
		t.Fatal(p)
	}
}
func TestCounterexampleGuidedOneDivisor(t *testing.T) {
	n := NewNetwork()
	a, b := n.AddInput(), n.AddInput()
	shared := n.AddNode([]int{a, b}, []uint32{8})
	root := n.AddNode([]int{shared, a}, []uint32{8})
	n.AddOutput(Literal(root << 1))
	want := table(t, n)
	stats, err := Optimize(n, nil)
	if err != nil {
		t.Fatal(err)
	}
	equalTable(t, n, want)
	if stats.Resubs != 1 || !n.Nodes[shared].Deleted || len(n.Nodes[root].Fanins) != 2 || stats.Counterexamples == 0 {
		t.Fatalf("%+v nodes%+v", stats, n.Nodes)
	}
}
func TestObservabilityAndReachabilityCare(t *testing.T) {
	for _, lev := range []int{0, 2} {
		n := NewNetwork()
		a, b := n.AddInput(), n.AddInput()
		target := n.AddNode([]int{a, b}, []uint32{8})
		root := n.AddNode([]int{target, a}, []uint32{14})
		n.AddOutput(Literal(root << 1))
		want := table(t, n)
		pars := DefaultParams()
		pars.Resub = false
		pars.WinTfoLevs = lev
		p := setup(t, n, pars, target)
		if !Abc_NtkMfsSolveSat(p, target) {
			t.Fatal("care failed")
		}
		care := uint32(0xffffffff)
		if lev != 0 {
			care = 0x55555555
		}
		if p.Care[0] != care {
			t.Fatalf("levels%d care%x want%x", lev, p.Care[0], care)
		}
		if !Abc_NtkMfsNode(p, target) {
			t.Fatal("node optimization failed")
		}
		equalTable(t, n, want)
		if lev != 0 && gateCount(n.Nodes[target].local) != 0 {
			t.Fatal("ODC did not remove gate")
		}
	}
}
func TestOneHotAndExternalCare(t *testing.T) {
	for _, onehot := range []bool{false, true} {
		n := NewNetwork()
		a, b := n.AddInput(), n.AddInput()
		target := n.AddNode([]int{a, b}, []uint32{8})
		n.AddOutput(Literal(target << 1))
		pars := DefaultParams()
		pars.OneHotness = onehot
		if !onehot {
			care := aig.NewGraph()
			x := care.NewInput()
			care.NewInput()
			care.AddOutput(x.Not())
			n.Care = care
		}
		stats, err := Optimize(n, &pars)
		if err != nil {
			t.Fatal(err)
		}
		if stats.Removes == 0 {
			t.Fatal("care did not allow edge removal")
		}
		if truthBit(n.Nodes[target].Truth, 0) {
			t.Fatal("expected constant-zero completion")
		}
	}
}
func TestExhaustiveRandomOutputPreservation(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	for mode := 0; mode < 3; mode++ {
		for trial := 0; trial < 70; trial++ {
			n := NewNetwork()
			ids := []int{n.AddInput(), n.AddInput(), n.AddInput(), n.AddInput()}
			for k := 0; k < 7; k++ {
				arity := 2 + r.Intn(2)
				fs := make([]int, arity)
				for i := range fs {
					fs[i] = ids[r.Intn(len(ids))]
				}
				id := n.AddNode(fs, []uint32{r.Uint32()})
				ids = append(ids, id)
			}
			for _, id := range ids[len(ids)-2:] {
				n.AddOutput(Literal(id << 1))
			}
			want := table(t, n)
			pars := DefaultParams()
			pars.Resub = mode != 0
			pars.MoreEffort = mode == 2
			_, err := Optimize(n, &pars)
			if err != nil {
				t.Fatalf("mode%d trial%d: %v", mode, trial, err)
			}
			equalTable(t, n, want)
		}
	}
}
func TestGraphAdaptersAndRegisterPartition(t *testing.T) {
	g := aig.NewGraph()
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	g.Name = "sequential"
	g.AddOutput(g.NewXor(a, q))
	g.AddOutput(g.AND(a.Not(), b))
	g.AddOutput(g.Mux(a, b, q))
	aig.Aig_ManSetRegNum(g, 1)
	n, err := FromAIG(g)
	if err != nil {
		t.Fatal(err)
	}
	before := table(t, n)
	_, err = Optimize(n, nil)
	if err != nil {
		t.Fatal(err)
	}
	h, err := n.ToAIG()
	if err != nil {
		t.Fatal(err)
	}
	if aig.Aig_ManRegNum(h) != 1 || h.Name != "sequential" {
		t.Fatal("metadata lost")
	}
	for a, want := range before {
		in := []bool{a&1 != 0, a&2 != 0, a&4 != 0}
		got, err := h.Evaluate(in)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal(a, got, want, err)
		}
	}
}
func TestPairedMiterProjectionAndDryRun(t *testing.T) {
	n := NewNetwork()
	a, b := n.AddInput(), n.AddInput()
	pnode := n.AddNode([]int{a, b}, []uint32{8})
	target := n.AddNode([]int{pnode, a}, []uint32{8})
	n.AddOutput(Literal(target << 1))
	p := setup(t, n, DefaultParams(), target)
	Abc_MfsCreateSolverResub(p, nil, false)
	before := table(t, n)
	if !Abc_NtkMfsSolveSatResub(p, target, 0, false, true) {
		t.Fatal("expected feasible dry run")
	}
	equalTable(t, n, before)
	if p.Stats.Removes != 0 || p.Stats.Resubs != 0 {
		t.Fatal("dry run counted unperformed mutations", p.Stats)
	}
	if n.Nodes[pnode].Deleted {
		t.Fatal("dry run mutated network")
	}
	miter := Abc_NtkMfsConstructGia(p)
	if err := miter.Check(); err != nil {
		t.Fatal(err)
	}
	if miter.InputCount() != 4 || miter.OutputCount() != len(p.Divisors)+1 {
		t.Fatal("miter interface")
	}
}
func TestInvalidNetworkRejectedBeforeMutation(t *testing.T) {
	n := NewNetwork()
	a := n.AddInput()
	id := n.AddNode([]int{a}, []uint32{2})
	n.AddOutput(Literal(id << 1))
	n.Nodes[id].Fanins[0] = id
	if _, err := Optimize(n, nil); err == nil {
		t.Fatal("accepted cycle")
	}
}

func TestProjectionSATMatchesIndependentWindowEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(622))
	queries := 0
	for trial := 0; trial < 40; trial++ {
		n := NewNetwork()
		ids := []int{n.AddInput(), n.AddInput(), n.AddInput(), n.AddInput()}
		for i := 0; i < 4; i++ {
			fs := []int{ids[rng.Intn(len(ids))], ids[rng.Intn(len(ids))]}
			ids = append(ids, n.AddNode(fs, []uint32{rng.Uint32()}))
		}
		target := ids[len(ids)-2]
		root := n.AddNode([]int{target, ids[len(ids)-1]}, []uint32{rng.Uint32()})
		n.AddOutput(Literal(root << 1))
		pars := DefaultParams()
		pars.OneHotness = trial%3 == 0
		p := setup(t, n, pars, target)
		Abc_MfsCreateSolverResub(p, nil, false)
		var rows [][]bool
		for a := 0; a < 1<<uint(p.Window.InputCount()); a++ {
			in := make([]bool, p.Window.InputCount())
			ones := 0
			for i := range in {
				in[i] = a&(1<<uint(i)) != 0
				if in[i] {
					ones++
				}
			}
			if pars.OneHotness && ones > 1 {
				continue
			}
			out, err := p.Window.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			care := out[0]
			for c := 1; c <= p.CareOutputs; c++ {
				care = care && out[c]
			}
			if care {
				rows = append(rows, out)
			}
		}
		width := min(6, len(p.Divisors))
		for mask := 0; mask < 1<<uint(width); mask++ {
			var cs []int
			for i := 0; i < width; i++ {
				if mask&(1<<uint(i)) != 0 {
					cs = append(cs, i)
				}
			}
			values := map[int]bool{}
			feasible := true
			for _, row := range rows {
				key := 0
				for i, c := range cs {
					if row[2+p.CareOutputs+c] {
						key |= 1 << uint(i)
					}
				}
				f := row[1+p.CareOutputs]
				if old, ok := values[key]; ok && old != f {
					feasible = false
				}
				values[key] = f
			}
			status := Abc_NtkMfsTryResubOnce(p, cs)
			queries++
			if (status == 1) != feasible {
				t.Fatalf("trial%d candidates%v got%d feasible%v", trial, cs, status, feasible)
			}
			if feasible {
				truth := Abc_NtkMfsInterplate(p, cs)
				if truth == nil {
					t.Fatal("interpolation unexpectedly timed out")
				}
				for _, row := range rows {
					key := 0
					for i, c := range cs {
						if row[2+p.CareOutputs+c] {
							key |= 1 << uint(i)
						}
					}
					if truthBit(truth, key) != row[1+p.CareOutputs] {
						t.Fatal("interpolant violates care projection")
					}
				}
			}
		}
	}
	if queries < 300 {
		t.Fatal("insufficient projected cases", queries)
	}
}
func TestPowerAndAreaPassesPreserveFunction(t *testing.T) {
	for mode := 0; mode < 4; mode++ {
		n := NewNetwork()
		a, b, c := n.AddInput(), n.AddInput(), n.AddInput()
		u := n.AddNode([]int{a, b}, []uint32{8})
		v := n.AddNode([]int{u, a, c}, []uint32{0xfe})
		n.AddOutput(Literal(v << 1))
		want := table(t, n)
		pars := DefaultParams()
		pars.Power = mode < 2
		pars.Resub = mode != 1
		pars.Area = mode == 2
		pars.RrOnly = mode == 3
		_, err := Optimize(n, &pars)
		if err != nil {
			t.Fatal(err)
		}
		equalTable(t, n, want)
	}
}
func TestWindowIgnoresDanglingFanoutBranches(t *testing.T) {
	n := NewNetwork()
	a, b := n.AddInput(), n.AddInput()
	target := n.AddNode([]int{a, b}, []uint32{8})
	live := n.AddNode([]int{target, a}, []uint32{14})
	n.AddNode([]int{target, b}, []uint32{8})
	n.AddOutput(Literal(live << 1))
	p, err := initialize(nil, n)
	if err != nil {
		t.Fatal(err)
	}
	roots := Abc_MfsComputeRoots(p, target, 2, 30)
	if !reflect.DeepEqual(roots, []int{live}) {
		t.Fatal("dangling branch became false observability root", roots)
	}
}

func TestBdcDoesNotApplyResubWindowSizeLimit(t *testing.T) {
	n := NewNetwork()
	a, b := n.AddInput(), n.AddInput()
	target := n.AddNode([]int{a, b}, []uint32{8})
	root := n.AddNode([]int{target, a}, []uint32{14})
	n.AddOutput(Literal(root << 1))
	pars := DefaultParams()
	pars.Resub = false
	pars.WinMax = 1
	p, err := initialize(&pars, n)
	if err != nil {
		t.Fatal(err)
	}
	if !Abc_NtkMfsNode(p, target) {
		t.Fatal("BDC incorrectly skipped a window above resub cap")
	}
	if gateCount(n.Nodes[target].local) != 0 {
		t.Fatal("missed valid don't-care simplification")
	}
}
func TestSequentialPowerProbabilityUsesReachableState(t *testing.T) {
	n := NewNetwork()
	q := n.AddInput()
	zero := n.AddNode(nil, []uint32{0})
	n.AddOutput(Literal(q << 1))
	n.AddOutput(Literal(zero << 1))
	n.Registers = 1
	ones := estimateProbabilities(n, true)
	if ones[q] != 0 {
		t.Fatalf("unreachable latch one treated as random PI: %v", ones[q])
	}
	n.Nodes[zero].Truth = []uint32{^uint32(0)}
	ones = estimateProbabilities(n, true)
	if ones[q] != 1 {
		t.Fatal("settled latch probability ignored", ones[q])
	}
}

func TestPowerPassContinuesAfterSuccessfulReplacement(t *testing.T) {
	n := NewNetwork()
	fs := make([]int, 5)
	for i := range fs {
		fs[i] = n.AddInput()
	}
	target := n.AddNode(fs, []uint32{0})
	n.AddOutput(Literal(target << 1))
	pars := DefaultParams()
	pars.Power = true
	stats, err := Optimize(n, &pars)
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Nodes[target].Fanins) > 1 || stats.Removes < 4 {
		t.Fatalf("power pass stopped after its first success: fanins%v stats%+v", n.Nodes[target].Fanins, stats)
	}
}

// Source Vec_PtrFillSimInfo fills 0xFF; TryResubOnce clears zero-valued
// projections. The scan accepts only an all-ones word (or pairwise OR),
// including untouched one padding beyond the final witness.
func TestPackedSourceCounterexampleCoverage(t *testing.T) {
	rng := rand.New(rand.NewSource(904))
	for count := 1; count < 100; count++ {
		p := &Mfs_Man_t{}
		words := make([][]uint32, 5)
		for i := range words {
			words[i] = make([]uint32, (count+31)/32)
			for w := range words[i] {
				words[i][w] = ^uint32(0)
			}
		}
		for c := 0; c < count; c++ {
			witness := make([]bool, 5)
			for d := range witness {
				witness[d] = rng.Intn(2) != 0
				if !witness[d] {
					words[d][c/32] &^= 1 << uint(c&31)
				}
			}
			p.witnesses = append(p.witnesses, witness)
		}
		for a := range words {
			single := true
			for _, v := range words[a] {
				single = single && v == ^uint32(0)
			}
			if p.covered(a) != single {
				t.Fatal("single-divisor criterion mismatch")
			}
			for b := 0; b < a; b++ {
				pair := true
				for w, v := range words[a] {
					pair = pair && (v|words[b][w]) == ^uint32(0)
				}
				if p.covered(a, b) != pair {
					t.Fatal("paired criterion mismatch")
				}
			}
		}
	}
}

func TestOptimizeReturnsErrorForInvalidParameters(t *testing.T) {
	n := NewNetwork()
	a := n.AddInput()
	n.AddOutput(Literal(a << 1))
	pars := DefaultParams()
	pars.BTLimit = -1
	if _, err := Optimize(n, &pars); err == nil {
		t.Fatal("negative conflict limit did not return an error")
	}
}

func TestAigAdapterRetainsDanglingLogic(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AND(a, b)
	g.AddOutput(a)
	n, err := FromAIG(g)
	if err != nil {
		t.Fatal(err)
	}
	nodes, _ := counts(n)
	if nodes != 1 {
		t.Fatalf("adapter swept source logic before MFS: nodes%d", nodes)
	}
	h, err := n.ToAIG()
	if err != nil {
		t.Fatal(err)
	}
	if gateCount(h) != 1 {
		t.Fatal("roundtrip omitted dangling gate")
	}
}
