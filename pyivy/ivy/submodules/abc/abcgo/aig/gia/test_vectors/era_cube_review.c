#include "aig/gia/gia.h"
static abctime review_clock(void){return 0;}
#define Abc_Clock review_clock
#include "aig/gia/giaEra2.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static Gia_Man_t *graph(int ni,int nr,int seed){
 Gia_Man_t *g=Gia_ManStart(256);int inputs[6],regs[65],next[65],i,a,b;
 for(i=0;i<ni;++i)inputs[i]=Gia_ManAppendCi(g);
 for(i=0;i<nr;++i)regs[i]=Gia_ManAppendCi(g);
 Gia_ManHashAlloc(g);a=ni?inputs[0]:1;b=ni?inputs[ni-1]:0;
 for(i=0;i<nr;++i){next[i]=regs[i];if(i>=3)continue;switch((seed+i)%4){case 0:next[i]=regs[i]^1;break;case 1:next[i]=a;break;case 2:next[i]=Gia_ManHashAnd(g,regs[i]^1,b^(seed&1));break;case 3:next[i]=regs[(i+1)%nr];break;}}
 Gia_ManAppendCo(g,regs[0]);Gia_ManAppendCo(g,Gia_ManHashAnd(g,a,regs[nr>1?1:0]^1)^(seed&1));for(i=0;i<nr;++i)Gia_ManAppendCo(g,next[i]);Gia_ManHashStop(g);Gia_ManSetRegNum(g,nr);return g;
}
static int relations(void){
 int positions[3][3]={{0,1,2},{0,16,32},{15,16,64}},mode,i,j,k,x,y;
 Gia_StaAre_t *a=calloc(1,sizeof(Gia_StaAre_t)+5*sizeof(unsigned));
 Gia_StaAre_t *b=calloc(1,sizeof(Gia_StaAre_t)+5*sizeof(unsigned));
 for(mode=0;mode<3;++mode)for(i=0;i<27;++i)for(j=0;j<27;++j){
  memset(a->pData,0,5*sizeof(unsigned));memset(b->pData,0,5*sizeof(unsigned));
  for(k=0,x=i,y=j;k<3;++k,x/=3,y/=3){
   if(x%3==1)Gia_StaSetValue0(a,positions[mode][k]);if(x%3==2)Gia_StaSetValue1(a,positions[mode][k]);
   if(y%3==1)Gia_StaSetValue0(b,positions[mode][k]);if(y%3==2)Gia_StaSetValue1(b,positions[mode][k]);
  }
  printf("%d %d %d: %d %d %d %d %d %d\n",mode,i,j,Gia_StaAreEqual(a,b,5),Gia_StaAreDisjoint(a,b,5),Gia_StaAreContain(a,b,5),Gia_StaAreDashNum(a,b,5),Gia_StaAreSharpVar(a,b,5),Gia_StaAreDisjointVar(a,b,5));
 }
 free(a);free(b);return 0;
}


static int reports(void) {
 Gia_Man_t *g=graph(1,3,0);Gia_ManAre_t *p=Gia_ManAreCreate(g);
 Gia_StaAre_t *a=Gia_ManAreCreateStaInit(p),*b;
 Gia_ManAreCubeAddToList(p,&p->Root,a);
 b=Gia_ManAreCreateSta(p);Gia_StaSetValue1(b,0);b->iPrev=Gia_Int2Ptr(1);
 Gia_ManAreCubeAddToList(p,&p->Root,b);p->iStaCur=2;
 puts("private reports");Gia_ManArePrintReport(p,0,0);Gia_ManArePrintReport(p,0,1);
 puts("private cubes");Gia_ManArePrintCube(p,a);Gia_ManArePrintCube(p,b);
 Gia_StaSetUnused(a);printf("print return %d\n",Gia_ManArePrintUsed(p));
 printf("counts %d %d %d\n",Gia_ManAreDepth(p,2),Gia_ManAreListCountUsed(p),Gia_ManCountMinterms(p));
 Gia_ManAreFree(p);return 0;
}

static int driver_reports(void) {
 int seed,verbose;Gia_Man_t *g;
 for(seed=0;seed<4;++seed)for(verbose=0;verbose<2;++verbose){
  g=graph(1,3,seed);g->pName=Abc_UtilStrsav("cube_review");
  printf("driver %d %d\n",seed,verbose);Gia_ManArePerform(g,100,0,verbose);Gia_ManStop(g);
 }
 return 0;
}
int main(int argc,char **argv){setvbuf(stdout,NULL,_IONBF,0);if(argc>1){if(!strcmp(argv[1],"--driver-reports"))return driver_reports();return !strcmp(argv[1],"--reports")?reports():relations();}int pis[]={0,1,3,5,6},rs[]={1,3,33,65},caps[]={100};int ip,ir,seed,limit,miter;
 for(ip=0;ip<5;++ip)for(ir=0;ir<4;++ir)for(seed=0;seed<4;++seed)for(limit=0;limit<1;++limit)for(miter=0;miter<2;++miter){
  Gia_Man_t *g=graph(pis[ip],rs[ir],seed);int dump=0,result;
  printf("case %d %d %d %d %d %d\n",pis[ip],rs[ir],seed,caps[limit],miter,dump);result=Gia_ManArePerform(g,caps[limit],miter,0);printf("result %d\n",result);printf("cex");if(g->pCexSeq){int bi;printf(" %d %d",g->pCexSeq->iPo,g->pCexSeq->iFrame);for(bi=0;bi<g->pCexSeq->nBits;++bi)printf("%s%d",bi?"":" ",Abc_InfoHasBit(g->pCexSeq->pData,bi));}puts("");
  if(dump){int c;FILE *f=fopen("test.stg","r");puts("stg begin");while((c=fgetc(f))!=EOF)putchar(c);puts("stg end");fclose(f);}
  Gia_ManStop(g);
 }return 0;}
