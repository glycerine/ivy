package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestCircuitSAT2ConflictAnalysisMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat2_review.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
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
	native, err := testoracle.Command(t, exe, "stats").CombinedOutput()
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
	native, err := testoracle.Command(t, exe, "miters").CombinedOutput()
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

func TestCircuitSAT2WatchedClausesMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat2_review.c")
	native, err := testoracle.Command(t, exe, "watch").CombinedOutput()
	if err != nil {
		t.Fatalf("native CBS2 watches: %v\n%s", err, native)
	}
	g := Gia_ManStart(16)
	a, b, c, d, e := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	p := Cbs2_ManAlloc(g)
	defer Cbs2_ManStop(p)
	p.pClauses.pData = make([]int32, 1, 2)
	learned := func(lits []int32) {
		h := p.pClauses.iHead
		Cbs2_QuePush(&p.pClauses, int32(len(lits)))
		for _, lit := range lits {
			Cbs2_QuePush(&p.pClauses, lit)
		}
		Cbs2_QuePush(&p.pClauses, 0)
		Cbs2_QuePush(&p.pClauses, 0)
		p.pClauses.iHead = p.pClauses.iTail
		Cbs2_ManWatchClause(p, h, lits[0])
		Cbs2_ManWatchClause(p, h, lits[1])
	}
	snapshot := func(label string, h int32) {
		fmt.Printf("watch %s h%d head%d tail%d clauses", label, h, p.pClauses.iHead, p.pClauses.iTail)
		for _, v := range p.pClauses.pData[1:p.pClauses.iTail] {
			fmt.Printf(" %d", v)
		}
		for _, item := range []struct {
			name   string
			values []int32
		}{
			{"trail", p.pProp.pData[:p.pProp.iTail]}, {"watches", p.vWatches},
		} {
			fmt.Print(" " + item.name)
			for _, v := range item.values {
				fmt.Printf(" %d", v)
			}
		}
		fmt.Print(" assigns")
		for _, v := range p.vAssign {
			fmt.Printf(" %d", v)
		}
		fmt.Print(" levels")
		for _, v := range p.vLevReason {
			fmt.Printf(" %d", v)
		}
		fmt.Print(" marks")
		for _, v := range p.vMark {
			fmt.Printf(" %d", v)
		}
		fmt.Printf(" calls%d conflicts%d\n", p.nPropCalls[1], p.nClauseConf)
	}
	got := hashReviewStdout(t, func() {
		learned([]int32{a, b, c, d})
		learned([]int32{a, e})
		Cbs2_ManAssign(p, a^1, 0, 0, 0)
		h := Cbs2_ManPropagateClauses(p, 0, a^1)
		snapshot("a", h)
		Cbs2_ManAssign(p, b^1, 1, 0, 0)
		h = Cbs2_ManPropagateClauses(p, 1, b^1)
		snapshot("b", h)
		Cbs2_ManAssign(p, c^1, 2, 0, 0)
		h = Cbs2_ManPropagateClauses(p, 2, c^1)
		snapshot("unit", h)
		Cbs2_ManCancelUntil(p, 4)
		Cbs2_ManAssign(p, d^1, 3, 0, 0)
		h = Cbs2_ManPropagateClauses(p, 3, c^1)
		snapshot("conflict", h)
		if h == 0 || Cbs2_ClauseSize(p, h) != 4 {
			t.Fatal("missing learned conflict")
		}
		Cbs2_ManPrintCube(p, 3, h)
		Cbs2_ManPrintClause(p, 3, h+Cbs2_ClauseSize(p, h)+1)
	})
	if got != string(native) {
		t.Fatalf("CBS2 watches differ\nGo:%s\nC:%s", got, native)
	}
	Cbs2_ManCancelUntil(p, 0)
	Cbs2_ManCleanWatch(p)
	for _, v := range p.vWatches {
		if v != 0 {
			t.Fatal("watch cleanup failed")
		}
	}
}

func TestCircuitSAT2PreparedFanoutsMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat2_review.c")
	native, err := testoracle.Command(t, exe, "fanouts").CombinedOutput()
	if err != nil {
		t.Fatalf("native CBS2 fanouts: %v\n%s", err, native)
	}
	g := Gia_ManStart(16)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	ab := Gia_ManAppendAnd(g, a, b)
	r0, r1 := Gia_ManAppendAnd(g, ab, c), Gia_ManAppendAnd(g, ab, d)
	p := Cbs2_ManAlloc(g)
	defer Cbs2_ManStop(p)
	Gia_ManIncrementTravId(g)
	Cbs2_ManCreateFanout_rec(p, r0>>1)
	Cbs2_ManCreateFanout_rec(p, r1>>1)
	heads, links := slices.Clone(p.vFanout0), slices.Clone(p.vFanoutN)
	got := hashReviewStdout(t, func() {
		Cbs2_ManPrintFanouts(p)
		fmt.Print("heads")
		for _, v := range p.vFanout0 {
			fmt.Printf(" %d", v)
		}
		fmt.Print(" links")
		for _, v := range p.vFanoutN {
			fmt.Printf(" %d", v)
		}
		fmt.Println()
		status := Cbs2_ManSolve2(p, r0, r1)
		fmt.Printf("prepared status%d model", status)
		for _, v := range Vec_IntArray(p.vModel) {
			fmt.Printf(" %d", v)
		}
		fmt.Println()
		Cbs2_ManPrintFanouts(p)
	})
	if got != string(native) {
		t.Fatalf("CBS2 fanouts differ\nGo:%s\nC:%s", got, native)
	}
	if !slices.Equal(heads, p.vFanout0) || !slices.Equal(links, p.vFanoutN) || !p.fanoutsReady {
		t.Fatal("solver changed caller-prepared fanouts")
	}
	Cbs2_ManDeleteFanout_rec(p, r0>>1)
	Cbs2_ManDeleteFanout_rec(p, r1>>1)
	Cbs2_ManCheckFanouts(p)
	if p.fanoutsReady {
		t.Fatal("fanout ownership was not released")
	}
	if Cbs2_ManSolve(p, r0) != 0 {
		t.Fatal("automatic fanout rebuild failed")
	}
	Cbs2_ManCheckFanouts(p)
}

func TestCircuitSAT2VerbosePropagationStatistics(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat2_review.c")
	native, err := testoracle.Command(t, exe, "verbose").CombinedOutput()
	if err != nil {
		t.Fatalf("native CBS2 verbose report: %v\n%s", err, native)
	}
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, 0)
	Gia_ManAppendCo(g, 1)
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b))
	got := hashReviewStdout(t, func() { Cbs2_ManSolveMiterNc(g, 0, nil, 1) })
	tail := func(report string) string {
		index := strings.Index(report, "Prop1 =")
		if index < 0 {
			t.Fatalf("missing propagation statistics:\n%s", report)
		}
		tail := report[index:]
		timing := regexp.MustCompile(`JFront = +[0-9]+\.[0-9]{2} sec\n$`)
		if !timing.MatchString(tail) {
			t.Fatalf("missing/malformed frontier timing: %q", tail)
		}
		return timing.ReplaceAllString(tail, "JFront = <elapsed>\n")
	}
	if want := tail(string(native)); tail(got) != want {
		t.Fatalf("propagation report differs\nGo:%s\nC:%s", tail(got), want)
	}
}

func TestCircuitSAT2DecisionHelpersMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat2_review.c")
	native, err := testoracle.Command(t, exe, "decisions").CombinedOutput()
	if err != nil {
		t.Fatalf("native CBS2 decision helpers: %v\n%s", err, native)
	}
	g := Gia_ManStart(16)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	ab, bc, cd := Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, b, c), Gia_ManAppendAnd(g, c, d)
	top := Gia_ManAppendAnd(g, ab, bc)
	Gia_ManLevelNum(g)
	p := Cbs2_ManAlloc(g)
	defer Cbs2_ManStop(p)
	for id := range p.vActivity {
		p.vActivity[id] = int32(id % 3)
	}
	got := hashReviewStdout(t, func() {
		for order := 0; order < 2; order++ {
			p.pJust.iHead, p.pJust.iTail = 0, 0
			Cbs2_QuePush(&p.pJust, top>>1)
			p.pJust.iHead = 1
			candidates := []int32{cd, ab, top}
			if order != 0 {
				candidates[0], candidates[1] = candidates[1], candidates[0]
			}
			r, r2, r3 := int32(-1), int32(-1), int32(-1)
			for _, lit := range candidates {
				Cbs2_QuePush(&p.pJust, lit>>1)
				r = Cbs2_ManUpdateDecVar(p, lit>>1, r)
				r2 = Cbs2_ManUpdateDecVar2(p, lit>>1, r2)
				r3 = Cbs2_ManUpdateDecVar3(p, lit>>1, r3)
			}
			fmt.Printf("decisions %d highest%d lowest%d maxff%d refs%d levels%d activity%d\n", order, Cbs2_ManDecideHighest(p), Gia_ObjId(g, Cbs2_ManDecideLowest(p)), Gia_ObjId(g, Cbs2_ManDecideMaxFF(p)), r, r2, r3)
		}
	})
	if got != string(native) {
		t.Fatalf("decision helpers differ\nGo:%s\nC:%s", got, native)
	}
}
