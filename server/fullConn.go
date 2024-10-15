package main

import (
	"net"
)

// fullConn 是一个包装了 net.Conn 的结构，允许我们"放回"已读取的字节
type fullConn struct {
	net.Conn
	peeked  []byte
	peekedN int
}

func (c *fullConn) Read(p []byte) (n int, err error) {
	if c.peekedN > 0 {
		n = copy(p, c.peeked[len(c.peeked)-c.peekedN:])
		c.peekedN -= n
		return
	}
	return c.Conn.Read(p)
}
