#include "proof/fraig/fraigInt.h"
#include <stdio.h>
int main(void) {
 Fraig_Man_t m={0}; Fraig_HashTable_t t={0}; Fraig_Node_t a={0},b={0},c={0};
 t.nBins=3;t.nEntries=3;t.pBins=calloc(3,sizeof(*t.pBins));m.pTableF0=&t;
 a.uHashD=6;b.uHashD=6;c.uHashD=0;a.pNextF=&b;b.pNextF=&c;t.pBins[0]=&a;
 for(int i=0;i<3;i++) {b.uHashD=6;int linked=Fraig_TableRehashF0(&m,1);int grouped=a.pNextD==&b||b.pNextD==&a;b.uHashD=5;Fraig_TableRehashF0(&m,0);printf("%d %d %d %d\n",linked,grouped,a.pNextD==0&&b.pNextD==0&&c.pNextD==0,t.nEntries);}
 free(t.pBins);return 0;
}
