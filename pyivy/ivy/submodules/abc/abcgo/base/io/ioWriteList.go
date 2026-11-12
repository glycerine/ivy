package io

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	stdio "io"
	"strings"
)

func Io_WriteListEdge(w stdio.Writer, o *abc.Object) error {
	var b strings.Builder
	fmt.Fprintf(&b, "%-10s >    ", exportName(o))
	for i, out := range o.Fanouts() {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, " %s ([%s_to_%s] = 0)", exportName(out), exportName(o), exportName(out))
	}
	b.WriteString(".\n")
	_, err := stdio.WriteString(w, b.String())
	return err
}
func Io_WriteListHost(w stdio.Writer, n *abc.Network) error {
	for _, o := range n.ObjectsOfKind(abc.ObjectPO) {
		if _, err := fmt.Fprintf(w, "%-10s >     HOST ([%s_to_HOST] = 0).\n", exportName(o), exportName(o)); err != nil {
			return err
		}
	}
	var b strings.Builder
	b.WriteString("HOST       >    ")
	for i, o := range n.ObjectsOfKind(abc.ObjectPI) {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, " %s ([HOST_to_%s] = 0)", exportName(o), exportName(o))
	}
	b.WriteString(".\n")
	_, err := stdio.WriteString(w, b.String())
	return err
}
func WriteList(w stdio.Writer, n *abc.Network, useHost bool) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	if _, err := fmt.Fprintf(w, "# Adjacency list for sequential AIG %q\n# written by ABC\n", n.Name); err != nil {
		return err
	}
	for _, kind := range []abc.ObjectKind{abc.ObjectConst1, abc.ObjectPI, abc.ObjectNode} {
		for _, o := range n.ObjectsOfKind(kind) {
			if kind == abc.ObjectConst1 && len(o.Fanouts()) == 0 {
				continue
			}
			if err := Io_WriteListEdge(w, o); err != nil {
				return err
			}
		}
	}
	if useHost {
		if err := Io_WriteListHost(w, n); err != nil {
			return err
		}
	} else {
		for _, o := range n.ObjectsOfKind(abc.ObjectPO) {
			if err := Io_WriteListEdge(w, o); err != nil {
				return err
			}
		}
	}
	_, err := stdio.WriteString(w, "\n")
	return err
}
func Io_WriteList(n *abc.Network, filename string, useHost int32) error {
	return exportFile(filename, func(w stdio.Writer) error { return WriteList(w, n, useHost != 0) })
}
func cellNetSink(o *abc.Object) *abc.Object {
	if o.Kind == abc.ObjectBO && len(o.Fanins()) == 1 {
		return o.Fanins()[0].Object
	}
	return o
}
func WriteCellNet(w stdio.Writer, n *abc.Network) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# CellNet file for network %q written by ABC\n", n.Name)
	var drivers []*abc.Object
	for _, kind := range []abc.ObjectKind{abc.ObjectPI, abc.ObjectPO, abc.ObjectLatch, abc.ObjectNode} {
		for _, o := range n.ObjectsOfKind(kind) {
			id, typ := o.ID, 0
			switch kind {
			case abc.ObjectPI:
				drivers = append(drivers, o)
			case abc.ObjectPO:
				typ = 1
			case abc.ObjectLatch:
				typ = 2
				if len(o.Fanins()) != 1 || len(o.Fanouts()) != 1 {
					return fmt.Errorf("latch %s lacks terminal pair", exportName(o))
				}
				id = o.Fanins()[0].Object.ID
				drivers = append(drivers, o)
			case abc.ObjectNode:
				typ = 3 + len(o.Fanins())
				drivers = append(drivers, o)
			}
			fmt.Fprintf(&b, "cell %d is %d\n", id, typ)
		}
	}
	for _, driver := range drivers {
		id := driver.ID
		source := driver
		if driver.Kind == abc.ObjectLatch {
			id = driver.Fanins()[0].Object.ID
			source = driver.Fanouts()[0]
		}
		fmt.Fprintf(&b, "net %d  %d 0", id, id)
		seen := make(map[*abc.Object]bool)
		for _, sink := range source.Fanouts() {
			if seen[sink] {
				continue
			}
			seen[sink] = true
			for pin, e := range sink.Fanins() {
				if e.Object == source {
					fmt.Fprintf(&b, "  %d %d", cellNetSink(sink).ID, pin+1)
				}
			}
		}
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	_, err := stdio.WriteString(w, b.String())
	return err
}
func Io_WriteCellNet(n *abc.Network, filename string) error {
	return exportFile(filename, func(w stdio.Writer) error { return WriteCellNet(w, n) })
}
func init() {
	RegisterFormat(IO_FILE_LIST, nil, func(n *abc.Network, path string) error { return Io_WriteList(n, path, 0) })
}
