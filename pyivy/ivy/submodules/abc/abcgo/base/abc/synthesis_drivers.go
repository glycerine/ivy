package abc

import (
	"sync"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
)

var collapseGia struct {
	sync.RWMutex
	fn func(*Network) *gia.Gia_Man_t
}

// RegisterCollapseGia supplies the logic-collapse operation from ABCI.
func RegisterCollapseGia(fn func(*Network) *gia.Gia_Man_t) {
	collapseGia.Lock()
	collapseGia.fn = fn
	collapseGia.Unlock()
}

func synthesisCurrent(frame *abcmain.Frame) *Network {
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*Network)
	if !ok || n == nil {
		panic("abc: synthesis produced no network")
	}
	return n
}

func Abc_SopSynthesize(sops *vec.Vec_Ptr_t) *vec.Vec_Wec_t {
	values := make([]string, vec.Vec_PtrSize(sops))
	for i, value := range vec.Vec_PtrArray(sops) {
		var ok bool
		values[i], ok = value.(string)
		if !ok {
			panic("abc: SOP entry is not a string")
		}
	}
	frame := abcmain.Abc_FrameGetGlobalFrame()
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, Abc_NtkCreateFromSops("top", values))
	abcmain.Abc_FrameSetBatchMode(1)
	defer abcmain.Abc_FrameSetBatchMode(0)
	frame.Execute("st; collapse; sop; fx; strash; &get; &ps; &deepsyn -I 4 -J 50 -T 5 -S 111 -t; &ps; &put; map -a")
	return abcSerializeMapped(synthesisCurrent(frame))
}

func Abc_GiaSynthesize(graphs *vec.Vec_Ptr_t, multi *gia.Gia_Man_t) *vec.Vec_Wec_t {
	frame := abcmain.Abc_FrameGetGlobalFrame()
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, synthesisFromGias(graphs, multi))
	abcmain.Abc_FrameSetBatchMode(1)
	defer abcmain.Abc_FrameSetBatchMode(0)
	frame.Execute("compress2rs; dch; map -a;  strash; compress2rs; dch; map -a;  strash; compress2rs; dch; map -a")
	return abcSerializeMapped(synthesisCurrent(frame))
}

func synthesisFromGias(graphs *vec.Vec_Ptr_t, multi *gia.Gia_Man_t) *Network {
	var inputs []NtkLiteralNetwork
	if graphs != nil {
		for _, value := range vec.Vec_PtrArray(graphs) {
			graph, ok := value.(NtkLiteralNetwork)
			if !ok {
				panic("abc: invalid GIA input")
			}
			inputs = append(inputs, graph)
		}
	}
	if multi != nil {
		return Abc_NtkCreateFromGias("top", inputs, multi)
	}
	return Abc_NtkCreateFromGias("top", inputs, nil)
}

func Abc_GiaSynthesizeInter(p *gia.Gia_Man_t) *gia.Gia_Man_t {
	collapseGia.RLock()
	collapse := collapseGia.fn
	collapseGia.RUnlock()
	if collapse == nil {
		panic("abc: logic collapse is not registered")
	}
	graphs := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(graphs, p)
	frame := abcmain.Abc_FrameGetGlobalFrame()
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, synthesisFromGias(graphs, nil))
	frame.Execute("balance; collapse; muxes; strash; dc2")
	return collapse(synthesisCurrent(frame))
}

func Abc_SopSynthesizeOne(sop string, collapse int32) *gia.Gia_Man_t {
	if len(sop) == 3 {
		if sop[1] != '0' && sop[1] != '1' {
			panic("abc: invalid constant SOP")
		}
		p := gia.Gia_ManStart(1)
		p.SetIdentity("top", "")
		lit := int32(0)
		if sop[1] == '1' {
			lit = 1
		}
		gia.Gia_ManAppendCo(p, lit)
		return p
	}
	frame := abcmain.Abc_FrameGetGlobalFrame()
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, Abc_NtkCreateFromSops("top", []string{sop}))
	abcmain.Abc_FrameSetBatchMode(1)
	defer abcmain.Abc_FrameSetBatchMode(0)
	if collapse != 0 {
		frame.Execute("clp; sop")
	}
	frame.Execute("fx; strash; balance; dc2")
	return gia.Gia_ManFromAig(Abc_NtkStrashToGia(synthesisCurrent(frame)))
}
