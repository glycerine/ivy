// Package st implements ABC's chained symbol tables using Go-managed storage.
package st

type CompareFunc func(any, any) int
type HashFunc func(any, int) int
type Entry struct {
	Key, Value any
	next       *Entry
}
type Table struct {
	compare        CompareFunc
	hash           HashFunc
	bins           []*Entry
	count, density int
	growth         float64
	reorder        bool
}
type Generator struct {
	table *Table
	entry *Entry
	index int
}
type VisitResult int

const (
	Continue VisitResult = iota
	Stop
	Delete
)
const st__CONTINUE = Continue
const st__STOP = Stop
const st__DELETE = Delete

type st__table = Table
type st__table_entry = Entry
type st__generator = Generator
type st__compare_func_type = CompareFunc
type st__hash_func_type = HashFunc
type st__PFSR func(any, any, any) VisitResult

const st__DEFAULT_MAX_DENSITY = 5
const st__DEFAULT_INIT_TABLE_SIZE = 11
const st__DEFAULT_GROW_FACTOR = 2.0
const st__DEFAULT_REORDER_FLAG = 0
const st__OUT_OF_MEM = -10000
