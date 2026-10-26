package abc

// Boolean cut substitution from sat/cnf/cnfCut.c:Cnf_CutCompose.

import (
	"strconv"
	"strings"
)

type CNFCompositionOptions struct {
	MaxInputs int // zero: nine, matching native nMergeLimit=10 including temporary input
	MaxVisits int // zero: one million assignments
	ISOP      TruthISOPOptions
	Stop      func() bool
}
type CNFComposition struct {
	Cut  AigCNFCut
	Cost int // sum of polarity cube counts, saturated at native sentinel 127
}

// FormatCNFCut replaces Cnf_CutPrint's stdout side effect with its exact leaf
// layout. Truth and root are intentionally not printed by the source helper.
func FormatCNFCut(c AigCNFCut) string {
	var output strings.Builder
	output.WriteByte('{')
	for _, leaf := range c.Leaves {
		output.WriteString(strconv.FormatUint(uint64(leaf), 10))
		output.WriteByte(' ')
	}
	output.WriteString(" } ")
	return output.String()
}

// ComposeCNFCuts substitutes fanin's function for its root in top. Leaves
// must be strictly ordered; functions and storage are owned and never mutated.
// Unlike the original native implementation, truth is read before removing
// the substituted leaf: native code otherwise reads a leaf ID as truth data.
// It also correctly aligns the substituted support and extracts packed mux
// cofactors, correcting native support-movement and small-table errors.
// No support minimization is performed after composition. Limits/cancellation
// return no result; callers may skip limit failures when exploring mappings.
func ComposeCNFCuts(top, fanin AigCNFCut, o CNFCompositionOptions) (*CNFComposition, error) {
	if o.MaxInputs < 0 || o.MaxInputs > 16 || o.MaxVisits < 0 || o.ISOP.MaxElements < 0 || o.ISOP.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if o.MaxInputs == 0 {
		o.MaxInputs = 9
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 1 << 20
	}
	check := func() error {
		if o.Stop != nil && o.Stop() || o.ISOP.Stop != nil && o.ISOP.Stop() {
			return ErrCancelled
		}
		return nil
	}
	if e := check(); e != nil {
		return nil, e
	}
	validate := func(c AigCNFCut) error {
		if c.Truth == nil || c.Truth.variables != len(c.Leaves) || len(c.Leaves) > 16 || uint32(c.Root) > MaxLiteralID {
			return ErrDimension
		}
		if e := c.Truth.Check(); e != nil {
			return e
		}
		for i, id := range c.Leaves {
			if id >= c.Root || i > 0 && id <= c.Leaves[i-1] {
				return ErrDimension
			}
		}
		return nil
	}
	if e := validate(top); e != nil {
		return nil, e
	}
	if e := validate(fanin); e != nil {
		return nil, e
	}
	substituted := -1
	for i, id := range top.Leaves {
		if id == fanin.Root {
			substituted = i
			break
		}
	}
	if substituted < 0 {
		return nil, ErrDimension
	}
	var leaves []AigID
	i, j := 0, 0
	for i < len(top.Leaves) || j < len(fanin.Leaves) {
		if i == substituted {
			i++
			continue
		}
		var id AigID
		if j == len(fanin.Leaves) || i < len(top.Leaves) && top.Leaves[i] < fanin.Leaves[j] {
			id = top.Leaves[i]
			i++
		} else if i == len(top.Leaves) || fanin.Leaves[j] < top.Leaves[i] {
			id = fanin.Leaves[j]
			j++
		} else {
			id = top.Leaves[i]
			i++
			j++
		}
		leaves = append(leaves, id)
	}
	if len(leaves) > o.MaxInputs {
		return nil, ErrLimit
	}
	positions := func(input []AigID) []int {
		p := make([]int, len(input))
		for i, id := range input {
			p[i] = -1
			for j, leaf := range leaves {
				if id == leaf {
					p[i] = j
					break
				}
			}
		}
		return p
	}
	topPositions, fanPositions := positions(top.Leaves), positions(fanin.Leaves)
	words := make([]uint64, 1<<uint(max(0, len(leaves)-6)))
	visits := 0
	for assignment := 0; assignment < len(words)*64; assignment++ {
		if e := check(); e != nil {
			return nil, e
		}
		if visits == o.MaxVisits {
			return nil, ErrLimit
		}
		visits++
		fanIndex := 0
		for i, p := range fanPositions {
			fanIndex |= ((assignment >> uint(p)) & 1) << uint(i)
		}
		fanValue := int((fanin.Truth.words[fanIndex/64] >> uint(fanIndex%64)) & 1)
		topIndex := 0
		for i, p := range topPositions {
			v := fanValue
			if i != substituted {
				v = (assignment >> uint(p)) & 1
			}
			topIndex |= v << uint(i)
		}
		words[assignment/64] |= ((top.Truth.words[topIndex/64] >> uint(topIndex%64)) & 1) << uint(assignment%64)
	}
	truth, e := TruthTableFromWords(len(leaves), words, 0)
	if e != nil {
		return nil, e
	}
	cost := 0
	if len(leaves) <= 4 {
		v := uint16(words[0])
		pos, e := CNFSOP4Cover(v)
		if e != nil {
			return nil, e
		}
		neg, e := CNFSOP4Cover(^v)
		if e != nil {
			return nil, e
		}
		cost = len(pos) + len(neg)
	} else {
		isop := o.ISOP
		isop.TryBoth = false
		isop.Stop = func() bool { return check() != nil }
		pos, e := truth.ISOP(isop)
		if e != nil {
			return nil, e
		}
		neg, e := truth.Not().ISOP(isop)
		if e != nil {
			return nil, e
		}
		cost = min(127, len(pos.Cubes)+len(neg.Cubes))
	}
	if e := check(); e != nil {
		return nil, e
	}
	return &CNFComposition{Cut: AigCNFCut{Root: top.Root, Leaves: leaves, Truth: truth}, Cost: cost}, nil
}
