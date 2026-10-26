package abc

// DAG-aware AIG rewriting from src/opt/rwr/rwr.h, rwrDec.c, rwrEva.c,
// rwrExp.c, rwrLib.c, rwrMan.c, rwrPrint.c, rwrUtil.c, and module.make.

import (
	"encoding/binary"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"time"
)

const rwrClassCount = 222

type RwrOptions struct {
	Precompute                          bool
	MaxNodes, MaxClasses, MaxOperations int
	Stop                                func() bool
}
type RwrNode struct {
	RwtNode
	Score, Gain, Added int16
}
type RwrStats struct {
	Considered, Added, Classes                                                       int
	NodesConsidered, NodesRewritten, NodesGained, NodesBeginning, NodesEnd           int
	CutsGood, CutsBad, Subgraphs                                                     int
	Scores                                                                           [rwrClassCount]int
	Start, CutTime, ResynthesisTime, EvaluationTime, MFFCTime, UpdateTime, TotalTime time.Duration
}
type RwrManager struct {
	core             *RwtManager
	nodes            []RwrNode
	stats            RwrStats
	result           *DecoGraph
	resultLeaves     []AigLit
	resultComplement bool
	options          RwrOptions
	closed           bool
}

func rwrNormalizeOptions(o RwrOptions) (RwrOptions, error) {
	if o.MaxNodes < 0 || o.MaxClasses < 0 || o.MaxOperations < 0 {
		return o, ErrDimension
	}
	if o.MaxNodes == 0 {
		o.MaxNodes = (1 << 18) + 5
	}
	if o.MaxClasses == 0 {
		o.MaxClasses = rwrClassCount
	}
	if o.MaxOperations == 0 {
		o.MaxOperations = 1 << 24
	}
	if o.MaxClasses < rwrClassCount {
		return o, ErrLimit
	}
	return o, nil
}
func (m *RwrManager) check() error {
	if m == nil || m.closed || m.core == nil {
		return ErrInvariant
	}
	return nil
}
func (m *RwrManager) stopped() bool { return m.options.Stop != nil && m.options.Stop() }
func (m *RwrManager) syncNodes() {
	base, _ := m.core.Nodes()
	old := m.nodes
	m.nodes = make([]RwrNode, len(base))
	for i, n := range base {
		m.nodes[i].RwtNode = n
		if i < len(old) {
			m.nodes[i].Score = old[i].Score
			m.nodes[i].Gain = old[i].Gain
			m.nodes[i].Added = old[i].Added
		}
	}
}

func NewRwrManager(o RwrOptions) (*RwrManager, error) {
	o, e := rwrNormalizeOptions(o)
	if e != nil {
		return nil, e
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	start := time.Now()
	core, e := NewRwtManager(RwtOptions{Precompute: o.Precompute, MaxNodes: o.MaxNodes, MaxClasses: o.MaxClasses, Stop: o.Stop})
	if e != nil {
		return nil, e
	}
	m := &RwrManager{core: core, options: o}
	m.syncNodes()
	s, _ := core.Stats()
	m.stats.Considered = s.Considered
	m.stats.Added = s.Added
	m.stats.Classes = s.Classes
	if o.Precompute {
		if e = m.Precompute(); e != nil {
			m.Close()
			return nil, e
		}
	}
	m.stats.Start = time.Since(start)
	return m, nil
}

func newRwrEmptyManager(o RwrOptions) (*RwrManager, error) {
	o, e := rwrNormalizeOptions(o)
	if e != nil {
		return nil, e
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	// Precompute mode in the RWT foundation creates exactly the five source
	// variables without loading the compiled-in forest.
	core, e := NewRwtManager(RwtOptions{Precompute: true, MaxNodes: o.MaxNodes, MaxClasses: o.MaxClasses, Stop: o.Stop})
	if e != nil {
		return nil, e
	}
	m := &RwrManager{core: core, options: o}
	m.syncNodes()
	s, _ := core.Stats()
	m.stats.Considered, m.stats.Added, m.stats.Classes = s.Considered, s.Added, s.Classes
	return m, nil
}

// NewRwrManagerFromBinary reconstructs the native rwrUtil.c file format into
// a fresh forest and preprocesses it into evaluator-ready source order.
func NewRwrManagerFromBinary(data []byte, maxEntries int, o RwrOptions) (*RwrManager, error) {
	start := time.Now()
	m, err := newRwrEmptyManager(o)
	if err != nil {
		return nil, err
	}
	if err = m.LoadBinaryLibrary(data, maxEntries); err == nil {
		err = m.Preprocess()
	}
	if err != nil {
		_ = m.Close()
		return nil, err
	}
	m.stats.Start = time.Since(start)
	return m, nil
}
func (m *RwrManager) Nodes() ([]RwrNode, error) {
	if e := m.check(); e != nil {
		return nil, e
	}
	return slices.Clone(m.nodes), nil
}
func (m *RwrManager) Classes() ([][]int, error) {
	if e := m.check(); e != nil {
		return nil, e
	}
	return m.core.Classes()
}
func (m *RwrManager) Practical(t uint16) (bool, error) {
	if e := m.check(); e != nil {
		return false, e
	}
	return m.core.Practical(t)
}
func (m *RwrManager) Canonical(t uint16) (uint16, uint8, uint8, uint8, error) {
	if e := m.check(); e != nil {
		return 0, 0, 0, 0, e
	}
	return m.core.Canonical(t)
}
func (m *RwrManager) Decomposition(id int) (*DecoGraph, error) {
	if e := m.check(); e != nil {
		return nil, e
	}
	return m.core.Decomposition(id)
}
func (m *RwrManager) Stats() (RwrStats, error) {
	if e := m.check(); e != nil {
		return RwrStats{}, e
	}
	return m.stats, nil
}
func (m *RwrManager) AddVariable(t uint16, p bool) (int, error) {
	if e := m.check(); e != nil {
		return -1, e
	}
	id, e := m.core.AddVariable(t, p)
	if e == nil {
		m.syncNodes()
	}
	return id, e
}
func (m *RwrManager) AddNode(a, b RwtLink, x bool, l, v int) (int, error) {
	if e := m.check(); e != nil {
		return -1, e
	}
	before := len(m.core.table)
	_ = before
	id, e := m.core.AddNode(a, b, x, l, v)
	if e == nil {
		m.syncNodes()
		m.stats.Considered++
		n := m.nodes[id]
		if n.Truth == m.core.canonical[n.Truth] {
			m.stats.Added++
			if len(m.core.table[n.Truth]) == 1 {
				m.stats.Classes++
			}
		}
	}
	return id, e
}
func (m *RwrManager) NodeVolume(a, b int) (int, error) {
	if e := m.check(); e != nil {
		return 0, e
	}
	return m.core.NodeVolume(a, b)
}
func (m *RwrManager) IncrementTraversalID() error {
	if e := m.check(); e != nil {
		return e
	}
	return m.core.IncrementTraversalID()
}
func (m *RwrManager) PracticalTable() ([]bool, error) {
	if e := m.check(); e != nil {
		return nil, e
	}
	return m.core.PracticalTable()
}
func (m *RwrManager) LoadStatic() error {
	if e := m.check(); e != nil {
		return e
	}
	if e := m.core.LoadStatic(); e != nil {
		return e
	}
	m.syncNodes()
	s, _ := m.core.Stats()
	m.stats.Considered = s.Considered
	m.stats.Added = s.Added
	m.stats.Classes = s.Classes
	return nil
}
func (m *RwrManager) Preprocess() error {
	if e := m.check(); e != nil {
		return e
	}
	if e := m.core.Preprocess(); e != nil {
		return e
	}
	m.syncNodes()
	return nil
}
func (m *RwrManager) TryNode(a, b RwtLink, x bool, l, v int) (int, bool, error) {
	if e := m.check(); e != nil {
		return -1, false, e
	}
	if l < 0 || l >= 64 || v < 0 || v >= 256 {
		return -1, false, ErrDimension
	}
	if m.stopped() {
		return -1, false, ErrCancelled
	}
	t0, e := m.core.truthOf(a)
	if e != nil {
		return -1, false, e
	}
	t1, e := m.core.truthOf(b)
	if e != nil {
		return -1, false, e
	}
	t := t0 & t1
	if x {
		if a.Complement || b.Complement {
			return -1, false, ErrInvariant
		}
		t = t0 ^ t1
	}
	m.stats.Considered++
	if l > 2 && !m.core.practical[m.core.canonical[t]] {
		return -1, false, nil
	}
	for _, id := range m.core.table[t] {
		n := m.core.nodes[id]
		if n.Level < l && n.Volume < v || n.Level == l && n.Volume < v {
			return -1, false, nil
		}
	}
	if len(m.core.nodes) >= m.options.MaxNodes {
		return -1, false, ErrLimit
	}
	id := len(m.core.nodes)
	m.core.nodes = append(m.core.nodes, RwtNode{ID: id, Truth: t, Level: l, Volume: v, XOR: x, Fan0: a, Fan1: b})
	if len(m.core.table[t]) == 0 && m.core.canonical[t] == t {
		m.stats.Classes++
	}
	m.core.table[t] = append(m.core.table[t], id)
	m.syncNodes()
	return id, true, nil
}
func (m *RwrManager) markUsed(id int) {
	n := &m.core.nodes[id]
	if n.Used || n.TraversalID == m.core.traversalID {
		return
	}
	n.TraversalID = m.core.traversalID
	n.Used = true
	m.markUsed(n.Fan0.Node)
	m.markUsed(n.Fan1.Node)
}
func (m *RwrManager) Precompute() error {
	if e := m.check(); e != nil {
		return e
	}
	if m.core.classes != nil {
		return ErrInvariant
	}
	bn := slices.Clone(m.core.nodes)
	bt := make([][]int, len(m.core.table))
	for i := range bt {
		bt[i] = slices.Clone(m.core.table[i])
	}
	bs := m.stats
	btid := m.core.traversalID
	fail := func(e error) error {
		m.core.nodes, m.core.table, m.core.traversalID, m.stats = bn, bt, btid, bs
		m.syncNodes()
		return e
	}
	ops := 0
	for i := 1; i < len(m.core.nodes); i++ {
		p0 := m.core.nodes[i]
		for k := 1; k < len(m.core.nodes); k++ {
			if k == i {
				break
			}
			p1 := m.core.nodes[k]
			if p0.Level+p1.Level > 5 {
				break
			}
			l := 1 + max(p0.Level, p1.Level)
			v, e := m.core.NodeVolume(i, k)
			if e != nil {
				return fail(e)
			}
			v++
			cases := []struct {
				a, b RwtLink
				x    bool
				v    int
			}{{RwtLink{Node: i}, RwtLink{Node: k}, false, v}, {RwtLink{Node: i, Complement: true}, RwtLink{Node: k}, false, v}, {RwtLink{Node: i}, RwtLink{Node: k, Complement: true}, false, v}, {RwtLink{Node: i, Complement: true}, RwtLink{Node: k, Complement: true}, false, v}, {RwtLink{Node: i}, RwtLink{Node: k}, true, v + 1}}
			for _, c := range cases {
				ops++
				if ops > m.options.MaxOperations {
					return fail(ErrLimit)
				}
				if m.stopped() {
					return fail(ErrCancelled)
				}
				_, _, e = m.TryNode(c.a, c.b, c.x, l, c.v)
				if e != nil {
					return fail(e)
				}
				if len(m.core.nodes) == m.options.MaxNodes {
					goto compact
				}
			}
		}
	}
compact:
	m.core.incrementTraversal()
	for id := 5; id < len(m.core.nodes); id++ {
		if m.core.nodes[id].Truth == m.core.canonical[m.core.nodes[id].Truth] {
			m.markUsed(id)
		}
	}
	remap := make([]int, len(m.core.nodes))
	for i := range remap {
		remap[i] = -1
	}
	kept := make([]RwtNode, 0, len(m.core.nodes))
	for i, n := range m.core.nodes {
		if i < 5 || n.Used {
			remap[i] = len(kept)
			n.ID = len(kept)
			kept = append(kept, n)
		}
	}
	for i := 5; i < len(kept); i++ {
		kept[i].Fan0.Node = remap[kept[i].Fan0.Node]
		kept[i].Fan1.Node = remap[kept[i].Fan1.Node]
	}
	m.core.nodes = kept
	m.core.table = make([][]int, 1<<16)
	for i, n := range kept {
		m.core.table[n.Truth] = append(m.core.table[n.Truth], i)
	}
	m.syncNodes()
	return nil
}

func (m *RwrManager) StaticArray() ([]uint16, error) {
	if e := m.check(); e != nil {
		return nil, e
	}
	out := make([]uint16, 0, 2*(len(m.nodes)-4))
	for _, n := range m.nodes[5:] {
		if n.Fan0.Node < 0 || n.Fan1.Node < 0 || n.Fan0.Node > math.MaxUint16>>2 || n.Fan1.Node > math.MaxUint16>>1 {
			return nil, ErrLimit
		}
		a := uint16(n.Fan0.Node << 1)
		if n.Fan0.Complement {
			a |= 1
		}
		a <<= 1
		if n.XOR {
			a |= 1
		}
		b := uint16(n.Fan1.Node << 1)
		if n.Fan1.Complement {
			b |= 1
		}
		out = append(out, a, b)
	}
	return append(out, 0, 0), nil
}
func (m *RwrManager) BinaryLibrary() ([]byte, error) {
	w, e := m.StaticArray()
	if e != nil {
		return nil, e
	}
	count := len(w)/2 - 1
	out := make([]byte, 4+8*count)
	binary.LittleEndian.PutUint32(out, uint32(count))
	for i := 0; i < 2*count; i++ {
		binary.LittleEndian.PutUint32(out[4+4*i:], uint32(w[i]))
	}
	return out, nil
}
func DecodeRwrBinaryLibrary(data []byte, maxEntries int) ([]uint16, error) {
	if maxEntries < 0 {
		return nil, ErrDimension
	}
	if len(data) < 4 {
		return nil, ErrFormat
	}
	n := int(binary.LittleEndian.Uint32(data))
	if maxEntries == 0 {
		maxEntries = 1 << 18
	}
	if n > maxEntries {
		return nil, ErrLimit
	}
	if n > (len(data)-4)/8 || len(data) != 4+8*n {
		return nil, ErrFormat
	}
	out := make([]uint16, 0, 2*n+2)
	for i := 0; i < 2*n; i++ {
		v := binary.LittleEndian.Uint32(data[4+4*i:])
		if v > math.MaxUint16 {
			return nil, ErrFormat
		}
		out = append(out, uint16(v))
	}
	return append(out, 0, 0), nil
}

func (m *RwrManager) LoadBinaryLibrary(data []byte, maxEntries int) error {
	if e := m.check(); e != nil {
		return e
	}
	words, e := DecodeRwrBinaryLibrary(data, maxEntries)
	if e != nil {
		return e
	}
	// Validate every serialized back-reference before considering manager
	// state. This keeps corrupt input distinguishable from a valid library
	// presented to a manager which has already been preprocessed.
	available := 5
	for i := 0; i+1 < len(words); i += 2 {
		a, b := uint32(words[i]), uint32(words[i+1])
		if a == 0 && b == 0 {
			break
		}
		if int(a>>2) >= available || int(b>>1) >= available {
			return ErrFormat
		}
		available++
	}
	// File indexes are relative to the five elementary nodes, and loading
	// after preprocessing would leave the class/decomposition indexes stale.
	if len(m.core.nodes) != 5 || m.core.classes != nil || m.core.graphs != nil || m.core.stats.Considered != 0 {
		return ErrInvariant
	}
	bn := slices.Clone(m.core.nodes)
	bt := make([][]int, len(m.core.table))
	for i := range bt {
		bt[i] = slices.Clone(m.core.table[i])
	}
	bs := m.stats
	btid := m.core.traversalID
	fail := func(e error) error {
		m.core.nodes, m.core.table, m.core.traversalID, m.stats = bn, bt, btid, bs
		m.syncNodes()
		return e
	}
	for i := 0; i+1 < len(words); i += 2 {
		a, b := uint32(words[i]), uint32(words[i+1])
		if a == 0 && b == 0 {
			break
		}
		x := a&1 != 0
		a >>= 1
		l0, l1 := RwtLink{Node: int(a >> 1), Complement: a&1 != 0}, RwtLink{Node: int(b >> 1), Complement: b&1 != 0}
		if l0.Node >= len(m.core.nodes) || l1.Node >= len(m.core.nodes) {
			return fail(ErrFormat)
		}
		level := 1 + max(m.core.nodes[l0.Node].Level, m.core.nodes[l1.Node].Level)
		volume, e := m.core.NodeVolume(l0.Node, l1.Node)
		if e != nil {
			return fail(e)
		}
		if _, e = m.AddNode(l0, l1, x, level, 1+volume+rwtBoolToInt(x)); e != nil {
			return fail(e)
		}
	}
	return nil
}

func (m *RwrManager) StaticArrayText() ([]byte, error) {
	words, e := m.StaticArray()
	if e != nil {
		return nil, e
	}
	var b strings.Builder
	b.WriteString("static unsigned short s_RwtAigSubgraphs[] = \n{")
	for i := 0; i < len(words)/2; i++ {
		if i%5 == 0 {
			b.WriteString("\n    ")
		}
		fmt.Fprintf(&b, "0x%04x,0x%04x, ", words[2*i], words[2*i+1])
	}
	b.WriteString("\n};\n")
	return []byte(b.String()), nil
}

func (m *RwrManager) expression(id int) string {
	n := m.nodes[id]
	if id == 0 {
		return "Const1"
	}
	if id < 5 {
		return string(rune('a' + id - 1))
	}
	child := func(l RwtLink) string {
		s := m.expression(l.Node)
		if l.Node >= 5 {
			s = "(" + s + ")"
		}
		if l.Complement {
			s += "'"
		}
		return s
	}
	op := ""
	if n.XOR {
		op = "+"
	}
	return child(n.Fan0) + op + child(n.Fan1)
}
func (m *RwrManager) NodeText(id int) (string, error) {
	if e := m.check(); e != nil {
		return "", e
	}
	if id < 0 || id >= len(m.nodes) {
		return "", ErrDimension
	}
	n := m.nodes[id]
	return fmt.Sprintf("%5d : %04x tt=%016b lev=%d vol=%d  %s\n", id, n.Truth, n.Truth, n.Level, n.Volume, m.expression(id)), nil
}
func (m *RwrManager) bush(entry int) (int, int, int, error) {
	if entry < 0 || entry >= 1<<16 {
		return 0, 0, 0, ErrDimension
	}
	unique := map[int]bool{}
	sum, roots := 0, 0
	var visit func(int, map[int]bool) int
	visit = func(id int, seen map[int]bool) int {
		n := m.nodes[id]
		if n.Used || seen[id] {
			return 0
		}
		seen[id] = true
		return 1 + visit(n.Fan0.Node, seen) + visit(n.Fan1.Node, seen)
	}
	for _, id := range m.core.table[entry] {
		if m.nodes[id].Truth != m.core.canonical[m.nodes[id].Truth] {
			continue
		}
		roots++
		sum += visit(id, map[int]bool{})
		var mark func(int)
		mark = func(x int) {
			n := m.nodes[x]
			if n.Used || unique[x] {
				return
			}
			unique[x] = true
			mark(n.Fan0.Node)
			mark(n.Fan1.Node)
		}
		mark(id)
	}
	return len(unique), roots, sum, nil
}
func (m *RwrManager) BushMetrics(entry int) (int, int, int, error) {
	if e := m.check(); e != nil {
		return 0, 0, 0, e
	}
	return m.bush(entry)
}
func (m *RwrManager) LibraryText() (string, error) {
	if e := m.check(); e != nil {
		return "", e
	}
	var b strings.Builder
	counter := 0
	for t := 0; t < 1<<16; t++ {
		if len(m.core.table[t]) == 0 || m.core.canonical[t] != uint16(t) {
			continue
		}
		v, r, s, _ := m.bush(t)
		fmt.Fprintf(&b, "\nClass %3d. Func %6d.  Roots = %3d. Vol = %3d. Sum = %3d.  %016b\n", m.core.classMap[t], counter, r, v, s, t)
		counter++
		for _, id := range m.core.table[t] {
			if m.nodes[id].Truth == m.core.canonical[m.nodes[id].Truth] {
				line, _ := m.NodeText(id)
				b.WriteString(line)
			}
		}
	}
	return b.String(), nil
}

func (m *RwrManager) ScoreClean() error {
	if e := m.check(); e != nil {
		return e
	}
	for i := range m.nodes {
		m.nodes[i].Score = 0
		m.nodes[i].Gain = 0
		m.nodes[i].Added = 0
	}
	return nil
}

type RwrClassScore struct {
	Class, TotalGain int
	Truth            uint16
	Nodes            []RwrNode
}

func (m *RwrManager) ScoreReport() ([]RwrClassScore, error) {
	if e := m.check(); e != nil {
		return nil, e
	}
	classes, _ := m.core.Classes()
	out := make([]RwrClassScore, 0)
	for c, ids := range classes {
		g := 0
		ns := make([]RwrNode, 0)
		for _, id := range ids {
			g += int(m.nodes[id].Gain)
			if m.nodes[id].Score != 0 {
				ns = append(ns, m.nodes[id])
			}
		}
		if g != 0 {
			t, _ := m.core.ClassRepresentative(c)
			out = append(out, RwrClassScore{c, g, t, ns})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].TotalGain != out[j].TotalGain {
			return out[i].TotalGain > out[j].TotalGain
		}
		return out[i].Class < out[j].Class
	})
	return out, nil
}
func (m *RwrManager) StatsText() (string, error) {
	if e := m.check(); e != nil {
		return "", e
	}
	used := 0
	for _, v := range m.stats.Scores {
		if v > 0 {
			used++
		}
	}
	gain := m.stats.NodesBeginning - m.stats.NodesEnd
	pct := 0.0
	if m.stats.NodesBeginning != 0 {
		pct = 100 * float64(gain) / float64(m.stats.NodesBeginning)
	}
	s := m.stats
	return fmt.Sprintf("Rewriting statistics:\nTotal cuts tries  = %8d.\nBad cuts found    = %8d.\nTotal subgraphs   = %8d.\nUsed NPN classes  = %8d.\nNodes considered  = %8d.\nNodes rewritten   = %8d.\nGain              = %8d. (%6.2f %%).\nStart        =%9.2f sec\nCuts         =%9.2f sec\nResynthesis  =%9.2f sec\n    Mffc     =%9.2f sec\n    Eval     =%9.2f sec\nUpdate       =%9.2f sec\nTOTAL        =%9.2f sec\n\n", s.CutsGood, s.CutsBad, s.Subgraphs, used, s.NodesConsidered, s.NodesRewritten, gain, pct, s.Start.Seconds(), s.CutTime.Seconds(), s.ResynthesisTime.Seconds(), s.MFFCTime.Seconds(), s.EvaluationTime.Seconds(), s.UpdateTime.Seconds(), s.TotalTime.Seconds()), nil
}
func (m *RwrManager) StatsRecord() ([]byte, error) {
	if e := m.check(); e != nil {
		return nil, e
	}
	s := m.stats
	return []byte(fmt.Sprintf("%d %d %d %d\n", s.CutsGood, s.Subgraphs, s.NodesRewritten, s.NodesGained)), nil
}
func (m *RwrManager) AddCutTime(d time.Duration) error {
	if e := m.check(); e != nil {
		return e
	}
	m.stats.CutTime += d
	return nil
}
func (m *RwrManager) AddUpdateTime(d time.Duration) error {
	if e := m.check(); e != nil {
		return e
	}
	m.stats.UpdateTime += d
	return nil
}
func (m *RwrManager) AddTotalTime(d time.Duration) error {
	if e := m.check(); e != nil {
		return e
	}
	m.stats.TotalTime += d
	return nil
}
func (m *RwrManager) Result() (*DecoGraph, []AigLit, bool, error) {
	if e := m.check(); e != nil {
		return nil, nil, false, e
	}
	var g *DecoGraph
	if m.result != nil {
		var e error
		g, e = m.result.Clone()
		if e != nil {
			return nil, nil, false, e
		}
	}
	return g, slices.Clone(m.resultLeaves), m.resultComplement, nil
}
func (m *RwrManager) Close() error {
	if e := m.check(); e != nil {
		return e
	}
	if m.result != nil {
		m.result.Close()
	}
	e := m.core.Close()
	m.core = nil
	m.nodes = nil
	m.result = nil
	m.resultLeaves = nil
	m.closed = true
	return e
}
func RwrPrecompute(o RwrOptions) error {
	o.Precompute = true
	m, e := NewRwrManager(o)
	if e != nil {
		return e
	}
	return m.Close()
}

type RwrExplorerReport struct {
	Cuts            int64
	Classes         int
	ZeroClasses     int
	Distribution    map[uint64]int
	Representatives map[uint64]uint32
	Dataset         []RewriteNPNClass
}
type RwrExplorer4 struct {
	canonical []uint16
	counts    []uint64
	closed    bool
}

func NewRwrExplorer4() (*RwrExplorer4, error) {
	d, e := NewDecManager()
	if e != nil {
		return nil, e
	}
	c, _, _, _, e := d.NPN4Tables()
	d.Close()
	if e != nil {
		return nil, e
	}
	return &RwrExplorer4{canonical: c, counts: make([]uint64, 1<<16)}, nil
}
func (x *RwrExplorer4) Count(t uint16) error {
	if x == nil || x.closed {
		return ErrInvariant
	}
	c := x.canonical[t]
	if x.counts[c] == math.MaxUint64 {
		return ErrLimit
	}
	x.counts[c]++
	return nil
}
func (x *RwrExplorer4) Report() (RwrExplorerReport, error) {
	if x == nil || x.closed {
		return RwrExplorerReport{}, ErrInvariant
	}
	r := RwrExplorerReport{Distribution: map[uint64]int{}, Representatives: map[uint64]uint32{}}
	for t, c := range x.counts {
		if c == 0 {
			continue
		}
		if c > math.MaxInt64-uint64(r.Cuts) {
			return RwrExplorerReport{}, ErrLimit
		}
		r.Cuts += int64(c)
		r.Classes++
		r.Distribution[c]++
		r.Representatives[c] = uint32(t)
		r.Dataset = append(r.Dataset, RewriteNPNClass{Truth: uint32(t), Frequency: int(c)})
	}
	r.ZeroClasses = 2288 - r.Classes
	return r, nil
}
func (x *RwrExplorer4) Close() error {
	if x == nil || x.closed {
		return ErrInvariant
	}
	x.canonical = nil
	x.counts = nil
	x.closed = true
	return nil
}

type RwrExplorer5 struct {
	counts map[uint32]uint64
	closed bool
}
type RwrExplorer5Report struct {
	NN, NPN      []RewriteNPNClass
	Cuts         int64
	Distribution map[uint64]int
}

func NewRwrExplorer5() *RwrExplorer5 { return &RwrExplorer5{counts: map[uint32]uint64{}} }
func (x *RwrExplorer5) Count(t uint32) error {
	if x == nil || x.closed {
		return ErrInvariant
	}
	if x.counts[t] == math.MaxUint64 {
		return ErrLimit
	}
	x.counts[t]++
	return nil
}
func (x *RwrExplorer5) Report() (RwrExplorer5Report, error) {
	if x == nil || x.closed {
		return RwrExplorer5Report{}, ErrInvariant
	}
	r := RwrExplorer5Report{Distribution: map[uint64]int{}}
	keys := make([]uint32, 0, len(x.counts))
	for k := range x.counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	npn := map[uint32]uint64{}
	for _, k := range keys {
		c := x.counts[k]
		r.Cuts += int64(c)
		r.Distribution[c]++
		r.NN = append(r.NN, RewriteNPNClass{k, int(c)})
		canon, e := ExtraCanonicalTruth32(k, 5, ExtraTruthCanonNPN)
		if e != nil {
			return RwrExplorer5Report{}, e
		}
		npn[canon] += c
	}
	keys = keys[:0]
	for k := range npn {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		r.NPN = append(r.NPN, RewriteNPNClass{k, int(npn[k])})
	}
	return r, nil
}
func (x *RwrExplorer5) Close() error {
	if x == nil || x.closed {
		return ErrInvariant
	}
	x.counts = nil
	x.closed = true
	return nil
}

type RwrRewriteOptions struct {
	UpdateLevel, UseZeroGain, PlacementEnable bool
	LevelMax, MaxVisits, MaxCandidates        int
	Stop                                      func() bool
}
type RwrRewriteResult struct {
	Gain        int
	Graph       *DecoGraph
	Leaves      []AigLit
	Complement  bool
	Truth       uint16
	LibraryNode int
}

func rwrApplyEdge(values []AigLit, e DecoEdge) AigLit { return values[e.Node].NotIf(e.Complement) }
func rwrNetworkCount(source *AigManager, graph *DecoGraph, leaves []AigLit, members []bool, nodeMax, levelMax int) (int, error) {
	if source == nil || graph == nil || len(leaves) != 4 {
		return -1, ErrDimension
	}
	work := source.Clone()
	for id, in := range members {
		if in && id < len(work.objects) {
			o := work.objects[id]
			if o.Type == AigAND {
				delete(work.hash, aigHashKey{o.Fanin0, o.Fanin1})
			}
		}
	}
	values := make([]AigLit, len(graph.nodes))
	copy(values, leaves)
	created := 0
	for i := graph.leaves; i < len(graph.nodes); i++ {
		n := graph.nodes[i]
		a, b := rwrApplyEdge(values, n.Edge0), rwrApplyEdge(values, n.Edge1)
		before := len(work.objects)
		lit, e := work.And(a, b)
		if e != nil {
			return -1, e
		}
		if len(work.objects) > before {
			created++
			if nodeMax >= 0 && created > nodeMax {
				return -1, nil
			}
		}
		if levelMax >= 0 && int(work.objects[lit.ID()].Level) > levelMax {
			return -1, nil
		}
		values[i] = lit
	}
	if !graph.constant {
		root := rwrApplyEdge(values, graph.root)
		if levelMax >= 0 && int(work.objects[root.ID()].Level) > levelMax {
			return -1, nil
		}
	}
	return created, nil
}
func (m *RwrManager) evaluateCut(source *AigManager, root AigID, cut AigCNFCut4, fanins []AigLit, levelMax, maxCandidates, maxVisits int, stop func() bool) (*DecoGraph, int, int, error) {
	class := int(m.core.classMap[cut.Truth])
	ids := m.core.classes[class]
	m.stats.Subgraphs += len(ids)
	if maxCandidates > 0 && len(ids) > maxCandidates {
		return nil, -1, -1, ErrLimit
	}
	leafIDs := make([]AigID, len(fanins))
	for i, l := range fanins {
		leafIDs[i] = l.ID()
	}
	mffc, e := source.MFFCLabelCut(root, leafIDs, AigMFFCOptions{MaxVisits: maxVisits, Stop: stop})
	if e != nil {
		return nil, -1, -1, e
	}
	bestGain, bestID := -1, -1
	var best *DecoGraph
	for _, id := range ids {
		if stop != nil && stop() {
			return nil, -1, -1, ErrCancelled
		}
		g := m.core.graphs[id]
		if g == nil {
			continue
		}
		added, e := rwrNetworkCount(source, g, fanins, mffc.Members, mffc.Size, levelMax)
		if e != nil {
			return nil, -1, -1, e
		}
		if added < 0 {
			continue
		}
		gain := mffc.Size - added
		if gain > bestGain {
			bestGain, bestID, best = gain, id, g
			if gain > 0 {
				n := &m.nodes[id]
				n.Score++
				n.Gain = int16(int(n.Gain) + gain)
				n.Added = int16(int(n.Added) + added)
			}
		}
	}
	if best == nil {
		return nil, -1, -1, nil
	}
	clone, e := best.Clone()
	return clone, bestGain, bestID, e
}

func (m *RwrManager) RewriteNode(source *AigManager, root AigID, cuts []AigCNFCut4, o RwrRewriteOptions) (RwrRewriteResult, error) {
	if e := m.check(); e != nil {
		return RwrRewriteResult{}, e
	}
	if source == nil || o.LevelMax < 0 || o.MaxVisits < 0 || o.MaxCandidates < 0 {
		return RwrRewriteResult{}, ErrDimension
	}
	if e := source.Check(); e != nil {
		return RwrRewriteResult{}, e
	}
	if uint64(root) >= uint64(len(source.objects)) || source.objects[root].Type != AigAND {
		return RwrRewriteResult{}, ErrInvalidLiteral
	}
	oldStats, oldNodes := m.stats, slices.Clone(m.nodes)
	oldResult, oldLeaves, oldCompl := m.result, slices.Clone(m.resultLeaves), m.resultComplement
	best := RwrRewriteResult{Gain: -1, LibraryNode: -1}
	fail := func(e error) (RwrRewriteResult, error) {
		if best.Graph != nil {
			best.Graph.Close()
			best.Graph = nil
		}
		m.stats = oldStats
		m.nodes = oldNodes
		m.result = oldResult
		m.resultLeaves = oldLeaves
		m.resultComplement = oldCompl
		return RwrRewriteResult{}, e
	}
	m.stats.NodesConsidered++
	levelMax := o.LevelMax
	if !o.UpdateLevel || levelMax == 0 {
		levelMax = math.MaxInt
	}
	maxVisits := o.MaxVisits
	if maxVisits == 0 {
		maxVisits = 1 << 20
	}
	perms := decPermutations(4)
	for _, cut := range cuts {
		if o.Stop != nil && o.Stop() {
			return fail(ErrCancelled)
		}
		if len(cut.Leaves) > 4 {
			return fail(ErrDimension)
		}
		if cut.Root != root || len(cut.Leaves) < 4 {
			continue
		}
		_, phase, perm, _, e := m.core.Canonical(cut.Truth)
		if e != nil {
			return fail(e)
		}
		fanins := make([]AigLit, 4)
		valid := true
		for i := 0; i < 4; i++ {
			id := cut.Leaves[perms[perm][i]]
			if uint64(id) >= uint64(len(source.objects)) || source.objects[id].Type == AigNone {
				valid = false
				break
			}
			fanins[i] = (AigLit(id) << 1).NotIf(phase&(1<<uint(i)) != 0)
		}
		if !valid {
			m.stats.CutsBad++
			continue
		}
		m.stats.CutsGood++
		single := 0
		for _, l := range fanins {
			if source.objects[l.ID()].RefCount == 1 {
				single++
			}
		}
		if single > 2 {
			continue
		}
		g, gain, id, e := m.evaluateCut(source, root, cut, fanins, levelMax, o.MaxCandidates, maxVisits, o.Stop)
		if e != nil {
			return fail(e)
		}
		if g != nil && gain > best.Gain {
			if best.Graph != nil {
				best.Graph.Close()
			}
			best = RwrRewriteResult{gain, g, fanins, phase&16 != 0, cut.Truth, id}
		}
	}
	if best.Gain < 0 {
		return best, nil
	}
	owned, e := best.Graph.Clone()
	if e != nil {
		return fail(e)
	}
	if oldResult != nil {
		oldResult.Close()
	}
	m.result = best.Graph
	m.resultLeaves = slices.Clone(best.Leaves)
	m.resultComplement = best.Complement
	class := m.core.classMap[best.Truth]
	m.stats.Scores[class]++
	m.stats.NodesGained += best.Gain
	if o.UseZeroGain || best.Gain > 0 {
		m.stats.NodesRewritten++
	}
	best.Graph = owned
	return best, nil
}

func RwrCutIsBoolean(source *AigManager, root AigID, leaves []AigID) (bool, error) {
	if source == nil {
		return false, ErrDimension
	}
	if uint64(root) >= uint64(len(source.objects)) || source.objects[root].Type != AigAND {
		return false, ErrInvalidLiteral
	}
	marks := map[AigID]int{}
	for _, id := range leaves {
		if uint64(id) >= uint64(len(source.objects)) || source.objects[id].Type == AigNone {
			return false, ErrInvalidLiteral
		}
		marks[id] = 0
	}
	var walk func(AigID, int) error
	walk = func(id AigID, side int) error {
		if _, ok := marks[id]; ok {
			marks[id] |= side
			return nil
		}
		if uint64(id) >= uint64(len(source.objects)) || source.objects[id].Type != AigAND {
			return ErrDimension
		}
		n := source.objects[id]
		if e := walk(n.Fanin0.ID(), side); e != nil {
			return e
		}
		return walk(n.Fanin1.ID(), side)
	}
	n := source.objects[root]
	if e := walk(n.Fanin0.ID(), 1); e != nil {
		return false, e
	}
	if e := walk(n.Fanin1.ID(), 2); e != nil {
		return false, e
	}
	for _, v := range marks {
		if v == 3 {
			return true, nil
		}
	}
	return false, nil
}
func RwrCutNodeCount(source *AigManager, root AigID, cuts []AigCNFCut4) (int, error) {
	if source == nil {
		return 0, ErrDimension
	}
	seen := map[AigID]bool{}
	var walk func(AigID, map[AigID]bool) error
	walk = func(id AigID, b map[AigID]bool) error {
		if b[id] {
			seen[id] = true
			return nil
		}
		if uint64(id) >= uint64(len(source.objects)) || source.objects[id].Type != AigAND {
			return ErrDimension
		}
		// Even when this internal node was collected for an earlier cut, its
		// fanins must be revisited because the current cut has a new boundary.
		// The C code uses fMarkC only for collection, never for recursion.
		seen[id] = true
		n := source.objects[id]
		if e := walk(n.Fanin0.ID(), b); e != nil {
			return e
		}
		return walk(n.Fanin1.ID(), b)
	}
	for _, c := range cuts {
		b := map[AigID]bool{}
		for _, id := range c.Leaves {
			if uint64(id) >= uint64(len(source.objects)) || source.objects[id].Type == AigNone {
				return 0, ErrInvalidLiteral
			}
			b[id] = true
		}
		if e := walk(root, b); e != nil {
			return 0, e
		}
	}
	return len(seen), nil
}
func RwrNodeDepth(source *AigManager, root AigID, leaves []AigID) (int, error) {
	if source == nil {
		return 0, ErrDimension
	}
	b := map[AigID]bool{}
	for _, id := range leaves {
		if uint64(id) >= uint64(len(source.objects)) || source.objects[id].Type == AigNone {
			return 0, ErrInvalidLiteral
		}
		b[id] = true
	}
	var walk func(AigID) (int, error)
	walk = func(id AigID) (int, error) {
		if uint64(id) >= uint64(len(source.objects)) {
			return 0, ErrInvalidLiteral
		}
		if b[id] || source.objects[id].Type == AigCI {
			return 0, nil
		}
		if uint64(id) >= uint64(len(source.objects)) || source.objects[id].Type != AigAND {
			return 0, ErrDimension
		}
		n := source.objects[id]
		a, e := walk(n.Fanin0.ID())
		if e != nil {
			return 0, e
		}
		c, e := walk(n.Fanin1.ID())
		return 1 + max(a, c), e
	}
	return walk(root)
}
