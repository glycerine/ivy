package bac

import "fmt"

/**CFile****************************************************************

  FileName    [bacNtk.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Netlist manipulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacNtk.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bac.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Bac_Pair_t struct {
    Type Bac_ObjType_t;
    pName         string;
    pSymb         string;
}
var s_Pref string = "ABC_";
var s_Types [BAC_BOX_UNKNOWN]Bac_Pair_t = [BAC_BOX_UNKNOWN]Bac_Pair_t{ 
    { BAC_OBJ_NONE,     "NONE",    nil    },
    { BAC_OBJ_PI,       "PI",      nil    },
    { BAC_OBJ_PO,       "PO",      nil    },
    { BAC_OBJ_BI,       "BI",      nil    },
    { BAC_OBJ_BO,       "BO",      nil    },
    { BAC_OBJ_BOX,      "BOX",     nil    },
    
    { BAC_BOX_CF,       "CF",      "o"     },
    { BAC_BOX_CT,       "CT",      "o"     }, 
    { BAC_BOX_CX,       "CX",      "o"     }, 
    { BAC_BOX_CZ,       "CZ",      "o"     },
    { BAC_BOX_BUF,      "BUF",     "ao"    },
    { BAC_BOX_INV,      "INV",     "ao"    },
    { BAC_BOX_AND,      "AND",     "abo"   },
    { BAC_BOX_NAND,     "NAND",    "abo"   },
    { BAC_BOX_OR,       "OR",      "abo"   },
    { BAC_BOX_NOR,      "NOR",     "abo"   },
    { BAC_BOX_XOR,      "XOR",     "abo"   },
    { BAC_BOX_XNOR,     "XNOR",    "abo"   },
    { BAC_BOX_SHARP,    "SHARP",   "abo"   },
    { BAC_BOX_SHARPL,   "SHARPL",  "abo"   },
    { BAC_BOX_MUX,      "MUX",     "cabo"  },
    { BAC_BOX_MAJ,      "MAJ",     "abco"  },
    
    { BAC_BOX_RAND,     "RAND",    "ao"    },
    { BAC_BOX_RNAND,    "RNAND",   "ao"    },
    { BAC_BOX_ROR,      "ROR",     "ao"    },
    { BAC_BOX_RNOR,     "RNOR",    "ao"    },
    { BAC_BOX_RXOR,     "RXOR",    "ao"    },
    { BAC_BOX_RXNOR,    "RXNOR",   "ao"    },
    
    { BAC_BOX_LAND,     "LAND",    "abo"   },
    { BAC_BOX_LNAND,    "LNAND",   "abo"   },
    { BAC_BOX_LOR,      "LOR",     "abo"   },
    { BAC_BOX_LNOR,     "LNOR",    "abo"   },
    { BAC_BOX_LXOR,     "LXOR",    "abo"   },
    { BAC_BOX_LXNOR,    "LXNOR",   "abo"   },
    
    { BAC_BOX_NMUX,     "NMUX",    "abo"   },
    { BAC_BOX_SEL,      "SEL",     "abo"   },
    { BAC_BOX_PSEL,     "PSEL",    "iabo"  },
    { BAC_BOX_ENC,      "ENC",     "ao"    },
    { BAC_BOX_PENC,     "PENC",    "ao"    },
    { BAC_BOX_DEC,      "DEC",     "ao"    },
    { BAC_BOX_EDEC,     "EDEC",    "abo"   },
    
    { BAC_BOX_ADD,      "ADD",     "iabso" },
    { BAC_BOX_SUB,      "SUB",     "abo"   },
    { BAC_BOX_MUL,      "MUL",     "abo"   },
    { BAC_BOX_DIV,      "DIV",     "abo"   },
    { BAC_BOX_MOD,      "MOD",     "abo"   },
    { BAC_BOX_REM,      "REM",     "abo"   },
    { BAC_BOX_POW,      "POW",     "abo"   },
    { BAC_BOX_MIN,      "MIN",     "ao"    },
    { BAC_BOX_ABS,      "ABS",     "ao"    },

    { BAC_BOX_LTHAN,    "LTHAN",   "iabo"  },
    { BAC_BOX_LETHAN,   "LETHAN",  "abo"   },
    { BAC_BOX_METHAN,   "METHAN",  "abo"   },
    { BAC_BOX_MTHAN,    "MTHAN",   "abo"   },
    { BAC_BOX_EQU,      "EQU",     "abo"   },
    { BAC_BOX_NEQU,     "NEQU",    "abo"   },
    
    { BAC_BOX_SHIL,     "SHIL",    "abo"   },
    { BAC_BOX_SHIR,     "SHIR",    "abo"   },
    { BAC_BOX_ROTL,     "ROTL",    "abo"   },
    { BAC_BOX_ROTR,     "ROTR",    "abo"   },

    { BAC_BOX_GATE,     "GATE",    "io"    },
    { BAC_BOX_LUT,      "LUT",     "io"    },
    { BAC_BOX_ASSIGN,   "ASSIGN",  "abo"   },
    
    { BAC_BOX_TRI,      "TRI",     "abo"   },
    { BAC_BOX_RAM,      "RAM",     "eadro" },
    { BAC_BOX_RAMR,     "RAMR",    "eamo"  },
    { BAC_BOX_RAMW,     "RAMW",    "eado"  },
    { BAC_BOX_RAMWC,    "RAMWC",   "ceado" },
    { BAC_BOX_RAMBOX,   "RAMBOX",  "io"    },
    
    { BAC_BOX_LATCH,    "LATCH",   "dvsgq" },
    { BAC_BOX_LATCHRS,  "LATCHRS", "dsrgq" },
    { BAC_BOX_DFF,      "DFF",     "dvscq" },
    { BAC_BOX_DFFRS,    "DFFRS",   "dsrcq" },
}; 
func Bac_GetTypeId(Type Bac_ObjType_t) int32 {
    var i int32;
    for  i = 1; i < BAC_BOX_UNKNOWN; i++  {
        if s_Types[i].Type == Type {
            return i; }
    }
    return -1;
}
func Bac_ManSetupTypes(pNames *string, pSymbs *string) {
    var Type int32;
    for  Type = 1; Type < BAC_BOX_UNKNOWN; Type++  {
        var Id int32  = Bac_GetTypeId( (Bac_ObjType_t)Type );
        pNames[Type] = s_Types[Id].pName;
        pSymbs[Type] = s_Types[Id].pSymb;
    }
}

var spStatic_Bac_NtkGenerateName_Buffer string // was [100]byte in C

func Bac_NtkGenerateName(p *Bac_Ntk_t, Type Bac_ObjType_t, vBits *Vec_Int_t) string {
     
    var pTemp string; var i, Bits int32;
    var pName string  = Bac_ManPrimName( p.pDesign, Type );
    var pSymb string  = Bac_ManPrimSymb( p.pDesign, Type );
    assert(Vec_IntSize(vBits) == int32(len(pSymb)), "Vec_IntSize(vBits) == int32(strlen(pSymb))");
    spStatic_Bac_NtkGenerateName_Buffer = fmt.Sprintf("%s%s_", s_Pref, pName);
    pTemp = spStatic_Bac_NtkGenerateName_Buffer + strlen(spStatic_Bac_NtkGenerateName_Buffer);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vBits)) && (((Bits) = Vec_IntEntry(vBits, i)), 1); i++  {
        pTemp = fmt.Sprintf("%c%d", pSymb[i], Bits);
        pTemp += len(pTemp);
    }
    //Vec_IntPrint( vBits );
    //printf( "%s\n", Buffer );
    return spStatic_Bac_NtkGenerateName_Buffer;
}

func Bac_NameToType(pName string) Bac_ObjType_t {
    var i int32;
    if strncmp(pName, s_Pref, len(s_Pref)) {
        return BAC_OBJ_NONE; }
    pName += len(s_Pref);
    for  i = 1; i < BAC_BOX_UNKNOWN; i++  {
        if !strncmp(pName, s_Types[i].pName, len(s_Types[i].pName)) {
            return s_Types[i].Type; }
    }
    return BAC_OBJ_NONE;
}
var spStatic_Bac_NameToRanges_Bits Vec_Int_t;
var spStatic_Bac_NameToRanges_vBits *Vec_Int_t = &spStatic_Bac_NameToRanges_Bits;
var spStatic_Bac_NameToRanges_pArray [10]int32;
func Bac_NameToRanges(pName string) *Vec_Int_t {
    
    
    var pTemp string; 
    var Num, Count int32;
    // initialize array
    spStatic_Bac_NameToRanges_vBits.pArray = spStatic_Bac_NameToRanges_pArray;
    spStatic_Bac_NameToRanges_vBits.nSize = 0;
    spStatic_Bac_NameToRanges_vBits.nCap = 10;
    // check the name
    assert(!strncmp(pName, s_Pref, len(s_Pref)), "!strncmp(pName, s_Pref, strlen(s_Pref))");
    for  pTemp = pName; *pTemp && !Bac_CharIsDigit(*pTemp); pTemp++  {}
    assert(Bac_CharIsDigit(*pTemp), "Bac_CharIsDigit(*pTemp)");
    for  ; *pTemp; pTemp++  {
        if Bac_CharIsDigit(*pTemp) {
            Num = 10 * Num + *pTemp - '0';
        } else {
            Vec_IntPush( spStatic_Bac_NameToRanges_vBits, Num );
Count += Num;
Num = 0; }
    }
    assert(Num > 0, "Num > 0");
    Vec_IntPush( spStatic_Bac_NameToRanges_vBits, Num );  Count += Num;
    assert(Vec_IntSize(spStatic_Bac_NameToRanges_vBits) <= 10, "Vec_IntSize(spStatic_Bac_NameToRanges_vBits) <= 10");
    return spStatic_Bac_NameToRanges_vBits;
}


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Replaces fanin iOld by iNew in all fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_NtkUpdateFanout(p *Bac_Ntk_t, iOld int32, iNew int32) {
    var iCo int32;
    assert(Bac_ObjIsCi(p, iOld), "Bac_ObjIsCi(p, iOld)");
    assert(Bac_ObjIsCi(p, iNew), "Bac_ObjIsCi(p, iNew)");
    // inlining of macro Bac_ObjForEachFanout
    for  iCo = Bac_ObjFanout(p, iOld); (iCo != 0); iCo = Bac_ObjNextFanout(p, iCo)  {
        assert(Bac_ObjFanin(p, iCo) == iOld, "Bac_ObjFanin(p, iCo) == iOld");
        Bac_ObjCleanFanin( p, iCo );
        Bac_ObjSetFanin( p, iCo, iNew );
    }
    Bac_ObjSetFanout( p, iNew, Bac_ObjFanout(p, iOld) );
    Bac_ObjSetFanout( p, iOld, 0 );
}

/**Function*************************************************************

  Synopsis    [Derives fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_NtkDeriveFanout(p *Bac_Ntk_t) {
    var iCi, iCo int32;
    assert(!Bac_NtkHasFanouts(p), "!Bac_NtkHasFanouts(p)");
    Bac_NtkStartFanouts( p );
    // inlining of macro Bac_NtkForEachCo
    for  iCo = 0; (iCo < Bac_NtkObjNum(p)); iCo++  { if !Bac_ObjIsCo(p, iCo) {} else {
        assert(!Bac_ObjNextFanout(p, iCo), "!Bac_ObjNextFanout(p, iCo)");
        iCi = Bac_ObjFanin(p, iCo);
        if Bac_ObjFanout(p, iCi) {
            Bac_ObjSetNextFanout( p, Bac_ObjFanout(p, iCi), iCo ); }
        Bac_ObjSetFanout( p, iCi, iCo );
    }
    }
    // inlining of macro Bac_NtkForEachCo
    for  iCo = 0; (iCo < Bac_NtkObjNum(p)); iCo++  { if !Bac_ObjIsCo(p, iCo) {} else
        if !Bac_ObjNextFanout(p, iCo) {
            Bac_ObjSetFanout( p, Bac_ObjFanin(p, iCo), iCo ); }
    }
}
func Bac_ManDeriveFanout(p *Bac_Man_t) {
    var pNtk *Bac_Ntk_t; var i int32;
    // inlining of macro Bac_ManForEachNtk
    for  i = 1; (i <= Bac_ManNtkNum(p)) && (((pNtk) = Bac_ManNtk(p, i)), 1); i++  { 
        Bac_NtkDeriveFanout( pNtk );
    }
}

/**Function*************************************************************

  Synopsis    [Assigns word-level names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_ManAssignInternTwo(p *Bac_Ntk_t, iNum int32, nDigits int32, pPref string, vMap *Vec_Int_t) int32 {
    var Buffer string // was [16]byte in C
 var i, NameId int32;
    for  i = 0; !(NameId != 0) || Vec_IntEntry(vMap, NameId); i++  {
        if i == 0 {
            Buffer = fmt.Sprintf("%s%0*d", pPref, nDigits, iNum);
        } else {
            Buffer = fmt.Sprintf("%s%0*d_%d", pPref, nDigits, iNum, i); }
        NameId = Abc_NamStrFindOrAdd( p.pDesign.pStrs, Buffer, nil );
    }
    Vec_IntWriteEntry( vMap, NameId, 1 );
    return NameId;
}
func Bac_ManAssignCountNames(p *Bac_Ntk_t) int32 {
    var i, iObj, iBox, Count int32;
    // inlining of macro Bac_NtkForEachPiMain
    for  i = 0; (i < Bac_NtkPiNum(p))  && (((iObj) = Bac_NtkPi(p, i)), 1); i++  {  if Bac_ObjBit(p, iObj) {} else
        if !Bac_ObjNameInt(p, iObj) {
            Count++; }
    }
    // inlining of macro Bac_NtkForEachBox
    for  iBox = 0; (iBox < Bac_NtkObjNum(p)); iBox++  { if !Bac_ObjIsBox(p, iBox) {} else {
        // inlining of macro Bac_BoxForEachBoMain
        for  iObj, i = iBox + 1, 0; iObj < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iObj); iObj, i = iObj+1, i+1  {  if Bac_ObjBit(p, iObj) {} else
            if !Bac_ObjNameInt(p, iObj) {
                Count++; }
        } }
    }
    return Count;
}
func Bac_ManAssignInternWordNamesNtk(p *Bac_Ntk_t, vMap *Vec_Int_t) {
    var k int32;
    var iObj int32;
    var iTerm int32;
    var iName int32 = -1;
    var iBit int32 = -1;
    var nDigits int32;
    var nPis int32;
    var nPos int32;
    var nNames int32 = 1;
    // start names
    if !Bac_NtkHasNames(p) {
        Bac_NtkStartNames(p); }
    nDigits = Abc_Base10Log( Bac_ManAssignCountNames(p) );
    // populate map with the currently used names
    // inlining of macro Bac_NtkForEachCi
    for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsCi(p, iObj) {} else
        if Bac_ObjNameInt(p, iObj) {
            Vec_IntWriteEntry( vMap, Bac_ObjNameId(p, iObj), 1 ); }
    }
    // inlining of macro Bac_NtkForEachBox
    for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsBox(p, iObj) {} else
        if Bac_ObjNameInt(p, iObj) {
            Vec_IntWriteEntry( vMap, Bac_ObjNameId(p, iObj), 1 ); }
    }
    // assign CI names
    // inlining of macro Bac_NtkForEachCi
    for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsCi(p, iObj) {} else {
        if Bac_ObjNameInt(p, iObj) {
            iName = -1;
            iBit = -1;
            continue;
        }
        if Bac_ObjBit(p, iObj) {
            assert(iBit > 0, "iBit > 0");
            Bac_ObjSetName( p, iObj, Abc_Var2Lit2(iBit++, BAC_NAME_INDEX) );
        
        } else {
            //int Type = Bac_ObjType(p, iObj);
            var Range int32;
            if Bac_ObjIsPi(p, iObj) {
                Range = Bac_ObjPiRange(p, iObj);
            } else {
                Range = Bac_BoxBoRange(p, iObj);
            }
            iName = Bac_ManAssignInternTwo( p, nNames++, nDigits, (char*)(Bac_ObjIsPi(p, iObj) ? "i":"n"), vMap );
            if Range == 1 {
                Bac_ObjSetName( p, iObj, Abc_Var2Lit2(iName, BAC_NAME_BIN) );
            } else {
                Bac_ObjSetName( p, iObj, Abc_Var2Lit2(iName, BAC_NAME_WORD) ); }
            iBit = 1;
        }
    }
    }
    // transfer names to the interface
    if Bac_NtkInfoNum(p) {
        for  k = 0; k < Bac_NtkInfoNum(p); k++  {
            //char * pName = Bac_NtkName(p);
            if Bac_NtkInfoType(p, k) == 1 { // PI
            
                iObj = Bac_NtkPi(p, nPis);
                assert(!Bac_ObjBit(p, iObj), "!Bac_ObjBit(p, iObj)");
                assert(Bac_ObjNameType(p, iObj) <= BAC_NAME_WORD, "Bac_ObjNameType(p, iObj) <= BAC_NAME_WORD");
                Bac_NtkSetInfoName( p, k, Abc_Var2Lit2(Bac_ObjNameId(p, iObj), 1) );
                nPis += Bac_NtkInfoRange(p, k);
            
            } else if Bac_NtkInfoType(p, k) == 2 { // PO
            
                iObj = Bac_NtkPo(p, nPos);
                assert(!Bac_ObjBit(p, iObj), "!Bac_ObjBit(p, iObj)");
                iObj = Bac_ObjFanin(p, iObj);
                assert(Bac_ObjNameType(p, iObj) <= BAC_NAME_WORD, "Bac_ObjNameType(p, iObj) <= BAC_NAME_WORD");
                Bac_NtkSetInfoName( p, k, Abc_Var2Lit2(Bac_ObjNameId(p, iObj), 2) );
                nPos += Bac_NtkInfoRange(p, k);
            
            } else { assert(false, "0"); }
        }
        assert(nPis == Bac_NtkPiNum(p), "nPis == Bac_NtkPiNum(p)");
        assert(nPos == Bac_NtkPoNum(p), "nPos == Bac_NtkPoNum(p)");
    }
    // assign instance names
    nDigits = Abc_Base10Log( Bac_NtkObjNum(p) );
    // inlining of macro Bac_NtkForEachBox
    for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsBox(p, iObj) {} else
        if !Bac_ObjNameInt(p, iObj) {
            iName = Bac_ManAssignInternTwo( p, iObj, nDigits, "g", vMap );
            Bac_ObjSetName( p, iObj, Abc_Var2Lit2(iName, BAC_NAME_BIN) );
        }
    }
    // unmark all names
    // inlining of macro Bac_NtkForEachPi
    for  k = 0; (k < Bac_NtkPiNum(p))  && (((iObj) = Bac_NtkPi(p, k)), 1); k++  { 
        if Bac_ObjNameType(p, iObj) <= BAC_NAME_WORD {
            Vec_IntWriteEntry( vMap, Bac_ObjNameId(p, iObj), 0 ); }
    }
    // inlining of macro Bac_NtkForEachBox
    for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsBox(p, iObj) {} else {
        Vec_IntWriteEntry( vMap, Bac_ObjNameId(p, iObj), 0 );
        // inlining of macro Bac_BoxForEachBo
        for  iTerm, k = iObj + 1, 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm, k = iTerm+1, k+1  {
            if Bac_ObjNameType(p, iTerm) <= BAC_NAME_WORD {
                Vec_IntWriteEntry( vMap, Bac_ObjNameId(p, iTerm), 0 ); }
        }
    }
    }
//    printf( "Generated %d word-level names.\n", nNames-1 );
}
func Bac_ManAssignInternWordNames(p *Bac_Man_t) {
    var vMap *Vec_Int_t  = Vec_IntStart( 2*Bac_ManObjNum(p) );
    var pNtk *Bac_Ntk_t; var i int32;
    // inlining of macro Bac_ManForEachNtk
    for  i = 1; (i <= Bac_ManNtkNum(p)) && (((pNtk) = Bac_ManNtk(p, i)), 1); i++  { 
        Bac_ManAssignInternWordNamesNtk( pNtk, vMap );
    }
    assert(Vec_IntCountEntry(vMap, 0) == Vec_IntSize(vMap), "Vec_IntCountEntry(vMap, 0) == Vec_IntSize(vMap)");
    Vec_IntFree( vMap );
}


/**Function*************************************************************

  Synopsis    [Count number of objects after collapsing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_ManClpObjNum_rec(p *Bac_Ntk_t) int32 {
    var i, Counter int32;
    if p.Count >= 0 {
        return p.Count; }
    // inlining of macro Bac_NtkForEachBox
    for  i = 0; (i < Bac_NtkObjNum(p)); i++  { if !Bac_ObjIsBox(p, i) {} else {
        Counter += Bac_ObjIsBoxUser(p, i) ? Bac_ManClpObjNum_rec( Bac_BoxNtk(p, i) ) + 3*Bac_BoxBoNum(p, i) : Bac_BoxSize(p, i); }
    }
    return (p.Count = Counter);
}
func Bac_ManClpObjNum(p *Bac_Man_t) int32 {
    var pNtk *Bac_Ntk_t; var i int32;
    // inlining of macro Bac_ManForEachNtk
    for  i = 1; (i <= Bac_ManNtkNum(p)) && (((pNtk) = Bac_ManNtk(p, i)), 1); i++  { 
        pNtk.Count = -1;
    }
    return Bac_NtkPioNum( Bac_ManRoot(p) ) + Bac_ManClpObjNum_rec( Bac_ManRoot(p) );
}

/**Function*************************************************************

  Synopsis    [Collects boxes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_NtkDfs_rec(p *Bac_Ntk_t, iObj int32, vBoxes *Vec_Int_t) {
    var k, iFanin int32;
    if Bac_ObjIsBo(p, iObj) == 1 {
        Bac_NtkDfs_rec( p, Bac_ObjFanin(p, iObj), vBoxes );
        return;
    }
    assert(Bac_ObjIsPi(p, iObj) || Bac_ObjIsBox(p, iObj), "Bac_ObjIsPi(p, iObj) || Bac_ObjIsBox(p, iObj)");
    if Bac_ObjCopy(p, iObj) > 0 { // visited
        return; }
    Bac_ObjSetCopy( p, iObj, 1 );
    // inlining of macro Bac_BoxForEachFanin
    for  k = 0; iObj - 1 - k >= 0 && Bac_ObjIsBi(p, iObj - 1 - k) && (((iFanin) = Bac_BoxFanin(p, iObj, k)), 1); k++  {
        Bac_NtkDfs_rec( p, iFanin, vBoxes );
    }
    Vec_IntPush( vBoxes, iObj );
}
func Bac_NtkDfs(p *Bac_Ntk_t) *Vec_Int_t {
    var i, iObj int32;
    var vBoxes *Vec_Int_t  = Vec_IntAlloc( Bac_NtkBoxNum(p) );
    Bac_NtkStartCopies( p ); // -1 = not visited; 1 = finished
    // inlining of macro Bac_NtkForEachPi
    for  i = 0; (i < Bac_NtkPiNum(p))  && (((iObj) = Bac_NtkPi(p, i)), 1); i++  { 
        Bac_ObjSetCopy( p, iObj, 1 );
    }
    // inlining of macro Bac_NtkForEachPo
    for  i = 0; (i < Bac_NtkPoNum(p))  && (((iObj) = Bac_NtkPo(p, i)), 1); i++  { 
        Bac_NtkDfs_rec( p, Bac_ObjFanin(p, iObj), vBoxes );
    }
    return vBoxes;
}

/**Function*************************************************************

  Synopsis    [Collects user boxes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_NtkDfsUserBoxes_rec(p *Bac_Ntk_t, iObj int32, vBoxes *Vec_Int_t) int32 {
    var k, iFanin int32;
    assert(Bac_ObjIsBoxUser(p, iObj), "Bac_ObjIsBoxUser(p, iObj)");
    if Bac_ObjCopy(p, iObj) == 1 { // visited
        return 1; }
    if Bac_ObjCopy(p, iObj) == 0 { // loop
        return 0; }
    Bac_ObjSetCopy( p, iObj, 0 );
    // inlining of macro Bac_BoxForEachFanin
    for  k = 0; iObj - 1 - k >= 0 && Bac_ObjIsBi(p, iObj - 1 - k) && (((iFanin) = Bac_BoxFanin(p, iObj, k)), 1); k++  {
        if Bac_ObjIsBo(p, iFanin) && Bac_ObjIsBoxUser(p, Bac_ObjFanin(p, iFanin)) {
            if !(Bac_NtkDfsUserBoxes_rec( p, Bac_ObjFanin(p, iFanin), vBoxes ) != 0) {
                return 0; } }
    }
    Vec_IntPush( vBoxes, iObj );
    Bac_ObjSetCopy( p, iObj, 1 );
    return 1;
}
func Bac_NtkDfsUserBoxes(p *Bac_Ntk_t) int32 {
    var iObj int32;
    Bac_NtkStartCopies( p ); // -1 = not visited; 0 = on the path; 1 = finished
    Vec_IntClear( &p.vArray );
    // inlining of macro Bac_NtkForEachBoxUser
    for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsBoxUser(p, iObj) {} else
        if !(Bac_NtkDfsUserBoxes_rec( p, iObj, &p.vArray ) != 0) {
            fmt.Printf("Cyclic dependency of user boxes is detected.\n");
            return 0;
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_NtkCollapse_rec(pNew *Bac_Ntk_t, p *Bac_Ntk_t, vSigs *Vec_Int_t) {
    var i, iObj, iObjNew, iTerm int32;
    Bac_NtkStartCopies( p );
    // set PI copies
    assert(Vec_IntSize(vSigs) == Bac_NtkPiNum(p), "Vec_IntSize(vSigs) == Bac_NtkPiNum(p)");
    // inlining of macro Bac_NtkForEachPi
    for  i = 0; (i < Bac_NtkPiNum(p))  && (((iObj) = Bac_NtkPi(p, i)), 1); i++  { 
        Bac_ObjSetCopy( p, iObj, Vec_IntEntry(vSigs, i) );
    }
    // duplicate internal objects and create buffers for hierarchy instances
    // inlining of macro Bac_NtkForEachBox
    for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsBox(p, iObj) {} else
        if Bac_ObjIsBoxPrim( p, iObj ) {
            Bac_BoxDup( pNew, p, iObj );
        } else {
            // inlining of macro Bac_BoxForEachBo
            for  iTerm, i = iObj + 1, 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm, i = iTerm+1, i+1  {
                iObjNew = Bac_ObjAlloc( pNew, BAC_OBJ_BI,  -1 );
                iObjNew = Bac_ObjAlloc( pNew, BAC_BOX_BUF, -1 ); // buffer
                iObjNew = Bac_ObjAlloc( pNew, BAC_OBJ_BO,  -1 );
                Bac_ObjSetCopy( p, iTerm, iObjNew );
            }
        }
    }
    // duplicate user modules and connect objects
    // inlining of macro Bac_NtkForEachBox
    for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsBox(p, iObj) {} else
        if Bac_ObjIsBoxPrim( p, iObj ) {
            // inlining of macro Bac_BoxForEachBi
            for  iTerm, i = iObj - 1, 0; iTerm >= 0 && Bac_ObjIsBi(p, iTerm); iTerm, i = iTerm-1, i+1  {
                Bac_ObjSetFanin( pNew, Bac_ObjCopy(p, iTerm), Bac_ObjCopy(p, Bac_ObjFanin(p, iTerm)) );
            }
        
        } else {
            Vec_IntClear( vSigs );
            // inlining of macro Bac_BoxForEachBi
            for  iTerm, i = iObj - 1, 0; iTerm >= 0 && Bac_ObjIsBi(p, iTerm); iTerm, i = iTerm-1, i+1  {
                Vec_IntPush( vSigs, Bac_ObjCopy(p, Bac_ObjFanin(p, iTerm)) );
            }
            Bac_NtkCollapse_rec( pNew, Bac_BoxNtk(p, iObj), vSigs );
            assert(Vec_IntSize(vSigs) == Bac_BoxBoNum(p, iObj), "Vec_IntSize(vSigs) == Bac_BoxBoNum(p, iObj)");
            // inlining of macro Bac_BoxForEachBo
            for  iTerm, i = iObj + 1, 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm, i = iTerm+1, i+1  {
                Bac_ObjSetFanin( pNew, Bac_ObjCopy(p, iTerm)-2, Vec_IntEntry(vSigs, i) );
            }
        }
    }
    // collect POs
    Vec_IntClear( vSigs );
    // inlining of macro Bac_NtkForEachPo
    for  i = 0; (i < Bac_NtkPoNum(p))  && (((iObj) = Bac_NtkPo(p, i)), 1); i++  { 
        Vec_IntPush( vSigs, Bac_ObjCopy(p, Bac_ObjFanin(p, iObj)) );
    }
}
func Bac_ManCollapse(p *Bac_Man_t) *Bac_Man_t {
    var i, iObj int32;
    var vSigs *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var pNew *Bac_Man_t  = Bac_ManStart( p, 1 );
    var pRoot *Bac_Ntk_t  = Bac_ManRoot( p );
    var pRootNew *Bac_Ntk_t  = Bac_ManRoot( pNew );
    Bac_NtkAlloc( pRootNew, Bac_NtkNameId(pRoot), Bac_NtkPiNum(pRoot), Bac_NtkPoNum(pRoot), Bac_ManClpObjNum(p) );
    if Vec_IntSize(&pRoot.vInfo) {
        Vec_IntAppend( &pRootNew.vInfo, &pRoot.vInfo ); }
    // inlining of macro Bac_NtkForEachPi
    for  i = 0; (i < Bac_NtkPiNum(pRoot))  && (((iObj) = Bac_NtkPi(pRoot, i)), 1); i++  { 
        Vec_IntPush( vSigs, Bac_ObjAlloc(pRootNew, BAC_OBJ_PI, -1) );
    }
    Bac_NtkCollapse_rec( pRootNew, pRoot, vSigs );
    assert(Vec_IntSize(vSigs) == Bac_NtkPoNum(pRoot), "Vec_IntSize(vSigs) == Bac_NtkPoNum(pRoot)");
    // inlining of macro Bac_NtkForEachPo
    for  i = 0; (i < Bac_NtkPoNum(pRoot))  && (((iObj) = Bac_NtkPo(pRoot, i)), 1); i++  { 
        Bac_ObjAlloc( pRootNew, BAC_OBJ_PO, Vec_IntEntry(vSigs, i) );
    }
    assert(Bac_NtkObjNum(pRootNew) == Bac_NtkObjNumAlloc(pRootNew), "Bac_NtkObjNum(pRootNew) == Bac_NtkObjNumAlloc(pRootNew)");
    Vec_IntFree( vSigs );
    // transfer PI/PO names
    if Bac_NtkHasNames(pRoot) {
        Bac_NtkStartNames( pRootNew );
        // inlining of macro Bac_NtkForEachPi
        for  i = 0; (i < Bac_NtkPiNum(pRoot))  && (((iObj) = Bac_NtkPi(pRoot, i)), 1); i++  { 
            Bac_ObjSetName( pRootNew, Bac_NtkPi(pRootNew, i), Bac_ObjName(pRoot, iObj) );
        }
        // inlining of macro Bac_NtkForEachPoDriver
        for  i = 0; (i < Bac_NtkPoNum(pRoot))  && (((iObj) = Bac_ObjFanin(pRoot, Bac_NtkPo(pRoot, i))), 1); i++  { 
            if !Bac_ObjIsPi(pRoot, iObj) {
                Bac_ObjSetName( pRootNew, Bac_ObjCopy(pRoot, iObj), Bac_ObjName(pRoot, iObj) ); }
        }
    }
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
