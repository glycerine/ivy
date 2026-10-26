package abc

// Concrete AMAP library, formula, subject-graph, and direct-match foundations
// from the remaining src/map/amap units. This file does not claim the native
// DSD rule synthesis, cut merge, iterative reference-driven mapper, or its
// exact-area dereference engine; callers of MapAMapGraph get only the explicit
// one-node primitive path documented there.

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const AMapMaximumGateInputs = 15

var AMapRemainingSourceInventory = [...]string{
	"amapGraph.c", "amapLib.c", "amapLiberty.c", "amapMatch.c",
	"amapMerge.c", "amapOutput.c", "amapParse.c", "amapPerm.c",
	"amapRead.c", "amapRule.c", "amapUniq.c", "module.make",
}

var ErrAMapDSDUnavailable = errors.New("AMAP DSD decomposition is unavailable")

type AMapPinPhase uint8

const (
	AMapPhaseUnknown AMapPinPhase = iota
	AMapPhaseInverting
	AMapPhaseNonInverting
)

type AMapPin struct {
	Name                                         string
	Phase                                        AMapPinPhase
	InputLoad, MaximumLoad                       float64
	BlockRise, FanoutRise, BlockFall, FanoutFall float64
	BlockMaximum                                 float64
}

type AMapGate struct {
	ID, Twin              int
	Name, Output, Formula string
	Area                  float64
	Pins                  []AMapPin
	Truth                 []uint64
	Mux                   bool
}

type AMapRuleNodeType uint8

const (
	AMapRuleInput AMapRuleNodeType = iota + 1
	AMapRuleAnd
	AMapRuleXor
	AMapRuleMux
)

type AMapRuleNode struct {
	ID, Fan0, Fan1, Fan2 int
	Type                 AMapRuleNodeType
	Support              int
}

type amapRuleKey struct {
	Type    AMapRuleNodeType
	A, B, C int
}

type AMapCellLibrary struct {
	Name                 string
	Gates                []AMapGate
	Sorted, Selected     []int
	Constant0, Constant1 int
	Buffer, Inverter     int
	HasXOR, HasMUX       bool
	RuleNodes            []AMapRuleNode
	ruleIndex            map[amapRuleKey]int
}

type AMapLimits struct {
	MaxBytes, MaxGates, MaxPins, MaxFormulaNodes, MaxRuleNodes int
	Stop                                                       func() bool
}

func normalizeAMapLimits(l AMapLimits) (AMapLimits, error) {
	if l.MaxBytes < 0 || l.MaxGates < 0 || l.MaxPins < 0 || l.MaxFormulaNodes < 0 || l.MaxRuleNodes < 0 {
		return l, ErrDimension
	}
	if l.MaxBytes == 0 {
		l.MaxBytes = 64 << 20
	}
	if l.MaxGates == 0 {
		l.MaxGates = 1 << 20
	}
	if l.MaxPins == 0 {
		l.MaxPins = AMapMaximumGateInputs
	}
	if l.MaxFormulaNodes == 0 {
		l.MaxFormulaNodes = 1 << 20
	}
	if l.MaxRuleNodes == 0 {
		l.MaxRuleNodes = 1 << 20
	}
	if l.MaxPins > AMapMaximumGateInputs {
		l.MaxPins = AMapMaximumGateInputs
	}
	return l, nil
}

func amapStopped(l AMapLimits) bool { return l.Stop != nil && l.Stop() }

type amapFormulaKind uint8

const (
	amapFormulaVar amapFormulaKind = iota
	amapFormulaZero
	amapFormulaOne
	amapFormulaNot
	amapFormulaAnd
	amapFormulaXor
	amapFormulaOr
)

type amapFormula struct {
	kind        amapFormulaKind
	name        string
	left, right *amapFormula
}

type amapFormulaParser struct {
	text            string
	pos, nodes, max int
	stop            func() bool
}

func (p *amapFormulaParser) skip() {
	for p.pos < len(p.text) && unicode.IsSpace(rune(p.text[p.pos])) {
		p.pos++
	}
}
func (p *amapFormulaParser) node(n *amapFormula) (*amapFormula, error) {
	p.nodes++
	if p.nodes > p.max {
		return nil, ErrLimit
	}
	if p.stop != nil && p.stop() {
		return nil, ErrCancelled
	}
	return n, nil
}
func (p *amapFormulaParser) parse() (*amapFormula, error) {
	n, e := p.parseOr()
	p.skip()
	if e == nil && p.pos != len(p.text) {
		e = ErrVerilogSyntax
	}
	return n, e
}
func (p *amapFormulaParser) parseOr() (*amapFormula, error) {
	n, e := p.parseXor()
	for e == nil {
		p.skip()
		if p.pos >= len(p.text) || (p.text[p.pos] != '+' && p.text[p.pos] != '|') {
			break
		}
		p.pos++
		var r *amapFormula
		r, e = p.parseXor()
		if e == nil {
			n, e = p.node(&amapFormula{kind: amapFormulaOr, left: n, right: r})
		}
	}
	return n, e
}
func (p *amapFormulaParser) parseXor() (*amapFormula, error) {
	n, e := p.parseAnd()
	for e == nil {
		p.skip()
		if p.pos >= len(p.text) || p.text[p.pos] != '^' {
			break
		}
		p.pos++
		var r *amapFormula
		r, e = p.parseAnd()
		if e == nil {
			n, e = p.node(&amapFormula{kind: amapFormulaXor, left: n, right: r})
		}
	}
	return n, e
}
func (p *amapFormulaParser) parseAnd() (*amapFormula, error) {
	n, e := p.parseUnary()
	for e == nil {
		p.skip()
		if p.pos >= len(p.text) {
			break
		}
		explicit := p.text[p.pos] == '*' || p.text[p.pos] == '&'
		implicit := p.text[p.pos] != ')' && p.text[p.pos] != '+' && p.text[p.pos] != '|' && p.text[p.pos] != '^' && p.text[p.pos] != '\''
		if !explicit && !implicit {
			break
		}
		if explicit {
			p.pos++
		}
		var r *amapFormula
		r, e = p.parseUnary()
		if e == nil {
			n, e = p.node(&amapFormula{kind: amapFormulaAnd, left: n, right: r})
		}
	}
	return n, e
}
func (p *amapFormulaParser) parseUnary() (*amapFormula, error) {
	p.skip()
	if p.pos >= len(p.text) {
		return nil, ErrVerilogSyntax
	}
	if p.text[p.pos] == '!' {
		p.pos++
		n, e := p.parseUnary()
		if e != nil {
			return nil, e
		}
		return p.node(&amapFormula{kind: amapFormulaNot, left: n})
	}
	var n *amapFormula
	var e error
	if p.text[p.pos] == '(' {
		p.pos++
		n, e = p.parseOr()
		p.skip()
		if e != nil {
			return nil, e
		}
		if p.pos >= len(p.text) || p.text[p.pos] != ')' {
			return nil, ErrVerilogSyntax
		}
		p.pos++
	} else {
		start := p.pos
		for p.pos < len(p.text) && !unicode.IsSpace(rune(p.text[p.pos])) && !strings.ContainsRune("*&+|^!()'", rune(p.text[p.pos])) {
			p.pos++
		}
		if start == p.pos {
			return nil, ErrVerilogSyntax
		}
		name := p.text[start:p.pos]
		kind := amapFormulaVar
		if name == "0" {
			kind = amapFormulaZero
		}
		if name == "1" {
			kind = amapFormulaOne
		}
		n, e = p.node(&amapFormula{kind: kind, name: name})
	}
	if e != nil {
		return nil, e
	}
	for {
		p.skip()
		if p.pos >= len(p.text) || p.text[p.pos] != '\'' {
			break
		}
		p.pos++
		n, e = p.node(&amapFormula{kind: amapFormulaNot, left: n})
		if e != nil {
			return nil, e
		}
	}
	return n, nil
}

func (n *amapFormula) eval(values map[string]bool) (bool, error) {
	if n == nil {
		return false, ErrInvariant
	}
	switch n.kind {
	case amapFormulaZero:
		return false, nil
	case amapFormulaOne:
		return true, nil
	case amapFormulaVar:
		v, ok := values[n.name]
		if !ok {
			return false, ErrVerilogUndefined
		}
		return v, nil
	case amapFormulaNot:
		v, e := n.left.eval(values)
		return !v, e
	case amapFormulaAnd, amapFormulaXor, amapFormulaOr:
		a, e := n.left.eval(values)
		if e != nil {
			return false, e
		}
		b, e := n.right.eval(values)
		if e != nil {
			return false, e
		}
		if n.kind == amapFormulaAnd {
			return a && b, nil
		}
		if n.kind == amapFormulaXor {
			return a != b, nil
		}
		return a || b, nil
	default:
		return false, ErrInvariant
	}
}

func amapFormulaNames(n *amapFormula, seen map[string]bool, out *[]string) {
	if n == nil {
		return
	}
	if n.kind == amapFormulaVar && !seen[n.name] {
		seen[n.name] = true
		*out = append(*out, n.name)
	}
	amapFormulaNames(n.left, seen, out)
	amapFormulaNames(n.right, seen, out)
}

func amapTruth(n *amapFormula, pins []AMapPin) ([]uint64, error) {
	if len(pins) > AMapMaximumGateInputs {
		return nil, ErrLimit
	}
	words := 1
	if len(pins) > 6 {
		words = 1 << (len(pins) - 6)
	}
	truth := make([]uint64, words)
	values := make(map[string]bool, len(pins))
	for assignment := 0; assignment < (1 << len(pins)); assignment++ {
		for i := range pins {
			values[pins[i].Name] = assignment&(1<<i) != 0
		}
		v, e := n.eval(values)
		if e != nil {
			return nil, e
		}
		if v {
			truth[assignment>>6] |= uint64(1) << uint(assignment&63)
		}
	}
	return truth, nil
}

func amapTruthHasFullSupport(truth []uint64, variables int) bool {
	if variables == 0 {
		return true
	}
	assignments := 1 << variables
	for variable := 0; variable < variables; variable++ {
		dependent := false
		bit := 1 << variable
		for assignment := 0; assignment < assignments; assignment++ {
			if assignment&bit != 0 {
				continue
			}
			low := truth[assignment>>6] >> (uint(assignment) & 63) & 1
			highAssignment := assignment | bit
			high := truth[highAssignment>>6] >> (uint(highAssignment) & 63) & 1
			if low != high {
				dependent = true
				break
			}
		}
		if !dependent {
			return false
		}
	}
	return true
}

func cleanAMapGenlib(data []byte) (string, error) {
	var out strings.Builder
	scan := bufio.NewScanner(bytes.NewReader(data))
	scan.Buffer(make([]byte, 1024), len(data)+1)
	continuation := ""
	for scan.Scan() {
		line := scan.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "\\") {
			continuation += strings.TrimSpace(strings.TrimSuffix(line, "\\")) + " "
			continue
		}
		out.WriteString(continuation)
		continuation = ""
		out.WriteString(line)
		out.WriteByte('\n')
	}
	if scan.Err() != nil {
		return "", scan.Err()
	}
	if continuation != "" {
		return "", ErrVerilogSyntax
	}
	return out.String(), nil
}

func ParseAMapLibrary(name string, data []byte, limits AMapLimits) (*AMapCellLibrary, error) {
	limits, err := normalizeAMapLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	if amapStopped(limits) {
		return nil, ErrCancelled
	}
	clean, err := cleanAMapGenlib(data)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(clean, "\n")
	lib := &AMapCellLibrary{Name: name, Constant0: -1, Constant1: -1, Buffer: -1, Inverter: -1, ruleIndex: make(map[amapRuleKey]int)}
	var gate *AMapGate
	skipLatch := false
	finish := func() error {
		if gate == nil {
			return nil
		}
		if len(gate.Pins) == 1 && gate.Pins[0].Name == "*" {
			p := amapFormulaParser{text: gate.Formula, max: limits.MaxFormulaNodes, stop: limits.Stop}
			tree, e := p.parse()
			if e != nil {
				return e
			}
			names := []string{}
			amapFormulaNames(tree, map[string]bool{}, &names)
			template := gate.Pins[0]
			gate.Pins = make([]AMapPin, len(names))
			for i, n := range names {
				gate.Pins[i] = template
				gate.Pins[i].Name = n
			}
		}
		if len(gate.Pins) > limits.MaxPins {
			return ErrLimit
		}
		formula := gate.Formula
		if len(gate.Pins) == 0 {
			if formula == "CONST0" {
				formula = "0"
			} else if formula == "CONST1" {
				formula = "1"
			}
		}
		p := amapFormulaParser{text: formula, max: limits.MaxFormulaNodes, stop: limits.Stop}
		tree, e := p.parse()
		if e != nil {
			return e
		}
		names := []string{}
		amapFormulaNames(tree, map[string]bool{}, &names)
		declared := map[string]bool{}
		for _, pin := range gate.Pins {
			if pin.Name == "" || declared[pin.Name] {
				return ErrVerilogDuplicate
			}
			declared[pin.Name] = true
		}
		for _, n := range names {
			if !declared[n] {
				return ErrVerilogUndefined
			}
		}
		gate.Truth, e = amapTruth(tree, gate.Pins)
		if e != nil {
			return e
		}
		if !amapTruthHasFullSupport(gate.Truth, len(gate.Pins)) {
			gate.Truth = nil
		}
		gate.ID = len(lib.Gates)
		gate.Twin = -1
		lib.Gates = append(lib.Gates, *gate)
		gate = nil
		return nil
	}
	for lineNo, line := range lines {
		if amapStopped(limits) {
			return nil, ErrCancelled
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case ".end":
			if err = finish(); err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo+1, err)
			}
			goto done
		case "GATE":
			skipLatch = false
			if err = finish(); err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo+1, err)
			}
			if len(lib.Gates) >= limits.MaxGates {
				return nil, ErrLimit
			}
			if len(fields) < 4 {
				return nil, fmt.Errorf("line %d: %w", lineNo+1, ErrVerilogSyntax)
			}
			area, e := strconv.ParseFloat(fields[2], 64)
			if e != nil || area < 0 || math.IsNaN(area) || math.IsInf(area, 0) {
				return nil, fmt.Errorf("line %d: %w", lineNo+1, ErrDimension)
			}
			equation := amapAfterFields(line, 3)
			semi := strings.LastIndexByte(equation, ';')
			eq := strings.IndexByte(equation, '=')
			if eq <= 0 || semi != len(equation)-1 {
				return nil, fmt.Errorf("line %d: %w", lineNo+1, ErrVerilogSyntax)
			}
			output := strings.TrimSpace(equation[:eq])
			formula := strings.TrimSpace(equation[eq+1 : semi])
			if output == "" || formula == "" {
				return nil, fmt.Errorf("line %d: %w", lineNo+1, ErrVerilogSyntax)
			}
			gate = &AMapGate{Name: fields[1], Area: area, Output: output, Formula: formula, Twin: -1}
		case "PIN":
			if skipLatch {
				continue
			}
			if gate == nil || len(fields) != 9 {
				return nil, fmt.Errorf("line %d: %w", lineNo+1, ErrVerilogSyntax)
			}
			phase := AMapPhaseUnknown
			if fields[2] == "INV" {
				phase = AMapPhaseInverting
			} else if fields[2] == "NONINV" {
				phase = AMapPhaseNonInverting
			} else if fields[2] != "UNKNOWN" {
				return nil, fmt.Errorf("line %d: %w", lineNo+1, ErrVerilogSyntax)
			}
			nums := make([]float64, 6)
			for i := range nums {
				nums[i], err = strconv.ParseFloat(fields[i+3], 64)
				if err != nil || math.IsNaN(nums[i]) || math.IsInf(nums[i], 0) {
					return nil, fmt.Errorf("line %d: %w", lineNo+1, ErrDimension)
				}
			}
			gate.Pins = append(gate.Pins, AMapPin{Name: fields[1], Phase: phase, InputLoad: nums[0], MaximumLoad: nums[1], BlockRise: nums[2], FanoutRise: nums[3], BlockFall: nums[4], FanoutFall: nums[5], BlockMaximum: math.Max(nums[2], nums[4])})
		case "LATCH":
			skipLatch = true
		default:
			if skipLatch {
				continue
			}
			return nil, fmt.Errorf("line %d: %w", lineNo+1, ErrVerilogSyntax)
		}
	}
	if err = finish(); err != nil {
		return nil, err
	}
done:
	if len(lib.Gates) == 0 {
		return nil, ErrVerilogSyntax
	}
	for i := 1; i < len(lib.Gates); i++ {
		if lib.Gates[i-1].Name == lib.Gates[i].Name {
			lib.Gates[i-1].Twin = i
			lib.Gates[i].Twin = i - 1
		}
	}
	lib.prepare()
	return lib, nil
}

func amapAfterFields(line string, count int) string {
	position := 0
	for field := 0; field < count; field++ {
		for position < len(line) && unicode.IsSpace(rune(line[position])) {
			position++
		}
		for position < len(line) && !unicode.IsSpace(rune(line[position])) {
			position++
		}
	}
	return strings.TrimSpace(line[position:])
}

func ReadAMapLibrary(path string, limits AMapLimits) (*AMapCellLibrary, error) {
	data, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	return ParseAMapLibrary(path, data, limits)
}

func (lib *AMapCellLibrary) prepare() {
	lib.Sorted = make([]int, len(lib.Gates))
	for i := range lib.Gates {
		lib.Sorted[i] = i
	}
	sort.SliceStable(lib.Sorted, func(i, j int) bool { return lib.Gates[lib.Sorted[i]].Area < lib.Gates[lib.Sorted[j]].Area })
	seen := make(map[string]bool)
	for _, i := range lib.Sorted {
		g := &lib.Gates[i]
		if g.Twin >= 0 || len(g.Truth) == 0 {
			continue
		}
		key := strconv.Itoa(len(g.Pins)) + ":" + fmt.Sprint(g.Truth)
		if seen[key] {
			continue
		}
		seen[key] = true
		lib.Selected = append(lib.Selected, i)
	}
	for _, i := range lib.Sorted {
		g := &lib.Gates[i]
		if len(g.Pins) > 5 || len(g.Truth) == 0 {
			continue
		}
		t := g.Truth[0]
		mask := uint64(1) << uint(1<<len(g.Pins))
		mask--
		switch {
		case t == 0 && lib.Constant0 < 0:
			lib.Constant0 = i
		case t == mask && lib.Constant1 < 0:
			lib.Constant1 = i
		case len(g.Pins) == 1 && t&3 == 2 && lib.Buffer < 0:
			lib.Buffer = i
		case len(g.Pins) == 1 && t&3 == 1 && lib.Inverter < 0:
			lib.Inverter = i
		}
		if len(g.Pins) == 2 && (t&15 == 6 || t&15 == 9) {
			lib.HasXOR = true
		}
		if len(g.Pins) == 3 && amapTruthIsMux(uint8(t)) {
			g.Mux = true
			lib.HasMUX = true
		}
	}
	lib.RuleNodes = []AMapRuleNode{{ID: 0, Type: AMapRuleInput, Support: 1}}
	lib.ruleIndex = make(map[amapRuleKey]int)
}

func amapTruthIsMux(truth uint8) bool {
	// A three-input mux has a selector whose two cofactors are (possibly
	// complemented) literals of the two remaining variables.
	for selector := 0; selector < 3; selector++ {
		remaining := make([]int, 0, 2)
		for variable := 0; variable < 3; variable++ {
			if variable != selector {
				remaining = append(remaining, variable)
			}
		}
		for swap := 0; swap < 2; swap++ {
			for inv0 := 0; inv0 < 2; inv0++ {
				for inv1 := 0; inv1 < 2; inv1++ {
					valid := true
					for assignment := 0; assignment < 8; assignment++ {
						s := assignment >> selector & 1
						variable := remaining[s^swap]
						want := assignment>>variable&1 != 0
						if (s == 0 && inv0 != 0) || (s == 1 && inv1 != 0) {
							want = !want
						}
						if (truth>>assignment&1 != 0) != want {
							valid = false
							break
						}
					}
					if valid {
						return true
					}
				}
			}
		}
	}
	return false
}

func (lib *AMapCellLibrary) Clone() *AMapCellLibrary {
	if lib == nil {
		return nil
	}
	c := *lib
	c.Gates = make([]AMapGate, len(lib.Gates))
	for i := range lib.Gates {
		c.Gates[i] = lib.Gates[i]
		c.Gates[i].Pins = append([]AMapPin(nil), lib.Gates[i].Pins...)
		c.Gates[i].Truth = append([]uint64(nil), lib.Gates[i].Truth...)
	}
	c.Sorted = append([]int(nil), lib.Sorted...)
	c.Selected = append([]int(nil), lib.Selected...)
	c.RuleNodes = append([]AMapRuleNode(nil), lib.RuleNodes...)
	c.ruleIndex = make(map[amapRuleKey]int, len(lib.ruleIndex))
	for k, v := range lib.ruleIndex {
		c.ruleIndex[k] = v
	}
	return &c
}

func (lib *AMapCellLibrary) MaxPins() int {
	n := 0
	if lib != nil {
		for _, g := range lib.Gates {
			if len(g.Pins) > n {
				n = len(g.Pins)
			}
		}
	}
	return n
}
func (lib *AMapCellLibrary) FindTruth(pinCount int, truth []uint64) (int, bool) {
	if lib == nil {
		return 0, false
	}
	for _, i := range lib.Sorted {
		g := lib.Gates[i]
		if len(g.Pins) == pinCount && slicesEqualUint64(g.Truth, truth) {
			return i, true
		}
	}
	return 0, false
}
func slicesEqualUint64(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (lib *AMapCellLibrary) WriteGenlib() ([]byte, error) {
	if lib == nil {
		return nil, ErrDimension
	}
	var b strings.Builder
	phases := []string{"UNKNOWN", "INV", "NONINV"}
	for _, g := range lib.Gates {
		fmt.Fprintf(&b, "GATE %s %.2f %s=%s;\n", g.Name, g.Area, g.Output, g.Formula)
		for _, p := range g.Pins {
			if int(p.Phase) >= len(phases) {
				return nil, ErrInvariant
			}
			fmt.Fprintf(&b, "    PIN %s %s %.6g %.6g %.6g %.6g %.6g %.6g\n", p.Name, phases[p.Phase], p.InputLoad, p.MaximumLoad, p.BlockRise, p.FanoutRise, p.BlockFall, p.FanoutFall)
		}
	}
	b.WriteString(".end\n")
	return []byte(b.String()), nil
}

func amapLitVar(lit int) int { return lit >> 1 }
func amapLitNot(lit int) int { return lit ^ 1 }

// CreateRuleNode provides the structural uniqueness primitive from amapUniq.c.
// Full rule creation still requires the absent DSD decomposition and gate-set
// permutation pipeline from amapRule.c and amapPerm.c.
func (lib *AMapCellLibrary) CreateRuleNode(kind AMapRuleNodeType, a, b, c int, limits AMapLimits) (int, error) {
	if lib == nil {
		return 0, ErrDimension
	}
	limits, e := normalizeAMapLimits(limits)
	if e != nil {
		return 0, e
	}
	if amapStopped(limits) {
		return 0, ErrCancelled
	}
	if kind != AMapRuleAnd && kind != AMapRuleXor && kind != AMapRuleMux {
		return 0, ErrDimension
	}
	if amapLitVar(a) >= len(lib.RuleNodes) || amapLitVar(b) >= len(lib.RuleNodes) || (kind == AMapRuleMux && amapLitVar(c) >= len(lib.RuleNodes)) {
		return 0, ErrDimension
	}
	if kind != AMapRuleMux && a < b {
		a, b = b, a
	}
	key := amapRuleKey{kind, a, b, c}
	if id, ok := lib.ruleIndex[key]; ok {
		return id, nil
	}
	if len(lib.RuleNodes) >= limits.MaxRuleNodes {
		return 0, ErrLimit
	}
	support := lib.RuleNodes[amapLitVar(a)].Support + lib.RuleNodes[amapLitVar(b)].Support
	if kind == AMapRuleMux {
		support += lib.RuleNodes[amapLitVar(c)].Support
	}
	id := len(lib.RuleNodes)
	lib.RuleNodes = append(lib.RuleNodes, AMapRuleNode{ID: id, Type: kind, Fan0: a, Fan1: b, Fan2: c, Support: support})
	lib.ruleIndex[key] = id
	return id, nil
}

type AMapGraphType uint8

const (
	AMapGraphConst1 AMapGraphType = iota + 1
	AMapGraphPI
	AMapGraphPO
	AMapGraphAND
	AMapGraphXOR
	AMapGraphMUX
)

type AMapGraphNode struct {
	ID                int
	Type              AMapGraphType
	Fanins            []int
	Level, References int
	Choice            int
	Phase             bool
}
type amapGraphKey struct {
	kind    AMapGraphType
	a, b, c int
}
type AMapGraph struct {
	Nodes           []AMapGraphNode
	Inputs, Outputs []int
	unique          map[amapGraphKey]int
	MaxLevel        int
	limits          AMapLimits
}

func NewAMapGraph(limits AMapLimits) (*AMapGraph, error) {
	limits, e := normalizeAMapLimits(limits)
	if e != nil {
		return nil, e
	}
	g := &AMapGraph{unique: make(map[amapGraphKey]int), limits: limits}
	_, e = g.add(AMapGraphConst1, nil)
	return g, e
}
func (g *AMapGraph) add(kind AMapGraphType, fans []int) (int, error) {
	if g == nil {
		return 0, ErrDimension
	}
	if amapStopped(g.limits) {
		return 0, ErrCancelled
	}
	if len(g.Nodes) >= g.limits.MaxRuleNodes {
		return 0, ErrLimit
	}
	level := 0
	phase := kind == AMapGraphConst1
	for _, lit := range fans {
		id := amapLitVar(lit)
		if id < 0 || id >= len(g.Nodes) {
			return 0, ErrDimension
		}
		if g.Nodes[id].Level > level {
			level = g.Nodes[id].Level
		}
	}
	for _, lit := range fans {
		g.Nodes[amapLitVar(lit)].References++
	}
	if kind == AMapGraphAND {
		level++
		phase = (g.Nodes[amapLitVar(fans[0])].Phase != (fans[0]&1 != 0)) && (g.Nodes[amapLitVar(fans[1])].Phase != (fans[1]&1 != 0))
	}
	if kind == AMapGraphXOR {
		level += 2
		phase = (g.Nodes[amapLitVar(fans[0])].Phase != (fans[0]&1 != 0)) != (g.Nodes[amapLitVar(fans[1])].Phase != (fans[1]&1 != 0))
	}
	if kind == AMapGraphMUX {
		level += 2
		value0 := g.Nodes[amapLitVar(fans[0])].Phase != (fans[0]&1 != 0)
		value1 := g.Nodes[amapLitVar(fans[1])].Phase != (fans[1]&1 != 0)
		control := g.Nodes[amapLitVar(fans[2])].Phase != (fans[2]&1 != 0)
		phase = value1 && control || value0 && !control
	}
	id := len(g.Nodes)
	g.Nodes = append(g.Nodes, AMapGraphNode{ID: id, Type: kind, Fanins: append([]int(nil), fans...), Level: level, Choice: -1, Phase: phase})
	if level > g.MaxLevel {
		g.MaxLevel = level
	}
	return id, nil
}
func (g *AMapGraph) Const1() int { return 0 }
func (g *AMapGraph) CreatePI() (int, error) {
	id, e := g.add(AMapGraphPI, nil)
	if e == nil {
		g.Inputs = append(g.Inputs, id)
	}
	return id << 1, e
}
func (g *AMapGraph) CreatePO(lit int) (int, error) {
	id, e := g.add(AMapGraphPO, []int{lit})
	if e == nil {
		g.Outputs = append(g.Outputs, id)
	}
	return id, e
}
func (g *AMapGraph) CreateNode(kind AMapGraphType, a, b, c int) (int, error) {
	if kind != AMapGraphAND && kind != AMapGraphXOR && kind != AMapGraphMUX {
		return 0, ErrDimension
	}
	if kind != AMapGraphMUX && a < b {
		a, b = b, a
	}
	key := amapGraphKey{kind, a, b, c}
	if id, ok := g.unique[key]; ok {
		return id << 1, nil
	}
	fans := []int{a, b}
	if kind == AMapGraphMUX {
		fans = append(fans, c)
	}
	id, e := g.add(kind, fans)
	if e == nil {
		g.unique[key] = id
	}
	return id << 1, e
}
func (g *AMapGraph) SetChoice(representative, equivalent int) error {
	if g == nil || representative < 0 || equivalent < 0 || representative >= len(g.Nodes) || equivalent >= len(g.Nodes) || representative == equivalent {
		return ErrDimension
	}
	if g.Nodes[representative].Choice >= 0 {
		return ErrVerilogDuplicate
	}
	g.Nodes[representative].Choice = equivalent
	if g.Nodes[equivalent].Level > g.Nodes[representative].Level {
		g.Nodes[representative].Level = g.Nodes[equivalent].Level
	}
	if g.Nodes[representative].Level > g.MaxLevel {
		g.MaxLevel = g.Nodes[representative].Level
	}
	return nil
}

// CreateAMapGraphFromAIG ports the source graph construction for the concrete
// append-only AigManager representation. Native XOR/MUX recognizers are not
// present in that manager (XOR is lowered to ANDs), so its live nodes are
// transferred source-faithfully as complemented-edge AND nodes.
func CreateAMapGraphFromAIG(aig *AigManager, limits AMapLimits) (*AMapGraph, error) {
	if aig == nil {
		return nil, ErrDimension
	}
	if err := aig.Check(); err != nil {
		return nil, err
	}
	graph, err := NewAMapGraph(limits)
	if err != nil {
		return nil, err
	}
	copyLiteral := make([]int, aig.ObjectCount())
	reachable := make([]bool, aig.ObjectCount())
	var mark func(AigID) error
	mark = func(id AigID) error {
		if int(id) < 0 || int(id) >= len(reachable) {
			return ErrDimension
		}
		if reachable[id] {
			return nil
		}
		reachable[id] = true
		object, objectErr := aig.Object(id)
		if objectErr != nil {
			return objectErr
		}
		switch object.Type {
		case AigAND, AigXOR:
			if objectErr = mark(object.Fanin0.ID()); objectErr != nil {
				return objectErr
			}
			return mark(object.Fanin1.ID())
		case AigBuffer, AigCO:
			return mark(object.Fanin0.ID())
		}
		return nil
	}
	for id := 1; id < aig.ObjectCount(); id++ {
		object, objectErr := aig.Object(AigID(id))
		if objectErr != nil {
			return nil, objectErr
		}
		if object.Type == AigCI {
			reachable[id] = true
		}
		if object.Type == AigCO {
			if objectErr = mark(AigID(id)); objectErr != nil {
				return nil, objectErr
			}
		}
	}
	copyLiteral[0] = graph.Const1()
	for id := 1; id < aig.ObjectCount(); id++ {
		if amapStopped(graph.limits) {
			return nil, ErrCancelled
		}
		object, objectErr := aig.Object(AigID(id))
		if objectErr != nil {
			return nil, objectErr
		}
		if !reachable[id] {
			continue
		}
		switch object.Type {
		case AigCI:
			copyLiteral[id], err = graph.CreatePI()
		case AigAND:
			fan0 := copyLiteral[object.Fanin0.ID()]
			fan1 := copyLiteral[object.Fanin1.ID()]
			if object.Fanin0.Complemented() {
				fan0 = amapLitNot(fan0)
			}
			if object.Fanin1.Complemented() {
				fan1 = amapLitNot(fan1)
			}
			copyLiteral[id], err = graph.CreateNode(AMapGraphAND, fan0, fan1, 0)
		case AigCO:
			fan := copyLiteral[object.Fanin0.ID()]
			if object.Fanin0.Complemented() {
				fan = amapLitNot(fan)
			}
			_, err = graph.CreatePO(fan)
		case AigBuffer:
			copyLiteral[id] = copyLiteral[object.Fanin0.ID()]
			if object.Fanin0.Complemented() {
				copyLiteral[id] = amapLitNot(copyLiteral[id])
			}
		case AigNone, AigVoid:
			continue
		default:
			return nil, ErrUnsupported
		}
		if err != nil {
			return nil, err
		}
	}
	return graph, nil
}

type AMapMatch struct {
	Node, Gate                 int
	Area, Delay, AverageFanout float64
	Inverted                   bool
	PinPermutation             []int
}

func CompareAMapMatches(a, b AMapMatch, areaFirst bool, epsilon float64) int {
	cmp := func(x, y float64, minimize bool) int {
		if minimize {
			if x < y-epsilon {
				return -1
			}
			if x > y+epsilon {
				return 1
			}
		} else {
			if x > y-epsilon {
				return -1
			}
			if x < y+epsilon {
				return 1
			}
		}
		return 0
	}
	if areaFirst {
		if c := cmp(a.Area, b.Area, true); c != 0 {
			return c
		}
		if c := cmp(a.AverageFanout, b.AverageFanout, false); c != 0 {
			return c
		}
		if c := cmp(a.Delay, b.Delay, true); c != 0 {
			return c
		}
	} else {
		if c := cmp(a.Delay, b.Delay, true); c != 0 {
			return c
		}
		if c := cmp(a.Area, b.Area, true); c != 0 {
			return c
		}
		if c := cmp(a.AverageFanout, b.AverageFanout, false); c != 0 {
			return c
		}
	}
	return 1
}

func primitiveTruth(kind AMapGraphType) []uint64 {
	switch kind {
	case AMapGraphAND:
		return []uint64{8}
	case AMapGraphXOR:
		return []uint64{6}
	case AMapGraphMUX:
		return []uint64{0xca}
	}
	return nil
}

// MapAMapGraph maps only explicit one-node AND/XOR/MUX primitives to an exact
// library truth. It is not Amap_ManMerge plus Amap_ManMap: DSD-derived gate
// rules, multi-level cuts, choice merging, flow/exact area iterations,
// polarity reference costing, and area-delay-ratio selection remain absent.
func MapAMapGraph(g *AMapGraph, lib *AMapCellLibrary, params AMapParams) ([]AMapMatch, error) {
	if g == nil || lib == nil {
		return nil, ErrDimension
	}
	if e := validateAMapParams(params); e != nil {
		return nil, e
	}
	matches := make([]AMapMatch, 0)
	delay := make([]float64, len(g.Nodes))
	for i := range g.Nodes {
		if amapStopped(g.limits) {
			return nil, ErrCancelled
		}
		n := g.Nodes[i]
		truth := primitiveTruth(n.Type)
		if truth == nil {
			continue
		}
		pins := len(n.Fanins)
		gate, ok := lib.FindTruth(pins, truth)
		if !ok {
			return nil, ErrVerilogUndefined
		}
		maxDelay := float64(0)
		fan := float64(0)
		for _, lit := range n.Fanins {
			d := delay[amapLitVar(lit)]
			if d > maxDelay {
				maxDelay = d
			}
			fan += float64(g.Nodes[amapLitVar(lit)].References)
		}
		m := AMapMatch{Node: i, Gate: gate, Area: lib.Gates[gate].Area, Delay: maxDelay + 1, AverageFanout: fan / float64(pins), PinPermutation: make([]int, pins)}
		for j := range m.PinPermutation {
			m.PinPermutation[j] = j
		}
		delay[i] = m.Delay
		matches = append(matches, m)
	}
	return matches, nil
}

// ProduceAMapOutput materializes direct matches with gate-pin permutation and
// shared polarity inverters. Native used-cut pruning depends on the absent
// reference-driven cut mapper and is therefore outside this direct path.
func ProduceAMapOutput(g *AMapGraph, lib *AMapCellLibrary, matches []AMapMatch) ([]AMapOutput, error) {
	if g == nil || lib == nil {
		return nil, ErrDimension
	}
	byNode := make(map[int]AMapMatch, len(matches))
	for _, m := range matches {
		if m.Node < 0 || m.Node >= len(g.Nodes) || m.Gate < 0 || m.Gate >= len(lib.Gates) {
			return nil, ErrInvariant
		}
		if _, exists := byNode[m.Node]; exists {
			return nil, ErrVerilogDuplicate
		}
		byNode[m.Node] = m
	}
	out := make([]AMapOutput, 0, len(g.Nodes))
	polarities := make([][2]int, len(g.Nodes))
	for index := range polarities {
		polarities[index] = [2]int{-1, -1}
	}
	var resolve func(int) (int, error)
	resolve = func(literal int) (int, error) {
		nodeID, complement := amapLitVar(literal), literal&1
		if nodeID < 0 || nodeID >= len(g.Nodes) {
			return 0, ErrInvariant
		}
		if nodeID == g.Const1() && polarities[nodeID][complement] < 0 {
			gateID := lib.Constant1
			if complement != 0 {
				gateID = lib.Constant0
			}
			if gateID < 0 || gateID >= len(lib.Gates) {
				return 0, ErrVerilogUndefined
			}
			polarities[nodeID][complement] = len(out)
			out = append(out, AMapOutput{Name: lib.Gates[gateID].Name, Type: 0})
		}
		if polarities[nodeID][complement] >= 0 {
			return polarities[nodeID][complement], nil
		}
		opposite := complement ^ 1
		if polarities[nodeID][opposite] < 0 || lib.Inverter < 0 || lib.Inverter >= len(lib.Gates) {
			return 0, ErrVerilogUndefined
		}
		polarities[nodeID][complement] = len(out)
		out = append(out, AMapOutput{Name: lib.Gates[lib.Inverter].Name, Type: 0, Fanins: []int{polarities[nodeID][opposite]}})
		return polarities[nodeID][complement], nil
	}
	for _, n := range g.Nodes {
		switch n.Type {
		case AMapGraphPI:
			polarities[n.ID][0] = len(out)
			out = append(out, AMapOutput{Name: "", Type: -1})
		case AMapGraphPO:
			if len(n.Fanins) != 1 {
				return nil, ErrInvariant
			}
			fanin, err := resolve(n.Fanins[0])
			if err != nil {
				return nil, err
			}
			out = append(out, AMapOutput{Name: "", Type: 1, Fanins: []int{fanin}})
		case AMapGraphAND, AMapGraphXOR, AMapGraphMUX:
			m, ok := byNode[n.ID]
			if !ok {
				return nil, ErrInvariant
			}
			gate := lib.Gates[m.Gate]
			if len(gate.Pins) != len(n.Fanins) || len(m.PinPermutation) != len(n.Fanins) {
				return nil, ErrInvariant
			}
			seen := make([]bool, len(n.Fanins))
			fanins := make([]int, len(n.Fanins))
			for gatePin, graphPin := range m.PinPermutation {
				if graphPin < 0 || graphPin >= len(n.Fanins) || seen[graphPin] {
					return nil, ErrInvariant
				}
				seen[graphPin] = true
				var err error
				fanins[gatePin], err = resolve(n.Fanins[graphPin])
				if err != nil {
					return nil, err
				}
			}
			polarity := 0
			if m.Inverted {
				polarity = 1
			}
			polarities[n.ID][polarity] = len(out)
			out = append(out, AMapOutput{Name: gate.Name, Type: 0, Fanins: fanins})
		}
	}
	return out, nil
}

// DeriveAMapPermutation returns formula first-occurrence order relative to the
// declared pins. It does not substitute for amapPerm.c's DSD-tree matching,
// complemented-leaf permutation, asymmetric symmetry expansion, or replay.
func DeriveAMapPermutation(formula string, pins []string, limits AMapLimits) ([]int, error) {
	limits, e := normalizeAMapLimits(limits)
	if e != nil {
		return nil, e
	}
	p := amapFormulaParser{text: formula, max: limits.MaxFormulaNodes, stop: limits.Stop}
	tree, e := p.parse()
	if e != nil {
		return nil, e
	}
	names := []string{}
	amapFormulaNames(tree, map[string]bool{}, &names)
	if len(names) != len(pins) {
		return nil, ErrDimension
	}
	index := make(map[string]int, len(pins))
	for i, n := range pins {
		if _, ok := index[n]; ok {
			return nil, ErrVerilogDuplicate
		}
		index[n] = i
	}
	perm := make([]int, len(names))
	for i, n := range names {
		v, ok := index[n]
		if !ok {
			return nil, ErrVerilogUndefined
		}
		perm[i] = v
	}
	return perm, nil
}

// ParseAMapLiberty converts the scalar combinational subset consumed by the
// native AMAP Liberty converter into a prepared library. Sequential cells,
// dont_use cells, and pins without functions are skipped.
func ParseAMapLiberty(name string, data []byte, limits AMapLimits) (*AMapCellLibrary, error) {
	limits, e := normalizeAMapLimits(limits)
	if e != nil {
		return nil, e
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	text := stripAMapCComments(string(data))
	cells, e := scanLibertyGroups(text, "cell")
	if e != nil {
		return nil, e
	}
	var gen strings.Builder
	gen.WriteString("GATE _const0_ 0 z=0;\n")
	gen.WriteString("GATE _const1_ 0 z=1;\n")
	for _, cell := range cells {
		if amapStopped(limits) {
			return nil, ErrCancelled
		}
		cellName, body := cell[0], cell[1]
		flops, groupErr := scanLibertyGroups(body, "ff")
		if groupErr != nil {
			return nil, groupErr
		}
		latches, groupErr := scanLibertyGroups(body, "latch")
		if groupErr != nil {
			return nil, groupErr
		}
		if libertyAttributePresent(body, "dont_use") || len(flops) != 0 || len(latches) != 0 {
			continue
		}
		areaText := libertyString(body, "area")
		if areaText == "" {
			continue
		}
		area, e := strconv.ParseFloat(areaText, 64)
		if e != nil || area < 0 || math.IsNaN(area) || math.IsInf(area, 0) {
			return nil, ErrDimension
		}
		pins, e := scanLibertyGroups(body, "pin")
		if e != nil {
			return nil, e
		}
		inputs := make([]string, 0)
		for _, pin := range pins {
			if libertyString(pin[1], "direction") == "input" {
				inputs = append(inputs, pin[0])
			}
		}
		for _, pin := range pins {
			if libertyString(pin[1], "direction") != "output" {
				continue
			}
			formula := libertyString(pin[1], "function")
			if formula == "" {
				continue
			}
			formula = strings.NewReplacer(" ", "", "~", "!", `\"`, "").Replace(formula)
			if formula == "0" || formula == "1" {
				continue
			}
			fmt.Fprintf(&gen, "GATE %s %.9g %s=%s;\n", cellName, area, pin[0], formula)
			for _, input := range inputs {
				fmt.Fprintf(&gen, "PIN %s UNKNOWN 1 999 1 0 1 0\n", input)
			}
		}
	}
	gen.WriteString(".end\n")
	return ParseAMapLibrary(name, []byte(gen.String()), limits)
}
func stripAMapCComments(s string) string {
	for {
		a := strings.Index(s, "/*")
		if a < 0 {
			return s
		}
		b := strings.Index(s[a+2:], "*/")
		if b < 0 {
			return s[:a]
		}
		s = s[:a] + s[a+2+b+2:]
	}
}
func scanLibertyGroups(s, keyword string) ([][2]string, error) {
	var out [][2]string
	for pos := 0; ; {
		i := strings.Index(s[pos:], keyword)
		if i < 0 {
			return out, nil
		}
		i += pos
		if i > 0 && (unicode.IsLetter(rune(s[i-1])) || s[i-1] == '_') {
			pos = i + len(keyword)
			continue
		}
		p := i + len(keyword)
		for p < len(s) && unicode.IsSpace(rune(s[p])) {
			p++
		}
		if p >= len(s) || s[p] != '(' {
			pos = p
			continue
		}
		q := strings.IndexByte(s[p+1:], ')')
		if q < 0 {
			return nil, ErrVerilogSyntax
		}
		q += p + 1
		name := strings.Trim(strings.TrimSpace(s[p+1:q]), `\"`)
		r := q + 1
		for r < len(s) && unicode.IsSpace(rune(s[r])) {
			r++
		}
		if r >= len(s) || s[r] != '{' {
			return nil, ErrVerilogSyntax
		}
		end, e := matchingBrace(s, r)
		if e != nil {
			return nil, e
		}
		out = append(out, [2]string{name, s[r+1 : end]})
		pos = end + 1
	}
}
func matchingBrace(s string, start int) (int, error) {
	depth := 0
	quoted := false
	for i := start; i < len(s); i++ {
		if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
			quoted = !quoted
		}
		if quoted {
			continue
		}
		if s[i] == '{' {
			depth++
		}
		if s[i] == '}' {
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return 0, ErrVerilogSyntax
}
func libertyString(body, key string) string {
	i := strings.Index(body, key)
	if i < 0 {
		return ""
	}
	offset := strings.Index(body[i:], ":")
	if offset < 0 {
		return ""
	}
	i += offset
	e := strings.IndexByte(body[i+1:], ';')
	if e < 0 {
		return ""
	}
	return strings.Trim(strings.TrimSpace(body[i+1:i+1+e]), `\"`)
}
func libertyNumber(body, key string) (float64, error) {
	s := libertyString(body, key)
	if s == "" {
		return 0, ErrVerilogSyntax
	}
	v, e := strconv.ParseFloat(s, 64)
	if e != nil || v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, ErrDimension
	}
	return v, nil
}
func hasLibertyAttribute(body, key, value string) bool { return libertyString(body, key) == value }

func libertyAttributePresent(body, key string) bool {
	for position := 0; position < len(body); {
		index := strings.Index(body[position:], key)
		if index < 0 {
			return false
		}
		index += position
		beforeOK := index == 0 || !(unicode.IsLetter(rune(body[index-1])) || unicode.IsDigit(rune(body[index-1])) || body[index-1] == '_')
		after := index + len(key)
		afterOK := after == len(body) || !(unicode.IsLetter(rune(body[after])) || unicode.IsDigit(rune(body[after])) || body[after] == '_')
		if beforeOK && afterOK {
			for after < len(body) && unicode.IsSpace(rune(body[after])) {
				after++
			}
			return after < len(body) && body[after] == ':'
		}
		position = index + len(key)
	}
	return false
}
