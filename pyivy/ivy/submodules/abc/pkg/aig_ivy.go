package abc

// Standalone behavior from src/aig/ivy/attr.h and the remaining Ivy units.

import (
	"context"
	"fmt"
	"math/bits"
	"slices"
	"sort"
	"strings"
)

// IvyExtendedSources is the exact source list selected by ivy/module.make.
// ivyMulti8.c and ivyRwrAlg.c are tracked separately because the build does
// not select them.
var IvyExtendedSources = [...]string{
	"ivyBalance.c", "ivyCanon.c", "ivyCheck.c", "ivyCut.c", "ivyCutTrav.c",
	"ivyDfs.c", "ivyDsd.c", "ivyFanout.c", "ivyFastMap.c", "ivyFraig.c",
	"ivyHaig.c", "ivyMan.c", "ivyMem.c", "ivyMulti.c", "ivyObj.c",
	"ivyOper.c", "ivyResyn.c", "ivyRwr.c", "ivySeq.c", "ivyShow.c",
	"ivyTable.c", "ivyUtil.c",
}

// IvyExtendedCoverage records the source boundaries of this standalone layer.
// Each entry carries its own status; ivyFraig.c is fully validated while the
// surrounding Ivy engines remain partial until their individual ports finish.
var IvyExtendedCoverage = [...]string{
	"attr.h:partial:concrete integer and owned-byte attributes",
	"ivyBalance.c:partial:combinational supergate balancing",
	"ivyCanon.c:partial:truth canonicalization through input phase and permutation",
	"ivyCut.c:partial:bounded structural cuts and truth",
	"ivyCutTrav.c:partial:owned cut traversal",
	"ivyDsd.c:partial:bounded truth classification",
	"ivyFanout.c:partial:deterministic fanout snapshots",
	"ivyFastMap.c:partial:bounded cut delay and area mapping",
	"ivyFraig.c:validated:deterministic simulation classes and SAT-certified full-support reduction/miter proof",
	"ivyHaig.c:partial:owned equivalence classes",
	"ivyMulti.c:partial:owned multi-input construction",
	"ivyMulti8.c:partial:eight-input bounded construction",
	"ivyResyn.c:partial:balance and cleanup pipeline",
	"ivyRwr.c:partial:truth-preserving local replacement analysis",
	"ivyRwrAlg.c:partial:algebraic cone analysis",
	"ivySeq.c:partial:sequential evaluation and cut behavior",
	"ivyShow.c:partial:deterministic DOT text",
	"ivyTable.c:partial:structural hash profile",
	"ivyUtil.c:partial:truth, MFFC, levels, mux and statistics",
}

type IvyExtendedOptions struct {
	MaxObjects int
	MaxEdges   int
	MaxCuts    int
	MaxLeaves  int
	MaxTruth   int
	MaxBytes   int
	MaxWork    int
	MaxDepth   int
	Stop       func() bool
}

type ivyExtendedBudget struct {
	o    IvyExtendedOptions
	work int
	used int
}

const ivyAttributeMaxBytes = 64 << 20

func newIvyExtendedBudget(o IvyExtendedOptions) (*ivyExtendedBudget, error) {
	if o.MaxObjects < 0 || o.MaxEdges < 0 || o.MaxCuts < 0 || o.MaxLeaves < 0 || o.MaxTruth < 0 || o.MaxBytes < 0 || o.MaxWork < 0 || o.MaxDepth < 0 {
		return nil, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxEdges == 0 {
		o.MaxEdges = 1 << 21
	}
	if o.MaxCuts == 0 {
		o.MaxCuts = 1 << 18
	}
	if o.MaxLeaves == 0 {
		o.MaxLeaves = 16
	}
	if o.MaxTruth == 0 {
		o.MaxTruth = 1 << 16
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = 64 << 20
	}
	if o.MaxWork == 0 {
		o.MaxWork = 1 << 24
	}
	if o.MaxDepth == 0 {
		o.MaxDepth = 1 << 16
	}
	return &ivyExtendedBudget{o: o}, nil
}

func (b *ivyExtendedBudget) step(n int) error {
	if n < 0 || b.work > b.o.MaxWork-n {
		return ErrLimit
	}
	b.work += n
	if b.o.Stop != nil && b.o.Stop() {
		return ErrCancelled
	}
	return nil
}

func (b *ivyExtendedBudget) bytes(n int) error {
	if n < 0 || b.used > b.o.MaxBytes-n {
		return ErrLimit
	}
	b.used += n
	return nil
}

func (m *IvyManager) ivyExtendedValidate(b *ivyExtendedBudget) error {
	if m == nil {
		return ErrInvariant
	}
	if len(m.objects) > b.o.MaxObjects {
		return ErrLimit
	}
	if err := b.step(len(m.objects)); err != nil {
		return err
	}
	if err := m.Check(); err != nil {
		return err
	}
	for id, x := range m.objects {
		if !x.Alive || id == 0 || x.FaninCount == 0 {
			continue
		}
		// Every constructor in the owned Ivy manager connects to an older
		// object. Enforcing that invariant rejects corrupted unary cycles and
		// forward edges that the core structural checker cannot otherwise see.
		if x.Fanin0.ID() >= uint32(id) || x.FaninCount == 2 && x.Fanin1.ID() >= uint32(id) {
			return ErrInvariant
		}
	}
	return nil
}

func ivyContextStop(ctx context.Context, stop func() bool) func() bool {
	return func() bool {
		if stop != nil && stop() {
			return true
		}
		return ctx != nil && ctx.Err() != nil
	}
}

// IvyIntAttributes is the integer specialization of Attr_Man_t.
type IvyIntAttributes struct {
	values []int
	closed bool
	limit  int
}

func NewIvyIntAttributes(size, maximum int) (*IvyIntAttributes, error) {
	if size < 0 || maximum < 0 || maximum > 0 && size > maximum {
		return nil, ErrDimension
	}
	if maximum == 0 {
		maximum = 1 << 20
	}
	if size > ivyAttributeMaxBytes/8 || maximum > ivyAttributeMaxBytes/8 {
		return nil, ErrLimit
	}
	return &IvyIntAttributes{values: make([]int, size), limit: maximum}, nil
}

func (a *IvyIntAttributes) Read(id int) (int, error) {
	if a == nil || a.closed {
		return 0, ErrInvariant
	}
	if id < 0 {
		return 0, ErrDimension
	}
	if id >= len(a.values) {
		return 0, nil
	}
	return a.values[id], nil
}

func (a *IvyIntAttributes) Write(id, value int) error {
	if a == nil || a.closed {
		return ErrInvariant
	}
	if id < 0 || id >= a.limit {
		return ErrDimension
	}
	if id >= len(a.values) {
		n := id + 1
		if n < len(a.values)*2 {
			n = len(a.values) * 2
		}
		if n > a.limit {
			n = a.limit
		}
		values := make([]int, n)
		copy(values, a.values)
		a.values = values
	}
	a.values[id] = value
	return nil
}

func (a *IvyIntAttributes) Snapshot() ([]int, error) {
	if a == nil || a.closed {
		return nil, ErrInvariant
	}
	return slices.Clone(a.values), nil
}

func (a *IvyIntAttributes) Close() error {
	if a == nil || a.closed {
		return ErrInvariant
	}
	a.values = nil
	a.closed = true
	return nil
}

// IvyByteAttributes is an owned replacement for pointer attributes and the
// fixed-entry allocator. Reads and writes never expose internal storage.
type IvyByteAttributes struct {
	values [][]byte
	width  int
	limit  int
	used   int
	closed bool
}

func NewIvyByteAttributes(size, width, maximum int) (*IvyByteAttributes, error) {
	if size < 0 || width < 0 || maximum < 0 {
		return nil, ErrDimension
	}
	if maximum == 0 {
		maximum = 1 << 20
	}
	if maximum > ivyAttributeMaxBytes/24 {
		return nil, ErrLimit
	}
	if size > maximum {
		return nil, ErrLimit
	}
	if width > ivyAttributeMaxBytes || width > 0 && size > ivyAttributeMaxBytes/width {
		return nil, ErrLimit
	}
	payload := size * width
	if size > (ivyAttributeMaxBytes-payload)/24 {
		return nil, ErrLimit
	}
	a := &IvyByteAttributes{values: make([][]byte, size), width: width, limit: maximum, used: payload}
	if width > 0 {
		for i := range a.values {
			a.values[i] = make([]byte, width)
		}
	}
	return a, nil
}

func (a *IvyByteAttributes) Read(id int) ([]byte, error) {
	if a == nil || a.closed {
		return nil, ErrInvariant
	}
	if id < 0 {
		return nil, ErrDimension
	}
	if id >= len(a.values) || a.values[id] == nil {
		return nil, nil
	}
	return slices.Clone(a.values[id]), nil
}

func (a *IvyByteAttributes) Write(id int, value []byte) error {
	if a == nil || a.closed {
		return ErrInvariant
	}
	if id < 0 || id >= a.limit || a.width > 0 && len(value) != a.width {
		return ErrDimension
	}
	old := len(a.valuesAt(id))
	if len(value) > ivyAttributeMaxBytes || a.used-old > ivyAttributeMaxBytes-len(value) {
		return ErrLimit
	}
	newUsed := a.used - old + len(value)
	if id >= len(a.values) {
		n := id + 1
		if n < len(a.values)*2 {
			n = len(a.values) * 2
		}
		if n > a.limit {
			n = a.limit
		}
		if n > (ivyAttributeMaxBytes-newUsed)/24 {
			return ErrLimit
		}
		values := make([][]byte, n)
		copy(values, a.values)
		a.values = values
	}
	a.used -= len(a.values[id])
	a.values[id] = slices.Clone(value)
	a.used += len(value)
	return nil
}

func (a *IvyByteAttributes) valuesAt(id int) []byte {
	if id < 0 || id >= len(a.values) {
		return nil
	}
	return a.values[id]
}

func (a *IvyByteAttributes) Close() error {
	if a == nil || a.closed {
		return ErrInvariant
	}
	for i := range a.values {
		a.values[i] = nil
	}
	a.values = nil
	a.used = 0
	a.closed = true
	return nil
}

type IvyExtendedStats struct {
	Objects, Inputs, Outputs, Latches, ANDs, EXORs, Buffers int
	Edges, MaximumLevel, HashEntries, Dangling              int
}

func (m *IvyManager) ExtendedStats(o IvyExtendedOptions) (IvyExtendedStats, error) {
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return IvyExtendedStats{}, err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return IvyExtendedStats{}, err
	}
	if err = b.bytes(len(m.objects) * 8); err != nil {
		return IvyExtendedStats{}, err
	}
	s := IvyExtendedStats{Objects: m.LiveCount(), Inputs: len(m.inputs), Outputs: len(m.outputs), HashEntries: len(m.hash)}
	for _, x := range m.objects {
		if !x.Alive {
			continue
		}
		if err = b.step(1); err != nil {
			return IvyExtendedStats{}, err
		}
		s.Edges += int(x.FaninCount)
		if s.Edges > b.o.MaxEdges {
			return IvyExtendedStats{}, ErrLimit
		}
		if int(x.Level) > s.MaximumLevel {
			s.MaximumLevel = int(x.Level)
		}
		switch x.Type {
		case IvyLatch:
			s.Latches++
		case IvyAND:
			s.ANDs++
		case IvyEXOR:
			s.EXORs++
		case IvyBuffer:
			s.Buffers++
		}
		if x.Refs == 0 && (x.Type == IvyAND || x.Type == IvyEXOR || x.Type == IvyBuffer || x.Type == IvyLatch) {
			s.Dangling++
		}
	}
	return s, nil
}

// IvyHashProfile returns deterministic source-shaped bucket occupancies.
type IvyHashProfile struct {
	TableSize, Entries int
	Buckets            []int
}

func (m *IvyManager) HashProfile(size int, o IvyExtendedOptions) (IvyHashProfile, error) {
	if size <= 0 {
		return IvyHashProfile{}, ErrDimension
	}
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return IvyHashProfile{}, err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return IvyHashProfile{}, err
	}
	if size > b.o.MaxObjects || size > b.o.MaxBytes/8 {
		return IvyHashProfile{}, ErrLimit
	}
	if err = b.bytes(size * 8); err != nil {
		return IvyHashProfile{}, err
	}
	r := IvyHashProfile{TableSize: size, Buckets: make([]int, size)}
	for _, x := range m.objects {
		if !x.Alive || (x.Type != IvyAND && x.Type != IvyEXOR) {
			continue
		}
		if err = b.step(1); err != nil {
			return IvyHashProfile{}, err
		}
		h := (uint64(x.Fanin0)*7937 ^ uint64(x.Fanin1)*2971 ^ uint64(x.Type)*911) % uint64(size)
		r.Buckets[h]++
		r.Entries++
	}
	return r, nil
}

// CreateBuffer adds a source-shaped explicit buffer.
func (m *IvyManager) CreateBuffer(driver IvyEdge) (IvyEdge, error) {
	if err := m.ready(); err != nil {
		return 0, err
	}
	if !m.validEdge(driver) {
		return 0, ErrInvalidLiteral
	}
	if err := m.room(1); err != nil {
		return 0, err
	}
	if m.objects[driver.ID()].Refs == ^uint32(0) {
		return 0, ErrLimit
	}
	id := m.appendObject(IvyObject{Type: IvyBuffer, Fanin0: driver, FaninCount: 1, Level: m.objects[driver.ID()].Level, Phase: m.phase(driver)})
	m.objects[driver.ID()].Refs++
	return IvyEdge(id << 1), nil
}

// CreateAssert adds a checked assertion output without placing it in the PO list.
func (m *IvyManager) CreateAssert(driver IvyEdge) (uint32, error) {
	if err := m.ready(); err != nil {
		return 0, err
	}
	if !m.validEdge(driver) {
		return 0, ErrInvalidLiteral
	}
	if err := m.room(1); err != nil {
		return 0, err
	}
	if m.objects[driver.ID()].Refs == ^uint32(0) {
		return 0, ErrLimit
	}
	id := m.appendObject(IvyObject{Type: IvyAssert, Fanin0: driver, FaninCount: 1, Level: m.objects[driver.ID()].Level, Phase: m.phase(driver)})
	m.objects[driver.ID()].Refs++
	return id, nil
}

// IvyStep evaluates POs and computes latch input values in stable latch-ID order.
func (m *IvyManager) IvyStep(primary, state []bool, o IvyExtendedOptions) ([]bool, []bool, error) {
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return nil, nil, err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return nil, nil, err
	}
	if len(primary) != len(m.inputs) || len(state) != m.Count(IvyLatch) {
		return nil, nil, ErrDimension
	}
	latches := m.Count(IvyLatch)
	if err = b.bytes(len(m.objects) + len(m.outputs) + latches); err != nil {
		return nil, nil, err
	}
	values := make([]bool, len(m.objects))
	values[0] = true
	for i, id := range m.inputs {
		values[id] = primary[i]
	}
	li := 0
	read := func(e IvyEdge) bool {
		v := values[e.ID()]
		if e.Complemented() {
			return !v
		}
		return v
	}
	for id, x := range m.objects {
		if !x.Alive || id == 0 {
			continue
		}
		if err = b.step(1); err != nil {
			return nil, nil, err
		}
		switch x.Type {
		case IvyLatch:
			values[id] = state[li]
			li++
		case IvyAND:
			values[id] = read(x.Fanin0) && read(x.Fanin1)
		case IvyEXOR:
			values[id] = read(x.Fanin0) != read(x.Fanin1)
		case IvyBuffer, IvyPO, IvyAssert:
			values[id] = read(x.Fanin0)
		}
	}
	outs := make([]bool, len(m.outputs))
	for i, id := range m.outputs {
		outs[i] = values[id]
	}
	next := make([]bool, 0, latches)
	for _, x := range m.objects {
		if x.Alive && x.Type == IvyLatch {
			next = append(next, read(x.Fanin0))
		}
	}
	return outs, next, nil
}

type IvyCut struct {
	Root   uint32
	Leaves []uint32
	Truth  []uint64
}

func ivyMergeLeaves(a, c []uint32, maximum int) ([]uint32, bool) {
	r := make([]uint32, 0, len(a)+len(c))
	i, j := 0, 0
	for i < len(a) || j < len(c) {
		var v uint32
		if j == len(c) || i < len(a) && a[i] < c[j] {
			v = a[i]
			i++
		} else if i == len(a) || c[j] < a[i] {
			v = c[j]
			j++
		} else {
			v = a[i]
			i++
			j++
		}
		if len(r) == maximum {
			return nil, false
		}
		r = append(r, v)
	}
	return r, true
}

func ivyLeavesSubset(a, c []uint32) bool {
	i, j := 0, 0
	for i < len(a) && j < len(c) {
		if a[i] == c[j] {
			i++
			j++
		} else if a[i] > c[j] {
			j++
		} else {
			return false
		}
	}
	return i == len(a)
}

// EnumerateCuts computes deterministic non-dominated structural cuts.
func (m *IvyManager) EnumerateCuts(k int, o IvyExtendedOptions) (map[uint32][]IvyCut, error) {
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return nil, err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return nil, err
	}
	return m.ivyEnumerateCuts(k, b)
}

func (m *IvyManager) ivyEnumerateCuts(k int, b *ivyExtendedBudget) (map[uint32][]IvyCut, error) {
	if k <= 0 {
		return nil, ErrDimension
	}
	if k > b.o.MaxLeaves {
		return nil, ErrLimit
	}
	cuts := make(map[uint32][]IvyCut, len(m.objects))
	if len(m.objects) > b.o.MaxObjects {
		return nil, ErrLimit
	}
	if err := b.bytes(len(m.objects) * 16); err != nil {
		return nil, err
	}
	reserveCut := func(leaves int) error { return b.bytes(32 + leaves*4) }
	total := 0
	for id, x := range m.objects {
		if !x.Alive {
			continue
		}
		uid := uint32(id)
		if x.Type == IvyPI || x.Type == IvyLatch {
			if err := reserveCut(1); err != nil {
				return nil, err
			}
			cuts[uid] = []IvyCut{{Root: uid, Leaves: []uint32{uid}}}
			total++
			continue
		}
		if x.Type != IvyAND && x.Type != IvyEXOR && x.Type != IvyBuffer {
			continue
		}
		if err := b.step(1); err != nil {
			return nil, err
		}
		if err := reserveCut(1); err != nil {
			return nil, err
		}
		list := []IvyCut{{Root: uid, Leaves: []uint32{uid}}}
		if x.FaninCount == 1 {
			for _, a := range cuts[x.Fanin0.ID()] {
				if err := reserveCut(len(a.Leaves)); err != nil {
					return nil, err
				}
				list = append(list, IvyCut{Root: uid, Leaves: slices.Clone(a.Leaves)})
			}
		} else {
			for _, a := range cuts[x.Fanin0.ID()] {
				for _, c := range cuts[x.Fanin1.ID()] {
					if err := b.step(1); err != nil {
						return nil, err
					}
					leaves, ok := ivyMergeLeaves(a.Leaves, c.Leaves, k)
					if !ok {
						continue
					}
					dominated := false
					for _, old := range list {
						if ivyLeavesSubset(old.Leaves, leaves) {
							dominated = true
							break
						}
					}
					if dominated {
						continue
					}
					if err := reserveCut(len(leaves)); err != nil {
						return nil, err
					}
					keep := list[:0]
					for _, old := range list {
						if !ivyLeavesSubset(leaves, old.Leaves) {
							keep = append(keep, old)
						}
					}
					list = append(keep, IvyCut{Root: uid, Leaves: leaves})
				}
			}
		}
		sort.Slice(list, func(i, j int) bool {
			if len(list[i].Leaves) != len(list[j].Leaves) {
				return len(list[i].Leaves) < len(list[j].Leaves)
			}
			return slices.Compare(list[i].Leaves, list[j].Leaves) < 0
		})
		total += len(list)
		if total > b.o.MaxCuts {
			return nil, ErrLimit
		}
		cuts[uid] = list
	}
	return cuts, nil
}

// CutTruth computes root truth over ordered leaves. Other terminals are rejected.
func (m *IvyManager) CutTruth(root IvyEdge, leaves []uint32, o IvyExtendedOptions) ([]uint64, error) {
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return nil, err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return nil, err
	}
	return m.ivyCutTruth(root, leaves, b)
}

func (m *IvyManager) ivyCutTruth(root IvyEdge, leaves []uint32, b *ivyExtendedBudget) ([]uint64, error) {
	if !m.validEdge(root) || len(leaves) > b.o.MaxLeaves || len(leaves) >= bits.UintSize {
		return nil, ErrDimension
	}
	assignments := 1 << uint(len(leaves))
	if assignments > b.o.MaxTruth {
		return nil, ErrLimit
	}
	words := (assignments + 63) / 64
	maxInt := int(^uint(0) >> 1)
	if len(m.objects) > maxInt/6 || len(leaves) > maxInt/32 {
		return nil, ErrLimit
	}
	if err := b.bytes(words*8 + len(m.objects)*6 + len(leaves)*32); err != nil {
		return nil, err
	}
	leaf := make(map[uint32]int, len(leaves))
	for i, id := range leaves {
		if uint64(id) >= uint64(len(m.objects)) || !m.validEdge(IvyEdge(id<<1)) {
			return nil, ErrInvalidLiteral
		}
		if _, ok := leaf[id]; ok {
			return nil, ErrDimension
		}
		leaf[id] = i
	}
	needed := make([]bool, len(m.objects))
	stack := []uint32{root.ID()}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if needed[id] {
			continue
		}
		if _, ok := leaf[id]; ok {
			continue
		}
		needed[id] = true
		x := m.objects[id]
		if x.Type == IvyPI || x.Type == IvyLatch {
			if _, ok := leaf[id]; !ok && id != 0 {
				return nil, ErrDimension
			}
			continue
		}
		if x.FaninCount >= 1 {
			stack = append(stack, x.Fanin0.ID())
		}
		if x.FaninCount == 2 {
			stack = append(stack, x.Fanin1.ID())
		}
	}
	truth := make([]uint64, words)
	values := make([]bool, len(m.objects))
	for a := 0; a < assignments; a++ {
		values[0] = true
		for id, pos := range leaf {
			values[id] = a&(1<<uint(pos)) != 0
		}
		read := func(e IvyEdge) bool {
			v := values[e.ID()]
			if e.Complemented() {
				v = !v
			}
			return v
		}
		for id, x := range m.objects {
			if !x.Alive || id == 0 || !needed[id] {
				continue
			}
			if err := b.step(1); err != nil {
				return nil, err
			}
			if _, ok := leaf[uint32(id)]; ok {
				continue
			}
			switch x.Type {
			case IvyPI, IvyLatch:
				return nil, ErrDimension
			case IvyAND:
				values[id] = read(x.Fanin0) && read(x.Fanin1)
			case IvyEXOR:
				values[id] = read(x.Fanin0) != read(x.Fanin1)
			case IvyBuffer:
				values[id] = read(x.Fanin0)
			}
		}
		if read(root) {
			truth[a/64] |= 1 << uint(a%64)
		}
	}
	return truth, nil
}

type IvyMFFC struct {
	Root            uint32
	Nodes, Boundary []uint32
}

func (m *IvyManager) MFFC(root uint32, o IvyExtendedOptions) (IvyMFFC, error) {
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return IvyMFFC{}, err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return IvyMFFC{}, err
	}
	if uint64(root) >= uint64(len(m.objects)) || !m.objects[root].Alive {
		return IvyMFFC{}, ErrInvalidLiteral
	}
	if len(m.objects) > b.o.MaxObjects {
		return IvyMFFC{}, ErrLimit
	}
	if err = b.bytes(len(m.objects) * 2); err != nil {
		return IvyMFFC{}, err
	}
	inside := make([]bool, len(m.objects))
	boundary := make([]bool, len(m.objects))
	stack := []uint32{root}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if inside[id] {
			continue
		}
		if err = b.step(1); err != nil {
			return IvyMFFC{}, err
		}
		x := m.objects[id]
		if id != root && (x.Refs > 1 || x.Type == IvyPI || x.Type == IvyLatch) {
			boundary[id] = true
			continue
		}
		if x.Type != IvyAND && x.Type != IvyEXOR && x.Type != IvyBuffer {
			return IvyMFFC{}, ErrDimension
		}
		inside[id] = true
		if x.FaninCount >= 1 {
			stack = append(stack, x.Fanin0.ID())
		}
		if x.FaninCount == 2 {
			stack = append(stack, x.Fanin1.ID())
		}
	}
	r := IvyMFFC{Root: root}
	for id := range inside {
		if inside[id] {
			r.Nodes = append(r.Nodes, uint32(id))
		}
		if boundary[id] {
			r.Boundary = append(r.Boundary, uint32(id))
		}
	}
	return r, nil
}

type IvyMapNode struct {
	Root        uint32
	Leaves      []uint32
	Delay, Area int
}
type IvyMapping struct {
	Nodes       []IvyMapNode
	Delay, Area int
}

// FastMap chooses one deterministic k-cut per reachable logic node.
func (m *IvyManager) FastMap(k int, o IvyExtendedOptions) (IvyMapping, error) {
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return IvyMapping{}, err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return IvyMapping{}, err
	}
	cuts, err := m.ivyEnumerateCuts(k, b)
	if err != nil {
		return IvyMapping{}, err
	}
	if len(m.objects) > b.o.MaxObjects || len(m.objects) > b.o.MaxBytes/8 {
		return IvyMapping{}, ErrLimit
	}
	if err = b.bytes(len(m.objects) * 8); err != nil {
		return IvyMapping{}, err
	}
	for _, list := range cuts {
		for _, cut := range list {
			if err = b.step(1 + len(cut.Leaves)); err != nil {
				return IvyMapping{}, err
			}
		}
	}
	delay := make([]int, len(m.objects))
	result := IvyMapping{}
	for id, x := range m.objects {
		if !x.Alive || (x.Type != IvyAND && x.Type != IvyEXOR && x.Type != IvyBuffer) {
			continue
		}
		if err = b.step(1); err != nil {
			return IvyMapping{}, err
		}
		best := IvyMapNode{Root: uint32(id), Delay: int(^uint(0) >> 1), Area: int(^uint(0) >> 1)}
		for _, c := range cuts[uint32(id)] {
			if len(c.Leaves) == 1 && c.Leaves[0] == uint32(id) {
				continue
			}
			d := 0
			for _, leaf := range c.Leaves {
				if delay[leaf] > d {
					d = delay[leaf]
				}
			}
			d++
			area := len(c.Leaves)
			if d < best.Delay || d == best.Delay && (area < best.Area || area == best.Area && slices.Compare(c.Leaves, best.Leaves) < 0) {
				best = IvyMapNode{Root: uint32(id), Leaves: slices.Clone(c.Leaves), Delay: d, Area: area}
			}
		}
		if best.Leaves == nil {
			return IvyMapping{}, ErrInvariant
		}
		if err = b.bytes(32 + len(best.Leaves)*4); err != nil {
			return IvyMapping{}, err
		}
		delay[id] = best.Delay
		result.Nodes = append(result.Nodes, best)
		result.Area++
		if best.Delay > result.Delay {
			result.Delay = best.Delay
		}
	}
	return result, nil
}

type IvyChoice struct {
	Representative uint32
	Equivalents    []IvyEdge
}
type IvyChoices struct {
	classes []IvyChoice
	objects int
	used    int
}

func NewIvyChoices(m *IvyManager) (*IvyChoices, error) {
	if m == nil {
		return nil, ErrInvariant
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	return &IvyChoices{objects: len(m.objects)}, nil
}
func (c *IvyChoices) Add(m *IvyManager, representative uint32, equivalents []IvyEdge, o IvyExtendedOptions) error {
	if c == nil || m == nil || c.objects != len(m.objects) || len(equivalents) == 0 {
		return ErrDimension
	}
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return err
	}
	return c.ivyAdd(m, representative, equivalents, b)
}

func (c *IvyChoices) ivyAdd(m *IvyManager, representative uint32, equivalents []IvyEdge, b *ivyExtendedBudget) error {
	if uint64(representative) >= uint64(len(m.objects)) || !m.validEdge(IvyEdge(representative<<1)) {
		return ErrInvalidLiteral
	}
	seen := map[IvyEdge]bool{}
	if len(equivalents) > (b.o.MaxBytes-b.used)/40 {
		return ErrLimit
	}
	if err := b.bytes(len(equivalents) * 40); err != nil {
		return err
	}
	copyEdges := make([]IvyEdge, 0, len(equivalents))
	for _, e := range equivalents {
		if !m.validEdge(e) {
			return ErrInvalidLiteral
		}
		if seen[e] {
			continue
		}
		seen[e] = true
		copyEdges = append(copyEdges, e)
		if err := b.step(1); err != nil {
			return err
		}
	}
	for _, cl := range c.classes {
		if err := b.step(1); err != nil {
			return err
		}
		if cl.Representative == representative {
			return ErrInvariant
		}
	}
	retained := 32 + len(copyEdges)*8
	if len(c.classes) >= b.o.MaxObjects || retained < 0 || c.used > b.o.MaxBytes-retained {
		return ErrLimit
	}
	c.classes = append(c.classes, IvyChoice{Representative: representative, Equivalents: copyEdges})
	c.used += retained
	return nil
}

// Equivalent performs bounded exact equivalence over all PIs and latch values.
func (m *IvyManager) Equivalent(a, c IvyEdge, o IvyExtendedOptions) (bool, error) {
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return false, err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return false, err
	}
	return m.ivyEquivalent(a, c, b)
}

func (m *IvyManager) ivyEquivalent(a, c IvyEdge, b *ivyExtendedBudget) (bool, error) {
	if !m.validEdge(a) || !m.validEdge(c) {
		return false, ErrInvalidLiteral
	}
	leaves := slices.Clone(m.inputs)
	if err := b.bytes(len(m.inputs) * 4); err != nil {
		return false, err
	}
	for id, x := range m.objects {
		if err := b.step(1); err != nil {
			return false, err
		}
		if x.Alive && x.Type == IvyLatch {
			if err := b.bytes(4); err != nil {
				return false, err
			}
			leaves = append(leaves, uint32(id))
		}
	}
	ta, err := m.ivyCutTruth(a, leaves, b)
	if err != nil {
		return false, err
	}
	tc, err := m.ivyCutTruth(c, leaves, b)
	if err != nil {
		return false, err
	}
	return slices.Equal(ta, tc), nil
}

// AddVerified admits a choice class only after exact bounded equivalence.
func (c *IvyChoices) AddVerified(m *IvyManager, representative IvyEdge, equivalents []IvyEdge, o IvyExtendedOptions) error {
	if c == nil || m == nil || c.objects != len(m.objects) || len(equivalents) == 0 {
		return ErrDimension
	}
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return err
	}
	for _, edge := range equivalents {
		equal, err := m.ivyEquivalent(representative, edge, b)
		if err != nil {
			return err
		}
		if !equal {
			return ErrInvariant
		}
	}
	return c.ivyAdd(m, representative.ID(), equivalents, b)
}
func (c *IvyChoices) Snapshot() []IvyChoice {
	if c == nil {
		return nil
	}
	r := make([]IvyChoice, len(c.classes))
	for i, x := range c.classes {
		r[i] = IvyChoice{Representative: x.Representative, Equivalents: slices.Clone(x.Equivalents)}
	}
	return r
}

// IvyCanonicalTruth minimizes a truth table across input negations and permutations.
type IvyCanonicalTruthResult struct {
	Truth            uint64
	Permutation      []int
	InputPhase       uint64
	OutputComplement bool
}

func IvyCanonicalTruth(truth uint64, variables int, o IvyExtendedOptions) (IvyCanonicalTruthResult, error) {
	if variables < 0 || variables > 6 {
		return IvyCanonicalTruthResult{}, ErrDimension
	}
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return IvyCanonicalTruthResult{}, err
	}
	n := 1 << uint(variables)
	if err = b.bytes(variables * 16); err != nil {
		return IvyCanonicalTruthResult{}, err
	}
	mask := uint64(1) << uint(n)
	if n == 64 {
		mask = ^uint64(0)
	} else {
		mask--
	}
	truth &= mask
	perm := make([]int, variables)
	for i := range perm {
		perm[i] = i
	}
	best := IvyCanonicalTruthResult{Truth: ^uint64(0)}
	var visit func(int) error
	visit = func(pos int) error {
		if pos == variables {
			for phase := 0; phase < (1 << uint(variables)); phase++ {
				if err = b.step(n); err != nil {
					return err
				}
				var t uint64
				for a := 0; a < n; a++ {
					source := 0
					for j, p := range perm {
						bit := (a >> uint(j)) & 1
						bit ^= (phase >> uint(j)) & 1
						source |= bit << uint(p)
					}
					if truth&(1<<uint(source)) != 0 {
						t |= 1 << uint(a)
					}
				}
				out := false
				if (^t)&mask < t {
					t = (^t) & mask
					out = true
				}
				if t < best.Truth {
					best = IvyCanonicalTruthResult{Truth: t, Permutation: slices.Clone(perm), InputPhase: uint64(phase), OutputComplement: out}
				}
			}
			return nil
		}
		for i := pos; i < variables; i++ {
			perm[pos], perm[i] = perm[i], perm[pos]
			if err := visit(pos + 1); err != nil {
				return err
			}
			perm[pos], perm[i] = perm[i], perm[pos]
		}
		return nil
	}
	if err = visit(0); err != nil {
		return IvyCanonicalTruthResult{}, err
	}
	return best, nil
}

// BalanceCombinational rebuilds a combinational manager with balanced AND cones.
func (m *IvyManager) BalanceCombinational(o IvyExtendedOptions) (*IvyManager, error) {
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return nil, err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return nil, err
	}
	if len(m.objects) > b.o.MaxObjects {
		return nil, ErrLimit
	}
	if err = b.bytes(len(m.objects) * 16); err != nil {
		return nil, err
	}
	if m.Count(IvyLatch) > 0 || m.Count(IvyAssert) > 0 {
		return nil, ErrDimension
	}
	n, err := NewIvyManager(m.options)
	if err != nil {
		return nil, err
	}
	memo := make([]IvyEdge, len(m.objects))
	memo[0] = IvyConst1
	ready := make([]bool, len(m.objects))
	ready[0] = true
	accountGrowth := func(before int) error {
		growth := len(n.objects) - before
		if growth < 0 || len(n.objects) > b.o.MaxObjects {
			return ErrLimit
		}
		if err := b.step(growth); err != nil {
			return err
		}
		return b.bytes(growth * 64)
	}
	for _, id := range m.inputs {
		before := len(n.objects)
		v, e := n.CreatePI()
		if e != nil {
			return nil, e
		}
		if e = accountGrowth(before); e != nil {
			return nil, e
		}
		memo[id] = v
		ready[id] = true
	}
	var translate func(IvyEdge, int) (IvyEdge, error)
	translate = func(e IvyEdge, depth int) (IvyEdge, error) {
		if depth > b.o.MaxDepth {
			return 0, ErrLimit
		}
		id := e.ID()
		if ready[id] {
			return memo[id].NotIf(e.Complemented()), nil
		}
		if err := b.step(1); err != nil {
			return 0, err
		}
		x := m.objects[id]
		var r IvyEdge
		switch x.Type {
		case IvyAND:
			leaves := []IvyEdge{}
			var collect func(IvyEdge, int) error
			collect = func(q IvyEdge, d int) error {
				if d > b.o.MaxDepth {
					return ErrLimit
				}
				z := m.objects[q.ID()]
				if !q.Complemented() && z.Type == IvyAND && z.Refs <= 1 {
					if err := collect(z.Fanin0, d+1); err != nil {
						return err
					}
					return collect(z.Fanin1, d+1)
				}
				v, er := translate(q, d+1)
				if er != nil {
					return er
				}
				leaves = append(leaves, v)
				return nil
			}
			if er := collect(x.Fanin0, depth+1); er != nil {
				return 0, er
			}
			if er := collect(x.Fanin1, depth+1); er != nil {
				return 0, er
			}
			for len(leaves) > 1 {
				sort.SliceStable(leaves, func(i, j int) bool {
					oi, _ := n.Object(leaves[i].ID())
					oj, _ := n.Object(leaves[j].ID())
					if oi.Level != oj.Level {
						return oi.Level < oj.Level
					}
					return leaves[i] < leaves[j]
				})
				before := len(n.objects)
				v, er := n.And(leaves[0], leaves[1])
				if er != nil {
					return 0, er
				}
				if er = accountGrowth(before); er != nil {
					return 0, er
				}
				leaves = append(leaves[2:], v)
			}
			r = leaves[0]
		case IvyEXOR:
			a, er := translate(x.Fanin0, depth+1)
			if er != nil {
				return 0, er
			}
			c, er := translate(x.Fanin1, depth+1)
			if er != nil {
				return 0, er
			}
			before := len(n.objects)
			r, er = n.Xor(a, c)
			if er != nil {
				return 0, er
			}
			if er = accountGrowth(before); er != nil {
				return 0, er
			}
		case IvyBuffer:
			a, er := translate(x.Fanin0, depth+1)
			if er != nil {
				return 0, er
			}
			before := len(n.objects)
			r, er = n.CreateBuffer(a)
			if er != nil {
				return 0, er
			}
			if er = accountGrowth(before); er != nil {
				return 0, er
			}
		default:
			return 0, ErrDimension
		}
		memo[id] = r
		ready[id] = true
		return r.NotIf(e.Complemented()), nil
	}
	for _, oid := range m.outputs {
		x := m.objects[oid]
		v, e := translate(x.Fanin0, 0)
		if e != nil {
			return nil, e
		}
		before := len(n.objects)
		if _, e = n.CreatePO(v); e != nil {
			return nil, e
		}
		if e = accountGrowth(before); e != nil {
			return nil, e
		}
	}
	if err = n.Check(); err != nil {
		return nil, err
	}
	return n, nil
}

// Resynthesize applies the source-shaped safe subset: balance then cleanup.
func (m *IvyManager) Resynthesize(o IvyExtendedOptions) (*IvyManager, error) {
	n, err := m.BalanceCombinational(o)
	if err != nil {
		return nil, err
	}
	if _, err = n.Cleanup(); err != nil {
		return nil, err
	}
	return n, n.Check()
}

func (m *IvyManager) FormatDOT(bold []uint32, o IvyExtendedOptions) (string, error) {
	b, err := newIvyExtendedBudget(o)
	if err != nil {
		return "", err
	}
	if err = m.ivyExtendedValidate(b); err != nil {
		return "", err
	}
	if len(m.objects) > b.o.MaxObjects {
		return "", ErrLimit
	}
	mark := make([]bool, len(m.objects))
	for _, id := range bold {
		if uint64(id) >= uint64(len(mark)) || !m.objects[id].Alive {
			return "", ErrInvalidLiteral
		}
		mark[id] = true
	}
	var s strings.Builder
	write := func(x string) error {
		if b.used > b.o.MaxBytes-len(x) {
			return ErrLimit
		}
		b.used += len(x)
		s.WriteString(x)
		return nil
	}
	if err = write("digraph ivy {\n"); err != nil {
		return "", err
	}
	edges := 0
	for id, x := range m.objects {
		if !x.Alive {
			continue
		}
		if err = b.step(1); err != nil {
			return "", err
		}
		shape := "ellipse"
		if x.Type == IvyPI {
			shape = "box"
		} else if x.Type == IvyPO {
			shape = "triangle"
		} else if x.Type == IvyLatch {
			shape = "diamond"
		}
		style := ""
		if mark[id] {
			style = ",style=bold"
		}
		if err = write(fmt.Sprintf("  n%d [label=\"%d:%d\",shape=%s%s];\n", id, id, x.Type, shape, style)); err != nil {
			return "", err
		}
		for k, e := range []IvyEdge{x.Fanin0, x.Fanin1} {
			if k >= int(x.FaninCount) {
				break
			}
			edges++
			if edges > b.o.MaxEdges {
				return "", ErrLimit
			}
			dash := ""
			if e.Complemented() {
				dash = ",style=dashed"
			}
			if err = write(fmt.Sprintf("  n%d -> n%d [label=\"%d\"%s];\n", e.ID(), id, k, dash)); err != nil {
				return "", err
			}
		}
	}
	if err = write("}\n"); err != nil {
		return "", err
	}
	return s.String(), nil
}
