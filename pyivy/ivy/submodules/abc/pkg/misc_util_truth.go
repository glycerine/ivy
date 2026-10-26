package abc

// Packed truth-table primitives from src/misc/util/utilTruth.h.

import (
	"math/bits"
	"slices"
)

// TruthTable owns an ABC-format packed Boolean function. Variable zero is the
// assignment's low bit; tables below six variables repeat their logical bits
// throughout one word. Use constructors; zero values are unconfigured. Binary
// operations return independent tables and require equal variable counts.
type TruthTable struct {
	variables, limit int
	words            []uint64
}

func NewConstantTruthTable(variables int, value bool, maxWords int) (*TruthTable, error) {
	if maxWords < 0 {
		return nil, ErrDimension
	}
	if maxWords == 0 {
		maxWords = 1 << 20
	}
	count, e := TruthTableWordCount(variables)
	if e != nil {
		return nil, e
	}
	if count > maxWords {
		return nil, ErrLimit
	}
	if _, e = checkedSize(0, 8, count, int(^uint(0)>>1)); e != nil {
		return nil, e
	}
	t := &TruthTable{variables, maxWords, make([]uint64, count)}
	if value {
		for i := range t.words {
			t.words[i] = ^uint64(0)
		}
	}
	return t, nil
}

// TruthTableFromWords requires canonical stretched small tables; malformed
// padding rejects rather than being silently reinterpreted. Input storage copies.
func TruthTableFromWords(variables int, words []uint64, maxWords int) (*TruthTable, error) {
	t, e := NewConstantTruthTable(variables, false, maxWords)
	if e != nil {
		return nil, e
	}
	if len(words) != len(t.words) {
		return nil, ErrDimension
	}
	copy(t.words, words)
	if e = t.Check(); e != nil {
		return nil, e
	}
	return t, nil
}
func truthVariablePattern(variable int) uint64 {
	switch variable {
	case 0:
		return 0xaaaaaaaaaaaaaaaa
	case 1:
		return 0xcccccccccccccccc
	case 2:
		return 0xf0f0f0f0f0f0f0f0
	case 3:
		return 0xff00ff00ff00ff00
	case 4:
		return 0xffff0000ffff0000
	case 5:
		return 0xffffffff00000000
	}
	return 0
}
func NewVariableTruthTable(variables, index, maxWords int) (*TruthTable, error) {
	if index < 0 || index >= variables {
		return nil, ErrDimension
	}
	t, e := NewConstantTruthTable(variables, false, maxWords)
	if e != nil {
		return nil, e
	}
	for i := range t.words {
		if index < 6 {
			t.words[i] = truthVariablePattern(index)
		} else if i&(1<<uint(index-6)) != 0 {
			t.words[i] = ^uint64(0)
		}
	}
	return t, nil
}
func (t *TruthTable) Variables() int {
	if t == nil {
		return 0
	}
	return t.variables
}
func (t *TruthTable) Words() []uint64 {
	if t == nil {
		return nil
	}
	return slices.Clone(t.words)
}
func (t *TruthTable) Clone() *TruthTable {
	if t == nil {
		return nil
	}
	return &TruthTable{t.variables, t.limit, slices.Clone(t.words)}
}
func (t *TruthTable) Check() error {
	if t == nil || t.limit <= 0 {
		return ErrDimension
	}
	count, e := TruthTableWordCount(t.variables)
	if e != nil {
		return e
	}
	if count != len(t.words) || count > t.limit {
		return ErrInvariant
	}
	if t.variables < 6 {
		period := 1 << uint(t.variables)
		for bit := period; bit < 64; bit++ {
			if ((t.words[0]>>uint(bit))^(t.words[0]>>uint(bit%period)))&1 != 0 {
				return ErrInvariant
			}
		}
	}
	return nil
}
func (t *TruthTable) Get(assignment uint64) (bool, error) {
	if e := t.Check(); e != nil {
		return false, e
	}
	if assignment >= uint64(1)<<uint(t.variables) {
		return false, ErrDimension
	}
	return t.words[assignment/64]&(uint64(1)<<uint(assignment%64)) != 0, nil
}

// Set updates one logical assignment and all its repeated small-table bits.
func (t *TruthTable) Set(assignment uint64, value bool) error {
	if e := t.Check(); e != nil {
		return e
	}
	if assignment >= uint64(1)<<uint(t.variables) {
		return ErrDimension
	}
	step := uint64(1) << uint(t.variables)
	end := uint64(len(t.words)) * 64
	for bit := assignment; bit < end; bit += step {
		mask := uint64(1) << uint(bit%64)
		if value {
			t.words[bit/64] |= mask
		} else {
			t.words[bit/64] &^= mask
		}
	}
	return nil
}
func (t *TruthTable) Not() *TruthTable {
	r := t.Clone()
	if r != nil {
		for i := range r.words {
			r.words[i] = ^r.words[i]
		}
	}
	return r
}
func (t *TruthTable) checkPair(other *TruthTable) error {
	if e := t.Check(); e != nil {
		return e
	}
	if e := other.Check(); e != nil {
		return e
	}
	if t.variables != other.variables {
		return ErrDimension
	}
	return nil
}

// And complements each input independently before combining, matching AndCompl.
func (t *TruthTable) And(other *TruthTable, complementSelf, complementOther bool) (*TruthTable, error) {
	if e := t.checkPair(other); e != nil {
		return nil, e
	}
	r := t.Clone()
	for i, a := range t.words {
		b := other.words[i]
		if complementSelf {
			a = ^a
		}
		if complementOther {
			b = ^b
		}
		r.words[i] = a & b
	}
	return r, nil
}
func (t *TruthTable) Or(other *TruthTable) (*TruthTable, error) {
	if e := t.checkPair(other); e != nil {
		return nil, e
	}
	r := t.Clone()
	for i, w := range r.words {
		r.words[i] = w | other.words[i]
	}
	return r, nil
}
func (t *TruthTable) Xor(other *TruthTable, complement bool) (*TruthTable, error) {
	if e := t.checkPair(other); e != nil {
		return nil, e
	}
	r := t.Clone()
	for i, w := range r.words {
		r.words[i] = w ^ other.words[i]
		if complement {
			r.words[i] = ^r.words[i]
		}
	}
	return r, nil
}
func MuxTruthTables(control, whenTrue, whenFalse *TruthTable) (*TruthTable, error) {
	if e := control.checkPair(whenTrue); e != nil {
		return nil, e
	}
	if e := control.checkPair(whenFalse); e != nil {
		return nil, e
	}
	r := control.Clone()
	for i, c := range control.words {
		r.words[i] = (c & whenTrue.words[i]) | (^c & whenFalse.words[i])
	}
	return r, nil
}
func MajorityTruthTables(a, b, c *TruthTable) (*TruthTable, error) {
	if e := a.checkPair(b); e != nil {
		return nil, e
	}
	if e := a.checkPair(c); e != nil {
		return nil, e
	}
	r := a.Clone()
	for i, x := range a.words {
		y, z := b.words[i], c.words[i]
		r.words[i] = (x & y) | (x & z) | (y & z)
	}
	return r, nil
}
func (t *TruthTable) Equal(other *TruthTable) bool {
	return t != nil && other != nil && t.variables == other.variables && slices.Equal(t.words, other.words)
}
func (t *TruthTable) Opposite(other *TruthTable) (bool, error) {
	if e := t.checkPair(other); e != nil {
		return false, e
	}
	for i, w := range t.words {
		if w != ^other.words[i] {
			return false, nil
		}
	}
	return true, nil
}
func (t *TruthTable) Implies(other *TruthTable) (bool, error) {
	if e := t.checkPair(other); e != nil {
		return false, e
	}
	for i, w := range t.words {
		if w&^other.words[i] != 0 {
			return false, nil
		}
	}
	return true, nil
}
func (t *TruthTable) Intersects(other *TruthTable, complementSelf bool) (bool, error) {
	if e := t.checkPair(other); e != nil {
		return false, e
	}
	for i, w := range t.words {
		if complementSelf {
			w = ^w
		}
		if w&other.words[i] != 0 {
			return true, nil
		}
	}
	return false, nil
}

// Compare scans low words first, or high words first when reverseWords is true.
func (t *TruthTable) Compare(other *TruthTable, reverseWords bool) (int, error) {
	if e := t.checkPair(other); e != nil {
		return 0, e
	}
	for i := 0; i < len(t.words); i++ {
		index := i
		if reverseWords {
			index = len(t.words) - 1 - i
		}
		a, b := t.words[index], other.words[index]
		if a < b {
			return -1, nil
		}
		if a > b {
			return 1, nil
		}
	}
	return 0, nil
}
func (t *TruthTable) IsConstant(value bool) bool {
	if t == nil || len(t.words) == 0 {
		return false
	}
	want := uint64(0)
	if value {
		want = ^uint64(0)
	}
	for _, w := range t.words {
		if w != want {
			return false
		}
	}
	return true
}

// CountOnes counts logical assignments, excluding stretched padding repeats.
func (t *TruthTable) CountOnes() int {
	if t == nil {
		return 0
	}
	count := 0
	for _, w := range t.words {
		count += bits.OnesCount64(w)
	}
	if t.variables < 6 {
		count >>= uint(6 - t.variables)
	}
	return count
}

// Cofactor retains the same variable count and stretches both selected halves,
// matching Cofactor0p/Cofactor1p. The fixed variable becomes unused.
func (t *TruthTable) Cofactor(variable int, value bool) (*TruthTable, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	if variable < 0 || variable >= t.variables {
		return nil, ErrDimension
	}
	r := t.Clone()
	if variable < 6 {
		shift := uint(1) << uint(variable)
		mask := ^truthVariablePattern(variable)
		for i, w := range t.words {
			if value {
				w = (w >> shift) & mask
			} else {
				w &= mask
			}
			r.words[i] = w | (w << shift)
		}
	} else {
		block := 1 << uint(variable-6)
		for base := 0; base < len(t.words); base += 2 * block {
			offset := base
			if value {
				offset += block
			}
			copy(r.words[base:base+block], t.words[offset:offset+block])
			copy(r.words[base+block:base+2*block], t.words[offset:offset+block])
		}
	}
	return r, nil
}
func (t *TruthTable) HasVariable(variable int) (bool, error) {
	if e := t.Check(); e != nil {
		return false, e
	}
	if variable < 0 || variable >= t.variables {
		return false, ErrDimension
	}
	if variable < 6 {
		shift := uint(1) << uint(variable)
		mask := ^truthVariablePattern(variable)
		for _, w := range t.words {
			if ((w>>shift)^w)&mask != 0 {
				return true, nil
			}
		}
	} else {
		block := 1 << uint(variable-6)
		for base := 0; base < len(t.words); base += 2 * block {
			if !slices.Equal(t.words[base:base+block], t.words[base+block:base+2*block]) {
				return true, nil
			}
		}
	}
	return false, nil
}
func (t *TruthTable) Support() ([]int, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	support := []int{}
	for variable := 0; variable < t.variables; variable++ {
		yes, e := t.HasVariable(variable)
		if e != nil {
			return nil, e
		}
		if yes {
			support = append(support, variable)
		}
	}
	return support, nil
}
