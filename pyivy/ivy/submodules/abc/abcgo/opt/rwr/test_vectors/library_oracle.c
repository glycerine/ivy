#include "opt/rwr/rwrLib.c"
#include <stdio.h>
int main(void) {
 Rwr_Man_t p={0};p.pTable=ABC_CALLOC(Rwr_Node_t *,1<<16);p.puCanons=ABC_ALLOC(unsigned short,1<<16);p.pPractical=ABC_ALLOC(char,1<<16);
 for(int i=0;i<(1<<16);i++){p.puCanons[i]=i;p.pPractical[i]=1;}
 p.vForest=Vec_PtrAlloc(8);p.pMmNode=Extra_MmFixedStart(sizeof(Rwr_Node_t));
 Rwr_Node_t *a=Rwr_ManAddVar(&p,0xaaaa,0),*b=Rwr_ManAddVar(&p,0xcccc,0);
 Rwr_Node_t *first=Rwr_ManTryNode(&p,a,Rwr_Not(b),0,1,1);
 Rwr_Node_t *reject=Rwr_ManTryNode(&p,a,Rwr_Not(b),0,2,2);
 Rwr_Node_t *equal=Rwr_ManTryNode(&p,a,Rwr_Not(b),0,1,1);
 Rwr_Node_t *xor=Rwr_ManAddNode(&p,first,a,1,2,3);
 int volume=Rwr_ManNodeVolume(&p,xor,first);
 printf("%u %d %d %d %d %d %d %d %d\n",first->uTruth,reject==NULL,first->pNext==equal,volume,p.nConsidered,p.nAdded,p.nClasses,Vec_PtrSize(p.vForest),xor->uTruth);
 Extra_MmFixedStop(p.pMmNode);Vec_PtrFree(p.vForest);ABC_FREE(p.pTable);ABC_FREE(p.puCanons);ABC_FREE(p.pPractical);
 return 0;
}
