package abc

import "unsafe"

// Abc_NtkMemoryVectorBytes reports the source-style allocated backing storage
// for the network's manager vectors and object connectivity vectors. It uses
// capacities, as Vec_PtrMemory/Vec_IntMemory do, rather than live lengths.
// The connectivity term uses ABC's int-sized fanin/fanout entries.
func Abc_NtkMemoryVectorBytes(n *Network) float64 {
	if n == nil {
		return 0
	}
	pointerBytes := unsafe.Sizeof((*Object)(nil))
	entries := cap(n.kindOrder[ObjectPI]) + cap(n.kindOrder[ObjectPO]) +
		cap(n.cis) + cap(n.cos) + cap(n.objects)
	bytes := uintptr(entries) * pointerBytes
	intBytes := unsafe.Sizeof(int32(0))
	for _, obj := range n.objects {
		if obj == nil {
			continue
		}
		bytes += uintptr(cap(obj.fanins)+cap(obj.fanouts)) * intBytes
	}
	return float64(bytes)
}
