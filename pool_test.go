package pool

import (
	"io"
	"net"
	"os"
	"time"
)

func ExamplePool() {
	go func() {
		if l, err := net.Listen("tcp", "localhost:10000"); err == nil {
			for {
				if c, err := l.Accept(); err == nil {
					go io.Copy(os.Stdout, c)
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

	p := NewPool(PoolLimiter(func() (net.Conn, error) {
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