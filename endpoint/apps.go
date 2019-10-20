package endpoint

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"net/http"
)

// Bind http handler
func BindApps(apps *hive.Apps, pattern string) {

	var upgrader = websocket.Upgrader{}

	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Auth
		var aid uuid.UUID
		{
			rAid, err := uuid.Parse(r.URL.Query().Get("uuid"))
			if err != nil {
				w.Header().Add("X-Error", "Invalid uuid")
				w.WriteHeader(400)
				return
			}

			token := r.URL.Query().Get("token")
			if token == "" {
				w.Header().Add("X-Error", "Empty token")
				w.WriteHeader(400)
				return
			}

			//@TODO: auth

			aid = rAid
		}

		// Accept connection
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("Upgrade connection error: %v", err)
			return
		}

		// Register app connection
		apps.AddConnection(aid, c)

		// Cleanup
		defer func() {
			// Remove app connection
			apps.RemoveConnection(aid, c)
			// close
			err := c.Close()
			if err != nil {
				log.Error("Connection close error: %v", err)
			}
		}()

		// Process
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				if err.Error() != "websocket: close 1005 (no status)" && err.Error() != "websocket: close 1001 (going away)" {
					log.Error("Connection read error: %v", err)
				}
				break
			}
			if mt == websocket.TextMessage {
				apps.DispatchMessage(aid, string(message))
			}
		}
	})
}
