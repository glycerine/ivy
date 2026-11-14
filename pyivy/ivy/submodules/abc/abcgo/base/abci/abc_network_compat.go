package abci

import baseabc "github.com/glycerine/abcgo/base/abc"

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
