package io

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
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
func gmlQuoted(s string) string {
	return `"` + strings.NewReplacer("&", "&amp;", "\"", "&quot;", "\n", "&#10;", "\r", "&#13;").Replace(s) + `"`
}
func WriteGML(w stdio.Writer, n *abc.Network) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# GML for %s written by ABC\ngraph [\n", exportQuoted(n.Name))
	for _, kind := range []abc.ObjectKind{abc.ObjectPO, abc.ObjectPI, abc.ObjectLatch, abc.ObjectNode, abc.ObjectConst1, abc.ObjectBI, abc.ObjectBO, abc.ObjectNet, abc.ObjectBlackbox, abc.ObjectWhitebox} {
		for _, o := range n.ObjectsOfKind(kind) {
			if kind == abc.ObjectConst1 && len(o.Fanouts()) == 0 {
				continue
			}
			shape, color := "ellipse", "#CCCCFF"
			switch kind {
			case abc.ObjectPO:
				shape, color = "triangle", "#00FFFF"
			case abc.ObjectPI:
				shape, color = "triangle", "#00FF00"
			case abc.ObjectLatch:
				shape, color = "rectangle", "#FF0000"
			}
			fmt.Fprintf(&b, "    node [ id %5d label %s\n        graphics [ type %s fill %s ]\n    ]\n", o.ID, gmlQuoted(exportName(o)), gmlQuoted(shape), gmlQuoted(color))
		}
	}
	for _, o := range n.Objects() {
		for _, e := range o.Fanins() {
			fmt.Fprintf(&b, "    edge [ source %5d   target %5d\n        graphics [ type \"line\" arrow \"first\" ]\n    ]\n", o.ID, e.Object.ID)
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
