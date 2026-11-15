#include "aig/gia/giaFrames.c"
static Gia_Man_t *graph(void){
 Gia_Man_t *g=Gia_ManStart(64);int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),r0=Gia_ManAppendCi(g),r1=Gia_ManAppendCi(g),r2=Gia_ManAppendCi(g);
 int x=Gia_ManAppendAnd(g,a,r0),y=Gia_ManAppendAnd(g,b,r1),z=Gia_ManAppendAnd(g,x,r2^1),t=Gia_ManAppendAnd(g,y,r0),u=Gia_ManAppendAnd(g,a,r1^1),v=Gia_ManAppendAnd(g,u,b);
 Gia_ManAppendCo(g,x);Gia_ManAppendCo(g,z);Gia_ManAppendCo(g,y);Gia_ManAppendCo(g,t);Gia_ManAppendCo(g,v);Gia_ManSetRegNum(g,3);return g;
}
static void dump(Gia_Man_t *g,const char *tag){int i;Gia_Obj_t *o;printf("%s ci%d co%d regs%d objs%d\n",tag,Gia_ManCiNum(g),Gia_ManCoNum(g),Gia_ManRegNum(g),Gia_ManObjNum(g));Gia_ManForEachObj(g,o,i)printf("%d %u %u %u %u %u %u\n",i,o->iDiff0,o->iDiff1,o->fCompl0,o->fCompl1,o->fTerm,o->Value);}
static int literal_value(Gia_Man_t *g,int lit,int bits){int values[256]={0},i;Gia_Obj_t *o;assert(g->nObjs<256);Gia_ManForEachCi(g,o,i)values[Gia_ObjId(g,o)]=(bits>>i)&1;Gia_ManForEachAnd(g,o,i)values[i]=(values[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o))&(values[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o));return values[Abc_Lit2Var(lit)]^Abc_LitIsCompl(lit);}
static void incremental(Gia_Man_t *g){Gia_ParFra_t p;Gia_Man_t *q=NULL;void *raw;int f,bits,lit;Gia_ManFraSetDefaultParams(&p);p.nFrames=3;p.fSaveLastLit=1;raw=Gia_ManUnrollStart(g,&p);for(f=1;f<=3;++f){q=(Gia_Man_t *)Gia_ManUnrollAdd(raw,f);lit=Gia_ManUnrollLastLit(raw);printf("step%d ci%d co%d objs%d lit%d\n",f,Gia_ManCiNum(q),Gia_ManCoNum(q),Gia_ManObjNum(q),lit);for(bits=0;bits<(1<<Gia_ManCiNum(q));++bits)printf("%d",literal_value(q,lit,bits));puts("");}Gia_ManUnrollStop(raw);Gia_ManStop(q);}
int main(int argc,char **argv){Gia_Man_t *g=graph(),*q;Gia_ParFra_t p;Vec_Int_t *limits,*row;Vec_Ptr_t *abs;int i,j,id;
 if(argc>1 && !strcmp(argv[1],"--incremental")){incremental(g);}
 else if(argc>1 && !strcmp(argv[1],"--frames2")){Gia_ManFraSetDefaultParams(&p);p.nFrames=3;q=Gia_ManFrames2(g,&p);dump(q,"frames2");Gia_ManStop(q);}
 else if(argc>1 && !strcmp(argv[1],"--abs")){abs=Gia_ManUnrollAbs(g,3);Vec_PtrForEachEntry(Vec_Int_t *,abs,row,i){printf("abs %d",i);Vec_IntForEachEntry(row,id,j)printf(" %d",id);puts("");}Vec_VecFree((Vec_Vec_t *)abs);}
 else {limits=Vec_IntAlloc(0);q=Gia_ManUnrollDup(g,limits);printf("limits");Vec_IntForEachEntry(limits,id,i)printf(" %d",id);puts("");dump(q,"ordered");Gia_ManStop(q);Vec_IntFree(limits);}
 Gia_ManStop(g);return 0;}
