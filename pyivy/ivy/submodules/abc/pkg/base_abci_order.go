package abc

// Combinational-input ordering from src/base/abci/abcOrder.c.

import (
	"slices"
	"strings"
)

type ABCCIOrderOptions struct {
	MaxCIs        int // retained order entries; zero: one million
	MaxVisited    int // support traversal visits; zero: network/default bound
	MaxInputBytes int // owned textual order bytes; zero: sixteen MiB
	Stop          func() bool
}

func normalizeABCCIOrderOptions(options ABCCIOrderOptions) (ABCCIOrderOptions, error) {
	if options.MaxCIs < 0 || options.MaxVisited < 0 || options.MaxInputBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxCIs == 0 {
		options.MaxCIs = 1 << 20
	}
	if options.MaxInputBytes == 0 {
		options.MaxInputBytes = 1 << 24
	}
	return options, nil
}

// ABCChangeCIOrder applies an exact CI permutation and rebuilds the PI order by
// filtering it, matching Abc_NtkChangeCiOrder. Failure is atomic. Reverse writes
// the supplied order from the opposite end, as in the source.
func ABCChangeCIOrder(network *ABCNetwork, order []int, reverse bool, options ABCCIOrderOptions) error {
	options, err := normalizeABCCIOrderOptions(options)
	if err != nil {
		return err
	}
	if network == nil {
		return ErrDimension
	}
	if err = network.Validate(); err != nil {
		return err
	}
	if len(order) != len(network.cis) {
		return ErrDimension
	}
	if len(order) > options.MaxCIs {
		return ErrLimit
	}
	seen := make([]bool, len(network.objects))
	result := make([]int, len(order))
	for index, id := range order {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if id < 0 || id >= len(network.objects) || seen[id] || !network.objects[id].Alive || !abcIsCI(network.objects[id].Type) {
			return ErrDimension
		}
		seen[id] = true
		destination := index
		if reverse {
			destination = len(order) - 1 - index
		}
		result[destination] = id
	}
	pis := make([]int, 0, len(network.pis))
	for _, id := range result {
		if network.objects[id].Type == ABCObjectPI {
			pis = append(pis, id)
		}
	}
	network.cis = result
	network.pis = pis
	return nil
}

// ABCFindCIOrder implements Abc_NtkFindCiOrder. CIs are ordered by their first
// encounter in the CO fanin traversal; unused CIs follow in their previous
// order. The returned slice is owned and reports the applied order.
func ABCFindCIOrder(network *ABCNetwork, reverse bool, options ABCCIOrderOptions) ([]int, error) {
	options, err := normalizeABCCIOrderOptions(options)
	if err != nil {
		return nil, err
	}
	if network == nil {
		return nil, ErrDimension
	}
	if len(network.cis) > options.MaxCIs {
		return nil, ErrLimit
	}
	order, err := ABCNetworkSupport(network, nil, ABCAnalysisOptions{MaxVisited: options.MaxVisited, Stop: options.Stop})
	if err != nil {
		return nil, err
	}
	seen := make([]bool, len(network.objects))
	for _, id := range order {
		seen[id] = true
	}
	for _, id := range network.cis {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if !seen[id] {
			seen[id] = true
			order = append(order, id)
		}
	}
	if len(order) != len(network.cis) {
		return nil, ErrInvariant
	}
	if err = ABCChangeCIOrder(network, order, reverse, options); err != nil {
		return nil, err
	}
	return network.CIs(), nil
}

// ABCImplementCIOrder is the owned-byte counterpart of
// Abc_NtkImplementCiOrder. ASCII whitespace separates exact CI names. The list
// must be a complete permutation; unknown, non-CI and duplicate names fail
// atomically instead of corrupting the terminal partition.
func ABCImplementCIOrder(network *ABCNetwork, data []byte, reverse bool, options ABCCIOrderOptions) ([]int, error) {
	options, err := normalizeABCCIOrderOptions(options)
	if err != nil {
		return nil, err
	}
	if network == nil {
		return nil, ErrDimension
	}
	if len(data) > options.MaxInputBytes {
		return nil, ErrLimit
	}
	if err = network.Validate(); err != nil {
		return nil, err
	}
	if len(network.cis) > options.MaxCIs {
		return nil, ErrLimit
	}
	names := strings.Fields(string(slices.Clone(data)))
	if len(names) != len(network.cis) {
		return nil, ErrDimension
	}
	order := make([]int, len(names))
	seen := make([]bool, len(network.objects))
	for index, name := range names {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		id, found := network.nameToID[name]
		if !found || seen[id] || !abcIsCI(network.objects[id].Type) {
			return nil, ErrDimension
		}
		seen[id] = true
		order[index] = id
	}
	if err = ABCChangeCIOrder(network, order, reverse, options); err != nil {
		return nil, err
	}
	return network.CIs(), nil
}
