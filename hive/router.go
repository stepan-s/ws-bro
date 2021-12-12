package hive

import (
	"github.com/google/uuid"
	"github.com/stepan-s/ws-bro/log"
)

type UserPayload struct {
	Receiver uuid.UUID
}

type AppPayload struct {
	Transmitter uint32
}

func RouterStart(users *Users, apps *Apps) {
	go func() {
		for {
			event := users.ReceiveEvent()
			action, err := MessageRawGetAction(event.RawMessage)
			if err != nil {
				log.Error("Fail get message action: %v, user:%d message:%s", err, event.Uid, event.RawMessage)
			} else {
				switch action {
				case ACTION_SEND_DATA:
					incomingMessage, err := MessageUserSendDataUnpack(event.RawMessage)
					if err != nil {
						log.Error("Fail unpack: %v, user:%d, message: %s", err, event.Uid, event.RawMessage)
					} else {
						outgoingMessage, err := MessageAppReceivedDataPack(&MessageAppReceivedData{
							Action: ACTION_RECEIVED_DATA,
							From:   event.Uid,
							Data:   incomingMessage.Data,
						})
						if err != nil {
							log.Error("Fail pack: %v, user:%d, message: %s", err, event.Uid, event.RawMessage)
						} else {
							apps.SendEvent(AppMessageToEvent{incomingMessage.To, event.Uid, outgoingMessage})
						}
					}
				case ACTION_GET_CONNECTED:
					incomingMessage, err := MessageUserGetConnectedUnpack(event.RawMessage)
					if err != nil {
						log.Error("Fail unpack: %v, user:%d, message: %s", err, event.Uid, event.RawMessage)
					} else {
						apps.getConnected(appConnectedEvent{
							uid:  event.Uid,
							aids: incomingMessage.List,
						})
					}
				default:
					log.Error("Invalid message action: %s, user:%d, message: %s", action, event.Uid, event.RawMessage)
				}
			}
		}
	}()

	go func() {
		for {
			event := apps.ReceiveEvent()
			action, err := MessageRawGetAction(event.RawMessage)
			if err != nil {
				log.Error("Fail get message action: %v, app:%d message:%s", err, event.Aid, event.RawMessage)
			} else {
				switch action {
				case ACTION_SEND_DATA:
					incomingMessage, err := MessageAppSendDataUnpack(event.RawMessage)
					if err != nil {
						log.Error("Fail unpack: %v, app:%d, message: %s", err, event.Aid, event.RawMessage)
					} else {
						outgoingMessage, err := MessageUserReceivedDataPack(&MessageUserReceivedData{
							Action: ACTION_RECEIVED_DATA,
							From:   event.Aid,
							Data:   incomingMessage.Data,
						})
						if err != nil {
							log.Error("Fail pack: %v, app:%d, message: %s", err, event.Aid, event.RawMessage)
						} else {
							// send to all users connected to the app
							for _, item := range event.Uids {
								users.SendEvent(UserMessageEvent{item, outgoingMessage})
							}
						}
					}
				case ACTION_CONNECTED, ACTION_DISCONNECTED, ACTION_ATTACHED, ACTION_DETACHED:
					for _, item := range event.Uids {
						users.SendEvent(UserMessageEvent{item, event.RawMessage})
					}
				default:
					log.Error("Invalid message action: %s, app:%d, message: %s", action, event.Aid, event.RawMessage)
				}
			}
		}
	}()
}
