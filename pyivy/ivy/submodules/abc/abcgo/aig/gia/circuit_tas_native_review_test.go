package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestTasMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_tas_review.c")
	native, err := exec.Command(exe).CombinedOutput()
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
	native, err := exec.Command(exe, "stats").CombinedOutput()
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
