#include <stdio.h>
#include <stdlib.h>
#include "sat/satoko/satoko.h"
int main(int argc,char **argv) {
 if(argc!=2)return 2;
 FILE *f=fopen(argv[1],"r");if(!f)return 3;
 int vars,clauses;if(fscanf(f,"%d%d",&vars,&clauses)!=2)return 4;
 satoko_t *s=satoko_create();
 for(int i=0;i<clauses;i++) {
  int n;if(fscanf(f,"%d",&n)!=1)return 5;
  int *lits=malloc((n?n:1)*sizeof(int));
  for(int k=0;k<n;k++)if(fscanf(f,"%d",&lits[k])!=1)return 6;
  int ok=satoko_add_clause(s,lits,n);free(lits);
  if(!ok){printf("-1 0\n");satoko_destroy(s);fclose(f);return 0;}
 }
 fclose(f);
 satoko_opts_t opts;satoko_default_opts(&opts);satoko_configure(s,&opts);
 if(satoko_simplify(s)!=SATOKO_OK){printf("-1 0\n");satoko_destroy(s);return 0;}
 int status=satoko_solve(s);printf("%d %ld\n",status,satoko_stats(s)->n_conflicts);
 satoko_destroy(s);return 0;
}
