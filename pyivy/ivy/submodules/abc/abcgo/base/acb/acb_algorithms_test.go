package acb

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Evaluate truth tables directly, independently of the GIA and SAT bridges.
func acbEval(p *Acb_Ntk_t, assignment int) []bool {
	values := make(map[int32]bool)
	for i, id := range Vec_IntArray(&p.vCis) {
		values[id] = assignment>>i&1 != 0
	}
	var visit func(int32) bool
	visit = func(id int32) bool {
		if v, ok := values[id]; ok {
			return v
		}
		if Acb_ObjIsCo(p, id) != 0 {
			return visit(Acb_ObjFanin0(p, id))
		}
		index := uint(0)
		for j := int32(0); j < Acb_ObjFaninNum(p, id); j++ {
			if visit(Acb_ObjFanin(p, id, j)) {
				index |= 1 << uint(j)
			}
		}
		v := Acb_ObjTruth(p, id)>>index&1 != 0
		values[id] = v
		return v
	}
	out := make([]bool, Acb_NtkCoNum(p))
	for i, id := range Vec_IntArray(&p.vCos) {
		out[i] = visit(id)
	}
	return out
}
func acbTable(p *Acb_Ntk_t) [][]bool {
	table := make([][]bool, 1<<uint(Acb_NtkCiNum(p)))
	for i := range table {
		table[i] = acbEval(p, i)
	}
	return table
}
func acbNetwork(t *testing.T, inputs int) *Acb_Ntk_t {
	t.Helper()
	m := Acb_ManAlloc("fixture.v", 1, nil, nil, nil, nil)
	p := Acb_NtkAlloc(m, Abc_NamStrFindOrAdd(m.pStrs, "fixture", nil), int32(inputs), 2, 32)
	p.nFaninMax = 6
	for i := 0; i < inputs; i++ {
		Acb_ObjAlloc(p, ABC_OPER_CI, 0, 0)
	}
	Acb_NtkCleanObjTruths(p)
	t.Cleanup(func() { Acb_ManFree(m) })
	return p
}
func acbNode(p *Acb_Ntk_t, truth uint64, fanins ...int32) int32 {
	id := Acb_ObjAlloc(p, ABC_OPER_LUT, int32(len(fanins)), 0)
	for _, f := range fanins {
		Acb_ObjAddFanin(p, id, f)
	}
	Acb_ObjSetTruth(p, id, truth)
	return id
}
func acbOutput(p *Acb_Ntk_t, id int32) {
	co := Acb_ObjAlloc(p, ABC_OPER_CO, 1, 0)
	Acb_ObjAddFanin(p, co, id)
}
func acbCheckLinks(t *testing.T, p *Acb_Ntk_t) {
	t.Helper()
	want := make(map[[2]int32]int)
	for id := int32(1); id < Acb_NtkObjNumMax(p); id++ {
		if Acb_ObjType(p, id) == 0 {
			continue
		}
		for j := int32(0); j < Acb_ObjFaninNum(p, id); j++ {
			want[[2]int32{Acb_ObjFanin(p, id, j), id}]++
		}
	}
	got := make(map[[2]int32]int)
	for id := int32(1); id < Acb_NtkObjNumMax(p); id++ {
		for _, out := range Vec_IntArray(Acb_ObjFanoutVec(p, id)) {
			got[[2]int32{id, out}]++
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fanin/fanout multiplicity mismatch: got %v want %v", got, want)
	}
}
func stretchTruth(bits uint64, n uint) uint64 {
	var out uint64
	for i := uint(0); i < 64; i++ {
		out |= (bits >> (i & ((1 << n) - 1)) & 1) << i
	}
	return out
}

func TestCNFExhaustiveTruthTables(t *testing.T) {
	var cover Vec_Int_t
	var clauses Vec_Str_t
	for n := uint(0); n <= 3; n++ {
		for bits := uint64(0); bits < 1<<(1<<n); bits++ {
			truth := stretchTruth(bits, n)
			count := Acb_DeriveCnfFromTruth(truth, int32(n), &cover, &clauses)
			data := Vec_StrArray(&clauses)
			for a := uint(0); a < 1<<(n+1); a++ {
				sat, clause, seen := true, false, 0
				for _, lit := range data {
					if lit == 255 {
						sat = sat && clause
						clause = false
						seen++
						continue
					}
					v := uint(lit >> 1)
					clause = clause || ((a>>v&1)^uint(lit&1)) != 0
				}
				want := uint64(a>>n&1) == (truth >> (a & ((1 << n) - 1)) & 1)
				if sat != want || seen != int(count) {
					t.Fatalf("CNF n=%d truth=%x assignment=%b: sat=%v want=%v clauses=%v", n, bits, a, sat, want, data)
				}
			}
		}
	}
}

func TestNetworkTraversalTimingAndMutation(t *testing.T) {
	p := acbNetwork(t, 3)
	and := s_Truths6[0] & s_Truths6[1]
	a := acbNode(p, and, 1, 2)
	b := acbNode(p, and, 1, 3)
	root := acbNode(p, s_Truths6[0]^s_Truths6[1], a, b)
	acbOutput(p, root)
	Acb_NtkCreateFanout(p)
	acbCheckLinks(t, p)
	want := []int32{1, 2, a, 3, b, root, Acb_NtkCo(p, 0)}
	if got := Vec_IntArray(Acb_ObjCollectTfi(p, -1, 1)); !reflect.DeepEqual(got, want) {
		t.Fatalf("postorder %v want %v", got, want)
	}
	Acb_NtkComputePaths(p)
	if p.LevelMax != 2 || Acb_ObjPathD(p, root) != 4 || Acb_ObjPathR(p, 1) != 2 {
		t.Fatalf("levels/paths: depth=%d root paths=%d input paths=%d", p.LevelMax, Acb_ObjPathD(p, root), Acb_ObjPathR(p, 1))
	}
	Acb_ObjPatchFanin(p, b, 3, 2)
	acbCheckLinks(t, p)
	for a := 0; a < 8; a++ {
		if acbEval(p, a)[0] {
			t.Fatal("identical fanin cones must XOR to zero")
		}
	}
}

func TestPushCleanupPreservesTruthAndLinks(t *testing.T) {
	for bits := uint64(0); bits < 16; bits++ {
		t.Run(fmt.Sprintf("duplicate_%x", bits), func(t *testing.T) {
			p := acbNetwork(t, 1)
			id := acbNode(p, stretchTruth(bits, 2), 1, 1)
			acbOutput(p, id)
			Acb_NtkCreateFanout(p)
			want := acbTable(p)
			Acb_ObjRemoveDupFanins(p, id)
			if got := acbTable(p); !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v want %v", got, want)
			}
			acbCheckLinks(t, p)
		})
	}
	for _, constant := range []uint64{0, ^uint64(0)} {
		for bits := uint64(0); bits < 16; bits++ {
			t.Run(fmt.Sprintf("constant_%x_%x", constant, bits), func(t *testing.T) {
				p := acbNetwork(t, 1)
				c := acbNode(p, constant)
				id := acbNode(p, stretchTruth(bits, 2), 1, c)
				acbOutput(p, id)
				Acb_NtkCreateFanout(p)
				want := acbTable(p)
				Acb_ObjRemoveConst(p, c)
				if got := acbTable(p); !reflect.DeepEqual(got, want) {
					t.Fatalf("got %v want %v", got, want)
				}
				acbCheckLinks(t, p)
			})
		}
	}
	for _, invert := range []bool{false, true} {
		for bits := uint64(0); bits < 16; bits++ {
			t.Run(fmt.Sprintf("buffer_%v_%x", invert, bits), func(t *testing.T) {
				p := acbNetwork(t, 1)
				truth := s_Truths6[0]
				if invert {
					truth = ^truth
				}
				b := acbNode(p, truth, 1)
				id := acbNode(p, stretchTruth(bits, 2), 1, b)
				acbOutput(p, id)
				Acb_NtkCreateFanout(p)
				want := acbTable(p)
				Acb_ObjRemoveBufInv(p, b)
				if got := acbTable(p); !reflect.DeepEqual(got, want) {
					t.Fatalf("got %v want %v", got, want)
				}
				acbCheckLinks(t, p)
			})
		}
	}
}

func TestMfsOptimizationTruthPreservation(t *testing.T) {
	for _, area := range []int32{0, 1} {
		for seed := int64(0); seed < 12; seed++ {
			t.Run(fmt.Sprintf("area_%d_seed_%d", area, seed), func(t *testing.T) {
				p := acbNetwork(t, 4)
				r := rand.New(rand.NewSource(seed))
				for j := 0; j < 6; j++ {
					a := Acb_NtkObjNum(p)
					b := a
					for b == a {
						b = int32(1 + r.Intn(int(Acb_NtkObjNum(p))))
					}
					acbNode(p, stretchTruth(uint64(r.Intn(16)), 2), a, b)
				}
				acbOutput(p, Acb_NtkObjNum(p))
				want := acbTable(p)
				var pars Acb_Par_t
				Acb_ParSetDefault(&pars)
				pars.fArea = area
				Acb_NtkOpt(p, &pars)
				if got := acbTable(p); !reflect.DeepEqual(got, want) {
					t.Fatalf("optimized truth %v want %v", got, want)
				}
				acbCheckLinks(t, p)
			})
		}
	}
}

func TestVerilogRoundTrip(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.v")
	output := filepath.Join(dir, "output.v")
	source := "module fixture(a,b,y);\ninput a,b;\noutput y;\nwire n;\nand (n,a,b);\nnot (y,n);\nendmodule\n"
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	p := Acb_VerilogSimpleRead(input, "")
	if p == nil {
		t.Fatal("read failed")
	}
	defer Acb_ManFree(p.pDesign)
	g := Acb_NtkToGia2(p, 0, 0, nil, 0)
	if g == nil || Gia_ManCiNum(g) != 2 || Gia_ManCoNum(g) != 1 {
		t.Fatal("GIA conversion failed")
	}
	Gia_ManStop(g)
	Acb_VerilogSimpleWrite(p, output)
	q := Acb_VerilogSimpleRead(output, "")
	if q == nil {
		t.Fatal("read output failed")
	}
	defer Acb_ManFree(q.pDesign)
	if Acb_NtkCiNum(q) != 2 || Acb_NtkCoNum(q) != 1 {
		t.Fatal("terminal counts changed")
	}
	for i := int32(0); i < 2; i++ {
		if Acb_ObjNameStr(q, Acb_NtkCi(q, i)) != Acb_ObjNameStr(p, Acb_NtkCi(p, i)) {
			t.Fatal("input names changed")
		}
	}
}

func TestPushDecompositionExhaustive(t *testing.T) {
	x, y := s_Truths6[0], s_Truths6[1]
	functions := []uint64{x & y, ^x & y, ^x | y, x | y, x ^ y}
	for k, f := range functions {
		for _, shared := range []bool{false, true} {
			t.Run(fmt.Sprintf("fanout_%d_shared_%v", k, shared), func(t *testing.T) {
				p := acbNetwork(t, 3)
				n := acbNode(p, f, 1, 2)
				other := int32(3)
				if shared {
					other = 1
				}
				root := acbNode(p, x^y, n, other)
				acbOutput(p, root)
				Acb_NtkCreateFanout(p)
				want := acbTable(p)
				Acb_ObjPushToFanout(p, n, 0, root)
				if got := acbTable(p); !reflect.DeepEqual(got, want) {
					t.Fatalf("got %v want %v", got, want)
				}
				acbCheckLinks(t, p)
			})
		}
	}
	for k, f := range []uint64{x & y, x &^ y, ^x & y, ^x &^ y, x ^ y} {
		for _, shared := range []bool{false, true} {
			t.Run(fmt.Sprintf("fanin_%d_shared_%v", k, shared), func(t *testing.T) {
				p := acbNetwork(t, 3)
				n := acbNode(p, x|y, 1, 2)
				other := int32(3)
				if shared {
					other = 1
				}
				root := acbNode(p, f, n, other)
				acbOutput(p, root)
				Acb_NtkCreateFanout(p)
				want := acbTable(p)
				Acb_ObjPushToFanin(p, root, 1, n)
				if got := acbTable(p); !reflect.DeepEqual(got, want) {
					t.Fatalf("got %v want %v", got, want)
				}
				acbCheckLinks(t, p)
			})
		}
	}
}

func TestMfsSmallAndConstantNetworks(t *testing.T) {
	for inputs := 0; inputs <= 2; inputs++ {
		for _, truth := range []uint64{0, ^uint64(0), s_Truths6[0]} {
			if inputs == 0 && truth == s_Truths6[0] {
				continue
			}
			for _, area := range []int32{0, 1} {
				t.Run(fmt.Sprintf("inputs_%d_truth_%x_area_%d", inputs, truth, area), func(t *testing.T) {
					p := acbNetwork(t, inputs)
					fanins := append([]int32(nil), Vec_IntArray(&p.vCis)...)
					id := acbNode(p, truth, fanins...)
					acbOutput(p, id)
					want := acbTable(p)
					var pars Acb_Par_t
					Acb_ParSetDefault(&pars)
					pars.fArea = area
					Acb_NtkOpt(p, &pars)
					if got := acbTable(p); !reflect.DeepEqual(got, want) {
						t.Fatalf("got %v want %v", got, want)
					}
					acbCheckLinks(t, p)
				})
			}
		}
	}
}

func TestEcoSupportAndPatchFunction(t *testing.T) {
	dir := t.TempDir()
	read := func(name, body string) *Acb_Ntk_t {
		path := filepath.Join(dir, name+".v")
		source := "module fixture(a,b,y);\ninput a,b;\noutput y;\n" + body + "\nendmodule\n"
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		p := Acb_VerilogSimpleRead(path, "")
		if p == nil {
			t.Fatal("read failed")
		}
		t.Cleanup(func() { Acb_ManFree(p.pDesign) })
		return p
	}
	p := read("bad", "wire t_fix;\nbuf(y,t_fix);")
	q := read("good", "wire n;\nand(n,a,b);\nbuf(y,n);")
	var block *Vec_Bit_t
	roots := Acb_NtkFindRoots(p, &p.vTargets, &block)
	if got := Vec_IntArray(roots); !reflect.DeepEqual(got, []int32{0}) {
		t.Fatalf("roots=%v", got)
	}
	supp := Vec_IntTwoMerge(Acb_NtkFindSupp(p, roots), Acb_NtkFindSupp(q, roots))
	if got := Vec_IntArray(supp); !reflect.DeepEqual(got, []int32{0, 1}) {
		t.Fatalf("support=%v", got)
	}
	divs := Acb_NtkFindDivsCis(p, supp)
	f := Acb_NtkToGia(p, supp, Acb_NtkFindNodes(p, roots, divs), roots, divs, &p.vTargets)
	g := Acb_NtkToGia(q, supp, Acb_NtkFindNodes(q, roots, nil), roots, nil, nil)
	defer Gia_ManStop(f)
	defer Gia_ManStop(g)
	m := Acb_CreateMiter(f, g)
	defer Gia_ManStop(m)
	cnf := Acb_NtkDeriveMiterCnf(m, 0, 1, 0)
	defer Cnf_DataFree(cnf)
	used := Acb_DerivePatchSupport(cnf, 0, 1, 2, divs, p, nil, 5)
	if used == nil || !reflect.DeepEqual(Vec_IntArray(used), []int32{0, 1}) {
		t.Fatalf("patch support=%v", used)
	}
	sop := Acb_DeriveOnePatchFunction(cnf, 0, 1, 2, used, 0)
	if got := Abc_SopToTruth(sop, 2); got != s_Truths6[0]&s_Truths6[1] {
		t.Fatalf("patch SOP %q has truth %x, want AND", sop, got)
	}
}

func TestMfsAshenhurstAndPushDrivers(t *testing.T) {
	for _, mode := range []string{"ashen_area", "ashen_delay", "push"} {
		for seed := int64(0); seed < 10; seed++ {
			t.Run(fmt.Sprintf("%s_%d", mode, seed), func(t *testing.T) {
				p := acbNetwork(t, 6)
				r := rand.New(rand.NewSource(seed))
				functions := []uint64{8, 14, 6, 7, 9}
				for j := 0; j < 12; j++ {
					a := Acb_NtkObjNum(p)
					b := int32(1 + r.Intn(int(a-1)))
					acbNode(p, stretchTruth(functions[r.Intn(len(functions))], 2), a, b)
				}
				acbOutput(p, Acb_NtkObjNum(p))
				want := acbTable(p)
				if mode == "push" {
					Acb_NtkPushLogic(p, 3, 0)
				} else {
					var pars Acb_Par_t
					Acb_ParSetDefault(&pars)
					pars.nLutSize = 3
					pars.fUseAshen = 1
					pars.fMoreEffort = 1
					if mode == "ashen_delay" {
						pars.fArea = 0
					}
					Acb_NtkOpt(p, &pars)
				}
				if got := acbTable(p); !reflect.DeepEqual(got, want) {
					t.Fatalf("%s changed function", mode)
				}
				acbCheckLinks(t, p)
			})
		}
	}
}

// This is test_vectors/ashen_unreachable.blif, also rejected by native C CEC
// after the original C mfse -d rewrite. All six input bits are enumerated.
func TestAshenhurstRejectsUnreachableDivisorRewrite(t *testing.T) {
	p := acbNetwork(t, 6)
	n7 := acbNode(p, stretchTruth(14, 2), 6, 5)
	n8 := acbNode(p, stretchTruth(8, 2), n7, 2)
	n11 := acbNode(p, stretchTruth(0xbd, 3), n8, 6, 5)
	n14 := acbNode(p, stretchTruth(0x4c, 3), n11, n7, 1)
	n15 := acbNode(p, stretchTruth(14, 2), n14, 4)
	n17 := acbNode(p, stretchTruth(0x6c, 3), n15, n11, 1)
	pivot := acbNode(p, stretchTruth(9, 2), n17, n14)
	acbOutput(p, pivot)
	want := acbTable(p)
	var pars Acb_Par_t
	Acb_ParSetDefault(&pars)
	pars.nLutSize = 3
	pars.fUseAshen = 1
	m := Acb_MfsStart(p, &pars)
	defer Acb_MfsStop(m)
	Acb_NtkCreateFanout(p)
	Acb_NtkCleanObjFuncs(p)
	Acb_NtkCleanObjCnfs(p)
	Acb_NtkUpdateLevelD(p, -1)
	if changed := Acb_NtkOptNode(m, pivot); changed != 0 {
		t.Fatalf("invalid candidate accepted: %d", changed)
	}
	if m.nTwoNodes != 1 {
		t.Fatal("did not exercise decomposition")
	}
	if !reflect.DeepEqual(acbTable(p), want) {
		t.Fatal("rejected candidate changed network")
	}
	acbCheckLinks(t, p)
}

func TestOptimizationReconvergentMultiOutput(t *testing.T) {
	for mode := 0; mode < 4; mode++ {
		for seed := int64(0); seed < 40; seed++ {
			t.Run(fmt.Sprintf("mode%d_seed%d", mode, seed), func(t *testing.T) {
				p := acbNetwork(t, 6)
				r := rand.New(rand.NewSource(seed))
				functions := []uint64{8, 14, 6, 7, 9, 0, 15, 10, 12}
				for j := 0; j < 15; j++ {
					prev := Acb_NtkObjNum(p)
					other := int32(1 + r.Intn(int(prev)))
					acbNode(p, stretchTruth(functions[r.Intn(len(functions))], 2), prev, other)
				}
				last := Acb_NtkObjNum(p)
				acbOutput(p, last)
				acbOutput(p, last-4)
				want := acbTable(p)
				var pars Acb_Par_t
				Acb_ParSetDefault(&pars)
				pars.nLutSize = 3
				pars.fArea = int32(mode & 1)
				pars.fUseAshen = int32(mode >> 1)
				pars.fMoreEffort = 1
				Acb_NtkOpt(p, &pars)
				if got := acbTable(p); !reflect.DeepEqual(got, want) {
					t.Fatal("optimization changed a primary output")
				}
				acbCheckLinks(t, p)
			})
		}
	}
}
