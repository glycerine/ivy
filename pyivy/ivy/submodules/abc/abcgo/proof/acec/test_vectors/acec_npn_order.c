#include <stdio.h>
#include "aig/gia/gia.h"
extern unsigned Extra_TruthCanonNPN2(unsigned,int,Vec_Int_t*);
int main(int argc, char **argv){Vec_Int_t*v=Vec_IntAlloc(128);if(argc>1){Vec_IntPush(v,86);Vec_IntPush(v,222);}unsigned min=Extra_TruthCanonNPN2(0x56,3,v);printf("min=%u count=%d\n",min,Vec_IntSize(v));for(int i=0;i<Vec_IntSize(v);i++)printf("%u ",(unsigned)Vec_IntEntry(v,i));puts("");return 0;}
