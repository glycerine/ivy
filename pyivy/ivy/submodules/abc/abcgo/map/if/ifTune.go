package if_

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/bool/kit"
	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

type Gia_Man_t = gia.Gia_Man_t
type Gia_Obj_t = gia.Gia_Obj_t
type Aig_Man_t = aig.Aig_Man_t
type sat_solver = bsat.Solver
type Cnf_Dat_t = cnf.Cnf_Dat_t

/**CFile****************************************************************

  FileName    [ifTune.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Library tuning.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifTune.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"
//#include "aig/gia/giaAig.h"
//#include "sat/bsat/satStore.h"
//#include "sat/cnf/cnf.h"
//#include "misc/extra/extra.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// network types
type Ifn_DsdType_t = int32

const (
	IFN_DSD_NONE Ifn_DsdType_t = 0
	// 0:  unknown
	IFN_DSD_CONST0 Ifn_DsdType_t = IFN_DSD_NONE + 1
	// 1:  constant
	IFN_DSD_VAR Ifn_DsdType_t = IFN_DSD_CONST0 + 1
	// 2:  variable
	IFN_DSD_AND Ifn_DsdType_t = IFN_DSD_VAR + 1
	// 3:  AND
	IFN_DSD_XOR Ifn_DsdType_t = IFN_DSD_AND + 1
	// 4:  XOR
	IFN_DSD_MUX Ifn_DsdType_t = IFN_DSD_XOR + 1
	// 5:  MUX
	IFN_DSD_PRIME Ifn_DsdType_t = IFN_DSD_MUX + 1 // 6:  PRIME

)

// object types
var Ifn_Symbs = [16]string{"", "const", "var", "()", "[]", "<>", "{}"}

func Ifn_ElemTruth(p *Ifn_Ntk_t, i int32) []word {
	start := int(i * abcutil.Abc_TtWordNum(p.nInps))
	return p.pTtElems[start : start+int(p.nWords)]
}
func Ifn_ObjTruth(p *Ifn_Ntk_t, i int32) []word {
	start := int(i * p.nWords)
	return p.pTtObjs[start : start+int(p.nWords)]
}

// variable ordering
// - primary inputs            [0;            p->nInps)
// - internal nodes            [p->nInps;     p->nObjs)
// - configuration params      [p->nObjs;     p->nParsVIni)
// - variable selection params [p->nParsVIni; p->nPars)

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Prepare network to check the given function.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ifn_Prepare(p *Ifn_Ntk_t, pTruth []word, nVars int32) int32 {
	var i, fVerbose int32
	assert(nVars <= p.nInps, "nVars <= p.nInps")
	p.pTruth = pTruth
	p.nVars = nVars
	p.nWords = abcutil.Abc_TtWordNum(nVars)
	p.nPars = p.nObjs
	for i = p.nInps; i < p.nObjs; i++ {
		if p.Nodes[i].Type != IFN_DSD_PRIME {
			continue
		}
		p.Nodes[i].iFirst = p.nPars
		p.nPars += (1 << p.Nodes[i].nFanins)
		if fVerbose != 0 {
			fmt.Printf("Node %d  Start %d  Vars %d\n", i, p.Nodes[i].iFirst, (1 << p.Nodes[i].nFanins))
		}
	}
	if fVerbose != 0 {
		fmt.Printf("Groups start %d\n", p.nPars)
	}
	p.nParsVIni = p.nPars
	p.nParsVNum = abcutil.Abc_Base2Log(uint32(nVars))
	p.nPars += p.nParsVNum * p.nInps
	assert(p.nPars <= IFN_PAR, "p.nPars <= IFN_PAR")
	for i := int32(0); i < p.nPars; i++ {
		p.Values[i] = -1
	}
	return p.nPars
}
func Ifn_NtkPrint(p *Ifn_Ntk_t) {
	var i, k int32
	if p == nil {
		fmt.Printf("String is empty.\n")
	}
	if p == nil {
		return
	}
	for i = p.nInps; i < p.nObjs; i++ {
		fmt.Printf("%c=", 'a'+i)
		fmt.Printf("%c", Ifn_Symbs[p.Nodes[i].Type][0])
		for k = 0; k < int32(p.Nodes[i].nFanins); k++ {
			fmt.Printf("%c", 'a'+p.Nodes[i].Fanins[k])
		}
		fmt.Printf("%c", Ifn_Symbs[p.Nodes[i].Type][1])
		fmt.Printf(";")
	}
	fmt.Printf("\n")
}
func Ifn_NtkLutSizeMax(p *Ifn_Ntk_t) int32 {
	var i, LutSize int32
	for i = p.nInps; i < p.nObjs; i++ {
		if p.Nodes[i].Type == IFN_DSD_PRIME {
			LutSize = Abc_MaxInt(LutSize, int32(p.Nodes[i].nFanins))
		}
	}
	return LutSize
}
func Ifn_NtkInputNum(p *Ifn_Ntk_t) int32 {
	return p.nInps
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ifn_ErrorMessage(format string, args ...any) int32 {
	fmt.Printf("%s", fmt.Sprintf(format, args...))
	return 0
}
func Inf_ManOpenSymb(pStr string) int32 {
	if len(pStr) == 0 {
		return 0
	}
	if pStr[0] == '(' {
		return 3
	}
	if pStr[0] == '[' {
		return 4
	}
	if pStr[0] == '<' {
		return 5
	}
	if pStr[0] == '{' {
		return 6
	}
	return 0
}
func Ifn_ManStrCheck(pStr string, pnInps *int32, pnObjs *int32) int32 {
	var i int32
	var nNodes int32
	var Marks [32]int32
	var MaxVar int32 = -1
	for i = 0; i < int32(len(pStr)); i++ {
		if Inf_ManOpenSymb(pStr[i:]) != 0 {
			nNodes++
		}
		if pStr[i] == ';' ||
			pStr[i] == '(' || pStr[i] == ')' ||
			pStr[i] == '[' || pStr[i] == ']' ||
			pStr[i] == '<' || pStr[i] == '>' ||
			pStr[i] == '{' || pStr[i] == '}' {
			continue
		}
		if pStr[i] >= 'A' && pStr[i] <= 'Z' {
			continue
		}
		if pStr[i] >= 'a' && pStr[i] <= 'z' {
			MaxVar = Abc_MaxInt(MaxVar, int32((pStr[i] - 'a')))
			Marks[pStr[i]-'a'] = 1
			continue
		}
		return Ifn_ErrorMessage("String \"%s\" contains unrecognized symbol '%c'.\n", pStr, pStr[i])
	}
	for i = 0; i <= MaxVar; i++ {
		if Marks[i] == 0 {
			return Ifn_ErrorMessage("String \"%s\" has no symbol '%c'.\n", pStr, 'a'+i)
		}
	}
	*pnInps = MaxVar + 1
	*pnObjs = MaxVar + 1 + nNodes
	return 1
}
func Ifn_NtkParseFindClosingParenthesis(text string, open, close byte) string {
	if len(text) == 0 || text[0] != open {
		return ""
	}
	depth := 0
	for i := 0; i < len(text); i++ {
		if text[i] == open {
			depth++
		}
		if text[i] == close {
			depth--
			if depth == 0 {
				return text[i:]
			}
		}
	}
	return ""
}
func Ifn_NtkParseInt_rec(text string, p *Ifn_Ntk_t, final *string, node *int32) int32 {
	*final = ""
	typ := Inf_ManOpenSymb(text)
	if typ < IFN_DSD_AND || typ > IFN_DSD_PRIME {
		return Ifn_ErrorMessage("Expected a DSD operator at %q", text)
	}
	pair := Ifn_Symbs[typ]
	closing := Ifn_NtkParseFindClosingParenthesis(text, pair[0], pair[1])
	if closing == "" {
		return Ifn_ErrorMessage("For symbol %q cannot find matching close", pair)
	}
	closeAt := len(text) - len(closing)
	body := text[1:closeAt]
	var fanins [IFN_INS]int32
	var count int32
	for len(body) > 0 {
		if count >= IFN_INS {
			return Ifn_ErrorMessage("DSD node exceeds %d fanins", IFN_INS)
		}
		if body[0] >= 'a' && body[0] <= 'z' {
			fanins[count] = int32(body[0] - 'a')
			count++
			body = body[1:]
			continue
		}
		childType := Inf_ManOpenSymb(body)
		if childType == 0 {
			return Ifn_ErrorMessage("Unrecognized DSD token %q", body[:1])
		}
		var rest string
		if Ifn_NtkParseInt_rec(body, p, &rest, node) == 0 {
			return 0
		}
		consumed := len(body) - len(rest)
		if consumed <= 0 {
			return Ifn_ErrorMessage("DSD parser made no progress")
		}
		fanins[count] = *node - 1
		count++
		body = rest
	}
	if typ == IFN_DSD_MUX && count != 3 {
		return Ifn_ErrorMessage("MUX should have exactly three fanins")
	}
	obj := &p.Nodes[*node]
	obj.Type = typ
	obj.nFanins = count
	copy(obj.Fanins[:], fanins[:count])
	*node++
	*final = text[closeAt+1:]
	return 1
}
func Ifn_NtkParseInt(text string, p *Ifn_Ntk_t) int32 {
	if Ifn_ManStrCheck(text, &p.nInps, &p.nObjs) == 0 {
		return 0
	}
	if p.nInps > IFN_INS {
		return Ifn_ErrorMessage("The number of variables (%d) exceeds limit %d", p.nInps, IFN_INS)
	}
	assert(p.nInps > 1 && p.nInps < p.nObjs && p.nInps <= IFN_INS && p.nObjs < 2*IFN_INS, "valid parsed network dimensions")
	if Inf_ManOpenSymb(text) == 0 {
		return Ifn_ErrorMessage("The first token must be a DSD operator")
	}
	node := p.nInps
	var rest string
	if Ifn_NtkParseInt_rec(text, p, &rest, &node) == 0 {
		return 0
	}
	if len(rest) > 0 && rest[0] != ';' {
		return Ifn_ErrorMessage("Unexpected trailing text %q", rest)
	}
	if node != p.nObjs {
		return Ifn_ErrorMessage("Mismatch in node count")
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ifn_ManStrType2(pStr string) int32 {
	var i int32
	for i = 0; i < int32(len(pStr)); i++ {
		if pStr[i] == '=' {
			return 1
		}
	}
	return 0
}
func Ifn_ManStrCheck2(pStr string, pnInps *int32, pnObjs *int32) int32 {
	var i int32
	var Marks [32]int32
	var MaxVar int32
	var MaxDef int32
	for i = 0; i < int32(len(pStr)); i++ {
		if pStr[i] == '=' || pStr[i] == ';' ||
			pStr[i] == '(' || pStr[i] == ')' ||
			pStr[i] == '[' || pStr[i] == ']' ||
			pStr[i] == '<' || pStr[i] == '>' ||
			pStr[i] == '{' || pStr[i] == '}' {
			continue
		}
		if pStr[i] >= 'A' && pStr[i] <= 'Z' {
			continue
		}
		if pStr[i] >= 'a' && pStr[i] <= 'z' {
			if i+1 < int32(len(pStr)) && pStr[i+1] == '=' {
				Marks[pStr[i]-'a'] = 2
				MaxDef = Abc_MaxInt(MaxDef, int32(pStr[i]-'a'))
			}
			continue
		}
		return Ifn_ErrorMessage("String \"%s\" contains unrecognized symbol '%c'.\n", pStr, pStr[i])
	}
	for i = 0; i < int32(len(pStr)); i++ {
		if pStr[i] == '=' || pStr[i] == ';' ||
			pStr[i] == '(' || pStr[i] == ')' ||
			pStr[i] == '[' || pStr[i] == ']' ||
			pStr[i] == '<' || pStr[i] == '>' ||
			pStr[i] == '{' || pStr[i] == '}' {
			continue
		}
		if pStr[i] >= 'A' && pStr[i] <= 'Z' {
			continue
		}
		if pStr[i] >= 'a' && pStr[i] <= 'z' {
			if (i+1 >= int32(len(pStr)) || pStr[i+1] != '=') && Marks[pStr[i]-'a'] != 2 {
				Marks[pStr[i]-'a'] = 1
				MaxVar = Abc_MaxInt(MaxVar, int32(pStr[i]-'a'))
			}
			continue
		}
		return Ifn_ErrorMessage("String \"%s\" contains unrecognized symbol '%c'.\n", pStr, pStr[i])
	}
	MaxVar++
	MaxDef++
	for i = 0; i < MaxDef; i++ {
		if Marks[i] == 0 {
			return Ifn_ErrorMessage("String \"%s\" has no symbol '%c'.\n", pStr, 'a'+i)
		}
	}
	for i = 0; i < MaxVar; i++ {
		if Marks[i] == 2 {
			return Ifn_ErrorMessage("String \"%s\" has definition of input variable '%c'.\n", pStr, 'a'+i)
		}
	}
	for i = MaxVar; i < MaxDef; i++ {
		if Marks[i] == 1 {
			return Ifn_ErrorMessage("String \"%s\" has no definition for internal variable '%c'.\n", pStr, 'a'+i)
		}
	}
	*pnInps = MaxVar
	*pnObjs = MaxDef
	return 1
}
func Ifn_NtkParseInt2(pStr string, p *Ifn_Ntk_t) int32 {
	var i, k, n, f, nFans, iFan int32
	if !(Ifn_ManStrCheck2(pStr, &p.nInps, &p.nObjs) != 0) {
		return 0
	}
	if p.nInps > IFN_INS {
		return Ifn_ErrorMessage("The number of variables (%d) exceeds predefined limit (%d). Recompile with different value of IFN_INS.\n", p.nInps, IFN_INS)
	}
	assert(p.nInps > 1 && p.nInps < p.nObjs && p.nInps <= IFN_INS && p.nObjs < 2*IFN_INS, "p.nInps > 1 && p.nInps < p.nObjs && p.nInps <= IFN_INS && p.nObjs < 2*IFN_INS")
	for i = p.nInps; i < p.nObjs; i++ {
		var Next byte
		for k = 0; k < int32(len(pStr)); k++ {
			if k+1 < int32(len(pStr)) && int32(pStr[k]) == 'a'+i && pStr[k+1] == '=' {
				break
			}
		}
		if k >= int32(len(pStr)) {
			return Ifn_ErrorMessage("Cannot find definition of signal '%c'.\n", 'a'+i)
		}
		if k+2 < int32(len(pStr)) && pStr[k+2] == '(' {
			p.Nodes[i].Type = int32(IFN_DSD_AND)
			Next = ')'
		} else if k+2 < int32(len(pStr)) && pStr[k+2] == '[' {
			p.Nodes[i].Type = int32(IFN_DSD_XOR)
			Next = ']'
		} else if k+2 < int32(len(pStr)) && pStr[k+2] == '<' {
			p.Nodes[i].Type = int32(IFN_DSD_MUX)
			Next = '>'
		} else if k+2 < int32(len(pStr)) && pStr[k+2] == '{' {
			p.Nodes[i].Type = int32(IFN_DSD_PRIME)
			Next = '}'
		} else {
			return Ifn_ErrorMessage("Cannot find openning operation symbol in the definition of signal '%c'.\n", 'a'+i)
		}
		for n = k + 3; n < int32(len(pStr)); n++ {
			if pStr[n] == Next {
				break
			}
		}
		if n >= int32(len(pStr)) {
			return Ifn_ErrorMessage("Cannot find closing operation symbol in the definition of signal '%c'.\n", 'a'+i)
		}
		nFans = n - k - 3
		if nFans > 8 {
			return Ifn_ErrorMessage("Cannot find matching operation symbol in the definition of signal '%c'.\n", 'a'+i)
		}
		for f = 0; f < nFans; f++ {
			iFan = int32(pStr[k+3+f]) - 'a'
			if iFan < 0 || iFan >= i {
				return Ifn_ErrorMessage("Fanin number %d is signal %d is out of range.\n", f, 'a'+i)
			}
			p.Nodes[i].Fanins[f] = iFan
		}
		p.Nodes[i].nFanins = nFans
	}
	return 1
}
func Ifn_NtkParseConstraints(pStr string, p *Ifn_Ntk_t) {
	var i, k int32
	// parse constraints
	p.nConstr = 0
	for i = 0; i < p.nInps; i++ {
		for k = 0; k < int32(len(pStr)); k++ {
			if k > 0 && k+1 < int32(len(pStr)) && int32(pStr[k]) == 'A'+i && pStr[k-1] == ';' {
				assert(p.nConstr < IFN_INS*IFN_INS, "p.nConstr < IFN_INS*IFN_INS")
				p.pConstr[p.nConstr] = (int32((pStr[k] - 'A')) << 16) | int32((pStr[k+1] - 'A'))
				p.nConstr++
				//                printf( "Added constraint (%c < %c)\n", pStr[k], pStr[k+1] );
			}
		}
	}
	// if ( p->nConstr )
	//
	//	printf( "Total constraints = %d\n", p->nConstr );
}

func Ifn_NtkParse(pStr string) *Ifn_Ntk_t {
	var p *Ifn_Ntk_t = new(Ifn_Ntk_t)
	if Ifn_ManStrType2(pStr) != 0 {
		if !(Ifn_NtkParseInt2(pStr, p) != 0) {
			return nil
		}

	} else {
		if !(Ifn_NtkParseInt(pStr, p) != 0) {
			return nil
		}
	}
	Ifn_NtkParseConstraints(pStr, p)
	abcutil.Abc_TtElemInit2(p.pTtElems[:], p.nInps)
	//    printf( "Finished parsing: " ); Ifn_NtkPrint(p);
	return p
}
func Ifn_NtkTtBits(pStr string) int32 {
	var i, Counter int32
	var pNtk *Ifn_Ntk_t = Ifn_NtkParse(pStr)
	for i = pNtk.nInps; i < pNtk.nObjs; i++ {
		if pNtk.Nodes[i].Type == IFN_DSD_PRIME {
			Counter += (1 << pNtk.Nodes[i].nFanins)
		}
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    [Derive AIG.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ifn_ManStrFindModel(p *Ifn_Ntk_t) *Gia_Man_t {
	var pNew, pTemp *Gia_Man_t
	var i int32
	var k int32
	var iLit int32
	pVarMap := make([]int32, p.nParsVIni)
	pNew = Gia_ManStart(1000)
	pNew.SetIdentity("model", "")
	Gia_ManHashStart(pNew)
	for i = 0; i < p.nInps; i++ {
		pVarMap[i] = Gia_ManAppendCi(pNew)
	}
	for i = p.nObjs; i < p.nParsVIni; i++ {
		pVarMap[i] = Gia_ManAppendCi(pNew)
	}
	for i = p.nInps; i < p.nObjs; i++ {
		var Type int32 = p.Nodes[i].Type
		var nFans int32 = p.Nodes[i].nFanins
		pFans := p.Nodes[i].Fanins[:]
		var iFanin int32 = p.Nodes[i].iFirst
		if Type == IFN_DSD_AND {
			iLit = 1
			for k = 0; k < nFans; k++ {
				iLit = Gia_ManHashAnd(pNew, iLit, pVarMap[pFans[k]])
			}
			pVarMap[i] = iLit

		} else if Type == IFN_DSD_XOR {
			iLit = 0
			for k = 0; k < nFans; k++ {
				iLit = Gia_ManHashXor(pNew, iLit, pVarMap[pFans[k]])
			}
			pVarMap[i] = iLit

		} else if Type == IFN_DSD_MUX {
			assert(nFans == 3, "nFans == 3")
			pVarMap[i] = Gia_ManHashMux(pNew, pVarMap[pFans[0]], pVarMap[pFans[1]], pVarMap[pFans[2]])

		} else if Type == IFN_DSD_PRIME {
			var n int32
			var Step int32
			var pVarsData [256]int32
			var nMints int32 = (1 << nFans)
			assert(nFans >= 0 && nFans <= 8, "nFans >= 0 && nFans <= 8")
			for k = 0; k < nMints; k++ {
				pVarsData[k] = pVarMap[iFanin+k]
			}
			Step = 1
			for k = 0; k < nFans; k++ {
				for n = 0; n < nMints; n += Step << 1 {
					pVarsData[n] = Gia_ManHashMux(pNew, pVarMap[pFans[k]], pVarsData[n+Step], pVarsData[n])
				}
				Step <<= 1
			}
			assert(Step == nMints, "Step == nMints")
			pVarMap[i] = pVarsData[0]

		} else {
			assert(false, "0")
		}
	}
	Gia_ManAppendCo(pNew, pVarMap[p.nObjs-1])
	pTemp = pNew
	pNew = Gia_ManCleanup(pTemp)
	Gia_ManStop(pTemp)
	assert(Gia_ManPiNum(pNew) == p.nParsVIni-(p.nObjs-p.nInps), "Gia_ManPiNum(pNew) == p.nParsVIni - (p.nObjs - p.nInps)")
	assert(Gia_ManPoNum(pNew) == 1, "Gia_ManPoNum(pNew) == 1")
	return pNew
}

// compute cofactors w.r.t. the first nIns variables
func Ifn_ManStrFindCofactors(nIns int32, p *Gia_Man_t) *Gia_Man_t {
	nCIs := Gia_ManCiNum(p)
	if nIns < 0 || nIns > nCIs {
		panic("if: cofactor input count out of range")
	}
	nMints := int32(1) << nIns
	pNew := Gia_ManStart(Gia_ManObjNum(p) * nMints)
	name, spec := p.Identity()
	pNew.SetIdentity(name, spec)
	Gia_ManHashAlloc(pNew)
	for i := nIns; i < nCIs; i++ {
		Gia_ManCi(p, i).Value = uint32(Gia_ManAppendCi(pNew))
	}
	for m := int32(0); m < nMints; m++ {
		for i := int32(0); i < nIns; i++ {
			Gia_ManCi(p, i).Value = uint32((m >> i) & 1)
		}
		for id := int32(1); id < Gia_ManObjNum(p); id++ {
			obj := Gia_ManObj(p, id)
			if Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			obj.Value = uint32(Gia_ManHashAnd(pNew, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		}
		for i := int32(0); i < Gia_ManPoNum(p); i++ {
			Gia_ManPo(p, i).Value = uint32(Gia_ManAppendCo(pNew, Gia_ObjFanin0Copy(Gia_ManPo(p, i))))
		}
	}
	cleaned := Gia_ManCleanup(pNew)
	Gia_ManStop(pNew)
	return cleaned
}

/*
*Function*************************************************************

	Synopsis    [Derive SAT solver.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cnf_DeriveGiaRemapped(p *Gia_Man_t) *Cnf_Dat_t {
	pAig := gia.Gia_ManToAigSimple(p)
	aig.Aig_ManSetRegNum(pAig, 0)
	pCnf := cnf.Cnf_Derive(pAig, aig.Aig_ManCoNum(pAig))
	aig.Aig_ManStop(pAig)
	return pCnf
}
func Ifn_ManStrFindSolver(p *Gia_Man_t, pvPiVars **Vec_Int_t, pvPoVars **Vec_Int_t) *sat_solver {
	pCnf := Cnf_DeriveGiaRemapped(p)
	pSat := bsat.NewSolver()
	for int(pSat.NumVars()) < pCnf.Variables {
		pSat.NewVariable()
	}
	for _, clause := range pCnf.Clauses {
		if !pSat.AddClause(clause...) {
			panic("if: CNF for cell network is inconsistent")
		}
	}
	// inputs/outputs
	*pvPiVars = Vec_IntAlloc(int(Gia_ManPiNum(p)))
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		obj := gia.Gia_ManCi(p, i)
		Vec_IntPush(*pvPiVars, pCnf.VarNums[gia.Gia_ObjId(p, obj)])
	}
	*pvPoVars = Vec_IntAlloc(int(Gia_ManPoNum(p)))
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		obj := gia.Gia_ManPo(p, i)
		Vec_IntPush(*pvPoVars, pCnf.VarNums[gia.Gia_ObjId(p, obj)])
	}
	cnf.Cnf_DataFree(pCnf)
	return pSat
}
func Ifn_ManSatBuild(p *Ifn_Ntk_t, pvPiVars **Vec_Int_t, pvPoVars **Vec_Int_t) *sat_solver {
	var p1, p2 *Gia_Man_t
	var pSat *sat_solver = nil
	*pvPiVars = nil
	*pvPoVars = nil
	p1 = Ifn_ManStrFindModel(p)
	//    Gia_AigerWrite( p1, "satbuild.aig", 0, 0, 0 );
	p2 = Ifn_ManStrFindCofactors(p.nInps, p1)
	Gia_ManStop(p1)
	//    Gia_AigerWrite( p2, "satbuild2.aig", 0, 0, 0 );
	pSat = Ifn_ManStrFindSolver(p2, pvPiVars, pvPoVars)
	Gia_ManStop(p2)
	return pSat
}
func If_ManSatBuildFromCell(pStr string, pvPiVars **Vec_Int_t, pvPoVars **Vec_Int_t, ppNtk **Ifn_Ntk_t) any {
	var p *Ifn_Ntk_t = Ifn_NtkParse(pStr)
	*ppNtk = p
	if p == nil {
		return nil
	}
	Ifn_Prepare(p, nil, p.nInps)
	//    Ifn_NtkPrint( p );
	return Ifn_ManSatBuild(p, pvPiVars, pvPoVars)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ifn_ManSatPrintPerm(pPerms string, nVars int32) {
	var i int32
	for i = 0; i < nVars; i++ {
		fmt.Printf("%c", 'a'+pPerms[i])
	}
	fmt.Printf("\n")
}
func Ifn_ManSatCheckOne(pSat *sat_solver, vPoVars *Vec_Int_t, pTruth []word, nVars int32, pPerm []int32, nInps int32, vLits *Vec_Int_t) int32 {
	var v int32
	var Value int32
	var m int32
	var mNew int32
	var nMints int32 = (1 << nVars) // (1 << nInps);
	assert((1<<nInps) == Vec_IntSize(vPoVars), "(1 << nInps) == Vec_IntSize(vPoVars)")
	assert(nVars <= nInps, "nVars <= nInps")
	// remap minterms
	vec.Vec_IntFill(vLits, Vec_IntSize(vPoVars), -1)
	for m = 0; m < nMints; m++ {
		mNew = 0
		for v = 0; v < nInps; v++ {
			assert(pPerm[v] < nVars, "pPerm[v] < nVars")
			if ((m >> pPerm[v]) & 1) != 0 {
				mNew |= (1 << v)
			}
		}
		assert(Vec_IntEntry(vLits, int(mNew)) == -1, "Vec_IntEntry(vLits, int(mNew)) == -1")
		Vec_IntWriteEntry(vLits, int(mNew), abcutil.Abc_TtGetBit(pTruth, m))
	}
	// find assumptions
	v = 0
	for m = 0; m < int32(Vec_IntSize(vLits)); m++ {
		Value = Vec_IntEntry(vLits, int(m))
		if Value >= 0 {
			Vec_IntWriteEntry(vLits, int(v), abcutil.Abc_Var2Lit(Vec_IntEntry(vPoVars, int(m)), ifTuneBoolInt(Value == 0)))
			v++
		}
	}
	vec.Vec_IntShrink(vLits, int(v))
	// run SAT solver
	return ifTuneBoolInt(pSat.Solve(Vec_IntArray(vLits)...) == bsat.Satisfiable)
}
func Ifn_ManSatDeriveOne(pSat *sat_solver, vPiVars *Vec_Int_t, vValues *Vec_Int_t) {
	var i, iVar int32
	vec.Vec_IntClear(vValues)
	for i = 0; i < int32(Vec_IntSize(vPiVars)); i++ {
		iVar = Vec_IntEntry(vPiVars, int(i))
		Vec_IntPush(vValues, satValue(pSat, iVar))
	}
}
func If_ManSatFindCofigBits(pSat any, vPiVars *Vec_Int_t, vPoVars *Vec_Int_t, pTruth []word, nVars int32, Perm word, nInps int32, vValues *Vec_Int_t) int32 {
	// extract permutation
	var RetValue int32
	var i int32
	var pPerm [IF_MAX_FUNC_LUTSIZE]int32
	assert(nInps <= IF_MAX_FUNC_LUTSIZE, "nInps <= IF_MAX_FUNC_LUTSIZE")
	for i = 0; i < nInps; i++ {
		pPerm[i] = abcutil.Abc_TtGetHex([]word{Perm}, i)
		assert(pPerm[i] < nVars, "pPerm[i] < nVars")
	}
	// perform SAT check
	solver, ok := pSat.(*sat_solver)
	if !ok {
		return 0
	}
	RetValue = Ifn_ManSatCheckOne(solver, vPoVars, pTruth, nVars, pPerm[:], nInps, vValues)
	vec.Vec_IntClear(vValues)
	if RetValue == 0 {
		return 0
	}
	Ifn_ManSatDeriveOne(solver, vPiVars, vValues)
	return 1
}
func Ifn_ManSatFindCofigBitsTest(p *Ifn_Ntk_t, pTruth []word, nVars int32, Perm word) int32 {
	var vValues *Vec_Int_t = Vec_IntAlloc(100)
	var vPiVars, vPoVars *Vec_Int_t
	var pSat *sat_solver = Ifn_ManSatBuild(p, &vPiVars, &vPoVars)
	var RetValue int32 = If_ManSatFindCofigBits(pSat, vPiVars, vPoVars, pTruth, nVars, Perm, p.nInps, vValues)
	vec.Vec_IntPrint(vValues)
	// cleanup
	vec.Vec_IntFreeP(&vPiVars)
	vec.Vec_IntFreeP(&vPoVars)
	vec.Vec_IntFreeP(&vValues)
	return RetValue
}

func ifTuneBoolInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
func satValue(s *sat_solver, variable int32) int32 {
	value, ok := s.Value(variable)
	if ok && value {
		return 1
	}
	return 0
}

func sat_solver_new() *sat_solver     { return bsat.NewSolver() }
func sat_solver_delete(_ *sat_solver) {}
func sat_solver_setnvars(s *sat_solver, count int32) {
	for s.NumVars() < count {
		s.NewVariable()
	}
}
func sat_solver_nvars(s *sat_solver) int32              { return s.NumVars() }
func sat_solver_nclauses(s *sat_solver) int32           { return int32(s.Statistics().Clauses) }
func sat_solver_nconflicts(s *sat_solver) int32         { return int32(s.Statistics().Conflicts) }
func sat_solver_var_value(s *sat_solver, v int32) int32 { return satValue(s, v) }

// ifnConfigWords adapts the int32-backed configuration vector to the source
// algorithm's word-oriented bit reader. The source stores each 64-bit word as
// two consecutive 32-bit entries, low half first.
func ifnConfigWords(v *Vec_Int_t, offset, nInts int32) []word {
	words := make([]word, (nInts+1)/2)
	for j := int32(0); j < nInts; j++ {
		words[j/2] |= word(uint32(Vec_IntEntry(v, int(offset+j)))) << (32 * (j & 1))
	}
	return words
}

/*
*Function*************************************************************

	Synopsis    [Derive GIA using programmable bits.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManSatDeriveGiaFromBits(pGia any, p *Ifn_Ntk_t, pConfigData []word, vLeaves *Vec_Int_t, vCover *Vec_Int_t) int32 {
	pNew, ok := pGia.(*Gia_Man_t)
	if !ok {
		return 0
	}
	var i int32
	var k int32
	var iLit int32
	var iVar int32
	var nVarsNew int32
	var pVarMap [1000]int32
	var nTtBits int32 = p.nParsVIni - p.nObjs
	var nPermBits int32 = abcutil.Abc_Base2Log(uint32(p.nInps+1)) + 1
	var fCompl int32 = abcutil.Abc_TtGetBit(pConfigData, nTtBits+nPermBits*p.nInps)
	assert(int32(Vec_IntSize(vLeaves)) <= p.nInps && p.nParsVIni < 1000, "int32(Vec_IntSize(vLeaves)) <= p.nInps && p.nParsVIni < 1000")
	for i = 0; i < p.nInps; i++ {
		iLit = 0
		for k = 0; k < nPermBits; k++ {
			if abcutil.Abc_TtGetBit(pConfigData, nTtBits+i*nPermBits+k) != 0 {
				iLit |= (1 << k)
			}
		}
		assert(Abc_Lit2Var(iLit) < int32(Vec_IntSize(vLeaves)), "Abc_Lit2Var(iLit) < int32(Vec_IntSize(vLeaves))")
		pVarMap[i] = Abc_Lit2LitL(Vec_IntArray(vLeaves), iLit)
	}
	for i = p.nInps; i < p.nObjs; i++ {
		var Type int32 = p.Nodes[i].Type
		var nFans int32 = p.Nodes[i].nFanins
		pFans := p.Nodes[i].Fanins[:]
		//int iFanin = p->Nodes[i].iFirst;
		assert(nFans <= 6, "nFans <= 6")
		if Type == IFN_DSD_AND {
			iLit = 1
			for k = 0; k < nFans; k++ {
				iLit = Gia_ManHashAnd(pNew, iLit, pVarMap[pFans[k]])
			}
			pVarMap[i] = iLit

		} else if Type == IFN_DSD_XOR {
			iLit = 0
			for k = 0; k < nFans; k++ {
				iLit = Gia_ManHashXor(pNew, iLit, pVarMap[pFans[k]])
			}
			pVarMap[i] = iLit

		} else if Type == IFN_DSD_MUX {
			assert(nFans == 3, "nFans == 3")
			pVarMap[i] = Gia_ManHashMux(pNew, pVarMap[pFans[0]], pVarMap[pFans[1]], pVarMap[pFans[2]])

		} else if Type == IFN_DSD_PRIME {
			var pFaninLits [16]int32
			// collect truth table
			var uTruth word
			var nMints int32 = (1 << nFans)
			for k = 0; k < nMints; k++ {
				if abcutil.Abc_TtGetBit(pConfigData, iVar) != 0 {
					uTruth |= word(1) << k
				}
				iVar++
			}
			uTruth = abcutil.Abc_Tt6Stretch(uTruth, nFans)
			// collect function
			for k = 0; k < nFans; k++ {
				pFaninLits[k] = pVarMap[pFans[k]]
			}
			// implement the function
			truth := []word{uTruth}
			nVarsNew = abcutil.Abc_TtMinBase(truth, pFaninLits[:], nFans, 6)
			if nVarsNew == 0 {
				pVarMap[i] = int32((uTruth & 1))
			} else {
				truth32 := []uint32{uint32(uTruth), uint32(uTruth >> 32)}
				pVarMap[i] = kit.Kit_TruthToGia(pNew, truth32, nVarsNew, nil, pFaninLits[:nFans], 1)
			}

		} else {
			assert(false, "0")
		}
	}
	assert(iVar == nTtBits, "iVar == nTtBits")
	return abcutil.Abc_LitNotCond(pVarMap[p.nObjs-1], fCompl)
}

/*
*Function*************************************************************

	Synopsis    [Derive GIA using programmable bits.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManDeriveGiaFromCells(pGia any) any {
	p, ok := pGia.(*Gia_Man_t)
	if !ok {
		return nil
	}
	var pNew, pTemp *Gia_Man_t
	var vCover, vLeaves *Vec_Int_t
	var pNtkCell *Ifn_Ntk_t
	var pObj *Gia_Obj_t
	var pConfigData []word
	//word * pTruth;
	var i, iLut int32
	var nConfigInts, Count int32
	cellStr, configs := p.CellConfiguration()
	assert(configs != nil, "configs != nil")
	assert(cellStr != "", "p.pCellStr != nil")
	assert(Gia_ManHasMapping(p) != 0, "Gia_ManHasMapping(p)")
	// derive cell network
	pNtkCell = Ifn_NtkParse(cellStr)
	Ifn_Prepare(pNtkCell, nil, pNtkCell.nInps)
	nConfigInts = Vec_IntEntry(configs, 1)
	// create new manager
	pNew = Gia_ManStart(6*Gia_ManObjNum(p)/5 + 100)
	name, spec := p.Identity()
	pNew.SetIdentity(name, spec)
	// map primary inputs
	Gia_ManFillValue(p)
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		obj := gia.Gia_ManCi(p, i)
		obj.Value = uint32(Gia_ManAppendCi(pNew))
	}
	// iterate through nodes used in the mapping
	vLeaves = Vec_IntAlloc(16)
	vCover = Vec_IntAlloc(1 << 16)
	Gia_ManHashStart(pNew)
	//Gia_ObjComputeTruthTableStart( p, Gia_ManLutSizeMax(p) );
	for iLut = 0; iLut < Gia_ManObjNum(p); iLut++ {
		pObj = gia.Gia_ManObj(p, iLut)
		if gia.Gia_ObjIsAnd(pObj) == 0 {
			continue
		}
		if gia.Gia_ObjIsBuf(pObj) != 0 {
			pObj.Value = uint32(Gia_ManAppendBuf(pNew, Gia_ObjFanin0Copy(pObj)))
			continue
		}
		if gia.Gia_ObjIsLut(p, iLut) == 0 {
			continue
		}
		// collect leaves
		//vec.Vec_IntClear( vLeaves );
		//Gia_LutForEachFanin( p, iLut, iVar, k )
		//    Vec_IntPush( vLeaves, iVar );
		//pTruth = Gia_ObjComputeTruthTableCut( p, Gia_ManObj(p, iLut), vLeaves );
		// collect incoming literals
		vec.Vec_IntClear(vLeaves)
		for _, iVar := range gia.Gia_ObjLutFanins(p, iLut) {
			Vec_IntPush(vLeaves, int32(gia.Gia_ManObj(p, iVar).Value))
		}
		offset := 2 + nConfigInts*Count
		pConfigData = ifnConfigWords(configs, offset, nConfigInts)
		Count++
		Gia_ManObj(p, iLut).Value = uint32(If_ManSatDeriveGiaFromBits(pNew, pNtkCell, pConfigData, vLeaves, vCover))
	}
	assert(Vec_IntEntry(configs, 0) == Count, "Vec_IntEntry(configs, 0) == Count")
	assert(Vec_IntSize(configs) == int(2+nConfigInts*Count), "Vec_IntSize(configs) == 2 + nConfigInts * Count")
	//Gia_ObjComputeTruthTableStop( p );
	for i = 0; i < gia.Gia_ManCoNum(p); i++ {
		pObj = gia.Gia_ManCo(p, i)
		pObj.Value = uint32(Gia_ManAppendCo(pNew, gia.Gia_ObjFanin0Copy(pObj)))
	}
	Gia_ManHashStop(pNew)
	Gia_ManSetRegNum(pNew, Gia_ManRegNum(p))
	Vec_IntFree(vLeaves)
	Vec_IntFree(vCover)
	// perform cleanup
	pTemp = pNew
	pNew = Gia_ManCleanup(pTemp)
	Gia_ManStop(pTemp)
	return pNew

}

/*
*Function*************************************************************

	Synopsis    [Derive truth table given the configulation values.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ifn_NtkDeriveTruth(p *Ifn_Ntk_t, pValues []int32) []word {
	var i, v, f, iVar, iStart int32
	// elementary variables
	for i = 0; i < p.nInps; i++ {
		// find variable
		iStart = p.nParsVIni + i*p.nParsVNum
		iVar = 0
		for v = 0; v < p.nParsVNum; v++ {
			if p.Values[iStart+v] != 0 {
				iVar += (1 << v)
			}
		}
		// assign variable
		abcutil.Abc_TtCopy(Ifn_ObjTruth(p, i), Ifn_ElemTruth(p, iVar), p.nWords, 0)
	}
	// internal variables
	for i = p.nInps; i < p.nObjs; i++ {
		var nFans int32 = p.Nodes[i].nFanins
		pFans := p.Nodes[i].Fanins[:]
		pTruth := Ifn_ObjTruth(p, i)
		if p.Nodes[i].Type == IFN_DSD_AND {
			abcutil.Abc_TtFill(pTruth, p.nWords)
			for f = 0; f < nFans; f++ {
				abcutil.Abc_TtAnd(pTruth, pTruth, Ifn_ObjTruth(p, pFans[f]), p.nWords, 0)
			}

		} else if p.Nodes[i].Type == IFN_DSD_XOR {
			abcutil.Abc_TtClear(pTruth, p.nWords)
			for f = 0; f < nFans; f++ {
				abcutil.Abc_TtXor(pTruth, pTruth, Ifn_ObjTruth(p, pFans[f]), p.nWords, 0)
			}

		} else if p.Nodes[i].Type == IFN_DSD_MUX {
			assert(nFans == 3, "nFans == 3")
			abcutil.Abc_TtMux(pTruth, Ifn_ObjTruth(p, pFans[0]), Ifn_ObjTruth(p, pFans[1]), Ifn_ObjTruth(p, pFans[2]), p.nWords)

		} else if p.Nodes[i].Type == IFN_DSD_PRIME {
			var nValues int32 = (1 << nFans)
			pTemp := Ifn_ObjTruth(p, p.nObjs)
			abcutil.Abc_TtClear(pTruth, p.nWords)
			for v = 0; v < nValues; v++ {
				if pValues[p.Nodes[i].iFirst+v] == 0 {
					continue
				}
				abcutil.Abc_TtFill(pTemp, p.nWords)
				for f = 0; f < nFans; f++ {
					if ((v >> f) & 1) != 0 {
						abcutil.Abc_TtAnd(pTemp, pTemp, Ifn_ObjTruth(p, pFans[f]), p.nWords, 0)
					} else {
						abcutil.Abc_TtSharp(pTemp, pTemp, Ifn_ObjTruth(p, pFans[f]), p.nWords)
					}
				}
				abcutil.Abc_TtOr(pTruth, pTruth, pTemp, p.nWords)
			}

		} else {
			assert(false, "0")
		}
		//dau.Dau_DsdPrintFromTruth( pTruth, p->nVars );
	}
	return Ifn_ObjTruth(p, p.nObjs-1)
}

/*
*Function*************************************************************

	Synopsis    [Compute more or equal]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ifn_TtComparisonConstr(pTruth []word, nVars int32, fMore int32, fEqual int32) {
	var Cond, Equa, Temp [4]word
	var s_TtElems [8][4]word = [8][4]word{
		{0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA},
		{0xCCCCCCCCCCCCCCCC, 0xCCCCCCCCCCCCCCCC, 0xCCCCCCCCCCCCCCCC, 0xCCCCCCCCCCCCCCCC},
		{0xF0F0F0F0F0F0F0F0, 0xF0F0F0F0F0F0F0F0, 0xF0F0F0F0F0F0F0F0, 0xF0F0F0F0F0F0F0F0},
		{0xFF00FF00FF00FF00, 0xFF00FF00FF00FF00, 0xFF00FF00FF00FF00, 0xFF00FF00FF00FF00},
		{0xFFFF0000FFFF0000, 0xFFFF0000FFFF0000, 0xFFFF0000FFFF0000, 0xFFFF0000FFFF0000},
		{0xFFFFFFFF00000000, 0xFFFFFFFF00000000, 0xFFFFFFFF00000000, 0xFFFFFFFF00000000},
		{0x0000000000000000, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000, 0xFFFFFFFFFFFFFFFF},
		{0x0000000000000000, 0x0000000000000000, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
	}
	var i int32
	var nWords int32 = abcutil.Abc_TtWordNum(2 * nVars)
	assert(nVars > 0 && nVars <= 4, "nVars > 0 && nVars <= 4")
	abcutil.Abc_TtClear(pTruth, nWords)
	abcutil.Abc_TtFill(Equa[:], nWords)
	for i = nVars - 1; i >= 0; i-- {
		if fMore != 0 {
			abcutil.Abc_TtSharp(Cond[:], s_TtElems[2*i+1][:], s_TtElems[2*i+0][:], nWords)
		} else {
			abcutil.Abc_TtSharp(Cond[:], s_TtElems[2*i+0][:], s_TtElems[2*i+1][:], nWords)
		}
		abcutil.Abc_TtAnd(Temp[:], Equa[:], Cond[:], nWords, 0)
		abcutil.Abc_TtOr(pTruth, pTruth, Temp[:], nWords)
		abcutil.Abc_TtXor(Temp[:], s_TtElems[2*i+0][:], s_TtElems[2*i+1][:], nWords, 1)
		abcutil.Abc_TtAnd(Equa[:], Equa[:], Temp[:], nWords, 0)
	}
	if fEqual != 0 {
		abcutil.Abc_TtNot(pTruth, nWords)
	}
}

/*
*Function*************************************************************

	Synopsis    [Adds parameter constraints.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ifn_AddClause(pSat *sat_solver, literals []int32) int32 {
	return ifTuneBoolInt(pSat.AddClause(literals...))
}
func Ifn_NtkAddConstrOne(pSat *sat_solver, vCover *Vec_Int_t, pVars []int32, nVars int32) {
	var RetValue int32
	var k int32
	var c int32
	var Cube int32
	var Literal int32
	var nLits int32
	var pLits [IFN_INS]int32
	for c = 0; c < int32(Vec_IntSize(vCover)); c++ {
		Cube = Vec_IntEntry(vCover, int(c))
		nLits = 0
		for k = 0; k < nVars; k++ {
			Literal = 3 & (Cube >> (k << 1))
			if Literal == 1 { // '0'  -> pos lit
				pLits[nLits] = abcutil.Abc_Var2Lit(pVars[k], 0)
				nLits++
			} else if Literal == 2 { // '1'  -> neg lit
				pLits[nLits] = abcutil.Abc_Var2Lit(pVars[k], 1)
				nLits++
			} else if Literal != 0 {
				assert(false, "0")
			}
		}
		RetValue = Ifn_AddClause(pSat, pLits[:nLits])
		assert((RetValue != 0), "RetValue")
	}
}
func Ifn_NtkAddConstraints(p *Ifn_Ntk_t, pSat *sat_solver) {
	var fAddConstr int32 = 1
	var vCover *Vec_Int_t = Vec_IntAlloc(0)
	var uTruth word = abcutil.Abc_Tt6Stretch(^abcutil.Abc_Tt6Mask(p.nVars), p.nParsVNum)
	assert(p.nParsVNum <= 4, "p.nParsVNum <= 4")
	if uTruth != 0 {
		var i int32
		var k int32
		var pVars [IFN_INS]int32
		var RetValue int32 = ifTuneTruthISOP([]word{uTruth}, p.nParsVNum, vCover)
		assert(RetValue == 0, "RetValue == 0")
		//        dau.Dau_DsdPrintFromTruth( &uTruth, p->nParsVNum );
		// add capacity constraints
		for i = 0; i < p.nInps; i++ {
			for k = 0; k < p.nParsVNum; k++ {
				pVars[k] = p.nParsVIni + i*p.nParsVNum + k
			}
			Ifn_NtkAddConstrOne(pSat, vCover, pVars[:], p.nParsVNum)
		}
	}
	// ordering constraints
	if fAddConstr != 0 && p.nConstr > 0 {
		var pTruth [4]word
		var i int32
		var k int32
		var RetValue int32
		var pVars [2 * IFN_INS]int32
		var fForceDiff int32 = ifTuneBoolInt(p.nVars == p.nInps)
		Ifn_TtComparisonConstr(pTruth[:], p.nParsVNum, fForceDiff, fForceDiff)
		RetValue = ifTuneTruthISOP(pTruth[:], 2*p.nParsVNum, vCover)
		assert(RetValue == 0, "RetValue == 0")
		//        Kit_TruthIsopPrintCover( vCover, 2*p->nParsVNum, 0 );
		for i = 0; i < p.nConstr; i++ {
			var iVar1 int32 = p.pConstr[i] >> 16
			var iVar2 int32 = p.pConstr[i] & 0xFFFF
			for k = 0; k < p.nParsVNum; k++ {
				pVars[2*k+0] = p.nParsVIni + iVar1*p.nParsVNum + k
				pVars[2*k+1] = p.nParsVIni + iVar2*p.nParsVNum + k
			}
			Ifn_NtkAddConstrOne(pSat, vCover, pVars[:], 2*p.nParsVNum)
			//            printf( "added constraint with %d clauses for %d and %d\n", Vec_IntSize(vCover), iVar1, iVar2 );
		}
	}
	Vec_IntFree(vCover)
}

func ifTuneTruthISOP(words []word, nVars int32, cover *Vec_Int_t) int32 {
	truth32 := make([]uint32, len(words)*2)
	for i, word := range words {
		truth32[2*i] = uint32(word)
		truth32[2*i+1] = uint32(word >> 32)
	}
	var cubes []uint32
	complement := kit.Kit_TruthIsop(truth32, nVars, &cubes, 0)
	vec.Vec_IntClear(cover)
	for _, cube := range cubes {
		Vec_IntPush(cover, int32(cube))
	}
	return complement
}

/*
*Function*************************************************************

	Synopsis    [Derive clauses given variable assignment.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ifn_NtkAddClauses(p *Ifn_Ntk_t, pValues []int32, pSat *sat_solver) int32 {
	var i int32
	var f int32
	var v int32
	var nLits int32
	var pLits [IFN_INS + 2]int32
	var pLits2 [IFN_INS + 2]int32
	// assign new variables
	var nSatVars int32 = sat_solver_nvars(pSat)
	//    for ( i = 0; i < p->nObjs-1; i++ )
	for i = 0; i < p.nObjs; i++ {
		p.Nodes[i].Var = nSatVars
		nSatVars++
	}
	//p->Nodes[p->nObjs-1].Var = 0xFFFF;
	sat_solver_setnvars(pSat, nSatVars)
	// verify variable values
	for i = 0; i < p.nVars; i++ {
		assert(pValues[i] != -1, "pValues[i] != -1")
	}
	for i = p.nVars; i < p.nObjs-1; i++ {
		assert(pValues[i] == -1, "pValues[i] == -1")
	}
	assert(pValues[p.nObjs-1] != -1, "pValues[p.nObjs-1] != -1")
	// internal variables
	//printf( "\n" );
	for i = 0; i < p.nInps; i++ {
		var iParStart int32 = p.nParsVIni + i*p.nParsVNum
		for v = 0; v < p.nVars; v++ {
			// add output literal
			pLits[0] = abcutil.Abc_Var2Lit(p.Nodes[i].Var, ifTuneBoolInt(pValues[v] == 0))
			// add clause literals
			for f = 0; f < p.nParsVNum; f++ {
				pLits[f+1] = abcutil.Abc_Var2Lit(iParStart+f, (v>>f)&1)
			}
			if !(Ifn_AddClause(pSat, pLits[:p.nParsVNum+1]) != 0) {
				return 0
			}
		}
	}
	//printf( "\n" );
	for i = p.nInps; i < p.nObjs; i++ {
		var nFans int32 = p.Nodes[i].nFanins
		pFans := p.Nodes[i].Fanins[:]
		if p.Nodes[i].Type == IFN_DSD_AND {
			nLits = 0
			pLits[nLits] = abcutil.Abc_Var2Lit(p.Nodes[i].Var, 0)
			nLits++
			for f = 0; f < nFans; f++ {
				pLits[nLits] = abcutil.Abc_Var2Lit(p.Nodes[pFans[f]].Var, 1)
				nLits++
				// add small clause
				pLits2[0] = abcutil.Abc_Var2Lit(p.Nodes[i].Var, 1)
				pLits2[1] = abcutil.Abc_Var2Lit(p.Nodes[pFans[f]].Var, 0)
				if !(Ifn_AddClause(pSat, pLits2[:2]) != 0) {
					return 0
				}
			}
			// add big clause
			if !(Ifn_AddClause(pSat, pLits[:nLits]) != 0) {
				return 0
			}

		} else if p.Nodes[i].Type == IFN_DSD_XOR {
			var m int32
			var nMints int32 = (1 << (nFans + 1))
			for m = 0; m < nMints; m++ {
				// skip even
				var Count int32
				for v = 0; v <= nFans; v++ {
					Count += ((m >> v) & 1)
				}
				if (Count & 1) == 0 {
					continue
				}
				// generate minterm
				pLits[0] = abcutil.Abc_Var2Lit(p.Nodes[i].Var, (m>>nFans)&1)
				for v = 0; v < nFans; v++ {
					pLits[v+1] = abcutil.Abc_Var2Lit(p.Nodes[pFans[v]].Var, (m>>v)&1)
				}
				if !(Ifn_AddClause(pSat, pLits[:nFans+1]) != 0) {
					return 0
				}
			}

		} else if p.Nodes[i].Type == IFN_DSD_MUX {
			pLits[0] = abcutil.Abc_Var2Lit(p.Nodes[i].Var, 0)
			pLits[1] = abcutil.Abc_Var2Lit(p.Nodes[pFans[0]].Var, 1) // ctrl
			pLits[2] = abcutil.Abc_Var2Lit(p.Nodes[pFans[1]].Var, 1)
			if !(Ifn_AddClause(pSat, pLits[:3]) != 0) {
				return 0
			}

			pLits[0] = abcutil.Abc_Var2Lit(p.Nodes[i].Var, 1)
			pLits[1] = abcutil.Abc_Var2Lit(p.Nodes[pFans[0]].Var, 1) // ctrl
			pLits[2] = abcutil.Abc_Var2Lit(p.Nodes[pFans[1]].Var, 0)
			if !(Ifn_AddClause(pSat, pLits[:3]) != 0) {
				return 0
			}

			pLits[0] = abcutil.Abc_Var2Lit(p.Nodes[i].Var, 0)
			pLits[1] = abcutil.Abc_Var2Lit(p.Nodes[pFans[0]].Var, 0) // ctrl
			pLits[2] = abcutil.Abc_Var2Lit(p.Nodes[pFans[2]].Var, 1)
			if !(Ifn_AddClause(pSat, pLits[:3]) != 0) {
				return 0
			}

			pLits[0] = abcutil.Abc_Var2Lit(p.Nodes[i].Var, 1)
			pLits[1] = abcutil.Abc_Var2Lit(p.Nodes[pFans[0]].Var, 0) // ctrl
			pLits[2] = abcutil.Abc_Var2Lit(p.Nodes[pFans[2]].Var, 0)
			if !(Ifn_AddClause(pSat, pLits[:3]) != 0) {
				return 0
			}

		} else if p.Nodes[i].Type == IFN_DSD_PRIME {
			var nValues int32 = (1 << nFans)
			var iParStart int32 = p.Nodes[i].iFirst
			for v = 0; v < nValues; v++ {
				nLits = 0
				if pValues[i] == -1 {
					pLits[nLits] = abcutil.Abc_Var2Lit(p.Nodes[i].Var, 0)
					pLits2[nLits] = abcutil.Abc_Var2Lit(p.Nodes[i].Var, 1)
					nLits++
				}
				for f = 0; f < nFans; f++ {
					lit := abcutil.Abc_Var2Lit(p.Nodes[pFans[f]].Var, (v>>f)&1)
					pLits[nLits] = lit
					pLits2[nLits] = lit
					nLits++
				}
				pLits[nLits] = abcutil.Abc_Var2Lit(iParStart+v, 1)
				pLits2[nLits] = abcutil.Abc_Var2Lit(iParStart+v, 0)
				nLits++
				if pValues[i] != 0 {
					if !(Ifn_AddClause(pSat, pLits2[:nLits]) != 0) {
						return 0
					}
				}
				if pValues[i] != 1 {
					if !(Ifn_AddClause(pSat, pLits[:nLits]) != 0) {
						return 0
					}
				}
			}

		} else {
			assert(false, "0")
		}
		//printf( "\n" );
	}
	// add last clause (not needed if the root node is IFN_DSD_PRIME)
	pLits[0] = abcutil.Abc_Var2Lit(p.Nodes[p.nObjs-1].Var, ifTuneBoolInt(pValues[p.nObjs-1] == 0))
	if !(Ifn_AddClause(pSat, pLits[:1]) != 0) {
		return 0
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Returns the minterm number for which there is a mismatch.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ifn_NtkMatchPrintStatus(p *sat_solver, Iter int32, status int32, iMint int32, Value int32, clk abctime) {
	fmt.Printf("Iter = %5d  ", Iter)
	fmt.Printf("Mint = %5d  ", iMint)
	fmt.Printf("Value = %2d  ", Value)
	fmt.Printf("Var = %6d  ", sat_solver_nvars(p))
	fmt.Printf("Cla = %6d  ", sat_solver_nclauses(p))
	fmt.Printf("Conf = %6d  ", sat_solver_nconflicts(p))
	if status == int32(bsat.Unsatisfiable) {
		fmt.Printf("status = unsat")
	} else if status == int32(bsat.Satisfiable) {
		fmt.Printf("status = sat  ")
	} else {
		fmt.Printf("status = undec")
	}
	Abc_PrintTime(1, "Time", clk)
}
func Ifn_NtkMatchPrintConfig(p *Ifn_Ntk_t, pSat *sat_solver) {
	var v, i int32
	for v = p.nObjs; v < p.nPars; v++ {
		for i = p.nInps; i < p.nObjs; i++ {
			if p.Nodes[i].Type == IFN_DSD_PRIME && int32(p.Nodes[i].iFirst) == v {
				break
			}
		}
		if i < p.nObjs {
			fmt.Printf(" ")
		} else if v >= p.nParsVIni && (v-p.nParsVIni)%p.nParsVNum == 0 {
			fmt.Printf(" %d=", (v-p.nParsVIni)/p.nParsVNum)
		}
		fmt.Printf("%d", sat_solver_var_value(pSat, v))
	}
}
func Ifn_NtkMatchCollectPerm(p *Ifn_Ntk_t, pSat *sat_solver) word {
	var Perm word
	permWords := []word{0}
	var i, v, Mint int32
	assert(p.nParsVNum <= 4, "p.nParsVNum <= 4")
	for i = 0; i < p.nInps; i++ {
		Mint = 0
		for v = 0; v < p.nParsVNum; v++ {
			if sat_solver_var_value(pSat, p.nParsVIni+i*p.nParsVNum+v) != 0 {
				Mint |= (1 << v)
			}
		}
		abcutil.Abc_TtSetHex(permWords, i, Mint)
	}
	Perm = permWords[0]
	return Perm
}
func Ifn_NtkMatchCollectConfig(p *Ifn_Ntk_t, pSat *sat_solver, pConfig []word) {
	var i, v, Mint int32
	assert(p.nParsVNum <= 4, "p.nParsVNum <= 4")
	for i = 0; i < p.nInps; i++ {
		Mint = 0
		for v = 0; v < p.nParsVNum; v++ {
			if sat_solver_var_value(pSat, p.nParsVIni+i*p.nParsVNum+v) != 0 {
				Mint |= (1 << v)
			}
		}
		abcutil.Abc_TtSetHex(pConfig, i, Mint)
	}
	for v = p.nObjs; v < p.nParsVIni; v++ {
		if sat_solver_var_value(pSat, v) != 0 {
			abcutil.Abc_TtSetBit(pConfig, 64+v-p.nObjs)
		}
	}
}
func Ifn_NtkMatchPrintPerm(Perm word, nInps int32) {
	var i int32
	assert(nInps <= 16, "nInps <= 16")
	for i = 0; i < nInps; i++ {
		fmt.Printf("%c", 'a'+abcutil.Abc_TtGetHex([]word{Perm}, i))
	}
	fmt.Printf("\n")
}
func Ifn_NtkMatch(p *Ifn_Ntk_t, pTruth []word, nVars int32, nConfls int32, fVerbose int32, fVeryVerbose int32, pConfig []word) int32 {
	var pTruth1 []word
	var RetValue int32
	var nIterMax int32 = (1 << nVars)
	var i, v, status, iMint int32
	var clk abctime = Abc_Clock()
	//    abctime clkTru = 0, clkSat = 0, clk2;
	var pSat *sat_solver
	if nVars == 0 {
		return 1
	}
	pSat = sat_solver_new()
	Ifn_Prepare(p, pTruth, nVars)
	sat_solver_setnvars(pSat, p.nPars)
	Ifn_NtkAddConstraints(p, pSat)
	if fVeryVerbose != 0 {
		Ifn_NtkMatchPrintStatus(pSat, 0, int32(bsat.Satisfiable), -1, -1, Abc_Clock()-clk)
	}
	if pConfig != nil {
		assert(len(pConfig) > 0 && pConfig[0] == 0, "configuration starts empty")
	}
	for i = 0; i < nIterMax; i++ {
		// get variable assignment
		for v = 0; v < p.nObjs; v++ {
			if v < p.nVars {
				p.Values[v] = (iMint >> v) & 1
			} else {
				p.Values[v] = -1
			}
		}
		p.Values[p.nObjs-1] = abcutil.Abc_TtGetBit(pTruth, iMint)
		// derive clauses
		if !(Ifn_NtkAddClauses(p, p.Values[:], pSat) != 0) {
			break
		}
		// find assignment of parameters
		//        clk2 = Abc_Clock();
		pSat.SetLimits(int64(nConfls), 0)
		status = int32(pSat.Solve())
		//        clkSat += Abc_Clock() - clk2;
		if fVeryVerbose != 0 {
			Ifn_NtkMatchPrintStatus(pSat, i+1, status, iMint, p.Values[p.nObjs-1], Abc_Clock()-clk)
		}
		if status != int32(bsat.Satisfiable) {
			break
		}
		// collect assignment
		for v = p.nObjs; v < p.nPars; v++ {
			p.Values[v] = sat_solver_var_value(pSat, v)
		}
		// find truth table
		//        clk2 = Abc_Clock();
		pTruth1 = Ifn_NtkDeriveTruth(p, p.Values[:])
		//        clkTru += Abc_Clock() - clk2;
		abcutil.Abc_TtXor(pTruth1, pTruth1, p.pTruth, p.nWords, 0)
		// find mismatch if present
		iMint = abcutil.Abc_TtFindFirstBit(pTruth1, p.nVars)
		if iMint == -1 {
			if pConfig != nil {
				Ifn_NtkMatchCollectConfig(p, pSat, pConfig)
			}
			/*
			   if ( pPerm )
			   {
			       int RetValue = Ifn_ManSatFindCofigBitsTest( p, pTruth, nVars, *pPerm );
			       Ifn_NtkMatchPrintPerm( *pPerm, p->nInps );
			       if ( RetValue == 0 )
			           printf( "Verification failed.\n" );
			   }
			*/
			RetValue = 1
			break
		}
	}
	assert(i < nIterMax, "i < nIterMax")
	if fVerbose != 0 {
		label := "no "
		if RetValue != 0 {
			label = "yes"
		}
		fmt.Printf("%s  Iter =%4d. Confl = %6d.  ", label, i, sat_solver_nconflicts(pSat))
		if RetValue != 0 {
			Ifn_NtkMatchPrintConfig(p, pSat)
		}
		fmt.Printf("\n")
	}
	sat_solver_delete(pSat)
	//    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
	//    Abc_PrintTime( 1, "Sat", clkSat );
	//    Abc_PrintTime( 1, "Tru", clkTru );
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ifn_NtkRead() {
	var nVars int32 = 8
	//    word * pTruth = dau.Dau_DsdToTruth( "(abcdefghi)", nVars );
	pTruth := dau.Dau_DsdToTruth("1008{(1008{(ab)cde}f)ghi}", nVars)
	//    word * pTruth = dau.Dau_DsdToTruth( "18{(1008{(ab)cde}f)gh}", nVars );
	//    word * pTruth = dau.Dau_DsdToTruth( "1008{(1008{[ab]cde}f)ghi}", nVars );
	//    word * pTruth = dau.Dau_DsdToTruth( "(abcd)", nVars );
	//    word * pTruth = dau.Dau_DsdToTruth( "(abc)", nVars );
	//    word * pTruth = dau.Dau_DsdToTruth( "18{(1008{(ab)cde}f)gh}", nVars );
	//    char * pStr = "e={abc};f={ed};";
	//    char * pStr = "d={ab};e={cd};";
	//    char * pStr = "j=(ab);k={jcde};l=(kf);m={lghi};";
	//    char * pStr = "i={abc};j={ide};k={ifg};l={jkh};";
	//    char * pStr = "h={abcde};i={abcdf};j=<ghi>;";
	//    char * pStr = "g=<abc>;h=<ade>;i={fgh};";
	//    char * pStr = "i=<abc>;j=(def);k=[gh];l={ijk};";
	var pStr string = "{({(ab)cde}f)ghi};AB;CD;DE;GH;HI"
	var p *Ifn_Ntk_t = Ifn_NtkParse(pStr)
	if p == nil {
		return
	}
	Ifn_NtkPrint(p)
	dau.Dau_DsdPrintFromTruth(pTruth, nVars)
	// get the given function
	config := make([]word, IFN_PAR/64)
	Ifn_NtkMatch(p, pTruth, nVars, 0, 1, 1, config)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
