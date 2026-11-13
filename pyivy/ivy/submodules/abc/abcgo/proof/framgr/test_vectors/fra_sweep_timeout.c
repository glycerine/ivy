#include "proof/fra/fra.h"
#include <stdio.h>
static int ids[1024],n;
static void cb(void*v,void*ignored){Fra_Man_t*p=v;Aig_Obj_t*obj;int i;printf("callback failures=%d logged=%d\n",p->nSatFailsReal,p->vTimeouts?Vec_PtrSize(p->vTimeouts):0);if(p->vTimeouts)Vec_PtrForEachEntry(Aig_Obj_t*,p->vTimeouts,obj,i)ids[n++]=obj->Id;}
static Aig_Obj_t *parity(Aig_Man_t*p,Aig_Obj_t**a,int n){if(n==1)return a[0];int m=n/2;Aig_Obj_t *left=parity(p,a,m);Aig_Obj_t *right=parity(p,a+m,n-m);return Aig_Exor(p,left,right);}
int main(void){Aig_Man_t*g=Aig_ManStart(256),*out;Aig_Obj_t*v[8],*left;Fra_Par_t p;for(int i=0;i<8;i++)v[i]=Aig_ObjCreateCi(g);left=v[0];for(int i=1;i<8;i++)left=Aig_Exor(g,left,v[i]);Aig_Obj_t *right=parity(g,v,8);Aig_ObjCreateCo(g,left);Aig_ObjCreateCo(g,right);Fra_ParamsDefault(&p);p.nBTLimitNode=1;p.fConeBias=0;p.nSimWords=1;g->pImpFunc=cb;out=Fra_FraigPerform(g,&p);printf("done nodes=%d\n",Aig_ManNodeNum(out));for(int i=0;i<n;i++)printf("timeout=%d repr=%d\n",ids[i],g->pReprs[ids[i]]?g->pReprs[ids[i]]->Id:-1);return 0;}
