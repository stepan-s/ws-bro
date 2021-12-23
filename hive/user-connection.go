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
	conn.SetPongHandler(func(appData string) error {
		_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
		return nil
	})
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

		_ = c.conn.SetReadDeadline(time.Now().Add(70 * time.Second))
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
		ticker := time.NewTicker(60 * time.Second)
		defer func() {
			ticker.Stop()

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
				_ = c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
				if !ok {
					_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				err := c.conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Error("Send error: %v", err)
					return
				}
			case <-ticker.C:
				err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(1*time.Second))
				if err != nil {
					log.Error("Ping error: %v", err)
					return
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
