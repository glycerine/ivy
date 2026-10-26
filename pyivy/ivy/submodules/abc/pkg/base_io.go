package abc

// Derived from the standalone semantics in ABC src/base/io, excluding the
// already separated binary-AIGER implementation.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// IOFormat identifies a concrete byte or text representation.
type IOFormat uint8

const (
	IOFormatUnknown IOFormat = iota
	IOFormatBLIF
	IOFormatBench
	IOFormatEquation
	IOFormatPLA
	IOFormatVerilog
	IOFormatJSON
	IOFormatBAF
	IOFormatBBLIF
	IOFormatBLIFMV
	IOFormatDSD
	IOFormatEDIF
	IOFormatCNF
	IOFormatDOT
	IOFormatGML
	IOFormatList
	IOFormatSMV
	IOFormatBook
)

// IOLimits bounds retained state and all parsing or formatting work.
type IOLimits struct {
	MaxInputBytes  int
	MaxOutputBytes int
	MaxObjects     int
	MaxEdges       int
	MaxCubes       int
	MaxOperations  int
	MaxBytes       int
	MaxDepth       int
	MaxFanins      int
	Stop           func() bool
}

func normalizeIOLimits(l IOLimits) (IOLimits, error) {
	vals := []*int{&l.MaxInputBytes, &l.MaxOutputBytes, &l.MaxObjects, &l.MaxEdges, &l.MaxCubes, &l.MaxOperations, &l.MaxBytes, &l.MaxDepth, &l.MaxFanins}
	defs := []int{1 << 26, 1 << 26, 1 << 20, 1 << 22, 1 << 22, 1 << 27, 1 << 28, 256, 12}
	for i := range vals {
		if *vals[i] < 0 {
			return l, ErrDimension
		}
		if *vals[i] == 0 {
			*vals[i] = defs[i]
		}
	}
	return l, nil
}

type ioBudget struct {
	limits IOLimits
	ops    int
}

func (b *ioBudget) use(n int) error {
	if n < 0 || b.ops > b.limits.MaxOperations-n {
		return ErrLimit
	}
	b.ops += n
	if b.limits.Stop != nil && b.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

// IOCube is one Boolean product. Pattern uses 0, 1 and - in fanin order.
type IOCube struct {
	Pattern string `json:"pattern"`
	Value   bool   `json:"value"`
}

// IONode is one topologically ordered Boolean node.
type IONode struct {
	Name   string   `json:"name"`
	Fanins []string `json:"fanins"`
	Cubes  []IOCube `json:"cubes"`
}

// IOLatch preserves a next-state signal, state name, and ternary initial value.
type IOLatch struct {
	Input  string    `json:"input"`
	Output string    `json:"output"`
	Init   LatchInit `json:"init"`
}

// IONetwork is the owned common representation used by standalone format APIs.
type IONetwork struct {
	Name    string    `json:"name"`
	Inputs  []string  `json:"inputs"`
	Outputs []string  `json:"outputs"`
	Nodes   []IONode  `json:"nodes"`
	Latches []IOLatch `json:"latches,omitempty"`
	closed  bool
}

func cloneIONetwork(n *IONetwork) *IONetwork {
	if n == nil {
		return nil
	}
	r := &IONetwork{Name: n.Name, Inputs: slices.Clone(n.Inputs), Outputs: slices.Clone(n.Outputs), Latches: slices.Clone(n.Latches)}
	r.Nodes = make([]IONode, len(n.Nodes))
	for i := range n.Nodes {
		r.Nodes[i] = IONode{Name: n.Nodes[i].Name, Fanins: slices.Clone(n.Nodes[i].Fanins), Cubes: slices.Clone(n.Nodes[i].Cubes)}
	}
	return r
}

func (n *IONetwork) Close() {
	if n == nil {
		return
	}
	n.Name = ""
	n.Inputs = nil
	n.Outputs = nil
	n.Nodes = nil
	n.Latches = nil
	n.closed = true
}

func validIOName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if !(unicode.IsLetter(r) || r == '_' || r == '$' || i > 0 && unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

func checkedIOAdd(a, b int) (int, error) {
	if a < 0 || b < 0 || a > math.MaxInt-b {
		return 0, ErrLimit
	}
	return a + b, nil
}

// Validate checks syntax, topology, ownership bounds, and cube semantics.
func (n *IONetwork) Validate(limits IOLimits) error {
	l, err := normalizeIOLimits(limits)
	if err != nil {
		return err
	}
	b := ioBudget{limits: l}
	return n.validate(l, &b)
}

func (n *IONetwork) validate(l IOLimits, b *ioBudget) error {
	if n == nil || n.closed {
		return ErrInvariant
	}
	objects, err := checkedIOAdd(len(n.Inputs), len(n.Nodes))
	if err != nil {
		return err
	}
	objects, err = checkedIOAdd(objects, len(n.Latches))
	if err != nil || objects > l.MaxObjects {
		return ErrLimit
	}
	if !validIOName(n.Name) {
		return ErrFormat
	}
	if objects > (l.MaxBytes-len(n.Name))/32 || len(n.Outputs) > (l.MaxBytes-len(n.Name))/8 {
		return ErrLimit
	}
	known := make(map[string]bool, objects+2)
	levels := map[string]int{"0": 0, "1": 0}
	retained := len(n.Name)
	add := func(name string) error {
		if !validIOName(name) || known[name] || name == "0" || name == "1" {
			return ErrFormat
		}
		known[name] = true
		var e error
		retained, e = checkedIOAdd(retained, len(name)+16)
		if e != nil || retained > l.MaxBytes {
			return ErrLimit
		}
		return b.use(1)
	}
	for _, name := range n.Inputs {
		if err = add(name); err != nil {
			return err
		}
		levels[name] = 0
	}
	for _, latch := range n.Latches {
		if latch.Init > LatchInitUnknown || latch.Input != "0" && latch.Input != "1" && !validIOName(latch.Input) {
			return ErrFormat
		}
		if err = add(latch.Output); err != nil {
			return err
		}
		retained, err = checkedIOAdd(retained, len(latch.Input)+16)
		if err != nil || retained > l.MaxBytes {
			return ErrLimit
		}
		levels[latch.Output] = 0
	}
	edges, cubes := len(n.Latches), 0
	if edges > l.MaxEdges {
		return ErrLimit
	}
	for _, node := range n.Nodes {
		if len(node.Fanins) > l.MaxFanins || len(node.Fanins) >= strconv.IntSize {
			return ErrLimit
		}
		if !validIOName(node.Name) || known[node.Name] || node.Name == "0" || node.Name == "1" {
			return ErrFormat
		}
		edges, err = checkedIOAdd(edges, len(node.Fanins))
		if err != nil || edges > l.MaxEdges {
			return ErrLimit
		}
		seenFanins := make(map[string]bool, len(node.Fanins))
		level := 0
		for _, fanin := range node.Fanins {
			if fanin != "0" && fanin != "1" && !known[fanin] {
				return ErrInvariant
			}
			if seenFanins[fanin] {
				return ErrFormat
			}
			seenFanins[fanin] = true
			if levels[fanin]+1 > level {
				level = levels[fanin] + 1
			}
		}
		if level > l.MaxDepth {
			return ErrLimit
		}
		retained, err = checkedIOAdd(retained, len(node.Fanins)*16+len(node.Cubes)*24)
		if err != nil || retained > l.MaxBytes {
			return ErrLimit
		}
		if err = add(node.Name); err != nil {
			return err
		}
		levels[node.Name] = level
		cubes, err = checkedIOAdd(cubes, len(node.Cubes))
		if err != nil || cubes > l.MaxCubes {
			return ErrLimit
		}
		seen := make(map[string]bool, len(node.Cubes))
		for _, cube := range node.Cubes {
			if len(cube.Pattern) != len(node.Fanins) || seen[cube.Pattern] {
				return ErrFormat
			}
			seen[cube.Pattern] = true
			for _, c := range cube.Pattern {
				if c != '0' && c != '1' && c != '-' {
					return ErrFormat
				}
			}
			retained, err = checkedIOAdd(retained, len(cube.Pattern))
			if err != nil || retained > l.MaxBytes {
				return ErrLimit
			}
			if err = b.use(len(cube.Pattern) + 1); err != nil {
				return err
			}
		}
	}
	for _, latch := range n.Latches {
		if latch.Input != "0" && latch.Input != "1" && !known[latch.Input] {
			return ErrInvariant
		}
	}
	seenOutputs := make(map[string]bool, len(n.Outputs))
	for _, output := range n.Outputs {
		if output != "0" && output != "1" && !known[output] {
			return ErrInvariant
		}
		if seenOutputs[output] {
			return ErrFormat
		}
		seenOutputs[output] = true
		retained, err = checkedIOAdd(retained, len(output)+8)
		if err != nil || retained > l.MaxBytes {
			return ErrLimit
		}
		if err = b.use(1); err != nil {
			return err
		}
	}
	return nil
}

func ioCubeMatches(pattern string, values []bool) bool {
	for i, c := range pattern {
		if c == '1' && !values[i] || c == '0' && values[i] {
			return false
		}
	}
	return true
}

// Evaluate returns output values for one input and present-state assignment.
func (n *IONetwork) Evaluate(inputs, state map[string]bool, limits IOLimits) ([]bool, error) {
	l, err := normalizeIOLimits(limits)
	if err != nil {
		return nil, err
	}
	b := ioBudget{limits: l}
	if err = n.validate(l, &b); err != nil {
		return nil, err
	}
	values, err := n.evaluateSignals(inputs, state, &b)
	if err != nil {
		return nil, err
	}
	result := make([]bool, len(n.Outputs))
	for i, name := range n.Outputs {
		result[i] = values[name]
	}
	return result, nil
}

func (n *IONetwork) evaluateSignals(inputs, state map[string]bool, b *ioBudget) (map[string]bool, error) {
	values := make(map[string]bool, len(n.Inputs)+len(n.Nodes)+len(n.Latches)+2)
	values["0"] = false
	values["1"] = true
	for _, name := range n.Inputs {
		v, ok := inputs[name]
		if !ok {
			return nil, ErrDimension
		}
		if err := b.use(1); err != nil {
			return nil, err
		}
		values[name] = v
	}
	for _, latch := range n.Latches {
		v, ok := state[latch.Output]
		if !ok {
			return nil, ErrDimension
		}
		if err := b.use(1); err != nil {
			return nil, err
		}
		values[latch.Output] = v
	}
	for _, node := range n.Nodes {
		args := make([]bool, len(node.Fanins))
		for i, name := range node.Fanins {
			args[i] = values[name]
		}
		value := false
		for _, cube := range node.Cubes {
			if err := b.use(len(args) + 1); err != nil {
				return nil, err
			}
			if ioCubeMatches(cube.Pattern, args) {
				value = cube.Value
			}
		}
		values[node.Name] = value
	}
	return values, nil
}

// NextState evaluates the latch inputs after applying the current state.
func (n *IONetwork) NextState(inputs, state map[string]bool, limits IOLimits) (map[string]bool, error) {
	l, err := normalizeIOLimits(limits)
	if err != nil {
		return nil, err
	}
	b := ioBudget{limits: l}
	if err = n.validate(l, &b); err != nil {
		return nil, err
	}
	values, err := n.evaluateSignals(inputs, state, &b)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(n.Latches))
	for _, latch := range n.Latches {
		result[latch.Output] = values[latch.Input]
	}
	return result, nil
}

// IODetectFormat maps conventional suffixes without touching the filesystem.
func IODetectFormat(name string) IOFormat {
	lower := strings.ToLower(name)
	for _, row := range []struct {
		suffix string
		kind   IOFormat
	}{{".blif", IOFormatBLIF}, {".bench", IOFormatBench}, {".eqn", IOFormatEquation}, {".pla", IOFormatPLA}, {".v", IOFormatVerilog}, {".sv", IOFormatVerilog}, {".json", IOFormatJSON}, {".baf", IOFormatBAF}, {".bblif", IOFormatBBLIF}, {".mv", IOFormatBLIFMV}, {".dsd", IOFormatDSD}, {".edif", IOFormatEDIF}, {".cnf", IOFormatCNF}, {".dot", IOFormatDOT}, {".gml", IOFormatGML}, {".list", IOFormatList}, {".smv", IOFormatSMV}} {
		if strings.HasSuffix(lower, row.suffix) {
			return row.kind
		}
	}
	return IOFormatUnknown
}

func ioLines(data []byte, l IOLimits) ([]string, error) {
	if len(data) > l.MaxInputBytes || len(data) > l.MaxBytes {
		return nil, ErrLimit
	}
	s := bufio.NewScanner(bytes.NewReader(data))
	maxToken := l.MaxInputBytes
	if maxToken < 1024 {
		maxToken = 1024
	}
	s.Buffer(make([]byte, 1024), maxToken)
	lines := make([]string, 0)
	for s.Scan() {
		if l.Stop != nil && l.Stop() {
			return nil, ErrCancelled
		}
		line := strings.TrimSpace(s.Text())
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	if s.Err() != nil {
		return nil, ErrLimit
	}
	return lines, nil
}

type ioBuilder struct {
	n      IONetwork
	known  map[string]bool
	levels map[string]int
	limits IOLimits
	budget ioBudget
	edges  int
	cubes  int
}

func newIOBuilder(name string, limits IOLimits) (*ioBuilder, error) {
	l, err := normalizeIOLimits(limits)
	if err != nil {
		return nil, err
	}
	if !validIOName(name) {
		return nil, ErrFormat
	}
	b := &ioBuilder{n: IONetwork{Name: name}, known: map[string]bool{"0": true, "1": true}, levels: map[string]int{"0": 0, "1": 0}, limits: l}
	b.budget.limits = l
	return b, nil
}

func (b *ioBuilder) input(name string) error {
	if !validIOName(name) || b.known[name] {
		return ErrFormat
	}
	if len(b.n.Inputs)+len(b.n.Nodes)+len(b.n.Latches) >= b.limits.MaxObjects {
		return ErrLimit
	}
	if err := b.budget.use(1); err != nil {
		return err
	}
	b.known[name] = true
	b.levels[name] = 0
	b.n.Inputs = append(b.n.Inputs, name)
	return nil
}

func (b *ioBuilder) node(node IONode) error {
	if !validIOName(node.Name) || b.known[node.Name] {
		return ErrFormat
	}
	if len(node.Fanins) > b.limits.MaxFanins || len(node.Fanins) >= strconv.IntSize || len(b.n.Inputs)+len(b.n.Nodes)+len(b.n.Latches) >= b.limits.MaxObjects {
		return ErrLimit
	}
	seenFanins := make(map[string]bool, len(node.Fanins))
	level := 0
	for _, f := range node.Fanins {
		if !b.known[f] {
			return ErrInvariant
		}
		if seenFanins[f] {
			return ErrFormat
		}
		seenFanins[f] = true
		if b.levels[f]+1 > level {
			level = b.levels[f] + 1
		}
	}
	if level > b.limits.MaxDepth {
		return ErrLimit
	}
	var err error
	b.edges, err = checkedIOAdd(b.edges, len(node.Fanins))
	if err != nil || b.edges > b.limits.MaxEdges {
		return ErrLimit
	}
	b.cubes, err = checkedIOAdd(b.cubes, len(node.Cubes))
	if err != nil || b.cubes > b.limits.MaxCubes {
		return ErrLimit
	}
	if err = b.budget.use(len(node.Fanins) + len(node.Cubes)); err != nil {
		return err
	}
	b.known[node.Name] = true
	b.levels[node.Name] = level
	node.Fanins = slices.Clone(node.Fanins)
	node.Cubes = slices.Clone(node.Cubes)
	b.n.Nodes = append(b.n.Nodes, node)
	return nil
}

func (b *ioBuilder) latch(latch IOLatch) error {
	if !validIOName(latch.Output) || latch.Input != "0" && latch.Input != "1" && !validIOName(latch.Input) || b.known[latch.Output] || latch.Init > LatchInitUnknown {
		return ErrFormat
	}
	if len(b.n.Inputs)+len(b.n.Nodes)+len(b.n.Latches) >= b.limits.MaxObjects {
		return ErrLimit
	}
	if b.edges >= b.limits.MaxEdges {
		return ErrLimit
	}
	if err := b.budget.use(2); err != nil {
		return err
	}
	b.known[latch.Output] = true
	b.levels[latch.Output] = 0
	b.edges++
	b.n.Latches = append(b.n.Latches, latch)
	return nil
}

func (b *ioBuilder) output(name string) error {
	if !b.known[name] {
		return ErrInvariant
	}
	if slices.Contains(b.n.Outputs, name) {
		return ErrFormat
	}
	b.n.Outputs = append(b.n.Outputs, name)
	return b.budget.use(1)
}

func (b *ioBuilder) finish() (*IONetwork, error) {
	if len(b.n.Outputs) == 0 {
		return nil, ErrFormat
	}
	if err := b.n.validate(b.limits, &b.budget); err != nil {
		return nil, err
	}
	return cloneIONetwork(&b.n), nil
}

func primitiveCubes(gate string, n int) ([]IOCube, error) {
	gate = strings.ToUpper(gate)
	if n < 0 || n > 12 {
		return nil, ErrLimit
	}
	if gate == "BUF" && n == 1 {
		return []IOCube{{Pattern: "1", Value: true}}, nil
	}
	if gate == "NOT" && n == 1 {
		return []IOCube{{Pattern: "0", Value: true}}, nil
	}
	pattern := strings.Repeat("1", n)
	switch gate {
	case "AND":
		return []IOCube{{Pattern: pattern, Value: true}}, nil
	case "NAND":
		return []IOCube{{Pattern: strings.Repeat("-", n), Value: true}, {Pattern: pattern, Value: false}}, nil
	case "OR":
		c := make([]IOCube, n)
		for i := range c {
			p := []byte(strings.Repeat("-", n))
			p[i] = '1'
			c[i] = IOCube{Pattern: string(p), Value: true}
		}
		return c, nil
	case "NOR":
		return []IOCube{{Pattern: strings.Repeat("0", n), Value: true}}, nil
	case "XOR", "XNOR":
		c := make([]IOCube, 0, 1<<n)
		for mask := 0; mask < 1<<n; mask++ {
			odd := bitsOnes(mask)&1 != 0
			if odd == (gate == "XOR") {
				c = append(c, IOCube{Pattern: binaryPattern(mask, n), Value: true})
			}
		}
		return c, nil
	}
	return nil, ErrUnsupported
}

func bitsOnes(x int) int {
	n := 0
	for x != 0 {
		n += x & 1
		x >>= 1
	}
	return n
}

func binaryPattern(mask, n int) string {
	p := make([]byte, n)
	for i := 0; i < n; i++ {
		if mask&(1<<i) != 0 {
			p[i] = '1'
		} else {
			p[i] = '0'
		}
	}
	return string(p)
}

func parseBench(data []byte, limits IOLimits) (*IONetwork, error) {
	lines, err := ioLines(data, limits)
	if err != nil {
		return nil, err
	}
	b, err := newIOBuilder("bench", limits)
	if err != nil {
		return nil, err
	}
	if err = b.budget.use(len(data)); err != nil {
		return nil, err
	}
	outs := make([]string, 0)
	for _, line := range lines {
		upper := strings.ToUpper(line)
		if strings.HasPrefix(upper, "INPUT(") && strings.HasSuffix(line, ")") {
			if err = b.input(strings.TrimSpace(line[6 : len(line)-1])); err != nil {
				return nil, err
			}
			continue
		}
		if strings.HasPrefix(upper, "OUTPUT(") && strings.HasSuffix(line, ")") {
			outs = append(outs, strings.TrimSpace(line[7:len(line)-1]))
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq < 1 {
			return nil, ErrFormat
		}
		name, rhs := strings.TrimSpace(line[:eq]), strings.TrimSpace(line[eq+1:])
		open := strings.IndexByte(rhs, '(')
		if open < 1 || !strings.HasSuffix(rhs, ")") {
			return nil, ErrFormat
		}
		gate := strings.TrimSpace(rhs[:open])
		parts := strings.Split(rhs[open+1:len(rhs)-1], ",")
		fanins := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				return nil, ErrFormat
			}
			fanins = append(fanins, p)
		}
		cubes, e := primitiveCubes(gate, len(fanins))
		if e != nil {
			return nil, e
		}
		if err = b.node(IONode{Name: name, Fanins: fanins, Cubes: cubes}); err != nil {
			return nil, err
		}
	}
	for _, name := range outs {
		if err = b.output(name); err != nil {
			return nil, err
		}
	}
	return b.finish()
}

func parseBLIF(data []byte, limits IOLimits) (*IONetwork, error) {
	lines, err := ioLines(data, limits)
	if err != nil {
		return nil, err
	}
	name := "blif"
	for _, line := range lines {
		if strings.HasPrefix(line, ".model ") {
			name = strings.TrimSpace(line[7:])
			break
		}
	}
	b, err := newIOBuilder(name, limits)
	if err != nil {
		return nil, err
	}
	if err = b.budget.use(len(data)); err != nil {
		return nil, err
	}
	outs := make([]string, 0)
	ended := false
	for i := 0; i < len(lines); {
		line := lines[i]
		fields := strings.Fields(line)
		if len(fields) == 0 {
			i++
			continue
		}
		switch fields[0] {
		case ".model":
			if len(fields) != 2 || i != 0 {
				return nil, ErrFormat
			}
			i++
		case ".inputs":
			for _, s := range fields[1:] {
				if err = b.input(s); err != nil {
					return nil, err
				}
			}
			i++
		case ".outputs":
			outs = append(outs, fields[1:]...)
			i++
		case ".latch":
			if len(fields) < 3 || len(fields) > 6 {
				return nil, ErrFormat
			}
			init := LatchInitUnknown
			last := fields[len(fields)-1]
			if last == "0" {
				init = LatchInitZero
			} else if last == "1" {
				init = LatchInitOne
			}
			if err = b.latch(IOLatch{Input: fields[1], Output: fields[2], Init: init}); err != nil {
				return nil, err
			}
			i++
		case ".names":
			if len(fields) < 2 {
				return nil, ErrFormat
			}
			node := IONode{Name: fields[len(fields)-1], Fanins: slices.Clone(fields[1 : len(fields)-1])}
			rows := make([]IOCube, 0)
			phaseSet, phase := false, false
			i++
			for i < len(lines) && !strings.HasPrefix(lines[i], ".") {
				cf := strings.Fields(lines[i])
				if len(node.Fanins) == 0 {
					if len(cf) != 1 || cf[0] != "0" && cf[0] != "1" {
						return nil, ErrFormat
					}
					value := cf[0] == "1"
					if phaseSet && phase != value {
						return nil, ErrUnsupported
					}
					phaseSet, phase = true, value
					rows = append(rows, IOCube{Pattern: "", Value: value})
				} else {
					if len(cf) != 2 || len(cf[0]) != len(node.Fanins) || cf[1] != "0" && cf[1] != "1" {
						return nil, ErrFormat
					}
					for _, ch := range cf[0] {
						if ch != '0' && ch != '1' && ch != '-' {
							return nil, ErrFormat
						}
					}
					value := cf[1] == "1"
					if phaseSet && phase != value {
						return nil, ErrUnsupported
					}
					phaseSet, phase = true, value
					rows = append(rows, IOCube{Pattern: cf[0], Value: value})
				}
				if len(rows) > b.limits.MaxCubes {
					return nil, ErrLimit
				}
				i++
			}
			if phaseSet && !phase {
				values := make([]bool, len(node.Fanins))
				for mask := 0; mask < 1<<len(node.Fanins); mask++ {
					if err = b.budget.use(len(node.Fanins) + len(rows) + 1); err != nil {
						return nil, err
					}
					for j := range values {
						values[j] = mask&(1<<j) != 0
					}
					matched := false
					for _, row := range rows {
						matched = matched || ioCubeMatches(row.Pattern, values)
					}
					if !matched {
						if len(node.Cubes) >= b.limits.MaxCubes-b.cubes {
							return nil, ErrLimit
						}
						node.Cubes = append(node.Cubes, IOCube{Pattern: binaryPattern(mask, len(values)), Value: true})
					}
				}
			} else if phaseSet {
				node.Cubes = rows
			}
			if err = b.node(node); err != nil {
				return nil, err
			}
		case ".end":
			if len(fields) != 1 || i != len(lines)-1 {
				return nil, ErrFormat
			}
			ended = true
			i++
		default:
			return nil, ErrUnsupported
		}
	}
	if !ended {
		return nil, ErrFormat
	}
	for _, s := range outs {
		if err = b.output(s); err != nil {
			return nil, err
		}
	}
	return b.finish()
}

func parsePLA(data []byte, limits IOLimits) (*IONetwork, error) {
	lines, err := ioLines(data, limits)
	if err != nil {
		return nil, err
	}
	nIn, nOut := -1, -1
	var inputs, outputs []string
	type row struct{ in, out string }
	rows := make([]row, 0)
	ended := false
	seenI, seenO, seenILB, seenOB, seenType := false, false, false, false, false
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if ended {
			return nil, ErrFormat
		}
		switch f[0] {
		case ".i", ".o":
			if len(f) != 2 {
				return nil, ErrFormat
			}
			n, e := strconv.Atoi(f[1])
			if e != nil || n < 0 {
				return nil, ErrFormat
			}
			if f[0] == ".i" {
				if seenI {
					return nil, ErrFormat
				}
				seenI = true
				nIn = n
			} else {
				if seenO {
					return nil, ErrFormat
				}
				seenO = true
				nOut = n
			}
		case ".ilb":
			if seenILB {
				return nil, ErrFormat
			}
			seenILB = true
			inputs = slices.Clone(f[1:])
		case ".ob":
			if seenOB {
				return nil, ErrFormat
			}
			seenOB = true
			outputs = slices.Clone(f[1:])
		case ".type":
			if seenType || len(f) != 2 || f[1] != "f" {
				return nil, ErrUnsupported
			}
			seenType = true
		case ".e":
			if len(f) != 1 {
				return nil, ErrFormat
			}
			ended = true
		default:
			if strings.HasPrefix(f[0], ".") || len(f) != 2 {
				return nil, ErrFormat
			}
			rows = append(rows, row{f[0], f[1]})
			if len(rows) > limits.MaxCubes && limits.MaxCubes > 0 {
				return nil, ErrLimit
			}
		}
	}
	if !ended || nIn < 0 || nOut < 1 {
		return nil, ErrFormat
	}
	if len(inputs) == 0 {
		for i := 0; i < nIn; i++ {
			inputs = append(inputs, fmt.Sprintf("x%d", i))
		}
	}
	if len(outputs) == 0 {
		for i := 0; i < nOut; i++ {
			outputs = append(outputs, fmt.Sprintf("y%d", i))
		}
	}
	if len(inputs) != nIn || len(outputs) != nOut {
		return nil, ErrFormat
	}
	b, err := newIOBuilder("pla", limits)
	if err != nil {
		return nil, err
	}
	if err = b.budget.use(len(data)); err != nil {
		return nil, err
	}
	if nOut != 0 && len(rows) > b.limits.MaxCubes/nOut {
		return nil, ErrLimit
	}
	for _, s := range inputs {
		if err = b.input(s); err != nil {
			return nil, err
		}
	}
	for j, out := range outputs {
		cubes := make([]IOCube, 0)
		for _, r := range rows {
			if len(r.in) != nIn || len(r.out) != nOut {
				return nil, ErrFormat
			}
			for _, c := range r.in {
				if c != '0' && c != '1' && c != '-' {
					return nil, ErrFormat
				}
			}
			if r.out[j] == '1' {
				cubes = append(cubes, IOCube{Pattern: r.in, Value: true})
			} else if r.out[j] != '0' && r.out[j] != '-' {
				return nil, ErrFormat
			}
		}
		if err = b.node(IONode{Name: out, Fanins: slices.Clone(inputs), Cubes: cubes}); err != nil {
			return nil, err
		}
		if err = b.output(out); err != nil {
			return nil, err
		}
	}
	return b.finish()
}

type ioExpr struct {
	op          byte
	name        string
	left, right *ioExpr
	depth       int
}
type ioExprParser struct {
	text       string
	pos, depth int
	limits     IOLimits
	budget     *ioBudget
}

func (p *ioExprParser) combine(op byte, left, right *ioExpr) (*ioExpr, error) {
	depth := left.depth + 1
	if right != nil && right.depth >= left.depth {
		depth = right.depth + 1
	}
	if depth > p.limits.MaxDepth {
		return nil, ErrLimit
	}
	if err := p.budget.use(1); err != nil {
		return nil, err
	}
	return &ioExpr{op: op, left: left, right: right, depth: depth}, nil
}

func (p *ioExprParser) skip() {
	for p.pos < len(p.text) && unicode.IsSpace(rune(p.text[p.pos])) {
		p.pos++
	}
}
func (p *ioExprParser) parse() (*ioExpr, error) {
	e, err := p.parseOr()
	p.skip()
	if err != nil {
		return nil, err
	}
	if p.pos != len(p.text) {
		return nil, ErrFormat
	}
	return e, nil
}
func (p *ioExprParser) parseOr() (*ioExpr, error) {
	e, err := p.parseXor()
	for err == nil {
		p.skip()
		if p.pos >= len(p.text) || (p.text[p.pos] != '|' && p.text[p.pos] != '+') {
			break
		}
		p.pos++
		var r *ioExpr
		r, err = p.parseXor()
		if err == nil {
			e, err = p.combine('|', e, r)
		}
	}
	return e, err
}
func (p *ioExprParser) parseXor() (*ioExpr, error) {
	e, err := p.parseAnd()
	for err == nil {
		p.skip()
		if p.pos >= len(p.text) || p.text[p.pos] != '^' {
			break
		}
		p.pos++
		var r *ioExpr
		r, err = p.parseAnd()
		if err == nil {
			e, err = p.combine('^', e, r)
		}
	}
	return e, err
}
func (p *ioExprParser) parseAnd() (*ioExpr, error) {
	e, err := p.parseUnary()
	for err == nil {
		p.skip()
		if p.pos >= len(p.text) || (p.text[p.pos] != '&' && p.text[p.pos] != '*') {
			break
		}
		p.pos++
		var r *ioExpr
		r, err = p.parseUnary()
		if err == nil {
			e, err = p.combine('&', e, r)
		}
	}
	return e, err
}
func (p *ioExprParser) parseUnary() (*ioExpr, error) {
	p.skip()
	negations := 0
	for p.pos < len(p.text) && (p.text[p.pos] == '!' || p.text[p.pos] == '~') {
		p.pos++
		negations++
		p.skip()
		if negations >= p.limits.MaxDepth {
			return nil, ErrLimit
		}
	}
	e, err := p.parseAtom()
	for err == nil && negations > 0 {
		e, err = p.combine('!', e, nil)
		negations--
	}
	return e, err
}
func (p *ioExprParser) parseAtom() (*ioExpr, error) {
	p.skip()
	if p.pos >= len(p.text) {
		return nil, ErrFormat
	}
	if p.text[p.pos] == '(' {
		p.depth++
		if p.depth > p.limits.MaxDepth {
			return nil, ErrLimit
		}
		p.pos++
		e, err := p.parseOr()
		p.skip()
		if err != nil {
			return nil, err
		}
		if p.pos >= len(p.text) || p.text[p.pos] != ')' {
			return nil, ErrFormat
		}
		p.pos++
		p.depth--
		return e, nil
	}
	start := p.pos
	for p.pos < len(p.text) && (unicode.IsLetter(rune(p.text[p.pos])) || unicode.IsDigit(rune(p.text[p.pos])) || p.text[p.pos] == '_' || p.text[p.pos] == '$') {
		p.pos++
	}
	if start == p.pos {
		return nil, ErrFormat
	}
	name := p.text[start:p.pos]
	if name != "0" && name != "1" && !validIOName(name) {
		return nil, ErrFormat
	}
	if err := p.budget.use(1); err != nil {
		return nil, err
	}
	return &ioExpr{op: 'v', name: name, depth: 1}, nil
}
func exprVars(e *ioExpr, seen map[string]bool, out *[]string) {
	if e == nil {
		return
	}
	if e.op == 'v' && e.name != "0" && e.name != "1" && !seen[e.name] {
		seen[e.name] = true
		*out = append(*out, e.name)
	}
	exprVars(e.left, seen, out)
	exprVars(e.right, seen, out)
}
func evalExpr(e *ioExpr, vals map[string]bool) bool {
	switch e.op {
	case 'v':
		if e.name == "1" {
			return true
		}
		return vals[e.name]
	case '!':
		return !evalExpr(e.left, vals)
	case '&':
		return evalExpr(e.left, vals) && evalExpr(e.right, vals)
	case '|':
		return evalExpr(e.left, vals) || evalExpr(e.right, vals)
	case '^':
		return evalExpr(e.left, vals) != evalExpr(e.right, vals)
	}
	return false
}
func exprNode(name, text string, b *ioBuilder) (IONode, error) {
	text = strings.ReplaceAll(text, "1'b0", "0")
	text = strings.ReplaceAll(text, "1'b1", "1")
	if err := b.budget.use(len(text)); err != nil {
		return IONode{}, err
	}
	p := ioExprParser{text: text, limits: b.limits, budget: &b.budget}
	e, err := p.parse()
	if err != nil {
		return IONode{}, err
	}
	vars := make([]string, 0)
	exprVars(e, make(map[string]bool), &vars)
	if len(vars) > b.limits.MaxFanins || len(vars) >= strconv.IntSize {
		return IONode{}, ErrLimit
	}
	for _, v := range vars {
		if !b.known[v] {
			return IONode{}, ErrInvariant
		}
	}
	cubes := make([]IOCube, 0)
	vals := make(map[string]bool, len(vars))
	for mask := 0; mask < 1<<len(vars); mask++ {
		if err = b.budget.use(len(vars) + 1); err != nil {
			return IONode{}, err
		}
		for i, v := range vars {
			vals[v] = mask&(1<<i) != 0
		}
		if evalExpr(e, vals) {
			cubes = append(cubes, IOCube{Pattern: binaryPattern(mask, len(vars)), Value: true})
		}
	}
	return IONode{Name: name, Fanins: vars, Cubes: cubes}, nil
}

func parseEquation(data []byte, limits IOLimits) (*IONetwork, error) {
	lines, err := ioLines(data, limits)
	if err != nil {
		return nil, err
	}
	b, err := newIOBuilder("eqn", limits)
	if err != nil {
		return nil, err
	}
	if err = b.budget.use(len(data)); err != nil {
		return nil, err
	}
	outs := []string{}
	for _, line := range lines {
		for _, stmt := range strings.Split(line, ";") {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			eq := strings.IndexByte(stmt, '=')
			if eq < 1 {
				return nil, ErrFormat
			}
			lhs, rhs := strings.TrimSpace(stmt[:eq]), strings.TrimSpace(stmt[eq+1:])
			switch strings.ToUpper(lhs) {
			case "INORDER":
				for _, v := range strings.Fields(rhs) {
					if err = b.input(v); err != nil {
						return nil, err
					}
				}
			case "OUTORDER":
				outs = append(outs, strings.Fields(rhs)...)
			default:
				n, er := exprNode(lhs, rhs, b)
				if er != nil {
					return nil, er
				}
				if err = b.node(n); err != nil {
					return nil, err
				}
			}
		}
	}
	for _, v := range outs {
		if err = b.output(v); err != nil {
			return nil, err
		}
	}
	return b.finish()
}

func parseVerilog(data []byte, limits IOLimits) (*IONetwork, error) {
	if len(data) > limits.MaxInputBytes && limits.MaxInputBytes > 0 {
		return nil, ErrLimit
	}
	text := string(data)
	text = strings.ReplaceAll(text, "\n", " ")
	stmts := strings.Split(text, ";")
	if len(stmts) < 3 {
		return nil, ErrFormat
	}
	header := strings.TrimSpace(stmts[0])
	if !strings.HasPrefix(header, "module ") {
		return nil, ErrFormat
	}
	header = strings.TrimSpace(strings.TrimPrefix(header, "module"))
	open, close := strings.IndexByte(header, '('), strings.LastIndexByte(header, ')')
	if open <= 0 || close < open || strings.TrimSpace(header[close+1:]) != "" {
		return nil, ErrFormat
	}
	name := strings.TrimSpace(header[:open])
	headerPorts := make(map[string]bool)
	for _, port := range strings.Split(header[open+1:close], ",") {
		port = strings.TrimSpace(port)
		if !validIOName(port) || headerPorts[port] {
			return nil, ErrFormat
		}
		headerPorts[port] = true
	}
	b, err := newIOBuilder(name, limits)
	if err != nil {
		return nil, err
	}
	if err = b.budget.use(len(data)); err != nil {
		return nil, err
	}
	outs := []string{}
	declaredPorts := make(map[string]bool, len(headerPorts))
	declaredWires := make(map[string]bool)
	driven := make(map[string]bool)
	ended := false
	for i := 1; i < len(stmts); i++ {
		s := strings.TrimSpace(stmts[i])
		if s == "" {
			continue
		}
		if s == "endmodule" {
			ended = true
			for _, tail := range stmts[i+1:] {
				if strings.TrimSpace(tail) != "" {
					return nil, ErrFormat
				}
			}
			break
		}
		f := strings.Fields(s)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "input", "output", "wire":
			names := strings.Split(strings.TrimSpace(strings.TrimPrefix(s, f[0])), ",")
			for _, v := range names {
				v = strings.TrimSpace(strings.Trim(v, "()"))
				if !validIOName(v) || declaredPorts[v] || declaredWires[v] {
					return nil, ErrFormat
				}
				if f[0] == "input" {
					if !headerPorts[v] {
						return nil, ErrFormat
					}
					if err = b.input(v); err != nil {
						return nil, err
					}
					declaredPorts[v] = true
				} else if f[0] == "output" {
					if !headerPorts[v] {
						return nil, ErrFormat
					}
					declaredPorts[v] = true
					outs = append(outs, v)
				} else {
					if headerPorts[v] {
						return nil, ErrFormat
					}
					declaredWires[v] = true
				}
			}
		case "assign":
			eq := strings.IndexByte(s, '=')
			if eq < 0 {
				return nil, ErrFormat
			}
			lhs := strings.TrimSpace(s[len("assign"):eq])
			rhs := strings.TrimSpace(s[eq+1:])
			if (!declaredPorts[lhs] && !declaredWires[lhs]) || driven[lhs] || slices.Contains(b.n.Inputs, lhs) {
				return nil, ErrFormat
			}
			n, er := exprNode(lhs, rhs, b)
			if er != nil {
				return nil, er
			}
			if err = b.node(n); err != nil {
				return nil, err
			}
			driven[lhs] = true
		default:
			return nil, ErrUnsupported
		}
	}
	if !ended {
		return nil, ErrFormat
	}
	if len(declaredPorts) != len(headerPorts) {
		return nil, ErrFormat
	}
	for _, v := range outs {
		if !driven[v] {
			return nil, ErrFormat
		}
		if err = b.output(v); err != nil {
			return nil, err
		}
	}
	return b.finish()
}

type ioJSONNetwork struct {
	Name    string    `json:"name"`
	Inputs  []string  `json:"inputs"`
	Outputs []string  `json:"outputs"`
	Nodes   []IONode  `json:"nodes"`
	Latches []IOLatch `json:"latches,omitempty"`
}

func parseJSON(data []byte, limits IOLimits) (*IONetwork, error) {
	l, err := normalizeIOLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > l.MaxInputBytes || len(data) > l.MaxBytes {
		return nil, ErrLimit
	}
	b := ioBudget{limits: l}
	if err = b.use(len(data)); err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var j ioJSONNetwork
	if err = dec.Decode(&j); err != nil {
		return nil, ErrFormat
	}
	var extra json.RawMessage
	if err = dec.Decode(&extra); err != io.EOF {
		return nil, ErrFormat
	}
	n := &IONetwork{Name: j.Name, Inputs: slices.Clone(j.Inputs), Outputs: slices.Clone(j.Outputs), Nodes: slices.Clone(j.Nodes), Latches: slices.Clone(j.Latches)}
	for i := range n.Nodes {
		n.Nodes[i].Fanins = slices.Clone(j.Nodes[i].Fanins)
		n.Nodes[i].Cubes = slices.Clone(j.Nodes[i].Cubes)
	}
	if err = n.validate(l, &b); err != nil {
		return nil, err
	}
	return n, nil
}

// ParseIONetwork parses one supported owned representation atomically.
func ParseIONetwork(data []byte, format IOFormat, limits IOLimits) (*IONetwork, error) {
	l, err := normalizeIOLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > l.MaxInputBytes || len(data) > l.MaxBytes {
		return nil, ErrLimit
	}
	if l.Stop != nil && l.Stop() {
		return nil, ErrCancelled
	}
	switch format {
	case IOFormatBench:
		return parseBench(data, l)
	case IOFormatBLIF:
		return parseBLIF(data, l)
	case IOFormatPLA:
		return parsePLA(data, l)
	case IOFormatEquation:
		return parseEquation(data, l)
	case IOFormatVerilog:
		return parseVerilog(data, l)
	case IOFormatJSON:
		return parseJSON(data, l)
	default:
		return nil, ErrUnsupported
	}
}

type ioTextWriter struct {
	b      strings.Builder
	limit  int
	budget *ioBudget
}

func (w *ioTextWriter) add(s string) error {
	if len(s) > w.limit-w.b.Len() {
		return ErrLimit
	}
	if err := w.budget.use(len(s) + 1); err != nil {
		return err
	}
	w.b.WriteString(s)
	return nil
}
func sopExpr(node IONode, budget *ioBudget) (string, error) {
	cubes, err := ioNodeOnset(node, budget)
	if err != nil {
		return "", err
	}
	if len(cubes) == 0 {
		return "1'b0", nil
	}
	terms := make([]string, 0, len(cubes))
	tempBytes := 0
	for _, c := range cubes {
		parts := make([]string, 0)
		for i, ch := range c.Pattern {
			if ch == '1' {
				tempBytes, err = checkedIOAdd(tempBytes, len(node.Fanins[i])+4)
				parts = append(parts, node.Fanins[i])
			} else if ch == '0' {
				tempBytes, err = checkedIOAdd(tempBytes, len(node.Fanins[i])+5)
				parts = append(parts, "!"+node.Fanins[i])
			}
			if err != nil || tempBytes > budget.limits.MaxBytes {
				return "", ErrLimit
			}
		}
		if len(parts) == 0 {
			terms = append(terms, "1'b1")
		} else {
			terms = append(terms, "("+strings.Join(parts, " & ")+")")
		}
	}
	if len(terms) == 0 {
		return "1'b0", nil
	}
	return strings.Join(terms, " | "), nil
}

func ioNodeOnset(node IONode, budget *ioBudget) ([]IOCube, error) {
	if len(node.Fanins) >= strconv.IntSize {
		return nil, ErrLimit
	}
	combinations := 1 << len(node.Fanins)
	unit := len(node.Fanins) + 24
	if unit <= 0 || combinations > budget.limits.MaxBytes/unit {
		return nil, ErrLimit
	}
	result := make([]IOCube, 0)
	values := make([]bool, len(node.Fanins))
	for mask := 0; mask < combinations; mask++ {
		if err := budget.use(len(node.Fanins) + len(node.Cubes) + 1); err != nil {
			return nil, err
		}
		for i := range values {
			values[i] = mask&(1<<i) != 0
		}
		value := false
		for _, cube := range node.Cubes {
			if ioCubeMatches(cube.Pattern, values) {
				value = cube.Value
			}
		}
		if value {
			result = append(result, IOCube{Pattern: binaryPattern(mask, len(values)), Value: true})
		}
	}
	return result, nil
}
func writeBLIF(n *IONetwork, w *ioTextWriter) error {
	if err := w.add(".model " + n.Name + "\n.inputs " + strings.Join(n.Inputs, " ") + "\n.outputs " + strings.Join(n.Outputs, " ") + "\n"); err != nil {
		return err
	}
	for _, l := range n.Latches {
		init := "2"
		if l.Init == LatchInitZero {
			init = "0"
		} else if l.Init == LatchInitOne {
			init = "1"
		}
		if err := w.add(".latch " + l.Input + " " + l.Output + " " + init + "\n"); err != nil {
			return err
		}
	}
	for _, node := range n.Nodes {
		line := ".names"
		if len(node.Fanins) > 0 {
			line += " " + strings.Join(node.Fanins, " ")
		}
		line += " " + node.Name + "\n"
		if err := w.add(line); err != nil {
			return err
		}
		onset, err := ioNodeOnset(node, w.budget)
		if err != nil {
			return err
		}
		for _, c := range onset {
			if len(node.Fanins) == 0 {
				line = "1\n"
			} else {
				line = c.Pattern + " 1\n"
			}
			if err := w.add(line); err != nil {
				return err
			}
		}
	}
	return w.add(".end\n")
}
func writeBench(n *IONetwork, w *ioTextWriter) error {
	if len(n.Latches) != 0 {
		return ErrUnsupported
	}
	for _, s := range n.Inputs {
		if err := w.add("INPUT(" + s + ")\n"); err != nil {
			return err
		}
	}
	for _, s := range n.Outputs {
		if err := w.add("OUTPUT(" + s + ")\n"); err != nil {
			return err
		}
	}
	for _, node := range n.Nodes {
		gate := ""
		c := node.Cubes
		if len(node.Fanins) == 1 && len(c) == 1 && c[0].Pattern == "1" {
			gate = "BUF"
		} else if len(node.Fanins) == 1 && len(c) == 1 && c[0].Pattern == "0" {
			gate = "NOT"
		} else if len(c) == 1 && c[0].Pattern == strings.Repeat("1", len(node.Fanins)) {
			gate = "AND"
		}
		if gate == "" {
			return ErrUnsupported
		}
		if err := w.add(node.Name + " = " + gate + "(" + strings.Join(node.Fanins, ", ") + ")\n"); err != nil {
			return err
		}
	}
	return nil
}
func writeEquation(n *IONetwork, w *ioTextWriter) error {
	if len(n.Latches) != 0 {
		return ErrUnsupported
	}
	if err := w.add("INORDER = " + strings.Join(n.Inputs, " ") + ";\nOUTORDER = " + strings.Join(n.Outputs, " ") + ";\n"); err != nil {
		return err
	}
	for _, node := range n.Nodes {
		expr, err := sopExpr(node, w.budget)
		if err != nil {
			return err
		}
		if err := w.add(node.Name + " = " + strings.ReplaceAll(expr, "1'b", " ") + ";\n"); err != nil {
			return err
		}
	}
	return nil
}
func writeVerilog(n *IONetwork, w *ioTextWriter) error {
	if len(n.Latches) != 0 {
		return ErrUnsupported
	}
	nodeNames := make(map[string]bool, len(n.Nodes))
	for _, node := range n.Nodes {
		nodeNames[node.Name] = true
	}
	inputNames := make(map[string]bool, len(n.Inputs))
	for _, input := range n.Inputs {
		inputNames[input] = true
	}
	for _, output := range n.Outputs {
		if inputNames[output] || !nodeNames[output] {
			return ErrUnsupported
		}
	}
	ports := append(slices.Clone(n.Inputs), n.Outputs...)
	if err := w.add("module " + n.Name + "(" + strings.Join(ports, ", ") + ");\n"); err != nil {
		return err
	}
	if len(n.Inputs) > 0 {
		if err := w.add("input " + strings.Join(n.Inputs, ", ") + ";\n"); err != nil {
			return err
		}
	}
	if len(n.Outputs) > 0 {
		if err := w.add("output " + strings.Join(n.Outputs, ", ") + ";\n"); err != nil {
			return err
		}
	}
	out := make(map[string]bool, len(n.Outputs))
	for _, s := range n.Outputs {
		out[s] = true
	}
	wires := []string{}
	for _, node := range n.Nodes {
		if !out[node.Name] {
			wires = append(wires, node.Name)
		}
	}
	if len(wires) > 0 {
		if err := w.add("wire " + strings.Join(wires, ", ") + ";\n"); err != nil {
			return err
		}
	}
	for _, node := range n.Nodes {
		expr, err := sopExpr(node, w.budget)
		if err != nil {
			return err
		}
		if err := w.add("assign " + node.Name + " = " + expr + ";\n"); err != nil {
			return err
		}
	}
	return w.add("endmodule;\n")
}
func writePLA(n *IONetwork, w *ioTextWriter) error {
	if len(n.Latches) > 0 {
		return ErrUnsupported
	}
	nodes := make(map[string]IONode, len(n.Nodes))
	for _, x := range n.Nodes {
		nodes[x.Name] = x
	}
	for _, o := range n.Outputs {
		x, ok := nodes[o]
		if !ok || !slices.Equal(x.Fanins, n.Inputs) {
			return ErrUnsupported
		}
	}
	if err := w.add(fmt.Sprintf(".i %d\n.o %d\n.ilb %s\n.ob %s\n.type f\n", len(n.Inputs), len(n.Outputs), strings.Join(n.Inputs, " "), strings.Join(n.Outputs, " "))); err != nil {
		return err
	}
	patterns := make(map[string][]byte)
	order := []string{}
	tempBytes := 0
	for j, o := range n.Outputs {
		onset, err := ioNodeOnset(nodes[o], w.budget)
		if err != nil {
			return err
		}
		for _, c := range onset {
			v, ok := patterns[c.Pattern]
			if !ok {
				var e error
				tempBytes, e = checkedIOAdd(tempBytes, len(c.Pattern)+len(n.Outputs)+32)
				if e != nil || tempBytes > w.budget.limits.MaxBytes {
					return ErrLimit
				}
				v = []byte(strings.Repeat("0", len(n.Outputs)))
				order = append(order, c.Pattern)
			}
			v[j] = '1'
			patterns[c.Pattern] = v
		}
	}
	slices.Sort(order)
	for _, p := range order {
		if err := w.add(p + " " + string(patterns[p]) + "\n"); err != nil {
			return err
		}
	}
	return w.add(".e\n")
}
func writeJSON(n *IONetwork, w *ioTextWriter) error {
	j := ioJSONNetwork{Name: n.Name, Inputs: n.Inputs, Outputs: n.Outputs, Nodes: n.Nodes, Latches: n.Latches}
	data, err := json.Marshal(j)
	if err != nil {
		return ErrInvariant
	}
	if err = w.add(string(data)); err != nil {
		return err
	}
	return w.add("\n")
}
func writeGraph(n *IONetwork, w *ioTextWriter, gml bool) error {
	if gml {
		if err := w.add("graph [\n  directed 1\n"); err != nil {
			return err
		}
		for i, node := range n.Nodes {
			if err := w.add(fmt.Sprintf("  node [ id %d label \"%s\" ]\n", i, node.Name)); err != nil {
				return err
			}
		}
		ids := make(map[string]int, len(n.Nodes))
		for i, node := range n.Nodes {
			ids[node.Name] = i
		}
		for _, node := range n.Nodes {
			for _, fanin := range node.Fanins {
				if source, ok := ids[fanin]; ok {
					if err := w.add(fmt.Sprintf("  edge [ source %d target %d ]\n", source, ids[node.Name])); err != nil {
						return err
					}
				}
			}
		}
		return w.add("]\n")
	}
	if err := w.add("digraph " + n.Name + " {\n"); err != nil {
		return err
	}
	for _, node := range n.Nodes {
		if err := w.add("  \"" + node.Name + "\";\n"); err != nil {
			return err
		}
		for _, f := range node.Fanins {
			if err := w.add("  \"" + f + "\" -> \"" + node.Name + "\";\n"); err != nil {
				return err
			}
		}
	}
	return w.add("}\n")
}
func writeList(n *IONetwork, w *ioTextWriter) error {
	for _, node := range n.Nodes {
		if err := w.add(node.Name + ": " + strings.Join(node.Fanins, " ") + "\n"); err != nil {
			return err
		}
	}
	return nil
}
func writeSMV(n *IONetwork, w *ioTextWriter) error {
	if len(n.Latches) != 0 {
		return ErrUnsupported
	}
	if err := w.add("MODULE main\nVAR\n"); err != nil {
		return err
	}
	for _, s := range n.Inputs {
		if err := w.add("  " + s + " : boolean;\n"); err != nil {
			return err
		}
	}
	if err := w.add("DEFINE\n"); err != nil {
		return err
	}
	for _, node := range n.Nodes {
		expr, err := sopExpr(node, w.budget)
		if err != nil {
			return err
		}
		expr = strings.ReplaceAll(expr, "1'b0", "FALSE")
		expr = strings.ReplaceAll(expr, "1'b1", "TRUE")
		if err := w.add("  " + node.Name + " := " + expr + ";\n"); err != nil {
			return err
		}
	}
	return nil
}
func writeCNF(n *IONetwork, w *ioTextWriter) error {
	if len(n.Latches) > 0 {
		return ErrUnsupported
	}
	names := append(slices.Clone(n.Inputs), func() []string {
		r := make([]string, len(n.Nodes))
		for i := range n.Nodes {
			r[i] = n.Nodes[i].Name
		}
		return r
	}()...)
	ids := make(map[string]int, len(names))
	for i, s := range names {
		ids[s] = i + 1
	}
	clauseCount := 0
	for _, node := range n.Nodes {
		variableCount := 0
		for _, fanin := range node.Fanins {
			if fanin != "0" && fanin != "1" {
				variableCount++
			}
		}
		if variableCount >= strconv.IntSize {
			return ErrLimit
		}
		count := 1 << variableCount
		var err error
		clauseCount, err = checkedIOAdd(clauseCount, count)
		if err != nil {
			return err
		}
	}
	if clauseCount > w.limit/4 || clauseCount > w.budget.limits.MaxOperations {
		return ErrLimit
	}
	if err := w.add(fmt.Sprintf("p cnf %d %d\n", len(names), clauseCount)); err != nil {
		return err
	}
	for _, node := range n.Nodes {
		variableCount := 0
		for _, fanin := range node.Fanins {
			if fanin != "0" && fanin != "1" {
				variableCount++
			}
		}
		for mask := 0; mask < 1<<variableCount; mask++ {
			if err := w.budget.use(len(node.Fanins) + len(node.Cubes) + 1); err != nil {
				return err
			}
			vals := make([]bool, len(node.Fanins))
			variable := 0
			for i, fanin := range node.Fanins {
				if fanin == "0" || fanin == "1" {
					vals[i] = fanin == "1"
					continue
				}
				vals[i] = mask&(1<<variable) != 0
				variable++
			}
			want := false
			for _, c := range node.Cubes {
				if ioCubeMatches(c.Pattern, vals) {
					want = c.Value
				}
			}
			parts := make([]string, 0, variableCount+2)
			for i, f := range node.Fanins {
				if f == "0" || f == "1" {
					continue
				}
				lit := ids[f]
				if vals[i] {
					lit = -lit
				}
				parts = append(parts, strconv.Itoa(lit))
			}
			lit := ids[node.Name]
			if !want {
				lit = -lit
			}
			parts = append(parts, strconv.Itoa(lit), "0")
			if err := w.add(strings.Join(parts, " ") + "\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

// FormatIONetwork emits one deterministic owned representation.
func FormatIONetwork(n *IONetwork, format IOFormat, limits IOLimits) ([]byte, error) {
	l, err := normalizeIOLimits(limits)
	if err != nil {
		return nil, err
	}
	b := ioBudget{limits: l}
	if err = n.validate(l, &b); err != nil {
		return nil, err
	}
	w := ioTextWriter{limit: l.MaxOutputBytes, budget: &b}
	switch format {
	case IOFormatBLIF:
		err = writeBLIF(n, &w)
	case IOFormatBench:
		err = writeBench(n, &w)
	case IOFormatEquation:
		err = writeEquation(n, &w)
	case IOFormatPLA:
		err = writePLA(n, &w)
	case IOFormatVerilog:
		err = writeVerilog(n, &w)
	case IOFormatJSON:
		err = writeJSON(n, &w)
	case IOFormatDOT:
		err = writeGraph(n, &w, false)
	case IOFormatGML:
		err = writeGraph(n, &w, true)
	case IOFormatList:
		err = writeList(n, &w)
	case IOFormatSMV:
		err = writeSMV(n, &w)
	case IOFormatCNF:
		err = writeCNF(n, &w)
	default:
		err = ErrUnsupported
	}
	if err != nil {
		return nil, err
	}
	return []byte(w.b.String()), nil
}

// IOBaseBuildSources is the exact default-build inventory from module.make.
var IOBaseBuildSources = [...]string{
	"io.c", "ioJson.c", "ioReadAiger.c", "ioReadBaf.c", "ioReadBblif.c", "ioReadBench.c", "ioReadBlif.c", "ioReadBlifAig.c", "ioReadBlifMv.c", "ioReadDsd.c", "ioReadEdif.c", "ioReadEqn.c", "ioReadPla.c", "ioReadPlaMo.c", "ioReadVerilog.c", "ioUtil.c", "ioWriteAiger.c", "ioWriteBaf.c", "ioWriteBblif.c", "ioWriteBench.c", "ioWriteBlif.c", "ioWriteBlifMv.c", "ioWriteBook.c", "ioWriteCnf.c", "ioWriteDot.c", "ioWriteEqn.c", "ioWriteGml.c", "ioWriteList.c", "ioWritePla.c", "ioWriteVerilog.c", "ioWriteSmv.c",
}

// IOBaseCoverage records the intentional standalone boundary for source review.
var IOBaseCoverage = [...]string{
	"io.c: command registration and process-file dispatch remain outside the owned byte API",
	"ioJson.c: the distinct source generic nested token arena, deterministic pretty-printer, neural-layer extraction and in-memory read/write round trip are fully covered; its whitespace-delimited scalar grammar intentionally remains distinct from the IONetwork JSON schema",
	"ioRead/WriteBench.c: Boolean primitive subset is covered; LUT and sequential extensions remain partial",
	"ioRead/WriteBlif.c: Boolean names and latches are covered; hierarchy, timing, gates, boxes, mapped cells and LUT structures remain partial",
	"ioRead/WriteEqn.c: bounded Boolean expressions are covered; the complete legacy token grammar remains partial",
	"ioRead/WritePla.c and ioReadPlaMo.c: type-f multi-output covers and deterministic reduction-free output are covered; native DD preprocessing remains partial",
	"ioRead/WriteVerilog.c: bounded single-module structural assignments are covered; hierarchy, cells and sequential syntax remain partial",
	"ioWriteCnf/Dot/Gml/List/Smv.c: deterministic standalone projections are covered; native formatting and integration remain partial",
	"BAF reading and writing use the complete owned sequential AIG/name model and checked ABCNetwork bridges; BBLIF writing is supported directly from ABCNetwork through BblManagerFromABCNetwork; and DSD formulas map to owned SOP networks with the source preprocessing and truth order. The EDIF entry point is source-disabled and returns unsupported before reading input. The distinct IONetwork BBLIF path remains unsupported, as do BlifAig, BlifMv and Bookshelf pending their native object ecosystems",
}
