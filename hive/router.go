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
	Receiver uint32
}

func RouterStart(users *Users, apps *Apps)  {
	go func() {
		for {
			msg := users.ReceiveMessage()
			var payload UserPayload
			err := json.Unmarshal(msg.Payload, payload)
			if err != nil {
				log.Error("User:%d say:%v", msg.Uid, msg.Payload)
			} else {
				apps.SendMessage(payload.Receiver, msg.Payload)
			}
		}
	}()

	go func() {
		for {
			msg := apps.ReceiveMessage()
			var payload AppPayload
			err := json.Unmarshal(msg.Payload, payload)
			if err != nil {
				log.Error("User:%d say:%v", msg.Aid, msg.Payload)
			} else {
				users.SendMessage(payload.Receiver, msg.Payload)
			}
		}
	}()
}
