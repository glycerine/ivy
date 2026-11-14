package abci

import (
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/amap"
)

type Amap_Out_t = amap.Amap_Out_t

// These compatibility functions preserve ABC's C-style names at the command
// layer while delegating network and object semantics to base/abc.
func Abc_NtkIsStrash(n *Abc_Ntk_t) int32                   { return baseabc.Abc_NtkIsStrash(n) }
func Abc_NtkIsLogic(n *Abc_Ntk_t) int32                    { return baseabc.Abc_NtkIsLogic(n) }
func Abc_NtkIsNetlist(n *Abc_Ntk_t) int32                  { return baseabc.Abc_NtkIsNetlist(n) }
func Abc_NtkIsComb(n *Abc_Ntk_t) int32                     { return baseabc.Abc_NtkIsComb(n) }
func Abc_NtkHasSop(n *Abc_Ntk_t) int32                     { return baseabc.Abc_NtkHasSop(n) }
func Abc_NtkHasBdd(n *Abc_Ntk_t) int32                     { return baseabc.Abc_NtkHasBdd(n) }
func Abc_NtkHasAig(n *Abc_Ntk_t) int32                     { return baseabc.Abc_NtkHasAig(n) }
func Abc_NtkHasMapping(n *Abc_Ntk_t) int32                 { return baseabc.Abc_NtkHasMapping(n) }
func Abc_NtkIsSopLogic(n *Abc_Ntk_t) int32                 { return baseabc.Abc_NtkIsSopLogic(n) }
func Abc_NtkIsBddLogic(n *Abc_Ntk_t) int32                 { return baseabc.Abc_NtkIsBddLogic(n) }
func Abc_NtkIsMappedLogic(n *Abc_Ntk_t) int32              { return baseabc.Abc_NtkIsMappedLogic(n) }
func Abc_NtkNodeNum(n *Abc_Ntk_t) int32                    { return baseabc.Abc_NtkNodeNum(n) }
func Abc_NtkObjNum(n *Abc_Ntk_t) int32                     { return baseabc.Abc_NtkObjNum(n) }
func Abc_NtkObjNumMax(n *Abc_Ntk_t) int32                  { return baseabc.Abc_NtkObjNumMax(n) }
func Abc_NtkObj(n *Abc_Ntk_t, id int32) *Abc_Obj_t         { return baseabc.Abc_NtkObj(n, id) }
func Abc_NtkFindNode(n *Abc_Ntk_t, name string) *Abc_Obj_t { return baseabc.Abc_NtkFindNode(n, name) }
func Abc_NtkDup(n *Abc_Ntk_t) *Abc_Ntk_t                   { return baseabc.Abc_NtkDup(n) }
func Abc_NtkCrossCut(n *Abc_Ntk_t) int32                   { return baseabc.Abc_NtkCrossCut(n) }
func Abc_NtkGetCubePairNum(n *Abc_Ntk_t) int32             { return baseabc.Abc_NtkGetCubePairNum(n) }
func Abc_NtkToBdd(n *Abc_Ntk_t) int32                      { return baseabc.Abc_NtkToBdd(n) }
func Abc_NtkToAig(n *Abc_Ntk_t) int32                      { return baseabc.Abc_NtkToAig(n) }
func Abc_NtkReadExdc(n *Abc_Ntk_t) *Abc_Ntk_t              { return baseabc.Abc_NtkReadExdc(n) }
func Abc_NtkSetExdc(n, exdc *Abc_Ntk_t)                    { baseabc.Abc_NtkSetExdc(n, exdc) }
func Abc_ObjName(o *Abc_Obj_t) string                      { return baseabc.Abc_ObjName(o) }
func Abc_ObjFanin0(o *Abc_Obj_t) *Abc_Obj_t                { return baseabc.Abc_ObjFanin0(o) }
func Abc_ObjFanin1(o *Abc_Obj_t) *Abc_Obj_t                { return baseabc.Abc_ObjFanin1(o) }
func Abc_ObjFaninNum(o *Abc_Obj_t) int32                   { return baseabc.Abc_ObjFaninNum(o) }
func Abc_ObjFanout0(o *Abc_Obj_t) *Abc_Obj_t               { return baseabc.Abc_ObjFanout0(o) }
func Abc_ObjIsNode(o *Abc_Obj_t) int32                     { return baseabc.Abc_ObjIsNode(o) }
func Abc_ObjIsPi(o *Abc_Obj_t) int32                       { return baseabc.Abc_ObjIsPi(o) }
func Abc_ObjIsPo(o *Abc_Obj_t) int32                       { return baseabc.Abc_ObjIsPo(o) }
func Abc_ObjChild0(o *Abc_Obj_t) baseabc.Edge              { return baseabc.Abc_ObjChild0(o) }
func Abc_ObjChild1(o *Abc_Obj_t) baseabc.Edge              { return baseabc.Abc_ObjChild1(o) }
func Abc_ObjAddFanin(o *Abc_Obj_t, fanin any)              { baseabc.Abc_ObjAddFanin(o, fanin) }
func Abc_ObjData(o *Abc_Obj_t) any                         { return baseabc.Abc_ObjData(o) }

func Abc_NtkName(n *Abc_Ntk_t) string                    { return baseabc.Abc_NtkName(n) }
func Abc_NtkSetName(n *Abc_Ntk_t, name string)           { baseabc.Abc_NtkSetName(n, name) }
func Abc_NtkReadModel(n *Abc_Ntk_t) []int32              { return baseabc.Abc_NtkReadModel(n) }
func Abc_NtkSetModel(n *Abc_Ntk_t, model []int32)        { baseabc.Abc_NtkSetModel(n, model) }
func Abc_NtkReadSeqModel(n *Abc_Ntk_t) *Abc_Cex_t        { return baseabc.Abc_NtkReadSeqModel(n) }
func Abc_NtkSetSeqModel(n *Abc_Ntk_t, model *Abc_Cex_t)  { baseabc.Abc_NtkSetSeqModel(n, model) }
func Abc_ObjReadCopy(o *Abc_Obj_t) *Abc_Obj_t            { return baseabc.Abc_ObjReadCopy(o) }
func Abc_ObjSetCopy(o, copy *Abc_Obj_t)                  { baseabc.Abc_ObjSetCopy(o, copy) }
func Abc_ObjReadCopyEdge(o *Abc_Obj_t) baseabc.Edge      { return o.ReadCopyEdge() }
func Abc_ObjSetCopyEdge(o *Abc_Obj_t, edge baseabc.Edge) { o.SetCopyEdge(edge) }

func abcCompatCopyEdge(value any) baseabc.Edge {
	switch v := value.(type) {
	case baseabc.Edge:
		return v
	case *Abc_Obj_t:
		return v.Signal()
	default:
		panic("expected ABC object or edge")
	}
}
