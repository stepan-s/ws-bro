package hive

import (
	"container/list"
	"github.com/gorilla/websocket"
	"github.com/stepan-s/ws-bro/log"
)

// A message to user
type Message struct {
	Uid     uint32
	Payload []byte
}

// A connection message
type connectionMessage struct {
	cmd  uint8
	uid  uint32
	conn *websocket.Conn
}

type Stats struct {
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
	chanIn   chan Message
	chanOut  chan Message
	chanConn chan connectionMessage
	Stats    Stats
}

// Instantiate users hive
func NewUsers() *Users {
	users := new(Users)
	users.conns = make(map[uint32]*list.List)
	users.chanIn = make(chan Message, 1000)
	users.chanOut = make(chan Message, 1000)
	users.chanConn = make(chan connectionMessage, 1000)
	go func() {
		for {
			select {
			case msg := <-users.chanConn:
				switch msg.cmd {
				case ADD:
					users.addConnection(msg.uid, msg.conn)
				case REMOVE:
					users.removeConnection(msg.uid, msg.conn)
				}
			case msg := <-users.chanIn:
				users.sendMessage(msg.Uid, msg.Payload)
			}
		}
	}()
	return users
}

// Register user connection
func (users *Users) addConnection(uid uint32, conn *websocket.Conn) {
	conns, exists := users.conns[uid]
	if !exists {
		log.Info("hello user: %d", uid)
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

func (users *Users) HandleUserConnection(conn *websocket.Conn, uid uint32)  {
	// Register user connection
	users.chanConn <- connectionMessage{ADD, uid, conn}

	// Cleanup
	defer func() {
		// Remove connection from user
		users.chanConn <- connectionMessage{REMOVE, uid, conn}
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
			if err.Error() != "websocket: close 1005 (no status)" && err.Error() != "websocket: close 1001 (going away)" {
				log.Error("Connection read error: %v", err)
			}
			break
		}
		if mt == websocket.TextMessage {
			users.Stats.MessagesReceived += 1
			// Dispatch message from user
			users.chanOut <- Message{uid, message}
		}
	}
}

// Send message to all user connections
func (users *Users) sendMessage(uid uint32, payload []byte) {
	conns, exists := users.conns[uid]
	if exists {
		item := conns.Front()
		for item != nil {
			err := item.Value.(*websocket.Conn).WriteMessage(websocket.TextMessage, payload)
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
func (users *Users) SendMessage(uid uint32, payload []byte) {
	users.chanIn <- Message{uid, payload}
}

// Read message from user, blocked
func (users *Users) ReceiveMessage() Message {
	return <-users.chanOut
}
