package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"
)

func cbs0ReviewGraph(seed int) (*Gia_Man_t, []int32) {
	g := Gia_ManStart(128)
	Gia_ManHashStart(g)
	lits := make([]int32, 5, 29)
	for i := range lits {
		lits[i] = Gia_ManAppendCi(g)
	}
	if seed < 8 {
		random := uint32(seed + 17)
		for i := 5; i < 29; i++ {
			random = random*1664525 + 1013904223
			a := lits[int(random)%i]
			random = random*1664525 + 1013904223
			a ^= int32(random >> 31)
			random = random*1664525 + 1013904223
			b := lits[int(random)%i]
			random = random*1664525 + 1013904223
			b ^= int32(random >> 31)
			var root int32
			if i%3 == 0 {
				root = Gia_ManHashXor(g, a, b)
			} else {
				root = Gia_ManHashAnd(g, a, b)
			}
			lits = append(lits, root)
		}

	} else {
		v := append([]int32(nil), lits...)
		for i := range v {
			v[i] ^= int32((seed - 8) >> i & 1)
		}
		lits = append(lits, make([]int32, 24)...)
		lits[5] = Gia_ManHashAnd(g, v[0], v[1])
		lits[6] = Gia_ManHashAnd(g, v[0], v[1]^1)
		lits[7] = Gia_ManHashOr(g, lits[5], lits[6])
		lits[8] = Gia_ManHashAnd(g, v[0]^1, lits[7])
		lits[9] = Gia_ManHashXor(g, v[0], v[1])
		lits[10] = Gia_ManHashXor(g, lits[9], v[2])
		lits[11] = Gia_ManHashXor(g, v[1], v[2])
		lits[12] = Gia_ManHashXor(g, v[0], lits[11])
		lits[13] = Gia_ManHashAnd(g, lits[10], lits[12]^1)
		lits[14] = Gia_ManHashXor(g, v[3], v[4])
		lits[15] = Gia_ManHashXor(g, lits[10], lits[14])
		lits[16] = Gia_ManHashXor(g, v[2], lits[14])
		lits[17] = Gia_ManHashXor(g, v[1], lits[16])
		lits[18] = Gia_ManHashXor(g, v[0], lits[17])
		lits[19] = Gia_ManHashAnd(g, lits[15], lits[18]^1)
		lits[20] = Gia_ManHashOr(g, lits[15], lits[18]^1)
		lits[21] = Gia_ManHashAnd(g, lits[13]^1, lits[8]^1)
		lits[22] = Gia_ManHashAnd(g, lits[9], lits[11])
		lits[23] = Gia_ManHashXor(g, lits[22], lits[19])
		lits[24] = Gia_ManHashXor(g, lits[20], lits[9])
		lits[25] = Gia_ManHashAnd(g, lits[24], lits[9])
		lits[26] = Gia_ManHashOr(g, lits[25], lits[22])
		lits[27] = Gia_ManHashAnd(g, lits[26], lits[22]^1)
		lits[28] = Gia_ManHashOr(g, lits[27], lits[8])
	}
	Gia_ManHashStop(g)
	return g, lits
}
func TestChronologicalCircuitSATMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat_old_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native CBS0: %v\n%s", err, native)
	}
	var got strings.Builder
	cases := 0
	var outcomes [3]int
	var maxBT, maxJust int32
	for seed := 0; seed < 12; seed++ {
		g, lits := cbs0ReviewGraph(seed)
		truth := make([]uint32, g.nObjs)
		for mask := 0; mask < 32; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
			values, _, err := g.evaluateObjects(inputs)
			if err != nil {
				t.Fatal(err)
			}
			for id, v := range values {
				if v {
					truth[id] |= 1 << uint(mask)
				}
			}
		}
		for heuristic := 0; heuristic < 3; heuristic++ {
			for _, budget := range []int32{0, 1, 3, 1000} {
				for _, frontier := range []int32{0, 1, 3, 100} {
					p := Cbs0_ManAlloc(g)
					p.Pars.fUseHighest, p.Pars.fUseLowest, p.Pars.fUseMaxFF = giaBool(heuristic == 0), giaBool(heuristic == 1), giaBool(heuristic == 2)
					p.Pars.nBTLimit, p.Pars.nJustLimit = budget, frontier
					for query, lit := range lits {
						if lit >= 2 {
							for phase := int32(0); phase < 2; phase++ {
								root := lit ^ phase
								status := Cbs0_ManSolve(p, Gia_ObjFromLit(g, root))
								outcomes[status+1]++
								maxBT = max(maxBT, p.Pars.nBTThis)
								maxJust = max(maxJust, p.Pars.nJustThis)
								fmt.Fprintf(&got, "case %d %d %d %d %d %d status%d bt%d just%d totals%d:%d model", seed, heuristic, budget, frontier, query, phase, status, p.Pars.nBTThis, p.Pars.nJustThis, p.Pars.nBTTotal, p.Pars.nJustTotal)
								bits := truth[root>>1]
								if root&1 != 0 {
									bits = ^bits
								}
								if status == 0 {
									if bits == 0 {
										t.Fatal("SAT for false root")
									}
									model := Vec_IntArray(Cbs0_ReadModel(p))
									mask, values := 0, 0
									for _, entry := range model {
										if entry < 0 || entry>>1 >= 5 || mask&(1<<uint(entry>>1)) != 0 {
											t.Fatalf("invalid partial cube literal %d", entry)
										}
										fmt.Fprintf(&got, " %d", entry)
										mask |= 1 << uint(entry>>1)
										if entry&1 == 0 {
											values |= 1 << uint(entry>>1)
										}
									}
									for assignment := 0; assignment < 32; assignment++ {
										if assignment&mask == values && bits>>uint(assignment)&1 == 0 {
											t.Fatalf("partial cube does not imply root: seed%d literal%d model%v", seed, root, model)
										}
									}
								} else {
									if Vec_IntSize(p.vModel) != 0 {
										t.Fatal("undecided/UNSAT query retained a stale witness")
									}
									if status == 1 && bits != 0 {
										t.Fatal("UNSAT for satisfiable root")
									}
								}
								for _, obj := range g.pObjs {
									if obj.fMark0 != 0 || obj.fMark1 != 0 {
										t.Fatal("assignment survived a completed solve")
									}
								}
								if p.pProp.iHead != 0 || p.pProp.iTail != 0 || p.pJust.iHead != 0 || p.pJust.iTail != 0 {
									t.Fatal("query did not restore queues")
								}
								got.WriteByte('\n')
								cases++
							}
						}
					}
					Cbs0_ManStop(p)
				}
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(got.String(), "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("CBS0 source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("native output length Go%d C%d", got.Len(), len(native))
	}
	if outcomes[0] == 0 || outcomes[1] == 0 || outcomes[2] == 0 || maxBT < 2 || maxJust < 2 {
		t.Fatalf("weak solver sweep: outcomes%v maxBT%d maxJust%d", outcomes, maxBT, maxJust)
	}
	t.Logf("%d exact native verdict/counter/ordered-model cases, outcomes%v maxBT%d maxJust%d; all partial cubes independently checked", cases, outcomes, maxBT, maxJust)
}

func TestChronologicalMiterMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat_old_review.c")
	native, err := exec.Command(exe, "miter").CombinedOutput()
	if err != nil {
		t.Fatalf("native CBS0 miter: %v\n%s", err, native)
	}
	got := hashReviewStdout(t, func() {
		for _, budget := range []int32{-1, 0, 1000} {
			g := Gia_ManStart(16)
			lit := Gia_ManAppendCi(g)
			for _, root := range []int32{0, 1, lit, lit ^ 1} {
				Gia_ManAppendCo(g, root)
			}
			var statuses *Vec_Str_t
			store := Cbs_ManSolveMiter(g, budget, &statuses, 0)
			fmt.Printf("miter %d statuses", budget)
			for _, status := range Vec_StrArray(statuses) {
				fmt.Printf(" %d", status)
			}
			fmt.Print(" store")
			for _, value := range Vec_IntArray(store) {
				fmt.Printf(" %d", value)
			}
			fmt.Println()
		}
	})
	if got != string(native) {
		t.Fatalf("miter source differs\nGo:%s\nC:%s", got, native)
	}
}

func TestChronologicalHardMitersMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat_old_review.c")
	native, err := exec.Command(exe, "hard-miters").CombinedOutput()
	if err != nil {
		t.Fatalf("native hard miters: %v\n%s", err, native)
	}
	queries := 0
	got := hashReviewStdout(t, func() {
		for seed := 0; seed < 12; seed++ {
			for _, budget := range []int32{-1, 0, 1, 3, 1000} {
				g, lits := cbs0ReviewGraph(seed)
				Gia_ManAppendCo(g, 0)
				Gia_ManAppendCo(g, 1)
				for _, root := range lits {
					Gia_ManAppendCo(g, root)
					Gia_ManAppendCo(g, root^1)
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
					for id, v := range values {
						if v {
							truth[id] |= 1 << uint(mask)
						}
					}
				}
				var statuses *Vec_Str_t
				store := Cbs_ManSolveMiter(g, budget, &statuses, 0)
				cbs0ReviewStore(t, g, statuses, store, truth)
				fmt.Printf("miter %d %d statuses", seed, budget)
				for _, status := range Vec_StrArray(statuses) {
					fmt.Printf(" %d", status)
				}
				fmt.Print(" store")
				for _, value := range Vec_IntArray(store) {
					fmt.Printf(" %d", value)
				}
				fmt.Println()
				queries += Vec_StrSize(statuses)
				for _, obj := range g.pObjs {
					if obj.fMark0 != 0 || obj.fMark1 != 0 {
						t.Fatal("miter retained assignments")
					}
				}
			}
		}
	})
	if got != string(native) {
		a, b := strings.Split(got, "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("miter source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("miter source lengths Go%d C%d", len(got), len(native))
	}
	t.Logf("%d exact source miter property statuses and stores, independently checked across interface permutations", queries)
}

func cbs0ReviewStore(t *testing.T, g *Gia_Man_t, statuses *Vec_Str_t, store *Vec_Int_t, truth []uint32) {
	t.Helper()
	seen := make([]bool, Vec_StrSize(statuses))
	data := Vec_IntArray(store)
	for pos := 0; pos < len(data); {
		if pos+2 > len(data) {
			t.Fatal("short miter store")
		}
		output, count := data[pos], data[pos+1]
		pos += 2
		if output < 0 || int(output) >= len(seen) || seen[output] {
			t.Fatal("invalid/duplicate store output")
		}
		seen[output] = true
		status := Vec_StrEntry(statuses, int(output))
		if status == 255 {
			if count != -1 {
				t.Fatal("timeout store marker")
			}
			continue
		}
		if status != 0 || count < 0 || count > 5 || pos+int(count) > len(data) {
			t.Fatal("invalid model record")
		}
		mask, values := 0, 0
		for _, lit := range data[pos : pos+int(count)] {
			if lit < 0 || lit>>1 >= 5 || mask&(1<<uint(lit>>1)) != 0 {
				t.Fatal("invalid CI model literal")
			}
			mask |= 1 << uint(lit>>1)
			if lit&1 == 0 {
				values |= 1 << uint(lit>>1)
			}
		}
		pos += int(count)
		bits := truth[Gia_ObjId(g, Gia_ManCo(g, output))]
		for assignment := 0; assignment < 32; assignment++ {
			if assignment&mask == values && bits>>uint(assignment)&1 == 0 {
				t.Fatal("miter cube does not imply its property")
			}
		}
	}
	for i, status := range Vec_StrArray(statuses) {
		if status != 0 && status != 1 && status != 255 {
			t.Fatal("invalid miter status")
		}
		if seen[i] != (status != 1) {
			t.Fatal("missing/extraneous miter record")
		}
		if status == 1 && truth[Gia_ObjId(g, Gia_ManCo(g, int32(i)))] != 0 {
			t.Fatal("false miter proof")
		}
	}
}

func TestChronologicalStatisticsMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_sat_old_review.c")
	native, err := exec.Command(exe, "stats").CombinedOutput()
	if err != nil {
		t.Fatalf("native CBS0 statistics: %v\n%s", err, native)
	}
	g := Gia_ManStart(16)
	a := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	got := hashReviewStdout(t, func() {
		p := Cbs0_ManAlloc(g)
		p.nSatTotal, p.nSatUnsat, p.nSatSat, p.nSatUndec = 4, 2, 1, 1
		p.nConfUnsat, p.nConfSat, p.nConfUndec = 5, 3, 7
		p.timeTotal, p.timeSatUnsat, p.timeSatSat, p.timeSatUndec = 4*time.Second, 2*time.Second, time.Second, time.Second/2
		Cbs0_ManSatPrintStats(p)
		Cbs0_ManStop(p)
		p = Cbs0_ManAlloc(g)
		Cbs0_ManSatPrintStats(p)
		Cbs0_ManStop(p)
	})
	if got != string(native) {
		t.Fatalf("source statistics differ\nGo:%s\nC:%s", got, native)
	}
}
