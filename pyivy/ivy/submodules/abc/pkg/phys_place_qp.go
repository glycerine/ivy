package abc

// Quadratic placement solver from src/phys/place/place_qpsolver.h and place_qpsolver.c.

import "math"

const (
	qpsTolerance     = 1e-3
	qpsLoopTolerance = 1e-3
	qpsMaxTolerance  = 0.1
	qpsMinStep       = 1e-6
)

type QPSPoint struct {
	X float64
	Y float64
}

// QPSEdge is one undirected quadratic connection. The source represents the
// same data as sentinel-terminated adjacency lists and retains only j > i.
type QPSEdge struct {
	CellA  int
	CellB  int
	Weight float64
}

// QPSCenterConstraint fixes the area-weighted center of the listed cells.
type QPSCenterConstraint struct {
	Cells  []int
	Center QPSPoint
}

// QPSLoopConstraint bounds the sum of squared lengths around a closed loop.
// Penalty is the initial nonnegative Lagrange multiplier.
type QPSLoopConstraint struct {
	Cells            []int
	MaxSquaredLength float64
	Penalty          float64
}

type QPSProblem struct {
	Positions []QPSPoint
	Fixed     []bool
	Edges     []QPSEdge
	Areas     []float64
	Centers   []QPSCenterConstraint
	Loops     []QPSLoopConstraint
	MaxX      float64
	MaxY      float64
	UseBounds bool
}

type QPSOptions struct {
	MaxCells           int
	MaxEdges           int
	MaxConstraints     int
	MaxOuterIterations int
	MaxInnerIterations int
	Tolerance          float64
	Stop               func() bool
}

type QPSResult struct {
	Positions       []QPSPoint
	Wirelength      float64
	Objective       float64
	OuterIterations int
	InnerIterations int
	LoopPenalties   []float64
	LoopResiduals   []float64
	BoundsSatisfied bool
	LoopsSatisfied  bool
}

type qpsCenter struct {
	cells     []int
	dependent int
	area      float64
	center    QPSPoint
}

// QPSSolver is an owned initialized problem. NewQPSSolver, Solve, and Close
// absorb qps_init, qps_solve, and qps_clean without exposing scratch storage.
type QPSSolver struct {
	problem     QPSProblem
	options     QPSOptions
	centers     []qpsCenter
	independent []int
	variable    []int
	closed      bool
}

func qpsFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func normalizeQPSOptions(options QPSOptions) (QPSOptions, error) {
	if options.MaxCells < 0 || options.MaxEdges < 0 || options.MaxConstraints < 0 ||
		options.MaxOuterIterations < 0 || options.MaxInnerIterations < 0 ||
		options.Tolerance < 0 || !qpsFinite(options.Tolerance) {
		return options, ErrDimension
	}
	if options.MaxCells == 0 {
		options.MaxCells = 1 << 20
	}
	if options.MaxEdges == 0 {
		options.MaxEdges = 1 << 22
	}
	if options.MaxConstraints == 0 {
		options.MaxConstraints = 1 << 20
	}
	if options.MaxOuterIterations == 0 {
		options.MaxOuterIterations = 200
	}
	if options.Tolerance == 0 {
		options.Tolerance = qpsTolerance
	}
	return options, nil
}

func cloneQPSProblem(problem QPSProblem) QPSProblem {
	result := problem
	result.Positions = append([]QPSPoint(nil), problem.Positions...)
	result.Fixed = append([]bool(nil), problem.Fixed...)
	result.Edges = append([]QPSEdge(nil), problem.Edges...)
	result.Areas = append([]float64(nil), problem.Areas...)
	result.Centers = make([]QPSCenterConstraint, len(problem.Centers))
	for i := range problem.Centers {
		result.Centers[i] = problem.Centers[i]
		result.Centers[i].Cells = append([]int(nil), problem.Centers[i].Cells...)
	}
	result.Loops = make([]QPSLoopConstraint, len(problem.Loops))
	for i := range problem.Loops {
		result.Loops[i] = problem.Loops[i]
		result.Loops[i].Cells = append([]int(nil), problem.Loops[i].Cells...)
	}
	return result
}

func NewQPSSolver(problem QPSProblem, options QPSOptions) (*QPSSolver, error) {
	options, err := normalizeQPSOptions(options)
	if err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	n := len(problem.Positions)
	if n == 0 || len(problem.Fixed) != n {
		return nil, ErrDimension
	}
	if n > options.MaxCells || len(problem.Edges) > options.MaxEdges ||
		len(problem.Centers)+len(problem.Loops) > options.MaxConstraints {
		return nil, ErrLimit
	}
	for i := range problem.Positions {
		if !qpsFinite(problem.Positions[i].X) || !qpsFinite(problem.Positions[i].Y) {
			return nil, ErrDimension
		}
	}
	for _, edge := range problem.Edges {
		if edge.CellA < 0 || edge.CellA >= n || edge.CellB < 0 || edge.CellB >= n ||
			edge.CellA == edge.CellB || edge.Weight < 0 || !qpsFinite(edge.Weight) {
			return nil, ErrDimension
		}
	}
	if problem.UseBounds && (problem.MaxX < 0 || problem.MaxY < 0 ||
		!qpsFinite(problem.MaxX) || !qpsFinite(problem.MaxY)) {
		return nil, ErrDimension
	}
	if len(problem.Centers) != 0 && len(problem.Areas) != n {
		return nil, ErrDimension
	}
	owned := cloneQPSProblem(problem)
	solver := &QPSSolver{problem: owned, options: options, variable: make([]int, n)}
	for i := range solver.variable {
		solver.variable[i] = -1
	}
	membership := make([]bool, n)
	dependent := make([]bool, n)
	for _, constraint := range owned.Centers {
		if len(constraint.Cells) == 0 || !qpsFinite(constraint.Center.X) || !qpsFinite(constraint.Center.Y) {
			return nil, ErrDimension
		}
		center := qpsCenter{cells: append([]int(nil), constraint.Cells...), dependent: -1, center: constraint.Center}
		largest := -1.0
		for _, cell := range constraint.Cells {
			if cell < 0 || cell >= n || membership[cell] || owned.Areas[cell] <= 0 || !qpsFinite(owned.Areas[cell]) {
				return nil, ErrDimension
			}
			membership[cell] = true
			center.area += owned.Areas[cell]
			if !owned.Fixed[cell] && owned.Areas[cell] > largest {
				largest = owned.Areas[cell]
				center.dependent = cell
			}
		}
		if center.dependent < 0 {
			return nil, ErrDimension
		}
		dependent[center.dependent] = true
		solver.centers = append(solver.centers, center)
	}
	for _, loop := range owned.Loops {
		if len(loop.Cells) == 0 || loop.MaxSquaredLength < 0 || loop.Penalty < 0 ||
			!qpsFinite(loop.MaxSquaredLength) || !qpsFinite(loop.Penalty) {
			return nil, ErrDimension
		}
		seen := make(map[int]bool, len(loop.Cells))
		for _, cell := range loop.Cells {
			if cell < 0 || cell >= n || seen[cell] {
				return nil, ErrDimension
			}
			seen[cell] = true
		}
	}
	for cell := 0; cell < n; cell++ {
		if !owned.Fixed[cell] && !dependent[cell] {
			solver.variable[cell] = 2 * len(solver.independent)
			solver.independent = append(solver.independent, cell)
		}
	}
	return solver, nil
}

func (solver *QPSSolver) check() error {
	if solver == nil || solver.closed {
		return ErrInvariant
	}
	if solver.options.Stop != nil && solver.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (solver *QPSSolver) Close() error {
	if solver == nil || solver.closed {
		return ErrInvariant
	}
	solver.problem = QPSProblem{}
	solver.centers = nil
	solver.independent = nil
	solver.variable = nil
	solver.closed = true
	return nil
}

func (solver *QPSSolver) setPositions(coordinates []float64, positions []QPSPoint) {
	copy(positions, solver.problem.Positions)
	for _, cell := range solver.independent {
		index := solver.variable[cell]
		positions[cell] = QPSPoint{X: coordinates[index], Y: coordinates[index+1]}
	}
	for _, center := range solver.centers {
		dependent := center.dependent
		x := center.center.X * center.area
		y := center.center.Y * center.area
		for _, cell := range center.cells {
			if cell != dependent {
				x -= solver.problem.Areas[cell] * positions[cell].X
				y -= solver.problem.Areas[cell] * positions[cell].Y
			}
		}
		positions[dependent] = QPSPoint{
			X: x / solver.problem.Areas[dependent],
			Y: y / solver.problem.Areas[dependent],
		}
	}
}

func qpsDistanceSquared(a, b QPSPoint) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}

func (solver *QPSSolver) evaluate(positions []QPSPoint, loopPenalty, lowX, highX, lowY, highY []float64) (float64, float64, []float64) {
	wirelength := 0.0
	for _, edge := range solver.problem.Edges {
		wirelength += edge.Weight * qpsDistanceSquared(positions[edge.CellA], positions[edge.CellB])
	}
	objective := wirelength
	residuals := make([]float64, len(solver.problem.Loops))
	for i, loop := range solver.problem.Loops {
		length := 0.0
		for j, cell := range loop.Cells {
			next := loop.Cells[(j+1)%len(loop.Cells)]
			length += qpsDistanceSquared(positions[cell], positions[next])
		}
		residuals[i] = length - loop.MaxSquaredLength
		objective += loopPenalty[i] * residuals[i]
	}
	if solver.problem.UseBounds {
		for i, point := range positions {
			objective += lowX[i]*(-point.X) + highX[i]*(point.X-solver.problem.MaxX)
			objective += lowY[i]*(-point.Y) + highY[i]*(point.Y-solver.problem.MaxY)
		}
	}
	return objective, wirelength, residuals
}

func (solver *QPSSolver) gradient(positions []QPSPoint, loopPenalty, lowX, highX, lowY, highY []float64) []float64 {
	partials := make([]QPSPoint, len(positions))
	for _, edge := range solver.problem.Edges {
		a, b := edge.CellA, edge.CellB
		dx := 2 * edge.Weight * (positions[a].X - positions[b].X)
		dy := 2 * edge.Weight * (positions[a].Y - positions[b].Y)
		partials[a].X += dx
		partials[a].Y += dy
		partials[b].X -= dx
		partials[b].Y -= dy
	}
	for i, loop := range solver.problem.Loops {
		weight := 2 * loopPenalty[i]
		for j, a := range loop.Cells {
			b := loop.Cells[(j+1)%len(loop.Cells)]
			dx := weight * (positions[a].X - positions[b].X)
			dy := weight * (positions[a].Y - positions[b].Y)
			partials[a].X += dx
			partials[a].Y += dy
			partials[b].X -= dx
			partials[b].Y -= dy
		}
	}
	if solver.problem.UseBounds {
		for i := range partials {
			partials[i].X += highX[i] - lowX[i]
			partials[i].Y += highY[i] - lowY[i]
		}
	}
	gradient := make([]float64, 2*len(solver.independent))
	for _, cell := range solver.independent {
		index := solver.variable[cell]
		gradient[index] += partials[cell].X
		gradient[index+1] += partials[cell].Y
	}
	for _, center := range solver.centers {
		dependent := center.dependent
		for _, cell := range center.cells {
			index := solver.variable[cell]
			if index >= 0 {
				ratio := solver.problem.Areas[cell] / solver.problem.Areas[dependent]
				gradient[index] -= partials[dependent].X * ratio
				gradient[index+1] -= partials[dependent].Y * ratio
			}
		}
	}
	return gradient
}

func (solver *QPSSolver) directionCurvature(direction []float64, loopPenalty []float64) float64 {
	velocity := make([]QPSPoint, len(solver.problem.Positions))
	for _, cell := range solver.independent {
		index := solver.variable[cell]
		velocity[cell] = QPSPoint{X: direction[index], Y: direction[index+1]}
	}
	for _, center := range solver.centers {
		dependent := center.dependent
		for _, cell := range center.cells {
			if cell != dependent {
				ratio := solver.problem.Areas[cell] / solver.problem.Areas[dependent]
				velocity[dependent].X -= ratio * velocity[cell].X
				velocity[dependent].Y -= ratio * velocity[cell].Y
			}
		}
	}
	curvature := 0.0
	for _, edge := range solver.problem.Edges {
		curvature += edge.Weight * qpsDistanceSquared(velocity[edge.CellA], velocity[edge.CellB])
	}
	for i, loop := range solver.problem.Loops {
		for j, cell := range loop.Cells {
			next := loop.Cells[(j+1)%len(loop.Cells)]
			curvature += loopPenalty[i] * qpsDistanceSquared(velocity[cell], velocity[next])
		}
	}
	return curvature
}

func qpsDot(a, b []float64) float64 {
	result := 0.0
	for i := range a {
		result += a[i] * b[i]
	}
	return result
}

func (solver *QPSSolver) minimize(coordinates []float64, positions []QPSPoint, loopPenalty, lowX, highX, lowY, highY []float64) (float64, float64, []float64, int, error) {
	solver.setPositions(coordinates, positions)
	objective, wirelength, residuals := solver.evaluate(positions, loopPenalty, lowX, highX, lowY, highY)
	gradient := solver.gradient(positions, loopPenalty, lowX, highX, lowY, highY)
	direction := make([]float64, len(gradient))
	for i := range gradient {
		direction[i] = -gradient[i]
	}
	norm := qpsDot(direction, direction)
	limit := solver.options.MaxInnerIterations
	if limit == 0 {
		limit = 2 * len(coordinates)
	}
	iterations := 0
	for ; iterations < limit && norm > 0; iterations++ {
		if solver.options.Stop != nil && solver.options.Stop() {
			return 0, 0, nil, iterations, ErrCancelled
		}
		curvature := solver.directionCurvature(direction, loopPenalty)
		if curvature <= math.SmallestNonzeroFloat64 || !qpsFinite(curvature) {
			break
		}
		step := norm / (2 * curvature)
		if !qpsFinite(step) {
			return 0, 0, nil, iterations, ErrInvariant
		}
		for i := range coordinates {
			coordinates[i] += step * direction[i]
		}
		solver.setPositions(coordinates, positions)
		newObjective, newWirelength, newResiduals := solver.evaluate(positions, loopPenalty, lowX, highX, lowY, highY)
		if math.Abs(newObjective-objective) <= (math.Abs(newObjective)+math.Abs(objective)+qpsTolerance*qpsTolerance)*solver.options.Tolerance/2 {
			return newObjective, newWirelength, newResiduals, iterations + 1, nil
		}
		newGradient := solver.gradient(positions, loopPenalty, lowX, highX, lowY, highY)
		newNorm := qpsDot(newGradient, newGradient)
		gamma := newNorm / norm
		for i := range direction {
			direction[i] = -newGradient[i] + gamma*direction[i]
		}
		gradient = newGradient
		norm = newNorm
		objective, wirelength, residuals = newObjective, newWirelength, newResiduals
	}
	return objective, wirelength, residuals, iterations, nil
}

func (solver *QPSSolver) Solve() (QPSResult, error) {
	if err := solver.check(); err != nil {
		return QPSResult{}, err
	}
	n := len(solver.problem.Positions)
	coordinates := make([]float64, 2*len(solver.independent))
	for _, cell := range solver.independent {
		index := solver.variable[cell]
		coordinates[index] = solver.problem.Positions[cell].X
		coordinates[index+1] = solver.problem.Positions[cell].Y
	}
	positions := make([]QPSPoint, n)
	loopPenalty := make([]float64, len(solver.problem.Loops))
	for i := range solver.problem.Loops {
		loopPenalty[i] = solver.problem.Loops[i].Penalty
	}
	lowX, highX := make([]float64, n), make([]float64, n)
	lowY, highY := make([]float64, n), make([]float64, n)
	epsilon := 2.0
	previous := math.Inf(1)
	retries := 0
	totalInner := 0
	for outer := 1; outer <= solver.options.MaxOuterIterations; outer++ {
		objective, wirelength, residuals, inner, err := solver.minimize(coordinates, positions, loopPenalty, lowX, highX, lowY, highY)
		totalInner += inner
		if err != nil {
			return QPSResult{}, err
		}
		loopsDone := true
		denominator := 0.0
		for i, residual := range residuals {
			threshold := qpsLoopTolerance * math.Max(1, solver.problem.Loops[i].MaxSquaredLength)
			if residual > threshold || (loopPenalty[i] > 0 && math.Abs(residual) >= qpsLoopTolerance) {
				loopsDone = false
			}
			if residual > 0 || loopPenalty[i] > 0 {
				denominator += residual * residual
			}
		}
		boundsDone := true
		if solver.problem.UseBounds {
			for i, point := range positions {
				violations := [4]float64{-point.X, point.X - solver.problem.MaxX, -point.Y, point.Y - solver.problem.MaxY}
				multipliers := [4]float64{lowX[i], highX[i], lowY[i], highY[i]}
				for j, violation := range violations {
					denominator += violation * violation
					if violation > solver.options.Tolerance ||
						(outer < 180 && multipliers[j]*violation < -qpsMaxTolerance) {
						boundsDone = false
					}
				}
			}
		}
		result := QPSResult{
			Positions: append([]QPSPoint(nil), positions...), Wirelength: wirelength, Objective: objective,
			OuterIterations: outer, InnerIterations: totalInner,
			LoopPenalties: append([]float64(nil), loopPenalty...), LoopResiduals: append([]float64(nil), residuals...),
			BoundsSatisfied: boundsDone, LoopsSatisfied: loopsDone,
		}
		if loopsDone && boundsDone {
			return result, nil
		}
		if denominator <= math.SmallestNonzeroFloat64 || !qpsFinite(denominator) {
			return QPSResult{}, ErrInvariant
		}
		if objective < previous && previous-objective > 0.01*math.Abs(previous) {
			retries++
			if retries > 2 {
				epsilon /= 2
				retries = 0
			}
		}
		previous = objective
		// Projected dual ascent uses the same residual vector as the source's
		// multiplier update. Normalizing by its Euclidean magnitude avoids the
		// source's division by a guessed optimum, which can be zero for a
		// perfectly anchored unconstrained placement and otherwise stalls.
		step := epsilon / math.Max(1, math.Sqrt(denominator))
		for i, residual := range residuals {
			loopPenalty[i] = math.Max(0, loopPenalty[i]+step*residual)
		}
		if solver.problem.UseBounds {
			for i, point := range positions {
				lowX[i] = math.Max(0, lowX[i]+step*(-point.X))
				highX[i] = math.Max(0, highX[i]+step*(point.X-solver.problem.MaxX))
				lowY[i] = math.Max(0, lowY[i]+step*(-point.Y))
				highY[i] = math.Max(0, highY[i]+step*(point.Y-solver.problem.MaxY))
			}
		}
		if epsilon < qpsMinStep {
			return QPSResult{}, ErrLimit
		}
	}
	return QPSResult{}, ErrLimit
}

func SolveQPS(problem QPSProblem, options QPSOptions) (QPSResult, error) {
	solver, err := NewQPSSolver(problem, options)
	if err != nil {
		return QPSResult{}, err
	}
	defer solver.Close()
	return solver.Solve()
}
