// Package cut computes K-feasible combinational and sequential cuts.
package cut

const CUT_SIZE_MIN int32 = 3
const CUT_SIZE_MAX int32 = 12
const CUT_SHIFT int32 = 8
const CUT_MASK int32 = 255

// Cut_Params_t retains the original independently selectable computation modes.
// Nonzero F* values enable the corresponding C option.
type Cut_Params_t struct {
	NVarsMax, NKeepMax, NIdsMax, NBitShift, NCutSet                                                                            int32
	FTruth, FFilter, FSeq, FDrop, FDag, FTree, FGlobal, FLocal, FRecord, FRecordAig, FFancy, FMap, FAdjust, FNpnSave, FVerbose int32
}
type Cut_Cut_t struct {
	Num0, Num1              uint16
	FSimul, FCompl          bool
	NVarsMax                int32
	USign, UCanon0, UCanon1 uint32
	Next                    *Cut_Cut_t
	leaves                  []int32
	truth                   []uint32
	proof                   *Cut_Proof
	owner                   *Cut_Man_t
	live                    bool
}

func (c *Cut_Cut_t) Leaves() []int32         { return c.leaves }
func (c *Cut_Cut_t) Truth() []uint32         { return c.truth }
func Cut_CutReadLeaveNum(c *Cut_Cut_t) int32 { return int32(len(c.leaves)) }
func Cut_CutReadLeaves(c *Cut_Cut_t) []int32 { return c.leaves }
func Cut_CutReadTruth(c *Cut_Cut_t) []uint32 { return c.truth }
func Cut_CutWriteTruth(c *Cut_Cut_t, truth []uint32) {
	copy(c.truth[:Cut_TruthWords(c.NVarsMax)], truth)
}
func DefaultParams(k int32) *Cut_Params_t {
	return &Cut_Params_t{NVarsMax: k, NKeepMax: 1000, FFilter: 1}
}
