/* Genuine original extraBddThresh.c bounded-search fixture. */
#include "misc/util/util_hack.h"
#include "misc/util/utilTruth.h"
#include <stdint.h>
extern int Extra_ThreshCheck(word *,int,int *);
extern int Extra_ThreshComputeChow(word *,int,int *);
static uint64_t hash=UINT64_C(14695981039346656037);
static void number(uint64_t value){int i;for(i=0;i<8;i++){hash^=(value>>(8*i))&255;hash*=UINT64_C(1099511628211);}}
int main(void){unsigned n,table;for(n=3;n<=4;n++){unsigned total=1u<<(1u<<n);for(table=0;table<total;table++){word truth=0;int weights[16]={0},chow[16]={0},i;for(i=0;i<64;i++)if((table>>(i&((1u<<n)-1)))&1)truth|=(word)1<<i;number(Extra_ThreshComputeChow(&truth,n,chow));for(i=0;i<(int)n;i++)number(chow[i]);number(Extra_ThreshCheck(&truth,n,weights));number(truth);for(i=0;i<(int)n;i++)number(weights[i]);}}printf("%016llx\n",(unsigned long long)hash);return 0;}
