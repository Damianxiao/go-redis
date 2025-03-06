package server

import (
	"net"
	"time"
)

type Conn struct {
	addr       string
	conn       net.Conn
	createTime time.Time
	msgCh      chan Message
}

func NewConn(conn net.Conn, msgCh chan Message) Conn {
	c := Conn{
		addr:       conn.RemoteAddr().String(),
		conn:       conn,
		createTime: time.Now(),
		msgCh:      msgCh,
	}
	return c
}

func (c *Conn) read() error {
	readBuf := make([]byte, 1024)
	// write msg to client
	go func() {
		for {
			select {}
		}
	}()
	for {
		count, err := c.conn.Read(readBuf)
		if err != nil {
			return err
		}
		msg := make([]byte, count)
		copy(msg, readBuf[:count])
		data := NewMessage(*c, msg)
		c.msgCh <- data
	}
}

func (c *Conn) Write(msg []byte) error {
	_, err := c.conn.Write(msg)
	if err != nil {
		return err
	}
	return nil
}
