// Go port of xsatClause.h by Bruno Schmitt (2016).
package xsat
import "fmt"
type xSAT_Clause_t struct{lits []int;learnt,marked,protected bool;lbd,activity uint32}
func xSAT_ClauseCompare(a,b *xSAT_Clause_t)bool{if len(a.lits)==2{return false};if len(b.lits)==2{return true};if a.lbd!=b.lbd{return a.lbd>b.lbd};return a.activity<b.activity}
func xSAT_ClausePrint(c *xSAT_Clause_t){fmt.Print("{ ");for _,lit:=range c.lits{fmt.Printf("%d ",lit)};fmt.Println("}")}
