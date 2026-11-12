package bsat

// satOrderUpdate repairs the max-activity heap after a variable's score rises.
func satOrderUpdate(heap, positions []int32, activity []uint64, v int32) {
	i := positions[v]
	if i < 0 || int(i) >= len(heap) {
		panic("variable not in decision heap")
	}
	x := heap[i]
	for i != 0 {
		parent := (i - 1) / 2
		if activity[x] <= activity[heap[parent]] {
			break
		}
		heap[i] = heap[parent]
		positions[heap[i]] = i
		i = parent
	}
	heap[i] = x
	positions[x] = i
}
func satOrderPop(heap, positions []int32, activity []uint64) ([]int32, int32) {
	if len(heap) == 0 {
		return heap, var_Undef
	}
	next := heap[0]
	last := len(heap) - 1
	x := heap[last]
	heap = heap[:last]
	positions[next] = -1
	if last > 0 {
		i := 0
		for child := 1; child < last; child = 2*i + 1 {
			if child+1 < last && activity[heap[child]] < activity[heap[child+1]] {
				child++
			}
			if activity[x] >= activity[heap[child]] {
				break
			}
			heap[i] = heap[child]
			positions[heap[i]] = int32(i)
			i = child
		}
		heap[i] = x
		positions[x] = int32(i)
	}
	return heap, next
}
