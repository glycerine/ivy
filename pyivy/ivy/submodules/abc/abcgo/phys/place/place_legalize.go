package place

import (
	"fmt"
	"math"
	"sort"
)

// Legalize packs movable cells into rows, treating fixed cells and pads as
// obstacles. Positions are committed only if every movable cell fits. The
// original C entry point was empty; this deterministic implementation provides
// a useful row legalizer without changing global placement's analytic flow.
func Legalize() error {
	if err := validatePlacementGeometry(); err != nil {
		return err
	}
	core := g_place_coreBounds
	row := g_place_rowHeight
	if row <= 0 || !qpFinite(row) || core.w < 0 || core.h < 0 || !qpFinite(core.x) || !qpFinite(core.y) || !qpFinite(core.w) || !qpFinite(core.h) {
		return fmt.Errorf("invalid core geometry")
	}
	type box struct{ x, y, w, h float32 }
	var occupied []box
	var movable []*ConcreteCell
	for _, c := range g_place_concreteCells {
		if c == nil {
			continue
		}
		if c.m_fixed || c.m_parent.m_pad {
			occupied = append(occupied, box{c.m_x - c.m_parent.m_width/2, c.m_y - c.m_parent.m_height/2, c.m_parent.m_width, c.m_parent.m_height})
		} else {
			movable = append(movable, c)
		}
	}
	sort.SliceStable(movable, func(i, j int) bool {
		if movable[i].m_y != movable[j].m_y {
			return movable[i].m_y < movable[j].m_y
		}
		if movable[i].m_x != movable[j].m_x {
			return movable[i].m_x < movable[j].m_x
		}
		return movable[i].m_id < movable[j].m_id
	})
	positions := make(map[*ConcreteCell][2]float32, len(movable))
	for _, c := range movable {
		w, h := c.m_parent.m_width, c.m_parent.m_height
		bestDistance := math.Inf(1)
		var best box
		found := false
		if w > core.w || h > core.h {
			return fmt.Errorf("cell %s exceeds core size", c.m_label)
		}
		rows := int(math.Floor(float64((core.h-h)/row))) + 1
		for r := 0; r < rows; r++ {
			bottom := core.y + float32(r)*row
			type interval struct{ left, right float32 }
			var blocked []interval
			for _, b := range occupied {
				if b.y < bottom+h && bottom < b.y+b.h && b.w > 0 && b.h > 0 {
					blocked = append(blocked, interval{max(core.x, b.x), min(core.x+core.w, b.x+b.w)})
				}
			}
			sort.Slice(blocked, func(i, j int) bool { return blocked[i].left < blocked[j].left })
			consider := func(left, right float32) {
				if right-left < w {
					return
				}
				x := clampPlacement(c.m_x-w/2, left, right-w)
				dx, dy := float64(x+w/2-c.m_x), float64(bottom+h/2-c.m_y)
				distance := dx*dx + dy*dy
				if distance < bestDistance {
					bestDistance = distance
					best = box{x, bottom, w, h}
					found = true
				}
			}
			cursor := core.x
			for _, b := range blocked {
				if b.right <= cursor || b.left >= core.x+core.w {
					continue
				}
				consider(cursor, b.left)
				cursor = max(cursor, b.right)
			}
			consider(cursor, core.x+core.w)
		}
		if !found {
			return fmt.Errorf("no legal row position for cell %s", c.m_label)
		}
		occupied = append(occupied, best)
		positions[c] = [2]float32{best.x + w/2, best.y + h/2}
	}
	for c, p := range positions {
		c.m_x, c.m_y = p[0], p[1]
	}
	return nil
}
func legalize() error { return Legalize() }
