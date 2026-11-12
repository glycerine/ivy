package gia

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/misc/extra"
)

// giaManShow keeps the adder recognizer and viewer at the package boundary.
func giaManShow(p *Gia_Man_t, bold *Vec_Int_t, adders, fullOnly, path int32, compute func(*Gia_Man_t, **Vec_Int_t, int32) *Vec_Int_t, show func(string)) {
	var adds, xors *Vec_Int_t
	if adders != 0 {
		adds = compute(p, &xors, 0)
	}
	filename := extra.Extra_FileNameGenericAppend(p.pName, ".dot")
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Cannot open the intermediate file %q.\n", filename)
		return
	}
	if err = f.Close(); err != nil {
		fmt.Printf("Cannot close the intermediate file %q: %v\n", filename, err)
		return
	}
	if path != 0 {
		Gia_ShowPath(p, filename)
	} else if adders != 0 {
		Gia_ShowProcess(p, filename, bold, adds, xors, fullOnly)
	} else {
		Gia_WriteDotAigSimple(p, filename, bold)
	}
	show(filename)
}
