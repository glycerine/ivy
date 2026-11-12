#include "sat/glucose2/SimpSolver.h"
#include <stdio.h>
using namespace Gluco2;
int main(){unsigned state=8231;unsigned long long hash=1469598103934665603ULL;
#define NEXT (state=state*1664525+1013904223)
for(int trial=0;trial<1500;trial++){SimpSolver s;int n=1+NEXT%9;for(int i=0;i<n;i++)s.newVar();int count=NEXT%45;for(int c=0;c<count;c++){vec<Lit> ps;int len=1+NEXT%5;for(int k=0;k<len;k++){int v=NEXT%n;bool sign=NEXT>>31;ps.push(mkLit(v,sign));}s.addClause(ps);}vec<Lit> as;count=NEXT%4;for(int k=0;k<count;k++){int v=NEXT%n;bool sign=NEXT>>31;as.push(mkLit(v,sign));}hash^=s.solve(as)?1:0;hash*=1099511628211ULL;}printf("%llu\n",hash);}
