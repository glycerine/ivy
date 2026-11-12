package fpga

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

func DefaultLibrary(size int) *Fpga_LutLib_t {
	p := &Fpga_LutLib_t{Name: fmt.Sprintf("lut%d", size), LutMax: int32(size), Areas: make([]float32, size+1), Delays: make([][]float32, size+1)}
	for k := 1; k <= size; k++ {
		p.Areas[k] = 1
		p.Delays[k] = []float32{1}
	}
	return p
}
func (p *Fpga_LutLib_t) Validate() error {
	if p == nil || p.LutMax < 1 || p.LutMax > FPGA_MAX_LUTSIZE {
		return fmt.Errorf("fpga: LUT size must be 1..%d", FPGA_MAX_LUTSIZE)
	}
	if len(p.Areas) <= int(p.LutMax) || len(p.Delays) <= int(p.LutMax) {
		return fmt.Errorf("fpga: missing LUT rows")
	}
	for k := 1; k <= int(p.LutMax); k++ {
		if !finite(p.Areas[k]) || p.Areas[k] < 0 {
			return fmt.Errorf("fpga: invalid area for LUT %d", k)
		}
		d := p.Delays[k]
		if len(d) != 1 && len(d) != k {
			return fmt.Errorf("fpga: LUT %d needs one or %d delays", k, k)
		}
		for _, v := range d {
			if !finite(v) || v < 0 {
				return fmt.Errorf("fpga: invalid LUT delay")
			}
		}
	}
	return nil
}
func finite(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }
func ReadLibrary(r io.Reader, name string) (*Fpga_LutLib_t, error) {
	p := &Fpga_LutLib_t{Name: name, Areas: []float32{0}, Delays: [][]float32{nil}}
	s := bufio.NewScanner(r)
	line := 0
	for s.Scan() {
		line++
		text := s.Text()
		if i := strings.IndexAny(text, "#"); i >= 0 {
			text = text[:i]
		}
		fields := strings.Fields(text)
		if len(fields) == 0 {
			continue
		}
		k, e := strconv.Atoi(fields[0])
		if e != nil || k != len(p.Areas) || k > FPGA_MAX_LUTSIZE || len(fields) < 3 {
			return nil, fmt.Errorf("%s:%d: expected consecutive LUT size, area, delay(s)", name, line)
		}
		values := make([]float32, len(fields)-1)
		for i, f := range fields[1:] {
			v, e := strconv.ParseFloat(f, 32)
			if e != nil {
				return nil, fmt.Errorf("%s:%d: %w", name, line, e)
			}
			values[i] = float32(v)
		}
		p.Areas = append(p.Areas, values[0])
		p.Delays = append(p.Delays, values[1:])
		p.LutMax = int32(k)
		p.VariablePinDelays = p.VariablePinDelays || len(values) > 2
	}
	if e := s.Err(); e != nil {
		return nil, e
	}
	if e := p.Validate(); e != nil {
		return nil, e
	}
	return p, nil
}
func Fpga_LutLibRead(path string, verbose int32) *Fpga_LutLib_t {
	f, e := os.Open(path)
	if e != nil {
		return nil
	}
	defer f.Close()
	p, e := ReadLibrary(f, path)
	if e != nil {
		if verbose != 0 {
			fmt.Fprintln(os.Stderr, e)
		}
		return nil
	}
	return p
}
func Fpga_LutLibReadVarMax(p *Fpga_LutLib_t) int32             { return p.LutMax }
func Fpga_LutLibReadLutAreas(p *Fpga_LutLib_t) []float32       { return p.Areas }
func Fpga_LutLibReadLutArea(p *Fpga_LutLib_t, k int32) float32 { return p.Areas[k] }
func Fpga_LutLibDup(p *Fpga_LutLib_t) *Fpga_LutLib_t {
	if p == nil {
		return nil
	}
	q := *p
	q.Areas = append([]float32(nil), p.Areas...)
	q.Delays = make([][]float32, len(p.Delays))
	for k, d := range p.Delays {
		q.Delays[k] = append([]float32(nil), d...)
	}
	return &q
}
func Fpga_LutLibFree(p *Fpga_LutLib_t) {
	if p != nil {
		p.Areas = nil
		p.Delays = nil
	}
}
func (p *Fpga_LutLib_t) Close() error { Fpga_LutLibFree(p); return nil }
func (p *Fpga_LutLib_t) Write(w io.Writer) error {
	for k := 1; k <= int(p.LutMax); k++ {
		if _, e := fmt.Fprintf(w, "%d %g", k, p.Areas[k]); e != nil {
			return e
		}
		for _, d := range p.Delays[k] {
			if _, e := fmt.Fprintf(w, " %g", d); e != nil {
				return e
			}
		}
		if _, e := fmt.Fprintln(w); e != nil {
			return e
		}
	}
	return nil
}
func Fpga_LutLibPrint(p *Fpga_LutLib_t) { _ = p.Write(os.Stdout) }
func Fpga_LutLibDelaysAreDiscrete(p *Fpga_LutLib_t) int32 {
	for _, row := range p.Delays {
		for _, d := range row {
			if d != float32(math.Trunc(float64(d))) {
				return 0
			}
		}
	}
	return 1
}
