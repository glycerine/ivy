package ivy

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ivyDsd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Disjoint-support decomposition.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyDsd.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// decomposition types
 type Ivy_DecType_t int32
const ( 
    IVY_DEC_PI Ivy_DecType_t = iota
             // 0: var
    IVY_DEC_CONST1
         // 1: CONST1
    IVY_DEC_BUF
            // 2: BUF
    IVY_DEC_AND
            // 3: AND
    IVY_DEC_EXOR
           // 4: EXOR
    IVY_DEC_MUX
            // 5: MUX
    IVY_DEC_MAJ
            // 6: MAJ
    IVY_DEC_PRIME           // 7: undecomposable

)

 

   
type Ivy_Dec_t struct {
    Type uint32; /* C bitfield: 4 bits; review packing and truncation */   // the node type (PI, CONST1, AND, EXOR, MUX, PRIME)
    fCompl uint32; /* C bitfield: 1 bits; review packing and truncation */   // shows if node is complemented (root node only)
    nFans uint32; /* C bitfield: 3 bits; review packing and truncation */   // the number of fanins
    Fan0 uint32; /* C bitfield: 4 bits; review packing and truncation */   // fanin 0
    Fan1 uint32; /* C bitfield: 4 bits; review packing and truncation */   // fanin 1 
    Fan2 uint32; /* C bitfield: 4 bits; review packing and truncation */   // fanin 2
    Fan3 uint32; /* C bitfield: 4 bits; review packing and truncation */   // fanin 3 
    Fan4 uint32; /* C bitfield: 4 bits; review packing and truncation */   // fanin 4 
    Fan5 uint32; /* C bitfield: 4 bits; review packing and truncation */   // fanin 5
}

func Ivy_DecToInt(m Ivy_Dec_t) int32 { union { var x Ivy_Dec_t; var y int32; }
 v; v.x = m; return v.y;  }
func Ivy_IntToDec(m int32) Ivy_Dec_t { union { var x Ivy_Dec_t; var y int32; }
 v; v.y = m; return v.x;  }
func Ivy_DecClear(pNode *Ivy_Dec_t) { *pNode = Ivy_IntToDec(0);                              }

//static inline int        Ivy_DecToInt( Ivy_Dec_t Node )     { return *((int *)&Node);       }
//static inline Ivy_Dec_t  Ivy_IntToDec( int Node )           { return *((Ivy_Dec_t *)&Node); }
//static inline void       Ivy_DecClear( Ivy_Dec_t * pNode )  { *((int *)pNode) = 0;          }


var s_Masks [6][2]uint32 = [6][2]uint32{
    { 0x55555555, 0xAAAAAAAA },
    { 0x33333333, 0xCCCCCCCC },
    { 0x0F0F0F0F, 0xF0F0F0F0 },
    { 0x00FF00FF, 0xFF00FF00 },
    { 0x0000FFFF, 0xFFFF0000 },
    { 0x00000000, 0xFFFFFFFF },
};

func Ivy_TruthWordCountOnes(uWord uint32) int32 {
    uWord = (uWord & 0x55555555) + ((uWord>>1) & 0x55555555);
    uWord = (uWord & 0x33333333) + ((uWord>>2) & 0x33333333);
    uWord = (uWord & 0x0F0F0F0F) + ((uWord>>4) & 0x0F0F0F0F);
    uWord = (uWord & 0x00FF00FF) + ((uWord>>8) & 0x00FF00FF);
    return  (uWord & 0x0000FFFF) + (uWord>>16);
}

func Ivy_TruthCofactorIsConst(uTruth uint32, Var int32, Cof int32, Const int32) int32 {
    if Const == 0 {
        return (uTruth & s_Masks[Var][Cof]) == 0;
    } else {
        return (uTruth & s_Masks[Var][Cof]) == s_Masks[Var][Cof]; }
}

func Ivy_TruthCofactorIsOne(uTruth uint32, Var int32) int32 {
    return (uTruth & s_Masks[Var][0]) == 0;
}

func Ivy_TruthCofactor(uTruth uint32, Var int32) uint32 {
    var uCofactor uint32  = uTruth & s_Masks[Var >> 1][(Var & 1) == 0];
    var Shift int32  = (1 << (Var >> 1));
    if Var & 1 {
        return uCofactor | (uCofactor << Shift); }
    return uCofactor | (uCofactor >> Shift);
}

func Ivy_TruthCofactor2(uTruth uint32, Var0 int32, Var1 int32) uint32 {
    return Ivy_TruthCofactor( Ivy_TruthCofactor(uTruth, Var0), Var1 );
}

// returns 1 if the truth table depends on this var (var is regular interger var)
func Ivy_TruthDepends(uTruth uint32, Var int32) int32 {
    return Ivy_TruthCofactor(uTruth, Var << 1) != Ivy_TruthCofactor(uTruth, (Var << 1) | 1);
}

func Ivy_DecSetVar(pNode *Ivy_Dec_t, iNum int32, Var uint32) {
    assert(iNum >= 0 && iNum <= 5, "iNum >= 0 && iNum <= 5");
    switch  iNum  {
        case 0: pNode.Fan0 = Var; break;
        case 1: pNode.Fan1 = Var; break;
        case 2: pNode.Fan2 = Var; break;
        case 3: pNode.Fan3 = Var; break;
        case 4: pNode.Fan4 = Var; break;
        case 5: pNode.Fan5 = Var; break;
    }
}

func Ivy_DecGetVar(pNode *Ivy_Dec_t, iNum int32) uint32 {
    assert(iNum >= 0 && iNum <= 5, "iNum >= 0 && iNum <= 5");
    switch  iNum  {
        case 0: return pNode.Fan0;
        case 1: return pNode.Fan1;
        case 2: return pNode.Fan2;
        case 3: return pNode.Fan3;
        case 4: return pNode.Fan4;
        case 5: return pNode.Fan5;
    }
    return ~0;
}

// static int   Ivy_TruthDecompose_rec( unsigned uTruth, Vec_Int_t * vTree );

// static int   Ivy_TruthRecognizeMuxMaj( unsigned uTruth, int * pSupp, int nSupp, Vec_Int_t * vTree );


//int nTruthDsd;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes DSD of truth table of 5 variables or less.]

  Description [Returns 1 if the function is a constant or is fully 
  DSD decomposable using AND/EXOR/MUX gates.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_TruthDsd(uTruth uint32, vTree *Vec_Int_t) int32 {
    var Node Ivy_Dec_t;
    var i, RetValue int32;
    // set the PI variables
    Vec_IntClear( vTree );
    for  i = 0; i < 5; i++  {
        Vec_IntPush( vTree, 0 );
    }
    // check if it is a constant
    if uTruth == 0 || ~uTruth == 0 {
        Ivy_DecClear( &Node );
        Node.Type = IVY_DEC_CONST1;
        Node.fCompl = (uTruth == 0);
        Vec_IntPush( vTree, Ivy_DecToInt(Node) );
        return 1;
    }
    // perform the decomposition
    RetValue = Ivy_TruthDecompose_rec( uTruth, vTree );
    if RetValue == -1 {
        return 0; }
    // get the topmost node
    if (RetValue >> 1) < 5 { // add buffer
        Ivy_DecClear( &Node );
        Node.Type = IVY_DEC_BUF;
        Node.fCompl = (RetValue & 1);
        Node.Fan0 = ((RetValue >> 1) << 1);
        Vec_IntPush( vTree, Ivy_DecToInt(Node) );
    
    } else if RetValue & 1 { // check if the topmost node has to be complemented
        Node = Ivy_IntToDec( Vec_IntPop(vTree) );
        assert(Node.fCompl == 0, "Node.fCompl == 0");
        Node.fCompl = (RetValue & 1);
        Vec_IntPush( vTree, Ivy_DecToInt(Node) );
    }
    if uTruth != Ivy_TruthDsdCompute(vTree) {
        fmt.Printf("Verification failed.\n"); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes DSD of truth table.]

  Description [Returns the number of topmost decomposition node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_TruthDecompose_rec(uTruth uint32, vTree *Vec_Int_t) int32 {
    var Node Ivy_Dec_t;
    var Supp [5]int32;
    var Vars0 [5]int32;
    var Vars1 [5]int32;
    var Vars2 [5]int32;
    var pVars *int32 = nil;
    var nSupp, Count0, Count1, Count2, nVars, RetValue, fCompl, i int32;
    var uTruthCof, uCof0, uCof1 uint32;

    // get constant confactors
    Count0 = Count1 = Count2 = nSupp = 0;
    for  i = 0; i < 5; i++  {
        if Ivy_TruthCofactorIsConst(uTruth, i, 0, 0) != 0 {
            Vars0[Count0++] = (i << 1) | 0;
        } else if Ivy_TruthCofactorIsConst(uTruth, i, 1, 0) != 0 {
            Vars0[Count0++] = (i << 1) | 1;
        } else if Ivy_TruthCofactorIsConst(uTruth, i, 0, 1) != 0 {
            Vars1[Count1++] = (i << 1) | 0;
        } else if Ivy_TruthCofactorIsConst(uTruth, i, 1, 1) != 0 {
            Vars1[Count1++] = (i << 1) | 1;
        } else {
            uCof0 = Ivy_TruthCofactor( uTruth, (i << 1) | 1 );
            uCof1 = Ivy_TruthCofactor( uTruth, (i << 1) | 0 );
            if uCof0 == ~uCof1 {
                Vars2[Count2++] = (i << 1) | 0;
            } else if uCof0 != uCof1 {
                Supp[nSupp++] = i; }
        }
    }
    assert(Count0 == 0 || Count1 == 0, "Count0 == 0 || Count1 == 0");
    assert(Count0 == 0 || Count2 == 0, "Count0 == 0 || Count2 == 0");
    assert(Count1 == 0 || Count2 == 0, "Count1 == 0 || Count2 == 0");

    // consider the case of a single variable
    if Count0 == 1 && nSupp == 0 {
        return Vars0[0]; }

    // consider more complex decompositions
    if Count0 == 0 && Count1 == 0 && Count2 == 0 {
        return Ivy_TruthRecognizeMuxMaj( uTruth, Supp, nSupp, vTree ); }

    // extract the nodes
    Ivy_DecClear( &Node );
    if Count0 > 0 {
        nVars = Count0;
pVars = Vars0;
Node.Type = IVY_DEC_AND;
fCompl = 0;
    } else if Count1 > 0 {
        nVars = Count1;
pVars = Vars1;
Node.Type = IVY_DEC_AND;
fCompl = 1;
uTruth = ~uTruth;
    } else if Count2 > 0 {
        nVars = Count2;
pVars = Vars2;
Node.Type = IVY_DEC_EXOR;
fCompl = 0;
    } else { 
        assert(false, "0"); }
    Node.nFans = nVars+(nSupp>0);

    // compute cofactor
    uTruthCof = uTruth;
    for  i = 0; i < nVars; i++  {
        uTruthCof = Ivy_TruthCofactor( uTruthCof, pVars[i] );
        Ivy_DecSetVar( &Node, i, pVars[i] );
    }

    if Node.Type == IVY_DEC_EXOR {
        fCompl ^= ((Node.nFans & 1) == 0); }

    if nSupp > 0 {
        assert(uTruthCof != 0 && ~uTruthCof != 0, "uTruthCof != 0 && ~uTruthCof != 0");
        // call recursively
        RetValue = Ivy_TruthDecompose_rec( uTruthCof, vTree );
        // quit if non-decomposable
        if RetValue == -1 {
            return -1; }
        // remove the complement from the child if the node is EXOR
        if Node.Type == IVY_DEC_EXOR && (RetValue & 1) {
            fCompl ^= 1;
            RetValue ^= 1;
        }
        // set the new decomposition
        Ivy_DecSetVar( &Node, nVars, RetValue );
    
    } else if Node.Type == IVY_DEC_EXOR {
        fCompl ^= (uTruthCof == 0); }

    Vec_IntPush( vTree, Ivy_DecToInt(Node) );
    return ((Vec_IntSize(vTree)-1) << 1) | fCompl;
}

/**Function*************************************************************

  Synopsis    [Returns a non-negative number if the truth table is a MUX.]

  Description [If the truth table is a MUX, returns the variable as follows:
  first, control variable; second, positive cofactor; third, negative cofactor.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_TruthRecognizeMuxMaj(uTruth uint32, pSupp *int32, nSupp int32, vTree *Vec_Int_t) int32 {
    var Node Ivy_Dec_t;
    var i, k, RetValue0, RetValue1 int32;
    var uCof0, uCof1, Num uint32;
    var Count [3]byte;
    assert(nSupp >= 3, "nSupp >= 3");
    // start the node
    Ivy_DecClear( &Node );
    Node.Type = IVY_DEC_MUX;
    Node.nFans = 3;
    // try each of the variables
    for  i = 0; i < nSupp; i++  {
        // get the cofactors with respect to these variables
        uCof0 = Ivy_TruthCofactor( uTruth, (pSupp[i] << 1) | 1 );
        uCof1 = Ivy_TruthCofactor( uTruth,  pSupp[i] << 1 );
        // go through all other variables and make sure 
        // each of them belongs to the support of one cofactor
        for  k = 0; k < nSupp; k++  {
            if k == i {
                continue; }
            if (Ivy_TruthDepends(uCof0, pSupp[k]) != 0) && (Ivy_TruthDepends(uCof1, pSupp[k]) != 0) {
                break; }
        }
        if k < nSupp {
            continue; }
        // MUX decomposition exists
        RetValue0 = Ivy_TruthDecompose_rec( uCof0, vTree );
        if RetValue0 == -1 {
            break; }
        RetValue1 = Ivy_TruthDecompose_rec( uCof1, vTree );
        if RetValue1 == -1 {
            break; }
        // both of them exist; create the node
        Ivy_DecSetVar( &Node, 0, pSupp[i] << 1 );
        Ivy_DecSetVar( &Node, 1, RetValue1 );
        Ivy_DecSetVar( &Node, 2, RetValue0 );
        Vec_IntPush( vTree, Ivy_DecToInt(Node) );
        return ((Vec_IntSize(vTree)-1) << 1) | 0;
    }
    // check majority gate
    if nSupp > 3 {
        return -1; }
    if Ivy_TruthWordCountOnes(uTruth) != 16 {
        return -1; }
    // this is a majority gate; determine polarity
    Node.Type = IVY_DEC_MAJ;
    Count[0] = Count[1] = Count[2] = 0;
    for  i = 0; i < 8; i++  {
        Num = 0;
        for  k = 0; k < 3; k++  {
            if i & (1 << k) {
                Num |= (1 << pSupp[k]); }
        }
        assert(Num < 32, "Num < 32");
        if (uTruth & (1 << Num)) == 0 {
            continue; }
        for  k = 0; k < 3; k++  {
            if i & (1 << k) {
                Count[k]++; }
        }
    }
    assert(Count[0] == 1 || Count[0] == 3, "Count[0] == 1 || Count[0] == 3");
    assert(Count[1] == 1 || Count[1] == 3, "Count[1] == 1 || Count[1] == 3");
    assert(Count[2] == 1 || Count[2] == 3, "Count[2] == 1 || Count[2] == 3");
    Ivy_DecSetVar( &Node, 0, (pSupp[0] << 1)|(Count[0] == 1) );
    Ivy_DecSetVar( &Node, 1, (pSupp[1] << 1)|(Count[1] == 1) );
    Ivy_DecSetVar( &Node, 2, (pSupp[2] << 1)|(Count[2] == 1) );
    Vec_IntPush( vTree, Ivy_DecToInt(Node) );
    return ((Vec_IntSize(vTree)-1) << 1) | 0;
}


/**Function*************************************************************

  Synopsis    [Computes truth table of decomposition tree for verification.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_TruthDsdCompute_rec(iNode int32, vTree *Vec_Int_t) uint32 {
    var uTruthChild, uTruthTotal uint32;
    var Var, i int32;
    // get the node
    var Node Ivy_Dec_t  = Ivy_IntToDec( Vec_IntEntry(vTree, iNode) );
    // compute the node function
    if Node.Type == IVY_DEC_CONST1 {
        return s_Masks[5][ !Node.fCompl ]; }
    if Node.Type == IVY_DEC_PI {
        return s_Masks[iNode][ !Node.fCompl ]; }
    if Node.Type == IVY_DEC_BUF {
        uTruthTotal = Ivy_TruthDsdCompute_rec( Node.Fan0 >> 1, vTree );
        if Node.fCompl {
            return ~uTruthTotal;
        } else {
            return uTruthTotal;
        }
    }
    if Node.Type == IVY_DEC_AND {
        uTruthTotal = s_Masks[5][1];
        for  i = 0; i < int32(Node.nFans); i++  {
            Var = Ivy_DecGetVar( &Node, i );
            uTruthChild = Ivy_TruthDsdCompute_rec( Var >> 1, vTree );
            if Var & 1 {
                uTruthTotal = uTruthTotal & ~uTruthChild;
            } else {
                uTruthTotal = uTruthTotal & uTruthChild;
            }
        }
        if Node.fCompl {
            return ~uTruthTotal;
        } else {
            return uTruthTotal;
        }
    }
    if Node.Type == IVY_DEC_EXOR {
        uTruthTotal = 0;
        for  i = 0; i < int32(Node.nFans); i++  {
            Var = Ivy_DecGetVar( &Node, i );
            uTruthTotal ^= Ivy_TruthDsdCompute_rec( Var >> 1, vTree );
            assert((Var & 1) == 0, "(Var & 1) == 0");
        }
        if Node.fCompl {
            return ~uTruthTotal;
        } else {
            return uTruthTotal;
        }
    }
    assert(Node.fCompl == 0, "Node.fCompl == 0");
    if Node.Type == IVY_DEC_MUX || Node.Type == IVY_DEC_MAJ {
        var uTruthChildC, uTruthChild1, uTruthChild0 uint32;
        var VarC, Var1, Var0 int32;
        VarC = Ivy_DecGetVar( &Node, 0 );
        Var1 = Ivy_DecGetVar( &Node, 1 );
        Var0 = Ivy_DecGetVar( &Node, 2 );
        uTruthChildC = Ivy_TruthDsdCompute_rec( VarC >> 1, vTree );
        uTruthChild1 = Ivy_TruthDsdCompute_rec( Var1 >> 1, vTree );
        uTruthChild0 = Ivy_TruthDsdCompute_rec( Var0 >> 1, vTree );
        assert(Node.Type == IVY_DEC_MAJ || (VarC & 1) == 0, "Node.Type == IVY_DEC_MAJ || (VarC & 1) == 0");
        if VarC & 1 {
            uTruthChildC = ~uTruthChildC;
        } else {
            uTruthChildC = uTruthChildC;
        }
        if Var1 & 1 {
            uTruthChild1 = ~uTruthChild1;
        } else {
            uTruthChild1 = uTruthChild1;
        }
        if Var0 & 1 {
            uTruthChild0 = ~uTruthChild0;
        } else {
            uTruthChild0 = uTruthChild0;
        }
        if Node.Type == IVY_DEC_MUX {
            return (uTruthChildC & uTruthChild1) | (~uTruthChildC & uTruthChild0);
        } else {
            return (uTruthChildC & uTruthChild1) | (uTruthChildC & uTruthChild0) | (uTruthChild1 & uTruthChild0); }
    }
    assert(false, "0");
    return 0;
}


/**Function*************************************************************

  Synopsis    [Computes truth table of decomposition tree for verification.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_TruthDsdCompute(vTree *Vec_Int_t) uint32 {
    return Ivy_TruthDsdCompute_rec( Vec_IntSize(vTree)-1, vTree );
}

/**Function*************************************************************

  Synopsis    [Prints the decomposition tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_TruthDsdPrint_rec(pFile *os.File, iNode int32, vTree *Vec_Int_t) {
    var Var, i int32;
    // get the node
    var Node Ivy_Dec_t  = Ivy_IntToDec( Vec_IntEntry(vTree, iNode) );
    // compute the node function
    if Node.Type == IVY_DEC_CONST1 {
        fmt.Fprintf(pFile, "Const1%s", (Node.fCompl? "\'" : ""));
    } else if Node.Type == IVY_DEC_PI {
        fmt.Fprintf(pFile, "%c%s", 'a' + iNode, (Node.fCompl? "\'" : ""));
    } else if Node.Type == IVY_DEC_BUF {
        Ivy_TruthDsdPrint_rec( pFile, Node.Fan0 >> 1, vTree );
        fmt.Fprintf(pFile, "%s", (Node.fCompl? "\'" : ""));
    
    } else if Node.Type == IVY_DEC_AND {
        fmt.Fprintf(pFile, "AND(");
        for  i = 0; i < int32(Node.nFans); i++  {
            Var = Ivy_DecGetVar( &Node, i );
            Ivy_TruthDsdPrint_rec( pFile, Var >> 1, vTree );
            fmt.Fprintf(pFile, "%s", (Var & 1)? "\'" : "");
            if i != int32(Node.nFans)-1 {
                fmt.Fprintf(pFile, ","); }
        }
        fmt.Fprintf(pFile, ")%s", (Node.fCompl? "\'" : ""));
    
    } else if Node.Type == IVY_DEC_EXOR {
        fmt.Fprintf(pFile, "EXOR(");
        for  i = 0; i < int32(Node.nFans); i++  {
            Var = Ivy_DecGetVar( &Node, i );
            Ivy_TruthDsdPrint_rec( pFile, Var >> 1, vTree );
            if i != int32(Node.nFans)-1 {
                fmt.Fprintf(pFile, ","); }
            assert((Var & 1) == 0, "(Var & 1) == 0");
        }
        fmt.Fprintf(pFile, ")%s", (Node.fCompl? "\'" : ""));
    
    } else if Node.Type == IVY_DEC_MUX || Node.Type == IVY_DEC_MAJ {
        var VarC, Var1, Var0 int32;
        assert(Node.fCompl == 0, "Node.fCompl == 0");
        VarC = Ivy_DecGetVar( &Node, 0 );
        Var1 = Ivy_DecGetVar( &Node, 1 );
        Var0 = Ivy_DecGetVar( &Node, 2 );
        fmt.Fprintf(pFile, "%s", (Node.Type == IVY_DEC_MUX)? "MUX(" : "MAJ(");
        Ivy_TruthDsdPrint_rec( pFile, VarC >> 1, vTree );
        fmt.Fprintf(pFile, "%s", (VarC & 1)? "\'" : "");
        fmt.Fprintf(pFile, ",");
        Ivy_TruthDsdPrint_rec( pFile, Var1 >> 1, vTree );
        fmt.Fprintf(pFile, "%s", (Var1 & 1)? "\'" : "");
        fmt.Fprintf(pFile, ",");
        Ivy_TruthDsdPrint_rec( pFile, Var0 >> 1, vTree );
        fmt.Fprintf(pFile, "%s", (Var0 & 1)? "\'" : "");
        fmt.Fprintf(pFile, ")");
    
    } else { assert(false, "0"); }
}


/**Function*************************************************************

  Synopsis    [Prints the decomposition tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_TruthDsdPrint(pFile *os.File, vTree *Vec_Int_t) {
    fmt.Fprintf(pFile, "F = ");
    Ivy_TruthDsdPrint_rec( pFile, Vec_IntSize(vTree)-1, vTree );
    fmt.Fprintf(pFile, "\n");
}

/**Function*************************************************************

  Synopsis    [Implement DSD in the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManDsdConstruct_rec(p *Ivy_Man_t, vFront *Vec_Int_t, iNode int32, vTree *Vec_Int_t) *Ivy_Obj_t {
    var pResult *Ivy_Obj_t;
    var pChild *Ivy_Obj_t;
    var pNodes [16]*Ivy_Obj_t;
    var Var, i int32;
    // get the node
    var Node Ivy_Dec_t  = Ivy_IntToDec( Vec_IntEntry(vTree, iNode) );
    // compute the node function
    if Node.Type == IVY_DEC_CONST1 {
        return Ivy_NotCond( Ivy_ManConst1(p), Node.fCompl ); }
    if Node.Type == IVY_DEC_PI {
        pResult = Ivy_ManObj( p, Vec_IntEntry(vFront, iNode) );
        return Ivy_NotCond( pResult, Node.fCompl );
    }
    if Node.Type == IVY_DEC_BUF {
        pResult = Ivy_ManDsdConstruct_rec( p, vFront, Node.Fan0 >> 1, vTree );
        return Ivy_NotCond( pResult, Node.fCompl );
    }
    if Node.Type == IVY_DEC_AND || Node.Type == IVY_DEC_EXOR {
        for  i = 0; i < int32(Node.nFans); i++  {
            Var = Ivy_DecGetVar( &Node, i );
            assert(Node.Type == IVY_DEC_AND || (Var & 1) == 0, "Node.Type == IVY_DEC_AND || (Var & 1) == 0");
            pChild = Ivy_ManDsdConstruct_rec( p, vFront, Var >> 1, vTree );
            pChild = Ivy_NotCond( pChild, (Var & 1) );
            pNodes[i] = pChild;
        }

//        Ivy_MultiEval( pNodes, Node.nFans, Node.Type == IVY_DEC_AND ? IVY_AND : IVY_EXOR );

        pResult = Ivy_Multi( p, pNodes, Node.nFans, Node.Type == IVY_DEC_AND ? IVY_AND : IVY_EXOR );
        return Ivy_NotCond( pResult, Node.fCompl );
    }
    assert(Node.fCompl == 0, "Node.fCompl == 0");
    if Node.Type == IVY_DEC_MUX || Node.Type == IVY_DEC_MAJ {
        var VarC, Var1, Var0 int32;
        VarC = Ivy_DecGetVar( &Node, 0 );
        Var1 = Ivy_DecGetVar( &Node, 1 );
        Var0 = Ivy_DecGetVar( &Node, 2 );
        pNodes[0] = Ivy_ManDsdConstruct_rec( p, vFront, VarC >> 1, vTree );
        pNodes[1] = Ivy_ManDsdConstruct_rec( p, vFront, Var1 >> 1, vTree );
        pNodes[2] = Ivy_ManDsdConstruct_rec( p, vFront, Var0 >> 1, vTree );
        assert(Node.Type == IVY_DEC_MAJ || (VarC & 1) == 0, "Node.Type == IVY_DEC_MAJ || (VarC & 1) == 0");
        pNodes[0] = Ivy_NotCond( pNodes[0], (VarC & 1) );
        pNodes[1] = Ivy_NotCond( pNodes[1], (Var1 & 1) );
        pNodes[2] = Ivy_NotCond( pNodes[2], (Var0 & 1) );
        if Node.Type == IVY_DEC_MUX {
            return Ivy_Mux( p, pNodes[0], pNodes[1], pNodes[2] );
        } else {
            return Ivy_Maj( p, pNodes[0], pNodes[1], pNodes[2] ); }
    }
    assert(false, "0");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Implement DSD in the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManDsdConstruct(p *Ivy_Man_t, vFront *Vec_Int_t, vTree *Vec_Int_t) *Ivy_Obj_t {
    var Entry, i int32;
    // implement latches on the frontier (TEMPORARY!!!)
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFront)) && (((Entry) = Vec_IntEntry(vFront, i)), 1); i++  {
        Vec_IntWriteEntry( vFront, i, Ivy_LeafId(Entry) );
    }
    // recursively construct the tree
    return Ivy_ManDsdConstruct_rec( p, vFront, Vec_IntSize(vTree)-1, vTree );
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ivy_TruthDsdComputePrint_vTree *Vec_Int_t = nil;
func Ivy_TruthDsdComputePrint(uTruth uint32) {
    
    if spStatic_Ivy_TruthDsdComputePrint_vTree == nil {
        spStatic_Ivy_TruthDsdComputePrint_vTree = Vec_IntAlloc( 12 ); }
    if Ivy_TruthDsd( uTruth, spStatic_Ivy_TruthDsdComputePrint_vTree ) != 0 {
        Ivy_TruthDsdPrint( os.Stdout, spStatic_Ivy_TruthDsdComputePrint_vTree );
    } else {
        fmt.Printf("Undecomposable\n"); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ivy_TruthTestOne_Counter int32;
var spStatic_Ivy_TruthTestOne_vTree *Vec_Int_t = nil;
func Ivy_TruthTestOne(uTruth uint32) {
    
    
    // decompose
    if spStatic_Ivy_TruthTestOne_vTree == nil {
        spStatic_Ivy_TruthTestOne_vTree = Vec_IntAlloc( 12 ); }

    if !(Ivy_TruthDsd( uTruth, spStatic_Ivy_TruthTestOne_vTree ) != 0) {
//        printf( "Undecomposable\n" );
    
    } else {
//        nTruthDsd++;
        fmt.Printf("%5d : ", spStatic_Ivy_TruthTestOne_Counter++);
        Extra_PrintBinary( os.Stdout, &uTruth, 32 );
        fmt.Printf("  ");
        Ivy_TruthDsdPrint( os.Stdout, spStatic_Ivy_TruthTestOne_vTree );
        if uTruth != Ivy_TruthDsdCompute(spStatic_Ivy_TruthTestOne_vTree) {
            fmt.Printf("Verification failed.\n"); }
    }
//    Vec_IntFree( vTree );
}

     

                                                                        

                

                
               
                

                

                                                                        
                    
 
                 
                     
                    
          

                                     
                               
                                       
                                    
     
                                      
                                                      
                                                        
                                 
                           
                                   
     
                    
 

                                                                        

                

                
               
                

                

                                                                        
                     
 
                 
                     
                    
          

                                     
                              
     
                                      
                                                      
                                                                          
                                   
     
                    
 

                                                                        

                

                
               
                

                

                                                                        
                     
 
                 
                     
                    
          

                                       
                                 
                                     
                                  
     
                                      
                                                        
                                                      
                                   
                           
                                   
     
                    
 

      


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

