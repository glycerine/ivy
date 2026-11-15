#include "aig/gia/gia.h"
static int repair_support;
static void review_collect_ands(Gia_Man_t *g,int *nodes,int n,Vec_Int_t *out,Vec_Int_t *leaves){
 if(repair_support)Gia_ManIncrementTravId(g);
 Gia_ManCollectAnds(g,nodes,n,out,leaves);
}
#define Gia_ManCollectAnds review_collect_ands
#include "aig/gia/giaFrames.c"
#undef Gia_ManCollectAnds
#include <sys/resource.h>
static Gia_Man_t *graph(void){
 Gia_Man_t *g=Gia_ManStart(64);int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),r0=Gia_ManAppendCi(g),r1=Gia_ManAppendCi(g),r2=Gia_ManAppendCi(g);
 int x=Gia_ManAppendAnd(g,a,r0),y=Gia_ManAppendAnd(g,b,r1),z=Gia_ManAppendAnd(g,x,r2^1),t=Gia_ManAppendAnd(g,y,r0),u=Gia_ManAppendAnd(g,a,r1^1),v=Gia_ManAppendAnd(g,u,b);
 Gia_ManAppendCo(g,x);Gia_ManAppendCo(g,z);Gia_ManAppendCo(g,y);Gia_ManAppendCo(g,t);Gia_ManAppendCo(g,v);Gia_ManSetRegNum(g,3);return g;
}
static void dump(Gia_Man_t *g,const char *tag){int i;Gia_Obj_t *o;printf("%s ci%d co%d regs%d objs%d\n",tag,Gia_ManCiNum(g),Gia_ManCoNum(g),Gia_ManRegNum(g),Gia_ManObjNum(g));Gia_ManForEachObj(g,o,i)printf("%d %u %u %u %u %u %u\n",i,o->iDiff0,o->iDiff1,o->fCompl0,o->fCompl1,o->fTerm,o->Value);}
static int literal_value(Gia_Man_t *g,int lit,int bits){int values[256]={0},i;Gia_Obj_t *o;assert(g->nObjs<256);Gia_ManForEachCi(g,o,i)values[Gia_ObjId(g,o)]=(bits>>i)&1;Gia_ManForEachAnd(g,o,i)values[i]=(values[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o))&(values[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o));return values[Abc_Lit2Var(lit)]^Abc_LitIsCompl(lit);}
static void incremental(Gia_Man_t *g,int save,int init){Gia_ParFra_t p;Gia_Man_t *q=NULL;void *raw;int f,bits,lit;Gia_ManFraSetDefaultParams(&p);p.nFrames=3;p.fSaveLastLit=save;p.fInit=init;raw=Gia_ManUnrollStart(g,&p);for(f=1;f<=3;++f){q=(Gia_Man_t *)Gia_ManUnrollAdd(raw,f);lit=Gia_ManUnrollLastLit(raw);printf("step%d ci%d co%d objs%d lit%d\n",f,Gia_ManCiNum(q),Gia_ManCoNum(q),Gia_ManObjNum(q),lit);for(bits=0;bits<(1<<Gia_ManCiNum(q));++bits)printf("%d",literal_value(q,lit,bits));puts("");}Gia_ManUnrollStop(raw);Gia_ManStop(q);}
static Gia_Man_t *live_init_graph(void){Gia_Man_t *g=graph();Gia_Obj_t *ri=Gia_ManRi(g,0);ri->iDiff0=Gia_ObjId(g,ri)-Gia_ObjId(g,Gia_ManPi(g,1));ri->fCompl0=0;return g;}
static void drivers(void){int frames,init,disable,or;Gia_ParFra_t p;Gia_Man_t *g,*q;repair_support=1;
 for(frames=1;frames<=3;++frames)for(init=0;init<2;++init)for(disable=0;disable<2;++disable)for(or=0;or<2;++or){g=live_init_graph();Gia_ManFraSetDefaultParams(&p);p.nFrames=frames;p.fInit=init;p.fDisableSt=disable;p.fOrPos=or;q=Gia_ManFrames(g,&p);printf("driver f%d init%d disable%d or%d\n",frames,init,disable,or);dump(q,"result");Gia_ManStop(q);Gia_ManStop(g);}}
static void init_defect(void){int mode,bits,co,packed;Gia_ParFra_t p;Gia_Man_t *g,*q;
 for(mode=0;mode<2;++mode){repair_support=mode;g=live_init_graph();Gia_ManFraSetDefaultParams(&p);p.nFrames=3;p.fInit=1;q=Gia_ManFrames(g,&p);for(bits=0;bits<64;++bits){packed=0;for(co=0;co<6;++co)packed|=literal_value(q,Gia_ObjFaninLit0p(q,Gia_ManCo(q,co)),bits)<<co;printf("init %d %d %d\n",mode,bits,packed);}Gia_ManStop(q);Gia_ManStop(g);}}
static void special_stale(void){Gia_Man_t *g=Gia_ManStart(8),*q;int r=Gia_ManAppendCi(g);struct rlimit limit={0,0};setrlimit(RLIMIT_CORE,&limit);Gia_ManAppendCo(g,1);Gia_ManAppendCo(g,r);Gia_ManSetRegNum(g,1);Gia_ManFillValue(g);q=Gia_ManFramesInitSpecial(g,1,0);Gia_ManStop(q);Gia_ManStop(g);}
static unsigned review_rng;
static unsigned next_graph_random(void){review_rng=review_rng*1664525u+1013904223u;return review_rng>>8;}
static Gia_Man_t *generated_graph(unsigned seed){Gia_Man_t *g=Gia_ManStart(128);int pool[32],n=1,i,ia,ib,a,b,lit;review_rng=seed;pool[0]=0;
 for(i=0;i<4;++i)pool[n++]=Gia_ManAppendCi(g);
 for(i=0;i<12;++i){ia=next_graph_random()%n;ib=next_graph_random()%n;if(ia==ib)ib=(ib+1)%n;a=pool[ia]^(next_graph_random()&1);b=pool[ib]^(next_graph_random()&1);pool[n++]=Gia_ManAppendAnd(g,a,b);}
 for(i=0;i<4;++i){ia=i<2?n-1-i:(int)(next_graph_random()%n);lit=pool[ia]^(next_graph_random()&1);Gia_ManAppendCo(g,lit);}Gia_ManSetRegNum(g,2);return g;}
static void generated_drivers(void){unsigned seed;int frames,init,disable,or;Gia_ParFra_t p;Gia_Man_t *g,*q;repair_support=1;
 for(seed=0;seed<32;++seed)for(frames=1;frames<=3;++frames)for(init=0;init<2;++init)for(disable=0;disable<2;++disable)for(or=0;or<2;++or){g=generated_graph(seed);Gia_ManFraSetDefaultParams(&p);p.nFrames=frames;p.fInit=init;p.fDisableSt=disable;p.fOrPos=or;q=Gia_ManFrames(g,&p);printf("generated %u f%d init%d disable%d or%d\n",seed,frames,init,disable,or);dump(q,"result");Gia_ManStop(q);Gia_ManStop(g);}}
static void print_vector(const char *tag,Vec_Int_t *v){int i,x;printf("%s",tag);Vec_IntForEachEntry(v,x,i)printf(" %d",x);puts("");}
static void history_snapshot(Gia_ManUnr_t *p){print_vector("limits",p->vLimit);print_vector("delays",p->vDegDiff);print_vector("first",p->vFirst);print_vector("history",p->vStore);}
static void generated_histories(void){unsigned seed;int save,step;Gia_ParFra_t pars;Gia_Man_t *g,*q;Gia_ManUnr_t *p;
 for(seed=0;seed<32;++seed)for(save=0;save<2;++save){g=generated_graph(seed);Gia_ManFraSetDefaultParams(&pars);pars.nFrames=3;pars.fSaveLastLit=save;p=(Gia_ManUnr_t *)Gia_ManUnrollStart(g,&pars);printf("history %u save%d step0\n",seed,save);history_snapshot(p);dump(p->pOrder,"ordered");
 for(step=1;step<=3;++step){q=(Gia_Man_t *)Gia_ManUnrollAdd(p,step);printf("history %u save%d step%d lit%d\n",seed,save,step,Gia_ManUnrollLastLit(p));history_snapshot(p);dump(q,"step");}q=p->pNew;Gia_ManUnrollStop(p);Gia_ManStop(q);Gia_ManStop(g);}}
static void generated_special(void){unsigned seed;int frames;Gia_Man_t *g,*q;for(seed=0;seed<32;++seed)for(frames=1;frames<=3;++frames){g=generated_graph(seed);q=Gia_ManFramesInitSpecial(g,frames,0);printf("special %u f%d\n",seed,frames);dump(q,"result");Gia_ManStop(q);Gia_ManStop(g);}}
int main(int argc,char **argv){Gia_Man_t *g=graph(),*q;Gia_ParFra_t p;Vec_Int_t *limits,*row;Vec_Ptr_t *abs;int i,j,id;
 if(argc>1 && !strcmp(argv[1],"--generated-special")){generated_special();}
 else if(argc>1 && !strcmp(argv[1],"--generated-drivers")){generated_drivers();}
 else if(argc>1 && !strcmp(argv[1],"--histories")){generated_histories();}
 else if(argc>1 && !strcmp(argv[1],"--special-stale")){special_stale();}
 else if(argc>1 && !strcmp(argv[1],"--drivers-safe")){drivers();}
 else if(argc>1 && !strcmp(argv[1],"--init-defect")){init_defect();}
 else if(argc>1 && !strcmp(argv[1],"--incremental")){incremental(g,1,0);}
 else if(argc>1 && !strcmp(argv[1],"--fixed")){incremental(g,0,0);}
 else if(argc>1 && !strcmp(argv[1],"--initialized-save")){struct rlimit limit={0,0};setrlimit(RLIMIT_CORE,&limit);incremental(g,1,1);}
 else if(argc>1 && !strcmp(argv[1],"--frames2")){Gia_ManFraSetDefaultParams(&p);p.nFrames=3;q=Gia_ManFrames2(g,&p);dump(q,"frames2");Gia_ManStop(q);}
 else if(argc>1 && !strcmp(argv[1],"--abs")){abs=Gia_ManUnrollAbs(g,3);Vec_PtrForEachEntry(Vec_Int_t *,abs,row,i){printf("abs %d",i);Vec_IntForEachEntry(row,id,j)printf(" %d",id);puts("");}Vec_VecFree((Vec_Vec_t *)abs);}
 else {limits=Vec_IntAlloc(0);q=Gia_ManUnrollDup(g,limits);printf("limits");Vec_IntForEachEntry(limits,id,i)printf(" %d",id);puts("");dump(q,"ordered");Gia_ManStop(q);Vec_IntFree(limits);}
 Gia_ManStop(g);return 0;}
