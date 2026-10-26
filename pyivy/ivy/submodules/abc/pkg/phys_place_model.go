package abc

// Native Go port of src/phys/place/place_base.h, place_base.c, place_bin.c,
// place_io.c, and place_pads.c.

import (
	"bytes"
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrPlacementDuplicate  = errors.New("placement ID already exists")
	ErrPlacementMissing    = errors.New("placement object does not exist")
	ErrPlacementReferenced = errors.New("placement cell is referenced by a net")
	ErrPlacementFormat     = errors.New("invalid Bookshelf placement data")
)

// PlacementRect is a lower-left corner followed by width and height.
type PlacementRect struct {
	X, Y, Width, Height float64
}

// PlacementCellType is the source AbstractCell with Go-owned text.
type PlacementCellType struct {
	Label         string
	Width, Height float64
	Pad           bool
}

// PlacementCell is the source ConcreteCell. Coordinates denote its center.
type PlacementCell struct {
	ID       int
	Label    string
	Type     PlacementCellType
	Fixed    bool
	X, Y     float64
	UserData int
}

// PlacementNet is the source ConcreteNet with term IDs instead of pointers.
type PlacementNet struct {
	ID       int
	Terms    []int
	Weight   float64
	UserData int
}

// PlacementLimits bounds sparse IDs and total stored connectivity. Zero uses
// conservative defaults. Stop is polled by operations that can do linear work.
type PlacementLimits struct {
	MaxCellID, MaxNetID, MaxTerms, MaxBytes int
	Stop                                    func() bool
}

// Placement owns a complete placement database. Sparse integer ID spaces are
// retained because the source explicitly permits them.
type Placement struct {
	cells      []*PlacementCell
	nets       []*PlacementNet
	termCount  int
	RowHeight  float64
	CoreBounds PlacementRect
	PadBounds  PlacementRect
	Limits     PlacementLimits
}

func placementLimit(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func placementFinite(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}

func (p *Placement) cancelled() bool { return p != nil && p.Limits.Stop != nil && p.Limits.Stop() }

// NewPlacement returns an empty placement database.
func NewPlacement(limits PlacementLimits) (*Placement, error) {
	if limits.MaxCellID < 0 || limits.MaxNetID < 0 || limits.MaxTerms < 0 || limits.MaxBytes < 0 {
		return nil, ErrDimension
	}
	return &Placement{RowHeight: 1, Limits: limits}, nil
}

func placementCloneCell(cell *PlacementCell) *PlacementCell {
	if cell == nil {
		return nil
	}
	copyCell := *cell
	return &copyCell
}

func placementCloneNet(net *PlacementNet) *PlacementNet {
	if net == nil {
		return nil
	}
	copyNet := *net
	copyNet.Terms = append([]int(nil), net.Terms...)
	return &copyNet
}

// Clone returns an independent database snapshot.
func (p *Placement) Clone() (*Placement, error) {
	if p == nil {
		return nil, ErrDimension
	}
	result := *p
	result.cells = make([]*PlacementCell, len(p.cells))
	for id, cell := range p.cells {
		result.cells[id] = placementCloneCell(cell)
	}
	result.nets = make([]*PlacementNet, len(p.nets))
	for id, net := range p.nets {
		result.nets[id] = placementCloneNet(net)
	}
	return &result, nil
}

func placementValidCell(cell PlacementCell) bool {
	return cell.ID >= 0 && cell.Type.Width > 0 && cell.Type.Height > 0 &&
		placementFinite(cell.Type.Width, cell.Type.Height, cell.X, cell.Y)
}

// AddCell adds an owned copy and grows the sparse ID space transactionally.
func (p *Placement) AddCell(cell PlacementCell) error {
	if p == nil || !placementValidCell(cell) {
		return ErrDimension
	}
	if p.cancelled() {
		return ErrCancelled
	}
	if cell.ID > placementLimit(p.Limits.MaxCellID, 1_000_000) {
		return ErrLimit
	}
	if cell.ID < len(p.cells) && p.cells[cell.ID] != nil {
		return ErrPlacementDuplicate
	}
	if cell.ID >= len(p.cells) {
		grown := make([]*PlacementCell, cell.ID+1)
		copy(grown, p.cells)
		p.cells = grown
	}
	p.cells[cell.ID] = placementCloneCell(&cell)
	return nil
}

// AddNet validates every term before committing an owned copy.
func (p *Placement) AddNet(net PlacementNet) error {
	if p == nil || net.ID < 0 || !placementFinite(net.Weight) || net.Weight < 0 {
		return ErrDimension
	}
	if p.cancelled() {
		return ErrCancelled
	}
	if net.ID > placementLimit(p.Limits.MaxNetID, 1_000_000) ||
		len(net.Terms) > placementLimit(p.Limits.MaxTerms, 10_000_000)-p.termCount {
		return ErrLimit
	}
	if net.ID < len(p.nets) && p.nets[net.ID] != nil {
		return ErrPlacementDuplicate
	}
	for _, id := range net.Terms {
		if id < 0 || id >= len(p.cells) || p.cells[id] == nil {
			return ErrPlacementMissing
		}
	}
	if net.ID >= len(p.nets) {
		grown := make([]*PlacementNet, net.ID+1)
		copy(grown, p.nets)
		p.nets = grown
	}
	copyNet := placementCloneNet(&net)
	p.nets[net.ID] = copyNet
	p.termCount += len(copyNet.Terms)
	return nil
}

// DeleteNet removes a net without renumbering other IDs.
func (p *Placement) DeleteNet(id int) error {
	if p == nil || id < 0 || id >= len(p.nets) || p.nets[id] == nil {
		return ErrPlacementMissing
	}
	p.termCount -= len(p.nets[id].Terms)
	p.nets[id] = nil
	for len(p.nets) != 0 && p.nets[len(p.nets)-1] == nil {
		p.nets = p.nets[:len(p.nets)-1]
	}
	return nil
}

// DeleteCell rejects dangling connectivity. The source left stale pointers.
// Partition-tree removal is intentionally absent until the Gordian partition
// manager is ported.
func (p *Placement) DeleteCell(id int) error {
	if p == nil || id < 0 || id >= len(p.cells) || p.cells[id] == nil {
		return ErrPlacementMissing
	}
	for _, net := range p.nets {
		if net == nil {
			continue
		}
		for _, term := range net.Terms {
			if term == id {
				return ErrPlacementReferenced
			}
		}
	}
	p.cells[id] = nil
	for len(p.cells) != 0 && p.cells[len(p.cells)-1] == nil {
		p.cells = p.cells[:len(p.cells)-1]
	}
	return nil
}

// Cell returns an owned cell snapshot.
func (p *Placement) Cell(id int) (PlacementCell, bool) {
	if p == nil || id < 0 || id >= len(p.cells) || p.cells[id] == nil {
		return PlacementCell{}, false
	}
	return *p.cells[id], true
}

// Net returns an owned net snapshot.
func (p *Placement) Net(id int) (PlacementNet, bool) {
	if p == nil || id < 0 || id >= len(p.nets) || p.nets[id] == nil {
		return PlacementNet{}, false
	}
	return *placementCloneNet(p.nets[id]), true
}

// CellIDs and NetIDs return deterministic increasing-ID snapshots.
func (p *Placement) CellIDs() []int {
	result := make([]int, 0)
	if p != nil {
		for id, cell := range p.cells {
			if cell != nil {
				result = append(result, id)
			}
		}
	}
	return result
}

func (p *Placement) NetIDs() []int {
	result := make([]int, 0)
	if p != nil {
		for id, net := range p.nets {
			if net != nil {
				result = append(result, id)
			}
		}
	}
	return result
}

// CellArea preserves getCellArea.
func (p *Placement) CellArea(id int) (float64, error) {
	cell, ok := p.Cell(id)
	if !ok {
		return 0, ErrPlacementMissing
	}
	return cell.Type.Width * cell.Type.Height, nil
}

// NetBBox returns the centers' bounding box. Empty nets are rejected instead
// of returning the source's overflow-derived rectangle.
func (p *Placement) NetBBox(id int) (PlacementRect, error) {
	if p == nil || id < 0 || id >= len(p.nets) || p.nets[id] == nil {
		return PlacementRect{}, ErrPlacementMissing
	}
	net := p.nets[id]
	if len(net.Terms) == 0 {
		return PlacementRect{}, ErrDimension
	}
	first := p.cells[net.Terms[0]]
	result := PlacementRect{X: first.X, Y: first.Y}
	maxX, maxY := first.X, first.Y
	for _, term := range net.Terms[1:] {
		cell := p.cells[term]
		result.X = math.Min(result.X, cell.X)
		result.Y = math.Min(result.Y, cell.Y)
		maxX = math.Max(maxX, cell.X)
		maxY = math.Max(maxY, cell.Y)
	}
	result.Width, result.Height = maxX-result.X, maxY-result.Y
	return result, nil
}

func (p *Placement) NetWirelength(id int) (float64, error) {
	box, err := p.NetBBox(id)
	if err != nil {
		return 0, err
	}
	return box.Width + box.Height, nil
}

func (p *Placement) TotalWirelength() (float64, error) {
	if p == nil {
		return 0, ErrDimension
	}
	total := 0.0
	for id, net := range p.nets {
		if net == nil {
			continue
		}
		value, err := p.NetWirelength(id)
		if err != nil {
			return 0, err
		}
		total += value
	}
	return total, nil
}

// PlacementCellOrder and PlacementNetOrder cover the source qsort helpers.
type PlacementCellOrder uint8

const (
	PlacementCellByX PlacementCellOrder = iota
	PlacementCellByY
	PlacementCellByID
)

type PlacementNetOrder uint8

const (
	PlacementNetByLeft PlacementNetOrder = iota
	PlacementNetByRight
	PlacementNetByBottom
	PlacementNetByTop
	PlacementNetByID
)

func (p *Placement) SortedCellIDs(order PlacementCellOrder) ([]int, error) {
	ids := p.CellIDs()
	if order > PlacementCellByID {
		return nil, ErrDimension
	}
	sort.SliceStable(ids, func(i, j int) bool {
		a, b := p.cells[ids[i]], p.cells[ids[j]]
		if order == PlacementCellByX && a.X != b.X {
			return a.X < b.X
		}
		if order == PlacementCellByY && a.Y != b.Y {
			return a.Y < b.Y
		}
		return a.ID < b.ID
	})
	return ids, nil
}

func (p *Placement) SortedNetIDs(order PlacementNetOrder) ([]int, error) {
	ids := p.NetIDs()
	if order > PlacementNetByID {
		return nil, ErrDimension
	}
	boxes := make(map[int]PlacementRect, len(ids))
	if order != PlacementNetByID {
		for _, id := range ids {
			box, err := p.NetBBox(id)
			if err != nil {
				return nil, err
			}
			boxes[id] = box
		}
	}
	value := func(id int) float64 {
		box := boxes[id]
		switch order {
		case PlacementNetByLeft:
			return box.X
		case PlacementNetByRight:
			return box.X + box.Width
		case PlacementNetByBottom:
			return box.Y + box.Height // source netSortByB
		case PlacementNetByTop:
			return box.Y // source netSortByT
		}
		return float64(id)
	}
	sort.SliceStable(ids, func(i, j int) bool {
		a, b := value(ids[i]), value(ids[j])
		return a < b || a == b && ids[i] < ids[j]
	})
	return ids, nil
}

// Preplace computes the core and pad ring and places unfixed pads in source
// north/south/east/west order. Bounds calculation fixes the source's inverted
// extrema and uninitialized pad-bound reads.
func (p *Placement) Preplace(utilization float64) error {
	if p == nil || utilization <= 0 || !placementFinite(utilization) {
		return ErrDimension
	}
	if p.cancelled() {
		return ErrCancelled
	}
	work, _ := p.Clone()
	coreArea, rowHeight := 0.0, 0.0
	padWidth, padHeight := 0.0, 0.0
	pads := make([]*PlacementCell, 0)
	minX, minY, maxX, maxY := 0.0, 0.0, 0.0, 0.0
	haveFixed := false
	for _, cell := range work.cells {
		if cell == nil {
			continue
		}
		if work.cancelled() {
			return ErrCancelled
		}
		if cell.Type.Pad {
			padWidth = math.Max(padWidth, cell.Type.Width)
			padHeight = math.Max(padHeight, cell.Type.Height)
			if !cell.Fixed {
				pads = append(pads, cell)
			}
		} else {
			coreArea += cell.Type.Width * cell.Type.Height
			rowHeight = math.Max(rowHeight, cell.Type.Height)
		}
		if cell.Fixed {
			if !haveFixed {
				minX, maxX, minY, maxY = cell.X, cell.X, cell.Y, cell.Y
				haveFixed = true
			} else {
				minX, maxX = math.Min(minX, cell.X), math.Max(maxX, cell.X)
				minY, maxY = math.Min(minY, cell.Y), math.Max(maxY, cell.Y)
			}
		}
	}
	if padWidth == 0 || padHeight == 0 || rowHeight == 0 || coreArea == 0 {
		return ErrDimension
	}
	targetArea := coreArea / utilization
	rows := int(math.Ceil(math.Sqrt(targetArea) / rowHeight))
	if rows < 1 {
		rows = 1
	}
	height := float64(rows) * rowHeight
	width := targetArea / height
	if haveFixed {
		work.CoreBounds = PlacementRect{X: math.Min(0, minX), Y: math.Min(0, minY), Width: math.Max(width, maxX-math.Min(0, minX)), Height: math.Max(height, maxY-math.Min(0, minY))}
	} else {
		work.CoreBounds = PlacementRect{Width: width, Height: height}
	}
	work.RowHeight = rowHeight
	work.PadBounds = PlacementRect{X: work.CoreBounds.X - padWidth, Y: work.CoreBounds.Y - padHeight, Width: work.CoreBounds.Width + 2*padWidth, Height: work.CoreBounds.Height + 2*padHeight}
	remaining, at := len(pads), 0
	counts := [4]int{remaining / 4, 0, 0, 0}
	remaining -= counts[0]
	counts[1] = remaining / 3
	remaining -= counts[1]
	counts[2] = remaining / 2
	remaining -= counts[2]
	counts[3] = remaining
	for side, count := range counts {
		for i := 0; i < count; i++ {
			cell := pads[at]
			at++
			switch side {
			case 0:
				cell.X = work.PadBounds.X + cell.Type.Width/2 + float64(i)*(work.PadBounds.Width-padWidth)/float64(count)
				cell.Y = work.PadBounds.Y + cell.Type.Height/2
			case 1:
				cell.X = work.PadBounds.X + work.PadBounds.Width - cell.Type.Width/2 - float64(i)*(work.PadBounds.Width-2*padWidth)/float64(count)
				cell.Y = work.PadBounds.Y + work.PadBounds.Height - cell.Type.Height/2
			case 2:
				cell.X = work.PadBounds.X + work.PadBounds.Width - cell.Type.Width/2
				cell.Y = work.PadBounds.Y + cell.Type.Height/2 + float64(i)*(work.PadBounds.Height-padHeight)/float64(count)
			case 3:
				cell.X = work.PadBounds.X + cell.Type.Width/2
				cell.Y = work.PadBounds.Y + work.PadBounds.Height - cell.Type.Height/2 - float64(i)*(work.PadBounds.Height-padHeight)/float64(count)
			}
		}
	}
	*p = *work
	return nil
}

func placementClampCenter(value, half, low, span float64) (float64, error) {
	if 2*half > span {
		return 0, ErrDimension
	}
	return math.Max(low+half, math.Min(low+span-half, value)), nil
}

func (p *Placement) spreadDensity(numBins int, maxMovement float64, spreadX bool) error {
	if p == nil || numBins <= 0 || maxMovement < 0 || !placementFinite(maxMovement) || p.CoreBounds.Width <= 0 || p.CoreBounds.Height <= 0 {
		return ErrDimension
	}
	if p.cancelled() {
		return ErrCancelled
	}
	work, _ := p.Clone()
	ids := make([]int, 0)
	totalArea := 0.0
	for id, cell := range work.cells {
		if work.cancelled() {
			return ErrCancelled
		}
		if cell != nil && !cell.Fixed && !cell.Type.Pad {
			if cell.Type.Width > work.CoreBounds.Width || cell.Type.Height > work.CoreBounds.Height {
				return ErrDimension
			}
			ids = append(ids, id)
			totalArea += cell.Type.Width * cell.Type.Height
		}
	}
	if len(ids) == 0 {
		return nil
	}
	outerValue := func(id int) float64 {
		if spreadX {
			return work.cells[id].Y
		}
		return work.cells[id].X
	}
	innerValue := func(cell *PlacementCell) float64 {
		if spreadX {
			return cell.X
		}
		return cell.Y
	}
	sort.SliceStable(ids, func(i, j int) bool {
		a, b := outerValue(ids[i]), outerValue(ids[j])
		return a < b || a == b && ids[i] < ids[j]
	})
	start, cumulative, outerBin := 0, 0.0, 0
	for end, id := range ids {
		if work.cancelled() {
			return ErrCancelled
		}
		cell := work.cells[id]
		cumulative += cell.Type.Width * cell.Type.Height
		if cumulative+1e-12 < totalArea*float64(outerBin+1)/float64(numBins) && end != len(ids)-1 {
			continue
		}
		group := append([]int(nil), ids[start:end+1]...)
		sort.SliceStable(group, func(i, j int) bool {
			a, b := innerValue(work.cells[group[i]]), innerValue(work.cells[group[j]])
			return a < b || a == b && group[i] < group[j]
		})
		groupArea := 0.0
		for _, groupID := range group {
			c := work.cells[groupID]
			groupArea += c.Type.Width * c.Type.Height
		}
		low, span := work.CoreBounds.Y, work.CoreBounds.Height
		if spreadX {
			low, span = work.CoreBounds.X, work.CoreBounds.Width
		}
		innerStart, innerArea, innerCum, innerBin := 0, 0.0, 0.0, 0
		lastOld, lastNew := low, low
		for innerEnd, groupID := range group {
			if work.cancelled() {
				return ErrCancelled
			}
			c := work.cells[groupID]
			area := c.Type.Width * c.Type.Height
			innerArea += area
			innerCum += area
			if innerCum+1e-12 < groupArea*float64(innerBin+1)/float64(numBins) && innerEnd != len(group)-1 {
				continue
			}
			curOld := innerValue(c)
			curNew := math.Min(low+span, lastNew+span*innerArea/groupArea)
			curNew = math.Max(curOld-maxMovement, math.Min(curOld+maxMovement, curNew))
			for offset := innerStart; offset <= innerEnd; offset++ {
				moving := work.cells[group[offset]]
				old := innerValue(moving)
				newValue := lastNew
				if curOld == lastOld {
					denom := innerEnd - innerStart + 1
					newValue += float64(offset-innerStart+1) * (curNew - lastNew) / float64(denom)
				} else {
					newValue += (old - lastOld) * (curNew - lastNew) / (curOld - lastOld)
				}
				newValue = math.Max(old-maxMovement, math.Min(old+maxMovement, newValue))
				half := moving.Type.Height / 2
				if spreadX {
					half = moving.Type.Width / 2
				}
				var err error
				newValue, err = placementClampCenter(newValue, half, low, span)
				if err != nil {
					return err
				}
				if spreadX {
					moving.X = newValue
				} else {
					moving.Y = newValue
				}
			}
			lastOld, lastNew = curOld, curNew
			innerStart, innerArea, innerBin = innerEnd+1, 0, innerBin+1
		}
		start, outerBin = end+1, outerBin+1
	}
	*p = *work
	return nil
}

func (p *Placement) SpreadDensityX(numBins int, maxMovement float64) error {
	return p.spreadDensity(numBins, maxMovement, true)
}

func (p *Placement) SpreadDensityY(numBins int, maxMovement float64) error {
	return p.spreadDensity(numBins, maxMovement, false)
}

// FixDensity preserves globalFixDensity's X-only public behavior.
func (p *Placement) FixDensity(numBins int, maxMovement float64) error {
	return p.SpreadDensityX(numBins, maxMovement)
}

// BookshelfFiles owns the two files emitted by this source slice.
type BookshelfFiles struct {
	Nodes, Placement []byte
}

func placementAppendFloat(dst []byte, value float64) []byte {
	return strconv.AppendFloat(dst, value, 'f', 6, 64)
}

// Bookshelf serializes cells in deterministic source ID order without global
// filenames or process termination on I/O failure.
func (p *Placement) Bookshelf() (BookshelfFiles, error) {
	if err := p.Validate(); err != nil {
		return BookshelfFiles{}, err
	}
	ids := p.CellIDs()
	terminals := 0
	for _, id := range ids {
		if p.cells[id].Type.Pad {
			terminals++
		}
	}
	var nodes, pl bytes.Buffer
	nodes.WriteString("UCLA nodes 1.0\nNumNodes : ")
	nodes.WriteString(strconv.Itoa(len(ids)))
	nodes.WriteString("\nNumTerminals : ")
	nodes.WriteString(strconv.Itoa(terminals))
	nodes.WriteByte('\n')
	pl.WriteString("UCLA pl 1.0\n")
	for _, id := range ids {
		cell := p.cells[id]
		line := []byte("CELL" + strconv.Itoa(id) + " ")
		line = placementAppendFloat(line, cell.Type.Width)
		line = append(line, ' ')
		line = placementAppendFloat(line, cell.Type.Height)
		line = append(line, ' ')
		if cell.Type.Pad {
			line = append(line, []byte(" terminal")...)
		}
		line = append(line, '\n')
		nodes.Write(line)
		line = []byte("CELL" + strconv.Itoa(id) + " ")
		line = placementAppendFloat(line, cell.X)
		line = append(line, ' ')
		line = placementAppendFloat(line, cell.Y)
		line = append(line, []byte(" : N ")...)
		if cell.Fixed {
			line = append(line, []byte("\\FIXED")...)
		}
		line = append(line, '\n')
		pl.Write(line)
	}
	return BookshelfFiles{Nodes: nodes.Bytes(), Placement: pl.Bytes()}, nil
}

func placementParseCellName(value string) (int, bool) {
	if !strings.HasPrefix(value, "CELL") || len(value) == 4 {
		return 0, false
	}
	id, err := strconv.Atoi(value[4:])
	return id, err == nil && id >= 0
}

// ParseBookshelf parses the deterministic nodes/pl subset written above.
func ParseBookshelf(nodes, pl []byte, limits PlacementLimits) (*Placement, error) {
	result, err := NewPlacement(limits)
	if err != nil {
		return nil, err
	}
	maxBytes := placementLimit(limits.MaxBytes, 64<<20)
	if len(nodes) > maxBytes || len(pl) > maxBytes-len(nodes) {
		return nil, ErrLimit
	}
	lines := strings.Split(string(nodes), "\n")
	declared, terminals := -1, -1
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "UCLA" {
			continue
		}
		if fields[0] == "NumNodes" && len(fields) == 3 {
			declared, err = strconv.Atoi(fields[2])
			if err != nil || declared < 0 {
				return nil, ErrPlacementFormat
			}
			continue
		}
		if fields[0] == "NumTerminals" && len(fields) == 3 {
			terminals, err = strconv.Atoi(fields[2])
			if err != nil || terminals < 0 {
				return nil, ErrPlacementFormat
			}
			continue
		}
		if len(fields) < 3 {
			return nil, ErrPlacementFormat
		}
		id, ok := placementParseCellName(fields[0])
		width, e1 := strconv.ParseFloat(fields[1], 64)
		height, e2 := strconv.ParseFloat(fields[2], 64)
		if !ok || e1 != nil || e2 != nil {
			return nil, ErrPlacementFormat
		}
		pad := len(fields) >= 4 && fields[len(fields)-1] == "terminal"
		if err = result.AddCell(PlacementCell{ID: id, Label: fields[0], Type: PlacementCellType{Label: fields[0], Width: width, Height: height, Pad: pad}}); err != nil {
			return nil, err
		}
	}
	if declared < 0 || declared != len(result.CellIDs()) {
		return nil, ErrPlacementFormat
	}
	actualTerminals := 0
	for _, id := range result.CellIDs() {
		if result.cells[id].Type.Pad {
			actualTerminals++
		}
	}
	if terminals < 0 || terminals != actualTerminals {
		return nil, ErrPlacementFormat
	}
	seen := make(map[int]bool, declared)
	for _, line := range strings.Split(string(pl), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "UCLA" {
			continue
		}
		if len(fields) < 5 {
			return nil, ErrPlacementFormat
		}
		id, ok := placementParseCellName(fields[0])
		x, e1 := strconv.ParseFloat(fields[1], 64)
		y, e2 := strconv.ParseFloat(fields[2], 64)
		if !ok || e1 != nil || e2 != nil || id >= len(result.cells) || result.cells[id] == nil || seen[id] || fields[3] != ":" || fields[4] != "N" {
			return nil, ErrPlacementFormat
		}
		result.cells[id].X, result.cells[id].Y = x, y
		result.cells[id].Fixed = len(fields) == 6 && fields[5] == "\\FIXED"
		if len(fields) > 6 || len(fields) == 6 && !result.cells[id].Fixed {
			return nil, ErrPlacementFormat
		}
		seen[id] = true
	}
	if len(seen) != declared {
		return nil, ErrPlacementFormat
	}
	return result, result.Validate()
}

// Validate checks all ownership, geometry, sparse-ID and connectivity invariants.
func (p *Placement) Validate() error {
	if p == nil || !placementFinite(p.RowHeight, p.CoreBounds.X, p.CoreBounds.Y, p.CoreBounds.Width, p.CoreBounds.Height, p.PadBounds.X, p.PadBounds.Y, p.PadBounds.Width, p.PadBounds.Height) || p.RowHeight <= 0 {
		return ErrDimension
	}
	terms := 0
	for id, cell := range p.cells {
		if cell != nil && (cell.ID != id || !placementValidCell(*cell)) {
			return ErrDimension
		}
	}
	for id, net := range p.nets {
		if net == nil {
			continue
		}
		if net.ID != id || net.Weight < 0 || !placementFinite(net.Weight) {
			return ErrDimension
		}
		terms += len(net.Terms)
		for _, term := range net.Terms {
			if term < 0 || term >= len(p.cells) || p.cells[term] == nil {
				return ErrPlacementMissing
			}
		}
	}
	if terms != p.termCount {
		return ErrDimension
	}
	return nil
}
