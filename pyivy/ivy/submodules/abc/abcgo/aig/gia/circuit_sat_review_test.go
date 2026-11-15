package gia

import "testing"

func TestCircuitSATRejectsInvalidIdleQueues(t *testing.T) {
	for field := 0; field < 6; field++ {
		g := Gia_ManStart(16)
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		root := Gia_ManAppendAnd(g, a, b)
		p := Cbs_ManAlloc(g)
		switch field {
		case 0:
			p.pProp.iHead = 1
		case 1:
			p.pProp.iTail = 1
		case 2:
			p.pJust.iHead = 1
		case 3:
			p.pJust.iTail = 1
		case 4:
			p.pClauses.iHead = 2
		case 5:
			p.pClauses.iTail = 2
		}
		// Keep a valid trail entry when checking an active propagation queue.
		if field == 1 {
			p.pProp.pData = append(p.pProp.pData, g.pObjs[a>>1])
			Cbs_VarAssign(g.pObjs[a>>1])
			Cbs_VarSetValue(g.pObjs[a>>1], 1)
		}
		cexReviewMustPanic(t, func() { Cbs_ManSolve(p, Gia_ObjFromLit(g, root)) })
		Cbs_ManStop(p)
	}
}
func TestCircuitSATRejectsForeignAndOutputRoots(t *testing.T) {
	for pair := 0; pair < 2; pair++ {
		for kind := 0; kind < 3; kind++ {
			if pair == 1 && kind == 0 {
				continue
			} // A nil second root is a valid single-root all-object query.
			g := Gia_ManStart(16)
			a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			root := Gia_ManAppendAnd(g, a, b)
			co := Gia_ManAppendCo(g, root)
			other := Gia_ManStart(8)
			foreign := Gia_ManAppendCi(other)
			bad := []*Gia_Obj_t{nil, Gia_ObjFromLit(other, foreign), Gia_ObjFromLit(g, co)}[kind]
			p := Cbs_ManAlloc(g)
			cexReviewMustPanic(t, func() {
				if pair == 0 {
					Cbs_ManSolve(p, bad)
				} else {
					Cbs_ManSolve2(p, Gia_ObjFromLit(g, root), bad)
				}
			})
			Cbs_ManStop(p)
		}
	}
}

func BenchmarkCircuitSATConflictScratch(b *testing.B) {
	for mode, name := range []string{"single", "paired"} {
		b.Run(name, func(b *testing.B) {
			g, lits := cbs0ReviewGraph(8)
			p := Cbs_ManAlloc(g)
			defer Cbs_ManStop(p)
			solve := func() int32 {
				if mode == 0 {
					return Cbs_ManSolve(p, Gia_ObjFromLit(g, lits[19]))
				}
				return Cbs_ManSolve2(p, Gia_ObjFromLit(g, lits[19]), Gia_ObjFromLit(g, lits[0]))
			}
			if solve() != 1 {
				b.Fatal("hard conflict fixture is not UNSAT")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if solve() != 1 {
					b.Fatal("conflict result changed")
				}
			}
		})
	}
}

func TestCircuitSATClauseReportForms(t *testing.T) {
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	p := Cbs_ManAlloc(g)
	defer Cbs_ManStop(p)
	Cbs_ManAssign(p, Gia_ObjFromLit(g, a), 0, nil, nil)
	Cbs_ManAssign(p, Gia_ObjFromLit(g, b^1), 1, nil, nil)
	Cbs_QuePush(&p.pClauses, Gia_ObjFromLit(g, a))
	Cbs_QuePush(&p.pClauses, Gia_ObjFromLit(g, b))
	clause := Cbs_QueFinish(&p.pClauses)
	if Cbs_ManClauseSize(p, clause) != 2 || Cbs_ClauseDecLevel(p, clause) != 0 {
		t.Fatal("clause size/decision level")
	}
	text := hashReviewStdout(t, func() { Cbs_ManPrintClause(p, 2, clause); Cbs_ManPrintClauseNew(p, 2, clause) })
	if text != "Level  2 : 1=1(0) 2=0(1) \nLevel  2 : +1 -2 \n" {
		t.Fatalf("source clause report forms differ: %q", text)
	}
}

func TestCircuitSATConstantFaninsAreFixedValues(t *testing.T) {
	for _, constant := range []int32{0, 1} {
		for phase := int32(0); phase < 2; phase++ {
			g := Gia_ManStart(16)
			a := Gia_ManAppendCi(g)
			root := Gia_ManAppendAnd(g, constant, a)
			p := Cbs_ManAlloc(g)
			status := Cbs_ManSolve(p, Gia_ObjFromLit(g, root^phase))
			want := giaBool(constant == 0 && phase == 0)
			if status != want {
				t.Fatalf("constant%d phase%d verdict%d want%d", constant, phase, status, want)
			}
			if status == 0 {
				for input := int32(0); input < 2; input++ {
					matches := true
					for _, lit := range Vec_IntArray(p.vModel) {
						if lit>>1 != 0 {
							t.Fatal("invalid CI model index")
						}
						matches = matches && (input != 0) == (lit&1 == 0)
					}
					if matches && (constant != 0 && input != 0) == (phase != 0) {
						t.Fatal("constant-fanin cube does not force root")
					}
				}
			}
			for _, obj := range g.pObjs {
				if obj.fMark0 != 0 || obj.fMark1 != 0 || obj.Value != ^uint32(0) {
					t.Fatal("constant-fanin scratch retained")
				}
			}
			Cbs_ManStop(p)
		}
	}
}

func TestCircuitSATRawAndExtensions(t *testing.T) {
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
		p := Cbs_ManAlloc(g)
		for heuristic := 0; heuristic < 3; heuristic++ {
			p.Pars.fUseHighest, p.Pars.fUseLowest, p.Pars.fUseMaxFF = giaBool(heuristic == 0), giaBool(heuristic == 1), giaBool(heuristic == 2)
			for query, first := range lits {
				for phase := int32(0); phase < 2; phase++ {
					for mode := 0; mode < 3; mode++ {
						root := first ^ phase
						second := lits[(query+5)%len(lits)] ^ phase
						expected := bits(root)
						var status int32
						if mode == 0 {
							status = Cbs_ManSolve(p, Gia_ObjFromLit(g, root))
						} else {
							if mode == 2 {
								second = root ^ 1
							}
							expected &= bits(second)
							status = Cbs_ManSolve2(p, Gia_ObjFromLit(g, root), Gia_ObjFromLit(g, second))
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
						for _, obj := range g.pObjs {
							if obj.fMark0 != 0 || obj.fMark1 != 0 || obj.Value != ^uint32(0) {
								t.Fatal("raw graph query retained assignments")
							}
						}
						if Vec_IntSize(p.vLevReas) != 0 || p.pClauses.iHead != 1 || p.pClauses.iTail != 1 {
							t.Fatal("raw graph retained reasons/clauses")
						}
						queries++
					}
				}
			}
		}
		Cbs_ManStop(p)
	}
	t.Logf("%d exhaustive raw-AND/constant/buffer/shared-fanin queries with validated witnesses", queries)
}
