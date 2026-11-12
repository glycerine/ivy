// Area-balanced bin spreading from place_bin.c.
package place

import (
	"fmt"
	"sort"
)

func globalFixDensity(numBins int32, maxMovement float32) { spreadDensityX(numBins, maxMovement) }
func spreadDensityX(numBins int32, maxMovement float32) {
	_ = spreadDensity(numBins, maxMovement, false)
}
func spreadDensityY(numBins int32, maxMovement float32) {
	_ = spreadDensity(numBins, maxMovement, true)
}

// FixDensity applies the original horizontal density pass. It does not resolve
// fixed-cell obstacles; Legalize performs that separate geometric operation.
func FixDensity(numBins int32, maxMovement float32) error {
	return spreadDensity(numBins, maxMovement, false)
}

func spreadDensity(numBins int32, maxMovement float32, vertical bool) error {
	if err := validatePlacementGeometry(); err != nil {
		return err
	}
	if numBins <= 0 || maxMovement < 0 || !qpFinite(maxMovement) {
		return fmt.Errorf("invalid density bin count or movement limit")
	}
	var cells []*ConcreteCell
	var total float32
	for _, c := range g_place_concreteCells {
		if c != nil && !c.m_fixed && !c.m_parent.m_pad {
			cells = append(cells, c)
			total += getCellArea(c)
		}
	}
	if total <= 0 || len(cells) == 0 {
		return nil
	}
	position := func(c *ConcreteCell) float32 {
		if vertical {
			return c.m_y
		}
		return c.m_x
	}
	cross := func(c *ConcreteCell) float32 {
		if vertical {
			return c.m_x
		}
		return c.m_y
	}
	sort.SliceStable(cells, func(i, j int) bool { return cross(cells[i]) < cross(cells[j]) })
	origin, extent := g_place_coreBounds.x, g_place_coreBounds.w
	if vertical {
		origin, extent = g_place_coreBounds.y, g_place_coreBounds.h
	}
	if extent < 0 || !qpFinite(extent) {
		return fmt.Errorf("invalid core extent")
	}
	outerStart, outerBin := 0, int32(0)
	var outerArea, cumulative float32
	for end, c := range cells {
		a := getCellArea(c)
		outerArea += a
		cumulative += a
		if end+1 < len(cells) && (outerArea <= 0 || cumulative < total*float32(outerBin+1)/float32(numBins)) {
			continue
		}
		row := append([]*ConcreteCell(nil), cells[outerStart:end+1]...)
		sort.SliceStable(row, func(i, j int) bool { return position(row[i]) < position(row[j]) })
		if outerArea > 0 {
			start, bin := 0, int32(0)
			oldEdge, newEdge := origin, origin
			var binArea, used float32
			for i, c := range row {
				a := getCellArea(c)
				binArea += a
				used += a
				if i+1 < len(row) && (binArea <= 0 || used < outerArea*float32(bin+1)/float32(numBins)) {
					continue
				}
				currentOld := position(c)
				currentNew := min(origin+extent, newEdge+extent*binArea/outerArea)
				currentNew = clampPlacement(currentNew, currentOld-maxMovement, currentOld+maxMovement)
				for j := start; j <= i; j++ {
					cell := row[j]
					old := position(cell)
					fraction := float32(j-start+1) / float32(i-start+2)
					if currentOld != oldEdge {
						fraction = (old - oldEdge) / (currentOld - oldEdge)
					}
					proposed := newEdge + fraction*(currentNew-newEdge)
					proposed = clampPlacement(proposed, old-maxMovement, old+maxMovement)
					size := cell.m_parent.m_width
					if vertical {
						size = cell.m_parent.m_height
					}
					proposed = clampPlacement(proposed, origin+size/2, origin+extent-size/2)
					if vertical {
						cell.m_y = proposed
					} else {
						cell.m_x = proposed
					}
				}
				oldEdge, newEdge = currentOld, currentNew
				start = i + 1
				bin++
				binArea = 0
			}
		}
		outerStart = end + 1
		outerBin++
		outerArea = 0
	}
	return nil
}
