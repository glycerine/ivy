package scl

import "math"

type SC_Time struct {
	Lib                                                     *SC_Lib
	CIs, COs, Gates                                         []int32
	Loads, Loads2, Departures, Times, Slews, Times2, Slews2 []SC_Pair
	Slack                                                   []float32
	WireLoad                                                *SC_WireLoad
	SumArea, MaxDelay, InitialArea, InitialDelay, BestDelay float32
}

func Scl_ManAlloc(lib *SC_Lib, cis, cos []int32, objects int) *SC_Time {
	p := &SC_Time{Lib: lib, CIs: cis, COs: cos, Loads: make([]SC_Pair, objects), Loads2: make([]SC_Pair, objects), Departures: make([]SC_Pair, objects), Times: make([]SC_Pair, objects), Slews: make([]SC_Pair, objects), Times2: make([]SC_Pair, objects), Slews2: make([]SC_Pair, objects), Slack: make([]float32, objects)}
	for i := range p.Slack {
		p.Slack[i] = math.Float32frombits(0xffffffff)
	}
	return p
}
func Scl_ManFree(p *SC_Time) {
	if p != nil {
		*p = SC_Time{}
	}
}
func Scl_ManCleanTime(p *SC_Time)                  { clear(p.Departures); clear(p.Times); clear(p.Slews) }
func Scl_ObjCell(p *SC_Time, i int) *SC_Cell       { return p.Lib.Cells[p.Gates[i]] }
func Scl_ObjSetCell(p *SC_Time, i int, c *SC_Cell) { p.Gates[i] = int32(c.ID) }
func Scl_ObjLoad(p *SC_Time, i int) *SC_Pair       { return &p.Loads[i] }
func Scl_ObjLoad2(p *SC_Time, i int) *SC_Pair      { return &p.Loads2[i] }
func Scl_ObjDept(p *SC_Time, i int) *SC_Pair       { return &p.Departures[i] }
func Scl_ObjTime(p *SC_Time, i int) *SC_Pair       { return &p.Times[i] }
func Scl_ObjSlew(p *SC_Time, i int) *SC_Pair       { return &p.Slews[i] }
func Scl_ObjTime2(p *SC_Time, i int) *SC_Pair      { return &p.Times2[i] }
func Scl_ObjSlew2(p *SC_Time, i int) *SC_Pair      { return &p.Slews2[i] }
func Scl_ObjTimeMax(p *SC_Time, i int) float32     { return SC_PairMax(&p.Times[i]) }
func Scl_ObjDepthMax(p *SC_Time, i int) float32    { return SC_PairMax(&p.Departures[i]) }
func Scl_ObjGetSlack(p *SC_Time, i int, d float32) float32 {
	return d - max(p.Times[i].Rise+p.Departures[i].Rise, p.Times[i].Fall+p.Departures[i].Fall)
}
func Scl_ObjGetSlackR(p *SC_Time, i int, d float32) float32 {
	return d - (p.Times[i].Rise + p.Departures[i].Rise)
}
func Scl_ObjGetSlackF(p *SC_Time, i int, d float32) float32 {
	return d - (p.Times[i].Fall + p.Departures[i].Fall)
}
func Scl_ObjSlack(p *SC_Time, i int) float32   { return p.Slack[i] }
func Scl_ObjDupFanin(p *SC_Time, i, fanin int) { p.Times[i] = p.Times[fanin] }
func Scl_ObjGain(p *SC_Time, i int) float32 {
	return .5 * ((p.Times2[i].Rise - p.Times[i].Rise) + (p.Times2[i].Fall - p.Times[i].Fall))
}
func Scl_ObjLegal(p *SC_Time, i int, d float32) bool {
	return p.Times[i].Rise <= p.Times2[i].Rise+Scl_ObjGetSlackR(p, i, d) && p.Times[i].Fall <= p.Times2[i].Fall+Scl_ObjGetSlackF(p, i, d)
}
func Scl_ObjLoadFf(p *SC_Time, i int, rise int32) float64 {
	if rise != 0 {
		return float64(p.Loads[i].Rise)
	}
	return float64(p.Loads[i].Fall)
}
func Scl_ObjTimePs(p *SC_Time, i int, rise int32) float64 {
	if rise != 0 {
		return float64(p.Times[i].Rise)
	}
	return float64(p.Times[i].Fall)
}
func Scl_ObjSlewPs(p *SC_Time, i int, rise int32) float64 {
	if rise != 0 {
		return float64(p.Slews[i].Rise)
	}
	return float64(p.Slews[i].Fall)
}
func Scl_ConeStore(p *SC_Time, cone []int32) {
	for _, i := range cone {
		p.Times2[i] = p.Times[i]
		p.Slews2[i] = p.Slews[i]
	}
}
func Scl_ConeRestore(p *SC_Time, cone []int32) {
	for _, i := range cone {
		p.Times[i] = p.Times2[i]
		p.Slews[i] = p.Slews2[i]
	}
}
func Scl_ConeClear(p *SC_Time, cone []int32) {
	for _, i := range cone {
		p.Times[i] = SC_Pair{}
		p.Slews[i] = SC_Pair{}
	}
}
func Scl_PinTimeArrival(p *SC_Time, t *SC_Timing, i, fanin int) {
	Scl_LibPinArrival(t, &p.Times[fanin], &p.Slews[fanin], &p.Loads[i], &p.Times[i], &p.Slews[i])
}
func Scl_PinTimeDeparture(p *SC_Time, t *SC_Timing, i, fanin int) {
	Scl_LibPinDeparture(t, &p.Departures[fanin], &p.Slews[fanin], &p.Loads[i], &p.Departures[i])
}
