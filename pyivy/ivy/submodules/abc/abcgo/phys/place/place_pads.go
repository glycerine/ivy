// Core sizing and perimeter-pad placement from place_pads.c.
package place

import (
	"fmt"
	"math"
)

func validatePlacementGeometry() error {
	for _, c := range g_place_concreteCells {
		if c == nil {
			continue
		}
		if c.m_parent == nil || !qpFinite(c.m_x) || !qpFinite(c.m_y) || !qpFinite(c.m_parent.m_width) || !qpFinite(c.m_parent.m_height) || c.m_parent.m_width < 0 || c.m_parent.m_height < 0 {
			return fmt.Errorf("invalid geometry for cell %q", c.m_label)
		}
	}
	return nil
}

func globalPreplace(utilization float32) error {
	if err := validatePlacementGeometry(); err != nil {
		return err
	}
	if !qpFinite(utilization) || utilization <= 0 || utilization > 1 {
		return fmt.Errorf("utilization must be in (0,1]")
	}
	var area, padWidth, padHeight float32
	var pads []*ConcreteCell
	var fixedBounds Rect
	haveFixed := false
	rowHeight := float32(0)
	for _, c := range g_place_concreteCells {
		if c == nil {
			continue
		}
		kind := c.m_parent
		if kind.m_pad {
			if !c.m_fixed {
				pads = append(pads, c)
				padWidth = max(padWidth, kind.m_width)
				padHeight = max(padHeight, kind.m_height)
			}
			continue
		}
		area += getCellArea(c)
		rowHeight = max(rowHeight, kind.m_height)
		if c.m_fixed {
			left, bottom := c.m_x-kind.m_width/2, c.m_y-kind.m_height/2
			right, top := c.m_x+kind.m_width/2, c.m_y+kind.m_height/2
			if !haveFixed {
				fixedBounds = Rect{left, bottom, right, top}
				haveFixed = true
			} else {
				fixedBounds.x = min(fixedBounds.x, left)
				fixedBounds.y = min(fixedBounds.y, bottom)
				fixedBounds.w = max(fixedBounds.w, right)
				fixedBounds.h = max(fixedBounds.h, top)
			}
		}
	}
	if rowHeight <= 0 {
		rowHeight = 1
	}
	g_place_rowHeight = rowHeight
	if area == 0 && !haveFixed && len(pads) == 0 {
		g_place_coreBounds = Rect{}
		g_place_padBounds = Rect{}
		return nil
	}
	targetArea := area / utilization
	if !qpFinite(targetArea) {
		return fmt.Errorf("core area overflow")
	}
	height := float32(math.Floor(math.Sqrt(float64(targetArea))/float64(rowHeight))+1) * rowHeight
	core := Rect{h: height}
	if haveFixed {
		core.x = fixedBounds.x
		core.y = fixedBounds.y
		core.w = fixedBounds.w - fixedBounds.x
		core.h = max(core.h, fixedBounds.h-fixedBounds.y)
	}
	if core.h > 0 {
		core.w = max(core.w, targetArea/core.h)
	}
	g_place_coreBounds = core
	ring := Rect{core.x - padWidth, core.y - padHeight, core.w + 2*padWidth, core.h + 2*padHeight}
	g_place_padBounds = ring
	// Divide in source order: bottom, top, right, left. Fixed pads are untouched.
	remaining, start := len(pads), 0
	for side := 0; side < 4; side++ {
		count := remaining / (4 - side)
		for i, c := range pads[start : start+count] {
			w, h := c.m_parent.m_width, c.m_parent.m_height
			fraction := float32(i) / float32(count)
			switch side {
			case 0:
				c.m_x = ring.x + w/2 + (ring.w-padWidth)*fraction
				c.m_y = ring.y + h/2
			case 1:
				c.m_x = ring.x + ring.w - w/2 - (ring.w-2*padWidth)*fraction
				c.m_y = ring.y + ring.h - h/2
			case 2:
				c.m_x = ring.x + ring.w - w/2
				c.m_y = ring.y + h/2 + (ring.h-padHeight)*fraction
			case 3:
				c.m_x = ring.x + w/2
				c.m_y = ring.y + ring.h - h/2 - (ring.h-padHeight)*fraction
			}
		}
		start += count
		remaining -= count
	}
	return nil
}
func Preplace(utilization float32) error { return globalPreplace(utilization) }
