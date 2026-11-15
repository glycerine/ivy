package gia

import (
	"fmt"
	"reflect"
	"slices"
	"testing"
)

func TestTasRejectsTypedAndChangedGraphs(t *testing.T) {
	for _, kind := range []string{"xor-constructor", "mux-constructor", "xor-query", "grown-query"} {
		t.Run(kind, func(t *testing.T) {
			g := Gia_ManStart(16)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			Gia_ManFillValue(g)
			if kind == "xor-constructor" || kind == "mux-constructor" {
				if kind == "xor-constructor" {
					Gia_ManAppendXorReal(g, a, b)
				} else {
					g.EnableMuxes()
					Gia_ManAppendMuxReal(g, a, b, c)
				}
				cexReviewMustPanic(t, func() { Tas_ManAlloc(g, 100) })
				return
			}
			p := Tas_ManAlloc(g, 100)
			defer Tas_ManStop(p)
			if kind == "xor-query" {
				Gia_ManAppendXorReal(g, a, b)
			} else {
				Gia_ManAppendCi(g)
			}
			cexReviewMustPanic(t, func() { Tas_ManSolve(p, Gia_ObjFromLit(g, a), nil) })
			if p.pProp.iTail != 0 {
				t.Fatal("invalid graph query assigned root")
			}
		})
	}
}

func TestTasRejectedAssumptionsPreserveState(t *testing.T) {
	for _, kind := range []string{"output", "foreign", "nil"} {
		t.Run(kind, func(t *testing.T) {
			g := Gia_ManStart(16)
			a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			root := Gia_ManAppendAnd(g, a, b)
			co := Gia_ManAppendCo(g, root)
			Gia_ManFillValue(g)
			p := Tas_ManAlloc(g, 100)
			defer Tas_ManStop(p)
			bad := Gia_ObjFromLit(g, co)
			if kind == "nil" {
				bad = nil
			}
			if kind == "foreign" {
				other := Gia_ManStart(8)
				bad = Gia_ObjFromLit(other, Gia_ManAppendCi(other))
			}
			Vec_IntPush(p.vModel, 1234)
			cexReviewMustPanic(t, func() { Tas_ManSolveArray(p, []*Gia_Obj_t{Gia_ObjFromLit(g, root), bad}) })
			if !slices.Equal(Vec_IntArray(p.vModel), []int32{1234}) || p.pProp.iTail != 0 {
				t.Fatal("rejected assumptions changed solver")
			}
			for _, obj := range g.pObjs {
				if obj.fMark0 != 0 || obj.Value != ^uint32(0) {
					t.Fatal("rejected assumptions changed graph")
				}
			}
		})
	}
}

func TestTasStopCancelsActiveAssignments(t *testing.T) {
	g := Gia_ManStart(16)
	a := Gia_ManAppendCi(g)
	Gia_ManFillValue(g)
	p := Tas_ManAlloc(g, 100)
	Tas_ManAssign(p, Gia_ObjFromLit(g, a), 0, nil, nil)
	Tas_ManStop(p)
	if obj := g.pObjs[a>>1]; obj.fMark0 != 0 || obj.fMark1 != 0 || obj.Value != ^uint32(0) {
		t.Fatal("active Stop retained graph assignment")
	}
	Tas_ManStop(p)
	Tas_ManStop(nil)
}

func BenchmarkTasConflictScratch(b *testing.B) {
	for _, pair := range []bool{false, true} {
		b.Run(fmt.Sprintf("pair%v", pair), func(b *testing.B) {
			g, lits := cbs0ReviewGraph(8)
			Gia_ManFillValue(g)
			Gia_ManCleanPhase(g)
			p := Tas_ManAlloc(g, 1000)
			defer Tas_ManStop(p)
			var second *Gia_Obj_t
			if pair {
				second = Gia_ObjFromLit(g, lits[0])
			}
			root := Gia_ObjFromLit(g, lits[19])
			if Tas_ManSolve(p, root, second) != 1 {
				b.Fatal("hard fixture is not UNSAT")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if Tas_ManSolve(p, root, second) != 1 {
					b.Fatal("conflict verdict changed")
				}
			}
		})
	}
}

func TestTasRawGateValues(t *testing.T) {
	for _, kind := range []string{"buffer", "repeated", "opposite", "zero", "one", "constant-zero", "constant-one"} {
		for mode := int32(0); mode < 4; mode++ {
			for phase := int32(0); phase < 2; phase++ {
				t.Run(fmt.Sprintf("%s/mode%d/phase%d", kind, mode, phase), func(t *testing.T) {
					g := Gia_ManStart(16)
					g.fGiaSimple = 1
					a := Gia_ManAppendCi(g)
					var root int32
					switch kind {
					case "buffer":
						root = Gia_ManAppendBuf(g, a)
					case "repeated":
						root = Gia_ManAppendAnd(g, a, a)
					case "opposite":
						root = Gia_ManAppendAnd(g, a, a^1)
					case "zero":
						root = Gia_ManAppendAnd(g, 0, a)
					case "one":
						root = Gia_ManAppendAnd(g, 1, a)
					case "constant-zero":
						root = Gia_ManAppendAnd(g, 0, 1)
					case "constant-one":
						root = Gia_ManAppendAnd(g, 1, 1)
					}
					Gia_ManFillValue(g)
					Gia_ManCleanPhase(g)
					p := Tas_ManAlloc(g, 1000)
					defer Tas_ManStop(p)
					p.Pars.fUseHighest, p.Pars.fUseLowest, p.Pars.fUseMaxFF, p.Pars.fUseActive = giaBool(mode == 0), giaBool(mode == 1), giaBool(mode == 2), giaBool(mode == 3)
					var truth uint8
					for input := 0; input < 2; input++ {
						values, _, err := g.evaluateObjects([]bool{input != 0})
						if err != nil {
							t.Fatal(err)
						}
						if values[root>>1] != (phase != 0) {
							truth |= 1 << input
						}
					}
					defer func() {
						if recovered := recover(); recovered != nil {
							t.Fatalf("raw gate query panicked: %v", recovered)
						}
					}()
					if status := Tas_ManSolve(p, Gia_ObjFromLit(g, root^phase), nil); status != giaBool(truth == 0) {
						t.Fatalf("raw gate verdict%d truth%x", status, truth)
					}
					if truth != 0 {
						completions := uint8(3)
						for _, lit := range Vec_IntArray(p.vModel) {
							if lit>>1 != 0 {
								t.Fatal("invalid CI witness")
							}
							if lit&1 == 0 {
								completions &= 2
							} else {
								completions &= 1
							}
						}
						if completions == 0 || completions & ^truth != 0 {
							t.Fatal("invalid raw-gate witness")
						}
					}
					for _, obj := range g.pObjs {
						if obj.fMark0 != 0 || obj.fMark1 != 0 || obj.fPhase != 0 || obj.Value != ^uint32(0) {
							t.Fatal("raw query retained scratch")
						}
					}
				})
			}
		}
	}
}

func TestTasForeignClauseRejectedBeforeMutation(t *testing.T) {
	for _, createConflict := range []bool{false, true} {
		t.Run(fmt.Sprintf("conflict%v", createConflict), func(t *testing.T) {
			g := Gia_ManStart(16)
			a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			Gia_ManFillValue(g)
			p, q := Tas_ManAlloc(g, 100), Tas_ManAlloc(g, 100)
			defer Tas_ManStop(p)
			defer Tas_ManStop(q)
			clause := Tas_ManAllocCls(q, 5)
			copy(clause.literals(), []int32{a, b})
			clause.setNext(0, 1234)
			before := slices.Clone(q.pStore.words)
			cexReviewMustPanic(t, func() {
				if createConflict {
					Tas_ManCreateFromCls(p, clause, 0)
				} else {
					Tas_ManWatchClause(p, clause, a)
				}
			})
			if !slices.Equal(q.pStore.words, before) || Vec_IntSize(p.vWatchLits) != 0 || p.pClauses.iHead != 1 || p.pClauses.iTail != 1 {
				t.Fatal("foreign clause rejection changed state")
			}
		})
	}
}

func TestTasRawAndExtensions(t *testing.T) {
	queries := 0
	for seed := 0; seed < 128; seed++ {
		g := Gia_ManStart(64)
		g.fGiaSimple = 1
		lits := []int32{0, 1, Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)}
		random := uint32(seed + 29)
		for i := 0; i < 18; i++ {
			random = random*1664525 + 1013904223
			a := lits[int(random)%len(lits)] ^ int32(random>>31)
			random = random*1664525 + 1013904223
			b := lits[int(random)%len(lits)] ^ int32(random>>31)
			if i%4 == 0 {
				b = a ^ int32((i/4)&1)
			}
			var lit int32
			if i%5 == 0 {
				lit = Gia_ManAppendBuf(g, a)
			} else {
				lit = Gia_ManAppendAnd(g, a, b)
			}
			lits = append(lits, lit)
		}
		truth := make([]uint8, g.nObjs)
		ciTruth := []uint8{0xaa, 0xcc, 0xf0}
		for mask := 0; mask < 8; mask++ {
			values, _, err := g.evaluateObjects([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
			if err != nil {
				t.Fatal(err)
			}
			for id, v := range values {
				if v {
					truth[id] |= 1 << uint(mask)
				}
			}
		}
		bits := func(lit int32) uint8 {
			v := truth[lit>>1]
			if lit&1 != 0 {
				v = ^v
			}
			return v
		}
		Gia_ManFillValue(g)
		Gia_ManCleanPhase(g)
		p := Tas_ManAlloc(g, 1000)
		for heuristic := 0; heuristic < 4; heuristic++ {
			p.Pars.fUseHighest, p.Pars.fUseLowest, p.Pars.fUseMaxFF, p.Pars.fUseActive = giaBool(heuristic == 0), giaBool(heuristic == 1), giaBool(heuristic == 2), giaBool(heuristic == 3)
			for query, first := range lits {
				for phase := int32(0); phase < 2; phase++ {
					for mode := 0; mode < 4; mode++ {
						root := first ^ phase
						second := lits[(query+5)%len(lits)] ^ phase
						expected := bits(root)
						var status int32
						if mode == 0 {
							status = Tas_ManSolve(p, Gia_ObjFromLit(g, root), nil)
						} else {
							if mode == 2 {
								second = root ^ 1
							}
							expected &= bits(second)
							if mode == 3 {
								third := lits[(query+11)%len(lits)] ^ (phase ^ 1)
								expected &= bits(third)
								status = Tas_ManSolveArray(p, []*Gia_Obj_t{Gia_ObjFromLit(g, root), Gia_ObjFromLit(g, second), Gia_ObjFromLit(g, third)})
							} else {
								status = Tas_ManSolve(p, Gia_ObjFromLit(g, root), Gia_ObjFromLit(g, second))
							}
						}
						if status != giaBool(expected == 0) {
							t.Fatalf("raw graph%d first%d second%d mode%d: verdict%d expected bits%x", seed, root, second, mode, status, expected)
						}
						if status == 0 {
							completions := uint8(0xff)
							seen := map[int32]bool{}
							for _, entry := range Vec_IntArray(p.vModel) {
								id := entry >> 1
								if id < 0 || seen[id] {
									t.Fatal("invalid raw model")
								}
								seen[id] = true
								if id >= 3 {
									t.Fatal("raw CI model index")
								}
								v := ciTruth[id]
								if entry&1 != 0 {
									v = ^v
								}
								completions &= v
							}
							if completions == 0 || completions&^expected != 0 {
								t.Fatalf("raw graph%d query%d phase%d mode%d model does not imply roots", seed, query, phase, mode)
							}
						}
						for _, obj := range g.pObjs {
							if obj.fMark0 != 0 || obj.fMark1 != 0 || obj.fPhase != 0 || obj.Value != ^uint32(0) {
								t.Fatal("raw graph query retained assignments")
							}
						}
						if Vec_IntSize(p.vLevReas) != 0 || p.pClauses.iHead != 1 || p.pClauses.iTail != 1 || p.nClauses != 0 || Vec_IntSize(p.vWatchLits) != 0 {
							t.Fatal("raw graph retained reasons/clauses")
						}
						queries++
					}
				}
			}
		}
		Tas_ManStop(p)
	}
	t.Logf("%d exhaustive raw-AND/constant/buffer/shared-fanin queries with validated witnesses", queries)
}

func TestTas2AllocationZeroesRecordsAndKeepsFlags(t *testing.T) {
	manager := &Tas2_Man_t{}
	first := Tas2_ManCreateVar(manager)
	first.fTerm, first.fAssign, first.fPhase = true, true, true
	first.Level, first.Reason0 = 19, -4
	expected := *first
	for i := 1; i < 4096; i++ {
		variable := Tas2_ManCreateVar(manager)
		if *variable != (Tas2_Var_t{Id: int32(i)}) {
			t.Fatal("new variable inherited nonzero storage")
		}
		variable.fVal, variable.fMark0 = true, true
	}
	if manager.pVars[0] != first || *first != expected {
		t.Fatal("growth changed existing variable state")
	}
	if Tas2_ManCreateVar(nil) != nil {
		t.Fatal("nil variable manager")
	}
	t.Logf("Go variable record: %d bytes", reflect.TypeOf(*first).Size())
}

func TestTasAllRootPairs(t *testing.T) {
	queries := 0
	ciTruth := [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	for seed := 0; seed < 12; seed++ {
		g, lits := cbs0ReviewGraph(seed)
		Gia_ManFillValue(g)
		Gia_ManCleanPhase(g)
		truth := make([]uint32, g.nObjs)
		for mask := 0; mask < 32; mask++ {
			values, _, err := g.evaluateObjects([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0})
			if err != nil {
				t.Fatal(err)
			}
			for id, value := range values {
				if value {
					truth[id] |= 1 << uint(mask)
				}
			}
		}
		bits := func(lit int32) uint32 {
			v := truth[lit>>1]
			if lit&1 != 0 {
				v = ^v
			}
			return v
		}
		roots := []int32{0, 1}
		seen := make([]bool, 2*g.nObjs)
		seen[0], seen[1] = true, true
		for _, lit := range lits {
			for phase := int32(0); phase < 2; phase++ {
				root := lit ^ phase
				if !seen[root] {
					roots = append(roots, root)
					seen[root] = true
				}
			}
		}
		p := Tas_ManAlloc(g, 1000)
		for mode := int32(0); mode < 4; mode++ {
			p.Pars.fUseHighest, p.Pars.fUseLowest, p.Pars.fUseMaxFF, p.Pars.fUseActive = giaBool(mode == 0), giaBool(mode == 1), giaBool(mode == 2), giaBool(mode == 3)
			for _, first := range roots {
				for _, second := range roots {
					expected := bits(first) & bits(second)
					status := Tas_ManSolve(p, Gia_ObjFromLit(g, first), Gia_ObjFromLit(g, second))
					if status != giaBool(expected == 0) {
						t.Fatalf("graph%d mode%d pair%d,%d verdict%d expected%x", seed, mode, first, second, status, expected)
					}
					if status == 0 {
						completions := ^uint32(0)
						used := make([]bool, 5)
						for _, lit := range Vec_IntArray(p.vModel) {
							id := lit >> 1
							if lit < 0 || id >= 5 || used[id] {
								t.Fatal("invalid paired witness literal")
							}
							used[id] = true
							v := ciTruth[id]
							if lit&1 != 0 {
								v = ^v
							}
							completions &= v
						}
						if completions == 0 || completions & ^expected != 0 {
							t.Fatalf("graph%d mode%d pair%d,%d invalid witness%v", seed, mode, first, second, Vec_IntArray(p.vModel))
						}
					} else if Vec_IntSize(p.vModel) != 0 {
						t.Fatal("failed query retained model")
					}
					for _, obj := range g.pObjs {
						if obj.fMark0 != 0 || obj.fMark1 != 0 || obj.fPhase != 0 || obj.Value != ^uint32(0) {
							t.Fatal("paired query retained graph scratch")
						}
					}
					if p.pProp.iHead != 0 || p.pProp.iTail != 0 || p.pJust.iHead != 0 || p.pJust.iTail != 0 || p.pClauses.iHead != 1 || p.pClauses.iTail != 1 || Vec_IntSize(p.vLevReas) != 0 || Vec_IntSize(p.vWatchLits) != 0 || p.nClauses != 0 || len(p.pStore.words) != 16 {
						t.Fatal("paired query retained solver scratch")
					}
					for _, v := range p.pWatches {
						if v != 0 {
							t.Fatal("paired query retained watch")
						}
					}
					queries++
				}
			}
		}
		Tas_ManStop(p)
	}
	t.Logf("%d complete five-input root-pair queries with exhaustive verdict and witness checks", queries)
}

func TestTasNestedQueuesAndLifecycle(t *testing.T) {
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManFillValue(g)
	p := Tas_ManAlloc(g, 100)
	defer Tas_ManStop(p)
	q := Tas_Que_t{}
	Tas_QuePush(&q, Gia_ObjFromLit(g, a))
	Tas_QuePush(&q, Gia_ObjFromLit(g, b))
	var h0, t0, h1, t1 int32
	Tas_QueStore(&q, &h0, &t0)
	for i := 0; i < 3000; i++ {
		Tas_QuePush(&q, Gia_ObjFromLit(g, a))
	}
	Tas_QueStore(&q, &h1, &t1)
	q.pData[q.iHead] = nil
	Tas_QueRestore(&q, h1, t1)
	if Tas_QueHasNode(&q, Gia_ObjFromLit(g, b)) != 1 {
		t.Fatal("inner frontier lost")
	}
	Tas_QueRestore(&q, h0, t0)
	if q.iHead != 0 || q.iTail != 2 || q.pData[0] != Gia_ObjFromLit(g, a) || q.pData[1] != Gia_ObjFromLit(g, b) {
		t.Fatal("outer frontier lost")
	}
	cexReviewMustPanic(t, func() { Tas_QueRestore(&q, -1, 2) })
	Tas_ManAssign(p, Gia_ObjFromLit(g, a), 0, nil, nil)
	Tas_ManAssign(p, Gia_ObjFromLit(g, b^1), 1, nil, nil)
	Tas_ManSaveModel(p, p.vModel)
	if !slices.Equal(Vec_IntArray(p.vModel), []int32{0, 3}) {
		t.Fatal("CI witness convention")
	}
	if Tas_VarReason0(p, g.pObjs[a>>1]) != g.pObjs[a>>1] || Tas_VarReason1(p, g.pObjs[a>>1]) != g.pObjs[a>>1] {
		t.Fatal("decision reason sentinel")
	}
	Tas_ManCancelUntil(p, 1)
	if Vec_IntSize(p.vLevReas) != 3 || Tas_VarIsAssigned(g.pObjs[a>>1]) != 1 || Tas_VarIsAssigned(g.pObjs[b>>1]) != 0 {
		t.Fatal("partial trail cancellation")
	}
	cexReviewMustPanic(t, func() { Tas_ManCancelUntil(p, 2) })
	p.Pars.nBTLimit, p.Pars.nBTThis, p.Pars.nJustLimit, p.Pars.nJustThis = 7, 7, 1, 1
	if Tas_ManCheckLimits(p) != 0 {
		t.Fatal("limits must be strict")
	}
	p.Pars.nBTThis++
	if Tas_ManCheckLimits(p) != 1 {
		t.Fatal("conflict limit ignored")
	}
	Tas_ManStop(p)
	Tas_ManStop(p)
	Tas_ManStop(nil)
	cexReviewMustPanic(t, func() { Tas_ManSolve(p, Gia_ObjFromLit(g, a), nil) })
	if p.pAig != nil || p.vModel != nil || p.vLevReas != nil || p.vTemp != nil || p.pStore.words != nil || p.pWatches != nil {
		t.Fatal("stopped manager retained graph/storage")
	}
	for _, obj := range g.pObjs {
		if obj.fMark0 != 0 || obj.fMark1 != 0 || obj.Value != ^uint32(0) {
			t.Fatal("active Stop retained scratch")
		}
	}
}
