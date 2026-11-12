#include <stdio.h>
#include <stdlib.h>
#include <string.h>
struct solver_t_;
extern void satoko_bookmark(struct solver_t_*);
extern void satoko_rollback(struct solver_t_*);
extern void satoko_reset(struct solver_t_*);
extern struct solver_t_* satoko_create(void);
extern int satoko_add_clause(struct solver_t_*,int*,int);
typedef struct solver_t_ satoko_t;
extern int satoko_parse_dimacs(char*,satoko_t**);
extern int satoko_solve(satoko_t*);
extern void satoko_setnvars(satoko_t*,int);
extern void satoko_destroy(satoko_t*);
extern int satoko_solve_assumptions(satoko_t*,int*,int);
int main(int argc,char **argv){satoko_t *s=0;if(argc<2)return 2;
if(strcmp(argv[1],"--lifecycle")==0){s=satoko_create();satoko_setnvars(s,2);int c[]={0,2},d[]={1,2},a[]={3},b[]={0,3};satoko_add_clause(s,c,2);satoko_bookmark(s);satoko_add_clause(s,d,2);printf("LIFE=%d,%d,",satoko_solve(s),satoko_solve_assumptions(s,a,1));satoko_reset(s);satoko_setnvars(s,2);satoko_add_clause(s,c,2);printf("%d\n",satoko_solve_assumptions(s,b,2));satoko_destroy(s);return 0;}
if(!satoko_parse_dimacs(argv[1],&s)){printf("RESULT=-1\n");return 0;}FILE *f=fopen(argv[1],"r");int nv,nc;fscanf(f,"p cnf %d %d",&nv,&nc);fclose(f);satoko_setnvars(s,nv);int result;if(argc>2){int *a=malloc((argc-2)*sizeof(int));for(int i=2;i<argc;i++)a[i-2]=atoi(argv[i]);result=satoko_solve_assumptions(s,a,argc-2);free(a);}else result=satoko_solve(s);printf("RESULT=%d\n",result);satoko_destroy(s);return 0;}
