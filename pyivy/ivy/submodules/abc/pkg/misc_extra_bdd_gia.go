package abc

// Plain-AND GIA conversion driver from extraUtilMult.c.

type ExtraBDDFromGiaResult struct {
	Manager     *ExtraBDDManager
	Outputs     []ExtraBDDLit
	SharedNodes int
	Text        []string
}

func BuildExtraBDDFromGia(source *GiaManager, maxNodes int, verbose bool, options ExtraBDDOptions) (ExtraBDDFromGiaResult, error) {
	if source == nil {
		return ExtraBDDFromGiaResult{}, ErrDimension
	}
	if err := source.Check(); err != nil {
		return ExtraBDDFromGiaResult{}, err
	}
	manager, err := NewExtraBDDManager(source.InputCount(), maxNodes)
	if err != nil {
		return ExtraBDDFromGiaResult{}, err
	}
	values := make([]ExtraBDDLit, len(source.objects))
	for index, id := range source.cis {
		values[id], _ = manager.Variable(index)
	}
	for index := 1; index < len(source.objects); index++ {
		id, object := GiaID(index), source.objects[index]
		if object.IsCI() || object.IsCO() {
			continue
		}
		if !object.IsAnd() || object.IsBuffer() || object.IsXor() {
			return ExtraBDDFromGiaResult{}, ErrUnsupported
		}
		if _, mux := source.MuxControl(id); mux {
			return ExtraBDDFromGiaResult{}, ErrUnsupported
		}
		first, firstErr := object.Fanin0(id)
		second, secondErr := object.Fanin1(id)
		if firstErr != nil || secondErr != nil {
			return ExtraBDDFromGiaResult{}, ErrInvariant
		}
		left, right := values[first.ID()], values[second.ID()]
		if first.Complemented() {
			left = left.Not()
		}
		if second.Complemented() {
			right = right.Not()
		}
		values[index], err = manager.And(left, right, options)
		if err != nil {
			return ExtraBDDFromGiaResult{}, err
		}
	}
	result := ExtraBDDFromGiaResult{Manager: manager, Outputs: make([]ExtraBDDLit, len(source.cos))}
	if verbose {
		result.Text = make([]string, len(source.cos))
	}
	for output, id := range source.cos {
		object := source.objects[id]
		driver, driverErr := object.Fanin0(id)
		if driverErr != nil {
			return ExtraBDDFromGiaResult{}, driverErr
		}
		literal := values[driver.ID()]
		if driver.Complemented() {
			literal = literal.Not()
		}
		result.Outputs[output] = literal
		if verbose {
			result.Text[output], err = manager.Format(literal, options)
			if err != nil {
				return ExtraBDDFromGiaResult{}, err
			}
		}
	}
	result.SharedNodes, err = manager.CountSeparateNodes(result.Outputs, options)
	if err != nil {
		return ExtraBDDFromGiaResult{}, err
	}
	return result, nil
}
