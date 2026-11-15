
#include "aig/gia/giaCut.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static Gia_Man_t *graph(int seed) {
 Gia_Man_t *g=Gia_ManStart(128);int i,a,b,lits[80],count=seed>=16?80:40;unsigned random=seed+307;
 Gia_ManHashStart(g);for(i=0;i<8;++i)lits[i]=Gia_ManAppendCi(g);
 for(i=8;i<12;++i)lits[i]=Gia_ManHashAnd(g,lits[2*(i-8)],lits[2*(i-8)+1]);
 lits[12]=Gia_ManHashAnd(g,lits[8],lits[9]);lits[13]=Gia_ManHashAnd(g,lits[10],lits[11]);lits[14]=Gia_ManHashAnd(g,lits[12],lits[13]);
 for(i=15;i<count;++i){
  random=random*1664525u+1013904223u;a=lits[random%i];random=random*1664525u+1013904223u;a^=random>>31;
  random=random*1664525u+1013904223u;b=lits[random%i];random=random*1664525u+1013904223u;b^=random>>31;
  lits[i]=(i%3==0)?((seed>=8&&seed<16&&(a>>1)!=(b>>1))?Gia_ManAppendXorReal(g,a,b):Gia_ManHashXor(g,a,b)):Gia_ManHashAnd(g,a,b);
 }
 Gia_ManAppendCo(g,lits[14]);Gia_ManAppendCo(g,lits[count-1]);Gia_ManAppendCo(g,lits[count-3]^1);Gia_ManHashStop(g);return g;
}
static void reports(void) {
 Gia_Man_t *g=graph(0);Gia_Sto_t *p=Gia_StoAlloc(g,4,10,1,1,0);Gia_Cut_t cut={0};
 cut.nLeaves=2;cut.pLeaves[0]=1;cut.pLeaves[1]=12;cut.Cost=7;cut.CostLev=9;cut.nTreeLeaves=1;
 Gia_CutPrint(p,3,&cut);Gia_CutPrint(p,3,NULL);Gia_StoFree(p);Gia_ManStop(g);
}
static void driver(void) {
 Gia_Man_t *g=graph(0);Gia_StoComputeCuts(g);Gia_ManStop(g);
}

int main(int argc,char **argv) {
 int first=0,last=8;
 if(argc>1){if(argv[1][0]=='x'){first=8;last=16;}else if(argv[1][0]=='l'){first=16;last=24;}else{if(argv[1][0]=='r')reports();else driver();return 0;}}
 int seed,size,ni,mode,i,j,v,n,nums[]={2,10,50};Gia_Man_t *g;Gia_Sto_t *p;Gia_Obj_t *obj;Vec_Int_t *list;word *truth;
 for(seed=first;seed<last;++seed)for(size=2;size<=8;++size)for(ni=0;ni<3;++ni)for(mode=0;mode<3;++mode){
  g=graph(seed);p=Gia_StoAlloc(g,size,nums[ni],mode>0,mode>1,0);
  Gia_ManForEachObj(g,obj,i)Gia_StoRefObj(p,i);
  Gia_StoComputeCutsConst0(p,0);Gia_ManForEachCi(g,obj,i)Gia_StoComputeCutsCi(p,Gia_ObjId(g,obj));
  Gia_ManForEachAnd(g,obj,i)Gia_StoComputeCutsNode(p,i);
  printf("case %d %d %d %d objs%d counts%.0f:%.0f:%.0f:%.0f over%d\n",seed,size,nums[ni],mode,Gia_ManObjNum(g),p->CutCount[0],p->CutCount[1],p->CutCount[2],p->CutCount[3],p->nCutsOver);
  Gia_ManForEachObj(g,obj,i){int kind=Gia_ObjIsConst0(obj)?0:Gia_ObjIsCi(obj)?1:Gia_ObjIsCo(obj)?6:Gia_ObjIsXor(obj)?3:2;printf("gate%d %d %d %d\n",i,kind,(kind==2||kind==3||kind==6)?Gia_ObjFaninLit0p(g,obj):0,(kind==2||kind==3)?Gia_ObjFaninLit1p(g,obj):0);}
  for(i=0;i<Gia_ManObjNum(g);++i){list=Vec_WecEntry(p->vCuts,i);printf("node%d",i);Vec_IntForEachEntry(list,v,j)printf(" %d",v);printf("\n");}
  if(mode>0)for(i=0;i<Vec_MemEntryNum(p->vTtMem);++i){truth=Vec_MemReadEntry(p->vTtMem,i);printf("truth%d",i);for(j=0;j<Abc_Truth6WordNum(size);++j)printf(" %016llx",(unsigned long long)truth[j]);printf("\n");}
  Gia_StoFree(p);Gia_ManStop(g);
 }
 return 0;
}
