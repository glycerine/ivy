package abc

// Four-variable SOP covers and cube emission from sat/cnf/cnfData.c and cnfWrite.c.

import (
	_ "embed"
	"slices"
	"sync"
)

//go:embed cnf_sop4.txt
var cnfSOP4Encoded string

const cnfSOPAlphabet = "!#&()*+,-.0123456789:;<=>?ABCDEFGHIJKLMNOPQRSTUVWXYZ[]abcdefghijklmnopqrstuvwxyz|"

type cnfSOP4Table struct {
	offsets []uint32
	cubes   []uint8
}

var cnfSOPOnce sync.Once
var cnfSOPTable cnfSOP4Table
var cnfSOPError error

// decodeCNFSOP4 preserves native table order and reverses ternary digit order
// into packed ISOP pairs, matching Cnf_SopConvertToVector. Padding after the
// final function is ignored, as in Cnf_ReadMsops.
func decodeCNFSOP4(data string) (cnfSOP4Table, error) {
	if len(data) > 1<<20 {
		return cnfSOP4Table{}, ErrLimit
	}
	var alphabet [256]int16
	for i := range alphabet {
		alphabet[i] = -1
	}
	for i, c := range []byte(cnfSOPAlphabet) {
		alphabet[c] = int16(i)
	}
	r := cnfSOP4Table{offsets: make([]uint32, 65537)}
	row := 1
	start := 0
	for i := 0; i < len(data) && row < 65536; i++ {
		c := data[i]
		if c == ' ' {
			if len(r.cubes) == start {
				return cnfSOP4Table{}, ErrDimension
			}
			r.offsets[row+1] = uint32(len(r.cubes))
			row++
			start = len(r.cubes)
			continue
		}
		value := int(alphabet[c])
		if value < 0 || len(r.cubes)-start >= 16 {
			return cnfSOP4Table{}, ErrDimension
		}
		var cube uint8
		for b := 0; b < 4; b++ {
			digit := value % 3
			value /= 3
			literal := uint8(0)
			if digit == 0 {
				literal = 1
			} else if digit == 1 {
				literal = 2
			}
			cube = cube<<2 | literal
		}
		r.cubes = append(r.cubes, cube)
	}
	if row != 65536 {
		return cnfSOP4Table{}, ErrDimension
	}
	return r, nil
}

// CNFSOP4Cover returns an owned native cover in packed two-bit ISOP form.
// Pair b describes variable b: 0 is absent, 1 requires false, 2 requires true.
// The zero function has no cubes; the one function has one empty cube. This is
// ABC's exact table, not a newly minimized or differently ordered cover.
func CNFSOP4Cover(truth uint16) ([]uint8, error) {
	cnfSOPOnce.Do(func() { cnfSOPTable, cnfSOPError = decodeCNFSOP4(cnfSOP4Encoded) })
	if cnfSOPError != nil {
		return nil, cnfSOPError
	}
	return slices.Clone(cnfSOPTable.cubes[cnfSOPTable.offsets[int(truth)]:cnfSOPTable.offsets[int(truth)+1]]), nil
}

// CNFISOPCubeLiterals emits the negated cube used as clause literals by native
// Cnf_IsopWriteCube. Pair 1 emits a positive variable, pair 2 a negative one;
// 0 and 3 are ignored. High pairs beyond variables are ignored. Only referenced
// variable IDs are validated. A uint32 cube has at most sixteen pairs.
func CNFISOPCubeLiterals(cube uint32, variables []SatVar) ([]SatLit, error) {
	if len(variables) > 16 {
		return nil, ErrDimension
	}
	var result []SatLit
	for _, v := range variables {
		pair := cube & 3
		cube >>= 2
		if pair != 1 && pair != 2 {
			continue
		}
		l, e := MakeSatLit(v, pair == 2)
		if e != nil {
			return nil, e
		}
		result = append(result, l)
	}
	return result, nil
}
func CNFISOPLiteralCount(cubes []uint32, variables int) (int, error) {
	if variables < 0 || variables > 16 {
		return 0, ErrDimension
	}
	total := 0
	for _, cube := range cubes {
		for b := 0; b < variables; b++ {
			pair := cube & 3
			cube >>= 2
			if pair == 1 || pair == 2 {
				if total == int(^uint32(0)>>1) {
					return 0, ErrLimit
				}
				total++
			}
		}
	}
	return total, nil
}
