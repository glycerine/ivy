package abc

// Ported from src/opt/nwk/nwk.h, ntlnwk.h, nwkAig.c, nwkBidec.c,
// nwkDfs.c, nwkFanio.c, nwkFlow.c, nwkFlow_depth.c, nwkMan.c,
// nwkMap.c, nwkMerge.c, nwkMerge.h, nwkObj.c, nwkSpeedup.c,
// nwkStrash.c, nwkTiming.c, nwkUtil.c, and module.make.

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
)

// NWKObjectType preserves the source Nwk_Type_t numbering.
type NWKObjectType uint8

const (
	NWKObjectNone NWKObjectType = iota
	NWKObjectCI
	NWKObjectCO
	NWKObjectNode
	NWKObjectLatch
	NWKObjectVoid
)

// NWKObject is an owned object snapshot. Truth uses the low 2^nFanins bits.
type NWKObject struct {
	ID        int
	Type      NWKObjectType
	Name      string
	Fanins    []int
	Truth     uint64
	Invert    bool
	Level     int
	Arrival   float64
	Required  float64
	Slack     float64
	Switching float64
}

// NWKLimits bounds all graph, truth, reporting, and traversal work.
type NWKLimits struct {
	MaxObjects, MaxEdges, MaxFanins, MaxOperations int
	MaxBytes, MaxDepth, MaxText, MaxCuts           int
}

func (l NWKLimits) normalized() NWKLimits {
	if l.MaxObjects <= 0 {
		l.MaxObjects = 1 << 16
	}
	if l.MaxEdges <= 0 {
		l.MaxEdges = 1 << 20
	}
	if l.MaxFanins <= 0 {
		l.MaxFanins = 6
	}
	if l.MaxFanins > 6 {
		l.MaxFanins = 6
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
	if l.MaxCuts <= 0 {
		l.MaxCuts = 1 << 18
	}
	return l
}

type nwkBudget struct {
	ctx        context.Context
	lim        NWKLimits
	ops, bytes int
}

func newNWKBudget(ctx context.Context, l NWKLimits) *nwkBudget {
	if ctx == nil {
		ctx = context.Background()
	}
	return &nwkBudget{ctx: ctx, lim: l.normalized()}
}

func (b *nwkBudget) use(ops, bytes int) error {
	if ops < 0 || bytes < 0 || b.ops > b.lim.MaxOperations-ops || b.bytes > b.lim.MaxBytes-bytes {
		return ErrLimit
	}
	b.ops += ops
	b.bytes += bytes
	select {
	case <-b.ctx.Done():
		return ErrCancelled
	default:
		return nil
	}
}

// NWKNetwork is an owned, stable-ID, topologically constructed logic network.
type NWKNetwork struct {
	name, spec        string
	objects           []NWKObject
	cis, cos, latches []int
	truePIs, truePOs  int
	edges, retained   int
	limits            NWKLimits
	partitions        bool
	closed            bool
}

func NewNWKNetwork(name string, limits NWKLimits) (*NWKNetwork, error) {
	l := limits.normalized()
	if len(name) > l.MaxText || len(name) > l.MaxBytes {
		return nil, ErrLimit
	}
	return &NWKNetwork{name: name, limits: l, retained: len(name)}, nil
}

func cloneNWKObject(o NWKObject) NWKObject { o.Fanins = append([]int(nil), o.Fanins...); return o }

func (n *NWKNetwork) checkLive() error {
	if n == nil || n.closed {
		return fmt.Errorf("%w: closed NWK network", ErrInvariant)
	}
	return nil
}

func (n *NWKNetwork) Name() string {
	if n == nil {
		return ""
	}
	return n.name
}
func (n *NWKNetwork) ObjectCount() int {
	if n == nil {
		return 0
	}
	return len(n.objects)
}
func (n *NWKNetwork) InputCount() int {
	if n == nil {
		return 0
	}
	return len(n.cis)
}
func (n *NWKNetwork) OutputCount() int {
	if n == nil {
		return 0
	}
	return len(n.cos)
}
func (n *NWKNetwork) NodeCount() int {
	if n == nil {
		return 0
	}
	c := 0
	for _, o := range n.objects {
		if o.Type == NWKObjectNode {
			c++
		}
	}
	return c
}
func (n *NWKNetwork) LatchCount() int {
	if n == nil {
		return 0
	}
	return len(n.latches)
}

func (n *NWKNetwork) Object(id int) (NWKObject, error) {
	if err := n.checkLive(); err != nil {
		return NWKObject{}, err
	}
	if id < 0 || id >= len(n.objects) || n.objects[id].Type == NWKObjectNone || n.objects[id].Type == NWKObjectVoid {
		return NWKObject{}, ErrDimension
	}
	return cloneNWKObject(n.objects[id]), nil
}

func (n *NWKNetwork) appendObject(ctx context.Context, typ NWKObjectType, name string, fanins []int, truth uint64) (int, error) {
	if err := n.checkLive(); err != nil {
		return 0, err
	}
	l := n.limits
	if typ <= NWKObjectNone || typ >= NWKObjectVoid || len(n.objects) >= l.MaxObjects || len(fanins) > l.MaxFanins || len(name) > l.MaxText {
		return 0, ErrLimit
	}
	if typ == NWKObjectCI || typ == NWKObjectLatch {
		if len(fanins) != 0 {
			return 0, ErrDimension
		}
	}
	if typ == NWKObjectCO && len(fanins) != 1 {
		return 0, ErrDimension
	}
	for _, f := range fanins {
		if f < 0 || f >= len(n.objects) || n.objects[f].Type == NWKObjectCO || n.objects[f].Type == NWKObjectNone || n.objects[f].Type == NWKObjectVoid {
			return 0, ErrInvariant
		}
	}
	if name != "" {
		for _, o := range n.objects {
			if o.Name == name {
				return 0, fmt.Errorf("%w: duplicate name", ErrInvariant)
			}
		}
	}
	bits := 1 << len(fanins)
	if bits < 64 && truth>>bits != 0 {
		return 0, ErrDimension
	}
	bytes := len(name) + len(fanins)*8 + 96
	if n.edges > l.MaxEdges-len(fanins) || n.retained > l.MaxBytes-bytes {
		return 0, ErrLimit
	}
	if err := newNWKBudget(ctx, l).use(len(fanins)+1, 0); err != nil {
		return 0, err
	}
	id := len(n.objects)
	cp := append([]int(nil), fanins...)
	level := 0
	for _, f := range cp {
		if n.objects[f].Level+1 > level {
			level = n.objects[f].Level + 1
		}
	}
	if typ == NWKObjectCO {
		level = n.objects[cp[0]].Level
	}
	o := NWKObject{ID: id, Type: typ, Name: name, Fanins: cp, Truth: truth, Level: level}
	n.objects = append(n.objects, o)
	n.edges += len(cp)
	n.retained += bytes
	switch typ {
	case NWKObjectCI:
		n.cis = append(n.cis, id)
	case NWKObjectCO:
		n.cos = append(n.cos, id)
	case NWKObjectLatch:
		n.latches = append(n.latches, id)
	}
	return id, nil
}

func (n *NWKNetwork) AddCI(ctx context.Context, name string) (int, error) {
	return n.appendObject(ctx, NWKObjectCI, name, nil, 0)
}
func (n *NWKNetwork) AddNode(ctx context.Context, name string, fanins []int, truth uint64) (int, error) {
	return n.appendObject(ctx, NWKObjectNode, name, fanins, truth)
}
func (n *NWKNetwork) AddCO(ctx context.Context, name string, driver int) (int, error) {
	return n.appendObject(ctx, NWKObjectCO, name, []int{driver}, 0)
}
func (n *NWKNetwork) AddLatch(ctx context.Context, name string) (int, error) {
	return n.appendObject(ctx, NWKObjectLatch, name, nil, 0)
}

func (n *NWKNetwork) SetSequentialPartitions(truePIs, truePOs int) error {
	if err := n.checkLive(); err != nil {
		return err
	}
	if truePIs < 0 || truePOs < 0 || truePIs > len(n.cis) || truePOs > len(n.cos) || len(n.cis)-truePIs != len(n.cos)-truePOs {
		return ErrDimension
	}
	n.truePIs = truePIs
	n.truePOs = truePOs
	n.partitions = true
	return nil
}

func (n *NWKNetwork) Snapshot(ctx context.Context) ([]NWKObject, error) {
	if err := n.checkLive(); err != nil {
		return nil, err
	}
	b := newNWKBudget(ctx, n.limits)
	if err := b.use(len(n.objects), len(n.objects)*96+n.edges*8); err != nil {
		return nil, err
	}
	r := make([]NWKObject, len(n.objects))
	for i, o := range n.objects {
		r[i] = cloneNWKObject(o)
	}
	return r, nil
}

func (n *NWKNetwork) fanouts(b *nwkBudget) ([][]int, error) {
	if err := b.use(len(n.objects)+n.edges, len(n.objects)*24+n.edges*8); err != nil {
		return nil, err
	}
	r := make([][]int, len(n.objects))
	for _, o := range n.objects {
		for _, f := range o.Fanins {
			r[f] = append(r[f], o.ID)
		}
	}
	return r, nil
}

func (n *NWKNetwork) validateBudget(b *nwkBudget) error {
	if err := n.checkLive(); err != nil {
		return err
	}
	if len(n.objects) > n.limits.MaxObjects || n.edges > n.limits.MaxEdges || n.retained > n.limits.MaxBytes {
		return ErrLimit
	}
	if err := b.use(0, len(n.objects)*32); err != nil {
		return err
	}
	names := make(map[string]int, len(n.objects))
	edges := 0
	retained := len(n.name) + len(n.spec)
	cis := make([]int, 0, len(n.cis))
	cos := make([]int, 0, len(n.cos))
	latches := make([]int, 0, len(n.latches))
	for i, o := range n.objects {
		if err := b.use(len(o.Fanins)+1, len(o.Name)); err != nil {
			return err
		}
		if o.ID != i || o.Type <= NWKObjectNone || o.Type >= NWKObjectVoid {
			return fmt.Errorf("%w: object %d metadata", ErrInvariant, i)
		}
		if o.Name != "" {
			if _, ok := names[o.Name]; ok {
				return fmt.Errorf("%w: duplicate name", ErrInvariant)
			}
			names[o.Name] = i
		}
		if len(o.Fanins) > n.limits.MaxFanins {
			return ErrLimit
		}
		edges += len(o.Fanins)
		if len(o.Name) > n.limits.MaxText || retained > n.limits.MaxBytes-(96+len(o.Name)+len(o.Fanins)*8) {
			return ErrLimit
		}
		retained += 96 + len(o.Name) + len(o.Fanins)*8
		for _, f := range o.Fanins {
			if f < 0 || f >= i || n.objects[f].Type == NWKObjectCO {
				return fmt.Errorf("%w: fanin", ErrInvariant)
			}
		}
		switch o.Type {
		case NWKObjectCI:
			if len(o.Fanins) != 0 {
				return ErrInvariant
			}
			cis = append(cis, i)
		case NWKObjectLatch:
			if len(o.Fanins) != 0 {
				return ErrInvariant
			}
			latches = append(latches, i)
		case NWKObjectCO:
			if len(o.Fanins) != 1 {
				return ErrInvariant
			}
			cos = append(cos, i)
		case NWKObjectNode:
			// The source uses zero-fanin nodes for constants.
		}
		bits := 1 << len(o.Fanins)
		if bits < 64 && o.Truth>>bits != 0 {
			return ErrDimension
		}
		level := 0
		for _, f := range o.Fanins {
			if n.objects[f].Level+1 > level {
				level = n.objects[f].Level + 1
			}
		}
		if o.Type == NWKObjectCO {
			level = n.objects[o.Fanins[0]].Level
		}
		if (o.Type == NWKObjectCI || o.Type == NWKObjectLatch) && o.Level != 0 || (o.Type == NWKObjectNode || o.Type == NWKObjectCO) && o.Level != level {
			return fmt.Errorf("%w: object %d level", ErrInvariant, i)
		}
	}
	if edges != n.edges {
		return fmt.Errorf("%w: edge count", ErrInvariant)
	}
	if retained != n.retained {
		return fmt.Errorf("%w: retained byte count", ErrInvariant)
	}
	if !equalNWKIDs(cis, n.cis) || !equalNWKIDs(cos, n.cos) || !equalNWKIDs(latches, n.latches) {
		return fmt.Errorf("%w: object index", ErrInvariant)
	}
	if n.truePIs < 0 || n.truePOs < 0 || n.truePIs > len(n.cis) || n.truePOs > len(n.cos) {
		return ErrInvariant
	}
	if n.partitions && len(n.cis)-n.truePIs != len(n.cos)-n.truePOs {
		return fmt.Errorf("%w: sequential partitions", ErrInvariant)
	}
	return nil
}

func equalNWKIDs(a, b []int) bool {
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

func (n *NWKNetwork) recomputeLevels() {
	for i := range n.objects {
		o := &n.objects[i]
		level := 0
		for _, f := range o.Fanins {
			if n.objects[f].Level+1 > level {
				level = n.objects[f].Level + 1
			}
		}
		if o.Type == NWKObjectCO {
			level = n.objects[o.Fanins[0]].Level
		}
		o.Level = level
	}
}

func (n *NWKNetwork) Validate(ctx context.Context) error {
	return n.validateBudget(newNWKBudget(ctx, n.limits))
}

func nwkTruthBit(truth uint64, values []bool) bool {
	idx := 0
	for i, v := range values {
		if v {
			idx |= 1 << i
		}
	}
	return truth&(uint64(1)<<idx) != 0
}

func (n *NWKNetwork) Evaluate(ctx context.Context, inputs []bool) ([]bool, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	if len(inputs) != len(n.cis) {
		return nil, ErrDimension
	}
	if err := b.use(len(inputs), len(n.objects)+len(n.cos)); err != nil {
		return nil, err
	}
	vals := make([]bool, len(n.objects))
	for i, id := range n.cis {
		vals[id] = inputs[i]
	}
	for _, o := range n.objects {
		if err := b.use(len(o.Fanins)+1, len(o.Fanins)); err != nil {
			return nil, err
		}
		switch o.Type {
		case NWKObjectNode:
			args := make([]bool, len(o.Fanins))
			for j, f := range o.Fanins {
				args[j] = vals[f]
			}
			vals[o.ID] = nwkTruthBit(o.Truth, args) != o.Invert
		case NWKObjectCO:
			vals[o.ID] = vals[o.Fanins[0]] != o.Invert
		}
	}
	r := make([]bool, len(n.cos))
	for i, id := range n.cos {
		r[i] = vals[id]
	}
	return r, nil
}

func (n *NWKNetwork) Fanouts(ctx context.Context) ([][]int, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	return n.fanouts(b)
}

func (n *NWKNetwork) DFS(ctx context.Context, roots []int) ([]int, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	return n.dfsBudget(b, roots)
}

func (n *NWKNetwork) dfsBudget(b *nwkBudget, roots []int) ([]int, error) {
	if err := b.use(0, len(n.objects)*9); err != nil {
		return nil, err
	}
	state := make([]uint8, len(n.objects))
	out := make([]int, 0, len(n.objects))
	all := roots == nil
	if all {
		roots = append([]int(nil), n.cos...)
	}
	var visit func(int, int) error
	visit = func(id, d int) error {
		if d > n.limits.MaxDepth {
			return ErrLimit
		}
		if id < 0 || id >= len(n.objects) {
			return ErrDimension
		}
		if state[id] == 2 {
			return nil
		}
		if state[id] == 1 {
			return ErrInvariant
		}
		state[id] = 1
		if err := b.use(1, 0); err != nil {
			return err
		}
		for _, f := range n.objects[id].Fanins {
			if err := visit(f, d+1); err != nil {
				return err
			}
		}
		state[id] = 2
		out = append(out, id)
		return nil
	}
	// Nwk_ManDfs places every CI first, including unused CIs, and then
	// recursively visits each CO cone.
	if all {
		for _, id := range n.cis {
			state[id] = 2
			out = append(out, id)
		}
	}
	for _, r := range roots {
		if err := visit(r, 0); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (n *NWKNetwork) ReverseDFS(ctx context.Context, starts []int) ([]int, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	fo, err := n.fanouts(b)
	if err != nil {
		return nil, err
	}
	if starts == nil {
		starts = append([]int(nil), n.cis...)
	}
	if err := b.use(0, len(n.objects)*10+len(starts)*8); err != nil {
		return nil, err
	}
	seen := make([]bool, len(n.objects))
	out := make([]int, 0, len(n.objects))
	var visit func(int, int) error
	visit = func(id, depth int) error {
		if depth > n.limits.MaxDepth {
			return ErrLimit
		}
		if id < 0 || id >= len(n.objects) {
			return ErrDimension
		}
		if seen[id] {
			return nil
		}
		seen[id] = true
		if err := b.use(1, 0); err != nil {
			return err
		}
		for _, next := range fo[id] {
			if err := visit(next, depth+1); err != nil {
				return err
			}
		}
		out = append(out, id)
		return nil
	}
	for _, id := range starts {
		if err := visit(id, 0); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (n *NWKNetwork) Support(ctx context.Context, roots []int) ([]int, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	order, err := n.dfsBudget(b, roots)
	if err != nil {
		return nil, err
	}
	if err := b.use(len(order)+len(n.cis), len(n.objects)+len(n.cis)*8); err != nil {
		return nil, err
	}
	mark := make([]bool, len(n.objects))
	for _, id := range order {
		if n.objects[id].Type == NWKObjectCI {
			mark[id] = true
		}
	}
	r := make([]int, 0)
	for _, id := range n.cis {
		if mark[id] {
			r = append(r, id)
		}
	}
	return r, nil
}

func (n *NWKNetwork) MFFC(ctx context.Context, root int) ([]int, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	if root < 0 || root >= len(n.objects) || n.objects[root].Type != NWKObjectNode {
		return nil, ErrDimension
	}
	fo, err := n.fanouts(b)
	if err != nil {
		return nil, err
	}
	if err := b.use(0, len(fo)*8); err != nil {
		return nil, err
	}
	refs := make([]int, len(fo))
	for i := range fo {
		refs[i] = len(fo[i])
	}
	out := make([]int, 0)
	var deref func(int, int) error
	deref = func(id, d int) error {
		if d > n.limits.MaxDepth {
			return ErrLimit
		}
		if err := b.use(1, 0); err != nil {
			return err
		}
		if n.objects[id].Type != NWKObjectNode {
			return nil
		}
		out = append(out, id)
		for _, f := range n.objects[id].Fanins {
			refs[f]--
			if refs[f] == 0 {
				if err := deref(f, d+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := deref(root, 0); err != nil {
		return nil, err
	}
	sort.Ints(out)
	return out, nil
}

// NWKTimingOptions models LUT pin delays and output required time.
type NWKTimingOptions struct {
	PinDelays []float64
	Required  float64
	Epsilon   float64
}
type NWKTimingResult struct {
	Delay                    float64
	Arrival, Required, Slack []float64
	Critical                 []int
}

func (n *NWKNetwork) TraceTiming(ctx context.Context, opt NWKTimingOptions) (NWKTimingResult, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return NWKTimingResult{}, err
	}
	if err := b.use(0, len(n.objects)*24); err != nil {
		return NWKTimingResult{}, err
	}
	if math.IsNaN(opt.Required) || math.IsInf(opt.Required, 0) {
		return NWKTimingResult{}, ErrDimension
	}
	for _, delay := range opt.PinDelays {
		if delay < 0 || math.IsNaN(delay) || math.IsInf(delay, 0) {
			return NWKTimingResult{}, ErrDimension
		}
	}
	arr := make([]float64, len(n.objects))
	req := make([]float64, len(n.objects))
	for i := range req {
		req[i] = math.Inf(1)
	}
	delay := 0.0
	for _, o := range n.objects {
		if err := b.use(len(o.Fanins)+1, 0); err != nil {
			return NWKTimingResult{}, err
		}
		switch o.Type {
		case NWKObjectNode:
			best := 0.0
			for j, f := range o.Fanins {
				pd := 1.0
				if j < len(opt.PinDelays) {
					pd = opt.PinDelays[j]
				}
				if arr[f]+pd > best {
					best = arr[f] + pd
				}
			}
			arr[o.ID] = best
		case NWKObjectCO:
			arr[o.ID] = arr[o.Fanins[0]]
		}
		if o.Type == NWKObjectCO && arr[o.ID] > delay {
			delay = arr[o.ID]
		}
	}
	target := opt.Required
	if target <= 0 {
		target = delay
	}
	for _, id := range n.cos {
		req[id] = target
		f := n.objects[id].Fanins[0]
		if target < req[f] {
			req[f] = target
		}
	}
	for i := len(n.objects) - 1; i >= 0; i-- {
		o := n.objects[i]
		if o.Type != NWKObjectNode || math.IsInf(req[i], 1) {
			continue
		}
		for j, f := range o.Fanins {
			pd := 1.0
			if j < len(opt.PinDelays) {
				pd = opt.PinDelays[j]
			}
			v := req[i] - pd
			if v < req[f] {
				req[f] = v
			}
		}
	}
	slack := make([]float64, len(arr))
	crit := make([]int, 0)
	eps := opt.Epsilon
	if eps <= 0 {
		eps = 1e-9
	}
	for i := range arr {
		slack[i] = req[i] - arr[i]
		if !math.IsInf(req[i], 1) && math.Abs(slack[i]) <= eps {
			crit = append(crit, i)
		}
	}
	return NWKTimingResult{Delay: delay, Arrival: arr, Required: req, Slack: slack, Critical: crit}, nil
}

func nwkRemoveTruthVar(truth uint64, nvars, drop int) uint64 {
	var r uint64
	for a := 0; a < (1 << (nvars - 1)); a++ {
		lo := a & ((1 << drop) - 1)
		hi := a >> drop
		old := lo | (hi << (drop + 1))
		if truth&(uint64(1)<<old) != 0 {
			r |= uint64(1) << a
		}
	}
	return r
}
func nwkTruthIndependent(truth uint64, nvars, v int) bool {
	for a := 0; a < (1 << nvars); a++ {
		if a&(1<<v) != 0 {
			continue
		}
		b := a | (1 << v)
		if ((truth>>a)^(truth>>b))&1 != 0 {
			return false
		}
	}
	return true
}

func (n *NWKNetwork) MinimizeBase(ctx context.Context, id int) (bool, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return false, err
	}
	if id < 0 || id >= len(n.objects) || n.objects[id].Type != NWKObjectNode {
		return false, ErrDimension
	}
	o := cloneNWKObject(n.objects[id])
	changed := false
	for v := len(o.Fanins) - 1; v >= 0; v-- {
		if err := b.use(1<<len(o.Fanins), 0); err != nil {
			return false, err
		}
		if nwkTruthIndependent(o.Truth, len(o.Fanins), v) {
			o.Truth = nwkRemoveTruthVar(o.Truth, len(o.Fanins), v)
			o.Fanins = append(o.Fanins[:v], o.Fanins[v+1:]...)
			changed = true
		}
	}
	if changed {
		delta := len(n.objects[id].Fanins) - len(o.Fanins)
		n.edges -= delta
		n.retained -= delta * 8
		n.objects[id] = o
		n.recomputeLevels()
	}
	return changed, nil
}

func nwkMergeDuplicateTruth(truth uint64, nvars, keep, drop int) uint64 {
	var r uint64
	for a := 0; a < (1 << (nvars - 1)); a++ {
		lo := a & ((1 << drop) - 1)
		hi := a >> drop
		old := lo | (hi << (drop + 1))
		kb := keep
		if keep > drop {
			kb--
		}
		if a&(1<<kb) != 0 {
			old |= 1 << drop
		} else {
			old &^= 1 << drop
		}
		if truth&(uint64(1)<<old) != 0 {
			r |= uint64(1) << a
		}
	}
	return r
}

func (n *NWKNetwork) RemoveDuplicateFanins(ctx context.Context, id int) (bool, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return false, err
	}
	if id < 0 || id >= len(n.objects) || n.objects[id].Type != NWKObjectNode {
		return false, ErrDimension
	}
	o := cloneNWKObject(n.objects[id])
	changed := false
	for j := len(o.Fanins) - 1; j >= 0; j-- {
		for i := 0; i < j; i++ {
			if o.Fanins[i] == o.Fanins[j] {
				if err := b.use(1<<len(o.Fanins), 0); err != nil {
					return false, err
				}
				o.Truth = nwkMergeDuplicateTruth(o.Truth, len(o.Fanins), i, j)
				o.Fanins = append(o.Fanins[:j], o.Fanins[j+1:]...)
				changed = true
				break
			}
		}
	}
	if changed {
		delta := len(n.objects[id].Fanins) - len(o.Fanins)
		n.edges -= delta
		n.retained -= delta * 8
		n.objects[id] = o
		n.recomputeLevels()
	}
	return changed, nil
}

func (n *NWKNetwork) ReplaceNode(ctx context.Context, id int, fanins []int, truth uint64) error {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return err
	}
	if id < 0 || id >= len(n.objects) || n.objects[id].Type != NWKObjectNode || len(fanins) > n.limits.MaxFanins {
		return ErrDimension
	}
	for _, f := range fanins {
		if f < 0 || f >= id || n.objects[f].Type == NWKObjectCO {
			return ErrInvariant
		}
	}
	bits := 1 << len(fanins)
	if bits < 64 && truth>>bits != 0 {
		return ErrDimension
	}
	delta := len(fanins) - len(n.objects[id].Fanins)
	if delta > 0 && (n.edges > n.limits.MaxEdges-delta || n.retained > n.limits.MaxBytes-delta*8) {
		return ErrLimit
	}
	if err := b.use(bits, len(fanins)*8); err != nil {
		return err
	}
	o := n.objects[id]
	o.Fanins = append([]int(nil), fanins...)
	o.Truth = truth
	level := 0
	for _, f := range fanins {
		if n.objects[f].Level+1 > level {
			level = n.objects[f].Level + 1
		}
	}
	o.Level = level
	n.objects[id] = o
	n.edges += delta
	n.retained += delta * 8
	n.recomputeLevels()
	return nil
}

func nwkTruthToAIG(m *AigManager, args []AigLit, truth uint64, b *nwkBudget) (AigLit, error) {
	var rec func([]AigLit, uint64) (AigLit, error)
	rec = func(v []AigLit, t uint64) (AigLit, error) {
		if err := b.use(1, 0); err != nil {
			return InvalidAigLit, err
		}
		if len(v) == 0 {
			if t&1 != 0 {
				return AigConst1, nil
			}
			return AigConst0, nil
		}
		half := 1 << (len(v) - 1)
		mask := uint64(1) << half
		mask--
		lo := t & mask
		hi := (t >> half) & mask
		if lo == hi {
			return rec(v[:len(v)-1], lo)
		}
		a, e := rec(v[:len(v)-1], lo)
		if e != nil {
			return InvalidAigLit, e
		}
		c, e := rec(v[:len(v)-1], hi)
		if e != nil {
			return InvalidAigLit, e
		}
		return m.Mux(v[len(v)-1], c, a)
	}
	return rec(args, truth)
}

func (n *NWKNetwork) Strash(ctx context.Context) (*AigManager, []AigLit, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, nil, err
	}
	if len(n.latches) > 0 {
		return nil, nil, ErrUnsupported
	}
	if err := b.use(0, len(n.objects)*4); err != nil {
		return nil, nil, err
	}
	m := NewAigManager()
	mapping := make([]AigLit, len(n.objects))
	for i := range mapping {
		mapping[i] = InvalidAigLit
	}
	for _, id := range n.cis {
		l, err := m.CreateCI()
		if err != nil {
			return nil, nil, err
		}
		mapping[id] = l
		if m.ObjectCount() > n.limits.MaxObjects {
			return nil, nil, ErrLimit
		}
	}
	for _, o := range n.objects {
		if o.Type != NWKObjectNode {
			continue
		}
		if err := b.use(0, len(o.Fanins)*4); err != nil {
			return nil, nil, err
		}
		args := make([]AigLit, len(o.Fanins))
		for i, f := range o.Fanins {
			args[i] = mapping[f]
			if args[i] == InvalidAigLit {
				return nil, nil, ErrInvariant
			}
		}
		l, err := nwkTruthToAIG(m, args, o.Truth, b)
		if err != nil {
			return nil, nil, err
		}
		if o.Invert {
			l ^= 1
		}
		mapping[o.ID] = l
		if m.ObjectCount() > n.limits.MaxObjects {
			return nil, nil, ErrLimit
		}
	}
	for _, id := range n.cos {
		o := n.objects[id]
		l := mapping[o.Fanins[0]]
		if o.Invert {
			l ^= 1
		}
		if _, err := m.CreateCO(l); err != nil {
			return nil, nil, err
		}
		mapping[id] = l
		if m.ObjectCount() > n.limits.MaxObjects {
			return nil, nil, ErrLimit
		}
	}
	if err := m.Check(); err != nil {
		return nil, nil, err
	}
	return m, mapping, nil
}

// NWKCut is a deterministic K-feasible cut used by the standalone mapper.
type NWKCut struct {
	Root   int
	Leaves []int
	Level  int
}

func (n *NWKNetwork) MapK(ctx context.Context, k int) ([]NWKCut, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return nil, err
	}
	if k <= 0 || k > n.limits.MaxFanins {
		return nil, ErrDimension
	}
	if n.NodeCount() > n.limits.MaxCuts {
		return nil, ErrLimit
	}
	if err := b.use(0, n.NodeCount()*32); err != nil {
		return nil, err
	}
	cuts := make([]NWKCut, 0, n.NodeCount())
	for _, o := range n.objects {
		if o.Type != NWKObjectNode {
			continue
		}
		leaves := append([]int(nil), o.Fanins...)
		if len(leaves) > k {
			return nil, ErrLimit
		}
		sort.Ints(leaves)
		if err := b.use(len(leaves)+1, len(leaves)*8); err != nil {
			return nil, err
		}
		cuts = append(cuts, NWKCut{Root: o.ID, Leaves: leaves, Level: o.Level})
	}
	return cuts, nil
}

// NWKMergeGraph is the source merge graph reduced to deterministic weighted matching.
type NWKMergeEdge struct{ A, B, Weight int }
type NWKMergeGraph struct {
	vertices int
	edges    map[[2]int]int
	limits   NWKLimits
	retained int
	closed   bool
}

func NewNWKMergeGraph(vertices int, limits NWKLimits) (*NWKMergeGraph, error) {
	l := limits.normalized()
	if vertices < 0 || vertices > l.MaxObjects {
		return nil, ErrLimit
	}
	if vertices > l.MaxBytes/8 {
		return nil, ErrLimit
	}
	return &NWKMergeGraph{vertices: vertices, edges: make(map[[2]int]int), limits: l, retained: vertices * 8}, nil
}
func (g *NWKMergeGraph) AddEdge(a, b, weight int) error {
	if g == nil || g.closed || g.edges == nil {
		return ErrInvariant
	}
	if a < 0 || b < 0 || a >= g.vertices || b >= g.vertices || a == b || weight < 0 {
		return ErrDimension
	}
	if a > b {
		a, b = b, a
	}
	k := [2]int{a, b}
	if _, ok := g.edges[k]; !ok {
		if len(g.edges) >= g.limits.MaxEdges || g.retained > g.limits.MaxBytes-32 {
			return ErrLimit
		}
		g.retained += 32
	}
	g.edges[k] = weight
	return nil
}
func (g *NWKMergeGraph) Solve(ctx context.Context) ([]NWKMergeEdge, error) {
	if g == nil || g.closed || g.edges == nil {
		return nil, ErrInvariant
	}
	b := newNWKBudget(ctx, g.limits)
	bytes := len(g.edges)*48 + g.vertices
	if err := b.use(len(g.edges), bytes); err != nil {
		return nil, err
	}
	all := make([]NWKMergeEdge, 0, len(g.edges))
	for k, w := range g.edges {
		all = append(all, NWKMergeEdge{k[0], k[1], w})
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Weight != all[j].Weight {
			return all[i].Weight < all[j].Weight
		}
		if all[i].A != all[j].A {
			return all[i].A < all[j].A
		}
		return all[i].B < all[j].B
	})
	used := make([]bool, g.vertices)
	out := make([]NWKMergeEdge, 0)
	for _, e := range all {
		if err := b.use(1, 0); err != nil {
			return nil, err
		}
		if !used[e.A] && !used[e.B] {
			used[e.A], used[e.B] = true, true
			out = append(out, e)
		}
	}
	return out, nil
}

func (g *NWKMergeGraph) Close() {
	if g == nil || g.closed {
		return
	}
	g.edges = nil
	g.vertices = 0
	g.retained = 0
	g.closed = true
}

type NWKStats struct {
	Objects, CIs, COs, Nodes, Latches, Edges, MaxFanin, Levels int
	Switching                                                  float64
}

func (n *NWKNetwork) Stats(ctx context.Context) (NWKStats, error) {
	b := newNWKBudget(ctx, n.limits)
	if err := n.validateBudget(b); err != nil {
		return NWKStats{}, err
	}
	s := NWKStats{Objects: len(n.objects), CIs: len(n.cis), COs: len(n.cos), Latches: len(n.latches), Edges: n.edges}
	for _, o := range n.objects {
		if err := b.use(1, 0); err != nil {
			return NWKStats{}, err
		}
		if o.Type == NWKObjectNode {
			s.Nodes++
		}
		if len(o.Fanins) > s.MaxFanin {
			s.MaxFanin = len(o.Fanins)
		}
		// Nwk_ManLevel is rooted at the COs; dangling internal logic does not
		// contribute to the reported network depth.
		if o.Type == NWKObjectCO && o.Level > s.Levels {
			s.Levels = o.Level
		}
		s.Switching += o.Switching
	}
	return s, nil
}
func (n *NWKNetwork) Report(ctx context.Context) (string, error) {
	s, err := n.Stats(ctx)
	if err != nil {
		return "", err
	}
	r := fmt.Sprintf("nwk %s objects=%d ci=%d co=%d nodes=%d latches=%d edges=%d levels=%d max_fanin=%d switching=%.6g", n.name, s.Objects, s.CIs, s.COs, s.Nodes, s.Latches, s.Edges, s.Levels, s.MaxFanin, s.Switching)
	if len(r) > n.limits.MaxText {
		return "", ErrLimit
	}
	return r, nil
}
func NWKGenericFileName(name string) string {
	// Nwk_FileNameGeneric preserves the directory and truncates at the last
	// dot, even when that dot begins the final path component.
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		return name[:i]
	}
	return name
}

func (n *NWKNetwork) Close() {
	if n == nil || n.closed {
		return
	}
	for i := range n.objects {
		n.objects[i].Fanins = nil
		n.objects[i].Name = ""
	}
	n.objects = nil
	n.cis = nil
	n.cos = nil
	n.latches = nil
	n.name = ""
	n.spec = ""
	n.closed = true
	n.edges = 0
	n.retained = 0
	n.partitions = false
}

// NWKBuildSources is the exact default-build inventory in module.make.
var NWKBuildSources = []string{"nwkAig.c", "nwkCheck.c", "nwkBidec.c", "nwkDfs.c", "nwkFanio.c", "nwkFlow.c", "nwkMan.c", "nwkMap.c", "nwkMerge.c", "nwkObj.c", "nwkSpeedup.c", "nwkStrash.c", "nwkTiming.c", "nwkUtil.c"}

// NWKDefinitionInventory records the audited callable-body count per source.
// nwkCheck.c and nwk_.c were already covered before this lane.
var NWKDefinitionInventory = map[string]int{"nwkAig.c": 5, "nwkBidec.c": 7, "nwkCheck.c": 1, "nwkDfs.c": 19, "nwkFanio.c": 11, "nwkFlow.c": 26, "nwkFlow_depth.c": 30, "nwkMan.c": 14, "nwkMap.c": 6, "nwkMerge.c": 26, "nwkObj.c": 7, "nwkSpeedup.c": 4, "nwkStrash.c": 3, "nwkTiming.c": 16, "nwkUtil.c": 18}

// Coverage boundary: the owned graph covers object/manager lifecycle, fanio,
// checking, DFS/support/MFFC, levels/timing, minimum-base and duplicate-fanin
// normalization, fixed-order truth strashing, direct K-cuts, statistics, and
// deterministic merge matching. Native NTL/IF/HOP/GIA/FRA/SSW integration,
// retiming max-flow internals, bidecomposition heuristics, LUT-library pin
// permutation, in-place speedup, power managers, filesystem drivers, and exact
// native allocation/pointer scratch behavior remain partial.
