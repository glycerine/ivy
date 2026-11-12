#define _GNU_SOURCE
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include "bdd/extrab/extraBdd.h"
#include <stdint.h>
static DdNode *truth(DdManager *dd,unsigned table,int level){DdNode *t,*e,*r;unsigned width,mask;if(level==3)return(table&1)?dd->one:Cudd_Not(dd->one);width=1u<<(2-level);mask=(1u<<width)-1;t=truth(dd,table>>width,level+1);Cudd_Ref(t);e=truth(dd,table&mask,level+1);Cudd_Ref(e);r=Cudd_bddIte(dd,dd->vars[level],t,e);Cudd_Ref(r);Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;}
int main(void){uint64_t hash=UINT64_C(14695981039346656037);DdManager *dd=Cudd_Init(3,0,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);unsigned table;for(table=0;table<256;table++){DdNode *f=truth(dd,table,0);char *data=NULL;size_t size=0,i;FILE *stream;Cudd_Ref(f);stream=open_memstream(&data,&size);Extra_PrintKMapRelation(stream,dd,f,Cudd_Not(f),1,2,dd->vars+2,dd->vars);fclose(stream);for(i=0;i<size;i++){hash^=(unsigned char)data[i];hash*=UINT64_C(1099511628211);}free(data);Cudd_RecursiveDeref(dd,f);}printf("%016llx\n",(unsigned long long)hash);Cudd_Quit(dd);return 0;}
