package abc

// Structural fault transforms from src/sat/bmc/bmcFault.c.

import "fmt"

type GiaFaultTransformOptions struct {
	MaxObjects        int
	MaxMappingEntries int
	Stop              func() bool
}

type GiaFaultTransform struct {
	Graph       *GiaManager
	SourceGates []GiaID
	Controls    [][]GiaID
}

type GiaTemporalFaultUnfold struct {
	Graph        *GiaManager
	Frame0ROs    []GiaID
	Frame0PIs    []GiaID
	Frame1PIs    []GiaID
	SourceFaults []GiaID
	Controls     []GiaID
}

func giaFaultLimits(options GiaFaultTransformOptions) (GiaFaultTransformOptions, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	return options, nil
}

func giaFaultOrdinaryGates(source *GiaManager, options GiaFaultTransformOptions) ([]GiaID, error) {
	if source == nil {
		return nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, err
	}
	if source.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	gates := make([]GiaID, 0, source.AndCount())
	for i, object := range source.objects {
		if !object.IsAnd() {
			continue
		}
		id := GiaID(i)
		if object.IsBuffer() || object.IsXor() {
			return nil, fmt.Errorf("%w: fault transform requires ordinary AND gates", ErrDimension)
		}
		if _, yes := source.MuxControl(id); yes {
			return nil, fmt.Errorf("%w: fault transform requires ordinary AND gates", ErrDimension)
		}
		gates = append(gates, id)
	}
	return gates, nil
}

func newGiaFaultGraph(source *GiaManager, gates []GiaID, controlsPerGate int, options GiaFaultTransformOptions) (*GiaFaultTransform, []GiaLit, [][]GiaLit, error) {
	graph, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, nil, nil, err
	}
	graph.Name, graph.Spec = source.Name, source.Spec
	if err = graph.HashStart(); err != nil {
		return nil, nil, nil, err
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	mapping := make([]GiaLit, source.ObjectCount())
	result := &GiaFaultTransform{Graph: graph, SourceGates: append([]GiaID(nil), gates...), Controls: make([][]GiaID, len(gates))}
	controls := make([][]GiaLit, len(gates))
	for i := 0; i < source.PrimaryInputCount(); i++ {
		if stopped() {
			return nil, nil, nil, ErrCancelled
		}
		mapping[source.cis[i]], err = graph.AppendCI()
		if err != nil {
			return nil, nil, nil, err
		}
	}
	for row := range controls {
		controls[row] = make([]GiaLit, controlsPerGate)
		result.Controls[row] = make([]GiaID, controlsPerGate)
		for column := range controls[row] {
			if stopped() {
				return nil, nil, nil, ErrCancelled
			}
			controls[row][column], err = graph.AppendCI()
			if err != nil {
				return nil, nil, nil, err
			}
			result.Controls[row][column] = controls[row][column].ID()
		}
	}
	for i := source.PrimaryInputCount(); i < source.InputCount(); i++ {
		if stopped() {
			return nil, nil, nil, ErrCancelled
		}
		mapping[source.cis[i]], err = graph.AppendCI()
		if err != nil {
			return nil, nil, nil, err
		}
	}
	return result, mapping, controls, nil
}

func finishGiaFaultGraph(source *GiaManager, result *GiaFaultTransform, mapping []GiaLit, options GiaFaultTransformOptions) error {
	for _, id := range source.cos {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		driver, err := source.objects[id].Fanin0(id)
		if err != nil {
			return err
		}
		copied, err := mappedGiaSignal(mapping, driver)
		if err != nil {
			return err
		}
		if _, err = result.Graph.AppendCO(copied); err != nil {
			return err
		}
	}
	if err := result.Graph.SetRegisterCount(source.RegisterCount()); err != nil {
		return err
	}
	if err := result.Graph.SetConstraintCount(source.ConstraintCount()); err != nil {
		return err
	}
	return result.Graph.Check()
}

func UnfoldGiaStuckAt(source *GiaManager, enabled []bool, options GiaFaultTransformOptions) (*GiaFaultTransform, error) {
	options, err := giaFaultLimits(options)
	if err != nil {
		return nil, err
	}
	gates, err := giaFaultOrdinaryGates(source, options)
	if err != nil || len(enabled) != 2*len(gates) {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	result, mapping, controls, err := newGiaFaultGraph(source, gates, 2, options)
	if err != nil {
		return nil, err
	}
	for row, id := range gates {
		value, err := hashGiaCopiedGate(source, result.Graph, id, mapping)
		if err != nil {
			return nil, err
		}
		if enabled[2*row] {
			value, err = result.Graph.HashAnd(controls[row][0].Not(), value)
		}
		if err == nil && enabled[2*row+1] {
			value, err = result.Graph.HashOr(controls[row][1], value)
		}
		if err != nil {
			return nil, err
		}
		mapping[id] = value
	}
	if err = finishGiaFaultGraph(source, result, mapping, options); err != nil {
		return nil, err
	}
	return result, nil
}

func UnfoldGiaFlip(source *GiaManager, enabled []bool, options GiaFaultTransformOptions) (*GiaFaultTransform, error) {
	options, err := giaFaultLimits(options)
	if err != nil {
		return nil, err
	}
	gates, err := giaFaultOrdinaryGates(source, options)
	if err != nil || len(enabled) != len(gates) {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	result, mapping, controls, err := newGiaFaultGraph(source, gates, 1, options)
	if err != nil {
		return nil, err
	}
	for row, id := range gates {
		value, err := hashGiaCopiedGate(source, result.Graph, id, mapping)
		if err == nil && enabled[row] {
			value, err = result.Graph.HashXor(controls[row][0], value)
		}
		if err != nil {
			return nil, err
		}
		mapping[id] = value
	}
	if err = finishGiaFaultGraph(source, result, mapping, options); err != nil {
		return nil, err
	}
	return result, nil
}

func UnfoldGiaFourOutputFunction(source *GiaManager, enabled []bool, options GiaFaultTransformOptions) (*GiaFaultTransform, error) {
	options, err := giaFaultLimits(options)
	if err != nil {
		return nil, err
	}
	gates, err := giaFaultOrdinaryGates(source, options)
	if err != nil || len(enabled) != 4*len(gates) {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	result, mapping, controls, err := newGiaFaultGraph(source, gates, 4, options)
	if err != nil {
		return nil, err
	}
	for row, id := range gates {
		object := source.objects[id]
		first, _ := object.Fanin0(id)
		second, _ := object.Fanin1(id)
		truth := [4]GiaLit{GiaConst0, GiaConst0, GiaConst0, GiaConst0}
		for k := range truth {
			if enabled[4*row+k] {
				truth[k] = controls[row][k]
			}
		}
		nominal := 3
		if first.Complemented() {
			nominal -= 1
		}
		if second.Complemented() {
			nominal -= 2
		}
		truth[nominal] = truth[nominal].Not()
		regularFirst := mapping[first.ID()]
		regularSecond := mapping[second.ID()]
		low, err := result.Graph.HashMux(regularFirst, truth[1], truth[0])
		if err != nil {
			return nil, err
		}
		high, err := result.Graph.HashMux(regularFirst, truth[3], truth[2])
		if err != nil {
			return nil, err
		}
		mapping[id], err = result.Graph.HashMux(regularSecond, high, low)
		if err != nil {
			return nil, err
		}
	}
	if err = finishGiaFaultGraph(source, result, mapping, options); err != nil {
		return nil, err
	}
	return result, nil
}

func CofactorGiaFaultInputs(source *GiaManager, values []bool, options GiaFaultTransformOptions) (*GiaManager, error) {
	options, err := giaFaultLimits(options)
	if err != nil || source == nil || len(values) > source.PrimaryInputCount() {
		return nil, ErrDimension
	}
	if err = source.Check(); err != nil {
		return nil, err
	}
	if source.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	graph, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	graph.Name, graph.Spec = source.Name, source.Spec
	if err = graph.HashStart(); err != nil {
		return nil, err
	}
	mapping := make([]GiaLit, source.ObjectCount())
	for i, id := range source.cis {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		mapping[id], err = graph.AppendCI()
		if err != nil {
			return nil, err
		}
		if i < len(values) {
			mapping[id] = GiaConst0.NotIf(values[i])
		}
	}
	for i, object := range source.objects {
		if !object.IsAnd() {
			continue
		}
		mapping[i], err = hashGiaCopiedGate(source, graph, GiaID(i), mapping)
		if err != nil {
			return nil, err
		}
	}
	result := &GiaFaultTransform{Graph: graph}
	if err = finishGiaFaultGraph(source, result, mapping, options); err != nil {
		return nil, err
	}
	return graph, nil
}

func UnfoldGiaTemporalFault(source *GiaManager, useMuxes, onlyRegisters bool, options GiaFaultTransformOptions) (*GiaTemporalFaultUnfold, error) {
	options, err := giaFaultLimits(options)
	if err != nil {
		return nil, err
	}
	gates, err := giaFaultOrdinaryGates(source, options)
	if err != nil {
		return nil, err
	}
	faults := gates
	if onlyRegisters {
		faults = make([]GiaID, source.RegisterCount())
		for i := range faults {
			faults[i] = source.cos[source.PrimaryOutputCount()+i]
		}
	}
	graph, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	graph.Name, graph.Spec = source.Name, source.Spec
	if err = graph.HashStart(); err != nil {
		return nil, err
	}
	result := &GiaTemporalFaultUnfold{Graph: graph, SourceFaults: append([]GiaID(nil), faults...)}
	appendInputs := func(count int, destination *[]GiaID) ([]GiaLit, error) {
		literals := make([]GiaLit, count)
		for i := range literals {
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			literals[i], err = graph.AppendCI()
			if err != nil {
				return nil, err
			}
			*destination = append(*destination, literals[i].ID())
		}
		return literals, nil
	}
	ros0, err := appendInputs(source.RegisterCount(), &result.Frame0ROs)
	if err != nil {
		return nil, err
	}
	pis0, err := appendInputs(source.PrimaryInputCount(), &result.Frame0PIs)
	if err != nil {
		return nil, err
	}
	pis1, err := appendInputs(source.PrimaryInputCount(), &result.Frame1PIs)
	if err != nil {
		return nil, err
	}
	controls, err := appendInputs(len(faults), &result.Controls)
	if err != nil {
		return nil, err
	}
	copyFrame := func(mapping []GiaLit) error {
		for _, id := range gates {
			mapping[id], err = hashGiaCopiedGate(source, graph, id, mapping)
			if err != nil {
				return err
			}
		}
		return nil
	}
	mapping0 := make([]GiaLit, source.ObjectCount())
	for i := 0; i < source.PrimaryInputCount(); i++ {
		mapping0[source.cis[i]] = pis0[i]
	}
	for i := 0; i < source.RegisterCount(); i++ {
		mapping0[source.cis[source.PrimaryInputCount()+i]] = ros0[i]
	}
	if err = copyFrame(mapping0); err != nil {
		return nil, err
	}
	co0 := make([]GiaLit, source.OutputCount())
	for i, id := range source.cos {
		driver, _ := source.objects[id].Fanin0(id)
		co0[i], err = mappedGiaSignal(mapping0, driver)
		if err != nil {
			return nil, err
		}
	}
	mapping1 := make([]GiaLit, source.ObjectCount())
	for i := 0; i < source.PrimaryInputCount(); i++ {
		mapping1[source.cis[i]] = pis1[i]
	}
	for i := 0; i < source.RegisterCount(); i++ {
		mapping1[source.cis[source.PrimaryInputCount()+i]] = co0[source.PrimaryOutputCount()+i]
	}
	for row, id := range gates {
		current, err := hashGiaCopiedGate(source, graph, id, mapping1)
		if err != nil {
			return nil, err
		}
		if !onlyRegisters && useMuxes {
			current, err = graph.HashMux(controls[row], mapping0[id], current)
			if err != nil {
				return nil, err
			}
		}
		mapping1[id] = current
	}
	for i, id := range source.cos {
		driver, _ := source.objects[id].Fanin0(id)
		current, err := mappedGiaSignal(mapping1, driver)
		if err != nil {
			return nil, err
		}
		if onlyRegisters && i >= source.PrimaryOutputCount() && useMuxes {
			current, err = graph.HashMux(controls[i-source.PrimaryOutputCount()], co0[i], current)
			if err != nil {
				return nil, err
			}
		}
		if _, err = graph.AppendCO(current); err != nil {
			return nil, err
		}
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	return result, nil
}
