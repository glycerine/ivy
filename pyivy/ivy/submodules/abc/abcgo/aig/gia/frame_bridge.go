package gia

import (
	"bytes"
	"fmt"

	"github.com/glycerine/abcgo/aig/ioa"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// FrameDataBridge connects the live manager to frame-owned plugin data. Other
// formats delegate to the bridge that was installed before this one.
type FrameDataBridge struct{ Other abcmain.DataBridge }

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
	if format != "gia-aiger" {
		if bridge.Other != nil {
			return bridge.Other.Export(frame, format)
		}
		return nil, fmt.Errorf("gia: unsupported export format %q", format)
	}
	p, ok := frame.Resource("pGia").(*Gia_Man_t)
	if !ok || p == nil {
		return nil, fmt.Errorf("gia: frame has no manager")
	}
	// Snapshot now: a later frame replacement must not invalidate the export.
	return p.AigerModel()
}

type frameWitness interface {
	CounterexampleData() (bits []int32, registers, inputs, frame, output int)
}

func (bridge *FrameDataBridge) Import(frame *abcmain.Frame, format string, value any) error {
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
		p.vFlopClasses = vec.Vec_IntAlloc(len(classes))
		for _, class := range classes {
			vec.Vec_IntPush(p.vFlopClasses, class)
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
