#include "aig/aig/aigMffc.c"
#include <stdio.h>
void Aig_ManIncrementTravId(Aig_Man_t *p) { p->nTravIds++; }
int main(void) {
 for(int shared=0;shared<=1;shared++) {
  Aig_Man_t m={0};Aig_Obj_t o[5]={0};
  for(int i=0;i<5;i++){o[i].Id=i;o[i].Type=i<3?AIG_OBJ_CI:AIG_OBJ_AND;}
  o[0].nRefs=o[1].nRefs=o[2].nRefs=1;
  o[3].pFanin0=o;o[3].pFanin1=o+1;o[3].nRefs=1+shared;o[3].Level=1;
  o[4].pFanin0=o+2;o[4].pFanin1=o+3;o[4].nRefs=1;o[4].Level=2;
  Vec_Ptr_t *leaves=Vec_PtrAlloc(5),*extended=Vec_PtrAlloc(5);
  int size=Aig_NodeMffcSupp(&m,o+4,0,leaves);
  printf("%d %d",size,Vec_PtrSize(leaves));Aig_Obj_t *p;int i;
  Vec_PtrForEachEntry(Aig_Obj_t *,leaves,p,i) printf(" %d",p->Id);
  int success=Aig_NodeMffcExtendCut(&m,o+4,leaves,extended);
  printf(" %d %d",success,Vec_PtrSize(extended));
  Vec_PtrForEachEntry(Aig_Obj_t *,extended,p,i) printf(" %d",p->Id);
  puts("");Vec_PtrFree(leaves);Vec_PtrFree(extended);
 }
 return 0;
}
