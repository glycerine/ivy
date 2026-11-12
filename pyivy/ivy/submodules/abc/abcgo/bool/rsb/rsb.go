// Package rsb implements truth-table resubstitution using counterexample-guided
// search over sets of up to four divisors.
package rsb

type Rsb_Man_t struct {
	leafMax, decMax          int32
	verbose                  bool
	cexes                    []uint64
	patterns                 []uint32
	fanins, faninsOld, tries []int32
}
