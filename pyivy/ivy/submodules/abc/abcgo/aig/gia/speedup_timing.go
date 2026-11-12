package gia

import "math"

// LUTPinTiming is the pin-delay interface supplied by a LUT library.
type LUTPinTiming interface {
	VariablePinDelays() bool
	PinDelay(inputs, pin int32) float32
}

func Gia_LutDelayTraceSortPins(p *Gia_Man_t, id int32, permutation []int32, delays []float32) {
	fans := Gia_ObjLutFanins(p, id)
	if len(permutation) < len(fans) || len(delays) < len(fans) {
		panic("gia: short LUT pin buffers")
	}
	for i, fan := range fans {
		permutation[i] = int32(i)
		delays[i] = Gia_ObjTimeArrival(p, fan)
	}
	for i := 0; i < len(fans)-1; i++ {
		best := i
		for j := i + 1; j < len(fans); j++ {
			if delays[permutation[j]] > delays[permutation[best]] {
				best = j
			}
		}
		permutation[i], permutation[best] = permutation[best], permutation[i]
	}
}
func Gia_LutWhereIsPin(p *Gia_Man_t, fanout, fanin int32, permutation []int32) int32 {
	fans := Gia_ObjLutFanins(p, fanout)
	for i := range fans {
		if fans[permutation[i]] == fanin {
			return int32(i)
		}
	}
	return -1
}
func giaLutTiming(p *Gia_Man_t) LUTPinTiming {
	if p.pLutLib == nil {
		return nil
	}
	lib, ok := p.pLutLib.(LUTPinTiming)
	if !ok {
		panic("gia: incompatible LUT timing library")
	}
	return lib
}
func giaLutPinAssignments(p *Gia_Man_t, id, sortPins int32) ([]int32, []float32) {
	fans := Gia_ObjLutFanins(p, id)
	order := make([]int32, len(fans))
	delays := make([]float32, len(fans))
	for i := range order {
		order[i] = int32(i)
	}
	lib := giaLutTiming(p)
	if lib != nil && lib.VariablePinDelays() && sortPins != 0 {
		arrivals := make([]float32, len(fans))
		Gia_LutDelayTraceSortPins(p, id, order, arrivals)
	}
	for i := range delays {
		delays[i] = 1
		if lib != nil {
			delays[i] = lib.PinDelay(int32(len(fans)), int32(i))
		}
	}
	return order, delays
}
func Gia_ObjComputeArrival(p *Gia_Man_t, id, sortPins int32) float32 {
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsCi(obj) != 0 {
		return Gia_ObjTimeArrival(p, id)
	}
	if Gia_ObjIsCo(obj) != 0 {
		return Gia_ObjTimeArrival(p, Gia_ObjFaninId0(obj, id))
	}
	if Gia_ObjIsLut(p, id) == 0 {
		panic("gia: arrival requires LUT")
	}
	fans := Gia_ObjLutFanins(p, id)
	if len(fans) == 0 {
		return 0
	}
	order, delays := giaLutPinAssignments(p, id, sortPins)
	arrival := float32(math.Inf(-1))
	for k, index := range order {
		arrival = max(arrival, Gia_ObjTimeArrival(p, fans[index])+delays[k])
	}
	return arrival
}
func Gia_ObjPropagateRequired(p *Gia_Man_t, id, sortPins int32) float32 {
	if Gia_ObjIsLut(p, id) == 0 {
		panic("gia: required time requires LUT")
	}
	fans := Gia_ObjLutFanins(p, id)
	order, delays := giaLutPinAssignments(p, id, sortPins)
	var required float32
	for k, index := range order {
		required = Gia_ObjTimeRequired(p, id) - delays[k]
		fan := fans[index]
		if Gia_ObjTimeRequired(p, fan) > required {
			Gia_ObjSetTimeRequired(p, fan, required)
		}
	}
	return required
}
