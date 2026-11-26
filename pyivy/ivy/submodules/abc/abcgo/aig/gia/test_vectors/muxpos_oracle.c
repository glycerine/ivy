// Runs ABC's Gia_ManDupWithMuxPos on an AIGER file and prints the number of
// original COs, added COs, and added COs whose driver also drives an
// original CO (which the function means to skip).
#include "aig/gia/gia.h"
#include <stdio.h>
extern Gia_Man_t * Gia_ManDupWithMuxPos( Gia_Man_t * p );
int main( int argc, char ** argv )
{
    Gia_Man_t * p = Gia_AigerRead( argv[1], 0, 0, 0 ), * q;
    int i, k, nOld = Gia_ManCoNum(p), nDup = 0;
    q = Gia_ManDupWithMuxPos( p );
    for ( i = nOld; i < Gia_ManCoNum(q); i++ )
        for ( k = 0; k < nOld; k++ )
            if ( Gia_ObjFaninId0p(q, Gia_ManCo(q, i)) == Gia_ObjFaninId0p(q, Gia_ManCo(q, k)) )
            { nDup++; break; }
    printf( "old=%d added=%d dup=%d\n", nOld, Gia_ManCoNum(q) - nOld, nDup );
    return 0;
}
