// Bookshelf harness from place_test.c, expressed as reusable Go readers.
package place

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type bookshelfLine struct {
	text   string
	number int
}

func bookshelfLines(filename, kind string) ([]bookshelfLine, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 4<<20)
	var lines []bookshelfLine
	number := 0
	for scanner.Scan() {
		number++
		s, _, _ := strings.Cut(scanner.Text(), "#")
		s = strings.TrimSpace(s)
		if s != "" {
			lines = append(lines, bookshelfLine{s, number})
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	if len(lines) == 0 || lines[0].text != "UCLA "+kind+" 1.0" {
		return nil, fmt.Errorf("%s: missing UCLA %s 1.0 header", filename, kind)
	}
	return lines[1:], nil
}
func bookshelfCount(line bookshelfLine, name string) (int, error) {
	key, value, ok := strings.Cut(line.text, ":")
	if !ok || strings.TrimSpace(key) != name {
		return 0, fmt.Errorf("line %d: expected %s", line.number, name)
	}
	n, err := strconv.ParseInt(strings.TrimSpace(value), 10, 32)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("line %d: invalid %s count", line.number, name)
	}
	return int(n), nil
}
func bookshelfFloat(s string) (float32, error) {
	v, err := strconv.ParseFloat(s, 32)
	if err != nil || !qpFinite(float32(v)) {
		return 0, fmt.Errorf("invalid finite number %q", s)
	}
	return float32(v), nil
}
func parseBookshelfNodes(filename string) ([]*ConcreteCell, error) {
	lines, err := bookshelfLines(filename, "nodes")
	if err != nil {
		return nil, err
	}
	if len(lines) < 2 {
		return nil, fmt.Errorf("%s: missing node counts", filename)
	}
	count, err := bookshelfCount(lines[0], "NumNodes")
	if err != nil {
		return nil, err
	}
	terminals, err := bookshelfCount(lines[1], "NumTerminals")
	if err != nil {
		return nil, err
	}
	if len(lines)-2 != count {
		return nil, fmt.Errorf("%s: node count mismatch", filename)
	}
	cells := make([]*ConcreteCell, 0, count)
	seen := make(map[string]bool)
	pads := 0
	for _, line := range lines[2:] {
		fields := strings.Fields(line.text)
		if len(fields) < 3 || len(fields) > 4 {
			return nil, fmt.Errorf("%s:%d: invalid node record", filename, line.number)
		}
		name := fields[0]
		if seen[name] {
			return nil, fmt.Errorf("%s:%d: duplicate cell %s", filename, line.number, name)
		}
		seen[name] = true
		width, e := bookshelfFloat(fields[1])
		if e != nil || width < 0 {
			return nil, fmt.Errorf("%s:%d: invalid width", filename, line.number)
		}
		height, e := bookshelfFloat(fields[2])
		if e != nil || height < 0 {
			return nil, fmt.Errorf("%s:%d: invalid height", filename, line.number)
		}
		pad := false
		if len(fields) == 4 {
			if fields[3] != "terminal" && fields[3] != "terminal_NI" {
				return nil, fmt.Errorf("%s:%d: invalid terminal marker", filename, line.number)
			}
			pad = true
			pads++
		}
		cells = append(cells, NewCell(int32(len(cells)), name, NewCellType(name, width, height, pad), 0, 0))
	}
	if pads != terminals {
		return nil, fmt.Errorf("%s: terminal count mismatch", filename)
	}
	return cells, nil
}
func bookshelfNames(cells []*ConcreteCell) (map[string]*ConcreteCell, error) {
	names := make(map[string]*ConcreteCell)
	for _, c := range cells {
		if c == nil {
			continue
		}
		name := bookshelfCellName(c)
		if names[name] != nil {
			return nil, fmt.Errorf("duplicate Bookshelf cell name %q", name)
		}
		names[name] = c
	}
	return names, nil
}
func parseBookshelfNets(filename string, cells []*ConcreteCell) ([]*ConcreteNet, error) {
	lines, err := bookshelfLines(filename, "nets")
	if err != nil {
		return nil, err
	}
	if len(lines) < 2 {
		return nil, fmt.Errorf("%s: missing net counts", filename)
	}
	count, err := bookshelfCount(lines[0], "NumNets")
	if err != nil {
		return nil, err
	}
	pins, err := bookshelfCount(lines[1], "NumPins")
	if err != nil {
		return nil, err
	}
	names, err := bookshelfNames(cells)
	if err != nil {
		return nil, err
	}
	var nets []*ConcreteNet
	usedPins := 0
	for i := 2; i < len(lines); {
		line := lines[i]
		i++
		key, value, ok := strings.Cut(line.text, ":")
		fields := strings.Fields(value)
		if !ok || strings.TrimSpace(key) != "NetDegree" || len(fields) == 0 {
			return nil, fmt.Errorf("%s:%d: expected NetDegree", filename, line.number)
		}
		degree, e := strconv.ParseInt(fields[0], 10, 32)
		if e != nil || degree < 0 || degree > int64(len(lines)-i) {
			return nil, fmt.Errorf("%s:%d: invalid net degree", filename, line.number)
		}
		terms := make([]*ConcreteCell, 0, int(degree))
		for j := int64(0); j < degree; j++ {
			pin := lines[i]
			i++
			fields := strings.Fields(pin.text)
			cell := names[fields[0]]
			if cell == nil {
				return nil, fmt.Errorf("%s:%d: unknown cell %s", filename, pin.number, fields[0])
			}
			terms = append(terms, cell)
		}
		nets = append(nets, NewNet(int32(len(nets)), terms, 1))
		usedPins += int(degree)
	}
	if len(nets) != count || usedPins != pins {
		return nil, fmt.Errorf("%s: net/pin count mismatch", filename)
	}
	return nets, nil
}

type bookshelfPosition struct {
	cell  *ConcreteCell
	x, y  float32
	fixed bool
}

func parseBookshelfPlacement(filename string, cells []*ConcreteCell) ([]bookshelfPosition, error) {
	lines, err := bookshelfLines(filename, "pl")
	if err != nil {
		return nil, err
	}
	names, err := bookshelfNames(cells)
	if err != nil {
		return nil, err
	}
	var positions []bookshelfPosition
	seen := make(map[*ConcreteCell]bool)
	for _, line := range lines {
		fields := strings.Fields(line.text)
		if len(fields) < 3 {
			return nil, fmt.Errorf("%s:%d: invalid placement", filename, line.number)
		}
		c := names[fields[0]]
		if c == nil || seen[c] {
			return nil, fmt.Errorf("%s:%d: unknown or repeated cell %s", filename, line.number, fields[0])
		}
		seen[c] = true
		x, e := bookshelfFloat(fields[1])
		if e != nil {
			return nil, fmt.Errorf("%s:%d: %w", filename, line.number, e)
		}
		y, e := bookshelfFloat(fields[2])
		if e != nil {
			return nil, fmt.Errorf("%s:%d: %w", filename, line.number, e)
		}
		fixed := false
		for _, s := range fields[3:] {
			if s == "/FIXED" || s == "\\FIXED" || s == "/FIXED_NI" {
				fixed = true
			}
		}
		positions = append(positions, bookshelfPosition{c, x, y, fixed})
	}
	return positions, nil
}
func installBookshelfNodes(cells []*ConcreteCell) {
	g_place_concreteCells = nil
	g_place_concreteNets = nil
	g_place_numCells = 0
	g_place_numNets = 0
	g_place_concreteCellsSize = 0
	g_place_concreteNetsSize = 0
	g_place_rootPartition = nil
	g_place_qpProb = nil
	g_place_numPartitions = 0
	for _, c := range cells {
		addConcreteCell(c)
	}
}
func applyBookshelfPositions(positions []bookshelfPosition) {
	for _, p := range positions {
		p.cell.m_x, p.cell.m_y, p.cell.m_fixed = p.x, p.y, p.fixed
	}
}
func readBookshelfNodes(filename string) error {
	cells, err := parseBookshelfNodes(filename)
	if err != nil {
		return err
	}
	installBookshelfNodes(cells)
	return nil
}
func readBookshelfNets(filename string) error {
	nets, err := parseBookshelfNets(filename, g_place_concreteCells)
	if err != nil {
		return err
	}
	g_place_concreteNets = nil
	g_place_numNets = 0
	g_place_concreteNetsSize = 0
	for _, n := range nets {
		addConcreteNet(n)
	}
	return nil
}
func readBookshelfPlacement(filename string) error {
	positions, err := parseBookshelfPlacement(filename, g_place_concreteCells)
	if err != nil {
		return err
	}
	applyBookshelfPositions(positions)
	return nil
}
func writeBookshelfPlacement(filename string) error { return writeBookshelfPl(filename) }
func ReadBookshelfNodes(filename string) error      { return readBookshelfNodes(filename) }
func ReadBookshelfNets(filename string) error       { return readBookshelfNets(filename) }
func ReadBookshelfPlacement(filename string) error  { return readBookshelfPlacement(filename) }

// ReadBookshelfFiles installs a complete netlist transactionally. Coordinates
// retain ABC's center convention; pin offsets and orientations are not modeled.
func ReadBookshelfFiles(nodes, nets, placement string) error {
	cells, err := parseBookshelfNodes(nodes)
	if err != nil {
		return err
	}
	connections, err := parseBookshelfNets(nets, cells)
	if err != nil {
		return err
	}
	positions, err := parseBookshelfPlacement(placement, cells)
	if err != nil {
		return err
	}
	applyBookshelfPositions(positions)
	installBookshelfNodes(cells)
	for _, n := range connections {
		addConcreteNet(n)
	}
	return nil
}
func delNetConnections(cell *ConcreteCell) {
	for _, net := range g_place_concreteNets {
		if net == nil {
			continue
		}
		terms := net.m_terms[:0]
		for _, c := range net.m_terms {
			if c != cell {
				terms = append(terms, c)
			}
		}
		clear(net.m_terms[len(terms):])
		net.m_terms = terms
		net.m_numTerms = int32(len(terms))
	}
}

// RunBookshelf is the original three-file harness, returning errors to its caller.
func RunBookshelf(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: bookshelf nodes nets placement")
	}
	if err := ReadBookshelfFiles(args[0], args[1], args[2]); err != nil {
		return err
	}
	if err := globalPreplace(.8); err != nil {
		return err
	}
	if err := globalPlace(); err != nil {
		return err
	}
	if err := globalIncremental(); err != nil {
		return err
	}
	return writeBookshelfPlacement(args[2])
}
