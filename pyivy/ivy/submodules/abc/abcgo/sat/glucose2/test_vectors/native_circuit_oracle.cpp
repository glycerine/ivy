#include "sat/glucose2/Solver.h"
#include <stdio.h>
using namespace Gluco2;
int main(){unsigned state=71;unsigned long long hash=1469598103934665603ULL;
#define NEXT (state=state*1664525+1013904223)
for(int trial=0;trial<500;trial++){Solver s;s.jftr=2;for(int i=0;i<9;i++)s.newVar();for(int v=3;v<9;v++){int a=NEXT%v,b=NEXT%v;while(b==a)b=NEXT%v;bool sa=NEXT>>31,sb=NEXT>>31;s.setVarFaninLits(v,mkLit(a,sa),mkLit(b,sb));}s.setNewRound();for(int v=0;v<9;v++)s.markCone(v);for(int round=0;round<10;round++){vec<Lit> as;int count=1+NEXT%5;for(int k=0;k<count;k++){int v=NEXT%9;bool sign=NEXT>>31;as.push(mkLit(v,sign));}hash^=s.solve(as)?1:0;hash*=1099511628211ULL;}}printf("%llu\n",hash);}
