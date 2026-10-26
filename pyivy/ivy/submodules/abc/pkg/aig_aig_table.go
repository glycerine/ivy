package abc

// Structural-hash lookup and profiling from src/aig/aig/aigTable.c.

type AigHashOptions struct {
	MaxObjects int // zero: one million objects
	Stop       func() bool
}

type AigHashProfile struct {
	TableSize         int
	Entries           int
	BucketOccupancies []int
}

// LookupHashedAnd implements Aig_TableLookupTwo. It includes native's trivial
// Boolean cases and never creates an object.
func (graph *AigManager) LookupHashedAnd(left, right AigLit) (AigLit, bool, error) {
	return graph.LookupAnd(left, right)
}

func aigNativeHash(object AigObject, tableSize int) int {
	key := uint64(object.Fanin0.ID()) * 7937
	key ^= uint64(object.Fanin1.ID()) * 2971
	if object.Fanin0.Complemented() {
		key ^= 911
	}
	if object.Fanin1.Complemented() {
		key ^= 353
	}
	return int(key % uint64(tableSize))
}

// StructuralHashProfile is the owned, non-printing equivalent of
// Aig_TableProfile. The requested table size makes native bucket placement
// reproducible even though Go's runtime owns the actual map capacity.
func (graph *AigManager) StructuralHashProfile(tableSize int, options AigHashOptions) (AigHashProfile, error) {
	if graph == nil || tableSize <= 0 || options.MaxObjects < 0 {
		return AigHashProfile{}, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return AigHashProfile{}, err
	}
	maximum := options.MaxObjects
	if maximum == 0 {
		maximum = 1 << 20
	}
	if len(graph.objects) > maximum {
		return AigHashProfile{}, ErrLimit
	}
	result := AigHashProfile{TableSize: tableSize, BucketOccupancies: make([]int, tableSize)}
	for _, object := range graph.objects {
		if options.Stop != nil && options.Stop() {
			return AigHashProfile{}, ErrCancelled
		}
		if object.Type != AigAND {
			continue
		}
		result.BucketOccupancies[aigNativeHash(object, tableSize)]++
		result.Entries++
	}
	return result, nil
}
