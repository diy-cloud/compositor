package bridge

import (
	"io"
	"net"
)

func Connect(left, right net.Conn) {
	go func() {
		defer left.Close()
		defer right.Close()
		io.Copy(left, right)
	}()
	go func() {
		defer left.Close()
		defer right.Close()
		io.Copy(right, left)
	}()
}
