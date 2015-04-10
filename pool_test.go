package pool

import (
	"io"
	"net"
	"os"
	"time"
)

func ExampleRecycler() {
	go func() {
		if l, err := net.Listen("tcp", "localhost:10001"); err == nil {
			p := NewPool(l.Accept)
			if c, err := p.Get(); err == nil {
				io.Copy(os.Stdout, c)
				c.Close()
			}
			l.Close()
		}
	}()
	// Output:
	// Hello World
	// Hello World
	// Hello World

	r := NewRecycler(func() (net.Conn, error) {
		return net.Dial("tcp", "localhost:10001")
	}, 1)
	p := NewPool(r.Get)

	for i := 0; i < 3; i++ {
		if c, err := p.Get(); err == nil {
			io.WriteString(c, "Hello World\n")
			r.Put(c)
		}
	}

	<-time.After(100 * time.Millisecond)
}

func ExamplePool() {
	go func() {
		if l, err := net.Listen("tcp", "localhost:10000"); err == nil {
			p := NewPool(l.Accept)
			for {
				if c, err := p.Get(); err == nil {
					go func() {
						io.Copy(os.Stdout, c)
						c.Close()
					}()
					// Output:
					// hello world
					// hello world
					// hello world
					// hello world
					// hello world
					// hello world
				}
			}
		}
	}()

	p := NewPool(Limiter(func() (net.Conn, error) {
		return net.Dial("tcp", "localhost:10000")
	}, 3))

	for i := 0; i < 6; i++ {
		if c, err := p.Get(); err == nil {
			c.Write([]byte("hello world\n"))
			go func() {
				<-time.After(100 * time.Millisecond)
				c.Close()
			}()
		}
	}

	<-time.After(100 * time.Millisecond)
}
