package if_

import (
	"fmt"
	"io"
	"os"

	"github.com/glycerine/abcgo/misc/vec"
)

// ifDsdDumpTruth writes the source's most-significant-digit-first hex format.
// A prime's formal-input table can be shorter than the requested support width;
// repeat it for the additional independent variables instead of reading past it.
func ifDsdDumpTruth(out io.Writer, truth []word, width int32) {
	// The original hex printer asserts width>=2. One digit also represents the
	// reserved constant/variable records in zero/one-input Go managers safely.
	width = max(width, 2)
	fmt.Fprint(out, "0x")
	if width < 6 {
		bits := uint(1) << uint(width)
		fmt.Fprintf(out, "%0*X", int(bits)/4, truth[0]&((word(1)<<bits)-1))
	} else {
		for i := dsdWordCount(width) - 1; i >= 0; i-- {
			fmt.Fprintf(out, "%016X", truth[i%len(truth)])
		}
	}
	fmt.Fprintln(out)
}

func If_DsdManDumpDsd(p *If_DsdMan_t, support int32) {
	const name = "tts_nondsd.txt"
	out, err := os.Create(name)
	if err != nil {
		fmt.Printf("Cannot open file \"%s\".\n", name)
		return
	}
	defer out.Close()
	for arity := int32(3); arity <= p.nVars; arity++ {
		seen := make([]bool, vec.Vec_MemEntryNum(p.vTtMem[arity]))
		for _, entry := range Vec_PtrArray(&p.vObjs) {
			o := entry.(*If_DsdObj_t)
			// IDs are local to each prime arity. The C loop omits this arity test,
			// which can conflate different tables or index beyond its seen bitmap.
			if o.Type != IF_DSD_PRIME || o.nFans != arity || support != 0 && o.nSupp != support {
				continue
			}
			id := If_DsdObjTruthId(p, o)
			if seen[id] {
				continue
			}
			seen[id] = true
			width := arity
			if support != 0 {
				width = max(support, 6)
			}
			ifDsdDumpTruth(out, If_DsdObjTruth(p, o), width)
		}
	}
}

func If_DsdManDumpAll(p *If_DsdMan_t, support int32) {
	const name = "tts_all.txt"
	out, err := os.Create(name)
	if err != nil {
		fmt.Printf("Cannot open file \"%s\".\n", name)
		return
	}
	defer out.Close()
	for id, entry := range Vec_PtrArray(&p.vObjs) {
		o := entry.(*If_DsdObj_t)
		if support != 0 && o.nSupp != support {
			continue
		}
		width := p.nVars
		if support != 0 {
			width = max(support, 6)
		}
		ifDsdDumpTruth(out, If_DsdManComputeTruth(p, int32(id)<<1, nil), width)
	}
}
