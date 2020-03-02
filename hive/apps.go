package hive

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stepan-s/ws-bro/log"
	"io/ioutil"
	"net/http"
)

// A message to app
type AppMessageToEvent struct {
	Aid        uuid.UUID
	Uid        uint32
	RawMessage []byte
}

// A message from app
type AppMessageFromEvent struct {
	Aid        uuid.UUID
	Uids       []uint32
	RawMessage []byte
}

// A connection message
type appConnectionEvent struct {
	cmd  uint8
	aid  uuid.UUID
	conn *websocket.Conn
}

type appGetUidsEvent struct {
	aid      uuid.UUID
	attempts byte
}

type AppUidsEvent struct {
	Cmd  uint8
	Aid  uuid.UUID
	Uids []uint32
}

type appConnectedEvent struct {
	uid uint32
	aids []uuid.UUID
}

type App struct {
	uids []uint32
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
	conns         map[uuid.UUID]*App
	chanIn        chan AppMessageToEvent
	chanOutUids   chan AppMessageFromEvent
	chanOut       chan AppMessageFromEvent
	chanConn      chan appConnectionEvent
	chanGetUids   chan appGetUidsEvent
	chanUids      chan AppUidsEvent
	chanConnected chan appConnectedEvent
	Stats         AppsStats
	uidsApiUrl    string
}

type uidsReponse struct {
	Uids []uint32
}

// Instantiate apps hive
func NewApps(uidsApiUrl string) *Apps {
	apps := new(Apps)
	apps.conns = make(map[uuid.UUID]*App)
	apps.chanIn = make(chan AppMessageToEvent, 10000)
	apps.chanOutUids = make(chan AppMessageFromEvent, 10000)
	apps.chanOut = make(chan AppMessageFromEvent, 10000)
	apps.chanConn = make(chan appConnectionEvent, 10000)
	apps.chanGetUids = make(chan appGetUidsEvent, 10000)
	apps.chanUids = make(chan AppUidsEvent, 10000)
	apps.chanConnected = make(chan appConnectedEvent, 10000)
	apps.uidsApiUrl = uidsApiUrl
	go func() {
		for {
			select {
			case event := <-apps.chanConn:
				switch event.cmd {
				case ADD:
					apps.addConnection(event.aid, event.conn)
				case REMOVE:
					apps.removeConnection(event.aid)
				}
			case event := <-apps.chanIn:
				apps.sendEvent(event)
			case event := <-apps.chanUids:
				switch event.Cmd {
				case ADD:
					apps.addUids(event)
				case REMOVE:
					apps.removeUids(event)
				}
			case event := <-apps.chanConnected:
				apps.replyConnected(event)
			case event := <-apps.chanOutUids:
				conn, exists := apps.conns[event.Aid]
				if exists && len(conn.uids) > 0 {
					event.Uids = conn.uids
					apps.chanOut <- event
				}
			}
		}
	}()
	for w := 0; w < 4; w++ {
		go apps.getUidsWorker()
	}
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
		uids: []uint32{},
		conn: conn,
	}
	apps.Stats.TotalConnectionsAccepted += 1
	apps.Stats.CurrentConnections += 1

	apps.chanGetUids <- appGetUidsEvent{aid, 0}
}

func (apps *Apps) getUidsWorker()  {
	for {
		select {
		case event := <-apps.chanGetUids:
			err, uids := apps.getUids(event.aid)
			if err != nil && event.attempts < 10 {
				event.attempts++
				apps.chanGetUids <- event
			} else {
				apps.chanUids <- AppUidsEvent{ADD,event.aid, uids}
			}
		}
	}
}

// Request uid list
func (apps *Apps) getUids(aid uuid.UUID) (error, []uint32) {
	req, err := http.NewRequest("GET", apps.uidsApiUrl, nil)
	if err != nil {
		log.Error("Fail init request: %v", err)
		return err, nil
	}

	q := req.URL.Query()
	q.Add("aid", aid.String())
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Fail do request: %v", err)
		return err, nil
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Fail get response: %v", err)
		return err, nil
	}

	var uids uidsReponse
	err = json.Unmarshal(buf, &uids)
	if err != nil {
		log.Error("Fail parse response: %v", err)
		return err, nil
	}

	return nil, uids.Uids
}

func (apps *Apps) addUids(event AppUidsEvent) {
	conn, exists := apps.conns[event.Aid]
	if !exists {
		return
	}

	var added []uint32
	for _, uid := range event.Uids {
		add := true
		// check exist
		for _, item := range conn.uids {
			if uid == item {
				add = false
				break
			}
		}
		if add {
			added = append(added, uid)
			conn.uids = append(conn.uids, uid)
		}
	}
	rawMessage, err := MessageUserConnectedPack(&MessageUserConnected{
		Action: ACTION_CONNECTED,
		List: []appConnection{{
			Aid: event.Aid,
			Ip:  conn.conn.RemoteAddr().String(),
		}},
	})
	if err != nil {
		log.Error("Fail pack %v", err)
	}
	apps.chanOut <- AppMessageFromEvent{
		Aid:        event.Aid,
		Uids:       added,
		RawMessage: rawMessage,
	}
}

func (apps *Apps) removeUids(event AppUidsEvent) {
	conn, exists := apps.conns[event.Aid]
	if !exists {
		return
	}

	var uids []uint32
	for _, item := range conn.uids {
		left := true
		// check remove
		for _, uid := range event.Uids {
			if uid == item {
				left = false
				break
			}
		}
		if left {
			uids = append(uids, item)
		}
	}
	conn.uids = uids

	rawMessage, err := MessageUserConnectedPack(&MessageUserConnected{
		Action: ACTION_DISCONNECTED,
		List: []appConnection{{
			Aid: event.Aid,
			Ip:  conn.conn.RemoteAddr().String(),
		}},
	})
	if err != nil {
		log.Error("Fail pack %v", err)
	}
	apps.chanOut <- AppMessageFromEvent{
		Aid:        event.Aid,
		Uids:       event.Uids,
		RawMessage: rawMessage,
	}
}

func (apps *Apps) replyConnected(event appConnectedEvent)  {
	list := []appConnection{}
	for _, aid := range event.aids {
		conn, exists := apps.conns[aid]
		if exists {
			for _, uid := range conn.uids {
				if uid == event.uid {
					list = append(list, appConnection{
						Aid: aid,
						Ip:  conn.conn.RemoteAddr().String(),
					})
					break
				}
			}
		}
	}
	rawMessage, err := MessageUserConnectedPack(&MessageUserConnected{
		Action: ACTION_CONNECTED,
		List: list,
	})
	if err != nil {
		log.Error("Fail pack %v", err)
	}
	apps.chanOut <- AppMessageFromEvent{
		Aid:        uuid.Nil,
		Uids:       []uint32{event.uid},
		RawMessage: rawMessage,
	}
}

// Unregister app connection
func (apps *Apps) removeConnection(aid uuid.UUID) {
	conn, exists := apps.conns[aid]
	if !exists {
		return
	}

	rawMessage, err := MessageUserDisconnectedPack(&MessageUserDisconnected{
		Action: ACTION_DISCONNECTED,
		List:   []uuid.UUID{aid},
	})
	if err != nil {
		log.Error("Fail pack %v", err)
	}
	apps.chanOut <- AppMessageFromEvent{
		Aid:        aid,
		Uids:       conn.uids,
		RawMessage: rawMessage,
	}

	// No connection left - remove app
	delete(apps.conns, aid)
	apps.Stats.CurrentConnections -= 1
	log.Info("Bye app: %v", aid)
}

func (apps *Apps) HandleConnection(conn *websocket.Conn, aid uuid.UUID) {
	// Register app connection
	apps.chanConn <- appConnectionEvent{ADD, aid, conn}

	// Cleanup
	defer func() {
		// Remove app connection
		apps.chanConn <- appConnectionEvent{REMOVE, aid, conn}
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
			apps.Stats.MessagesReceived += 1
			// Dispatch message from app
			apps.chanOutUids <- AppMessageFromEvent{aid, nil, message}
		}
	}
}

// Send message to all app connections
func (apps *Apps) sendEvent(event AppMessageToEvent) {
	app, exists := apps.conns[event.Aid]
	if exists {
		send := false
		if event.Uid == SYSUID {
			send = true
		} else {
			// check uid is linked to app
			if app.uids != nil {
				for _, item := range app.uids {
					if event.Uid == item {
						send = true
						break
					}
				}
			}
		}

		if send {
			// uid can send to app
			err := app.conn.WriteMessage(websocket.TextMessage, event.RawMessage)
			if err != nil {
				log.Error("Send error: %v", err)
			} else {
				apps.Stats.MessagesTransmitted += 1
			}
		}
	}
}

// Send message to all app connections
func (apps *Apps) SendEvent(event AppMessageToEvent) {
	apps.chanIn <- event
}

// Read message from app, blocked
func (apps *Apps) ReceiveEvent() AppMessageFromEvent {
	return <-apps.chanOut
}

// Unregister app connection
func (apps *Apps) UpdateUids(event AppUidsEvent) {
	apps.chanUids <- event
}

func (apps *Apps) getConnected(event appConnectedEvent) {
	apps.chanConnected <- event
}
