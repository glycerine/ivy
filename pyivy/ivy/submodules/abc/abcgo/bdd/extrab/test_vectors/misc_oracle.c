#define _GNU_SOURCE
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include "bdd/extrab/extraBdd.h"
#include <stdint.h>
#include <unistd.h>
static uint64_t transform=UINT64_C(14695981039346656037),counts=UINT64_C(14695981039346656037),pla=UINT64_C(14695981039346656037);
static void number(uint64_t *hash,uint64_t v){int i;for(i=0;i<8;i++){*hash^=(v>>(8*i))&255;*hash*=UINT64_C(1099511628211);}}
static DdNode *truth(DdManager *dd,unsigned table,int level){DdNode *t,*e,*r;unsigned width,mask;if(level==3)return(table&1)?dd->one:Cudd_Not(dd->one);width=1u<<(2-level);mask=(1u<<width)-1;t=truth(dd,table>>width,level+1);Cudd_Ref(t);e=truth(dd,table&mask,level+1);Cudd_Ref(e);r=Cudd_bddIte(dd,dd->vars[level],t,e);Cudd_Ref(r);Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;}
static uint64_t bits(DdManager *dd,DdNode *f){uint64_t result=0;int a,v,input[6];for(a=0;a<64;a++){for(v=0;v<6;v++)input[v]=(a>>(5-v))&1;if(Cudd_Eval(dd,f,input)==dd->one)result|=UINT64_C(1)<<a;}return result;}
static void collect(DdManager *dd,DdNode *f){Cudd_Ref(f);number(&transform,bits(dd,f));Cudd_RecursiveDeref(dd,f);}
int main(void){DdManager *dd=Cudd_Init(6,12,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);unsigned table;char path[]="/tmp/extrab-pla-oracle.XXXXXX";int fd=mkstemp(path);close(fd);Cudd_zddVarsFromBddVars(dd,2);for(table=0;table<256;table++){DdNode *f=truth(dd,table,0),*cube,*temp,*cover;int mask,v,c;FILE *file;Cudd_Ref(f);collect(dd,Extra_bddGetOneCube(dd,f));collect(dd,Extra_bddMove(dd,f,2));collect(dd,Extra_bddRemapUp(dd,f));for(mask=0;mask<8;mask++){cube=dd->one;Cudd_Ref(cube);for(v=0;v<3;v++)if(mask&(1<<v)){temp=Cudd_bddAnd(dd,cube,dd->vars[v]);Cudd_Ref(temp);Cudd_RecursiveDeref(dd,cube);cube=temp;}collect(dd,Extra_bddChangePolarity(dd,f,cube));Cudd_RecursiveDeref(dd,cube);}cover=Extra_zddPrimes(dd,f);Cudd_Ref(cover);Extra_zddDumpPla(dd,cover,3,path);file=fopen(path,"rb");while((c=fgetc(file))!=EOF){pla^=(unsigned char)c;pla*=UINT64_C(1099511628211);}fclose(file);Cudd_RecursiveDerefZdd(dd,cover);Cudd_RecursiveDeref(dd,f);}unlink(path);Cudd_Quit(dd);
for(table=0;table<256;table++){int mode;for(mode=-1;mode<=1;mode++){DdManager *manager=Cudd_Init(3,0,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);DdNode *f=truth(manager,table,0);int guide=9;Cudd_Ref(f);number(&counts,Extra_bddCountCubes(manager,&f,1,mode,100,&guide));number(&counts,guide);Cudd_Quit(manager);}}
printf("%016llx %016llx %016llx\n",(unsigned long long)transform,(unsigned long long)counts,(unsigned long long)pla);return 0;}
