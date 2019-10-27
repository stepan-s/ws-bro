package hive

import (
	"encoding/json"
	"github.com/google/uuid"
)

type Message struct {
	Action string
}

// in
type MessageUserSendData struct {
	Action string
	To     uuid.UUID
	Data   json.RawMessage
}

// out
type MessageUserReceivedData struct {
	Action string
	From   uuid.UUID
	Data   json.RawMessage
}

// in
type MessageUserGetConnected struct {
	Action string
	List []uuid.UUID
}

type appConnection struct {
	Aid uuid.UUID
	Ip string
}
// out
type MessageUserConnected struct {
	Action string
	List []appConnection
}

// out
type MessageUserDisconnected struct {
	Action string
	List []uuid.UUID
}

// out
type MessageUserAttached struct {
	Action string
	List []uuid.UUID
}

// out
type MessageUserDetached struct {
	Action string
	List []uuid.UUID
}

// in
type MessageAppSendData struct {
	Action string
	Data   json.RawMessage
}

// out
type MessageAppReceivedData struct {
	Action string
	From   uint32
	Data   json.RawMessage
}

func MessageRawGetAction(rawMessage []byte) (string, error) {
	var message Message
	err := json.Unmarshal(rawMessage, &message)
	if err != nil {
		return "", err
	} else {
		return message.Action, nil
	}
}

func MessageUserSendDataUnpack(rawMessage []byte) (*MessageUserSendData, error) {
	var message MessageUserSendData
	err := json.Unmarshal(rawMessage, &message)
	if err != nil {
		return nil, err
	} else {
		return &message, nil
	}
}

func MessageAppReceivedDataPack(message *MessageAppReceivedData) ([]byte, error) {
	rawMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	} else {
		return rawMessage, nil
	}
}

func MessageAppSendDataUnpack(rawMessage []byte) (*MessageAppSendData, error) {
	var message MessageAppSendData
	err := json.Unmarshal(rawMessage, &message)
	if err != nil {
		return nil, err
	} else {
		return &message, nil
	}
}

func MessageUserReceivedDataPack(message *MessageUserReceivedData) ([]byte, error) {
	rawMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	} else {
		return rawMessage, nil
	}
}

func MessageUserConnectedPack(message *MessageUserConnected) ([]byte, error) {
	rawMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	} else {
		return rawMessage, nil
	}
}

func MessageUserDisconnectedPack(message *MessageUserDisconnected) ([]byte, error) {
	rawMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	} else {
		return rawMessage, nil
	}
}

func MessageUserGetConnectedUnpack(rawMessage []byte) (*MessageUserGetConnected, error) {
	var message MessageUserGetConnected
	err := json.Unmarshal(rawMessage, &message)
	if err != nil {
		return nil, err
	} else {
		return &message, nil
	}
}

func MessageUserGetConnectedPack(message *MessageUserGetConnected) ([]byte, error) {
	rawMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	} else {
		return rawMessage, nil
	}
}
