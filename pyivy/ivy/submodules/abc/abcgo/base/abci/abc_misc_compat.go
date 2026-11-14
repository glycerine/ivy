package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Extra_UtilStrsav(s string) string { return extra.Extra_UtilStrsav(s) }

func Extra_FileGetSimilarName(wrong, s1, s2, s3, s4, s5 string) string {
	return extra.Extra_FileGetSimilarName(wrong, s1, s2, s3, s4, s5)
}

func Extra_TimeStamp() string { return extra.Extra_TimeStamp() }

func Extra_FileNameGeneric(name string) string { return extra.Extra_FileNameGeneric(name) }

func Abc_Base10Log(n int32) int32 { return util.Abc_Base10Log(uint32(n)) }

func Vec_WrdSize(v *Vec_Wrd_t) int32 { return int32(vec.Vec_WrdSize(v)) }

func Vec_PtrWriteEntry(v *Vec_Ptr_t, i int, value any) {
	vec.Vec_PtrWriteEntry(v, i, value)
}

func Vec_PtrShrink(v *Vec_Ptr_t, size int) { vec.Vec_PtrShrink(v, size) }

// ABC_PRT reports elapsed clock time using ABC's CLOCKS_PER_SEC conversion.
func ABC_PRT(label string, elapsed int64) {
	fmt.Printf("%s = %7.2f sec\n", label, float64(elapsed)/float64(util.CLOCKS_PER_SEC))
}
