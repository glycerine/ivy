package scl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
)

const (
	SCL_NUM         = 1000
	SCL_INFINITY    = int32(0x3fffffff)
	SCL_INPUT_CELL  = "input_cell"
	SCL_INPUT_ARR   = "input_arrival"
	SCL_INPUT_SLEW  = "input_slew"
	SCL_INPUT_LOAD  = "input_load"
	SCL_OUTPUT_REQ  = "output_required"
	SCL_OUTPUT_LOAD = "output_load"
)

func Scl_Flt2Int(v float32) int32 { return int32(SCL_NUM * v) }
func Scl_Int2Flt(v int32) float32 { return float32(v) / SCL_NUM }

type Scl_Con_t struct {
	FileName, ModelName, InputCellDefault                                                             string
	InputArrivalDefault, InputSlewDefault, InputLoadDefault, OutputRequiredDefault, OutputLoadDefault int32
	InputCells                                                                                        []string
	InputCellPointers                                                                                 []*SC_Cell
	InputArrivals, InputSlews, InputLoads, OutputRequired, OutputLoads                                []int32
	InputNames, OutputNames                                                                           *util.Abc_Nam_t
}

func Scl_ConAlloc(path string, inputs, outputs *util.Abc_Nam_t) *Scl_Con_t {
	ni, no := int(util.Abc_NamObjNumMax(inputs))-1, int(util.Abc_NamObjNumMax(outputs))-1
	makeValues := func(n int) []int32 {
		v := make([]int32, n)
		for i := range v {
			v[i] = SCL_INFINITY
		}
		return v
	}
	return &Scl_Con_t{FileName: path, InputNames: inputs, OutputNames: outputs, InputCellDefault: "", InputArrivalDefault: SCL_INFINITY, InputSlewDefault: SCL_INFINITY, InputLoadDefault: SCL_INFINITY, OutputRequiredDefault: SCL_INFINITY, OutputLoadDefault: SCL_INFINITY, InputCells: make([]string, ni), InputArrivals: makeValues(ni), InputSlews: makeValues(ni), InputLoads: makeValues(ni), OutputRequired: makeValues(no), OutputLoads: makeValues(no)}
}
func Scl_ConFree(p *Scl_Con_t) {
	if p == nil {
		return
	}
	util.Abc_NamDeref(p.InputNames)
	util.Abc_NamDeref(p.OutputNames)
	*p = Scl_Con_t{}
}
func (p *Scl_Con_t) Close() error { Scl_ConFree(p); return nil }
func ParseConstraints(p *Scl_Con_t, r io.Reader, diagnostics io.Writer) error {
	defaults := map[string]*int32{SCL_INPUT_ARR: &p.InputArrivalDefault, SCL_INPUT_SLEW: &p.InputSlewDefault, SCL_INPUT_LOAD: &p.InputLoadDefault, SCL_OUTPUT_REQ: &p.OutputRequiredDefault, SCL_OUTPUT_LOAD: &p.OutputLoadDefault}
	vectors := map[string][]int32{SCL_INPUT_ARR: p.InputArrivals, SCL_INPUT_SLEW: p.InputSlews, SCL_INPUT_LOAD: p.InputLoads, SCL_OUTPUT_REQ: p.OutputRequired, SCL_OUTPUT_LOAD: p.OutputLoads}
	scan := bufio.NewScanner(r)
	scan.Buffer(make([]byte, 4096), 1024*1024)
	line := 0
	for scan.Scan() {
		line++
		fields := strings.Fields(scan.Text())
		if len(fields) == 0 {
			continue
		}
		if len(fields) < 2 {
			fmt.Fprintf(diagnostics, "Line %d: Skipping directive %q without argument.\n", line, fields[0])
			continue
		}
		key := strings.TrimPrefix(fields[0], ".")
		if key == "model" {
			p.ModelName = fields[1]
			continue
		}
		if key == "default_"+SCL_INPUT_CELL {
			p.InputCellDefault = fields[1]
			continue
		}
		number := func(s string) (int32, error) {
			v, e := strconv.ParseFloat(s, 32)
			if e != nil || v != v || v > 2147483.647 || v < -2147483.648 {
				return 0, fmt.Errorf("%s:%d: invalid constraint value %q", p.FileName, line, s)
			}
			return Scl_Flt2Int(float32(v)), nil
		}
		if d := defaults[strings.TrimPrefix(key, "default_")]; strings.HasPrefix(key, "default_") && d != nil {
			v, e := number(fields[1])
			if e != nil {
				return e
			}
			*d = v
			continue
		}
		if len(fields) < 3 {
			fmt.Fprintf(diagnostics, "Directive %s should be followed by two arguments.\n", fields[0])
			continue
		}
		if key != SCL_INPUT_CELL && vectors[key] == nil {
			fmt.Fprintf(diagnostics, "Line %d: Skipping unrecognized directive %q.\n", line, fields[0])
			continue
		}
		names := p.InputNames
		kind := "input"
		if key == SCL_OUTPUT_REQ || key == SCL_OUTPUT_LOAD {
			names = p.OutputNames
			kind = "output"
		}
		id := int(util.Abc_NamStrFind(names, fields[1])) - 1
		if id < 0 {
			fmt.Fprintf(diagnostics, "Line %d: Cannot find %s %q.\n", line, kind, fields[1])
			continue
		}
		if key == SCL_INPUT_CELL {
			p.InputCells[id] = fields[2]
		} else {
			v, e := number(fields[2])
			if e != nil {
				return e
			}
			vectors[key][id] = v
		}
	}
	if e := scan.Err(); e != nil {
		return e
	}
	for _, d := range defaults {
		if *d == SCL_INFINITY {
			*d = 0
		}
	}
	for key, v := range vectors {
		for i, value := range v {
			if value == SCL_INFINITY {
				v[i] = *defaults[key]
			}
		}
	}
	for i, name := range p.InputCells {
		if name == "" {
			p.InputCells[i] = p.InputCellDefault
		}
	}
	return nil
}
func Scl_ConParse(p *Scl_Con_t, inputs, outputs *util.Abc_Nam_t) error {
	f, e := os.Open(p.FileName)
	if e != nil {
		return e
	}
	defer f.Close()
	return ParseConstraints(p, f, os.Stdout)
}
func Scl_ConRead(path string, inputs, outputs *util.Abc_Nam_t) (*Scl_Con_t, error) {
	p := Scl_ConAlloc(path, inputs, outputs)
	if e := Scl_ConParse(p, inputs, outputs); e != nil {
		Scl_ConFree(p)
		return nil, e
	}
	return p, nil
}
func WriteConstraints(w io.Writer, p *Scl_Con_t) error {
	if _, e := fmt.Fprintf(w, ".model %s\n", p.ModelName); e != nil {
		return e
	}
	if p.InputCellDefault != "" {
		if _, e := fmt.Fprintf(w, ".default_%s %s\n", SCL_INPUT_CELL, p.InputCellDefault); e != nil {
			return e
		}
	}
	specs := []struct {
		key    string
		def    int32
		values []int32
		names  *util.Abc_Nam_t
	}{{SCL_INPUT_ARR, p.InputArrivalDefault, p.InputArrivals, p.InputNames}, {SCL_INPUT_SLEW, p.InputSlewDefault, p.InputSlews, p.InputNames}, {SCL_INPUT_LOAD, p.InputLoadDefault, p.InputLoads, p.InputNames}, {SCL_OUTPUT_REQ, p.OutputRequiredDefault, p.OutputRequired, p.OutputNames}, {SCL_OUTPUT_LOAD, p.OutputLoadDefault, p.OutputLoads, p.OutputNames}}
	for _, s := range specs {
		if s.def != 0 {
			if _, e := fmt.Fprintf(w, ".default_%s %.2f\n", s.key, Scl_Int2Flt(s.def)); e != nil {
				return e
			}
		}
	}
	for i, name := range p.InputCells {
		if name != "" && name != p.InputCellDefault {
			if _, e := fmt.Fprintf(w, ".%s %s %s\n", SCL_INPUT_CELL, util.Abc_NamIntToStr(p.InputNames, int32(i+1)), name); e != nil {
				return e
			}
		}
	}
	for _, s := range specs {
		for i, value := range s.values {
			if value != s.def {
				if _, e := fmt.Fprintf(w, ".%s %s %.2f\n", s.key, util.Abc_NamIntToStr(s.names, int32(i+1)), Scl_Int2Flt(value)); e != nil {
					return e
				}
			}
		}
	}
	return nil
}
func Scl_ConWrite(p *Scl_Con_t, path string) error {
	if path == "" {
		return WriteConstraints(os.Stdout, p)
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	e = WriteConstraints(f, p)
	closeErr := f.Close()
	if e != nil {
		return e
	}
	return closeErr
}
func Scl_ConReadMan() *Scl_Con_t {
	p, _ := abcmain.Abc_FrameGetGlobalFrame().Resource("pAbcCon").(*Scl_Con_t)
	return p
}
func Scl_ConIsRunning() bool { return Scl_ConReadMan() != nil }

func Scl_ConHasInCells_(p *Scl_Con_t) bool {
	for _, s := range p.InputCells {
		if s != "" {
			return true
		}
	}
	return false
}
func Scl_ConHasInCells() bool                      { return Scl_ConHasInCells_(Scl_ConReadMan()) }
func Scl_ConGetInCell_(p *Scl_Con_t, i int) string { return p.InputCells[i] }
func Scl_ConGetInCell(i int) string                { return Scl_ConGetInCell_(Scl_ConReadMan(), i) }
func Scl_ConHasInArrs_(p *Scl_Con_t) bool {
	for _, v := range p.InputArrivals {
		if v != 0 {
			return true
		}
	}
	return false
}
func Scl_ConHasInArrs() bool                     { return Scl_ConHasInArrs_(Scl_ConReadMan()) }
func Scl_ConGetInArr_(p *Scl_Con_t, i int) int32 { return p.InputArrivals[i] }
func Scl_ConGetInArr(i int) int32                { return Scl_ConGetInArr_(Scl_ConReadMan(), i) }
func Scl_ConGetInArrFloat(i int) float32         { return Scl_Int2Flt(Scl_ConGetInArr(i)) }
func Scl_ConHasInSlews_(p *Scl_Con_t) bool {
	for _, v := range p.InputSlews {
		if v != 0 {
			return true
		}
	}
	return false
}
func Scl_ConHasInSlews() bool                     { return Scl_ConHasInSlews_(Scl_ConReadMan()) }
func Scl_ConGetInSlew_(p *Scl_Con_t, i int) int32 { return p.InputSlews[i] }
func Scl_ConGetInSlew(i int) int32                { return Scl_ConGetInSlew_(Scl_ConReadMan(), i) }
func Scl_ConGetInSlewFloat(i int) float32         { return Scl_Int2Flt(Scl_ConGetInSlew(i)) }
func Scl_ConHasInLoads_(p *Scl_Con_t) bool {
	for _, v := range p.InputLoads {
		if v != 0 {
			return true
		}
	}
	return false
}
func Scl_ConHasInLoads() bool                     { return Scl_ConHasInLoads_(Scl_ConReadMan()) }
func Scl_ConGetInLoad_(p *Scl_Con_t, i int) int32 { return p.InputLoads[i] }
func Scl_ConGetInLoad(i int) int32                { return Scl_ConGetInLoad_(Scl_ConReadMan(), i) }
func Scl_ConGetInLoadFloat(i int) float32         { return Scl_Int2Flt(Scl_ConGetInLoad(i)) }
func Scl_ConHasOutReqs_(p *Scl_Con_t) bool {
	for _, v := range p.OutputRequired {
		if v != 0 {
			return true
		}
	}
	return false
}
func Scl_ConHasOutReqs() bool                     { return Scl_ConHasOutReqs_(Scl_ConReadMan()) }
func Scl_ConGetOutReq_(p *Scl_Con_t, i int) int32 { return p.OutputRequired[i] }
func Scl_ConGetOutReq(i int) int32                { return Scl_ConGetOutReq_(Scl_ConReadMan(), i) }
func Scl_ConGetOutReqFloat(i int) float32         { return Scl_Int2Flt(Scl_ConGetOutReq(i)) }
func Scl_ConHasOutLoads_(p *Scl_Con_t) bool {
	for _, v := range p.OutputLoads {
		if v != 0 {
			return true
		}
	}
	return false
}
func Scl_ConHasOutLoads() bool                     { return Scl_ConHasOutLoads_(Scl_ConReadMan()) }
func Scl_ConGetOutLoad_(p *Scl_Con_t, i int) int32 { return p.OutputLoads[i] }
func Scl_ConGetOutLoad(i int) int32                { return Scl_ConGetOutLoad_(Scl_ConReadMan(), i) }
func Scl_ConGetOutLoadFloat(i int) float32         { return Scl_Int2Flt(Scl_ConGetOutLoad(i)) }
