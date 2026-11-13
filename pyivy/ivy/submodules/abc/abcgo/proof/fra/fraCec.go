package fra

// CEC driver ported from ABC's proof/fra/fraCec.c (Alan Mishchenko, 2007).

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/fracomb"
	"github.com/glycerine/abcgo/proof/framgr"
)

// Fra_FraigSat checks a combinational miter with the shared FRA SAT implementation.
func Fra_FraigSat(graph *aig.Graph, conflictLimit, propagationLimit int64, learnedStart, learnedDelta, learnedPercent, flipBits, andOutputs, newSolver, verbose int32) int32 {
	return fracomb.Fra_FraigSat(graph, conflictLimit, propagationLimit, learnedStart, learnedDelta, learnedPercent, flipBits, andOutputs, newSolver, verbose)
}

// Aig_ManCountXors counts XOR structures recognized by the AIG manager.
func Aig_ManCountXors(graph *aig.Graph) int32 { return fracomb.Aig_ManCountXors(graph) }

// Fra_FraigCec runs ABC's SAT, rewriting and FRAIG sequence.
func Fra_FraigCec(graph **aig.Graph, conflictLimit, verbose int32) int32 {
	return fracomb.Fra_FraigCec(graph, conflictLimit, verbose, framgr.SweepForCEC)
}

// Fra_FraigCecPartitioned checks matching output partitions of two networks.
func Fra_FraigCecPartitioned(left, right *aig.Graph, conflictLimit, partSize, smart, verbose int32) int32 {
	parts := aig.Aig_ManMiterPartitioned(left, right, partSize, smart)
	status := int32(1)
	completed := 0
	for i, miter := range parts {
		if verbose != 0 {
			fmt.Printf("Verifying part %4d (out of %4d) PI = %5d. PO = %5d. And = %6d. Lev = %4d.\r", i+1, len(parts), aig.Aig_ManCiNum(miter), aig.Aig_ManCoNum(miter), aig.Aig_ManNodeNum(miter), aig.Aig_ManLevelNum(miter))
		}
		status = Fra_FraigMiterStatus(miter)
		if status == -1 {
			status = Fra_FraigCec(&miter, conflictLimit, 0)
			parts[i] = miter
		}
		if status != 1 {
			break
		}
		completed++
	}
	if status == -1 {
		fmt.Printf("Timed out after verifying %d partitions (out of %d).\n", completed, len(parts))
	}
	return status
}

// Fra_FraigCecTop verifies two combinational AIGs, optionally partitioned.
func Fra_FraigCecTop(left, right *aig.Graph, conflictLimit, partSize, smart, verbose int32) int32 {
	if left == nil || right == nil || left.InputCount() != right.InputCount() || left.OutputCount() != right.OutputCount() {
		return 0
	}
	if aig.Aig_ManNodeNum(left) < aig.Aig_ManNodeNum(right) {
		left, right = right, left
	}
	if partSize == 0 {
		partSize, smart = int32(left.OutputCount()), 0
	}
	return Fra_FraigCecPartitioned(left, right, conflictLimit, partSize, smart, verbose)
}
