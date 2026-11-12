// Standard-cell library lifecycle and area selection from amapLib.c.
package amap

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
)

func Amap_LibAlloc() *Amap_Lib_t {
	return &Amap_Lib_t{vGates: vec.Vec_PtrAlloc(100), vRules: vec.Vec_PtrAlloc(100), vRulesX: vec.Vec_PtrAlloc(100), vRules3: vec.Vec_IntAlloc(100)}
}
func Amap_LibFree(p *Amap_Lib_t) {
	if p != nil {
		*p = Amap_Lib_t{}
	}
}
func (p *Amap_Lib_t_) Close() error { Amap_LibFree(p); return nil }
func (p *Amap_Lib_t_) Name() string { return p.pName }
func (p *Amap_Lib_t_) Gates() []*Amap_Gat_t {
	if p == nil || p.vGates == nil {
		return nil
	}
	out := make([]*Amap_Gat_t, 0, vec.Vec_PtrSize(p.vGates))
	for _, entry := range vec.Vec_PtrArray(p.vGates) {
		if entry != nil {
			out = append(out, entry.(*Amap_Gat_t))
		}
	}
	return out
}
func Amap_LibNumPinsMax(p *Amap_Lib_t) int32 {
	var maximum uint32
	for _, gate := range p.Gates() {
		maximum = max(maximum, gate.nPins)
	}
	return int32(maximum)
}
func Amap_LibWritePin(w io.Writer, pin *Amap_Pin_t) error {
	phases := [3]string{"UNKNOWN", "INV", "NONINV"}
	if pin == nil || pin.Phase < 0 || pin.Phase > 2 {
		return fmt.Errorf("invalid AMAP pin phase")
	}
	_, err := fmt.Fprintf(w, "    PIN %s %s %.17g %.17g %.17g %.17g %.17g %.17g\n", pin.pName, phases[pin.Phase], pin.dLoadInput, pin.dLoadMax, pin.dDelayBlockRise, pin.dDelayFanoutRise, pin.dDelayBlockFall, pin.dDelayFanoutFall)
	return err
}

// amapCheckedWriter retains errors from legacy printers with no error result.
type amapCheckedWriter struct {
	io.Writer
	err error
}

func (w *amapCheckedWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	n, err := w.Writer.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	w.err = err
	return n, err
}

func Amap_LibWriteGate(w io.Writer, gate *Amap_Gat_t, printDsd int32) error {
	if gate == nil || int(gate.nPins) != len(gate.Pins) {
		return fmt.Errorf("invalid AMAP gate pin count")
	}
	if _, err := fmt.Fprintf(w, "GATE %s %.17g %s=%s;\n", gate.pName, gate.dArea, gate.pOutName, gate.pForm); err != nil {
		return err
	}
	if printDsd != 0 && gate.pFunc != nil {
		if _, err := fmt.Fprint(w, "# DSD: "); err != nil {
			return err
		}
		checked := &amapCheckedWriter{Writer: w}
		kit.Kit_DsdPrintFromTruth2(checked, gate.pFunc, int32(gate.nPins))
		if checked.err != nil {
			return checked.err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	for i := range gate.Pins {
		if err := Amap_LibWritePin(w, &gate.Pins[i]); err != nil {
			return err
		}
	}
	return nil
}
func Amap_LibWrite(w io.Writer, lib *Amap_Lib_t, printDsd int32) error {
	if lib == nil {
		return fmt.Errorf("nil AMAP library")
	}
	if _, err := fmt.Fprintf(w, "# Genlib library %q\n", lib.pName); err != nil {
		return err
	}
	for _, gate := range lib.Gates() {
		if err := Amap_LibWriteGate(w, gate, printDsd); err != nil {
			return err
		}
	}
	return nil
}
func Amap_LibCompareGatesByArea(a, b *Amap_Gat_t) int32 {
	if a.dArea < b.dArea {
		return -1
	}
	if a.dArea > b.dArea {
		return 1
	}
	return 0
}
func Amap_LibSortGatesByArea(lib *Amap_Lib_t) *Vec_Ptr_t {
	gates := lib.Gates()
	sort.SliceStable(gates, func(i, j int) bool { return gates[i].dArea < gates[j].dArea })
	sorted := vec.Vec_PtrAlloc(len(gates))
	for _, gate := range gates {
		vec.Vec_PtrPush(sorted, gate)
	}
	return sorted
}
func Amap_LibFindGate(lib *Amap_Lib_t, truth uint32) *Amap_Gat_t {
	if lib == nil {
		return nil
	}
	if lib.vSorted == nil {
		lib.vSorted = Amap_LibSortGatesByArea(lib)
	}
	for _, entry := range vec.Vec_PtrArray(lib.vSorted) {
		gate, ok := entry.(*Amap_Gat_t)
		if ok && gate != nil && gate.nPins <= 5 && len(gate.pFunc) > 0 && gate.pFunc[0] == truth {
			return gate
		}
	}
	return nil
}
func amapTruthKey(gate *Amap_Gat_t) string {
	words := 1
	if gate.nPins > 5 {
		words = 1 << uint(gate.nPins-5)
	}
	if len(gate.pFunc) < words {
		panic("short AMAP gate truth table")
	}
	key := make([]byte, 4+4*words)
	binary.LittleEndian.PutUint32(key, gate.nPins)
	for i, value := range gate.pFunc[:words] {
		binary.LittleEndian.PutUint32(key[4+4*i:], value)
	}
	return string(key)
}
func Amap_LibSelectGates(lib *Amap_Lib_t, verbose int32) *Vec_Ptr_t {
	if lib == nil {
		return nil
	}
	if lib.vSorted == nil {
		lib.vSorted = Amap_LibSortGatesByArea(lib)
	}
	lib.pGate0 = Amap_LibFindGate(lib, 0)
	lib.pGate1 = Amap_LibFindGate(lib, ^uint32(0))
	lib.pGateBuf = Amap_LibFindGate(lib, 0xaaaaaaaa)
	lib.pGateInv = Amap_LibFindGate(lib, 0x55555555)
	selected := vec.Vec_PtrAlloc(100)
	seen := make(map[string]bool)
	for _, entry := range vec.Vec_PtrArray(lib.vSorted) {
		gate, ok := entry.(*Amap_Gat_t)
		if !ok || gate == nil || gate.pFunc == nil || gate.pTwin != nil || gate.nPins > AMAP_MAXINS {
			continue
		}
		key := amapTruthKey(gate)
		if seen[key] {
			continue
		}
		seen[key] = true
		vec.Vec_PtrPush(selected, gate)
	}
	return selected
}
func Amap_LibPrintSelectedGates(lib *Amap_Lib_t, all int32) error {
	array := lib.vSelect
	if all != 0 {
		array = lib.vGates
	}
	if array == nil {
		return nil
	}
	for index, entry := range vec.Vec_PtrArray(array) {
		gate := entry.(*Amap_Gat_t)
		if _, err := fmt.Fprintf(os.Stdout, "%3d: %s inputs=%d area=%.2f %s=%s\n", index, gate.pName, gate.nPins, gate.dArea, gate.pOutName, gate.pForm); err != nil {
			return err
		}
		if gate.pFunc != nil {
			kit.Kit_DsdPrintFromTruth2(os.Stdout, gate.pFunc, int32(gate.nPins))
			fmt.Println()
		}
	}
	return nil
}
func Amap_LibReadAndPrepare(filename, buffer string, verbose, veryVerbose int32) *Amap_Lib_t {
	var lib *Amap_Lib_t
	if buffer == "" {
		lib = Amap_LibReadFile(filename, verbose)
	} else {
		lib = Amap_LibReadBuffer(buffer, verbose)
		if lib != nil {
			lib.pName = filename
		}
	}
	if lib == nil {
		return nil
	}
	if Amap_LibParseEquations(lib, verbose) == 0 {
		Amap_LibFree(lib)
		return nil
	}
	lib.vSorted = Amap_LibSortGatesByArea(lib)
	lib.vSelect = Amap_LibSelectGates(lib, verbose)
	Amap_LibCreateRules(lib, veryVerbose)
	return lib
}
