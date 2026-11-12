package mvc

func Mvc_ManagerStart() *Mvc_Manager_t { return &Mvc_Manager_t{} }

// Go's collector owns all covers and cubes associated with a manager.
func Mvc_ManagerFree(_ *Mvc_Manager_t) {}
