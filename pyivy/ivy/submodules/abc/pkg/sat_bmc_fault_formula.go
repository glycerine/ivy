package abc

// Fault-model formulas from src/sat/bmc/bmcFault.c.

import (
	"fmt"
	"slices"
	"strings"
)

type FaultFormulaOptions struct {
	MaxBytes   int
	MaxNodes   int
	MaxObjects int
	Stop       func() bool
}

type faultFormulaKind uint8

const (
	faultFormulaZero faultFormulaKind = iota
	faultFormulaOne
	faultFormulaVariable
	faultFormulaParameter
	faultFormulaNot
	faultFormulaAnd
	faultFormulaOr
	faultFormulaXor
	faultFormulaMux
)

type faultFormulaNode struct {
	kind   faultFormulaKind
	index  int
	first  int
	second int
	third  int
}

type FaultFormula struct {
	nodes      []faultFormulaNode
	root       int
	parameters int
	text       string
}

func faultFormulaLimits(options FaultFormulaOptions) (FaultFormulaOptions, error) {
	if options.MaxBytes < 0 || options.MaxNodes < 0 || options.MaxObjects < 0 {
		return options, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 1 << 20
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 16
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	return options, nil
}

type faultFormulaParser struct {
	text           string
	position       int
	nodes          []faultFormulaNode
	seenVariables  [2]bool
	seenParameters [8]bool
	parameterCount int
	options        FaultFormulaOptions
}

func (p *faultFormulaParser) add(node faultFormulaNode) (int, error) {
	if p.options.Stop != nil && p.options.Stop() {
		return 0, ErrCancelled
	}
	if len(p.nodes) >= p.options.MaxNodes {
		return 0, ErrLimit
	}
	p.nodes = append(p.nodes, node)
	return len(p.nodes) - 1, nil
}

func (p *faultFormulaParser) expression() (int, error) {
	if p.position >= len(p.text) {
		return 0, fmt.Errorf("%w: unexpected end of fault formula", ErrDimension)
	}
	character := p.text[p.position]
	if character == '~' {
		p.position++
		child, err := p.expression()
		if err != nil {
			return 0, err
		}
		return p.add(faultFormulaNode{kind: faultFormulaNot, first: child})
	}
	if character == '0' || character == '1' {
		p.position++
		kind := faultFormulaZero
		if character == '1' {
			kind = faultFormulaOne
		}
		return p.add(faultFormulaNode{kind: kind})
	}
	if character >= 'a' && character <= 'b' {
		p.position++
		index := int(character - 'a')
		p.seenVariables[index] = true
		return p.add(faultFormulaNode{kind: faultFormulaVariable, index: index})
	}
	if character >= 'p' && character <= 'w' {
		p.position++
		index := int(character - 'p')
		p.seenParameters[index] = true
		p.parameterCount = max(p.parameterCount, index+1)
		return p.add(faultFormulaNode{kind: faultFormulaParameter, index: index})
	}
	if character != '(' {
		return 0, fmt.Errorf("%w: unexpected fault-formula symbol %q", ErrDimension, character)
	}
	p.position++
	first, err := p.expression()
	if err != nil {
		return 0, err
	}
	if p.position >= len(p.text) {
		return 0, fmt.Errorf("%w: unclosed fault-formula group", ErrDimension)
	}
	if p.text[p.position] == ')' {
		p.position++
		return first, nil
	}
	operator := p.text[p.position]
	if operator != '&' && operator != '|' && operator != '^' && operator != '?' {
		return 0, fmt.Errorf("%w: unexpected fault-formula operator %q", ErrDimension, operator)
	}
	p.position++
	second, err := p.expression()
	if err != nil {
		return 0, err
	}
	third := -1
	kind := faultFormulaAnd
	switch operator {
	case '|':
		kind = faultFormulaOr
	case '^':
		kind = faultFormulaXor
	case '?':
		kind = faultFormulaMux
		if p.position >= len(p.text) || p.text[p.position] != ':' {
			return 0, fmt.Errorf("%w: conditional fault formula has no false branch", ErrDimension)
		}
		p.position++
		third, err = p.expression()
		if err != nil {
			return 0, err
		}
	}
	if p.position >= len(p.text) || p.text[p.position] != ')' {
		return 0, fmt.Errorf("%w: fault-formula group has trailing or missing syntax", ErrDimension)
	}
	p.position++
	return p.add(faultFormulaNode{kind: kind, first: first, second: second, third: third})
}

func ParseFaultFormula(text string, options FaultFormulaOptions) (FaultFormula, error) {
	options, err := faultFormulaLimits(options)
	if err != nil {
		return FaultFormula{}, err
	}
	if len(text) == 0 || len(text) > options.MaxBytes || text[0] != '(' || text[len(text)-1] != ')' {
		if len(text) > options.MaxBytes {
			return FaultFormula{}, ErrLimit
		}
		return FaultFormula{}, ErrDimension
	}
	p := faultFormulaParser{text: text, options: options}
	root, err := p.expression()
	if err != nil {
		return FaultFormula{}, err
	}
	if p.position != len(text) || !p.seenVariables[0] || !p.seenVariables[1] || p.parameterCount == 0 {
		return FaultFormula{}, ErrDimension
	}
	formula := FaultFormula{nodes: p.nodes, root: root, parameters: p.parameterCount}
	formula.text = formula.render(root)
	return formula, nil
}

func (f FaultFormula) render(index int) string {
	node := f.nodes[index]
	switch node.kind {
	case faultFormulaZero:
		return "0"
	case faultFormulaOne:
		return "1"
	case faultFormulaVariable:
		return string(byte('a' + node.index))
	case faultFormulaParameter:
		return string(byte('p' + node.index))
	case faultFormulaNot:
		return "~" + f.render(node.first)
	case faultFormulaMux:
		return "(" + f.render(node.first) + "?" + f.render(node.second) + ":" + f.render(node.third) + ")"
	default:
		operator := byte('&')
		if node.kind == faultFormulaOr {
			operator = '|'
		} else if node.kind == faultFormulaXor {
			operator = '^'
		}
		return "(" + f.render(node.first) + string(operator) + f.render(node.second) + ")"
	}
}

func (f FaultFormula) String() string      { return f.text }
func (f FaultFormula) ParameterCount() int { return f.parameters }
func (f FaultFormula) NodeCount() int      { return len(f.nodes) }

func (f FaultFormula) Evaluate(variables [2]bool, parameters []bool) (bool, error) {
	if f.root < 0 || f.root >= len(f.nodes) || len(parameters) != f.parameters {
		return false, ErrDimension
	}
	values := make([]bool, len(f.nodes))
	for i, node := range f.nodes {
		switch node.kind {
		case faultFormulaOne:
			values[i] = true
		case faultFormulaVariable:
			values[i] = variables[node.index]
		case faultFormulaParameter:
			values[i] = parameters[node.index]
		case faultFormulaNot:
			values[i] = !values[node.first]
		case faultFormulaAnd:
			values[i] = values[node.first] && values[node.second]
		case faultFormulaOr:
			values[i] = values[node.first] || values[node.second]
		case faultFormulaXor:
			values[i] = values[node.first] != values[node.second]
		case faultFormulaMux:
			values[i] = values[node.third]
			if values[node.first] {
				values[i] = values[node.second]
			}
		}
	}
	return values[f.root], nil
}

func realizeFaultFormula(manager *GiaManager, formula FaultFormula, variables [2]GiaLit, parameters []GiaLit) (GiaLit, error) {
	if formula.root < 0 || formula.root >= len(formula.nodes) || len(parameters) != formula.parameters {
		return InvalidGiaLit, ErrDimension
	}
	values := make([]GiaLit, len(formula.nodes))
	for i, node := range formula.nodes {
		var err error
		switch node.kind {
		case faultFormulaZero:
			values[i] = GiaConst0
		case faultFormulaOne:
			values[i] = GiaConst1
		case faultFormulaVariable:
			values[i] = variables[node.index]
		case faultFormulaParameter:
			values[i] = parameters[node.index]
		case faultFormulaNot:
			values[i] = values[node.first].Not()
		case faultFormulaAnd:
			values[i], err = manager.HashAnd(values[node.first], values[node.second])
		case faultFormulaOr:
			values[i], err = manager.HashOr(values[node.first], values[node.second])
		case faultFormulaXor:
			values[i], err = manager.HashXor(values[node.first], values[node.second])
		case faultFormulaMux:
			values[i], err = manager.HashMux(values[node.first], values[node.second], values[node.third])
		}
		if err != nil {
			return InvalidGiaLit, err
		}
	}
	return values[formula.root], nil
}

func RealizeFaultFormula(manager *GiaManager, formula FaultFormula, variables [2]GiaLit, parameters []GiaLit) (GiaLit, error) {
	if manager == nil {
		return InvalidGiaLit, ErrDimension
	}
	trial := manager.Clone()
	if trial.hash == nil {
		if err := trial.HashStart(); err != nil {
			return InvalidGiaLit, err
		}
	}
	result, err := realizeFaultFormula(trial, formula, variables, slices.Clone(parameters))
	if err != nil {
		return InvalidGiaLit, err
	}
	*manager = *trial
	return result, nil
}

func DeriveGiaFormula(manager *GiaManager, output int, names []string, options FaultFormulaOptions) (string, error) {
	options, err := faultFormulaLimits(options)
	if err != nil || manager == nil || output < 0 || output >= manager.OutputCount() || len(names) != manager.InputCount() {
		return "", ErrDimension
	}
	if err = manager.Check(); err != nil {
		return "", err
	}
	for _, name := range names {
		if name == "" || strings.ContainsAny(name, "()&|^?:~") {
			return "", ErrDimension
		}
	}
	visits, bytes := 0, 0
	var render func(GiaLit) (string, error)
	render = func(literal GiaLit) (string, error) {
		visits++
		if visits > options.MaxNodes || options.Stop != nil && options.Stop() {
			if visits > options.MaxNodes {
				return "", ErrLimit
			}
			return "", ErrCancelled
		}
		complemented := literal.Complemented()
		id := literal.ID()
		object := manager.objects[id]
		var text string
		switch {
		case id == 0:
			text = "0"
		case object.IsCI():
			text = names[object.Diff1()]
		case object.IsBuffer():
			fanin, _ := object.Fanin0(id)
			text, err = render(fanin)
		case object.IsXor():
			a, _ := object.Fanin0(id)
			b, _ := object.Fanin1(id)
			var x, y string
			x, err = render(a)
			if err == nil {
				y, err = render(b)
			}
			text = "(" + x + "^" + y + ")"
		default:
			if control, yes := manager.MuxControl(id); yes {
				falseBranch, _ := object.Fanin0(id)
				trueBranch, _ := object.Fanin1(id)
				var c, x, y string
				c, err = render(control)
				if err == nil {
					x, err = render(trueBranch)
				}
				if err == nil {
					y, err = render(falseBranch)
				}
				text = "(" + c + "?" + x + ":" + y + ")"
			} else {
				a, _ := object.Fanin0(id)
				b, _ := object.Fanin1(id)
				var x, y string
				x, err = render(a)
				if err == nil {
					y, err = render(b)
				}
				text = "(" + x + "&" + y + ")"
			}
		}
		if err != nil {
			return "", err
		}
		if complemented {
			text = "~" + text
		}
		bytes += len(text)
		if bytes > options.MaxBytes {
			return "", ErrLimit
		}
		return text, nil
	}
	co := manager.objects[manager.cos[output]]
	driver, _ := co.Fanin0(manager.cos[output])
	text, err := render(driver)
	if err != nil {
		return "", err
	}
	return "(" + text + ")", nil
}

type GiaFaultFormulaUnfold struct {
	Graph           *GiaManager
	Formula         FaultFormula
	SourceGates     []GiaID
	ParameterInputs [][]GiaID
}

func UnfoldGiaFaultFormula(source *GiaManager, text string, onlyRegisterDrivers bool, options FaultFormulaOptions) (*GiaFaultFormulaUnfold, error) {
	options, err := faultFormulaLimits(options)
	if err != nil || source == nil {
		return nil, ErrDimension
	}
	if err = source.Check(); err != nil {
		return nil, err
	}
	formula, err := ParseFaultFormula(text, options)
	if err != nil {
		return nil, err
	}
	graph, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	graph.Name, graph.Spec = source.Name, source.Spec
	if err = graph.HashStart(); err != nil {
		return nil, err
	}
	selected := make([]bool, source.ObjectCount())
	if onlyRegisterDrivers {
		for i := source.PrimaryOutputCount(); i < source.OutputCount(); i++ {
			co := source.objects[source.cos[i]]
			driver, _ := co.Fanin0(source.cos[i])
			if source.objects[driver.ID()].IsAnd() {
				selected[driver.ID()] = true
			}
		}
	} else {
		for i, object := range source.objects {
			selected[i] = object.IsAnd()
		}
	}
	result := &GiaFaultFormulaUnfold{Graph: graph, Formula: formula}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	gateRow := make([]int, source.ObjectCount())
	for i := range gateRow {
		gateRow[i] = -1
	}
	for i, yes := range selected {
		if yes {
			gateRow[i] = len(result.SourceGates)
			result.SourceGates = append(result.SourceGates, GiaID(i))
			result.ParameterInputs = append(result.ParameterInputs, make([]GiaID, formula.parameters))
		}
	}
	mapping := make([]GiaLit, source.ObjectCount())
	for i := 0; i < source.PrimaryInputCount(); i++ {
		if stopped() {
			return nil, ErrCancelled
		}
		id := source.cis[i]
		mapping[id], err = graph.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	parameterLiterals := make([][]GiaLit, len(result.SourceGates))
	for row := range parameterLiterals {
		parameterLiterals[row] = make([]GiaLit, formula.parameters)
		for k := range parameterLiterals[row] {
			if stopped() {
				return nil, ErrCancelled
			}
			parameterLiterals[row][k], err = graph.AppendCI()
			if err != nil {
				return nil, err
			}
			result.ParameterInputs[row][k] = parameterLiterals[row][k].ID()
		}
	}
	for i := source.PrimaryInputCount(); i < source.InputCount(); i++ {
		if stopped() {
			return nil, ErrCancelled
		}
		id := source.cis[i]
		mapping[id], err = graph.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	copyLiteral := func(literal GiaLit) GiaLit { return mapping[literal.ID()].NotIf(literal.Complemented()) }
	for i, object := range source.objects {
		if stopped() {
			return nil, ErrCancelled
		}
		id := GiaID(i)
		if !object.IsAnd() {
			continue
		}
		if object.IsBuffer() || object.IsXor() {
			return nil, fmt.Errorf("%w: fault-formula unfold requires ordinary AND gates", ErrDimension)
		}
		if _, yes := source.MuxControl(id); yes {
			return nil, fmt.Errorf("%w: fault-formula unfold requires ordinary AND gates", ErrDimension)
		}
		a, _ := object.Fanin0(id)
		b, _ := object.Fanin1(id)
		if selected[id] {
			mapping[id], err = realizeFaultFormula(graph, formula, [2]GiaLit{copyLiteral(a), copyLiteral(b)}, parameterLiterals[gateRow[id]])
		} else {
			mapping[id], err = graph.HashAnd(copyLiteral(a), copyLiteral(b))
		}
		if err != nil {
			return nil, err
		}
	}
	for _, id := range source.cos {
		if stopped() {
			return nil, ErrCancelled
		}
		object := source.objects[id]
		driver, _ := object.Fanin0(id)
		if _, err = graph.AppendCO(copyLiteral(driver)); err != nil {
			return nil, err
		}
	}
	if err = graph.SetRegisterCount(source.RegisterCount()); err != nil {
		return nil, err
	}
	if err = graph.SetConstraintCount(source.ConstraintCount()); err != nil {
		return nil, err
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	return result, nil
}
