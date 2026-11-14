// ABC's public network and object API, backed by Go-owned graph storage.
package abc

type Abc_Ntk_t = Network
type Abc_Obj_t = Object
type Abc_ObjType_t = ObjectKind
type Abc_NtkType_t uint8
type Abc_NtkFunc_t uint8
type Abc_InitType_t uint8

const (
	ABC_NTK_NONE Abc_NtkType_t = iota
	ABC_NTK_NETLIST
	ABC_NTK_LOGIC
	ABC_NTK_STRASH
	ABC_NTK_OTHER
)
const (
	ABC_FUNC_NONE Abc_NtkFunc_t = iota
	ABC_FUNC_SOP
	ABC_FUNC_BDD
	ABC_FUNC_AIG
	ABC_FUNC_MAP
	ABC_FUNC_BLIFMV
	ABC_FUNC_BLACKBOX
	ABC_FUNC_OTHER
)
const (
	ABC_OBJ_NONE     = ObjectNone
	ABC_OBJ_CONST1   = ObjectConst1
	ABC_OBJ_PI       = ObjectPI
	ABC_OBJ_PO       = ObjectPO
	ABC_OBJ_BI       = ObjectBI
	ABC_OBJ_BO       = ObjectBO
	ABC_OBJ_NET      = ObjectNet
	ABC_OBJ_NODE     = ObjectNode
	ABC_OBJ_LATCH    = ObjectLatch
	ABC_OBJ_WHITEBOX = ObjectWhitebox
	ABC_OBJ_BLACKBOX = ObjectBlackbox
	ABC_OBJ_NUMBER   = 11
)
const (
	ABC_INIT_NONE Abc_InitType_t = iota
	ABC_INIT_ZERO
	ABC_INIT_ONE
	ABC_INIT_DC
	ABC_INIT_OTHER
)

type Abc_Time_t struct{ Rise, Fall float32 }

func abcBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func objectEdge(value any) Edge {
	switch v := value.(type) {
	case Edge:
		return v
	case *Object:
		return v.Signal()
	default:
		panic("expected ABC object or edge")
	}
}
func Abc_ObjRegular(value any) *Object { return objectEdge(value).Object }
func Abc_ObjNot(value any) Edge        { return objectEdge(value).Not() }
func Abc_ObjNotCond(value any, condition int32) Edge {
	e := objectEdge(value)
	if condition != 0 {
		e = e.Not()
	}
	return e
}
func Abc_ObjIsComplement(value any) int32   { return abcBool(objectEdge(value).Complement) }
func Abc_ObjId(o *Object) int32             { return int32(o.ID) }
func Abc_ObjNtk(o *Object) *Network         { return o.network }
func Abc_ObjType(o *Object) ObjectKind      { return o.Kind }
func Abc_ObjIsPi(o *Object) int32           { return abcBool(o.Kind == ObjectPI) }
func Abc_ObjIsPo(o *Object) int32           { return abcBool(o.Kind == ObjectPO) }
func Abc_ObjIsBi(o *Object) int32           { return abcBool(o.Kind == ObjectBI) }
func Abc_ObjIsBo(o *Object) int32           { return abcBool(o.Kind == ObjectBO) }
func Abc_ObjIsCi(o *Object) int32           { return abcBool(o.IsCI()) }
func Abc_ObjIsCo(o *Object) int32           { return abcBool(o.IsCO()) }
func Abc_ObjIsNet(o *Object) int32          { return abcBool(o.Kind == ObjectNet) }
func Abc_ObjIsNode(o *Object) int32         { return abcBool(o.Kind == ObjectNode) }
func Abc_ObjIsBarBuf(o *Object) int32       { return abcBool(o.Barrier) }
func Abc_ObjIsLatch(o *Object) int32        { return abcBool(o.Kind == ObjectLatch) }
func Abc_ObjIsBox(o *Object) int32          { return abcBool(o.Kind >= ObjectLatch && o.Kind <= ObjectBlackbox) }
func Abc_ObjIsConst1(o *Object) int32       { return abcBool(o.Kind == ObjectConst1) }
func Abc_ObjFaninNum(o *Object) int32       { return int32(len(o.fanins)) }
func Abc_ObjTemp(o *Object) int32           { return int32(o.Temp) }
func Abc_ObjSetTemp(o *Object, value int32) { o.Temp = int(value) }
func Abc_ObjFanoutNum(o *Object) int32 {
	if o.network.refs != nil {
		return int32(o.network.refs[o])
	}
	return int32(len(o.fanouts))
}
func Abc_ObjFanin(o *Object, i int32) *Object  { return o.fanins[i].Object }
func Abc_ObjFanin0(o *Object) *Object          { return Abc_ObjFanin(o, 0) }
func Abc_ObjFanin1(o *Object) *Object          { return Abc_ObjFanin(o, 1) }
func Abc_ObjFanout(o *Object, i int32) *Object { return o.fanouts[i] }
func Abc_ObjFanout0(o *Object) *Object         { return Abc_ObjFanout(o, 0) }
func Abc_ObjChild(o *Object, i int32) Edge     { return o.fanins[i] }
func Abc_ObjChild0(o *Object) Edge             { return Abc_ObjChild(o, 0) }
func Abc_ObjChild1(o *Object) Edge             { return Abc_ObjChild(o, 1) }
func Abc_ObjFaninC(o *Object, i int32) int32   { return abcBool(o.fanins[i].Complement) }
func Abc_ObjSetFaninC(o *Object, i int32)      { o.fanins[i].Complement = true; clear(o.network.hashed) }
func Abc_ObjXorFaninC(o *Object, i int32) {
	o.fanins[i].Complement = !o.fanins[i].Complement
	clear(o.network.hashed)
}
func Abc_ObjLevel(o *Object) int32            { return int32(o.Level) }
func Abc_ObjSetLevel(o *Object, level int32)  { o.Level = int(level) }
func Abc_ObjData(o *Object) any               { return o.Data }
func Abc_ObjSetData(o *Object, data any)      { o.Data = data; clear(o.network.hashed) }
func Abc_ObjCopyData(o *Object) any           { return o.CopyData }
func Abc_ObjSetCopyData(o *Object, data any)  { o.CopyData = data }
func Abc_ObjCopyEdge(o *Object) Edge          { return getCopy(o) }
func Abc_NtkObj(n *Network, id int32) *Object { return n.Object(int(id)) }
func Abc_NtkObjNum(n *Network) int32          { return int32(len(n.Objects())) }
func Abc_NtkObjNumMax(n *Network) int32       { return int32(len(n.objects)) }
func Abc_NtkPiNum(n *Network) int32           { return int32(len(n.ObjectsOfKind(ObjectPI))) }
func Abc_NtkPoNum(n *Network) int32           { return int32(len(n.ObjectsOfKind(ObjectPO))) }
func Abc_NtkNodeNum(n *Network) int32         { return int32(len(n.ObjectsOfKind(ObjectNode))) }
func Abc_NtkName(n *Network) string           { return n.Name }

func Abc_NtkCreatePi(n *Network) *Object       { return n.NewObject(ObjectPI) }
func Abc_NtkCreatePo(n *Network) *Object       { return n.NewObject(ObjectPO) }
func Abc_NtkCreateBi(n *Network) *Object       { return n.NewObject(ObjectBI) }
func Abc_NtkCreateBo(n *Network) *Object       { return n.NewObject(ObjectBO) }
func Abc_NtkCreateNet(n *Network) *Object      { return n.NewObject(ObjectNet) }
func Abc_NtkCreateNode(n *Network) *Object     { return n.NewObject(ObjectNode) }
func Abc_NtkCreateLatch(n *Network) *Object    { return n.NewObject(ObjectLatch) }
func Abc_NtkCreateBlackbox(n *Network) *Object { return n.NewObject(ObjectBlackbox) }
func Abc_NtkCreateWhitebox(n *Network) *Object { return n.NewObject(ObjectWhitebox) }
func Abc_NtkCiNum(n *Network) int32 {
	count := int32(0)
	for _, o := range n.Objects() {
		if o.IsCI() {
			count++
		}
	}
	return count
}
func Abc_NtkCoNum(n *Network) int32 {
	count := int32(0)
	for _, o := range n.Objects() {
		if o.IsCO() {
			count++
		}
	}
	return count
}
func Abc_NtkBoxNum(n *Network) int32 {
	count := int32(0)
	for _, o := range n.Objects() {
		if Abc_ObjIsBox(o) != 0 {
			count++
		}
	}
	return count
}
func Abc_NtkPi(n *Network, i int32) *Object { return n.ObjectsOfKind(ObjectPI)[i] }
func Abc_NtkPo(n *Network, i int32) *Object { return n.ObjectsOfKind(ObjectPO)[i] }
func Abc_NtkBarBufNum(n *Network) int32     { return int32(n.Barriers) }
func Abc_NtkNodes(n *Network) []*Object     { return n.ObjectsOfKind(ObjectNode) }

func Abc_AigNodeIsAnd(o *Object) int32 {
	return abcBool(o != nil && o.Kind == ObjectNode && o.Data == aigANDMarker && len(o.fanins) == 2)
}
func Abc_AigNodeIsConst(o *Object) int32 { return abcBool(o.Kind == ObjectConst1) }
func Abc_AigNodeIsChoice(o *Object) int32 {
	return abcBool(o.Kind == ObjectNode && o.Choice != nil && len(o.fanouts) > 0)
}
func Abc_ObjChoice(o *Object) *Object { return o.Choice }
func Abc_ObjFaninC0(o *Object) int32  { return Abc_ObjFaninC(o, 0) }
func Abc_ObjFaninC1(o *Object) int32  { return Abc_ObjFaninC(o, 1) }
