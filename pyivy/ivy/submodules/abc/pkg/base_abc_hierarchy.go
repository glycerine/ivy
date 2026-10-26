package abc

// Derived from ABC src/base/abc/{abcBlifMv.c,abcHie.c,abcHieCec.c,abcHieGia.c,abcHieNew.c}.
//
// This unit ports the owned relation, hierarchy validation/traversal, flattening,
// bounded truth comparison, and flat-GIA semantics. The source-only ABC pointer
// network rewrites, black-box mutation, HOP scratch fields, CBLIF file reader,
// ternary simulator, latch reconstruction, SAT command integration, and Au page
// allocator/ABI are intentionally outside this stand-alone representation.

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// ABCHierarchyOptions bounds retained hierarchy and deterministic work.
type ABCHierarchyOptions struct {
	MaxModules, MaxNodes, MaxInstances, MaxEdges int
	MaxDepth, MaxOperations, MaxTextBytes        int
	Stop                                         func() bool
}

func (o ABCHierarchyOptions) normalized() ABCHierarchyOptions {
	if o.MaxModules <= 0 {
		o.MaxModules = 1 << 12
	}
	if o.MaxNodes <= 0 {
		o.MaxNodes = 1 << 20
	}
	if o.MaxInstances <= 0 {
		o.MaxInstances = 1 << 18
	}
	if o.MaxEdges <= 0 {
		o.MaxEdges = 1 << 22
	}
	if o.MaxDepth <= 0 {
		o.MaxDepth = 256
	}
	if o.MaxOperations <= 0 {
		o.MaxOperations = 1 << 24
	}
	if o.MaxTextBytes <= 0 {
		o.MaxTextBytes = 1 << 24
	}
	return o
}

type abcHierarchyBudget struct {
	options    ABCHierarchyOptions
	operations int
}

func newABCHierarchyBudget(options ABCHierarchyOptions) (*abcHierarchyBudget, error) {
	if options.MaxModules < 0 || options.MaxNodes < 0 || options.MaxInstances < 0 || options.MaxEdges < 0 || options.MaxDepth < 0 || options.MaxOperations < 0 || options.MaxTextBytes < 0 {
		return nil, ErrDimension
	}
	return &abcHierarchyBudget{options: options.normalized()}, nil
}

func (b *abcHierarchyBudget) tick(n int) error {
	if n < 0 || b.operations > b.options.MaxOperations-n {
		return ErrLimit
	}
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	b.operations += n
	return nil
}

// ABCMVVariable is an owned multi-valued signal descriptor.
type ABCMVVariable struct {
	Values int
	Names  []string
}

func (v ABCMVVariable) Check() error {
	if v.Values < 2 || (len(v.Names) != 0 && len(v.Names) != v.Values) {
		return ErrDimension
	}
	for _, name := range v.Names {
		if name == "" || strings.IndexByte(name, 0) >= 0 {
			return ErrFormat
		}
	}
	return nil
}

// ABCMVCube uses -1 for don't-care and otherwise an exact input value.
type ABCMVCube struct {
	Inputs []int
	Output int
}

// ABCMVRelation is the checked stand-alone counterpart of a BLIF-MV node formula.
// DefaultInput >= 0 copies that input value. Otherwise DefaultOutput >= 0 is
// used; both fields being -1 represents the source format's absent default.
type ABCMVRelation struct {
	InputValues   []int
	OutputValues  int
	DefaultOutput int
	DefaultInput  int
	Cubes         []ABCMVCube
}

func (r ABCMVRelation) Clone() ABCMVRelation {
	r.InputValues = slices.Clone(r.InputValues)
	r.Cubes = slices.Clone(r.Cubes)
	for i := range r.Cubes {
		r.Cubes[i].Inputs = slices.Clone(r.Cubes[i].Inputs)
	}
	return r
}

func (r ABCMVRelation) Check(options ABCHierarchyOptions) error {
	b, err := newABCHierarchyBudget(options)
	if err != nil {
		return err
	}
	if err := b.tick(0); err != nil {
		return err
	}
	o := b.options
	if r.OutputValues < 2 || len(r.InputValues) > 63 || len(r.Cubes) > o.MaxNodes {
		return ErrDimension
	}
	for _, n := range r.InputValues {
		if n < 2 {
			return ErrDimension
		}
	}
	if r.DefaultInput >= len(r.InputValues) || r.DefaultInput < -1 {
		return ErrDimension
	}
	if r.DefaultInput >= 0 && r.InputValues[r.DefaultInput] != r.OutputValues {
		return ErrDimension
	}
	if r.DefaultInput < 0 && (r.DefaultOutput < -1 || r.DefaultOutput >= r.OutputValues) {
		return ErrDimension
	}
	for _, cube := range r.Cubes {
		if err := b.tick(len(cube.Inputs) + 1); err != nil {
			return err
		}
		if len(cube.Inputs) != len(r.InputValues) || cube.Output < 0 || cube.Output >= r.OutputValues {
			return ErrDimension
		}
		for i, value := range cube.Inputs {
			if value < -1 || value >= r.InputValues[i] {
				return ErrDimension
			}
		}
	}
	return nil
}

// ParseABCMVRelation parses deterministic source-shaped rows. An optional first
// row is "d N" or "d =I". Cube rows contain one value or '-' per input followed
// by the output value.
func ParseABCMVRelation(text string, inputValues []int, outputValues int, options ABCHierarchyOptions) (ABCMVRelation, error) {
	b, err := newABCHierarchyBudget(options)
	if err != nil {
		return ABCMVRelation{}, err
	}
	if len(text) > b.options.MaxTextBytes {
		return ABCMVRelation{}, ErrLimit
	}
	if len(inputValues) > 63 || outputValues < 2 {
		return ABCMVRelation{}, ErrDimension
	}
	if strings.IndexByte(text, 0) >= 0 {
		return ABCMVRelation{}, ErrFormat
	}
	r := ABCMVRelation{InputValues: slices.Clone(inputValues), OutputValues: outputValues, DefaultOutput: -1, DefaultInput: -1}
	lines := strings.Split(text, "\n")
	defaultSeen := false
	for _, raw := range lines {
		if err := b.tick(1); err != nil {
			return ABCMVRelation{}, err
		}
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if fields[0] == "d" {
			if len(fields) != 2 || len(r.Cubes) != 0 || defaultSeen {
				return ABCMVRelation{}, ErrFormat
			}
			defaultSeen = true
			if strings.HasPrefix(fields[1], "=") {
				value, e := strconv.Atoi(fields[1][1:])
				if e != nil {
					return ABCMVRelation{}, ErrFormat
				}
				r.DefaultInput = value
			} else {
				value, e := strconv.Atoi(fields[1])
				if e != nil {
					return ABCMVRelation{}, ErrFormat
				}
				r.DefaultOutput = value
			}
			continue
		}
		if len(fields) != len(inputValues)+1 {
			return ABCMVRelation{}, ErrFormat
		}
		cube := ABCMVCube{Inputs: make([]int, len(inputValues))}
		for i, field := range fields[:len(inputValues)] {
			if field == "-" {
				cube.Inputs[i] = -1
				continue
			}
			value, e := strconv.Atoi(field)
			if e != nil {
				return ABCMVRelation{}, ErrFormat
			}
			cube.Inputs[i] = value
		}
		cube.Output, err = strconv.Atoi(fields[len(inputValues)])
		if err != nil {
			return ABCMVRelation{}, ErrFormat
		}
		r.Cubes = append(r.Cubes, cube)
	}
	if err := r.Check(options); err != nil {
		return ABCMVRelation{}, err
	}
	return r, nil
}

func (r ABCMVRelation) Format(options ABCHierarchyOptions) (string, error) {
	if err := r.Check(options); err != nil {
		return "", err
	}
	var out strings.Builder
	if r.DefaultInput >= 0 {
		fmt.Fprintf(&out, "d =%d\n", r.DefaultInput)
	} else if r.DefaultOutput >= 0 {
		fmt.Fprintf(&out, "d %d\n", r.DefaultOutput)
	}
	for _, cube := range r.Cubes {
		for _, value := range cube.Inputs {
			if value < 0 {
				out.WriteString("- ")
			} else {
				fmt.Fprintf(&out, "%d ", value)
			}
		}
		fmt.Fprintf(&out, "%d\n", cube.Output)
	}
	if out.Len() > options.normalized().MaxTextBytes {
		return "", ErrLimit
	}
	return out.String(), nil
}

func (r ABCMVRelation) Evaluate(inputs []int) ([]bool, error) {
	if err := r.Check(ABCHierarchyOptions{}); err != nil {
		return nil, err
	}
	if len(inputs) != len(r.InputValues) {
		return nil, ErrDimension
	}
	for i, value := range inputs {
		if value < 0 || value >= r.InputValues[i] {
			return nil, ErrDimension
		}
	}
	result := make([]bool, r.OutputValues)
	matched := false
	for _, cube := range r.Cubes {
		good := true
		for i, value := range cube.Inputs {
			if value >= 0 && value != inputs[i] {
				good = false
				break
			}
		}
		if good {
			result[cube.Output] = true
			matched = true
		}
	}
	if !matched {
		value := r.DefaultOutput
		if r.DefaultInput >= 0 {
			value = inputs[r.DefaultInput]
		}
		if value < 0 {
			return result, nil
		}
		if value < 0 || value >= len(result) {
			return nil, ErrDimension
		}
		result[value] = true
	}
	return result, nil
}

func (r ABCMVRelation) Cost() (int, error) {
	if err := r.Check(ABCHierarchyOptions{}); err != nil {
		return 0, err
	}
	cost := 0
	for _, cube := range r.Cubes {
		cost++
		for _, value := range cube.Inputs {
			if value >= 0 {
				cost++
			}
		}
	}
	return cost, nil
}

type ABCHierarchyEdge struct {
	Node         int
	Complemented bool
}

func (e ABCHierarchyEdge) Not() ABCHierarchyEdge { e.Complemented = !e.Complemented; return e }

type ABCHierarchyNodeKind uint8

const (
	ABCHierarchyConstant ABCHierarchyNodeKind = iota
	ABCHierarchyInput
	ABCHierarchyGate
	ABCHierarchyInstanceOutput
)

type ABCHierarchyNode struct {
	Name           string
	Kind           ABCHierarchyNodeKind
	Fanins         []ABCHierarchyEdge
	SOP            ABCSOP
	Instance, Port int
}

func (n ABCHierarchyNode) clone() ABCHierarchyNode {
	n.Fanins = slices.Clone(n.Fanins)
	n.SOP = n.SOP.Clone()
	return n
}

type ABCHierarchyInstance struct {
	Name, Model string
	Inputs      []ABCHierarchyEdge
	Outputs     []int
}

func (i ABCHierarchyInstance) clone() ABCHierarchyInstance {
	i.Inputs = slices.Clone(i.Inputs)
	i.Outputs = slices.Clone(i.Outputs)
	return i
}

type ABCHierarchyModule struct {
	Name        string
	Nodes       []ABCHierarchyNode
	Inputs      []int
	Outputs     []ABCHierarchyEdge
	OutputNames []string
	Instances   []ABCHierarchyInstance
	options     ABCHierarchyOptions
}

func NewABCHierarchyModule(name string, inputs []string, options ABCHierarchyOptions) (*ABCHierarchyModule, error) {
	if name == "" || strings.IndexByte(name, 0) >= 0 {
		return nil, ErrFormat
	}
	o := options.normalized()
	if len(name) > o.MaxTextBytes {
		return nil, ErrLimit
	}
	if len(inputs)+1 > o.MaxNodes {
		return nil, ErrLimit
	}
	m := &ABCHierarchyModule{Name: name, options: o, Nodes: []ABCHierarchyNode{{Kind: ABCHierarchyConstant, Name: "$false"}}}
	seen := map[string]bool{"$false": true}
	for _, name := range inputs {
		if len(name) > o.MaxTextBytes {
			return nil, ErrLimit
		}
		if name == "" || seen[name] || strings.IndexByte(name, 0) >= 0 {
			return nil, ErrFormat
		}
		seen[name] = true
		m.Inputs = append(m.Inputs, len(m.Nodes))
		m.Nodes = append(m.Nodes, ABCHierarchyNode{Name: name, Kind: ABCHierarchyInput})
	}
	return m, nil
}

func (m *ABCHierarchyModule) Clone() *ABCHierarchyModule {
	if m == nil {
		return nil
	}
	c := *m
	c.Nodes = make([]ABCHierarchyNode, len(m.Nodes))
	for i := range m.Nodes {
		c.Nodes[i] = m.Nodes[i].clone()
	}
	c.Inputs = slices.Clone(m.Inputs)
	c.Outputs = slices.Clone(m.Outputs)
	c.OutputNames = slices.Clone(m.OutputNames)
	c.Instances = make([]ABCHierarchyInstance, len(m.Instances))
	for i := range m.Instances {
		c.Instances[i] = m.Instances[i].clone()
	}
	return &c
}

func (m *ABCHierarchyModule) validEdge(e ABCHierarchyEdge, before int) bool {
	return e.Node >= 0 && e.Node < before
}

func (m *ABCHierarchyModule) AddGate(name string, sop ABCSOP, fanins []ABCHierarchyEdge) (ABCHierarchyEdge, error) {
	if m == nil || name == "" || len(fanins) != sop.Variables {
		return ABCHierarchyEdge{}, ErrDimension
	}
	if len(m.Nodes) >= m.options.MaxNodes {
		return ABCHierarchyEdge{}, ErrLimit
	}
	if len(name) > m.options.MaxTextBytes || len(fanins) > m.options.MaxEdges-m.edgeCount() {
		return ABCHierarchyEdge{}, ErrLimit
	}
	if strings.IndexByte(name, 0) >= 0 {
		return ABCHierarchyEdge{}, ErrFormat
	}
	if err := sop.Check(ABCMetadataOptions{MaxVariables: 63}); err != nil {
		return ABCHierarchyEdge{}, err
	}
	for _, node := range m.Nodes {
		if node.Name == name {
			return ABCHierarchyEdge{}, ErrFormat
		}
	}
	for _, edge := range fanins {
		if !m.validEdge(edge, len(m.Nodes)) {
			return ABCHierarchyEdge{}, ErrDimension
		}
	}
	id := len(m.Nodes)
	m.Nodes = append(m.Nodes, ABCHierarchyNode{Name: name, Kind: ABCHierarchyGate, Fanins: slices.Clone(fanins), SOP: sop.Clone(), Instance: -1, Port: -1})
	return ABCHierarchyEdge{Node: id}, nil
}

func (m *ABCHierarchyModule) AddInstance(name, model string, inputs []ABCHierarchyEdge, outputs int) ([]ABCHierarchyEdge, error) {
	if m == nil || name == "" || model == "" || outputs < 1 {
		return nil, ErrDimension
	}
	if len(m.Instances) >= m.options.MaxInstances || len(m.Nodes) > m.options.MaxNodes-outputs {
		return nil, ErrLimit
	}
	if len(name) > m.options.MaxTextBytes || len(model) > m.options.MaxTextBytes || len(inputs) > m.options.MaxEdges-m.edgeCount() {
		return nil, ErrLimit
	}
	if strings.IndexByte(name, 0) >= 0 || strings.IndexByte(model, 0) >= 0 {
		return nil, ErrFormat
	}
	for _, inst := range m.Instances {
		if inst.Name == name {
			return nil, ErrFormat
		}
	}
	for port := 0; port < outputs; port++ {
		candidate := fmt.Sprintf("%s[%d]", name, port)
		for _, node := range m.Nodes {
			if node.Name == candidate {
				return nil, ErrFormat
			}
		}
	}
	for _, edge := range inputs {
		if !m.validEdge(edge, len(m.Nodes)) {
			return nil, ErrDimension
		}
	}
	index := len(m.Instances)
	inst := ABCHierarchyInstance{Name: name, Model: model, Inputs: slices.Clone(inputs)}
	result := make([]ABCHierarchyEdge, outputs)
	for port := 0; port < outputs; port++ {
		id := len(m.Nodes)
		inst.Outputs = append(inst.Outputs, id)
		m.Nodes = append(m.Nodes, ABCHierarchyNode{Name: fmt.Sprintf("%s[%d]", name, port), Kind: ABCHierarchyInstanceOutput, Instance: index, Port: port})
		result[port] = ABCHierarchyEdge{Node: id}
	}
	m.Instances = append(m.Instances, inst)
	return result, nil
}

func (m *ABCHierarchyModule) SetOutputs(names []string, outputs []ABCHierarchyEdge) error {
	if m == nil || len(names) != len(outputs) {
		return ErrDimension
	}
	seen := make(map[string]bool, len(names))
	for i, name := range names {
		if len(name) > m.options.MaxTextBytes {
			return ErrLimit
		}
		if name == "" || seen[name] || strings.IndexByte(name, 0) >= 0 || !m.validEdge(outputs[i], len(m.Nodes)) {
			return ErrFormat
		}
		seen[name] = true
	}
	m.OutputNames = slices.Clone(names)
	m.Outputs = slices.Clone(outputs)
	return nil
}

func (m *ABCHierarchyModule) edgeCount() int {
	count := 0
	for i := range m.Nodes {
		count += len(m.Nodes[i].Fanins)
	}
	for i := range m.Instances {
		count += len(m.Instances[i].Inputs)
	}
	return count
}

func (m *ABCHierarchyModule) Check() error {
	if m == nil || m.Name == "" || len(m.Nodes) == 0 || m.Nodes[0].Kind != ABCHierarchyConstant || len(m.Outputs) != len(m.OutputNames) {
		return ErrInvariant
	}
	if len(m.Name) > m.options.MaxTextBytes || strings.IndexByte(m.Name, 0) >= 0 {
		return ErrInvariant
	}
	if len(m.Nodes) > m.options.MaxNodes || len(m.Instances) > m.options.MaxInstances {
		return ErrLimit
	}
	if m.edgeCount() > m.options.MaxEdges {
		return ErrLimit
	}
	seenNames := make(map[string]bool, len(m.Nodes))
	seenInputs := make(map[int]bool, len(m.Inputs))
	seenOutputs := make(map[string]bool, len(m.OutputNames))
	seenInstances := make(map[string]bool, len(m.Instances))
	for _, id := range m.Inputs {
		if id <= 0 || id >= len(m.Nodes) || seenInputs[id] || m.Nodes[id].Kind != ABCHierarchyInput {
			return ErrInvariant
		}
		seenInputs[id] = true
	}
	for id, node := range m.Nodes {
		if node.Kind > ABCHierarchyInstanceOutput || (id != 0 && node.Name == "") || len(node.Name) > m.options.MaxTextBytes || strings.IndexByte(node.Name, 0) >= 0 {
			return ErrInvariant
		}
		if seenNames[node.Name] {
			return ErrInvariant
		}
		seenNames[node.Name] = true
		if node.Kind == ABCHierarchyInput && !seenInputs[id] {
			return ErrInvariant
		}
		if node.Kind == ABCHierarchyConstant && (id != 0 || node.Name != "$false" || len(node.Fanins) != 0) {
			return ErrInvariant
		}
		if node.Kind != ABCHierarchyGate && len(node.Fanins) != 0 {
			return ErrInvariant
		}
		for _, edge := range node.Fanins {
			if !m.validEdge(edge, id) {
				return ErrInvariant
			}
		}
		if node.Kind == ABCHierarchyGate {
			if len(node.Fanins) != node.SOP.Variables || node.SOP.Check(ABCMetadataOptions{MaxVariables: 63}) != nil {
				return ErrInvariant
			}
		}
		if node.Kind == ABCHierarchyInstanceOutput {
			if node.Instance < 0 || node.Instance >= len(m.Instances) || node.Port < 0 || node.Port >= len(m.Instances[node.Instance].Outputs) || m.Instances[node.Instance].Outputs[node.Port] != id {
				return ErrInvariant
			}
		}
	}
	for index, inst := range m.Instances {
		if inst.Name == "" || inst.Model == "" || len(inst.Outputs) == 0 || seenInstances[inst.Name] || len(inst.Name) > m.options.MaxTextBytes || len(inst.Model) > m.options.MaxTextBytes || strings.IndexByte(inst.Name, 0) >= 0 || strings.IndexByte(inst.Model, 0) >= 0 {
			return ErrInvariant
		}
		seenInstances[inst.Name] = true
		firstOutput := inst.Outputs[0]
		for _, edge := range inst.Inputs {
			if !m.validEdge(edge, inst.Outputs[0]) {
				return ErrInvariant
			}
		}
		for port, id := range inst.Outputs {
			if id != firstOutput+port || id <= 0 || id >= len(m.Nodes) || m.Nodes[id].Kind != ABCHierarchyInstanceOutput || m.Nodes[id].Instance != index || m.Nodes[id].Port != port {
				return ErrInvariant
			}
		}
	}
	for _, output := range m.Outputs {
		if !m.validEdge(output, len(m.Nodes)) {
			return ErrInvariant
		}
	}
	for _, name := range m.OutputNames {
		if name == "" || seenOutputs[name] || len(name) > m.options.MaxTextBytes || strings.IndexByte(name, 0) >= 0 {
			return ErrInvariant
		}
		seenOutputs[name] = true
	}
	return nil
}

type ABCHierarchyDesign struct {
	Name, Root string
	Modules    []ABCHierarchyModule
	options    ABCHierarchyOptions
	byName     map[string]int
}

func NewABCHierarchyDesign(name string, options ABCHierarchyOptions) (*ABCHierarchyDesign, error) {
	if name == "" || strings.IndexByte(name, 0) >= 0 {
		return nil, ErrFormat
	}
	o := options.normalized()
	if len(name) > o.MaxTextBytes {
		return nil, ErrLimit
	}
	return &ABCHierarchyDesign{Name: name, options: o, byName: make(map[string]int)}, nil
}

func (d *ABCHierarchyDesign) AddModule(module *ABCHierarchyModule, root bool) error {
	if d == nil || module == nil {
		return ErrInvariant
	}
	if d.byName == nil {
		return ErrInvariant
	}
	if len(d.Modules) >= d.options.MaxModules {
		return ErrLimit
	}
	candidate := module.Clone()
	candidate.options = d.options
	if err := candidate.Check(); err != nil {
		return err
	}
	if _, exists := d.byName[module.Name]; exists {
		return ErrFormat
	}
	if root && d.Root != "" {
		return ErrFormat
	}
	nodes, instances, edges := len(candidate.Nodes), len(candidate.Instances), candidate.edgeCount()
	for i := range d.Modules {
		if nodes > d.options.MaxNodes-len(d.Modules[i].Nodes) || instances > d.options.MaxInstances-len(d.Modules[i].Instances) || edges > d.options.MaxEdges-d.Modules[i].edgeCount() {
			return ErrLimit
		}
		nodes += len(d.Modules[i].Nodes)
		instances += len(d.Modules[i].Instances)
		edges += d.Modules[i].edgeCount()
	}
	d.byName[module.Name] = len(d.Modules)
	d.Modules = append(d.Modules, *candidate)
	if root {
		d.Root = module.Name
	}
	return nil
}

func (d *ABCHierarchyDesign) module(name string) (*ABCHierarchyModule, bool) {
	index, ok := d.byName[name]
	if !ok {
		return nil, false
	}
	return &d.Modules[index], true
}

func (d *ABCHierarchyDesign) Validate() error {
	if d == nil || d.Name == "" || d.Root == "" || len(d.Modules) == 0 {
		return ErrInvariant
	}
	budget, err := newABCHierarchyBudget(d.options)
	if err != nil {
		return err
	}
	if len(d.byName) != len(d.Modules) {
		return ErrInvariant
	}
	for i := range d.Modules {
		index, ok := d.byName[d.Modules[i].Name]
		if !ok || index != i {
			return ErrInvariant
		}
	}
	if _, ok := d.module(d.Root); !ok {
		return ErrInvariant
	}
	totalNodes, totalInstances, totalEdges := 0, 0, 0
	for i := range d.Modules {
		if err := budget.tick(1); err != nil {
			return err
		}
		m := &d.Modules[i]
		if err := m.Check(); err != nil {
			return err
		}
		if err := budget.tick(len(m.Nodes) + len(m.Instances) + m.edgeCount()); err != nil {
			return err
		}
		totalNodes += len(m.Nodes)
		totalInstances += len(m.Instances)
		for _, node := range m.Nodes {
			totalEdges += len(node.Fanins)
		}
		for _, inst := range m.Instances {
			totalEdges += len(inst.Inputs)
			child, ok := d.module(inst.Model)
			if !ok || len(inst.Inputs) != len(child.Inputs) || len(inst.Outputs) != len(child.Outputs) {
				return ErrInvariant
			}
		}
	}
	if totalNodes > d.options.MaxNodes || totalInstances > d.options.MaxInstances || totalEdges > d.options.MaxEdges {
		return ErrLimit
	}
	state := make([]uint8, len(d.Modules))
	var visit func(int, int) error
	visit = func(index, depth int) error {
		if err := budget.tick(1); err != nil {
			return err
		}
		if depth > d.options.MaxDepth {
			return ErrLimit
		}
		if state[index] == 1 {
			return ErrInvariant
		}
		if state[index] == 2 {
			return nil
		}
		state[index] = 1
		for _, inst := range d.Modules[index].Instances {
			next := d.byName[inst.Model]
			if err := visit(next, depth+1); err != nil {
				return err
			}
		}
		state[index] = 2
		return nil
	}
	if err := visit(d.byName[d.Root], 0); err != nil {
		return err
	}
	for index := range d.Modules {
		if err := visit(index, 0); err != nil {
			return err
		}
	}
	return nil
}

// ABCHierarchyFlattenMap records deterministic hierarchical names for flat nodes.
type ABCHierarchyFlattenMap struct {
	FlatNames  []string
	Instances  int
	ModuleUses map[string]int
}

// ABCHierarchyCECResult reports an exact bounded truth comparison. A
// counterexample is ordered like the root module's inputs.
type ABCHierarchyCECResult struct {
	Equivalent     bool
	Counterexample []bool
}

// CompareTruth is a small, deterministic CEC cone for independently flattened
// designs. It is exact up to maxInputs and returns ErrLimit instead of silently
// approximating larger cones.
func (d *ABCHierarchyDesign) CompareTruth(other *ABCHierarchyDesign, maxInputs int) (ABCHierarchyCECResult, error) {
	var result ABCHierarchyCECResult
	if d == nil || other == nil || maxInputs < 0 || maxInputs > 20 {
		return result, ErrDimension
	}
	left, _, err := d.Flatten()
	if err != nil {
		return result, err
	}
	right, _, err := other.Flatten()
	if err != nil {
		return result, err
	}
	if len(left.Inputs) != len(right.Inputs) || len(left.Outputs) != len(right.Outputs) {
		return result, ErrDimension
	}
	if len(left.Inputs) > maxInputs {
		return result, ErrLimit
	}
	budget, err := newABCHierarchyBudget(d.options)
	if err != nil {
		return result, err
	}
	assignments := 1 << uint(len(left.Inputs))
	for assignment := 0; assignment < assignments; assignment++ {
		if err := budget.tick(len(left.Nodes) + len(right.Nodes)); err != nil {
			return result, err
		}
		inputs := make([]bool, len(left.Inputs))
		for index := range inputs {
			inputs[index] = assignment&(1<<uint(index)) != 0
		}
		leftOutputs, err := left.Evaluate(inputs)
		if err != nil {
			return result, err
		}
		rightOutputs, err := right.Evaluate(inputs)
		if err != nil {
			return result, err
		}
		if !slices.Equal(leftOutputs, rightOutputs) {
			result.Counterexample = inputs
			return result, nil
		}
	}
	result.Equivalent = true
	return result, nil
}

// Format returns a deterministic diagnostic serialization. Module, terminal,
// instance, fanin, cube, and output order are retained exactly; lookup maps are
// never enumerated.
func (d *ABCHierarchyDesign) Format(options ABCHierarchyOptions) (string, error) {
	if err := d.Validate(); err != nil {
		return "", err
	}
	budget, err := newABCHierarchyBudget(options)
	if err != nil {
		return "", err
	}
	o := budget.options
	var out strings.Builder
	fmt.Fprintf(&out, "design %q root %q\n", d.Name, d.Root)
	for moduleIndex := range d.Modules {
		module := &d.Modules[moduleIndex]
		if err := budget.tick(len(module.Nodes) + len(module.Instances) + len(module.Outputs)); err != nil {
			return "", err
		}
		fmt.Fprintf(&out, "module %d %q\n", moduleIndex, module.Name)
		for port, id := range module.Inputs {
			fmt.Fprintf(&out, " input %d node %d %q\n", port, id, module.Nodes[id].Name)
		}
		for id, node := range module.Nodes {
			if node.Kind != ABCHierarchyGate {
				continue
			}
			fmt.Fprintf(&out, " gate %d %q phase %c", id, node.Name, node.SOP.Phase)
			for _, edge := range node.Fanins {
				fmt.Fprintf(&out, " %d:%t", edge.Node, edge.Complemented)
			}
			out.WriteByte('\n')
			for _, cube := range node.SOP.Cubes {
				fmt.Fprintf(&out, "  cube %q\n", cube)
			}
		}
		for instanceIndex, instance := range module.Instances {
			fmt.Fprintf(&out, " instance %d %q model %q", instanceIndex, instance.Name, instance.Model)
			for _, edge := range instance.Inputs {
				fmt.Fprintf(&out, " in=%d:%t", edge.Node, edge.Complemented)
			}
			for _, id := range instance.Outputs {
				fmt.Fprintf(&out, " out=%d", id)
			}
			out.WriteByte('\n')
		}
		for port, edge := range module.Outputs {
			fmt.Fprintf(&out, " output %d %q node %d:%t\n", port, module.OutputNames[port], edge.Node, edge.Complemented)
		}
		if out.Len() > o.MaxTextBytes {
			return "", ErrLimit
		}
	}
	return out.String(), nil
}

func (d *ABCHierarchyDesign) Flatten() (*ABCHierarchyModule, ABCHierarchyFlattenMap, error) {
	var info ABCHierarchyFlattenMap
	if err := d.Validate(); err != nil {
		return nil, info, err
	}
	budget, err := newABCHierarchyBudget(d.options)
	if err != nil {
		return nil, info, err
	}
	root, _ := d.module(d.Root)
	flat, err := NewABCHierarchyModule(root.Name+"$flat", nil, d.options)
	if err != nil {
		return nil, info, err
	}
	info.ModuleUses = make(map[string]int)
	bindings := make([]ABCHierarchyEdge, len(root.Inputs))
	for i, id := range root.Inputs {
		if len(flat.Nodes) >= flat.options.MaxNodes {
			return nil, info, ErrLimit
		}
		nodeID := len(flat.Nodes)
		flat.Nodes = append(flat.Nodes, ABCHierarchyNode{Name: root.Nodes[id].Name, Kind: ABCHierarchyInput, Instance: -1, Port: -1})
		flat.Inputs = append(flat.Inputs, nodeID)
		bindings[i] = ABCHierarchyEdge{Node: nodeID}
	}
	var instantiate func(*ABCHierarchyModule, []ABCHierarchyEdge, string, int) ([]ABCHierarchyEdge, error)
	instantiate = func(module *ABCHierarchyModule, inputs []ABCHierarchyEdge, prefix string, depth int) ([]ABCHierarchyEdge, error) {
		if depth > d.options.MaxDepth || len(inputs) != len(module.Inputs) {
			return nil, ErrLimit
		}
		if err := budget.tick(len(module.Nodes) + len(module.Instances)); err != nil {
			return nil, err
		}
		info.ModuleUses[module.Name]++
		mapping := make([]ABCHierarchyEdge, len(module.Nodes))
		mapping[0] = ABCHierarchyEdge{Node: 0}
		for i, id := range module.Inputs {
			mapping[id] = inputs[i]
		}
		processed := make([]bool, len(module.Instances))
		for id := 1; id < len(module.Nodes); id++ {
			node := module.Nodes[id]
			switch node.Kind {
			case ABCHierarchyInput:
				continue
			case ABCHierarchyGate:
				if len(prefix) > d.options.MaxTextBytes-len(node.Name) {
					return nil, ErrLimit
				}
				flatName := prefix + node.Name
				fanins := make([]ABCHierarchyEdge, len(node.Fanins))
				for i, e := range node.Fanins {
					fanins[i] = mapping[e.Node]
					if e.Complemented {
						fanins[i] = fanins[i].Not()
					}
				}
				edge, e := flat.AddGate(flatName, node.SOP, fanins)
				if e != nil {
					return nil, e
				}
				mapping[id] = edge
				info.FlatNames = append(info.FlatNames, flatName)
			case ABCHierarchyInstanceOutput:
				if processed[node.Instance] {
					continue
				}
				processed[node.Instance] = true
				inst := module.Instances[node.Instance]
				if len(prefix) > d.options.MaxTextBytes-len(inst.Name)-1 {
					return nil, ErrLimit
				}
				childPrefix := prefix + inst.Name + "/"
				child, _ := d.module(inst.Model)
				childInputs := make([]ABCHierarchyEdge, len(inst.Inputs))
				for i, e := range inst.Inputs {
					childInputs[i] = mapping[e.Node]
					if e.Complemented {
						childInputs[i] = childInputs[i].Not()
					}
				}
				childOutputs, e := instantiate(child, childInputs, childPrefix, depth+1)
				if e != nil {
					return nil, e
				}
				for i, outID := range inst.Outputs {
					mapping[outID] = childOutputs[i]
				}
				info.Instances++
			default:
				return nil, ErrInvariant
			}
		}
		outputs := make([]ABCHierarchyEdge, len(module.Outputs))
		for i, e := range module.Outputs {
			outputs[i] = mapping[e.Node]
			if e.Complemented {
				outputs[i] = outputs[i].Not()
			}
		}
		return outputs, nil
	}
	outputs, err := instantiate(root, bindings, "", 0)
	if err != nil {
		return nil, info, err
	}
	if err = flat.SetOutputs(root.OutputNames, outputs); err != nil {
		return nil, info, err
	}
	return flat, info, nil
}

func (m *ABCHierarchyModule) Evaluate(inputs []bool) ([]bool, error) {
	if err := m.Check(); err != nil {
		return nil, err
	}
	if len(inputs) != len(m.Inputs) {
		return nil, ErrDimension
	}
	values := make([]bool, len(m.Nodes))
	for i, id := range m.Inputs {
		values[id] = inputs[i]
	}
	for id, node := range m.Nodes {
		if node.Kind != ABCHierarchyGate {
			if node.Kind == ABCHierarchyInstanceOutput {
				return nil, ErrInvariant
			}
			continue
		}
		value, err := evaluateHierarchySOP(node.SOP, node.Fanins, values)
		if err != nil {
			return nil, err
		}
		values[id] = value
	}
	outputs := make([]bool, len(m.Outputs))
	for i, e := range m.Outputs {
		outputs[i] = values[e.Node] != e.Complemented
	}
	return outputs, nil
}

// evaluateHierarchySOP preserves ABC's special XOR/XNOR marker semantics:
// phases x/n mean parity of all fanins and do not describe an ordinary cover.
func evaluateHierarchySOP(sop ABCSOP, fanins []ABCHierarchyEdge, values []bool) (bool, error) {
	if sop.IsXORType() {
		value := false
		for _, edge := range fanins {
			value = value != (values[edge.Node] != edge.Complemented)
		}
		return value != (sop.Phase == 'n'), nil
	}
	var assignment uint64
	for i, edge := range fanins {
		if values[edge.Node] != edge.Complemented {
			assignment |= uint64(1) << uint(i)
		}
	}
	return sop.Evaluate(assignment)
}

func giaHierarchyAnd(manager *GiaManager, a, b GiaLit) (GiaLit, error) {
	if a == GiaConst0 || b == GiaConst0 {
		return GiaConst0, nil
	}
	if a == GiaConst1 {
		return b, nil
	}
	if b == GiaConst1 {
		return a, nil
	}
	if a == b {
		return a, nil
	}
	if a == b.Not() {
		return GiaConst0, nil
	}
	return manager.AppendAnd(a, b)
}
func giaHierarchyOr(manager *GiaManager, a, b GiaLit) (GiaLit, error) {
	v, e := giaHierarchyAnd(manager, a.Not(), b.Not())
	if e != nil {
		return InvalidGiaLit, e
	}
	return v.Not(), nil
}

func giaHierarchyXor(manager *GiaManager, a, b GiaLit) (GiaLit, error) {
	left, err := giaHierarchyAnd(manager, a, b.Not())
	if err != nil {
		return InvalidGiaLit, err
	}
	right, err := giaHierarchyAnd(manager, a.Not(), b)
	if err != nil {
		return InvalidGiaLit, err
	}
	return giaHierarchyOr(manager, left, right)
}

func (d *ABCHierarchyDesign) ToGia() (*GiaManager, error) {
	flat, _, err := d.Flatten()
	if err != nil {
		return nil, err
	}
	limit := d.options.MaxNodes
	if limit > int(MaxGiaID)/8 {
		limit = int(MaxGiaID) / 8
	}
	limit = limit*8 + len(flat.Outputs) + 1
	manager, err := NewGiaManagerWithLimit(limit)
	if err != nil {
		return nil, err
	}
	manager.Name = d.Name
	manager.Spec = d.Root
	values := make([]GiaLit, len(flat.Nodes))
	values[0] = GiaConst0
	for _, id := range flat.Inputs {
		values[id], err = manager.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	for id, node := range flat.Nodes {
		if node.Kind != ABCHierarchyGate {
			continue
		}
		result := GiaConst0
		if node.SOP.IsXORType() {
			for _, fanin := range node.Fanins {
				result, err = giaHierarchyXor(manager, result, values[fanin.Node].NotIf(fanin.Complemented))
				if err != nil {
					return nil, err
				}
			}
			result = result.NotIf(node.SOP.Phase == 'n')
		} else if node.SOP.IsConst1() {
			result = GiaConst1
		} else if !node.SOP.IsConst0() {
			for _, cube := range node.SOP.Cubes {
				product := GiaConst1
				for i, c := range []byte(cube) {
					if c == '-' {
						continue
					}
					lit := values[node.Fanins[i].Node].NotIf(node.Fanins[i].Complemented).NotIf(c == '0')
					product, err = giaHierarchyAnd(manager, product, lit)
					if err != nil {
						return nil, err
					}
				}
				result, err = giaHierarchyOr(manager, result, product)
				if err != nil {
					return nil, err
				}
			}
			result = result.NotIf(node.SOP.Phase == '0')
		}
		values[id] = result
	}
	for _, out := range flat.Outputs {
		if _, err = manager.AppendCO(values[out.Node].NotIf(out.Complemented)); err != nil {
			return nil, err
		}
	}
	return manager, nil
}

func (d *ABCHierarchyDesign) Statistics() (modules, instances, nodes int, err error) {
	if err = d.Validate(); err != nil {
		return
	}
	modules = len(d.Modules)
	for i := range d.Modules {
		instances += len(d.Modules[i].Instances)
		nodes += len(d.Modules[i].Nodes)
	}
	return
}
