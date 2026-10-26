package abc

// Most-forward GIA retiming from src/aig/gia/giaRetime.c.

import "slices"

type GiaRetimeOptions struct {
	MaxIterations int
	MaxObjects    int
	MaxVisits     int
	Stop          func() bool
}

type GiaRetimeRound struct {
	Fixed, Moves int
	Cut          []GiaID
	CutPhases    []bool
}

type GiaRetimeResult struct {
	Graph   *GiaManager
	Classes []int
	Rounds  []GiaRetimeRound
}

type giaRetimeBudget struct {
	visits int
	limit  int
	stop   func() bool
}

func (b *giaRetimeBudget) spend() error {
	if b.stop != nil && b.stop() {
		return ErrCancelled
	}
	if b.visits == b.limit {
		return ErrLimit
	}
	b.visits++
	return nil
}

func checkGiaRetimeOptions(options GiaRetimeOptions) (GiaRetimeOptions, error) {
	if options.MaxIterations < 0 || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return options, ErrDimension
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 1 << 20
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 26
	}
	return options, nil
}

type giaRetimeAutonomousFrame struct {
	id    GiaID
	stage uint8
}

// giaRetimeAutonomous preserves the source's shared traversal memoization and
// early fanin-zero success, including its deterministic handling of sequential
// cycles. It returns the marks that Gia_ManMarkAutonomous leaves on ROs.
func (m *GiaManager) giaRetimeAutonomous(budget *giaRetimeBudget) ([]bool, error) {
	seen := make([]bool, len(m.objects))
	autonomous := make([]bool, len(m.objects))
	evaluate := func(root GiaID) (bool, error) {
		stack := []giaRetimeAutonomousFrame{{id: root}}
		last := false
		for len(stack) != 0 {
			if err := budget.spend(); err != nil {
				return false, err
			}
			frame := &stack[len(stack)-1]
			object := m.objects[frame.id]
			switch frame.stage {
			case 0:
				if seen[frame.id] {
					last = autonomous[frame.id]
					stack = stack[:len(stack)-1]
					continue
				}
				seen[frame.id] = true
				if frame.id == 0 || object.IsCI() && int(object.Diff1()) < m.PrimaryInputCount() {
					autonomous[frame.id] = true
					last = true
					stack = stack[:len(stack)-1]
					continue
				}
				frame.stage = 1
				var child GiaID
				if object.IsCO() || object.IsAnd() {
					fanin, err := object.Fanin0(frame.id)
					if err != nil {
						return false, err
					}
					child = fanin.ID()
				} else if object.IsCI() {
					index := int(object.Diff1()) - m.PrimaryInputCount()
					if index < 0 || index >= m.registers {
						return false, ErrInvariant
					}
					child = m.cos[m.PrimaryOutputCount()+index]
				} else {
					return false, ErrInvariant
				}
				stack = append(stack, giaRetimeAutonomousFrame{id: child})
			case 1:
				if !object.IsAnd() || last {
					autonomous[frame.id] = last
					stack = stack[:len(stack)-1]
					continue
				}
				frame.stage = 2
				fanin, err := object.Fanin1(frame.id)
				if err != nil {
					return false, err
				}
				stack = append(stack, giaRetimeAutonomousFrame{id: fanin.ID()})
			case 2:
				autonomous[frame.id] = last
				stack = stack[:len(stack)-1]
			}
		}
		return last, nil
	}
	for _, id := range m.cis[m.PrimaryInputCount():] {
		if _, err := evaluate(id); err != nil {
			return nil, err
		}
	}
	result := make([]bool, len(m.objects))
	for _, id := range m.cis[m.PrimaryInputCount():] {
		result[id] = autonomous[id]
	}
	return result, nil
}

func (m *GiaManager) giaRetimeCut(classes []int, budget *giaRetimeBudget) (GiaRetimeRound, []int, error) {
	autonomous, err := m.giaRetimeAutonomous(budget)
	if err != nil {
		return GiaRetimeRound{}, nil, err
	}
	marked := make([]bool, len(m.objects))
	objectClasses := make([]int, len(m.objects))
	for i := range objectClasses {
		objectClasses[i] = -1
	}
	round := GiaRetimeRound{}
	for register, id := range m.cis[m.PrimaryInputCount():] {
		if autonomous[id] {
			marked[id] = true
		} else {
			round.Fixed++
		}
		if classes != nil {
			objectClasses[id] = classes[register]
		}
	}
	for i, object := range m.objects {
		if err = budget.spend(); err != nil {
			return GiaRetimeRound{}, nil, err
		}
		if !object.IsAnd() {
			continue
		}
		id := GiaID(i)
		first, firstErr := object.Fanin0(id)
		if firstErr != nil {
			return GiaRetimeRound{}, nil, firstErr
		}
		second, secondErr := object.Fanin1(id)
		if secondErr != nil {
			return GiaRetimeRound{}, nil, secondErr
		}
		if !marked[first.ID()] || !marked[second.ID()] {
			continue
		}
		if classes != nil && objectClasses[first.ID()] != objectClasses[second.ID()] {
			continue
		}
		marked[id] = true
		round.Moves++
		if classes != nil {
			objectClasses[id] = objectClasses[first.ID()]
		}
	}
	for _, id := range m.cis[m.PrimaryInputCount():] {
		marked[id] = true
	}
	cutSeen := make([]bool, len(m.objects))
	cutClasses := []int(nil)
	if classes != nil {
		cutClasses = make([]int, 0, m.registers)
	}
	add := func(id GiaID) {
		if !marked[id] || cutSeen[id] {
			return
		}
		cutSeen[id] = true
		round.Cut = append(round.Cut, id)
		if classes != nil {
			cutClasses = append(cutClasses, objectClasses[id])
		}
	}
	for i, object := range m.objects {
		if err = budget.spend(); err != nil {
			return GiaRetimeRound{}, nil, err
		}
		if marked[i] || !object.IsAnd() && !object.IsCO() {
			continue
		}
		id := GiaID(i)
		first, firstErr := object.Fanin0(id)
		if firstErr != nil {
			return GiaRetimeRound{}, nil, firstErr
		}
		add(first.ID())
		if object.IsAnd() {
			second, secondErr := object.Fanin1(id)
			if secondErr != nil {
				return GiaRetimeRound{}, nil, secondErr
			}
			add(second.ID())
		}
	}
	return round, cutClasses, nil
}

type giaRetimeCopyFrame struct {
	id   GiaID
	exit bool
}

func (m *GiaManager) giaRetimeCopyRoot(target *GiaManager, mapping []GiaLit, root GiaID, budget *giaRetimeBudget) error {
	stack := []giaRetimeCopyFrame{{id: root}}
	active := make([]bool, len(m.objects))
	for len(stack) != 0 {
		if err := budget.spend(); err != nil {
			return err
		}
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if mapping[frame.id].Valid() {
			continue
		}
		object := m.objects[frame.id]
		if frame.exit {
			first, err := object.Fanin0(frame.id)
			if err != nil {
				return err
			}
			second, err := object.Fanin1(frame.id)
			if err != nil {
				return err
			}
			mapping[frame.id], err = target.HashAnd(mapping[first.ID()].NotIf(first.Complemented()), mapping[second.ID()].NotIf(second.Complemented()))
			if err != nil {
				return err
			}
			active[frame.id] = false
			continue
		}
		if !object.IsAnd() || active[frame.id] {
			return ErrInvariant
		}
		active[frame.id] = true
		first, err := object.Fanin0(frame.id)
		if err != nil {
			return err
		}
		second, err := object.Fanin1(frame.id)
		if err != nil {
			return err
		}
		stack = append(stack, giaRetimeCopyFrame{id: frame.id, exit: true})
		if !mapping[second.ID()].Valid() {
			stack = append(stack, giaRetimeCopyFrame{id: second.ID()})
		}
		if !mapping[first.ID()].Valid() {
			stack = append(stack, giaRetimeCopyFrame{id: first.ID()})
		}
	}
	return nil
}

func (m *GiaManager) giaRetimeDuplicate(round GiaRetimeRound, classes []int, options GiaRetimeOptions, budget *giaRetimeBudget) (*GiaManager, []int, GiaRetimeRound, error) {
	zeroValues, err := m.evaluateValues(make([]bool, len(m.cis)))
	if err != nil {
		return nil, nil, GiaRetimeRound{}, err
	}
	round.CutPhases = make([]bool, len(round.Cut))
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, nil, GiaRetimeRound{}, err
	}
	target.Name, target.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for _, id := range m.cis[:m.PrimaryInputCount()] {
		if err = budget.spend(); err != nil {
			return nil, nil, GiaRetimeRound{}, err
		}
		mapping[id], err = target.AppendCI()
		if err != nil {
			return nil, nil, GiaRetimeRound{}, err
		}
	}
	for index, id := range round.Cut {
		if err = budget.spend(); err != nil {
			return nil, nil, GiaRetimeRound{}, err
		}
		literal, appendErr := target.AppendCI()
		if appendErr != nil {
			return nil, nil, GiaRetimeRound{}, appendErr
		}
		round.CutPhases[index] = zeroValues[id]
		mapping[id] = literal.NotIf(zeroValues[id])
	}
	if err = target.HashStart(); err != nil {
		return nil, nil, GiaRetimeRound{}, err
	}
	for _, co := range m.cos {
		driver, driverErr := m.objects[co].Fanin0(co)
		if driverErr != nil {
			return nil, nil, GiaRetimeRound{}, driverErr
		}
		if err = m.giaRetimeCopyRoot(target, mapping, driver.ID(), budget); err != nil {
			return nil, nil, GiaRetimeRound{}, err
		}
	}
	for _, co := range m.cos[:m.PrimaryOutputCount()] {
		driver, _ := m.objects[co].Fanin0(co)
		if _, err = target.AppendCO(mapping[driver.ID()].NotIf(driver.Complemented())); err != nil {
			return nil, nil, GiaRetimeRound{}, err
		}
	}
	for register, ro := range m.cis[m.PrimaryInputCount():] {
		ri := m.cos[m.PrimaryOutputCount()+register]
		driver, driverErr := m.objects[ri].Fanin0(ri)
		if driverErr != nil {
			return nil, nil, GiaRetimeRound{}, driverErr
		}
		mapping[ro] = mapping[driver.ID()].NotIf(driver.Complemented())
	}
	for _, id := range round.Cut {
		if m.objects[id].IsAnd() {
			mapping[id] = InvalidGiaLit
		}
	}
	for index, id := range round.Cut {
		if err = m.giaRetimeCopyRoot(target, mapping, id, budget); err != nil {
			return nil, nil, GiaRetimeRound{}, err
		}
		if _, err = target.AppendCO(mapping[id].NotIf(round.CutPhases[index])); err != nil {
			return nil, nil, GiaRetimeRound{}, err
		}
	}
	target.HashStop()
	if err = target.SetRegisterCount(len(round.Cut)); err != nil {
		return nil, nil, GiaRetimeRound{}, err
	}
	if err = target.SetConstraintCount(m.constraints); err != nil {
		return nil, nil, GiaRetimeRound{}, err
	}
	clean, _, err := target.Cleanup()
	if err != nil {
		return nil, nil, GiaRetimeRound{}, err
	}
	return clean, classes, round, nil
}

func (m *GiaManager) giaRetimeOnce(classes []int, options GiaRetimeOptions, budget *giaRetimeBudget) (*GiaManager, []int, GiaRetimeRound, error) {
	round, nextClasses, err := m.giaRetimeCut(classes, budget)
	if err != nil {
		return nil, nil, GiaRetimeRound{}, err
	}
	return m.giaRetimeDuplicate(round, nextClasses, options, budget)
}

func checkGiaRetimeGraph(m *GiaManager, classes []int, options GiaRetimeOptions) error {
	if m == nil {
		return ErrDimension
	}
	if err := m.Check(); err != nil {
		return err
	}
	if len(m.objects) > options.MaxObjects {
		return ErrLimit
	}
	if classes != nil && len(classes) != m.registers {
		return ErrDimension
	}
	if m.XorCount() != 0 || m.MuxCount() != 0 {
		return ErrUnsupported
	}
	if options.Stop != nil && options.Stop() {
		return ErrCancelled
	}
	return nil
}

// RetimeForwardOne performs one source cut/duplicate round. Register classes
// are explicit owned data; nil disables class restrictions.
func (m *GiaManager) RetimeForwardOne(classes []int, options GiaRetimeOptions) (GiaRetimeResult, error) {
	options, err := checkGiaRetimeOptions(options)
	if err != nil {
		return GiaRetimeResult{}, err
	}
	if err = checkGiaRetimeGraph(m, classes, options); err != nil {
		return GiaRetimeResult{}, err
	}
	budget := giaRetimeBudget{limit: options.MaxVisits, stop: options.Stop}
	graph, nextClasses, round, err := m.giaRetimeOnce(slices.Clone(classes), options, &budget)
	if err != nil {
		return GiaRetimeResult{}, err
	}
	return GiaRetimeResult{Graph: graph, Classes: nextClasses, Rounds: []GiaRetimeRound{round}}, nil
}

// RetimeForward repeats forward retiming through the first zero-move round or
// MaxIterations, matching Gia_ManRetimeForward while always returning an owned
// graph and typed progress instead of timing stdout.
func (m *GiaManager) RetimeForward(classes []int, options GiaRetimeOptions) (GiaRetimeResult, error) {
	options, err := checkGiaRetimeOptions(options)
	if err != nil {
		return GiaRetimeResult{}, err
	}
	if err = checkGiaRetimeGraph(m, classes, options); err != nil {
		return GiaRetimeResult{}, err
	}
	budget := giaRetimeBudget{limit: options.MaxVisits, stop: options.Stop}
	current := m
	currentClasses := slices.Clone(classes)
	result := GiaRetimeResult{}
	for iteration := 0; iteration < options.MaxIterations; iteration++ {
		graph, nextClasses, round, roundErr := current.giaRetimeOnce(currentClasses, options, &budget)
		if roundErr != nil {
			return GiaRetimeResult{}, roundErr
		}
		result.Rounds = append(result.Rounds, round)
		current, currentClasses = graph, nextClasses
		if round.Moves == 0 {
			break
		}
	}
	result.Graph = current
	result.Classes = currentClasses
	return result, nil
}
