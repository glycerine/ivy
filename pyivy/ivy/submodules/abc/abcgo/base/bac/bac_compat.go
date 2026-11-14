package bac

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/base/cmd"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/proof/cec"
)

// These aliases map the C types included by bac.h to their source packages.
type Abc_Nam_t = util.Abc_Nam_t
type Gia_Man_t = gia.Gia_Man_t
type Gia_Obj_t = gia.Gia_Obj_t
type Abc_Obj_t = abc.Abc_Obj_t
type Abc_Ntk_t = abc.Abc_Ntk_t
type Mio_Library_t = mio.Mio_Library_t
type Mio_Gate_t = mio.Mio_Gate_t
type Mio_Pin_t = mio.Mio_Pin_t
type Abc_Frame_t = abcmain.Abc_Frame_t
type Cmd_CommandFuncType = cmd.Cmd_CommandFuncType
type Cec_ParCec_t = cec.Cec_ParCec_t

func Cec_ManCecSetDefaultParams(p *Cec_ParCec_t)        { cec.Cec_ManCecSetDefaultParams(p) }
func Cec_ParCecSetVerbose(p *Cec_ParCec_t, v int32)     { cec.Cec_ParCecSetVerbose(p, v) }
func Cec_ParCecVerbose(p *Cec_ParCec_t) int32           { return cec.Cec_ParCecVerbose(p) }
func Cec_ManVerify(g *Gia_Man_t, p *Cec_ParCec_t) int32 { return cec.Cec_ManVerify(g, p) }
func Abc_FrameReadLibGen() *Mio_Library_t {
	lib, _ := abcmain.Abc_FrameReadLibGen().(*mio.Mio_Library_t)
	return lib
}
func Mio_LibraryReadGateByName(p *Mio_Library_t, name, out string) *Mio_Gate_t {
	return mio.Mio_LibraryReadGateByName(p, name, out)
}
func Mio_LibraryReadGateByTruth(p *Mio_Library_t, truth uint64) *Mio_Gate_t {
	return mio.Mio_LibraryReadGateByTruth(p, truth)
}
func Mio_GateReadName(p *Mio_Gate_t) string             { return mio.Mio_GateReadName(p) }
func Mio_GateReadPins(p *Mio_Gate_t) *Mio_Pin_t         { return mio.Mio_GateReadPins(p) }
func Mio_GateReadOutName(p *Mio_Gate_t) string          { return mio.Mio_GateReadOutName(p) }
func Mio_GateReadPinName(p *Mio_Gate_t, i int32) string { return mio.Mio_GateReadPinName(p, i) }
func Mio_PinReadNext(p *Mio_Pin_t) *Mio_Pin_t           { return mio.Mio_PinReadNext(p) }
func Mio_PinReadName(p *Mio_Pin_t) string               { return mio.Mio_PinReadName(p) }

func Abc_FrameGetGlobalFrame() *Abc_Frame_t    { return abcmain.Abc_FrameGetGlobalFrame() }
func Abc_FrameReadGia(f *Abc_Frame_t) any      { return abcmain.Abc_FrameReadGia(f) }
func Abc_FrameUpdateGia(f *Abc_Frame_t, g any) { abcmain.Abc_FrameUpdateGia(f, g) }
