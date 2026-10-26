package abc

// Owned direct/reverse level analysis from src/aig/aig/aigTiming.c.

import "slices"

type AigTimingOptions struct {
	MaxObjects int // zero: one million objects
	Stop       func() bool
}

// AigTiming owns a checked graph snapshot and its direct, reverse and required
// level state. Native manager-attached scratch vectors become ordinary owned
// slices; Refresh atomically replaces the complete snapshot.
type AigTiming struct {
	objects  []AigObject
	forward  []uint32
	reverse  []uint32
	maximum  int64
	increase int64
}

func buildAigTiming(graph *AigManager, increase int, options AigTimingOptions) (*AigTiming, error) {
	if graph == nil || increase < 0 || options.MaxObjects < 0 {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	maximumObjects := options.MaxObjects
	if maximumObjects == 0 {
		maximumObjects = 1 << 20
	}
	if len(graph.objects) > maximumObjects {
		return nil, ErrLimit
	}
	result := &AigTiming{
		objects:  slices.Clone(graph.objects),
		forward:  make([]uint32, len(graph.objects)),
		reverse:  make([]uint32, len(graph.objects)),
		increase: int64(increase),
	}
	visit := func() error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		return nil
	}
	for id, object := range graph.objects {
		if err := visit(); err != nil {
			return nil, err
		}
		if object.Type == AigBuffer {
			return nil, ErrInvariant
		}
		result.forward[id] = object.Level
		if object.Type == AigCO {
			result.maximum = max(result.maximum, int64(graph.objects[object.Fanin0.ID()].Level))
		}
	}
	if result.maximum > int64(^uint32(0))-int64(increase)-1 {
		return nil, ErrLimit
	}
	result.maximum += int64(increase)
	// Stable topological IDs let a descending pass reproduce native reverse
	// DFS. Every internal object has a minimum reverse level of one, including
	// dangling cones; CIs, COs and constant one retain zero.
	for id := len(graph.objects) - 1; id > 0; id-- {
		if err := visit(); err != nil {
			return nil, err
		}
		object := graph.objects[id]
		if object.Type != AigAND && object.Type != AigBuffer {
			continue
		}
		result.reverse[id] = max(result.reverse[id], 1)
		propagate := func(literal AigLit) {
			fanin := literal.ID()
			kind := graph.objects[fanin].Type
			if kind == AigAND || kind == AigBuffer {
				result.reverse[fanin] = max(result.reverse[fanin], result.reverse[id]+1)
			}
		}
		propagate(object.Fanin0)
		if object.Type == AigAND {
			propagate(object.Fanin1)
		}
	}
	return result, nil
}

// BuildAigTiming implements Aig_ManStartReverseLevels. increase reserves the
// same extra direct-level margin as native nMaxLevelIncrease.
func BuildAigTiming(graph *AigManager, increase int, options AigTimingOptions) (*AigTiming, error) {
	return buildAigTiming(graph, increase, options)
}

func (timing *AigTiming) ObjectCount() int {
	if timing == nil {
		return 0
	}
	return len(timing.objects)
}

func (timing *AigTiming) MaximumLevel() int64 {
	if timing == nil {
		return 0
	}
	return timing.maximum
}

func (timing *AigTiming) ForwardLevels() []uint32 {
	if timing == nil {
		return nil
	}
	return slices.Clone(timing.forward)
}

func (timing *AigTiming) ReverseLevels() []uint32 {
	if timing == nil {
		return nil
	}
	return slices.Clone(timing.reverse)
}

func (timing *AigTiming) validObject(id AigID) bool {
	return timing != nil && uint64(id) < uint64(len(timing.objects)) && timing.objects[id].Type != AigNone
}

func (timing *AigTiming) ReverseLevel(id AigID) (uint32, error) {
	if !timing.validObject(id) {
		return 0, ErrInvalidLiteral
	}
	return timing.reverse[id], nil
}

// ClearReverseLevel implements Aig_ObjClearReverseLevel on the owned snapshot.
func (timing *AigTiming) ClearReverseLevel(id AigID) error {
	if !timing.validObject(id) {
		return ErrInvalidLiteral
	}
	timing.reverse[id] = 0
	return nil
}

// RequiredLevel implements Aig_ObjRequiredLevel.
func (timing *AigTiming) RequiredLevel(id AigID) (int64, error) {
	if !timing.validObject(id) {
		return 0, ErrInvalidLiteral
	}
	return timing.maximum + 1 - int64(timing.reverse[id]), nil
}

// ReverseLevelNew implements Aig_ObjReverseLevelNew using the owned graph.
func (timing *AigTiming) ReverseLevelNew(id AigID) (uint32, error) {
	if !timing.validObject(id) {
		return 0, ErrInvalidLiteral
	}
	object := timing.objects[id]
	if object.Type != AigAND && object.Type != AigBuffer {
		return 0, ErrUnsupported
	}
	level := uint32(0)
	for fanoutID, fanout := range timing.objects {
		switch fanout.Type {
		case AigCO, AigBuffer:
			if fanout.Fanin0.ID() == id {
				level = max(level, timing.reverse[fanoutID])
			}
		case AigAND:
			if fanout.Fanin0.ID() == id || fanout.Fanin1.ID() == id {
				level = max(level, timing.reverse[fanoutID])
			}
		}
	}
	if level == ^uint32(0) {
		return 0, ErrLimit
	}
	return level + 1, nil
}

// Refresh represents both native incremental update routines. Because public
// AigManager edits are atomic and always restore all direct levels, rebuilding
// the owned timing snapshot is simpler and cannot expose partially updated
// level buckets.
func (timing *AigTiming) Refresh(graph *AigManager, options AigTimingOptions) error {
	if timing == nil {
		return ErrDimension
	}
	if timing.increase > int64(^uint(0)>>1) {
		return ErrLimit
	}
	fresh, err := buildAigTiming(graph, int(timing.increase), options)
	if err != nil {
		return err
	}
	*timing = *fresh
	return nil
}

// Verify implements the two native diagnostic verifiers without printing.
func (timing *AigTiming) Verify() error {
	if timing == nil || len(timing.objects) != len(timing.forward) || len(timing.objects) != len(timing.reverse) {
		return ErrDimension
	}
	for id, object := range timing.objects {
		if timing.forward[id] != object.Level {
			return ErrInvariant
		}
		if object.Type != AigAND && object.Type != AigBuffer {
			continue
		}
		expected, err := timing.ReverseLevelNew(AigID(id))
		if err != nil || timing.reverse[id] != expected {
			return ErrInvariant
		}
	}
	return nil
}
