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
				log.Error("User:%d say:%s", msg.Uid, msg.Payload)
			} else {
				apps.SendMessage(payload.Receiver, msg.Uid, msg.Payload)
			}
		}
	}()

	go func() {
		for {
			msg := apps.ReceiveMessage()
			// send to all users connected to the app
			if msg.Uids != nil {
				item := msg.Uids.Front()
				for item != nil {
					users.SendMessage(item.Value.(uint32), msg.Payload)
					item = item.Next()
				}
			}
		}
	}()
}
