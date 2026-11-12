package core

// CountISOPBounded uses the same authoritative Minato recursion as the ordinary
// BDD/ZDD ISOP operations, without building a ZDD cover. It returns nil,-1 when
// the cube or newly allocated BDD-node limit is exceeded. Failed allocations
// are rolled back; existing diagrams and manager allocation limits are retained.
func CountISOPBounded(dd *DdManager, lower, upper *DdNode, cubeLimit, nodeLimit int) (function *DdNode, count int) {
	count = -1
	if !dd.valid(lower, upper) || cubeLimit < 0 || nodeLimit < 0 || !Cudd_bddLeq(dd, lower, upper) {
		return
	}
	oldLimited, oldCeiling := dd.allocationLimited, dd.allocationCeiling
	firstID := dd.nextID
	maxInt := int(^uint(0) >> 1)
	ceiling := len(dd.unique) + min(nodeLimit, maxInt-len(dd.unique))
	if oldLimited {
		ceiling = min(ceiling, oldCeiling)
	}
	dd.allocationLimited, dd.allocationCeiling = true, ceiling
	defer func() {
		dd.allocationLimited, dd.allocationCeiling = oldLimited, oldCeiling
		if failure := recover(); failure != nil {
			if _, limited := failure.(bddAllocationLimit); !limited {
				panic(failure)
			}
			function = nil
			count = -1
		}
		if function == nil {
			for key, n := range dd.unique {
				if n.ID >= firstID {
					delete(dd.unique, key)
				}
			}
			clear(dd.cache)
			dd.nextID = firstID
		}
	}()
	result := dd.isopBounded(lower, upper, false, cubeLimit)
	if result.function != nil {
		function, count = result.function, result.cubes
	}
	return
}
