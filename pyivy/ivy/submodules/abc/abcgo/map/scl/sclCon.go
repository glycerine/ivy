package scl

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [sclCon.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Constraint manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclCon.h,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/
 



//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Scl_Con_t struct {
    pFileName          string;   // constraint file name
    pModelName          string;  // current design name
    pInCellDef          string;  // default input driving gate
    tInArrDef            int32;   // default input arrival time
    tInSlewDef            int32;  // default input slew
    tInLoadDef            int32;  // default input load
    tOutReqDef            int32;  // default output required time
    tOutLoadDef            int32; // default output load
    vInCells      Vec_Ptr_t;    // input driving gate names
    vInCellsPtr      Vec_Ptr_t; // input driving gates
    vInArrs      Vec_Int_t;     // input arrival times
    vInSlews      Vec_Int_t;    // input slews
    vInLoads      Vec_Int_t;    // input loads
    vOutReqs      Vec_Int_t;    // output required times
    vOutLoads      Vec_Int_t;   // output loads
    pNamI     *Abc_Nam_t;       // input names
    pNamO     *Abc_Nam_t;       // output names
}

const SCL_INPUT_CELL =   "input_cell"
const SCL_INPUT_ARR =    "input_arrival"
const SCL_INPUT_SLEW =   "input_slew"
const SCL_INPUT_LOAD =   "input_load"
const SCL_OUTPUT_REQ =   "output_required"
const SCL_OUTPUT_LOAD =  "output_load"

// Original macro:
// #define SCL_DIRECTIVE(ITEM)     "."#ITEM

// Original macro:
// #define SCL_DEF_DIRECTIVE(ITEM) ".default_"#ITEM


const SCL_NUM =          1000
#define SCL_INFINITY    (0x3FFFFFFF)

func Scl_Flt2Int(w float32) int32 { return SCL_NUM*w;        }
func Scl_Int2Flt(i int32) float32 { return float32(i)/SCL_NUM; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Manager construction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_ConAlloc(pFileName string, pNamI *Abc_Nam_t, pNamO *Abc_Nam_t) *Scl_Con_t {
    var p *Scl_Con_t   = new(Scl_Con_t);
    p.pFileName   = Abc_UtilStrsav( pFileName );
    p.pNamI       = pNamI;
    p.pNamO       = pNamO;
    p.pModelName  = nil;
    p.pInCellDef  = nil;
    p.tInArrDef   = SCL_INFINITY;
    p.tInSlewDef  = SCL_INFINITY;
    p.tInLoadDef  = SCL_INFINITY;
    p.tOutReqDef  = SCL_INFINITY;
    p.tOutLoadDef = SCL_INFINITY;
    Vec_PtrFill( &p.vInCells,  Abc_NamObjNumMax(pNamI)-1, nil );
    Vec_IntFill( &p.vInArrs,   Abc_NamObjNumMax(pNamI)-1, SCL_INFINITY );
    Vec_IntFill( &p.vInSlews,  Abc_NamObjNumMax(pNamI)-1, SCL_INFINITY );
    Vec_IntFill( &p.vInLoads,  Abc_NamObjNumMax(pNamI)-1, SCL_INFINITY );
    Vec_IntFill( &p.vOutReqs,  Abc_NamObjNumMax(pNamO)-1, SCL_INFINITY );
    Vec_IntFill( &p.vOutLoads, Abc_NamObjNumMax(pNamO)-1, SCL_INFINITY );
    return p;
}
func Scl_ConFree(p *Scl_Con_t) {
    Vec_PtrErase( &p.vInCellsPtr );
    Vec_PtrFreeData( &p.vInCells );
    Vec_PtrErase( &p.vInCells );
    Vec_IntErase( &p.vInArrs );
    Vec_IntErase( &p.vInSlews );
    Vec_IntErase( &p.vInLoads );
    Vec_IntErase( &p.vOutReqs );
    Vec_IntErase( &p.vOutLoads );
    Abc_NamDeref( p.pNamI );
    Abc_NamDeref( p.pNamO );
    ABC_FREE( p.pInCellDef );
    ABC_FREE( p.pModelName );
    ABC_FREE( p.pFileName );
    ABC_FREE( p );
}


/**Function*************************************************************

  Synopsis    [Manager serialization.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_ConParse(p *Scl_Con_t, pNamI *Abc_Nam_t, pNamO *Abc_Nam_t) int32 {
    var Buffer [1000]byte;
    var pToken, pToken2, pToken3, pName string; 
    var i int32;
    var Num int32 = -1;
    var nLines int32; var Value int32;
    var pFile *os.File  = os.Open(p.pFileName);
    for  fgets( Buffer, 1000, pFile )  {
        nLines++;
        pToken = strtok( Buffer, " \t\r\n" );
        if pToken == nil {
            continue; }
        pToken2 = strtok( nil, " \t\r\n" );
        if pToken2 == nil {
            fmt.Printf("Line %d: Skipping directive \"%s\" without argument.\n", nLines, pToken);
            continue;
        }
        pToken3 = strtok( nil, " \t\r\n" );
        if !(int32(strings.Compare(pToken, ".model")) != 0) {                                p.pModelName  = Abc_UtilStrsav(pToken2);
        } else if !(int32(strings.Compare(pToken, SCL_DEF_DIRECTIVE(SCL_INPUT_CELL))) != 0) {  p.pInCellDef  = Abc_UtilStrsav(pToken2);
        } else if !(int32(strings.Compare(pToken, SCL_DEF_DIRECTIVE(SCL_INPUT_ARR))) != 0) {  p.tInArrDef   = Scl_Flt2Int(atof(pToken2));
        } else if !(int32(strings.Compare(pToken, SCL_DEF_DIRECTIVE(SCL_INPUT_SLEW))) != 0) {  p.tInSlewDef  = Scl_Flt2Int(atof(pToken2));
        } else if !(int32(strings.Compare(pToken, SCL_DEF_DIRECTIVE(SCL_INPUT_LOAD))) != 0) {  p.tInLoadDef  = Scl_Flt2Int(atof(pToken2));
        } else if !(int32(strings.Compare(pToken, SCL_DEF_DIRECTIVE(SCL_OUTPUT_REQ))) != 0) {  p.tOutReqDef  = Scl_Flt2Int(atof(pToken2));
        } else if !(int32(strings.Compare(pToken, SCL_DEF_DIRECTIVE(SCL_OUTPUT_LOAD))) != 0) {  p.tOutLoadDef = Scl_Flt2Int(atof(pToken2));
        } else if pToken3 == nil { fmt.Printf("Directive %s should be followed by two arguments.\n", pToken); continue; 
        } else if !(int32(strings.Compare(pToken, SCL_DIRECTIVE(SCL_INPUT_CELL))) != 0) {     if (Num = Abc_NamStrFind(pNamI, pToken2)) > 0 {  Vec_PtrWriteEntry( &p.vInCells,  Num-1, Abc_UtilStrsav(pToken3) ); }    else { fmt.Printf("Line %d: Cannot find input \"%s\".\n", nLines, pToken2); }
        } else if !(int32(strings.Compare(pToken, SCL_DIRECTIVE(SCL_INPUT_ARR))) != 0) {     if (Num = Abc_NamStrFind(pNamI, pToken2)) > 0 {  Vec_IntWriteEntry( &p.vInArrs,   Num-1, Scl_Flt2Int(atof(pToken3)) ); } else { fmt.Printf("Line %d: Cannot find input \"%s\".\n", nLines, pToken2); }
        } else if !(int32(strings.Compare(pToken, SCL_DIRECTIVE(SCL_INPUT_SLEW))) != 0) {     if (Num = Abc_NamStrFind(pNamI, pToken2)) > 0 {  Vec_IntWriteEntry( &p.vInSlews,  Num-1, Scl_Flt2Int(atof(pToken3)) ); } else { fmt.Printf("Line %d: Cannot find input \"%s\".\n", nLines, pToken2); }
        } else if !(int32(strings.Compare(pToken, SCL_DIRECTIVE(SCL_INPUT_LOAD))) != 0) {     if (Num = Abc_NamStrFind(pNamI, pToken2)) > 0 {  Vec_IntWriteEntry( &p.vInLoads,  Num-1, Scl_Flt2Int(atof(pToken3)) ); } else { fmt.Printf("Line %d: Cannot find input \"%s\".\n", nLines, pToken2); }
        } else if !(int32(strings.Compare(pToken, SCL_DIRECTIVE(SCL_OUTPUT_REQ))) != 0) {     if (Num = Abc_NamStrFind(pNamO, pToken2)) > 0 {  Vec_IntWriteEntry( &p.vOutReqs,  Num-1, Scl_Flt2Int(atof(pToken3)) ); } else { fmt.Printf("Line %d: Cannot find output \"%s\".\n", nLines, pToken2); }
        } else if !(int32(strings.Compare(pToken, SCL_DIRECTIVE(SCL_OUTPUT_LOAD))) != 0) {     if (Num = Abc_NamStrFind(pNamO, pToken2)) > 0 {  Vec_IntWriteEntry( &p.vOutLoads, Num-1, Scl_Flt2Int(atof(pToken3)) ); } else { fmt.Printf("Line %d: Cannot find output \"%s\".\n", nLines, pToken2); }
        } else { fmt.Printf("Line %d: Skipping unrecognized directive \"%s\".\n", nLines, pToken); }
    }
    // set missing defaults
    if p.pInCellDef  == nil {          p.pInCellDef  = nil; } // consider using buffer from the current library
    if p.tInArrDef   == SCL_INFINITY {  p.tInArrDef   = 0; }
    if p.tInSlewDef  == SCL_INFINITY {  p.tInSlewDef  = 0; }
    if p.tInLoadDef  == SCL_INFINITY {  p.tInLoadDef  = 0; }
    if p.tOutReqDef  == SCL_INFINITY {  p.tOutReqDef  = 0; }
    if p.tOutLoadDef == SCL_INFINITY {  p.tOutLoadDef = 0; }
    // set individual defaults
    if p.pInCellDef {
    Vec_PtrForEachEntry(char *, &p.vInCells, pName, i) if ( pName == nil )          Vec_PtrWriteEntry( &p.vInCells,  i, Abc_UtilStrsav(p.pInCellDef) ); }
    Vec_IntForEachEntry( &p.vInArrs, Value, i )        if ( Value == SCL_INFINITY )  Vec_IntWriteEntry( &p.vInArrs,   i, p.tInArrDef );
    Vec_IntForEachEntry( &p.vInSlews, Value, i )       if ( Value == SCL_INFINITY )  Vec_IntWriteEntry( &p.vInSlews,  i, p.tInSlewDef );
    Vec_IntForEachEntry( &p.vInLoads, Value, i )       if ( Value == SCL_INFINITY )  Vec_IntWriteEntry( &p.vInLoads,  i, p.tInLoadDef );
    Vec_IntForEachEntry( &p.vOutReqs, Value, i )       if ( Value == SCL_INFINITY )  Vec_IntWriteEntry( &p.vOutReqs,  i, p.tOutReqDef );
    Vec_IntForEachEntry( &p.vOutLoads, Value, i )      if ( Value == SCL_INFINITY )  Vec_IntWriteEntry( &p.vOutLoads, i, p.tOutLoadDef );

    _ = (pFile).Close();
    return 1;
}
func Scl_ConRead(pFileName string, pNamI *Abc_Nam_t, pNamO *Abc_Nam_t) *Scl_Con_t {
    var p *Scl_Con_t  = Scl_ConAlloc( pFileName, pNamI, pNamO );
    if Scl_ConParse(p, pNamI, pNamO) != 0 {
        return p; }
    Scl_ConFree( p );
    return nil;
}
func Scl_ConWrite(p *Scl_Con_t, pFileName string) {
    var pName string; var Value int32; var i int32;
    var pFile *os.File  = pFileName ? os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666) : os.Stdout;
    if pFile == nil {
        fmt.Printf("Cannot open output file \"%s\".\n", pFileName);
        return;
    }
    fmt.Fprintf(pFile, ".model %s\n", p.pModelName);

    if p.pInCellDef {        fmt.Fprintf(pFile, ".default_%s %s\n", SCL_INPUT_CELL, p.pInCellDef); }
    if p.tInArrDef   != 0 {  fmt.Fprintf(pFile, ".default_%s %.2f\n", SCL_INPUT_ARR, Scl_Int2Flt(p.tInArrDef)); }
    if p.tInSlewDef  != 0 {  fmt.Fprintf(pFile, ".default_%s %.2f\n", SCL_INPUT_SLEW, Scl_Int2Flt(p.tInSlewDef)); }
    if p.tInLoadDef  != 0 {  fmt.Fprintf(pFile, ".default_%s %.2f\n", SCL_INPUT_LOAD, Scl_Int2Flt(p.tInLoadDef)); }
    if p.tOutReqDef  != 0 {  fmt.Fprintf(pFile, ".default_%s %.2f\n", SCL_OUTPUT_REQ, Scl_Int2Flt(p.tOutReqDef)); }
    if p.tOutLoadDef != 0 {  fmt.Fprintf(pFile, ".default_%s %.2f\n", SCL_OUTPUT_LOAD, Scl_Int2Flt(p.tOutLoadDef)); }

    Vec_PtrForEachEntry(char *, &p.vInCells, pName, i) if ( pName && (!p.pInCellDef || int32(strings.Compare(pName, p.pInCellDef))) )  fmt.Fprintf(pFile, ".%s %s %s\n", SCL_INPUT_CELL, Abc_NamStr(p.pNamI, i+1), pName);
    Vec_IntForEachEntry( &p.vInArrs, Value, i )        if ( Value != p.tInArrDef )   fmt.Fprintf(pFile, ".%s %s %.2f\n", SCL_INPUT_ARR, Abc_NamStr(p.pNamI, i+1), Scl_Int2Flt(Value));
    Vec_IntForEachEntry( &p.vInSlews, Value, i )       if ( Value != p.tInSlewDef )  fmt.Fprintf(pFile, ".%s %s %.2f\n", SCL_INPUT_SLEW, Abc_NamStr(p.pNamI, i+1), Scl_Int2Flt(Value));
    Vec_IntForEachEntry( &p.vInLoads, Value, i )       if ( Value != p.tInLoadDef )  fmt.Fprintf(pFile, ".%s %s %.2f\n", SCL_INPUT_LOAD, Abc_NamStr(p.pNamI, i+1), Scl_Int2Flt(Value));
    Vec_IntForEachEntry( &p.vOutReqs, Value, i )       if ( Value != p.tOutReqDef )  fmt.Fprintf(pFile, ".%s %s %.2f\n", SCL_OUTPUT_REQ, Abc_NamStr(p.pNamO, i+1), Scl_Int2Flt(Value));
    Vec_IntForEachEntry( &p.vOutLoads, Value, i )      if ( Value != p.tOutLoadDef ) fmt.Fprintf(pFile, ".%s %s %.2f\n", SCL_OUTPUT_LOAD, Abc_NamStr(p.pNamO, i+1), Scl_Int2Flt(Value));

    if pFile != os.Stdout {
        _ = (pFile).Close(); }
}

/**Function*************************************************************

  Synopsis    [Internal APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_ConHasInCells_(p *Scl_Con_t) int32 { return Vec_PtrCountZero(&p.vInCells)  != Vec_PtrSize(&p.vInCells);  }
func Scl_ConHasInArrs_(p *Scl_Con_t) int32 { return Vec_IntCountZero(&p.vInArrs)   != Vec_IntSize(&p.vInArrs);   }
func Scl_ConHasInSlews_(p *Scl_Con_t) int32 { return Vec_IntCountZero(&p.vInSlews)  != Vec_IntSize(&p.vInSlews);  }
func Scl_ConHasInLoads_(p *Scl_Con_t) int32 { return Vec_IntCountZero(&p.vInLoads)  != Vec_IntSize(&p.vInLoads);  }
func Scl_ConHasOutReqs_(p *Scl_Con_t) int32 { return Vec_IntCountZero(&p.vOutReqs)  != Vec_IntSize(&p.vOutReqs);  }
func Scl_ConHasOutLoads_(p *Scl_Con_t) int32 { return Vec_IntCountZero(&p.vOutLoads) != Vec_IntSize(&p.vOutLoads); }

func Scl_ConGetInCell_(p *Scl_Con_t, i int32) string { return (char*)Vec_PtrEntry( &p.vInCells,  i ); }
func Scl_ConGetInArr_(p *Scl_Con_t, i int32) int32 { return Vec_IntEntry( &p.vInArrs,   i ); }
func Scl_ConGetInSlew_(p *Scl_Con_t, i int32) int32 { return Vec_IntEntry( &p.vInSlews,  i ); }
func Scl_ConGetInLoad_(p *Scl_Con_t, i int32) int32 { return Vec_IntEntry( &p.vInLoads,  i ); }
func Scl_ConGetOutReq_(p *Scl_Con_t, i int32) int32 { return Vec_IntEntry( &p.vOutReqs,  i ); }
func Scl_ConGetOutLoad_(p *Scl_Con_t, i int32) int32 { return Vec_IntEntry( &p.vOutLoads, i ); }

/**Function*************************************************************

  Synopsis    [External APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// extern Scl_Con_t *   Scl_ConReadMan();


func Scl_ConIsRunning() int32 { return Scl_ConReadMan() != nil;                  }

func Scl_ConHasInCells() int32 { return Scl_ConHasInCells_ ( Scl_ConReadMan() );   }
func Scl_ConHasInArrs() int32 { return Scl_ConHasInArrs_  ( Scl_ConReadMan() );   }
func Scl_ConHasInSlews() int32 { return Scl_ConHasInSlews_ ( Scl_ConReadMan() );   }
func Scl_ConHasInLoads() int32 { return Scl_ConHasInLoads_ ( Scl_ConReadMan() );   }
func Scl_ConHasOutReqs() int32 { return Scl_ConHasOutReqs_ ( Scl_ConReadMan() );   }
func Scl_ConHasOutLoads() int32 { return Scl_ConHasOutLoads_( Scl_ConReadMan() );   }

func Scl_ConGetInCell(i int32) string { return Scl_ConGetInCell_ ( Scl_ConReadMan(), i ); }
func Scl_ConGetInArr(i int32) int32 { return Scl_ConGetInArr_  ( Scl_ConReadMan(), i ); }
func Scl_ConGetInSlew(i int32) int32 { return Scl_ConGetInSlew_ ( Scl_ConReadMan(), i ); }
func Scl_ConGetInLoad(i int32) int32 { return Scl_ConGetInLoad_ ( Scl_ConReadMan(), i ); }
func Scl_ConGetOutReq(i int32) int32 { return Scl_ConGetOutReq_ ( Scl_ConReadMan(), i ); }
func Scl_ConGetOutLoad(i int32) int32 { return Scl_ConGetOutLoad_( Scl_ConReadMan(), i ); }

func Scl_ConGetInArrFloat(i int32) float32 { return Scl_Int2Flt( Scl_ConGetInArr(i) );         }
func Scl_ConGetInSlewFloat(i int32) float32 { return Scl_Int2Flt( Scl_ConGetInSlew(i) );        }
func Scl_ConGetInLoadFloat(i int32) float32 { return Scl_Int2Flt( Scl_ConGetInLoad(i) );        }
func Scl_ConGetOutReqFloat(i int32) float32 { return Scl_Int2Flt( Scl_ConGetOutReq(i) );        }
func Scl_ConGetOutLoadFloat(i int32) float32 { return Scl_Int2Flt( Scl_ConGetOutLoad(i) );       }


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

