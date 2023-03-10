package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex // protects m
	m  map[string]*call
}

// for lots of concurrent calls, fn is only called once
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	// lazy initilization
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()         // request is ongoing, plesae wait
		return c.val, c.err // request return
	}
	c := new(call)
	c.wg.Add(1)  // lock before request
	g.m[key] = c // add this request to g.m, indicating this is the first request, later comers wait for this result
	g.mu.Unlock()

	c.val, c.err = fn() // call fn() to issue the request
	c.wg.Done()         // request end

	g.mu.Lock()
	delete(g.m, key) // update g.m
	g.mu.Unlock()

	return c.val, c.err
}
