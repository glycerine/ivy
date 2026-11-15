#include "aig/gia/gia.h"
static abctime review_clock(void){return 0;}
#define Abc_Clock review_clock
#include "aig/gia/giaEra.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static Gia_Man_t *graph(int ni,int nr,int seed){
 Gia_Man_t *g=Gia_ManStart(256);int inputs[6],regs[65],next[65],i,a,b;
 for(i=0;i<ni;++i)inputs[i]=Gia_ManAppendCi(g);
 for(i=0;i<nr;++i)regs[i]=Gia_ManAppendCi(g);
 Gia_ManHashAlloc(g);a=ni?inputs[0]:1;b=ni?inputs[ni-1]:0;
 for(i=0;i<nr;++i){next[i]=regs[i];if(i>=3)continue;switch((seed+i)%4){case 0:next[i]=regs[i]^1;break;case 1:next[i]=a;break;case 2:next[i]=Gia_ManHashAnd(g,regs[i]^1,b^(seed&1));break;case 3:next[i]=regs[(i+1)%nr];break;}}
 Gia_ManAppendCo(g,regs[0]);Gia_ManAppendCo(g,Gia_ManHashAnd(g,a,regs[nr>1?1:0]^1)^(seed&1));for(i=0;i<nr;++i)Gia_ManAppendCo(g,next[i]);Gia_ManHashStop(g);Gia_ManSetRegNum(g,nr);return g;
}

/* Exercise private state/trace tables, with small buckets to force resizing. */
static void inspect(Gia_Man_t *g, int limit, int miter) {
 Gia_ManEra_t *p=Gia_ManEraCreate(g); Gia_ObjEra_t *s;
 int result=1,i,j; unsigned *place; int found;
 ABC_FREE(p->pBins);p->nBins=3;p->pBins=ABC_CALLOC(unsigned,p->nBins);
 s=Gia_ManEraCreateState(p);s->Cond=s->iPrev=s->iNext=0;memset(s->pData,0,sizeof(unsigned)*p->nWordsDat);p->pBins[Gia_ManEraStateHash(s->pData,p->nWordsDat,p->nBins)]=s->Num;
 while(p->iCurState<Vec_PtrSize(p->vStates)-1){
  if(Vec_PtrSize(p->vStates)>=limit){result=-1;break;}
  s=Gia_ManEraState(p,++p->iCurState);
  if(p->iCurState>1 && s->iPrev==0)continue;
  Gia_ManInsertState(p,s);Gia_ManPerformOneIter(p);
  if(Gia_ManAnalyzeResult(p,s,miter,1)&&miter){result=0;break;}
 }
 printf("manager %d %d %d %d %d %d %d\n",result,p->iCurState,Vec_PtrSize(p->vStates),p->nBins,p->nWordsSim,p->pStateNew?p->pStateNew->Num:0,Gia_ManCountDepth(p));
 for(i=1;i<Vec_PtrSize(p->vStates);++i){
  s=Gia_ManEraState(p,i);printf("state %d %d %d",s->Num,s==p->pStateNew?0:s->Cond,s->iPrev);
  for(j=0;j<p->nWordsDat;++j)printf(" %u",s->pData[j]);
  place=Gia_ManEraHashFind(p,s,&found);
  printf(" found%d\n",place?-1:found);
 }
 printf("trace");if(p->vBugTrace)for(i=0;i<Vec_IntSize(p->vBugTrace);++i)printf(" %d",Vec_IntEntry(p->vBugTrace,i));puts("");
 printf("transitions");for(i=0;i<Vec_IntSize(p->vStgDump);++i)printf(" %d",Vec_IntEntry(p->vStgDump,i));puts("");
 Gia_ManEraFree(p);
}

static int dump_cases(void) {
 int ni,no,toggle,i,ro;Gia_Man_t *g;FILE *f;int c;
 for(ni=0;ni<2;++ni)for(no=0;no<2;++no)for(toggle=0;toggle<2;++toggle){
  g=Gia_ManStart(8);if(ni)Gia_ManAppendCi(g);ro=Gia_ManAppendCi(g);
  if(no)Gia_ManAppendCo(g,ro);Gia_ManAppendCo(g,ro^toggle);Gia_ManSetRegNum(g,1);
  printf("dump %d %d %d\n",ni,no,toggle);
  Gia_ManCollectReachable(g,100,0,1,0);
  puts("stg begin");f=fopen("test.stg","r");if(!f)return 2;
  while((c=fgetc(f))!=EOF)putchar(c);fclose(f);puts("stg end");Gia_ManStop(g);
 }
 return 0;
}


static int verbose_case(void) {
 Gia_Man_t *g=Gia_ManStart(256);int r[13],next[13],carry=1,i;
 for(i=0;i<13;++i)r[i]=Gia_ManAppendCi(g);
 Gia_ManHashAlloc(g);
 for(i=0;i<13;++i){next[i]=Gia_ManHashXor(g,r[i],carry);carry=Gia_ManHashAnd(g,carry,r[i]);}
 Gia_ManAppendCo(g,r[12]);for(i=0;i<13;++i)Gia_ManAppendCo(g,next[i]);
 Gia_ManHashStop(g);Gia_ManSetRegNum(g,13);
 Gia_ManCollectReachable(g,20000,0,0,1);Gia_ManStop(g);return 0;
}

int main(int argc,char **argv){int pis[]={0,1,3,5,6},rs[]={1,3,33,65},caps[]={2,4,100};int ip,ir,seed,limit,miter;
 if(argc>1)return !strcmp(argv[1],"--verbose")?verbose_case():dump_cases();
 for(ip=0;ip<5;++ip)for(ir=0;ir<4;++ir)for(seed=0;seed<4;++seed)for(limit=0;limit<3;++limit)for(miter=0;miter<2;++miter){
  Gia_Man_t *g=graph(pis[ip],rs[ir],seed);int result;
  printf("case %d %d %d %d %d\n",pis[ip],rs[ir],seed,caps[limit],miter);
  result=Gia_ManCollectReachable(g,caps[limit],miter,0,0);printf("result %d\n",result);
  inspect(g,caps[limit],miter);Gia_ManStop(g);
 }return 0;}
