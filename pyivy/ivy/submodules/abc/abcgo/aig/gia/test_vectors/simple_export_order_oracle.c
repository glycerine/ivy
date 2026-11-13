#include "aig/gia/gia.h"
#include "aig/gia/giaAig.h"
#include "aig/aig/aig.h"
#include <stdio.h>
int main(int argc,char **argv){Gia_Man_t *g=Gia_ManStart(12);int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),n=Gia_ManAppendAnd(g,a,b),late;Aig_Man_t *out;
 if(argc>1) {late=Gia_ManAppendAnd(g,a,b);Gia_ManAppendCo(g,late);} else {late=Gia_ManAppendCi(g);Gia_ManAppendCo(g,n);Gia_ManAppendCo(g,late);}
 out=Gia_ManToAigSimple(g);printf("node=%u late=%u ci2=%d\n",Gia_ManObj(g,n>>1)->Value,Gia_ManObj(g,late>>1)->Value,Aig_ObjId(Aig_ManCi(out,2)));Aig_ManStop(out);Gia_ManStop(g);return 0;}
