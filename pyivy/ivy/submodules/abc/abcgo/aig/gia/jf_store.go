package gia

// A positive comparison means the new cut is better than the old one.
func Jf_CutCompareDelay(old, next *Jf_Cut_t) float32 {
	if old.Time != next.Time {
		return float32(old.Time - next.Time)
	}
	if old.pCut[0] != next.pCut[0] {
		return float32(old.pCut[0] - next.pCut[0])
	}
	if old.Flow < next.Flow-JF_EPSILON {
		return -1
	}
	if old.Flow > next.Flow+JF_EPSILON {
		return 1
	}
	return 0
}
func Jf_CutCompareArea(old, next *Jf_Cut_t) float32 {
	if old.Flow < next.Flow-JF_EPSILON {
		return -1
	}
	if old.Flow > next.Flow+JF_EPSILON {
		return 1
	}
	if old.pCut[0] != next.pCut[0] {
		return float32(old.pCut[0] - next.pCut[0])
	}
	if old.Time != next.Time {
		return float32(old.Time - next.Time)
	}
	return 0
}
func jfCutContains(base, cut *Jf_Cut_t, cutMin bool) bool {
	if base.pCut[0] < cut.pCut[0] || base.Sign&cut.Sign != cut.Sign {
		return false
	}
	if cutMin {
		return Jf_CutIsContained1(base.pCut[:], cut.pCut[:])
	}
	return Jf_CutIsContainedOrder(base.pCut[:], cut.pCut[:])
}
func jfObjCutFilter(store []*Jf_Cut_t, count int32, cutMin bool) bool {
	for k := int32(0); k < count; k++ {
		if jfCutContains(store[count], store[k], cutMin) {
			return false
		}
	}
	return true
}
func jfObjCutFilterBoth(store []*Jf_Cut_t, count int32) int32 {
	candidate := store[count]
	for k := int32(0); k < count; k++ {
		if jfCutContains(candidate, store[k], true) {
			candidate.pCut[0] = -1
			return count
		}
	}
	last := int32(0)
	for k := int32(0); k < count; k++ {
		if candidate.pCut[0] < store[k].pCut[0] && jfCutContains(store[k], candidate, true) {
			continue
		}
		store[last], store[k] = store[k], store[last]
		last++
	}
	if last < count {
		store[last], store[count] = store[count], store[last]
	}
	return last
}
func Jf_ObjSortCuts(store []*Jf_Cut_t, n int32) {
	for i := int32(0); i < n-1; i++ {
		best := i
		for j := i + 1; j < n; j++ {
			if store[j].pCut[0] < store[best].pCut[0] {
				best = j
			}
		}
		store[i], store[best] = store[best], store[i]
	}
}
func jfObjAddCutToStore(store []*Jf_Cut_t, count, max int32, cutMin bool, cmp func(*Jf_Cut_t, *Jf_Cut_t) float32) int32 {
	if count == 0 {
		return 1
	}
	if count == max && cmp(store[count-1], store[count]) <= 0 {
		return count
	}
	if count > max {
		panic("JF cut store exceeds capacity")
	}
	pivot := count - 1
	for ; pivot >= 0; pivot-- {
		if cmp(store[pivot], store[count]) < 0 {
			break
		}
	}
	for k := int32(0); k <= pivot; k++ {
		if jfCutContains(store[count], store[k], cutMin) {
			return count
		}
	}
	candidate := store[count]
	pivot++
	for k := count; k > pivot; k-- {
		store[k] = store[k-1]
	}
	store[pivot] = candidate
	count++
	last := pivot + 1
	for k := last; k < count; k++ {
		if jfCutContains(store[k], candidate, cutMin) {
			continue
		}
		store[last], store[k] = store[k], store[last]
		last++
	}
	if last == max+1 {
		return max
	}
	return last
}
