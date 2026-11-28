package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestTasMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_tas_review.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native TAS: %v\n%s", err, native)
	}
	var got strings.Builder
	var outcomes [3]int
	var maxBT, maxJust int32
	cases := 0
	for seed := 0; seed < 12; seed++ {
		g, lits := cbs0ReviewGraph(seed)
		Gia_ManFillValue(g)
		Gia_ManCleanPhase(g)
		truth := make([]uint32, g.nObjs)
		ciTruth := make([]uint32, 5)
		for mask := 0; mask < 32; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
			values, _, err := g.evaluateObjects(inputs)
			if err != nil {
				t.Fatal(err)
			}
			for id, value := range values {
				if value {
					truth[id] |= 1 << uint(mask)
				}
			}
			for i, value := range inputs {
				if value {
					ciTruth[i] |= 1 << uint(mask)
				}
			}
		}
		bits := func(lit int32) uint32 {
			value := truth[lit>>1]
			if lit&1 != 0 {
				value = ^value
			}
			return value
		}
		for heuristic := 0; heuristic < 4; heuristic++ {
			for _, budget := range []int32{0, 1, 3, 1000} {
				for _, frontier := range []int32{0, 1, 3, 100} {
					p := Tas_ManAlloc(g, budget)
					p.Pars.fUseHighest, p.Pars.fUseLowest, p.Pars.fUseMaxFF, p.Pars.fUseActive = giaBool(heuristic == 0), giaBool(heuristic == 1), giaBool(heuristic == 2), giaBool(heuristic == 3)
					p.Pars.nBTLimit, p.Pars.nJustLimit = budget, frontier
					for query, lit := range lits {
						if lit >= 2 {
							for phase := int32(0); phase < 2; phase++ {
								for mode := 0; mode < 3; mode++ {
									other := lits[(query+7)%29] ^ (phase ^ 1)
									if mode == 1 && (other < 2 || other>>1 == lit>>1) {
										continue
									}
									if mode == 2 && query%5 != 0 {
										continue
									}
									root := lit ^ phase
									var status int32
									if mode == 0 {
										status = Tas_ManSolve(p, Gia_ObjFromLit(g, root), nil)
									} else if mode == 1 {
										status = Tas_ManSolve(p, Gia_ObjFromLit(g, root), Gia_ObjFromLit(g, other))
									} else {
										status = Tas_ManSolveArray(p, []*Gia_Obj_t{Gia_ObjFromLit(g, root)})
									}
									outcomes[status+1]++
									maxBT = max(maxBT, p.Pars.nBTThis)
									maxJust = max(maxJust, p.Pars.nJustThis)
									fmt.Fprintf(&got, "case %d %d %d %d %d %d %d status%d bt%d nc%d just%d totals%d:%d model", seed, heuristic, budget, frontier, query, phase, mode, status, p.Pars.nBTThis, p.Pars.nBTThisNc, p.Pars.nJustThis, p.Pars.nBTTotal, p.Pars.nJustTotal)
									expected := bits(root)
									if mode == 1 {
										expected &= bits(other)
									}
									if status == 0 {
										if expected == 0 {
											t.Fatal("SAT for inconsistent roots")
										}
										completions := ^uint32(0)
										seen := map[int32]bool{}
										for _, entry := range Vec_IntArray(Tas_ReadModel(p)) {
											fmt.Fprintf(&got, " %d", entry)
											id := entry >> 1
											if id < 0 || seen[id] {
												t.Fatal("invalid model object")
											}
											seen[id] = true
											if id >= 5 {
												t.Fatal("invalid CI model index")
											}
											values := ciTruth[id]
											if entry&1 != 0 {
												values = ^values
											}
											completions &= values
										}
										if completions == 0 || completions&^expected != 0 {
											t.Fatalf("invalid TAS cube: seed%d query%d phase%d mode%d", seed, query, phase, mode)
										}
									} else {
										if Vec_IntSize(p.vModel) != 0 {
											t.Fatal("stale model after UNSAT/undecided")
										}
										if status == 1 && expected != 0 {
											t.Fatal("false TAS proof")
										}
									}
									if p.pProp.iHead != 0 || p.pProp.iTail != 0 || p.pJust.iHead != 0 || p.pJust.iTail != 0 || p.pClauses.iHead != 1 || p.pClauses.iTail != 1 || Vec_IntSize(p.vLevReas) != 0 {
										t.Fatal("TAS queues/reasons retained after query")
									}
									for _, obj := range g.pObjs {
										if obj.fMark0 != 0 || obj.fMark1 != 0 || obj.fPhase != 0 || obj.Value != ^uint32(0) {
											t.Fatal("TAS assignment/reason scratch leaked")
										}
									}
									if p.nClauses != 0 || len(p.pStore.words) != 16 || Vec_IntSize(p.vWatchLits) != 0 || Vec_IntSize(p.vActiveVars) != 0 {
										t.Fatal("TAS retained clause/activity store")
									}
									for _, h := range p.pWatches {
										if h != 0 {
											t.Fatal("TAS retained watch")
										}
									}
									got.WriteByte('\n')
									cases++
								}
							}
						}
					}
					Tas_ManStop(p)
				}
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(got.String(), "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("TAS source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("native output length Go%d C%d", got.Len(), len(native))
	}
	if outcomes[0] == 0 || outcomes[1] == 0 || outcomes[2] == 0 || maxBT < 2 || maxJust < 2 {
		t.Fatal("weak conflict-analysis sweep")
	}
	t.Logf("%d exact native single/paired/array cases, outcomes%v maxBT%d maxJust%d; all cubes exhaustively checked", cases, outcomes, maxBT, maxJust)
}

func TestTasStatisticsMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_tas_review.c")
	native, err := testoracle.Command(t, exe, "stats").CombinedOutput()
	if err != nil {
		t.Fatalf("native TAS statistics: %v\n%s", err, native)
	}
	g := Gia_ManStart(16)
	a := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	got := hashReviewStdout(t, func() {
		p := Tas_ManAlloc(g, 1000)
		p.nSatTotal, p.nSatUnsat, p.nSatSat, p.nSatUndec = 4, 2, 1, 1
		p.nConfUnsat, p.nConfSat, p.nConfUndec = 5, 3, 7
		p.timeTotal, p.timeSatUnsat, p.timeSatSat, p.timeSatUndec = 4*time.Second, 2*time.Second, time.Second, time.Second/2
		Tas_ManSatPrintStats(p)
		Tas_ManStop(p)
		p = Tas_ManAlloc(g, 1000)
		Tas_ManSatPrintStats(p)
		Tas_ManStop(p)
	})
	if got != string(native) {
		t.Fatalf("TAS statistics differ\nGo:%s\nC:%s", got, native)
	}
}

func TestTasMitersAndPhasesMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_tas_review.c")
	native, err := testoracle.Command(t, exe, "miters").CombinedOutput()
	if err != nil {
		t.Fatalf("native TAS miter: %v\n%s", err, native)
	}
	got := hashReviewStdout(t, func() {
		for seed := 0; seed < 12; seed++ {
			for _, budget := range []int32{-1, 0, 1, 3, 1000} {
				g, lits := cbs0ReviewGraph(seed)
				Gia_ManAppendCo(g, 0)
				Gia_ManAppendCo(g, 1)
				for _, lit := range lits {
					Gia_ManAppendCo(g, lit)
					Gia_ManAppendCo(g, lit^1)
				}
				for _, ports := range []struct {
					v       *Vec_Int_t
					reverse bool
				}{{g.vCos, seed&1 != 0}, {g.vCis, seed&2 != 0}} {
					if ports.reverse {
						slices.Reverse(Vec_IntArray(ports.v))
						for i, id := range Vec_IntArray(ports.v) {
							Gia_ObjSetCioId(g.pObjs[id], int32(i))
						}
					}
				}
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
				for _, obj := range g.pObjs {
					obj.fPhase = 1
				}
				var statuses *Vec_Str_t
				store := Tas_ManSolveMiterNc(g, budget, &statuses, 0)
				cbs0ReviewStore(t, g, statuses, store, truth)
				fmt.Printf("miter %d %d statuses", seed, budget)
				for _, status := range Vec_StrArray(statuses) {
					fmt.Printf(" %d", status)
				}
				fmt.Print(" store")
				for _, value := range Vec_IntArray(store) {
					fmt.Printf(" %d", value)
				}
				fmt.Print(" drivers")
				for i := int32(0); i < Gia_ManCoNum(g); i++ {
					lit, _, _, _ := g.Gate(Gia_ObjId(g, Gia_ManCo(g, i)))
					fmt.Printf(" %d", lit)
				}
				fmt.Print(" phases")
				for _, obj := range g.pObjs {
					fmt.Printf("%d", obj.fPhase)
				}
				fmt.Println()
			}
		}
	})
	if got != string(native) {
		a, b := strings.Split(got, "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("TAS miter source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("miter source lengths Go%d C%d", len(got), len(native))
	}
	t.Log("3600 exact source miter properties, witnesses and cleared phases")
}

func TestTasWatchedClausesMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_tas_review.c")
	native, err := testoracle.Command(t, exe, "watch").CombinedOutput()
	if err != nil {
		t.Fatalf("native TAS watches: %v\n%s", err, native)
	}
	g := Gia_ManStart(16)
	a, b, c, d, e := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManFillValue(g)
	p := Tas_ManAlloc(g, 1000)
	defer Tas_ManStop(p)
	p.pStore.words = make([]int32, 16, 16)
	p.pClauses.pData = make([]*Gia_Obj_t, 1, 2)
	learned := func(lits []int32) {
		c := Tas_ManAllocCls(p, int32(len(lits)+3))
		copy(c.literals(), lits)
		Tas_ManWatchClause(p, c, lits[0])
		Tas_ManWatchClause(p, c, lits[1])
	}
	snapshot := func(label string, h int32) {
		fmt.Printf("watch %s h%d head%d tail%d clauses", label, h, p.pClauses.iHead, p.pClauses.iTail)
		for _, obj := range p.pClauses.pData[1:p.pClauses.iTail] {
			id := int32(-1)
			if obj != nil {
				id = obj.id
			}
			fmt.Printf(" %d", id)
		}
		fmt.Print(" store")
		for _, v := range p.pStore.words[16:] {
			fmt.Printf(" %d", v)
		}
		fmt.Print(" trail")
		for _, obj := range p.pProp.pData[:p.pProp.iTail] {
			fmt.Printf(" %d", Tas_VarToLit(p, obj))
		}
		fmt.Print(" watches")
		for _, v := range p.pWatches {
			fmt.Printf(" %d", v)
		}
		fmt.Print(" levels")
		for _, v := range Vec_IntArray(p.vLevReas) {
			fmt.Printf(" %d", v)
		}
		fmt.Print(" marks")
		for _, obj := range g.pObjs {
			fmt.Printf(" %d%d%d:%d", obj.fMark0, obj.fMark1, obj.fPhase, obj.Value)
		}
		fmt.Printf(" count%d\n", p.nClauses)
	}
	got := hashReviewStdout(t, func() {
		learned([]int32{a, b, c, d})
		learned([]int32{a, e})
		Tas_ManAssign(p, Gia_ObjFromLit(g, a^1), 0, nil, nil)
		h := Tas_ManPropagateWatch(p, 0, a^1)
		snapshot("a", h)
		Tas_ManAssign(p, Gia_ObjFromLit(g, b^1), 1, nil, nil)
		h = Tas_ManPropagateWatch(p, 1, b^1)
		snapshot("b", h)
		Tas_ManAssign(p, Gia_ObjFromLit(g, c^1), 2, nil, nil)
		h = Tas_ManPropagateWatch(p, 2, c^1)
		snapshot("unit", h)
		Tas_ManCancelUntil(p, 4)
		Tas_ManAssign(p, Gia_ObjFromLit(g, d^1), 3, nil, nil)
		h = Tas_ManPropagateWatch(p, 3, c^1)
		if h == 0 || Tas_ManClauseSize(p, h) != 4 {
			t.Fatal("missing learned conflict")
		}
		Tas_ManCreateCls(p, h)
		snapshot("conflict", h)
		Tas_ManPrintClause(p, 3, h)
		Tas_ManPrintClauseNew(p, 3, h)
	})
	if got != string(native) {
		t.Fatalf("TAS watched clauses differ\nGo:%s\nC:%s", got, native)
	}
	Tas_ManCancelUntil(p, 0)
}

func TestTasPackedMiterMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_tas_review.c")
	native, err := testoracle.Command(t, exe, "Packed").CombinedOutput()
	if err != nil {
		t.Fatalf("native TAS packed miter: %v\n%s", err, native)
	}
	snapshots := func(report string) string {
		var b strings.Builder
		for _, line := range strings.Split(report, "\n") {
			if strings.HasPrefix(line, "packed ") {
				b.WriteString(line)
				b.WriteByte('\n')
			}
		}
		return b.String()
	}
	got := hashReviewStdout(t, func() {
		for _, seed := range []int{0, 3, 8, 11} {
			for _, budget := range []int32{-1, 0, 1, 1000} {
				for _, count := range []int{17, 1000, 1001, 2000, 2001} {
					g, lits := cbs0ReviewGraph(seed)
					var pool []int32
					for _, lit := range lits {
						if lit >= 2 {
							pool = append(pool, lit, lit^1)
						}
					}
					for i := 0; i < count; i++ {
						Gia_ManAppendCo(g, pool[(7*i+seed)%len(pool)])
					}
					for _, ports := range []struct {
						v       *Vec_Int_t
						reverse bool
					}{{g.vCos, seed&1 != 0}, {g.vCis, seed&2 != 0}} {
						if ports.reverse {
							slices.Reverse(Vec_IntArray(ports.v))
							for i, id := range Vec_IntArray(ports.v) {
								Gia_ObjSetCioId(g.pObjs[id], int32(i))
							}
						}
					}
					trueOutputs := make([]bool, count)
					for mask := 0; mask < 32; mask++ {
						_, outputs, e := g.evaluateObjects([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0})
						if e != nil {
							t.Fatal(e)
						}
						for i, value := range outputs {
							trueOutputs[i] = trueOutputs[i] || value
						}
					}
					old := Gia_ManStart(int32(count + 1))
					roots := make([]*Gia_Obj_t, count)
					for i := range roots {
						roots[i] = Gia_ObjFromLit(old, Gia_ManAppendCi(old))
					}
					old.PReprs = make([]Gia_Rpr_t, old.nObjs)
					sim := make([][]uint32, 5)
					random := uint32(seed + 211)
					for i := range sim {
						random = random*1664525 + 1013904223
						sim[i] = []uint32{random}
					}
					Tas_ManSolveMiterNc2(g, budget, old, roots, sim)
					fmt.Printf("packed %d %d %d proved", seed, budget, count)
					step := count/1000 + 1
					for i, root := range roots {
						proved := Gia_ObjProved(old, root.id)
						if proved != 0 && (i%step != 0 || trueOutputs[i]) {
							t.Fatal("proved a skipped or satisfiable output")
						}
						if budget == 1000 && i%step == 0 && !trueOutputs[i] && proved == 0 {
							t.Fatal("missed an exhaustively false output")
						}
						fmt.Printf("%d", proved)
					}
					fmt.Print(" sim")
					for _, words := range sim {
						fmt.Printf(" %d", words[0])
					}
					fmt.Println()
					Gia_ManStop(old)
					Gia_ManStop(g)
				}
			}
		}
	})
	actual, want := snapshots(got), snapshots(string(native))
	if actual != want || strings.Count(actual, "packed ") != 80 {
		t.Fatal("TAS packed miter source snapshots differ")
	}
	t.Log("80 exact native packed miters across budget, logical port and 1000/2000-output sampling boundaries")
}
