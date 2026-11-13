package bmc

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Bmc_MeshVarValue2 reads a variable in the solver's current model.
func Bmc_MeshVarValue2(s *bsat.Solver, v int32) int32 {
	value, ok := s.Value(v)
	if !ok {
		panic("bmc: mesh variable has no model")
	}
	if value {
		return 1
	}
	return 0
}

func mesh2OneHot(s *bsat.Solver, model []bool, first, last int32) int32 {
	if first < 0 || first >= last || int(last) > len(model) {
		panic("bmc: invalid mesh variable range")
	}
	var selected []int32
	for v := first; v < last; v++ {
		if model[v] {
			selected = append(selected, v)
		}
	}
	var count int32
	for i, a := range selected {
		for _, b := range selected[i+1:] {
			if !s.AddClause(bsat.Negative(a), bsat.Negative(b)) {
				panic("bmc: inconsistent mesh one-hot constraint")
			}
			count++
		}
	}
	return count
}

// Bmc_MeshAddOneHotness2 adds mutual exclusion for active variables in the model.
func Bmc_MeshAddOneHotness2(s *bsat.Solver, first, last int32) int32 {
	return mesh2OneHot(s, s.Model(), first, last)
}

// Bmc_MeshTest2 encodes a graph placement and iteratively refines one-hot
// constraints until it finds a valid mesh or proves it impossible.
func Bmc_MeshTest2(p *gia.Gia_Man_t, X, Y, T, verbose int32) {
	if p == nil || X < 3 || Y < 3 || T < 1 || X > 100 || Y > 100 || T > 100 {
		panic("bmc: invalid mesh dimensions")
	}
	start := time.Now()
	inputs := p.InputLiterals()
	I := int32(len(inputs))
	type fanins struct{ a, b int32 }
	nodes := make([]fanins, I)
	ids := make(map[int32]int32)
	for i, lit := range inputs {
		ids[lit>>1] = int32(i)
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, _, kind := p.Gate(id)
		if kind == 0 || kind == 1 || kind == 6 {
			continue
		}
		if kind != 2 {
			panic("bmc: mesh requires an AND-inverter graph")
		}
		aa, ok := ids[a>>1]
		if !ok {
			panic("bmc: missing mesh fanin")
		}
		bb, ok := ids[b>>1]
		if !ok {
			panic("bmc: missing mesh fanin")
		}
		ids[id] = int32(len(nodes))
		nodes = append(nodes, fanins{aa, bb})
	}
	G := int32(len(nodes))
	if G == 0 || G > 100 {
		panic("bmc: mesh graph size out of range")
	}
	const configs int32 = 16
	stride := T + G + configs + 1
	base := func(x, y int32) int32 { return (y*X + x) * stride }
	tv := func(x, y int32) int32 { return base(x, y) }
	gv := func(x, y int32) int32 { return base(x, y) + T }
	cv := func(x, y int32) int32 { return base(x, y) + T + G }
	uv := func(x, y int32) int32 { return base(x, y) + T + G + configs }
	s := bsat.NewSolver()
	for i := int32(0); i < X*Y*stride; i++ {
		s.NewVariable()
	}
	if verbose != 0 {
		fmt.Printf("SAT variable count is %d (%d time vars + %d graph vars + %d config vars + %d aux vars)\n", s.NumVars(), X*Y*T, X*Y*G, X*Y*configs, X*Y)
	}
	clause := func(lits ...bsat.Literal) bool { return s.AddClause(lits...) }
	boundary := func(x, y int32) bool { return x == 0 || x == X-1 || y == 0 || y == Y-1 }
	for x := int32(0); x < X; x++ {
		for y := int32(0); y < Y; y++ {
			if boundary(x, y) {
				for t := int32(0); t < T; t++ {
					lit := bsat.Negative(tv(x, y) + t)
					if t == 0 {
						lit = bsat.Positive(tv(x, y) + t)
					}
					if !clause(lit) {
						fmt.Println("Problem has no solution.")
						return
					}
				}
				for g := I; g < G; g++ {
					if !clause(bsat.Negative(gv(x, y) + g)) {
						fmt.Println("Problem has no solution.")
						return
					}
				}
			} else if !clause(bsat.Negative(tv(x, y))) {
				fmt.Println("Problem has no solution.")
				return
			}
		}
	}
	pairs := [12][2]int32{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}, {1, 0}, {2, 0}, {3, 0}, {2, 1}, {3, 1}, {3, 2}}
	for x := int32(1); x < X-1; x++ {
		for y := int32(1); y < Y-1; y++ {
			tim, graph, config, used := tv(x, y), gv(x, y), cv(x, y), uv(x, y)
			times := [4]int32{tv(x-1, y), tv(x, y-1), tv(x+1, y), tv(x, y+1)}
			graphs := [4]int32{gv(x-1, y), gv(x, y-1), gv(x+1, y), gv(x, y+1)}
			for g := int32(0); g < G; g++ {
				clause(bsat.Negative(graph+g), bsat.Positive(used))
			}
			timeClause := []bsat.Literal{bsat.Negative(used)}
			for t := int32(1); t < T; t++ {
				timeClause = append(timeClause, bsat.Positive(tim+t))
			}
			clause(timeClause...)
			confClause := []bsat.Literal{bsat.Negative(used)}
			for c := int32(0); c < configs; c++ {
				confClause = append(confClause, bsat.Positive(config+c))
			}
			clause(confClause...)
			for t := int32(1); t < T; t++ {
				for g := int32(0); g < G; g++ {
					for c := int32(0); c < 4; c++ {
						head := []bsat.Literal{bsat.Negative(tim + t), bsat.Negative(graph + g), bsat.Negative(config + c)}
						clause(append(append([]bsat.Literal(nil), head...), bsat.Positive(times[c]+t-1))...)
						clause(append(append([]bsat.Literal(nil), head...), bsat.Positive(graphs[c]+g))...)
					}
				}
				for g := int32(0); g < I; g++ {
					for c := int32(4); c < configs; c++ {
						clause(bsat.Negative(graph+g), bsat.Negative(config+c))
					}
				}
				for g := I; g < G; g++ {
					for c := int32(0); c < 12; c++ {
						head := []bsat.Literal{bsat.Negative(tim + t), bsat.Negative(graph + g), bsat.Negative(config + c + 4)}
						for _, lit := range []bsat.Literal{bsat.Positive(times[pairs[c][0]] + t - 1), bsat.Positive(times[pairs[c][1]] + t - 1), bsat.Positive(graphs[pairs[c][0]] + nodes[g].a), bsat.Positive(graphs[pairs[c][1]] + nodes[g].b)} {
							clause(append(append([]bsat.Literal(nil), head...), lit)...)
						}
					}
				}
			}
		}
	}
	if !clause(bsat.Positive(gv(1, 1) + G - 1)) {
		fmt.Println("Problem has no solution.")
		return
	}
	var model []bool
	for {
		status := s.Solve()
		if status == bsat.Unsatisfiable {
			fmt.Printf("Problem has no solution. Time: %s\n", time.Since(start))
			return
		}
		if status != bsat.Satisfiable {
			fmt.Printf("Computation timed out. Time: %s\n", time.Since(start))
			return
		}
		model = s.Model()
		var added int32
		for x := int32(0); x < X; x++ {
			for y := int32(0); y < Y; y++ {
				if boundary(x, y) {
					added += mesh2OneHot(s, model, gv(x, y), gv(x, y)+G)
				} else {
					added += mesh2OneHot(s, model, tv(x, y), tv(x, y)+T)
					added += mesh2OneHot(s, model, gv(x, y), gv(x, y)+G)
					added += mesh2OneHot(s, model, cv(x, y), cv(x, y)+configs)
				}
			}
		}
		if added == 0 {
			break
		}
		if verbose != 0 {
			fmt.Printf("Adding %d one-hotness clauses.\n", added)
		}
	}
	fmt.Printf("Satisfying solution found. Time: %s\n", time.Since(start))
	var buffers, nodesCount int32
	for y := int32(1); y < Y-1; y++ {
		for x := int32(1); x < X-1; x++ {
			for c := int32(0); c < configs; c++ {
				if model[cv(x, y)+c] {
					if c < 4 {
						buffers++
					} else {
						nodesCount++
					}
				}
			}
		}
	}
	fmt.Printf("The %d x %d mesh with latency %d with %d active cells (%d nodes and %d buffers):\n", X, Y, T, nodesCount+buffers, nodesCount, buffers)
	fmt.Print(" Y\\X ")
	for x := int32(0); x < X; x++ {
		fmt.Printf("  %-2d ", x)
	}
	fmt.Println()
	for y := int32(0); y < Y; y++ {
		fmt.Printf(" %-2d  ", y)
		for x := int32(0); x < X; x++ {
			found := false
			for t := int32(0); t < T; t++ {
				for g := int32(0); g < G; g++ {
					if model[tv(x, y)+t] && model[gv(x, y)+g] {
						fmt.Printf(" %c%-2d ", 'a'+g, t)
						found = true
					}
				}
			}
			if !found {
				if boundary(x, y) {
					fmt.Print("  *  ")
				} else {
					fmt.Print("     ")
				}
			}
		}
		fmt.Println()
	}
}
