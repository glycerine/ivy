package gia

func Str_MuxRestructAreaThree(dst *Gia_Man_t, mux *Str_Mux_t, delays *Vec_Int_t, verbose int32) int32 {
	a, b := Str_MuxFanin(mux, 0), Str_MuxFanin(mux, 1)
	if mux.Copy != -1 {
		panic("gia: mux area node already processed")
	}
	mux.Copy = -2
	if a.Edge[2].Copy == b.Edge[2].Copy {
		return 0
	}
	lit := Gia_ManHashMuxReal(dst, mux.Edge[2].Copy, b.Edge[2].Copy, a.Edge[2].Copy)
	Str_ObjDelay(dst, lit>>1, mux.nLutSize, delays)
	a.Edge[2].Copy = lit
	b.Edge[2].Copy = lit
	return 0
}
func Str_MuxRestructArea_rec(dst *Gia_Man_t, tree []Str_Mux_t, root *Str_Mux_t, edge int32, delays *Vec_Int_t, verbose int32) int32 {
	if root.Edge[edge].Fan <= 0 {
		return 0
	}
	node := Str_MuxFanin(root, edge)
	a := Str_MuxRestructArea_rec(dst, tree, node, 0, delays, verbose)
	b := Str_MuxRestructArea_rec(dst, tree, node, 1, delays, verbose)
	if a+b < 2 {
		return 1 + a + b
	}
	if a+b == 2 {
		if a == 2 || b == 2 {
			branch := int32(0)
			if b == 2 {
				branch = 1
			}
			child := Str_MuxFanin(node, branch)
			left, right := Str_MuxHasFanin(child, 0), Str_MuxHasFanin(child, 1)
			if left == right {
				panic("gia: invalid mux area path")
			}
			path := []int32{2*child.Id + right, 2*node.Id + branch, 2*root.Id + edge}
			Str_MuxChangeOnce(tree, path, 0, 2, nil, dst, delays)
		}
		Str_MuxRestructAreaThree(dst, Str_MuxFanin(root, edge), delays, verbose)
		return 0
	}
	if a+b != 3 && a+b != 4 || a != 2 && b != 2 {
		panic("gia: invalid mux area subtree")
	}
	Str_MuxRestructAreaThree(dst, node, delays, verbose)
	return 0
}
func Str_MuxRestructureArea(dst *Gia_Man_t, ntk *Str_Ntk_t, start, count int32, delays *Vec_Int_t, lutSize, verbose int32) int32 {
	if count >= MAX_TREE || count < 2 {
		return -1
	}
	tree := make([]Str_Mux_t, count+1)
	Str_MuxCreate(tree, ntk, start, count, delays, lutSize)
	result := Str_MuxRestructArea_rec(dst, tree, &tree[0], 0, delays, verbose)
	if result < 0 || result > 2 {
		panic("gia: invalid mux area result")
	}
	return Str_MuxToGia_rec(dst, &tree[0], 0, delays)
}
func Str_MuxRestructure(dst *Gia_Man_t, ntk *Str_Ntk_t, start, count int32, delays *Vec_Int_t, lutSize, recursive, optArea, verbose int32) int32 {
	if optArea != 0 {
		if count < 2 {
			return Str_MuxRestructure1(dst, ntk, start, count, delays, lutSize, verbose)
		}
		return Str_MuxRestructureArea(dst, ntk, start, count, delays, lutSize, verbose)
	}
	if recursive != 0 {
		return Str_MuxRestructure2(dst, ntk, start, count, delays, lutSize, verbose)
	}
	return Str_MuxRestructure1(dst, ntk, start, count, delays, lutSize, verbose)
}
