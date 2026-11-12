package dau

import "fmt"

/**CFile****************************************************************

  FileName    [dauArray.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    [Array representation of DSD.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauArray.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Dau_Dsd_t struct {
    iVar uint32; /* C bitfield: 5 bits; review packing and truncation */  // variable
    nFans uint32; /* C bitfield: 5 bits; review packing and truncation */  // fanin count
    Depth uint32; /* C bitfield: 5 bits; review packing and truncation */  // tree depth
    Offset uint32; /* C bitfield: 5 bits; review packing and truncation */  // the diff between this and other node
    Data uint32; /* C bitfield: 5 bits; review packing and truncation */  // user data
    Type uint32; /* C bitfield: 3 bits; review packing and truncation */  // node type
    fCompl uint32; /* C bitfield: 1 bits; review packing and truncation */  // the complemented attribute
    fUnused uint32; /* C bitfield: 1 bits; review packing and truncation */  // this vertice is unused
}

func Dau_DsdClean(p *Dau_Dsd_t) { *((int *)p) = 0; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
func Dau_DsdCountAnd(p *Dau_Dsd_t) int32 {
    var Count int32;
    var Costs [7]int32 = [7]int32{0, 0, 0, 1, 3, 3, 3};
    for  Count = 0; p.Type; p++  {
        Count += Costs[p.Type];
    }
    return Count;
}

/*
void Dau_DsdMark( Dau_Dsd_t * p, int nSize, int * pMarks )
{
    int pStore[DAU_MAX_VAR] = {0};
    Dau_Dsd_t * q;
    if ( p->Type == DAU_DSD_CONST || p->Type == DAU_DSD_VAR )
        return;
    for ( q = p + nSize - 1; q >= p; q-- )
    {
        if ( q->Type == DAU_DSD_VAR )
            pStore[q->Depth] += pMarks[q->iVar];
        else 
        {
            q->Data = pStore[q->Depth+1]; pStore[q->Depth+1] = 0;
            pStore[q->Depth] += (q->Data == q->nFans);            
        }
    }
}
*/

func Dau_DsdConstruct(pDsd string, pStore *Dau_Dsd_t) int32 {
    var pLevel [DAU_MAX_VAR]*Dau_Dsd_t;
    var q *Dau_Dsd_t  = pStore;
    var d int32 = -1;
    var fCompl int32;
    if Dau_DsdIsConst(pDsd) {
        Dau_DsdClean( q );
        q.Type   = DAU_DSD_CONST0;
        q.fCompl = Dau_DsdIsConst1(pDsd);
        return 1;
    }
    for  --q; *pDsd; pDsd++  {
        if *pDsd == '!' {
            fCompl ^= 1;
            continue;
        }
        if *pDsd == ')' || *pDsd == ']' || *pDsd == '>' || *pDsd == '}' {
            assert(fCompl == 0, "fCompl == 0");
            if d--; d >= 0 {
                pLevel[d].nFans++;
                if pLevel[d].Data > pLevel[d+1].Data {
                    pLevel[d].Data = pLevel[d+1].Data; }
            }
            continue;
        }
        Dau_DsdClean( ++q );
        q.Data = 31;
        q.fCompl = fCompl;
        fCompl = 0;
        if *pDsd >= 'a' && *pDsd <= 'z' {
            q.Type   = DAU_DSD_VAR;
            q.iVar   = *pDsd - 'a';
            q.Depth  = d + 1;
            if d >= 0 {
                pLevel[d].nFans++;
                if pLevel[d].Data > q.iVar {
                    pLevel[d].Data = q.iVar; }
            }
            continue;
        }
        if *pDsd == '(' {
            q.Type = DAU_DSD_AND;
        } else if *pDsd == '[' {
            q.Type = DAU_DSD_XOR;
        } else if *pDsd == '<' {
            q.Type = DAU_DSD_MUX;
        } else if *pDsd == '{' {
            q.Type = DAU_DSD_PRIME;
        } else { assert(false, "0"); }
        pLevel[++d] = q;
        q.Depth = d;
    }
    assert(d == -1, "d == -1");
    Dau_DsdClean( ++q );
    return q - pStore;
}

func Dau_DsdPrint(p *Dau_Dsd_t) {
    var OpenType [7]byte = [7]byte{0, 0, 0, '(', '[', '<', '{'};
    var CloseType [7]byte = [7]byte{0, 0, 0, ')', ']', '>', '}'};
    var pTypes [DAU_MAX_VAR]byte;
    var d int32;
    var pVisits [DAU_MAX_VAR]int32;
    if p.Type == DAU_DSD_CONST0 {
        fmt.Printf("%d\n", p.fCompl);
        return;
    }
    pVisits[0] = 1;
    for  d = 0; p.Type; p++  {
        if p.fCompl {
            fmt.Printf("!"); }
        if p.Type == DAU_DSD_VAR {
            fmt.Printf("%c", 'a' + p.iVar);
            for  d > 0 && --pVisits[d] == 0  {
                fmt.Printf("%c", pTypes[d--]);
            }
        
        } else {
            pVisits[++d] = p.nFans;
            fmt.Printf("%c", OpenType[p.Type]);
            fmt.Printf("%c", 'a' + p.Data);
            fmt.Printf("%d", p.Depth);
            pTypes[d]  = CloseType[p.Type];
        }
    }
    assert(d == 0, "d == 0");
    fmt.Printf("\n");
}

func Dau_DsdDepth(p *Dau_Dsd_t) {
    var d int32;
    var pVisits [DAU_MAX_VAR]int32;
    if p.Type == DAU_DSD_CONST0 {
        return; }
    pVisits[0] = 1;
    for  d = 0; p.Type; p++  {
        p.Depth = d;
        if p.Type == DAU_DSD_VAR {
            for  d > 0 && --pVisits[d] == 0  {
                d--;
            }
        } else {
            pVisits[++d] = p.nFans; }
    }
    assert(d == 0, "d == 0");
}

func Dau_DsdRemoveUseless(p *Dau_Dsd_t) {
    var q *Dau_Dsd_t = p;
    var pLevel [DAU_MAX_VAR]*Dau_Dsd_t;
    var d int32;
    var fChange int32;
    var pVisits [DAU_MAX_VAR]int32;
    if p.Type == DAU_DSD_CONST0 {
        return; }
    pVisits[0] = 1;
    for  d = 0; p.Type; p++  {
        p.Depth = d;
        if p.Type == DAU_DSD_VAR {
            for  d > 0 && --pVisits[d] == 0  {
                d--;
            }
        } else {
            if d > 0 && (pLevel[d-1].Type == DAU_DSD_XOR && p.Type == DAU_DSD_XOR ||
                           pLevel[d-1].Type == DAU_DSD_AND && p.Type == DAU_DSD_AND && !p.fCompl) {
                pLevel[d-1].nFans += p.nFans - 1;
                pVisits[d] += p.nFans - 1;
                p.fUnused = 1;
                fChange = 1;
            
            } else {
                pLevel[d++] = p;
                pVisits[d] = p.nFans;
            }
        }
    }
    assert(d == 0, "d == 0");
    // compact
    if fChange != 0 {
        for  p = q; p.Type; p++  {
            if !p.fUnused {
                *q++ = *p; }
        }
        Dau_DsdClean( q );
    }
}

func Dau_DsdTest22() {
    var pStore [2 * DAU_MAX_VAR]Dau_Dsd_t;
//    char * pDsd = "[(ab)c(f!(he))]";
//    char * pDsd = "[(abd)cf(f!{she})]";
    var pDsd string  = "[(abd)[cf](f(sg(he)))]";
//    char * pDsd = "[(ab)[cf]]";
    var i int32;
    var nSize int32 = Dau_DsdConstruct( pDsd, pStore );
//    Dau_DsdDepth( pStore );
    Dau_DsdPrint( pStore );

    Dau_DsdRemoveUseless( pStore );

    Dau_DsdPrint( pStore );
    i = 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

