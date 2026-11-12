package io

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/mio"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type bookCell struct {
	name     string
	width    int
	terminal bool
	object   *abc.Object
}
type bookPin struct {
	cell      *bookCell
	direction string
}
type bookNet struct {
	name string
	pins []bookPin
}
type bookModel struct {
	cells []*bookCell
	nets  []bookNet
	terms []*bookCell
	area  int
}

func bookSOP(o *abc.Object) string {
	switch f := o.Data.(type) {
	case string:
		return f
	case *mio.Mio_Gate_t:
		return mio.Mio_GateReadSop(f)
	}
	return ""
}
func bookTwo(o *abc.Object, truth byte) int32 {
	if len(o.Fanins()) != 2 {
		return 0
	}
	s := bookSOP(o)
	if s == "" {
		return 0
	}
	if abc.Abc_SopGetVarNum(s) != 2 {
		return 0
	}
	if byte(abc.Abc_SopToTruth(s, 2)&15) == truth {
		return 1
	}
	return 0
}
func Abc_NodeIsNand2(o *abc.Object) int32 { return bookTwo(o, 7) }
func Abc_NodeIsNor2(o *abc.Object) int32  { return bookTwo(o, 1) }
func Abc_NodeIsAnd2(o *abc.Object) int32 {
	s := bookSOP(o)
	if len(o.Fanins()) == 2 && s != "" {
		return abc.Abc_SopIsAndType(s)
	}
	return 0
}
func Abc_NodeIsOr2(o *abc.Object) int32 {
	if len(o.Fanins()) != 2 {
		return 0
	}
	s := bookSOP(o)
	if s == "" {
		return 0
	}
	if abc.Abc_SopIsOrType(s) != 0 || s == "01 0\n" || s == "10 0\n" || s == "00 0\n" {
		return 1
	}
	return 0
}
func Abc_NodeIsXor2(o *abc.Object) int32  { return bookTwo(o, 6) }
func Abc_NodeIsXnor2(o *abc.Object) int32 { return bookTwo(o, 9) }
func bookWidth(o *abc.Object) (int, error) {
	if o.Kind == abc.ObjectLatch {
		return 6, nil
	}
	if o.Kind == abc.ObjectNode && abc.Abc_NtkIsStrash(o.Network()) != 0 {
		return 3, nil
	}
	if o.Kind == abc.ObjectConst1 {
		return 0, nil
	}
	if o.Kind == abc.ObjectBlackbox || o.Kind == abc.ObjectWhitebox {
		return max(1, len(o.Fanins())), nil
	}
	if gate, ok := o.Data.(*mio.Mio_Gate_t); ok {
		area := mio.Mio_GateReadArea(gate)
		if math.IsNaN(area) || math.IsInf(area, 0) || area < 0 {
			return 0, fmt.Errorf("invalid mapped cell area")
		}
		return int(area), nil
	}
	s := bookSOP(o)
	if s == "" {
		return 0, fmt.Errorf("cell %s has no SOP or mapped gate", exportName(o))
	}
	if abc.Abc_SopIsConst0(s) != 0 || abc.Abc_SopIsConst1(s) != 0 {
		return 0, nil
	}
	if abc.Abc_SopIsInv(s) != 0 {
		return 1, nil
	}
	if abc.Abc_SopIsBuf(s) != 0 {
		return 2, nil
	}
	if Abc_NodeIsNand2(o) != 0 || Abc_NodeIsNor2(o) != 0 {
		return 2, nil
	}
	if Abc_NodeIsAnd2(o) != 0 || Abc_NodeIsOr2(o) != 0 {
		return 3, nil
	}
	if Abc_NodeIsXor2(o) != 0 || Abc_NodeIsXnor2(o) != 0 {
		return 5, nil
	}
	return len(o.Fanins()) + int(abc.Abc_SopGetCubeNum(s)), nil
}
func bookInputNet(o *abc.Object) *abc.Object {
	in := o.Fanins()
	if len(in) == 1 && in[0].Object.Kind == abc.ObjectNet {
		return in[0].Object
	}
	return o
}
func bookOutputNet(o *abc.Object) *abc.Object {
	for _, out := range o.Fanouts() {
		if out.Kind == abc.ObjectNet {
			return out
		}
	}
	return o
}
func bookCellName(o *abc.Object) string {
	switch o.Kind {
	case abc.ObjectPI:
		return "i" + exportName(bookOutputNet(o)) + "_input"
	case abc.ObjectPO:
		return "o" + exportName(bookInputNet(o)) + "_output"
	case abc.ObjectLatch:
		in, out := "in", "out"
		if len(o.Fanins()) == 1 {
			in = exportName(bookInputNet(o.Fanins()[0].Object))
		}
		if len(o.Fanouts()) == 1 {
			out = exportName(bookOutputNet(o.Fanouts()[0]))
		}
		return in + "_" + out + "_latch"
	}
	var b strings.Builder
	for _, e := range o.Fanins() {
		b.WriteString(exportName(e.Object))
		b.WriteByte('_')
	}
	for _, out := range o.Fanouts() {
		b.WriteString(exportName(out))
		b.WriteByte('_')
	}
	if gate, ok := o.Data.(*mio.Mio_Gate_t); ok {
		b.WriteString(mio.Mio_GateReadName(gate))
	} else {
		b.WriteString("name")
	}
	return b.String()
}
func makeBook(n *abc.Network) (*bookModel, error) {
	if n == nil {
		return nil, fmt.Errorf("nil network")
	}
	model := &bookModel{}
	modules := []*abc.Network{n}
	if n.Exdc != nil {
		modules = append(modules, n.Exdc)
	}
	if n.Design != nil {
		for _, m := range n.Design.Modules {
			if m != n && m != n.Exdc {
				modules = append(modules, m)
			}
		}
	}
	names := map[string]bool{}
	for mi, module := range modules {
		cells := map[*abc.Object]*bookCell{}
		prefix := ""
		if mi > 0 {
			prefix = fmt.Sprintf("m%d_", mi)
		}
		for _, kind := range []abc.ObjectKind{abc.ObjectPI, abc.ObjectPO, abc.ObjectLatch, abc.ObjectNode, abc.ObjectConst1, abc.ObjectBlackbox, abc.ObjectWhitebox} {
			for _, o := range module.ObjectsOfKind(kind) {
				if kind == abc.ObjectConst1 && len(o.Fanouts()) == 0 {
					continue
				}
				terminal := kind == abc.ObjectPI || kind == abc.ObjectPO
				width := 1
				var err error
				if !terminal {
					width, err = bookWidth(o)
					if err != nil {
						return nil, err
					}
				}
				name := prefix + bookCellName(o)
				if strings.ContainsAny(name, " \t\n\r") {
					return nil, fmt.Errorf("Bookshelf cell name contains whitespace: %q", name)
				}
				if names[name] {
					name = fmt.Sprintf("%s_%d", name, o.ID)
				}
				if names[name] {
					return nil, fmt.Errorf("duplicate Bookshelf cell name %q", name)
				}
				names[name] = true
				c := &bookCell{name, width, terminal, o}
				cells[o] = c
				model.cells = append(model.cells, c)
				if terminal {
					model.terms = append(model.terms, c)
				} else {
					model.area += width
				}
			}
		}
		owner := func(o *abc.Object) *bookCell {
			if o.Kind == abc.ObjectBI {
				for _, out := range o.Fanouts() {
					if out.Kind == abc.ObjectLatch {
						return cells[out]
					}
				}
			}
			if o.Kind == abc.ObjectBO {
				for _, e := range o.Fanins() {
					if e.Object.Kind == abc.ObjectLatch {
						return cells[e.Object]
					}
				}
			}
			return cells[o]
		}
		if nets := module.ObjectsOfKind(abc.ObjectNet); len(nets) > 0 {
			for _, net := range nets {
				inputs := net.Fanins()
				if len(inputs) != 1 {
					return nil, fmt.Errorf("net %s needs one driver", exportName(net))
				}
				driver := owner(inputs[0].Object)
				if driver == nil {
					return nil, fmt.Errorf("net %s driver has no cell", exportName(net))
				}
				bn := bookNet{name: prefix + exportName(net), pins: []bookPin{{driver, "I"}}}
				for _, sink := range net.Fanouts() {
					cell := owner(sink)
					if cell == nil {
						return nil, fmt.Errorf("net %s sink has no cell", exportName(net))
					}
					bn.pins = append(bn.pins, bookPin{cell, "O"})
				}
				model.nets = append(model.nets, bn)
			}
		} else {
			for _, driver := range module.Objects() {
				cell := owner(driver)
				if cell == nil || driver.Kind == abc.ObjectPO || driver.Kind == abc.ObjectLatch || driver.Kind == abc.ObjectBI {
					continue
				}
				bn := bookNet{name: prefix + exportName(driver), pins: []bookPin{{cell, "I"}}}
				for _, sink := range driver.Fanouts() {
					c := owner(sink)
					if c == nil {
						return nil, fmt.Errorf("logic fanout has no cell")
					}
					bn.pins = append(bn.pins, bookPin{c, "O"})
				}
				if len(bn.pins) > 1 {
					model.nets = append(model.nets, bn)
				}
			}
		}
	}
	return model, nil
}

// Io_NtkBfsPads finds the closest unordered PI/PO through undirected incidence.
// Sets are keyed by object pointers, avoiding the source's term-count-sized
// array indexed by arbitrary object IDs.
func Io_NtkBfsPads(n *abc.Network, start *abc.Object, ordered map[*abc.Object]bool) *abc.Object {
	queue := []*abc.Object{start}
	seen := map[*abc.Object]bool{start: true}
	for len(queue) > 0 {
		o := queue[0]
		queue = queue[1:]
		if o != start && (o.Kind == abc.ObjectPI || o.Kind == abc.ObjectPO) && !ordered[o] {
			return o
		}
		neighbors := append([]*abc.Object(nil), o.Fanouts()...)
		for _, e := range o.Fanins() {
			neighbors = append(neighbors, e.Object)
		}
		for _, out := range neighbors {
			if !seen[out] {
				seen[out] = true
				queue = append(queue, out)
			}
		}
	}
	return nil
}
func Io_NtkOrderingPads(n *abc.Network, terms []*abc.Object) []*abc.Object {
	ordered := map[*abc.Object]bool{}
	out := make([]*abc.Object, 0, len(terms))
	var current *abc.Object
	for len(out) < len(terms) {
		if current == nil || ordered[current] {
			for _, term := range terms {
				if !ordered[term] {
					current = term
					break
				}
			}
		}
		ordered[current] = true
		out = append(out, current)
		current = Io_NtkBfsPads(n, current, ordered)
	}
	return out
}
func bookFiles(model *bookModel, base string) map[string]string {
	files := map[string]string{}
	var nodes, nets, rows, placement strings.Builder
	fmt.Fprintf(&nodes, "UCLA    nodes    1.0\nNumNodes : %d\nNumTerminals : %d\n", len(model.cells), len(model.terms))
	for _, c := range model.cells {
		if c.terminal {
			fmt.Fprintf(&nodes, "%s\t 1 1 terminal\n", c.name)
		} else {
			fmt.Fprintf(&nodes, "%s\t %d 1\n", c.name, c.width)
		}
	}
	pins := 0
	for _, net := range model.nets {
		pins += len(net.pins)
	}
	fmt.Fprintf(&nets, "UCLA    nets    1.0\nNumNets : %d\nNumPins : %d\n", len(model.nets), pins)
	for _, net := range model.nets {
		fmt.Fprintf(&nets, "NetDegree\t:\t\t%d\t\t%s\n", len(net.pins), net.name)
		for _, pin := range net.pins {
			if pin.cell.terminal {
				fmt.Fprintf(&nets, "%s %s\n", pin.cell.name, pin.direction)
			} else {
				fmt.Fprintf(&nets, "%s %s : 0.00 0.00\n", pin.cell.name, pin.direction)
			}
		}
	}
	area := float64(max(model.area, 1)) / 0.9
	height := max(1, int(math.Sqrt(area)))
	width := max(1, int(area/float64(height)))
	fmt.Fprintf(&rows, "UCLA    scl    1.0\n\nNumrows : %d\n\n", height)
	for row := 0; row < height; row++ {
		orient := "N"
		if row%2 != 0 {
			orient = "FS"
		}
		fmt.Fprintf(&rows, "CoreRow Horizontal\n Coordinate   : \t%d\n Height       : \t1\n Sitewidth    : \t1\n Sitespacing  : \t1\n Siteorient   : \t%s\n Sitesymmetry : \tY\n SubrowOrigin : \t0 Numsites :    \t%d\nEnd\n", row, orient, width)
	}
	placement.WriteString("UCLA    pl    1.0\n\n")
	perimeter := 2 * float64(width+height)
	top := int(float64(len(model.terms)) * float64(width) / perimeter)
	bottom := top
	right := (len(model.terms) - top - bottom + 1) / 2
	left := len(model.terms) - top - bottom - right
	counts := []int{top, bottom, left, right}
	var ordered []*bookCell
	byObject := map[*abc.Object]*bookCell{}
	networks := []*abc.Network{}
	grouped := map[*abc.Network][]*abc.Object{}
	for _, c := range model.terms {
		n := c.object.Network()
		if _, ok := grouped[n]; !ok {
			networks = append(networks, n)
		}
		grouped[n] = append(grouped[n], c.object)
		byObject[c.object] = c
	}
	for _, n := range networks {
		for _, o := range Io_NtkOrderingPads(n, grouped[n]) {
			ordered = append(ordered, byObject[o])
		}
	}
	index := 0
	for side, count := range counts {
		extent := float64(width)
		if side >= 2 {
			extent = float64(height)
		}
		delta := 0.0
		if count > 0 {
			delta = extent / float64(count)
		}
		coordinate := 0.0
		for step := 0; step < count; step++ {
			if (side >= 2 || step > 0) && int(coordinate) < int(coordinate-delta)+1 {
				coordinate++
			}
			x, y, orientation := int(coordinate), height+2, "FS"
			switch side {
			case 1:
				y, orientation = -3, "N"
			case 2:
				x, y, orientation = -3, int(coordinate), "E"
			case 3:
				x, y, orientation = width+2, int(coordinate), "FW"
			}
			fmt.Fprintf(&placement, "%s\t\t%d\t\t%d\t: %s /FIXED\n", ordered[index].name, x, y, orientation)
			coordinate += delta
			index++
		}
	}
	for _, c := range model.cells {
		if !c.terminal {
			fmt.Fprintf(&placement, "%s\t\t0\t\t0\t: N\n", c.name)
		}
	}
	files[".nodes"], files[".nets"], files[".scl"], files[".pl"], files[".wts"] = nodes.String(), nets.String(), rows.String(), placement.String(), ""
	files[".aux"] = fmt.Sprintf("RowBasedPlacement : %s.nodes %s.nets %s.scl %s.pl %s.wts", base, base, base, base, base)
	return files
}
func Io_WriteBook(n *abc.Network, path string) error {
	model, err := makeBook(n)
	if err != nil {
		return err
	}
	files := bookFiles(model, filepath.Base(path))
	for _, suffix := range []string{".nodes", ".nets", ".scl", ".pl", ".wts", ".aux"} {
		if err := os.WriteFile(path+suffix, []byte(files[suffix]), 0644); err != nil {
			return err
		}
	}
	return nil
}
func Io_WriteBookLogic(n *abc.Network, path string) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	if abc.Abc_NtkIsStrash(n) == 0 && abc.Abc_NtkHasSop(n) == 0 && abc.Abc_NtkHasMapping(n) == 0 {
		n = abc.Abc_NtkDup(n)
		if n == nil || abc.Abc_NtkToSop(n, -1, 1000000) == 0 {
			return fmt.Errorf("Bookshelf SOP conversion failed")
		}
	}
	return Io_WriteBook(n, path)
}
func init() { RegisterFormat(IO_FILE_BOOK, nil, Io_WriteBookLogic) }
