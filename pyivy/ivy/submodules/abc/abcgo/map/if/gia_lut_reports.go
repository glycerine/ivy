package if_

import (
	"fmt"
	"io"

	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/extra"
)

// The GIA source uses Kit_DsdPrint's uppercase prime hex, compact separators
// and F prefix. Keep this diagnostic format local to its source caller.
func giaWriteSourceDsd(w io.Writer, ntk *kit.Kit_DsdNtk_t, lit int32) {
	if lit&1 != 0 {
		fmt.Fprint(w, "!")
	}
	id := lit >> 1
	obj := kit.Kit_DsdNtkObj(ntk, id)
	if obj == nil {
		fmt.Fprintf(w, "%c", 'a'+id)
		return
	}
	if obj.Type == kit.KIT_DSD_CONST1 {
		fmt.Fprint(w, "Const1")
		return
	}
	separator := ","
	switch obj.Type {
	case kit.KIT_DSD_AND:
		separator = "*"
	case kit.KIT_DSD_XOR:
		separator = "+"
	case kit.KIT_DSD_PRIME:
		extra.Extra_PrintHex2(w, kit.Kit_DsdObjTruth(obj), kit.Kit_DsdObjFanNum(obj))
	}
	fmt.Fprint(w, "(")
	for index, fan := range kit.Kit_DsdObjForEachFanin(obj) {
		if index != 0 {
			fmt.Fprint(w, separator)
		}
		giaWriteSourceDsd(w, ntk, int32(fan))
	}
	fmt.Fprint(w, ")")
}

func giaWriteLutFailure(w io.Writer, truth []word, n, objectID int32) {
	words := make([]uint32, 2*len(truth))
	for i, value := range truth {
		words[2*i], words[2*i+1] = uint32(value), uint32(value>>32)
	}
	extra.Extra_PrintHex(w, words, n)
	fmt.Fprint(w, "    F = ")
	original := kit.Kit_DsdDecomposeMux(words, n, 8)
	defer kit.Kit_DsdNtkFree(original)
	expanded := kit.Kit_DsdExpand(original)
	defer kit.Kit_DsdNtkFree(expanded)
	giaWriteSourceDsd(w, expanded, kit.Kit_DsdNtkRootLit(expanded))
	kit.Kit_DsdVerify(expanded, words, n)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Node %d is not decomposable. Deriving LUT structures has failed.\n", objectID)
}
