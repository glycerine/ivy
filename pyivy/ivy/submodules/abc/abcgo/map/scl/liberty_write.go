package scl

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func LibertyText(p *SC_Lib) (string, error) {
	var b strings.Builder
	q := strconv.Quote
	fmt.Fprintf(&b, "library (%s) {\n time_unit : %s;\n capacitive_load_unit(%g,%s);\n", q(p.Name), q(map[int]string{9: "1ns", 10: "100ps", 11: "10ps", 12: "1ps"}[p.UnitTime]), p.UnitCapMultiplier, map[int]string{12: "pf", 15: "ff"}[p.UnitCapExponent])
	if p.DefaultWireLoad != "" {
		fmt.Fprintf(&b, " default_wire_load : %s;\n", q(p.DefaultWireLoad))
	}
	if p.DefaultWireLoadSelection != "" {
		fmt.Fprintf(&b, " default_wire_load_selection : %s;\n", q(p.DefaultWireLoadSelection))
	}
	fmt.Fprintf(&b, " default_max_transition : %g;\n", p.DefaultMaxOutSlew)
	for _, w := range p.WireLoads {
		fmt.Fprintf(&b, " wire_load(%s) { capacitance:%g; slope:%g;\n", q(w.Name), w.Cap, w.Slope)
		if len(w.Fanout) != len(w.Length) {
			return "", fmt.Errorf("wire-load vector length mismatch")
		}
		for i, f := range w.Fanout {
			fmt.Fprintf(&b, "  fanout_length(%d,%g);\n", f, w.Length[i])
		}
		b.WriteString(" }\n")
	}
	for _, w := range p.WireLoadSelections {
		fmt.Fprintf(&b, " wire_load_selection(%s) {\n", q(w.Name))
		if len(w.Models) != len(w.AreaFrom) || len(w.Models) != len(w.AreaTo) {
			return "", fmt.Errorf("wire-load selection vector mismatch")
		}
		for i, name := range w.Models {
			fmt.Fprintf(&b, "  wire_load_from_area(%g,%g,%s);\n", w.AreaFrom[i], w.AreaTo[i], q(name))
		}
		b.WriteString(" }\n")
	}
	for _, c := range p.Cells {
		if c.Sequential || c.Unsupported {
			continue
		}
		fmt.Fprintf(&b, " cell(%s) { area:%g; cell_leakage_power:%g; drive_strength:%d;\n", q(c.Name), c.Area, c.Leakage, c.DriveStrength)
		for i, pin := range c.Pins {
			fmt.Fprintf(&b, "  pin(%s) {\n", q(pin.Name))
			if i < c.Inputs {
				fmt.Fprintf(&b, "   direction:input; rise_capacitance:%g; fall_capacitance:%g;\n", pin.RiseCap, pin.FallCap)
			} else {
				if pin.Function == "" {
					return "", fmt.Errorf("cell %s pin %s has no textual function", c.Name, pin.Name)
				}
				fmt.Fprintf(&b, "   direction:output; function:%s; max_capacitance:%g; max_transition:%g;\n", q(pin.Function), pin.MaxOutCap, pin.MaxOutSlew)
				for _, g := range pin.Timings {
					for _, t := range g.Timings {
						fmt.Fprintf(&b, "   timing() { related_pin:%s; timing_sense:%s;\n", q(g.Name), map[SC_TSense]string{ScTSPos: "positive_unate", ScTSNeg: "negative_unate", ScTSNon: "non_unate"}[t.Sense])
						if t.When != "" {
							fmt.Fprintf(&b, "    when:%s;\n", q(t.When))
						}
						for _, entry := range []struct {
							name string
							s    *SC_Surface
						}{{"cell_rise", &t.CellRise}, {"cell_fall", &t.CellFall}, {"rise_transition", &t.RiseTrans}, {"fall_transition", &t.FallTrans}} {
							s := entry.s
							fmt.Fprintf(&b, "    %s() { index_1(%s); index_2(%s); values(\n", entry.name, q(floatList(s.Index0)), q(floatList(s.Index1)))
							for j, row := range s.Data {
								if j > 0 {
									b.WriteString(",\n")
								}
								fmt.Fprintf(&b, "     %s", q(floatList(row)))
							}
							b.WriteString("); }\n")
						}
						b.WriteString("   }\n")
					}
				}
			}
			b.WriteString("  }\n")
		}
		b.WriteString(" }\n")
	}
	b.WriteString("}\n")
	return b.String(), nil
}
func floatList(v []float32) string {
	values := make([]string, len(v))
	for i, x := range v {
		values[i] = strconv.FormatFloat(float64(x), 'g', -1, 32)
	}
	return strings.Join(values, ",")
}
func Abc_SclWriteLiberty(path string, p *SC_Lib) error {
	text, e := LibertyText(p)
	if e != nil {
		return e
	}
	return os.WriteFile(path, []byte(text), 0644)
}
