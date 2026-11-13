#include "opt/dar/darInt.h"
#include "bool/kit/kit.h"
#include <stdio.h>
/* Empty strash table: all nontrivial factored nodes are new. */
Aig_Obj_t *Aig_TableLookupTwo(Aig_Man_t *p, Aig_Obj_t *a, Aig_Obj_t *b) { (void)p; (void)a; (void)b; return NULL; }
int main(void) {
 unsigned truth; int maxNodes,maxLevel;
 Vec_Int_t *memory=Vec_IntAlloc(1<<12);
 Aig_Man_t manager={0}; manager.nTravIds=1;
 Aig_Obj_t objects[6]={0}; Vec_Ptr_t *leaves=Vec_PtrAlloc(5);
 for(int i=0;i<5;i++) { objects[i].Id=i+1; objects[i].Type=AIG_OBJ_CI; objects[i].Level=i%3; Vec_PtrPush(leaves,objects+i); }
 objects[5].Id=6;objects[5].Type=AIG_OBJ_AND;
 while(scanf("%x %d %d",&truth,&maxNodes,&maxLevel)==3) {
  for(int phase=0;phase<2;phase++) {
   unsigned t=phase?~truth:truth;
   int r=Kit_TruthIsop(&t,5,memory,0);
   if(r<0){puts("-1");continue;}
   Kit_Graph_t *f=Kit_SopFactor(memory,phase,5,memory);
   printf("%d\n",Dar_RefactTryGraph(&manager,objects+5,leaves,f,maxNodes,maxLevel));
   Kit_GraphFree(f);
  }
 }
 Vec_PtrFree(leaves);Vec_IntFree(memory);return 0;
}
