package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"time"
)

func Ga2_ManAbsPrintFrame(p *Ga2_Man_t, frames, conflicts, cexes int32, elapsed int64, final int32) {
	newline := final != 0 && cexes != 0 || p.pPars.FVeryVerbose != 0
	if abcmain.Abc_FrameIsBatchMode() != 0 && !newline {
		return
	}
	p.fUseNewLine = 0
	if newline {
		p.fUseNewLine = 1
	}
	percent := int32(0)
	if p.nMarked > 0 {
		percent = min(100, 100*int32(vec.Vec_IntSize(p.vAbs))/p.nMarked)
	}
	stats := p.pSat.Statistics()
	memory, proof := p.pSat.Memory(false)
	fmt.Printf("%4d :%4d%6d%5d%5d%6d%8d", frames, percent, vec.Vec_IntSize(p.vAbs), vec.Vec_IntSize(p.vValues)-vec.Vec_IntSize(p.vAbs)-1, Ga2_GlaAbsCount(p, 1, 0), Ga2_GlaAbsCount(p, 0, 1), conflicts)
	if cexes == 0 {
		fmt.Printf("%5c", '-')
	} else {
		fmt.Printf("%5d", cexes)
	}
	util.Abc_PrintInt(stats.Variables)
	util.Abc_PrintInt(int32(stats.Clauses))
	util.Abc_PrintInt(int32(stats.Learned))
	fmt.Printf("%9.2f sec%5.0f MB", float64(elapsed)/float64(time.Second), (memory+proof)/(1<<20))
	if newline {
		fmt.Print("\n")
	} else {
		fmt.Print("\r")
	}
}
func Ga2_GlaGetFileName(p *Ga2_Man_t, abstract int32) string {
	if p.pPars.PFileVabs != "" {
		return p.pPars.PFileVabs
	}
	if spec := p.pGia.NetworkSpec(); spec != "" {
		suffix := "_gla.aig"
		if abstract != 0 {
			suffix = "_abs.aig"
		}
		return extra.Extra_FileNameGenericAppend(spec, suffix)
	}
	return "glabs.aig"
}
func Ga2_GlaDumpAbsracted(p *Ga2_Man_t, verbose int32) {
	if p.pPars.FDumpMabs != 0 {
		name := Ga2_GlaGetFileName(p, 0)
		if verbose != 0 {
			fmt.Printf("Dumping miter with abstraction map into file %q...\n", name)
		}
		gia.Gia_ManSetGateClasses(p.pGia, Ga2_ManAbsTranslate(p))
		if err := gia.Gia_AigerWrite(p.pGia, name, 0, 0, 0); err != nil {
			fmt.Printf("Cannot write %s: %v\n", name, err)
		}
		return
	}
	if p.pPars.FDumpVabs != 0 {
		name := Ga2_GlaGetFileName(p, 1)
		if verbose != 0 {
			fmt.Printf("Dumping abstracted model into file %q...\n", name)
		}
		abstract := Gia_ManDupAbsGates(p.pGia, Ga2_ManAbsTranslate(p))
		gia.Gia_ManCleanValue(p.pGia)
		if err := gia.Gia_AigerWrite(abstract, name, 0, 0, 0); err != nil {
			fmt.Printf("Cannot write %s: %v\n", name, err)
		}
		gia.Gia_ManStop(abstract)
		return
	}
	panic("abs: no GLA dump mode")
}
func Gia_Ga2SendAbsracted(p *Ga2_Man_t, verbose int32) {
	if abcmain.Abc_FrameIsBridgeMode() == 0 {
		panic("abs: bridge mode required")
	}
	abstract := Gia_ManDupAbsGates(p.pGia, Ga2_ManAbsTranslate(p))
	gia.Gia_ManCleanValue(p.pGia)
	net, err := gia.Gia_ManToBridgeNetlist(abstract)
	if err != nil {
		panic(err)
	}
	util.Gia_ManToBridgeAbsNetlist(os.Stdout, net, 107)
	gia.Gia_ManStop(abstract)
}
func Gia_Ga2SendCancel(p *Ga2_Man_t, verbose int32) {
	if abcmain.Abc_FrameIsBridgeMode() == 0 {
		panic("abs: bridge mode required")
	}
	util.Gia_ManToBridgeBadAbs(os.Stdout)
}

func Ga2_ManWriteStatus(p *Ga2_Man_t, frames int32) {
	abcmain.Abc_FrameSetStatus(-1)
	abcmain.Abc_FrameSetCex(nil)
	abcmain.Abc_FrameSetNFrames(frames)
	name := p.pPars.PFileVabs
	if name == "" {
		name = "glabs.aig"
	}
	command := fmt.Sprintf("write_status %s", extra.Extra_FileNameGenericAppend(name, ".status"))
	abcmain.Cmd_CommandExecute(abcmain.Abc_FrameGetGlobalFrame(), command)
}

func Ga2_ManDumpInitialMap(p *Ga2_Man_t) {
	previous := gia.Gia_ManGateClasses(p.pGia)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p.pGia)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	for id := int32(0); id < gia.Gia_ManObjNum(p.pGia); id++ {
		obj := gia.Gia_ManObj(p.pGia, id)
		if gia.Gia_ObjIsAnd(obj) != 0 || gia.Gia_ObjIsRo(p.pGia, obj) != 0 {
			vec.Vec_IntWriteEntry(classes, int(id), 1)
		}
	}
	gia.Gia_ManSetGateClasses(p.pGia, classes)
	name := Ga2_GlaGetFileName(p, 0)
	if err := gia.Gia_AigerWrite(p.pGia, name, 0, 0, 0); err != nil {
		fmt.Printf("Cannot write %s: %v\n", name, err)
	}
	gia.Gia_ManSetGateClasses(p.pGia, previous)
	if p.pPars.FVerbose != 0 {
		fmt.Printf("Dumping miter with abstraction map into file %q...\n", name)
	}
}
