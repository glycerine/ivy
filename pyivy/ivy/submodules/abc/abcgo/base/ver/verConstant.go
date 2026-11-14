// Binary constant parsing from verCore.c by Alan Mishchenko (2006).
package ver

import (
	"github.com/glycerine/abcgo/misc/vec"
	"strconv"
	"strings"
)

// Ver_ParseConstant stores bits in MSB-to-LSB order. As in C, x and 0
// both use the null pointer representation and 1 uses the integer value one.
func Ver_ParseConstant(p *Ver_Man_t, word string) int32 {
	if len(word) == 0 || word[0] < '1' || len(word) > 1 && word[1] > '9' {
		panic("ver: invalid binary constant width")
	}
	digits := 0
	for digits < len(word) && word[digits] >= '0' && word[digits] <= '9' {
		digits++
	}
	width, _ := strconv.Atoi(word[:digits])
	quote := strings.IndexByte(word, '\'')
	fail := func(message string) int32 {
		clear(p.sError[:])
		copy(p.sError[:], message)
		Ver_ParsePrintErrorMessage(p)
		return 0
	}
	if quote < 0 {
		return fail("Cannot find symbol ' in the constant.")
	}
	rest := word[quote+1:]
	if len(rest) == 0 || rest[0] != 'b' {
		return fail("Currently can only handle binary constants.")
	}
	rest = rest[1:]
	vec.Vec_PtrClear(p.vNames)
	for i := 0; i < width; i++ {
		if i >= len(rest) || rest[i] != '0' && rest[i] != '1' && rest[i] != 'x' {
			return fail("Having problem parsing the binary constant.")
		}
		if rest[i] == '1' {
			vec.Vec_PtrPush(p.vNames, int32(1))
		} else {
			vec.Vec_PtrPush(p.vNames, nil)
		}
	}
	return 1
}
