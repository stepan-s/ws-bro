package hive

import (
	"github.com/google/uuid"
	"net"
)

type AConnection interface {
	Start()
	RemoteAddr() net.Addr
	Send([]byte)
	Close()
}

type AUserHandler interface {
	ConnectionAdd(uint32, AConnection)
	ConnectionRemove(uint32, AConnection)
	ConnectionMessage(uint32, []byte)
}

type AUserStat interface {
	Connected()
	ConnectionAdded()
	ConnectionRemoved()
	Disconnected()
	Received()
	Transmitted()
}

type AAppHandler interface {
	ConnectionAdd(uuid.UUID, AConnection)
	ConnectionRemove(uuid.UUID, AConnection)
	ConnectionMessage(uuid.UUID, []byte)
}

type AAppStat interface {
	Connected()
	Disconnected()
	Reconnected()
	Received()
	Transmitted()
}
