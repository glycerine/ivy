package abc

// Experimental structural isomorphism signatures from src/aig/gia/giaIso3.c.

import (
	"slices"
	"sort"
)

var giaIsoNodeSeeds = [6]uint32{0x04892ed6, 0xc2523d7d, 0xdc36cd2e, 0xf2db64f8, 0xde3126bb, 0xdebbdff0}
var giaIsoDirectionSeeds = [2]uint32{0x855ee0cf, 0x946e1b5f}
var giaIsoComplementSeeds = [2]uint32{0x8ba63e50, 0x14d87f02}

type GiaIsoOptions struct {
	MaxObjects    int
	MaxEntries    int
	MaxIterations int
	Stop          func() bool
}

type GiaIso3Result struct {
	Signatures   []uint32
	UniqueCounts []int
	Converged    bool
}

type GiaIso4Result struct {
	Levels [][]GiaID
	Values []uint32
}

type GiaIsoLevelComparison struct {
	Level                 int
	OnlyFirst, OnlySecond int
	Common                int
}

type GiaIsoComparison struct {
	First, Second GiaIso4Result
	Levels        []GiaIsoLevelComparison
}

func checkGiaIsoOptions(options GiaIsoOptions) (GiaIsoOptions, error) {
	if options.MaxObjects < 0 || options.MaxEntries < 0 || options.MaxIterations < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 24
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 500
	}
	return options, nil
}

func (m *GiaManager) checkIsoGraph(options GiaIsoOptions) error {
	if m == nil {
		return ErrDimension
	}
	if err := m.Check(); err != nil {
		return err
	}
	if len(m.objects) > options.MaxObjects {
		return ErrLimit
	}
	// These experiments inspect only the two packed fanins of every AND-like
	// object. Native XOR/mux encodings would therefore receive misleading
	// signatures, and mux controls would be omitted.
	if m.XorCount() != 0 || m.MuxCount() != 0 {
		return ErrUnsupported
	}
	if options.Stop != nil && options.Stop() {
		return ErrCancelled
	}
	return nil
}

func giaIsoUnique(values []uint32) int {
	seen := make(map[uint32]struct{}, len(values))
	for _, value := range values {
		seen[value] = struct{}{}
	}
	return len(seen)
}

// Iso3InitialSignatures returns Gia_Iso3Init's type/polarity seeds without
// borrowing GiaObject.Value scratch.
func (m *GiaManager) Iso3InitialSignatures(options GiaIsoOptions) ([]uint32, error) {
	options, err := checkGiaIsoOptions(options)
	if err != nil {
		return nil, err
	}
	if err = m.checkIsoGraph(options); err != nil {
		return nil, err
	}
	if len(m.objects) > options.MaxEntries {
		return nil, ErrLimit
	}
	result := make([]uint32, len(m.objects))
	for i, object := range m.objects {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		switch {
		case object.IsAnd():
			first, firstErr := object.Fanin0(GiaID(i))
			if firstErr != nil {
				return nil, firstErr
			}
			second, secondErr := object.Fanin1(GiaID(i))
			if secondErr != nil {
				return nil, secondErr
			}
			index := 0
			if first.Complemented() {
				index++
			}
			if second.Complemented() {
				index++
			}
			result[i] = giaIsoNodeSeeds[index]
		case object.IsCI():
			result[i] = giaIsoNodeSeeds[3]
		case object.IsCO():
			result[i] = giaIsoNodeSeeds[4]
		default:
			result[i] = giaIsoNodeSeeds[5]
		}
	}
	return result, nil
}

// Iso3Refine performs one simultaneous Gia_Iso3Compute round. Current
// signatures are also the additive base, matching the source's accumulated
// Value fields.
func (m *GiaManager) Iso3Refine(signatures []uint32, options GiaIsoOptions) ([]uint32, error) {
	options, err := checkGiaIsoOptions(options)
	if err != nil {
		return nil, err
	}
	if err = m.checkIsoGraph(options); err != nil {
		return nil, err
	}
	if len(signatures) != len(m.objects) {
		return nil, ErrDimension
	}
	if len(signatures) > options.MaxEntries {
		return nil, ErrLimit
	}
	next := slices.Clone(signatures)
	addEdge := func(object GiaID, fanin GiaLit) {
		phase := 0
		if fanin.Complemented() {
			phase = 1
		}
		next[object] += signatures[fanin.ID()] + giaIsoComplementSeeds[phase] + giaIsoDirectionSeeds[0]
		next[fanin.ID()] += signatures[object] + giaIsoComplementSeeds[phase] + giaIsoDirectionSeeds[1]
	}
	for i, object := range m.objects {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if !object.IsAnd() && !object.IsCO() {
			continue
		}
		id := GiaID(i)
		first, firstErr := object.Fanin0(id)
		if firstErr != nil {
			return nil, firstErr
		}
		addEdge(id, first)
		if object.IsAnd() {
			second, secondErr := object.Fanin1(id)
			if secondErr != nil {
				return nil, secondErr
			}
			addEdge(id, second)
		}
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	return next, nil
}

// AnalyzeIso3 runs the source's unique-count stopping rule for at most 500
// rounds by default. Convergence means two adjacent rounds have the same
// number of distinct signatures; it does not claim graph isomorphism.
func (m *GiaManager) AnalyzeIso3(options GiaIsoOptions) (GiaIso3Result, error) {
	options, err := checkGiaIsoOptions(options)
	if err != nil {
		return GiaIso3Result{}, err
	}
	signatures, err := m.Iso3InitialSignatures(options)
	if err != nil {
		return GiaIso3Result{}, err
	}
	if _, err = checkedSize(len(signatures), 1, options.MaxIterations, options.MaxEntries); err != nil {
		return GiaIso3Result{}, err
	}
	result := GiaIso3Result{Signatures: signatures}
	previous := -1
	for iteration := 0; iteration < options.MaxIterations; iteration++ {
		if options.Stop != nil && options.Stop() {
			return GiaIso3Result{}, ErrCancelled
		}
		unique := giaIsoUnique(signatures)
		result.UniqueCounts = append(result.UniqueCounts, unique)
		if unique == previous {
			result.Converged = true
			break
		}
		previous = unique
		signatures, err = m.Iso3Refine(signatures, options)
		if err != nil {
			return GiaIso3Result{}, err
		}
		result.Signatures = signatures
	}
	return result, nil
}

// AnalyzeIso4 applies Gia_Iso4Gia's fixed ABC random sequence along reverse
// levels. Values begin at zero because manager scratch is not part of the Go
// representation; CO values are overwritten exactly as in the source.
func (m *GiaManager) AnalyzeIso4(options GiaIsoOptions) (GiaIso4Result, error) {
	options, err := checkGiaIsoOptions(options)
	if err != nil {
		return GiaIso4Result{}, err
	}
	if err = m.checkIsoGraph(options); err != nil {
		return GiaIso4Result{}, err
	}
	if _, err = checkedSize(0, 2, len(m.objects), options.MaxEntries); err != nil {
		return GiaIso4Result{}, err
	}
	levels, err := m.ReverseLevelize(options.MaxObjects)
	if err != nil {
		return GiaIso4Result{}, err
	}
	result := GiaIso4Result{Levels: make([][]GiaID, len(levels)), Values: make([]uint32, len(m.objects))}
	for i := range levels {
		result.Levels[i] = slices.Clone(levels[i])
	}
	var random ABCRandom
	_ = random.Uint32(true) // Gia_Iso4Gia resets and discards this value.
	for level, objects := range levels {
		if options.Stop != nil && options.Stop() {
			return GiaIso4Result{}, ErrCancelled
		}
		randomComplement := [2]uint32{random.Uint32(false), random.Uint32(false)}
		for _, id := range objects {
			if options.Stop != nil && options.Stop() {
				return GiaIso4Result{}, ErrCancelled
			}
			object := m.objects[id]
			if level == 0 {
				if !object.IsCO() {
					return GiaIso4Result{}, ErrInvariant
				}
				result.Values[id] = random.Uint32(false)
				fanin, faninErr := object.Fanin0(id)
				if faninErr != nil {
					return GiaIso4Result{}, faninErr
				}
				phase := 0
				if fanin.Complemented() {
					phase = 1
				}
				result.Values[fanin.ID()] += result.Values[id] + randomComplement[phase]
				continue
			}
			if !object.IsAnd() {
				continue
			}
			first, firstErr := object.Fanin0(id)
			if firstErr != nil {
				return GiaIso4Result{}, firstErr
			}
			second, secondErr := object.Fanin1(id)
			if secondErr != nil {
				return GiaIso4Result{}, secondErr
			}
			firstPhase, secondPhase := 0, 0
			if first.Complemented() {
				firstPhase = 1
			}
			if second.Complemented() {
				secondPhase = 1
			}
			result.Values[first.ID()] += result.Values[id] + randomComplement[firstPhase]
			result.Values[second.ID()] += result.Values[id] + randomComplement[secondPhase]
		}
	}
	return result, nil
}

func giaIsoLevelValues(result GiaIso4Result, level int) ([]uint32, error) {
	if level < 0 || level >= len(result.Levels) {
		return nil, ErrDimension
	}
	values := make([]uint32, len(result.Levels[level]))
	for i, id := range result.Levels[level] {
		if uint64(id) >= uint64(len(result.Values)) {
			return nil, ErrInvariant
		}
		values[i] = result.Values[id]
	}
	return values, nil
}

func giaIsoCommon(first, second []uint32) int {
	first = slices.Clone(first)
	second = slices.Clone(second)
	sort.Slice(first, func(i, j int) bool { return first[i] < first[j] })
	sort.Slice(second, func(i, j int) bool { return second[i] < second[j] })
	i, j, common := 0, 0, 0
	for i < len(first) && j < len(second) {
		switch {
		case first[i] < second[j]:
			i++
		case second[j] < first[i]:
			j++
		default:
			i++
			j++
			common++
		}
	}
	return common
}

// CompareIso4 replaces Gia_IsoCompareVecs/Gia_Iso4TestTwo stdout with owned
// multiset counts for every shared reverse level.
func CompareIso4(first, second *GiaManager, options GiaIsoOptions) (GiaIsoComparison, error) {
	if first == nil || second == nil {
		return GiaIsoComparison{}, ErrDimension
	}
	left, err := first.AnalyzeIso4(options)
	if err != nil {
		return GiaIsoComparison{}, err
	}
	right, err := second.AnalyzeIso4(options)
	if err != nil {
		return GiaIsoComparison{}, err
	}
	result := GiaIsoComparison{First: left, Second: right}
	count := min(len(left.Levels), len(right.Levels))
	result.Levels = make([]GiaIsoLevelComparison, count)
	for level := 0; level < count; level++ {
		if options.Stop != nil && options.Stop() {
			return GiaIsoComparison{}, ErrCancelled
		}
		leftValues, err := giaIsoLevelValues(left, level)
		if err != nil {
			return GiaIsoComparison{}, err
		}
		rightValues, err := giaIsoLevelValues(right, level)
		if err != nil {
			return GiaIsoComparison{}, err
		}
		common := giaIsoCommon(leftValues, rightValues)
		result.Levels[level] = GiaIsoLevelComparison{Level: level, OnlyFirst: len(leftValues) - common, OnlySecond: len(rightValues) - common, Common: common}
	}
	return result, nil
}
