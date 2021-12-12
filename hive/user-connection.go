package hive

import (
	"github.com/gorilla/websocket"
	"github.com/stepan-s/ws-bro/log"
	"net"
	"time"
)

type UserConnection struct {
	handler AUserHandler
	uid     uint32
	conn    *websocket.Conn
	send    chan []byte
}

func NewUserConnection(handler AUserHandler, uid uint32, conn *websocket.Conn) *UserConnection {
	c := &UserConnection{
		handler: handler,
		uid:     uid,
		conn:    conn,
		send:    make(chan []byte, 10),
	}
	handler.ConnectionAdd(uid, c)
	return c
}

func (c *UserConnection) Start() {
	go func() {
		defer func() {
			// Remove app connection
			c.handler.ConnectionRemove(c.uid, c)

			// close
			err := c.conn.Close()
			if err != nil {
				log.Error("Connection close error: %v", err)
			}
		}()

		for {
			mt, message, err := c.conn.ReadMessage()
			if err != nil {
				if err.Error() != "websocket: close 1005 (no status)" &&
					err.Error() != "websocket: close 1001 (going away)" &&
					err.Error() != "websocket: close 1000 (normal)" {
					log.Error("Connection read error: %v", err)
				}
				break
			}
			if mt == websocket.TextMessage {
				c.handler.ConnectionMessage(c.uid, message)
			}
		}
	}()

	go func() {
		defer func() {
			// Remove app connection
			c.handler.ConnectionRemove(c.uid, c)

			// close
			err := c.conn.Close()
			if err != nil {
				log.Error("Connection close error: %v", err)
			}
		}()

		for {
			select {
			case msg, ok := <-c.send:
				_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if !ok {
					_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				err := c.conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Error("Send error: %v", err)
				}
			}
		}
	}()
}

func (c *UserConnection) Send(message []byte) {
	if len(c.send) < cap(c.send) {
		c.send <- message
	}
}

func (c *UserConnection) Close() {
	close(c.send)
}

func (c *UserConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
