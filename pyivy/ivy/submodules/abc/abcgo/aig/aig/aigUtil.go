package aig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Traversal IDs are bounded so stale node marks cannot alias a new traversal.
func Aig_ManIncrementTravId(p *Graph) {
	if p.travID >= (1<<30)-1 {
		// Preserve the generation that becomes Previous. Clearing every mark at
		// wrap would destroy seeds between the two steps of cone traversals.
		previous := p.travID
		for i := range p.nodes {
			if p.nodes[i].TravID == previous {
				p.nodes[i].TravID = 1
			} else {
				p.nodes[i].TravID = 0
			}
		}
		p.travID = 2
		return
	}
	p.travID++
}
func Aig_TimeStamp() string { return time.Now().Format("Mon Jan _2 15:04:05 2006") }
func Aig_ManHasNoGaps(p *Graph) int32 {
	if len(p.deleted) == 0 && int(Aig_ManObjNum(p)) == len(p.inputs)+len(p.outputs)+int(Aig_ManNodeNum(p))+1 {
		return 1
	}
	return 0
}
func Aig_ManLevels(p *Graph) int32 {
	level := 0
	for _, e := range p.outputs {
		level = max(level, p.node(p.outputDriver(e)).Level)
	}
	return int32(level)
}
func Aig_ManResetRefs(p *Graph) {
	for i := range p.nodes {
		p.nodes[i].Refs = 0
	}
	for i, n := range p.nodes {
		if p.deleted[i] {
			continue
		}
		for _, e := range nodeFanins(n) {
			p.node(e).Refs++
		}
	}
	for _, e := range p.outputs {
		p.node(e).Refs++
	}
}
func Aig_ManCleanMarkA(p *Graph) {
	for i := range p.nodes {
		p.nodes[i].MarkA = false
	}
}
func Aig_ManCleanMarkB(p *Graph) {
	for i := range p.nodes {
		p.nodes[i].MarkB = false
	}
}
func Aig_ManCleanMarkAB(p *Graph) { Aig_ManCleanMarkA(p); Aig_ManCleanMarkB(p) }
func Aig_ManCleanData(p *Graph)   { clear(p.objectData); clear(p.copies) }
func Aig_ManCleanNext(p *Graph)   { clear(p.next) }
func Aig_ObjCleanData_rec(p *Graph, root Edge) {
	for _, e := range p.dfs([]Edge{root}, false, false, false) {
		delete(p.objectData, e.NodeID())
		delete(p.copies, e.NodeID())
	}
}

// CollectMulti collects the unique frontier of an associative gate cone. A
// complemented edge or a different node kind stops expansion.
func Aig_ObjCollectMulti(p *Graph, root Edge) []Edge {
	result := []Edge{}
	Aig_ObjCollectMulti_rec(p, Aig_Regular(root), root, &result)
	return result
}
func Aig_ObjCollectMulti_rec(p *Graph, root, start Edge, result *[]Edge) {
	kind := p.node(root).Kind
	frontier := map[Edge]bool{}
	for _, e := range *result {
		frontier[e] = true
	}
	active, done := map[int]bool{}, map[int]bool{}
	var collect func(Edge)
	collect = func(e Edge) {
		if e != root && (e.Complemented() || p.node(e).Kind != kind) || kind != And && kind != ExclusiveOr {
			if !frontier[e] {
				*result = append(*result, e)
				frontier[e] = true
			}
			return
		}
		if active[e.NodeID()] {
			panic("aig: cycle in supergate")
		}
		if done[e.NodeID()] {
			return
		}
		active[e.NodeID()] = true
		for _, fanin := range nodeFanins(*p.node(e)) {
			collect(fanin)
		}
		delete(active, e.NodeID())
		done[e.NodeID()] = true
	}
	collect(start)
}
func Aig_ObjIsMuxType(p *Graph, e Edge) bool { _, _, _, ok := aigMux(p, e); return ok }
func aigMux(p *Graph, e Edge) (control, yes, no Edge, ok bool) {
	n := p.node(e)
	if n.Kind != And || !n.Fanins[0].Complemented() || !n.Fanins[1].Complemented() {
		return 0, 0, 0, false
	}
	a, b := p.node(n.Fanins[0]), p.node(n.Fanins[1])
	if a.Kind != And || b.Kind != And {
		return 0, 0, 0, false
	}
	for i, x := range a.Fanins {
		for j, y := range b.Fanins {
			if x != y.Not() {
				continue
			}
			yes, no = a.Fanins[1-i].Not(), b.Fanins[1-j].Not()
			control = x
			if x.Complemented() {
				control = y
				yes, no = no, yes
			}
			if e.Complemented() {
				yes, no = yes.Not(), no.Not()
			}
			return control, yes, no, true
		}
	}
	return 0, 0, 0, false
}
func Aig_ObjRecognizeMux(p *Graph, e Edge) (control, yes, no Edge) {
	control, yes, no, ok := aigMux(p, e)
	if !ok {
		panic("aig: node is not a multiplexer")
	}
	return
}
func Aig_ObjRecognizeExor(p *Graph, e Edge) (Edge, Edge, bool) {
	n := p.node(e)
	if n.Kind == ExclusiveOr {
		a := n.Fanins[0]
		if e.Complemented() {
			a = a.Not()
		}
		return a, n.Fanins[1], true
	}
	if n.Kind != And || !n.Fanins[0].Complemented() || !n.Fanins[1].Complemented() {
		return 0, 0, false
	}
	a, b := p.node(n.Fanins[0]), p.node(n.Fanins[1])
	if a.Kind != And || b.Kind != And {
		return 0, 0, false
	}
	if (a.Fanins[0] == b.Fanins[0].Not() && a.Fanins[1] == b.Fanins[1].Not()) || (a.Fanins[0] == b.Fanins[1].Not() && a.Fanins[1] == b.Fanins[0].Not()) {
		x := a.Fanins[0]
		if e.Complemented() {
			x = x.Not()
		}
		return x, a.Fanins[1], true
	}
	return 0, 0, false
}
func Aig_ObjReal_rec(p *Graph, e Edge) Edge {
	seen := map[int]bool{}
	for p.node(e).Kind == Buffer {
		if seen[e.NodeID()] {
			panic("aig: buffer cycle")
		}
		seen[e.NodeID()] = true
		next := p.node(e).Fanins[0]
		if e.Complemented() {
			next = next.Not()
		}
		e = next
	}
	return e
}
func Aig_ObjCompareIdIncrease(a, b Edge) int {
	if a.NodeID() < b.NodeID() {
		return -1
	}
	if a.NodeID() > b.NodeID() {
		return 1
	}
	return 0
}

// Expressions use typed node IDs and local builders; no temporary object data
// is written into the graph, and repeated XOR operands retain their parity.
func aigExpression(p *Graph, e Edge, verilog bool) string {
	n := p.node(e)
	neg := "!"
	and, or := " * ", " + "
	if verilog {
		neg = "~"
		and, or = " & ", " | "
	}
	switch n.Kind {
	case Constant:
		value := "1"
		if e.Complemented() {
			value = "0"
		}
		if verilog {
			return "1'b" + value
		}
		return value
	case Input:
		name, ok := p.objectData[e.NodeID()].(string)
		if !ok || name == "" {
			name = fmt.Sprintf("pi%d", Aig_ObjCioId(p, Aig_Regular(e)))
		}
		if e.Complemented() {
			return neg + name
		}
		return name
	case Buffer, Output:
		driver := n.Fanins[0]
		if e.Complemented() {
			driver = driver.Not()
		}
		return aigExpression(p, driver, verilog)
	case ExclusiveOr:
		text := "(" + aigExpression(p, n.Fanins[0], verilog) + " ^ " + aigExpression(p, n.Fanins[1], verilog) + ")"
		if e.Complemented() {
			text = neg + text
		}
		return text
	case And:
		if verilog {
			if c, t, f, ok := aigMux(p, e); ok {

				return "(" + aigExpression(p, c, true) + " ? " + aigExpression(p, t, true) + " : " + aigExpression(p, f, true) + ")"
			}
		}
		separator := and
		if e.Complemented() {
			separator = or
		}
		terms := []string{}
		for _, f := range Aig_ObjCollectMulti(p, Aig_Regular(e)) {
			if e.Complemented() {
				f = f.Not()
			}
			terms = append(terms, aigExpression(p, f, verilog))
		}
		return "(" + strings.Join(terms, separator) + ")"
	}
	panic("aig: unknown node kind")
}
func Aig_ObjPrintEqn(w io.Writer, p *Graph, e Edge) error {
	return aigWriteText(w, aigExpression(p, e, false))
}
func Aig_ObjPrintVerilog(w io.Writer, p *Graph, e Edge) error {
	return aigWriteText(w, aigExpression(p, e, true))
}
func Aig_ObjPrintVerbose(p *Graph, e Edge, history int32) {
	n := p.node(e)
	fmt.Printf("Node %d : %v", e.NodeID(), n.Kind)
	if fanins := nodeFanins(*n); len(fanins) > 0 {
		fmt.Print("(")
		for i, f := range fanins {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%d", f.NodeID())
			if f.Complemented() {
				fmt.Print("'")
			}
		}
		fmt.Print(")")
	}
	fmt.Printf(" (refs = %3d)", n.Refs)
}
func Aig_ObjPrintVerboseCone(p *Graph, root Edge, history int32) {
	for _, e := range p.dfs([]Edge{root}, false, false, false) {
		Aig_ObjPrintVerbose(p, e, history)
		fmt.Println()
	}
	fmt.Println()
}
func Aig_ManPrintVerbose(p *Graph, history int32) {
	fmt.Print("PIs:")
	for _, e := range p.inputs {
		fmt.Printf(" %d", e.NodeID())
	}
	fmt.Println()
	for _, e := range Aig_ManDfs(p, 1) {
		Aig_ObjPrintVerbose(p, e, history)
		fmt.Println()
	}
	fmt.Println()
}

var aigDumpCounter atomic.Uint64

func Aig_ManDump(p *Graph) error {
	if err := os.MkdirAll("aigbug", 0755); err != nil {
		return err
	}
	path := filepath.Join("aigbug", fmt.Sprintf("%03d.blif", aigDumpCounter.Add(1)))
	if err := Aig_ManDumpBlif(p, path, nil, nil); err != nil {
		return err
	}
	fmt.Printf("Intermediate AIG with %d nodes was written into file %q.\n", Aig_ManNodeNum(p), path)
	return nil
}

// All dump names are local snapshots rather than cached IDs in graph nodes.
func aigDumpNames(p *Graph, inputs, outputs []string) map[int]string {
	names := map[int]string{}
	for i := range p.nodes {
		names[i] = fmt.Sprintf("n%d", i)
	}
	if inputs != nil {
		if len(inputs) != len(p.inputs) {
			panic("aig: input name count")
		}
		for i, e := range p.inputs {
			names[e.NodeID()] = inputs[i]
		}
	}
	if outputs != nil {
		if len(outputs) != len(p.outputs) {
			panic("aig: output name count")
		}
		for i, e := range p.outputs {
			names[e.NodeID()] = outputs[i]
		}
	}
	return names
}
func aigModelName(p *Graph) string {
	if p.Name == "" {
		return "test"
	}
	return p.Name
}
func Aig_ManDumpBlif(p *Graph, path string, inputNames, outputNames []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = WriteBLIF(f, p, inputNames, outputNames)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func WriteBLIF(w io.Writer, p *Graph, inputNames, outputNames []string) error {
	if err := p.Check(); err != nil {
		return err
	}
	names := aigDumpNames(p, inputNames, outputNames)
	var b strings.Builder
	fmt.Fprintf(&b, "# BLIF file written by Aig_ManDumpBlif\n.model %s\n.inputs", aigModelName(p))
	for _, e := range p.inputs[:len(p.inputs)-int(p.nRegs)] {
		fmt.Fprintf(&b, " %s", names[e.NodeID()])
	}
	b.WriteString("\n.outputs")
	for _, e := range p.outputs[:len(p.outputs)-int(p.nRegs)] {
		fmt.Fprintf(&b, " %s", names[e.NodeID()])
	}
	b.WriteByte('\n')
	for i := 0; i < int(p.nRegs); i++ {
		li, lo := p.outputs[len(p.outputs)-int(p.nRegs)+i], p.inputs[len(p.inputs)-int(p.nRegs)+i]
		fmt.Fprintf(&b, ".latch %s %s 0\n", names[li.NodeID()], names[lo.NodeID()])
	}
	fmt.Fprintf(&b, ".names %s\n1\n", names[0])
	for _, e := range aigUtilOrder(p) {
		n := p.node(e)
		fanins := nodeFanins(*n)
		if len(fanins) == 0 {
			continue
		}
		b.WriteString(".names")
		for _, f := range fanins {
			fmt.Fprintf(&b, " %s", names[f.NodeID()])
		}
		fmt.Fprintf(&b, " %s\n", names[e.NodeID()])
		if len(fanins) == 1 {
			bit := 1
			if fanins[0].Complemented() {
				bit = 0
			}
			fmt.Fprintf(&b, "%d 1\n", bit)
		} else {
			for a := 0; a < 2; a++ {
				for c := 0; c < 2; c++ {
					av, bv := (a != 0) != fanins[0].Complemented(), (c != 0) != fanins[1].Complemented()
					value := av && bv
					if n.Kind == ExclusiveOr {
						value = av != bv
					}
					if value {
						fmt.Fprintf(&b, "%d%d 1\n", a, c)
					}
				}
			}
		}
	}
	b.WriteString(".end\n")
	return aigWriteText(w, b.String())
}
func Aig_ManDumpVerilog(p *Graph, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = WriteVerilog(f, p)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func WriteVerilog(w io.Writer, p *Graph) error {
	if err := p.Check(); err != nil {
		return err
	}
	names := aigDumpNames(p, nil, nil)
	var b strings.Builder
	ports := []string{}
	if p.nRegs > 0 {
		ports = append(ports, "clock")
	}
	for _, e := range p.inputs[:len(p.inputs)-int(p.nRegs)] {
		ports = append(ports, names[e.NodeID()])
	}
	for _, e := range p.outputs[:len(p.outputs)-int(p.nRegs)] {
		ports = append(ports, names[e.NodeID()])
	}
	fmt.Fprintf(&b, "// Verilog file written by Aig_ManDumpVerilog\nmodule %s (%s);\n", aigModelName(p), strings.Join(ports, ", "))
	if p.nRegs > 0 {
		b.WriteString("input clock;\n")
	}
	for i, e := range p.inputs {
		kind := "input"
		if i >= len(p.inputs)-int(p.nRegs) {
			kind = "reg"
		}
		fmt.Fprintf(&b, "%s %s;\n", kind, names[e.NodeID()])
	}
	for i, e := range p.outputs {
		kind := "output"
		if i >= len(p.outputs)-int(p.nRegs) {
			kind = "wire"
		}
		fmt.Fprintf(&b, "%s %s;\n", kind, names[e.NodeID()])
	}
	fmt.Fprintf(&b, "wire %s;\nassign %s = 1'b1;\n", names[0], names[0])
	edgeName := func(e Edge) string {
		s := names[e.NodeID()]
		if e.Complemented() {
			s = "~" + s
		}
		return s
	}
	for _, e := range aigUtilOrder(p) {
		n := p.node(e)
		fanins := nodeFanins(*n)
		if len(fanins) == 0 {
			continue
		}
		if n.Kind != Output {
			fmt.Fprintf(&b, "wire %s;\n", names[e.NodeID()])
		}
		text := edgeName(fanins[0])
		if len(fanins) == 2 {
			op := " & "
			if n.Kind == ExclusiveOr {
				op = " ^ "
			}
			text += op + edgeName(fanins[1])
		}
		fmt.Fprintf(&b, "assign %s = %s;\n", names[e.NodeID()], text)
	}
	for i := 0; i < int(p.nRegs); i++ {
		li, lo := p.outputs[len(p.outputs)-int(p.nRegs)+i], p.inputs[len(p.inputs)-int(p.nRegs)+i]
		fmt.Fprintf(&b, "always @(posedge clock) %s <= %s;\ninitial %s = 1'b0;\n", names[lo.NodeID()], names[li.NodeID()], names[lo.NodeID()])
	}
	b.WriteString("endmodule\n")
	return aigWriteText(w, b.String())
}
func Aig_ManSetCioIds(p *Graph) map[Edge]int32 {
	ids := map[Edge]int32{}
	for i, e := range p.inputs {
		p.node(e).CioID = int32(i)
		ids[e] = int32(i)
	}
	for i, e := range p.outputs {
		p.node(e).CioID = int32(i)
		ids[e] = int32(i)
	}
	return ids
}

// Aig_ManCleanCioIds clears the terminal pNext/CioId union, as in ABC.
func Aig_ManCleanCioIds(p *Graph) {
	for _, e := range p.inputs {
		p.node(e).CioID = 0
	}
	for _, e := range p.outputs {
		p.node(e).CioID = 0
	}
}
func Aig_ManChoiceNum(p *Graph) int32 {
	var count int32
	for i, n := range p.nodes {
		if !p.deleted[i] && (n.Kind == And || n.Kind == ExclusiveOr) && Aig_ObjIsChoice(p, Edge(i)<<1) {
			count++
		}
	}
	return count
}
func Aig_ManPrintControlFanouts(p *Graph) {
	if len(p.inputs) == 0 {
		return
	}
	control := p.inputs[len(p.inputs)-1]
	fmt.Println("Control signal:")
	Aig_ObjPrintVerbose(p, control, 0)
	fmt.Print("\n\n")
	for i, n := range p.nodes {
		if p.deleted[i] || len(nodeFanins(n)) != 2 {
			continue
		}
		for j, f := range n.Fanins {
			other := n.Fanins[1-j]
			if Aig_Regular(f) == control && p.node(other).Kind == Input {
				Aig_ObjPrintVerbose(p, Edge(i)<<1, 0)
				fmt.Println()
				Aig_ObjPrintVerbose(p, Aig_Regular(other), 0)
				fmt.Print("\n\n")
			}
		}
	}
}
func Aig_FileNameGenericAppend(base, suffix string) string {
	base = strings.ReplaceAll(base, "\\", "/")
	base = filepath.Base(base)
	if dot := strings.LastIndexByte(base, '.'); dot >= 0 {
		base = base[:dot]
	}
	return base + suffix
}

// AigRandomSource is ABC's 32-bit multiply-with-carry stream, with explicit
// state for independent reproducible consumers. Compatibility functions lock
// the shared stream so a 64-bit draw cannot interleave its two halves.
type AigRandomSource struct{ z, w uint32 }

func NewAigRandomSource() *AigRandomSource { return &AigRandomSource{z: 3716960521, w: 2174103536} }
func (r *AigRandomSource) Uint32() uint32 {
	if *r == (AigRandomSource{}) {
		*r = *NewAigRandomSource()
	}
	r.z = 36969*(r.z&65535) + (r.z >> 16)
	r.w = 18000*(r.w&65535) + (r.w >> 16)
	return r.z<<16 + r.w
}

var aigRandom struct {
	sync.Mutex
	source AigRandomSource
}

func aigRandomDraw(reset bool) uint32 {
	if reset || aigRandom.source == (AigRandomSource{}) {
		aigRandom.source = *NewAigRandomSource()
	}
	return aigRandom.source.Uint32()
}
func Aig_ManRandom(reset int32) uint32 {
	aigRandom.Lock()
	defer aigRandom.Unlock()
	return aigRandomDraw(reset != 0)
}
func Aig_ManRandom64(reset int32) uint64 {
	aigRandom.Lock()
	defer aigRandom.Unlock()
	lo := aigRandomDraw(reset != 0)
	return uint64(lo) | uint64(aigRandomDraw(false))<<32
}
func Aig_ManRandomTest1() error {
	f, err := os.Create("rand.txt")
	if err != nil {
		return err
	}
	for i := 1; i <= 20000; i++ {
		v := Aig_ManRandom(0)
		if _, err := fmt.Fprintf(f, "%10d : %10d \n", i, v); err != nil {
			f.Close()
			return err
		}
		if v == 1 {
			break
		}
	}
	return f.Close()
}
func Aig_ManRandomTest2() error {
	return os.WriteFile("rand.txt", []byte(fmt.Sprintf("%10d : %10d \n", 1, 1)), 0644)
}
func Aig_ManRandomInfo(info *vec.Vec_Ptr_t, inputStart, wordStart, wordStop int) {
	for _, entry := range vec.Vec_PtrArray(info)[inputStart:] {
		words := entry.([]uint32)
		for i := wordStart; i < wordStop; i++ {
			words[i] = Aig_ManRandom(0)
		}
	}
}
func Aig_NodeUnionLists(a, b []Edge) []Edge     { return aigMergeLists(a, b, false) }
func Aig_NodeIntersectLists(a, b []Edge) []Edge { return aigMergeLists(a, b, true) }
func aigMergeLists(a, b []Edge, intersection bool) []Edge {
	result := make([]Edge, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i].NodeID() == b[j].NodeID() {
			result = append(result, a[i])
			i++
			j++
		} else if a[i].NodeID() < b[j].NodeID() {
			if !intersection {
				result = append(result, a[i])
			}
			i++
		} else {
			if !intersection {
				result = append(result, b[j])
			}
			j++
		}
	}
	if !intersection {
		result = append(result, a[i:]...)
		result = append(result, b[j:]...)
	}
	return result
}

func aigUtilOrder(p *Graph) []Edge {
	roots := make([]Edge, 0, len(p.nodes))
	for i := range p.nodes {
		if !p.deleted[i] {
			roots = append(roots, Edge(i)<<1)
		}
	}
	return p.dfs(roots, false, false, false)
}
func aigSimulate(p *Graph, order []Edge, inputs []bool) []bool {
	if len(inputs) != len(p.inputs) {
		panic("aig: simulation input count")
	}
	values := make([]bool, len(p.nodes))
	values[0] = true
	for i, e := range p.inputs {
		values[e.NodeID()] = inputs[i]
	}
	for _, e := range order {
		n := p.node(e)
		switch n.Kind {
		case And:
			values[e.NodeID()] = aigUtilEdgeValue(n.Fanins[0], values) && aigUtilEdgeValue(n.Fanins[1], values)
		case ExclusiveOr:
			values[e.NodeID()] = aigUtilEdgeValue(n.Fanins[0], values) != aigUtilEdgeValue(n.Fanins[1], values)
		case Buffer, Output:
			values[e.NodeID()] = aigUtilEdgeValue(n.Fanins[0], values)
		}
	}
	return values
}
func Aig_ManCounterExampleValueStart(p *Graph, cex *util.Abc_Cex_t) {
	if p.nRegs <= 0 || p.cexValues != nil {
		panic("aig: invalid counterexample session")
	}
	pis := len(p.inputs) - int(p.nRegs)
	if cex.NRegs != p.nRegs || int(cex.NPis) != pis || cex.IFrame < 0 || cex.IPo < 0 || int(cex.IPo) >= len(p.outputs)-int(p.nRegs) {
		panic("aig: counterexample dimensions")
	}
	bits := int(p.nRegs) + (int(cex.IFrame)+1)*pis
	if int(cex.NBits) != bits || len(cex.PData)*32 < bits {
		panic("aig: counterexample bit count")
	}
	bit := func(i int) bool { return cex.PData[i/32]>>(uint(i)&31)&1 != 0 }
	for i := 0; i < int(p.nRegs); i++ {
		if bit(i) {
			panic("aig: nonzero initial register state")
		}
	}
	order := aigUtilOrder(p)
	frames := make([][]bool, int(cex.IFrame)+1)
	state := make([]bool, int(p.nRegs))
	position := int(p.nRegs)
	for frame := range frames {
		inputs := make([]bool, len(p.inputs))
		for i := 0; i < pis; i++ {
			inputs[i] = bit(position)
			position++
		}
		copy(inputs[pis:], state)
		values := aigSimulate(p, order, inputs)
		frames[frame] = values
		for i := range state {
			state[i] = aigUtilEdgeValue(p.outputs[len(p.outputs)-int(p.nRegs)+i], values)
		}
	}
	if !aigUtilEdgeValue(p.outputs[cex.IPo], frames[len(frames)-1]) {
		panic("aig: counterexample does not assert its output")
	}
	p.cexValues = frames
}
func Aig_ManCounterExampleValueStop(p *Graph) {
	if p.cexValues == nil {
		panic("aig: no counterexample session")
	}
	p.cexValues = nil
}
func Aig_ManCounterExampleValueLookup(p *Graph, id, frame int32) int32 {
	if id < 0 || int(id) >= len(p.nodes) || p.deleted[int(id)] || frame < 0 || int(frame) >= len(p.cexValues) {
		panic("aig: counterexample lookup bounds")
	}
	if p.cexValues[frame][id] {
		return 1
	}
	return 0
}
func Aig_ManCounterExampleValueTest(p *Graph, cex *util.Abc_Cex_t) {
	fmt.Printf("\nUsing counter-example, which asserts output %d in frame %d.\n", cex.IPo, cex.IFrame)
	Aig_ManCounterExampleValueStart(p, cex)
	defer Aig_ManCounterExampleValueStop(p)
	id := int32(len(p.nodes) / 2)
	frame := max(0, cex.IFrame-1)
	fmt.Printf("Value of object %d in frame %d is %d.\n", id, frame, Aig_ManCounterExampleValueLookup(p, id, frame))
}
func Aig_ManSetPhase(p *Graph) {
	values := aigSimulate(p, aigUtilOrder(p), make([]bool, len(p.inputs)))
	for i := range p.nodes {
		if !p.deleted[i] {
			p.nodes[i].Phase = values[i]
		}
	}
}
func Aig_ManMuxesCollect(p *Graph) []Edge {
	result := []Edge{}
	for i, n := range p.nodes {
		if !p.deleted[i] && n.Kind == And && Aig_ObjIsMuxType(p, Edge(i)<<1) {
			result = append(result, Edge(i)<<1)
		}
	}
	return result
}
func aigMuxRefs(p *Graph, muxes []Edge, delta int) {
	for _, e := range muxes {
		a, b, xor := Aig_ObjRecognizeExor(p, e)
		if xor {
			p.node(a).Refs += delta
			p.node(b).Refs += delta
		} else {
			c, _, _ := Aig_ObjRecognizeMux(p, e)
			p.node(c).Refs += delta
		}
	}
}
func Aig_ManMuxesDeref(p *Graph, muxes []Edge) { aigMuxRefs(p, muxes, -1) }
func Aig_ManMuxesRef(p *Graph, muxes []Edge)   { aigMuxRefs(p, muxes, 1) }
func Aig_ManInvertConstraints(p *Graph) {
	count := int(p.nConstrs)
	primaryOutputs := len(p.outputs) - int(p.nRegs)
	if count < 0 || count > primaryOutputs {
		panic("aig: constraint count")
	}
	for i := primaryOutputs - count; i < primaryOutputs; i++ {
		e := p.outputs[i]
		if p.node(e).Kind == Output {
			p.node(e).Fanins[0] = p.node(e).Fanins[0].Not()
			p.node(e).Phase = !p.node(e).Phase
		} else {
			p.outputs[i] = e.Not()
		}
	}
	p.refreshFanouts()
}

func aigUtilEdgeValue(e Edge, values []bool) bool { return values[e.NodeID()] != e.Complemented() }

func aigWriteText(w io.Writer, text string) error {
	n, err := io.WriteString(w, text)
	if err == nil && n != len(text) {
		return io.ErrShortWrite
	}
	return err
}
