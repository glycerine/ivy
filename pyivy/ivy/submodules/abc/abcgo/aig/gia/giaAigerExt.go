package gia

/**CFile****************************************************************

  FileName    [giaAigerExt.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Custom AIGER extensions.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaAigerExt.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Read/write equivalence classes information.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
Gia_Rpr_t * Gia_AigerReadEquivClasses( unsigned char ** ppPos, int nSize )
{
    var pReprs *Gia_Rpr_t;
    unsigned char * pStop;
    var i, Item, fProved, iRepr, iNode int32;
    pStop = *ppPos;
    pStop += Gia_AigerReadInt( *ppPos ); *ppPos += 4;
    // inlining of macro ABC_CALLOC
    pReprs = ((Gia_Rpr_t *) calloc((size_t)(nSize), sizeof(Gia_Rpr_t)));
    for  i = 0; i < nSize; i++  {
        pReprs[i].iRepr = GIA_VOID;
    }
    iRepr = iNode = 0;
    for  *ppPos < pStop  {
        Item = Gia_AigerReadUnsigned( ppPos );
        if Item & 1 {
            iRepr += (Item >> 1);
            iNode = iRepr;
            continue;
        }
        Item >>= 1;
        fProved = (Item & 1);
        Item >>= 1;
        iNode += Item;
        pReprs[iNode].fProved = fProved;
        pReprs[iNode].iRepr = iRepr;
        assert(iRepr < iNode, "iRepr < iNode");
    }
    return pReprs;
}
func Gia_WriteEquivClassesInt(p *Gia_Man_t, pEquivSize *int32) *unsigned char {
    unsigned char * pBuffer;
    var iRepr, iNode, iPrevRepr, iPrevNode, iLit, nItems, iPos int32;
    assert(p.pReprs && p.pNexts, "p.pReprs && p.pNexts");
    // count the number of entries to be written
    nItems = 0;
    for  iRepr = 1; iRepr < Gia_ManObjNum(p); iRepr++  {
        nItems += Gia_ObjIsConst( p, iRepr );
        if !Gia_ObjIsHead(p, iRepr) {
            continue; }
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, iRepr) && iRepr, "Gia_ObjIsHead(p, iRepr) && iRepr"), iNode = iRepr; iNode > 0; iNode = Gia_ObjNext(p, iNode)  {
            nItems++;
        }
    }
    // inlining of macro ABC_ALLOC
    pBuffer = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(sizeof(int) * (nItems + 10))));
    // write constant class
    iPos = Gia_AigerWriteUnsignedBuffer( pBuffer, 4, Abc_Var2Lit(0, 1) );
    iPrevNode = 0;
    for  iNode = 1; iNode < Gia_ManObjNum(p); iNode++  {
        if Gia_ObjIsConst(p, iNode) {
            iLit = Abc_Var2Lit( iNode - iPrevNode, Gia_ObjProved(p, iNode) );
            iPrevNode = iNode;
            iPos = Gia_AigerWriteUnsignedBuffer( pBuffer, iPos, Abc_Var2Lit(iLit, 0) );
        }
    }
    // write non-constant classes
    iPrevRepr = 0;
    // inlining of macro Gia_ManForEachClass
    for  iRepr = 1; iRepr < Gia_ManObjNum(p); iRepr++  { if !Gia_ObjIsHead(p, iRepr) {} else {
        iPos = Gia_AigerWriteUnsignedBuffer( pBuffer, iPos, Abc_Var2Lit(iRepr - iPrevRepr, 1) );
        iPrevRepr = iPrevNode = iRepr;
        // inlining of macro Gia_ClassForEachObj1
        for  assert(Gia_ObjIsHead(p, iRepr), "Gia_ObjIsHead(p, iRepr)"), iNode = Gia_ObjNext(p, iRepr); iNode > 0; iNode = Gia_ObjNext(p, iNode)  {
            iLit = Abc_Var2Lit( iNode - iPrevNode, Gia_ObjProved(p, iNode) );
            iPrevNode = iNode;
            iPos = Gia_AigerWriteUnsignedBuffer( pBuffer, iPos, Abc_Var2Lit(iLit, 0) );
        }
    }
    }
    Gia_AigerWriteInt( pBuffer, iPos );
    *pEquivSize = iPos;
    return pBuffer;
}
func Gia_WriteEquivClasses(p *Gia_Man_t) *Vec_Str_t {
    var nEquivSize int32;
    unsigned char * pBuffer = Gia_WriteEquivClassesInt( p, &nEquivSize );
    return Vec_StrAllocArray( (char *)pBuffer, nEquivSize );
}

/**Function*************************************************************

  Synopsis    [Read/write mapping information.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
static inline unsigned Gia_AigerReadDiffValue( unsigned char ** ppPos, int iPrev )
{
    var Item int32  = Gia_AigerReadUnsigned( ppPos );
    if Item & 1 {
        return iPrev + (Item >> 1); }
    return iPrev - (Item >> 1);
}
int * Gia_AigerReadMapping( unsigned char ** ppPos, int nSize )
{
    var pMapping *int32;
    unsigned char * pStop;
    var k int32;
    var j int32;
    var nFanins int32;
    var nAlloc int32;
    var iNode int32;
    var iOffset int32 = nSize;
    pStop = *ppPos;
    pStop += Gia_AigerReadInt( *ppPos ); *ppPos += 4;
    nAlloc = nSize + pStop - *ppPos;
    // inlining of macro ABC_CALLOC
    pMapping = ((int *) calloc((size_t)(nAlloc), sizeof(int)));
    for  *ppPos < pStop  {
        k = iOffset;
        pMapping[k++] = nFanins = Gia_AigerReadUnsigned( ppPos );
        for  j = 0; j <= nFanins; j++  {
            pMapping[k++] = iNode = Gia_AigerReadDiffValue( ppPos, iNode );
        }
        pMapping[iNode] = iOffset;
        iOffset = k;
    }
    assert(iOffset <= nAlloc, "iOffset <= nAlloc");
    return pMapping;
}
static inline int Gia_AigerWriteDiffValue( unsigned char * pPos, int iPos, int iPrev, int iThis )
{
    if iPrev < iThis {
        return Gia_AigerWriteUnsignedBuffer( pPos, iPos, Abc_Var2Lit(iThis - iPrev, 1) ); }
    return Gia_AigerWriteUnsignedBuffer( pPos, iPos, Abc_Var2Lit(iPrev - iThis, 0) );
}
func Gia_AigerWriteMappingInt(p *Gia_Man_t, pMapSize *int32) *unsigned char {
    unsigned char * pBuffer;
    var i int32;
    var k int32;
    var iPrev int32;
    var iFan int32;
    var nItems int32;
    var iPos int32 = 4;
    assert(Gia_ManHasMapping(p), "Gia_ManHasMapping(p)");
    // count the number of entries to be written
    nItems = 0;
    // inlining of macro Gia_ManForEachLut
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsLut(p, i) {} else {
        nItems += 2 + Gia_ObjLutSize( p, i ); }
    }
    // inlining of macro ABC_ALLOC
    pBuffer = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(sizeof(int) * (nItems + 1))));
    // write non-constant classes
    iPrev = 0;
    // inlining of macro Gia_ManForEachLut
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsLut(p, i) {} else {
//printf( "\nSize = %d ", Gia_ObjLutSize(p, i) );
        iPos = Gia_AigerWriteUnsignedBuffer( pBuffer, iPos, Gia_ObjLutSize(p, i) );
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,i) && ((iFan = Gia_ObjLutFanins(p,i)[k]),1); k++  {
//printf( "Fan = %d ", iFan );
            iPos = Gia_AigerWriteDiffValue( pBuffer, iPos, iPrev, iFan );
            iPrev = iFan;
        }
        iPos = Gia_AigerWriteDiffValue( pBuffer, iPos, iPrev, i );
        iPrev = i;
//printf( "Node = %d ", i );
    }
    }
//printf( "\n" );
    Gia_AigerWriteInt( pBuffer, iPos );
    *pMapSize = iPos;
    return pBuffer;
}
func Gia_AigerWriteMapping(p *Gia_Man_t) *Vec_Str_t {
    var nMapSize int32;
    unsigned char * pBuffer = Gia_AigerWriteMappingInt( p, &nMapSize );
    return Vec_StrAllocArray( (char *)pBuffer, nMapSize );
}

/**Function*************************************************************

  Synopsis    [Read/write mapping information.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
int * Gia_AigerReadMappingSimple( unsigned char ** ppPos, int nSize )
{
    // inlining of macro ABC_ALLOC
    var pMapping *int32  = ((int *) malloc(sizeof(int) * (size_t)((size_t)nSize/4)));
    memcpy( pMapping, *ppPos, nSize );
    assert(nSize % 4 == 0, "nSize % 4 == 0");
    return pMapping;
}
func Gia_AigerWriteMappingSimple(p *Gia_Man_t) *Vec_Str_t {
    // inlining of macro ABC_ALLOC
    unsigned char * pBuffer = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(4*Vec_IntSize(p.vMapping))));
    memcpy( pBuffer, Vec_IntArray(p.vMapping), (size_t)4*Vec_IntSize(p.vMapping) );
    assert(Vec_IntSize(p.vMapping) >= Gia_ManObjNum(p), "Vec_IntSize(p.vMapping) >= Gia_ManObjNum(p)");
    return Vec_StrAllocArray( (char *)pBuffer, 4*Vec_IntSize(p.vMapping) );
}

/**Function*************************************************************

  Synopsis    [Read/write mapping information.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
Vec_Int_t * Gia_AigerReadMappingDoc( unsigned char ** ppPos, int nObjs )
{
    var pMapping *int32;
    var nLuts int32;
    var LutSize int32;
    var iRoot int32;
    var nFanins int32;
    var i int32;
    var k int32;
    var nOffset int32;
    nLuts    = Gia_AigerReadInt( *ppPos ); *ppPos += 4;
    LutSize  = Gia_AigerReadInt( *ppPos ); *ppPos += 4;
    // inlining of macro ABC_CALLOC
    pMapping = ((int *) calloc((size_t)(nObjs + (LutSize + 2) * nLuts), sizeof(int)));
    nOffset = nObjs;
    for  i = 0; i < nLuts; i++  {
        iRoot   = Gia_AigerReadInt( *ppPos ); *ppPos += 4;
        nFanins = Gia_AigerReadInt( *ppPos ); *ppPos += 4;
        pMapping[iRoot] = nOffset;
        // write one
        pMapping[ nOffset++ ] = nFanins; 
        for  k = 0; k < nFanins; k++  {
            pMapping[ nOffset++ ] = Gia_AigerReadInt( *ppPos ); *ppPos += 4;
        }
        pMapping[ nOffset++ ] = iRoot; 
    }
    return Vec_IntAllocArray( pMapping, nOffset );
}
func Gia_AigerWriteMappingDoc(p *Gia_Man_t) *Vec_Str_t {
    unsigned char * pBuffer;
    var i int32;
    var k int32;
    var iFan int32;
    var nLuts int32;
    var LutSize int32;
    var nSize int32 = 2;
    var nSize2 int32;
    // inlining of macro Gia_ManForEachLut
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsLut(p, i) {} else {
        nLuts++;
        nSize += Gia_ObjLutSize(p, i) + 2;
        LutSize = Abc_MaxInt( LutSize, Gia_ObjLutSize(p, i) );
    }
    }
    // inlining of macro ABC_ALLOC
    pBuffer = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(4 * nSize)));
    Gia_AigerWriteInt( pBuffer + 4 * nSize2++, nLuts );  
    Gia_AigerWriteInt( pBuffer + 4 * nSize2++, LutSize );
    // inlining of macro Gia_ManForEachLut
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsLut(p, i) {} else {
        Gia_AigerWriteInt( pBuffer + 4 * nSize2++, i );
        Gia_AigerWriteInt( pBuffer + 4 * nSize2++, Gia_ObjLutSize(p, i) );
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,i) && ((iFan = Gia_ObjLutFanins(p,i)[k]),1); k++  {
            Gia_AigerWriteInt( pBuffer + 4 * nSize2++, iFan );
        }
    }
    }
    assert(nSize2 == nSize, "nSize2 == nSize");
    return Vec_StrAllocArray( (char *)pBuffer, 4*nSize );
}

/**Function*************************************************************

  Synopsis    [Read/write packing information.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
Vec_Int_t * Gia_AigerReadPacking( unsigned char ** ppPos, int nSize )
{
    var vPacking *Vec_Int_t  = Vec_IntAlloc( nSize/4 );
    var i int32;
    assert(nSize % 4 == 0, "nSize % 4 == 0");
    for  i = 0; i < nSize/4; i++, *ppPos += 4  {
        Vec_IntPush( vPacking, Gia_AigerReadInt( *ppPos ) );
    }
    return vPacking;
}
func Gia_WritePacking(vPacking *Vec_Int_t) *Vec_Str_t {
    // inlining of macro ABC_ALLOC
    unsigned char * pBuffer = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(4*Vec_IntSize(vPacking))));
    var i, Entry, nSize int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPacking)) && (((Entry) = Vec_IntEntry(vPacking, i)), 1); i++  {
        Gia_AigerWriteInt( pBuffer + 4 * nSize++, Entry );
    }
    return Vec_StrAllocArray( (char *)pBuffer, 4*Vec_IntSize(vPacking) );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

