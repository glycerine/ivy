/* Tests execute the actual ABC routines, with real ABC object/vector layouts. */
#include "opt/sim/simUtils.c"
#include "opt/sim/simSwitch.c"
#include "opt/sim/simSupp.c"
#include "opt/sim/simSymStr.c"
#include "opt/sim/simSymSim.c"
#include "misc/extra/extraUtilBitMatrix.c"
int main(void) {
 char mode;
 while (scanf(" %c", &mode)==1) {
  if (mode=='W') {
   int words,c0,c1;scanf("%d%d%d",&words,&c0,&c1);
   unsigned a[128],b[128],z[128];for(int i=0;i<words;i++)scanf("%x",a+i);for(int i=0;i<words;i++)scanf("%x",b+i);
   Abc_Ntk_t ntk={0};Abc_Obj_t obj={0};int fanins[2]={0,1};obj.Type=ABC_OBJ_NODE;obj.Id=2;obj.pNtk=&ntk;obj.vFanins.nSize=2;obj.vFanins.pArray=fanins;obj.fCompl0=c0;obj.fCompl1=c1;
   void *rows[]={a,b,z};Vec_Ptr_t info={0};info.nSize=3;info.pArray=rows;Sim_UtilSimulateNodeOne(&obj,&info,words,0);
   for(int i=0;i<words;i++)printf("%x ",z[i]);
   printf("%d %d %d %d %.9g\n",Sim_UtilCountOnes(a,words),Sim_UtilInfoIsEqual(a,b,words),Sim_UtilInfoIsImp(a,b,words),Sim_UtilInfoIsClause(a,b,words),Sim_ComputeSwitching(a,words));
  } else if (mode=='D') {
   int n;scanf("%d",&n);int words=SIM_NUM_WORDS(n);unsigned support[8]={0},pat[8],vals[8];for(int i=0;i<n;i++)Sim_SetBit(support,i);for(int w=0;w<words;w++)scanf("%x",pat+w);for(int w=0;w<words;w++)scanf("%x",vals+w);
   Sym_Man_t p={0};Abc_Obj_t node={0};node.Id=0;p.nSimWords=words;p.vSim=Vec_PtrAlloc(1);Vec_PtrPush(p.vSim,vals);p.vSuppFun=Vec_PtrAlloc(1);Vec_PtrPush(p.vSuppFun,support);p.vSupports=Vec_VecStart(1);for(int i=0;i<n;i++)Vec_VecPushInt(p.vSupports,0,i);p.uPatCol=calloc(words,4);p.uPatRow=calloc(words,4);
   Extra_BitMat_t *mat=Extra_BitMatrixStart(n);Vec_Ptr_t *mats=Vec_PtrAlloc(1);Vec_PtrPush(mats,mat);Sim_SymmsDeriveInfo(&p,pat,&node,mats,0);
   for(int i=0;i<n;i++)for(int j=i+1;j<n;j++)printf("%d",Extra_BitMatrixLookup1(mat,i,j));puts("");
   Extra_BitMatrixStop(mat);Vec_PtrFree(mats);Vec_PtrFree(p.vSim);Vec_PtrFree(p.vSuppFun);Vec_VecFree(p.vSupports);free(p.uPatCol);free(p.uPatRow);
  } else if (mode=='S') {
   int ni,ng;scanf("%d%d",&ni,&ng);int count=1+ni+ng;Abc_Ntk_t ntk={0};ntk.vObjs=Vec_PtrStart(count);ntk.vCis=Vec_PtrAlloc(ni);ntk.vCos=Vec_PtrAlloc(0);Abc_Obj_t *objs=calloc(count,sizeof(Abc_Obj_t));
   for(int i=0;i<count;i++){Abc_Obj_t *o=objs+i;o->Id=i;o->pNtk=&ntk;Vec_PtrWriteEntry(ntk.vObjs,i,o);o->Type=i==0?ABC_OBJ_CONST1:i<=ni?ABC_OBJ_PI:ABC_OBJ_NODE;if(i&&i<=ni){Vec_PtrPush(ntk.vCis,o);o->pCopy=(Abc_Obj_t*)Vec_IntAlloc(0);}if(i>ni){int a,b;scanf("%d%d",&a,&b);o->vFanins.nSize=o->vFanins.nCap=2;o->vFanins.pArray=malloc(8);o->vFanins.pArray[0]=a>>1;o->vFanins.pArray[1]=b>>1;o->fCompl0=a&1;o->fCompl1=b&1;}}
   ntk.vSupps=Sim_ComputeStrSupp(&ntk);int *map=Sim_SymmsCreateMap(&ntk);for(int i=ni+1;i<count;i++)Sim_SymmsStructComputeOne(&ntk,objs+i,map);
   for(int i=ni+1;i<count;i++){Vec_Int_t *pairs=(Vec_Int_t*)objs[i].pCopy;printf("%x:",((unsigned*)ntk.vSupps->pArray[i])[0]);for(int a=0;a<ni;a++)for(int b=a+1;b<ni;b++)printf("%d",Vec_IntFind(pairs,(a<<16)|b)>=0);printf(" ");}puts("");
   for(int i=1;i<count;i++){Vec_IntFree((Vec_Int_t*)objs[i].pCopy);if(i>ni)free(objs[i].vFanins.pArray);}free(map);Sim_UtilInfoFree(ntk.vSupps);Vec_PtrFree(ntk.vObjs);Vec_PtrFree(ntk.vCis);Vec_PtrFree(ntk.vCos);free(objs);
  } else return 2;
 }
 return 0;
}
