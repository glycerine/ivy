package abc

// Technology-independent mux recognition and MFFC diagnostics from
// src/aig/gia/giaTis.c.

type GiaTisOptions struct {
	MaxObjects int
	MaxVisits  int
	MaxReports int
	Stop       func() bool
}

type GiaTisMFFC struct {
	Root   GiaID
	Nodes  []GiaID // fanin postorder with Root last
	Leaves []GiaID // first-encounter DFS order
}

type GiaTisExperiment struct {
	Graph   *GiaManager
	Reports []GiaTisMFFC
}

func checkGiaTisOptions(options GiaTisOptions) (GiaTisOptions, error) {
	if options.MaxObjects < 0 || options.MaxVisits < 0 || options.MaxReports < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if options.MaxReports == 0 {
		options.MaxReports = 1 << 20
	}
	return options, nil
}

// DuplicateWithRecognizedMuxes converts ordinary complemented-AND mux and XOR
// patterns into native GIA nodes. As in Gia_ManTisDupMuxes, a pattern remains
// an ordinary AND when both of the root's children have multiple references.
// The source is unchanged and dangling logic is removed by cleanup.
func (m *GiaManager) DuplicateWithRecognizedMuxes(options GiaTisOptions) (*GiaManager, error) {
	options, err := checkGiaTisOptions(options)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrDimension
	}
	if err = m.Check(); err != nil {
		return nil, err
	}
	if len(m.objects) > options.MaxObjects || len(m.objects) > options.MaxVisits {
		return nil, ErrLimit
	}
	// The source asserts that no native mux array exists. Native XORs and
	// buffers are also outside this routine's ordinary-AND representation:
	// treating either as an AND would change its Boolean function.
	if m.MuxCount() != 0 || m.XorCount() != 0 || m.BufferCount() != 0 {
		return nil, ErrUnsupported
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	refs, err := m.ReferenceCounts(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target, err := NewGiaManagerWithLimit(m.objectLimit)
	if err != nil {
		return nil, err
	}
	target.Name, target.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	copyLiteral := func(l GiaLit) GiaLit {
		return mapping[l.ID()].NotIf(l.Complemented())
	}
	for _, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = target.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	for i, object := range m.objects {
		id := GiaID(i)
		if !object.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		first, err := object.Fanin0(id)
		if err != nil {
			return nil, err
		}
		second, err := object.Fanin1(id)
		if err != nil {
			return nil, err
		}
		_, _, _, muxPattern, err := m.RecognizeMux(id)
		if err != nil {
			return nil, err
		}
		if !muxPattern || refs[first.ID()] > 1 && refs[second.ID()] > 1 {
			mapping[id], err = target.HashAnd(copyLiteral(first), copyLiteral(second))
		} else if left, right, xorPattern, recognizeErr := m.RecognizeXor(id); recognizeErr != nil {
			return nil, recognizeErr
		} else if xorPattern {
			mapping[id], err = target.HashXorReal(copyLiteral(left), copyLiteral(right))
		} else {
			control, whenTrue, whenFalse, _, recognizeErr := m.RecognizeMux(id)
			if recognizeErr != nil {
				return nil, recognizeErr
			}
			mapping[id], err = target.HashMuxReal(copyLiteral(control), copyLiteral(whenTrue), copyLiteral(whenFalse))
		}
		if err != nil {
			return nil, err
		}
	}
	target.HashStop()
	for _, id := range m.cos {
		if stopped() {
			return nil, ErrCancelled
		}
		driver, driverErr := m.objects[id].Fanin0(id)
		if driverErr != nil {
			return nil, driverErr
		}
		if _, err = target.AppendCO(copyLiteral(driver)); err != nil {
			return nil, err
		}
	}
	if err = target.SetRegisterCount(m.registers); err != nil {
		return nil, err
	}
	// Constraints are manager metadata not consulted by the C experiment,
	// but retaining them makes the owned duplicate safe for sequential use.
	if err = target.SetConstraintCount(m.constraints); err != nil {
		return nil, err
	}
	if stopped() {
		return nil, ErrCancelled
	}
	clean, _, err := target.Cleanup()
	if err != nil {
		return nil, err
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return clean, nil
}

type giaTisFrame struct {
	id       GiaID
	expanded bool
}

func (m *GiaManager) collectTisMFFC(root GiaID, refs []uint64, options GiaTisOptions, visits *int) (GiaTisMFFC, error) {
	result := GiaTisMFFC{Root: root}
	if uint64(root) >= uint64(len(m.objects)) || !m.objects[root].IsAnd() {
		return GiaTisMFFC{}, ErrInvalidLiteral
	}
	if *visits == options.MaxVisits {
		return GiaTisMFFC{}, ErrLimit
	}
	*visits = *visits + 1
	seen := make([]bool, len(m.objects))
	children, count, err := m.crossCutFanins(root)
	if err != nil {
		return GiaTisMFFC{}, err
	}
	stack := make([]giaTisFrame, 0, count)
	for index := count - 1; index >= 0; index-- {
		stack = append(stack, giaTisFrame{id: children[index]})
	}
	for len(stack) != 0 {
		if options.Stop != nil && options.Stop() {
			return GiaTisMFFC{}, ErrCancelled
		}
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if frame.expanded {
			result.Nodes = append(result.Nodes, frame.id)
			continue
		}
		if seen[frame.id] {
			continue
		}
		if *visits == options.MaxVisits {
			return GiaTisMFFC{}, ErrLimit
		}
		*visits = *visits + 1
		seen[frame.id] = true
		object := m.objects[frame.id]
		if refs[frame.id] > 1 || object.IsCI() || frame.id == 0 {
			result.Leaves = append(result.Leaves, frame.id)
			continue
		}
		if !object.IsAnd() {
			return GiaTisMFFC{}, ErrInvariant
		}
		stack = append(stack, giaTisFrame{id: frame.id, expanded: true})
		children, count, err = m.crossCutFanins(frame.id)
		if err != nil {
			return GiaTisMFFC{}, err
		}
		for index := count - 1; index >= 0; index-- {
			stack = append(stack, giaTisFrame{id: children[index]})
		}
	}
	result.Nodes = append(result.Nodes, root)
	return result, nil
}

// CollectTisMFFC returns the exact diagnostic node and leaf orders used by
// Gia_ManTisCollectMffc. The root is always included even when it has multiple
// references. Constants are treated as leaves instead of reaching C's assert.
func (m *GiaManager) CollectTisMFFC(root GiaID, options GiaTisOptions) (GiaTisMFFC, error) {
	options, err := checkGiaTisOptions(options)
	if err != nil {
		return GiaTisMFFC{}, err
	}
	if m == nil {
		return GiaTisMFFC{}, ErrDimension
	}
	if err = m.Check(); err != nil {
		return GiaTisMFFC{}, err
	}
	if len(m.objects) > options.MaxObjects {
		return GiaTisMFFC{}, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return GiaTisMFFC{}, ErrCancelled
	}
	refs, err := m.ReferenceCounts(options.MaxObjects)
	if err != nil {
		return GiaTisMFFC{}, err
	}
	visits := 0
	return m.collectTisMFFC(root, refs, options, &visits)
}

// AnalyzeTis replaces Gia_ManTisTest and Gia_ManTisPrintMffc with an owned
// converted graph and typed reports. It reports every AND-like object whose
// reference count is not one, followed by eligible CO drivers in CO order.
func (m *GiaManager) AnalyzeTis(options GiaTisOptions) (GiaTisExperiment, error) {
	options, err := checkGiaTisOptions(options)
	if err != nil {
		return GiaTisExperiment{}, err
	}
	graph, err := m.DuplicateWithRecognizedMuxes(options)
	if err != nil {
		return GiaTisExperiment{}, err
	}
	refs, err := graph.ReferenceCounts(options.MaxObjects)
	if err != nil {
		return GiaTisExperiment{}, err
	}
	result := GiaTisExperiment{Graph: graph}
	visits := 0
	appendReport := func(root GiaID) error {
		if len(result.Reports) == options.MaxReports {
			return ErrLimit
		}
		report, collectErr := graph.collectTisMFFC(root, refs, options, &visits)
		if collectErr != nil {
			return collectErr
		}
		result.Reports = append(result.Reports, report)
		return nil
	}
	for i, object := range graph.objects {
		if options.Stop != nil && options.Stop() {
			return GiaTisExperiment{}, ErrCancelled
		}
		if object.IsAnd() && refs[i] != 1 {
			if err = appendReport(GiaID(i)); err != nil {
				return GiaTisExperiment{}, err
			}
		}
	}
	for _, id := range graph.cos {
		if options.Stop != nil && options.Stop() {
			return GiaTisExperiment{}, ErrCancelled
		}
		driver, driverErr := graph.objects[id].Fanin0(id)
		if driverErr != nil {
			return GiaTisExperiment{}, driverErr
		}
		// The C diagnostic asserts on a terminal driver. A typed Go report has
		// no gate MFFC to return, so safely omit such outputs.
		if refs[driver.ID()] <= 1 && graph.objects[driver.ID()].IsAnd() {
			if err = appendReport(driver.ID()); err != nil {
				return GiaTisExperiment{}, err
			}
		}
	}
	return result, nil
}
