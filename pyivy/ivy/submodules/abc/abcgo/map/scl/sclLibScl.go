package scl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// SCL version 8 uses unsigned base-128 lengths, NUL-terminated strings,
// little-endian IEEE floats and uncompressed little-endian truth words.
type sclCodec struct {
	data []byte
	pos  int
	err  error
}

func (c *sclCodec) integer() int {
	if c.err != nil {
		return 0
	}
	v, n := binary.Uvarint(c.data[c.pos:])
	if n <= 0 || v > 1<<30 {
		c.err = fmt.Errorf("invalid SCL integer at byte %d", c.pos)
		return 0
	}
	c.pos += n
	return int(v)
}
func (c *sclCodec) take(n int) []byte {
	if c.err != nil {
		return make([]byte, n)
	}
	if n < 0 || n > len(c.data)-c.pos {
		c.err = io.ErrUnexpectedEOF
		return make([]byte, max(0, n))
	}
	b := c.data[c.pos : c.pos+n]
	c.pos += n
	return b
}
func (c *sclCodec) str() string {
	if c.err != nil {
		return ""
	}
	i := bytes.IndexByte(c.data[c.pos:], 0)
	if i < 0 {
		c.err = io.ErrUnexpectedEOF
		return ""
	}
	s := string(c.data[c.pos : c.pos+i])
	c.pos += i + 1
	return s
}
func (c *sclCodec) flt() float32 { return math.Float32frombits(binary.LittleEndian.Uint32(c.take(4))) }
func (c *sclCodec) count() int {
	n := c.integer()
	if n > len(c.data)-c.pos {
		c.err = fmt.Errorf("SCL count %d exceeds remaining data", n)
		return 0
	}
	return n
}
func (c *sclCodec) surface() SC_Surface {
	s := SC_Surface{}
	for n := c.count(); n > 0; n-- {
		s.Index0 = append(s.Index0, c.flt())
	}
	for n := c.count(); n > 0; n-- {
		s.Index1 = append(s.Index1, c.flt())
	}
	if len(s.Index0)*len(s.Index1) > (len(c.data)-c.pos)/4 {
		c.err = io.ErrUnexpectedEOF
		return s
	}
	for range s.Index0 {
		row := make([]float32, len(s.Index1))
		for j := range row {
			row[j] = c.flt()
		}
		s.Data = append(s.Data, row)
	}
	for i, n := range []int{3, 4, 6} {
		for j := 0; j < n; j++ {
			s.Approx[i][j] = c.flt()
		}
	}
	return s
}
func Abc_SclReadFromStr(data []byte) (*SC_Lib, error) {
	r := sclCodec{data: data}
	if v := r.integer(); v != ABC_SCL_CUR_VERSION {
		return nil, fmt.Errorf("unsupported SCL version %d", v)
	}
	p := &SC_Lib{Name: r.str(), DefaultWireLoad: r.str(), DefaultWireLoadSelection: r.str(), DefaultMaxOutSlew: r.flt(), UnitTime: r.integer(), UnitCapMultiplier: r.flt(), UnitCapExponent: r.integer()}
	for n := r.count(); n > 0; n-- {
		w := &SC_WireLoad{Name: r.str(), Cap: r.flt(), Slope: r.flt()}
		for j := r.count(); j > 0; j-- {
			w.Fanout = append(w.Fanout, int32(r.integer()))
			w.Length = append(w.Length, r.flt())
		}
		p.WireLoads = append(p.WireLoads, w)
	}
	for n := r.count(); n > 0; n-- {
		w := &SC_WireLoadSel{Name: r.str()}
		for j := r.count(); j > 0; j-- {
			w.AreaFrom = append(w.AreaFrom, r.flt())
			w.AreaTo = append(w.AreaTo, r.flt())
			w.Models = append(w.Models, r.str())
		}
		p.WireLoadSelections = append(p.WireLoadSelections, w)
	}
	for n := r.count(); n > 0; n-- {
		cell := &SC_Cell{Name: r.str(), ID: len(p.Cells), Area: r.flt(), Leakage: r.flt(), DriveStrength: r.integer(), Inputs: r.count(), Outputs: r.count()}
		if cell.Inputs > 20 {
			return nil, fmt.Errorf("SCL cell %s exceeds 20 inputs", cell.Name)
		}
		for j := 0; j < cell.Inputs; j++ {
			cell.Pins = append(cell.Pins, &SC_Pin{Name: r.str(), Direction: ScDirInput, RiseCap: r.flt(), FallCap: r.flt()})
		}
		for j := 0; j < cell.Outputs; j++ {
			pin := &SC_Pin{Name: r.str(), Direction: ScDirOutput, MaxOutCap: r.flt(), MaxOutSlew: r.flt()}
			if r.integer() != cell.Inputs {
				return nil, fmt.Errorf("SCL cell %s inconsistent input count", cell.Name)
			}
			pin.Function = r.str()
			for k := 0; k < max(1, 1<<max(0, cell.Inputs-6)); k++ {
				pin.Truth = append(pin.Truth, binary.LittleEndian.Uint64(r.take(8)))
			}
			if pin.Function != "" {
				truth, e := formulaTruth(pin.Function, cell.Pins[:cell.Inputs])
				if e != nil {
					return nil, e
				}
				for k, w := range truth {
					if pin.Truth[k] != w {
						return nil, fmt.Errorf("SCL cell %s formula/truth mismatch", cell.Name)
					}
				}
			}
			for k := 0; k < cell.Inputs; k++ {
				group := &SC_Timings{Name: r.str()}
				count := r.integer()
				if count > 1 {
					return nil, fmt.Errorf("SCL timing group has %d entries", count)
				}
				if count == 1 {
					t := &SC_Timing{RelatedPin: group.Name, Sense: SC_TSense(r.integer())}
					t.CellRise = r.surface()
					t.CellFall = r.surface()
					t.RiseTrans = r.surface()
					t.FallTrans = r.surface()
					group.Timings = append(group.Timings, t)
				}
				pin.Timings = append(pin.Timings, group)
			}
			cell.Pins = append(cell.Pins, pin)
		}
		p.Cells = append(p.Cells, cell)
	}
	if r.err != nil {
		return nil, fmt.Errorf("SCL byte %d: %w", r.pos, r.err)
	}
	if r.pos != len(data) {
		return nil, fmt.Errorf("SCL trailing data at byte %d", r.pos)
	}
	initializeIntegerTables(p)
	Abc_SclHashCells(p)
	Abc_SclLinkCells(p)
	return p, nil
}

type sclWriter struct{ data []byte }

func (w *sclWriter) integer(v int) { w.data = binary.AppendUvarint(w.data, uint64(v)) }
func (w *sclWriter) str(v string)  { w.data = append(w.data, v...); w.data = append(w.data, 0) }
func (w *sclWriter) flt(v float32) {
	w.data = binary.LittleEndian.AppendUint32(w.data, math.Float32bits(v))
}
func (w *sclWriter) surface(s *SC_Surface) {
	w.integer(len(s.Index0))
	for _, v := range s.Index0 {
		w.flt(v)
	}
	w.integer(len(s.Index1))
	for _, v := range s.Index1 {
		w.flt(v)
	}
	for _, row := range s.Data {
		for _, v := range row {
			w.flt(v)
		}
	}
	for i, n := range []int{3, 4, 6} {
		for j := 0; j < n; j++ {
			w.flt(s.Approx[i][j])
		}
	}
}
func MarshalSCL(p *SC_Lib) ([]byte, error) {
	if p.UnitTime < 0 || p.UnitCapExponent < 0 {
		return nil, fmt.Errorf("negative library unit exponent")
	}
	w := sclWriter{}
	w.integer(ABC_SCL_CUR_VERSION)
	w.str(p.Name)
	w.str(p.DefaultWireLoad)
	w.str(p.DefaultWireLoadSelection)
	w.flt(p.DefaultMaxOutSlew)
	w.integer(p.UnitTime)
	w.flt(p.UnitCapMultiplier)
	w.integer(p.UnitCapExponent)
	w.integer(len(p.WireLoads))
	for _, x := range p.WireLoads {
		if len(x.Fanout) != len(x.Length) {
			return nil, fmt.Errorf("wire-load lengths differ")
		}
		w.str(x.Name)
		w.flt(x.Cap)
		w.flt(x.Slope)
		w.integer(len(x.Fanout))
		for i, v := range x.Fanout {
			w.integer(int(v))
			w.flt(x.Length[i])
		}
	}
	w.integer(len(p.WireLoadSelections))
	for _, x := range p.WireLoadSelections {
		if len(x.Models) != len(x.AreaFrom) || len(x.Models) != len(x.AreaTo) {
			return nil, fmt.Errorf("wire-load selection lengths differ")
		}
		w.str(x.Name)
		w.integer(len(x.Models))
		for i, name := range x.Models {
			w.flt(x.AreaFrom[i])
			w.flt(x.AreaTo[i])
			w.str(name)
		}
	}
	count := 0
	for _, c := range p.Cells {
		if !c.Sequential && !c.Unsupported {
			count++
		}
	}
	w.integer(count)
	for _, c := range p.Cells {
		if c.Sequential || c.Unsupported {
			continue
		}
		if c.Inputs < 0 || c.Inputs > 20 || c.Outputs < 0 || c.DriveStrength < 0 || len(c.Pins) != c.Inputs+c.Outputs {
			return nil, fmt.Errorf("cell %s invalid pin count", c.Name)
		}
		w.str(c.Name)
		w.flt(c.Area)
		w.flt(c.Leakage)
		w.integer(c.DriveStrength)
		w.integer(c.Inputs)
		w.integer(c.Outputs)
		for _, pin := range c.Pins[:c.Inputs] {
			w.str(pin.Name)
			w.flt(pin.RiseCap)
			w.flt(pin.FallCap)
		}
		for _, pin := range c.Pins[c.Inputs:] {
			w.str(pin.Name)
			w.flt(pin.MaxOutCap)
			w.flt(pin.MaxOutSlew)
			w.integer(c.Inputs)
			w.str(pin.Function)
			if len(pin.Truth) != max(1, 1<<max(0, c.Inputs-6)) {
				return nil, fmt.Errorf("cell %s truth size mismatch", c.Name)
			}
			for _, v := range pin.Truth {
				w.data = binary.LittleEndian.AppendUint64(w.data, v)
			}
			if len(pin.Timings) != c.Inputs {
				return nil, fmt.Errorf("cell %s timing group count mismatch", c.Name)
			}
			for _, g := range pin.Timings {
				w.str(g.Name)
				if len(g.Timings) > 1 {
					return nil, fmt.Errorf("cell %s unmerged timing groups", c.Name)
				}
				w.integer(len(g.Timings))
				for _, t := range g.Timings {
					w.integer(int(t.Sense))
					for _, s := range []*SC_Surface{&t.CellRise, &t.CellFall, &t.RiseTrans, &t.FallTrans} {
						if len(s.Data) != len(s.Index0) {
							return nil, fmt.Errorf("invalid surface row count")
						}
						for _, row := range s.Data {
							if len(row) != len(s.Index1) {
								return nil, fmt.Errorf("invalid surface column count")
							}
						}
						w.surface(s)
					}
				}
			}
		}
	}
	return w.data, nil
}
func Abc_SclWriteScl(path string, p *SC_Lib) error {
	data, e := MarshalSCL(p)
	if e != nil {
		return e
	}
	return os.WriteFile(path, data, 0644)
}
func Abc_SclReadFromFile(path string) (*SC_Lib, error) {
	data, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	p, e := Abc_SclReadFromStr(data)
	if e == nil {
		p.FileName = path
		Abc_SclLibNormalize(p)
	}
	return p, e
}
