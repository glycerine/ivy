// Structural MUX recognition ported from ABC ivyUtil.c.
package ivy

func Ivy_ObjIsMuxType(n *Ivy_Obj_t) int32 {
	if n == nil || Ivy_IsComplement(n) != 0 || Ivy_ObjIsAnd(n) == 0 || Ivy_ObjFaninC0(n) == 0 || Ivy_ObjFaninC1(n) == 0 {
		return 0
	}
	a, b := Ivy_ObjFanin0(n), Ivy_ObjFanin1(n)
	if Ivy_ObjIsAnd(a) == 0 || Ivy_ObjIsAnd(b) == 0 {
		return 0
	}
	for _, x := range []*Ivy_Obj_t{a.pFanin0, a.pFanin1} {
		for _, y := range []*Ivy_Obj_t{b.pFanin0, b.pFanin1} {
			if x == Ivy_Not(y) {
				return 1
			}
		}
	}
	return 0
}
func Ivy_ObjRecognizeMux(n *Ivy_Obj_t, then, otherwise **Ivy_Obj_t) *Ivy_Obj_t {
	if Ivy_ObjIsMuxType(n) == 0 {
		panic("ivy: node is not a structural MUX")
	}
	a, b := Ivy_ObjFanin0(n), Ivy_ObjFanin1(n)
	left, right := [2]*Ivy_Obj_t{a.pFanin0, a.pFanin1}, [2]*Ivy_Obj_t{b.pFanin0, b.pFanin1}
	for i, x := range left {
		for j, y := range right {
			if x != Ivy_Not(y) {
				continue
			}
			if Ivy_IsComplement(x) != 0 {
				*then, *otherwise = Ivy_Not(right[1-j]), Ivy_Not(left[1-i])
				return y
			}
			*then, *otherwise = Ivy_Not(left[1-i]), Ivy_Not(right[1-j])
			return x
		}
	}
	panic("ivy: structural MUX lost its control")
}
