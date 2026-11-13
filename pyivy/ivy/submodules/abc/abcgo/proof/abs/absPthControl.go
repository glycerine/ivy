package abs

import "sync"

// absProverControl tracks the current asynchronous proof. A stale worker is
// never allowed to publish a proof for a newer abstraction.
type absProverControl struct {
	mu     sync.Mutex
	runID  int32
	proved bool
}

func (c *absProverControl) start(solve func(int32) int32) int32 {
	c.mu.Lock()
	c.runID++
	id := c.runID
	c.proved = false
	c.mu.Unlock()
	go func() {
		result := solve(id)
		c.mu.Lock()
		if result == 1 && c.runID == id {
			c.proved = true
		}
		c.mu.Unlock()
	}()
	return id
}

func (c *absProverControl) cancel() {
	c.mu.Lock()
	c.runID++
	c.proved = false
	c.mu.Unlock()
}

func (c *absProverControl) stopped(id int32) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return id != c.runID
}

func (c *absProverControl) check() int32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.proved {
		return 0
	}
	c.proved = false
	return 1
}
