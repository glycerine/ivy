package abc

import (
	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/vec"
)

// abcSerializeMapped emits ABC's operation records in PI, node, PO order.
func abcSerializeMapped(n *Network) *vec.Vec_Wec_t {
	pis, nodes, pos := n.ObjectsOfKind(ObjectPI), n.ObjectsOfKind(ObjectNode), n.ObjectsOfKind(ObjectPO)
	result := vec.Vec_WecStart(len(pis) + len(nodes) + len(pos))
	index := 0
	for _, obj := range pis {
		obj.Temp = index
		index++
	}
	for _, obj := range nodes {
		row := vec.Vec_WecEntry(result, index)
		vec.Vec_IntPush(row, Abc_GateToType(obj))
		vec.Vec_IntPush(row, int32(index))
		for _, fan := range obj.fanins {
			vec.Vec_IntPush(row, int32(fan.Object.Temp))
		}
		obj.Temp = index
		index++
	}
	for _, obj := range pos {
		if len(obj.fanins) != 1 {
			panic("abc: invalid synthesized output")
		}
		vec.Vec_IntPushTwo(vec.Vec_WecEntry(result, index), miniaig.ABC_OPER_BIT_BUF, int32(obj.fanins[0].Object.Temp))
		index++
	}
	return result
}
