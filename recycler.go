package pool

import "net"

// Recycler allows net.Conn's to be re-used.
type Recycler struct {
	buf  chan net.Conn
	next NextConn
}

// NewRecycler creates a Recycler which will create new net.Conn's with next
// and can hold up to size net.Conn's for re-use.
func NewRecycler(next NextConn, size uint) *Recycler {
	return &Recycler{
		buf:  make(chan net.Conn, size),
		next: next,
	}
}

// Put will add c to the re-use queue for this Recycler. Connections are returned from
// Get in the order they were added with Put. Put will close the connection if
// the re-use queue is full. Put is safe for concurrent calls with Get.
func (r *Recycler) Put(c net.Conn) error {
	select {
	case r.buf <- c:
		return nil
	default:
		return c.Close()
	}
}

// Get will return a net.Conn added with Put(). If no connections are available,
// a new one is created via the supplied NextConn.
// Get can be used as a NextConn for a Pool.
func (r *Recycler) Get() (net.Conn, error) {
	select {
	case c := <-r.buf:
		return c, nil
	default:
		return r.next()
	}
}

// Close will close any net.Conn's currently in the re-use queue.
func (r *Recyler) Close() error {
	for {
		select {
		case c := <-r.buf:
			c.Close()
		default:
			return nil
		}
	}
}
