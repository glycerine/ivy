package abci

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// These aliases map the C names used by abc.c to the owning Go packages.
type Abc_Frame_t = abcmain.Frame
type Abc_Cex_t = util.Abc_Cex_t
type Gia_Man_t = gia.Gia_Man_t
type Abc_Ntk_t = abc.Abc_Ntk_t
type Abc_Obj_t = abc.Abc_Obj_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Vec_t = vec.Vec_Vec_t
type If_Man_t = ifpkg.If_Man_t
type If_Obj_t = ifpkg.If_Obj_t
type If_Par_t = ifpkg.If_Par_t
type If_Cut_t = ifpkg.If_Cut_t
type Ifif_Par_t = ifpkg.Ifif_Par_t
type Hop_Man_t = hop.Hop_Man_t
type Hop_Obj_t = hop.Hop_Obj_t

func abcFrameState(frame *Abc_Frame_t) *abcmain.ABCFrameState {
	return abcmain.ReadABCFrameState(frame)
}

func Abc_FrameReadNtk(frame *Abc_Frame_t) *Abc_Ntk_t {
	network, _ := abcmain.Abc_FrameReadNtk(frame).(*Abc_Ntk_t)
	return network
}

func abcYesNo(value int32) string {
	return abcSelect(value, "yes", "no")
}

func abcSelect(value int32, yes, no string) string {
	if value != 0 {
		return yes
	}
	return no
}
