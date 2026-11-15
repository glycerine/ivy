#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern void Gia_ManPrepareWin(Gia_Man_t *,Vec_Int_t *,Vec_Int_t **,Vec_Int_t **,Vec_Int_t **,int);
extern Gia_Man_t *Gia_ManExtractWin(Gia_Man_t *,Vec_Int_t *,int);
extern Gia_Man_t *Gia_ManInsertWin(Gia_Man_t *,Vec_Int_t *,Gia_Man_t *);
extern Vec_Int_t *Gia_ManFindLatest(Gia_Man_t *,int,int);
static Gia_Man_t *graph(int signs,int layout){
 Gia_Man_t *g=Gia_ManStart(32);int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=0,d=0,x,y,z,m,i;Gia_Obj_t *o;
 if(!(layout&1)){c=Gia_ManAppendCi(g);d=Gia_ManAppendCi(g);}
 x=Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
 if(layout&1){Gia_ManAppendCo(g,x^(signs&1));c=Gia_ManAppendCi(g);}
 y=Gia_ManAppendAnd(g,x,c^((signs>>2)&1));
 if(layout&1){Gia_ManAppendCo(g,y^((signs>>1)&1));d=Gia_ManAppendCi(g);}
 z=Gia_ManAppendAnd(g,b^1,d);m=Gia_ManAppendAnd(g,y^1,z);
 if(!(layout&1)){Gia_ManAppendCo(g,x^(signs&1));Gia_ManAppendCo(g,y^((signs>>1)&1));}
 Gia_ManAppendCo(g,z^((signs>>2)&1));Gia_ManAppendCo(g,m^1);
 if(layout&2){Vec_IntReverseOrder(g->vCis);Vec_IntReverseOrder(g->vCos);Gia_ManForEachCi(g,o,i)Gia_ObjSetCioId(o,i);Gia_ManForEachCo(g,o,i)Gia_ObjSetCioId(o,i);}
 return g;
}
static Vec_Int_t *selection(int mask){Vec_Int_t *v=Vec_IntAlloc(4);int i;for(i=0;i<4;i++)if((mask>>i)&1)Vec_IntPush(v,i);return v;}
static void list(char *label,Vec_Int_t *v){int entry,i;printf("%s",label);Vec_IntForEachEntry(v,entry,i)printf(" %d",entry);puts("");}
static void print_graph(Gia_Man_t *g){
 Gia_Obj_t *o;int i,j,mask,bits,*values=ABC_CALLOC(int,g->nObjs);
 printf("graph %d %d %d nodes",g->nObjs,Gia_ManCiNum(g),Gia_ManCoNum(g));
 Gia_ManForEachObj(g,o,i){
  if(Gia_ObjIsConst0(o))printf(" 0");else if(Gia_ObjIsCi(o))printf(" 1");else if(Gia_ObjIsCo(o))printf(" 6:%d",Gia_ObjFaninLit0p(g,o));else if(Gia_ObjIsBuf(o))printf(" 5:%d",Gia_ObjFaninLit0p(g,o));else printf(" 2:%d:%d",Gia_ObjFaninLit0p(g,o),Gia_ObjFaninLit1p(g,o));
 }
 puts("");list("cis",g->vCis);list("cos",g->vCos);printf("truth");
 for(mask=0;mask<(1<<Gia_ManCiNum(g));mask++){
  Gia_ManForEachCi(g,o,j)values[Gia_ObjId(g,o)]=(mask>>j)&1;
  Gia_ManForEachAnd(g,o,i)values[i]=(values[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o))&(values[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o));
  bits=0;Gia_ManForEachCo(g,o,j)bits|=(values[Gia_ObjFaninId0p(g,o)]^Gia_ObjFaninC0(o))<<j;printf(" %d",bits);
 }
 puts("");ABC_FREE(values);
}
static void print_source(Gia_Man_t *g){Gia_Obj_t *o;int i;printf("source");Gia_ManForEachObj(g,o,i)printf(" %u:%d:%d:%d",o->Value,o->fMark0,o->fMark1,Gia_ObjIsTravIdCurrentId(g,i));puts("");}
static void support(Gia_Man_t *g,int id,Vec_Int_t *leaves){Gia_Obj_t *o=Gia_ManObj(g,id);if(Gia_ObjIsCi(o)){Vec_IntPushUnique(leaves,id);return;}if(Gia_ObjIsAnd(o)){support(g,Gia_ObjFaninId0(o,id),leaves);support(g,Gia_ObjFaninId1(o,id),leaves);}}
static void mapping(Gia_Man_t *g,int mode){int i,k,leaf;Gia_Obj_t *o;Vec_Int_t *leaves;if(!mode)return;g->vMapping=Vec_IntStart(g->nObjs);
 Gia_ManForEachAnd(g,o,i){Vec_IntWriteEntry(g->vMapping,i,Vec_IntSize(g->vMapping));Vec_IntPush(g->vMapping,2);Vec_IntPush(g->vMapping,Gia_ObjFaninId0(o,i));Vec_IntPush(g->vMapping,Gia_ObjFaninId1(o,i));Vec_IntPush(g->vMapping,i);}
 if(mode==2){Gia_ManForEachAnd(g,o,i){leaves=Vec_IntAlloc(4);support(g,i,leaves);Vec_IntSort(leaves,0);Vec_IntWriteEntry(g->vMapping,i,Vec_IntSize(g->vMapping));Vec_IntPush(g->vMapping,Vec_IntSize(leaves));Vec_IntForEachEntry(leaves,leaf,k)Vec_IntPush(g->vMapping,leaf);Vec_IntPush(g->vMapping,i);Vec_IntFree(leaves);}}
}
int main(void){
 int signs,layout,mask,po,i,mode,q,choices[14][2]={{-1,0},{0,0},{1,0},{2,0},{3,0},{4,0},{100,0},{1,-10},{1,1},{1,25},{1,50},{1,75},{1,100},{1,150}};
 Gia_Obj_t *o;Gia_Man_t *g,*w,*r;Vec_Int_t *selected,*inputs,*outputs,*gates,*latest;
 for(signs=0;signs<8;signs++)for(layout=0;layout<4;layout++)for(mask=0;mask<16;mask++)for(po=0;po<2;po++){
  printf("case %d %d %d %d\n",signs,layout,mask,po);g=graph(signs,layout);Gia_ManFillValue(g);Gia_ManForEachObj(g,o,i){o->fMark0=1;o->fMark1=i&1;}
  selected=selection(mask);Gia_ManPrepareWin(g,selected,&inputs,&outputs,&gates,po);list("inputs",inputs);list("boundary",outputs);list("gates",gates);Vec_IntFree(inputs);Vec_IntFree(outputs);Vec_IntFree(gates);
  w=Gia_ManExtractWin(g,selected,po);print_graph(w);print_source(g);
  if(!po){r=Gia_ManInsertWin(g,selected,w);print_graph(r);print_source(g);Gia_ManStop(r);}
  Gia_ManStop(w);Gia_ManStop(g);Vec_IntFree(selected);
 }
 for(layout=0;layout<4;layout++)for(mode=0;mode<3;mode++)for(q=0;q<14;q++){
  g=graph(5,layout);mapping(g,mode);printf("latest %d %d %d %d\n",layout,mode,choices[q][0],choices[q][1]);latest=Gia_ManFindLatest(g,choices[q][0],choices[q][1]);list("outputs",latest);Vec_IntFree(latest);Gia_ManStop(g);
 }
 return 0;
}
