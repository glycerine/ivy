// Oracle for CBUGS_FIXED.md #14: Sim_UtilAssignFromFifo flips bit i of
// input i in each series of nSuppWords words, but the last series may be
// shorter, and then the flip lands in a following object's row.
#include "base/main/main.h"
#include "base/io/ioAbc.h"
#include "opt/sim/sim.h"
#include "opt/sim/simSupp.c" // Sim_UtilAssignFromFifo is static

static int CountOnes( unsigned w ) { int n = 0; for ( ; w; w &= w - 1 ) n++; return n; }

int main( int argc, char ** argv )
{
    Abc_Ntk_t * pNtk, * pAig;
    Abc_Obj_t * pObj;
    Sim_Man_t * p;
    int i, k, w, Bad = 0;
    Abc_Start();
    pNtk = Io_Read( argv[1], IO_FILE_AIGER, 1, 0 );
    pAig = Abc_NtkStrash( pNtk, 0, 1, 0 );
    p = Sim_ManStart( pAig, 0 );
    for ( i = 0; i < Abc_NtkObjNumMax(pAig); i++ )
        memset( p->vSim0->pArray[i], 0, sizeof(unsigned) * p->nSimWords );
    for ( k = 0; k < 8; k++ ) // all-zero patterns
        Vec_PtrPush( p->vFifo, Sim_ManPatAlloc(p) );
    Sim_UtilAssignFromFifo( p );
    // each CI row holds only its own flips; every other row stays zero
    Abc_NtkForEachObj( pAig, pObj, i )
    {
        unsigned * pSim = (unsigned *)p->vSim0->pArray[pObj->Id];
        int nOnes = 0;
        for ( w = 0; w < p->nSimWords; w++ )
            nOnes += CountOnes( pSim[w] );
        pObj->iTemp = nOnes;
    }
    Abc_NtkForEachCi( pAig, pObj, k )
    {
        int iWord, nExp = 0;
        for ( iWord = 0; iWord < p->nSimWords; iWord += p->nSuppWords )
            nExp += iWord + (k >> 5) < Abc_MinInt(iWord + p->nSuppWords, p->nSimWords);
        pObj->iTemp -= nExp;
    }
    Abc_NtkForEachObj( pAig, pObj, i )
        Bad += pObj->iTemp != 0;
    printf( "words=%d supp=%d corrupted=%d\n", p->nSimWords, p->nSuppWords, Bad );
    return 0;
}
