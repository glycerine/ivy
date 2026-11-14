package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
	"strings"
)

func Gia_ManGenUsed(p *Gia_Man_t, buffer int32) *Vec_Bit_t {
	used := vec.Vec_BitStart(int(p.nObjs))
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			if buffer != 0 {
				vec.Vec_BitWriteEntry(used, int(obj.id), 1)
			}
			if Gia_ObjFaninC0(obj)^buffer != 0 {
				vec.Vec_BitWriteEntry(used, int(Gia_ObjFaninId0(obj, obj.id)), 1)
			}
			if Gia_ObjFaninC1(obj)^buffer != 0 {
				vec.Vec_BitWriteEntry(used, int(Gia_ObjFaninId1(obj, obj.id)), 1)
			}
		} else if Gia_ObjIsCo(obj) != 0 && Gia_ObjFaninC0(obj)^buffer != 0 {
			vec.Vec_BitWriteEntry(used, int(Gia_ObjFaninId0(obj, obj.id)), 1)
		}
	}
	vec.Vec_BitWriteEntry(used, 0, 0)
	return used
}
func Gia_ManNameIsLegalInVerilog(name string) int32 {
	if name == "" {
		panic("empty Verilog name")
	}
	if name[0] == '\\' {
		return 1
	}
	letter := func(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' }
	if !letter(name[0]) {
		return 0
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !letter(c) && !(c >= '0' && c <= '9') && c != '$' {
			return 0
		}
	}
	return 1
}
func Gia_ObjGetDumpName(names *Vec_Ptr_t, prefix byte, index, digits int32) string {
	if names != nil {
		name, ok := vec.Vec_PtrEntry(names, int(index)).(string)
		if !ok {
			panic("invalid Verilog name")
		}
		if Gia_ManNameIsLegalInVerilog(name) != 0 {
			return name
		}
		return "\\" + name + " "
	}
	return fmt.Sprintf("%c%0*d%c", prefix, int(digits), index, prefix)
}
func giaDecimalDigits(n int32) int32 {
	digits := int32(1)
	for n >= 10 {
		n /= 10
		digits++
	}
	return digits
}
func Gia_ManWriteNames(w io.Writer, prefix byte, n int32, names *Vec_Ptr_t, start, skip int32, objects *Vec_Bit_t) {
	digits := giaDecimalDigits(n)
	length := start
	first := true
	for i := int32(0); i < n; i++ {
		if objects != nil && vec.Vec_BitEntry(objects, int(i)) == 0 {
			continue
		}
		name := Gia_ObjGetDumpName(names, prefix, i, digits)
		length += int32(len(name)) + 2
		if length > 60 {
			fmt.Fprint(w, ",\n    ")
			length = skip
			first = true
		}
		if !first {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprint(w, name)
		first = false
	}
}
func Gia_ManDumpVerilog(p *Gia_Man_t, filename string, objects *Vec_Int_t) {
	if Gia_ManRegNum(p) != 0 {
		fmt.Println("Currently cannot write sequential AIG.")
		return
	}
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Cannot open output file %q.\n", filename)
		return
	}
	defer file.Close()
	invs, used := Gia_ManGenUsed(p, 0), Gia_ManGenUsed(p, 1)
	digits := giaDecimalDigits(p.nObjs)
	digits2 := giaDecimalDigits(Gia_ManPiNum(p))
	name := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		return '_'
	}, p.pName)
	fmt.Fprintf(file, "module %s (\n    ", name)
	Gia_ManWriteNames(file, 'x', Gia_ManPiNum(p), p.VNamesIn, 4, 4, nil)
	fmt.Fprint(file, ",\n    ")
	Gia_ManWriteNames(file, 'z', Gia_ManPoNum(p), p.VNamesOut, 4, 4, nil)
	fmt.Fprint(file, "\n  );\n\n  input ")
	Gia_ManWriteNames(file, 'x', Gia_ManPiNum(p), p.VNamesIn, 8, 4, nil)
	fmt.Fprint(file, ";\n\n  output ")
	Gia_ManWriteNames(file, 'z', Gia_ManPoNum(p), p.VNamesOut, 9, 4, nil)
	fmt.Fprint(file, ";\n\n")
	if vec.Vec_BitCount(used) != 0 {
		fmt.Fprint(file, "  wire ")
		Gia_ManWriteNames(file, 'n', p.nObjs, nil, 7, 4, used)
		fmt.Fprint(file, ";\n\n")
	}
	if vec.Vec_BitCount(invs) != 0 {
		fmt.Fprint(file, "  wire ")
		Gia_ManWriteNames(file, 'i', p.nObjs, nil, 7, 4, invs)
		fmt.Fprint(file, ";\n\n")
	}
	skip := map[int32]bool{}
	if objects != nil {
		fmt.Fprint(file, "  wire ")
		for i, id := range vec.Vec_IntArray(objects) {
			skip[id] = true
			if i > 0 {
				fmt.Fprint(file, ",")
			}
			fmt.Fprintf(file, " t_%d", i)
		}
		fmt.Fprint(file, ";\n\n")
		for i, id := range vec.Vec_IntArray(objects) {
			fmt.Fprintf(file, "  buf( %s, t_%d );\n", Gia_ObjGetDumpName(nil, 'n', id, digits), i)
		}
		fmt.Fprintln(file)
	}
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		obj := Gia_ManPi(p, i)
		input := Gia_ObjGetDumpName(p.VNamesIn, 'x', i, digits2)
		if vec.Vec_BitEntry(used, int(obj.id)) != 0 {
			fmt.Fprintf(file, "  buf( %s, %s );\n", Gia_ObjGetDumpName(nil, 'n', obj.id, digits), input)
		}
		if vec.Vec_BitEntry(invs, int(obj.id)) != 0 {
			fmt.Fprintf(file, "  not( %s, %s );\n", Gia_ObjGetDumpName(nil, 'i', obj.id, digits), input)
		}
	}
	fmt.Fprintln(file)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		if !skip[id] {
			a, b := byte('n'), byte('n')
			if Gia_ObjFaninC0(obj) != 0 {
				a = 'i'
			}
			if Gia_ObjFaninC1(obj) != 0 {
				b = 'i'
			}
			fmt.Fprintf(file, "  and( %s, %s, %s );\n", Gia_ObjGetDumpName(nil, 'n', id, digits), Gia_ObjGetDumpName(nil, a, Gia_ObjFaninId0(obj, id), digits), Gia_ObjGetDumpName(nil, b, Gia_ObjFaninId1(obj, id), digits))
		}
		if vec.Vec_BitEntry(invs, int(id)) != 0 {
			fmt.Fprintf(file, "  not( %s, %s );\n", Gia_ObjGetDumpName(nil, 'i', id, digits), Gia_ObjGetDumpName(nil, 'n', id, digits))
		}
	}
	fmt.Fprintln(file)
	digits2 = giaDecimalDigits(Gia_ManPoNum(p))
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		obj := Gia_ManPo(p, i)
		var input string
		if Gia_ObjIsConst0(Gia_ObjFanin0(obj)) != 0 {
			input = fmt.Sprintf("1'b%d", Gia_ObjFaninC0(obj))
		} else {
			prefix := byte('n')
			if Gia_ObjFaninC0(obj) != 0 {
				prefix = 'i'
			}
			input = Gia_ObjGetDumpName(nil, prefix, Gia_ObjFanin0(obj).id, digits)
		}
		fmt.Fprintf(file, "  buf( %s, %s );\n", Gia_ObjGetDumpName(p.VNamesOut, 'z', i, digits2), input)
	}
	fmt.Fprint(file, "\nendmodule\n\n")
}
