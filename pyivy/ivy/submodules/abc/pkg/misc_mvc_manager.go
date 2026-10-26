package abc

// Memory-manager lifecycle from src/misc/mvc/mvcMan.c.

// MvcManager owns MVC covers and cubes created by later MVC port stages. Go
// allocation replaces the native fixed-size pools. A manager must not be
// copied; construct it with NewMvcManager and release it with Close.
type MvcManager struct {
	closed bool
}

// NewMvcManager replaces Mvc_ManagerStart. Its zero allocation state is ready
// for use; concrete cube and cover storage remains Go-owned.
func NewMvcManager() *MvcManager { return &MvcManager{} }

func (manager *MvcManager) check() error {
	if manager == nil {
		return ErrDimension
	}
	if manager.closed {
		return ErrInvariant
	}
	return nil
}

// Close replaces Mvc_ManagerFree. It rejects nil and repeated release instead
// of reproducing native invalid memory access.
func (manager *MvcManager) Close() error {
	if err := manager.check(); err != nil {
		return err
	}
	manager.closed = true
	return nil
}
