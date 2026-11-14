package abc

import "fmt"

// Public inline predicates use the original integer truth convention.
func Abc_NtkIsNetlist(n *Network) int32       { return abcBool(n.Type == uint8(ABC_NTK_NETLIST)) }
func Abc_NtkIsLogic(n *Network) int32         { return abcBool(n.Type == uint8(ABC_NTK_LOGIC)) }
func Abc_NtkIsStrash(n *Network) int32        { return abcBool(n.Type == uint8(ABC_NTK_STRASH)) }
func Abc_NtkHasSop(n *Network) int32          { return abcBool(n.Function == uint8(ABC_FUNC_SOP)) }
func Abc_NtkHasBdd(n *Network) int32          { return abcBool(n.Function == uint8(ABC_FUNC_BDD)) }
func Abc_NtkHasAig(n *Network) int32          { return abcBool(n.Function == uint8(ABC_FUNC_AIG)) }
func Abc_NtkHasMapping(n *Network) int32      { return abcBool(n.Function == uint8(ABC_FUNC_MAP)) }
func Abc_NtkHasBlifMv(n *Network) int32       { return abcBool(n.Function == uint8(ABC_FUNC_BLIFMV)) }
func Abc_NtkHasBlackbox(n *Network) int32     { return abcBool(n.Function == uint8(ABC_FUNC_BLACKBOX)) }
func Abc_NtkIsSopNetlist(n *Network) int32    { return Abc_NtkIsNetlist(n) & Abc_NtkHasSop(n) }
func Abc_NtkIsBddNetlist(n *Network) int32    { return Abc_NtkIsNetlist(n) & Abc_NtkHasBdd(n) }
func Abc_NtkIsAigNetlist(n *Network) int32    { return Abc_NtkIsNetlist(n) & Abc_NtkHasAig(n) }
func Abc_NtkIsMappedNetlist(n *Network) int32 { return Abc_NtkIsNetlist(n) & Abc_NtkHasMapping(n) }
func Abc_NtkIsBlifMvNetlist(n *Network) int32 { return Abc_NtkIsNetlist(n) & Abc_NtkHasBlifMv(n) }
func Abc_NtkIsSopLogic(n *Network) int32      { return Abc_NtkIsLogic(n) & Abc_NtkHasSop(n) }
func Abc_NtkIsBddLogic(n *Network) int32      { return Abc_NtkIsLogic(n) & Abc_NtkHasBdd(n) }
func Abc_NtkIsAigLogic(n *Network) int32      { return Abc_NtkIsLogic(n) & Abc_NtkHasAig(n) }
func Abc_NtkIsMappedLogic(n *Network) int32   { return Abc_NtkIsLogic(n) & Abc_NtkHasMapping(n) }
func Abc_NtkBiNum(n *Network) int32           { return int32(len(n.ObjectsOfKind(ObjectBI))) }
func Abc_NtkBoNum(n *Network) int32           { return int32(len(n.ObjectsOfKind(ObjectBO))) }
func Abc_NtkNetNum(n *Network) int32          { return int32(len(n.ObjectsOfKind(ObjectNet))) }
func Abc_NtkLatchNum(n *Network) int32        { return int32(len(n.ObjectsOfKind(ObjectLatch))) }
func Abc_NtkWhiteboxNum(n *Network) int32     { return int32(len(n.ObjectsOfKind(ObjectWhitebox))) }
func Abc_NtkBlackboxNum(n *Network) int32     { return int32(len(n.ObjectsOfKind(ObjectBlackbox))) }

// Abc_ObjBlackboxToWhitebox mirrors ABC's type conversion and keeps Go's
// per-kind indexes synchronized with the object's new type.
func Abc_ObjBlackboxToWhitebox(o *Object) {
	if o == nil || o.Kind != ObjectBlackbox {
		panic("Abc_ObjBlackboxToWhitebox requires a blackbox object")
	}
	if err := o.network.ChangeObjectKind(o, ObjectWhitebox); err != nil {
		panic(err)
	}
}
func Abc_NtkIsComb(n *Network) int32 { return abcBool(Abc_NtkLatchNum(n) == 0) }
func Abc_NtkHasOnlyLatchBoxes(n *Network) int32 {
	return abcBool(Abc_NtkLatchNum(n) == Abc_NtkBoxNum(n))
}
func Abc_NtkCi(n *Network, index int32) *Object  { return n.cis[index] }
func Abc_NtkCo(n *Network, index int32) *Object  { return n.cos[index] }
func Abc_NtkBox(n *Network, index int32) *Object { return n.boxes[index] }
func Abc_NtkSetName(n *Network, name string)     { n.Name = name }
func Abc_NtkSetSpec(n *Network, spec string)     { n.Spec = spec }
func setLatchInit(o *Object, value Abc_InitType_t) {
	if o.Kind != ObjectLatch {
		panic("initialization requires latch")
	}
	o.Data = value
}
func latchIsInit(o *Object, value Abc_InitType_t) int32 {
	if o.Kind != ObjectLatch {
		panic("initialization requires latch")
	}
	if o.Data == nil {
		return abcBool(value == ABC_INIT_NONE)
	}
	return abcBool(latchInitValue(o) == int(value))
}
func Abc_LatchSetInitNone(o *Object)      { setLatchInit(o, ABC_INIT_NONE) }
func Abc_LatchSetInit0(o *Object)         { setLatchInit(o, ABC_INIT_ZERO) }
func Abc_LatchSetInit1(o *Object)         { setLatchInit(o, ABC_INIT_ONE) }
func Abc_LatchSetInitDc(o *Object)        { setLatchInit(o, ABC_INIT_DC) }
func Abc_LatchIsInitNone(o *Object) int32 { return latchIsInit(o, ABC_INIT_NONE) }
func Abc_LatchIsInit0(o *Object) int32    { return latchIsInit(o, ABC_INIT_ZERO) }
func Abc_LatchIsInit1(o *Object) int32    { return latchIsInit(o, ABC_INIT_ONE) }
func Abc_LatchIsInitDc(o *Object) int32   { return latchIsInit(o, ABC_INIT_DC) }

// SetCopyEdge preserves a complemented conversion binding in the explicit Go
// representation. Copy alone stores only its regular object.
func (o *Object) SetCopyEdge(edge Edge) { setCopy(o, edge) }
func (o *Object) ReadCopyEdge() Edge    { return getCopy(o) }

func Abc_NtkStep(n *Network) int32          { return n.Step }
func Abc_NtkSetStep(n *Network, step int32) { n.Step = step }
func Abc_NtkOrderCisCos(n *Network) {
	n.cis = append([]*Object(nil), n.ObjectsOfKind(ObjectPI)...)
	n.cos = append([]*Object(nil), n.ObjectsOfKind(ObjectPO)...)
	for _, box := range n.boxes {
		n.cis = append(n.cis, box.fanouts...)
		for _, edge := range box.fanins {
			n.cos = append(n.cos, edge.Object)
		}
	}
	n.kindOrder[ObjectBO] = nil
	n.kindOrder[ObjectBI] = nil
	for _, ci := range n.cis {
		if ci.Kind == ObjectBO {
			n.kindOrder[ObjectBO] = append(n.kindOrder[ObjectBO], ci)
		}
	}
	for _, co := range n.cos {
		if co.Kind == ObjectBI {
			n.kindOrder[ObjectBI] = append(n.kindOrder[ObjectBI], co)
		}
	}
}

// SetBoxOrder validates the complete box permutation and updates the associated
// terminal and per-kind iteration orders together.
func (n *Network) SetBoxOrder(ids []int) error {
	if len(ids) != len(n.boxes) {
		return fmt.Errorf("box order must contain every box")
	}
	seen := make(map[*Object]bool)
	boxes := make([]*Object, len(ids))
	for i, id := range ids {
		o := n.Object(id)
		if o == nil || Abc_ObjIsBox(o) == 0 || seen[o] {
			return fmt.Errorf("invalid box order")
		}
		seen[o] = true
		boxes[i] = o
	}
	n.boxes = boxes
	for _, kind := range []ObjectKind{ObjectLatch, ObjectWhitebox, ObjectBlackbox} {
		n.kindOrder[kind] = nil
		for _, o := range boxes {
			if o.Kind == kind {
				n.kindOrder[kind] = append(n.kindOrder[kind], o)
			}
		}
	}
	Abc_NtkOrderCisCos(n)
	return nil
}

// SAT models are snapshots in ordered CI space; values are solver-provided.
func Abc_NtkReadModel(n *Network) []int32       { return append([]int32(nil), n.Model...) }
func Abc_NtkSetModel(n *Network, model []int32) { n.Model = append([]int32(nil), model...) }
