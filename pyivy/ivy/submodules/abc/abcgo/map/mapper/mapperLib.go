package mapper

import (
	"fmt"
	"unsafe"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	superpkg "github.com/glycerine/abcgo/map/super"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
)

func Map_SuperLibCreate(genlib *mio.Mio_Library_t, text *vec.Vec_Str_t, path, exclude string, algorithm, verbose int32) *Map_SuperLib_t {
	if genlib == nil && (text != nil || algorithm != 0) {
		return nil
	}
	lib := &Map_SuperLibStruct_t_{pName: path, fVerbose: verbose}
	lib.mmSupers = extra.Extra_MmFixedStart(int32(unsafe.Sizeof(Map_SuperStruct_t_{})))
	lib.mmEntries = extra.Extra_MmFixedStart(int32(unsafe.Sizeof(Map_HashEntryStruct_t_{})))
	lib.mmForms = extra.Extra_MmFlexStart()
	Map_MappingSetupTruthTables(lib.uTruths[:])
	lib.tTableC, lib.tTable = Map_SuperTableCreate(lib), Map_SuperTableCreate(lib)
	start := Abc_Clock()
	loaded := int32(0)
	if text != nil {
		loaded = Map_LibraryReadFileTreeStr(lib, genlib, text, path)
		if loaded != 0 {
			loaded = Map_LibraryDeriveGateInfo(lib, nil)
		}
	} else if algorithm != 0 {
		loaded = Map_LibraryReadTree(lib, genlib, path, exclude)
	} else if exclude == "" {
		loaded = Map_LibraryRead(lib, path)
	} else {
		fmt.Println("Error: Exclude file support not present for old format. Stop.")
	}
	if loaded == 0 || lib.nVarsMax <= 0 || len(lib.ppSupers) == 0 {
		Map_SuperLibFree(lib)
		return nil
	}
	if verbose != 0 {
		fmt.Printf("Loaded %d unique %d-input supergates from %q. %.2f sec\n", lib.nSupersReal, lib.nVarsMax, path, float64(Abc_Clock()-start)/1e9)
	}
	lib.pGateInv = mio.Mio_LibraryReadInv(lib.pGenlib)
	lib.tDelayInv.Rise = mio.Mio_LibraryReadDelayInvRise(lib.pGenlib)
	lib.tDelayInv.Fall = mio.Mio_LibraryReadDelayInvFall(lib.pGenlib)
	lib.tDelayInv.Worst = max(lib.tDelayInv.Rise, lib.tDelayInv.Fall)
	lib.AreaInv, lib.AreaBuf = mio.Mio_LibraryReadAreaInv(lib.pGenlib), mio.Mio_LibraryReadAreaBuf(lib.pGenlib)
	lib.pSuperInv = &Map_SuperStruct_t_{Num: -1, nGates: 1, nFanins: 1, nFanLimit: 10, pRoot: lib.pGateInv, Area: lib.AreaInv, tDelayMax: lib.tDelayInv}
	lib.pSuperInv.pFanins[0] = lib.ppSupers[0]
	lib.pSuperInv.tDelaysR[0] = Map_Time_t{Rise: MAP_NO_VAR, Fall: lib.tDelayInv.Rise}
	lib.pSuperInv.tDelaysF[0] = Map_Time_t{Rise: lib.tDelayInv.Fall, Fall: MAP_NO_VAR}
	return lib
}

func Map_SuperLibFree(lib *Map_SuperLib_t) {
	if lib == nil {
		return
	}
	if lib.tTableC != nil {
		Map_SuperTableFree(lib.tTableC)
	}
	if lib.tTable != nil {
		Map_SuperTableFree(lib.tTable)
	}
	if lib.pGenlib != nil {
		if current, _ := abcmain.Abc_FrameGetGlobalFrame().Resource("pLibGen").(*mio.Mio_Library_t); current != lib.pGenlib {
			mio.Mio_LibraryDelete(lib.pGenlib)
		}
	}
	extra.Extra_MmFixedStop(lib.mmSupers)
	extra.Extra_MmFixedStop(lib.mmEntries)
	extra.Extra_MmFlexStop(lib.mmForms)
	lib.pGenlib, lib.ppSupers, lib.pSuperInv = nil, nil, nil
	lib.tTableC, lib.tTable = nil, nil
}

func Map_SuperLibDeriveFromGenlib(genlib *mio.Mio_Library_t, verbose int32) int32 {
	if genlib == nil {
		return 0
	}
	text := superpkg.Super_PrecomputeStr(genlib, 5, 1, 100000000, 10000000, 10000000, 100, 1, 0)
	if text == nil {
		return 0
	}
	path := mio.Mio_LibraryReadName(genlib) + ".super"
	lib := Map_SuperLibCreate(genlib, text, path, "", 1, 0)
	vec.Vec_StrFree(text)
	if lib == nil {
		return 0
	}
	frame := abcmain.Abc_FrameGetGlobalFrame()
	if old, ok := frame.TakeResource("pLibSuper").(*Map_SuperLib_t); ok {
		Map_SuperLibFree(old)
	}
	frame.SetResource("pLibSuper", lib)
	return 1
}

func Map_SuperLibDeriveFromGenlib2(genlib *mio.Mio_Library_t, verbose int32) int32 {
	if genlib == nil {
		return 0
	}
	path := mio.Mio_LibraryReadName(genlib) + ".super"
	if err := superpkg.Super_Precompute(genlib, 5, 1, 100000000, 10000000, 10000000, 100, 1, verbose, path); err != nil {
		return 0
	}
	lib := Map_SuperLibCreate(genlib, nil, path, "", 1, verbose)
	if lib == nil {
		return 0
	}
	frame := abcmain.Abc_FrameGetGlobalFrame()
	if old, ok := frame.TakeResource("pLibSuper").(*Map_SuperLib_t); ok {
		Map_SuperLibFree(old)
	}
	frame.SetResource("pLibSuper", lib)
	return 1
}
