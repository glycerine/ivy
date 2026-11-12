// Analytic placement driver from place_gordian.c.
package place

import "fmt"

var g_place_numPartitions int32

func globalPlace() error {
	if g_place_numCells == 0 {
		return nil
	}
	initPartitioning()
	if err := constructQuadraticProblem(); err != nil {
		return err
	}
	for iteration := int32(0); ; iteration++ {
		if iteration > g_place_numCells+1 {
			return fmt.Errorf("partition refinement did not converge")
		}
		if err := solveQuadraticProblem(!IGNORE_COG); err != nil {
			return err
		}
		if REALLOCATE_PARTITIONS {
			reallocPartitions()
		}
		if refinePartitions() {
			break
		}
	}
	if FINAL_REALLOCATE_PARTITIONS {
		reallocPartitions()
	}
	if err := solveQuadraticProblem(!IGNORE_COG); err != nil {
		return err
	}
	sanitizePlacement()
	globalFixDensity(25, g_place_rowHeight*5)
	return nil
}
func globalIncremental() error {
	if g_place_rootPartition == nil {
		return globalPlace()
	}
	incrementalPartition()
	if err := constructQuadraticProblem(); err != nil {
		return err
	}
	if err := solveQuadraticProblem(!IGNORE_COG); err != nil {
		return err
	}
	sanitizePlacement()
	globalFixDensity(25, g_place_rowHeight*5)
	return nil
}

func clampPlacement(value, low, high float32) float32 {
	if low > high {
		return (low + high) / 2
	}
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
func sanitizeCoordinate(value, size, origin, extent, margin float32) float32 {
	low, high := origin+size/2, origin+extent-size/2
	if value < low {
		value = low + margin/(1+low-value)
	} else if value > high {
		value = high - margin/(1+value-high)
	}
	return clampPlacement(value, low, high)
}

// sanitizePlacement keeps movable cells inside the core, preserving edge order.
func sanitizePlacement() {
	for _, c := range g_place_concreteCells {
		if c == nil || c.m_fixed || c.m_parent.m_pad {
			continue
		}
		c.m_x = sanitizeCoordinate(c.m_x, c.m_parent.m_width, g_place_coreBounds.x, g_place_coreBounds.w, g_place_rowHeight)
		c.m_y = sanitizeCoordinate(c.m_y, c.m_parent.m_height, g_place_coreBounds.y, g_place_coreBounds.h, g_place_rowHeight)
	}
}
func GlobalPlace() error       { return globalPlace() }
func GlobalIncremental() error { return globalIncremental() }
func SanitizePlacement()       { sanitizePlacement() }
