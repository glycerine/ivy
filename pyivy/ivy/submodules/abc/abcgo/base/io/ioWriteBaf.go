package io

import (
	"bytes"
	"encoding/binary"
	"fmt"
	stdio "io"
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
)

// WriteBAF writes a strashed network as Io_WriteBaf does.
func WriteBAF(w stdio.Writer, n *abc.Network) error {
	if n == nil || abc.Abc_NtkIsStrash(n) == 0 {
		return fmt.Errorf("Writing this format is only possible for structurally hashed AIGs.")
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "# BAF (Binary Aig Format) for \"%s\" written by ABC on %s\n", n.Name, extra.Extra_TimeStamp())
	fmt.Fprintf(&out, "%s%c", n.Name, 0)
	fmt.Fprintf(&out, "%d%c", abc.Abc_NtkPiNum(n), 0)
	fmt.Fprintf(&out, "%d%c", abc.Abc_NtkPoNum(n), 0)
	fmt.Fprintf(&out, "%d%c", abc.Abc_NtkLatchNum(n), 0)
	fmt.Fprintf(&out, "%d%c", abc.Abc_NtkNodeNum(n), 0)
	for i := int32(0); i < abc.Abc_NtkPiNum(n); i++ {
		fmt.Fprintf(&out, "%s%c", abc.Abc_ObjName(abc.Abc_NtkPi(n, i)), 0)
	}
	for i := int32(0); i < abc.Abc_NtkPoNum(n); i++ {
		fmt.Fprintf(&out, "%s%c", abc.Abc_ObjName(abc.Abc_NtkPo(n, i)), 0)
	}
	for _, latch := range n.ObjectsOfKind(abc.ObjectLatch) {
		fmt.Fprintf(&out, "%s%c", abc.Abc_ObjName(latch), 0)
		fmt.Fprintf(&out, "%s%c", abc.Abc_ObjName(abc.Abc_ObjFanin0(latch)), 0)
		fmt.Fprintf(&out, "%s%c", abc.Abc_ObjName(abc.Abc_ObjFanout0(latch)), 0)
	}
	// CIs are numbered from 1, then the ANDs in id order; the constant is 0
	number := map[*abc.Object]uint32{}
	next := uint32(1)
	for i := int32(0); i < abc.Abc_NtkCiNum(n); i++ {
		number[abc.Abc_NtkCi(n, i)] = next
		next++
	}
	var ands []*abc.Object
	for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
		if o := abc.Abc_NtkObj(n, i); o != nil && abc.Abc_ObjIsNode(o) != 0 {
			number[o] = next
			next++
			ands = append(ands, o)
		}
	}
	word := func(v uint32) {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], v)
		out.Write(b[:])
	}
	lit := func(o *abc.Object, k int32) uint32 {
		return number[abc.Abc_ObjFanin(o, k)]<<1 | uint32(abc.Abc_ObjFaninC(o, k))
	}
	for _, o := range ands {
		word(lit(o, 0))
		word(lit(o, 1))
	}
	for i := int32(0); i < abc.Abc_NtkCoNum(n); i++ {
		co := abc.Abc_NtkCo(n, i)
		v := lit(co, 0)
		if abc.Abc_ObjFanoutNum(co) > 0 && abc.Abc_ObjIsLatch(abc.Abc_ObjFanout0(co)) != 0 {
			v = v<<2 | uint32(bafLatchInit(abc.Abc_ObjFanout0(co)))&3
		}
		word(v)
	}
	_, err := w.Write(out.Bytes())
	return err
}

// bafLatchInit is the latch's init value (Abc_ObjData of a latch).
func bafLatchInit(latch *abc.Object) abc.Abc_InitType_t {
	switch {
	case abc.Abc_LatchIsInit0(latch) != 0:
		return abc.ABC_INIT_ZERO
	case abc.Abc_LatchIsInit1(latch) != 0:
		return abc.ABC_INIT_ONE
	case abc.Abc_LatchIsInitDc(latch) != 0:
		return abc.ABC_INIT_DC
	}
	return abc.ABC_INIT_NONE
}

func WriteBAFFile(n *abc.Network, filename string) error {
	if abc.Abc_NtkIsStrash(n) == 0 {
		fmt.Printf("Writing this format is only possible for structurally hashed AIGs.\n")
		return nil
	}
	var out bytes.Buffer
	if err := WriteBAF(&out, n); err != nil {
		return err
	}
	if os.WriteFile(filename, out.Bytes(), 0644) != nil {
		fmt.Printf("Io_WriteBaf(): Cannot open the output file \"%s\".\n", filename)
	}
	return nil
}
func Io_WriteBaf(n *abc.Network, filename string) error { return WriteBAFFile(n, filename) }
func init()                                             { RegisterFormat(IO_FILE_BAF, nil, WriteBAFFile) }
