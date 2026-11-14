package abc

// MappingTiming carries the per-CI arrival and per-CO required times used by
// the FPGA mapper. Defaults cover interface entries without explicit timing.
type MappingTiming struct {
	ArrivalDefault  float32
	RequiredDefault float32
	CIArrivals      []float32
	CORequired      []float32
}

func Abc_NtkSetMappingTiming(n *Network, timing *MappingTiming) {
	if timing == nil {
		n.MappingTiming = nil
		return
	}
	copy := *timing
	copy.CIArrivals = append([]float32(nil), timing.CIArrivals...)
	copy.CORequired = append([]float32(nil), timing.CORequired...)
	n.MappingTiming = &copy
}

func Abc_NtkHasMappingTiming(n *Network) bool { return n != nil && n.MappingTiming != nil }

func Abc_NtkMappingArrivalTimes(n *Network) []float32 {
	if !Abc_NtkHasMappingTiming(n) {
		return nil
	}
	t := n.MappingTiming
	times := make([]float32, len(n.cis))
	for i := range times {
		times[i] = t.ArrivalDefault
		if i < len(t.CIArrivals) {
			times[i] = t.CIArrivals[i]
		}
	}
	return times
}

func Abc_NtkMappingRequiredTimes(n *Network) []float32 {
	if !Abc_NtkHasMappingTiming(n) {
		return nil
	}
	t := n.MappingTiming
	times := make([]float32, len(n.cos))
	for i := range times {
		times[i] = t.RequiredDefault
		if i < len(t.CORequired) {
			times[i] = t.CORequired[i]
		}
	}
	return times
}
