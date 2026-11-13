#include "proof/fra/fra.h"
#include <stdio.h>
static int ids[1024],n;
static void cb(void*v,void*ignored){Fra_Man_t*p=v;Aig_Obj_t*obj;int i;printf("failures=%d sat=%d",p->nSatFailsReal,p->nSatCallsSat);if(p->vTimeouts)Vec_PtrForEachEntry(Aig_Obj_t*,p->vTimeouts,obj,i)ids[n++]=obj->Id;printf(" ids=");for(i=0;i<n;i++)printf("%s%d",i?",":"",ids[i]);puts("");}
static Aig_Obj_t *parity(Aig_Man_t*p,Aig_Obj_t**a,int n){if(n==1)return a[0];int m=n/2;Aig_Obj_t*l=parity(p,a,m),*r=parity(p,a+m,n-m);return Aig_Exor(p,l,r);}
int main(void){Aig_Man_t*g=Aig_ManStart(256),*out;Aig_Obj_t*v[20],*left,*right,*rare;Fra_Par_t p;for(int i=0;i<20;i++)v[i]=Aig_ObjCreateCi(g);left=v[0];for(int i=1;i<8;i++)left=Aig_Exor(g,left,v[i]);right=parity(g,v,8);rare=v[0];for(int i=1;i<20;i++)rare=Aig_And(g,rare,Aig_NotCond(v[i],(i-1)%2==0));Aig_ObjCreateCo(g,left);Aig_ObjCreateCo(g,right);Aig_ObjCreateCo(g,rare);Fra_ParamsDefault(&p);p.nBTLimitNode=1;p.fConeBias=0;p.nSimWords=1;g->pImpFunc=cb;out=Fra_FraigPerform(g,&p);printf("rare=%d nodes=%d\n",rare->Id,Aig_ManNodeNum(out));for(int i=0;i<n;i++)printf("timeout=%d repr=%d\n",ids[i],g->pReprs[ids[i]]?g->pReprs[ids[i]]->Id:-1);return 0;}
