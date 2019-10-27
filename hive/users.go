package hive

import (
	"container/list"
	"github.com/gorilla/websocket"
	"github.com/stepan-s/ws-bro/log"
)

// A message to user
type UserMessageEvent struct {
	Uid        uint32
	RawMessage []byte
}

// A connection message
type userConnectionEvent struct {
	cmd  uint8
	uid  uint32
	conn *websocket.Conn
}

type UsersStats struct {
	TotalConnectionsAccepted uint64
	CurrentConnections       uint32
	TotalUsersConnected      uint64
	CurrentUsersConnected    uint32
	MessagesReceived         uint64
	MessagesTransmitted      uint64
}

// A users hive
type Users struct {
	conns    map[uint32]*list.List
	chanIn   chan UserMessageEvent
	chanOut  chan UserMessageEvent
	chanConn chan userConnectionEvent
	Stats    UsersStats
}

// Instantiate users hive
func NewUsers() *Users {
	users := new(Users)
	users.conns = make(map[uint32]*list.List)
	users.chanIn = make(chan UserMessageEvent, 1000)
	users.chanOut = make(chan UserMessageEvent, 1000)
	users.chanConn = make(chan userConnectionEvent, 1000)
	go func() {
		for {
			select {
			case event := <-users.chanConn:
				switch event.cmd {
				case ADD:
					users.addConnection(event.uid, event.conn)
				case REMOVE:
					users.removeConnection(event.uid, event.conn)
				}
			case event := <-users.chanIn:
				users.sendEvent(event)
			}
		}
	}()
	return users
}

// Register user connection
func (users *Users) addConnection(uid uint32, conn *websocket.Conn) {
	conns, exists := users.conns[uid]
	if !exists {
		log.Info("Hello user: %d", uid)
		conns = list.New()
	} else {
		log.Debug("Add connection for user: %d", uid)
	}
	conns.PushBack(conn)
	if !exists {
		users.conns[uid] = conns

		users.Stats.TotalUsersConnected += 1
		users.Stats.CurrentUsersConnected += 1
	}
	users.Stats.TotalConnectionsAccepted += 1
	users.Stats.CurrentConnections += 1
}

// Unregister user connection
func (users *Users) removeConnection(uid uint32, conn *websocket.Conn) {
	conns, exists := users.conns[uid]
	if !exists {
		return
	}

	item := conns.Front()
	for item != nil {
		if item.Value == conn {
			conns.Remove(item)
			users.Stats.CurrentConnections -= 1
			break
		}
		item = item.Next()
	}

	if conns.Front() == nil {
		// No connection left - remove user
		delete(users.conns, uid)
		users.Stats.CurrentUsersConnected -= 1
		log.Info("Bye user: %d", uid)
	} else {
		log.Debug("Remove connection for user: %d", uid)
	}
}

func (users *Users) HandleConnection(conn *websocket.Conn, uid uint32)  {
	// Register user connection
	users.chanConn <- userConnectionEvent{ADD, uid, conn}

	// Cleanup
	defer func() {
		// Remove connection from user
		users.chanConn <- userConnectionEvent{REMOVE, uid, conn}
		// close
		err := conn.Close()
		if err != nil {
			log.Error("Connection close error: %v", err)
		}
	}()

	// Process
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			if err.Error() != "websocket: close 1005 (no status)" &&
				err.Error() != "websocket: close 1001 (going away)" &&
				err.Error() != "websocket: close 1000 (normal)" {
				log.Error("Connection read error: %v", err)
			}
			break
		}
		if mt == websocket.TextMessage {
			users.Stats.MessagesReceived += 1
			// Dispatch message from user
			users.chanOut <- UserMessageEvent{uid, message}
		}
	}
}

// Send message to all user connections
func (users *Users) sendEvent(event UserMessageEvent) {
	conns, exists := users.conns[event.Uid]
	if exists {
		item := conns.Front()
		for item != nil {
			err := item.Value.(*websocket.Conn).WriteMessage(websocket.TextMessage, event.RawMessage)
			if err != nil {
				log.Error("Send error: %v", err)
			} else {
				users.Stats.MessagesTransmitted += 1
			}
			item = item.Next()
		}
	}
}

// Send message to all user connections
func (users *Users) SendEvent(event UserMessageEvent) {
	users.chanIn <- event
}

// Read message from user, blocked
func (users *Users) ReceiveEvent() UserMessageEvent {
	return <-users.chanOut
}
