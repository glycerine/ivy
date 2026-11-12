package abcmain

import (
	"errors"
	"fmt"
	"io"
)

func (f *Frame) Import(format string, value any) error {
	if f.Bridge == nil {
		return fmt.Errorf("abcmain: %s import requires a data bridge", format)
	}
	return f.Bridge.Import(f, format, value)
}
func (f *Frame) Export(format string) (any, error) {
	if f.Bridge == nil {
		return nil, fmt.Errorf("abcmain: %s export requires a data bridge", format)
	}
	return f.Bridge.Export(f, format)
}
func Abc_NtkInputMiniAig(f *Frame, value any) error      { return f.Import("network-mini-aig", value) }
func Abc_NtkOutputMiniAig(f *Frame) (any, error)         { return f.Export("network-mini-aig") }
func Abc_FrameGiaInputMiniAig(f *Frame, value any) error { return f.Import("gia-mini-aig", value) }
func Abc_FrameGiaOutputMiniAig(f *Frame) (any, error)    { return f.Export("gia-mini-aig") }
func Abc_FrameGiaInputMiniLut(f *Frame, value any) error { return f.Import("gia-mini-lut", value) }
func Abc_FrameGiaOutputMiniLut(f *Frame) (any, error)    { return f.Export("gia-mini-lut") }
func Abc_FrameGiaOutputMiniLutAttr(f *Frame, value any) (string, error) {
	if err := f.Import("gia-mini-lut-attributes", value); err != nil {
		return "", err
	}
	result, err := f.Export("gia-mini-lut-attributes")
	if err != nil {
		return "", err
	}
	text, ok := result.(string)
	if !ok {
		return "", errors.New("abcmain: mini-LUT attributes bridge returned a non-string")
	}
	return text, nil
}
func Abc_NtkSetFlopNum(f *Frame, count int32) error { return f.Import("network-flop-count", count) }
func Abc_NtkSetCiArrivalTime(f *Frame, index int32, rise, fall float32) error {
	if f.Timing == nil {
		return errors.New("abcmain: CI arrival time requires a timing bridge")
	}
	return f.Timing.SetTiming(f, "ci-arrival", int(index), rise, fall)
}
func Abc_NtkSetCoRequiredTime(f *Frame, index int32, rise, fall float32) error {
	if f.Timing == nil {
		return errors.New("abcmain: CO required time requires a timing bridge")
	}
	return f.Timing.SetTiming(f, "co-required", int(index), rise, fall)
}
func Abc_NtkSetAndGateDelay(f *Frame, delay float32) error { return f.Import("and-gate-delay", delay) }
func Abc_NtkOutputMiniMapping(f *Frame) ([]int32, error) {
	value, err := f.Export("mini-mapping")
	if err != nil {
		return nil, err
	}
	mapping, ok := value.([]int32)
	if !ok {
		return nil, errors.New("abcmain: mini-mapping bridge returned a non-integer slice")
	}
	return mapping, nil
}
func Abc_NtkPrintMiniMapping(mapping []int32) error {
	f := Abc_FrameGetGlobalFrame()
	printer, ok := f.Bridge.(interface {
		PrintMiniMapping(io.Writer, []int32) error
	})
	if !ok {
		return errors.New("abcmain: mini-mapping printing requires a data bridge")
	}
	return printer.PrintMiniMapping(f.Out, mapping)
}
func Abc_FrameReadMiniAigEquivClasses(f *Frame) ([]int32, error) {
	value, err := f.Export("mini-aig-equivalence-classes")
	if err != nil {
		return nil, err
	}
	classes, ok := value.([]int32)
	if !ok {
		return nil, errors.New("abcmain: equivalence bridge returned a non-integer slice")
	}
	return classes, nil
}
