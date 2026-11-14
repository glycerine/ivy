package abci

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/base/cmd"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/amap"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/cgt"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/proof/fra"
	"github.com/glycerine/abcgo/proof/int"
	"github.com/glycerine/abcgo/proof/pdr"
	"github.com/glycerine/abcgo/proof/ssw"
	"github.com/glycerine/abcgo/sat/bmc"
	"github.com/glycerine/abcgo/sat/cnf"
)

// These aliases map the C names used by abc.c to the owning Go packages.
type Abc_Frame_t = abcmain.Frame
type Abc_Cex_t = util.Abc_Cex_t
type Gia_Man_t = gia.Gia_Man_t
type Abc_Ntk_t = abc.Abc_Ntk_t
type Abc_Obj_t = abc.Abc_Obj_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Mem_t = vec.Vec_Mem_t
type Vec_Str_t = vec.Vec_Str_t
type Vec_Wrd_t = vec.Vec_Wrd_t
type Vec_Vec_t = vec.Vec_Vec_t
type abctime = int64
type Cmd_CommandId_t = cmd.Cmd_CommandId_t
type Dar_RwrPar_t = dar.Dar_RwrPar_t
type Dar_RefPar_t = dar.Dar_RefPar_t
type Cnf_Man_t = cnf.Cnf_Man_t
type Fra_Ssw_t = fra.Fra_Ssw_t
type Fra_Sec_t = fra.Fra_Sec_t
type Ssw_Pars_t = ssw.Ssw_Pars_t
type Saig_ParBmc_t = bmc.Saig_ParBmc_t
type Inter_ManParams_t = int.Inter_ManParams_t
type Pdr_Par_t = pdr.Pdr_Par_t
type Ssw_RarPars_t = ssw.Ssw_RarPars_t
type Cgt_Par_t = cgt.Cgt_Par_t
type Saig_ParBbr_t = saig.Saig_ParBbr_t
type Amap_Par_t = amap.Amap_Par_t

func Abc_CexFree(p *Abc_Cex_t)                          { util.Abc_CexFree(p) }
func Vec_PtrFreeFree(p *Vec_Ptr_t)                      { vec.Vec_PtrFreeFree(p) }
func Vec_VecFree(p *Vec_Vec_t)                          { vec.Vec_VecFree(p) }
func Abc_Print(level int32, format string, args ...any) { util.Abc_Print(level, format, args...) }
func Gia_ManCiNum(p *Gia_Man_t) int32                   { return gia.Gia_ManCiNum(p) }
func Gia_ManCoNum(p *Gia_Man_t) int32                   { return gia.Gia_ManCoNum(p) }
func Gia_ManPoNum(p *Gia_Man_t) int32                   { return gia.Gia_ManPoNum(p) }

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
