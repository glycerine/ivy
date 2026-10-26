package abc

// Standalone state-model and fuzz tests for src/misc/vec/vecAtt.h.

import (
	"errors"
	"slices"
	"testing"
)

type attributeTestManager struct {
	next         int
	started      []int
	freed        []int
	managerFreed bool
}

type attributeTestObject struct{ value int }

func TestAttributeTypesAndDirectVector(t *testing.T) {
	wantTypes := []AttributeType{
		AttributeNone, AttributeCopy, AttributeLocalAIG, AttributeLocalSOP,
		AttributeLocalBDD, AttributeGlobalAIG, AttributeGlobalSOP,
		AttributeGlobalBDD, AttributeLevel, AttributeReverseLevel,
		AttributeRetimeLag, AttributeFraig, AttributeMVVar, AttributeData1,
		AttributeData2,
	}
	for index, value := range wantTypes {
		if int(value) != index {
			t.Fatal(index, value)
		}
	}
	if AttributeTypeCount != 15 {
		t.Fatal(AttributeTypeCount)
	}

	manager := &attributeTestManager{}
	vector, err := NewAttributeVector[attributeTestObject](0, manager, nil, nil,
		func(manager *attributeTestManager, object *attributeTestObject) {
			manager.freed = append(manager.freed, object.value)
		}, AttributeVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if capacity, err := vector.Capacity(); err != nil || capacity != 16 {
		t.Fatal(capacity, err)
	}
	first, second, far := &attributeTestObject{1}, &attributeTestObject{2}, &attributeTestObject{99}
	if err = vector.Write(2, first); err != nil {
		t.Fatal(err)
	}
	if err = vector.Write(15, second); err != nil {
		t.Fatal(err)
	}
	if err = vector.Write(40, far); err != nil {
		t.Fatal(err)
	}
	if capacity, err := vector.Capacity(); err != nil || capacity != 50 {
		t.Fatal(capacity, err)
	}
	if err = vector.Grow(40); err != nil {
		t.Fatal(err)
	}
	if err = vector.Grow(60); err != nil {
		t.Fatal(err)
	}
	if capacity, err := vector.Capacity(); err != nil || capacity != 60 {
		t.Fatal(capacity, err)
	}
	if got, err := vector.Entry(2); err != nil || got != first {
		t.Fatal(got, err)
	}
	if got, err := vector.Entry(3); err != nil || got != nil {
		t.Fatal(got, err)
	}
	snapshot, err := vector.Snapshot()
	if err != nil || snapshot[2] != first || snapshot[15] != second || snapshot[40] != far {
		t.Fatal(snapshot, err)
	}
	snapshot[2] = nil
	if got, _ := vector.Entry(2); got == nil {
		t.Fatal("snapshot aliases vector slots")
	}
	if err = vector.FreeEntry(15); err != nil || !slices.Equal(manager.freed, []int{2}) {
		t.Fatal(manager.freed, err)
	}
	if err = vector.FreeEntry(1000); err != nil {
		t.Fatal(err)
	}
	if err = vector.Clear(); err != nil || !slices.Equal(manager.freed, []int{2, 1, 99}) {
		t.Fatal(manager.freed, err)
	}
	for _, value := range mustAttributeSnapshot(t, vector) {
		if value != nil {
			t.Fatal("clear retained an entry")
		}
	}
	retained, err := vector.Close(false)
	if err != nil || retained != manager || manager.managerFreed {
		t.Fatal(retained, manager, err)
	}
	if _, err = vector.Capacity(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func mustAttributeSnapshot(t *testing.T, vector *AttributeVector[attributeTestObject, attributeTestManager]) []*attributeTestObject {
	t.Helper()
	result, err := vector.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestAttributeVectorLazyLifecycleLimitsAndSourceFreeEntryRule(t *testing.T) {
	manager := &attributeTestManager{next: 10}
	vector, err := NewAttributeVector[attributeTestObject](2, manager,
		func(manager *attributeTestManager) { manager.managerFreed = true },
		func(manager *attributeTestManager) *attributeTestObject {
			value := manager.next
			manager.next++
			manager.started = append(manager.started, value)
			return &attributeTestObject{value}
		},
		func(manager *attributeTestManager, object *attributeTestObject) {
			manager.freed = append(manager.freed, object.value)
		}, AttributeVectorLimits{MaxEntries: 20})
	if err != nil {
		t.Fatal(err)
	}
	first, err := vector.Entry(0)
	again, againErr := vector.Entry(0)
	far, farErr := vector.Entry(5)
	if err != nil || againErr != nil || farErr != nil || first != again || first.value != 10 || far.value != 11 || !slices.Equal(manager.started, []int{10, 11}) {
		t.Fatal(first, again, far, manager.started, err, againErr, farErr)
	}
	if capacity, err := vector.Capacity(); err != nil || capacity != 15 {
		t.Fatal(capacity, err)
	}
	if err = vector.Write(1, &attributeTestObject{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = vector.FreeEntry(0); err != nil || !slices.Equal(manager.freed, []int{10}) {
		t.Fatal(manager.freed, err)
	}
	newFirst, err := vector.Entry(0)
	if err != nil || newFirst.value != 12 {
		t.Fatal(newFirst, err)
	}
	if _, err = vector.Entry(20); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if retained, err := vector.Close(true); err != nil || retained != nil || !manager.managerFreed || !slices.Equal(manager.freed, []int{10, 12, 11}) {
		t.Fatal(retained, manager, err)
	}

	freeCount := 0
	nilManagerVector, err := NewAttributeVector[attributeTestObject, attributeTestManager](1, nil, nil, nil,
		func(_ *attributeTestManager, _ *attributeTestObject) { freeCount++ }, AttributeVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = nilManagerVector.Write(0, &attributeTestObject{7}); err != nil {
		t.Fatal(err)
	}
	if err = nilManagerVector.FreeEntry(0); err != nil || freeCount != 0 {
		t.Fatal(freeCount, err)
	}
	if err = nilManagerVector.Write(0, &attributeTestObject{8}); err != nil {
		t.Fatal(err)
	}
	if err = nilManagerVector.Clear(); err != nil || freeCount != 1 {
		t.Fatal(freeCount, err)
	}

	for _, create := range []func() error{
		func() error {
			_, e := NewAttributeVector[int, int](-1, nil, nil, nil, nil, AttributeVectorLimits{})
			return e
		},
		func() error {
			_, e := NewAttributeVector[int, int](17, nil, nil, nil, nil, AttributeVectorLimits{MaxEntries: 16})
			return e
		},
		func() error {
			_, e := NewAttributeVector[int, int](1, nil, nil, nil, nil, AttributeVectorLimits{MaxEntries: -1})
			return e
		},
		func() error {
			_, e := NewAttributeVector[int, int](1, nil, nil, nil, nil, AttributeVectorLimits{Stop: func() bool { return true }})
			return e
		},
	} {
		if err = create(); err == nil {
			t.Fatal("expected constructor error")
		}
	}
	missingFree, err := NewAttributeVector[int](1, new(int), nil, nil, nil, AttributeVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = missingFree.Close(true); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzAttributeVectorAgainstSlotModel(f *testing.F) {
	f.Add([]byte{0, 1, 2, 16, 31, 255})
	f.Add([]byte{255, 255, 0, 17, 0})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 512 {
			operations = operations[:512]
		}
		vector, err := NewAttributeVector[attributeTestObject, struct{}](1, nil, nil, nil, nil, AttributeVectorLimits{MaxEntries: 256})
		if err != nil {
			t.Fatal(err)
		}
		model := make([]*attributeTestObject, 1)
		for step, operation := range operations {
			index := int(operation)
			switch step % 3 {
			case 0:
				value := &attributeTestObject{step}
				err = vector.Write(index, value)
				if err == nil {
					if index >= len(model) {
						capacity := 2 * len(model)
						if capacity <= index {
							capacity = index + 10
						}
						if capacity > 256 {
							capacity = 256
						}
						grown := make([]*attributeTestObject, capacity)
						copy(grown, model)
						model = grown
					}
					model[index] = value
				} else if !errors.Is(err, ErrLimit) {
					t.Fatal(step, index, err)
				}
			case 1:
				err = vector.FreeEntry(index)
				if err != nil {
					t.Fatal(step, index, err)
				}
				if index < len(model) {
					model[index] = nil
				}
			case 2:
				value, entryErr := vector.Entry(index)
				if index >= 256 {
					if !errors.Is(entryErr, ErrLimit) {
						t.Fatal(step, index, value, entryErr)
					}
					continue
				}
				if entryErr != nil {
					t.Fatal(step, index, entryErr)
				}
				if index >= len(model) {
					capacity := 2 * len(model)
					if capacity <= index {
						capacity = index + 10
					}
					if capacity > 256 {
						capacity = 256
					}
					grown := make([]*attributeTestObject, capacity)
					copy(grown, model)
					model = grown
				}
				if value != model[index] {
					t.Fatal(step, index, value, model[index])
				}
			}
			got, snapshotErr := vector.Snapshot()
			if snapshotErr != nil || !slices.Equal(got, model) {
				t.Fatal(step, got, model, snapshotErr)
			}
		}
	})
}
