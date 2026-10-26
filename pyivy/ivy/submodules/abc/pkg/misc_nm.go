package abc

// Name-manager API and table behavior from src/misc/nm/nmApi.c and nmTable.c.

import (
	"strconv"
	"strings"

	"github.com/glycerine/insdict"
)

type nameManagerEntry struct {
	id, kind int
	name     string
}

// NameManagerProfile is an owned deterministic replacement for native bucket
// histogram printing.
type NameManagerProfile struct {
	Entries, UniqueNames, DuplicateGroups, MaxNamesakes int
}

// NameManager owns unique ID-to-name entries and ordered name-to-ID groups.
// It must not be copied.
type NameManager struct {
	byID   *insdict.Dict[int, *nameManagerEntry]
	byName *insdict.Dict[string, []int]
	closed bool
}

func NewNameManager(capacity int) (*NameManager, error) {
	if capacity < 0 {
		return nil, ErrDimension
	}
	return &NameManager{
		byID:   insdict.NewDictSize[int, *nameManagerEntry](capacity),
		byName: insdict.NewDictSize[string, []int](capacity),
	}, nil
}

func (manager *NameManager) check() error {
	if manager == nil || manager.closed || manager.byID == nil || manager.byName == nil {
		return ErrInvariant
	}
	return nil
}

func (manager *NameManager) Close() error {
	if err := manager.check(); err != nil {
		return err
	}
	manager.byID.DeleteAll()
	manager.byName.DeleteAll()
	manager.byID = nil
	manager.byName = nil
	manager.closed = true
	return nil
}

func (manager *NameManager) EntryCount() (int, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	return manager.byID.Len(), nil
}

// Store adds one unique ID. Names may repeat across entries and kinds.
func (manager *NameManager) Store(id, kind int, name, suffix string) (string, error) {
	if err := manager.check(); err != nil {
		return "", err
	}
	if id < 0 || kind < 0 || strings.IndexByte(name, 0) >= 0 || strings.IndexByte(suffix, 0) >= 0 {
		return "", ErrDimension
	}
	if _, found := manager.byID.Get2(id); found {
		return "", ErrInvariant
	}
	fullName := name + suffix
	entry := &nameManagerEntry{id: id, kind: kind, name: fullName}
	ids, found := manager.byName.Get2(fullName)
	if !found {
		manager.byName.Set(fullName, []int{id})
	} else {
		ordered := make([]int, len(ids)+1)
		ordered[0] = ids[0]
		ordered[1] = id
		copy(ordered[2:], ids[1:])
		manager.byName.Set(fullName, ordered)
	}
	manager.byID.Set(id, entry)
	return fullName, nil
}

// Delete removes an ID and preserves native namesake-anchor rotation.
func (manager *NameManager) Delete(id int) error {
	if err := manager.check(); err != nil {
		return err
	}
	if id < 0 {
		return ErrDimension
	}
	entry, found := manager.byID.Get2(id)
	if !found {
		return ErrInvariant
	}
	ids, found := manager.byName.Get2(entry.name)
	if !found || len(ids) == 0 {
		return ErrInvariant
	}
	position := -1
	for index, candidate := range ids {
		if candidate == id {
			position = index
			break
		}
	}
	if position < 0 {
		return ErrInvariant
	}
	if len(ids) == 1 {
		manager.byName.Del(entry.name)
	} else if position == 0 {
		ordered := make([]int, 0, len(ids)-1)
		ordered = append(ordered, ids[len(ids)-1])
		ordered = append(ordered, ids[1:len(ids)-1]...)
		manager.byName.Set(entry.name, ordered)
	} else {
		ordered := make([]int, 0, len(ids)-1)
		ordered = append(ordered, ids[:position]...)
		ordered = append(ordered, ids[position+1:]...)
		manager.byName.Set(entry.name, ordered)
	}
	if !manager.byID.Del(id) {
		return ErrInvariant
	}
	return nil
}

func (manager *NameManager) FindName(id int) (string, bool, error) {
	if err := manager.check(); err != nil {
		return "", false, err
	}
	if id < 0 {
		return "", false, ErrDimension
	}
	entry, found := manager.byID.Get2(id)
	if !found {
		return "", false, nil
	}
	return entry.name, true, nil
}

// FindID returns the first ID in native namesake order matching kind. Kind -1
// accepts any kind.
func (manager *NameManager) FindID(name string, kind int) (int, bool, error) {
	if err := manager.check(); err != nil {
		return 0, false, err
	}
	if kind < -1 || strings.IndexByte(name, 0) >= 0 {
		return 0, false, ErrDimension
	}
	ids, found := manager.byName.Get2(name)
	if !found {
		return 0, false, nil
	}
	for _, id := range ids {
		entry, entryFound := manager.byID.Get2(id)
		if !entryFound {
			return 0, false, ErrInvariant
		}
		if kind == -1 || entry.kind == kind {
			return id, true, nil
		}
	}
	return 0, false, nil
}

func (manager *NameManager) FindIDTwoKinds(name string, firstKind, secondKind int) (int, bool, error) {
	if err := manager.check(); err != nil {
		return 0, false, err
	}
	if firstKind < -1 || secondKind < -1 || strings.IndexByte(name, 0) >= 0 {
		return 0, false, ErrDimension
	}
	id, found, err := manager.FindID(name, firstKind)
	if err != nil || found {
		return id, found, err
	}
	return manager.FindID(name, secondKind)
}

// UniqueName returns an existing ID's name or the first unused n<ID>[_N].
func (manager *NameManager) UniqueName(id int) (string, error) {
	if err := manager.check(); err != nil {
		return "", err
	}
	if id < 0 {
		return "", ErrDimension
	}
	if entry, found := manager.byID.Get2(id); found {
		return entry.name, nil
	}
	base := "n" + strconv.Itoa(id)
	if _, found := manager.byName.Get2(base); !found {
		return base, nil
	}
	for suffix := 1; ; suffix++ {
		candidate := base + "_" + strconv.Itoa(suffix)
		if _, found := manager.byName.Get2(candidate); !found {
			return candidate, nil
		}
	}
}

// NameIDs returns every live ID in insertion order. This intentionally fixes
// native Nm_ManReturnNameIds, which drops all but one ID per occupied bucket.
func (manager *NameManager) NameIDs() ([]int, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	result := make([]int, 0, manager.byID.Len())
	for id := range manager.byID.All() {
		result = append(result, id)
	}
	return result, nil
}

func (manager *NameManager) Profile() (NameManagerProfile, error) {
	if err := manager.check(); err != nil {
		return NameManagerProfile{}, err
	}
	profile := NameManagerProfile{Entries: manager.byID.Len(), UniqueNames: manager.byName.Len()}
	for _, ids := range manager.byName.All() {
		if len(ids) > 1 {
			profile.DuplicateGroups++
		}
		if len(ids) > profile.MaxNamesakes {
			profile.MaxNamesakes = len(ids)
		}
	}
	return profile, nil
}
