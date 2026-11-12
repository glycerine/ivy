package if_

import "fmt"

// ifDanglingChoices walks both structural fanins and choice alternatives from
// the outputs, returning the IDs of AND nodes that are not reachable.
func ifDanglingChoices(p *If_Man_t) []int32 {
	for i := int32(0); i < If_ManObjNum(p); i++ {
		If_ManObj(p, i).fDriver = 0
	}
	var visit func(*If_Obj_t)
	visit = func(obj *If_Obj_t) {
		if obj == nil || obj.Type != IF_AND || obj.fDriver != 0 {
			return
		}
		obj.fDriver = 1
		visit(obj.pFanin0)
		visit(obj.pFanin1)
		visit(obj.pEquiv)
	}
	for i := int32(0); i < If_ManCoNum(p); i++ {
		visit(If_ManCo(p, i).pFanin0)
	}
	var dangling []int32
	for i := int32(0); i < If_ManObjNum(p); i++ {
		obj := If_ManObj(p, i)
		if obj.Type == IF_AND && obj.fDriver == 0 {
			dangling = append(dangling, i)
		}
		obj.fDriver = 0
	}
	return dangling
}

// If_ManCheckChoices reports nodes that do not drive an output, including
// nodes reachable only through a choice alternative.
func If_ManCheckChoices(p *If_Man_t) {
	dangling := ifDanglingChoices(p)
	if len(dangling) == 0 {
		fmt.Println("There are no dangling objects.")
		return
	}
	for _, id := range dangling {
		fmt.Printf("Object %d is dangling.\n", id)
	}
}
