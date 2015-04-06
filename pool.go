package pool

import (
	"errors"
	"net"
	"sync"
	"time"
)

// NextConn returns a new net.Conn
type NextConn func() (net.Conn, error)

type limitConn struct {
	net.Conn
	*sync.Cond
	limit *uint64
}

func (c *limitConn) Close() error {
	go func() {
		c.Cond.L.Lock()
		defer c.Cond.L.Unlock()
		(*c.limit)++
		c.Cond.Signal()
	}()
	return c.Conn.Close()
}

// Limiter limits the number of concurrent connections returned by next to limit.
// The number of connections is calculated as # of connections returned - # closed.
func Limiter(next NextConn, limit uint64) NextConn {
	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	return func() (net.Conn, error) {
		mu.Lock()
		defer mu.Unlock()
		for limit == 0 {
			cond.Wait()
		}
		c, err := next()
		if err == nil {
			limit--
			c = &limitConn{
				Conn:  c,
				Cond:  cond,
				limit: &limit,
			}
		}
		return c, err
	}
}

// Pool manages connections
type Pool struct {
	next NextConn
	reqs *requests
	dead chan struct{}
}

// NewPool creates a new Pool object for managing connections
func NewPool(next NextConn) *Pool {
	p := &Pool{
		next: next,
		reqs: newRequests(),
		dead: make(chan struct{}),
	}
	go p.manage()
	return p
}

var ErrPoolClosed = errors.New("pool closed")

// Get will return a net.Conn, it can be called concurrently
// and calls to Get() will be returned in the order they were called.
func (p *Pool) Get() (net.Conn, error) {
	if c, ok := <-p.reqs.submit(); ok {
		return c, nil
	}
	return nil, ErrPoolClosed
}

func (p *Pool) manage() {
	for {
		req, ok := p.reqs.next()
		if !ok {
			return
		}

		var c net.Conn
		err := errors.New("temp error")

		for err != nil {
			c, err = p.next()
			if err != nil {
				select {
				case <-time.After(5 * time.Second):
				case <-p.dead:
					close(req)
					return
				}
			}
		}

		req <- c
	}
}

// Close closes the pool and cancels calls to Get()
func (p *Pool) Close() error {
	p.reqs.close()
	close(p.dead)
	return nil
}

type requests struct {
	mu   sync.Mutex
	cond *sync.Cond
	dead bool
	reqs []chan net.Conn
}

func newRequests() *requests {
	r := &requests{reqs: make([]chan net.Conn, 0)}
	r.cond = sync.NewCond(&r.mu)
	return r
}

func (r *requests) submit() <-chan net.Conn {
	r.mu.Lock()
	defer r.mu.Unlock()
	req := make(chan net.Conn)
	if r.dead {
		close(req)
	} else {
		r.reqs = append(r.reqs, req)
		r.cond.Signal()
	}
	return req
}

func (r *requests) next() (req chan<- net.Conn, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for len(r.reqs) == 0 && !r.dead {
		r.cond.Wait()
	}
	if r.dead {
		return nil, false
	} else {
		req, r.reqs = r.reqs[0], r.reqs[1:]
		return req, true
	}
}

func (r *requests) close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dead = true
	for _, req := range r.reqs {
		close(req)
	}
	r.reqs = nil
	r.cond.Broadcast()
}
