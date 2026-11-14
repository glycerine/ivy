package abci

import (
	aigpkg "github.com/glycerine/abcgo/aig/aig"
	giapkg "github.com/glycerine/abcgo/aig/gia"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/base/cmd"
	abcmain "github.com/glycerine/abcgo/base/main"
	extrapkg "github.com/glycerine/abcgo/misc/extra"
	darpkg "github.com/glycerine/abcgo/opt/dar"
	livepkg "github.com/glycerine/abcgo/proof/live"
	cnfpkg "github.com/glycerine/abcgo/sat/cnf"
)

const (
	CMD_COMMAND_STRASH      = cmd.CMD_COMMAND_STRASH
	CMD_COMMAND_PRINT_STATS = cmd.CMD_COMMAND_PRINT_STATS
	CMD_COMMAND_PDR         = cmd.CMD_COMMAND_PDR
)

func Dar_LibStart() { darpkg.Dar_LibStart() }
func Dar_LibStop()  { darpkg.Dar_LibStop() }
func Cnf_ManFree()  { cnfpkg.Cnf_ManFree() }
func Aig_RManQuit() { _ = aigpkg.Aig_RManQuit() }
func Sdm_ManQuit()  { extrapkg.Sdm_ManQuit() }
func Abc_FrameReplaceCurrentNetwork(frame *Abc_Frame_t, network *Abc_Ntk_t) {
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, network)
}

func abcRunLiveCommand(frame *Abc_Frame_t, argc int32, argv []string, run func(*livepkg.Frame, int32, []string) int32) int32 {
	liveFrame := &livepkg.Frame{}
	if frame != nil {
		liveFrame.In, liveFrame.Out, liveFrame.Err = frame.In, frame.Out, frame.Err
	}
	if ntk := abcFrameNtk(frame); ntk != nil {
		graph := Abc_NtkToDar(ntk, 0, 1)
		if graph != nil {
			liveFrame.Network = &livepkg.Network{
				Graph:       graph,
				InputNames:  baseabc.Abc_NtkCollectCioNames(ntk, 0),
				OutputNames: baseabc.Abc_NtkCollectCioNames(ntk, 1),
			}
		}
	}
	oldGraph := (*aigpkg.Graph)(nil)
	if liveFrame.Network != nil {
		oldGraph = liveFrame.Network.Graph
	}
	status := run(liveFrame, argc, argv)
	if liveFrame.Network != nil && liveFrame.Network.Graph != nil && liveFrame.Network.Graph != oldGraph {
		newNtk := Abc_NtkFromAigPhase(liveFrame.Network.Graph)
		if newNtk != nil {
			for i, name := range liveFrame.Network.InputNames {
				if int32(i) >= baseabc.Abc_NtkCiNum(newNtk) {
					break
				}
				baseabc.Abc_ObjAssignName(baseabc.Abc_NtkCi(newNtk, int32(i)), name, nil)
			}
			for i, name := range liveFrame.Network.OutputNames {
				if int32(i) >= baseabc.Abc_NtkCoNum(newNtk) {
					break
				}
				baseabc.Abc_ObjAssignName(baseabc.Abc_NtkCo(newNtk, int32(i)), name, nil)
			}
			Abc_FrameReplaceCurrentNetwork(frame, newNtk)
		}
	}
	if oldGraph != nil {
		aigpkg.Aig_ManStop(oldGraph)
	}
	if liveFrame.Network != nil && liveFrame.Network.Graph != nil && liveFrame.Network.Graph != oldGraph {
		aigpkg.Aig_ManStop(liveFrame.Network.Graph)
	}
	return status
}

func Abc_CommandAbcLivenessToSafety(frame *Abc_Frame_t, argc int32, argv []string) int32 {
	return abcRunLiveCommand(frame, argc, argv, livepkg.Abc_CommandAbcLivenessToSafety)
}
func Abc_CommandAbcLivenessToSafetySim(frame *Abc_Frame_t, argc int32, argv []string) int32 {
	return abcRunLiveCommand(frame, argc, argv, livepkg.Abc_CommandAbcLivenessToSafetySim)
}
func Abc_CommandAbcLivenessToSafetyWithLTL(frame *Abc_Frame_t, argc int32, argv []string) int32 {
	return abcRunLiveCommand(frame, argc, argv, livepkg.Abc_CommandAbcLivenessToSafetyWithLTL)
}
func Abc_CommandCS_kLiveness(frame *Abc_Frame_t, argc int32, argv []string) int32 {
	return abcRunLiveCommand(frame, argc, argv, livepkg.Abc_CommandCS_kLiveness)
}
func Abc_CommandNChooseK(frame *Abc_Frame_t, argc int32, argv []string) int32 {
	return abcRunLiveCommand(frame, argc, argv, livepkg.Abc_CommandNChooseK)
}

func Cmd_CommandAdd(frame *Abc_Frame_t, group, name string, handler cmd.Cmd_CommandFuncType, changes int32) {
	cmd.Cmd_CommandAdd(frame, group, name, handler, changes)
}

// The C frame has fields that belong to several Go packages. Keep their
// ownership in base/main's Frame and expose typed views here, at the ABC
// command boundary.
func abcFrameResource[T any](frame *Abc_Frame_t, key string) T {
	var zero T
	if frame == nil {
		return zero
	}
	value := frame.Resource(key)
	if value == nil {
		return zero
	}
	result, ok := value.(T)
	if !ok {
		panic("abci: incompatible frame resource " + key)
	}
	return result
}

func abcFrameSetResource[T any](frame *Abc_Frame_t, key string, value T) {
	if frame == nil {
		return
	}
	frame.SetResource(key, value)
}

func abcFrameTakeResource[T any](frame *Abc_Frame_t, key string) T {
	var zero T
	if frame == nil {
		return zero
	}
	value := frame.TakeResource(key)
	if value == nil {
		return zero
	}
	result, ok := value.(T)
	if !ok {
		panic("abci: incompatible frame resource " + key)
	}
	return result
}

func abcFrameGia(frame *Abc_Frame_t) *Gia_Man_t {
	if frame == nil {
		return nil
	}
	value, _ := abcmain.Abc_FrameReadGia(frame).(*Gia_Man_t)
	return value
}
func abcFrameSetGia(frame *Abc_Frame_t, gia *Gia_Man_t) {
	if frame == nil {
		return
	}
	abcmain.Abc_FrameUpdateGia(frame, gia)
}
func abcFrameTakeGia(frame *Abc_Frame_t) *Gia_Man_t {
	if frame == nil {
		return nil
	}
	gia, _ := abcmain.Abc_FrameGetGia(frame).(*Gia_Man_t)
	return gia
}
func abcFrameStopGia(frame *Abc_Frame_t, key string) {
	if gia := abcFrameTakeResource[*Gia_Man_t](frame, key); gia != nil {
		Gia_ManStop(gia)
	}
}
func abcGiaInputNames(gia *Gia_Man_t) *Vec_Ptr_t {
	return giapkg.Gia_ManInputNames(gia)
}
func abcGiaSetInputNames(gia *Gia_Man_t, names *Vec_Ptr_t) {
	giapkg.Gia_ManSetInputNames(gia, names)
}
func abcGiaOutputNames(gia *Gia_Man_t) *Vec_Ptr_t {
	return giapkg.Gia_ManOutputNames(gia)
}
func abcGiaSetOutputNames(gia *Gia_Man_t, names *Vec_Ptr_t) {
	giapkg.Gia_ManSetOutputNames(gia, names)
}

func abcFrameNtk(frame *Abc_Frame_t) *Abc_Ntk_t {
	if frame == nil {
		return nil
	}
	value, _ := abcmain.Abc_FrameReadNtk(frame).(*Abc_Ntk_t)
	return value
}
func abcFrameSetNtk(frame *Abc_Frame_t, network *Abc_Ntk_t) {
	if frame == nil || network == nil {
		return
	}
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, network)
}

func abcFrameCex(frame *Abc_Frame_t) *Abc_Cex_t {
	if frame == nil {
		return nil
	}
	value, _ := abcFrameState(frame).Cex().(*Abc_Cex_t)
	return value
}
func abcFrameSetCex(frame *Abc_Frame_t, cex *Abc_Cex_t) {
	if frame == nil {
		return
	}
	var value any = cex
	abcFrameState(frame).ReplaceCex(&value)
}
func abcFrameStatus(frame *Abc_Frame_t) int32 {
	if frame == nil {
		return 0
	}
	status, _ := abcFrameState(frame).VerificationStatus()
	return status
}
func abcFrameSetStatus(frame *Abc_Frame_t, status int32) {
	if frame == nil {
		return
	}
	frame.Status = status
	_, frames := abcFrameState(frame).VerificationStatus()
	abcFrameState(frame).SetVerificationStatus(status, frames)
}
func abcFrameNFrames(frame *Abc_Frame_t) int32 {
	if frame == nil {
		return 0
	}
	_, frames := abcFrameState(frame).VerificationStatus()
	return frames
}
func abcFrameSetNFrames(frame *Abc_Frame_t, frames int32) {
	if frame == nil {
		return
	}
	frame.NFrames = frames
	status, _ := abcFrameState(frame).VerificationStatus()
	abcFrameState(frame).SetVerificationStatus(status, frames)
}
func abcFrameScalar[T any](frame *Abc_Frame_t, key string) T {
	return abcFrameResource[T](frame, key)
}
func abcFrameSetScalar[T any](frame *Abc_Frame_t, key string, value T) {
	abcFrameSetResource(frame, key, value)
}
func abcFrameFreeIntVec(frame *Abc_Frame_t, key string) {
	if vector := abcFrameTakeResource[*Vec_Int_t](frame, key); vector != nil {
		Vec_IntFree(vector)
	}
}
func abcFrameSlot[T any](frame *Abc_Frame_t, key string) *T {
	if frame == nil {
		return nil
	}
	if slot := abcFrameResource[*T](frame, key); slot != nil {
		return slot
	}
	slot := new(T)
	abcFrameSetResource(frame, key, slot)
	return slot
}
func abcFrameCallback(frame *Abc_Frame_t) abcmain.Abc_Frame_Callback_BmcFrameDone_Func {
	return abcFrameResource[abcmain.Abc_Frame_Callback_BmcFrameDone_Func](frame, "pFuncOnFrameDone")
}
func abcFrameBatchMode(frame *Abc_Frame_t) int32 {
	if frame != nil && frame.BatchMode {
		return 1
	}
	return 0
}
func abcFrameSetBatchMode(frame *Abc_Frame_t, mode int32) {
	if frame != nil {
		frame.BatchMode = mode != 0
	}
}
func abcFrameBridgeMode(frame *Abc_Frame_t) int32 {
	if frame != nil && frame.BridgeMode {
		return 1
	}
	return 0
}
func abcFrameSetBridgeMode(frame *Abc_Frame_t, mode int32) {
	if frame != nil {
		frame.BridgeMode = mode != 0
	}
}

func abcFrameCexVector(frame *Abc_Frame_t) *Vec_Ptr_t {
	if frame == nil {
		return nil
	}
	value, _ := abcFrameState(frame).CexVec().(*Vec_Ptr_t)
	return value
}
func abcFrameSetCexVector(frame *Abc_Frame_t, value *Vec_Ptr_t) {
	if frame == nil {
		return
	}
	var opaque any = value
	abcFrameState(frame).ReplaceCexVec(&opaque)
}
func abcFramePoEquivs(frame *Abc_Frame_t) *Vec_Ptr_t {
	if frame == nil {
		return nil
	}
	value, _ := abcFrameState(frame).PoEquivs().(*Vec_Ptr_t)
	return value
}
func abcFrameSetPoEquivs(frame *Abc_Frame_t, value *Vec_Ptr_t) {
	if frame == nil {
		return
	}
	var opaque any = value
	abcFrameState(frame).ReplacePoEquivs(&opaque)
}
func abcFrameStatuses(frame *Abc_Frame_t) *Vec_Int_t {
	if frame == nil {
		return nil
	}
	value, _ := abcFrameState(frame).Statuses().(*Vec_Int_t)
	return value
}
func abcFrameSetStatuses(frame *Abc_Frame_t, value *Vec_Int_t) {
	if frame == nil {
		return
	}
	var opaque any = value
	abcFrameState(frame).ReplaceStatuses(&opaque)
}

func abcFrameCex2(frame *Abc_Frame_t) *Abc_Cex_t {
	return abcFrameResource[*Abc_Cex_t](frame, "pCex2")
}
func abcFrameSetCex2(frame *Abc_Frame_t, value *Abc_Cex_t) {
	abcFrameSetResource(frame, "pCex2", value)
}
