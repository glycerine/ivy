// BLIF and BLIF-MV hierarchical netlist reader.
package io

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
)

type blifLine struct {
	number int
	text   string
}
type blifModel struct {
	network      *abc.Network
	lines        []blifLine
	nets         map[string]*abc.Object
	domains      map[string]*abc.MVVariable
	instantiated bool
	resetFlag    string
}

func blifLines(name, text string) ([]blifLine, error) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 4096), 16<<20)
	var result []blifLine
	var joined strings.Builder
	start, line := 1, 0
	for scanner.Scan() {
		line++
		part := scanner.Text()
		if at := strings.IndexByte(part, '#'); at >= 0 {
			part = part[:at]
		}
		part = strings.TrimSpace(part)
		if joined.Len() == 0 {
			start = line
		}
		continuation := strings.HasSuffix(part, "\\")
		part = strings.TrimSuffix(part, "\\")
		joined.WriteString(part)
		if continuation {
			joined.WriteByte(' ')
			continue
		}
		if value := strings.TrimSpace(joined.String()); value != "" {
			result = append(result, blifLine{start, value})
		}
		joined.Reset()
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s:%d: %w", name, line, err)
	}
	if joined.Len() > 0 {
		return nil, fmt.Errorf("%s:%d: unterminated line continuation", name, start)
	}
	return result, nil
}
func blifError(name string, line blifLine, format string, args ...any) error {
	return fmt.Errorf("%s:%d: %s", name, line.number, fmt.Sprintf(format, args...))
}
func (m *blifModel) net(name string) *abc.Object {
	if o := m.nets[name]; o != nil {
		return o
	}
	o := m.network.NewObject(abc.ObjectNet)
	o.SetName(name)
	m.nets[name] = o
	return o
}
func (m *blifModel) connect(out, in *abc.Object) error { return m.network.AddFanin(out, in.Signal()) }
func (m *blifModel) input(name string) error {
	net := m.net(name)
	if len(net.Fanins()) > 0 {
		return fmt.Errorf("duplicate driver for %s", name)
	}
	pi := m.network.NewObject(abc.ObjectPI)
	pi.SetName(name)
	return m.connect(net, pi)
}
func (m *blifModel) output(name string) error {
	for _, o := range m.network.ObjectsOfKind(abc.ObjectPO) {
		if o.Name() == name {
			return fmt.Errorf("duplicate output %s", name)
		}
	}
	po := m.network.NewObject(abc.ObjectPO)
	po.SetName(name)
	return m.connect(po, m.net(name))
}
func (m *blifModel) node(names []string, data any) (*abc.Object, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("node missing output")
	}
	output := m.net(names[len(names)-1])
	if len(output.Fanins()) > 0 {
		return nil, fmt.Errorf("duplicate driver for %s", output.Name())
	}
	node := m.network.NewObject(abc.ObjectNode)
	node.Data = data
	for _, name := range names[:len(names)-1] {
		if err := m.connect(node, m.net(name)); err != nil {
			return nil, err
		}
	}
	return node, m.connect(output, node)
}
func (m *blifModel) domain(name string) int {
	if d := m.domains[name]; d != nil {
		return d.Values
	}
	return 2
}
func parseMvDeclaration(m *blifModel, fields []string) error {
	// Commas join multiple signals before the first numeric domain size.
	fields = strings.Fields(strings.ReplaceAll(strings.Join(fields, " "), ",", " , "))
	sizeAt := -1
	size := 0
	for i, f := range fields {
		if n, err := strconv.Atoi(f); err == nil {
			sizeAt = i
			size = n
			break
		}
	}
	if sizeAt < 1 || size < 2 || size > 256 {
		return fmt.Errorf("invalid .mv domain (expected 2..256)")
	}
	var names []string
	for _, f := range fields[:sizeAt] {
		if f != "," {
			names = append(names, f)
		}
	}
	labels := fields[sizeAt+1:]
	if len(labels) != 0 && len(labels) != size {
		return fmt.Errorf("expected %d symbolic values, got %d", size, len(labels))
	}
	seen := map[string]bool{}
	for _, label := range labels {
		if seen[label] {
			return fmt.Errorf("duplicate symbolic value %s", label)
		}
		seen[label] = true
	}
	for _, name := range names {
		if m.domains[name] != nil {
			return fmt.Errorf("duplicate domain for %s", name)
		}
		m.domains[name] = &abc.MVVariable{Values: size, Names: append([]string(nil), labels...)}
		m.net(name)
	}
	return nil
}
func blifValueFields(line string) ([]string, error) {
	var fields []string
	start, depth := -1, 0
	for i, c := range line {
		switch c {
		case '(', '{':
			depth++
		case ')', '}':
			depth--
			if depth < 0 {
				return nil, fmt.Errorf("unbalanced value set")
			}
		}
		if (c == ' ' || c == '\t') && depth == 0 {
			if start >= 0 {
				fields = append(fields, line[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if depth != 0 {
		return nil, fmt.Errorf("unbalanced value set")
	}
	if start >= 0 {
		fields = append(fields, line[start:])
	}
	return fields, nil
}
func normalizeMvTerm(text string, index int, names []string, m *blifModel) (string, error) {
	if text == "-" {
		return text, nil
	}
	if strings.HasPrefix(text, "=") {
		ref := strings.TrimPrefix(text, "=")

		for i, name := range names[:len(names)-1] {
			if name == ref {
				return fmt.Sprintf("=%d", i), nil
			}
		}
		if n, err := strconv.Atoi(ref); err == nil {
			if n < 0 || n >= len(names)-1 {
				return "", fmt.Errorf("invalid equality %s", text)
			}
			return text, nil
		}
		return "", fmt.Errorf("unknown equality signal %s", ref)
	}
	prefix := ""
	if strings.HasPrefix(text, "!") {
		prefix = "!"
		text = text[1:]
	}
	text = strings.TrimSpace(text)
	if len(text) >= 2 && ((text[0] == '(' && text[len(text)-1] == ')') || (text[0] == '{' && text[len(text)-1] == '}')) {
		text = text[1 : len(text)-1]
	}
	parts := strings.Split(text, ",")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		domain := m.domains[names[index]]
		if domain == nil || len(domain.Names) == 0 {
			if _, err := strconv.Atoi(part); err == nil {
				parts[i] = part
				continue
			}
		}
		found := -1
		if domain != nil {
			for j, label := range domain.Names {
				if label == part {
					found = j
					break
				}
			}
		}
		if found < 0 {
			return "", fmt.Errorf("unknown value %s for %s", part, names[index])
		}
		parts[i] = strconv.Itoa(found)
	}
	if len(parts) > 1 {
		return prefix + "(" + strings.Join(parts, ",") + ")", nil
	}
	return prefix + parts[0], nil
}
func mvCover(m *blifModel, names []string, rows []blifLine) (*abc.MVTable, error) {
	domains := make([]int, len(names))
	for i, name := range names {
		domains[i] = m.domain(name)
	}
	var text strings.Builder
	for _, row := range rows {
		line := row.text
		defaultRow := strings.HasPrefix(line, ".default") || strings.HasPrefix(line, ".def ")
		if defaultRow {
			line = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, ".default"), ".def"))
		}
		fields, err := blifValueFields(line)
		if err != nil {
			return nil, err
		}
		if defaultRow {
			if len(fields) != 1 {
				return nil, fmt.Errorf("invalid default row")
			}
			term, err := normalizeMvTerm(fields[0], len(names)-1, names, m)
			if err != nil {
				return nil, err
			}
			fmt.Fprintln(&text, "d", term)
			continue
		}
		if len(fields) != len(names) {
			return nil, fmt.Errorf("expected %d table terms, got %d", len(names), len(fields))
		}
		for i, field := range fields {
			term, err := normalizeMvTerm(field, i, names, m)
			if err != nil {
				return nil, err
			}
			if i > 0 {
				text.WriteByte(' ')
			}
			text.WriteString(term)
		}
		text.WriteByte('\n')
	}
	return abc.ParseMVTable(text.String(), domains)
}

// ParseBLIF reads all models before resolving instances, so forward module
// references retain their original port order and shared model identity.
func ParseBLIF(name, text string, multiValued, check bool) (*abc.Network, error) {
	lines, err := blifLines(name, text)
	if err != nil {
		return nil, err
	}
	design := abc.NewDesign(name)
	var models []*blifModel
	byName := map[string]*blifModel{}
	var current *blifModel
	for _, line := range lines {
		f := strings.Fields(strings.ReplaceAll(line.text, "->", " -> "))
		switch f[0] {
		case ".model":
			if len(f) != 2 || current != nil {
				return nil, blifError(name, line, "invalid or nested .model")
			}
			if byName[f[1]] != nil {
				return nil, blifError(name, line, "duplicate model %s", f[1])
			}
			n := abc.NewNetwork(f[1])
			n.Spec = name
			n.Type = uint8(abc.ABC_NTK_NETLIST)
			n.Function = uint8(abc.ABC_FUNC_SOP)
			if multiValued {
				n.Function = uint8(abc.ABC_FUNC_BLIFMV)
			}
			current = &blifModel{network: n, nets: map[string]*abc.Object{}, domains: map[string]*abc.MVVariable{}}
			models = append(models, current)
			byName[f[1]] = current
			if err := design.AddModule(n); err != nil {
				return nil, err
			}
		case ".end":
			if current == nil {
				return nil, blifError(name, line, ".end outside model")
			}
			current = nil
		default:
			if current == nil {
				return nil, blifError(name, line, "content outside .model")
			}
			current.lines = append(current.lines, line)
		}
	}
	if current != nil {
		return nil, fmt.Errorf("%s: model %s has no .end", name, current.network.Name)
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("%s: no models", name)
	}
	// Declare interfaces and domains first, including those of forward models.
	for _, m := range models {
		for _, line := range m.lines {
			f := strings.Fields(strings.ReplaceAll(line.text, "->", " -> "))
			var err error
			switch f[0] {
			case ".inputs":
				for _, s := range f[1:] {
					if err = m.input(s); err != nil {
						break
					}
				}
			case ".outputs":
				for _, s := range f[1:] {
					if err = m.output(s); err != nil {
						break
					}
				}
			case ".mv":
				if !multiValued {
					return nil, blifError(name, line, ".mv requires BLIF-MV mode")
				}
				err = parseMvDeclaration(m, f[1:])
			case ".blackbox":
				m.network.Function = uint8(abc.ABC_FUNC_BLACKBOX)
			}
			if err != nil {
				return nil, blifError(name, line, "%v", err)
			}
		}
	}
	for _, m := range models {
		for _, line := range m.lines {
			if strings.HasPrefix(line.text, ".reset ") || strings.HasPrefix(line.text, ".r ") {
				if !multiValued {
					return nil, blifError(name, line, "reset tables require BLIF-MV")
				}
				m.resetFlag = "__blif_reset_active"
				if m.nets[m.resetFlag] != nil || m.nets[m.resetFlag+"_one"] != nil {
					return nil, blifError(name, line, "reset helper name collision")
				}
				if _, err = m.node([]string{m.resetFlag + "_one"}, &abc.MVTable{DomainSizes: []int{2}, Cubes: []abc.MVCube{{Output: abc.MVTerm{EqualInput: -1, Values: []int{1}}}}}); err != nil {
					return nil, err
				}
				if err = blifLatch(m, []string{m.resetFlag + "_one", m.resetFlag, "0"}); err != nil {
					return nil, err
				}
				break
			}
		}
		for i := 0; i < len(m.lines); i++ {
			line := m.lines[i]
			f := strings.Fields(strings.ReplaceAll(line.text, "->", " -> "))
			var err error
			switch f[0] {
			case ".inputs", ".outputs", ".mv", ".blackbox":
				continue
			case ".names", ".table", ".reset", ".r":
				names := f[1:]
				if len(names) == 0 {
					return nil, blifError(name, line, "missing table output")
				}
				var rows []blifLine
				for i+1 < len(m.lines) {
					next := m.lines[i+1]
					if strings.HasPrefix(next.text, ".") && !strings.HasPrefix(next.text, ".default") && !strings.HasPrefix(next.text, ".def ") {
						break
					}
					i++
					rows = append(rows, next)
				}
				if f[0] == ".reset" || f[0] == ".r" {
					err = blifResetTable(m, names, rows)
				} else {
					err = blifTables(m, names, rows, multiValued)
				}

			case ".subckt":
				err = blifInstance(m, f[1:], byName, multiValued)
			case ".latch":
				latchFields := append([]string(nil), f[1:]...)
				if multiValued && len(latchFields) >= 2 {
					if m.resetFlag != "" {
						original := latchFields[1]
						latchFields[1] += "_out"
						if domain := m.domains[original]; domain != nil {
							m.domains[latchFields[1]] = domain
						}
					}
					latchFields = append(latchFields[:2], "0")
				}
				err = blifLatch(m, latchFields)
			default:
				err = fmt.Errorf("unsupported BLIF directive %s", f[0])
			}
			if err != nil {
				return nil, blifError(name, line, "%v", err)
			}
		}
		for signal, domain := range m.domains {
			if err := abc.SetMVVariable(m.net(signal), domain); err != nil {
				return nil, err
			}
		}
		if check {
			for signal, net := range m.nets {
				if len(net.Fanins()) == 0 && m.network.Function != uint8(abc.ABC_FUNC_BLACKBOX) {
					return nil, fmt.Errorf("%s: model %s signal %s has no driver", name, m.network.Name, signal)
				}
			}
			if err := m.network.Check(); err != nil {
				return nil, err
			}
			roots := append(m.network.ObjectsOfKind(abc.ObjectPO), m.network.ObjectsOfKind(abc.ObjectBI)...)
			if _, err := m.network.Topological(roots); err != nil {
				return nil, fmt.Errorf("%s: model %s: %w", name, m.network.Name, err)
			}
		}
	}
	for _, m := range models {
		if !m.instantiated {
			design.Tops = append(design.Tops, m.network)
		}
	}
	if check {
		state := map[*abc.Network]byte{}
		var visit func(*abc.Network) error
		visit = func(n *abc.Network) error {
			if state[n] == 1 {
				return fmt.Errorf("%s: recursive model %s", name, n.Name)
			}
			if state[n] == 2 {
				return nil
			}
			state[n] = 1
			for _, box := range n.Boxes() {
				if model, ok := box.Data.(*abc.Network); ok {
					if err := visit(model); err != nil {
						return err
					}
				}
			}
			state[n] = 2
			return nil
		}
		for _, m := range models {
			if err := visit(m.network); err != nil {
				return nil, err
			}
		}
	}
	if len(design.Tops) == 0 {
		return nil, fmt.Errorf("%s: recursive hierarchy has no top", name)
	}
	return design.Tops[0], nil
}
func binaryBlifCover(inputs int, rows []blifLine) (string, error) {
	if len(rows) == 0 {
		return " 0\n", nil
	}
	var cover strings.Builder
	polarity := ""
	for _, row := range rows {
		f := strings.Fields(row.text)
		pattern, value := "", ""
		if inputs == 0 && len(f) == 1 {
			value = f[0]
		} else if len(f) == 2 {
			pattern, value = f[0], f[1]
		} else {
			return "", fmt.Errorf("invalid binary cube %s", row.text)
		}
		if len(pattern) != inputs || strings.Trim(pattern, "01-") != "" || (value != "0" && value != "1") {
			return "", fmt.Errorf("invalid binary cube %s", row.text)
		}
		if polarity != "" && polarity != value {
			return "", fmt.Errorf("mixed output polarities in binary cover")
		}
		polarity = value
		fmt.Fprintf(&cover, "%s %s\n", pattern, value)
	}
	return cover.String(), nil
}
func blifLatch(m *blifModel, f []string) error {
	if len(f) < 2 || len(f) > 5 {
		return fmt.Errorf("invalid .latch")
	}
	input, output := m.net(f[0]), m.net(f[1])
	if len(output.Fanins()) > 0 {
		return fmt.Errorf("duplicate latch output")
	}
	bi := m.network.NewObject(abc.ObjectBI)
	latch := m.network.NewObject(abc.ObjectLatch)
	bo := m.network.NewObject(abc.ObjectBO)
	latch.SetName(f[1] + "L")
	if err := m.connect(bi, input); err != nil {
		return err
	}
	if err := m.connect(latch, bi); err != nil {
		return err
	}
	if err := m.connect(bo, latch); err != nil {
		return err
	}
	if err := m.connect(output, bo); err != nil {
		return err
	}
	init := "2"
	if len(f) == 3 || len(f) == 5 {
		init = f[len(f)-1]
	}
	switch init {
	case "0":
		abc.Abc_LatchSetInit0(latch)
	case "1":
		abc.Abc_LatchSetInit1(latch)
	case "2", "3":
		abc.Abc_LatchSetInitDc(latch)
	default:
		return fmt.Errorf("invalid latch initial value %s", init)
	}
	return nil
}
func blifInstance(m *blifModel, f []string, models map[string]*blifModel, mv bool) error {
	if len(f) == 0 {
		return fmt.Errorf("missing subcircuit model")
	}
	modelName := strings.SplitN(f[0], "|", 2)[0]
	model := models[modelName]
	if model == nil {
		return fmt.Errorf("unknown model %s", modelName)
	}
	model.instantiated = true
	bindings := map[string]string{}
	instance := ""
	start := 1
	if mv && len(f) > 1 && !strings.Contains(f[1], "=") {
		instance = f[1]
		start = 2
	}
	for _, binding := range f[start:] {
		pair := strings.SplitN(binding, "=", 2)
		if len(pair) != 2 || pair[0] == "" || pair[1] == "" {
			return fmt.Errorf("invalid port binding %s", binding)
		}
		if bindings[pair[0]] != "" {
			return fmt.Errorf("duplicate port binding %s", pair[0])
		}
		bindings[pair[0]] = pair[1]
	}
	kind := abc.ObjectWhitebox
	if model.network.Function == uint8(abc.ABC_FUNC_BLACKBOX) {
		kind = abc.ObjectBlackbox
	}
	box := m.network.NewObject(kind)
	box.Data = model.network
	if instance != "" {
		box.SetName(instance)
	}
	for _, port := range model.network.ObjectsOfKind(abc.ObjectPI) {
		signal, ok := bindings[port.Name()]
		if !ok {
			signal = m.uniqueNetName("__unconnected_input")
			var data any = " 0\n"
			if mv {
				domain := model.domain(port.Name())
				data = &abc.MVTable{DomainSizes: []int{domain}, Cubes: []abc.MVCube{{Output: abc.MVTerm{EqualInput: -1, Values: []int{0}}}}}
				if v := model.domains[port.Name()]; v != nil {
					m.domains[signal] = v
				}
			}
			if _, err := m.node([]string{signal}, data); err != nil {
				return err
			}
		}
		delete(bindings, port.Name())
		bi := m.network.NewObject(abc.ObjectBI)
		if err := m.connect(bi, m.net(signal)); err != nil {
			return err
		}
		if err := m.connect(box, bi); err != nil {
			return err
		}
	}
	for _, port := range model.network.ObjectsOfKind(abc.ObjectPO) {
		signal, ok := bindings[port.Name()]
		if !ok {
			signal = m.uniqueNetName("__unconnected_output")
		}
		delete(bindings, port.Name())
		net := m.net(signal)
		if len(net.Fanins()) > 0 {
			return fmt.Errorf("duplicate subcircuit output driver %s", signal)
		}
		bo := m.network.NewObject(abc.ObjectBO)
		if err := m.connect(bo, box); err != nil {
			return err
		}
		if err := m.connect(net, bo); err != nil {
			return err
		}
	}
	if len(bindings) > 0 {
		return fmt.Errorf("unknown subcircuit ports")
	}
	return nil
}
func ReadBLIF(filename string, multiValued, check bool) (*abc.Network, error) {
	text, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseBLIF(filename, string(text), multiValued, check)
}
func Io_ReadBlifMv(filename string, multiValued, check int32) *abc.Network {
	n, err := ReadBLIF(filename, multiValued != 0, check != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}
func init() { abc.SetHierarchyReader(ReadBLIF) }

func blifTables(m *blifModel, names []string, rows []blifLine, mv bool) error {
	arrow := -1
	for i, name := range names {
		if name == "->" {
			if arrow >= 0 {
				return fmt.Errorf("duplicate output arrow")
			}
			arrow = i
		}
	}
	if arrow < 0 {
		var data any
		var err error
		if mv {
			data, err = mvCover(m, names, rows)
		} else {
			data, err = binaryBlifCover(len(names)-1, rows)
		}
		if err != nil {
			return err
		}
		_, err = m.node(names, data)
		return err
	}
	if !mv {
		return fmt.Errorf("multi-output tables require BLIF-MV")
	}
	inputs, outputs := names[:arrow], names[arrow+1:]
	if len(outputs) == 0 {
		return fmt.Errorf("table has no outputs")
	}
	projected := make([][]blifLine, len(outputs))
	for _, row := range rows {
		defaultRow := strings.HasPrefix(row.text, ".default") || strings.HasPrefix(row.text, ".def ")
		line := row.text
		if defaultRow {
			line = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, ".default"), ".def"))
		}
		fields, err := blifValueFields(line)
		if err != nil {
			return err
		}
		count := len(inputs) + len(outputs)
		if defaultRow {
			count = len(outputs)
		}
		if len(fields) != count {
			return fmt.Errorf("expected %d terms for multi-output table", count)
		}
		for i := range outputs {
			var value string
			if defaultRow {
				value = ".default " + fields[i]
			} else {
				value = strings.Join(append(append([]string(nil), fields[:len(inputs)]...), fields[len(inputs)+i]), " ")
			}
			projected[i] = append(projected[i], blifLine{row.number, value})
		}
	}
	for i, output := range outputs {
		signals := append(append([]string(nil), inputs...), output)
		data, err := mvCover(m, signals, projected[i])
		if err != nil {
			return err
		}
		if _, err = m.node(signals, data); err != nil {
			return err
		}
	}
	return nil
}

func blifResetTable(m *blifModel, names []string, rows []blifLine) error {
	if len(names) == 0 {
		return fmt.Errorf("reset table missing output")
	}
	arrow := -1
	for i, name := range names {
		if name == "->" {
			arrow = i
		}
	}
	start := len(names) - 1
	if arrow >= 0 {
		start = arrow + 1
	}
	resetNames := append([]string(nil), names...)
	for i := start; i < len(names); i++ {
		output := names[i]
		if output == "->" {
			continue
		}
		resetNames[i] = output + "_reset"
		domain := m.domain(output)
		if variable := m.domains[output]; variable != nil {
			m.domains[output+"_reset"] = variable
			m.domains[output+"_out"] = variable
		}
		table, err := abc.ParseMVTable("1 - - =1\n0 - - =2\n", []int{2, domain, domain, domain})
		if err != nil {
			return err
		}
		if _, err = m.node([]string{m.resetFlag, output + "_out", output + "_reset", output}, table); err != nil {
			return err
		}
	}
	return blifTables(m, resetNames, rows, true)
}

func (m *blifModel) uniqueNetName(prefix string) string {
	for i := 0; ; i++ {
		name := fmt.Sprintf("%s_%d", prefix, i)
		if m.nets[name] == nil {
			m.net(name)
			return name
		}
	}
}
