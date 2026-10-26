package abc

// Structural Verilog parsing from src/base/ver/ver.h, verCore.c,
// verFormula.c, verParse.c, and verStream.c.

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

var (
	ErrVerilogSyntax    = errors.New("invalid structural Verilog syntax")
	ErrVerilogDuplicate = errors.New("duplicate Verilog declaration")
	ErrVerilogUndefined = errors.New("undefined Verilog signal or module")
	ErrVerilogDriven    = errors.New("Verilog signal has multiple drivers")
	ErrVerilogEOF       = errors.New("end of Verilog input")
)

type VerilogLimits struct {
	MaxBytes      int
	MaxWordBytes  int
	MaxTokens     int
	MaxModules    int
	MaxNets       int
	MaxStatements int
	MaxExprNodes  int
	MaxVectorBits int
	Stop          func() bool
}

func normalizeVerilogLimits(limits VerilogLimits) (VerilogLimits, error) {
	values := []*int{&limits.MaxBytes, &limits.MaxWordBytes, &limits.MaxTokens, &limits.MaxModules, &limits.MaxNets, &limits.MaxStatements, &limits.MaxExprNodes, &limits.MaxVectorBits}
	defaults := []int{1 << 30, 1 << 20, 1 << 24, 1 << 18, 1 << 24, 1 << 24, 1 << 24, 1 << 20}
	for i := range values {
		if *values[i] < 0 {
			return limits, ErrDimension
		}
		if *values[i] == 0 {
			*values[i] = defaults[i]
		}
	}
	return limits, nil
}

func verilogStopped(limits VerilogLimits) bool { return limits.Stop != nil && limits.Stop() }

// VerilogStream is an owned, checked replacement for Ver_Stream_t. It keeps
// one-based line numbers and byte positions, but never lends its backing data.
type VerilogStream struct {
	name   string
	data   []byte
	pos    int
	line   int
	limits VerilogLimits
	closed bool
}

func NewVerilogStream(name string, data []byte, limits VerilogLimits) (*VerilogStream, error) {
	limits, err := normalizeVerilogLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	if verilogStopped(limits) {
		return nil, ErrCancelled
	}
	owned := append([]byte(nil), data...)
	return &VerilogStream{name: name, data: owned, line: 1, limits: limits}, nil
}

func OpenVerilogStream(path string, limits VerilogLimits) (*VerilogStream, error) {
	limits, err := normalizeVerilogLimits(limits)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > int64(limits.MaxBytes) {
		return nil, ErrLimit
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewVerilogStream(path, data, limits)
}

func (stream *VerilogStream) check() error {
	if stream == nil || stream.closed || stream.pos < 0 || stream.pos > len(stream.data) || stream.line < 1 {
		return ErrInvariant
	}
	return nil
}

func (stream *VerilogStream) FileName() (string, error) {
	if err := stream.check(); err != nil {
		return "", err
	}
	return stream.name, nil
}

func (stream *VerilogStream) FileSize() (int, error) {
	if err := stream.check(); err != nil {
		return 0, err
	}
	return len(stream.data), nil
}

func (stream *VerilogStream) Position() (int, error) {
	if err := stream.check(); err != nil {
		return 0, err
	}
	return stream.pos, nil
}

func (stream *VerilogStream) Line() (int, error) {
	if err := stream.check(); err != nil {
		return 0, err
	}
	return stream.line, nil
}

func (stream *VerilogStream) OK() bool {
	return stream != nil && !stream.closed && stream.pos < len(stream.data)
}

func (stream *VerilogStream) Scan() (byte, error) {
	if err := stream.check(); err != nil {
		return 0, err
	}
	if verilogStopped(stream.limits) {
		return 0, ErrCancelled
	}
	if stream.pos == len(stream.data) {
		return 0, ErrVerilogEOF
	}
	return stream.data[stream.pos], nil
}

func (stream *VerilogStream) Pop() (byte, error) {
	value, err := stream.Scan()
	if err != nil {
		return 0, err
	}
	stream.pos++
	if value == '\n' {
		stream.line++
	}
	return value, nil
}

func byteInSet(value byte, set string) bool { return strings.IndexByte(set, value) >= 0 }

func (stream *VerilogStream) SkipChars(set string) error {
	if err := stream.check(); err != nil {
		return err
	}
	if set == "" {
		return ErrDimension
	}
	for stream.pos < len(stream.data) && byteInSet(stream.data[stream.pos], set) {
		if _, err := stream.Pop(); err != nil {
			return err
		}
	}
	return nil
}

func (stream *VerilogStream) SkipToChars(set string) error {
	if err := stream.check(); err != nil {
		return err
	}
	if set == "" {
		return ErrDimension
	}
	for stream.pos < len(stream.data) && !byteInSet(stream.data[stream.pos], set) {
		if _, err := stream.Pop(); err != nil {
			return err
		}
	}
	return nil
}

func (stream *VerilogStream) Word(stops string) (string, error) {
	if err := stream.check(); err != nil {
		return "", err
	}
	if stops == "" {
		return "", ErrDimension
	}
	start := stream.pos
	for stream.pos < len(stream.data) && !byteInSet(stream.data[stream.pos], stops) {
		if stream.pos-start == stream.limits.MaxWordBytes {
			return "", ErrLimit
		}
		if _, err := stream.Pop(); err != nil {
			return "", err
		}
	}
	return string(stream.data[start:stream.pos]), nil
}

func (stream *VerilogStream) Close() error {
	if stream == nil {
		return ErrInvariant
	}
	if stream.closed {
		return nil
	}
	stream.data = nil
	stream.pos = 0
	stream.closed = true
	return nil
}

type VerilogExprKind uint8

const (
	VerilogExprConst VerilogExprKind = iota
	VerilogExprName
	VerilogExprNot
	VerilogExprAnd
	VerilogExprXor
	VerilogExprOr
	VerilogExprMux
)

type VerilogExpr struct {
	Kind  VerilogExprKind
	Name  string
	Value bool
	Args  []*VerilogExpr
}

func (expr *VerilogExpr) Clone() *VerilogExpr {
	if expr == nil {
		return nil
	}
	clone := &VerilogExpr{Kind: expr.Kind, Name: expr.Name, Value: expr.Value, Args: make([]*VerilogExpr, len(expr.Args))}
	for i := range expr.Args {
		clone.Args[i] = expr.Args[i].Clone()
	}
	return clone
}

func (expr *VerilogExpr) Eval(values map[string]bool) (bool, error) {
	if expr == nil {
		return false, ErrInvariant
	}
	switch expr.Kind {
	case VerilogExprConst:
		return expr.Value, nil
	case VerilogExprName:
		value, ok := values[expr.Name]
		if !ok {
			return false, ErrVerilogUndefined
		}
		return value, nil
	case VerilogExprNot:
		if len(expr.Args) != 1 {
			return false, ErrInvariant
		}
		value, err := expr.Args[0].Eval(values)
		return !value, err
	case VerilogExprAnd, VerilogExprXor, VerilogExprOr:
		if len(expr.Args) < 1 {
			return false, ErrInvariant
		}
		value, err := expr.Args[0].Eval(values)
		if err != nil {
			return false, err
		}
		for _, arg := range expr.Args[1:] {
			next, err := arg.Eval(values)
			if err != nil {
				return false, err
			}
			if expr.Kind == VerilogExprAnd {
				value = value && next
			} else if expr.Kind == VerilogExprOr {
				value = value || next
			} else {
				value = value != next
			}
		}
		return value, nil
	case VerilogExprMux:
		if len(expr.Args) != 3 {
			return false, ErrInvariant
		}
		selectValue, err := expr.Args[0].Eval(values)
		if err != nil {
			return false, err
		}
		if selectValue {
			return expr.Args[1].Eval(values)
		}
		return expr.Args[2].Eval(values)
	default:
		return false, ErrInvariant
	}
}

type verilogFormulaParser struct {
	tokens []string
	pos    int
	nodes  int
	limits VerilogLimits
	vars   []string
	seen   map[string]bool
}

func formulaTokens(text string, limits VerilogLimits) ([]string, error) {
	if len(text) > limits.MaxBytes {
		return nil, ErrLimit
	}
	tokens := make([]string, 0, 16)
	for pos := 0; pos < len(text); {
		if verilogStopped(limits) {
			return nil, ErrCancelled
		}
		if strings.ContainsRune(" \t\r\n", rune(text[pos])) {
			pos++
			continue
		}
		if strings.ContainsRune("()!~&|^?:{},", rune(text[pos])) {
			tokens = append(tokens, text[pos:pos+1])
			pos++
			continue
		}
		start := pos
		if text[pos] == '\\' {
			pos++
			for pos < len(text) && !strings.ContainsRune(" \t\r\n", rune(text[pos])) {
				pos++
			}
		} else {
			for pos < len(text) && !strings.ContainsRune(" \t\r\n()!~&|^?:{},", rune(text[pos])) {
				pos++
			}
		}
		if pos-start > limits.MaxWordBytes {
			return nil, ErrLimit
		}
		tokens = append(tokens, text[start:pos])
		if len(tokens) > limits.MaxTokens {
			return nil, ErrLimit
		}
	}
	return tokens, nil
}

func newFormulaNode(parser *verilogFormulaParser, kind VerilogExprKind, args ...*VerilogExpr) (*VerilogExpr, error) {
	parser.nodes++
	if parser.nodes > parser.limits.MaxExprNodes {
		return nil, ErrLimit
	}
	return &VerilogExpr{Kind: kind, Args: args}, nil
}

func (parser *verilogFormulaParser) primary() (*VerilogExpr, error) {
	if parser.pos >= len(parser.tokens) {
		return nil, ErrVerilogSyntax
	}
	token := parser.tokens[parser.pos]
	parser.pos++
	if token == "!" || token == "~" {
		arg, err := parser.primary()
		if err != nil {
			return nil, err
		}
		return newFormulaNode(parser, VerilogExprNot, arg)
	}
	if token == "(" {
		expr, err := parser.expression(0)
		if err != nil {
			return nil, err
		}
		if parser.pos >= len(parser.tokens) || parser.tokens[parser.pos] != ")" {
			return nil, ErrVerilogSyntax
		}
		parser.pos++
		return expr, nil
	}
	if token == "0" || token == "1'b0" || token == "1'h0" || token == "1'bx" {
		expr, err := newFormulaNode(parser, VerilogExprConst)
		if err != nil {
			return nil, err
		}
		expr.Value = false
		return expr, nil
	}
	if token == "1" || token == "1'b1" || token == "1'h1" {
		expr, err := newFormulaNode(parser, VerilogExprConst)
		if err != nil {
			return nil, err
		}
		expr.Value = true
		return expr, nil
	}
	if strings.ContainsAny(token, "=;[]{}") && !strings.Contains(token, "[") {
		return nil, ErrVerilogSyntax
	}
	expr, err := newFormulaNode(parser, VerilogExprName)
	if err != nil {
		return nil, err
	}
	expr.Name = strings.TrimPrefix(token, "\\")
	if !parser.seen[expr.Name] {
		parser.seen[expr.Name] = true
		parser.vars = append(parser.vars, expr.Name)
	}
	return expr, nil
}

func formulaPrecedence(token string) int {
	switch token {
	case "&":
		return 3
	case "^":
		return 2
	case "|":
		return 1
	}
	return -1
}

func formulaKind(token string) VerilogExprKind {
	switch token {
	case "&":
		return VerilogExprAnd
	case "^":
		return VerilogExprXor
	default:
		return VerilogExprOr
	}
}

func (parser *verilogFormulaParser) expression(minimum int) (*VerilogExpr, error) {
	left, err := parser.primary()
	if err != nil {
		return nil, err
	}
	for parser.pos < len(parser.tokens) {
		token := parser.tokens[parser.pos]
		precedence := formulaPrecedence(token)
		if precedence < minimum {
			break
		}
		parser.pos++
		right, err := parser.expression(precedence + 1)
		if err != nil {
			return nil, err
		}
		left, err = newFormulaNode(parser, formulaKind(token), left, right)
		if err != nil {
			return nil, err
		}
	}
	if minimum == 0 && parser.pos < len(parser.tokens) && parser.tokens[parser.pos] == "?" {
		parser.pos++
		whenTrue, err := parser.expression(0)
		if err != nil {
			return nil, err
		}
		if parser.pos >= len(parser.tokens) || parser.tokens[parser.pos] != ":" {
			return nil, ErrVerilogSyntax
		}
		parser.pos++
		whenFalse, err := parser.expression(0)
		if err != nil {
			return nil, err
		}
		left, err = newFormulaNode(parser, VerilogExprMux, left, whenTrue, whenFalse)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func ParseVerilogFormula(text string, limits VerilogLimits) (*VerilogExpr, []string, error) {
	limits, err := normalizeVerilogLimits(limits)
	if err != nil {
		return nil, nil, err
	}
	if len(text) > limits.MaxBytes {
		return nil, nil, ErrLimit
	}
	if verilogStopped(limits) {
		return nil, nil, ErrCancelled
	}
	tokens, err := formulaTokens(text, limits)
	if err != nil {
		return nil, nil, err
	}
	if len(tokens) == 0 {
		return nil, nil, ErrVerilogSyntax
	}
	parser := verilogFormulaParser{tokens: tokens, limits: limits, seen: make(map[string]bool)}
	expr, err := parser.expression(0)
	if err != nil || parser.pos != len(tokens) {
		if err == nil {
			err = ErrVerilogSyntax
		}
		return nil, nil, err
	}
	return expr, append([]string(nil), parser.vars...), nil
}

func ParseVerilogReduction(text string, limits VerilogLimits) (*VerilogExpr, []string, error) {
	limits, err := normalizeVerilogLimits(limits)
	if err != nil {
		return nil, nil, err
	}
	tokens, err := formulaTokens(text, limits)
	if err != nil {
		return nil, nil, err
	}
	pos, negate := 0, false
	if pos < len(tokens) && tokens[pos] == "~" {
		negate = true
		pos++
	}
	if pos >= len(tokens) || (tokens[pos] != "&" && tokens[pos] != "|" && tokens[pos] != "^") {
		return nil, nil, ErrVerilogSyntax
	}
	op := tokens[pos]
	pos++
	if pos >= len(tokens) || tokens[pos] != "{" {
		return nil, nil, ErrVerilogSyntax
	}
	pos++
	args, vars := make([]*VerilogExpr, 0), make([]string, 0)
	seen := make(map[string]bool)
	for pos < len(tokens) && tokens[pos] != "}" {
		name := strings.TrimPrefix(tokens[pos], "\\")
		pos++
		if name == "," {
			continue
		}
		if name == "" || strings.ContainsAny(name, "()!~&|^?:{}") {
			return nil, nil, ErrVerilogSyntax
		}
		args = append(args, &VerilogExpr{Kind: VerilogExprName, Name: name})
		if !seen[name] {
			seen[name] = true
			vars = append(vars, name)
		}
		if len(args) > limits.MaxExprNodes {
			return nil, nil, ErrLimit
		}
	}
	if pos >= len(tokens) || tokens[pos] != "}" || pos+1 != len(tokens) || len(args) == 0 {
		return nil, nil, ErrVerilogSyntax
	}
	nodes := len(args) + 1
	if negate {
		nodes++
	}
	if nodes > limits.MaxExprNodes {
		return nil, nil, ErrLimit
	}
	expr := &VerilogExpr{Kind: formulaKind(op), Args: args}
	if negate {
		expr = &VerilogExpr{Kind: VerilogExprNot, Args: []*VerilogExpr{expr}}
	}
	return expr, vars, nil
}

type VerilogSignalKind uint8

const (
	VerilogWire VerilogSignalKind = iota
	VerilogInput
	VerilogOutput
	VerilogInout
	VerilogReg
)

type VerilogNet struct {
	Name     string
	Kind     VerilogSignalKind
	RangeMSB int
	RangeLSB int
	Driver   string
}

type VerilogRange struct{ MSB, LSB int }

type VerilogAssignment struct {
	Output     string
	Expression *VerilogExpr
	Inputs     []string
}
type VerilogGate struct {
	Type, Name, Output string
	Inputs             []string
	Cell, TwinOutput   string
	PinPermutation     []int
	Mapped             bool
}
type VerilogLatch struct {
	Output, Input, Clock string
	Inverted             bool
	Initial              int8
}
type VerilogConnection struct {
	Formal string
	Actual []string
}
type VerilogTerminal struct {
	Formal, Actual string
	Kind           VerilogSignalKind
	Bit            int
}
type VerilogInstance struct {
	Module, Name string
	Connections  []VerilogConnection
	Terminals    []VerilogTerminal
}

type VerilogModule struct {
	Name        string
	Ports       []string
	Nets        map[string]*VerilogNet
	Ranges      map[string]VerilogRange
	NetOrder    []string
	Assignments []VerilogAssignment
	Gates       []VerilogGate
	Latches     []VerilogLatch
	Instances   []VerilogInstance
	Undefined   bool
	Finalized   bool
	InstancesOf int
}

type VerilogDesign struct {
	Name            string
	Modules         []*VerilogModule
	Mapped          bool
	Finalized       bool
	HierarchyReport string
	Warnings        []string
	byName          map[string]*VerilogModule
}

type VerilogMappedCell struct {
	Name, Output, TwinOutput string
	Inputs                   []string
}

type VerilogGateLibrary struct {
	Cells  []VerilogMappedCell
	byName map[string]int
	cells  []VerilogMappedCell
}

func NewVerilogGateLibrary(cells []VerilogMappedCell) (*VerilogGateLibrary, error) {
	library := &VerilogGateLibrary{Cells: make([]VerilogMappedCell, len(cells)), cells: make([]VerilogMappedCell, len(cells)), byName: make(map[string]int, len(cells))}
	for index, source := range cells {
		if source.Name == "" || source.Output == "" || source.Output == source.TwinOutput {
			return nil, ErrDimension
		}
		if _, exists := library.byName[source.Name]; exists {
			return nil, ErrVerilogDuplicate
		}
		seen := make(map[string]bool, len(source.Inputs)+2)
		for _, name := range source.Inputs {
			if name == "" || name == source.Output || name == source.TwinOutput || seen[name] {
				return nil, ErrDimension
			}
			seen[name] = true
		}
		copyCell := source
		copyCell.Inputs = append([]string(nil), source.Inputs...)
		library.Cells[index] = copyCell
		library.cells[index] = copyCell
		library.cells[index].Inputs = append([]string(nil), copyCell.Inputs...)
		library.byName[source.Name] = index
	}
	return library, nil
}

func (library *VerilogGateLibrary) cell(name string) (VerilogMappedCell, bool) {
	if library == nil {
		return VerilogMappedCell{}, false
	}
	index, found := library.byName[name]
	if !found || index < 0 || index >= len(library.cells) {
		return VerilogMappedCell{}, false
	}
	cell := library.cells[index]
	cell.Inputs = append([]string(nil), cell.Inputs...)
	return cell, true
}

type VerilogParseOptions struct {
	Limits              VerilogLimits
	Mapped              bool
	GateLibrary         *VerilogGateLibrary
	Check               bool
	EmitHierarchyReport bool
}

func (design *VerilogDesign) Module(name string) (*VerilogModule, bool) {
	if design == nil {
		return nil, false
	}
	module, ok := design.byName[name]
	return module, ok
}

func (design *VerilogDesign) Clone() *VerilogDesign {
	if design == nil {
		return nil
	}
	clone := &VerilogDesign{Name: design.Name, Modules: make([]*VerilogModule, len(design.Modules)), Mapped: design.Mapped, Finalized: design.Finalized, HierarchyReport: design.HierarchyReport, Warnings: append([]string(nil), design.Warnings...), byName: make(map[string]*VerilogModule)}
	for i, source := range design.Modules {
		module := &VerilogModule{Name: source.Name, Ports: append([]string(nil), source.Ports...), NetOrder: append([]string(nil), source.NetOrder...), Nets: make(map[string]*VerilogNet), Ranges: make(map[string]VerilogRange), Gates: append([]VerilogGate(nil), source.Gates...), Latches: append([]VerilogLatch(nil), source.Latches...), Instances: append([]VerilogInstance(nil), source.Instances...), Undefined: source.Undefined, Finalized: source.Finalized, InstancesOf: source.InstancesOf}
		for name, net := range source.Nets {
			copyNet := *net
			module.Nets[name] = &copyNet
		}
		for name, signalRange := range source.Ranges {
			module.Ranges[name] = signalRange
		}
		module.Assignments = make([]VerilogAssignment, len(source.Assignments))
		for j := range source.Assignments {
			module.Assignments[j] = VerilogAssignment{Output: source.Assignments[j].Output, Expression: source.Assignments[j].Expression.Clone(), Inputs: append([]string(nil), source.Assignments[j].Inputs...)}
		}
		for j := range module.Gates {
			module.Gates[j].Inputs = append([]string(nil), source.Gates[j].Inputs...)
			module.Gates[j].PinPermutation = append([]int(nil), source.Gates[j].PinPermutation...)
		}
		for j := range module.Instances {
			module.Instances[j].Connections = cloneVerilogConnections(source.Instances[j].Connections)
			module.Instances[j].Terminals = append([]VerilogTerminal(nil), source.Instances[j].Terminals...)
		}
		clone.Modules[i] = module
		clone.byName[module.Name] = module
	}
	return clone
}

func cloneVerilogConnections(source []VerilogConnection) []VerilogConnection {
	result := make([]VerilogConnection, len(source))
	for i := range source {
		result[i] = VerilogConnection{Formal: source[i].Formal, Actual: append([]string(nil), source[i].Actual...)}
	}
	return result
}

type verilogToken struct {
	text string
	line int
}

type verilogLexer struct {
	data             []byte
	pos, line, count int
	limits           VerilogLimits
	peeked           *verilogToken
}

func (lexer *verilogLexer) skip() error {
	for lexer.pos < len(lexer.data) {
		if verilogStopped(lexer.limits) {
			return ErrCancelled
		}
		if strings.ContainsRune(" \t\r\n", rune(lexer.data[lexer.pos])) {
			if lexer.data[lexer.pos] == '\n' {
				lexer.line++
			}
			lexer.pos++
			continue
		}
		if lexer.pos+1 < len(lexer.data) && lexer.data[lexer.pos] == '/' && lexer.data[lexer.pos+1] == '/' {
			lexer.pos += 2
			for lexer.pos < len(lexer.data) && lexer.data[lexer.pos] != '\n' {
				lexer.pos++
			}
			continue
		}
		if lexer.pos+1 < len(lexer.data) && lexer.data[lexer.pos] == '/' && lexer.data[lexer.pos+1] == '*' {
			lexer.pos += 2
			closed := false
			for lexer.pos+1 < len(lexer.data) {
				if lexer.data[lexer.pos] == '\n' {
					lexer.line++
				}
				if lexer.data[lexer.pos] == '*' && lexer.data[lexer.pos+1] == '/' {
					lexer.pos += 2
					closed = true
					break
				}
				lexer.pos++
			}
			if !closed {
				return ErrVerilogSyntax
			}
			continue
		}
		break
	}
	return nil
}

func (lexer *verilogLexer) nextRaw() (verilogToken, error) {
	if verilogStopped(lexer.limits) {
		return verilogToken{}, ErrCancelled
	}
	if err := lexer.skip(); err != nil {
		return verilogToken{}, err
	}
	if lexer.pos == len(lexer.data) {
		return verilogToken{}, ErrVerilogEOF
	}
	line, start := lexer.line, lexer.pos
	single := "()[]{}.,;:=?~!&|^@#<>"
	if strings.IndexByte(single, lexer.data[lexer.pos]) >= 0 {
		lexer.pos++
	} else if lexer.data[lexer.pos] == '\\' {
		lexer.pos++
		for lexer.pos < len(lexer.data) && !strings.ContainsRune(" \t\r\n", rune(lexer.data[lexer.pos])) {
			lexer.pos++
		}
	} else {
		for lexer.pos < len(lexer.data) && !strings.ContainsRune(" \t\r\n()[]{}.,;:=?~!&|^@#<>", rune(lexer.data[lexer.pos])) {
			lexer.pos++
		}
	}
	if lexer.pos-start > lexer.limits.MaxWordBytes {
		return verilogToken{}, ErrLimit
	}
	lexer.count++
	if lexer.count > lexer.limits.MaxTokens {
		return verilogToken{}, ErrLimit
	}
	return verilogToken{text: strings.TrimPrefix(string(lexer.data[start:lexer.pos]), "\\"), line: line}, nil
}

func (lexer *verilogLexer) next() (verilogToken, error) {
	if lexer.peeked != nil {
		token := *lexer.peeked
		lexer.peeked = nil
		return token, nil
	}
	return lexer.nextRaw()
}
func (lexer *verilogLexer) peek() (verilogToken, error) {
	if lexer.peeked == nil {
		token, err := lexer.nextRaw()
		if err != nil {
			return verilogToken{}, err
		}
		lexer.peeked = &token
	}
	return *lexer.peeked, nil
}

type verilogDesignParser struct {
	lexer      verilogLexer
	limits     VerilogLimits
	options    VerilogParseOptions
	design     *VerilogDesign
	statements int
}

type VerilogSyntaxError struct{ Line int }

func (parseError VerilogSyntaxError) Error() string {
	return ErrVerilogSyntax.Error() + " at line " + strconv.Itoa(parseError.Line)
}

func (parseError VerilogSyntaxError) Unwrap() error { return ErrVerilogSyntax }

func syntaxAt(token verilogToken) error { return VerilogSyntaxError{Line: token.line} }
func (parser *verilogDesignParser) expect(text string) (verilogToken, error) {
	token, err := parser.lexer.next()
	if err != nil {
		return token, err
	}
	if token.text != text {
		return token, syntaxAt(token)
	}
	return token, nil
}

func (parser *verilogDesignParser) nextSignalName() (verilogToken, error) {
	token, err := parser.lexer.next()
	if err != nil {
		return token, err
	}
	return parser.finishSignalName(token)
}

func (parser *verilogDesignParser) finishSignalName(token verilogToken) (verilogToken, error) {
	next, peekErr := parser.lexer.peek()
	if peekErr == nil && next.text == "[" {
		parser.lexer.next()
		index, err := parser.lexer.next()
		if err != nil {
			return token, err
		}
		end, err := parser.lexer.next()
		if err != nil {
			return token, err
		}
		if end.text == ":" {
			last, err := parser.lexer.next()
			if err != nil {
				return token, err
			}
			if _, err = parser.expect("]"); err != nil {
				return token, err
			}
			token.text += "[" + index.text + ":" + last.text + "]"
			return token, nil
		}
		if end.text != "]" {
			return token, syntaxAt(end)
		}
		token.text += "[" + index.text + "]"
	}
	return token, nil
}

func ParseStructuralVerilog(name string, data []byte, limits VerilogLimits) (*VerilogDesign, error) {
	return ParseStructuralVerilogWithOptions(name, data, VerilogParseOptions{Limits: limits, Check: true})
}

func ParseStructuralVerilogWithOptions(name string, data []byte, options VerilogParseOptions) (*VerilogDesign, error) {
	limits, err := normalizeVerilogLimits(options.Limits)
	if err != nil {
		return nil, err
	}
	options.Limits = limits
	if options.Mapped && options.GateLibrary == nil {
		return nil, ErrDimension
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	if verilogStopped(limits) {
		return nil, ErrCancelled
	}
	parser := verilogDesignParser{lexer: verilogLexer{data: append([]byte(nil), data...), line: 1, limits: limits}, limits: limits, options: options}
	parser.design = &VerilogDesign{Name: name, Mapped: options.Mapped, byName: make(map[string]*VerilogModule)}
	for {
		token, err := parser.lexer.peek()
		if errors.Is(err, ErrVerilogEOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if token.text != "module" {
			return nil, syntaxAt(token)
		}
		module, err := parser.parseModule()
		if err != nil {
			return nil, err
		}
		if _, exists := parser.design.byName[module.Name]; exists {
			return nil, ErrVerilogDuplicate
		}
		if len(parser.design.Modules) == limits.MaxModules {
			return nil, ErrLimit
		}
		parser.design.Modules = append(parser.design.Modules, module)
		parser.design.byName[module.Name] = module
	}
	if len(parser.design.Modules) == 0 {
		return nil, ErrVerilogSyntax
	}
	if err := parser.finalizeDesign(); err != nil {
		return nil, err
	}
	return parser.design, nil
}

func ParseStructuralVerilogFile(path string, limits VerilogLimits) (*VerilogDesign, error) {
	return ParseStructuralVerilogFileWithOptions(path, VerilogParseOptions{Limits: limits, Check: true})
}

func ParseStructuralVerilogFileWithOptions(path string, options VerilogParseOptions) (*VerilogDesign, error) {
	limits, err := normalizeVerilogLimits(options.Limits)
	if err != nil {
		return nil, err
	}
	options.Limits = limits
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > int64(limits.MaxBytes) {
		return nil, ErrLimit
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseStructuralVerilogWithOptions(path, data, options)
}

func (parser *verilogDesignParser) parseModule() (*VerilogModule, error) {
	if _, err := parser.expect("module"); err != nil {
		return nil, err
	}
	name, err := parser.lexer.next()
	if err != nil || name.text == "" {
		return nil, ErrVerilogSyntax
	}
	module := &VerilogModule{Name: name.text, Nets: make(map[string]*VerilogNet), Ranges: make(map[string]VerilogRange)}
	if _, err = parser.expect("("); err != nil {
		return nil, err
	}
	for {
		token, err := parser.lexer.next()
		if err != nil {
			return nil, err
		}
		if token.text == ")" {
			break
		}
		if token.text == "," {
			continue
		}
		module.Ports = append(module.Ports, token.text)
	}
	if _, err = parser.expect(";"); err != nil {
		return nil, err
	}
	for {
		token, err := parser.lexer.next()
		if err != nil {
			return nil, err
		}
		if token.text == "endmodule" {
			break
		}
		parser.statements++
		if parser.statements > parser.limits.MaxStatements {
			return nil, ErrLimit
		}
		switch token.text {
		case "input", "output", "inout", "wire", "reg":
			err = parser.parseDeclaration(module, token.text)
		case "assign":
			err = parser.parseAssign(module)
		case "and", "or", "xor", "buf", "nand", "nor", "xnor", "not":
			err = parser.parseGate(module, token.text)
		case "dff":
			err = parser.parseDFF(module)
		case "always":
			err = parser.parseAlways(module)
		case "initial":
			err = parser.parseInitial(module)
		default:
			if cell, found := parser.options.GateLibrary.cell(token.text); found {
				err = parser.parseMappedGate(module, cell)
			} else {
				err = parser.parseInstance(module, token.text)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return module, nil
}

func signalKind(word string) VerilogSignalKind {
	switch word {
	case "input":
		return VerilogInput
	case "output":
		return VerilogOutput
	case "inout":
		return VerilogInout
	case "reg":
		return VerilogReg
	}
	return VerilogWire
}

func (parser *verilogDesignParser) parseRange() (int, int, bool, error) {
	token, err := parser.lexer.peek()
	if err != nil {
		return 0, 0, false, err
	}
	if token.text != "[" {
		return -1, -1, false, nil
	}
	parser.lexer.next()
	msbToken, err := parser.lexer.next()
	if err != nil {
		return 0, 0, false, err
	}
	msb, err := strconv.Atoi(msbToken.text)
	if err != nil || msb < 0 {
		return 0, 0, false, syntaxAt(msbToken)
	}
	lsb := msb
	next, err := parser.lexer.next()
	if err != nil {
		return 0, 0, false, err
	}
	if next.text == ":" {
		lsbToken, err := parser.lexer.next()
		if err != nil {
			return 0, 0, false, err
		}
		lsb, err = strconv.Atoi(lsbToken.text)
		if err != nil || lsb < 0 {
			return 0, 0, false, syntaxAt(lsbToken)
		}
		if _, err = parser.expect("]"); err != nil {
			return 0, 0, false, err
		}
	} else if next.text != "]" {
		return 0, 0, false, syntaxAt(next)
	}
	width := msb - lsb
	if width < 0 {
		width = -width
	}
	width++
	if width > parser.limits.MaxVectorBits {
		return 0, 0, false, ErrLimit
	}
	return msb, lsb, true, nil
}

func (parser *verilogDesignParser) addNet(module *VerilogModule, name string, kind VerilogSignalKind, msb, lsb int) error {
	if name == "" {
		return ErrVerilogSyntax
	}
	if old, exists := module.Nets[name]; exists {
		compatibleOutputReg := (old.Kind == VerilogOutput && kind == VerilogReg) || (old.Kind == VerilogReg && kind == VerilogOutput)
		if old.Kind != kind && old.Kind != VerilogWire && kind != VerilogWire && !compatibleOutputReg {
			return ErrVerilogDuplicate
		}
		if old.Kind == VerilogWire || (old.Kind == VerilogReg && kind == VerilogOutput) {
			old.Kind = kind
		}
		return nil
	}
	if len(module.Nets) == parser.limits.MaxNets {
		return ErrLimit
	}
	module.Nets[name] = &VerilogNet{Name: name, Kind: kind, RangeMSB: msb, RangeLSB: lsb}
	module.NetOrder = append(module.NetOrder, name)
	return nil
}

func expandRange(name string, msb, lsb int) []string {
	if msb < 0 {
		return []string{name}
	}
	count := msb - lsb
	if count < 0 {
		count = -count
	}
	count++
	result := make([]string, 0, count)
	step := 1
	if msb < lsb {
		step = -1
	}
	for bit, i := lsb, 0; i < count; bit, i = bit+step, i+1 {
		result = append(result, name+"["+strconv.Itoa(bit)+"]")
	}
	return result
}

func (parser *verilogDesignParser) parseDeclaration(module *VerilogModule, word string) error {
	kind := signalKind(word)
	if word == "output" {
		token, err := parser.lexer.peek()
		if err == nil && token.text == "reg" {
			parser.lexer.next()
		}
	}
	msb, lsb, ranged, err := parser.parseRange()
	if err != nil {
		return err
	}
	if !ranged {
		msb, lsb = -1, -1
	}
	for {
		name, err := parser.lexer.next()
		if err != nil {
			return err
		}
		if ranged {
			signalRange := VerilogRange{MSB: msb, LSB: lsb}
			if old, exists := module.Ranges[name.text]; exists && old != signalRange {
				return ErrVerilogDuplicate
			}
			module.Ranges[name.text] = signalRange
		}
		for _, expanded := range expandRange(name.text, msb, lsb) {
			if err = parser.addNet(module, expanded, kind, msb, lsb); err != nil {
				return err
			}
		}
		end, err := parser.lexer.next()
		if err != nil {
			return err
		}
		if end.text == ";" {
			return nil
		}
		if end.text != "," {
			return syntaxAt(end)
		}
	}
}

func (parser *verilogDesignParser) collectUntil(stop string) ([]string, error) {
	result, depth := make([]string, 0), 0
	for {
		token, err := parser.lexer.next()
		if err != nil {
			return nil, err
		}
		if token.text == stop && depth == 0 {
			return result, nil
		}
		if token.text == "(" || token.text == "{" {
			depth++
		}
		if token.text == ")" || token.text == "}" {
			depth--
		}
		result = append(result, token.text)
	}
}

func (parser *verilogDesignParser) collectUntilAssignmentEnd() ([]string, string, error) {
	result, depth := make([]string, 0), 0
	for {
		token, err := parser.lexer.next()
		if err != nil {
			return nil, "", err
		}
		if depth == 0 && (token.text == "," || token.text == ";") {
			return result, token.text, nil
		}
		if token.text == "(" || token.text == "{" {
			depth++
		}
		if token.text == ")" || token.text == "}" {
			depth--
		}
		result = append(result, token.text)
	}
}

func normalizeIndexedFormulaTokens(tokens []string) ([]string, error) {
	result := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		if i+3 < len(tokens) && tokens[i+1] == "[" && tokens[i+3] == "]" {
			result = append(result, tokens[i]+"["+tokens[i+2]+"]")
			i += 3
			continue
		}
		if tokens[i] == "[" || tokens[i] == "]" {
			return nil, ErrVerilogSyntax
		}
		result = append(result, tokens[i])
	}
	return result, nil
}

func joinFormula(tokens []string) string { return strings.Join(tokens, " ") }

func (parser *verilogDesignParser) ensureNet(module *VerilogModule, name string) error {
	if _, ok := module.Nets[name]; !ok {
		return ErrVerilogUndefined
	}
	return nil
}

func (parser *verilogDesignParser) markDriver(module *VerilogModule, output, driver string) error {
	net, ok := module.Nets[output]
	if !ok {
		return ErrVerilogUndefined
	}
	if net.Driver != "" {
		return ErrVerilogDriven
	}
	net.Driver = driver
	return nil
}

func (parser *verilogDesignParser) parseAssign(module *VerilogModule) error {
	first, err := parser.lexer.peek()
	if err != nil {
		return err
	}
	if first.text == "#" {
		parser.lexer.next()
		delay, err := parser.lexer.next()
		if err != nil {
			return err
		}
		if delay.text != "1" {
			return ErrVerilogSyntax
		}
	}
	for {
		output, err := parser.nextSignalName()
		if err != nil {
			return err
		}
		reduction := false
		if output.text == "{" {
			reduction = true
			output, err = parser.nextSignalName()
			if err != nil {
				return err
			}
			if _, err = parser.expect("}"); err != nil {
				return err
			}
		}
		if _, err = parser.expect("="); err != nil {
			return err
		}
		tokens, end, err := parser.collectUntilAssignmentEnd()
		if err != nil {
			return err
		}
		if err = parser.applyAssignment(module, output.text, tokens, reduction); err != nil {
			return err
		}
		if end == ";" {
			return nil
		}
	}
}

func (parser *verilogDesignParser) applyAssignment(module *VerilogModule, output string, tokens []string, reduction bool) error {
	var err error
	tokens, err = normalizeIndexedFormulaTokens(tokens)
	if err != nil {
		return err
	}
	if signalRange, vector := module.Ranges[output]; vector {
		if len(tokens) != 1 {
			return ErrVerilogSyntax
		}
		width := signalRange.MSB - signalRange.LSB
		if width < 0 {
			width = -width
		}
		width++
		bits, constantErr := ParseVerilogBinaryConstant(tokens[0], width)
		if constantErr != nil || len(bits) != width {
			return ErrVerilogSyntax
		}
		outputs := expandRange(output, signalRange.MSB, signalRange.LSB)
		for i, bitOutput := range outputs {
			if err = parser.markDriver(module, bitOutput, "assign"); err != nil {
				return err
			}
			module.Assignments = append(module.Assignments, VerilogAssignment{Output: bitOutput, Expression: &VerilogExpr{Kind: VerilogExprConst, Value: bits[len(bits)-1-i]}})
		}
		return nil
	}
	if err = parser.ensureNet(module, output); err != nil {
		return err
	}
	var expr *VerilogExpr
	var inputs []string
	if reduction {
		expr, inputs, err = ParseVerilogReduction(joinFormula(tokens), parser.limits)
	} else {
		expr, inputs, err = ParseVerilogFormula(joinFormula(tokens), parser.limits)
	}
	if err != nil {
		return err
	}
	for _, input := range inputs {
		if err = parser.ensureNet(module, input); err != nil {
			return err
		}
	}
	if err = parser.markDriver(module, output, "assign"); err != nil {
		return err
	}
	module.Assignments = append(module.Assignments, VerilogAssignment{Output: output, Expression: expr, Inputs: inputs})
	return nil
}

func (parser *verilogDesignParser) parseGate(module *VerilogModule, kind string) error {
	first, err := parser.lexer.next()
	if err != nil {
		return err
	}
	name := ""
	if first.text != "(" {
		name = first.text
		if _, err = parser.expect("("); err != nil {
			return err
		}
	}
	nets := make([]string, 0)
	for {
		token, err := parser.nextSignalName()
		if err != nil {
			return err
		}
		if token.text == ")" {
			break
		}
		if token.text == "," {
			continue
		}
		nets = append(nets, token.text)
	}
	if _, err = parser.expect(";"); err != nil {
		return err
	}
	if len(nets) < 2 || ((kind == "buf" || kind == "not") && len(nets) != 2) {
		return ErrVerilogSyntax
	}
	for _, net := range nets {
		if err = parser.ensureNet(module, net); err != nil {
			return err
		}
	}
	if err = parser.markDriver(module, nets[0], kind); err != nil {
		return err
	}
	module.Gates = append(module.Gates, VerilogGate{Type: kind, Name: name, Output: nets[0], Inputs: append([]string(nil), nets[1:]...)})
	return nil
}

func mappedGateFormalIndex(cell VerilogMappedCell, formal string) int {
	for index, input := range cell.Inputs {
		if input == formal {
			return index
		}
	}
	if formal == cell.Output {
		return len(cell.Inputs)
	}
	if cell.TwinOutput != "" && formal == cell.TwinOutput {
		return len(cell.Inputs) + 1
	}
	return -1
}

func (parser *verilogDesignParser) parseMappedGate(module *VerilogModule, cell VerilogMappedCell) error {
	if !parser.options.Mapped {
		return ErrUnsupported
	}
	instance, err := parser.lexer.next()
	if err != nil || instance.text == "" {
		return ErrVerilogSyntax
	}
	if _, err = parser.expect("("); err != nil {
		return err
	}
	inputs := make([]string, len(cell.Inputs))
	permutation := make([]int, 0, len(cell.Inputs))
	seen := make([]bool, len(cell.Inputs)+1)
	if cell.TwinOutput != "" {
		seen = append(seen, false)
	}
	output, twinOutput := "", ""
	for {
		token, nextErr := parser.lexer.peek()
		if nextErr != nil {
			return nextErr
		}
		if token.text == ")" {
			parser.lexer.next()
			break
		}
		if _, err = parser.expect("."); err != nil {
			return err
		}
		formal, formalErr := parser.lexer.next()
		if formalErr != nil {
			return formalErr
		}
		pin := mappedGateFormalIndex(cell, formal.text)
		if pin < 0 || pin >= len(seen) || seen[pin] {
			return ErrVerilogDuplicate
		}
		seen[pin] = true
		if _, err = parser.expect("("); err != nil {
			return err
		}
		actualToken, actualErr := parser.nextSignalName()
		if actualErr != nil {
			return actualErr
		}
		actuals, actualErr := verilogActualNames(module, actualToken.text, parser.limits.MaxVectorBits)
		if actualErr != nil || len(actuals) != 1 {
			return ErrVerilogSyntax
		}
		actual := actuals[0]
		if _, err = parser.expect(")"); err != nil {
			return err
		}
		if pin < len(cell.Inputs) {
			inputs[pin] = actual
			permutation = append(permutation, pin)
		} else if pin == len(cell.Inputs) {
			output = actual
		} else {
			twinOutput = actual
		}
		next, nextErr := parser.lexer.next()
		if nextErr != nil {
			return nextErr
		}
		if next.text == ")" {
			break
		}
		if next.text != "," {
			return syntaxAt(next)
		}
	}
	if _, err = parser.expect(";"); err != nil {
		return err
	}
	for pin := range cell.Inputs {
		if !seen[pin] {
			return ErrVerilogSyntax
		}
	}
	if !seen[len(cell.Inputs)] || cell.TwinOutput != "" && !seen[len(cell.Inputs)+1] {
		return ErrVerilogSyntax
	}
	if err = parser.markDriver(module, output, "mapped "+cell.Name); err != nil {
		return err
	}
	if twinOutput != "" {
		if err = parser.markDriver(module, twinOutput, "mapped "+cell.TwinOutput); err != nil {
			return err
		}
	}
	module.Gates = append(module.Gates, VerilogGate{Type: cell.Name, Cell: cell.Name, Name: instance.text, Output: output, TwinOutput: twinOutput, Inputs: inputs, PinPermutation: permutation, Mapped: true})
	return nil
}

func (parser *verilogDesignParser) parseDFF(module *VerilogModule) error {
	first, err := parser.lexer.next()
	if err != nil {
		return err
	}
	if first.text != "(" {
		if _, err = parser.expect("("); err != nil {
			return err
		}
	}
	output, err := parser.nextSignalName()
	if err != nil {
		return err
	}
	if _, err = parser.expect(","); err != nil {
		return err
	}
	input, err := parser.nextSignalName()
	if err != nil {
		return err
	}
	if _, err = parser.expect(")"); err != nil {
		return err
	}
	if _, err = parser.expect(";"); err != nil {
		return err
	}
	if err = parser.ensureNet(module, output.text); err != nil {
		return err
	}
	if err = parser.ensureNet(module, input.text); err != nil {
		return err
	}
	if err = parser.markDriver(module, output.text, "dff"); err != nil {
		return err
	}
	// Ver_ParseFlopStandard discards the optional instance name and creates a
	// latch initialized to zero. The source does not interpret this name as a
	// clock.
	module.Latches = append(module.Latches, VerilogLatch{Output: output.text, Input: input.text, Initial: 0})
	return nil
}

func (parser *verilogDesignParser) parseAlways(module *VerilogModule) error {
	clock := ""
	token, err := parser.lexer.next()
	if err != nil {
		return err
	}
	if token.text == "@" {
		if _, err = parser.expect("("); err != nil {
			return err
		}
		// The source skips everything through the next closing parenthesis. This
		// accepts wildcard and compound sensitivity lists; retain only a useful
		// first event name as diagnostic metadata.
		events := make([]string, 0, 4)
		for {
			event, eventErr := parser.lexer.next()
			if eventErr != nil {
				return eventErr
			}
			if event.text == ")" {
				break
			}
			events = append(events, event.text)
		}
		if len(events) > 1 && (events[0] == "posedge" || events[0] == "negedge") {
			clock = events[1]
		} else if len(events) > 0 {
			clock = events[0]
		}
		token, err = parser.lexer.next()
		if err != nil {
			return err
		}
	}
	// Ver_ParseAlways skips the complete event-control text. Retain the parsed
	// spelling as owned metadata, but do not reject a source-accepted statement
	// merely because its event name is not a declared data net.
	multiple := token.text == "begin"
	for {
		if multiple {
			token, err = parser.lexer.next()
			if err != nil {
				return err
			}
			if token.text == "end" {
				return nil
			}
		}
		outputToken := token
		if next, peekErr := parser.lexer.peek(); peekErr == nil && next.text == "[" {
			outputToken, err = parser.finishSignalName(outputToken)
			if err != nil {
				return err
			}
		}
		output := outputToken.text
		op, err := parser.lexer.next()
		if err != nil {
			return err
		}
		if op.text == "<" {
			if _, err = parser.expect("="); err != nil {
				return err
			}
		} else if op.text != "=" {
			return syntaxAt(op)
		}
		inverted := false
		input, err := parser.nextSignalName()
		if err != nil {
			return err
		}
		if input.text == "~" {
			inverted = true
			input, err = parser.nextSignalName()
			if err != nil {
				return err
			}
		}
		if _, err = parser.expect(";"); err != nil {
			return err
		}
		if err = parser.ensureNet(module, output); err != nil {
			return err
		}
		if err = parser.ensureNet(module, input.text); err != nil {
			return err
		}
		if err = parser.markDriver(module, output, "always"); err != nil {
			return err
		}
		module.Latches = append(module.Latches, VerilogLatch{Output: output, Input: input.text, Clock: clock, Inverted: inverted, Initial: -1})
		if !multiple {
			return nil
		}
	}
}

func (parser *verilogDesignParser) parseInitial(module *VerilogModule) error {
	token, err := parser.lexer.next()
	if err != nil {
		return err
	}
	multiple := token.text == "begin"
	for {
		if multiple {
			token, err = parser.lexer.next()
			if err != nil {
				return err
			}
			if token.text == "end" {
				return nil
			}
		}
		output := token.text
		op, err := parser.lexer.next()
		if err != nil {
			return err
		}
		if op.text == "<" {
			if _, err = parser.expect("="); err != nil {
				return err
			}
		} else if op.text != "=" {
			return syntaxAt(op)
		}
		value, err := parser.lexer.next()
		if err != nil {
			return err
		}
		if _, err = parser.expect(";"); err != nil {
			return err
		}
		initial := int8(-1)
		if value.text == "0" || value.text == "1'b0" {
			initial = 0
		} else if value.text == "1" || value.text == "1'b1" {
			initial = 1
		} else {
			return ErrVerilogSyntax
		}
		found := false
		for i := len(module.Latches) - 1; i >= 0; i-- {
			if module.Latches[i].Output == output {
				module.Latches[i].Initial = initial
				found = true
				break
			}
		}
		if !found {
			return ErrVerilogUndefined
		}
		if !multiple {
			return nil
		}
	}
}

func (parser *verilogDesignParser) parseActual(module *VerilogModule) ([]string, error) {
	next, err := parser.lexer.peek()
	if err != nil {
		return nil, err
	}
	if next.text == ")" {
		return nil, nil
	}
	token, err := parser.nextSignalName()
	if err != nil {
		return nil, err
	}
	if token.text == "{" {
		values := make([]string, 0)
		for {
			token, err = parser.nextSignalName()
			if err != nil {
				return nil, err
			}
			if token.text == "}" {
				break
			}
			if token.text == "," {
				continue
			}
			expanded, expandErr := verilogActualNames(module, token.text, parser.limits.MaxVectorBits)
			if expandErr != nil {
				return nil, expandErr
			}
			values = append(values, expanded...)
		}
		return values, nil
	}
	return verilogActualNames(module, token.text, parser.limits.MaxVectorBits)
}

func verilogActualNames(module *VerilogModule, name string, maxBits int) ([]string, error) {
	if strings.HasPrefix(name, "Open_") || strings.HasPrefix(name, "dct_unconnected") {
		// The source creates an unnamed net for these synthesis-tool spellings.
		return nil, nil
	}
	if strings.Contains(name, "'") {
		bits, err := ParseVerilogBinaryConstant(name, maxBits)
		if err != nil {
			return nil, err
		}
		actuals := make([]string, len(bits))
		for i, bit := range bits {
			if bit {
				actuals[i] = "1'b1"
			} else {
				actuals[i] = "1'b0"
			}
		}
		return actuals, nil
	}
	if open := strings.LastIndexByte(name, '['); open > 0 && strings.HasSuffix(name, "]") && strings.Contains(name[open+1:len(name)-1], ":") {
		bounds := strings.Split(name[open+1:len(name)-1], ":")
		if len(bounds) != 2 {
			return nil, ErrVerilogSyntax
		}
		msb, err0 := strconv.Atoi(bounds[0])
		lsb, err1 := strconv.Atoi(bounds[1])
		if err0 != nil || err1 != nil || msb < 0 || lsb < 0 {
			return nil, ErrVerilogSyntax
		}
		base := name[:open]
		count := msb - lsb
		if count < 0 {
			count = -count
		}
		count++
		step := -1
		if msb < lsb {
			step = 1
		}
		result := make([]string, 0, count)
		for bit, i := msb, 0; i < count; bit, i = bit+step, i+1 {
			actual := base + "[" + strconv.Itoa(bit) + "]"
			if _, ok := module.Nets[actual]; !ok {
				return nil, ErrVerilogUndefined
			}
			result = append(result, actual)
		}
		return result, nil
	}
	if signalRange, ok := module.Ranges[name]; ok {
		bits := expandRange(name, signalRange.MSB, signalRange.LSB)
		result := make([]string, 0, len(bits))
		for i := len(bits) - 1; i >= 0; i-- {
			result = append(result, bits[i])
		}
		return result, nil
	}
	if _, ok := module.Nets[name]; !ok && !strings.Contains(name, "'") {
		return nil, ErrVerilogUndefined
	}
	return []string{name}, nil
}

func (parser *verilogDesignParser) parseInstance(module *VerilogModule, moduleName string) error {
	name, err := parser.lexer.next()
	if err != nil {
		return err
	}
	if _, err = parser.expect("("); err != nil {
		return err
	}
	instance := VerilogInstance{Module: moduleName, Name: name.text}
	for {
		token, err := parser.lexer.peek()
		if err != nil {
			return err
		}
		if token.text == ")" {
			parser.lexer.next()
			break
		}
		formal := ""
		if token.text == "." {
			parser.lexer.next()
			formalToken, err := parser.nextSignalName()
			if err != nil {
				return err
			}
			formal = formalToken.text
			if _, err = parser.expect("("); err != nil {
				return err
			}
			actual, err := parser.parseActual(module)
			if err != nil {
				return err
			}
			if _, err = parser.expect(")"); err != nil {
				return err
			}
			instance.Connections = append(instance.Connections, VerilogConnection{Formal: formal, Actual: actual})
		} else {
			actual, err := parser.parseActual(module)
			if err != nil {
				return err
			}
			instance.Connections = append(instance.Connections, VerilogConnection{Actual: actual})
		}
		next, err := parser.lexer.next()
		if err != nil {
			return err
		}
		if next.text == ")" {
			break
		}
		if next.text != "," {
			return syntaxAt(next)
		}
	}
	if _, err = parser.expect(";"); err != nil {
		return err
	}
	module.Instances = append(module.Instances, instance)
	return nil
}

func (parser *verilogDesignParser) validateInstances() error {
	for _, module := range parser.design.Modules {
		for instanceIndex := range module.Instances {
			if verilogStopped(parser.limits) {
				return ErrCancelled
			}
			instance := &module.Instances[instanceIndex]
			instance.Terminals = nil
			target, exists := parser.design.byName[instance.Module]
			if !exists || target.Undefined {
				continue
			}
			target.InstancesOf++
			named := len(instance.Connections) > 0 && instance.Connections[0].Formal != ""
			if !named {
				if err := parser.connectPositionalInstance(module, target, instance); err != nil {
					return err
				}
				continue
			}
			formalBits := 0
			for _, formal := range target.Ports {
				_, width, found := verilogFormalInfo(target, formal)
				if !found {
					return ErrVerilogUndefined
				}
				formalBits += width
			}
			if len(instance.Connections) > formalBits {
				return ErrVerilogSyntax
			}
			seenFormal := make(map[string]bool)
			for _, connection := range instance.Connections {
				if (connection.Formal != "") != named {
					return ErrVerilogSyntax
				}
				formal := connection.Formal
				if seenFormal[formal] || verilogFormalOverlapsSeen(target, formal, seenFormal) {
					return ErrVerilogDuplicate
				}
				seenFormal[formal] = true
				kind, width, found := verilogFormalInfo(target, formal)
				if !found {
					return ErrVerilogUndefined
				}
				if len(connection.Actual) == 0 {
					// Ver_ParseBox creates an anonymous net for an empty actual,
					// regardless of the formal direction. Empty Actual strings are
					// the pointer-free representation of those disconnected nets.
					for bit := 0; bit < width; bit++ {
						instance.Terminals = append(instance.Terminals, VerilogTerminal{Formal: formal, Kind: kind, Bit: bit})
					}
					continue
				}
				if len(connection.Actual) != width {
					return ErrVerilogSyntax
				}
				if kind == VerilogOutput || kind == VerilogInout {
					for bit, actual := range connection.Actual {
						if strings.Contains(actual, "'") {
							return ErrVerilogSyntax
						}
						if err := parser.markDriver(module, actual, "instance "+instance.Name); err != nil {
							return err
						}
						instance.Terminals = append(instance.Terminals, VerilogTerminal{Formal: formal, Actual: actual, Kind: kind, Bit: bit})
					}
				} else {
					for bit, actual := range connection.Actual {
						instance.Terminals = append(instance.Terminals, VerilogTerminal{Formal: formal, Actual: actual, Kind: kind, Bit: bit})
					}
				}
			}
			if named {
				for _, formal := range target.Ports {
					kind, width, found := verilogFormalInfo(target, formal)
					if !found {
						return ErrVerilogUndefined
					}
					covered := make([]bool, width)
					if seenFormal[formal] {
						for bit := range covered {
							covered[bit] = true
						}
					} else if signalRange, vector := target.Ranges[formal]; vector {
						for bit, bitName := range expandRange(formal, signalRange.MSB, signalRange.LSB) {
							covered[bit] = seenFormal[bitName]
						}
					}
					for bit, present := range covered {
						if present {
							continue
						}
						if kind == VerilogInput || kind == VerilogInout {
							return ErrVerilogUndefined
						}
						instance.Terminals = append(instance.Terminals, VerilogTerminal{Formal: formal, Kind: kind, Bit: bit})
					}
				}
			}
		}
	}
	return nil
}

func (parser *verilogDesignParser) connectPositionalInstance(parent, target *VerilogModule, instance *VerilogInstance) error {
	actuals := make([]string, 0)
	for _, connection := range instance.Connections {
		if len(connection.Actual) == 0 {
			actuals = append(actuals, "")
		} else {
			actuals = append(actuals, connection.Actual...)
		}
	}
	formals := make([]VerilogTerminal, 0)
	for _, kind := range []VerilogSignalKind{VerilogInput, VerilogOutput} {
		for _, name := range target.NetOrder {
			net := target.Nets[name]
			if net == nil {
				return ErrInvariant
			}
			if net.Kind != kind && net.Kind != VerilogInout {
				continue
			}
			formals = append(formals, VerilogTerminal{Formal: name, Kind: kind})
		}
	}
	if len(actuals) != len(formals) {
		return ErrVerilogSyntax
	}
	for index := range formals {
		if verilogStopped(parser.limits) {
			return ErrCancelled
		}
		actual := actuals[index]
		formals[index].Actual = actual
		if actual == "" {
			continue
		}
		if formals[index].Kind == VerilogOutput {
			if verilogIsConstant(actual) {
				return ErrVerilogSyntax
			}
			if err := parser.markDriver(parent, actual, "instance "+instance.Name); err != nil {
				return err
			}
		} else if !verilogIsConstant(actual) {
			if _, found := parent.Nets[actual]; !found {
				return ErrVerilogUndefined
			}
		}
	}
	instance.Terminals = formals
	return nil
}

func verilogFormalOverlapsSeen(module *VerilogModule, formal string, seen map[string]bool) bool {
	if signalRange, vector := module.Ranges[formal]; vector {
		for _, bitName := range expandRange(formal, signalRange.MSB, signalRange.LSB) {
			if seen[bitName] {
				return true
			}
		}
		return false
	}
	if open := strings.LastIndexByte(formal, '['); open > 0 && strings.HasSuffix(formal, "]") {
		return seen[formal[:open]]
	}
	return false
}

type verilogInstanceRef struct {
	parent   *VerilogModule
	instance *VerilogInstance
}

func verilogConnectionAt(instance *VerilogInstance, slot int, formals []string) (*VerilogConnection, bool) {
	if instance == nil || slot < 0 {
		return nil, false
	}
	named := len(instance.Connections) > 0 && instance.Connections[0].Formal != ""
	if !named {
		if slot >= len(instance.Connections) {
			return nil, false
		}
		return &instance.Connections[slot], true
	}
	if slot >= len(formals) {
		return nil, false
	}
	for index := range instance.Connections {
		if instance.Connections[index].Formal == formals[slot] {
			return &instance.Connections[index], true
		}
	}
	return nil, false
}

func verilogIsConstant(name string) bool { return strings.Contains(name, "'") }

func (parser *verilogDesignParser) inferUndefinedModules() error {
	groupOrder := make([]string, 0)
	groups := make(map[string][]verilogInstanceRef)
	definedCount := len(parser.design.Modules)
	for moduleIndex := 0; moduleIndex < definedCount; moduleIndex++ {
		parent := parser.design.Modules[moduleIndex]
		for instanceIndex := range parent.Instances {
			instance := &parent.Instances[instanceIndex]
			if _, defined := parser.design.byName[instance.Module]; defined {
				continue
			}
			if _, found := groups[instance.Module]; !found {
				groupOrder = append(groupOrder, instance.Module)
			}
			groups[instance.Module] = append(groups[instance.Module], verilogInstanceRef{parent: parent, instance: instance})
		}
	}
	for _, moduleName := range groupOrder {
		if verilogStopped(parser.limits) {
			return ErrCancelled
		}
		refs := groups[moduleName]
		if len(refs) == 0 {
			continue
		}
		first := refs[0].instance
		named := len(first.Connections) > 0 && first.Connections[0].Formal != ""
		formals := make([]string, len(first.Connections))
		widths := make([]int, len(first.Connections))
		seenFormals := make(map[string]bool, len(first.Connections))
		for slot, connection := range first.Connections {
			if verilogStopped(parser.limits) {
				return ErrCancelled
			}
			if (connection.Formal != "") != named {
				return ErrVerilogSyntax
			}
			if named {
				formals[slot] = connection.Formal
			} else {
				formals[slot] = "port" + strconv.Itoa(slot)
			}
			if seenFormals[formals[slot]] {
				return ErrVerilogDuplicate
			}
			seenFormals[formals[slot]] = true
			widths[slot] = len(connection.Actual)
			if widths[slot] == 0 {
				widths[slot] = 1
			}
		}
		for _, ref := range refs {
			if verilogStopped(parser.limits) {
				return ErrCancelled
			}
			if len(ref.instance.Connections) != len(formals) {
				return ErrVerilogSyntax
			}
			for slot := range formals {
				connection, found := verilogConnectionAt(ref.instance, slot, formals)
				if !found || (connection.Formal != "") != named {
					return ErrVerilogSyntax
				}
				width := len(connection.Actual)
				if width == 0 {
					width = 1
				}
				if width != widths[slot] {
					return ErrVerilogSyntax
				}
			}
		}
		kinds := make([]VerilogSignalKind, len(formals))
		for slot := len(formals) - 1; slot >= 0; slot-- {
			if verilogStopped(parser.limits) {
				return ErrCancelled
			}
			anyUndriven, anyDriven := false, false
			for _, ref := range refs {
				connection, _ := verilogConnectionAt(ref.instance, slot, formals)
				if len(connection.Actual) == 0 {
					// Ver_ParseBox creates an anonymous, initially undriven net
					// for an empty actual. It therefore participates in undefined-
					// box direction inference as an output candidate.
					anyUndriven = true
				}
				for _, actual := range connection.Actual {
					if verilogIsConstant(actual) {
						continue
					}
					net := ref.parent.Nets[actual]
					if net == nil {
						return ErrVerilogUndefined
					}
					// A primary input is driven by its implicit PI terminal.  Driver
					// records explicit structural drivers only, so its empty spelling
					// must not make an input look like an undefined-box output.
					if net.Driver == "" && net.Kind != VerilogInput && net.Kind != VerilogInout {
						anyUndriven = true
					} else {
						anyDriven = true
					}
				}
			}
			if anyUndriven && !anyDriven {
				kinds[slot] = VerilogOutput
				for _, ref := range refs {
					connection, _ := verilogConnectionAt(ref.instance, slot, formals)
					for _, actual := range connection.Actual {
						if verilogIsConstant(actual) {
							return ErrVerilogSyntax
						}
						if err := parser.markDriver(ref.parent, actual, "instance "+ref.instance.Name); err != nil {
							return err
						}
					}
				}
			} else {
				kinds[slot] = VerilogInput
			}
		}
		module := &VerilogModule{Name: moduleName, Ports: append([]string(nil), formals...), Nets: make(map[string]*VerilogNet), Ranges: make(map[string]VerilogRange), Undefined: true, Finalized: true, InstancesOf: len(refs)}
		for slot, formal := range formals {
			if widths[slot] == 1 {
				if len(module.Nets) >= parser.limits.MaxNets {
					return ErrLimit
				}
				module.Nets[formal] = &VerilogNet{Name: formal, Kind: kinds[slot], RangeMSB: -1, RangeLSB: -1}
				module.NetOrder = append(module.NetOrder, formal)
			} else {
				module.Ranges[formal] = VerilogRange{MSB: widths[slot] - 1, LSB: 0}
				for _, bitName := range expandRange(formal, widths[slot]-1, 0) {
					if len(module.Nets) >= parser.limits.MaxNets {
						return ErrLimit
					}
					module.Nets[bitName] = &VerilogNet{Name: bitName, Kind: kinds[slot], RangeMSB: widths[slot] - 1, RangeLSB: 0}
					module.NetOrder = append(module.NetOrder, bitName)
				}
			}
		}
		for _, ref := range refs {
			ref.instance.Terminals = nil
			for slot, formal := range formals {
				connection, _ := verilogConnectionAt(ref.instance, slot, formals)
				for bit, actual := range connection.Actual {
					ref.instance.Terminals = append(ref.instance.Terminals, VerilogTerminal{Formal: formal, Actual: actual, Kind: kinds[slot], Bit: bit})
				}
			}
		}
		if len(parser.design.Modules) >= parser.limits.MaxModules {
			return ErrLimit
		}
		parser.design.Modules = append(parser.design.Modules, module)
		parser.design.byName[module.Name] = module
		parser.design.Warnings = append(parser.design.Warnings, "undefined module "+moduleName+" inferred from "+strconv.Itoa(len(refs))+" instance(s)")
	}
	return nil
}

func (parser *verilogDesignParser) finalizeNonDrivenNets() error {
	for _, module := range parser.design.Modules {
		if module.Undefined {
			continue
		}
		// Abc_NtkFixNonDrivenNets deliberately leaves a pure terminal-only
		// network alone. Once the module contains a node or box, every remaining
		// non-driven net receives an explicit constant-zero driver.
		if len(module.Assignments) == 0 && len(module.Gates) == 0 && len(module.Latches) == 0 && len(module.Instances) == 0 {
			continue
		}
		fixed := make([]string, 0)
		for _, name := range module.NetOrder {
			if verilogStopped(parser.limits) {
				return ErrCancelled
			}
			net := module.Nets[name]
			if net == nil {
				return ErrInvariant
			}
			if net.Driver != "" || net.Kind == VerilogInput || net.Kind == VerilogInout {
				continue
			}
			net.Driver = "constant 0"
			module.Assignments = append(module.Assignments, VerilogAssignment{Output: name, Expression: &VerilogExpr{Kind: VerilogExprConst}})
			fixed = append(fixed, name)
		}
		if len(fixed) != 0 {
			parser.design.Warnings = append(parser.design.Warnings, "constant-0 drivers added to "+strconv.Itoa(len(fixed))+" non-driven net(s) in module "+module.Name+": "+strings.Join(fixed, ", "))
		}
	}
	return nil
}

func (parser *verilogDesignParser) checkFinalizedDesign() error {
	for _, module := range parser.design.Modules {
		if verilogStopped(parser.limits) {
			return ErrCancelled
		}
		for _, port := range module.Ports {
			if _, _, found := verilogFormalInfo(module, port); !found {
				return ErrVerilogUndefined
			}
		}
		if !module.Undefined {
			for _, name := range module.NetOrder {
				net := module.Nets[name]
				if net == nil {
					return ErrInvariant
				}
				if net.Kind == VerilogOutput && net.Driver == "" {
					return ErrVerilogUndefined
				}
			}
		}
		module.Finalized = true
	}
	return nil
}

func (parser *verilogDesignParser) hierarchyReport() (string, error) {
	var builder strings.Builder
	builder.WriteString("Hierarchical design ")
	builder.WriteString(parser.design.Name)
	builder.WriteString(" contains ")
	builder.WriteString(strconv.Itoa(len(parser.design.Modules)))
	builder.WriteString(" modules:\n")
	for _, module := range parser.design.Modules {
		if verilogStopped(parser.limits) {
			return "", ErrCancelled
		}
		kind := "logicbox"
		if module.Undefined {
			kind = "undefbox"
		} else {
			// Abc_NtkHasBlackbox reports whether a defined module contains a
			// black-box instance; ordinary hierarchy does not itself make the
			// containing module a black box.
			for _, instance := range module.Instances {
				target := parser.design.byName[instance.Module]
				if target != nil && target.Undefined {
					kind = "blackbox"
					break
				}
			}
		}
		builder.WriteString(module.Name)
		builder.WriteByte(' ')
		builder.WriteString(kind)
		builder.WriteString(" instantiated ")
		builder.WriteString(strconv.Itoa(module.InstancesOf))
		builder.WriteString(" times pi=")
		inputs, outputs := 0, 0
		for _, name := range module.NetOrder {
			if verilogStopped(parser.limits) {
				return "", ErrCancelled
			}
			net := module.Nets[name]
			if net.Kind == VerilogInput || net.Kind == VerilogInout {
				inputs++
			}
			if net.Kind == VerilogOutput || net.Kind == VerilogInout {
				outputs++
			}
		}
		builder.WriteString(strconv.Itoa(inputs))
		builder.WriteString(" po=")
		builder.WriteString(strconv.Itoa(outputs))
		nodes := len(module.Assignments)
		for _, gate := range module.Gates {
			nodes++
			if gate.Mapped && gate.TwinOutput != "" {
				nodes++
			}
		}
		for _, latch := range module.Latches {
			if latch.Inverted {
				nodes++
			}
		}
		boxes := len(module.Instances)
		builder.WriteString(" nd=")
		builder.WriteString(strconv.Itoa(nodes))
		builder.WriteString(" lat=")
		builder.WriteString(strconv.Itoa(len(module.Latches)))
		builder.WriteString(" box=")
		builder.WriteString(strconv.Itoa(boxes))
		builder.WriteByte('\n')
	}
	oneOutput := 0
	for _, module := range parser.design.Modules {
		if verilogStopped(parser.limits) {
			return "", ErrCancelled
		}
		outputs := 0
		for _, name := range module.NetOrder {
			net := module.Nets[name]
			if net != nil && (net.Kind == VerilogOutput || net.Kind == VerilogInout) {
				outputs++
			}
		}
		if outputs == 1 {
			oneOutput++
		}
	}
	builder.WriteString("The number of modules with one output = ")
	builder.WriteString(strconv.Itoa(oneOutput))
	builder.WriteString(".\n")
	dangling := 0
	for _, module := range parser.design.Modules {
		if verilogStopped(parser.limits) {
			return "", ErrCancelled
		}
		for _, instance := range module.Instances {
			for _, terminal := range instance.Terminals {
				// The source report compares total actual and formal net counts,
				// so any omitted actual makes this instance incomplete.
				if terminal.Actual == "" {
					dangling++
					break
				}
			}
		}
	}
	if dangling == 0 {
		builder.WriteString("The outputs of all box instances are connected.\n")
	} else {
		builder.WriteString("The outputs of ")
		builder.WriteString(strconv.Itoa(dangling))
		builder.WriteString(" box instance(s) are not connected.\n")
	}
	return builder.String(), nil
}

func (parser *verilogDesignParser) finalizeDesign() error {
	for _, module := range parser.design.Modules {
		module.InstancesOf = 0
	}
	if err := parser.validateInstances(); err != nil {
		return err
	}
	if err := parser.inferUndefinedModules(); err != nil {
		return err
	}
	if err := parser.finalizeNonDrivenNets(); err != nil {
		return err
	}
	if parser.options.Check {
		if err := parser.checkFinalizedDesign(); err != nil {
			return err
		}
	} else {
		for _, module := range parser.design.Modules {
			module.Finalized = true
		}
	}
	parser.design.Finalized = true
	if parser.options.EmitHierarchyReport {
		report, err := parser.hierarchyReport()
		if err != nil {
			return err
		}
		parser.design.HierarchyReport = report
	}
	return nil
}

func verilogFormalInfo(module *VerilogModule, formal string) (VerilogSignalKind, int, bool) {
	if signalRange, vector := module.Ranges[formal]; vector {
		bits := expandRange(formal, signalRange.MSB, signalRange.LSB)
		if len(bits) == 0 {
			return VerilogWire, 0, false
		}
		net, found := module.Nets[bits[0]]
		if !found {
			return VerilogWire, 0, false
		}
		return net.Kind, len(bits), true
	}
	net, found := module.Nets[formal]
	if !found {
		return VerilogWire, 0, false
	}
	return net.Kind, 1, true
}

func ParseVerilogBinaryConstant(text string, maxBits int) ([]bool, error) {
	quote := strings.IndexByte(text, '\'')
	if quote <= 0 || quote+2 > len(text) || text[quote+1] != 'b' {
		return nil, ErrVerilogSyntax
	}
	width, err := strconv.Atoi(text[:quote])
	if err != nil || width < 1 {
		return nil, ErrVerilogSyntax
	}
	if maxBits > 0 && width > maxBits {
		return nil, ErrLimit
	}
	digits := text[quote+2:]
	if len(digits) != width {
		return nil, ErrVerilogSyntax
	}
	result := make([]bool, width)
	for i, digit := range digits {
		if digit == '1' {
			result[i] = true
		} else if digit != '0' && digit != 'x' && digit != 'X' {
			return nil, ErrVerilogSyntax
		}
	}
	return result, nil
}
