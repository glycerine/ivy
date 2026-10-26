package abc

// Standalone mapper primitives from mapperCut.c, mapperCutUtils.c,
// mapperMatch.c, mapperTime.c, mapperTree.c, and mapperUtils.c.

import (
	"math"
)

var mapperCompiledSources = [...]string{
	"mapper.c", "mapperCanon.c", "mapperCore.c", "mapperCreate.c",
	"mapperCut.c", "mapperCutUtils.c", "mapperLib.c", "mapperMatch.c",
	"mapperRefs.c", "mapperSuper.c", "mapperSwitch.c", "mapperTable.c",
	"mapperTime.c", "mapperTree.c", "mapperTruth.c", "mapperUtils.c",
	"mapperVec.c",
}

func MapperCompiledSources() []string {
	return append([]string(nil), mapperCompiledSources[:]...)
}

type MapperLimits struct {
	MaxWork    uint64
	MaxObjects int
	MaxBytes   uint64
	Stop       func() bool
}

type MapperTime struct {
	Rise  float64
	Fall  float64
	Worst float64
}

type MapperPinDelay struct {
	ToRise MapperTime
	ToFall MapperTime
}

type MapperSupergate struct {
	Area        float64
	FanoutLimit int
	Fanins      int
	PinDelays   []MapperPinDelay
}

type MapperMatch struct {
	Arrival  MapperTime
	AreaFlow float64
	Super    MapperSupergate
	Phase    uint32
}

type MapperMatchOrder struct {
	Epsilon   float64
	AreaFirst bool
}

type mapperBudget struct {
	limits MapperLimits
	work   uint64
	bytes  uint64
}

func (b *mapperBudget) poll() error {
	if b.limits.Stop != nil && b.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

func (b *mapperBudget) takeWork(n uint64) error {
	if ^uint64(0)-b.work < n {
		return ErrLimit
	}
	b.work += n
	if b.limits.MaxWork != 0 && b.work > b.limits.MaxWork {
		return ErrLimit
	}
	return b.poll()
}

func (b *mapperBudget) takeBytes(n uint64) error {
	if ^uint64(0)-b.bytes < n {
		return ErrLimit
	}
	b.bytes += n
	if b.limits.MaxBytes != 0 && b.bytes > b.limits.MaxBytes {
		return ErrLimit
	}
	return b.poll()
}

func mapperLimitsValid(l MapperLimits) bool {
	return l.MaxObjects >= 0
}

func mapperFinite(x float64) bool {
	return !math.IsNaN(x) && !math.IsInf(x, 0)
}

// MapperMergeCutLeaves is Map_CutMergeTwo over stable node numbers. Inputs
// must be strictly increasing, as mapper cut leaves are ordered by Node.Num.
func MapperMergeCutLeaves(a, c []int, maxLeaves int, limits MapperLimits) ([]int, error) {
	if !mapperLimitsValid(limits) || maxLeaves < 0 || maxLeaves > 6 || len(a) > maxLeaves || len(c) > maxLeaves {
		return nil, ErrDimension
	}
	if limits.MaxObjects != 0 && (len(a) > limits.MaxObjects || len(c) > limits.MaxObjects || maxLeaves > limits.MaxObjects) {
		return nil, ErrLimit
	}
	b := mapperBudget{limits: limits}
	if err := b.takeBytes(uint64(maxLeaves) * 8); err != nil {
		return nil, err
	}
	for _, leaves := range [][]int{a, c} {
		for i, id := range leaves {
			if err := b.takeWork(1); err != nil {
				return nil, err
			}
			if id < 0 || i > 0 && leaves[i-1] >= id {
				return nil, ErrDimension
			}
		}
	}
	out := make([]int, 0, maxLeaves)
	i, j := 0, 0
	for i < len(a) || j < len(c) {
		if err := b.takeWork(1); err != nil {
			return nil, err
		}
		var id int
		switch {
		case j == len(c) || i < len(a) && a[i] < c[j]:
			id, i = a[i], i+1
		case i == len(a) || c[j] < a[i]:
			id, j = c[j], j+1
		default:
			id, i, j = a[i], i+1, j+1
		}
		if len(out) == maxLeaves {
			return nil, ErrLimit
		}
		out = append(out, id)
	}
	return out, nil
}

// MapperElementaryTruths reproduces the two 32-bit words initialized by
// Map_MappingSetupTruthTables.  Functions with fewer than six variables are
// stretched over all 64 minterms, as they are in the mapper truth storage.
func MapperElementaryTruths(nVars int, limits MapperLimits) ([]uint64, error) {
	if !mapperLimitsValid(limits) || nVars < 0 || nVars > 6 {
		return nil, ErrDimension
	}
	if limits.MaxObjects != 0 && nVars > limits.MaxObjects {
		return nil, ErrLimit
	}
	b := mapperBudget{limits: limits}
	if err := b.takeBytes(uint64(nVars) * 8); err != nil {
		return nil, err
	}
	truths := make([]uint64, nVars)
	for v := 0; v < nVars; v++ {
		for m := 0; m < 64; m++ {
			if err := b.takeWork(1); err != nil {
				return nil, err
			}
			if m&(1<<uint(v)) != 0 {
				truths[v] |= uint64(1) << uint(m)
			}
		}
	}
	return truths, nil
}

// MapperTransformTruth applies the input-negation phase transformation used by
// Map_CalculatePhase and Map_CalculatePhase6 to the mapper's complete two-word
// truth representation.  Inactive dimensions are retained rather than
// silently discarding the stretched copies.
func MapperTransformTruth(truth uint64, nVars int, phase uint32, limits MapperLimits) (uint64, error) {
	if !mapperLimitsValid(limits) || nVars < 0 || nVars > 6 || phase>>uint(nVars) != 0 {
		return 0, ErrDimension
	}
	b := mapperBudget{limits: limits}
	var out uint64
	for m := 0; m < 64; m++ {
		if err := b.takeWork(1); err != nil {
			return 0, err
		}
		source := m ^ int(phase)
		if truth&(uint64(1)<<uint(source)) != 0 {
			out |= uint64(1) << uint(m)
		}
	}
	return out, nil
}

func MapperCutLeafPhase(phase uint32, leaf, leafCount int) (bool, error) {
	if leafCount < 0 || leafCount > 32 || leaf < 0 || leaf >= leafCount || phase>>uint(leafCount) != 0 {
		return false, ErrDimension
	}
	return phase&(1<<uint(leaf)) == 0, nil
}

// MapperCompareMatches returns true exactly when candidate is preferred over
// current by Map_MatchCompare's delay-first or area-first ordering.
func MapperCompareMatches(current, candidate MapperMatch, order MapperMatchOrder) (bool, error) {
	if !mapperFinite(order.Epsilon) || order.Epsilon < 0 || !mapperMatchValid(current) || !mapperMatchValid(candidate) {
		return false, ErrDimension
	}
	cmp := func(a, c float64) int {
		if a < c-order.Epsilon {
			return -1
		}
		if a > c+order.Epsilon {
			return 1
		}
		return 0
	}
	if order.AreaFirst {
		if c := cmp(current.AreaFlow, candidate.AreaFlow); c != 0 {
			return c > 0, nil
		}
		if c := cmp(current.Arrival.Worst, candidate.Arrival.Worst); c != 0 {
			return c > 0, nil
		}
	} else {
		if c := cmp(current.Arrival.Worst, candidate.Arrival.Worst); c != 0 {
			return c > 0, nil
		}
		if c := cmp(current.AreaFlow, candidate.AreaFlow); c != 0 {
			return c > 0, nil
		}
	}
	if current.Super.FanoutLimit != candidate.Super.FanoutLimit {
		return current.Super.FanoutLimit < candidate.Super.FanoutLimit, nil
	}
	if current.Super.Fanins != candidate.Super.Fanins {
		return current.Super.Fanins > candidate.Super.Fanins, nil
	}
	return false, nil
}

func mapperMatchValid(m MapperMatch) bool {
	return mapperTimeValid(m.Arrival) && mapperFinite(m.AreaFlow) && mapperFinite(m.Super.Area) && m.Super.Area >= 0 && m.Super.FanoutLimit >= 0 && m.Super.Fanins >= 0
}

func mapperTimeValid(t MapperTime) bool {
	return mapperFinite(t.Rise) && mapperFinite(t.Fall) && mapperFinite(t.Worst) &&
		t.Rise >= 0 && t.Fall >= 0 && t.Worst == math.Max(t.Rise, t.Fall)
}

// Mapper uses nonpositive pin delays to mark transitions which do not exist.
// Their Worst field is storage-compatible metadata and is not used by the
// source arrival calculation.
func mapperPinDelayValid(t MapperTime) bool {
	return mapperFinite(t.Rise) && mapperFinite(t.Fall) && mapperFinite(t.Worst)
}

// MapperCutArrival reproduces Map_TimeCutComputeArrival for one selected
// supergate. arrivals[i][0] is negative phase and arrivals[i][1] positive.
func MapperCutArrival(arrivals [][2]MapperTime, gate MapperSupergate, phase uint32, nodeDelay, worstLimit float64, limits MapperLimits) (MapperTime, error) {
	var result MapperTime
	if !mapperLimitsValid(limits) || len(arrivals) != gate.Fanins || len(gate.PinDelays) != gate.Fanins || gate.Fanins < 0 || gate.Fanins > 6 || phase>>uint(gate.Fanins) != 0 || !mapperFinite(nodeDelay) || nodeDelay < 0 || !mapperFinite(worstLimit) || worstLimit < 0 || !mapperFinite(gate.Area) || gate.Area < 0 || gate.FanoutLimit < 0 {
		return MapperTime{}, ErrDimension
	}
	if limits.MaxObjects != 0 && gate.Fanins > limits.MaxObjects {
		return MapperTime{}, ErrLimit
	}
	b := mapperBudget{limits: limits}
	// The returned value plus the source-shaped selected-arrival scratch.
	if err := b.takeBytes(24 + uint64(gate.Fanins)*24); err != nil {
		return MapperTime{}, err
	}
	for i := gate.Fanins - 1; i >= 0; i-- {
		if err := b.takeWork(4); err != nil {
			return MapperTime{}, err
		}
		for p := 0; p < 2; p++ {
			if !mapperTimeValid(arrivals[i][p]) {
				return MapperTime{}, ErrDimension
			}
		}
		pin := gate.PinDelays[i]
		if !mapperPinDelayValid(pin.ToRise) || !mapperPinDelayValid(pin.ToFall) {
			return MapperTime{}, ErrDimension
		}
		selected := 0
		if phase&(1<<uint(i)) == 0 {
			selected = 1
		}
		in := arrivals[i][selected]
		paths := [4]struct {
			arrival, delay float64
			rise           bool
		}{
			{in.Rise, pin.ToRise.Rise, true}, {in.Fall, pin.ToRise.Fall, true},
			{in.Rise, pin.ToFall.Rise, false}, {in.Fall, pin.ToFall.Fall, false},
		}
		for _, path := range paths {
			if path.delay <= 0 {
				continue
			}
			value := path.arrival + path.delay + nodeDelay
			if !mapperFinite(value) {
				return MapperTime{}, ErrDimension
			}
			if value > worstLimit {
				return MapperTime{}, ErrLimit
			}
			if path.rise && value > result.Rise {
				result.Rise = value
			}
			if !path.rise && value > result.Fall {
				result.Fall = value
			}
		}
	}
	result.Worst = math.Max(result.Rise, result.Fall)
	return result, nil
}
