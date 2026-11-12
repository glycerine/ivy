package sbd

import (
	"fmt"
	"io"
	"os"

	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
)

func sbdPrintCnf(out io.Writer, cnf *vec.Vec_Str_t) {
	for _, entry := range vec.Vec_StrArray(cnf) {
		if entry == 255 {
			fmt.Fprintln(out)
			continue
		}
		sign := ""
		if entry&1 != 0 {
			sign = "-"
		}
		fmt.Fprintf(out, "%s%d ", sign, entry>>1)
	}
}
func Sbd_PrintCnf(cnf *vec.Vec_Str_t) { sbdPrintCnf(os.Stdout, cnf) }

// Sbd_TruthToCnf encodes output == truth(inputs), using ISOP covers of both
// polarities. Entries are ABC literals; 255 terminates each clause.
func Sbd_TruthToCnf(truth uint64, n int32, cover *vec.Vec_Int_t, cnf *vec.Vec_Str_t) int32 {
	if n < 0 || n > 6 {
		panic("SBD CNF requires 0..6 inputs")
	}
	vec.Vec_StrClear(cnf)
	if truth == 0 || truth == ^uint64(0) {
		literal := byte(2 * n)
		if truth == 0 {
			literal |= 1
		}
		vec.Vec_StrPush(cnf, literal)
		vec.Vec_StrPush(cnf, 255)
		return 1
	}
	if n == 0 {
		panic("nonconstant zero-input truth")
	}
	var count int32
	for polarity := int32(0); polarity < 2; polarity++ {
		value := truth
		if polarity != 0 {
			value = ^truth
		}
		var cubes []uint32
		if kit.Kit_TruthIsop([]uint32{uint32(value), uint32(value >> 32)}, n, &cubes, 0) != 0 {
			panic("SBD ISOP failed")
		}
		vec.Vec_IntClear(cover)
		for _, cube := range cubes {
			vec.Vec_IntPush(cover, int32(cube))
			for v := int32(0); v < n; v++ {
				switch cube >> (2 * v) & 3 {
				case 1:
					vec.Vec_StrPush(cnf, byte(2*v))
				case 2:
					vec.Vec_StrPush(cnf, byte(2*v+1))
				case 3:
					panic("invalid ISOP cube")
				}
			}
			vec.Vec_StrPush(cnf, byte(2*n+polarity))
			vec.Vec_StrPush(cnf, 255)
			count++
		}
	}
	return count
}

// Sbd_TranslateCnf maps formal variables to literals and flips the pivot's
// polarity. As in C, a clause separator starts the following (possibly empty)
// level; callers can distinguish a terminal empty level by its position.
func Sbd_TranslateCnf(result *vec.Vec_Wec_t, cnf *vec.Vec_Str_t, faninMap *vec.Vec_Int_t, pivot int32) {
	vec.Vec_WecClear(result)
	clause := vec.Vec_WecPushLevel(result)
	for _, entry := range vec.Vec_StrArray(cnf) {
		if entry == 255 {
			clause = vec.Vec_WecPushLevel(result)
			continue
		}
		literal := vec.Vec_IntEntry(faninMap, int(entry>>1)) ^ int32(entry&1)
		if literal>>1 == pivot {
			literal ^= 1
		}
		vec.Vec_IntPush(clause, literal)
	}
}
