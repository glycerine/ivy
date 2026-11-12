// Package scl implements ABC's standard-cell timing library representation.
package scl

import "math"

const ABC_SCL_CUR_VERSION = 8

type SC_Dir int32

const (
	ScDirNull SC_Dir = iota
	ScDirInput
	ScDirOutput
	ScDirInOut
	ScDirInternal
)

type SC_TSense int32

const (
	ScTSNull SC_TSense = iota
	ScTSPos
	ScTSNeg
	ScTSNon
)

type SC_Pair struct{ Rise, Fall float32 }
type SC_PairI struct{ Rise, Fall int32 }
type SC_WireLoad struct {
	Name       string
	Cap, Slope float32
	Fanout     []int32
	Length     []float32
}
type SC_WireLoadSel struct {
	Name             string
	AreaFrom, AreaTo []float32
	Models           []string
}
type SC_TableTempl struct {
	Name      string
	Variables []string
	Index     [][]float32
}
type SC_Surface struct {
	Name             string
	Index0, Index1   []float32
	Data             [][]float32
	Index0I, Index1I []int32
	DataI            [][]int32
	Approx           [3][6]float32
}
type SC_Timing struct {
	RelatedPin                               string
	Sense                                    SC_TSense
	When                                     string
	CellRise, CellFall, RiseTrans, FallTrans SC_Surface
}
type SC_Timings struct {
	Name    string
	Timings []*SC_Timing
}
type SC_Pin struct {
	Name                  string
	Direction             SC_Dir
	Cap, RiseCap, FallCap float32
	RiseCapI, FallCapI    int32
	MaxOutCap, MaxOutSlew float32
	Function              string
	Truth                 []uint64
	Timings               []*SC_Timings
}
type SC_Cell struct {
	Name                          string
	ID                            int
	Skip, Sequential, Unsupported bool
	Area, Leakage                 float32
	AreaI, LeakageI               int32
	DriveStrength                 int
	Pins                          []*SC_Pin
	Inputs, Outputs               int
	Next, Prev, Repr, Average     *SC_Cell
	Order, Gates                  int
}
type SC_Lib struct {
	Name, FileName, DefaultWireLoad, DefaultWireLoadSelection string
	DefaultMaxOutSlew                                         float32
	UnitTime                                                  int
	UnitCapMultiplier                                         float32
	UnitCapExponent                                           int
	WireLoads                                                 []*SC_WireLoad
	WireLoadSelections                                        []*SC_WireLoadSel
	Templates                                                 []*SC_TableTempl
	Cells, CellClasses                                        []*SC_Cell
	cellIndex                                                 map[string]int
}

func SC_PairClean(d *SC_Pair)         { *d = SC_Pair{} }
func SC_PairMax(d *SC_Pair) float32   { return max(d.Rise, d.Fall) }
func SC_PairMin(d *SC_Pair) float32   { return min(d.Rise, d.Fall) }
func SC_PairAve(d *SC_Pair) float32   { return .5*d.Rise + .5*d.Fall }
func SC_PairDup(d, s *SC_Pair)        { *d = *s }
func SC_PairMove(d, s *SC_Pair)       { *d = *s; *s = SC_Pair{} }
func SC_PairAdd(d, s *SC_Pair)        { d.Rise += s.Rise; d.Fall += s.Fall }
func SC_PairEqual(d, s *SC_Pair) bool { return *d == *s }
func SC_PairEqualE(d, s *SC_Pair, e float32) bool {
	return float32(math.Abs(float64(d.Rise-s.Rise))) < e && float32(math.Abs(float64(d.Fall-s.Fall))) < e
}
func SC_LibCellNum(p *SC_Lib) int             { return len(p.Cells) }
func SC_LibCell(p *SC_Lib, i int) *SC_Cell    { return p.Cells[i] }
func SC_CellPin(p *SC_Cell, i int) *SC_Pin    { return p.Pins[i] }
func SC_CellFunc(p *SC_Cell) []uint64         { return p.Pins[p.Inputs].Truth }
func SC_CellPinCap(p *SC_Cell, i int) float32 { return .5 * (p.Pins[i].RiseCap + p.Pins[i].FallCap) }
func SC_CellPinCapAve(p *SC_Cell) float32 {
	var c float32
	for i := 0; i < p.Inputs; i++ {
		c += SC_CellPinCap(p, i)
	}
	return c / float32(max(1, p.Inputs))
}
func SC_CellPinOutFunc(p *SC_Cell, i int) string { return p.Pins[p.Inputs+i].Function }
func SC_CellPinName(p *SC_Cell, i int) string    { return p.Pins[i].Name }
func Abc_SclWireLoadAlloc() *SC_WireLoad         { return &SC_WireLoad{} }
func Abc_SclWireLoadSelAlloc() *SC_WireLoadSel   { return &SC_WireLoadSel{} }
func Abc_SclTableTemplAlloc() *SC_TableTempl     { return &SC_TableTempl{} }
func Abc_SclSurfaceAlloc() *SC_Surface           { return &SC_Surface{} }
func Abc_SclTimingAlloc() *SC_Timing             { return &SC_Timing{} }
func Abc_SclTimingsAlloc() *SC_Timings           { return &SC_Timings{} }
func Abc_SclPinAlloc() *SC_Pin                   { return &SC_Pin{} }
func Abc_SclCellAlloc() *SC_Cell                 { return &SC_Cell{} }
func Abc_SclLibAlloc() *SC_Lib                   { return &SC_Lib{} }
func Abc_SclWireLoadFree(p *SC_WireLoad)         { *p = SC_WireLoad{} }
func Abc_SclWireLoadSelFree(p *SC_WireLoadSel)   { *p = SC_WireLoadSel{} }
func Abc_SclTableTemplFree(p *SC_TableTempl)     { *p = SC_TableTempl{} }
func Abc_SclSurfaceFree(p *SC_Surface)           { *p = SC_Surface{} }
func Abc_SclTimingFree(p *SC_Timing)             { *p = SC_Timing{} }
func Abc_SclTimingsFree(p *SC_Timings)           { *p = SC_Timings{} }
func Abc_SclPinFree(p *SC_Pin)                   { *p = SC_Pin{} }
func Abc_SclCellFree(p *SC_Cell)                 { *p = SC_Cell{} }
func Abc_SclLibFree(p *SC_Lib) {
	if p != nil {
		*p = SC_Lib{}
	}
}

// Scl_LibLookup performs the same bilinear interpolation and boundary
// extrapolation as sclLib.h. The caller supplies a rectangular timing surface.
func Scl_LibLookup(p *SC_Surface, slew, load float32) float32 {
	if len(p.Index0) == 1 && len(p.Index1) == 1 {
		return p.Data[0][0]
	}
	s, l := 0, 0
	for s+1 < len(p.Index0)-1 && p.Index0[s+1] <= slew {
		s++
	}
	for l+1 < len(p.Index1)-1 && p.Index1[l+1] <= load {
		l++
	}
	sf := (slew - p.Index0[s]) / (p.Index0[s+1] - p.Index0[s])
	lf := (load - p.Index1[l]) / (p.Index1[l+1] - p.Index1[l])
	a := p.Data[s][l] + lf*(p.Data[s][l+1]-p.Data[s][l])
	b := p.Data[s+1][l] + lf*(p.Data[s+1][l+1]-p.Data[s+1][l])
	return a + sf*(b-a)
}
func Scl_LibPinArrival(t *SC_Timing, arr, slew, load, out, outSlew *SC_Pair) {
	if t.Sense == ScTSPos || t.Sense == ScTSNon {
		out.Rise = max(out.Rise, arr.Rise+Scl_LibLookup(&t.CellRise, slew.Rise, load.Rise))
		out.Fall = max(out.Fall, arr.Fall+Scl_LibLookup(&t.CellFall, slew.Fall, load.Fall))
		outSlew.Rise = max(outSlew.Rise, Scl_LibLookup(&t.RiseTrans, slew.Rise, load.Rise))
		outSlew.Fall = max(outSlew.Fall, Scl_LibLookup(&t.FallTrans, slew.Fall, load.Fall))
	}
	if t.Sense == ScTSNeg || t.Sense == ScTSNon {
		out.Rise = max(out.Rise, arr.Fall+Scl_LibLookup(&t.CellRise, slew.Fall, load.Rise))
		out.Fall = max(out.Fall, arr.Rise+Scl_LibLookup(&t.CellFall, slew.Rise, load.Fall))
		outSlew.Rise = max(outSlew.Rise, Scl_LibLookup(&t.RiseTrans, slew.Fall, load.Rise))
		outSlew.Fall = max(outSlew.Fall, Scl_LibLookup(&t.FallTrans, slew.Rise, load.Fall))
	}
}

func Scl_LibLookupI(p *SC_Surface, slew, load int32) int32 {
	if len(p.Index0I) == 1 && len(p.Index1I) == 1 {
		return p.DataI[0][0]
	}
	s, l := 0, 0
	for s+1 < len(p.Index0I)-1 && p.Index0I[s+1] <= slew {
		s++
	}
	for l+1 < len(p.Index1I)-1 && p.Index1I[l+1] <= load {
		l++
	}
	a := p.DataI[s][l] + int32(int64(p.DataI[s][l+1]-p.DataI[s][l])*int64(load-p.Index1I[l])/int64(p.Index1I[l+1]-p.Index1I[l]))
	b := p.DataI[s+1][l] + int32(int64(p.DataI[s+1][l+1]-p.DataI[s+1][l])*int64(load-p.Index1I[l])/int64(p.Index1I[l+1]-p.Index1I[l]))
	return a + int32(int64(b-a)*int64(slew-p.Index0I[s])/int64(p.Index0I[s+1]-p.Index0I[s]))
}
func Scl_LibPinArrivalI(t *SC_Timing, arr, slew, load, out, outSlew *SC_PairI, delays *[4]int32) {
	if t.Sense == ScTSPos || t.Sense == ScTSNon {
		delays[0] = Scl_LibLookupI(&t.CellRise, slew.Rise, load.Rise)
		delays[1] = Scl_LibLookupI(&t.CellFall, slew.Fall, load.Fall)
		out.Rise = max(out.Rise, arr.Rise+delays[0])
		out.Fall = max(out.Fall, arr.Fall+delays[1])
		outSlew.Rise = max(outSlew.Rise, Scl_LibLookupI(&t.RiseTrans, slew.Rise, load.Rise))
		outSlew.Fall = max(outSlew.Fall, Scl_LibLookupI(&t.FallTrans, slew.Fall, load.Fall))
	}
	if t.Sense == ScTSNeg || t.Sense == ScTSNon {
		delays[2] = Scl_LibLookupI(&t.CellRise, slew.Fall, load.Rise)
		delays[3] = Scl_LibLookupI(&t.CellFall, slew.Rise, load.Fall)
		out.Rise = max(out.Rise, arr.Fall+delays[2])
		out.Fall = max(out.Fall, arr.Rise+delays[3])
		outSlew.Rise = max(outSlew.Rise, Scl_LibLookupI(&t.RiseTrans, slew.Fall, load.Rise))
		outSlew.Fall = max(outSlew.Fall, Scl_LibLookupI(&t.FallTrans, slew.Rise, load.Fall))
	}
}
func Scl_LibPinRequiredI(t *SC_Timing, in, out *SC_PairI, delays *[4]int32) {
	if t.Sense == ScTSPos || t.Sense == ScTSNon {
		in.Rise = min(in.Rise, out.Rise-delays[0])
		in.Fall = min(in.Fall, out.Fall-delays[1])
	}
	if t.Sense == ScTSNeg || t.Sense == ScTSNon {
		in.Fall = min(in.Fall, out.Rise-delays[2])
		in.Rise = min(in.Rise, out.Fall-delays[3])
	}
}
func Scl_LibPinDeparture(t *SC_Timing, in, slew, load, out *SC_Pair) {
	if t.Sense == ScTSPos || t.Sense == ScTSNon {
		in.Rise = max(in.Rise, out.Rise+Scl_LibLookup(&t.CellRise, slew.Rise, load.Rise))
		in.Fall = max(in.Fall, out.Fall+Scl_LibLookup(&t.CellFall, slew.Fall, load.Fall))
	}
	if t.Sense == ScTSNeg || t.Sense == ScTSNon {
		in.Fall = max(in.Fall, out.Rise+Scl_LibLookup(&t.CellRise, slew.Fall, load.Rise))
		in.Rise = max(in.Rise, out.Fall+Scl_LibLookup(&t.CellFall, slew.Rise, load.Fall))
	}
}
func Scl_LibPinArrivalEstimate(c *SC_Cell, i int, slew, load float32) float32 {
	a, s, l, out, outS := SC_Pair{}, SC_Pair{slew, slew}, SC_Pair{load, load}, SC_Pair{}, SC_Pair{}
	Scl_LibPinArrival(Scl_CellPinTime(c, i), &a, &s, &l, &out, &outS)
	return SC_PairAve(&out)
}
func Scl_LibPinArrivalEstimateI(c *SC_Cell, i int, slew, load int32) int32 {
	a, s, l, out, outS := SC_PairI{}, SC_PairI{slew, slew}, SC_PairI{load, load}, SC_PairI{}, SC_PairI{}
	var delays [4]int32
	Scl_LibPinArrivalI(Scl_CellPinTime(c, i), &a, &s, &l, &out, &outS, &delays)
	return (out.Rise + out.Fall) >> 1
}
func Scl_LibHandleInputDriver(c *SC_Cell, load, arr, outSlew *SC_Pair) {
	zero, base, out, s := SC_Pair{}, SC_Pair{}, SC_Pair{}, SC_Pair{}
	*outSlew = SC_Pair{}
	t := Scl_CellPinTime(c, 0)
	Scl_LibPinArrival(t, &zero, &zero, &zero, &base, &s)
	Scl_LibPinArrival(t, &zero, &zero, load, &out, outSlew)
	arr.Rise = out.Rise - base.Rise
	arr.Fall = out.Fall - base.Fall
}
func Scl_LibHandleInputDriver2(c *SC_Cell, load, arr, outSlew *SC_PairI) {
	zero, base, out, s := SC_PairI{}, SC_PairI{}, SC_PairI{}, SC_PairI{}
	*outSlew = SC_PairI{}
	var delays [4]int32
	t := Scl_CellPinTime(c, 0)
	Scl_LibPinArrivalI(t, &zero, &zero, &zero, &base, &s, &delays)
	Scl_LibPinArrivalI(t, &zero, &zero, load, &out, outSlew, &delays)
	arr.Rise = out.Rise - base.Rise
	arr.Fall = out.Fall - base.Fall
}

func initializeIntegerTables(p *SC_Lib) {
	for _, c := range p.Cells {
		c.AreaI = int32(1000 * c.Area)
		c.LeakageI = int32(1000 * c.Leakage)
		for _, pin := range c.Pins {
			pin.RiseCapI = int32(1000 * pin.RiseCap)
			pin.FallCapI = int32(1000 * pin.FallCap)
			for _, g := range pin.Timings {
				for _, t := range g.Timings {
					for _, s := range []*SC_Surface{&t.CellRise, &t.CellFall, &t.RiseTrans, &t.FallTrans} {
						s.Index0I = make([]int32, len(s.Index0))
						for i, v := range s.Index0 {
							s.Index0I[i] = int32(1000 * v)
						}
						s.Index1I = make([]int32, len(s.Index1))
						for i, v := range s.Index1 {
							s.Index1I[i] = int32(1000 * v)
						}
						s.DataI = make([][]int32, len(s.Data))
						for i, row := range s.Data {
							s.DataI[i] = make([]int32, len(row))
							for j, v := range row {
								s.DataI[i][j] = int32(1000 * v)
							}
						}
					}
				}
			}
		}
	}
}

type SC_SizePars struct {
	Iterations, IterationsNoChange, Window, Ratio, Notches, DelayUser, DelayGap, Timeout, BufferTreeEstimate, BypassFrequency int
	UseDepartures, DumpStats, UseWireLoads, Verbose, VeryVerbose                                                              bool
}
type SC_BusPars struct {
	GainRatio, Slew, Degree                                             int
	SizeOnly, AddBuffers, BufferPIs, UseWireLoads, Verbose, VeryVerbose bool
}
