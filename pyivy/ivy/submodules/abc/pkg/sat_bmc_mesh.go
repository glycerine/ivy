package abc

// Mesh embedding from src/sat/bmc/bmcMesh.c and bmcMesh2.c.

const giaMeshConfigurations = 16

type GiaMeshOptions struct {
	X, Y, Latency                         int
	MaxVariables, MaxClauses, MaxLiterals int
	MaxIterations                         int
	SAT                                   SatLimits
	Stop                                  func() bool
}

type GiaMeshCell struct {
	X, Y, Time, Graph, Configuration int
	Boundary, Used                   bool
}

type GiaMeshPlacement struct {
	Status                              SatStatus
	Reason                              SatStopReason
	Stats                               SatStats
	X, Y, Latency                       int
	InputCount, GraphCount, TargetGraph int
	Variables, Clauses, Literals        int
	Iterations, RefinementRounds        int
	RefinementClauses                   int
	Buffers, Nodes                      int
	Cells                               []GiaMeshCell
}

type giaMeshLayout struct {
	x, latency, graphs, block int
}

func (layout giaMeshLayout) base(x, y int) int { return (y*layout.x + x) * layout.block }
func (layout giaMeshLayout) time(x, y int) int { return layout.base(x, y) }
func (layout giaMeshLayout) graph(x, y int) int {
	return layout.base(x, y) + layout.latency
}
func (layout giaMeshLayout) configuration(x, y int) int {
	return layout.base(x, y) + layout.latency + layout.graphs
}
func (layout giaMeshLayout) used(x, y int) int {
	return layout.configuration(x, y) + giaMeshConfigurations
}

func normalizeGiaMeshOptions(options GiaMeshOptions) (GiaMeshOptions, error) {
	if options.X < 3 || options.X > 100 || options.Y < 3 || options.Y > 100 || options.Latency < 2 || options.Latency > 100 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 || options.MaxIterations < 0 {
		return options, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 4 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 16 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 64 << 20
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 1 << 20
	}
	return options, nil
}

func giaMeshLiteral(variable int, negative bool) SatLit {
	literal := SatLit(variable) << 1
	if negative {
		literal |= 1
	}
	return literal
}

func giaMeshGraph(manager *GiaManager) (inputs int, fanins [][2]int, err error) {
	if manager == nil {
		return 0, nil, ErrDimension
	}
	if err = manager.Check(); err != nil {
		return 0, nil, err
	}
	if manager.registers != 0 || manager.constraints != 0 || manager.PrimaryOutputCount() != 1 {
		return 0, nil, ErrDimension
	}
	inputs = manager.PrimaryInputCount()
	if inputs < 1 {
		return 0, nil, ErrDimension
	}
	for id, object := range manager.objects {
		if object.IsAnd() && (object.IsBuffer() || object.IsXor()) {
			return 0, nil, ErrUnsupported
		}
		if _, mux := manager.MuxControl(GiaID(id)); mux {
			return 0, nil, ErrUnsupported
		}
		if !object.IsAnd() {
			continue
		}
		expected := inputs + 1 + len(fanins)
		if id != expected {
			return 0, nil, ErrUnsupported
		}
		left, leftErr := object.Fanin0(GiaID(id))
		right, rightErr := object.Fanin1(GiaID(id))
		if leftErr != nil || rightErr != nil {
			return 0, nil, ErrInvariant
		}
		if left.ID() == 0 || right.ID() == 0 || int(left.ID()) > inputs+len(fanins) || int(right.ID()) > inputs+len(fanins) {
			return 0, nil, ErrUnsupported
		}
		fanins = append(fanins, [2]int{int(left.ID()) - 1, int(right.ID()) - 1})
	}
	if len(manager.objects) != 1+inputs+len(fanins)+1 || manager.cos[0] != GiaID(len(manager.objects)-1) {
		return 0, nil, ErrUnsupported
	}
	if inputs+len(fanins) > 100 {
		return 0, nil, ErrLimit
	}
	return inputs, fanins, nil
}

func giaMeshAddOneHotness(model []bool, first, last int, add func([]SatLit) error) (int, error) {
	if first < 0 || first >= last || last > len(model) || last-first > 100 {
		return 0, ErrDimension
	}
	selected := make([]int, 0, last-first)
	for variable := first; variable < last; variable++ {
		if model[variable] {
			selected = append(selected, variable)
		}
	}
	count := 0
	for firstIndex := 0; firstIndex < len(selected); firstIndex++ {
		for secondIndex := firstIndex + 1; secondIndex < len(selected); secondIndex++ {
			if err := add([]SatLit{giaMeshLiteral(selected[firstIndex], true), giaMeshLiteral(selected[secondIndex], true)}); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

// EmbedGiaMesh implements the common model in Bmc_MeshTest and
// Bmc_MeshTest2. The C files differ only by SAT backend, so this concrete Go
// operation uses the owned native solver for both. It preserves lazy pairwise
// one-hot refinement and targets graph item G-1 at mesh cell (1,1), even when
// that item is not the sole CO's driver.
func (manager *GiaManager) EmbedGiaMesh(options GiaMeshOptions) (GiaMeshPlacement, error) {
	fail := func(err error) (GiaMeshPlacement, error) { return GiaMeshPlacement{}, err }
	options, err := normalizeGiaMeshOptions(options)
	if err != nil {
		return fail(err)
	}
	inputs, fanins, err := giaMeshGraph(manager)
	if err != nil {
		return fail(err)
	}
	graphs := inputs + len(fanins)
	block, err := checkedSize(options.Latency, 1, graphs+giaMeshConfigurations+1, options.MaxVariables)
	if err != nil {
		return fail(err)
	}
	variables, err := checkedSize(0, options.X*options.Y, block, options.MaxVariables)
	if err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return fail(err)
	}
	if err = solver.SetVariableCount(variables); err != nil {
		return fail(err)
	}
	layout := giaMeshLayout{x: options.X, latency: options.Latency, graphs: graphs, block: block}
	result := GiaMeshPlacement{
		X: options.X, Y: options.Y, Latency: options.Latency,
		InputCount: inputs, GraphCount: graphs, TargetGraph: graphs - 1,
		Variables: variables, Status: SatUnknown,
	}
	rootUnsatisfiable := false
	add := func(clause []SatLit) error {
		if rootUnsatisfiable {
			return nil
		}
		if stopped() {
			return ErrCancelled
		}
		if result.Clauses >= options.MaxClauses || len(clause) > options.MaxLiterals-result.Literals {
			return ErrLimit
		}
		ok, addErr := solver.AddClause(clause)
		if addErr != nil {
			return addErr
		}
		result.Clauses++
		result.Literals += len(clause)
		if !ok {
			rootUnsatisfiable = true
		}
		return nil
	}
	unit := func(variable int, value bool) error {
		return add([]SatLit{giaMeshLiteral(variable, !value)})
	}
	boundary := func(x, y int) bool { return x == 0 || x == options.X-1 || y == 0 || y == options.Y-1 }
	for x := 0; x < options.X; x++ {
		for y := 0; y < options.Y; y++ {
			if boundary(x, y) {
				for time := 0; time < options.Latency; time++ {
					if err = unit(layout.time(x, y)+time, time == 0); err != nil {
						return fail(err)
					}
				}
				for graph := inputs; graph < graphs; graph++ {
					if err = unit(layout.graph(x, y)+graph, false); err != nil {
						return fail(err)
					}
				}
			} else if err = unit(layout.time(x, y), false); err != nil {
				return fail(err)
			}
		}
	}
	configurations := [12][2]int{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}, {1, 0}, {2, 0}, {3, 0}, {2, 1}, {3, 1}, {3, 2}}
	for x := 1; x < options.X-1; x++ {
		for y := 1; y < options.Y-1; y++ {
			neighborX := [4]int{x - 1, x, x + 1, x}
			neighborY := [4]int{y, y - 1, y, y + 1}
			for graph := 0; graph < graphs; graph++ {
				if err = add([]SatLit{giaMeshLiteral(layout.graph(x, y)+graph, true), giaMeshLiteral(layout.used(x, y), false)}); err != nil {
					return fail(err)
				}
			}
			clause := make([]SatLit, 1, options.Latency)
			clause[0] = giaMeshLiteral(layout.used(x, y), true)
			for time := 1; time < options.Latency; time++ {
				clause = append(clause, giaMeshLiteral(layout.time(x, y)+time, false))
			}
			if err = add(clause); err != nil {
				return fail(err)
			}
			clause = make([]SatLit, 1, giaMeshConfigurations+1)
			clause[0] = giaMeshLiteral(layout.used(x, y), true)
			for configuration := 0; configuration < giaMeshConfigurations; configuration++ {
				clause = append(clause, giaMeshLiteral(layout.configuration(x, y)+configuration, false))
			}
			if err = add(clause); err != nil {
				return fail(err)
			}
			for time := 1; time < options.Latency; time++ {
				for graph := 0; graph < graphs; graph++ {
					for configuration := 0; configuration < 4; configuration++ {
						prefix := []SatLit{
							giaMeshLiteral(layout.time(x, y)+time, true),
							giaMeshLiteral(layout.graph(x, y)+graph, true),
							giaMeshLiteral(layout.configuration(x, y)+configuration, true),
						}
						if err = add(append(prefix, giaMeshLiteral(layout.time(neighborX[configuration], neighborY[configuration])+time-1, false))); err != nil {
							return fail(err)
						}
						if err = add(append(prefix, giaMeshLiteral(layout.graph(neighborX[configuration], neighborY[configuration])+graph, false))); err != nil {
							return fail(err)
						}
					}
				}
				for graph := 0; graph < inputs; graph++ {
					for configuration := 4; configuration < giaMeshConfigurations; configuration++ {
						if err = add([]SatLit{giaMeshLiteral(layout.graph(x, y)+graph, true), giaMeshLiteral(layout.configuration(x, y)+configuration, true)}); err != nil {
							return fail(err)
						}
					}
				}
				for graph := inputs; graph < graphs; graph++ {
					fanin := fanins[graph-inputs]
					for configuration, directions := range configurations {
						prefix := []SatLit{
							giaMeshLiteral(layout.time(x, y)+time, true),
							giaMeshLiteral(layout.graph(x, y)+graph, true),
							giaMeshLiteral(layout.configuration(x, y)+configuration+4, true),
						}
						for input := 0; input < 2; input++ {
							direction := directions[input]
							if err = add(append(prefix, giaMeshLiteral(layout.time(neighborX[direction], neighborY[direction])+time-1, false))); err != nil {
								return fail(err)
							}
							if err = add(append(prefix, giaMeshLiteral(layout.graph(neighborX[direction], neighborY[direction])+fanin[input], false))); err != nil {
								return fail(err)
							}
						}
					}
				}
			}
		}
	}
	if err = unit(layout.graph(1, 1)+graphs-1, true); err != nil {
		return fail(err)
	}
	limits := options.SAT
	callerStop := limits.Stop
	limits.Stop = func() bool { return stopped() || callerStop != nil && callerStop() }
	for result.Iterations < options.MaxIterations {
		result.Iterations++
		answer, solveErr := solver.Solve(nil, limits)
		if solveErr != nil {
			return fail(solveErr)
		}
		result.Status, result.Reason, result.Stats = answer.Status, answer.Reason, answer.Stats
		if answer.Status != SatSatisfiable {
			return result, nil
		}
		added := 0
		for x := 0; x < options.X; x++ {
			for y := 0; y < options.Y; y++ {
				groups := [][2]int{{layout.graph(x, y), layout.graph(x, y) + graphs}}
				if !boundary(x, y) {
					groups = [][2]int{
						{layout.time(x, y), layout.time(x, y) + options.Latency},
						{layout.graph(x, y), layout.graph(x, y) + graphs},
						{layout.configuration(x, y), layout.configuration(x, y) + giaMeshConfigurations},
					}
				}
				for _, group := range groups {
					count, refineErr := giaMeshAddOneHotness(answer.Model, group[0], group[1], add)
					if refineErr != nil {
						return fail(refineErr)
					}
					added += count
				}
			}
		}
		if added != 0 {
			result.RefinementRounds++
			result.RefinementClauses += added
			continue
		}
		result.Cells = make([]GiaMeshCell, 0, options.X*options.Y)
		for y := 0; y < options.Y; y++ {
			for x := 0; x < options.X; x++ {
				cell := GiaMeshCell{X: x, Y: y, Time: -1, Graph: -1, Configuration: -1, Boundary: boundary(x, y), Used: answer.Model[layout.used(x, y)]}
				for time := 0; time < options.Latency; time++ {
					if answer.Model[layout.time(x, y)+time] {
						cell.Time = time
					}
				}
				for graph := 0; graph < graphs; graph++ {
					if answer.Model[layout.graph(x, y)+graph] {
						cell.Graph = graph
					}
				}
				if !cell.Boundary {
					for configuration := 0; configuration < giaMeshConfigurations; configuration++ {
						if answer.Model[layout.configuration(x, y)+configuration] {
							cell.Configuration = configuration
							if configuration < 4 {
								result.Buffers++
							} else {
								result.Nodes++
							}
						}
					}
				}
				result.Cells = append(result.Cells, cell)
			}
		}
		return result, nil
	}
	return fail(ErrLimit)
}
