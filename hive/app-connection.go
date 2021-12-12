package hive

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stepan-s/ws-bro/log"
	"net"
	"time"
)

type AppConnection struct {
	handler AAppHandler
	aid     uuid.UUID
	conn    *websocket.Conn
	send    chan []byte
}

func NewAppConnection(handler AAppHandler, aid uuid.UUID, conn *websocket.Conn) *AppConnection {
	c := &AppConnection{
		handler: handler,
		aid:     aid,
		conn:    conn,
		send:    make(chan []byte, 10),
	}
	handler.ConnectionAdd(aid, c)
	return c
}

func (c *AppConnection) Start() {
	go func() {
		defer func() {
			// Remove app connection
			c.handler.ConnectionRemove(c.aid, c)

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
				c.handler.ConnectionMessage(c.aid, message)
			}
		}
	}()

	go func() {
		defer func() {
			// Remove app connection
			c.handler.ConnectionRemove(c.aid, c)

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

func (c *AppConnection) Send(message []byte) {
	if len(c.send) < cap(c.send) {
		c.send <- message
	}
}

func (c *AppConnection) Close() {
	close(c.send)
}

func (c *AppConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
