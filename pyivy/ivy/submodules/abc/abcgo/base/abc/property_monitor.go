package abc

import "github.com/glycerine/abcgo/misc/vec"

// Abc_NtkDeriveWithOnePo duplicates a logic network and adds a monitor output
// for the conjunction of the requested object values.
func Abc_NtkDeriveWithOnePo(src *Network, ids, values *vec.Vec_Int_t) *Network {
	if src == nil || Abc_NtkIsLogic(src) == 0 || ids == nil || values == nil || vec.Vec_IntSize(ids) == 0 || vec.Vec_IntSize(ids) != vec.Vec_IntSize(values) {
		panic("abc: invalid property monitor inputs")
	}
	for i, id := range vec.Vec_IntArray(ids) {
		node := src.Object(int(id))
		value := vec.Vec_IntEntry(values, i)
		if node == nil || node.Kind == ObjectPO || (value != 0 && value != 1) {
			panic("abc: invalid monitor object or value")
		}
	}
	dst := ntkFreshLike(src, Abc_NtkType_t(src.Type), Abc_NtkFunc_t(src.Function))
	dst.Name, dst.Spec = src.Name, src.Spec
	ntkClearCopies(src)
	setCopy(src.constant, dst.One())
	for _, input := range src.ObjectsOfKind(ObjectPI) {
		Abc_NtkDupObj(dst, input, 1)
	}
	output := dst.NewObject(ObjectPO)
	output.SetName("monitor")
	for _, box := range src.boxes {
		Abc_NtkDupBox(dst, box, 1)
	}
	for _, object := range src.Objects() {
		if getCopy(object).Object == nil && object.Kind != ObjectPO {
			Abc_NtkDupObj(dst, object, 0)
		}
	}
	for _, object := range src.Objects() {
		if object.Kind == ObjectPO || Abc_ObjIsBox(object) != 0 || object.Kind == ObjectBO {
			continue
		}
		for _, edge := range object.fanins {
			Abc_ObjAddFanin(getCopy(object).Object, ntkCopiedEdge(edge))
		}
	}
	var monitor *Object
	for i, id := range vec.Vec_IntArray(ids) {
		node := getCopy(src.Object(int(id))).Object
		if vec.Vec_IntEntry(values, i) == 0 {
			node = Abc_NtkCreateNodeInv(dst, node)
		}
		if monitor == nil {
			monitor = node
		} else {
			monitor = Abc_NtkCreateNodeAnd(dst, []*Object{monitor, node})
		}
	}
	Abc_ObjAddFanin(output, monitor)
	Abc_NtkOrderCisCos(dst)
	return dst
}
