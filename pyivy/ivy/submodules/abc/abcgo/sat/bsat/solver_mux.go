package bsat

// AddMux adds ABC's original CNF for output = control ? then : otherwise.
// Complement flags apply to the corresponding variables. It returns the
// number of clauses emitted by sat_solver_add_mux.
func (s *sat_solver_t) AddMux(output, control, then, otherwise int32, complementControl, complementThen, complementElse, complementOutput bool) int32 {
	s.lastStatus = Unknown
	return sat_solver_add_mux(s, output, control, then, otherwise, boolInt32(complementControl), boolInt32(complementThen), boolInt32(complementElse), boolInt32(complementOutput))
}

// AddMux4 adds ABC's original four-way MUX clauses. control0 is the low
// selector bit, and data0 through data3 retain the source variable order.
func (s *sat_solver_t) AddMux4(output, control0, control1, data0, data1, data2, data3 int32) int32 {
	s.lastStatus = Unknown
	return sat_solver_add_mux41(s, output, control0, control1, data0, data1, data2, data3)
}
