#include "aig/gia/giaCut.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static void make_cut(Gia_Cut_t *c,int mask) {
 int i,ids[]={0,1,63,64,65,127};memset(c,0,sizeof(*c));
 for(i=0;i<6;++i)if(mask&(1<<i))c->pLeaves[c->nLeaves++]=ids[i];
 c->Sign=Gia_CutGetSign(c);
}
int main(void) {
 Gia_Cut_t a,b,out,storage[GIA_MAX_CUTNUM],*cuts[GIA_MAX_CUTNUM];
 int ma,mb,limit,method,status,i,j,step,n,accepted,mask,limits[]={2,5,50};unsigned random;
 for(ma=0;ma<64;++ma)for(mb=0;mb<64;++mb){
  make_cut(&a,ma);make_cut(&b,mb);
  printf("check %d %d %d %d %d\n",ma,mb,Gia_CutCheck(&a,&b),a.nLeaves>=b.nLeaves?Gia_CutSetCutIsContainedOrder(&a,&b):-1,Gia_CutCountBits(a.Sign|b.Sign));
  for(limit=0;limit<=8;++limit)for(method=0;method<2;++method){
   memset(&out,0,sizeof(out));out.nLeaves=GIA_CUT_NO_LEAF;out.iFunc=123;
   status=method?Gia_CutMergeOrder2(&a,&b,&out,limit):Gia_CutMergeOrder(&a,&b,&out,limit);
   printf("merge %d %d %d %d %d",ma,mb,limit,method,status);
   if(status){printf(" %u %016llx %d",out.nLeaves,(unsigned long long)out.Sign,out.iFunc);for(i=0;i<(int)out.nLeaves;++i)printf(" %d",out.pLeaves[i]);}printf("\n");
  }
 }
 for(j=0;j<3;++j){
  memset(storage,0,sizeof(storage));for(i=0;i<GIA_MAX_CUTNUM;++i)cuts[i]=&storage[i];n=0;random=limits[j]+409;
  for(step=0;step<64;++step){
   random=random*1664525u+1013904223u;mask=step==63?0:1+random%63;
   make_cut(cuts[n],mask);cuts[n]->nTreeLeaves=mask?((random>>19)&3):0;if(step%7==0)cuts[n]->nTreeLeaves+=1<<20;
   accepted=!Gia_CutSetLastCutIsContained(cuts,n);if(accepted)n=Gia_CutSetAddCut(cuts,n,limits[j]);
   printf("set %d %d %d %d",limits[j],step,accepted,n);
   for(i=0;i<n;++i)printf(" %ld:%u:%u:%016llx",(long)(cuts[i]-storage),cuts[i]->nLeaves,cuts[i]->nTreeLeaves,(unsigned long long)cuts[i]->Sign);printf("\n");
  }
 }
 return 0;
}
