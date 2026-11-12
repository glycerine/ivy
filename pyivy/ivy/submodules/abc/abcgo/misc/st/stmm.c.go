package st

// The memory-managed variant shares Go-managed table storage.
func stmm_init_table(c CompareFunc, h HashFunc) *Table { return st__init_table(c, h) }
func stmm_init_table_with_params(c CompareFunc, h HashFunc, size, density int, growth float64, reorder int) *Table {
	return st__init_table_with_params(c, h, size, density, growth, reorder)
}
func stmm_free_table(t *Table)                        { st__free_table(t) }
func stmm_lookup(t *Table, k any, out *any) int       { return st__lookup(t, k, out) }
func stmm_lookup_int(t *Table, k any, out *int) int   { return st__lookup_int(t, k, out) }
func stmm_insert(t *Table, k, v any) int              { return st__insert(t, k, v) }
func stmm_add_direct(t *Table, k, v any) int          { return st__add_direct(t, k, v) }
func stmm_find(t *Table, k any, out **any) int        { return st__find(t, k, out) }
func stmm_find_or_add(t *Table, k any, out **any) int { return st__find_or_add(t, k, out) }
func stmm_copy(t *Table) *Table                       { return st__copy(t) }
func stmm_delete(t *Table, k *any, out *any) int      { return st__delete(t, k, out) }
func stmm_delete_int(t *Table, k *int, out *any) int  { return st__delete_int(t, k, out) }
func stmm_foreach(t *Table, f st__PFSR, arg any) int  { return st__foreach(t, f, arg) }
func stmm_init_gen(t *Table) *Generator               { return st__init_gen(t) }
func stmm_gen(g *Generator, k, v *any) int            { return st__gen(g, k, v) }
func stmm_gen_int(g *Generator, k *any, v *int) int   { return st__gen_int(g, k, v) }
func stmm_free_gen(g *Generator)                      { st__free_gen(g) }
func stmm_is_member(t *Table, k any) int              { return st__is_member(t, k) }
func stmm_count(t *Table) int                         { return st__count(t) }
func stmm_strcmp(a, b any) int                        { return st__strcmp(a, b) }
func stmm_strhash(k any, n int) int                   { return st__strhash(k, n) }
func stmm_numhash(k any, n int) int                   { return st__numhash(k, n) }
func stmm_ptrhash(k any, n int) int                   { return st__ptrhash(k, n) }
func stmm_numcmp(a, b any) int                        { return st__numcmp(a, b) }
func stmm_ptrcmp(a, b any) int                        { return st__ptrcmp(a, b) }
func stmm_clean(t *Table)                             { t.Clear() }
