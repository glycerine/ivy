package gia

import (
	"bytes"
	"fmt"

	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/aig/miniaig"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// FrameDataBridge connects the live manager to frame-owned plugin data. Other
// formats delegate to the bridge that was installed before this one.
type FrameDataBridge struct {
	Other         abcmain.DataBridge
	ComputeEquivs func(*Gia_Man_t, int32, int32) *Gia_Man_t
}

// InstallFrameEquivalenceComputer connects the SAT sweeping package without
// introducing a GIA-to-CEC import cycle.
func InstallFrameEquivalenceComputer(frame *abcmain.Frame, compute func(*Gia_Man_t, int32, int32) *Gia_Man_t) {
	InstallFrameBridge(frame)
	frame.Bridge.(*FrameDataBridge).ComputeEquivs = compute
}

func InstallFrameBridge(frame *abcmain.Frame) {
	if _, ok := frame.Bridge.(*FrameDataBridge); !ok {
		frame.Bridge = &FrameDataBridge{Other: frame.Bridge}
	}
}

// CloseFrameResource implements frame ownership without exposing C allocation.
func (p *Gia_Man_t) CloseFrameResource() error {
	if p != nil {
		Gia_ManStop(p)
	}
	return nil
}

// AigerModel returns independent wire data; unsupported native metadata or real
// gates produce the codec's explicit error rather than an incomplete snapshot.
func (p *Gia_Man_t) AigerModel() (*ioa.Model, error) {
	data, err := MarshalAiger(p, AigerWriteOptions{})
	if err != nil {
		return nil, err
	}
	return ioa.Read(data)
}

func (bridge *FrameDataBridge) Export(frame *abcmain.Frame, format string) (any, error) {
	if format == "mini-lut-name-mapping" {
		aig, ok1 := frame.Resource("pGiaMiniAig").(*Gia_Man_t)
		lut, ok2 := frame.Resource("pGiaMiniLut").(*Gia_Man_t)
		aigMap, ok3 := frame.Resource("vCopyMiniAig").(*Vec_Int_t)
		lutMap, ok4 := frame.Resource("vCopyMiniLut").(*Vec_Int_t)
		if !ok1 || !ok2 || !ok3 || !ok4 || aig == nil || lut == nil || aigMap == nil || lutMap == nil || bridge.ComputeEquivs == nil {
			return nil, fmt.Errorf("gia: MiniLUT name mapping needs snapshots and an equivalence computer")
		}
		if Gia_ManCiNum(aig) != Gia_ManCiNum(lut) || Gia_ManCoNum(aig) != Gia_ManCoNum(lut) {
			return nil, fmt.Errorf("gia: MiniAIG and MiniLUT interfaces differ")
		}
		combined := Gia_ManDup2(aig, lut)
		defer Gia_ManStop(combined)
		reduced := bridge.ComputeEquivs(combined, 1000, 0)
		if reduced == nil || combined.PReprs == nil {
			return nil, fmt.Errorf("gia: equivalence computation failed")
		}
		Gia_ManStop(reduced)
		return Gia_ManMapMiniLut2MiniAig(combined, aig, lut, aigMap, lutMap), nil
	}
	if format == "mini-aig-equivalence-classes" {
		original, ok1 := frame.Resource("pGiaMiniAig").(*Gia_Man_t)
		mapped, ok2 := frame.Resource("vCopyMiniAig").(*Vec_Int_t)
		classes, ok3 := frame.Resource("pGia2").(*Gia_Man_t)
		if !ok1 || !ok2 || !ok3 || original == nil || mapped == nil || classes == nil || classes.PReprs == nil || classes.nObjs != original.nObjs {
			return nil, fmt.Errorf("gia: MiniAIG equivalence data is incomplete")
		}
		result := Gia_ManMapEquivAfterScorr(classes, mapped)
		return append([]int32(nil), vec.Vec_IntArray(result)...), nil
	}
	if format != "gia-aiger" && format != "gia-mini-aig" && format != "gia-mini-lut" && format != "gia-mini-lut-attributes" {
		if bridge.Other != nil {
			return bridge.Other.Export(frame, format)
		}
		return nil, fmt.Errorf("gia: unsupported export format %q", format)
	}
	p, ok := frame.Resource("pGia").(*Gia_Man_t)
	if !ok || p == nil {
		return nil, fmt.Errorf("gia: frame has no manager")
	}
	switch format {
	case "gia-mini-aig":
		return Gia_ManToMiniAig(p), nil
	case "gia-mini-lut":
		if Gia_ManHasMapping(p) == 0 {
			return nil, fmt.Errorf("gia: MiniLUT export requires mapping")
		}
		mini := Gia_ManToMiniLut(p)
		var copies *Vec_Int_t
		snapshot := Gia_ManFromMiniLut(mini, &copies)
		frame.SetResource("pGiaMiniLut", snapshot)
		frame.SetResource("vCopyMiniLut", copies)
		return mini, nil
	case "gia-mini-lut-attributes":
		mini, ok := frame.Resource("miniLutAttrInput").(*miniaig.Mini_Lut_t)
		if !ok || mini == nil || Gia_ManHasMapping(p) == 0 {
			return nil, fmt.Errorf("gia: MiniLUT attributes require a mapped manager and LUT")
		}
		for id := int32(1); id < p.nObjs; id++ {
			if Gia_ObjIsLut(p, id) != 0 && p.pObjs[id].Value >= uint32(miniaig.Mini_LutNodeNum(mini)) {
				return nil, fmt.Errorf("gia: MiniLUT does not match the last export")
			}
		}
		return string(Gia_ManToMiniLutAttr(p, mini)), nil
	default:
		return p.AigerModel()
	}
}

type frameWitness interface {
	CounterexampleData() (bits []int32, registers, inputs, frame, output int)
}

func (bridge *FrameDataBridge) Import(frame *abcmain.Frame, format string, value any) error {
	if format == "gia-mini-lut-attributes" {
		mini, ok := value.(*miniaig.Mini_Lut_t)
		if !ok || mini == nil || !miniaig.Mini_LutCheck(mini) {
			return fmt.Errorf("gia: invalid MiniLUT")
		}
		frame.SetResource("miniLutAttrInput", mini)
		return nil
	}
	if format == "gia-mini-aig" {
		mini, ok := value.(*miniaig.Mini_Aig_t)
		if !ok || mini == nil || !miniaig.Mini_AigCheck(mini) {
			return fmt.Errorf("gia: invalid MiniAIG")
		}
		var copies *Vec_Int_t
		p := Gia_ManFromMiniAig(mini, &copies)
		frame.SetResource("pGiaMiniAig", Gia_ManDup(p))
		frame.SetResource("vCopyMiniAig", copies)
		frame.SetResource("pGia", p)
		return nil
	}
	if format == "gia-mini-lut" {
		mini, ok := value.(*miniaig.Mini_Lut_t)
		if !ok || mini == nil || !miniaig.Mini_LutCheck(mini) {
			return fmt.Errorf("gia: invalid MiniLUT")
		}
		frame.SetResource("pGia", Gia_ManFromMiniLut(mini, nil))
		return nil
	}
	if format == "gia-aiger" {
		provider, ok := value.(ioa.AigerProvider)
		if !ok {
			return fmt.Errorf("gia: AIGER import needs an AigerProvider")
		}
		var buffer bytes.Buffer
		if err := ioa.Write(&buffer, provider, ioa.WriteOptions{Symbols: true}); err != nil {
			return err
		}
		if buffer.Len() > int(^uint32(0)>>1) {
			return fmt.Errorf("gia: AIGER import exceeds supported size")
		}
		p, err := Gia_AigerReadFromMemory(buffer.Bytes(), int32(buffer.Len()), 0, 1, 1)
		if err != nil {
			return err
		}
		frame.SetResource("pGia", p)
		return nil
	}
	if format != "gia-flop-classes" && format != "gia-counterexample" {
		if bridge.Other != nil {
			return bridge.Other.Import(frame, format, value)
		}
		return fmt.Errorf("gia: unsupported import format %q", format)
	}
	p, ok := frame.Resource("pGia").(*Gia_Man_t)
	if !ok || p == nil {
		return fmt.Errorf("gia: frame has no manager")
	}
	if format == "gia-flop-classes" {
		classes, ok := value.([]int32)
		if !ok || len(classes) != int(p.nRegs) {
			return fmt.Errorf("gia: invalid flop class dimensions")
		}
		p.VFlopClasses = vec.Vec_IntAlloc(len(classes))
		for _, class := range classes {
			vec.Vec_IntPush(p.VFlopClasses, class)
		}
		return nil
	}
	witness, ok := value.(frameWitness)
	if !ok {
		return fmt.Errorf("gia: counterexample import needs CounterexampleData")
	}
	bits, registers, inputs, lastFrame, output := witness.CounterexampleData()
	if registers != int(p.nRegs) || inputs != int(Gia_ManPiNum(p)) || lastFrame < 0 || lastFrame >= int(^uint32(0)>>1) || output < 0 || output >= int(Gia_ManPoNum(p)) {
		return fmt.Errorf("gia: invalid imported counterexample dimensions")
	}
	count := int64(registers) + (int64(lastFrame)+1)*int64(inputs)
	if count > int64(^uint32(0)>>1) || count != int64(len(bits)) {
		return fmt.Errorf("gia: invalid imported counterexample bit count")
	}
	cex := util.Abc_CexAlloc(int32(registers), int32(inputs), int32(lastFrame+1))
	cex.IPo, cex.IFrame = int32(output), int32(lastFrame)
	packed := make([]uint64, (count+63)/64)
	for i, bit := range bits {
		if bit != 0 && bit != 1 {
			return fmt.Errorf("gia: invalid counterexample bit %d", i)
		}
		if bit != 0 {
			cex.PData[i>>5] |= 1 << uint(i&31)
			packed[i>>6] |= 1 << uint(i&63)
		}
	}
	if err := p.ValidateSequentialCounterexample(cex, true); err != nil {
		return err
	}
	if err := p.SetSequentialCounterexample(cex); err != nil {
		return err
	}
	model := &abcmain.Counterexample{PIs: int32(inputs), Registers: int32(registers), Output: int32(output), Frame: int32(lastFrame), Bits: packed}
	abcmain.Abc_FrameReplaceCex(frame, &model)
	return nil
}

// Abc_FrameMiniAigSetCiArrivals installs independent CI arrival times.
func Abc_FrameMiniAigSetCiArrivals(frame *abcmain.Frame, arrivals []int32) {
	if arrivals == nil {
		fmt.Println("Arrival times are not given.")
		return
	}
	if frame == nil {
		fmt.Println("ABC framework is not initialized by calling Abc_Start().")
		return
	}
	p, ok := frame.Resource("pGia").(*Gia_Man_t)
	if !ok || p == nil {
		fmt.Println("Current network in ABC framework is not defined.")
		return
	}
	if len(arrivals) != int(Gia_ManCiNum(p)) {
		panic("gia: CI arrival count mismatch")
	}
	p.vCiArrs = vec.Vec_IntAllocArrayCopy(arrivals, len(arrivals))
}

// Abc_FrameMiniAigSetCoRequireds installs independent CO required times.
func Abc_FrameMiniAigSetCoRequireds(frame *abcmain.Frame, required []int32) {
	if required == nil {
		fmt.Println("Required times are not given.")
		return
	}
	if frame == nil {
		fmt.Println("ABC framework is not initialized by calling Abc_Start().")
		return
	}
	p, ok := frame.Resource("pGia").(*Gia_Man_t)
	if !ok || p == nil {
		fmt.Println("Current network in ABC framework is not defined.")
		return
	}
	if len(required) != int(Gia_ManCoNum(p)) {
		panic("gia: CO required count mismatch")
	}
	p.vCoReqs = vec.Vec_IntAllocArrayCopy(required, len(required))
}

// Abc_FrameMiniAigReadCoArrivals returns an independent copy of CO arrivals.
func Abc_FrameMiniAigReadCoArrivals(frame *abcmain.Frame) []int32 {
	if frame == nil {
		fmt.Println("ABC framework is not initialized by calling Abc_Start().")
		return nil
	}
	p, ok := frame.Resource("pGia").(*Gia_Man_t)
	if !ok || p == nil {
		fmt.Println("Current network in ABC framework is not defined.")
		return nil
	}
	if p.vCoArrs == nil {
		fmt.Println("Current network in ABC framework has no CO arrival times.")
		return nil
	}
	return append([]int32(nil), vec.Vec_IntArray(p.vCoArrs)...)
}

func Abc_FrameMiniAigSetAndGateDelay(frame *abcmain.Frame, delay int32) {
	if frame == nil {
		fmt.Println("ABC framework is not initialized by calling Abc_Start().")
		return
	}
	p, ok := frame.Resource("pGia").(*Gia_Man_t)
	if !ok || p == nil {
		fmt.Println("Current network in ABC framework is not defined.")
		return
	}
	p.And2Delay = delay
}
