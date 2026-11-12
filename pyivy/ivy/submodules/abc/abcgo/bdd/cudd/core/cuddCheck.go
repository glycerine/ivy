package core

import (
	"fmt"
	"strings"
)

// Cudd_CheckConsistency verifies canonical-table keys, node ownership,
// complemented-edge pairs, child ordering, and variable permutations.
func Cudd_CheckConsistency(dd *DdManager) error {
	if dd == nil {
		return fmt.Errorf("nil manager")
	}
	if len(dd.Perm) != len(dd.Vars) || len(dd.InvPerm) != len(dd.Vars) {
		return fmt.Errorf("BDD permutation lengths do not match variable count")
	}
	for i, v := range dd.Vars {
		if v == nil || v.manager != dd || int(v.Index) != i {
			return fmt.Errorf("BDD variable %d has invalid identity", i)
		}
		level := dd.Perm[i]
		if level < 0 || level >= len(dd.InvPerm) || dd.InvPerm[level] != i {
			return fmt.Errorf("BDD permutation is inconsistent at variable %d", i)
		}
	}
	if err := checkUniqueTable(dd, dd.unique, dd.Perm, "BDD"); err != nil {
		return err
	}
	if err := checkUniqueTable(dd, dd.addUnique, dd.Perm, "ADD"); err != nil {
		return err
	}
	if err := checkUniqueTable(dd, dd.zddUnique, dd.PermZ, "ZDD"); err != nil {
		return err
	}
	return nil
}

func checkUniqueTable(dd *DdManager, table map[bddUniqueKey]*DdNode, permutation []int, family string) error {
	for key, node := range table {
		if node == nil || node.manager != dd {
			return fmt.Errorf("%s unique table contains a foreign or nil node", family)
		}
		if node.Index != key.index || node.T != key.t || node.E != key.e {
			return fmt.Errorf("%s unique table key disagrees with node %d", family, node.ID)
		}
		if node.T == nil || node.E == nil || node.T.manager != dd || node.E.manager != dd {
			return fmt.Errorf("%s node %d has invalid children", family, node.ID)
		}
		if node.regular != nil {
			if node.regular.inverse != node || node.inverse != node.regular || node.regular.regular != nil {
				return fmt.Errorf("node %d has inconsistent complemented-edge pair", node.ID)
			}
		}
		if int(node.Index) < len(permutation) {
			parentLevel := permutation[node.Index]
			for _, child := range []*DdNode{node.T, node.E} {
				child = Cudd_Regular(child)
				if int(child.Index) < len(permutation) && permutation[child.Index] <= parentLevel {
					return fmt.Errorf("%s node %d violates variable order", family, node.ID)
				}
			}
		}
	}
	return nil
}

// Cudd_ConsistencySummary returns a compact diagnostic useful to callers.
func Cudd_ConsistencySummary(dd *DdManager) string {
	if err := Cudd_CheckConsistency(dd); err != nil {
		return strings.TrimSpace(err.Error())
	}
	return ""
}
