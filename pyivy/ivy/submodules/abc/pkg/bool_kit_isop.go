package abc

import (
	"errors"
	"strconv"
	"strings"
)

// Ordered irredundant covers from bool/kit/kitIsop.c:Kit_TruthIsop and recursion.

type TruthISOPOptions struct {
	TryBoth                bool
	MaxElements, MaxVisits int // cumulative truth/cube allocations and work; zero: one million each
	Stop                   func() bool
}
type TruthISOP struct {
	Cubes        []uint32 // pair 1 requires false, pair 2 true; variable zero uses the low pair
	Literals     int
	Complemented bool // cover represents the complement when true
	Table        *TruthTable
}
type kitISOPResult struct {
	truth    []uint64
	cubes    []uint32
	literals int
}

func kitISOPTable(variables int, cubes []uint32) *TruthTable {
	wordCount := 1
	if variables > 6 {
		wordCount = 1 << uint(variables-6)
	}
	words := make([]uint64, wordCount)
	for _, cube := range cubes {
		for word := range words {
			value := ^uint64(0)
			for variable := 0; variable < variables; variable++ {
				switch cube >> uint(2*variable) & 3 {
				case 1:
					if variable < 6 {
						value &^= truthVariablePattern(variable)
					} else if word>>uint(variable-6)&1 != 0 {
						value = 0
					}
				case 2:
					if variable < 6 {
						value &= truthVariablePattern(variable)
					} else if word>>uint(variable-6)&1 == 0 {
						value = 0
					}
				}
			}
			words[word] |= value
		}
	}
	return &TruthTable{variables: variables, limit: wordCount, words: words}
}

// ISOP preserves native ordered zero/one/shared decomposition for an exact
// truth table. It is the equal-on/allowed case of ISOPInterval.
func (t *TruthTable) ISOP(o TruthISOPOptions) (TruthISOP, error) {
	return truthISOPInterval(t, t, o)
}

// ISOPInterval implements Kit_TruthIsop2 over an on-set and an allowed
// on/don't-care set. The selected cover contains on and is contained by
// allowed. TryBoth may instead return a cover of the complementary interval;
// Complemented reports that polarity. Table owns the realized truth in the
// returned cover polarity. A failed limited complement attempt retains a valid
// direct cover, while cancellation returns no result. Inputs never mutate.
func (on *TruthTable) ISOPInterval(allowed *TruthTable, o TruthISOPOptions) (TruthISOP, error) {
	return truthISOPInterval(on, allowed, o)
}

func truthISOPInterval(on, allowed *TruthTable, o TruthISOPOptions) (TruthISOP, error) {
	fail := func(e error) (TruthISOP, error) { return TruthISOP{}, e }
	if on == nil || allowed == nil || on.variables != allowed.variables || on.variables > 16 || o.MaxElements < 0 || o.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if e := on.Check(); e != nil {
		return fail(e)
	}
	if e := allowed.Check(); e != nil {
		return fail(e)
	}
	for i := range on.words {
		if on.words[i]&^allowed.words[i] != 0 {
			return fail(ErrDimension)
		}
	}
	if o.MaxElements == 0 {
		o.MaxElements = 1 << 20
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 1 << 20
	}
	elements, visits := 0, 0
	charge := func() error {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		if visits == o.MaxVisits {
			return ErrLimit
		}
		visits++
		return nil
	}
	reserve := func(n int) error {
		if e := charge(); e != nil {
			return e
		}
		next, e := checkedSize(elements, 1, n, o.MaxElements)
		if e == nil {
			elements = next
		}
		return e
	}
	words := func(n int) ([]uint64, error) {
		if e := reserve(n); e != nil {
			return nil, e
		}
		return make([]uint64, n), nil
	}
	var rec func([]uint64, []uint64, int) (kitISOPResult, error)
	rec = func(on, dc []uint64, n int) (kitISOPResult, error) {
		if e := charge(); e != nil {
			return kitISOPResult{}, e
		}
		result, e := words(len(on))
		if e != nil {
			return kitISOPResult{}, e
		}
		zero, one := true, true
		for i, v := range on {
			if e := charge(); e != nil {
				return kitISOPResult{}, e
			}
			zero = zero && v == 0
			one = one && dc[i] == ^uint64(0)
		}
		if zero {
			return kitISOPResult{truth: result}, nil
		}
		if one {
			if e := reserve(1); e != nil {
				return kitISOPResult{}, e
			}
			for i := range result {
				result[i] = ^uint64(0)
			}
			return kitISOPResult{truth: result, cubes: []uint32{0}}, nil
		}
		variable := -1
		for v := n - 1; v >= 0; v-- {
			different := false
			if v < 6 {
				shift := uint(1 << uint(v))
				mask := ^truthVariablePattern(v)
				for i := range on {
					if e := charge(); e != nil {
						return kitISOPResult{}, e
					}
					if ((on[i]^(on[i]>>shift))|(dc[i]^(dc[i]>>shift)))&mask != 0 {
						different = true
						break
					}
				}
			} else {
				half := 1 << uint(v-6)
				for start := 0; start < len(on) && !different; start += 2 * half {
					for i := 0; i < half; i++ {
						if e := charge(); e != nil {
							return kitISOPResult{}, e
						}
						if on[start+i] != on[start+half+i] || dc[start+i] != dc[start+half+i] {
							different = true
							break
						}
					}
				}
			}
			if different {
				variable = v
				break
			}
		}
		if variable < 0 {
			return kitISOPResult{}, ErrInvariant
		}
		var on0, on1, dc0, dc1 []uint64
		if variable < 6 {
			cofactor := func(v uint64, high bool) uint64 {
				mask := truthVariablePattern(variable)
				shift := uint(1 << uint(variable))
				if high {
					v &= mask
					return v | v>>shift
				}
				v &= ^mask
				return v | v<<shift
			}
			if e := reserve(4); e != nil {
				return kitISOPResult{}, e
			}
			on0 = []uint64{cofactor(on[0], false)}
			on1 = []uint64{cofactor(on[0], true)}
			dc0 = []uint64{cofactor(dc[0], false)}
			dc1 = []uint64{cofactor(dc[0], true)}
		} else {
			half := 1 << uint(variable-6)
			on0, on1 = on[:half], on[half:2*half]
			dc0, dc1 = dc[:half], dc[half:2*half]
		}
		temp0, e := words(len(on0))
		if e != nil {
			return kitISOPResult{}, e
		}
		for i := range temp0 {
			if e := charge(); e != nil {
				return kitISOPResult{}, e
			}
			temp0[i] = on0[i] &^ dc1[i]
		}
		r0, e := rec(temp0, dc0, variable)
		if e != nil {
			return kitISOPResult{}, e
		}
		temp1, e := words(len(on0))
		if e != nil {
			return kitISOPResult{}, e
		}
		for i := range temp1 {
			if e := charge(); e != nil {
				return kitISOPResult{}, e
			}
			temp1[i] = on1[i] &^ dc0[i]
		}
		r1, e := rec(temp1, dc1, variable)
		if e != nil {
			return kitISOPResult{}, e
		}
		for i := range temp0 {
			if e := charge(); e != nil {
				return kitISOPResult{}, e
			}
			temp0[i] = on0[i]&^r0.truth[i] | on1[i]&^r1.truth[i]
			temp1[i] = dc0[i] & dc1[i]
		}
		shared, e := rec(temp0, temp1, variable)
		if e != nil {
			return kitISOPResult{}, e
		}
		count := len(r0.cubes) + len(r1.cubes) + len(shared.cubes)
		if e := reserve(count); e != nil {
			return kitISOPResult{}, e
		}
		cubes := make([]uint32, 0, count)
		for _, cube := range r0.cubes {
			if e := charge(); e != nil {
				return kitISOPResult{}, e
			}
			cubes = append(cubes, cube|uint32(1)<<uint(2*variable))
		}
		for _, cube := range r1.cubes {
			if e := charge(); e != nil {
				return kitISOPResult{}, e
			}
			cubes = append(cubes, cube|uint32(2)<<uint(2*variable))
		}
		cubes = append(cubes, shared.cubes...)
		if variable < 6 {
			v := shared.truth[0] | r0.truth[0]&^truthVariablePattern(variable) | r1.truth[0]&truthVariablePattern(variable)
			for i := range result {
				if e := charge(); e != nil {
					return kitISOPResult{}, e
				}
				result[i] = v
			}
		} else {
			half := len(on0)
			for i := range result {
				if e := charge(); e != nil {
					return kitISOPResult{}, e
				}
				j := i % (2 * half)
				if j < half {
					result[i] = r0.truth[j] | shared.truth[j]
				} else {
					result[i] = r1.truth[j-half] | shared.truth[j-half]
				}
			}
		}
		return kitISOPResult{truth: result, cubes: cubes, literals: r0.literals + r1.literals + shared.literals + len(r0.cubes) + len(r1.cubes)}, nil
	}
	direct, e := rec(on.words, allowed.words, on.variables)
	if e != nil {
		return fail(e)
	}
	r := TruthISOP{Cubes: direct.cubes, Literals: direct.literals, Table: kitISOPTable(on.variables, direct.cubes)}
	retained := func() (TruthISOP, error) {
		if o.Stop != nil && o.Stop() {
			return fail(ErrCancelled)
		}
		return r, nil
	}
	if o.TryBoth && len(r.Cubes) > 0 && !(len(r.Cubes) == 1 && r.Cubes[0] == 0) {
		complementOn, e := words(len(on.words))
		if e != nil {
			if errors.Is(e, ErrLimit) {
				return retained()
			}
			return fail(e)
		}
		complementAllowed, e := words(len(on.words))
		if e != nil {
			if errors.Is(e, ErrLimit) {
				return retained()
			}
			return fail(e)
		}
		for i := range on.words {
			if e := charge(); e != nil {
				if errors.Is(e, ErrLimit) {
					return retained()
				}
				return fail(e)
			}
			complementOn[i] = ^allowed.words[i]
			complementAllowed[i] = ^on.words[i]
		}
		other, e := rec(complementOn, complementAllowed, on.variables)
		if e != nil {
			if errors.Is(e, ErrLimit) {
				return retained()
			}
			return fail(e)
		}
		if len(other.cubes) < len(r.Cubes) || len(other.cubes) == len(r.Cubes) && other.literals < r.Literals {
			r = TruthISOP{Cubes: other.cubes, Literals: other.literals, Complemented: true, Table: kitISOPTable(on.variables, other.cubes)}
		}
	}
	if o.Stop != nil && o.Stop() {
		return fail(ErrCancelled)
	}
	return r, nil
}

// Format returns Kit_TruthIsopPrintCover's deterministic PLA-style output.
func (r TruthISOP) Format(variables int) (string, error) {
	if variables < 0 || variables > 16 {
		return "", ErrDimension
	}
	if len(r.Cubes) == 0 || len(r.Cubes) == 1 && r.Cubes[0] == 0 {
		return "Constant " + strconv.Itoa(len(r.Cubes)) + "\n", nil
	}
	var output strings.Builder
	for _, cube := range r.Cubes {
		for variable := 0; variable < variables; variable++ {
			switch cube >> uint(2*variable) & 3 {
			case 0:
				output.WriteByte('-')
			case 1:
				output.WriteByte('0')
			case 2:
				output.WriteByte('1')
			default:
				return "", ErrDimension
			}
		}
		if r.Complemented {
			output.WriteString(" 0\n")
		} else {
			output.WriteString(" 1\n")
		}
	}
	return output.String(), nil
}
