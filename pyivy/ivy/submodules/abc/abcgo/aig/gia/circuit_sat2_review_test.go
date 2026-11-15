package gia

import (
	"fmt"
	"slices"
	"testing"
)

func TestCircuitSAT2MiterConstantsPrecedeBudget(t *testing.T) {
	g := Gia_ManStart(16)
	a := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, 1)
	Gia_ManAppendCo(g, 0)
	Gia_ManAppendCo(g, a)
	for _, obj := range g.pObjs {
		obj.fPhase = 1
	}
	var statuses *Vec_Str_t
	store := Cbs2_ManSolveMiterNc(g, -1, &statuses, 0)
	if !slices.Equal(Vec_StrArray(statuses), []byte{0, 1, 255}) || !slices.Equal(Vec_IntArray(store), []int32{0, 0, 2, -1}) {
		t.Fatalf("constant properties must bypass solver budget: statuses%v store%v", Vec_StrArray(statuses), Vec_IntArray(store))
	}
	// Unlike CBS, source CBS2 deliberately leaves phase initialization disabled.
	for _, obj := range g.pObjs {
		if obj.fPhase != 1 {
			t.Fatal("CBS2 changed phases")
		}
	}
}

func TestCircuitSAT2RejectsInvalidIdleQueues(t *testing.T) {
	for field := 0; field < 6; field++ {
		t.Run(string(rune('a'+field)), func(t *testing.T) {
			g := Gia_ManStart(16)
			a := Gia_ManAppendCi(g)
			p := Cbs2_ManAlloc(g)
			defer Cbs2_ManStop(p)
			switch field {
			case 0:
				p.pProp.iHead = 1
			case 1:
				Cbs2_ManAssign(p, a, 0, 0, 0)
			case 2:
				p.pJust.iHead = 1
			case 3:
				p.pJust.iTail = 1
				p.pJust.pData = append(p.pJust.pData, a>>1)
			case 4:
				p.pClauses.iHead = 2
			case 5:
				p.pClauses.iTail = 2
			}
			cexReviewMustPanic(t, func() { Cbs2_ManSolve(p, a) })
		})
	}
}

func TestCircuitSAT2RejectedPairDoesNotModifySolver(t *testing.T) {
	for _, kind := range []string{"negative", "outside", "output", "grown", "typed"} {
		t.Run(kind, func(t *testing.T) {
			g := Gia_ManStart(16)
			a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			root := Gia_ManAppendAnd(g, a, b)
			co := Gia_ManAppendCo(g, root)
			p := Cbs2_ManAlloc(g)
			defer Cbs2_ManStop(p)
			bad := int32(-1)
			switch kind {
			case "outside":
				bad = 2 * g.nObjs
			case "output":
				bad = co
			case "grown":
				bad = Gia_ManAppendCi(g)
			case "typed":
				bad = Gia_ManAppendXorReal(g, a, b)
			}
			Vec_IntPush(p.vModel, 1234)
			cexReviewMustPanic(t, func() { Cbs2_ManSolve2(p, root, bad) })
			if p.pProp.iTail != 0 || !slices.Equal(Vec_IntArray(p.vModel), []int32{1234}) {
				t.Fatal("invalid pair modified solver")
			}
			for _, value := range p.vAssign {
				if value < 2 {
					t.Fatal("invalid pair assigned a root")
				}
			}
		})
	}
}

func TestCircuitSAT2ClauseReportForms(t *testing.T) {
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	p := Cbs2_ManAlloc(g)
	defer Cbs2_ManStop(p)
	Cbs2_ManAssign(p, a, 0, 0, 0)
	Cbs2_ManAssign(p, b^1, 1, 0, 0)
	Cbs2_QuePush(&p.pClauses, 0)
	Cbs2_QuePush(&p.pClauses, a>>1)
	Cbs2_QuePush(&p.pClauses, b>>1)
	cube := Cbs2_QueFinish(p, 1)
	// The adjacent learned clause contains complemented assignment literals.
	clause := cube + 3
	got := hashReviewStdout(t, func() { Cbs2_ManPrintCube(p, 2, cube); Cbs2_ManPrintClause(p, 2, clause) })
	if got != "Level  2 : +1 -2 \nLevel  2 : -1 +2 \n" {
		t.Fatalf("cube/clause reports differ: %q", got)
	}
}

func BenchmarkCircuitSAT2ConflictScratch(b *testing.B) {
	for mode, name := range []string{"top-down", "fanout"} {
		b.Run(name, func(b *testing.B) {
			g, lits := cbs0ReviewGraph(8)
			p := Cbs2_ManAlloc(g)
			defer Cbs2_ManStop(p)
			p.Pars.fUseFanout = int32(mode)
			if Cbs2_ManSolve(p, lits[19]) != 1 {
				b.Fatal("hard fixture is not UNSAT")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if Cbs2_ManSolve(p, lits[19]) != 1 {
					b.Fatal("conflict verdict changed")
				}
			}
		})
	}
}

func TestCircuitSAT2RawGateValues(t *testing.T) {
	for _, kind := range []string{"buffer", "repeated", "opposite", "zero", "one", "constant-zero", "constant-one"} {
		for mode := int32(0); mode < 2; mode++ {
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
					p := Cbs2_ManAlloc(g)
					defer Cbs2_ManStop(p)
					p.Pars.fUseFanout = mode
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
					if status := Cbs2_ManSolve(p, root^phase); status != giaBool(truth == 0) {
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
					Cbs2_ManCheckFanouts(p)
				})
			}
		}
	}
}

func TestCircuitSAT2RawAndExtensions(t *testing.T) {
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
		p := Cbs2_ManAlloc(g)
		for modeFanout := int32(0); modeFanout < 2; modeFanout++ {
			p.Pars.fUseFanout = modeFanout
			for query, first := range lits {
				for phase := int32(0); phase < 2; phase++ {
					for mode := 0; mode < 3; mode++ {
						root := first ^ phase
						second := lits[(query+5)%len(lits)] ^ phase
						expected := bits(root)
						var status int32
						if mode == 0 {
							status = Cbs2_ManSolve(p, root)
						} else {
							if mode == 2 {
								second = root ^ 1
							}
							if second != 0 {
								expected &= bits(second)
							}
							status = Cbs2_ManSolve2(p, root, second)
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
								var v uint8
								if mode == 0 {
									if id >= 3 {
										t.Fatal("raw CI model index")
									}
									v = ciTruth[id]
								} else {
									if id >= g.nObjs {
										t.Fatal("raw all-object model index")
									}
									v = truth[id]
								}
								if entry&1 != 0 {
									v = ^v
								}
								completions &= v
							}
							if completions == 0 || completions&^expected != 0 {
								t.Fatalf("raw graph%d query%d phase%d mode%d model does not imply roots", seed, query, phase, mode)
							}
						}
						for id, value := range p.vAssign {
							if value < 2 || p.vMark[id] != 0 || p.vLevReason[3*id] != -1 {
								t.Fatal("raw graph retained assignments")
							}
						}
						if p.pProp.iHead != 0 || p.pProp.iTail != 0 || p.pJust.iHead != 0 || p.pJust.iTail != 0 || p.pClauses.iHead != 1 || p.pClauses.iTail != 1 || len(p.vWatchUpds) != 0 {
							t.Fatal("raw graph retained queues/watches")
						}
						Cbs2_ManCheckFanouts(p)
						queries++
					}
				}
			}
		}
		Cbs2_ManStop(p)
	}
	t.Logf("%d exhaustive raw-AND/constant/buffer/shared-fanin queries with validated witnesses", queries)
}

func TestCircuitSAT2NestedQueuesAndLifecycle(t *testing.T) {
	q := Cbs2_Que_t{}
	Cbs2_QuePush(&q, 11)
	Cbs2_QuePush(&q, 23)
	var h0, t0, h1, t1 int32
	Cbs2_QueStore(&q, &h0, &t0)
	for i := 0; i < 3000; i++ {
		Cbs2_QuePush(&q, int32(i))
	}
	Cbs2_QueStore(&q, &h1, &t1)
	q.pData[q.iHead] = -1
	Cbs2_QueGrow(&q, 2*int32(cap(q.pData)))
	Cbs2_QueRestore(&q, h1, t1)
	if Cbs2_QueHasNode(&q, 23) != 1 {
		t.Fatal("inner saved frontier lost")
	}
	Cbs2_QueRestore(&q, h0, t0)
	if q.iHead != 0 || q.iTail != 2 || !slices.Equal(q.pData[:2], []int32{11, 23}) {
		t.Fatal("outer saved frontier lost")
	}
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	for _, obj := range g.pObjs {
		obj.Value = 1234
		obj.fMark0 = 1
		obj.fMark1 = 1
	}
	p := Cbs2_ManAlloc(g)
	Cbs2_ManAssign(p, a, 0, 0, 0)
	Cbs2_ManAssign(p, b^1, 1, 0, 0)
	model := Vec_IntAlloc(0)
	Cbs2_ManSaveModel(p, model)
	if !slices.Equal(Vec_IntArray(model), []int32{0, 3}) {
		t.Fatal("CI witness convention")
	}
	Cbs2_ManSaveModelAll(p, model)
	if !slices.Equal(Vec_IntArray(model), []int32{2, 5}) {
		t.Fatal("all-object witness convention")
	}
	Cbs2_ManCancelUntil(p, 1)
	if p.pProp.iHead != 1 || p.pProp.iTail != 1 || Cbs2_VarIsAssigned(p, a>>1) != 1 || Cbs2_VarUnused(p, b>>1) != 1 {
		t.Fatal("partial cancellation")
	}
	cexReviewMustPanic(t, func() { Cbs2_ManCancelUntil(p, 2) })
	Cbs2_ManSetConflictNum(p, 7)
	p.Pars.nJustLimit = 1
	p.Pars.nBTThis = 7
	p.Pars.nJustThis = 1
	if Cbs2_ManCheckLimits(p) != 0 {
		t.Fatal("limits must be strict")
	}
	p.Pars.nBTThis++
	if Cbs2_ManCheckLimits(p) != 1 || p.nFails[0] != 0 || p.nFails[1] != 1 {
		t.Fatal("conflict budget/counter")
	}
	p.Pars.nBTThis = 0
	p.Pars.nJustThis++
	if Cbs2_ManCheckLimits(p) != 1 || p.nFails[0] != 1 || p.nFails[1] != 1 {
		t.Fatal("frontier budget/counter")
	}
	p.vActivity[a>>1] = 5
	p.vActStore = append(p.vActStore, a>>1)
	Cbs2_ManBumpClean(p)
	if p.vActivity[a>>1] != 0 {
		t.Fatal("activity cleanup")
	}
	Cbs2_ManStop(p)
	Cbs2_ManStop(p)
	Cbs2_ManStop(nil)
	cexReviewMustPanic(t, func() { Cbs2_ManSolve(p, a) })
	if p.pAig != nil || p.vAssign != nil || p.vLevReason != nil || p.vWatches != nil || p.vTemp != nil || p.vModel != nil || p.pProp.pData != nil {
		t.Fatal("stopped manager retains storage")
	}
	for _, obj := range g.pObjs {
		if obj.Value != 1234 || obj.fMark0 != 1 || obj.fMark1 != 1 {
			t.Fatal("CBS2 changed borrowed graph marks/values")
		}
	}
}

func TestCircuitSAT2AllRootPairs(t *testing.T) {
	queries := 0
	for seed := 0; seed < 12; seed++ {
		g, lits := cbs0ReviewGraph(seed)
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
		p := Cbs2_ManAlloc(g)
		for mode := int32(0); mode < 2; mode++ {
			p.Pars.fUseFanout = mode
			for _, first := range roots {
				for _, second := range roots {
					expected := bits(first)
					// Zero is the source API's absent-second-root sentinel, not a false root.
					if second != 0 {
						expected &= bits(second)
					}
					status := Cbs2_ManSolve2(p, first, second)
					if status != giaBool(expected == 0) {
						t.Fatalf("graph%d mode%d pair%d,%d verdict%d expected%x", seed, mode, first, second, status, expected)
					}
					if status == 0 {
						completions := ^uint32(0)
						used := make([]bool, g.nObjs)
						for _, lit := range Vec_IntArray(p.vModel) {
							id := lit >> 1
							if lit < 0 || id >= g.nObjs || used[id] {
								t.Fatal("invalid paired witness literal")
							}
							used[id] = true
							completions &= bits(lit)
						}
						if completions == 0 || completions & ^expected != 0 {
							t.Fatalf("graph%d mode%d pair%d,%d invalid witness%v", seed, mode, first, second, Vec_IntArray(p.vModel))
						}
					} else if Vec_IntSize(p.vModel) != 0 {
						t.Fatal("failed query retained model")
					}
					for id, v := range p.vAssign {
						if v < 2 || p.vMark[id] != 0 || p.vLevReason[3*id] != -1 {
							t.Fatal("paired query retained assignment/reason")
						}
					}
					if p.pProp.iHead != 0 || p.pProp.iTail != 0 || p.pJust.iHead != 0 || p.pJust.iTail != 0 || p.pClauses.iHead != 1 || p.pClauses.iTail != 1 || len(p.vWatchUpds) != 0 {
						t.Fatal("paired query retained queues/watches")
					}
					for _, v := range p.vWatches {
						if v != 0 {
							t.Fatal("paired query retained watch")
						}
					}
					Cbs2_ManCheckFanouts(p)
					queries++
				}
			}
		}
		Cbs2_ManStop(p)
	}
	t.Logf("%d complete five-input root-pair queries with exhaustive verdict and witness checks", queries)
}
