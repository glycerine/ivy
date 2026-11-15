package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestCircuitSAT2ConflictAnalysisMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat2_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native CBS2: %v\n%s", err, native)
	}
	var got strings.Builder
	var outcomes [3]int
	var maxBT, maxJust int32
	cases := 0
	for seed := 0; seed < 12; seed++ {
		g, lits := cbs0ReviewGraph(seed)
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
		for fanout := 0; fanout < 2; fanout++ {
			for _, budget := range []int32{0, 1, 3, 1000} {
				for _, frontier := range []int32{0, 1, 3, 100} {
					p := Cbs2_ManAlloc(g)
					p.Pars.fUseFanout = int32(fanout)
					p.Pars.nBTLimit, p.Pars.nJustLimit = budget, frontier
					for query, lit := range lits {
						if lit >= 2 {
							for phase := int32(0); phase < 2; phase++ {
								for _, mode := range []int{0, 2} {
									if mode == 2 && query%5 != 0 {
										continue
									}
									root := lit ^ phase
									var status int32
									if mode == 0 {
										status = Cbs2_ManSolve(p, root)
									} else {
										status = Cbs2_ManSolve2(p, root, 0)
									}
									outcomes[status+1]++
									maxBT = max(maxBT, p.Pars.nBTThis)
									maxJust = max(maxJust, p.Pars.nJustThis)
									fmt.Fprintf(&got, "case %d %d %d %d %d %d %d status%d bt%d nc%d just%d totals%d:%d prop%d:%d:%d fails%d:%d clauses%d model", seed, fanout, budget, frontier, query, phase, mode, status, p.Pars.nBTThis, p.Pars.nBTThisNc, p.Pars.nJustThis, p.Pars.nBTTotal, p.Pars.nJustTotal, p.nPropCalls[0], p.nPropCalls[1], p.nPropCalls[2], p.nFails[0], p.nFails[1], p.nClauseConf)
									expected := bits(root)
									if status == 0 {
										if expected == 0 {
											t.Fatal("SAT for inconsistent roots")
										}
										completions := ^uint32(0)
										seen := map[int32]bool{}
										for _, entry := range Vec_IntArray(Cbs2_ReadModel(p)) {
											fmt.Fprintf(&got, " %d", entry)
											id := entry >> 1
											if id < 0 || seen[id] {
												t.Fatal("invalid model object")
											}
											seen[id] = true
											var values uint32
											if mode == 0 {
												if id >= 5 {
													t.Fatal("invalid CI model index")
												}
												values = ciTruth[id]
											} else {
												if id >= g.nObjs {
													t.Fatal("invalid all-object model index")
												}
												values = truth[id]
											}
											if entry&1 != 0 {
												values = ^values
											}
											completions &= values
										}
										if completions == 0 || completions&^expected != 0 {
											t.Fatalf("invalid CBS2 cube: seed%d query%d phase%d mode%d", seed, query, phase, mode)
										}
									} else {
										if Vec_IntSize(p.vModel) != 0 {
											t.Fatal("stale model after UNSAT/undecided")
										}
										if status == 1 && expected != 0 {
											t.Fatal("false CBS2 proof")
										}
									}
									if p.pProp.iHead != 0 || p.pProp.iTail != 0 || p.pJust.iHead != 0 || p.pJust.iTail != 0 || p.pClauses.iHead != 1 || p.pClauses.iTail != 1 || len(p.vWatchUpds) != 0 {
										t.Fatal("CBS2 queues/reasons retained after query")
									}
									for id, value := range p.vAssign {
										if value < 2 || p.vMark[id] != 0 || p.vLevReason[3*id] != -1 {
											t.Fatal("CBS2 assignment/reason scratch leaked")
										}
									}
									for _, watch := range p.vWatches {
										if watch != 0 {
											t.Fatal("CBS2 stale watch")
										}
									}
									Cbs2_ManCheckFanouts(p)
									got.WriteByte('\n')
									cases++
								}
							}
						}
					}
					Cbs2_ManStop(p)
				}
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(got.String(), "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("CBS2 source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("native output length Go%d C%d", got.Len(), len(native))
	}
	if outcomes[0] == 0 || outcomes[1] == 0 || outcomes[2] == 0 || maxBT < 2 || maxJust < 2 {
		t.Fatal("weak conflict-analysis sweep")
	}
	t.Logf("%d exact native single-root/CI-or-all-object cases, outcomes%v maxBT%d maxJust%d; all cubes exhaustively checked", cases, outcomes, maxBT, maxJust)
}

func TestCircuitSAT2StatisticsMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat2_review.c")
	native, err := exec.Command(exe, "stats").CombinedOutput()
	if err != nil {
		t.Fatalf("native CBS2 statistics: %v\n%s", err, native)
	}
	g := Gia_ManStart(16)
	a := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	got := hashReviewStdout(t, func() {
		p := Cbs2_ManAlloc(g)
		p.nSatTotal, p.nSatUnsat, p.nSatSat, p.nSatUndec = 4, 2, 1, 1
		p.nConfUnsat, p.nConfSat, p.nConfUndec = 5, 3, 7
		p.timeTotal, p.timeSatUnsat, p.timeSatSat, p.timeSatUndec = 4*time.Second, 2*time.Second, time.Second, time.Second/2
		Cbs2_ManSatPrintStats(p)
		Cbs2_ManStop(p)
		p = Cbs2_ManAlloc(g)
		Cbs2_ManSatPrintStats(p)
		Cbs2_ManStop(p)
	})
	if got != string(native) {
		t.Fatalf("CBS2 statistics differ\nGo:%s\nC:%s", got, native)
	}
}

func TestCircuitSAT2MitersAndPhasesMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat2_review.c")
	native, err := exec.Command(exe, "miters").CombinedOutput()
	if err != nil {
		t.Fatalf("native CBS2 miter: %v\n%s", err, native)
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
				store := Cbs2_ManSolveMiterNc(g, budget, &statuses, 0)
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
				t.Fatalf("CBS2 miter source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("miter source lengths Go%d C%d", len(got), len(native))
	}
	t.Log("3600 exact source miter properties, witnesses and unchanged phases")
}
