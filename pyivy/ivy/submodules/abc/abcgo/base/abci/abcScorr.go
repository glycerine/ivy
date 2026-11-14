package abci

import (
	"fmt"
	"os"

	aiggraph "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
	baseabc "github.com/glycerine/abcgo/base/abc"
	abcio "github.com/glycerine/abcgo/base/io"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec"
	"github.com/glycerine/abcgo/proof/ssw"
)

type Tst_Dat_t struct {
	pNetlist     *Abc_Ntk_t
	pAig         *Aig_Man_t
	pGia         *Gia_Man_t
	vId2Name     *Vec_Int_t
	pFileNameOut string
	fFlopOnly    int32
	fFfNdOnly    int32
	fDumpBmc     int32
}

// Abc_NtkMapGiaIntoNameId maps each GIA/AIG object to the source netlist ID
// associated with its net.  Later CI/CO passes intentionally overwrite names
// assigned by the general-net pass, matching ABC's precedence.
func Abc_NtkMapGiaIntoNameId(pNetlist *Abc_Ntk_t, pAig *Aig_Man_t, pGia *Gia_Man_t) *Vec_Int_t {
	if pNetlist == nil || pAig == nil {
		return nil
	}
	size := int32(0)
	if pGia != nil {
		size = gia.Gia_ManObjNum(pGia)
	} else {
		size = aiggraph.Aig_ManObjNumMax(pAig)
	}
	mapping := vec.Vec_IntAlloc(int(size))
	vec.Vec_IntFill(mapping, int(size), -1)
	mapObject := func(obj *Abc_Obj_t) (int32, bool) {
		lit, ok := Abc_NtkDarMappedObjLit(obj)
		if !ok {
			return 0, false
		}
		if pGia == nil {
			return int32(lit.NodeID()), true
		}
		copy := aiggraph.Aig_ObjCopy(pAig, aiggraph.Aig_Regular(lit))
		if copy == aiggraph.InvalidEdge {
			return 0, false
		}
		return int32(copy) >> 1, true
	}
	put := func(net *Abc_Obj_t) {
		if net == nil || net.Kind != baseabc.ABC_OBJ_NET {
			return
		}
		fanins := net.Fanins()
		if len(fanins) == 0 {
			return
		}
		// ABC names the AIG node driving a net, not the net object itself.
		// For PI and output passes this also preserves the source's net ID while
		// associating it with the corresponding CI/CO driver node.
		id, ok := mapObject(fanins[0].Object)
		if ok && id >= 0 && id < size {
			vec.Vec_IntWriteEntry(mapping, int(id), int32(net.ID))
		}
	}
	for _, object := range pNetlist.ObjectsOfKind(baseabc.ABC_OBJ_NET) {
		put(object)
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNetlist); i++ {
		co := baseabc.Abc_NtkCo(pNetlist, i)
		fanins := co.Fanins()
		if len(fanins) != 0 {
			put(fanins[0].Object)
		}
	}
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNetlist); i++ {
		ci := baseabc.Abc_NtkCi(pNetlist, i)
		fanouts := ci.Fanouts()
		if len(fanouts) != 0 {
			put(fanouts[0])
		}
	}
	return mapping
}

func Abc_NtkTestScorrGetName(pNetlist *Abc_Ntk_t, vId2Name *Vec_Int_t, id int32) string {
	if pNetlist == nil || vId2Name == nil || id < 0 || id >= Vec_IntSize(vId2Name) {
		return ""
	}
	netID := Vec_IntEntry(vId2Name, id)
	if netID < 0 {
		return ""
	}
	obj := baseabc.Abc_NtkObj(pNetlist, netID)
	if obj == nil {
		return ""
	}
	return baseabc.Abc_ObjName(obj)
}

func Abc_NtkTestScorrWriteEquivPair(pNetlist *Abc_Ntk_t, vId2Name *Vec_Int_t, id1, id2 int32, file *os.File, polarity int32) int32 {
	name1 := Abc_NtkTestScorrGetName(pNetlist, vId2Name, id1)
	name2 := Abc_NtkTestScorrGetName(pNetlist, vId2Name, id2)
	if name1 == "" || name2 == "" || file == nil {
		return 0
	}
	if polarity != 0 {
		fmt.Fprintf(file, "%s=~%s\n", name1, name2)
	} else {
		fmt.Fprintf(file, "%s=%s\n", name1, name2)
	}
	return 1
}

func Abc_NtkTestScorrWriteEquivConst(pNetlist *Abc_Ntk_t, vId2Name *Vec_Int_t, id int32, file *os.File, polarity int32) int32 {
	name := Abc_NtkTestScorrGetName(pNetlist, vId2Name, id)
	if name == "" || file == nil {
		return 0
	}
	if polarity != 0 {
		fmt.Fprintf(file, "%s=~const0\n", name)
	} else {
		fmt.Fprintf(file, "%s=const0\n", name)
	}
	return 1
}

func Abc_NtkBmcFileName(name string) string {
	generic := extra.Extra_FileNameGeneric(name)
	if len(generic) > len(name) {
		return name + "_bmc"
	}
	return generic + "_bmc" + name[len(generic):]
}

func scorrOpenOutput(data *Tst_Dat_t) (*os.File, string) {
	filename := data.pFileNameOut
	if data.fDumpBmc != 0 {
		data.fDumpBmc = 0
		filename = Abc_NtkBmcFileName(filename)
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Cannot open equivalence output file %q: %v\n", filename, err)
		return nil, filename
	}
	return file, filename
}

func Abc_NtkTestScorrWriteEquivGia(data *Tst_Dat_t) int32 {
	if data == nil || data.pGia == nil {
		return 0
	}
	file, filename := scorrOpenOutput(data)
	if file == nil {
		return 0
	}
	defer file.Close()
	gia.Gia_ManSetPhase(data.pGia)
	var count int32
	for id := int32(0); id < gia.Gia_ManObjNum(data.pGia); id++ {
		if gia.Gia_ObjHasRepr(data.pGia, id) == 0 {
			continue
		}
		reprID := gia.Gia_ObjRepr(data.pGia, id)
		obj, repr := gia.Gia_ManObj(data.pGia, id), gia.Gia_ManObj(data.pGia, reprID)
		if data.fFlopOnly != 0 {
			if gia.Gia_ObjIsRo(data.pGia, obj) == 0 || (gia.Gia_ObjIsRo(data.pGia, repr) == 0 && gia.Gia_ObjIsConst0(repr) == 0) {
				continue
			}
		} else if data.fFfNdOnly != 0 {
			if gia.Gia_ObjIsRo(data.pGia, obj) == 0 && (gia.Gia_ObjIsRo(data.pGia, repr) == 0 && gia.Gia_ObjIsConst0(repr) == 0) {
				continue
			}
		}
		if reprID == 0 {
			count += Abc_NtkTestScorrWriteEquivConst(data.pNetlist, data.vId2Name, id, file, gia.Gia_ObjPhase(obj))
		} else {
			count += Abc_NtkTestScorrWriteEquivPair(data.pNetlist, data.vId2Name, reprID, id, file, gia.Gia_ObjPhase(repr)^gia.Gia_ObjPhase(obj))
		}
	}
	fmt.Printf("%d pairs of sequentially equivalent nodes are written into file %q.\n", count, filename)
	return count
}

func Abc_NtkTestScorrWriteEquivAig(data *Tst_Dat_t) int32 {
	if data == nil || data.pAig == nil {
		return 0
	}
	file, filename := scorrOpenOutput(data)
	if file == nil {
		return 0
	}
	defer file.Close()
	var count int32
	for id := int32(0); id < aiggraph.Aig_ManObjNumMax(data.pAig); id++ {
		obj := aiggraph.Aig_ManObj(data.pAig, id)
		if aiggraph.Aig_ObjIsNone(data.pAig, obj) {
			continue
		}
		repr := aiggraph.Aig_ObjRepr(data.pAig, obj)
		if repr == aiggraph.InvalidEdge {
			continue
		}
		repr = aiggraph.Aig_Regular(repr)
		if data.fFlopOnly != 0 {
			if !saig.Saig_ObjIsLo(data.pAig, obj) || (!saig.Saig_ObjIsLo(data.pAig, repr) && repr != aiggraph.Aig_ManConst1(data.pAig)) {
				continue
			}
		} else if data.fFfNdOnly != 0 {
			if !saig.Saig_ObjIsLo(data.pAig, obj) && (!saig.Saig_ObjIsLo(data.pAig, repr) && repr != aiggraph.Aig_ManConst1(data.pAig)) {
				continue
			}
		}
		if repr == aiggraph.Aig_ManConst1(data.pAig) {
			count += Abc_NtkTestScorrWriteEquivConst(data.pNetlist, data.vId2Name, id, file, scorrPolarity(aiggraph.Aig_ObjPhase(data.pAig, obj)))
		} else {
			polarity := scorrPolarity(aiggraph.Aig_ObjPhase(data.pAig, repr) != aiggraph.Aig_ObjPhase(data.pAig, obj))
			count += Abc_NtkTestScorrWriteEquivPair(data.pNetlist, data.vId2Name, int32(repr.NodeID()), id, file, polarity)
		}
	}
	fmt.Printf("%d pairs of sequentially equivalent nodes are written into file %q.\n", count, filename)
	return count
}

func scorrPolarity(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

// Abc_NtkTestScorr runs either the Gia/Cec or AIG/Ssw sequential
// correspondence engine and returns the sequentially swept network.
func Abc_NtkTestScorr(inputFile, outputFile string, stepsMax, backtrackLimit, newAlgo, flopOnly, ffNodeOnly, verbose int32) *Abc_Ntk_t {
	if file, err := os.Open(inputFile); err != nil {
		fmt.Printf("Input file %q cannot be opened.\n", inputFile)
		return nil
	} else {
		file.Close()
	}
	if file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666); err != nil {
		fmt.Printf("Output file %q cannot be opened.\n", outputFile)
		return nil
	} else {
		file.Close()
	}

	netlist := abcio.Io_ReadNetlist(inputFile, abcio.Io_ReadFileType(inputFile), 1)
	if netlist == nil {
		fmt.Printf("Reading input file %q has failed.\n", inputFile)
		return nil
	}
	logic := baseabc.Abc_NtkToLogic(netlist)
	if logic == nil {
		baseabc.Abc_NtkDelete(netlist)
		fmt.Printf("Deriving logic network from input file %s has failed.\n", inputFile)
		return nil
	}
	cleanup := func() {
		baseabc.Abc_NtkDelete(logic)
		baseabc.Abc_NtkDelete(netlist)
	}
	if extra.Extra_FileIsType(inputFile, ".bench", ".BENCH", "") != 0 {
		initFile := extra.Extra_FileNameGenericAppend(baseabc.Abc_NtkSpec(logic), ".init")
		file, err := os.Open(initFile)
		if err != nil {
			cleanup()
			fmt.Printf("Init file %q cannot be opened.\n", initFile)
			return nil
		}
		file.Close()
		abcio.Io_ReadBenchInit(logic, initFile)
		baseabc.Abc_NtkConvertDcLatches(logic)
		if verbose != 0 {
			fmt.Printf("Initial state was derived from file %q.\n", initFile)
		}
	}
	strash := Abc_NtkStrash(logic, 0, 1, 0)
	if strash == nil {
		cleanup()
		fmt.Printf("Deriving strashed network from input file %s has failed.\n", inputFile)
		return nil
	}
	aig := Abc_NtkToDar(strash, 0, 1)
	if aig == nil {
		baseabc.Abc_NtkDelete(strash)
		cleanup()
		return nil
	}
	var resultAig *Aig_Man_t
	var resultGia *Gia_Man_t
	data := &Tst_Dat_t{pNetlist: netlist, pFileNameOut: outputFile, fFlopOnly: flopOnly, fFfNdOnly: ffNodeOnly, fDumpBmc: 1}
	if newAlgo != 0 {
		params := &cec.Cec_ParCor_t{}
		cec.Cec_ManCorSetDefaultParams(params)
		giaMan := gia.Gia_ManFromAig(aig)
		data.pGia = giaMan
		data.vId2Name = Abc_NtkMapGiaIntoNameId(netlist, aig, giaMan)
		callback := func(value any) { Abc_NtkTestScorrWriteEquivGia(value.(*Tst_Dat_t)) }
		cec.Cec_ParCorSetScorr(params, backtrackLimit, stepsMax, verbose, data, callback)
		resultGia = cec.Cec_ManLSCorrespondence(giaMan, params)
		if resultGia != nil {
			resultAig = gia.Gia_ManToAigSimple(resultGia)
			gia.Gia_ManStop(resultGia)
		}
		gia.Gia_ManStop(giaMan)
	} else {
		params := &ssw.Ssw_Pars_t{}
		ssw.Ssw_ManSetDefaultParams(params)
		data.pAig = aig
		data.vId2Name = Abc_NtkMapGiaIntoNameId(netlist, aig, nil)
		callback := func(value any) { Abc_NtkTestScorrWriteEquivAig(value.(*Tst_Dat_t)) }
		ssw.Ssw_ParsSetScorr(params, backtrackLimit, stepsMax, verbose, data, callback)
		resultAig = ssw.Ssw_SignalCorrespondence(aig, params)
	}
	var result *Abc_Ntk_t
	if resultAig != nil {
		result = Abc_NtkFromDarSeqSweep(strash, resultAig)
	}
	Vec_IntFree(data.vId2Name)
	if resultAig != nil {
		aiggraph.Aig_ManStop(resultAig)
	}
	aiggraph.Aig_ManStop(aig)
	baseabc.Abc_NtkDelete(strash)
	cleanup()
	return result
}
