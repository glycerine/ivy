package if_

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/abcgo/aig/gia"
	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManFromIfLogic reconstructs a mapped IF network and attaches its LUT,
// optional packing, and programmable-cell configuration metadata to the GIA.
func Gia_ManFromIfLogic(p *If_Man_t) *Gia_Man_t {
	pars := p.pPars
	assert(pars.FDeriveLuts == 0 || pars.FTruth != 0, "LUT derivation requires truth tables")
	offsets := vec.Vec_IntStart(int(If_ManObjNum(p)))
	records := vec.Vec_IntStart(1)
	var packing, configs *Vec_Int_t
	var configText *Vec_Str_t
	if pars.FDeriveLuts != 0 && (pars.PLutStruct != "" || pars.FEnableCheck75 != 0 || pars.FEnableCheck75u != 0 || pars.FEnableCheck07 != 0) {
		packing = vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	}
	if pars.FUseDsdTune != 0 {
		count := (If_DsdManTtBitNum(p.pIfDsdMan) + If_DsdManPermBitNum(p.pIfDsdMan) + 32) / 32
		configs = vec.Vec_IntAllocArrayCopy([]int32{0, count}, 2)
		configText = vec.Vec_StrAlloc(1000)
	}
	dst := Gia_ManStart(If_ManObjNum(p))
	leaves, temp, cover, lits := Vec_IntAlloc(16), Vec_IntAlloc(16), Vec_IntAlloc(1<<16), Vec_IntAlloc(1000)
	defer Vec_IntFree(leaves)
	defer Vec_IntFree(temp)
	defer Vec_IntFree(cover)
	defer Vec_IntFree(lits)
	defer Vec_IntFree(records)
	If_ManCleanCutData(p)
	for i := int32(0); i < If_ManObjNum(p); i++ {
		obj := If_ManObj(p, i)
		if obj.nRefs == 0 && If_ObjIsTerm(obj) == 0 {
			continue
		}
		switch obj.Type {
		case IF_AND:
			cut := If_ObjCutBest(obj)
			if pars.FUseTtPerm == 0 && pars.FDelayOpt == 0 && pars.FDelayOptLut == 0 && pars.FDsdBalance == 0 && pars.PLutStruct == "" && pars.FUserRecLib == 0 && pars.FUserSesLib == 0 && pars.NGateSize == 0 && pars.FEnableCheck75 == 0 && pars.FEnableCheck75u == 0 && pars.FEnableCheck07 == 0 && pars.FUseDsdTune == 0 && pars.FUseCofVars == 0 && pars.FUseAndVars == 0 {
				If_CutRotatePins(p, cut)
			}
			vec.Vec_IntClear(leaves)
			for k := int32(0); k < int32(cut.nLeaves); k++ {
				Vec_IntPush(leaves, If_CutLeaf(p, cut, k).iCopy)
			}
			switch {
			case pars.FUseDsd != 0 && pars.PLutStruct != "":
				if pars.FDeriveLuts != 0 {
					obj.iCopy = Gia_ManFromIfLogicFindLut(p, dst, cut, nil, leaves, lits, cover, offsets, records, packing)
				} else {
					if If_CutTruthW(p, cut) == nil {
						panic("if: missing cut truth")
					}
					obj.iCopy = Gia_ManFromIfLogicCreateLut(dst, p.puTempW[:p.nTruth6Words[cut.nLeaves]], leaves, cover, offsets, records)
				}
				obj.iCopy ^= int32(cut.fCompl)
			case pars.FUseAndVars != 0 && pars.FDeriveLuts != 0 && int32(cut.nLeaves) > pars.NLutSize/2 && (pars.FUseCofVars == 0 || Vec_IntEntry(p.vTtDecs[cut.nLeaves], int(Abc_Lit2Var(cut.iCutFunc))) != 0):
				obj.iCopy = Gia_ManFromIfLogicAndVars(dst, p, cut, leaves, temp, cover, offsets, records)
			case pars.FUseCofVars != 0 && pars.FDeriveLuts != 0 && int32(cut.nLeaves) > pars.NLutSize/2:
				obj.iCopy = Gia_ManFromIfLogicCofVars(dst, p, cut, leaves, temp, cover, offsets, records)
			case pars.FDeriveLuts != 0 && pars.FTruth != 0 || pars.FUseDsd != 0 || pars.FUseTtPerm != 0 || pars.pFuncCell2 != nil:
				if If_CutTruthW(p, cut) == nil {
					panic("if: missing cut truth")
				}
				truth := p.puTempW[:p.nTruth6Words[cut.nLeaves]]
				if pars.FUseTtPerm != 0 {
					for k := int32(0); k < int32(cut.nLeaves); k++ {
						if If_CutLeafBit(cut, k) != 0 {
							abcutil.Abc_TtFlip(truth, abcutil.Abc_TtWordNum(int32(cut.nLeaves)), k)
						}
					}
				}
				obj.iCopy = Gia_ManFromIfLogicNode(p, dst, i, leaves, temp, truth, pars.PLutStruct, cover, offsets, records, packing, ifTuneBoolInt(pars.FEnableCheck75 != 0 || pars.FEnableCheck75u != 0), pars.FEnableCheck07)
				if obj.iCopy < 0 {
					Gia_ManStop(dst)
					return nil
				}
				obj.iCopy ^= int32(cut.fCompl)
				if configs != nil && Vec_IntSize(leaves) > 1 && obj.iCopy > 1 && gia.Gia_ObjIsCi(Gia_ManObj(dst, obj.iCopy>>1)) == 0 {
					Gia_ManFromIfGetConfig(configs, p, cut, obj.iCopy, configText)
				}
			default:
				obj.iCopy = Gia_ManNodeIfToGia(dst, p, obj, leaves, 0)
				if obj.iCopy < 0 {
					Gia_ManStop(dst)
					return nil
				}
				vec.Vec_IntSetEntry(offsets, int(obj.iCopy>>1), int32(Vec_IntSize(records)))
				Vec_IntPush(records, int32(Vec_IntSize(leaves)))
				for _, lit := range Vec_IntArray(leaves) {
					assert(lit>>1 < obj.iCopy>>1, "cut leaf precedes LUT root")
					Vec_IntPush(records, lit>>1)
				}
				Vec_IntPush(records, obj.iCopy>>1)
			}
		case IF_CI:
			obj.iCopy = Gia_ManAppendCi(dst)
		case IF_CO:
			obj.iCopy = Gia_ManAppendCo(dst, obj.pFanin0.iCopy^int32(obj.fCompl0))
		case IF_CONST1:
			obj.iCopy = 1
			vec.Vec_IntWriteEntry(offsets, 0, int32(Vec_IntSize(records)))
			Vec_IntPush(records, 0)
			Vec_IntPush(records, 0)
		default:
			panic("if: unknown mapping object")
		}
	}
	count := int(Gia_ManObjNum(dst))
	if Vec_IntSize(offsets) > count {
		vec.Vec_IntShrink(offsets, count)
	} else {
		vec.Vec_IntFillExtra(offsets, count, 0)
	}
	for i, offset := range Vec_IntArray(offsets) {
		if offset > 0 {
			vec.Vec_IntWriteEntry(offsets, i, offset+int32(count))
		}
	}
	vec.Vec_IntAppend(offsets, records)
	cell := ""
	if configs != nil {
		cell = If_DsdManGetCellStr(p.pIfDsdMan)
	}
	if err := dst.SetMappedData(offsets, packing, configs, cell); err != nil {
		panic(err)
	}
	for i := int32(0); i < Gia_ManCoNum(dst); i++ {
		obj := Gia_ManCo(dst, i)
		fanin := gia.Gia_ObjFanin0(obj)
		assert(gia.Gia_ObjIsAnd(fanin) == 0 || gia.Gia_ObjIsLut(dst, gia.Gia_ObjId(dst, fanin)) != 0, "CO driver is mapped")
	}
	for id := int32(1); id < Gia_ManObjNum(dst); id++ {
		if gia.Gia_ObjIsLut(dst, id) == 0 {
			continue
		}
		for _, fan := range gia.Gia_ObjLutFanins(dst, id) {
			assert(gia.Gia_ObjIsAnd(Gia_ManObj(dst, fan)) == 0 || gia.Gia_ObjIsLut(dst, fan) != 0, "LUT fanin is mapped")
		}
	}
	if configText != nil {
		name := p.pName
		if name == "" {
			name = "nameless_"
		} else if dot, slash := strings.LastIndexByte(name, '.'), strings.LastIndexAny(name, "/\\"); dot > slash {
			name = name[:dot]
		}
		path := name + "_configs.txt"
		if err := os.WriteFile(path, vec.Vec_StrArray(configText), 0666); err != nil {
			fmt.Printf("Cannot open file %q.\n", path)
		} else {
			fmt.Printf("Finished dumping configs into file %q.\n", path)
		}
	}
	return dst
}
