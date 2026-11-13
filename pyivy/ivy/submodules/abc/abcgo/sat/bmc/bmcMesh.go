package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
	"time"
)

const NCPARS = 16

func Bmc_MeshTVar(m [][102]int32, x, y int32) int32 { return m[x][y] }
func Bmc_MeshGVar(m [][102]int32, x, y int32) int32 { return m[x][y] + m[101][100] }
func Bmc_MeshCVar(m [][102]int32, x, y int32) int32 { return m[x][y] + m[101][100] + m[101][101] }
func Bmc_MeshUVar(m [][102]int32, x, y int32) int32 {
	return m[x][y] + m[101][100] + m[101][101] + NCPARS
}
func Bmc_MeshVarValue(s *satoko.Solver, v int32) int32 {
	if s.ModelValue(int(v)) == 1 {
		return 1
	}
	return 0
}
func meshOneHotClauses(model []bool, first, last int32) [][]int {
	if first < 0 || first >= last || int(last) > len(model) {
		panic("bmc: invalid one-hot range")
	}
	var vars []int32
	for i := first; i < last; i++ {
		if model[i] {
			vars = append(vars, i)
		}
	}
	var result [][]int
	for i := 0; i < len(vars); i++ {
		for j := i + 1; j < len(vars); j++ {
			result = append(result, []int{int(vars[i]*2 + 1), int(vars[j]*2 + 1)})
		}
	}
	return result
}
func Bmc_MeshAddOneHotness(s *satoko.Solver, first, last int32) int32 {
	model := make([]bool, s.VariableCount())
	for i := range model {
		model[i] = s.ModelValue(i) == 1
	}
	clauses := meshOneHotClauses(model, first, last)
	for _, c := range clauses {
		if !s.AddClause(c) {
			panic("bmc: inconsistent one-hot clause")
		}
	}
	return int32(len(clauses))
}

// Bmc_MeshTest maps a combinational AND graph to a time-stepped LUT mesh.
func Bmc_MeshTest(p *gia.Gia_Man_t, X, Y, T, verbose int32) {
	start := time.Now()
	if p == nil || X < 3 || Y < 3 || T < 2 || X > 100 || Y > 100 || T > 100 {
		panic("bmc: invalid mesh dimensions")
	}
	inputs, gates := gia.Gia_ManPiNum(p), gia.Gia_ManAndNum(p)
	G := inputs + gates
	if G < 1 || G > 100 || p.RegisterCount() != 0 {
		panic("bmc: invalid mesh graph")
	}
	var nodes [100][2]int32
	for i := int32(0); i < inputs; i++ {
		nodes[i] = [2]int32{-1, -1}
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, _, kind := p.Gate(id)
		if kind == 2 {
			if id > G || a>>1 == 0 || b>>1 == 0 {
				panic("bmc: mesh requires ordered nonconstant AND graph")
			}
			nodes[id-1] = [2]int32{(a >> 1) - 1, (b >> 1) - 1}
		} else if kind >= 3 && kind <= 5 {
			panic("bmc: mesh requires an AND graph")
		}
	}
	if verbose != 0 {
		fmt.Printf("The graph has %d inputs and %d nodes.\n", inputs, gates)
	}
	var me [102][102]int32
	nvars := int32(0)
	for y := int32(0); y < Y; y++ {
		for x := int32(0); x < X; x++ {
			me[x][y] = nvars
			nvars += T + G + NCPARS + 1
		}
	}
	me[101][100], me[101][101] = T, G
	s := satoko.New()
	s.SetVariables(int(nvars))
	nclauses := 0
	add := func(lits ...int32) bool {
		c := make([]int, len(lits))
		for i, l := range lits {
			c[i] = int(l)
		}
		nclauses++
		return s.AddClause(c)
	}
	pos := func(v int32) int32 { return v * 2 }
	neg := func(v int32) int32 { return v*2 + 1 }
	if verbose != 0 {
		fmt.Printf("SAT variable count is %d (%d time vars + %d graph vars + %d config vars + %d aux vars)\n", nvars, X*Y*T, X*Y*G, X*Y*NCPARS, X*Y)
	}
	for x := int32(0); x < X; x++ {
		for y := int32(0); y < Y; y++ {
			tv, gv := Bmc_MeshTVar(me[:], x, y), Bmc_MeshGVar(me[:], x, y)
			if x == 0 || x == X-1 || y == 0 || y == Y-1 {
				for t := int32(0); t < T; t++ {
					lit := neg(tv + t)
					if t == 0 {
						lit = pos(tv)
					}
					if !add(lit) {
						fmt.Println("Problem has no solution.")
						return
					}
				}
				for g := inputs; g < G; g++ {
					if !add(neg(gv + g)) {
						fmt.Println("Problem has no solution.")
						return
					}
				}
			} else if !add(neg(tv)) {
				fmt.Println("Problem has no solution.")
				return
			}
		}
	}
	conf := [12][2]int32{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}, {1, 0}, {2, 0}, {3, 0}, {2, 1}, {3, 1}, {3, 2}}
	for x := int32(1); x < X-1; x++ {
		for y := int32(1); y < Y-1; y++ {
			tv, gv, cv, uv := Bmc_MeshTVar(me[:], x, y), Bmc_MeshGVar(me[:], x, y), Bmc_MeshCVar(me[:], x, y), Bmc_MeshUVar(me[:], x, y)
			ts := [4]int32{Bmc_MeshTVar(me[:], x-1, y), Bmc_MeshTVar(me[:], x, y-1), Bmc_MeshTVar(me[:], x+1, y), Bmc_MeshTVar(me[:], x, y+1)}
			gs := [4]int32{Bmc_MeshGVar(me[:], x-1, y), Bmc_MeshGVar(me[:], x, y-1), Bmc_MeshGVar(me[:], x+1, y), Bmc_MeshGVar(me[:], x, y+1)}
			for g := int32(0); g < G; g++ {
				if !add(neg(gv+g), pos(uv)) {
					fmt.Println("Problem has no solution.")
					return
				}
			}
			clause := []int32{neg(uv)}
			for t := int32(1); t < T; t++ {
				clause = append(clause, pos(tv+t))
			}
			if !add(clause...) {
				fmt.Println("Problem has no solution.")
				return
			}
			clause = []int32{neg(uv)}
			for c := int32(0); c < NCPARS; c++ {
				clause = append(clause, pos(cv+c))
			}
			if !add(clause...) {
				fmt.Println("Problem has no solution.")
				return
			}
			for t := int32(1); t < T; t++ {
				for g := int32(0); g < G; g++ {
					for c := int32(0); c < 4; c++ {
						for _, lit := range []int32{pos(ts[c] + t - 1), pos(gs[c] + g)} {
							if !add(neg(tv+t), neg(gv+g), neg(cv+c), lit) {
								fmt.Println("Problem has no solution.")
								return
							}
						}
					}
				}
				for g := int32(0); g < inputs; g++ {
					for c := int32(4); c < NCPARS; c++ {
						if !add(neg(gv+g), neg(cv+c)) {
							fmt.Println("Problem has no solution.")
							return
						}
					}
				}
				for g := inputs; g < G; g++ {
					for c := int32(0); c < 12; c++ {
						a, b := conf[c][0], conf[c][1]
						for _, lit := range []int32{pos(ts[a] + t - 1), pos(ts[b] + t - 1), pos(gs[a] + nodes[g][0]), pos(gs[b] + nodes[g][1])} {
							if !add(neg(tv+t), neg(gv+g), neg(cv+c+4), lit) {
								fmt.Println("Problem has no solution.")
								return
							}
						}
					}
				}
			}
		}
	}
	if !add(pos(Bmc_MeshGVar(me[:], 1, 1) + G - 1)) {
		fmt.Println("Problem has no solution.")
		return
	}
	if verbose != 0 {
		fmt.Printf("Finished adding %d clauses. Started solving...\n", nclauses)
	}
	var model []bool
	for {
		status := s.Solve()
		if status == satoko.SATOKO_UNSAT {
			fmt.Printf("Problem has no solution. Time %s\n", time.Since(start))
			return
		}
		if status != satoko.SATOKO_SAT {
			fmt.Printf("Computation timed out. Time %s\n", time.Since(start))
			return
		}
		model = make([]bool, nvars)
		for i := range model {
			model[i] = s.ModelValue(i) == 1
		}
		var clauses [][]int
		for x := int32(0); x < X; x++ {
			for y := int32(0); y < Y; y++ {
				gv := Bmc_MeshGVar(me[:], x, y)
				clauses = append(clauses, meshOneHotClauses(model, gv, gv+G)...)
				if x > 0 && x < X-1 && y > 0 && y < Y-1 {
					tv, cv := Bmc_MeshTVar(me[:], x, y), Bmc_MeshCVar(me[:], x, y)
					clauses = append(clauses, meshOneHotClauses(model, tv, tv+T)...)
					clauses = append(clauses, meshOneHotClauses(model, cv, cv+NCPARS)...)
				}
			}
		}
		if len(clauses) == 0 {
			break
		}
		fmt.Printf("Adding %d one-hotness clauses.\n", len(clauses))
		for _, clause := range clauses {
			if !s.AddClause(clause) {
				fmt.Println("Problem has no solution.")
				return
			}
		}
	}
	fmt.Printf("Satisfying solution found. Time %s\n", time.Since(start))
	var buffers, active int32
	for y := int32(1); y < Y-1; y++ {
		for x := int32(1); x < X-1; x++ {
			cv := Bmc_MeshCVar(me[:], x, y)
			for c := int32(0); c < NCPARS; c++ {
				if model[cv+c] {
					if c < 4 {
						buffers++
					} else {
						active++
					}
				}
			}
		}
	}
	fmt.Printf("The %d x %d mesh with latency %d with %d active cells (%d nodes and %d buffers):\n", X, Y, T, buffers+active, active, buffers)
	fmt.Print(" Y\\X ")
	for x := int32(0); x < X; x++ {
		fmt.Printf("  %-2d ", x)
	}
	fmt.Println()
	for y := int32(0); y < Y; y++ {
		fmt.Printf(" %-2d  ", y)
		for x := int32(0); x < X; x++ {
			tv, gv := Bmc_MeshTVar(me[:], x, y), Bmc_MeshGVar(me[:], x, y)
			found := false
			for t := int32(0); t < T; t++ {
				for g := int32(0); g < G; g++ {
					if model[tv+t] && model[gv+g] {
						fmt.Printf(" %c%-2d ", 'a'+g, t)
						found = true
					}
				}
			}
			if !found {
				if x == 0 || x == X-1 || y == 0 || y == Y-1 {
					fmt.Print("  *  ")
				} else {
					fmt.Print("     ")
				}
			}
		}
		fmt.Println()
	}
}
