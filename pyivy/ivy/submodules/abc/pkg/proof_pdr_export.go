package abc

// Port of the full-register PLA path in src/proof/pdr/pdrInv.c:
// Pdr_ManDumpClauses. Timestamp and host-dependent headers are omitted.

import (
	"fmt"
	"slices"
	"strings"
)

// MarshalPLA writes the excluded-state cover, which is the complement of
// this invariant. A PLA output of one means a blocked state, not safety.
// This serialization does not verify initiation, induction or safety.
// maxBytes zero selects 64 MiB. Cube order follows ABC's Pdr_SetCompare;
// no model names or time-dependent comments are emitted.
func (invariant *PdrInvariant) MarshalPLA(registers, maxBytes int) ([]byte, error) {
	if invariant == nil || registers < 0 || maxBytes < 0 {
		return nil, ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 64 << 20
	}
	for _, cube := range invariant.clauses {
		if err := cube.Check(); err != nil {
			return nil, err
		}
		for _, literal := range cube.state {
			if uint64(literal.Index()) >= uint64(registers) {
				return nil, ErrDimension
			}
		}
	}
	header := fmt.Sprintf(".i %d\n.o 1\n.p %d\n", registers, len(invariant.clauses))
	// Each row contains all register columns, a space, output one and LF.
	// Check the complete allocation before creating rows or sorting cubes.
	row, err := checkedSize(3, 1, registers, maxBytes)
	if err != nil {
		return nil, err
	}
	total, err := checkedSize(len(header)+3, len(invariant.clauses), row, maxBytes)
	if err != nil {
		return nil, err
	}
	ordered := slices.Clone(invariant.clauses)
	slices.SortStableFunc(ordered, func(a, b *PdrCube) int { return a.Compare(b) })
	var out strings.Builder
	out.Grow(total)
	out.WriteString(header)
	for _, cube := range ordered {
		line, err := cube.StateString(registers)
		if err != nil {
			return nil, err
		}
		out.WriteString(line)
		out.WriteString(" 1\n")
	}
	out.WriteString(".e\n")
	return []byte(out.String()), nil
}
