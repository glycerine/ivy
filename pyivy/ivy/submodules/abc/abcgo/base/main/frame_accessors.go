package abcmain

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/dec"
)

func (f *Frame) StringValue(key string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.strings[key]
}
func (f *Frame) SetStringValue(key, value string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.closed {
		f.strings[key] = value
	}
}
func (f *Frame) IntValues(key string) []int32 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return append([]int32(nil), f.ints[key]...)
}
func (f *Frame) SetIntValues(key string, values []int32) {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return
	}
	f.ints[key] = append([]int32(nil), values...)
	f.mu.Unlock()
}
func (f *Frame) ListValues(key string) []any {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return append([]any(nil), f.lists[key]...)
}
func (f *Frame) SetListValues(key string, values []any) {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return
	}
	f.lists[key] = append([]any(nil), values...)
	f.mu.Unlock()
}
func (f *Frame) Rows(key string) [][]int32 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	rows := make([][]int32, len(f.rows[key]))
	for i, row := range f.rows[key] {
		rows[i] = append([]int32(nil), row...)
	}
	return rows
}
func (f *Frame) SetRows(key string, rows [][]int32) {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return
	}
	copy := make([][]int32, len(rows))
	for i, row := range rows {
		copy[i] = append([]int32(nil), row...)
	}
	f.rows[key] = copy
	f.mu.Unlock()
}
func (f *Frame) AppendStore(network any) {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return
	}
	f.lists["vStore"] = append(f.lists["vStore"], network)
	f.mu.Unlock()
}
func Abc_FrameReadLibLut() any       { return Abc_FrameGetGlobalFrame().Resource("pLibLut") }
func Abc_FrameSetLibLut(value any)   { Abc_FrameGetGlobalFrame().SetResource("pLibLut", value) }
func Abc_FrameReadLibBox() any       { return Abc_FrameGetGlobalFrame().Resource("pLibBox") }
func Abc_FrameSetLibBox(value any)   { Abc_FrameGetGlobalFrame().SetResource("pLibBox", value) }
func Abc_FrameReadLibGen() any       { return Abc_FrameGetGlobalFrame().Resource("pLibGen") }
func Abc_FrameSetLibGen(value any)   { Abc_FrameGetGlobalFrame().SetResource("pLibGen", value) }
func Abc_FrameReadLibGen2() any      { return Abc_FrameGetGlobalFrame().Resource("pLibGen2") }
func Abc_FrameSetLibGen2(value any)  { Abc_FrameGetGlobalFrame().SetResource("pLibGen2", value) }
func Abc_FrameReadLibSuper() any     { return Abc_FrameGetGlobalFrame().Resource("pLibSuper") }
func Abc_FrameSetLibSuper(value any) { Abc_FrameGetGlobalFrame().SetResource("pLibSuper", value) }
func Abc_FrameReadLibScl() any       { return Abc_FrameGetGlobalFrame().Resource("pLibScl") }
func Abc_FrameSetLibScl(value any)   { Abc_FrameGetGlobalFrame().SetResource("pLibScl", value) }
func Abc_FrameReadLibVer() any       { return Abc_FrameGetGlobalFrame().Resource("pLibVer") }
func Abc_FrameSetLibVer(value any)   { Abc_FrameGetGlobalFrame().SetResource("pLibVer", value) }
func Abc_FrameReadManDsd() any       { return Abc_FrameGetGlobalFrame().Resource("pManDsd") }
func Abc_FrameSetManDsd(value any)   { Abc_FrameGetGlobalFrame().SetResource("pManDsd", value) }
func Abc_FrameReadManDsd2() any      { return Abc_FrameGetGlobalFrame().Resource("pManDsd2") }
func Abc_FrameSetManDsd2(value any)  { Abc_FrameGetGlobalFrame().SetResource("pManDsd2", value) }
func Abc_FrameReadManDd() any        { return Abc_FrameGetGlobalFrame().Resource("dd") }
func Abc_FrameSetManDd(value any)    { Abc_FrameGetGlobalFrame().SetResource("dd", value) }
func Abc_FrameReadManDec() any {
	f := Abc_FrameGetGlobalFrame()
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.manDec == nil {
		f.manDec = dec.Dec_ManStart()
	}
	return f.manDec
}
func Abc_FrameReadGia(f *Frame) any          { return f.Resource("pGia") }
func Abc_FrameUpdateGia(f *Frame, value any) { f.SetResource("pGia", value) }
func Abc_FrameGetGia(f *Frame) any           { return f.TakeResource("pGia") }
func Abc_FrameReadStore() []any              { return Abc_FrameGetGlobalFrame().ListValues("vStore") }
func Abc_FrameReadStoreSize() int32          { return int32(len(Abc_FrameReadStore())) }
func Abc_FrameSetNtkStore(network any)       { Abc_FrameGetGlobalFrame().AppendStore(network) }
func Abc_FrameSetNtkStoreSize(size int32) {
	if size < 0 {
		panic("abcmain: negative store size")
	}
	f := Abc_FrameGetGlobalFrame()
	f.mu.Lock()
	defer f.mu.Unlock()
	store := f.lists["vStore"]
	if int(size) > len(store) {
		store = append(store, make([]any, int(size)-len(store))...)
	} else {
		store = store[:size]
	}
	f.lists["vStore"] = store
}
func Abc_FrameReadFlag(name string) string {
	value, _ := Abc_FrameGetGlobalFrame().LookupFlag(name)
	return value
}
func Abc_FrameSetFlag(name, value string) { Abc_FrameGetGlobalFrame().SetFlag(name, value) }
func Abc_FrameIsFlagEnabled(name string) int32 {
	if Abc_FrameGetGlobalFrame().FlagEnabled(name) {
		return 1
	}
	return 0
}
func Abc_FrameIsBatchMode() int32 {
	f := Abc_FrameReadGlobalFrame()
	if f == nil {
		return 0
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.BatchMode {
		return 1
	}
	return 0
}
func Abc_FrameSetBatchMode(mode int32) {
	if f := Abc_FrameReadGlobalFrame(); f != nil {
		f.mu.Lock()
		f.BatchMode = mode != 0
		f.mu.Unlock()
	}
}
func Abc_FrameIsBridgeMode() int32 {
	f := Abc_FrameReadGlobalFrame()
	if f == nil {
		return 0
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.BridgeMode {
		return 1
	}
	return 0
}
func Abc_FrameSetBridgeMode() {
	if f := Abc_FrameReadGlobalFrame(); f != nil {
		f.mu.Lock()
		f.BridgeMode = true
		f.mu.Unlock()
	}
}
func Abc_FrameReadBmcFrames(f *Frame) int32  { f.mu.RLock(); defer f.mu.RUnlock(); return f.NFrames }
func Abc_FrameReadProbStatus(f *Frame) int32 { f.mu.RLock(); defer f.mu.RUnlock(); return f.Status }
func Abc_FrameReadCex(f *Frame) any          { f.mu.RLock(); defer f.mu.RUnlock(); return f.Cex }
func Abc_FrameSetCex(cex *Counterexample) {
	f := Abc_FrameGetGlobalFrame()
	f.mu.Lock()
	f.Cex = cex
	f.mu.Unlock()
}
func Abc_FrameSetNFrames(frames int32) {
	f := Abc_FrameGetGlobalFrame()
	f.mu.Lock()
	f.Cex = nil
	f.NFrames = frames
	f.mu.Unlock()
}
func Abc_FrameSetStatus(status int32) {
	f := Abc_FrameGetGlobalFrame()
	f.mu.Lock()
	f.Cex = nil
	f.Status = status
	f.mu.Unlock()
}
func frameCex(f *Frame) *Counterexample {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.Cex == nil {
		return &Counterexample{Output: -1, Frame: -1}
	}
	copy := *f.Cex
	return &copy
}
func Abc_FrameReadCexPiNum(f *Frame) int32  { return frameCex(f).PIs }
func Abc_FrameReadCexRegNum(f *Frame) int32 { return frameCex(f).Registers }
func Abc_FrameReadCexPo(f *Frame) int32     { return frameCex(f).Output }
func Abc_FrameReadCexFrame(f *Frame) int32  { return frameCex(f).Frame }
func Abc_FrameReplaceCex(f *Frame, value **Counterexample) {
	f.mu.Lock()
	f.Cex, *value = *value, nil
	delete(f.lists, "vCexVec")
	f.mu.Unlock()
}
func Abc_FrameReadSignalNames() []any { return Abc_FrameGetGlobalFrame().ListValues("vSignalNames") }
func Abc_FrameSetSignalNames(value []any) {
	Abc_FrameGetGlobalFrame().SetListValues("vSignalNames", value)
}
func Abc_FrameReadCexVec(f *Frame) []any       { return f.ListValues("vCexVec") }
func Abc_FrameSetCexVec(value []any)           { Abc_FrameGetGlobalFrame().SetListValues("vCexVec", value) }
func Abc_FrameReadPoEquivs(f *Frame) []any     { return f.ListValues("vPoEquivs") }
func Abc_FrameSetPoEquivs(value []any)         { Abc_FrameGetGlobalFrame().SetListValues("vPoEquivs", value) }
func Abc_FrameReadStatusVec(f *Frame) []int32  { return f.IntValues("vStatuses") }
func Abc_FrameSetStatusVec(value []int32)      { Abc_FrameGetGlobalFrame().SetIntValues("vStatuses", value) }
func Abc_FrameReadPoStatuses(f *Frame) []int32 { return f.IntValues("vStatuses") }
func Abc_FrameSetPoStatuses(value []int32) {
	Abc_FrameGetGlobalFrame().SetIntValues("vStatuses", value)
}
func Abc_FrameReadObjIds(f *Frame) []int32       { return f.IntValues("vAbcObjIds") }
func Abc_FrameSetObjIds(value []int32)           { Abc_FrameGetGlobalFrame().SetIntValues("vAbcObjIds", value) }
func Abc_FrameReadArrayMapping(f *Frame) []int32 { return f.IntValues("pArray") }
func Abc_FrameSetArrayMapping(value []int32)     { Abc_FrameGetGlobalFrame().SetIntValues("pArray", value) }
func Abc_FrameReadBoxes(f *Frame) []int32        { return f.IntValues("pBoxes") }
func Abc_FrameSetBoxes(value []int32)            { Abc_FrameGetGlobalFrame().SetIntValues("pBoxes", value) }
func Abc_FrameReadSpecName() string              { return Abc_FrameGetGlobalFrame().StringValue("pSpecName") }
func Abc_FrameSetSpecName(value string)          { Abc_FrameGetGlobalFrame().SetStringValue("pSpecName", value) }
func Abc_FrameReadDrivingCell() string           { return Abc_FrameGetGlobalFrame().StringValue("pDrivingCell") }
func Abc_FrameSetDrivingCell(value string) {
	Abc_FrameGetGlobalFrame().SetStringValue("pDrivingCell", value)
}
func Abc_FrameSetInv(value []int32) { Abc_FrameGetGlobalFrame().SetIntValues("pAbcWlcInv", value) }
func Abc_FrameSetCnf(value []int32) { Abc_FrameGetGlobalFrame().SetIntValues("cnf", value) }
func Abc_FrameSetStr(value string)  { Abc_FrameGetGlobalFrame().SetStringValue("str", value) }
func Abc_FrameReadMaxLoad() float32 {
	f := Abc_FrameGetGlobalFrame()
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.MaxLoad
}
func Abc_FrameSetMaxLoad(load float32) {
	f := Abc_FrameGetGlobalFrame()
	f.mu.Lock()
	f.MaxLoad = load
	f.mu.Unlock()
}
func Abc_FrameReadJsonStrs(f *Frame) any       { return f.Resource("pJsonStrs") }
func Abc_FrameSetJsonStrs(value any)           { Abc_FrameGetGlobalFrame().SetResource("pJsonStrs", value) }
func Abc_FrameReadJsonObjs(f *Frame) [][]int32 { return f.Rows("vJsonObjs") }
func Abc_FrameSetJsonObjs(value [][]int32)     { Abc_FrameGetGlobalFrame().SetRows("vJsonObjs", value) }
func Abc_FrameInputNdr(f *Frame, value any)    { f.SetResource("pNdr", value) }
func Abc_FrameOutputNdr(f *Frame) any          { return f.TakeResource("pNdr") }
func Abc_FrameOutputNdrArray(f *Frame) []int32 {
	f.mu.Lock()
	defer f.mu.Unlock()
	array := f.ints["pNdrArray"]
	delete(f.ints, "pNdrArray")
	return array
}
func Abc_FrameSetSave1(value any) { Abc_FrameGetGlobalFrame().SetResource("pSave1", value) }
func Abc_FrameSetSave2(value any) { Abc_FrameGetGlobalFrame().SetResource("pSave2", value) }
func Abc_FrameReadSave1() any     { return Abc_FrameGetGlobalFrame().TakeResource("pSave1") }
func Abc_FrameReadSave2() any     { return Abc_FrameGetGlobalFrame().TakeResource("pSave2") }
func Abc_FrameCheckPoConst(f *Frame, index int32) int32 {
	network := f.Network()
	if resourceNil(network) {
		return -1
	}
	if checker, ok := f.Networks.(interface{ OutputConstant(any, int) int32 }); ok {
		return checker.OutputConstant(network, int(index))
	}
	return -1
}
func Abc_FrameCheckPoConstTest(f *Frame) {
	if f.Networks == nil {
		return
	}
	for i := 0; i < f.Networks.Outputs(f.Network()); i++ {
		fmt.Fprintf(f.Out, "%d = %d\n", i, Abc_FrameCheckPoConst(f, int32(i)))
	}
}
