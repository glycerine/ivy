/* Match optional exported functions in the existing ABC archive so its
   callers resolve to this freshly compiled source, not a second ifDsd.o. */
#define ABC_USE_CUDD
#define ABC_USE_PTHREADS
#include "map/if/ifDsd.c"
int main(void)
{
    If_DsdMan_t *p=If_DsdManAlloc(3,2);
    word truth=ABC_CONST(0x8080808080808080);unsigned char perm[3];int lit,i;
    printf("initial %d %s\n",If_DsdManObjNum(p),If_DsdManFileName(p));
    lit=If_DsdManCompute(p,&truth,3,perm,"22");
    printf("computed %d %d %d permutation",lit,If_DsdManObjNum(p),If_DsdManSuppSize(p,lit));
    for(i=0;i<3;i++)printf(" %d",perm[i]);puts("");
    printf("partition %x\n",If_DsdManCheckXY(p,lit,2,1,0,1,0));
    lit=If_DsdManCompute(p,&truth,3,perm,"22");
    printf("reused %d %d\n",lit,If_DsdManObjNum(p));
    printf("masked-feasibility %x derived %x\n",If_DsdManCheckXY(p,lit,2,0,~0u,0,0),If_DsdManCheckXY(p,lit,2,1,~0u,0,0));
    If_DsdManAllocIsops(p,3);
    printf("isop-lut-size %d\n",If_DsdManLutSize(p));
    If_DsdManFree(p,0);return 0;
}
