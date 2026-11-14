// Parser manager lifecycle and network conversion from verCore.c.
package ver

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
)

func Ver_ParseStart(filename string, library *abc.Design) *Ver_Man_t {
	reader := Ver_StreamAlloc(filename)
	if reader == nil {
		return nil
	}
	return &Ver_Man_t{pFileName: filename, pReader: reader, Output: os.Stdout, vNames: vec.Vec_PtrAlloc(100), vStackFn: vec.Vec_PtrAlloc(100), vStackOp: vec.Vec_IntAlloc(100), vPerm: vec.Vec_IntAlloc(100), pDesign: abc.Abc_DesCreate(filename), pGateLib: library, pGenlib: abcmain.Abc_FrameReadLibGen()}
}
func Ver_ParseStop(p *Ver_Man_t) {
	if p.pProgress != nil {
		extra.Extra_ProgressBarStop(p.pProgress)
	}
	Ver_StreamFree(p.pReader)
	vec.Vec_PtrFree(p.vNames)
	vec.Vec_PtrFree(p.vStackFn)
	vec.Vec_IntFree(p.vStackOp)
	vec.Vec_IntFree(p.vPerm)
}
func Ver_ParseConvertNetwork(p *Ver_Man_t, n *abc.Network, mapped int32) int32 {
	target := abc.ABC_FUNC_AIG
	var manager any = p.pDesign.FunctionManager
	if mapped != 0 {
		target = abc.ABC_FUNC_MAP
		manager = p.pGenlib
	}
	if n.Function == uint8(abc.ABC_FUNC_BLACKBOX) {
		if n.FunctionManager != nil {
			panic("ver: blackbox has a function manager")
		}
		n.Function = uint8(target)
		n.FunctionManager = manager
		return 1
	}
	if n.Function != uint8(target) {
		return verRangeError(p, fmt.Sprintf("The network %s appears to have both gates and assign statements. Currently such network are not allowed. One way to fix this problem might be to replace assigns by buffers from the library.", n.Name))
	}
	return 1
}
