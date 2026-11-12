/* Original Chow/inequality-search oracle. For monotone functions the complete
   prime cover is exactly the minimal true/maximal false boundary; the adapter
   enumerates this mathematical cover in Abc_ConvertZddToSop_rec order. */
#include "misc/util/util_hack.h"
#include "misc/util/utilTruth.h"
#include <stdint.h>
#include <string.h>
extern int Extra_ThreshComputeChow(word *,int,int *);
extern void Extra_ThreshSortByChowInverted(word *,int,int *);
extern int Extra_ThreshInitializeChow(int,int *);
extern int Extra_ThreshAssignWeights(word *,char *,char *,int,int *,int *,int,int);
static uint64_t hash=UINT64_C(14695981039346656037);
static void number(uint64_t v){int i;for(i=0;i<8;i++){hash^=(v>>(8*i))&255;hash*=UINT64_C(1099511628211);}}
static void primes(word truth,int n,int on,char *sop){int rank,v,a,prime,len=0;for(rank=0;rank<(1<<n);rank++){a=0;for(v=0;v<n;v++){int included=!(rank&(1<<(n-1-v)));if(included==on)a|=1<<v;}if(Abc_TtGetBit(&truth,a)!=on)continue;prime=1;for(v=0;v<n;v++){int bit=1<<v;if(on&&(a&bit)&&Abc_TtGetBit(&truth,a^bit))prime=0;if(!on&&!(a&bit)&&!Abc_TtGetBit(&truth,a|bit))prime=0;}if(!prime)continue;for(v=0;v<n;v++)sop[len++]=(on&&(a&(1<<v)))?'1':(!on&&!(a&(1<<v)))?'0':'-';sop[len++]=' ';sop[len++]='1';sop[len++]='\n';}sop[len]=0;}
int main(void){int n,table;for(n=3;n<=4;n++){int total=1<<(1<<n);for(table=1;table<total-1;table++){word truth=0;int weights[16]={0},chow[16]={0},i,threshold=0,groups;char on[4096],off[4096];for(i=0;i<64;i++)if((table>>(i&((1<<n)-1)))&1)truth|=(word)1<<i;if(Abc_TtIsUnate(&truth,n)){Abc_TtMakePosUnate(&truth,n);Extra_ThreshComputeChow(&truth,n,chow);Extra_ThreshSortByChowInverted(&truth,n,chow);groups=Extra_ThreshInitializeChow(n,chow);primes(truth,n,1,on);primes(truth,n,0,off);truth=~truth;threshold=Extra_ThreshAssignWeights(&truth,on,off,n,weights,chow,groups,1);}number(threshold);number(truth);for(i=0;i<n;i++)number(weights[i]);}}printf("%016llx\n",(unsigned long long)hash);return 0;}
