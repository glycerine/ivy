package scl

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

func Abc_SclHashCells(p *SC_Lib) {
	p.cellIndex = make(map[string]int, len(p.Cells))
	for i, c := range p.Cells {
		p.cellIndex[c.Name] = i
	}
}
func Abc_SclCellFind(p *SC_Lib, name string) int {
	if p.cellIndex == nil {
		Abc_SclHashCells(p)
	}
	i, ok := p.cellIndex[name]
	if !ok {
		return -1
	}
	return i
}
func Abc_SclClassCellNum(p *SC_Cell) int {
	if p == nil {
		return 0
	}
	n := 1
	for c := p.Next; c != nil && c != p; c = c.Next {
		n++
	}
	return n
}
func Abc_SclLibClassNum(p *SC_Lib) int { return len(p.CellClasses) }
func Abc_SclLinkCells(p *SC_Lib) {
	classes := map[string][]*SC_Cell{}
	var keys []string
	for _, c := range p.Cells {
		var b strings.Builder
		fmt.Fprintf(&b, "%d/%d/", c.Inputs, c.Outputs)
		for _, pin := range c.Pins[c.Inputs : c.Inputs+1] {
			for _, w := range pin.Truth {
				fmt.Fprintf(&b, "%016x", w)
			}
			b.WriteByte('/')
		}
		key := b.String()
		if _, ok := classes[key]; !ok {
			keys = append(keys, key)
		}
		classes[key] = append(classes[key], c)
	}
	less := func(a, b *SC_Cell) bool {
		if a.Inputs != b.Inputs {
			return a.Inputs < b.Inputs
		}
		ca, cb := SC_CellPinCapAve(a), SC_CellPinCapAve(b)
		if ca != cb {
			return ca < cb
		}
		return a.Name < b.Name
	}
	sort.SliceStable(keys, func(i, j int) bool { return less(classes[keys[i]][0], classes[keys[j]][0]) })
	p.CellClasses = nil
	for _, key := range keys {
		cells := classes[key]
		sort.SliceStable(cells, func(i, j int) bool { return less(cells[i], cells[j]) })
		for i, c := range cells {
			c.Next = cells[(i+1)%len(cells)]
			c.Prev = cells[(i+len(cells)-1)%len(cells)]
			c.Repr = cells[0]
			c.Average = cells[len(cells)/2]
			c.Order = i
			c.Gates = len(cells)
		}
		p.CellClasses = append(p.CellClasses, cells[0])
	}
}
func Abc_SclFindInvertor(p *SC_Lib, buffer int32) *SC_Cell {
	truth := uint64(0x5555555555555555)
	if buffer != 0 {
		truth = 0xaaaaaaaaaaaaaaaa
	}
	for _, c := range p.CellClasses {
		if c.Inputs == 1 && len(SC_CellFunc(c)) == 1 && SC_CellFunc(c)[0] == truth {
			return c.Repr
		}
	}
	return nil
}

func Abc_SclFindSmallestGate(p *SC_Cell, capMin float32) *SC_Cell {
	r := p.Repr
	if r == nil {
		r = p
	}
	for c := r; ; {
		if SC_CellPinCapAve(c) > capMin {
			return c
		}
		c = c.Next
		if c == nil || c == r {
			break
		}
	}
	return r.Prev
}

func Abc_SclFetchWireLoadModel(p *SC_Lib, name string) *SC_WireLoad {
	for _, w := range p.WireLoads {
		if w.Name == name {
			return w
		}
	}
	return nil
}
func Abc_SclFindWireLoadModel(p *SC_Lib, area float32) *SC_WireLoad {
	for _, s := range p.WireLoadSelections {
		if s.Name != p.DefaultWireLoadSelection {
			continue
		}
		for i := range s.Models {
			if area >= s.AreaFrom[i] && area < s.AreaTo[i] {
				return Abc_SclFetchWireLoadModel(p, s.Models[i])
			}
		}
		if len(s.Models) > 0 {
			return Abc_SclFetchWireLoadModel(p, s.Models[len(s.Models)-1])
		}
	}
	return Abc_SclFetchWireLoadModel(p, p.DefaultWireLoad)
}
func Abc_SclConvertLeakageIntoArea(p *SC_Lib, a, b float32) {
	for _, c := range p.Cells {
		c.Area = a*c.Area + b*c.Leakage
	}
}
func Abc_SclShortNames(p *SC_Lib) {
	digits := len(strconv.Itoa(len(p.CellClasses)))
	for i, r := range p.CellClasses {
		width := len(strconv.Itoa(Abc_SclClassCellNum(r)))
		for c, j := r, 0; ; j++ {
			c.Name = fmt.Sprintf("g%0*d_%0*d", digits, i, width, j)
			names := map[string]string{}
			for k, pin := range c.Pins[:c.Inputs] {
				names[pin.Name] = string(rune('a' + k))
			}
			for k, pin := range c.Pins[c.Inputs:] {
				var b strings.Builder
				for pos := 0; pos < len(pin.Function); {
					ch := pin.Function[pos]
					if ch == '_' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' {
						end := pos + 1
						for end < len(pin.Function) {
							v := pin.Function[end]
							if !(v == '_' || v >= 'a' && v <= 'z' || v >= 'A' && v <= 'Z' || v >= '0' && v <= '9') {
								break
							}
							end++
						}
						word := pin.Function[pos:end]
						if replacement, ok := names[word]; ok {
							b.WriteString(replacement)
						} else {
							b.WriteString(word)
						}
						pos = end
					} else {
						b.WriteByte(ch)
						pos++
					}
				}
				pin.Function = b.String()
				pin.Name = string(rune('z' - k))
				for _, g := range pin.Timings {
					if replacement, ok := names[g.Name]; ok {
						g.Name = replacement
					}
					for _, t := range g.Timings {
						if replacement, ok := names[t.RelatedPin]; ok {
							t.RelatedPin = replacement
						}
					}
				}
			}
			for k, pin := range c.Pins[:c.Inputs] {
				pin.Name = string(rune('a' + k))
			}
			c = c.Next
			if c == nil || c == r {
				break
			}
		}
	}
	p.Name = fmt.Sprintf("lib%d", len(p.Cells))
	Abc_SclHashCells(p)
}

func Abc_SclHasDelayInfo(p *SC_Lib) bool {
	c := Abc_SclFindInvertor(p, 0)
	return c != nil && Scl_CellPinTime(c, 0) != nil
}

func Abc_SclLibNormalizeSurface(s *SC_Surface, time, load float32) {
	for i := range s.Index0 {
		s.Index0[i] *= time
	}
	for i := range s.Index1 {
		s.Index1[i] *= load
	}
	for i := range s.Data {
		for j := range s.Data[i] {
			s.Data[i][j] *= time
		}
	}
}
func Abc_SclLibNormalize(p *SC_Lib) {
	time := float32(math.Pow10(12 - p.UnitTime))
	load := p.UnitCapMultiplier * float32(math.Pow10(15-p.UnitCapExponent))
	if time == 1 && load == 1 {
		return
	}
	p.UnitTime = 12
	p.UnitCapMultiplier = 1
	p.UnitCapExponent = 15
	p.DefaultMaxOutSlew *= time
	for _, w := range p.WireLoads {
		w.Cap *= load
	}
	for _, c := range p.Cells {
		for _, pin := range c.Pins {
			pin.Cap *= load
			pin.RiseCap *= load
			pin.FallCap *= load
			pin.MaxOutCap *= load
			pin.MaxOutSlew *= time
			for _, g := range pin.Timings {
				for _, t := range g.Timings {
					for _, s := range []*SC_Surface{&t.CellRise, &t.CellFall, &t.RiseTrans, &t.FallTrans} {
						Abc_SclLibNormalizeSurface(s, time, load)
					}
				}
			}
		}
	}
}
func Scl_CellPinTime(c *SC_Cell, i int) *SC_Timing {
	if c.Outputs != 1 || i < 0 || i >= c.Inputs {
		return nil
	}
	groups := c.Pins[c.Inputs].Timings
	if i >= len(groups) || len(groups[i].Timings) == 0 {
		return nil
	}
	return groups[i].Timings[0]
}
func Abc_SclComputeAverageSlew(p *SC_Lib) float32 {
	c := Abc_SclFindInvertor(p, 0)
	if c == nil {
		return 0
	}
	t := Scl_CellPinTime(c, 0)
	if t == nil || len(t.CellRise.Index0) == 0 {
		return 0
	}
	return t.CellRise.Index0[len(t.CellRise.Index0)/3]
}
func Abc_SclComputeParametersPin(p *SC_Lib, c *SC_Cell, i int, slew float32, ld, pd *float32) bool {
	t := Scl_CellPinTime(c, i)
	if t == nil {
		return false
	}
	index := t.CellRise.Index1
	if len(index) == 0 {
		return false
	}
	if len(index) == 1 {
		*ld = 0
		*pd = t.CellRise.Data[0][0]
		return true
	}
	upper := len(index) - 2
	if upper == 0 {
		upper = 1
	} // Two-point tables still need two distinct tangent samples.
	loads := [3]float32{0, index[0], index[upper]}
	var delays [3]float32
	for j, load := range loads {
		arr, inSlew, l, out, outSlew := SC_Pair{}, SC_Pair{slew, slew}, SC_Pair{load, load}, SC_Pair{}, SC_Pair{}
		Scl_LibPinArrival(t, &arr, &inSlew, &l, &out, &outSlew)
		delays[j] = SC_PairAve(&out)
	}
	*ld = (delays[2] - delays[1]) / ((loads[2] - loads[1]) / SC_CellPinCap(c, i))
	*pd = delays[0]
	return true
}
func Abc_SclComputeDelayCellPin(p *SC_Lib, c *SC_Cell, i int, slew, gain float32) float32 {
	var ld, pd float32
	Abc_SclComputeParametersPin(p, c, i, slew, &ld, &pd)
	return .01*ld*gain + pd
}
func Abc_SclComputeDelayClassPin(p *SC_Lib, c *SC_Cell, i int, slew, gain float32) float32 {
	var delay float32
	count := 0
	for q := c; ; {
		if !q.Skip {
			delay += Abc_SclComputeDelayCellPin(p, q, i, slew, gain)
			count++
		}
		q = q.Next
		if q == nil || q == c {
			break
		}
	}
	return delay / float32(max(1, count))
}
func Abc_SclProduceGenlibStrSimple(p *SC_Lib) string {
	var b strings.Builder
	b.WriteString("GATE _const0_            0.00 z=CONST0;\nGATE _const1_            0.00 z=CONST1;\n")
	for _, c := range p.Cells {
		if c.Inputs == 0 {
			continue
		}
		for _, out := range c.Pins[c.Inputs:] {
			fmt.Fprintf(&b, "GATE %-16s %7.2f %s=%s;\n", c.Name, c.Area, out.Name, out.Function)
			for _, in := range c.Pins[:c.Inputs] {
				fmt.Fprintf(&b, "         PIN %-4s UNKNOWN  1  999  1.00  0.00  1.00  0.00\n", in.Name)
			}
		}
	}
	b.WriteString("\n.end\n")
	return b.String()
}

func GenlibText(p *SC_Lib, slew, gain float32, minimum int) (string, error) {
	return genlibText(p, p.CellClasses, slew, gain, minimum)
}
func genlibText(p *SC_Lib, cells []*SC_Cell, slew, gain float32, minimum int) (string, error) {
	if e := Abc_SclMarkSkippedCells(p); e != nil {
		return "", e
	}
	if slew == 0 {
		slew = Abc_SclComputeAverageSlew(p)
	}
	largest := 0
	for _, r := range p.CellClasses {
		if r.Outputs == 1 {
			largest = max(largest, Abc_SclClassCellNum(r))
		}
	}
	if minimum >= largest {
		minimum = 0
	}
	var b strings.Builder
	b.WriteString("GATE _const0_            0.00 z=CONST0;\nGATE _const1_            0.00 z=CONST1;\n")
	for _, r := range cells {
		if r.Inputs == 0 || r.Outputs != 1 || minimum > 0 && r.Inputs > 2 && Abc_SclClassCellNum(r) < minimum {
			continue
		}
		out := r.Pins[r.Inputs]
		fmt.Fprintf(&b, "GATE %-16s %7.2f %s=%s;\n", r.Name, r.Area, out.Name, out.Function)
		for i, pin := range r.Pins[:r.Inputs] {
			delay := Abc_SclComputeDelayClassPin(p, r, i, slew, gain)
			if delay <= 0 || math.IsNaN(float64(delay)) || math.IsInf(float64(delay), 0) {
				return "", fmt.Errorf("cell %s pin %s has invalid delay %g", r.Name, pin.Name, delay)
			}
			fmt.Fprintf(&b, "         PIN %-4s UNKNOWN  1  999  %7.2f  0.00  %7.2f  0.00\n", pin.Name, delay, delay)
		}
	}
	b.WriteString("\n.end\n")
	return b.String(), nil
}

func Abc_SclComputeParametersCell(p *SC_Lib, c *SC_Cell, slew float32, ld, pd *float32) bool {
	var a, b float32
	for i := 0; i < c.Inputs; i++ {
		var x, y float32
		if !Abc_SclComputeParametersPin(p, c, i, slew, &x, &y) {
			return false
		}
		a += x
		b += y
	}
	*ld = a / float32(max(1, c.Inputs))
	*pd = b / float32(max(1, c.Inputs))
	return true
}
func Abc_SclComputeParametersClass(p *SC_Lib, r *SC_Cell, slew float32, ld, pd *float32) {
	var a, b float32
	n := 0
	for c := r; ; {
		var x, y float32
		Abc_SclComputeParametersCell(p, c, slew, &x, &y)
		a += x
		b += y
		n++
		c = c.Next
		if c == nil || c == r {
			break
		}
	}
	*ld = a / float32(max(1, n))
	*pd = b / float32(max(1, n))
}
func Abc_SclComputeParametersClassPin(p *SC_Lib, r *SC_Cell, i int, slew float32, ld, pd *float32) {
	var a, b float32
	n := 0
	for c := r; ; {
		var x, y float32
		Abc_SclComputeParametersPin(p, c, i, slew, &x, &y)
		a += x
		b += y
		n++
		c = c.Next
		if c == nil || c == r {
			break
		}
	}
	*ld = a / float32(max(1, n))
	*pd = b / float32(max(1, n))
}
func Abc_SclComputeAreaClass(r *SC_Cell) float32 {
	var a float32
	n := 0
	for c := r; ; {
		if !c.Skip {
			a += c.Area
			n++
		}
		c = c.Next
		if c == nil || c == r {
			break
		}
	}
	return a / float32(max(1, n))
}
func MarkSkippedCells(p *SC_Lib, path string, out io.Writer) error {
	f, e := os.Open(path)
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	count := 0
	for scan.Scan() {
		fields := strings.Fields(scan.Text())
		if len(fields) == 0 {
			continue
		}
		i := Abc_SclCellFind(p, fields[0])
		if i < 0 {
			fmt.Fprintf(out, "Cannot find cell %q in the library %q.\n", fields[0], p.Name)
			continue
		}
		p.Cells[i].Skip = true
		count++
	}
	if e = scan.Err(); e != nil {
		return e
	}
	fmt.Fprintf(out, "Marked %d cells for skipping in the library %q.\n", count, p.Name)
	return nil
}
func Abc_SclMarkSkippedCells(p *SC_Lib) error { return MarkSkippedCells(p, p.Name+".skip", os.Stdout) }
func Abc_SclDumpGenlib(path string, p *SC_Lib, slew, gain float32, minimum int) error {
	text, e := GenlibText(p, slew, gain, minimum)
	if e != nil {
		return e
	}
	return os.WriteFile(path, []byte(text), 0644)
}
func Abc_SclProduceGenlibStr(p *SC_Lib, slew, gain float32, minimum int, count *int) (string, error) {
	text, e := GenlibText(p, slew, gain, minimum)
	if count != nil {
		*count = strings.Count(text, "GATE ")
	}
	return text, e
}

func PrintCells(out io.Writer, p *SC_Lib, slew, gain float32, invertersOnly, short bool) error {
	if e := MarkSkippedCells(p, p.Name+".skip", out); e != nil {
		return e
	}
	if slew == 0 {
		slew = Abc_SclComputeAverageSlew(p)
	}
	fmt.Fprintf(out, "Library %q has %d cells in %d classes.\n", p.Name, len(p.Cells), len(p.CellClasses))
	for i, r := range p.CellClasses {
		if invertersOnly && r.Inputs != 1 {
			continue
		}
		for _, pin := range r.Pins[r.Inputs:] {
			fmt.Fprintf(out, "Class%4d : Cells =%3d   Ins =%2d  Outs =%2d  %s\n", i, Abc_SclClassCellNum(r), r.Inputs, r.Outputs, pin.Function)
		}
		if short {
			continue
		}
		for c := r; ; {
			mark := " "
			if c.Skip {
				mark = "s"
			}
			fmt.Fprintf(out, "  %3d %s : %s  %2d   A =%8.2f  L =%8.2f", c.Order+1, mark, c.Name, c.DriveStrength, c.Area, c.Leakage)
			if c.Outputs == 1 {
				var ld, pd float32
				if Abc_SclComputeParametersCell(p, c, slew, &ld, &pd) {
					fmt.Fprintf(out, "  D =%6.1f ps  LD =%6.1f ps  PD =%6.1f ps  C =%5.1f ff  Cm =%5.0f ff  Sm =%5.1f ps", .01*gain*ld+pd, ld, pd, SC_CellPinCapAve(c), c.Pins[c.Inputs].MaxOutCap, c.Pins[c.Inputs].MaxOutSlew)
				}
			}
			fmt.Fprintln(out)
			c = c.Next
			if c == nil || c == r {
				break
			}
		}
	}
	return nil
}
func Abc_SclPrintCells(p *SC_Lib, slew, gain float32, invertersOnly, short int32) error {
	return PrintCells(os.Stdout, p, slew, gain, invertersOnly != 0, short != 0)
}
