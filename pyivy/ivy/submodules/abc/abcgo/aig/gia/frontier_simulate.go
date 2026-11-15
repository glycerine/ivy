package gia

// Frontier data pins are slots. MUX controls retain their original object IDs;
// each control object's Value identifies its slot when this node is visited.
// Only managers with real XORs normalize slot order to preserve the XOR tag.
func giaFrontSimulateWords(g *Gia_Man_t, obj *Gia_Obj_t, data []uint32, words int32) {
	slot := func(index uint32) []uint32 {
		start := int(index) * int(words)
		return data[start : start+int(words)]
	}
	dst, a, b := slot(obj.Value), slot(obj.iDiff0), slot(obj.iDiff1)
	// A complemented pin XORs every bit with all ones. Select the operation
	// once per node, keeping gate-kind and phase branches out of word loops.
	mask0, mask1 := uint32(0)-obj.fCompl0, uint32(0)-obj.fCompl1
	switch {
	case g.nMuxes != 0 && Gia_ObjIsMux(g, obj) != 0:
		lit := g.PMuxes[obj.id]
		control := slot(g.pObjs[lit>>1].Value)
		controlMask := uint32(0) - (lit & 1)
		for i := range dst {
			c := control[i] ^ controlMask
			dst[i] = (a[i]^mask0)&^c | (b[i]^mask1)&c
		}
	case g.nXors != 0 && Gia_ObjIsXor(obj) != 0:
		for i := range dst {
			dst[i] = a[i] ^ mask0 ^ b[i] ^ mask1
		}
	default:
		for i := range dst {
			dst[i] = (a[i] ^ mask0) & (b[i] ^ mask1)
		}
	}
}

func giaFrontSimulateTernary(g *Gia_Man_t, obj *Gia_Obj_t, data []uint32) {
	a := Gia_XsimNotCond(Gia_ManTerSimInfoGet(data, int32(obj.iDiff0)), int32(obj.fCompl0))
	b := Gia_XsimNotCond(Gia_ManTerSimInfoGet(data, int32(obj.iDiff1)), int32(obj.fCompl1))
	var value int32
	switch {
	case Gia_ObjIsMux(g, obj) != 0:
		lit := g.PMuxes[obj.id]
		c := Gia_XsimNotCond(Gia_ManTerSimInfoGet(data, int32(g.pObjs[lit>>1].Value)), int32(lit&1))
		// Evaluate the source's three-AND MUX expansion in the existing
		// ternary algebra, retaining its conservative unknown behavior.
		u := Gia_XsimAndCond(c, 0, b, 0)
		v := Gia_XsimAndCond(c, 1, a, 0)
		value = Gia_XsimNotCond(Gia_XsimAndCond(u, 1, v, 1), 1)
	case g.nXors != 0 && Gia_ObjIsXor(obj) != 0:
		u := Gia_XsimAndCond(a, 0, b, 1)
		v := Gia_XsimAndCond(a, 1, b, 0)
		value = Gia_XsimNotCond(Gia_XsimAndCond(u, 1, v, 1), 1)
	default:
		value = Gia_XsimAndCond(a, 0, b, 0)
	}
	Gia_ManTerSimInfoSet(data, int32(obj.Value), value)
}
