package abc

// Selected-cut mapping decode from aig/gia/giaSatLut.c:
// Sbl_ManGetCurrentMapping.

import "math/bits"

// GiaWindowCut uses two 128-bit masks for leaf and internal-node ordinals.
// Node is the destination index in nodes, not a GIA object ID. Unselected
// records may be leaf/sentinel entries and are not interpreted.
type GiaWindowCut struct {
	Node          int
	Leaves, Nodes [2]uint64
}

// DecodeGiaWindowCuts expands selected cut masks in native category/bit order.
// Later selected records for the same destination replace earlier lists.
// Values, including duplicate/zero IDs, are retained verbatim; topology, LUT
// width, candidate closure and Boolean correctness are separate validations.
// An empty selection rejects C's assertion precondition. MaxObjects bounds
// total input entries; MaxVisits counts selected reads and expanded bits,
// including overwritten records. Errors/cancellation expose no partial lists.
func DecodeGiaWindowCuts(leaves, nodes []GiaID, cuts []GiaWindowCut, selected []int, options GiaLUTAnalysisOptions) ([][]GiaID, error) {
	if options.MaxObjects < 0 || options.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	total := 0
	for _, count := range []int{len(leaves), len(nodes), len(cuts), len(selected)} {
		var err error
		total, err = checkedSize(total, 1, count, options.MaxObjects)
		if err != nil {
			return nil, err
		}
	}
	if len(selected) == 0 {
		return nil, ErrInvariant
	}
	work := 0
	charge := func() error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if work == options.MaxVisits {
			return ErrLimit
		}
		work++
		return nil
	}
	window := make([][]GiaID, len(nodes))
	for _, index := range selected {
		if err := charge(); err != nil {
			return nil, err
		}
		if index < 0 || index >= len(cuts) {
			return nil, ErrDimension
		}
		cut := cuts[index]
		if cut.Node < 0 || cut.Node >= len(nodes) {
			return nil, ErrDimension
		}
		result := window[cut.Node][:0]
		for category := 0; category < 2; category++ {
			values, mask := leaves, cut.Leaves
			if category == 1 {
				values, mask = nodes, cut.Nodes
			}
			for half, word := range mask {
				for word != 0 {
					if err := charge(); err != nil {
						return nil, err
					}
					bit := bits.TrailingZeros64(word)
					position := half*64 + bit
					if position >= len(values) {
						return nil, ErrDimension
					}
					result = append(result, values[position])
					word &= word - 1
				}
			}
		}
		window[cut.Node] = result
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	return window, nil
}
