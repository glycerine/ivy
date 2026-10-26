package abc

// Port of src/aig/aig/aigPack.c using concrete owned 64-pattern words.

import (
	"math/bits"
	"slices"
)

const (
	aigPackRandomZ = uint32(3716960521)
	aigPackRandomW = uint32(2174103536)
)

type AigPatternPackOptions struct {
	// ResetRandom reproduces a call to Aig_ManRandom(1), including the value
	// returned (and consumed) by that reset call, before randomization.
	ResetRandom bool
	MaxObjects  int // zero: 1M graph objects
	MaxPatterns int // zero: 1M AddPattern calls
	Stop        func() bool
}

type AigPatternPackStats struct {
	PatternsTotal, PatternsSkipped, PatternsRepeated int
	CareBits                                         int
	NodeHistogram                                    [33]int
}

type AigPatternPack struct {
	graph               *AigManager
	signs               []uint64
	inputPatterns       []uint64
	inputCare           []uint64
	patternsTotal       int
	patternsSkipped     int
	patternsRepeated    int
	maxPatterns         int
	maxObjects          int
	stop                func() bool
	randomZ, randomW    uint32
	originalObjectCount int
}

func (p *AigPatternPack) check() error {
	if p == nil || p.graph == nil {
		return ErrDimension
	}
	if err := p.graph.Check(); err != nil {
		return err
	}
	if p.graph.ObjectCount() != p.originalObjectCount || len(p.signs) != p.graph.ObjectCount() || len(p.inputPatterns) != p.graph.InputCount() || len(p.inputCare) != p.graph.InputCount() {
		return ErrInvariant
	}
	for _, object := range p.graph.objects {
		if object.Type == AigBuffer || object.Type == AigXOR {
			return ErrUnsupported
		}
	}
	return nil
}

func (p *AigPatternPack) random() uint32 {
	p.randomZ = 36969*(p.randomZ&65535) + (p.randomZ >> 16)
	p.randomW = 18000*(p.randomW&65535) + (p.randomW >> 16)
	return (p.randomZ << 16) + p.randomW
}

func NewAigPatternPack(graph *AigManager, options AigPatternPackOptions) (*AigPatternPack, error) {
	if graph == nil || options.MaxObjects < 0 || options.MaxPatterns < 0 {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	for _, object := range graph.objects {
		if object.Type == AigBuffer || object.Type == AigXOR {
			return nil, ErrUnsupported
		}
	}
	if options.MaxPatterns == 0 {
		options.MaxPatterns = 1 << 20
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if graph.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	p := &AigPatternPack{
		graph:               graph,
		signs:               make([]uint64, graph.ObjectCount()),
		inputPatterns:       make([]uint64, graph.InputCount()),
		inputCare:           make([]uint64, graph.InputCount()),
		maxPatterns:         options.MaxPatterns,
		maxObjects:          options.MaxObjects,
		stop:                options.Stop,
		randomZ:             aigPackRandomZ,
		randomW:             aigPackRandomW,
		originalObjectCount: graph.ObjectCount(),
	}
	if options.ResetRandom {
		p.ResetRandom()
	}
	if err := p.SetRandom(); err != nil {
		return nil, err
	}
	if err := p.Simulate(); err != nil {
		return nil, err
	}
	return p, nil
}

// ResetRandom mirrors Aig_ManRandom(1): it restores the two native seeds and
// consumes/returns the first value in that reset stream.
func (p *AigPatternPack) ResetRandom() uint32 {
	p.randomZ, p.randomW = aigPackRandomZ, aigPackRandomW
	return p.random()
}

// SetRandom implements Aig_ManPackSetRandom using this pack's local native RNG
// state. Keeping the state local avoids cross-manager global interference.
func (p *AigPatternPack) SetRandom() error {
	if err := p.check(); err != nil {
		return err
	}
	for i := range p.inputPatterns {
		if p.stop != nil && p.stop() {
			return ErrCancelled
		}
		high, low := p.random(), p.random()
		p.inputPatterns[i] = (uint64(high)<<32 | uint64(low)) << 1
	}
	return nil
}

// SetInputPatterns installs an owned deterministic assignment block.
func (p *AigPatternPack) SetInputPatterns(patterns []uint64) error {
	if err := p.check(); err != nil {
		return err
	}
	if len(patterns) != len(p.inputPatterns) {
		return ErrDimension
	}
	copy(p.inputPatterns, patterns)
	return nil
}

// Simulate refreshes all object signatures from the packed CI assignments.
func (p *AigPatternPack) Simulate() error {
	if err := p.check(); err != nil {
		return err
	}
	if p.stop != nil && p.stop() {
		return ErrCancelled
	}
	p.signs[0] = ^uint64(0)
	for index, id := range p.graph.cis {
		p.signs[id] = p.inputPatterns[index]
	}
	for id, object := range p.graph.objects {
		if p.stop != nil && id&1023 == 0 && p.stop() {
			return ErrCancelled
		}
		if object.Type != AigAND {
			continue
		}
		left, right := p.signs[object.Fanin0.ID()], p.signs[object.Fanin1.ID()]
		if object.Fanin0.Complemented() {
			left = ^left
		}
		if object.Fanin1.Complemented() {
			right = ^right
		}
		p.signs[id] = left & right
	}
	for index, id := range p.graph.cos {
		if p.stop != nil && index&1023 == 0 && p.stop() {
			return ErrCancelled
		}
		object := p.graph.objects[id]
		value := p.signs[object.Fanin0.ID()]
		if object.Fanin0.Complemented() {
			value = ^value
		}
		p.signs[id] = value
	}
	return nil
}

func (p *AigPatternPack) Signs() []uint64         { return slices.Clone(p.signs) }
func (p *AigPatternPack) InputPatterns() []uint64 { return slices.Clone(p.inputPatterns) }
func (p *AigPatternPack) InputCare() []uint64     { return slices.Clone(p.inputCare) }

func (p *AigPatternPack) CountCares() int {
	total := 0
	for _, care := range p.inputCare {
		total += bits.OnesCount64(care)
	}
	return total
}

// RareNodes implements Aig_ManPackConstNodes. It includes signatures that are
// constant or differ from a constant in one packed assignment.
func (p *AigPatternPack) RareNodes() ([]AigID, error) {
	if err := p.check(); err != nil {
		return nil, err
	}
	var result []AigID
	for id, object := range p.graph.objects {
		if object.Type != AigAND {
			continue
		}
		sign := p.signs[id]
		if sign == 0 || sign == ^uint64(0) || sign&(sign-1) == 0 || (^sign)&(^sign-1) == 0 {
			result = append(result, AigID(id))
		}
	}
	return result, nil
}

// TryPattern implements Aig_ManPackAddPatternTry for one bit. A conflicting
// already-cared assignment leaves all state unchanged.
func (p *AigPatternPack) TryPattern(bit int, literals []SatLit) (bool, error) {
	if err := p.check(); err != nil {
		return false, err
	}
	if bit < 0 || bit >= 64 {
		return false, ErrDimension
	}
	if p.stop != nil && p.stop() {
		return false, ErrCancelled
	}
	for _, literal := range literals {
		variable := int(literal.Var())
		if variable < 0 || variable >= len(p.inputPatterns) {
			return false, ErrInvalidLiteral
		}
		mask := uint64(1) << bit
		if p.inputCare[variable]&mask != 0 && (p.inputPatterns[variable]&mask != 0) == literal.Negative() {
			return false, nil
		}
	}
	for _, literal := range literals {
		variable := int(literal.Var())
		mask := uint64(1) << bit
		p.inputCare[variable] |= mask
		if (p.inputPatterns[variable]&mask != 0) == literal.Negative() {
			p.inputPatterns[variable] ^= mask
		}
	}
	return true, nil
}

// AddPattern packs a partial CI assignment into the first compatible bit from
// 1 through 63. It returns false when all bits conflict.
func (p *AigPatternPack) AddPattern(literals []SatLit) (bool, error) {
	if p == nil {
		return false, ErrDimension
	}
	if p.patternsTotal == p.maxPatterns {
		return false, ErrLimit
	}
	for bit := 1; bit < 64; bit++ {
		added, err := p.TryPattern(bit, literals)
		if err != nil {
			return false, err
		}
		if added {
			p.patternsTotal++
			return true, nil
		}
	}
	p.patternsTotal++
	p.patternsSkipped++
	return false, nil
}

func (p *AigPatternPack) Stats() (AigPatternPackStats, error) {
	var result AigPatternPackStats
	if err := p.check(); err != nil {
		return result, err
	}
	result.PatternsTotal = p.patternsTotal
	result.PatternsSkipped = p.patternsSkipped
	result.PatternsRepeated = p.patternsRepeated
	result.CareBits = p.CountCares()
	for id, object := range p.graph.objects {
		if object.Type != AigAND {
			continue
		}
		count := bits.OnesCount64(p.signs[id])
		if count > 32 {
			count = 64 - count
		}
		result.NodeHistogram[count]++
	}
	return result, nil
}
