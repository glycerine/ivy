package kit

// Packed MUX records have fixed portable bit positions, replacing the C union.
type Kit_Mux_t struct{ v, t, e, c, i uint32 }

func Kit_Mux2Int(m Kit_Mux_t) int32 {
	if m.v >= 32 || m.t >= 4096 || m.e >= 4096 || m.c > 1 || m.i > 1 {
		panic("kit: MUX field overflow")
	}
	return int32(m.v | m.t<<5 | m.e<<17 | m.c<<29 | m.i<<30)
}
func Kit_Int2Mux(value int32) Kit_Mux_t {
	v := uint32(value)
	return Kit_Mux_t{v & 31, v >> 5 & 4095, v >> 17 & 4095, v >> 29 & 1, v >> 30 & 1}
}
func Kit_TruthToCloud5_rec(dd *CloudManager, truth uint32, nVars, total int32) *CloudNode {
	return Kit_TruthToCloud_rec(dd, []uint32{truth}, nVars, total)
}
func Kit_TruthToCloud_rec(dd *CloudManager, truth []uint32, nVars, total int32) *CloudNode {
	if Kit_TruthIsConst0(truth, nVars) != 0 {
		return dd.zero
	}
	if Kit_TruthIsConst1(truth, nVars) != 0 {
		return dd.one
	}
	if nVars == 0 {
		panic("kit: nonconstant zero-variable truth table")
	}
	low, high := kitIsopCofs(truth, nVars-1)
	if Kit_TruthIsEqual(low, high, nVars-1) != 0 {
		return Kit_TruthToCloud_rec(dd, low, nVars-1, total)
	}
	a := Kit_TruthToCloud_rec(dd, low, nVars-1, total)
	if a == nil {
		return nil
	}
	var b *CloudNode
	if Kit_TruthIsOpposite(low, high, nVars-1) != 0 {
		b = Cloud_Not(a)
	} else {
		b = Kit_TruthToCloud_rec(dd, high, nVars-1, total)
	}
	if b == nil {
		return nil
	}
	return Cloud_MakeNode(dd, uint32(total-nVars), b, a)
}
func Kit_TruthToCloud(dd *CloudManager, truth []uint32, nVars int32) *CloudNode {
	if nVars < 0 || nVars > dd.nVars {
		panic("kit: cloud variable count out of bounds")
	}
	return Kit_TruthToCloud_rec(dd, truth, nVars, nVars)
}
func Kit_CreateCloud(dd *CloudManager, function *CloudNode, nodes *[]int32) int32 {
	if function == nil {
		return 0
	}
	n := Cloud_DagCollect(dd, function)
	if n >= 4096 {
		return 0
	}
	*nodes = (*nodes)[:0]
	indices := map[*CloudNode]uint32{}
	for i, node := range dd.ppNodes {
		indices[node] = uint32(i)
		if i == 0 {
			record := Kit_Mux_t{}
			if n == 1 && Cloud_IsComplement(function) {
				record.i = 1
			}
			*nodes = append(*nodes, Kit_Mux2Int(record))
			continue
		}
		record := Kit_Mux_t{v: Cloud_V(node), t: indices[Cloud_Regular(node.t)], e: indices[Cloud_Regular(node.e)]}
		if Cloud_IsComplement(node.e) {
			record.c = 1
		}
		if i == int(n)-1 && Cloud_IsComplement(function) {
			record.i = 1
		}
		*nodes = append(*nodes, Kit_Mux2Int(record))
	}
	return 1
}
func Kit_CreateCloudFromTruth(dd *CloudManager, truth []uint32, nVars int32, nodes *[]int32) int32 {
	Cloud_Restart(dd)
	function := Kit_TruthToCloud(dd, truth, nVars)
	*nodes = (*nodes)[:0]
	return Kit_CreateCloud(dd, function, nodes)
}
func kitCloudStore(store *[][]uint32, count int, nVars int32) {
	words := int(Kit_TruthWordNum(nVars))
	for len(*store) < count {
		*store = append(*store, nil)
	}
	for i := 0; i < count; i++ {
		if cap((*store)[i]) < words {
			(*store)[i] = make([]uint32, words)
		} else {
			(*store)[i] = (*store)[i][:words]
		}
	}
}
func Kit_CloudToTruth(nodes []int32, nVars int32, store *[][]uint32, inverse int32) []uint32 {
	if len(nodes) == 0 {
		panic("kit: empty MUX program")
	}
	kitCloudStore(store, len(nodes), nVars)
	Kit_TruthFill((*store)[0], nVars)
	last := Kit_Int2Mux(nodes[0])
	for i := 1; i < len(nodes); i++ {
		m := Kit_Int2Mux(nodes[i])
		if int(m.e) >= i || int(m.t) >= i || m.v >= uint32(nVars) {
			panic("kit: invalid MUX topology")
		}
		v := nVars - 1 - int32(m.v)
		if inverse != 0 {
			v = int32(m.v)
		}
		Kit_TruthMuxVarPhase((*store)[i], (*store)[m.e], (*store)[m.t], nVars, v, int32(m.c))
		last = m
	}
	out := (*store)[len(nodes)-1]
	if last.i != 0 {
		Kit_TruthNot(out, out, nVars)
	}
	return out
}
func Kit_TruthCompose(dd *CloudManager, truth []uint32, nVars int32, inputs [][]uint32, total int32, store *[][]uint32, nodes *[]int32) []uint32 {
	if len(inputs) < int(nVars) {
		panic("kit: insufficient composition inputs")
	}
	if Kit_CreateCloudFromTruth(dd, truth, nVars, nodes) == 0 {
		return nil
	}
	kitCloudStore(store, len(*nodes), total)
	Kit_TruthFill((*store)[0], total)
	last := Kit_Int2Mux((*nodes)[0])
	for i := 1; i < len(*nodes); i++ {
		m := Kit_Int2Mux((*nodes)[i])
		Kit_TruthMuxPhase((*store)[i], (*store)[m.e], (*store)[m.t], inputs[nVars-1-int32(m.v)], total, int32(m.c))
		last = m
	}
	out := (*store)[len(*nodes)-1]
	if last.i != 0 {
		Kit_TruthNot(out, out, total)
	}
	return out
}
func Kit_TruthCofSupports(direct, inverse []int32, nVars int32, memory *[]uint32, supports []uint32) {
	width := 2 * nVars
	if len(supports) < int(width) {
		panic("kit: short cofactor support output")
	}
	run := func(nodes []int32, reverse bool) []uint32 {
		if len(nodes) == 0 {
			panic("kit: empty cofactor MUX program")
		}
		*memory = make([]uint32, int(width)*len(nodes))
		var out []uint32
		if width == 0 {
			return nil
		}
		out = (*memory)[:width]
		for i := 1; i < len(nodes); i++ {
			m := Kit_Int2Mux(nodes[i])
			v := int32(m.v)
			if reverse {
				v = nVars - 1 - v
			}
			lo := (*memory)[width*int32(m.e) : width*(int32(m.e)+1)]
			hi := (*memory)[width*int32(m.t) : width*(int32(m.t)+1)]
			out = (*memory)[width*int32(i) : width*int32(i+1)]
			for j := range out {
				out[j] = lo[j] | hi[j] | 1<<uint(v)
			}
			out[2*v] = lo[2*v]
			out[2*v+1] = hi[2*v]
		}
		return append([]uint32(nil), out...)
	}
	copy(supports, run(direct, true))
	other := run(inverse, false)
	for v := int32(0); v < width; v++ {
		mask := Kit_BitMask(v / 2)
		supports[v] = supports[v]&mask | other[v]&^mask
	}
}
