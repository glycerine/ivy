package fra

// CEC driver ported from ABC's proof/fra/fraCec.c (Alan Mishchenko, 2007).

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/proof/fracomb"
)

// Fra_FraigSat checks a combinational miter with the shared FRA SAT implementation.
func Fra_FraigSat(graph *aig.Graph, conflictLimit, propagationLimit int64, learnedStart, learnedDelta, learnedPercent, flipBits, andOutputs, newSolver, verbose int32) int32 {
	return fracomb.Fra_FraigSat(graph, conflictLimit, propagationLimit, learnedStart, learnedDelta, learnedPercent, flipBits, andOutputs, newSolver, verbose)
}

// Aig_ManCountXors counts XOR structures recognized by the AIG manager.
func Aig_ManCountXors(graph *aig.Graph) int32 {
	var count int32
	for _, edge := range graph.LiveEdges() {
		if aig.Aig_ObjIsMuxType(graph, edge) {
			if _, _, ok := aig.Aig_ObjRecognizeExor(graph, edge); ok {
				count++
			}
		}
	}
	return count
}

// Fra_FraigCec applies the original sequence of SAT, rewriting, and FRAIG
// attempts to the given miter. Rewrites replace the caller's graph.
func Fra_FraigCec(graph **aig.Graph, conflictLimit, verbose int32) int32 {
	if graph == nil || *graph == nil {
		panic("fra: nil CEC miter")
	}
	miter := *graph
	if verbose != 0 {
		fmt.Printf("Original miter:   Nodes = %6d.\n", aig.Aig_ManNodeNum(miter))
	}
	status := Fra_FraigMiterStatus(miter)
	if status == 0 {
		miter.Data = make([]int32, miter.InputCount())
		return status
	}
	status = Fra_FraigSat(miter, 600, 0, 0, 0, 0, 1, 0, 0, 0)
	if status >= 0 {
		return status
	}
	var err error
	miter, err = dar.Dar_ManRwsat(miter, true, false)
	if err != nil || miter == nil {
		return -1
	}
	var params Fra_Par_t
	Fra_ParamsDefault(&params)
	params.nBTLimitNode, params.nBTLimitMiter = 2, 300
	params.fDontShowBar, params.fProve = 1, 1
	for i := 0; i < 6; i++ {
		if Aig_ManCountXors(miter)*30 > aig.Aig_ManNodeNum(miter)+300 {
			miter = dar.Dar_ManBalanceXor(miter, true, false, false)
		}
		miter = Fra_FraigPerform(miter, &params)
		if miter == nil {
			return -1
		}
		status = Fra_FraigMiterStatus(miter)
		if status >= 0 {
			break
		}
		miter, err = dar.Dar_ManRewriteDefault(miter)
		if err != nil || miter == nil {
			return -1
		}
		status = Fra_FraigMiterStatus(miter)
		if status >= 0 {
			break
		}
		params.nBTLimitNode *= 8
		params.nBTLimitMiter *= 2
	}
	if status == -1 {
		status = Fra_FraigSat(miter, int64(conflictLimit), 0, 0, 0, 0, 1, 0, 0, 0)
	}
	*graph = miter
	return status
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
