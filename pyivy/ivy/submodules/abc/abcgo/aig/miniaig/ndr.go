// Port of ABC src/aig/miniaig/ndr.h (Alan Mishchenko, UC Berkeley).
package miniaig

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	NDR_NONE = iota
	NDR_DESIGN
	NDR_MODULE
	NDR_OBJECT
	NDR_INPUT
	NDR_OUTPUT
	NDR_OPERTYPE
	NDR_NAME
	NDR_RANGE
	NDR_FUNCTION
	NDR_TARGET
	NDR_UNKNOWN
)

type Ndr_RecordType_t = int
type Ndr_Data_t struct {
	nSize, nCap int
	pHead       []byte
	pBody       []int
}

func Ndr_DataType(p *Ndr_Data_t, i int) int { return int(p.pHead[i]) }
func Ndr_DataSize(p *Ndr_Data_t, i int) int {
	if Ndr_DataType(p, i) > NDR_OBJECT {
		return 1
	}
	return p.pBody[i]
}
func Ndr_DataEntry(p *Ndr_Data_t, i int) int    { return p.pBody[i] }
func Ndr_DataEntryP(p *Ndr_Data_t, i int) []int { return p.pBody[i:] }
func Ndr_DataEnd(p *Ndr_Data_t, i int) int      { return i + p.pBody[i] }
func Ndr_DataAddTo(p *Ndr_Data_t, i, add int) {
	if Ndr_DataType(p, i) > NDR_OBJECT {
		panic("not a container")
	}
	p.pBody[i] += add
}
func Ndr_DataResize(p *Ndr_Data_t, add int) {
	if p.nSize+add <= p.nCap {
		return
	}
	n := 2 * p.nCap
	if n < p.nSize+add {
		n = p.nSize + add
	}
	h := make([]byte, len(p.pHead), n)
	copy(h, p.pHead)
	b := make([]int, len(p.pBody), n)
	copy(b, p.pBody)
	p.pHead = h
	p.pBody = b
	p.nCap = n
}
func Ndr_DataPush(p *Ndr_Data_t, typ, entry int) {
	Ndr_DataResize(p, 1)
	p.pHead = append(p.pHead, byte(typ))
	p.pBody = append(p.pBody, entry)
	p.nSize++
}
func Ndr_DataPushRange(p *Ndr_Data_t, left, right, signed int) {
	if signed != 0 {
		Ndr_DataPushArray(p, NDR_RANGE, 3, []int{left, right, signed})
	} else if left != 0 || right != 0 {
		if left == right {
			Ndr_DataPush(p, NDR_RANGE, left)
		} else {
			Ndr_DataPushArray(p, NDR_RANGE, 2, []int{left, right})
		}
	}
}
func Ndr_DataPushArray(p *Ndr_Data_t, typ, n int, a []int) {
	if n < 0 || n > len(a) {
		panic("invalid record array length")
	}
	for _, v := range a[:n] {
		Ndr_DataPush(p, typ, v)
	}
}

// LUT functions carry their 64-bit truth value directly, rather than casting
// pointer bits to a string as the C representation does.
func Ndr_DataPushString(p *Ndr_Data_t, obj, typ int, function any) {
	if function == nil {
		return
	}
	if obj == ABC_OPER_LUT {
		var v uint64
		switch x := function.(type) {
		case uint64:
			v = x
		case int:
			v = uint64(x)
		default:
			panic("LUT needs uint64 truth")
		}
		Ndr_DataPushArray(p, typ, 2, []int{int(uint32(v)), int(uint32(v >> 32))})
		return
	}
	s, ok := function.(string)
	if !ok {
		panic("function needs string")
	}
	data := make([]byte, 4*((len(s)+4)/4))
	copy(data, s)
	for i := 0; i < len(data); i += 4 {
		Ndr_DataPush(p, typ, int(binary.LittleEndian.Uint32(data[i:])))
	}
}
func Ndr_ObjReadEntry(p *Ndr_Data_t, obj, typ int) int {
	for i := obj + 1; i < Ndr_DataEnd(p, obj); i += Ndr_DataSize(p, i) {
		if Ndr_DataType(p, i) == typ {
			return p.pBody[i]
		}
	}
	return -1
}
func Ndr_ObjReadArray(p *Ndr_Data_t, obj, typ int, out *[]int) int {
	*out = nil
	start, end := -1, -1
	for i := obj + 1; i < Ndr_DataEnd(p, obj); i += Ndr_DataSize(p, i) {
		if Ndr_DataType(p, i) == typ {
			if start < 0 {
				start = i
			}
			end = i + 1
		} else if start >= 0 {
			break
		}
	}
	if start >= 0 {
		*out = p.pBody[start:end]
	}
	return len(*out)
}
func Ndr_ObjIsType(p *Ndr_Data_t, obj, typ int) bool {
	return Ndr_ObjReadEntry(p, obj, NDR_OPERTYPE) == typ
}
func Ndr_ObjReadBody(p *Ndr_Data_t, obj, typ int) int { return Ndr_ObjReadEntry(p, obj, typ) }
func Ndr_ObjReadBodyP(p *Ndr_Data_t, obj, typ int) []int {
	var a []int
	Ndr_ObjReadArray(p, obj, typ, &a)
	return a
}
func Ndr_ObjWriteRange(p *Ndr_Data_t, obj int, w io.Writer, skip int) {
	a := Ndr_ObjReadBodyP(p, obj, NDR_RANGE)
	if len(a) <= 1 && skip != 0 {
		return
	}
	if len(a) == 3 && skip != 0 {
		fmt.Fprint(w, "signed ")
	}
	if len(a) == 0 {
		fmt.Fprint(w, "[0]")
	} else if len(a) == 1 {
		fmt.Fprintf(w, "[%d]", a[0])
	} else {
		fmt.Fprintf(w, "[%d:%d]", a[0], a[1])
	}
}
func Ndr_ObjReadOutName(p *Ndr_Data_t, obj int, names []string) string {
	return names[Ndr_ObjReadBody(p, obj, NDR_OUTPUT)]
}
func Ndr_ObjReadInName(p *Ndr_Data_t, obj int, names []string) string {
	return names[Ndr_ObjReadBody(p, obj, NDR_INPUT)]
}
func ndrObjects(p *Ndr_Data_t, mod int) []int {
	var result []int
	for i := mod + 1; i < Ndr_DataEnd(p, mod); i += Ndr_DataSize(p, i) {
		if Ndr_DataType(p, i) == NDR_OBJECT {
			result = append(result, i)
		}
	}
	return result
}
func Ndr_DataCiNum(p *Ndr_Data_t, mod int) int {
	n := 0
	for _, o := range ndrObjects(p, mod) {
		if Ndr_ObjIsType(p, o, ABC_OPER_CI) {
			n++
		}
	}
	return n
}
func Ndr_DataCoNum(p *Ndr_Data_t, mod int) int {
	n := 0
	for _, o := range ndrObjects(p, mod) {
		if Ndr_ObjIsType(p, o, ABC_OPER_CO) {
			n++
		}
	}
	return n
}
func Ndr_DataObjNum(p *Ndr_Data_t, mod int) int { return len(ndrObjects(p, mod)) }
func Ndr_Create(name int) *Ndr_Data_t {
	if name <= 0 {
		panic("design name must be positive")
	}
	p := &Ndr_Data_t{}
	Ndr_DataPush(p, NDR_DESIGN, 2)
	Ndr_DataPush(p, NDR_NAME, name)
	return p
}
func Ndr_AddModule(p *Ndr_Data_t, name int) int {
	mod := p.nSize
	Ndr_DataPush(p, NDR_MODULE, 2)
	Ndr_DataPush(p, NDR_NAME, name)
	Ndr_DataAddTo(p, 0, 2)
	return mod + 256
}
func Ndr_AddObject(p *Ndr_Data_t, module, typ, name, left, right, signed, ni int, inputs []int, no int, outputs []int, function any) {
	mod := module - 256
	if typ == 0 || mod < 0 || mod >= p.nSize || Ndr_DataType(p, mod) != NDR_MODULE || Ndr_DataEnd(p, mod) != p.nSize {
		panic("objects must be appended to latest module")
	}
	obj := p.nSize
	Ndr_DataPush(p, NDR_OBJECT, 0)
	Ndr_DataPush(p, NDR_OPERTYPE, typ)
	Ndr_DataPushRange(p, left, right, signed)
	if name != 0 {
		Ndr_DataPush(p, NDR_NAME, name)
	}
	Ndr_DataPushArray(p, NDR_INPUT, ni, inputs)
	Ndr_DataPushArray(p, NDR_OUTPUT, no, outputs)
	Ndr_DataPushString(p, typ, NDR_FUNCTION, function)
	size := p.nSize - obj
	Ndr_DataAddTo(p, obj, size)
	Ndr_DataAddTo(p, mod, size)
	Ndr_DataAddTo(p, 0, size)
}
func Ndr_Delete(p *Ndr_Data_t) {
	if p != nil {
		p.pHead = nil
		p.pBody = nil
		p.nSize = 0
		p.nCap = 0
	}
}
func Ndr_Write(name string, p *Ndr_Data_t) {
	d := make([]byte, 5*p.nSize)
	for i, v := range p.pBody {
		binary.LittleEndian.PutUint32(d[4*i:], uint32(v))
	}
	copy(d[4*p.nSize:], p.pHead)
	if err := os.WriteFile(name, d, 0644); err != nil {
		panic(err)
	}
}
func Ndr_Read(name string) *Ndr_Data_t {
	d, e := os.ReadFile(name)
	if e != nil || len(d) < 10 || len(d)%5 != 0 {
		return nil
	}
	n := len(d) / 5
	p := &Ndr_Data_t{nSize: n, nCap: n, pHead: append([]byte(nil), d[4*n:]...), pBody: make([]int, n)}
	for i := range p.pBody {
		p.pBody[i] = int(int32(binary.LittleEndian.Uint32(d[4*i:])))
	}
	if p.pHead[0] != NDR_DESIGN || p.pBody[0] != n {
		return nil
	}
	for i, t := range p.pHead {
		if t == 0 || int(t) >= NDR_UNKNOWN {
			return nil
		}
		if int(t) <= NDR_OBJECT && (p.pBody[i] <= 0 || p.pBody[i] > n-i) {
			return nil
		}
	}
	return p
}
func ndrFunction(p *Ndr_Data_t, obj int) string {
	a := Ndr_ObjReadBodyP(p, obj, NDR_FUNCTION)
	var b []byte
	for _, v := range a {
		var word [4]byte
		binary.LittleEndian.PutUint32(word[:], uint32(v))
		b = append(b, word[:]...)
	}
	return strings.TrimRight(string(b), "\x00")
}
func Ndr_WriteVerilog(name string, p *Ndr_Data_t, names []string) {
	var w io.Writer = os.Stdout
	var f *os.File
	if name != "" {
		var e error
		f, e = os.Create(name)
		if e != nil {
			panic(e)
		}
		defer f.Close()
		w = f
	}
	for m := 1; m < p.nSize; m += Ndr_DataSize(p, m) {
		if Ndr_DataType(p, m) == NDR_MODULE {
			Ndr_WriteVerilogModule(w, p, m, names)
		}
	}
}
func Ndr_WriteVerilogModule(w io.Writer, p *Ndr_Data_t, mod int, names []string) {
	objects := ndrObjects(p, mod)
	var ports []string
	outs := map[int]bool{}
	for _, o := range objects {
		if Ndr_ObjIsType(p, o, ABC_OPER_CI) {
			ports = append(ports, Ndr_ObjReadOutName(p, o, names))
		}
	}
	for _, o := range objects {
		if Ndr_ObjIsType(p, o, ABC_OPER_CO) {
			ports = append(ports, Ndr_ObjReadInName(p, o, names))
			outs[Ndr_ObjReadBody(p, o, NDR_INPUT)] = true
		}
	}
	fmt.Fprintf(w, "\nmodule %s (%s);\n", names[Ndr_ObjReadEntry(p, mod, NDR_NAME)], strings.Join(ports, ", "))
	for _, o := range objects {
		typ := Ndr_ObjReadBody(p, o, NDR_OPERTYPE)
		keyword := "wire"
		id := Ndr_ObjReadBody(p, o, NDR_OUTPUT)
		if typ == ABC_OPER_CI {
			keyword = "input"
		} else if typ == ABC_OPER_CO {
			keyword = "output"
			id = Ndr_ObjReadBody(p, o, NDR_INPUT)
		} else if outs[id] {
			continue
		}
		fmt.Fprintf(w, "  %s ", keyword)
		Ndr_ObjWriteRange(p, o, w, 1)
		fmt.Fprintf(w, " %s;\n", names[id])
	}
	for _, o := range objects {
		typ := Ndr_ObjReadBody(p, o, NDR_OPERTYPE)
		if typ == ABC_OPER_CI || typ == ABC_OPER_CO {
			continue
		}
		a := Ndr_ObjReadBodyP(p, o, NDR_INPUT)
		var args []string
		for _, id := range a {
			args = append(args, names[id])
		}
		out := Ndr_ObjReadOutName(p, o, names)
		inst := ""
		if id := Ndr_ObjReadBody(p, o, NDR_NAME); id > 0 {
			inst = names[id]
		}
		if typ >= 256 {
			fmt.Fprintf(w, "  %s %s (%s);\n", names[Ndr_ObjReadEntry(p, typ-256, NDR_NAME)], inst, strings.Join(args, ", "))
			continue
		}
		var cell string
		var pins []string
		switch typ {
		case ABC_OPER_DFF:
			cell = "ABC_DFF"
			pins = []string{"q", "d", "init"}
		case ABC_OPER_DFFRSE:
			cell = "ABC_DFFRSE"
			pins = []string{"q", "d", "clk", "reset", "set", "enable", "async", "sre", "init"}
		case ABC_OPER_RAMR:
			cell = "ABC_READ"
			pins = []string{"data", "mem_in", "addr"}
		case ABC_OPER_RAMW:
			cell = "ABC_WRITE"
			pins = []string{"mem_out", "mem_in", "addr", "data"}
		}
		if cell != "" {
			values := append([]string{out}, args...)
			var conns []string
			for i, pin := range pins {
				if i >= len(values) {
					panic("missing cell input")
				}
				conns = append(conns, fmt.Sprintf(".%s(%s)", pin, values[i]))
			}
			fmt.Fprintf(w, "  %s %s (%s);\n", cell, inst, strings.Join(conns, ", "))
			continue
		}
		fmt.Fprintf(w, "  assign %s = ", out)
		switch {
		case len(args) == 0:
			fmt.Fprint(w, ndrFunction(p, o))
		case typ == ABC_OPER_BIT_BUF:
			fmt.Fprint(w, args[0])
		case typ == ABC_OPER_SLICE:
			fmt.Fprint(w, args[0])
			Ndr_ObjWriteRange(p, o, w, 0)
		case typ == ABC_OPER_CONCAT:
			fmt.Fprintf(w, "{%s}", strings.Join(args, ", "))
		case len(args) == 1:
			fmt.Fprintf(w, "%s %s", Abc_OperName(typ), args[0])
		case len(args) == 2:
			fmt.Fprintf(w, "%s %s %s", args[0], Abc_OperName(typ), args[1])
		case len(args) == 3 && typ == ABC_OPER_ARI_ADD:
			fmt.Fprint(w, strings.Join(args, " + "))
		case typ == ABC_OPER_BIT_MUX:
			fmt.Fprintf(w, "%s ? %s : %s", args[0], args[1], args[2])
		default:
			fmt.Fprintf(w, "<cannot write operation %s>", Abc_OperName(typ))
		}
		fmt.Fprintln(w, ";")
	}
	fmt.Fprintln(w, "\nendmodule")
}
