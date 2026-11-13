package abc

// Abc_NtkSetCiOrderSource implements the vCis/vPis update in
// src/base/abci/abcOrder.c. The source checks only the length; it permits
// duplicate entries and does not reorder box terminal vectors.
func Abc_NtkSetCiOrderSource(n *Network, order []*Object, reverse int32) {
	if len(order) != len(n.cis) {
		panic("abc: CI order length mismatch")
	}
	for i, o := range order {
		j := i
		if reverse != 0 {
			j = len(order) - 1 - i
		}
		n.cis[j] = o
	}
	n.kindOrder[ObjectPI] = nil
	for _, o := range n.cis {
		if Abc_ObjIsPi(o) != 0 {
			n.kindOrder[ObjectPI] = append(n.kindOrder[ObjectPI], o)
		}
	}
}
