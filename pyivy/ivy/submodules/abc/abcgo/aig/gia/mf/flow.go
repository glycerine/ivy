package mf

// discountFlowReferences ports the uncoarsened-AIG adjustment in
// Mf_ManSetFlowRefs. Coarsened graphs have real MUX storage and skip this step.
func discountFlowReferences(network Network, refs []int32) {
	if storage, ok := network.(interface{ MuxesEnabled() bool }); ok && storage.MuxesEnabled() {
		return
	}
	for id := 0; id < network.ObjectCount(); id++ {
		a, b, _, kind := network.Gate(int32(id))
		if kind != 2 || a>>1 == b>>1 || a&1 == 0 || b&1 == 0 {
			continue
		}
		x0, x1, _, xKind := network.Gate(a >> 1)
		y0, y1, _, yKind := network.Gate(b >> 1)
		if !gateNode(xKind) || !gateNode(yKind) {
			continue
		}
		if xKind == 5 {
			x1 = x0
		}
		if yKind == 5 {
			y1 = y0
		}
		left, right := [2]int32{x0, x1}, [2]int32{y0, y1}
		// Match Gia_ObjRecognizeMux's precedence, including XOR structures.
		for _, pair := range [][2]int{{1, 1}, {0, 0}, {0, 1}, {1, 0}} {
			i, j := pair[0], pair[1]
			if left[i]>>1 != right[j]>>1 || (left[i]^right[j])&1 == 0 {
				continue
			}
			control, data0, data1 := left[i]>>1, left[1-i]>>1, right[1-j]>>1
			_, _, _, ctrlKind := network.Gate(control)
			if gateNode(ctrlKind) {
				refs[control]--
			}
			if data0 == data1 {
				_, _, _, dataKind := network.Gate(data0)
				if gateNode(dataKind) {
					refs[data0]--
				}
			}
			break
		}
	}
}
