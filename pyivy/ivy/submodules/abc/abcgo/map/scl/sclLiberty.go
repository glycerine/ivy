package scl

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// LibertyItem preserves the generic Liberty hierarchy, including attributes
// not interpreted by the standard-cell reader. Line/column locate diagnostics.
type LibertyItem struct {
	Key          string
	Arguments    []string
	Value        string
	Children     []*LibertyItem
	Line, Column int
}
type Scl_Tree_t struct {
	Name, Contents string
	Root           *LibertyItem
}
type Scl_Item_t = LibertyItem
type libertyToken struct {
	text      string
	quoted    bool
	line, col int
}
type libertyParser struct {
	tokens []libertyToken
	pos    int
	name   string
}

func libertyLex(name, text string) ([]libertyToken, error) {
	var out []libertyToken
	line, col := 1, 1
	advance := func(c byte) {
		if c == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	for i := 0; i < len(text); {
		c := text[i]
		if unicode.IsSpace(rune(c)) || c == '\\' && i+1 < len(text) && (text[i+1] == '\n' || text[i+1] == '\r') {
			advance(c)
			i++
			continue
		}
		if c == '/' && i+1 < len(text) && text[i+1] == '/' {
			for i < len(text) && text[i] != '\n' {
				advance(text[i])
				i++
			}
			continue
		}
		if c == '/' && i+1 < len(text) && text[i+1] == '*' {
			start := i
			advance(c)
			advance('*')
			i += 2
			for i+1 < len(text) && !(text[i] == '*' && text[i+1] == '/') {
				advance(text[i])
				i++
			}
			if i+1 >= len(text) {
				return nil, fmt.Errorf("%s:%d:%d: unterminated comment at byte %d", name, line, col, start)
			}
			advance('*')
			advance('/')
			i += 2
			continue
		}
		t := libertyToken{line: line, col: col}
		if c == '"' {
			t.quoted = true
			i++
			advance(c)
			var b strings.Builder
			closed := false
			for i < len(text) {
				c = text[i]
				if c == '"' {
					advance(c)
					i++
					closed = true
					break
				}
				if c == '\\' && i+1 < len(text) {
					advance(c)
					i++
					c = text[i]
					if c == '\n' || c == '\r' {
						advance(c)
						i++
						continue
					}
					if c != '"' && c != '\\' {
						b.WriteByte('\\')
					}
				}
				b.WriteByte(c)
				advance(c)
				i++
			}
			if !closed {
				return nil, fmt.Errorf("%s:%d:%d: unterminated string", name, t.line, t.col)
			}
			t.text = b.String()
		} else if strings.ContainsRune("(){}:;,", rune(c)) {
			t.text = string(c)
			advance(c)
			i++
		} else {
			start := i
			for i < len(text) && !unicode.IsSpace(rune(text[i])) && !strings.ContainsRune("(){}:;,\"", rune(text[i])) {
				if text[i] == '/' && i+1 < len(text) && (text[i+1] == '*' || text[i+1] == '/') {
					break
				}
				advance(text[i])
				i++
			}
			t.text = text[start:i]
		}
		out = append(out, t)
	}
	return out, nil
}
func (p *libertyParser) err(t libertyToken, s string) error {
	return fmt.Errorf("%s:%d:%d: %s", p.name, t.line, t.col, s)
}
func (p *libertyParser) items(nested bool) ([]*LibertyItem, error) {
	var items []*LibertyItem
	for p.pos < len(p.tokens) {
		t := p.tokens[p.pos]
		p.pos++
		if t.text == "}" {
			if !nested {
				return nil, p.err(t, "unexpected closing brace")
			}
			return items, nil
		}
		if t.text == ";" {
			continue
		}
		if strings.Contains("({):,", t.text) {
			return nil, p.err(t, "expected attribute or group name")
		}
		item := &LibertyItem{Key: t.text, Line: t.line, Column: t.col}
		if p.pos >= len(p.tokens) {
			return nil, p.err(t, "missing attribute or group body")
		}
		op := p.tokens[p.pos]
		p.pos++
		switch op.text {
		case ":":
			var values []string
			for p.pos < len(p.tokens) && (p.tokens[p.pos].quoted || p.tokens[p.pos].text != ";" && p.tokens[p.pos].text != "}") {
				values = append(values, p.tokens[p.pos].text)
				p.pos++
			}
			item.Value = strings.Join(values, " ")
			if p.pos < len(p.tokens) && p.tokens[p.pos].text == ";" {
				p.pos++
			}
		case "(":
			var arg []string
			depth := 0
			closed := false
			for p.pos < len(p.tokens) {
				v := p.tokens[p.pos]
				p.pos++
				if !v.quoted && v.text == ")" && depth == 0 {
					if len(arg) > 0 {
						item.Arguments = append(item.Arguments, strings.Join(arg, " "))
					}
					closed = true
					break
				}
				if !v.quoted && v.text == "," && depth == 0 {
					item.Arguments = append(item.Arguments, strings.Join(arg, " "))
					arg = nil
					continue
				}
				if !v.quoted && v.text == "(" {
					depth++
				}
				if !v.quoted && v.text == ")" {
					depth--
				}
				arg = append(arg, v.text)
			}
			if !closed {
				return nil, p.err(t, "unclosed argument list")
			}
			if p.pos >= len(p.tokens) {
				return nil, p.err(t, "missing group body or semicolon")
			}
			end := p.tokens[p.pos]
			p.pos++
			if end.text == "{" {
				var err error
				item.Children, err = p.items(true)
				if err != nil {
					return nil, err
				}
			} else if end.text != ";" {
				return nil, p.err(end, "expected group body or semicolon")
			}
		default:
			return nil, p.err(op, "expected ':' or '('")
		}
		items = append(items, item)
	}
	if nested {
		return nil, fmt.Errorf("%s: unexpected end of file in group", p.name)
	}
	return items, nil
}
func ParseLibertyTree(name, text string) (*Scl_Tree_t, error) {
	tokens, err := libertyLex(name, text)
	if err != nil {
		return nil, err
	}
	p := libertyParser{tokens: tokens, name: name}
	items, err := p.items(false)
	if err != nil {
		return nil, err
	}
	if len(items) != 1 || items[0].Key != "library" {
		return nil, fmt.Errorf("%s: expected one library group", name)
	}
	return &Scl_Tree_t{Name: name, Contents: text, Root: items[0]}, nil
}
func Scl_LibertyParse(path string, verbose int32) (*Scl_Tree_t, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseLibertyTree(path, string(data))
}
func Scl_LibertyRoot(p *Scl_Tree_t) *LibertyItem { return p.Root }
func Scl_LibertyStop(p *Scl_Tree_t, verbose int32) {
	if p != nil {
		*p = Scl_Tree_t{}
	}
}
func (n *LibertyItem) child(key string) *LibertyItem {
	for _, c := range n.Children {
		if c.Key == key {
			return c
		}
	}
	return nil
}
func (n *LibertyItem) value(key string) string {
	if c := n.child(key); c != nil {
		return c.Value
	}
	return ""
}
func (n *LibertyItem) number(key string) float32 {
	v, _ := strconv.ParseFloat(n.value(key), 32)
	return float32(v)
}
func (n *LibertyItem) name() string {
	if len(n.Arguments) > 0 {
		return n.Arguments[0]
	}
	return ""
}
func floatVector(text string) ([]float32, error) {
	fields := strings.FieldsFunc(text, func(r rune) bool { return r == ',' || unicode.IsSpace(r) || r == '\\' })
	v := make([]float32, len(fields))
	for i, s := range fields {
		x, err := strconv.ParseFloat(s, 32)
		if err != nil || x != x || x > 3.402823466e38 || x < -3.402823466e38 {
			return nil, fmt.Errorf("invalid finite table number %q", s)
		}
		v[i] = float32(x)
	}
	return v, nil
}
func Scl_LibertyReadFloatVec(text string) ([]float32, error) { return floatVector(text) }

func ReadLiberty(path string) (*SC_Lib, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseLiberty(path, string(data))
}
func Abc_SclReadLiberty(path string, verbose, veryVerbose int32) (*SC_Lib, error) {
	return ReadLiberty(path)
}
func ParseLiberty(name, text string) (*SC_Lib, error) {
	tree, err := ParseLibertyTree(name, text)
	if err != nil {
		return nil, err
	}
	return libraryFromTree(tree)
}
func libraryFromTree(tree *Scl_Tree_t) (*SC_Lib, error) {
	root := tree.Root
	if e := validateNumericAttributes(tree.Name, root); e != nil {
		return nil, e
	}
	lib := &SC_Lib{Name: root.name(), FileName: tree.Name, DefaultWireLoad: root.value("default_wire_load"), DefaultWireLoadSelection: root.value("default_wire_load_selection"), DefaultMaxOutSlew: root.number("default_max_transition"), UnitTime: 9, UnitCapMultiplier: 1, UnitCapExponent: 12}
	if unit := root.value("time_unit"); unit != "" {
		switch unit {
		case "1ns":
			lib.UnitTime = 9
		case "100ps":
			lib.UnitTime = 10
		case "10ps":
			lib.UnitTime = 11
		case "1ps":
			lib.UnitTime = 12
		default:
			return nil, fmt.Errorf("%s: unsupported time_unit %q", tree.Name, unit)
		}
	}
	if c := root.child("capacitive_load_unit"); c != nil {
		if len(c.Arguments) != 2 {
			return nil, fmt.Errorf("capacitive_load_unit requires multiplier and unit")
		}
		v, err := strconv.ParseFloat(c.Arguments[0], 32)
		if err != nil {
			return nil, err
		}
		lib.UnitCapMultiplier = float32(v)
		switch c.Arguments[1] {
		case "pf":
			lib.UnitCapExponent = 12
		case "ff":
			lib.UnitCapExponent = 15
		default:
			return nil, fmt.Errorf("unsupported capacitance unit %q", c.Arguments[1])
		}
	}
	templates := map[string]*SC_TableTempl{}
	for _, n := range root.Children {
		switch n.Key {
		case "lu_table_template":
			t := &SC_TableTempl{Name: n.name()}
			for i := 1; i <= 2; i++ {
				v := n.value(fmt.Sprintf("variable_%d", i))
				if v != "" {
					t.Variables = append(t.Variables, v)
				}
				if c := n.child(fmt.Sprintf("index_%d", i)); c != nil {
					a, e := floatVector(strings.Join(c.Arguments, ","))
					if e != nil {
						return nil, e
					}
					t.Index = append(t.Index, a)
				}
			}
			templates[t.Name] = t
			lib.Templates = append(lib.Templates, t)
		case "wire_load":
			w := &SC_WireLoad{Name: n.name(), Cap: n.number("capacitance"), Slope: n.number("slope")}
			for _, a := range n.Children {
				if a.Key == "fanout_length" && len(a.Arguments) == 2 {
					x, e := strconv.Atoi(a.Arguments[0])
					if e != nil {
						return nil, e
					}
					y, e := strconv.ParseFloat(a.Arguments[1], 32)
					if e != nil {
						return nil, e
					}
					w.Fanout = append(w.Fanout, int32(x))
					w.Length = append(w.Length, float32(y))
				}
			}
			lib.WireLoads = append(lib.WireLoads, w)
		case "wire_load_selection":
			w := &SC_WireLoadSel{Name: n.name()}
			for _, a := range n.Children {
				if a.Key == "wire_load_from_area" && len(a.Arguments) == 3 {
					x, e := strconv.ParseFloat(a.Arguments[0], 32)
					if e != nil {
						return nil, e
					}
					y, e := strconv.ParseFloat(a.Arguments[1], 32)
					if e != nil {
						return nil, e
					}
					w.AreaFrom = append(w.AreaFrom, float32(x))
					w.AreaTo = append(w.AreaTo, float32(y))
					w.Models = append(w.Models, a.Arguments[2])
				}
			}
			lib.WireLoadSelections = append(lib.WireLoadSelections, w)
		}
	}
	for _, n := range root.Children {
		if n.Key != "cell" {
			continue
		}
		if n.child("ff") != nil || n.child("latch") != nil || n.child("statetable") != nil || n.value("dont_use") == "true" {
			continue
		}
		cell := &SC_Cell{Name: n.name(), ID: len(lib.Cells), Area: n.number("area"), Leakage: n.number("cell_leakage_power"), DriveStrength: int(n.number("drive_strength"))}
		if n.value("area") == "" {
			cell.Area = 1
		}
		var inputs, outputs []*SC_Pin
		var outputNodes []*LibertyItem
		skip := false
		for _, pn := range n.Children {
			if pn.Key != "pin" {
				continue
			}
			if pn.child("three_state") != nil {
				skip = true
			}
			pin := &SC_Pin{Name: pn.name(), Cap: pn.number("capacitance"), RiseCap: pn.number("rise_capacitance"), FallCap: pn.number("fall_capacitance"), Function: pn.value("function"), MaxOutCap: pn.number("max_capacitance"), MaxOutSlew: pn.number("max_transition")}
			if pin.RiseCap == 0 {
				pin.RiseCap = pin.Cap
			}
			if pin.FallCap == 0 {
				pin.FallCap = pin.Cap
			}
			if pin.Function != "" {
				pin.Direction = ScDirOutput

				outputs = append(outputs, pin)
				outputNodes = append(outputNodes, pn)
			} else {
				pin.Direction = ScDirInput
				inputs = append(inputs, pin)
			}
		}
		if skip || len(outputs) == 0 {
			continue
		}
		cell.Inputs = len(inputs)
		cell.Outputs = len(outputs)
		cell.Pins = append(inputs, outputs...)
		for i, pin := range outputs {
			truth, e := formulaTruth(pin.Function, inputs)
			if e != nil {
				return nil, fmt.Errorf("cell %s pin %s: %w", cell.Name, pin.Name, e)
			}
			pin.Truth = truth
			for _, input := range inputs {
				group := &SC_Timings{Name: input.Name}
				for _, tn := range outputNodes[i].Children {
					if tn.Key != "timing" {
						continue
					}
					related := strings.Fields(tn.value("related_pin"))
					found := false
					for _, r := range related {
						found = found || r == input.Name
					}
					if !found {
						continue
					}
					timing := &SC_Timing{RelatedPin: input.Name, When: tn.value("when"), Sense: ScTSNon}
					switch tn.value("timing_sense") {
					case "positive_unate":
						timing.Sense = ScTSPos
					case "negative_unate":
						timing.Sense = ScTSNeg
					}
					for _, spec := range []struct {
						key string
						dst *SC_Surface
					}{{"cell_rise", &timing.CellRise}, {"cell_fall", &timing.CellFall}, {"rise_transition", &timing.RiseTrans}, {"fall_transition", &timing.FallTrans}} {
						table := tn.child(spec.key)
						if table == nil {
							other := map[string]string{"cell_rise": "cell_fall", "cell_fall": "cell_rise", "rise_transition": "fall_transition", "fall_transition": "rise_transition"}
							table = tn.child(other[spec.key])
						}
						if table == nil {
							return nil, fmt.Errorf("cell %s pin %s missing %s and paired table", cell.Name, pin.Name, spec.key)
						}
						surface, e := readSurface(table, templates)
						if e != nil {
							return nil, fmt.Errorf("cell %s pin %s %s: %w", cell.Name, pin.Name, spec.key, e)
						}
						*spec.dst = surface
					}
					if len(group.Timings) == 0 {
						group.Timings = append(group.Timings, timing)
					} else {
						first := group.Timings[0]
						for j, a := range []*SC_Surface{&first.CellRise, &first.CellFall, &first.RiseTrans, &first.FallTrans} {
							b := []*SC_Surface{&timing.CellRise, &timing.CellFall, &timing.RiseTrans, &timing.FallTrans}[j]
							mergeWorstSurface(a, b)
						}
					}
				}
				pin.Timings = append(pin.Timings, group)
			}
		}
		lib.Cells = append(lib.Cells, cell)
	}
	initializeIntegerTables(lib)
	Abc_SclHashCells(lib)
	Abc_SclLinkCells(lib)
	Abc_SclLibNormalize(lib)
	return lib, nil
}
func readSurface(n *LibertyItem, templates map[string]*SC_TableTempl) (SC_Surface, error) {
	if n == nil {
		return SC_Surface{}, nil
	}
	s := SC_Surface{Name: n.name()}
	t := templates[s.Name]
	if t != nil && len(t.Index) >= 2 {
		s.Index0 = append([]float32(nil), t.Index[0]...)
		s.Index1 = append([]float32(nil), t.Index[1]...)
	}
	for i := 1; i <= 2; i++ {
		if c := n.child(fmt.Sprintf("index_%d", i)); c != nil {
			v, e := floatVector(strings.Join(c.Arguments, ","))
			if e != nil {
				return s, e
			}
			if i == 1 {
				s.Index0 = v
			} else {
				s.Index1 = v
			}
		}
	}
	values := n.child("values")
	if values == nil {
		return s, fmt.Errorf("missing values")
	}
	v, e := floatVector(strings.Join(values.Arguments, ","))
	if e != nil {
		return s, e
	}
	if len(v) == 1 && len(s.Index0) == 0 && len(s.Index1) == 0 {
		s.Index0 = []float32{0}
		s.Index1 = []float32{0}
	}
	if len(s.Index0) == 0 || len(s.Index1) == 0 || len(v) != len(s.Index0)*len(s.Index1) {
		return s, fmt.Errorf("table dimensions %dx%d do not match %d values", len(s.Index0), len(s.Index1), len(v))
	}

	if len(s.Index0) != 1 || len(s.Index1) != 1 {
		if len(s.Index0) < 2 || len(s.Index1) < 2 {
			return s, fmt.Errorf("non-scalar table requires two samples per axis")
		}
	}
	for _, axis := range [][]float32{s.Index0, s.Index1} {
		for i := 1; i < len(axis); i++ {
			if axis[i] <= axis[i-1] {
				return s, fmt.Errorf("table indices must increase strictly")
			}
		}
	}
	for i := 0; i < len(s.Index0); i++ {
		s.Data = append(s.Data, append([]float32(nil), v[i*len(s.Index1):(i+1)*len(s.Index1)]...))
	}
	// ABC normalizes templates with reversed slew/load axes by transposing.
	if t != nil && len(t.Variables) == 2 && t.Variables[0] == "total_output_net_capacitance" && t.Variables[1] == "input_net_transition" {
		s.Index0, s.Index1 = s.Index1, s.Index0
		data := make([][]float32, len(s.Index0))
		for i := range data {
			data[i] = make([]float32, len(s.Index1))
			for j := range data[i] {
				data[i][j] = s.Data[j][i]
			}
		}
		s.Data = data
	}
	return s, nil
}

func mergeWorstSurface(a, b *SC_Surface) {
	if len(a.Index0) != len(b.Index0) || len(a.Index1) != len(b.Index1) {
		return
	}
	for i, v := range a.Index0 {
		if b.Index0[i] != v {
			return
		}
	}
	for i, v := range a.Index1 {
		if b.Index1[i] != v {
			return
		}
	}
	for i := range a.Data {
		for j := range a.Data[i] {
			a.Data[i][j] = max(a.Data[i][j], b.Data[i][j])
		}
	}
}

func validateNumericAttributes(name string, n *LibertyItem) error {
	numeric := map[string]bool{"area": true, "cell_leakage_power": true, "drive_strength": true, "capacitance": true, "rise_capacitance": true, "fall_capacitance": true, "max_capacitance": true, "max_transition": true, "default_max_transition": true, "slope": true}
	if numeric[n.Key] && n.Value != "" {
		v, e := strconv.ParseFloat(n.Value, 32)
		if e != nil || v != v || v > 3.402823466e38 || v < -3.402823466e38 {
			return fmt.Errorf("%s:%d:%d: invalid %s number %q", name, n.Line, n.Column, n.Key, n.Value)
		}
	}
	for _, c := range n.Children {
		if e := validateNumericAttributes(name, c); e != nil {
			return e
		}
	}
	return nil
}
