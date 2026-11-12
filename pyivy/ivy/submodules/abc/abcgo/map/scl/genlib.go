package scl

import (
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"strings"
)

func Abc_SclDeriveGenlibSimple(p *SC_Lib) (*mio.Mio_Library_t, error) {
	return mio.ReadLibrary(p.FileName, Abc_SclProduceGenlibStrSimple(p), nil)
}

func (p *SC_Lib) MioCellID(name string) (int32, bool) {
	id := Abc_SclCellFind(p, name)
	return int32(id), id >= 0
}

type LibertyMIOReader struct{}

func (LibertyMIOReader) ReadMioLiberty(path string, verbose bool) (*mio.Mio_Library_t, error) {
	lib, e := ReadLiberty(path)
	if e != nil {
		return nil, e
	}
	return Abc_SclDeriveGenlibSimple(lib)
}
func Abc_SclDeriveGenlib(p *SC_Lib, profile *mio.Mio_Library_t, slew, gain float32, minimum int, verbose int32) (*mio.Mio_Library_t, error) {
	if profile != nil {
		return deriveProfileGenlib(p, profile, slew, gain, minimum)
	}
	text, e := GenlibText(p, slew, gain, minimum)
	if e != nil {
		return nil, e
	}
	return mio.ReadLibrary(p.FileName, text, nil)
}

func profileGenlibText(p *SC_Lib, profile *mio.Mio_Library_t, slew, gain float32, minimum int) (string, error) {
	cells := make([]*SC_Cell, 0, len(p.Cells))
	for _, c := range p.Cells {
		if c.Inputs > 1 {
			g := mio.Mio_LibraryReadGateByName(profile, c.Name, "")
			if g == nil || mio.Mio_GateReadProfile(g) == 0 {
				continue
			}
		}
		cells = append(cells, c)
	}
	return genlibText(p, cells, slew, gain, minimum)
}
func Abc_SclProduceGenlibStrProfile(p *SC_Lib, profile *mio.Mio_Library_t, slew, gain float32, minimum int, count *int) (string, error) {
	text, e := profileGenlibText(p, profile, slew, gain, minimum)
	if count != nil {
		*count = strings.Count(text, "GATE ")
	}
	return text, e
}
func deriveProfileGenlib(p *SC_Lib, profile *mio.Mio_Library_t, slew, gain float32, minimum int) (*mio.Mio_Library_t, error) {
	text, e := profileGenlibText(p, profile, slew, gain, minimum)
	if e != nil {
		return nil, e
	}
	return mio.ReadLibrary(p.FileName, text, nil)
}

// Abc_SclReadFromGenlib preserves ABC's fallback unit-delay timing model;
// GENLIB's linear block/fanout delays are not NLDM measurement surfaces.
func Abc_SclReadFromGenlib(lib *mio.Mio_Library_t) (*SC_Lib, error) {
	p := &SC_Lib{Name: mio.Mio_LibraryReadName(lib), UnitTime: 12, UnitCapMultiplier: 1, UnitCapExponent: 15}
	for g := mio.Mio_LibraryReadGates(lib); g != nil; g = mio.Mio_GateReadNext(g) {
		c := &SC_Cell{Name: mio.Mio_GateReadName(g), ID: len(p.Cells), Area: float32(mio.Mio_GateReadArea(g)), Inputs: int(mio.Mio_GateReadPinNum(g)), Outputs: 1}
		for pin := mio.Mio_GateReadPins(g); pin != nil; pin = mio.Mio_PinReadNext(pin) {
			c.Pins = append(c.Pins, &SC_Pin{Name: mio.Mio_PinReadName(pin), Direction: ScDirInput})
		}
		out := &SC_Pin{Name: mio.Mio_GateReadOutName(g), Direction: ScDirOutput, Function: mio.Mio_GateReadForm(g)}
		truth, e := formulaTruth(out.Function, c.Pins)
		if e != nil {
			return nil, e
		}
		out.Truth = truth
		for pin := mio.Mio_GateReadPins(g); pin != nil; pin = mio.Mio_PinReadNext(pin) {
			t := &SC_Timing{RelatedPin: mio.Mio_PinReadName(pin), Sense: ScTSNon}
			switch mio.Mio_PinReadPhase(pin) {
			case mio.MIO_PHASE_INV:
				t.Sense = ScTSNeg
			case mio.MIO_PHASE_NONINV:
				t.Sense = ScTSPos
			}
			for _, s := range []*SC_Surface{&t.CellRise, &t.CellFall, &t.RiseTrans, &t.FallTrans} {
				s.Index0 = []float32{0}
				s.Index1 = []float32{0}
				s.Data = [][]float32{{1}}
			}
			out.Timings = append(out.Timings, &SC_Timings{Name: t.RelatedPin, Timings: []*SC_Timing{t}})
		}
		c.Pins = append(c.Pins, out)
		p.Cells = append(p.Cells, c)
	}
	initializeIntegerTables(p)
	Abc_SclHashCells(p)
	Abc_SclLinkCells(p)
	return p, nil
}

func Abc_SclInstallGenlib(p *SC_Lib, slew, gain float32, minimum int) error {
	lib, e := Abc_SclDeriveGenlib(p, nil, slew, gain, minimum, 1)
	if e != nil {
		return e
	}
	f := abcmain.Abc_FrameGetGlobalFrame()
	f.SetResource("pLibGen", lib)
	f.SetResource("mio.libertyReader", LibertyMIOReader{})
	return nil
}
