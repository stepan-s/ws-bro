package hive

import (
	"container/list"
	"github.com/stepan-s/ws-bro/log"
)

// UserMessageEvent A message to user
type UserMessageEvent struct {
	Uid        uint32
	RawMessage []byte
}

// A connection message
type userConnectionEvent struct {
	cmd  uint8
	uid  uint32
	conn AConnection
}

// Users A users hive
type Users struct {
	conns    map[uint32]*list.List
	chanIn   chan UserMessageEvent
	chanOut  chan UserMessageEvent
	chanConn chan userConnectionEvent
	stats    AUserStat
}

// NewUsers Instantiate users hive
func NewUsers(stats AUserStat) *Users {
	users := new(Users)
	users.conns = make(map[uint32]*list.List)
	users.chanIn = make(chan UserMessageEvent, 1000)
	users.chanOut = make(chan UserMessageEvent, 1000)
	users.chanConn = make(chan userConnectionEvent, 1000)
	users.stats = stats
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
func (users *Users) addConnection(uid uint32, conn AConnection) {
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
		users.stats.Connected()
	} else {
		users.stats.ConnectionAdded()
	}
	conn.Start()
}

// Unregister user connection
func (users *Users) removeConnection(uid uint32, conn AConnection) {
	conns, exists := users.conns[uid]
	if !exists {
		return
	}

	removed := false
	item := conns.Front()
	for item != nil {
		if item.Value == conn {
			conns.Remove(item)
			removed = true
			break
		}
		item = item.Next()
	}

	disconnected := false
	if conns.Front() == nil {
		// No connection left - remove user
		delete(users.conns, uid)
		disconnected = true
		log.Info("Bye user: %d", uid)
	} else {
		log.Debug("Remove connection for user: %d", uid)
	}

	if disconnected {
		users.stats.Disconnected()
	} else if removed {
		users.stats.ConnectionRemoved()
	}
}

// Send message to all user connections
func (users *Users) sendEvent(event UserMessageEvent) {
	conns, exists := users.conns[event.Uid]
	if exists {
		item := conns.Front()
		for item != nil {
			item.Value.(AConnection).Send(event.RawMessage)
			users.stats.Transmitted()
			item = item.Next()
		}
	}
}

// SendEvent Send message to all user connections
func (users *Users) SendEvent(event UserMessageEvent) {
	users.chanIn <- event
}

// ReceiveEvent Read message from user, blocked
func (users *Users) ReceiveEvent() UserMessageEvent {
	return <-users.chanOut
}

func (users *Users) ConnectionAdd(uid uint32, conn AConnection) {
	users.chanConn <- userConnectionEvent{ADD, uid, conn}
}

func (users *Users) ConnectionRemove(uid uint32, conn AConnection) {
	users.chanConn <- userConnectionEvent{REMOVE, uid, conn}
}

func (users *Users) ConnectionMessage(uid uint32, message []byte) {
	users.stats.Received()
	users.chanOut <- UserMessageEvent{uid, message}
}
