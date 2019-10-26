package hive

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stepan-s/ws-bro/log"
)

type UserPayload struct {
	Receiver uuid.UUID
}

type AppPayload struct {
	Transmitter uint32
}

func RouterStart(users *Users, apps *Apps)  {
	go func() {
		for {
			msg := users.ReceiveMessage()
			aid, err := extractReceiverUUID(msg)
			if err != nil {
				log.Error("User:%d say:%s", msg.Uid, msg.Payload)
			} else {
				apps.SendMessage(*aid, msg.Uid, msg.Payload)
			}
		}
	}()

	go func() {
		for {
			msg := apps.ReceiveMessage()
			// send to all users connected to the app
			if msg.Uids != nil {
				err := injectTransmitterUUID(&msg)
				if err != nil {
					log.Error("App:%v say:%s", msg.Aid, msg.Payload)
				} else {
					item := msg.Uids.Front()
					for item != nil {
						users.SendMessage(item.Value.(uint32), msg.Payload)
						item = item.Next()
					}
				}
			}
		}
	}()
}

func extractReceiverUUID(msg Message) (*uuid.UUID, error)  {
	var payload UserPayload
	err := json.Unmarshal(msg.Payload, payload)
	if err != nil {
		return nil, err
	} else {
		return &payload.Receiver, nil
	}
}

func injectTransmitterUUID(msg *AppMessageFrom) error  {
	var objmap map[string]*json.RawMessage
	err := json.Unmarshal(msg.Payload, &objmap)
	if err != nil {
		return err
	}

	json_value := json.RawMessage(msg.Aid.String())
	objmap["Transmitter"] = &json_value
	payload, err := json.Marshal(objmap);
	if err != nil {
		return err
	}

	msg.Payload = payload

	return nil
}
