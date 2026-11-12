/* Original root/divisor algorithms on real ABC graph storage. */
#include "opt/mfs/mfsWin.c"
#include "opt/mfs/mfsDiv.c"
#include "base/abc/abcUtil.c"
int main(void){int ni,ng,no,target,levels,fanouts,cap,maxlev;
 while(scanf("%d%d%d%d%d%d%d%d",&ni,&ng,&no,&target,&levels,&fanouts,&cap,&maxlev)==8){
 int total=1+ni+ng+no;Abc_Ntk_t n={0};n.vObjs=Vec_PtrStart(total);n.vCis=Vec_PtrAlloc(ni);n.vCos=Vec_PtrAlloc(no);Abc_Obj_t *objs=calloc(total,sizeof(Abc_Obj_t));
 for(int i=0;i<total;i++){Abc_Obj_t *v=objs+i;v->pNtk=&n;v->Id=i;v->Type=i==0?ABC_OBJ_CONST1:i<=ni?ABC_OBJ_PI:i<=ni+ng?ABC_OBJ_NODE:ABC_OBJ_PO;Vec_PtrWriteEntry(n.vObjs,i,v);if(i&&i<=ni)Vec_PtrPush(n.vCis,v);
 if(i>ni){int count=1;if(i<=ni+ng)scanf("%d",&count);for(int k=0;k<count;k++){int f;scanf("%d",&f);Vec_IntPush(&v->vFanins,f);Vec_IntPush(&objs[f].vFanouts,i);if(v->Level<=objs[f].Level)v->Level=objs[f].Level+1;}if(i>ni+ng)Vec_PtrPush(n.vCos,v);}}
 Mfs_Par_t pars={0};pars.nWinMax=cap;pars.nFanoutsMax=fanouts;Mfs_Man_t p={0};p.pPars=&pars;p.pNtk=&n;p.vNodes=Vec_PtrAlloc(0);
 Vec_Ptr_t *roots=Abc_MfsComputeRoots(objs+target,levels,fanouts);for(int i=0;i<roots->nSize;i++)printf("%d,",((Abc_Obj_t*)roots->pArray[i])->Id);printf(" ");Vec_Ptr_t *divs=Abc_MfsComputeDivisors(&p,objs+target,maxlev);for(int i=0;i<divs->nSize;i++)printf("%d,",((Abc_Obj_t*)divs->pArray[i])->Id);puts("");
 Vec_PtrFree(roots);Vec_PtrFree(divs);Vec_PtrFree(p.vNodes);for(int i=0;i<total;i++){ABC_FREE(objs[i].vFanins.pArray);ABC_FREE(objs[i].vFanouts.pArray);}ABC_FREE(n.vTravIds.pArray);Vec_PtrFree(n.vObjs);Vec_PtrFree(n.vCis);Vec_PtrFree(n.vCos);free(objs);
 }return 0;}
