// Library lifetime, formatting and technology mapping utilities from mioUtils.c.
package mio

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
)

func (lib *Mio_Library_t) Close() error { Mio_LibraryDelete(lib); return nil }
func Mio_LibraryDelete(lib *Mio_Library_t) {
	if lib == nil {
		return
	}
	*lib = Mio_Library_t{}
}
func Mio_GateDelete(g *Mio_Gate_t) {
	if g != nil {
		*g = Mio_Gate_t{}
	}
}
func Mio_PinDelete(p *Mio_Pin_t) {
	if p != nil {
		*p = Mio_Pin_t{}
	}
}
func Mio_PinDup(p *Mio_Pin_t) *Mio_Pin_t {
	if p == nil {
		return nil
	}
	copy := *p
	copy.pNext = nil
	return &copy
}
func Mio_CheckPins(a, b *Mio_Pin_t) int32 {
	if a == nil || b == nil {
		return mioBool(a == b)
	}
	x, y := *a, *b
	x.pName = ""
	y.pName = ""
	x.pNext = nil
	y.pNext = nil
	return mioBool(x == y)
}
func Mio_CheckGates(lib *Mio_Library_t) int32 {
	for g := lib.pGates; g != nil; g = g.pNext {
		for p := g.pPins; p != nil && p.pNext != nil; p = p.pNext {
			if Mio_CheckPins(p, p.pNext) == 0 {
				return 0
			}
		}
	}
	return 1
}
func Mio_WritePin(w io.Writer, p *Mio_Pin_t, nameLen, all int32) error {
	name := p.pName
	if all != 0 {
		name = "*"
	}
	phases := [3]string{MIO_STRING_UNKNOWN, MIO_STRING_INV, MIO_STRING_NONINV}
	if p.Phase < 0 || p.Phase > 2 {
		return fmt.Errorf("invalid pin phase")
	}
	_, err := fmt.Fprintf(w, "\n    PIN %s %s %.17g %.17g %.17g %.17g %.17g %.17g", name, phases[p.Phase], p.dLoadInput, p.dLoadMax, p.dDelayBlockRise, p.dDelayFanoutRise, p.dDelayBlockFall, p.dDelayFanoutFall)
	return err
}
func Mio_WriteGate(w io.Writer, g *Mio_Gate_t, gateLen, nameLen, formLen, sops, all int32) error {
	if _, err := fmt.Fprintf(w, "GATE %s %.17g %s=%s;", g.pName, g.dArea, g.pOutName, g.pForm); err != nil {
		return err
	}
	for p := g.pPins; p != nil; p = p.pNext {
		if err := Mio_WritePin(w, p, nameLen, all); err != nil {
			return err
		}
		if all != 0 {
			break
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if sops != 0 {
		for _, line := range strings.Split(strings.TrimSuffix(g.pSop, "\n"), "\n") {
			if _, err := fmt.Fprintln(w, "# SOP "+line); err != nil {
				return err
			}
		}
	}
	return nil
}
func mioSelectedGates(lib *Mio_Library_t, selected int32) ([]*Mio_Gate_t, error) {
	if selected == 0 {
		return lib.ppGates0, nil
	}
	ids := make(map[*Mio_Gate_t]int32)
	for g := lib.pGates; g != nil; g = g.pNext {
		ids[g] = g.Cell
	}
	defer func() {
		for g, id := range ids {
			g.Cell = id
		}
	}()
	cells := Mio_CollectRootsNew2(lib, 6, nil, 0)
	if cells == nil {
		return nil, fmt.Errorf("library lacks constants, buffer or inverter")
	}
	gates := make([]*Mio_Gate_t, len(cells))
	for i, c := range cells {
		gates[i] = c.pMioGate
	}
	return gates, nil
}
func Mio_WriteLibrary(w io.Writer, lib *Mio_Library_t, sops, short, selected int32) error {
	if lib == nil {
		return fmt.Errorf("nil mapping library")
	}
	gates, err := mioSelectedGates(lib, selected)
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintf(w, "# The genlib library %q with %d gates\n", lib.pName, len(gates)); err != nil {
		return err
	}
	all := mioBool(short != 0 || Mio_CheckGates(lib) != 0)
	for _, g := range gates {
		if err = Mio_WriteGate(w, g, 0, 0, 0, sops, all); err != nil {
			return err
		}
	}
	return nil
}
func expVerilog(n int32, p *Vec_Int_t, names []string, lit int32) string {
	if lit == EXP_CONST0 {
		return "1'b0"
	}
	if lit == EXP_CONST1 {
		return "1'b1"
	}
	var s string
	if lit < 2*n {
		s = mioVerilogName(names[lit/2])
	} else {
		a := vec.Vec_IntArray(p)
		node := int(lit/2 - n)
		s = "(" + expVerilog(n, p, names, a[2*node]) + " & " + expVerilog(n, p, names, a[2*node+1]) + ")"
	}
	if lit&1 != 0 {
		return "~(" + s + ")"
	}
	return s
}
func Exp_PrintLitVerilog(w io.Writer, n int32, p *Vec_Int_t, names *Vec_Ptr_t, lit int32) error {
	array := make([]string, vec.Vec_PtrSize(names))
	for i := range array {
		array[i] = vec.Vec_PtrEntry(names, i).(string)
	}
	_, err := io.WriteString(w, expVerilog(n, p, array, lit))
	return err
}
func Exp_PrintNodeVerilog(w io.Writer, n int32, p *Vec_Int_t, names *Vec_Ptr_t, node, compl int32) error {
	return Exp_PrintLitVerilog(w, n, p, names, 2*(n+node)^compl)
}
func Exp_PrintVerilog(w io.Writer, n int32, p *Vec_Int_t, names *Vec_Ptr_t) error {
	a := vec.Vec_IntArray(p)
	return Exp_PrintLitVerilog(w, n, p, names, a[len(a)-1])
}
func mioVerilogName(name string) string {
	valid := len(name) > 0
	for i, c := range []byte(name) {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || i > 0 && (c >= '0' && c <= '9' || c == '$')) {
			valid = false
			break
		}
	}
	if valid {
		return name
	}
	return "\\" + name + " "
}
func Mio_WriteGateVerilog(w io.Writer, g *Mio_Gate_t, _ *Vec_Ptr_t) error {
	gates := []*Mio_Gate_t{g}
	if g.pTwin != nil {
		gates = append(gates, g.pTwin)
	}
	var inputs, outputs []string
	seen := make(map[string]bool)
	for _, gate := range gates {
		if seen[gate.pOutName] {
			return fmt.Errorf("duplicate Verilog output %s", gate.pOutName)
		}
		seen[gate.pOutName] = true
		outputs = append(outputs, gate.pOutName)
	}
	for _, gate := range gates {
		for p := gate.pPins; p != nil; p = p.pNext {
			if seen[p.pName] {
				for _, output := range outputs {
					if output == p.pName {
						return fmt.Errorf("pin %s is both input and output", p.pName)
					}
				}
				continue
			}
			seen[p.pName] = true
			inputs = append(inputs, p.pName)
		}
	}
	escaped := func(names []string) string {
		result := make([]string, len(names))
		for i, name := range names {
			result[i] = mioVerilogName(name)
		}
		return strings.Join(result, ", ")
	}
	ports := append(append([]string(nil), outputs...), inputs...)
	if _, err := fmt.Fprintf(w, "module %s (%s);\n  output %s;\n", mioVerilogName(g.pName), escaped(ports), escaped(outputs)); err != nil {
		return err
	}
	if len(inputs) > 0 {
		if _, err := fmt.Fprintf(w, "  input %s;\n", escaped(inputs)); err != nil {
			return err
		}
	}
	for _, gate := range gates {
		var names []string
		for p := gate.pPins; p != nil; p = p.pNext {
			names = append(names, p.pName)
		}
		a := vec.Vec_IntArray(gate.vExpr)
		if _, err := fmt.Fprintf(w, "  assign %s = %s;\n", mioVerilogName(gate.pOutName), expVerilog(gate.nInputs, gate.vExpr, names, a[len(a)-1])); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, "endmodule")
	return err
}
func Mio_WriteLibraryVerilog(w io.Writer, lib *Mio_Library_t, sops, short, selected int32) error {
	if lib == nil {
		return fmt.Errorf("nil mapping library")
	}
	gates, err := mioSelectedGates(lib, selected)
	if err != nil {
		return err
	}
	written := make(map[string]bool)
	for _, g := range gates {
		if written[g.pName] {
			continue
		}
		if err = Mio_WriteGateVerilog(w, g, nil); err != nil {
			return err
		}
		written[g.pName] = true
	}
	return nil
}
func Mio_DelayCompare(a, b *Mio_Gate_t) int32 {
	if a.dDelayMax < b.dDelayMax {
		return -1
	}
	if a.dDelayMax > b.dDelayMax {
		return 1
	}
	return int32(strings.Compare(a.pName, b.pName))
}
func Mio_AreaCompare(a, b *Mio_Cell_t) int32 {
	if a.nFanins < b.nFanins {
		return -1
	}
	if a.nFanins > b.nFanins {
		return 1
	}
	if a.Area < b.Area {
		return -1
	}
	if a.Area > b.Area {
		return 1
	}
	return int32(strings.Compare(a.pName, b.pName))
}
func Mio_AreaCompare2(a, b *Mio_Cell2_t) int32 {
	if a.nFanins < b.nFanins {
		return -1
	}
	if a.nFanins > b.nFanins {
		return 1
	}
	if a.AreaW < b.AreaW {
		return -1
	}
	if a.AreaW > b.AreaW {
		return 1
	}
	return int32(strings.Compare(a.pName, b.pName))
}
func Mio_CellDelayAve(c *Mio_Cell_t) float32 {
	var sum float32
	for i := uint32(0); i < c.nFanins; i++ {
		sum += c.Delays[i]
	}
	if c.nFanins > 0 {
		return sum / float32(c.nFanins)
	}
	return 0
}
func Mio_GateDelayAve(g *Mio_Gate_t) float32 {
	var sum float32
	for p := g.pPins; p != nil; p = p.pNext {
		sum += float32((p.dDelayBlockRise + p.dDelayBlockFall) / 2)
	}
	if g.nInputs > 0 {
		return sum / float32(g.nInputs)
	}
	return 0
}
func mioBetterGate(a, b *Mio_Gate_t) bool {
	const eps = .0094636
	if math.Abs(a.dArea-b.dArea) > eps {
		return a.dArea < b.dArea
	}
	da, db := Mio_GateDelayAve(a), Mio_GateDelayAve(b)
	if math.Abs(float64(da-db)) > eps {
		return da < db
	}
	return a.pName < b.pName
}
func Mio_CompareTwoGates(a, b *Mio_Gate_t) int32 { return mioBool(mioBetterGate(b, a)) }
func Mio_CollectRoots(lib *Mio_Library_t, n int32, delay float32, skipInv int32, count *int32, verbose int32) []*Mio_Gate_t {
	var roots []*Mio_Gate_t
	positions := make(map[uint64]int)
	profile := Mio_LibraryHasProfile(lib) != 0
	for g := lib.pGates; g != nil; g = g.pNext {
		if g.nInputs > n || g.nInputs > 6 || g.pTwin != nil || profile && g.Profile == 0 && g.nInputs > 1 || delay > 0 && g.dDelayMax > float64(delay) || g.uTruth == 0 || g.uTruth == ^uint64(0) || g.uTruth == expTruth6[0] || skipInv != 0 && g.uTruth == ^expTruth6[0] {
			continue
		}
		if i, ok := positions[g.uTruth]; ok {
			if mioBetterGate(g, roots[i]) {
				roots[i] = g
			}
		} else {
			positions[g.uTruth] = len(roots)
			roots = append(roots, g)
		}
	}
	sort.SliceStable(roots, func(i, j int) bool { return Mio_DelayCompare(roots[i], roots[j]) < 0 })
	if count != nil {
		*count = int32(len(roots))
	}
	return roots
}
func Mio_CollectCopy(c *Mio_Cell_t, g *Mio_Gate_t) {
	*c = Mio_Cell_t{pName: g.pName, nFanins: uint32(g.nInputs), Area: float32(g.dArea), uTruth: g.uTruth}
	for i, p := 0, g.pPins; p != nil; i, p = i+1, p.pNext {
		if i >= len(c.Delays) {
			panic("technology cell has more than six inputs")
		}
		c.Delays[i] = float32((p.dDelayBlockRise + p.dDelayBlockFall) / 2)
	}
}
func Mio_CollectCopy2(c *Mio_Cell2_t, g *Mio_Gate_t) {
	*c = Mio_Cell2_t{pName: g.pName, vExpr: g.vExpr, nFanins: uint32(g.nInputs), AreaF: float32(g.dArea), AreaW: uint64(1000 * g.dArea), uTruth: g.uTruth, pMioGate: g}
	for i, p := 0, g.pPins; p != nil; i, p = i+1, p.pNext {
		if i >= len(c.iDelays) {
			panic("technology cell has more than six inputs")
		}
		c.iDelays[i] = int32(500 * (p.dDelayBlockRise + p.dDelayBlockFall))
		c.iDelayAve += c.iDelays[i]
	}
	if c.nFanins > 0 {
		c.iDelayAve /= int32(c.nFanins)
	}
}
func Mio_CompareTwo(a *Mio_Cell_t, g *Mio_Gate_t) int32 {
	const eps = .0094636
	if math.Abs(float64(a.Area)-g.dArea) > eps {
		return mioBool(float64(a.Area) > g.dArea)
	}
	if math.Abs(float64(Mio_CellDelayAve(a)-Mio_GateDelayAve(g))) > eps {
		return mioBool(Mio_CellDelayAve(a) > Mio_GateDelayAve(g))
	}
	return mioBool(a.pName > g.pName)
}
func Mio_CompareTwo2(a, b *Mio_Cell2_t) int32 {
	if a.AreaW != b.AreaW {
		return mioBool(a.AreaW > b.AreaW)
	}
	if a.iDelayAve != b.iDelayAve {
		return mioBool(a.iDelayAve > b.iDelayAve)
	}
	return mioBool(a.pName > b.pName)
}
func mioBasicIndex(truth uint64) int {
	switch truth {
	case 0:
		return 0
	case ^uint64(0):
		return 1
	case 0xaaaaaaaaaaaaaaaa:
		return 2
	case 0x5555555555555555:
		return 3
	}
	return -1
}
func Mio_CollectRootsNew(lib *Mio_Library_t, n int32, count *int32, verbose int32) []Mio_Cell_t {
	if count != nil {
		*count = 0
	}
	if lib == nil || n > 6 || n < 0 {
		return nil
	}
	cells := make([]Mio_Cell_t, 4)
	indices := make(map[uint64]int)
	for g := lib.pGates; g != nil; g = g.pNext {
		if g.nInputs > n || g.pTwin != nil {
			continue
		}
		index, exists := indices[g.uTruth]
		if !exists {
			index = mioBasicIndex(g.uTruth)
			if index < 0 {
				index = len(cells)
				cells = append(cells, Mio_Cell_t{})
			}
			indices[g.uTruth] = index
		}
		if cells[index].pName == "" || Mio_CompareTwo(&cells[index], g) != 0 {
			Mio_CollectCopy(&cells[index], g)
		}
	}
	for _, c := range cells[:4] {
		if c.pName == "" {
			return nil
		}
	}
	sort.SliceStable(cells[4:], func(i, j int) bool { return Mio_AreaCompare(&cells[i+4], &cells[j+4]) < 0 })
	for i := range cells {
		cells[i].Id = uint32(i)
	}
	if count != nil {
		*count = int32(len(cells))
	}
	return cells
}
func Mio_CollectRootsNew2(lib *Mio_Library_t, n int32, count *int32, verbose int32) []Mio_Cell2_t {
	if count != nil {
		*count = 0
	}
	if lib == nil || n > 6 || n < 0 {
		return nil
	}
	cells := make([]Mio_Cell2_t, 4)
	indices := make(map[uint64]int)
	for g := lib.pGates; g != nil; g = g.pNext {
		if g.nInputs > n || g.pTwin != nil {
			continue
		}
		index, exists := indices[g.uTruth]
		if !exists {
			index = mioBasicIndex(g.uTruth)
			if index < 0 {
				index = len(cells)
				cells = append(cells, Mio_Cell2_t{})
			}
			indices[g.uTruth] = index
		}
		var candidate Mio_Cell2_t
		Mio_CollectCopy2(&candidate, g)
		if cells[index].pName == "" || Mio_CompareTwo2(&cells[index], &candidate) != 0 {
			cells[index] = candidate
		}
	}
	for _, c := range cells[:4] {
		if c.pName == "" {
			return nil
		}
	}
	sort.SliceStable(cells[4:], func(i, j int) bool { return Mio_AreaCompare2(&cells[i+4], &cells[j+4]) < 0 })
	for g := lib.pGates; g != nil; g = g.pNext {
		g.Cell = -1
	}
	for i := range cells {
		cells[i].Id = uint32(i)
		cells[i].pMioGate.Cell = int32(i)
	}
	if count != nil {
		*count = int32(len(cells))
	}
	return cells
}
func mioCurrentLibrary() *Mio_Library_t {
	lib, _ := abcmain.Abc_FrameReadLibGen().(*Mio_Library_t)
	return lib
}
func Mio_CollectRootsNewDefault(n int32, count *int32, verbose int32) []Mio_Cell_t {
	return Mio_CollectRootsNew(mioCurrentLibrary(), n, count, verbose)
}
func Mio_CollectRootsNewDefault2(n int32, count *int32, verbose int32) []Mio_Cell2_t {
	return Mio_CollectRootsNew2(mioCurrentLibrary(), n, count, verbose)
}
func Mio_CollectRootsNewDefault3(n int32, names **Vec_Ptr_t, truths **Vec_Wrd_t) int32 {
	lib := mioCurrentLibrary()
	if lib == nil {
		return 0
	}
	*names = vec.Vec_PtrAlloc(0)
	*truths = vec.Vec_WrdAlloc(0)
	var count int32
	for _, g := range lib.ppGates0 {
		if g.nInputs > n || g.nInputs > 8 || g.pTwin != nil {
			continue
		}
		vec.Vec_PtrPush(*names, g.pName)
		for w := 0; w < 4; w++ {
			truth := g.uTruth
			if g.nInputs > 6 {
				truth = g.pTruth[w%len(g.pTruth)]
			}
			vec.Vec_WrdPush(*truths, truth)
		}
		count++
	}
	return count
}
func Mio_DeriveTruthTable6(g *Mio_Gate_t) uint64 { return Exp_Truth6(g.nInputs, g.vExpr, nil) }
func Mio_DeriveTruthTable(g *Mio_Gate_t, input [][2]uint32, signs, n int32, result []uint32) {
	if signs != g.nInputs || int(signs) > len(input) || len(result) < 2 {
		panic("invalid gate truth dimensions")
	}
	fanins := make([]uint64, signs)
	for i := range fanins {
		fanins[i] = uint64(input[i][0]) | uint64(input[i][1])<<32
	}
	truth := Exp_Truth6(signs, g.vExpr, fanins)
	result[0], result[1] = uint32(truth), uint32(truth>>32)
}
func Mio_DeriveTruthTable2(g *Mio_Gate_t, input [][2]uint32, signs, n int32, result []uint32) {
	Mio_DeriveTruthTable(g, input, signs, n, result)
	if n < 5 {
		result[0] &= MIO_MASK(1 << uint(n))
		result[1] = 0
	} else if n == 5 {
		result[1] = 0
	}
}
func Mio_SopGetVarNum(s string) int32 {
	end := strings.IndexByte(s, '\n')
	if end < 2 {
		return -1
	}
	return int32(end - 2)
}
func Mio_DeriveGateDelays(g *Mio_Gate_t, inputs [][]float32, pins, n int32, zero float32, result []float32, maximum *float32) {
	if pins != g.nInputs || int(pins) > len(inputs) || int(n) > len(result) {
		panic("invalid gate delay dimensions")
	}
	*maximum = 0
	for i := int32(0); i < n; i++ {
		result[i] = zero
		for k, p := 0, g.pPins; p != nil; k, p = k+1, p.pNext {
			if inputs[k][i] >= 0 {
				result[i] = max(result[i], inputs[k][i]+float32(p.dDelayBlockMax))
			}
		}
		*maximum = max(*maximum, result[i])
	}
}
func Mio_GateCreatePseudo(n int32) *Mio_Gate_t {
	if n < 0 {
		panic("negative pseudo-gate inputs")
	}
	g := &Mio_Gate_t{nInputs: n}
	for i := int32(0); i < n; i++ {
		g.pPins = &Mio_Pin_t{pNext: g.pPins}
	}
	return g
}
func Mio_LibraryShiftDelay(lib *Mio_Library_t, shift float64) {
	Mio_LibraryMatchesStop(lib)
	Mio_LibraryMatches2Stop(lib)
	for g := lib.pGates; g != nil; g = g.pNext {
		g.dDelayMax += shift
		for p := g.pPins; p != nil; p = p.pNext {
			p.dDelayBlockRise += shift
			p.dDelayBlockFall += shift
			p.dDelayBlockMax += shift
		}
	}
}
func Mio_LibraryMultiArea(lib *Mio_Library_t, multi float64) {
	Mio_LibraryMatchesStop(lib)
	Mio_LibraryMatches2Stop(lib)
	for g := lib.pGates; g != nil; g = g.pNext {
		if g.nInputs >= 2 {
			g.dArea *= math.Pow(float64(g.nInputs), multi)
		}
	}
}
func Mio_LibraryMultiDelay(lib *Mio_Library_t, multi float64) {
	Mio_LibraryMatchesStop(lib)
	Mio_LibraryMatches2Stop(lib)
	for g := lib.pGates; g != nil; g = g.pNext {
		if g.nInputs < 2 {
			continue
		}
		scale := math.Pow(float64(g.nInputs), multi)
		g.dDelayMax *= scale
		for p := g.pPins; p != nil; p = p.pNext {
			p.dDelayBlockRise *= scale
			p.dDelayBlockFall *= scale
			p.dDelayBlockMax *= scale
		}
	}
}
func Mio_LibraryTransferDelays(dst, src *Mio_Library_t) {
	Mio_LibraryMatchesStop(dst)
	Mio_LibraryMatches2Stop(dst)
	for s := src.pGates; s != nil; s = s.pNext {
		for d := dst.pGates; d != nil; d = d.pNext {
			if d.nInputs != s.nInputs || d.nInputs > 6 || d.uTruth != s.uTruth {
				continue
			}
			for p, q := d.pPins, s.pPins; p != nil; p = p.pNext {
				p.dDelayBlockRise, p.dDelayBlockFall, p.dDelayBlockMax = 0, 0, 0
				if q != nil {
					p.dDelayBlockRise, p.dDelayBlockFall, p.dDelayBlockMax = q.dDelayBlockRise, q.dDelayBlockFall, q.dDelayBlockMax
					q = q.pNext
				}
			}
			d.dDelayMax = 0
			for p := d.pPins; p != nil; p = p.pNext {
				d.dDelayMax = max(d.dDelayMax, p.dDelayBlockMax)
			}
		}
	}
}
func Mio_LibraryReadProfile(r io.Reader, lib *Mio_Library_t) error {
	scanner := bufio.NewScanner(r)
	staged := make(map[*Mio_Gate_t]int32)
	line := 0
	for scanner.Scan() {
		line++
		s, _, _ := strings.Cut(scanner.Text(), "#")
		fields := strings.Fields(s)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 2 {
			return fmt.Errorf("profile line %d: expected gate and count", line)
		}
		count, err := strconv.ParseInt(fields[1], 10, 32)
		if err != nil || count < 0 {
			return fmt.Errorf("profile line %d: invalid count", line)
		}
		g := Mio_LibraryReadGateByName(lib, fields[0], "")
		if g == nil {
			return fmt.Errorf("profile line %d: unknown gate %s", line, fields[0])
		}
		staged[g] = int32(count)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	for g, count := range staged {
		g.Profile = count
	}
	return nil
}
func Mio_LibraryWriteProfile(w io.Writer, lib *Mio_Library_t) error {
	for g := lib.pGates; g != nil; g = g.pNext {
		if g.Profile > 0 {
			if _, err := fmt.Fprintf(w, "%-24s %d\n", g.pName, g.Profile); err != nil {
				return err
			}
		}
	}
	return nil
}
func Mio_LibraryHasProfile(lib *Mio_Library_t) int32 {
	for g := lib.pGates; g != nil; g = g.pNext {
		if g.Profile > 0 {
			return 1
		}
	}
	return 0
}
func mioTransferProfile(dst, src *Mio_Library_t, second bool) {
	for g := dst.pGates; g != nil; g = g.pNext {
		if second {
			g.Profile2 = 0
		} else {
			g.Profile = 0
		}
	}
	for s := src.pGates; s != nil; s = s.pNext {
		value := s.Profile
		if second {
			value = s.Profile2
		}
		if value <= 0 {
			continue
		}
		d := Mio_LibraryReadGateByName(dst, s.pName, s.pOutName)
		if d == nil && s.nInputs <= 6 {
			d = Mio_LibraryReadGateByTruth(dst, s.uTruth)
		}
		if d != nil {
			if second {
				d.Profile2 += value
			} else {
				d.Profile += value
			}
		}
	}
}
func Mio_LibraryTransferProfile(dst, src *Mio_Library_t)  { mioTransferProfile(dst, src, false) }
func Mio_LibraryTransferProfile2(dst, src *Mio_Library_t) { mioTransferProfile(dst, src, true) }
func Mio_LibraryCleanProfile2(lib *Mio_Library_t) {
	for g := lib.pGates; g != nil; g = g.pNext {
		g.Profile2 = 0
	}
}
func Mio_LibraryHashGates(lib *Mio_Library_t) {
	lib.tName2Gate = make(map[string]*Mio_Gate_t)
	for g := lib.pGates; g != nil; g = g.pNext {
		if lib.tName2Gate[g.pName] == nil {
			lib.tName2Gate[g.pName] = g
		}
	}
}
func Mio_LibraryShortFormula(g *Mio_Gate_t, formula string) string {
	names := make(map[string]string)
	for i, p := 0, g.pPins; p != nil; i, p = i+1, p.pNext {
		names[p.pName] = string(rune('a' + i))
	}
	tokens, err := formulaTokens(formula)
	if err != nil {
		return formula
	}
	var output strings.Builder
	for _, t := range tokens {
		if replacement, ok := names[t.text]; ok {
			output.WriteString(replacement)
		} else {
			output.WriteString(t.text)
		}
	}
	return output.String()
}
func Mio_LibraryShortNames(lib *Mio_Library_t) {
	Mio_LibraryMatchesStop(lib)
	Mio_LibraryMatches2Stop(lib)
	digits := len(strconv.Itoa(int(lib.nGates)))
	for i, g := 0, lib.pGates; g != nil; i, g = i+1, g.pNext {
		g.pForm = Mio_LibraryShortFormula(g, g.pForm)
		g.pName = fmt.Sprintf("g%0*d", digits, i+1)
		for j, p := 0, g.pPins; p != nil; j, p = j+1, p.pNext {
			p.pName = string(rune('a' + j))
		}
		g.pOutName = "z"
	}
	lib.pName = fmt.Sprintf("lib%d", lib.nGates)
	Mio_LibraryHashGates(lib)
}
