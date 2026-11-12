package gia

// Gia_ManFxTopoOrder_rec appends the transitive fanins before their owner.
func Gia_ManFxTopoOrder_rec(cubes *Vec_Wec_t, first, count, visited, order *Vec_Int_t, obj int32) {
	if Vec_IntEntry(visited, int(obj)) != 0 {
		panic("gia: repeated extraction order visit")
	}
	Vec_IntWriteEntry(visited, int(obj), 1)
	for c := int32(0); c < Vec_IntEntry(count, int(obj)); c++ {
		cube := Vec_WecEntry(cubes, int(Vec_IntEntry(first, int(obj))+c))
		if Vec_IntEntry(cube, 0) != obj {
			panic("gia: extraction cubes are not grouped by owner")
		}
		for i := 1; i < Vec_IntSize(cube); i++ {
			fanin := Vec_IntEntry(cube, i) >> 1
			if Vec_IntEntry(visited, int(fanin)) == 0 {
				Gia_ManFxTopoOrder_rec(cubes, first, count, visited, order, fanin)
			}
		}
	}
	Vec_IntPush(order, obj)
}

// Gia_ManFxTopoOrder returns nil when extraction introduced no new nodes.
// Otherwise it returns the original-sized input prefix followed by nodes in
// fanin-first order, and reports each owner's first cube and cube count.
func Gia_ManFxTopoOrder(cubes *Vec_Wec_t, nInputs, nStart int32, pFirst, pCount **Vec_Int_t) *Vec_Int_t {
	var nodeMax int32
	for i := 0; i < Vec_WecSize(cubes); i++ {
		cube := Vec_WecEntry(cubes, i)
		owner := Vec_IntEntry(cube, 0)
		if owner+1 > nodeMax {
			nodeMax = owner + 1
		}
	}
	if nodeMax == nStart {
		return nil
	}
	if nodeMax < nInputs {
		panic("gia: invalid extraction input count")
	}
	first, count := Vec_IntStart(int(nodeMax)), Vec_IntStart(int(nodeMax))
	for i := 0; i < Vec_WecSize(cubes); i++ {
		owner := Vec_IntEntry(Vec_WecEntry(cubes, i), 0)
		if owner < nInputs {
			panic("gia: input owns extraction cube")
		}
		if Vec_IntEntry(count, int(owner)) == 0 {
			Vec_IntWriteEntry(first, int(owner), int32(i))
		}
		Vec_IntAddToEntry(count, int(owner), 1)
	}
	order, visited := Vec_IntStart(int(nInputs)), Vec_IntStart(int(nodeMax))
	for i := int32(0); i < nInputs; i++ {
		Vec_IntWriteEntry(visited, int(i), 1)
	}
	for i := nInputs; i < nodeMax; i++ {
		if Vec_IntEntry(visited, int(i)) == 0 {
			Gia_ManFxTopoOrder_rec(cubes, first, count, visited, order, i)
		}
	}
	*pFirst, *pCount = first, count
	return order
}
