#include "map/if/ifDec16.c"
int main(void) {
 char mode;int n,leaf,root,start,stop;unsigned long long x;word t[1024];
 while(scanf(" %c%d%d%d%d%d",&mode,&n,&leaf,&root,&start,&stop)==6){
  for(int i=0;i<If_CluWordNum(n);i++){if(scanf("%llx",&x)!=1)return 2;t[i]=x;}
  if(mode=='D'){
   If_Grp_t g,r={0};word f0=0,f1=0;
   g=If_CluCheck(NULL,t,n,start,stop,leaf,root,&r,&f0,&f1,NULL,0);
   printf("%d %d",g.nVars,g.nMyu);for(int i=0;i<leaf;i++)printf(" %d",g.pVars[i]);
   printf(" %d",r.nVars);for(int i=0;i<r.nVars;i++)printf(" %d",r.pVars[i]);
   printf(" %llx %llx\n",(unsigned long long)f0,(unsigned long long)f1);
  }else if(mode=='S'){
   int perm[16]={0};unsigned phase=If_CluSemiCanonicize(t,n,perm);printf("%u",phase);
   for(int i=0;i<n;i++)printf(" %d",perm[i]);for(int i=0;i<If_CluWordNum(n);i++)printf(" %llx",(unsigned long long)t[i]);puts("");
  }else return 3;
 }
 return 0;
}
