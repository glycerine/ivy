package io

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
)

type CexServices struct {
	Extend   func(*aig.Graph, *util.Abc_Cex_t) *util.Abc_Cex_t
	Minimize func(*aig.Graph, *util.Abc_Cex_t, bool, bool, bool, bool, bool) *util.Abc_Cex_t
}

var cexServices struct {
	sync.RWMutex
	value CexServices
}

func RegisterCexServices(value CexServices) {
	cexServices.Lock()
	cexServices.value = value
	cexServices.Unlock()
}

func Abc_NtkCheckSpecialPi(n *abc.Network) int32 {
	for _, pi := range n.ObjectsOfKind(abc.ObjectPI) {
		if abc.Abc_ObjName(pi) == "_abc_190121_abc_" {
			return 1
		}
	}
	return 0
}

func writeOneCex(w io.Writer, n *abc.Network, cex *util.Abc_Cex_t, flags map[rune]bool) error {
	if cex == nil || cex.IFrame < 0 || cex.NPis != abc.Abc_NtkPiNum(n) || cex.NRegs != abc.Abc_NtkLatchNum(n) {
		return fmt.Errorf("io: counterexample dimensions do not match network")
	}
	if int64(cex.NRegs)+int64(cex.NPis)*(int64(cex.IFrame)+1) > int64(len(cex.PData))*32 {
		return fmt.Errorf("io: truncated counterexample")
	}
	pis, latches := n.ObjectsOfKind(abc.ObjectPI), n.ObjectsOfKind(abc.ObjectLatch)
	cexServices.RLock()
	services := cexServices.value
	cexServices.RUnlock()
	if flags['f'] {
		if services.Extend == nil {
			return fmt.Errorf("io: counterexample extension is not registered")
		}
		extended := services.Extend(abc.Abc_NtkToDar(n, 0, 1), cex)
		if extended == nil {
			return fmt.Errorf("io: counterexample extension failed")
		}
		cis := append(append([]*abc.Object(nil), pis...), n.ObjectsOfKind(abc.ObjectBO)...)
		for frame := int32(0); frame <= cex.IFrame; frame++ {
			for i, ci := range cis {
				bit := util.Abc_InfoHasBit(extended.PData, int32(len(cis))*frame+int32(i))
				fmt.Fprintf(w, "%s@%d=%d ", abc.Abc_ObjName(ci), frame, boolBit(bit))
			}
		}
		return nil
	}
	if !flags['n'] {
		for _, latch := range latches {
			fmt.Fprintf(w, "%d", 1-abc.Abc_LatchIsInit0(latch))
		}
		for i := cex.NRegs; i < cex.NBits; i++ {
			if flags['a'] && cex.NPis > 0 && (i-cex.NRegs)%cex.NPis == 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprint(w, boolBit(util.Abc_InfoHasBit(cex.PData, i)))
		}
		return nil
	}
	var care *util.Abc_Cex_t
	if flags['m'] {
		if services.Minimize == nil {
			return fmt.Errorf("io: counterexample minimizer is not registered")
		}
		care = services.Minimize(abc.Abc_NtkToDar(n, 0, 1), cex, flags['o'], flags['u'], flags['e'], flags['c'], flags['v'])
	}
	outputs := n.Outputs()
	if cex.IPo < 0 || int(cex.IPo) >= len(outputs) {
		return fmt.Errorf("io: counterexample output is out of range")
	}
	fmt.Fprintf(w, "# FALSIFYING OUTPUTS: %s\n# COUNTEREXAMPLE LENGTH: %d\n", abc.Abc_ObjName(outputs[cex.IPo]), cex.IFrame+1)
	if flags['z'] && Abc_NtkCheckSpecialPi(n) != 0 {
		pivot := 0
		for abc.Abc_ObjName(pis[pivot]) != "_abc_190121_abc_" {
			pivot++
		}
		nX := 0
		for _, pi := range pis[pivot+1:] {
			if strings.HasPrefix(abc.Abc_ObjName(pi), "x") {
				nX++
			}
		}
		idx := 0
		for _, pi := range pis[pivot+1:] {
			name := abc.Abc_ObjName(pi)
			if len(name) == 0 {
				continue
			}
			if name[0] != 'x' {
				fmt.Fprintf(w, "%s@0=%c\n", name[1:], name[0])
				continue
			}
			pos := int32(pivot - nX + idx)
			idx++
			if care == nil || util.Abc_InfoHasBit(care.PData, care.NRegs+pos) {
				fmt.Fprintf(w, "%s@0=%d\n", name[1:], boolBit(util.Abc_InfoHasBit(cex.PData, cex.NRegs+pos)))
			}
		}
		pis = pis[:pivot-nX]
	} else {
		for _, latch := range latches {
			fmt.Fprintf(w, "%s@0=%d\n", abc.Abc_ObjName(abc.Abc_ObjFanout0(latch)), 1-abc.Abc_LatchIsInit0(latch))
		}
	}
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		for i, pi := range pis {
			bit := cex.NRegs + cex.NPis*frame + int32(i)
			if care == nil || util.Abc_InfoHasBit(care.PData, care.NRegs+care.NPis*frame+int32(i)) {
				fmt.Fprintf(w, "%s@%d=%d\n", abc.Abc_ObjName(pi), frame, boolBit(util.Abc_InfoHasBit(cex.PData, bit)))
			}
		}
	}
	return nil
}
func boolBit(value bool) int {
	if value {
		return 1
	}
	return 0
}

func IoCommandWriteCex(frame *abcmain.Frame, args []string) int32 {
	flags, filename, ok := ioReadOptions(frame, args, "write_cex", "snmueocafzv")
	if !ok {
		return 1
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Out, "Empty network.")
		return 0
	}
	_, _, snapshot := frame.VerificationResult()
	multiple := abcmain.Abc_FrameReadCexVec(frame)
	if n.Model == nil && snapshot == nil && multiple == nil {
		fmt.Fprintln(frame.Out, "Counter-example is not available.")
		return 0
	}
	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	defer f.Close()
	if snapshot != nil || multiple != nil {
		if snapshot != nil {
			cex, e := frameCounterexample(snapshot)
			if e == nil {
				e = writeOneCex(f, n, cex, flags)
			}
			if e != nil {
				fmt.Fprintln(frame.Err, e)
				return 1
			}
		} else {
			for i, value := range multiple {
				if value == nil {
					continue
				}
				cex, ok := value.(*util.Abc_Cex_t)
				if !ok {
					fmt.Fprintln(frame.Err, "io: invalid counterexample entry")
					return 1
				}
				fmt.Fprintf(f, "#\n#\n# CEX for output %d\n#\n", i)
				if err := writeOneCex(f, n, cex, flags); err != nil {
					fmt.Fprintln(frame.Err, err)
					return 1
				}
			}
		}
		fmt.Fprintln(f, "# DONE")
	} else {
		for i, pi := range n.ObjectsOfKind(abc.ObjectPI) {
			if i >= len(n.Model) {
				fmt.Fprintln(frame.Err, "io: incomplete combinational model")
				return 1
			}
			bit := 0
			if n.Model[i] == 1 {
				bit = 1
			}
			if flags['n'] {
				suffix := ""
				if flags['s'] {
					suffix = "@0"
				}
				fmt.Fprintf(f, "%s%s=%d\n", abc.Abc_ObjName(pi), suffix, bit)
			} else {
				fmt.Fprint(f, bit)
			}
		}
		fmt.Fprintln(f)
	}
	return 0
}
