package abc

import "github.com/glycerine/abcgo/misc/vec"

// Abc_NtkDfsCheck avoids paths through pivot while respecting traversal marks.
func Abc_NtkDfsCheck(obj, pivot *Object) bool {
	if obj == pivot {
		return false
	}
	if Abc_NodeIsTravIdCurrent(obj) != 0 {
		return true
	}
	Abc_NodeSetTravIdCurrent(obj)
	if obj.IsCI() {
		return true
	}
	if obj.Kind != ObjectNode {
		panic("abc: expected logic node")
	}
	for _, edge := range obj.fanins {
		if !Abc_NtkDfsCheck(edge.Object, pivot) {
			return false
		}
	}
	return true
}

// Abc_NtkDfsReverseOne visits a bounded fanout cone in reverse DFS order.
func Abc_NtkDfsReverseOne(obj *Object, tfo *vec.Vec_Int_t, maxLevel, maxFanouts int) {
	if Abc_NodeIsTravIdCurrent(obj) != 0 {
		return
	}
	Abc_NodeSetTravIdCurrent(obj)
	if obj.IsCO() || obj.Level > maxLevel {
		return
	}
	if obj.Kind != ObjectNode {
		panic("abc: expected logic node")
	}
	expand := len(obj.fanouts) <= maxFanouts
	if expand {
		for _, fan := range obj.fanouts {
			if fan.IsCO() || fan.Level > maxLevel {
				expand = false
				break
			}
		}
	}
	if expand {
		for _, fan := range obj.fanouts {
			Abc_NtkDfsReverseOne(fan, tfo, maxLevel, maxFanouts)
		}
	}
	vec.Vec_IntPush(tfo, int32(obj.ID))
	obj.Temp = 0
}

// Abc_NtkDfsOne labels and orders a bounded fanin cone; visit runs postorder.
func Abc_NtkDfsOne(obj *Object, tfi *vec.Vec_Int_t, minLevel, ciLabel int, visit func(*Object)) int {
	if Abc_NodeIsTravIdCurrent(obj) != 0 {
		return obj.Temp
	}
	Abc_NodeSetTravIdCurrent(obj)
	if obj.IsCI() || obj.Level < minLevel && len(obj.fanins) > 0 {
		vec.Vec_IntPush(tfi, int32(obj.ID))
		obj.Temp = ciLabel
		return obj.Temp
	}
	if obj.Kind != ObjectNode {
		panic("abc: expected logic node")
	}
	obj.Temp = 0
	if len(obj.fanins) == 0 {
		obj.Temp = ciLabel
	}
	for _, fan := range obj.fanins {
		obj.Temp |= Abc_NtkDfsOne(fan.Object, tfi, minLevel, ciLabel, visit)
	}
	vec.Vec_IntPush(tfi, int32(obj.ID))
	if visit != nil {
		visit(obj)
	}
	return obj.Temp
}
