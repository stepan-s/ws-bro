package hive

import (
	"container/list"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stepan-s/ws-bro/log"
)

// A message to app
type AppMessage struct {
	Aid     uuid.UUID
	Payload []byte
}

// A connection message
type appConnectionMessage struct {
	cmd  uint8
	aid  uuid.UUID
	conn *websocket.Conn
}

type App struct {
	uids *list.List
	conn *websocket.Conn
}

type AppsStats struct {
	TotalConnectionsAccepted uint64
	CurrentConnections       uint32
	MessagesReceived         uint64
	MessagesTransmitted      uint64
}

// A apps hive
type Apps struct {
	conns    map[uuid.UUID]*App
	chanIn   chan AppMessage
	chanOut  chan AppMessage
	chanConn chan appConnectionMessage
	Stats    AppsStats
}

// Instantiate apps hive
func NewApps() *Apps {
	apps := new(Apps)
	apps.conns = make(map[uuid.UUID]*App)
	apps.chanIn = make(chan AppMessage, 1000)
	apps.chanOut = make(chan AppMessage, 1000)
	apps.chanConn = make(chan appConnectionMessage, 1000)
	go func() {
		for {
			select {
			case msg := <-apps.chanConn:
				switch msg.cmd {
				case ADD:
					apps.addConnection(msg.aid, msg.conn)
				case REMOVE:
					apps.removeConnection(msg.aid)
				}
			case msg := <-apps.chanIn:
				apps.sendMessage(msg.Aid, msg.Payload)
			}
		}
	}()
	return apps
}

// Register app connection
func (apps *Apps) addConnection(aid uuid.UUID, conn *websocket.Conn) {
	_, exists := apps.conns[aid]
	if exists {
		log.Error("App already connected, disconnect: %v", aid)
		err := conn.Close()
		if err != nil {
			log.Error("Fail disconnect app: %v", aid)
		}
		//@TODO: notify app or|and user?
		return
	}

	log.Info("Hello app: %v", aid)
	apps.conns[aid] = &App{
		uids: list.New(),
		conn: conn,
	}
	apps.Stats.TotalConnectionsAccepted += 1
	apps.Stats.CurrentConnections += 1

	//@TODO: request uid list
}

// Unregister app connection
func (apps *Apps) removeConnection(aid uuid.UUID) {
	_, exists := apps.conns[aid]
	if !exists {
		return
	}

	// No connection left - remove app
	delete(apps.conns, aid)
	apps.Stats.CurrentConnections -= 1
	log.Info("Bye app: %v", aid)
}

func (apps *Apps) HandleAppConnection(conn *websocket.Conn, aid uuid.UUID) {
	// Register app connection
	apps.chanConn <- appConnectionMessage{ADD, aid, conn}

	// Cleanup
	defer func() {
		// Remove app connection
		apps.chanConn <- appConnectionMessage{REMOVE, aid, conn}
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
			apps.Stats.MessagesReceived += 1
			// Dispatch message from app
			apps.chanOut <- AppMessage{aid, message}
		}
	}
}

// Send message to all app connections
func (apps *Apps) sendMessage(aid uuid.UUID, payload []byte) {
	app, exists := apps.conns[aid]
	if exists {
		err := app.conn.WriteMessage(websocket.TextMessage, payload)
		if err != nil {
			log.Error("Send error: %v", err)
		} else {
			apps.Stats.MessagesTransmitted += 1
		}
	}
}

// Send message to all app connections
func (apps *Apps) SendMessage(aid uuid.UUID, payload []byte) {
	apps.chanIn <- AppMessage{aid, payload}
}

// Read message from app, blocked
func (apps *Apps) ReceiveMessage() AppMessage {
	return <-apps.chanOut
}
