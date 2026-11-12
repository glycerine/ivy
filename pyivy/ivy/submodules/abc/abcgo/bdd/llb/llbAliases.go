package llb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge
type Gia_Man_t = gia.Gia_Man_t
type DdManager = core.DdManager
type DdNode = core.DdNode
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Abc_Cex_t = util.Abc_Cex_t
type abctime = int64

func Aig_ObjId(obj *Aig_Obj_t) int32           { return aig.Aig_ObjId(*obj) }
func Vec_IntEntry(v *Vec_Int_t, i int32) int32 { return vec.Vec_IntEntry(v, int(i)) }
func Vec_IntSize(v *Vec_Int_t) int32           { return int32(vec.Vec_IntSize(v)) }
func Vec_PtrEntry(v *Vec_Ptr_t, i int32) any   { return vec.Vec_PtrEntry(v, int(i)) }
func Vec_PtrSize(v *Vec_Ptr_t) int32           { return int32(vec.Vec_PtrSize(v)) }
func Saig_ManPiNum(p *Aig_Man_t) int32         { return int32(p.InputCount()) - p.RegisterCount() }
func Saig_ManRegNum(p *Aig_Man_t) int32        { return p.RegisterCount() }
