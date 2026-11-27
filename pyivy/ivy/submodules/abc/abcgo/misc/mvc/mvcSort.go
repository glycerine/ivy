package mvc


// Mvc_CoverSort is mvcSort.c's merge sort of the cube list. On a tie the
// merge takes the second list's cube, so equal cubes are not kept in
// order; factored forms depend on it.
func Mvc_CoverSort(c *Mvc_Cover_t, m *Mvc_Cube_t, compare func(*Mvc_Cube_t, *Mvc_Cube_t, *Mvc_Cube_t) int) {
	n := c.Cubes.Items
	if n <= 1 {
		return
	}
	head := mvcCoverSortRec(c.Cubes.Head, n, m, compare)
	c.Cubes.Head = head
	tail := head
	for tail.Next != nil {
		tail = tail.Next
	}
	c.Cubes.Tail = tail
}

func mvcCoverSortRec(list *Mvc_Cube_t, n int, m *Mvc_Cube_t, compare func(*Mvc_Cube_t, *Mvc_Cube_t, *Mvc_Cube_t) int) *Mvc_Cube_t {
	if n == 1 {
		list.Next = nil
		return list
	}
	n1 := n / 2
	list2 := list
	for i := 0; i < n1; i++ {
		list2 = list2.Next
	}
	list1 := mvcCoverSortRec(list, n1, m, compare)
	list2 = mvcCoverSortRec(list2, n-n1, m, compare)
	// Mvc_CoverSortMerge
	var result *Mvc_Cube_t
	tail := &result
	for list1 != nil && list2 != nil {
		var cube *Mvc_Cube_t
		if compare(list1, list2, m) < 0 {
			cube, list1 = list1, list1.Next
		} else {
			cube, list2 = list2, list2.Next
		}
		*tail = cube
		tail = &cube.Next
	}
	if list1 != nil {
		*tail = list1
	} else {
		*tail = list2
	}
	return result
}
