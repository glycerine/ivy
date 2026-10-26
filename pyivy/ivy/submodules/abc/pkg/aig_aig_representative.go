package abc

// Owned representative tables from src/aig/aig/aigRepr.c.

import (
	"slices"
	"sync"
)

type AigRepresentativeOptions struct {
	MaxObjects int
	MaxWork    int
	MaxBytes   int
	Stop       func() bool
}

type AigRepresentatives struct {
	mu             sync.RWMutex
	representative []AigID
	objectTypes    []AigType
	phases         []bool
	maxWork        int
	stop           func() bool
}

var aigCompiledSourceInventory = [...]string{
	"aigCheck.c", "aigCanon.c", "aigCuts.c", "aigDfs.c", "aigDup.c",
	"aigFanout.c", "aigFrames.c", "aigInter.c", "aigJust.c", "aigMan.c",
	"aigMem.c", "aigMffc.c", "aigObj.c", "aigOper.c", "aigOrder.c",
	"aigPack.c", "aigPart.c", "aigPartReg.c", "aigPartSat.c", "aigRepr.c",
	"aigRet.c", "aigRetF.c", "aigScl.c", "aigShow.c", "aigSplit.c",
	"aigTable.c", "aigTiming.c", "aigTruth.c", "aigTsim.c", "aigUtil.c",
	"aigWin.c",
}

func AigCompiledSourceInventory() []string {
	result := make([]string, len(aigCompiledSourceInventory))
	copy(result, aigCompiledSourceInventory[:])
	return result
}

func normalizeAigRepresentativeOptions(options AigRepresentativeOptions) (AigRepresentativeOptions, error) {
	if options.MaxObjects < 0 || options.MaxWork < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxWork == 0 {
		options.MaxWork = 1 << 20
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	return options, nil
}

func NewAigRepresentatives(graph *AigManager, options AigRepresentativeOptions) (*AigRepresentatives, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	options, err := normalizeAigRepresentativeOptions(options)
	if err != nil {
		return nil, err
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	count := graph.ObjectCount()
	if count > options.MaxObjects || count > options.MaxBytes/16 {
		return nil, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	representatives := make([]AigID, count)
	types := make([]AigType, count)
	phases := make([]bool, count)
	for index, object := range graph.objects {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if index >= options.MaxWork {
			return nil, ErrLimit
		}
		if object.Type == AigBuffer {
			return nil, ErrDimension
		}
		representatives[index] = InvalidAigID
		types[index] = object.Type
		phases[index] = object.Phase
	}
	return &AigRepresentatives{representative: representatives, objectTypes: types, phases: phases, maxWork: options.MaxWork, stop: options.Stop}, nil
}

func (table *AigRepresentatives) validObjectLocked(id AigID) bool {
	return table != nil && uint64(id) < uint64(len(table.representative)) && table.objectTypes[id] != AigNone && table.objectTypes[id] != AigCO && table.objectTypes[id] != AigBuffer
}

// SetRepresentative implements Aig_ObjCreateRepr's strict lower-ID contract.
func (table *AigRepresentatives) SetRepresentative(representative, object AigID) error {
	if table == nil {
		return ErrDimension
	}
	table.mu.Lock()
	defer table.mu.Unlock()
	if !table.validObjectLocked(representative) || !table.validObjectLocked(object) {
		return ErrInvalidLiteral
	}
	if representative >= object {
		return ErrDimension
	}
	if table.stop != nil && table.stop() {
		return ErrCancelled
	}
	table.representative[object] = representative
	return nil
}

// SetLowerPair implements Aig_ObjSetRepr_: the smaller ID represents the
// larger, while identical IDs are a no-op.
func (table *AigRepresentatives) SetLowerPair(first, second AigID) error {
	if table == nil {
		return ErrDimension
	}
	if first == second {
		table.mu.RLock()
		defer table.mu.RUnlock()
		if !table.validObjectLocked(first) {
			return ErrInvalidLiteral
		}
		return nil
	}
	if first > second {
		first, second = second, first
	}
	return table.SetRepresentative(first, second)
}

func (table *AigRepresentatives) Representative(object AigID) (AigID, bool, error) {
	if table == nil {
		return InvalidAigID, false, ErrDimension
	}
	table.mu.RLock()
	defer table.mu.RUnlock()
	if !table.validObjectLocked(object) {
		return InvalidAigID, false, ErrInvalidLiteral
	}
	representative := table.representative[object]
	return representative, representative != InvalidAigID, nil
}

func (table *AigRepresentatives) TransitiveRepresentative(object AigID) (AigID, bool, error) {
	if table == nil {
		return InvalidAigID, false, ErrDimension
	}
	table.mu.RLock()
	defer table.mu.RUnlock()
	if !table.validObjectLocked(object) {
		return InvalidAigID, false, ErrInvalidLiteral
	}
	current := table.representative[object]
	if current == InvalidAigID {
		return InvalidAigID, false, nil
	}
	for work := 0; ; work++ {
		if table.stop != nil && table.stop() {
			return InvalidAigID, false, ErrCancelled
		}
		if work >= table.maxWork {
			return InvalidAigID, false, ErrLimit
		}
		next := table.representative[current]
		if next == InvalidAigID {
			return current, true, nil
		}
		if next >= current {
			return InvalidAigID, false, ErrInvariant
		}
		current = next
	}
}

func (table *AigRepresentatives) Clear(object AigID) error {
	if table == nil {
		return ErrDimension
	}
	table.mu.Lock()
	defer table.mu.Unlock()
	if !table.validObjectLocked(object) {
		return ErrInvalidLiteral
	}
	if table.stop != nil && table.stop() {
		return ErrCancelled
	}
	table.representative[object] = InvalidAigID
	return nil
}

func (table *AigRepresentatives) Count() int {
	if table == nil {
		return 0
	}
	table.mu.RLock()
	defer table.mu.RUnlock()
	count := 0
	for _, representative := range table.representative {
		if representative != InvalidAigID {
			count++
		}
	}
	return count
}

func (table *AigRepresentatives) Snapshot() []AigID {
	if table == nil {
		return nil
	}
	table.mu.RLock()
	defer table.mu.RUnlock()
	return slices.Clone(table.representative)
}

// ResolveMapped implements Aig_ObjGetRepr plus an incoming edge complement.
// mapped contains target-manager literals indexed by source object ID.
func (table *AigRepresentatives) ResolveMapped(source AigLit, mapped []AigLit) (AigLit, error) {
	if table == nil || !source.Valid() {
		return InvalidAigLit, ErrInvalidLiteral
	}
	table.mu.RLock()
	defer table.mu.RUnlock()
	id := source.ID()
	if !table.validObjectLocked(id) || len(mapped) != len(table.representative) {
		return InvalidAigLit, ErrDimension
	}
	targetID := id
	phase := false
	if representative := table.representative[id]; representative != InvalidAigID {
		targetID = representative
		phase = table.phases[id] != table.phases[representative]
	}
	result := mapped[targetID]
	if !result.Valid() {
		return InvalidAigLit, ErrInvalidLiteral
	}
	return result.NotIf(phase != source.Complemented()), nil
}
