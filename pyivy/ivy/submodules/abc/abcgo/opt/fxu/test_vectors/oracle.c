// Original-source oracle: link the unchanged ABC FXU implementation.
#include "opt/fxu/fxuInt.h"
#include "opt/fxu/fxu.h"
extern int Fxu_FastExtract(Fxu_Data_t *);
int main(int argc,char **argv) {
 Fxu_Data_t d = {0}; int i,j,mode=argc>1?atoi(argv[1]):0;
 d.vSopsNew=Vec_PtrStart(26); d.vFaninsNew=Vec_PtrStart(26); d.nNodesOld=6; d.fOnlyS=1; d.nNodesExt=20; d.nSingleMax=20000; d.nPairsMax=30000;
 d.vSops=Vec_PtrStart(6); d.vFanins=Vec_PtrStart(6); d.pManSop=Mem_FlexStart();
 Vec_PtrWriteEntry(d.vSops,5,"111-- 1\n11-1- 1\n11--1 1\n");
 Vec_Int_t *v=Vec_IntAlloc(5); for(i=0;i<5;i++) Vec_IntPush(v,i); Vec_PtrWriteEntry(d.vFanins,5,v);
 if(mode) {
 d.fOnlyS=0; d.fUse0=1; d.fUseCompl=mode==2; d.nNodesOld=8;
 Vec_PtrFree(d.vSops); Vec_PtrFree(d.vFanins);
 d.vSops=Vec_PtrStart(8); d.vFanins=Vec_PtrStart(8);
 Vec_PtrWriteEntry(d.vSops,5,"1-1-- 1\n-11-- 1\n");
 Vec_PtrWriteEntry(d.vSops,6,"1--1- 1\n-1-1- 1\n");
 Vec_PtrWriteEntry(d.vSops,7,"00--1 1\n");
 for(i=5;i<8;i++){v=Vec_IntAlloc(5);for(j=0;j<5;j++)Vec_IntPush(v,j);Vec_PtrWriteEntry(d.vFanins,i,v);}
 }
 printf("COUNT %d\n",Fxu_FastExtract(&d));
 for(i=0;i<Vec_PtrSize(d.vSopsNew);i++) if(Vec_PtrEntry(d.vSopsNew,i)) {
  printf("NODE %d",i); v=(Vec_Int_t*)Vec_PtrEntry(d.vFaninsNew,i);
  for(j=0;j<Vec_IntSize(v);j++) printf(" %d",Vec_IntEntry(v,j));
  printf("\n%s",(char*)Vec_PtrEntry(d.vSopsNew,i));
 }
 return 0;
}
