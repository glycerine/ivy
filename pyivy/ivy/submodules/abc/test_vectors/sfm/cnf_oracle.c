// Oracle for Sfm_TruthToCnf (src/opt/sfm/sfmCnf.c).
// Input lines: nVars followed by Abc_Truth6WordNum(nVars) hex words.
// Output lines: the return value, then the CNF entries (-1 ends a clause).
#include "opt/sfm/sfmCnf.c"
#include "bool/kit/kitIsop.c"
#include "bool/kit/kitTruth.c"
int main(void)
{
    int nVars;
    Vec_Int_t * vCover = Vec_IntAlloc( 1 << 16 );
    Vec_Str_t * vCnf = Vec_StrAlloc( 100 );
    while ( scanf( "%d", &nVars ) == 1 )
    {
        word pTruth[1 << 10];
        int w, i, nWords = Abc_Truth6WordNum( nVars ), nCubes;
        for ( w = 0; w < nWords; w++ )
            if ( scanf( "%lx", (unsigned long *)&pTruth[w] ) != 1 )
                return 1;
        nCubes = Sfm_TruthToCnf( pTruth[0], pTruth, nVars, vCover, vCnf );
        printf( "%d", nCubes );
        for ( i = 0; i < Vec_StrSize(vCnf); i++ )
            printf( " %d", (int)(signed char)Vec_StrEntry(vCnf, i) );
        printf( "\n" );
    }
    Vec_IntFree( vCover );
    Vec_StrFree( vCnf );
    return 0;
}
