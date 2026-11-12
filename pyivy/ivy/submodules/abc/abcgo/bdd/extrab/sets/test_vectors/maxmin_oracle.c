/* Original MaxMin algorithms; fused operations receive their required
   maximal/minimal antichains. Ordinary products receive general families. */
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include "bdd/extrab/extraBdd.h"
#include <stdint.h>
static uint64_t hash=UINT64_C(14695981039346656037);
static void byte(unsigned b){hash^=b;hash*=UINT64_C(1099511628211);}
static DdNode *from_bits(DdManager *dd,unsigned bits){int set,v;DdNode *f=dd->zero,*cube,*temp;Cudd_Ref(f);for(set=0;set<8;set++){if(!(bits&(1u<<set)))continue;cube=dd->one;Cudd_Ref(cube);for(v=0;v<3;v++)if(set&(1<<v)){temp=Cudd_zddChange(dd,cube,v);Cudd_Ref(temp);Cudd_RecursiveDerefZdd(dd,cube);cube=temp;}temp=Cudd_zddUnion(dd,f,cube);Cudd_Ref(temp);Cudd_RecursiveDerefZdd(dd,f);Cudd_RecursiveDerefZdd(dd,cube);f=temp;}Cudd_Deref(f);return f;}
static unsigned family_bits(DdManager *dd,DdNode *f,unsigned prefix){if(f==dd->zero)return 0;if(f==dd->one)return 1u<<prefix;return family_bits(dd,cuddE(f),prefix)|family_bits(dd,cuddT(f),prefix|(1u<<f->index));}
static void collect(DdManager *dd,DdNode *f){Cudd_Ref(f);byte(family_bits(dd,f,0));Cudd_RecursiveDerefZdd(dd,f);}
int main(void){DdManager *dd=Cudd_Init(0,3,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);DdNode *families[256],*maximal[256],*minimal[256];int x,y;for(x=0;x<256;x++){families[x]=from_bits(dd,x);Cudd_Ref(families[x]);maximal[x]=Extra_zddMaximal(dd,families[x]);Cudd_Ref(maximal[x]);minimal[x]=Extra_zddMinimal(dd,families[x]);Cudd_Ref(minimal[x]);}for(x=0;x<256;x++){collect(dd,minimal[x]);for(y=0;y<256;y++){collect(dd,Extra_zddDotProduct(dd,families[x],families[y]));collect(dd,Extra_zddCrossProduct(dd,families[x],families[y]));collect(dd,Extra_zddMaxUnion(dd,maximal[x],maximal[y]));collect(dd,Extra_zddMinUnion(dd,minimal[x],minimal[y]));collect(dd,Extra_zddMaxDotProduct(dd,maximal[x],maximal[y]));}}printf("%016llx\n",(unsigned long long)hash);Cudd_Quit(dd);return 0;}
