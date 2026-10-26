package abc

// Tests for src/misc/nm/nmApi.c and src/misc/nm/nmTable.c.

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strconv"
	"testing"
)

type nameManagerOracle struct {
	Order          []int    `json:"order"`
	Typed          []int    `json:"typed"`
	Unique         []string `json:"unique"`
	Suffix         string   `json:"suffix"`
	ResizeChecksum int      `json:"resize_checksum"`
	Entries        int      `json:"entries"`
	NativeIDs      []int    `json:"native_ids"`
	NativeCount    int      `json:"native_count"`
}

func mustNameManagerID(t *testing.T, manager *NameManager, name string, kind int) int {
	t.Helper()
	id, found, err := manager.FindID(name, kind)
	if err != nil || !found {
		t.Fatal(name, kind, id, found, err)
	}
	return id
}

func TestNameManagerNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/nm.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle nameManagerOracle
	if err = json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}

	order, _ := NewNameManager(1)
	for _, id := range []int{10, 20, 30} {
		if _, err = order.Store(id, 1, "dup", ""); err != nil {
			t.Fatal(err)
		}
	}
	gotOrder := []int{mustNameManagerID(t, order, "dup", -1)}
	_ = order.Delete(10)
	gotOrder = append(gotOrder, mustNameManagerID(t, order, "dup", -1))
	_ = order.Delete(30)
	gotOrder = append(gotOrder, mustNameManagerID(t, order, "dup", -1))
	_, _ = order.Store(40, 1, "dup", "")
	_ = order.Delete(20)
	gotOrder = append(gotOrder, mustNameManagerID(t, order, "dup", -1))
	if !slices.Equal(gotOrder, oracle.Order) {
		t.Fatal(gotOrder, oracle.Order)
	}

	typed, _ := NewNameManager(1)
	for _, row := range []struct{ id, kind int }{{1, 1}, {2, 2}, {3, 1}} {
		_, _ = typed.Store(row.id, row.kind, "same", "")
	}
	gotTyped := []int{
		mustNameManagerID(t, typed, "same", -1),
		mustNameManagerID(t, typed, "same", 1),
		mustNameManagerID(t, typed, "same", 2),
	}
	_ = typed.Delete(1)
	gotTyped = append(gotTyped, mustNameManagerID(t, typed, "same", -1))
	if id, found, findErr := typed.FindIDTwoKinds("same", 7, 1); findErr != nil || !found {
		t.Fatal(id, found, findErr)
	} else {
		gotTyped = append(gotTyped, id)
	}
	if !slices.Equal(gotTyped, oracle.Typed) {
		t.Fatal(gotTyped, oracle.Typed)
	}

	unique, _ := NewNameManager(1)
	_, _ = unique.Store(100, 0, "n77", "")
	_, _ = unique.Store(101, 0, "n77_1", "")
	first, _ := unique.UniqueName(77)
	second, _ := unique.UniqueName(100)
	suffixed, _ := unique.Store(102, 3, "base", "_s")
	if !slices.Equal([]string{first, second}, oracle.Unique) || suffixed != oracle.Suffix {
		t.Fatal(first, second, suffixed, oracle)
	}

	large, _ := NewNameManager(1)
	for index := 0; index < 64; index++ {
		_, _ = large.Store(index, index%4, "entry_"+strconv.Itoa(index), "")
	}
	checksum := 0
	for index := 0; index < 64; index++ {
		checksum += mustNameManagerID(t, large, "entry_"+strconv.Itoa(index), index%4) * (index + 1)
	}
	ids, _ := large.NameIDs()
	wantIDs := make([]int, 64)
	for index := range wantIDs {
		wantIDs[index] = index
	}
	if checksum != oracle.ResizeChecksum || len(ids) != oracle.Entries || !slices.Equal(ids, wantIDs) {
		t.Fatal(checksum, len(ids), ids)
	}
	if oracle.NativeCount != len(oracle.NativeIDs) || oracle.NativeCount >= oracle.Entries {
		t.Fatal("fixture must expose lossy native enumeration", oracle.NativeCount, oracle.Entries)
	}

	for _, manager := range []*NameManager{large, unique, typed, order} {
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNameManagerLifecycleAndErrors(t *testing.T) {
	if manager, err := NewNameManager(-1); manager != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(manager, err)
	}
	manager, err := NewNameManager(0)
	if err != nil {
		t.Fatal(err)
	}
	if name, err := manager.Store(4, 2, "alpha", "_0"); name != "alpha_0" || err != nil {
		t.Fatal(name, err)
	}
	if name, err := manager.Store(4, 3, "other", ""); name != "" || !errors.Is(err, ErrInvariant) {
		t.Fatal(name, err)
	}
	if count, _ := manager.EntryCount(); count != 1 {
		t.Fatal(count)
	}
	if name, found, findErr := manager.FindName(4); name != "alpha_0" || !found || findErr != nil {
		t.Fatal(name, found, findErr)
	}
	if name, found, findErr := manager.FindName(99); name != "" || found || findErr != nil {
		t.Fatal(name, found, findErr)
	}
	if _, err = manager.Store(5, -1, "bad", ""); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = manager.Store(-1, 1, "bad", ""); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = manager.Store(5, 1, "bad\x00name", ""); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = manager.Store(5, 1, "bad", "\x00suffix"); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err = manager.FindID("alpha_0", -2); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err = manager.FindID("bad\x00name", -1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err = manager.FindIDTwoKinds("alpha_0", 2, -2); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err = manager.FindName(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = manager.UniqueName(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = manager.Delete(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = manager.Delete(99); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	profile, err := manager.Profile()
	if err != nil || profile != (NameManagerProfile{Entries: 1, UniqueNames: 1, MaxNamesakes: 1}) {
		t.Fatal(profile, err)
	}
	ids, _ := manager.NameIDs()
	ids[0] = 999
	idsAgain, _ := manager.NameIDs()
	if !slices.Equal(idsAgain, []int{4}) {
		t.Fatal(idsAgain)
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
	if err = manager.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = manager.EntryCount(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, _, err = (*NameManager)(nil).FindName(0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestNameManagerOrderAndProfile(t *testing.T) {
	manager, _ := NewNameManager(2)
	for _, row := range []struct {
		id, kind int
		name     string
	}{{8, 0, "x"}, {3, 1, "y"}, {5, 2, "x"}, {1, 3, "x"}} {
		_, _ = manager.Store(row.id, row.kind, row.name, "")
	}
	profile, err := manager.Profile()
	if err != nil || profile != (NameManagerProfile{Entries: 4, UniqueNames: 2, DuplicateGroups: 1, MaxNamesakes: 3}) {
		t.Fatal(profile, err)
	}
	_ = manager.Delete(3)
	_ = manager.Delete(8)
	_, _ = manager.Store(3, 4, "z", "")
	ids, _ := manager.NameIDs()
	if !slices.Equal(ids, []int{5, 1, 3}) {
		t.Fatal(ids)
	}
	if id := mustNameManagerID(t, manager, "x", -1); id != 5 {
		t.Fatal(id)
	}
}

type nameManagerModelEntry struct {
	kind int
	name string
}

type nameManagerModel struct {
	entries map[int]nameManagerModelEntry
	groups  map[string][]int
	order   []int
}

func newNameManagerModel() *nameManagerModel {
	return &nameManagerModel{entries: make(map[int]nameManagerModelEntry), groups: make(map[string][]int)}
}

func (model *nameManagerModel) store(id, kind int, name string) bool {
	if _, found := model.entries[id]; found {
		return false
	}
	model.entries[id] = nameManagerModelEntry{kind: kind, name: name}
	ids := model.groups[name]
	if len(ids) == 0 {
		model.groups[name] = []int{id}
	} else {
		model.groups[name] = append([]int{ids[0], id}, ids[1:]...)
	}
	model.order = append(model.order, id)
	return true
}

func (model *nameManagerModel) delete(id int) bool {
	entry, found := model.entries[id]
	if !found {
		return false
	}
	ids := model.groups[entry.name]
	position := slices.Index(ids, id)
	if len(ids) == 1 {
		delete(model.groups, entry.name)
	} else if position == 0 {
		model.groups[entry.name] = append([]int{ids[len(ids)-1]}, ids[1:len(ids)-1]...)
	} else {
		model.groups[entry.name] = append(slices.Clone(ids[:position]), ids[position+1:]...)
	}
	delete(model.entries, id)
	model.order = slices.Delete(model.order, slices.Index(model.order, id), slices.Index(model.order, id)+1)
	return true
}

func (model *nameManagerModel) find(name string, kind int) (int, bool) {
	for _, id := range model.groups[name] {
		if kind == -1 || model.entries[id].kind == kind {
			return id, true
		}
	}
	return 0, false
}

func (model *nameManagerModel) unique(id int) string {
	if entry, found := model.entries[id]; found {
		return entry.name
	}
	base := "n" + strconv.Itoa(id)
	if len(model.groups[base]) == 0 {
		return base
	}
	for suffix := 1; ; suffix++ {
		candidate := base + "_" + strconv.Itoa(suffix)
		if len(model.groups[candidate]) == 0 {
			return candidate
		}
	}
}

func FuzzNameManager(f *testing.F) {
	f.Add([]byte{0, 1, 2, 0, 0, 2, 1, 0, 2, 1, 2, 0, 1, 1, 0, 0})
	f.Add([]byte{0, 7, 0, 3, 0, 8, 1, 3, 3, 7, 0, 3, 5, 8, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 512 {
			data = data[:512]
		}
		names := []string{"a", "b", "n0", "n1", "same"}
		manager, err := NewNameManager(1)
		if err != nil {
			t.Fatal(err)
		}
		model := newNameManagerModel()
		for offset := 0; offset+3 < len(data); offset += 4 {
			op := data[offset] % 7
			id := int(data[offset+1] % 12)
			kind := int(data[offset+2] % 4)
			name := names[int(data[offset+3])%len(names)]
			if data[offset]&0x80 != 0 {
				name += "_s"
			}
			switch op {
			case 0:
				_, gotErr := manager.Store(id, kind, name, "")
				wantOK := model.store(id, kind, name)
				if (gotErr == nil) != wantOK {
					t.Fatal("store", offset, id, gotErr, wantOK)
				}
			case 1:
				gotErr := manager.Delete(id)
				wantOK := model.delete(id)
				if (gotErr == nil) != wantOK {
					t.Fatal("delete", offset, id, gotErr, wantOK)
				}
			case 2:
				queryKind := kind
				if data[offset]&0x40 != 0 {
					queryKind = -1
				}
				gotID, gotFound, gotErr := manager.FindID(name, queryKind)
				wantID, wantFound := model.find(name, queryKind)
				if gotErr != nil || gotID != wantID || gotFound != wantFound {
					t.Fatal("find", offset, gotID, gotFound, gotErr, wantID, wantFound)
				}
			case 3:
				got, gotErr := manager.UniqueName(id)
				if want := model.unique(id); gotErr != nil || got != want {
					t.Fatal("unique", offset, got, gotErr, want)
				}
			case 4:
				got, found, gotErr := manager.FindName(id)
				want, wantFound := model.entries[id]
				if gotErr != nil || found != wantFound || got != want.name {
					t.Fatal("name", offset, got, found, gotErr, want, wantFound)
				}
			case 5:
				first, second := kind, (kind+1)%4
				gotID, gotFound, gotErr := manager.FindIDTwoKinds(name, first, second)
				wantID, wantFound := model.find(name, first)
				if !wantFound {
					wantID, wantFound = model.find(name, second)
				}
				if gotErr != nil || gotID != wantID || gotFound != wantFound {
					t.Fatal("two kinds", offset, gotID, gotFound, gotErr, wantID, wantFound)
				}
			case 6:
				got, gotErr := manager.NameIDs()
				if gotErr != nil || !slices.Equal(got, model.order) {
					t.Fatal("ids", offset, got, gotErr, model.order)
				}
			}
			count, countErr := manager.EntryCount()
			ids, idsErr := manager.NameIDs()
			if countErr != nil || idsErr != nil || count != len(model.entries) || !slices.Equal(ids, model.order) {
				t.Fatal("state", offset, count, countErr, ids, idsErr, model.order)
			}
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
