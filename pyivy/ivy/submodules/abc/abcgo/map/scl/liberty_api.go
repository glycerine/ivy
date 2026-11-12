package scl

import (
	"fmt"
	"os"
)

func Scl_LibertyItemNum(p *Scl_Tree_t, root *LibertyItem, name string) int {
	n := 0
	for _, c := range root.Children {
		if c.Key == name {
			n++
		}
	}
	return n
}
func Scl_LibertyReadCellIsFlop(p *Scl_Tree_t, cell *LibertyItem) bool {
	return cell.child("ff") != nil || cell.child("latch") != nil || cell.child("statetable") != nil
}
func Scl_LibertyReadCellIsDontUse(p *Scl_Tree_t, cell *LibertyItem) bool {
	return cell.value("dont_use") == "true"
}
func Scl_LibertyReadCellArea(p *Scl_Tree_t, cell *LibertyItem) string { return cell.value("area") }
func Scl_LibertyReadCellLeakage(p *Scl_Tree_t, cell *LibertyItem) string {
	return cell.value("cell_leakage_power")
}
func Scl_LibertyReadPinFormula(p *Scl_Tree_t, pin *LibertyItem) string { return pin.value("function") }
func Scl_LibertyReadCellIsThreeState(p *Scl_Tree_t, cell *LibertyItem) bool {
	for _, pin := range cell.Children {
		if pin.Key == "pin" && pin.child("three_state") != nil {
			return true
		}
	}
	return false
}
func Scl_LibertyReadCellOutputNum(p *Scl_Tree_t, cell *LibertyItem) int {
	n := 0
	for _, pin := range cell.Children {
		if pin.Key == "pin" && pin.value("function") != "" {
			n++
		}
	}
	return n
}
func Scl_LibertyReadDefaultWireLoad(p *Scl_Tree_t) string { return p.Root.value("default_wire_load") }
func Scl_LibertyReadDefaultWireLoadSel(p *Scl_Tree_t) string {
	return p.Root.value("default_wire_load_selection")
}
func Scl_LibertyReadDefaultMaxTrans(p *Scl_Tree_t) float32 {
	return p.Root.number("default_max_transition")
}
func Scl_LibertyReadDeriveStrength(p *Scl_Tree_t, cell *LibertyItem) int {
	return int(cell.number("drive_strength"))
}
func Scl_LibertyReadPinDirection(p *Scl_Tree_t, pin *LibertyItem) int {
	switch pin.value("direction") {
	case "input":
		return 0
	case "output":
		return 1
	default:
		return -1
	}
}
func Scl_LibertyReadPinCap(p *Scl_Tree_t, pin *LibertyItem, name string) float32 {
	return pin.number(name)
}
func Scl_LibertyReadTimingSense(p *Scl_Tree_t, timing *LibertyItem) SC_TSense {
	switch timing.value("timing_sense") {
	case "positive_unate":
		return ScTSPos
	case "negative_unate":
		return ScTSNeg
	default:
		return ScTSNon
	}
}
func Scl_LibertyReadGenlibStr(p *Scl_Tree_t, verbose int32) (string, error) {
	lib, e := libraryFromTree(p)
	if e != nil {
		return "", e
	}
	return Abc_SclProduceGenlibStrSimple(lib), nil
}
func Scl_LibertyParseGenlibStr(path string, verbose int32) (string, error) {
	p, e := Scl_LibertyParse(path, verbose)
	if e != nil {
		return "", e
	}
	return Scl_LibertyReadGenlibStr(p, verbose)
}
func Scl_LibertyReadSclStr(p *Scl_Tree_t, verbose, veryVerbose int32) ([]byte, error) {
	lib, e := libraryFromTree(p)
	if e != nil {
		return nil, e
	}
	return MarshalSCL(lib)
}
func Scl_LibertyStringDump(path string, data []byte) error { return os.WriteFile(path, data, 0644) }
func Scl_LibertyFileSize(path string) (int64, error) {
	s, e := os.Stat(path)
	if e != nil {
		return 0, e
	}
	return s.Size(), nil
}
func Scl_LibertyFileContents(path string, size int) (string, error) {
	data, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	if len(data) != size {
		return "", fmt.Errorf("Liberty file size changed: expected %d got %d", size, len(data))
	}
	return string(data), nil
}
