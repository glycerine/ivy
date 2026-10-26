package abc

// Ported from src/base/acb/acb.h, acbAbc.c, acbCom.c, acbFunc.c,
// acbMfs.c, acbPar.h, acbPush.c, acbSets.h, acbTest.c, acbUtil.c,
// and module.make.

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// ACBObjectType is the owned counterpart of the source object-kind field.
type ACBObjectType uint8

const (
	ACBObjectNone ACBObjectType = iota
	ACBObjectInput
	ACBObjectOutput
	ACBObjectNode
	ACBObjectBox
)

// ACBRange preserves an HDL range and its signedness.
type ACBRange struct {
	Left   int
	Right  int
	Signed bool
}

func (r ACBRange) Width() (int, error) {
	var d uint64
	if r.Left >= r.Right {
		d = uint64(r.Left) - uint64(r.Right)
	} else {
		d = uint64(r.Right) - uint64(r.Left)
	}
	if d >= uint64(^uint(0)>>1) {
		return 0, ErrLimit
	}
	return int(d + 1), nil
}

// ACBAttribute is ordered so snapshots and reports are deterministic.
type ACBAttribute struct {
	Key   int
	Value int
}

// ACBObject is a caller-owned snapshot when returned by this package.
type ACBObject struct {
	ID         int
	Type       ACBObjectType
	Name       string
	Fanins     []int
	Truth      uint64
	Range      ACBRange
	Attributes []ACBAttribute
	Weight     int
	Sequential bool
}

// ACBLimits bounds aggregate work and allocations.
type ACBLimits struct {
	MaxObjects    int
	MaxEdges      int
	MaxOperations int
	MaxBytes      int
	MaxDepth      int
	MaxText       int
}

func (l ACBLimits) normalized() ACBLimits {
	if l.MaxObjects <= 0 {
		l.MaxObjects = 1 << 16
	}
	if l.MaxEdges <= 0 {
		l.MaxEdges = 1 << 20
	}
	if l.MaxOperations <= 0 {
		l.MaxOperations = 1 << 24
	}
	if l.MaxBytes <= 0 {
		l.MaxBytes = 64 << 20
	}
	if l.MaxDepth <= 0 {
		l.MaxDepth = 1 << 14
	}
	if l.MaxText <= 0 {
		l.MaxText = 8 << 20
	}
	return l
}

type acbBudget struct {
	ctx context.Context
	lim ACBLimits
	ops int
	mem int
}

func newACBBudget(ctx context.Context, limits ACBLimits) *acbBudget {
	if ctx == nil {
		ctx = context.Background()
	}
	return &acbBudget{ctx: ctx, lim: limits.normalized()}
}

func (b *acbBudget) use(ops, bytes int) error {
	if ops < 0 || bytes < 0 || b.ops > b.lim.MaxOperations-ops || b.mem > b.lim.MaxBytes-bytes {
		return ErrLimit
	}
	b.ops += ops
	b.mem += bytes
	select {
	case <-b.ctx.Done():
		return ErrCancelled
	default:
		return nil
	}
}

// ACBNetwork owns an acyclic, source-ordered logic network. Index zero is a
// permanent sentinel, matching the source layout.
type ACBNetwork struct {
	name    string
	objects []ACBObject
	inputs  []int
	outputs []int
	regs    int
	closed  bool
	limits  ACBLimits
}

func NewACBNetwork(name string, limits ACBLimits) (*ACBNetwork, error) {
	limits = limits.normalized()
	if name == "" || len(name) > limits.MaxText || limits.MaxObjects < 1 {
		return nil, ErrInvariant
	}
	n := &ACBNetwork{name: name, limits: limits, objects: []ACBObject{{}}}
	return n, nil
}

func (n *ACBNetwork) Name() string { return n.name }

func cloneACBObject(o ACBObject) ACBObject {
	r := o
	r.Fanins = append([]int(nil), o.Fanins...)
	r.Attributes = append([]ACBAttribute(nil), o.Attributes...)
	return r
}

func (n *ACBNetwork) Snapshot() ([]ACBObject, error) {
	if n == nil || n.closed {
		return nil, ErrInvariant
	}
	r := make([]ACBObject, len(n.objects)-1)
	for i := 1; i < len(n.objects); i++ {
		r[i-1] = cloneACBObject(n.objects[i])
	}
	return r, nil
}

func (n *ACBNetwork) Close() {
	if n == nil {
		return
	}
	n.closed = true
	n.objects = nil
	n.inputs = nil
	n.outputs = nil
}

func acbTruthMask(vars int) uint64 {
	if vars >= 6 {
		return ^uint64(0)
	}
	return uint64(1)<<(1<<vars) - 1
}

func acbValidateAttributes(attrs []ACBAttribute) error {
	last := -1
	for i, a := range attrs {
		if a.Key < 0 || (i > 0 && a.Key <= last) {
			return ErrInvariant
		}
		last = a.Key
	}
	return nil
}

func (n *ACBNetwork) addObject(ctx context.Context, o ACBObject) (int, error) {
	if n == nil || n.closed {
		return 0, ErrInvariant
	}
	b := newACBBudget(ctx, n.limits)
	if len(n.objects) >= n.limits.MaxObjects || len(o.Fanins) > 6 {
		return 0, ErrLimit
	}
	bytes := 96 + 8*len(o.Fanins) + 16*len(o.Attributes) + len(o.Name)
	for i := 1; i < len(n.objects); i++ {
		bytes += 96 + 8*len(n.objects[i].Fanins) + 16*len(n.objects[i].Attributes) + len(n.objects[i].Name)
	}
	if err := b.use(1+len(o.Fanins), bytes); err != nil {
		return 0, err
	}
	if o.Type < ACBObjectInput || o.Type > ACBObjectBox || acbValidateAttributes(o.Attributes) != nil {
		return 0, ErrInvariant
	}
	if _, err := o.Range.Width(); err != nil {
		return 0, err
	}
	if o.Type == ACBObjectInput && len(o.Fanins) != 0 || o.Type == ACBObjectOutput && len(o.Fanins) != 1 {
		return 0, ErrInvariant
	}
	if o.Type == ACBObjectNode && len(o.Fanins) == 0 {
		o.Truth &= 1
	} else if o.Type == ACBObjectNode {
		o.Truth &= acbTruthMask(len(o.Fanins))
	}
	for _, f := range o.Fanins {
		if f <= 0 || f >= len(n.objects) || n.objects[f].Type == ACBObjectOutput {
			return 0, ErrInvariant
		}
	}
	if o.Name != "" {
		for i := 1; i < len(n.objects); i++ {
			if n.objects[i].Name == o.Name {
				return 0, ErrInvariant
			}
		}
	}
	edges := len(o.Fanins)
	for i := 1; i < len(n.objects); i++ {
		edges += len(n.objects[i].Fanins)
	}
	if edges > n.limits.MaxEdges {
		return 0, ErrLimit
	}
	o.ID = len(n.objects)
	o.Fanins = append([]int(nil), o.Fanins...)
	o.Attributes = append([]ACBAttribute(nil), o.Attributes...)
	n.objects = append(n.objects, o)
	if o.Type == ACBObjectInput {
		n.inputs = append(n.inputs, o.ID)
	}
	if o.Type == ACBObjectOutput {
		n.outputs = append(n.outputs, o.ID)
	}
	return o.ID, nil
}

func (n *ACBNetwork) AddInput(ctx context.Context, name string, r ACBRange) (int, error) {
	return n.addObject(ctx, ACBObject{Type: ACBObjectInput, Name: name, Range: r})
}

func (n *ACBNetwork) AddNode(ctx context.Context, name string, fanins []int, truth uint64) (int, error) {
	return n.addObject(ctx, ACBObject{Type: ACBObjectNode, Name: name, Fanins: fanins, Truth: truth})
}

func (n *ACBNetwork) AddOutput(ctx context.Context, name string, driver int) (int, error) {
	return n.addObject(ctx, ACBObject{Type: ACBObjectOutput, Name: name, Fanins: []int{driver}})
}

func (n *ACBNetwork) SetRegisters(count int) error {
	if n == nil || n.closed || count < 0 || count > len(n.inputs) || count > len(n.outputs) {
		return ErrInvariant
	}
	n.regs = count
	return nil
}

func (n *ACBNetwork) Object(id int) (ACBObject, error) {
	if n == nil || n.closed || id <= 0 || id >= len(n.objects) {
		return ACBObject{}, ErrInvariant
	}
	return cloneACBObject(n.objects[id]), nil
}

func (n *ACBNetwork) Inputs() []int  { return append([]int(nil), n.inputs...) }
func (n *ACBNetwork) Outputs() []int { return append([]int(nil), n.outputs...) }

func (n *ACBNetwork) validateBudget(b *acbBudget) error {
	if n == nil || n.closed || len(n.objects) == 0 || n.objects[0].Type != ACBObjectNone {
		return ErrInvariant
	}
	edges := 0
	names := make(map[string]struct{}, len(n.objects))
	inputs, outputs := 0, 0
	for i := 1; i < len(n.objects); i++ {
		if err := b.use(1+len(n.objects[i].Fanins), 0); err != nil {
			return err
		}
		o := &n.objects[i]
		if o.ID != i || o.Type < ACBObjectInput || o.Type > ACBObjectBox || len(o.Fanins) > 6 {
			return ErrInvariant
		}
		if _, err := o.Range.Width(); err != nil {
			return err
		}
		if o.Type == ACBObjectInput {
			inputs++
			if len(o.Fanins) != 0 {
				return ErrInvariant
			}
		}
		if o.Type == ACBObjectOutput {
			outputs++
			if len(o.Fanins) != 1 {
				return ErrInvariant
			}
		}
		for _, f := range o.Fanins {
			if f <= 0 || f >= i || n.objects[f].Type == ACBObjectOutput {
				return ErrInvariant
			}
		}
		edges += len(o.Fanins)
		if o.Name != "" {
			if _, ok := names[o.Name]; ok {
				return ErrInvariant
			}
			names[o.Name] = struct{}{}
		}
		if err := acbValidateAttributes(o.Attributes); err != nil {
			return err
		}
	}
	if len(n.objects) > n.limits.MaxObjects || edges > n.limits.MaxEdges || inputs != len(n.inputs) || outputs != len(n.outputs) || n.regs > inputs || n.regs > outputs {
		return ErrInvariant
	}
	for i, id := range n.inputs {
		if id <= 0 || id >= len(n.objects) || n.objects[id].Type != ACBObjectInput || i > 0 && id <= n.inputs[i-1] {
			return ErrInvariant
		}
	}
	for i, id := range n.outputs {
		if id <= 0 || id >= len(n.objects) || n.objects[id].Type != ACBObjectOutput || i > 0 && id <= n.outputs[i-1] {
			return ErrInvariant
		}
	}
	return nil
}

func (n *ACBNetwork) Validate(ctx context.Context) error {
	if n == nil {
		return ErrInvariant
	}
	return n.validateBudget(newACBBudget(ctx, n.limits))
}

// Evaluate returns output values in interface order.
func (n *ACBNetwork) Evaluate(ctx context.Context, inputValues []bool) ([]bool, error) {
	if n == nil {
		return nil, ErrInvariant
	}
	b := newACBBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	if len(inputValues) != len(n.inputs) {
		return nil, ErrDimension
	}
	values := make([]bool, len(n.objects))
	if err := b.use(0, len(values)); err != nil {
		return nil, err
	}
	for i, id := range n.inputs {
		values[id] = inputValues[i]
	}
	for i := 1; i < len(n.objects); i++ {
		o := &n.objects[i]
		if err := b.use(1+len(o.Fanins), 0); err != nil {
			return nil, err
		}
		switch o.Type {
		case ACBObjectNode:
			index := 0
			for k, f := range o.Fanins {
				if values[f] {
					index |= 1 << k
				}
			}
			values[i] = (o.Truth>>index)&1 != 0
		case ACBObjectOutput, ACBObjectBox:
			if len(o.Fanins) > 0 {
				values[i] = values[o.Fanins[0]]
			}
		}
	}
	result := make([]bool, len(n.outputs))
	for i, id := range n.outputs {
		result[i] = values[id]
	}
	return result, nil
}

// ACBProfile contains source-shaped forward/reverse levels and path counts.
type ACBProfile struct {
	Forward []int
	Reverse []int
	Paths   []uint64
	Maximum int
}

func (n *ACBNetwork) Profile(ctx context.Context) (ACBProfile, error) {
	if n == nil {
		return ACBProfile{}, ErrInvariant
	}
	b := newACBBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return ACBProfile{}, err
	}
	p := ACBProfile{Forward: make([]int, len(n.objects)), Reverse: make([]int, len(n.objects)), Paths: make([]uint64, len(n.objects))}
	if err := b.use(0, len(n.objects)*24); err != nil {
		return ACBProfile{}, err
	}
	maximumAny := 0
	for i := 1; i < len(n.objects); i++ {
		increment := 1
		if n.objects[i].Type == ACBObjectInput || n.objects[i].Type == ACBObjectOutput {
			increment = 0
		}
		for _, f := range n.objects[i].Fanins {
			if p.Forward[i] < p.Forward[f]+increment {
				p.Forward[i] = p.Forward[f] + increment
			}
		}
		if n.objects[i].Type == ACBObjectOutput && p.Maximum < p.Forward[i] {
			p.Maximum = p.Forward[i]
		}
		if maximumAny < p.Forward[i] {
			maximumAny = p.Forward[i]
		}
		if err := b.use(1+len(n.objects[i].Fanins), 0); err != nil {
			return ACBProfile{}, err
		}
	}
	if len(n.outputs) == 0 {
		p.Maximum = maximumAny
	}
	fanouts, err := n.fanoutsBudget(b)
	if err != nil {
		return ACBProfile{}, err
	}
	for i := len(n.objects) - 1; i > 0; i-- {
		increment := 1
		if n.objects[i].Type == ACBObjectInput || n.objects[i].Type == ACBObjectOutput {
			increment = 0
		}
		for _, f := range fanouts[i] {
			if p.Reverse[i] < p.Reverse[f]+increment {
				p.Reverse[i] = p.Reverse[f] + increment
			}
		}
		if err := b.use(1+len(fanouts[i]), 0); err != nil {
			return ACBProfile{}, err
		}
	}
	for i := 1; i < len(n.objects); i++ {
		increment := 1
		if n.objects[i].Type == ACBObjectInput || n.objects[i].Type == ACBObjectOutput {
			increment = 0
		}
		slack := increment + p.Maximum - p.Forward[i] - p.Reverse[i]
		if slack < 0 {
			if err := b.use(1, 0); err != nil {
				return ACBProfile{}, err
			}
			continue
		}
		if slack == 0 {
			if n.objects[i].Type == ACBObjectInput {
				p.Paths[i] = 1
			} else {
				for _, f := range n.objects[i].Fanins {
					fIncrement := 1
					if n.objects[f].Type == ACBObjectInput || n.objects[f].Type == ACBObjectOutput {
						fIncrement = 0
					}
					if fIncrement+p.Maximum-p.Forward[f]-p.Reverse[f] != 0 {
						continue
					}
					if ^uint64(0)-p.Paths[i] < p.Paths[f] {
						return ACBProfile{}, ErrLimit
					}
					p.Paths[i] += p.Paths[f]
				}
			}
		}
		if err := b.use(1+len(n.objects[i].Fanins), 0); err != nil {
			return ACBProfile{}, err
		}
	}
	return p, nil
}

func (n *ACBNetwork) fanoutsBudget(b *acbBudget) ([][]int, error) {
	r := make([][]int, len(n.objects))
	if err := b.use(0, len(r)*24); err != nil {
		return nil, err
	}
	for i := 1; i < len(n.objects); i++ {
		for _, f := range n.objects[i].Fanins {
			if err := b.use(1, 8); err != nil {
				return nil, err
			}
			r[f] = append(r[f], i)
		}
	}
	return r, nil
}

func (n *ACBNetwork) Fanouts(ctx context.Context) ([][]int, error) {
	if n == nil {
		return nil, ErrInvariant
	}
	b := newACBBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	return n.fanoutsBudget(b)
}

func (n *ACBNetwork) TFI(ctx context.Context, roots []int, includeInputs bool) ([]int, error) {
	if n == nil {
		return nil, ErrInvariant
	}
	b := newACBBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	seen := make([]bool, len(n.objects))
	result := make([]int, 0)
	var visit func(int, int) error
	visit = func(id, depth int) error {
		if depth > b.lim.MaxDepth {
			return ErrLimit
		}
		if id <= 0 || id >= len(n.objects) {
			return ErrInvariant
		}
		if seen[id] {
			return nil
		}
		seen[id] = true
		if err := b.use(1, 0); err != nil {
			return err
		}
		if n.objects[id].Type == ACBObjectInput && !includeInputs {
			return nil
		}
		for _, f := range n.objects[id].Fanins {
			if err := visit(f, depth+1); err != nil {
				return err
			}
		}
		result = append(result, id)
		return nil
	}
	for _, root := range roots {
		if err := visit(root, 0); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (n *ACBNetwork) TFO(ctx context.Context, starts []int, includeOutputs bool) ([]int, error) {
	if n == nil {
		return nil, ErrInvariant
	}
	b := newACBBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	fanouts, err := n.fanoutsBudget(b)
	if err != nil {
		return nil, err
	}
	seen := make([]bool, len(n.objects))
	result := make([]int, 0)
	var visit func(int, int) error
	visit = func(id, depth int) error {
		if depth > b.lim.MaxDepth {
			return ErrLimit
		}
		if id <= 0 || id >= len(n.objects) {
			return ErrInvariant
		}
		if seen[id] {
			return nil
		}
		seen[id] = true
		if err := b.use(1, 0); err != nil {
			return err
		}
		if n.objects[id].Type == ACBObjectOutput && !includeOutputs {
			return nil
		}
		for _, f := range fanouts[id] {
			if err := visit(f, depth+1); err != nil {
				return err
			}
		}
		result = append(result, id)
		return nil
	}
	for _, start := range starts {
		if err := visit(start, 0); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (n *ACBNetwork) Support(ctx context.Context, roots []int) ([]int, error) {
	cone, err := n.TFI(ctx, roots, true)
	if err != nil {
		return nil, err
	}
	r := make([]int, 0)
	for _, id := range cone {
		if n.objects[id].Type == ACBObjectInput {
			r = append(r, id)
		}
	}
	sort.Ints(r)
	return r, nil
}

// MFFC returns the root and recursively dominated internal fanins in postorder.
func (n *ACBNetwork) MFFC(ctx context.Context, root int) ([]int, error) {
	if n == nil {
		return nil, ErrInvariant
	}
	b := newACBBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	fanouts, err := n.fanoutsBudget(b)
	if err != nil {
		return nil, err
	}
	if root <= 0 || root >= len(n.objects) || n.objects[root].Type != ACBObjectNode {
		return nil, ErrInvariant
	}
	seen := make([]bool, len(n.objects))
	result := make([]int, 0)
	var visit func(int, int) error
	visit = func(id, depth int) error {
		if depth > b.lim.MaxDepth {
			return ErrLimit
		}
		if seen[id] || n.objects[id].Type != ACBObjectNode {
			return nil
		}
		if id != root && len(fanouts[id]) != 1 {
			return nil
		}
		seen[id] = true
		if err := b.use(1, 0); err != nil {
			return err
		}
		for _, f := range n.objects[id].Fanins {
			if err := visit(f, depth+1); err != nil {
				return err
			}
		}
		result = append(result, id)
		return nil
	}
	if err := visit(root, 0); err != nil {
		return nil, err
	}
	return result, nil
}

func acbReduceTruth(truth uint64, oldCount int, keep []int) uint64 {
	var result uint64
	for a := 0; a < 1<<len(keep); a++ {
		old := 0
		for k, pos := range keep {
			if a&(1<<k) != 0 {
				old |= 1 << pos
			}
		}
		if truth&(uint64(1)<<old) != 0 {
			result |= uint64(1) << a
		}
	}
	return result & acbTruthMask(len(keep))
}

// MinimizeNodeBase removes functionally inessential inputs transactionally.
func (n *ACBNetwork) MinimizeNodeBase(ctx context.Context, id int) (bool, error) {
	if n == nil {
		return false, ErrInvariant
	}
	b := newACBBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return false, err
	}
	if id <= 0 || id >= len(n.objects) || n.objects[id].Type != ACBObjectNode {
		return false, ErrInvariant
	}
	o := n.objects[id]
	keep := make([]int, 0, len(o.Fanins))
	for v := range o.Fanins {
		essential := false
		for a := 0; a < 1<<len(o.Fanins); a++ {
			if a&(1<<v) != 0 {
				continue
			}
			if err := b.use(1, 0); err != nil {
				return false, err
			}
			if ((o.Truth>>a)^(o.Truth>>(a|(1<<v))))&1 != 0 {
				essential = true
				break
			}
		}
		if essential {
			keep = append(keep, v)
		}
	}
	if len(keep) == len(o.Fanins) {
		return false, nil
	}
	newFanins := make([]int, len(keep))
	for i, k := range keep {
		newFanins[i] = o.Fanins[k]
	}
	n.objects[id].Fanins = newFanins
	n.objects[id].Truth = acbReduceTruth(o.Truth, len(o.Fanins), keep)
	return true, nil
}

// RemoveDuplicateFanins merges repeated signals while preserving the truth.
func (n *ACBNetwork) RemoveDuplicateFanins(ctx context.Context, id int) (bool, error) {
	if n == nil {
		return false, ErrInvariant
	}
	b := newACBBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return false, err
	}
	if id <= 0 || id >= len(n.objects) || n.objects[id].Type != ACBObjectNode {
		return false, ErrInvariant
	}
	o := n.objects[id]
	unique := make([]int, 0, len(o.Fanins))
	position := make([]int, len(o.Fanins))
	for i, f := range o.Fanins {
		pos := -1
		for k, u := range unique {
			if u == f {
				pos = k
				break
			}
		}
		if pos < 0 {
			pos = len(unique)
			unique = append(unique, f)
		}
		position[i] = pos
	}
	if len(unique) == len(o.Fanins) {
		return false, nil
	}
	var truth uint64
	for a := 0; a < 1<<len(unique); a++ {
		old := 0
		for i, pos := range position {
			if a&(1<<pos) != 0 {
				old |= 1 << i
			}
		}
		if err := b.use(1, 0); err != nil {
			return false, err
		}
		if o.Truth&(uint64(1)<<old) != 0 {
			truth |= uint64(1) << a
		}
	}
	n.objects[id].Fanins = unique
	n.objects[id].Truth = truth
	return true, nil
}

// ReplaceNode changes a node only after the complete candidate validates.
func (n *ACBNetwork) ReplaceNode(ctx context.Context, id int, fanins []int, truth uint64) error {
	if n == nil || n.closed || id <= 0 || id >= len(n.objects) || n.objects[id].Type != ACBObjectNode || len(fanins) > 6 {
		return ErrInvariant
	}
	copyN := *n
	copyN.objects = make([]ACBObject, len(n.objects))
	for i := range n.objects {
		copyN.objects[i] = cloneACBObject(n.objects[i])
	}
	copyN.objects[id].Fanins = append([]int(nil), fanins...)
	copyN.objects[id].Truth = truth & acbTruthMask(len(fanins))
	if err := copyN.Validate(ctx); err != nil {
		return err
	}
	n.objects[id] = copyN.objects[id]
	return nil
}

// ACBStats is a deterministic structural summary.
type ACBStats struct {
	Objects int
	Inputs  int
	Outputs int
	Nodes   int
	Boxes   int
	Edges   int
	Regs    int
	Depth   int
}

func (n *ACBNetwork) Stats(ctx context.Context) (ACBStats, error) {
	profile, err := n.Profile(ctx)
	if err != nil {
		return ACBStats{}, err
	}
	s := ACBStats{Objects: len(n.objects) - 1, Inputs: len(n.inputs), Outputs: len(n.outputs), Regs: n.regs, Depth: profile.Maximum}
	for i := 1; i < len(n.objects); i++ {
		s.Edges += len(n.objects[i].Fanins)
		if n.objects[i].Type == ACBObjectNode {
			s.Nodes++
		}
		if n.objects[i].Type == ACBObjectBox {
			s.Boxes++
		}
	}
	return s, nil
}

func (n *ACBNetwork) Report(ctx context.Context) (string, error) {
	s, err := n.Stats(ctx)
	if err != nil {
		return "", err
	}
	r := fmt.Sprintf("network %s objects=%d inputs=%d outputs=%d nodes=%d boxes=%d edges=%d regs=%d depth=%d", n.name, s.Objects, s.Inputs, s.Outputs, s.Nodes, s.Boxes, s.Edges, s.Regs, s.Depth)
	if len(r) > n.limits.MaxText {
		return "", ErrLimit
	}
	return r, nil
}

// ACBManager owns named modules and a root module.
type ACBManager struct {
	name     string
	modules  []*ACBNetwork
	root     int
	closed   bool
	limits   ACBLimits
	nameMap  []int
	nameMap2 []int
}

func NewACBManager(name string, limits ACBLimits) (*ACBManager, error) {
	limits = limits.normalized()
	if name == "" || len(name) > limits.MaxText {
		return nil, ErrInvariant
	}
	return &ACBManager{name: name, modules: []*ACBNetwork{nil}, limits: limits}, nil
}

func (m *ACBManager) AddNetwork(ctx context.Context, n *ACBNetwork) (int, error) {
	if m == nil || m.closed || n == nil || n.closed || len(m.modules) >= m.limits.MaxObjects {
		return 0, ErrInvariant
	}
	b := newACBBudget(ctx, m.limits)
	objects, edges := 0, 0
	all := append(append([]*ACBNetwork(nil), m.modules[1:]...), n)
	for _, module := range all {
		if module == nil || module.closed {
			return 0, ErrInvariant
		}
		if err := module.validateBudget(b); err != nil {
			return 0, err
		}
		objects += len(module.objects) - 1
		bytes := len(module.name)
		for i := 1; i < len(module.objects); i++ {
			o := &module.objects[i]
			edges += len(o.Fanins)
			bytes += 96 + 8*len(o.Fanins) + 16*len(o.Attributes) + len(o.Name)
		}
		if objects > m.limits.MaxObjects || edges > m.limits.MaxEdges {
			return 0, ErrLimit
		}
		if err := b.use(0, bytes); err != nil {
			return 0, err
		}
	}
	for i := 1; i < len(m.modules); i++ {
		if m.modules[i].name == n.name {
			return 0, ErrInvariant
		}
	}
	m.modules = append(m.modules, n)
	if m.root == 0 {
		m.root = len(m.modules) - 1
	}
	return len(m.modules) - 1, nil
}

func (m *ACBManager) SetRoot(id int) error {
	if m == nil || m.closed || id <= 0 || id >= len(m.modules) {
		return ErrInvariant
	}
	m.root = id
	return nil
}

func (m *ACBManager) Root() (*ACBNetwork, error) {
	if m == nil || m.closed || m.root <= 0 || m.root >= len(m.modules) {
		return nil, ErrInvariant
	}
	return m.modules[m.root], nil
}

func (m *ACBManager) Find(name string) (int, bool) {
	if m == nil || m.closed {
		return 0, false
	}
	for i := 1; i < len(m.modules); i++ {
		if m.modules[i].name == name {
			return i, true
		}
	}
	return 0, false
}

func (m *ACBManager) Close() {
	if m == nil {
		return
	}
	for i := 1; i < len(m.modules); i++ {
		m.modules[i].Close()
	}
	m.closed = true
	m.modules = nil
	m.nameMap = nil
	m.nameMap2 = nil
}

// ACBSetStore provides deterministic canonical integer-set storage.
type ACBSetStore struct {
	sets   [][]int
	bytes  int
	closed bool
	limits ACBLimits
}

func NewACBSetStore(limits ACBLimits) *ACBSetStore {
	return &ACBSetStore{limits: limits.normalized(), sets: [][]int{nil}}
}

func canonicalACBSet(values []int) ([]int, error) {
	r := append([]int(nil), values...)
	sort.Ints(r)
	w := 0
	for _, v := range r {
		if v < 0 {
			return nil, ErrInvariant
		}
		if w == 0 || r[w-1] != v {
			r[w] = v
			w++
		}
	}
	return r[:w], nil
}

func (s *ACBSetStore) Intern(ctx context.Context, values []int) (int, error) {
	if s == nil || s.closed {
		return 0, ErrInvariant
	}
	b := newACBBudget(ctx, s.limits)
	if len(values) > b.lim.MaxOperations || len(values) > b.lim.MaxBytes/8 {
		return 0, ErrLimit
	}
	if err := b.use(len(values), len(values)*8); err != nil {
		return 0, err
	}
	sortOps := 0
	for width := len(values); width > 1; width = (width + 1) / 2 {
		if sortOps > b.lim.MaxOperations-len(values) {
			return 0, ErrLimit
		}
		sortOps += len(values)
	}
	if err := b.use(sortOps, 0); err != nil {
		return 0, err
	}
	r, err := canonicalACBSet(values)
	if err != nil {
		return 0, err
	}
	if err := b.use(0, 0); err != nil {
		return 0, err
	}
	return s.internCanonical(b, r)
}

func (s *ACBSetStore) internCanonical(b *acbBudget, r []int) (int, error) {
	for i := 1; i < len(s.sets); i++ {
		if err := b.use(1+len(r), 0); err != nil {
			return 0, err
		}
		if len(s.sets[i]) != len(r) {
			continue
		}
		equal := true
		for k := range r {
			if s.sets[i][k] != r[k] {
				equal = false
				break
			}
		}
		if equal {
			return i, nil
		}
	}
	if len(s.sets) >= s.limits.MaxObjects {
		return 0, ErrLimit
	}
	bytes := len(r) * 8
	if len(r) > s.limits.MaxBytes/8 || s.bytes > s.limits.MaxBytes-bytes {
		return 0, ErrLimit
	}
	if err := b.use(0, bytes); err != nil {
		return 0, err
	}
	s.sets = append(s.sets, append([]int(nil), r...))
	s.bytes += bytes
	return len(s.sets) - 1, nil
}

func (s *ACBSetStore) Values(id int) ([]int, error) {
	if s == nil || s.closed || id <= 0 || id >= len(s.sets) {
		return nil, ErrInvariant
	}
	return append([]int(nil), s.sets[id]...), nil
}

func (s *ACBSetStore) combine(ctx context.Context, left, right int, mode uint8) (int, error) {
	if s == nil || s.closed || left <= 0 || right <= 0 || left >= len(s.sets) || right >= len(s.sets) {
		return 0, ErrInvariant
	}
	a, b := s.sets[left], s.sets[right]
	if len(a) > s.limits.MaxOperations-len(b) || len(a) > s.limits.MaxBytes/8-len(b) {
		return 0, ErrLimit
	}
	budget := newACBBudget(ctx, s.limits)
	if err := budget.use(0, (len(a)+len(b))*8); err != nil {
		return 0, err
	}
	result := make([]int, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		if err := budget.use(1, 8); err != nil {
			return 0, err
		}
		if j == len(b) || i < len(a) && a[i] < b[j] {
			if mode == 0 || mode == 2 {
				result = append(result, a[i])
			}
			i++
		} else if i == len(a) || b[j] < a[i] {
			if mode == 0 {
				result = append(result, b[j])
			}
			j++
		} else {
			if mode == 0 || mode == 1 {
				result = append(result, a[i])
			}
			i++
			j++
		}
	}
	return s.internCanonical(budget, result)
}

func (s *ACBSetStore) Union(ctx context.Context, left, right int) (int, error) {
	return s.combine(ctx, left, right, 0)
}

func (s *ACBSetStore) Intersection(ctx context.Context, left, right int) (int, error) {
	return s.combine(ctx, left, right, 1)
}

func (s *ACBSetStore) Difference(ctx context.Context, left, right int) (int, error) {
	return s.combine(ctx, left, right, 2)
}

func (s *ACBSetStore) IsSubset(left, right int) (bool, error) {
	return s.IsSubsetContext(context.Background(), left, right)
}

func (s *ACBSetStore) IsSubsetContext(ctx context.Context, left, right int) (bool, error) {
	if s == nil || s.closed || left <= 0 || right <= 0 || left >= len(s.sets) || right >= len(s.sets) {
		return false, ErrInvariant
	}
	a, b := s.sets[left], s.sets[right]
	if len(a) > s.limits.MaxOperations-len(b) {
		return false, ErrLimit
	}
	budget := newACBBudget(ctx, s.limits)
	if err := budget.use(0, 0); err != nil {
		return false, err
	}
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if err := budget.use(1, 0); err != nil {
			return false, err
		}
		if a[i] == b[j] {
			i++
			j++
		} else if a[i] > b[j] {
			j++
		} else {
			return false, nil
		}
	}
	return i == len(a), nil
}

func (s *ACBSetStore) Close() {
	if s != nil {
		s.closed = true
		s.sets = nil
		s.bytes = 0
	}
}

// ACBParameters mirrors the source optimization defaults with Go booleans.
type ACBParameters struct {
	LUTSize       int
	TFOMaxLevels  int
	TFIMaxLevels  int
	FanoutMax     int
	WindowNodeMax int
	GrowthLevel   int
	ConflictLimit int
	NodesMax      int
	UseAshenhurst bool
	OneNode       int
	Area          bool
	MoreEffort    bool
	Verbose       bool
	VeryVerbose   bool
}

func DefaultACBParameters() ACBParameters {
	return ACBParameters{LUTSize: 4, TFOMaxLevels: 2, TFIMaxLevels: 3, FanoutMax: 20, WindowNodeMax: 100, Area: true}
}

// ACBModuleBuildInventory returns the exact implementation list in module.make.
func ACBModuleBuildInventory() []string {
	return []string{"acbAbc.c", "acbAig.c", "acbCom.c", "acbFunc.c", "acbMfs.c", "acbPush.c", "acbSets.c", "acbTest.c", "acbUtil.c"}
}

// ACBSourceCoverage summarizes intentionally represented source areas. It is
// deterministic diagnostic data, not a claim that broad integration paths exist.
func ACBSourceCoverage() string {
	parts := []string{
		"acb.h:owned network manager objects ranges attributes traversal fanout",
		"acbAbc.c:parameters and concrete structural conversion model",
		"acbCom.c:owned lifecycle without command-frame mutation",
		"acbFunc.c:truth evaluation cones support and deterministic names",
		"acbMfs.c:window primitives profiles MFFC and transactional base reduction",
		"acbPush.c:duplicate-fanin and minimum-base normalization",
		"acbSets.h:empty declaration surface; canonical ordered sets are Go support",
		"acbTest.c:standalone equivalence models",
		"acbUtil.c:fanouts TFI TFO support levels paths statistics",
	}
	return strings.Join(parts, "\n")
}
