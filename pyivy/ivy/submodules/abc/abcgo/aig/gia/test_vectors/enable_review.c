#include "aig/gia/giaEnable.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static Gia_Man_t *graph(int kind,int phase){
 Gia_Man_t *g=Gia_ManStart(32);int c=Gia_ManAppendCi(g),d=Gia_ManAppendCi(g),r[2],next[2],i;
 for(i=0;i<2;++i)r[i]=Gia_ManAppendCi(g);
 for(i=0;i<2;++i){
  if(kind)next[i]=Gia_ManAppendMux(g,c^(phase&1),d^((phase>>1)&1),r[i]^((phase>>2)&1))^((phase>>2)&1);
  else next[i]=Gia_ManAppendAnd(g,c^(phase&1),r[i]^((phase>>1)&1))^((phase>>2)&1);
 }
 Gia_ManAppendCo(g,r[0]);Gia_ManAppendCo(g,d);for(i=0;i<2;++i)Gia_ManAppendCo(g,next[i]);Gia_ManSetRegNum(g,2);return g;
}
static void dump(const char *name,Gia_Man_t *g){
 int values[512],mask,id,i;Gia_Obj_t *obj;
 printf("%s objects%d cis%d cos%d regs%d\n",name,Gia_ManObjNum(g),Gia_ManCiNum(g),Gia_ManCoNum(g),Gia_ManRegNum(g));
 for(mask=0;mask<(1<<Gia_ManCiNum(g));++mask){
  values[0]=0;Gia_ManForEachObj1(g,obj,id){if(Gia_ObjIsCi(obj))values[id]=(mask>>Gia_ObjCioId(obj))&1;else if(Gia_ObjIsAnd(obj))values[id]=(values[Gia_ObjFaninId0(obj,id)]^Gia_ObjFaninC0(obj))&(values[Gia_ObjFaninId1(obj,id)]^Gia_ObjFaninC1(obj));else values[id]=values[Gia_ObjFaninId0(obj,id)]^Gia_ObjFaninC0(obj);}
  printf("truth %d",mask);Gia_ManForEachCo(g,obj,i)printf(" %d",values[Gia_ObjId(g,obj)]);puts("");
 }
 printf("classes");if(g->vFlopClasses)for(i=0;i<Vec_IntSize(g->vFlopClasses);++i)printf(" %d",Vec_IntEntry(g->vFlopClasses,i));puts("");
}
int main(void){int kind,phase,nf,limit,limits[]={0,1,3};for(kind=0;kind<2;++kind)for(phase=0;phase<8;++phase){
 Gia_Man_t *g=graph(kind,phase),*q;Vec_Int_t *s;int i;
 printf("case %d %d\n",kind,phase);Gia_ManDetectSeqSignals(g,1,1);
 for(limit=0;limit<3;++limit){s=Gia_ManDetectSeqSignalsWithFanout(g,limits[limit],1);printf("signals %d",limits[limit]);for(i=0;i<Vec_IntSize(s);++i)printf(" %d",Vec_IntEntry(s,i));puts("");Vec_IntFree(s);}
 for(nf=1;nf<=3;++nf){printf("frames %d\n",nf);q=Gia_ManUnrollInit(g,nf);dump("unrolled",q);printf("copies");for(i=0;i<nf*Gia_ManObjNum(g);++i)printf(" %d",Vec_IntEntry(&g->vCopies,i));puts("");Gia_ManStop(q);}
 q=Gia_ManRemoveEnables(g);dump("removed",q);Gia_ManStop(q);if(kind){q=Gia_ManRemoveEnables2(g);dump("removed2",q);Gia_ManStop(q);}
 Gia_ManStop(g);
 }return 0;}
