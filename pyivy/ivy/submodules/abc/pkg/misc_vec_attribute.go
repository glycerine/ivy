package abc

// Managed object attributes from src/misc/vec/vecAtt.h.

type AttributeType uint8

const (
	AttributeNone AttributeType = iota
	AttributeCopy
	AttributeLocalAIG
	AttributeLocalSOP
	AttributeLocalBDD
	AttributeGlobalAIG
	AttributeGlobalSOP
	AttributeGlobalBDD
	AttributeLevel
	AttributeReverseLevel
	AttributeRetimeLag
	AttributeFraig
	AttributeMVVar
	AttributeData1
	AttributeData2
	AttributeTypeCount
)

type AttributeVectorLimits struct {
	MaxEntries int
	Stop       func() bool
}

type AttributeManagerFreeFunc[M any] func(*M)
type AttributeObjectStartFunc[T, M any] func(*M) *T
type AttributeObjectFreeFunc[T, M any] func(*M, *T)

// AttributeVector stores optional typed object attributes. A starter makes
// Entry lazy; vectors without a starter accept explicit Write operations.
type AttributeVector[T, M any] struct {
	values      []*T
	manager     *M
	freeManager AttributeManagerFreeFunc[M]
	startObject AttributeObjectStartFunc[T, M]
	freeObject  AttributeObjectFreeFunc[T, M]
	limits      AttributeVectorLimits
	closed      bool
}

func normalizeAttributeVectorLimits(limits AttributeVectorLimits) (AttributeVectorLimits, error) {
	if limits.MaxEntries < 0 {
		return limits, ErrDimension
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 1 << 24
	}
	return limits, nil
}

func NewAttributeVector[T, M any](initialSize int, manager *M, freeManager AttributeManagerFreeFunc[M], startObject AttributeObjectStartFunc[T, M], freeObject AttributeObjectFreeFunc[T, M], limits AttributeVectorLimits) (*AttributeVector[T, M], error) {
	limits, err := normalizeAttributeVectorLimits(limits)
	if err != nil || initialSize < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	capacity := initialSize
	if capacity == 0 {
		capacity = 16
	}
	if capacity > limits.MaxEntries {
		return nil, ErrLimit
	}
	return &AttributeVector[T, M]{
		values:      make([]*T, capacity),
		manager:     manager,
		freeManager: freeManager,
		startObject: startObject,
		freeObject:  freeObject,
		limits:      limits,
	}, nil
}

func (attributes *AttributeVector[T, M]) check() error {
	if attributes == nil {
		return ErrDimension
	}
	if attributes.closed || len(attributes.values) == 0 || len(attributes.values) > attributes.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}

func (attributes *AttributeVector[T, M]) grow(minimum int) error {
	if err := attributes.check(); err != nil {
		return err
	}
	if minimum < 0 {
		return ErrDimension
	}
	if minimum <= len(attributes.values) {
		return nil
	}
	if minimum > attributes.limits.MaxEntries {
		return ErrLimit
	}
	if attributes.limits.Stop != nil && attributes.limits.Stop() {
		return ErrCancelled
	}
	values := make([]*T, minimum)
	copy(values, attributes.values)
	attributes.values = values
	return nil
}

func (attributes *AttributeVector[T, M]) growForIndex(index int) error {
	if index < 0 {
		return ErrDimension
	}
	if index < len(attributes.values) {
		return nil
	}
	if index >= attributes.limits.MaxEntries {
		return ErrLimit
	}
	capacity := attributes.limits.MaxEntries
	if len(attributes.values) <= attributes.limits.MaxEntries/2 {
		capacity = 2 * len(attributes.values)
	}
	if capacity <= index {
		capacity = attributes.limits.MaxEntries
		if attributes.limits.MaxEntries-index >= 10 {
			capacity = index + 10
		}
	}
	return attributes.grow(capacity)
}

func (attributes *AttributeVector[T, M]) Grow(minimum int) error {
	return attributes.grow(minimum)
}

func (attributes *AttributeVector[T, M]) Capacity() (int, error) {
	if err := attributes.check(); err != nil {
		return 0, err
	}
	return len(attributes.values), nil
}

func (attributes *AttributeVector[T, M]) Manager() (*M, error) {
	if err := attributes.check(); err != nil {
		return nil, err
	}
	return attributes.manager, nil
}

// Snapshot owns its slice; nonnil entries retain the source's object identity.
func (attributes *AttributeVector[T, M]) Snapshot() ([]*T, error) {
	if err := attributes.check(); err != nil {
		return nil, err
	}
	return append([]*T(nil), attributes.values...), nil
}

func (attributes *AttributeVector[T, M]) Write(index int, value *T) error {
	if err := attributes.check(); err != nil {
		return err
	}
	if attributes.startObject != nil {
		return ErrInvariant
	}
	if err := attributes.growForIndex(index); err != nil {
		return err
	}
	attributes.values[index] = value
	return nil
}

func (attributes *AttributeVector[T, M]) Entry(index int) (*T, error) {
	if err := attributes.check(); err != nil {
		return nil, err
	}
	if err := attributes.growForIndex(index); err != nil {
		return nil, err
	}
	if attributes.values[index] == nil && attributes.startObject != nil {
		if attributes.limits.Stop != nil && attributes.limits.Stop() {
			return nil, ErrCancelled
		}
		attributes.values[index] = attributes.startObject(attributes.manager)
	}
	return attributes.values[index], nil
}

func (attributes *AttributeVector[T, M]) FreeEntry(index int) error {
	if err := attributes.check(); err != nil {
		return err
	}
	if index < 0 {
		return ErrDimension
	}
	if index >= len(attributes.values) {
		return nil
	}
	value := attributes.values[index]
	// Vec_AttFreeEntry uniquely conditions its callback on a nonnil manager.
	if value != nil && attributes.manager != nil && attributes.freeObject != nil {
		attributes.freeObject(attributes.manager, value)
	}
	attributes.values[index] = nil
	return nil
}

func (attributes *AttributeVector[T, M]) Clear() error {
	if err := attributes.check(); err != nil {
		return err
	}
	if attributes.freeObject != nil {
		for index, value := range attributes.values {
			if value != nil {
				attributes.freeObject(attributes.manager, value)
				attributes.values[index] = nil
			}
		}
		return nil
	}
	clear(attributes.values)
	return nil
}

// Close frees every present object. If freeManager is false, the returned
// manager retains caller ownership, matching Vec_AttFree's return value.
func (attributes *AttributeVector[T, M]) Close(freeManager bool) (*M, error) {
	if err := attributes.check(); err != nil {
		return nil, err
	}
	if freeManager && attributes.manager != nil && attributes.freeManager == nil {
		return nil, ErrInvariant
	}
	if err := attributes.Clear(); err != nil {
		return nil, err
	}
	manager := attributes.manager
	if freeManager {
		if manager != nil {
			attributes.freeManager(manager)
		}
		manager = nil
	}
	attributes.values = nil
	attributes.manager = nil
	attributes.closed = true
	return manager, nil
}
