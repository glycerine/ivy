package io

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	stdio "io"
	"os"
	"strconv"
	"strings"
)

func exportName(o *abc.Object) string {
	if name := o.Name(); name != "" {
		return name
	}
	return fmt.Sprintf("n%d", o.ID)
}
func exportFile(filename string, write func(stdio.Writer) error) error {
	var b strings.Builder
	if err := write(&b); err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(b.String()), 0644)
}
func exportQuoted(s string) string { return strconv.Quote(s) }

// WriteGML follows Io_WriteGml: POs, PIs, latches and nodes (each section
// after an empty line), then one edge per fanin.
func WriteGML(w stdio.Writer, n *abc.Network) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# GML for \"%s\" written by ABC on %s\n", n.Name, extra.Extra_TimeStamp())
	b.WriteString("graph [\n")
	node := func(o *abc.Object, shape, color string) {
		fmt.Fprintf(&b, "    node [ id %5d label \"%s\"\n", o.ID, abc.Abc_ObjName(o))
		fmt.Fprintf(&b, "        graphics [ type \"%s\" fill \"%s\" ]\n", shape, color)
		b.WriteString("    ]\n")
	}
	b.WriteString("\n")
	for i := int32(0); i < abc.Abc_NtkPoNum(n); i++ {
		node(abc.Abc_NtkPo(n, i), "triangle", "#00FFFF")
	}
	b.WriteString("\n")
	for i := int32(0); i < abc.Abc_NtkPiNum(n); i++ {
		node(abc.Abc_NtkPi(n, i), "triangle", "#00FF00")
	}
	b.WriteString("\n")
	for _, latch := range n.ObjectsOfKind(abc.ObjectLatch) {
		node(latch, "rectangle", "#FF0000")
	}
	b.WriteString("\n")
	for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
		if o := abc.Abc_NtkObj(n, i); o != nil && abc.Abc_ObjIsNode(o) != 0 {
			node(o, "ellipse", "#CCCCFF")
		}
	}
	b.WriteString("\n")
	for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
		o := abc.Abc_NtkObj(n, i)
		if o == nil {
			continue
		}
		for _, e := range o.Fanins() {
			fmt.Fprintf(&b, "    edge [ source %5d   target %5d\n", o.ID, e.Object.ID)
			b.WriteString("        graphics [ type \"line\" arrow \"first\" ]\n")
			b.WriteString("    ]\n")
		}
	}
	b.WriteString("]\n\n")
	_, err := stdio.WriteString(w, b.String())
	return err
}
func Io_WriteGml(n *abc.Network, filename string) error {
	return exportFile(filename, func(w stdio.Writer) error { return WriteGML(w, n) })
}
func init() { RegisterFormat(IO_FILE_GML, nil, Io_WriteGml) }
