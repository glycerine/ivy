package gia

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [giaUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaUtil.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "base/main/mainInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates a sequence or random numbers.]

  Description []
               
  SideEffects []

  SeeAlso     [http://www.codeproject.com/KB/recipes/SimpleRNG.aspx]

***********************************************************************/





/**Function*************************************************************

  Synopsis    [Creates random info for the primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Returns the time stamp.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns the composite name of the file.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Cleans the value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Cleans the value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets the phase of one object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns the number of COs whose value is 1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prepares copies for the model.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prepares copies for the model.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Assigns levels using CI level information.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Compute reverse levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Compute required levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Compute slacks measured using the number of AIG levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns references.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns references.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Computes the maximum frontier size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Collects PO Ids into one array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the root of MUX or EXOR/NEXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Recognizes what nodes are inputs of the EXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Recognizes what nodes are control and data inputs of a MUX.]

  Description [If the node is a MUX, returns the control variable C.
  Assigns nodes T and E to be the then and else variables of the MUX. 
  Node C is never complemented. Nodes T and E can be complemented.
  This function also recognizes EXOR/NEXOR gates as MUXes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Dereferences the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [References the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [References the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns the number of internal nodes in the MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Returns the number of internal nodes in the MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has dangling nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has dangling nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has dangling nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Verbose printing of the AIG node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Complements the constraint outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Testing the speedup due to grouping POs into batches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Marks nodes that appear as faninis of other nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Swaps PO number 0 with PO number i.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Save/load value from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Returns the array containing the first fanout of each object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Check if two logic cones have overlap.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    [Count PIs with fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Count POs driven by non-zero driver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupWithMuxPos(p *Gia_Man_t) *Gia_Man_t {
    var vPoints *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vQuads *Vec_Int_t   = Vec_IntAlloc( 1000 );
    var vHeads *Vec_Bit_t   = Vec_BitStart( Gia_ManObjNum(p) );
    var vDatas *Vec_Bit_t   = Vec_BitStart( Gia_ManObjNum(p) );
    var pObj, pCtrl, pData0, pData1 *Gia_Obj_t;
    var pNew *Gia_Man_t  = Gia_ManDup( p ); var i, iObj int32;
    assert(Gia_ManRegNum(pNew) == 0, "Gia_ManRegNum(pNew) == 0");
    Gia_ManForEachAnd( pNew, pObj, i )
    {
        if !(Gia_ObjIsMuxType(pObj) != 0) {
            continue; }
        pCtrl  = Gia_ObjRecognizeMux( pObj, &pData1, &pData0 );
        pCtrl  = Gia_Regular(pCtrl);
        pData0 = Gia_Regular(pData0);
        pData1 = Gia_Regular(pData1);
        Vec_IntPushTwo( vQuads, Gia_ObjId(pNew, pObj),   Gia_ObjId(pNew, pCtrl)  );
        Vec_IntPushTwo( vQuads, Gia_ObjId(pNew, pData0), Gia_ObjId(pNew, pData1) );
        Vec_BitWriteEntry( vHeads, Gia_ObjId(pNew, pObj),   1 );
        Vec_BitWriteEntry( vDatas, Gia_ObjId(pNew, pData0), 1 );
        Vec_BitWriteEntry( vDatas, Gia_ObjId(pNew, pData1), 1 );
    }
    Gia_ManForEachCo( pNew, pObj, i )
        Gia_ObjFanin0(pObj).fMark0 = 1;
    for  i = 0; i < Vec_IntSize(vQuads)/4; i++  {
        var iObj int32    = Vec_IntEntry( vQuads, 4*i+0 );
        var iCtrl int32   = Vec_IntEntry( vQuads, 4*i+1 );
        var iData0 int32  = Vec_IntEntry( vQuads, 4*i+2 );
        var iData1 int32  = Vec_IntEntry( vQuads, 4*i+3 );
        if (Vec_BitEntry(vHeads, iObj)   && Vec_BitEntry(vDatas, iObj))   || 
             (Vec_BitEntry(vHeads, iData0) && Vec_BitEntry(vDatas, iData0)) || 
             (Vec_BitEntry(vHeads, iData1) && Vec_BitEntry(vDatas, iData1)) {
            var pObj *Gia_Obj_t    = Gia_ManObj( p, iObj );
            var pCtrl *Gia_Obj_t   = Gia_ManObj( p, iCtrl );
            var pData0 *Gia_Obj_t  = Gia_ManObj( p, iData0 );
            var pData1 *Gia_Obj_t  = Gia_ManObj( p, iData1 );
            if Gia_ObjIsAnd(pObj)   && !pObj.fMark0 {  Vec_IntPush( vPoints, iObj   ); }
            if Gia_ObjIsAnd(pCtrl)  && !pCtrl.fMark0 {  Vec_IntPush( vPoints, iCtrl  ); }
            if Gia_ObjIsAnd(pData0) && !pData0.fMark0 {  Vec_IntPush( vPoints, iData0 ); }
            if Gia_ObjIsAnd(pData1) && !pData1.fMark0 {  Vec_IntPush( vPoints, iData1 ); }
        }
    }
    Gia_ManCleanMark0( pNew );
    Vec_IntUniqify( vPoints );
    Vec_IntForEachEntry( vPoints, iObj, i )
        Gia_ManAppendCo( pNew, Abc_Var2Lit(iObj, 0) );
    Vec_IntFree( vPoints );
    Vec_IntFree( vQuads );
    Vec_BitFree( vHeads );
    Vec_BitFree( vDatas );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Collect distance info.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ComputeTest() {
    var pStart string;
    var Line [1000]byte; var Total float32;
    var pFileName string  = "data.txt";
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil { fmt.Printf("Input file \"%s\" cannot be opened.\n", pFileName); return; }
    for  fgets( Line, 1000, pFile ) != nil  {
        if !strstr(Line, "xxx") {
            continue; }
        if !strstr(Line, "yyy") {
            continue; }
        //printf( "%s", Line );
        pStart = strstr(Line, "zzz");
        if pStart == nil {
            continue; }
        //printf( "%s", pStart + 4 );
        Total += -atof( pStart + 4 );
    }
    fmt.Printf("Total = %.2f\n", Total);
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

