package abc

// Incidence-matrix scheduling and clustering from src/bdd/llb/llb1Matrix.c,
// llb1Sched.c, and llb1Cluster.c. Native AIG, CUDD, image, and reachability
// managers are intentionally outside this standalone boundary.

type LLBPartition struct {
	ID        int
	Variables []int
}

type LLBMatrixSpec struct {
	// Rows are ordered as PI, current-state, internal, next-state. Columns[0]
	// is the current-state boundary and the last column is next-state.
	PrimaryInputs int
	FlipFlops     int
	Rows          int
	Columns       []LLBPartition
}

type LLBLimits struct {
	MaxWork    int64
	MaxObjects int64
	MaxBytes   int64
	Stop       func() bool
}

type LLBIncidenceMatrix struct {
	primaryInputs int
	flipFlops     int
	rows          int
	ids           []int
	columns       [][]int
	present       [][]bool
	rowSums       []int
}

type LLBScheduleResult struct {
	ColumnIDs       []int
	LiveAfter       [][]int
	QuantifiedAfter [][]int
	PeakLive        int
	Work            int64
}

type LLBClusterGroup struct {
	PartitionIDs []int
	Variables    []int
}

type LLBClusterResult struct {
	Groups []LLBClusterGroup
	Work   int64
}

type llbBudget struct {
	limits LLBLimits
	work   int64
}

func (b *llbBudget) take(n int64) error {
	if n < 0 || b.work > int64(^uint64(0)>>1)-n {
		return ErrLimit
	}
	if b.limits.Stop != nil && b.limits.Stop() {
		return ErrCancelled
	}
	b.work += n
	if b.limits.MaxWork != 0 && b.work > b.limits.MaxWork {
		return ErrLimit
	}
	return nil
}

func llbCheckedAdd(a, b int64) (int64, error) {
	if a < 0 || b < 0 || a > int64(^uint64(0)>>1)-b {
		return 0, ErrLimit
	}
	return a + b, nil
}

func llbCheckedMul(a, b int64) (int64, error) {
	if a < 0 || b < 0 || (a != 0 && b > int64(^uint64(0)>>1)/a) {
		return 0, ErrLimit
	}
	return a * b, nil
}

func llbPreflightBytes(limits LLBLimits, amounts ...int64) error {
	if !llbLimitsValid(limits) {
		return ErrDimension
	}
	var total int64
	for _, amount := range amounts {
		var err error
		total, err = llbCheckedAdd(total, amount)
		if err != nil {
			return err
		}
	}
	if limits.MaxBytes != 0 && total > limits.MaxBytes {
		return ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return ErrCancelled
	}
	return nil
}

func llbLimitsValid(limits LLBLimits) bool {
	return limits.MaxWork >= 0 && limits.MaxObjects >= 0 && limits.MaxBytes >= 0
}

func llbPreflightObjects(limits LLBLimits, amounts ...int64) error {
	if !llbLimitsValid(limits) {
		return ErrDimension
	}
	var total int64
	for _, amount := range amounts {
		var err error
		total, err = llbCheckedAdd(total, amount)
		if err != nil {
			return err
		}
	}
	if limits.MaxObjects != 0 && total > limits.MaxObjects {
		return ErrLimit
	}
	return nil
}

func NewLLBIncidenceMatrix(spec LLBMatrixSpec, limits LLBLimits) (*LLBIncidenceMatrix, error) {
	maxInt := int(^uint(0) >> 1)
	if spec.PrimaryInputs < 0 || spec.FlipFlops <= 0 || spec.FlipFlops > (maxInt-spec.PrimaryInputs)/2 || spec.Rows < spec.PrimaryInputs+2*spec.FlipFlops || len(spec.Columns) < 2 || !llbLimitsValid(limits) {
		return nil, ErrDimension
	}
	cellBytes, err := llbCheckedMul(int64(spec.Rows), int64(len(spec.Columns)))
	if err != nil {
		return nil, err
	}
	b := llbBudget{limits: limits}
	var variableCount int64
	for _, column := range spec.Columns {
		if err := b.take(1); err != nil {
			return nil, err
		}
		variableCount, err = llbCheckedAdd(variableCount, int64(len(column.Variables)))
		if err != nil {
			return nil, err
		}
	}
	variableBytes, err := llbCheckedMul(variableCount, 8)
	if err != nil {
		return nil, err
	}
	vectorCount, err := llbCheckedAdd(int64(spec.Rows), int64(len(spec.Columns)))
	if err != nil {
		return nil, err
	}
	vectorBytes, err := llbCheckedMul(vectorCount, 8)
	if err != nil {
		return nil, err
	}
	headerBytes, err := llbCheckedMul(int64(len(spec.Columns)), 48)
	if err != nil {
		return nil, err
	}
	if err := llbPreflightObjects(limits, int64(spec.Rows), int64(len(spec.Columns)), variableCount); err != nil {
		return nil, err
	}
	if err := llbPreflightBytes(limits, 64, cellBytes, variableBytes, vectorBytes, headerBytes); err != nil {
		return nil, err
	}
	m := &LLBIncidenceMatrix{
		primaryInputs: spec.PrimaryInputs,
		flipFlops:     spec.FlipFlops,
		rows:          spec.Rows,
		ids:           make([]int, len(spec.Columns)),
		columns:       make([][]int, len(spec.Columns)),
		present:       make([][]bool, len(spec.Columns)),
		rowSums:       make([]int, spec.Rows),
	}
	for c, column := range spec.Columns {
		if err := b.take(1); err != nil {
			return nil, err
		}
		if column.ID < 0 {
			return nil, ErrDimension
		}
		for prior := 0; prior < c; prior++ {
			if err := b.take(1); err != nil {
				return nil, err
			}
			if m.ids[prior] == column.ID {
				return nil, ErrDimension
			}
		}
		m.ids[c] = column.ID
		m.present[c] = make([]bool, spec.Rows)
		m.columns[c] = make([]int, len(column.Variables))
		last := -1
		for i, variable := range column.Variables {
			if err := b.take(1); err != nil {
				return nil, err
			}
			if variable < 0 || variable >= spec.Rows || variable <= last {
				return nil, ErrDimension
			}
			last = variable
			m.columns[c][i] = variable
			m.present[c][variable] = true
			m.rowSums[variable]++
		}
	}
	firstOK, err := llbExactRangeBudget(m.columns[0], spec.PrimaryInputs, spec.PrimaryInputs+spec.FlipFlops, &b)
	if err != nil {
		return nil, err
	}
	lastOK, err := llbExactRangeBudget(m.columns[len(m.columns)-1], spec.Rows-spec.FlipFlops, spec.Rows, &b)
	if err != nil {
		return nil, err
	}
	if !firstOK || !lastOK {
		return nil, ErrDimension
	}
	for row, sum := range m.rowSums {
		if err := b.take(1); err != nil {
			return nil, err
		}
		if sum == 0 || (row >= spec.PrimaryInputs && row < spec.PrimaryInputs+spec.FlipFlops && sum < 2) {
			return nil, ErrDimension
		}
	}
	return m, nil
}

func llbExactRangeBudget(values []int, begin, end int, b *llbBudget) (bool, error) {
	if len(values) != end-begin {
		return false, nil
	}
	for i, value := range values {
		if err := b.take(1); err != nil {
			return false, err
		}
		if value != begin+i {
			return false, nil
		}
	}
	return true, nil
}

func (m *LLBIncidenceMatrix) Snapshot(limits LLBLimits) (LLBMatrixSpec, error) {
	if m == nil || !llbLimitsValid(limits) {
		return LLBMatrixSpec{}, ErrDimension
	}
	b := llbBudget{limits: limits}
	var variableCount int64
	for _, column := range m.columns {
		if err := b.take(1); err != nil {
			return LLBMatrixSpec{}, err
		}
		var err error
		variableCount, err = llbCheckedAdd(variableCount, int64(len(column)))
		if err != nil {
			return LLBMatrixSpec{}, err
		}
	}
	variableBytes, err := llbCheckedMul(variableCount, 8)
	if err != nil {
		return LLBMatrixSpec{}, err
	}
	headerBytes, err := llbCheckedMul(int64(len(m.columns)), 40)
	if err != nil {
		return LLBMatrixSpec{}, err
	}
	if err := llbPreflightObjects(limits, int64(len(m.columns)), variableCount); err != nil {
		return LLBMatrixSpec{}, err
	}
	if err := llbPreflightBytes(limits, 64, variableBytes, headerBytes); err != nil {
		return LLBMatrixSpec{}, err
	}
	result := LLBMatrixSpec{PrimaryInputs: m.primaryInputs, FlipFlops: m.flipFlops, Rows: m.rows, Columns: make([]LLBPartition, len(m.columns))}
	for i := range m.columns {
		variables := make([]int, len(m.columns[i]))
		for k, variable := range m.columns[i] {
			if err := b.take(1); err != nil {
				return LLBMatrixSpec{}, err
			}
			variables[k] = variable
		}
		result.Columns[i] = LLBPartition{ID: m.ids[i], Variables: variables}
	}
	return result, nil
}

func (m *LLBIncidenceMatrix) VariableOrder(limits LLBLimits) ([]int, error) {
	if m == nil || !llbLimitsValid(limits) {
		return nil, ErrDimension
	}
	bytes, err := llbCheckedMul(int64(m.rows), 16)
	if err != nil {
		return nil, err
	}
	if err := llbPreflightObjects(limits, int64(m.rows), int64(m.rows)); err != nil {
		return nil, err
	}
	if err := llbPreflightBytes(limits, 64, bytes); err != nil {
		return nil, err
	}
	b := llbBudget{limits: limits}
	order := make([]int, m.rows)
	last := make([]int, m.rows)
	for row := 0; row < m.rows; row++ {
		order[row] = row
		for col := len(m.present) - 1; col >= 0; col-- {
			if err := b.take(1); err != nil {
				return nil, err
			}
			if m.present[col][row] {
				last[row] = col
				break
			}
		}
	}
	for i := 1; i < len(order); i++ {
		value, valueLast := order[i], last[i]
		j := i
		for j > 0 {
			if err := b.take(1); err != nil {
				return nil, err
			}
			if last[j-1] <= valueLast {
				break
			}
			order[j], last[j] = order[j-1], last[j-1]
			j--
		}
		order[j], last[j] = value, valueLast
	}
	return order, nil
}

func (m *LLBIncidenceMatrix) Schedule(limits LLBLimits) (LLBScheduleResult, error) {
	var zero LLBScheduleResult
	if m == nil || !llbLimitsValid(limits) {
		return zero, ErrDimension
	}
	cellBytes, err := llbCheckedMul(int64(m.rows), int64(len(m.columns)))
	if err != nil {
		return zero, err
	}
	traceFactor, err := llbCheckedAdd(int64(len(m.columns)), int64(len(m.columns)))
	if err != nil {
		return zero, err
	}
	traceFactor, err = llbCheckedAdd(traceFactor, 3)
	if err != nil {
		return zero, err
	}
	traceBytes, err := llbCheckedMul(int64(m.rows), traceFactor)
	if err != nil {
		return zero, err
	}
	traceBytes, err = llbCheckedMul(traceBytes, 8)
	if err != nil {
		return zero, err
	}
	headerBytes, err := llbCheckedMul(int64(len(m.columns)), 64)
	if err != nil {
		return zero, err
	}
	traceObjects, err := llbCheckedMul(int64(m.rows), traceFactor)
	if err != nil {
		return zero, err
	}
	if err := llbPreflightObjects(limits, cellBytes, traceObjects, int64(len(m.columns))); err != nil {
		return zero, err
	}
	if err := llbPreflightBytes(limits, cellBytes, traceBytes, headerBytes); err != nil {
		return zero, err
	}
	b := llbBudget{limits: limits}
	cols := make([][]bool, len(m.present))
	for i := range cols {
		cols[i] = make([]bool, m.rows)
		for row, value := range m.present[i] {
			if err := b.take(1); err != nil {
				return zero, err
			}
			cols[i][row] = value
		}
	}
	ids := make([]int, len(m.ids))
	for i, id := range m.ids {
		if err := b.take(1); err != nil {
			return zero, err
		}
		ids[i] = id
	}
	live := make([]bool, m.rows)
	remaining := make([]int, m.rows)
	for row, sum := range m.rowSums {
		if err := b.take(1); err != nil {
			return zero, err
		}
		remaining[row] = sum
	}
	for row := m.primaryInputs; row < m.primaryInputs+m.flipFlops; row++ {
		live[row] = true
		remaining[row]--
	}
	result := LLBScheduleResult{ColumnIDs: make([]int, len(ids)), LiveAfter: make([][]int, 0, len(ids)-2), QuantifiedAfter: make([][]int, 0, len(ids)-2)}
	result.ColumnIDs[0] = ids[0]
	result.ColumnIDs[len(ids)-1] = ids[len(ids)-1]
	result.PeakLive = m.flipFlops
	for position := 1; position < len(cols)-1; position++ {
		best, err := llbFindBestColumn(cols, m.rowSums, live, remaining, position, m.rows-m.flipFlops, &b)
		if err != nil {
			return zero, err
		}
		if best < position {
			return zero, ErrDimension
		}
		quantified := make([]int, 0, m.rows)
		for row := 0; row < m.rows; row++ {
			if err := b.take(1); err != nil {
				return zero, err
			}
			if !cols[best][row] {
				continue
			}
			if live[row] && remaining[row] == 1 {
				live[row], remaining[row] = false, 0
				quantified = append(quantified, row)
				continue
			}
			if !live[row] {
				live[row], remaining[row] = true, m.rowSums[row]
			}
			remaining[row]--
			if remaining[row] < 0 {
				return zero, ErrDimension
			}
		}
		cols[position], cols[best] = cols[best], cols[position]
		ids[position], ids[best] = ids[best], ids[position]
		result.ColumnIDs[position] = ids[position]
		liveNow, err := llbTrueIndicesBudget(live, &b)
		if err != nil {
			return zero, err
		}
		result.LiveAfter = append(result.LiveAfter, liveNow)
		result.QuantifiedAfter = append(result.QuantifiedAfter, quantified)
		if len(result.LiveAfter[len(result.LiveAfter)-1]) > result.PeakLive {
			result.PeakLive = len(result.LiveAfter[len(result.LiveAfter)-1])
		}
	}
	result.Work = b.work
	return result, nil
}

func llbFindBestColumn(cols [][]bool, rowSums []int, live []bool, remaining []int, start, nonNextStateRows int, b *llbBudget) (int, error) {
	best, bestWeight := -1, -int(^uint(0)>>1)
	for row := 0; row < nonNextStateRows; row++ {
		if err := b.take(1); err != nil {
			return -1, err
		}
		if rowSums[row] < 2 || !live[row] || remaining[row] != 1 {
			continue
		}
		candidate, count := -1, 0
		for col := start; col < len(cols)-1; col++ {
			if err := b.take(1); err != nil {
				return -1, err
			}
			if cols[col][row] {
				candidate, count = col, count+1
			}
		}
		if count != 1 {
			return -1, ErrDimension
		}
		weight := 0
		for k := range live {
			if err := b.take(1); err != nil {
				return -1, err
			}
			if live[k] && cols[candidate][k] && remaining[k] == 1 {
				weight += 2
			}
			if !live[k] && cols[candidate][k] {
				weight--
			}
		}
		if weight > 0 && bestWeight < weight {
			best, bestWeight = candidate, weight
		}
	}
	if best >= 0 {
		return best, nil
	}
	bestCost, bestSecondary := int(^uint(0)>>1), int(^uint(0)>>1)
	for col := start; col < len(cols)-1; col++ {
		cost, secondary := 0, 0
		for row := range live {
			if err := b.take(1); err != nil {
				return -1, err
			}
			if !live[row] && cols[col][row] {
				cost++
				secondary += remaining[row]
			}
		}
		if cost < bestCost || (cost == bestCost && secondary > bestSecondary) {
			best, bestCost, bestSecondary = col, cost, secondary
		}
	}
	return best, nil
}

func llbTrueIndicesBudget(values []bool, b *llbBudget) ([]int, error) {
	count := 0
	for _, value := range values {
		if err := b.take(1); err != nil {
			return nil, err
		}
		if value {
			count++
		}
	}
	result := make([]int, 0, count)
	for i, value := range values {
		if err := b.take(1); err != nil {
			return nil, err
		}
		if value {
			result = append(result, i)
		}
	}
	return result, nil
}

func (m *LLBIncidenceMatrix) Cluster(clusterMax int, limits LLBLimits) (LLBClusterResult, error) {
	var zero LLBClusterResult
	if m == nil || clusterMax <= 0 || !llbLimitsValid(limits) {
		return zero, ErrDimension
	}
	cellBytes, err := llbCheckedMul(int64(m.rows), int64(len(m.columns)))
	if err != nil {
		return zero, err
	}
	resultFactor, err := llbCheckedAdd(int64(m.rows), int64(len(m.columns)))
	if err != nil {
		return zero, err
	}
	resultBytes, err := llbCheckedMul(resultFactor, int64(len(m.columns)))
	if err != nil {
		return zero, err
	}
	resultBytes, err = llbCheckedMul(resultBytes, 8)
	if err != nil {
		return zero, err
	}
	headerBytes, err := llbCheckedMul(int64(len(m.columns)), 96)
	if err != nil {
		return zero, err
	}
	resultObjects, err := llbCheckedMul(resultFactor, int64(len(m.columns)))
	if err != nil {
		return zero, err
	}
	if err := llbPreflightObjects(limits, cellBytes, resultObjects, int64(m.rows), int64(len(m.columns))); err != nil {
		return zero, err
	}
	if err := llbPreflightBytes(limits, cellBytes, resultBytes, headerBytes); err != nil {
		return zero, err
	}
	b := llbBudget{limits: limits}
	cols := make([][]bool, len(m.present))
	colSums := make([]int, len(cols))
	members := make([][]int, len(cols))
	active := make([]bool, len(cols))
	rowSums := make([]int, m.rows)
	for row, sum := range m.rowSums {
		if err := b.take(1); err != nil {
			return zero, err
		}
		rowSums[row] = sum
	}
	for i := range cols {
		cols[i] = make([]bool, m.rows)
		for row, value := range m.present[i] {
			if err := b.take(1); err != nil {
				return zero, err
			}
			cols[i][row] = value
		}
		colSums[i] = len(m.columns[i])
		members[i] = []int{m.ids[i]}
		active[i] = true
	}
	merge := func(i, k int) error {
		for row := 0; row < m.rows; row++ {
			if err := b.take(1); err != nil {
				return err
			}
			if cols[i][row] && cols[k][row] {
				rowSums[row]--
			}
			if !cols[i][row] && cols[k][row] {
				cols[i][row] = true
				colSums[i]++
			}
		}
		if err := b.take(int64(len(members[k]))); err != nil {
			return err
		}
		members[i] = append(members[i], members[k]...)
		colSums[k] = 0
		active[k] = false
		return nil
	}
	for {
		for {
			i, k, score, err := llbBestQuantPair(cols, colSums, active, rowSums, clusterMax, m.rows-m.flipFlops, &b)
			if err != nil {
				return zero, err
			}
			if i < 0 || score <= 0 {
				break
			}
			if err := merge(i, k); err != nil {
				return zero, err
			}
		}
		i, k, err := llbBestAttractionPair(cols, colSums, active, clusterMax, m.rows-m.flipFlops, &b)
		if err != nil {
			return zero, err
		}
		if i < 0 {
			break
		}
		if err := merge(i, k); err != nil {
			return zero, err
		}
	}
	result := LLBClusterResult{Groups: make([]LLBClusterGroup, 0, len(cols)), Work: b.work}
	for col := range cols {
		if err := b.take(1); err != nil {
			return zero, err
		}
		if !active[col] {
			continue
		}
		variables, err := llbTrueIndicesBudget(cols[col], &b)
		if err != nil {
			return zero, err
		}
		result.Groups = append(result.Groups, LLBClusterGroup{PartitionIDs: append([]int(nil), members[col]...), Variables: variables})
	}
	result.Work = b.work
	return result, nil
}

func llbBestQuantPair(cols [][]bool, colSums []int, active []bool, rowSums []int, clusterMax, nonNextStateRows int, b *llbBudget) (int, int, int64, error) {
	bestI, bestK, bestScore := -1, -1, int64(-100000)
	for i := 1; i < len(cols)-1; i++ {
		if err := b.take(1); err != nil {
			return -1, -1, 0, err
		}
		if !active[i] || colSums[i] == 0 || colSums[i] > clusterMax {
			continue
		}
		for k := i + 1; k < len(cols)-1; k++ {
			if err := b.take(1); err != nil {
				return -1, -1, 0, err
			}
			if !active[k] || colSums[k] == 0 || colSums[k] > clusterMax {
				continue
			}
			var score int64
			for row := 0; row < nonNextStateRows; row++ {
				if err := b.take(1); err != nil {
					return -1, -1, 0, err
				}
				a, c := cols[i][row], cols[k][row]
				if a && c && rowSums[row] == 2 {
					score += 2
				} else if a != c {
					score--
				}
			}
			if score > 0 && score > bestScore {
				bestI, bestK, bestScore = i, k, score
			}
		}
	}
	return bestI, bestK, bestScore, nil
}

func llbBestAttractionPair(cols [][]bool, colSums []int, active []bool, clusterMax, nonNextStateRows int, b *llbBudget) (int, int, error) {
	bestI, bestK, bestDiff, bestTotal := -1, -1, 0, 1
	for i := 1; i < len(cols)-1; i++ {
		if err := b.take(1); err != nil {
			return -1, -1, err
		}
		if !active[i] || colSums[i] == 0 || colSums[i] > clusterMax {
			continue
		}
		for k := i + 1; k < len(cols)-1; k++ {
			if err := b.take(1); err != nil {
				return -1, -1, err
			}
			if !active[k] || colSums[k] == 0 || colSums[k] > clusterMax {
				continue
			}
			common, different := 0, 0
			for row := 0; row < nonNextStateRows; row++ {
				if err := b.take(1); err != nil {
					return -1, -1, err
				}
				a, c := cols[i][row], cols[k][row]
				if a && c {
					common++
				} else if a != c {
					different++
				}
			}
			total := common + different
			if total == 0 {
				// The native float expression is 0/0 for columns containing only
				// next-state variables. Such a pair has no defined attraction.
				continue
			}
			if bestI < 0 || float64(different)/float64(total) < float64(bestDiff)/float64(bestTotal) {
				bestI, bestK, bestDiff, bestTotal = i, k, different, total
			}
		}
	}
	return bestI, bestK, nil
}

func LLBCompiledSources() []string {
	return []string{
		"llb1Cluster.c", "llb1Constr.c", "llb1Core.c", "llb1Group.c", "llb1Hint.c", "llb1Man.c",
		"llb1Matrix.c", "llb1Pivot.c", "llb1Reach.c", "llb1Sched.c", "llb2Bad.c", "llb2Core.c",
		"llb2Driver.c", "llb2Dump.c", "llb2Flow.c", "llb2Image.c", "llb3Image.c", "llb3Nonlin.c",
		"llb4Cex.c", "llb4Image.c", "llb4Nonlin.c", "llb4Sweep.c",
	}
}
