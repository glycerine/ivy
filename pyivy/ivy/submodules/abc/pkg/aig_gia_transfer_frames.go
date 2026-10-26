package abc

// Signal mapping from src/aig/gia/giaEnable.c:Gia_ManTransferFrames.

// TransferGiaFrameSignals composes an unrolling's source-to-frame mapping with
// frameToTarget and returns unique target IDs in signal-major/frame-major
// order. Literal phase is discarded, matching C. Invalid/missing copies and
// target constant-zero literals are skipped; constant-one maps to object zero.
// source must be the unrolling's source graph; numeric mappings cannot detect
// a different same-sized graph. Zero maxSignals selects one million results.
func TransferGiaFrameSignals(source *GiaManager, frames *GiaUnrolling, target *GiaManager, frameToTarget []GiaLit, signals []GiaID, maxSignals int) ([]GiaID, error) {
	if frames == nil || maxSignals < 0 {
		return nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, err
	}
	if err := target.Check(); err != nil {
		return nil, err
	}
	if frames.graph == nil || frames.sourceObjects != source.ObjectCount() || len(frameToTarget) != frames.graph.ObjectCount() {
		return nil, ErrDimension
	}
	if maxSignals == 0 {
		maxSignals = 1 << 20
	}
	if len(signals) > maxSignals {
		return nil, ErrLimit
	}
	for _, id := range signals {
		if uint64(id) >= uint64(source.ObjectCount()) || (!source.objects[id].IsCI() && !source.objects[id].IsAnd()) {
			return nil, ErrInvalidLiteral
		}
	}
	result := make([]GiaID, 0)
	seen := make(map[GiaID]bool)
	for _, signal := range signals {
		for frame := 0; frame < frames.frames; frame++ {
			copy := frames.mapping[frame*frames.sourceObjects+int(signal)]
			if copy == InvalidGiaLit {
				continue
			}
			if uint64(copy.ID()) >= uint64(len(frameToTarget)) {
				return nil, ErrInvariant
			}
			mapped := frameToTarget[copy.ID()]
			if mapped == InvalidGiaLit || mapped == GiaConst0 {
				continue
			}
			if !mapped.Valid() || uint64(mapped.ID()) >= uint64(target.ObjectCount()) {
				return nil, ErrInvalidLiteral
			}
			id := mapped.ID()
			if seen[id] {
				continue
			}
			if len(result) == maxSignals {
				return nil, ErrLimit
			}
			seen[id] = true
			result = append(result, id)
		}
	}
	return result, nil
}
