// LUT area/delay libraries from ifLibLut.c.
package if_

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func lutNumber(text string) (float32, error) {
	value, err := strconv.ParseFloat(text, 32)
	if err != nil || value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("invalid nonnegative finite number %q", text)
	}
	return float32(value), nil
}

// ParseLUTLibrary accepts consecutive rows: input count, area, then one shared
// delay or a delay per pin. A shared row expands if other rows use pin delays.
func ParseLUTLibrary(name, text string) (*If_LibLut_t, error) {
	p := &If_LibLut_t{pName: name}
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	line := 0
	var counts [IF_MAX_LUTSIZE + 1]int
	for scanner.Scan() {
		line++
		row, _, _ := strings.Cut(scanner.Text(), "#")
		fields := strings.Fields(row)
		if len(fields) == 0 {
			continue
		}
		size, err := strconv.ParseInt(fields[0], 10, 32)
		expected := int64(p.LutMax) + 1
		if err != nil || size != expected || size > IF_MAX_LUTSIZE {
			return nil, fmt.Errorf("%s:%d: expected LUT size %d within 1..%d", name, line, expected, IF_MAX_LUTSIZE)
		}
		if len(fields) < 3 {
			return nil, fmt.Errorf("%s:%d: LUT %d requires area and delay", name, line, size)
		}
		pins := len(fields) - 2
		if pins != 1 && int64(pins) != size {
			return nil, fmt.Errorf("%s:%d: LUT %d requires one or %d pin delays", name, line, size, size)
		}
		p.pLutAreas[size], err = lutNumber(fields[1])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", name, line, err)
		}
		for pin, value := range fields[2:] {
			delay, err := lutNumber(value)
			if err != nil {
				return nil, fmt.Errorf("%s:%d: %w", name, line, err)
			}
			if pin > 0 && delay < p.pLutDelays[size][pin-1] {
				return nil, fmt.Errorf("%s:%d: pin delays must be nondecreasing", name, line)
			}
			p.pLutDelays[size][pin] = delay
		}
		counts[size] = pins
		if pins > 1 {
			p.fVarPinDelays = 1
		}
		p.LutMax = int32(size)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if p.LutMax == 0 {
		return nil, fmt.Errorf("%s: LUT library is empty", name)
	}
	if p.fVarPinDelays != 0 {
		for size := int32(1); size <= p.LutMax; size++ {
			if counts[size] == 1 {
				for pin := int32(1); pin < size; pin++ {
					p.pLutDelays[size][pin] = p.pLutDelays[size][0]
				}
			}
		}
	}
	return p, nil
}
func ReadLUTLibrary(filename string) (*If_LibLut_t, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseLUTLibrary(filename, string(data))
}
func If_LibLutReadString(text string) *If_LibLut_t {
	p, err := ParseLUTLibrary("<string>", text)
	if err != nil {
		return nil
	}
	return p
}
func If_LibLutRead(filename string) *If_LibLut_t {
	p, err := ReadLUTLibrary(filename)
	if err != nil {
		return nil
	}
	return p
}
func If_LibLutDup(p *If_LibLut_t) *If_LibLut_t {
	if p == nil {
		return nil
	}
	copy := *p
	return &copy
}
func If_LibLutFree(p *If_LibLut_t) {
	if p != nil {
		*p = If_LibLut_t{}
	}
}
func (p *If_LibLut_t_) Close() error { If_LibLutFree(p); return nil }
func (p *If_LibLut_t_) Name() string { return p.pName }
func (p *If_LibLut_t_) Area(inputs int32) float32 {
	if inputs < 1 || inputs > p.LutMax {
		panic("LUT input count outside library")
	}
	return p.pLutAreas[inputs]
}
func (p *If_LibLut_t_) PinDelay(inputs, pin int32) float32 {
	if inputs < 1 || inputs > p.LutMax || pin < 0 || pin >= inputs {
		panic("LUT pin outside library")
	}
	if p.fVarPinDelays == 0 {
		pin = 0
	}
	return p.pLutDelays[inputs][pin]
}
func WriteLUTLibrary(w io.Writer, p *If_LibLut_t) error {
	if p == nil || p.LutMax < 1 || p.LutMax > IF_MAX_LUTSIZE {
		return fmt.Errorf("invalid LUT library")
	}
	if _, err := fmt.Fprintln(w, "# k area delay(s)"); err != nil {
		return err
	}
	for size := int32(1); size <= p.LutMax; size++ {
		if _, err := fmt.Fprintf(w, "%d %.9g", size, p.pLutAreas[size]); err != nil {
			return err
		}
		pins := int32(1)
		if p.fVarPinDelays != 0 {
			pins = size
		}
		for pin := int32(0); pin < pins; pin++ {
			if _, err := fmt.Fprintf(w, " %.9g", p.pLutDelays[size][pin]); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	return nil
}
func If_LibLutPrint(p *If_LibLut_t) error { return WriteLUTLibrary(os.Stdout, p) }
func If_LibLutDelaysAreDiscrete(p *If_LibLut_t) int32 {
	if p == nil {
		return 1
	}
	for size := int32(1); size <= p.LutMax; size++ {
		pins := int32(1)
		if p.fVarPinDelays != 0 {
			pins = size
		}
		for pin := int32(0); pin < pins; pin++ {
			v := float64(p.pLutDelays[size][pin])
			if math.IsNaN(v) || math.IsInf(v, 0) || math.Trunc(v) != v {
				return 0
			}
		}
	}
	return 1
}
func If_LibLutDelaysAreDifferent(p *If_LibLut_t) int32 {
	if p == nil || p.LutMax < 1 {
		return 0
	}
	first := p.pLutDelays[1][0]
	for size := int32(1); size <= p.LutMax; size++ {
		pins := int32(1)
		if p.fVarPinDelays != 0 {
			pins = size
		}
		for pin := int32(0); pin < pins; pin++ {
			if p.pLutDelays[size][pin] != first {
				return 1
			}
		}
	}
	return 0
}
func If_LibLutSetSimple(size int32) *If_LibLut_t {
	if size < 3 || size > 10 {
		return nil
	}
	p := &If_LibLut_t{pName: "lutlib", LutMax: size}
	for i := int32(1); i <= size; i++ {
		p.pLutAreas[i] = 1
		p.pLutDelays[i][0] = 1
	}
	return p
}
func If_LibLutFastestPinDelay(p *If_LibLut_t) float32 {
	if p == nil {
		return 1
	}
	return p.pLutDelays[p.LutMax][0]
}
func If_LibLutSlowestPinDelay(p *If_LibLut_t) float32 {
	if p == nil {
		return 1
	}
	if p.fVarPinDelays != 0 && p.LutMax > 0 {
		return p.pLutDelays[p.LutMax][p.LutMax-1]
	}
	return p.pLutDelays[p.LutMax][0]
}
func Abc_FrameSetLutLibrary(frame *abcmain.Frame, text string) int32 {
	p := If_LibLutReadString(text)
	if p == nil || frame == nil {
		return 0
	}
	frame.SetResource("pLibLut", p)
	return 1
}
func Abc_FrameSetLutLibraryTest(frame *abcmain.Frame) int32 {
	return Abc_FrameSetLutLibrary(frame, "1 1 1000\n2 1 1000 1200\n3 1 1000 1200 1400\n4 1 1000 1200 1400 1600\n5 1 1000 1200 1400 1600 1800\n6 1 1000 1200 1400 1600 1800 2000\n")
}

func (p *If_LibLut_t_) VariablePinDelays() bool { return p.fVarPinDelays != 0 }
