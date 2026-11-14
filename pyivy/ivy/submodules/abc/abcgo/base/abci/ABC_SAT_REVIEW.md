# abcSat.go port review

- The combinational AIG path assigns variables to the constant, CIs, and DFS
  nodes, emits equivalence clauses for AND nodes, and asserts each CO driver.
  This preserves the miter's Boolean relation while avoiding tagged C pointers.
- SAT limits use the bsat.Solver2 source-limit API. Return values preserve
  ABC's -1 timeout, 0 satisfiable, and 1 unsatisfiable convention; a SAT
  result stores the CI assignment in Network.Model.
- The default BDD path remains enabled. It calls Abc_NodeBddToCnf for each
  logic node and emits the two phase cover clauses. C-owned Mem_Flex storage
  is omitted because SOP values are Go-owned strings.
- Sorter CNF writing retains ABC's gen / read; strash command sequence,
  selected adjacent output roots, variable ordering, and AIG AND clause count.
- Focused tests cover a satisfiable AND miter and contradictory positive /
  negative outputs. Package tests could not reach type checking because
  unrelated raw C syntax remains in abcMfs.go and abcSaucy.go.
