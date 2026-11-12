package lpk

import (
	"time"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/kit"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/vec"
)

type Abc_Ntk_t = abc.Abc_Ntk_t
type Abc_Obj_t = abc.Abc_Obj_t
type Hop_Man_t = hop.Hop_Man_t
type Hop_Obj_t = hop.Hop_Obj_t
type If_Man_t = ifpkg.If_Man_t
type If_Obj_t = ifpkg.If_Obj_t
type If_Par_t = ifpkg.If_Par_t
type If_LibLut_t = ifpkg.If_LibLut_t
type Kit_DsdMan_t = kit.Kit_DsdMan_t
type Kit_DsdNtk_t = kit.Kit_DsdNtk_t
type Kit_DsdObj_t = kit.Kit_DsdObj_t
type Kit_Graph_t = kit.Kit_Graph_t
type Kit_Node_t = kit.Kit_Node_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Vec_t = vec.Vec_Vec_t
type Mem_Flex_t = mem.Mem_Flex_t
type abctime = time.Duration

type word = uint64
type CloudManager = kit.CloudManager
type CloudNode = kit.CloudNode
