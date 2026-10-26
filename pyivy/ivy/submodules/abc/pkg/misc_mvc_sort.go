package abc

// Concrete cover sorting from src/misc/mvc/mvcSort.c.

// MvcCubeOrder selects one of the five native MVC cube comparators without an
// interface or arbitrary callback.
type MvcCubeOrder uint8

const (
	MvcOrderInteger MvcCubeOrder = iota
	MvcOrderSizeAndInteger
	MvcOrderUnderMask
	MvcOrderOutsideMask
	MvcOrderOutsideAndUnderMask
)

func (order MvcCubeOrder) valid() bool {
	return order <= MvcOrderOutsideAndUnderMask
}

func mvcCompareForOrder(first, second, mask *MvcCube, order MvcCubeOrder) (int, error) {
	switch order {
	case MvcOrderInteger:
		return first.CompareInt(second)
	case MvcOrderSizeAndInteger:
		return first.CompareSizeAndInt(second)
	case MvcOrderUnderMask:
		return first.CompareIntUnderMask(second, mask)
	case MvcOrderOutsideMask:
		return first.CompareIntOutsideMask(second, mask)
	case MvcOrderOutsideAndUnderMask:
		return first.CompareIntOutsideAndUnderMask(second, mask)
	default:
		return 0, ErrDimension
	}
}

func mvcMergeSorted(left, right []*MvcCube, mask *MvcCube, order MvcCubeOrder) ([]*MvcCube, error) {
	merged := make([]*MvcCube, 0, len(left)+len(right))
	for len(left) > 0 && len(right) > 0 {
		comparison, err := mvcCompareForOrder(left[0], right[0], mask, order)
		if err != nil {
			return nil, err
		}
		// Native takes from the right list on equality.
		if comparison < 0 {
			merged = append(merged, left[0])
			left = left[1:]
		} else {
			merged = append(merged, right[0])
			right = right[1:]
		}
	}
	merged = append(merged, left...)
	merged = append(merged, right...)
	return merged, nil
}

func mvcMergeSort(cubes []*MvcCube, mask *MvcCube, order MvcCubeOrder) ([]*MvcCube, error) {
	if len(cubes) <= 1 {
		return append([]*MvcCube(nil), cubes...), nil
	}
	middle := len(cubes) / 2
	left, err := mvcMergeSort(cubes[:middle], mask, order)
	if err != nil {
		return nil, err
	}
	right, err := mvcMergeSort(cubes[middle:], mask, order)
	if err != nil {
		return nil, err
	}
	return mvcMergeSorted(left, right, mask, order)
}

// Sort reorders cover in place using the native right-biased merge ordering.
// Mask is required only for the three mask-based orders.
func (cover *MvcCover) Sort(order MvcCubeOrder, mask *MvcCube) error {
	if err := cover.check(); err != nil {
		return err
	}
	if !order.valid() {
		return ErrDimension
	}
	for _, cube := range cover.cubes {
		if err := cube.check(); err != nil {
			return err
		}
		if cube.cover != cover || !cube.listed {
			return ErrInvariant
		}
	}
	if order >= MvcOrderUnderMask {
		if err := mask.check(); err != nil {
			return err
		}
		if mask.cover.bits != cover.bits || len(mask.words) != cover.words {
			return ErrDimension
		}
	}
	if len(cover.cubes) <= 1 {
		return nil
	}
	sorted, err := mvcMergeSort(cover.cubes, mask, order)
	if err != nil {
		return err
	}
	copy(cover.cubes, sorted)
	return nil
}
