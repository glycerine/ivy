package sfm

import "github.com/glycerine/abcgo/map/mio"

// Sfm_Mit_t and its operations mirror the upstream sfmMit.c placeholders.
// ABC's alternate timing manager is not implemented in that source file.
type Sfm_Mit_t struct{ Value1, Value2, Value3, Value4 int32 }

func Sfm_MitStart(_ *mio.Mio_Library_t, _, _ any, _ *Sfm_Ntk_t, _ int32) *Sfm_Mit_t { return nil }
func Sfm_MitStop(_ *Sfm_Mit_t)                                                      {}
func Sfm_MitReadNtkDelay(_ *Sfm_Mit_t) int32                                        { return 0 }
func Sfm_MitReadNtkMinSlack(_ *Sfm_Mit_t) int32                                     { return 0 }
func Sfm_MitReadObjDelay(_ *Sfm_Mit_t, _ int32) int32                               { return 0 }
func Sfm_MitTransferLoad(_ *Sfm_Mit_t, _, _ int)                                    {}
func Sfm_MitTimingGrow(_ *Sfm_Mit_t)                                                {}
func Sfm_MitUpdateLoad(_ *Sfm_Mit_t, _ []int, _ int32)                              {}
func Sfm_MitUpdateTiming(_ *Sfm_Mit_t, _ []int)                                     {}
func Sfm_MitSortArrayByArrival(_ *Sfm_Mit_t, _ []int, _ int32) int32                { return 0 }
func Sfm_MitPriorityNodes(_ *Sfm_Mit_t, _ []int, _ int32) int32                     { return 0 }
func Sfm_MitNodeIsNonCritical(_ *Sfm_Mit_t, _, _ int) int32                         { return 0 }
func Sfm_MitEvalRemapping(_ *Sfm_Mit_t, _ []int, _ int, _, _ []int, _ *mio.Mio_Gate_t, _ string, _ *mio.Mio_Gate_t, _ string) int32 {
	return 0
}
