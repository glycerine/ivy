package abc

// Shared checked gate copying for the giaDup.c cofactor paths.

func mappedGiaSignal(mapping []GiaLit, literal GiaLit) (GiaLit, error) {
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) || !mapping[literal.ID()].Valid() {
		return InvalidGiaLit, ErrInvariant
	}
	return mapping[literal.ID()].NotIf(literal.Complemented()), nil
}
func hashGiaCopiedGate(source, target *GiaManager, id GiaID, mapping []GiaLit) (GiaLit, error) {
	obj, err := source.Object(id)
	if err != nil {
		return InvalidGiaLit, err
	}
	if !obj.IsAnd() {
		return InvalidGiaLit, ErrDimension
	}
	first, err := obj.Fanin0(id)
	if err != nil {
		return InvalidGiaLit, err
	}
	a, err := mappedGiaSignal(mapping, first)
	if err != nil {
		return InvalidGiaLit, err
	}
	if obj.IsBuffer() {
		return a, nil
	}
	second, err := obj.Fanin1(id)
	if err != nil {
		return InvalidGiaLit, err
	}
	b, err := mappedGiaSignal(mapping, second)
	if err != nil {
		return InvalidGiaLit, err
	}
	if control, yes := source.MuxControl(id); yes {
		c, err := mappedGiaSignal(mapping, control)
		if err != nil {
			return InvalidGiaLit, err
		}
		return target.HashMux(c, b, a)
	}
	if obj.IsXor() {
		return target.HashXor(a, b)
	}
	return target.HashAnd(a, b)
}

// appendGiaCopiedGate preserves raw node kinds and source fanin polarity.
func appendGiaCopiedGate(source, target *GiaManager, id GiaID, mapping []GiaLit) (GiaLit, error) {
	obj, err := source.Object(id)
	if err != nil {
		return InvalidGiaLit, err
	}
	if !obj.IsAnd() {
		return InvalidGiaLit, ErrDimension
	}
	first, err := obj.Fanin0(id)
	if err != nil {
		return InvalidGiaLit, err
	}
	a, err := mappedGiaSignal(mapping, first)
	if err != nil {
		return InvalidGiaLit, err
	}
	if obj.IsBuffer() {
		return target.AppendBuffer(a)
	}
	second, err := obj.Fanin1(id)
	if err != nil {
		return InvalidGiaLit, err
	}
	b, err := mappedGiaSignal(mapping, second)
	if err != nil {
		return InvalidGiaLit, err
	}
	if control, yes := source.MuxControl(id); yes {
		c, err := mappedGiaSignal(mapping, control)
		if err != nil {
			return InvalidGiaLit, err
		}
		return target.AppendMuxReal(c, b, a)
	}
	if obj.IsXor() {
		return target.AppendXorReal(a, b)
	}
	return target.AppendAnd(a, b)
}
