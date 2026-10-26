package abc

// Ported from src/opt/fxu/fxuInt.h, fxuHeapD.c, fxuHeapS.c, fxuList.c,
// fxuMatrix.c, fxuPair.c, fxuSelect.c, and fxuSingle.c.

import (
	"sort"
	"strconv"
	"strings"
)

// FXUOptions bounds the owned sparse kernel. Zero values select defaults.
type FXUOptions struct {
	MaxVariables  int
	MaxCubes      int
	MaxEntries    int
	MaxDivisors   int
	MaxOperations int
	Stop          func() bool
}

type FXULiteralRef struct {
	Cube     int
	Variable int
}

type FXUVariable struct {
	ID          int
	Occurrences []int
}

type FXUCube struct {
	ID       int
	Output   int
	Local    int
	Literals []int
	Active   bool
}

type FXUPair struct {
	Cube1 int
	Cube2 int
	Base  int
	Lits1 int
	Lits2 int
}

type FXUSingle struct {
	ID      int
	Var1    int
	Var2    int
	Weight  int
	HeapPos int
}

type FXUDouble struct {
	ID      int
	Key     uint32
	Left    []int
	Right   []int
	Pairs   []FXUPair
	Weight  int
	HeapPos int
}

type FXUSelection struct {
	Weight   int
	SingleID int
	DoubleID int
	Var1     int
	Var2     int
}

type fxuHeap struct {
	items []int
	pos   []int
}

// FXUMatrix replaces pointer-linked rows, columns, hash chains and heaps with
// stable integer IDs and owned ordered slices.
type FXUMatrix struct {
	options     FXUOptions
	variables   []FXUVariable
	cubes       []FXUCube
	entries     int
	singles     []FXUSingle
	doubles     []FXUDouble
	doubleByKey map[string]int
	singleHeap  fxuHeap
	doubleHeap  fxuHeap
	operations  int
	closed      bool
}

func NewFXUMatrix(options FXUOptions) *FXUMatrix {
	if options.MaxVariables <= 0 {
		options.MaxVariables = 1 << 16
	}
	if options.MaxCubes <= 0 {
		options.MaxCubes = 1 << 18
	}
	if options.MaxEntries <= 0 {
		options.MaxEntries = 1 << 22
	}
	if options.MaxDivisors <= 0 {
		options.MaxDivisors = 1 << 20
	}
	if options.MaxOperations <= 0 {
		options.MaxOperations = 1 << 27
	}
	return &FXUMatrix{options: options, doubleByKey: make(map[string]int)}
}

func (m *FXUMatrix) tick(n int) error {
	if m == nil || m.closed {
		return ErrInvariant
	}
	if n < 0 || m.operations > m.options.MaxOperations-n {
		return ErrLimit
	}
	if m.options.Stop != nil && m.options.Stop() {
		return ErrCancelled
	}
	m.operations += n
	return nil
}

func (m *FXUMatrix) AddVariable() (int, error) {
	beforeOps := m.operations
	if err := m.tick(1); err != nil {
		return -1, err
	}
	if len(m.variables) >= m.options.MaxVariables {
		m.operations = beforeOps
		return -1, ErrLimit
	}
	id := len(m.variables)
	m.variables = append(m.variables, FXUVariable{ID: id})
	return id, nil
}

func (m *FXUMatrix) AddCube(output, local int) (int, error) {
	beforeOps := m.operations
	if err := m.tick(1); err != nil {
		return -1, err
	}
	if output < 0 || output >= len(m.variables) || local < 0 {
		m.operations = beforeOps
		return -1, ErrDimension
	}
	if len(m.cubes) >= m.options.MaxCubes {
		m.operations = beforeOps
		return -1, ErrLimit
	}
	for i := range m.cubes {
		if m.cubes[i].Active && m.cubes[i].Output == output && m.cubes[i].Local == local {
			m.operations = beforeOps
			return -1, ErrInvariant
		}
	}
	id := len(m.cubes)
	m.cubes = append(m.cubes, FXUCube{ID: id, Output: output, Local: local, Active: true})
	return id, nil
}

func sortedInsertUnique(values []int, value int) ([]int, bool) {
	i := sort.SearchInts(values, value)
	if i < len(values) && values[i] == value {
		return values, false
	}
	values = append(values, 0)
	copy(values[i+1:], values[i:])
	values[i] = value
	return values, true
}

func sortedDelete(values []int, value int) ([]int, bool) {
	i := sort.SearchInts(values, value)
	if i == len(values) || values[i] != value {
		return values, false
	}
	copy(values[i:], values[i+1:])
	return values[:len(values)-1], true
}

func (m *FXUMatrix) AddLiteral(cube, variable int) error {
	beforeOps := m.operations
	if err := m.tick(1); err != nil {
		return err
	}
	if cube < 0 || cube >= len(m.cubes) || !m.cubes[cube].Active || variable < 0 || variable >= len(m.variables) {
		m.operations = beforeOps
		return ErrDimension
	}
	if m.entries >= m.options.MaxEntries {
		m.operations = beforeOps
		return ErrLimit
	}
	updated, added := sortedInsertUnique(m.cubes[cube].Literals, variable)
	if !added {
		m.operations = beforeOps
		return ErrInvariant
	}
	m.invalidateCubeDerived(cube)
	m.cubes[cube].Literals = updated
	m.variables[variable].Occurrences, _ = sortedInsertUnique(m.variables[variable].Occurrences, cube)
	m.entries++
	return nil
}

func (m *FXUMatrix) DeleteLiteral(cube, variable int) error {
	beforeOps := m.operations
	if err := m.tick(1); err != nil {
		return err
	}
	if cube < 0 || cube >= len(m.cubes) || variable < 0 || variable >= len(m.variables) {
		m.operations = beforeOps
		return ErrDimension
	}
	updated, found := sortedDelete(m.cubes[cube].Literals, variable)
	if !found {
		m.operations = beforeOps
		return ErrInvariant
	}
	m.invalidateCubeDerived(cube)
	m.cubes[cube].Literals = updated
	m.variables[variable].Occurrences, _ = sortedDelete(m.variables[variable].Occurrences, cube)
	m.entries--
	return nil
}

func (m *FXUMatrix) Coincidence(var1, var2 int) (int, error) {
	beforeOps := m.operations
	if err := m.tick(1); err != nil {
		return 0, err
	}
	if var1 < 0 || var2 < 0 || var1 >= len(m.variables) || var2 >= len(m.variables) {
		m.operations = beforeOps
		return 0, ErrDimension
	}
	a, b := m.variables[var1].Occurrences, m.variables[var2].Occurrences
	if err := m.tick(len(a) + len(b)); err != nil {
		m.operations = beforeOps
		return 0, err
	}
	i, j, count := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			count++
			i++
			j++
		} else if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return count, nil
}

func (m *FXUMatrix) rebuildSingleHeap() {
	m.singleHeap = fxuHeap{}
	for i := range m.singles {
		m.singles[i].HeapPos = -1
		m.singleHeap.insertSingle(m, i)
	}
}
func (m *FXUMatrix) rebuildDoubleHeap() {
	m.doubleHeap = fxuHeap{}
	for i := range m.doubles {
		m.doubles[i].HeapPos = -1
		if len(m.doubles[i].Pairs) != 0 {
			m.doubleHeap.insertDouble(m, i)
		}
	}
}

func (m *FXUMatrix) invalidateCubeDerived(cube int) {
	m.singles = nil
	m.singleHeap = fxuHeap{}
	for i := range m.doubles {
		d := &m.doubles[i]
		kept := d.Pairs[:0]
		for _, pair := range d.Pairs {
			if pair.Cube1 == cube || pair.Cube2 == cube {
				d.Weight -= pair.Lits1 + pair.Lits2 - 1 + pair.Base
				continue
			}
			kept = append(kept, pair)
		}
		d.Pairs = kept
		if len(d.Pairs) == 0 {
			delete(m.doubleByKey, fxuDivisorKey(d.Left, d.Right))
		}
	}
	m.rebuildDoubleHeap()
}

func (m *FXUMatrix) ComputeSingles(useZero bool, maximum int) error {
	startOps := m.operations
	if err := m.tick(1); err != nil {
		return err
	}
	candidates := make([]FXUSingle, 0)
	limit := 1
	if useZero {
		limit = 0
	}
	for b := 1; b < len(m.variables); b++ {
		for a := 0; a < b; a++ {
			if err := m.tick(1 + len(m.variables[a].Occurrences) + len(m.variables[b].Occurrences)); err != nil {
				m.operations = startOps
				return err
			}
			coincidence := intersectCount(m.variables[a].Occurrences, m.variables[b].Occurrences)
			weight := coincidence - 2
			if coincidence > 0 && weight >= limit {
				candidates = append(candidates, FXUSingle{Var1: a, Var2: b, Weight: weight})
			}
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Weight != candidates[j].Weight {
			return candidates[i].Weight > candidates[j].Weight
		}
		if candidates[i].Var1 != candidates[j].Var1 {
			return candidates[i].Var1 < candidates[j].Var1
		}
		return candidates[i].Var2 < candidates[j].Var2
	})
	if maximum >= 0 && len(candidates) > maximum {
		candidates = candidates[:maximum]
	}
	if len(candidates) > m.options.MaxDivisors {
		candidates = candidates[:m.options.MaxDivisors]
	}
	for i := range candidates {
		candidates[i].ID = i
	}
	m.singles = candidates
	m.rebuildSingleHeap()
	return nil
}

func intersectCount(a, b []int) int {
	i, j, n := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			n++
			i++
			j++
		} else if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return n
}

func residues(a, b []int) (base, left, right []int) {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			base = append(base, a[i])
			i++
			j++
		} else if a[i] < b[j] {
			left = append(left, a[i])
			i++
		} else {
			right = append(right, b[j])
			j++
		}
	}
	left = append(left, a[i:]...)
	right = append(right, b[j:]...)
	return
}

func compareInts(a, b []int) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func fxuDivisorKey(left, right []int) string {
	var b strings.Builder
	for _, v := range left {
		b.WriteString(strconv.Itoa(v))
		b.WriteByte(',')
	}
	b.WriteByte('|')
	for _, v := range right {
		b.WriteString(strconv.Itoa(v))
		b.WriteByte(',')
	}
	return b.String()
}

func FXUPairHash(left, right []int) uint32 {
	h := uint32(2166136261)
	for _, side := range [][]int{left, right} {
		for _, v := range side {
			h ^= uint32(v + 1)
			h *= 16777619
		}
		h ^= 0xff
		h *= 16777619
	}
	return h
}

func (m *FXUMatrix) AddDouble(cube1, cube2 int) (int, error) {
	beforeOps := m.operations
	if err := m.tick(1); err != nil {
		return -1, err
	}
	if cube1 < 0 || cube2 < 0 || cube1 >= len(m.cubes) || cube2 >= len(m.cubes) || cube1 == cube2 || !m.cubes[cube1].Active || !m.cubes[cube2].Active {
		m.operations = beforeOps
		return -1, ErrDimension
	}
	if m.cubes[cube1].Output != m.cubes[cube2].Output {
		m.operations = beforeOps
		return -1, ErrInvariant
	}
	if err := m.tick(len(m.cubes[cube1].Literals) + len(m.cubes[cube2].Literals)); err != nil {
		m.operations = beforeOps
		return -1, err
	}
	base, left, right := residues(m.cubes[cube1].Literals, m.cubes[cube2].Literals)
	if len(left) == 0 || len(right) == 0 {
		m.operations = beforeOps
		return -1, ErrInvariant
	}
	if compareInts(left, right) > 0 || (compareInts(left, right) == 0 && cube1 > cube2) {
		left, right = right, left
		cube1, cube2 = cube2, cube1
	}
	key := fxuDivisorKey(left, right)
	id, found := m.doubleByKey[key]
	if !found {
		if len(m.doubles) >= m.options.MaxDivisors {
			m.operations = beforeOps
			return -1, ErrLimit
		}
		id = len(m.doubles)
		m.doubles = append(m.doubles, FXUDouble{ID: id, Key: FXUPairHash(left, right), Left: append([]int(nil), left...), Right: append([]int(nil), right...), Weight: -(len(left) + len(right)), HeapPos: -1})
		m.doubleByKey[key] = id
	}
	for _, p := range m.doubles[id].Pairs {
		if (p.Cube1 == cube1 && p.Cube2 == cube2) || (p.Cube1 == cube2 && p.Cube2 == cube1) {
			m.operations = beforeOps
			return -1, ErrInvariant
		}
	}
	p := FXUPair{Cube1: cube1, Cube2: cube2, Base: len(base), Lits1: len(left), Lits2: len(right)}
	m.doubles[id].Pairs = append(m.doubles[id].Pairs, p)
	m.doubles[id].Weight += len(left) + len(right) - 1 + len(base)
	m.rebuildDoubleHeap()
	return id, nil
}

func (m *FXUMatrix) DeleteDoublePair(id, cube1, cube2 int) error {
	beforeOps := m.operations
	if err := m.tick(1); err != nil {
		return err
	}
	if id < 0 || id >= len(m.doubles) {
		m.operations = beforeOps
		return ErrDimension
	}
	d := &m.doubles[id]
	found := -1
	for i, p := range d.Pairs {
		if (p.Cube1 == cube1 && p.Cube2 == cube2) || (p.Cube1 == cube2 && p.Cube2 == cube1) {
			found = i
			break
		}
	}
	if found < 0 {
		m.operations = beforeOps
		return ErrInvariant
	}
	p := d.Pairs[found]
	d.Weight -= p.Lits1 + p.Lits2 - 1 + p.Base
	copy(d.Pairs[found:], d.Pairs[found+1:])
	d.Pairs = d.Pairs[:len(d.Pairs)-1]
	if len(d.Pairs) == 0 {
		delete(m.doubleByKey, fxuDivisorKey(d.Left, d.Right))
	}
	m.rebuildDoubleHeap()
	return nil
}

func (m *FXUMatrix) Select() (FXUSelection, error) {
	if err := m.tick(1); err != nil {
		return FXUSelection{}, err
	}
	result := FXUSelection{Weight: -1, SingleID: -1, DoubleID: -1, Var1: -1, Var2: -1}
	if id, ok := m.singleHeap.max(); ok {
		result.Weight = m.singles[id].Weight
		result.SingleID = id
	}
	if id, ok := m.doubleHeap.max(); ok {
		d := m.doubles[id]
		if d.Weight > result.Weight {
			result = FXUSelection{Weight: d.Weight, SingleID: -1, DoubleID: id, Var1: -1, Var2: -1}
		}
	}
	return result, nil
}

func (m *FXUMatrix) SelectSingleCoincidence(weightLimit int) (FXUSelection, error) {
	beforeOps := m.operations
	if err := m.tick(1); err != nil {
		return FXUSelection{}, err
	}
	best := FXUSelection{Weight: -10, SingleID: -1, DoubleID: -1, Var1: -1, Var2: -1}
	for b := 1; b < len(m.variables); b++ {
		for a := 0; a < b; a++ {
			if err := m.tick(1 + len(m.variables[a].Occurrences) + len(m.variables[b].Occurrences)); err != nil {
				m.operations = beforeOps
				return FXUSelection{}, err
			}
			w := intersectCount(m.variables[a].Occurrences, m.variables[b].Occurrences) - 2
			if w > best.Weight {
				best.Weight = w
				best.Var1 = a
				best.Var2 = b
			}
		}
	}
	if best.Weight < weightLimit {
		return best, nil
	}
	return best, nil
}

func (h *fxuHeap) ensure(n int) {
	for len(h.pos) < n {
		h.pos = append(h.pos, -1)
	}
}
func (h *fxuHeap) max() (int, bool) {
	if len(h.items) == 0 {
		return -1, false
	}
	return h.items[0], true
}
func (h *fxuHeap) insertSingle(m *FXUMatrix, id int) {
	h.ensure(id + 1)
	h.items = append(h.items, id)
	h.pos[id] = len(h.items) - 1
	h.upSingle(m, len(h.items)-1)
	for i, v := range h.items {
		m.singles[v].HeapPos = i
	}
}
func (h *fxuHeap) insertDouble(m *FXUMatrix, id int) {
	h.ensure(id + 1)
	h.items = append(h.items, id)
	h.pos[id] = len(h.items) - 1
	h.upDouble(m, len(h.items)-1)
	for i, v := range h.items {
		m.doubles[v].HeapPos = i
	}
}
func better(weightA, idA, weightB, idB int) bool {
	return weightA > weightB || (weightA == weightB && idA < idB)
}
func (h *fxuHeap) upSingle(m *FXUMatrix, i int) {
	for i > 0 {
		p := (i - 1) / 2
		if !better(m.singles[h.items[i]].Weight, h.items[i], m.singles[h.items[p]].Weight, h.items[p]) {
			break
		}
		h.items[i], h.items[p] = h.items[p], h.items[i]
		h.pos[h.items[i]], h.pos[h.items[p]] = i, p
		i = p
	}
}
func (h *fxuHeap) upDouble(m *FXUMatrix, i int) {
	for i > 0 {
		p := (i - 1) / 2
		if !better(m.doubles[h.items[i]].Weight, h.items[i], m.doubles[h.items[p]].Weight, h.items[p]) {
			break
		}
		h.items[i], h.items[p] = h.items[p], h.items[i]
		h.pos[h.items[i]], h.pos[h.items[p]] = i, p
		i = p
	}
}

func (m *FXUMatrix) Validate() error {
	if m == nil || m.closed || m.entries < 0 {
		return ErrInvariant
	}
	count := 0
	for i, v := range m.variables {
		if v.ID != i || !sort.IntsAreSorted(v.Occurrences) {
			return ErrInvariant
		}
		for _, c := range v.Occurrences {
			if c < 0 || c >= len(m.cubes) || !m.cubes[c].Active || sort.SearchInts(m.cubes[c].Literals, i) >= len(m.cubes[c].Literals) || m.cubes[c].Literals[sort.SearchInts(m.cubes[c].Literals, i)] != i {
				return ErrInvariant
			}
		}
	}
	for i, c := range m.cubes {
		if c.ID != i || !sort.IntsAreSorted(c.Literals) {
			return ErrInvariant
		}
		if !c.Active {
			continue
		}
		for _, v := range c.Literals {
			if v < 0 || v >= len(m.variables) {
				return ErrInvariant
			}
			count++
		}
	}
	if count != m.entries {
		return ErrInvariant
	}
	if err := m.validateDivisors(); err != nil {
		return err
	}
	if err := m.validateSingleHeap(); err != nil {
		return err
	}
	if err := m.validateDoubleHeap(); err != nil {
		return err
	}
	return nil
}

func (m *FXUMatrix) validateSingleHeap() error {
	if len(m.singleHeap.items) != len(m.singles) {
		return ErrInvariant
	}
	for i, id := range m.singleHeap.items {
		if id < 0 || id >= len(m.singles) || id >= len(m.singleHeap.pos) || m.singleHeap.pos[id] != i || m.singles[id].HeapPos != i {
			return ErrInvariant
		}
		for _, ch := range []int{2*i + 1, 2*i + 2} {
			if ch < len(m.singleHeap.items) {
				cid := m.singleHeap.items[ch]
				if better(m.singles[cid].Weight, cid, m.singles[id].Weight, id) {
					return ErrInvariant
				}
			}
		}
	}
	return nil
}
func (m *FXUMatrix) validateDoubleHeap() error {
	want := 0
	for _, d := range m.doubles {
		if len(d.Pairs) != 0 {
			want++
		}
	}
	if len(m.doubleHeap.items) != want {
		return ErrInvariant
	}
	for i, id := range m.doubleHeap.items {
		if id < 0 || id >= len(m.doubles) || id >= len(m.doubleHeap.pos) || m.doubleHeap.pos[id] != i || m.doubles[id].HeapPos != i || len(m.doubles[id].Pairs) == 0 {
			return ErrInvariant
		}
		for _, ch := range []int{2*i + 1, 2*i + 2} {
			if ch < len(m.doubleHeap.items) {
				cid := m.doubleHeap.items[ch]
				if better(m.doubles[cid].Weight, cid, m.doubles[id].Weight, id) {
					return ErrInvariant
				}
			}
		}
	}
	return nil
}

func (m *FXUMatrix) validateDivisors() error {
	for i, single := range m.singles {
		if single.ID != i || single.Var1 < 0 || single.Var2 <= single.Var1 || single.Var2 >= len(m.variables) || single.Weight != intersectCount(m.variables[single.Var1].Occurrences, m.variables[single.Var2].Occurrences)-2 {
			return ErrInvariant
		}
	}
	for i, d := range m.doubles {
		if d.ID != i || !sort.IntsAreSorted(d.Left) || !sort.IntsAreSorted(d.Right) || len(d.Left) == 0 || len(d.Right) == 0 || compareInts(d.Left, d.Right) > 0 || d.Key != FXUPairHash(d.Left, d.Right) {
			return ErrInvariant
		}
		weight := -(len(d.Left) + len(d.Right))
		seen := make(map[[2]int]bool, len(d.Pairs))
		for _, pair := range d.Pairs {
			if pair.Cube1 < 0 || pair.Cube2 < 0 || pair.Cube1 >= len(m.cubes) || pair.Cube2 >= len(m.cubes) || pair.Cube1 == pair.Cube2 || seen[[2]int{pair.Cube1, pair.Cube2}] {
				return ErrInvariant
			}
			seen[[2]int{pair.Cube1, pair.Cube2}] = true
			base, left, right := residues(m.cubes[pair.Cube1].Literals, m.cubes[pair.Cube2].Literals)
			if compareInts(left, right) > 0 {
				left, right = right, left
			}
			if !equalInts(left, d.Left) || !equalInts(right, d.Right) || pair.Base != len(base) || pair.Lits1 != len(left) || pair.Lits2 != len(right) {
				return ErrInvariant
			}
			weight += pair.Lits1 + pair.Lits2 - 1 + pair.Base
		}
		if d.Weight != weight {
			return ErrInvariant
		}
		key := fxuDivisorKey(d.Left, d.Right)
		if len(d.Pairs) == 0 {
			if d.HeapPos != -1 {
				return ErrInvariant
			}
		} else if id, ok := m.doubleByKey[key]; !ok || id != i {
			return ErrInvariant
		}
	}
	return nil
}

func equalInts(a, b []int) bool {
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

func (m *FXUMatrix) Variables() []FXUVariable {
	out := make([]FXUVariable, len(m.variables))
	for i, v := range m.variables {
		out[i] = v
		out[i].Occurrences = append([]int(nil), v.Occurrences...)
	}
	return out
}
func (m *FXUMatrix) Cubes() []FXUCube {
	out := make([]FXUCube, len(m.cubes))
	for i, c := range m.cubes {
		out[i] = c
		out[i].Literals = append([]int(nil), c.Literals...)
	}
	return out
}
func (m *FXUMatrix) Singles() []FXUSingle { return append([]FXUSingle(nil), m.singles...) }
func (m *FXUMatrix) Doubles() []FXUDouble {
	out := make([]FXUDouble, len(m.doubles))
	for i, d := range m.doubles {
		out[i] = d
		out[i].Left = append([]int(nil), d.Left...)
		out[i].Right = append([]int(nil), d.Right...)
		out[i].Pairs = append([]FXUPair(nil), d.Pairs...)
	}
	return out
}
func (m *FXUMatrix) EntryCount() int {
	if m == nil {
		return 0
	}
	return m.entries
}
func (m *FXUMatrix) Close() {
	if m == nil {
		return
	}
	m.closed = true
	m.variables = nil
	m.cubes = nil
	m.singles = nil
	m.doubles = nil
	m.doubleByKey = nil
	m.singleHeap = fxuHeap{}
	m.doubleHeap = fxuHeap{}
	m.entries = 0
}
