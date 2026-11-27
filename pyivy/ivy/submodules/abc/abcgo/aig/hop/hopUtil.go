// Graph recognition and output, ported from ABC hopUtil.c (Alan Mishchenko).
package hop

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func Hop_ManIncrementTravId(p *Hop_Man_t) {
	if p.nTravIds >= (1<<30)-1 {
		Hop_ManCleanData(p)
	}
	p.nTravIds++
}
func Hop_ManCleanData(p *Hop_Man_t) {
	p.nTravIds = 1
	clean := func(o *Hop_Obj_t) { o.pData = nil; o.iData = 0; o.travId = 0 }
	clean(p.pConst1)
	for _, v := range Vec_PtrArray(p.vPis) {
		clean(v.(*Hop_Obj_t))
	}
	for _, v := range Vec_PtrArray(p.vPos) {
		clean(v.(*Hop_Obj_t))
	}
	for _, head := range p.pTable {
		for o := head; o != nil; o = o.pNext {
			clean(o)
		}
	}
}
func Hop_ObjCleanData_rec(o *Hop_Obj_t) {
	o = Hop_Regular(o)
	if Hop_ObjIsNode(o) != 0 {
		Hop_ObjCleanData_rec(Hop_ObjFanin0(o))
		Hop_ObjCleanData_rec(Hop_ObjFanin1(o))
	}
	o.pData = nil
	o.iData = 0
	o.travId = 0
}
func Hop_ObjCollectMulti_rec(root, o *Hop_Obj_t, v *Vec_Ptr_t) {
	if root != o && (Hop_IsComplement(o) != 0 || Hop_ObjIsNode(Hop_Regular(o)) == 0 || root.Type != o.Type) {
		for _, old := range Vec_PtrArray(v) {
			if old == o {
				return
			}
		}
		Vec_PtrPush(v, o)
		return
	}
	Hop_ObjCollectMulti_rec(root, o.pFanin0, v)
	Hop_ObjCollectMulti_rec(root, o.pFanin1, v)
}
func Hop_ObjCollectMulti(root *Hop_Obj_t, v *Vec_Ptr_t) {
	Vec_PtrClear(v)
	if Hop_ObjIsNode(root) == 0 {
		Vec_PtrPush(v, root)
		return
	}
	Hop_ObjCollectMulti_rec(root, root, v)
}
func Hop_ObjIsMuxType(o *Hop_Obj_t) int32 {
	if Hop_ObjIsAnd(o) == 0 || Hop_ObjFaninC0(o) == 0 || Hop_ObjFaninC1(o) == 0 {
		return 0
	}
	n0, n1 := Hop_ObjFanin0(o), Hop_ObjFanin1(o)
	if Hop_ObjIsAnd(n0) == 0 || Hop_ObjIsAnd(n1) == 0 {
		return 0
	}
	return hopBool((Hop_ObjFanin0(n0) == Hop_ObjFanin0(n1) && Hop_ObjFaninC0(n0)^Hop_ObjFaninC0(n1) != 0) ||
		(Hop_ObjFanin0(n0) == Hop_ObjFanin1(n1) && Hop_ObjFaninC0(n0)^Hop_ObjFaninC1(n1) != 0) ||
		(Hop_ObjFanin1(n0) == Hop_ObjFanin0(n1) && Hop_ObjFaninC1(n0)^Hop_ObjFaninC0(n1) != 0) ||
		(Hop_ObjFanin1(n0) == Hop_ObjFanin1(n1) && Hop_ObjFaninC1(n0)^Hop_ObjFaninC1(n1) != 0))
}

// Hop_ObjRecognizeMux returns the (never complemented) control of a MUX and
// sets T and E, checking the fanin pairs in hopUtil.c's order.
func Hop_ObjRecognizeMux(o *Hop_Obj_t, t, e **Hop_Obj_t) *Hop_Obj_t {
	if Hop_ObjIsMuxType(o) == 0 {
		panic("not a mux")
	}
	n0, n1 := Hop_ObjFanin0(o), Hop_ObjFanin1(o)
	switch {
	case Hop_ObjFanin1(n0) == Hop_ObjFanin1(n1) && Hop_ObjFaninC1(n0)^Hop_ObjFaninC1(n1) != 0:
		if Hop_ObjFaninC1(n0) != 0 {
			*t, *e = Hop_Not(Hop_ObjChild0(n1)), Hop_Not(Hop_ObjChild0(n0))
			return Hop_ObjChild1(n1)
		}
		*t, *e = Hop_Not(Hop_ObjChild0(n0)), Hop_Not(Hop_ObjChild0(n1))
		return Hop_ObjChild1(n0)
	case Hop_ObjFanin0(n0) == Hop_ObjFanin0(n1) && Hop_ObjFaninC0(n0)^Hop_ObjFaninC0(n1) != 0:
		if Hop_ObjFaninC0(n0) != 0 {
			*t, *e = Hop_Not(Hop_ObjChild1(n1)), Hop_Not(Hop_ObjChild1(n0))
			return Hop_ObjChild0(n1)
		}
		*t, *e = Hop_Not(Hop_ObjChild1(n0)), Hop_Not(Hop_ObjChild1(n1))
		return Hop_ObjChild0(n0)
	case Hop_ObjFanin0(n0) == Hop_ObjFanin1(n1) && Hop_ObjFaninC0(n0)^Hop_ObjFaninC1(n1) != 0:
		if Hop_ObjFaninC0(n0) != 0 {
			*t, *e = Hop_Not(Hop_ObjChild0(n1)), Hop_Not(Hop_ObjChild1(n0))
			return Hop_ObjChild1(n1)
		}
		*t, *e = Hop_Not(Hop_ObjChild1(n0)), Hop_Not(Hop_ObjChild0(n1))
		return Hop_ObjChild0(n0)
	case Hop_ObjFanin1(n0) == Hop_ObjFanin0(n1) && Hop_ObjFaninC1(n0)^Hop_ObjFaninC0(n1) != 0:
		if Hop_ObjFaninC1(n0) != 0 {
			*t, *e = Hop_Not(Hop_ObjChild1(n1)), Hop_Not(Hop_ObjChild0(n0))
			return Hop_ObjChild0(n1)
		}
		*t, *e = Hop_Not(Hop_ObjChild0(n0)), Hop_Not(Hop_ObjChild1(n1))
		return Hop_ObjChild1(n0)
	}
	panic("not a mux")
}
func Hop_ObjRecognizeExor(o *Hop_Obj_t, a, b **Hop_Obj_t) int32 {
	if o.Type == AIG_EXOR {
		*a, *b = o.pFanin0, o.pFanin1
		return 1
	}
	if o.Type != AIG_AND || Hop_ObjFaninC0(o) == 0 || Hop_ObjFaninC1(o) == 0 {
		return 0
	}
	x, y := Hop_ObjFanin0(o), Hop_ObjFanin1(o)
	if x.Type != AIG_AND || y.Type != AIG_AND || Hop_ObjFanin0(x) != Hop_ObjFanin0(y) || Hop_ObjFanin1(x) != Hop_ObjFanin1(y) || Hop_ObjFaninC0(x) == Hop_ObjFaninC0(y) || Hop_ObjFaninC1(x) == Hop_ObjFaninC1(y) {
		return 0
	}
	*a, *b = x.pFanin0, x.pFanin1
	return 1
}
func hopFormula(edge *Hop_Obj_t, verilog bool) string {
	o := Hop_Regular(edge)
	neg := Hop_IsComplement(edge) != 0
	var value string
	switch o.Type {
	case AIG_CONST1:
		if verilog {
			if neg {
				return "1'b0"
			}
			return "1'b1"
		}
		if neg {
			return "0"
		}
		return "1"
	case AIG_PI:
		value, _ = o.pData.(string)
		if value == "" {
			value = fmt.Sprintf("x%d", o.PioNum)
		}
	case AIG_PO:
		return hopFormula(Hop_NotCond(o.pFanin0, hopBool(neg)), verilog)
	default:
		op := " * "
		if verilog {
			op = " & "
		}
		if o.Type == AIG_EXOR {
			op = " ^ "
		}
		value = "(" + hopFormula(o.pFanin0, verilog) + op + hopFormula(o.pFanin1, verilog) + ")"
	}
	if neg {
		if verilog {
			return "~" + value
		}
		return "!" + value
	}
	return value
}
func hopPrintLevel(levels *Vec_Vec_t, level int32) *Vec_Ptr_t {
	Vec_VecExpand(levels, level)
	return Vec_VecEntry(levels, level)
}
func hopPiName(o *Hop_Obj_t) string {
	name, _ := o.pData.(string)
	return name
}

// Hop_ObjPrintEqn prints the formula rooted at o (hopUtil.c); PIs carry
// their names in pData.
func Hop_ObjPrintEqn(w io.Writer, o *Hop_Obj_t, levels *Vec_Vec_t, level int32) {
	if levels == nil {
		levels = Vec_VecAlloc(10)
	}
	compl := Hop_IsComplement(o) != 0
	o = Hop_Regular(o)
	if Hop_ObjIsConst1(o) != 0 {
		fmt.Fprintf(w, "%d", hopBool(!compl))
		return
	}
	if Hop_ObjIsPi(o) != 0 {
		if compl {
			fmt.Fprint(w, "!")
		}
		fmt.Fprint(w, hopPiName(o))
		return
	}
	super := hopPrintLevel(levels, level)
	Hop_ObjCollectMulti(o, super)
	if level != 0 {
		fmt.Fprint(w, "(")
	}
	op := " * "
	if compl {
		op = " + "
	}
	for i, n := int32(0), Vec_PtrSize(super); i < n; i++ {
		Hop_ObjPrintEqn(w, Hop_NotCond(Vec_PtrEntry(super, i).(*Hop_Obj_t), hopBool(compl)), levels, level+1)
		if i < n-1 {
			fmt.Fprint(w, op)
		}
	}
	if level != 0 {
		fmt.Fprint(w, ")")
	}
}

// Hop_ObjPrintVerilog prints the Verilog formula rooted at o (hopUtil.c).
func Hop_ObjPrintVerilog(w io.Writer, o *Hop_Obj_t, levels *Vec_Vec_t, level, onlyAnds int32) {
	if levels == nil {
		levels = Vec_VecAlloc(10)
	}
	compl := Hop_IsComplement(o) != 0
	o = Hop_Regular(o)
	if Hop_ObjIsConst1(o) != 0 {
		fmt.Fprintf(w, "1'b%d", hopBool(!compl))
		return
	}
	if Hop_ObjIsPi(o) != 0 {
		if compl {
			fmt.Fprint(w, "~")
		}
		fmt.Fprint(w, hopPiName(o))
		return
	}
	open, closeParen := "", ""
	if level != 0 {
		open, closeParen = "(", ")"
	}
	if onlyAnds == 0 && Hop_ObjIsExor(o) != 0 {
		super := hopPrintLevel(levels, level)
		Hop_ObjCollectMulti(o, super)
		fmt.Fprint(w, open)
		for i, n := int32(0), Vec_PtrSize(super); i < n; i++ {
			Hop_ObjPrintVerilog(w, Hop_NotCond(Vec_PtrEntry(super, i).(*Hop_Obj_t), hopBool(compl && i == 0)), levels, level+1, onlyAnds)
			if i < n-1 {
				fmt.Fprint(w, " ^ ")
			}
		}
		fmt.Fprint(w, closeParen)
		return
	}
	if onlyAnds == 0 && Hop_ObjIsMuxType(o) != 0 {
		var f0, f1 *Hop_Obj_t
		if Hop_ObjRecognizeExor(o, &f0, &f1) != 0 {
			fmt.Fprint(w, open)
			Hop_ObjPrintVerilog(w, Hop_NotCond(f0, hopBool(compl)), levels, level+1, onlyAnds)
			fmt.Fprint(w, " ^ ")
			Hop_ObjPrintVerilog(w, f1, levels, level+1, onlyAnds)
			fmt.Fprint(w, closeParen)
		} else {
			c := Hop_ObjRecognizeMux(o, &f1, &f0)
			fmt.Fprint(w, open)
			Hop_ObjPrintVerilog(w, c, levels, level+1, onlyAnds)
			fmt.Fprint(w, " ? ")
			Hop_ObjPrintVerilog(w, Hop_NotCond(f1, hopBool(compl)), levels, level+1, onlyAnds)
			fmt.Fprint(w, " : ")
			Hop_ObjPrintVerilog(w, Hop_NotCond(f0, hopBool(compl)), levels, level+1, onlyAnds)
			fmt.Fprint(w, closeParen)
		}
		return
	}
	super := hopPrintLevel(levels, level)
	Hop_ObjCollectMulti(o, super)
	fmt.Fprint(w, open)
	op := " & "
	if compl {
		op = " | "
	}
	for i, n := int32(0), Vec_PtrSize(super); i < n; i++ {
		Hop_ObjPrintVerilog(w, Hop_NotCond(Vec_PtrEntry(super, i).(*Hop_Obj_t), hopBool(compl)), levels, level+1, onlyAnds)
		if i < n-1 {
			fmt.Fprint(w, op)
		}
	}
	fmt.Fprint(w, closeParen)
}
func Hop_ObjPrintVerbose(o *Hop_Obj_t, haig int32) {
	fmt.Printf("Node %p : type %d (refs = %3d)", o, o.Type, o.nRefs)
}
func Hop_ManPrintVerbose(p *Hop_Man_t, haig int32) {
	fmt.Print("PIs:")
	for _, v := range Vec_PtrArray(p.vPis) {
		fmt.Printf(" %p", v)
	}
	fmt.Println()
	nodes := Hop_ManDfs(p)
	for _, v := range Vec_PtrArray(nodes) {
		Hop_ObjPrintVerbose(v.(*Hop_Obj_t), haig)
		fmt.Println()
	}
}
func WriteBlif(p *Hop_Man_t, w io.Writer) error {
	var b strings.Builder
	b.WriteString("# BLIF generated by ABC Go Hop\n.model test\n.inputs")
	for _, v := range Vec_PtrArray(p.vPis) {
		fmt.Fprintf(&b, " n%d", v.(*Hop_Obj_t).Id)
	}
	b.WriteString("\n.outputs")
	for _, v := range Vec_PtrArray(p.vPos) {
		fmt.Fprintf(&b, " n%d", v.(*Hop_Obj_t).Id)
	}
	b.WriteByte('\n')
	fmt.Fprintf(&b, ".names n%d\n1\n", p.pConst1.Id)
	nodes := Hop_ManDfs(p)
	for _, v := range Vec_PtrArray(nodes) {
		o := v.(*Hop_Obj_t)
		fmt.Fprintf(&b, ".names n%d n%d n%d\n", Hop_ObjFanin0(o).Id, Hop_ObjFanin1(o).Id, o.Id)
		if o.Type == AIG_EXOR {
			for x := int32(0); x < 2; x++ {
				fmt.Fprintf(&b, "%d%d 1\n", x^Hop_ObjFaninC0(o), (x^1)^Hop_ObjFaninC1(o))
			}
		} else {
			fmt.Fprintf(&b, "%d%d 1\n", 1^Hop_ObjFaninC0(o), 1^Hop_ObjFaninC1(o))
		}
	}
	for _, v := range Vec_PtrArray(p.vPos) {
		o := v.(*Hop_Obj_t)
		fmt.Fprintf(&b, ".names n%d n%d\n%d 1\n", Hop_ObjFanin0(o).Id, o.Id, 1^Hop_ObjFaninC0(o))
	}
	b.WriteString(".end\n")
	_, err := io.WriteString(w, b.String())
	return err
}
func Hop_ManDumpBlif(p *Hop_Man_t, name string) {
	if Hop_ManPoNum(p) == 0 {
		fmt.Println("Hop_ManDumpBlif(): AIG manager does not have POs.")
		return
	}
	f, err := os.Create(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	err = WriteBlif(p, f)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
