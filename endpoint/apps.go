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
		aid, err := uuid.Parse(r.URL.Query().Get("uuid"))
		if err != nil {
			w.Header().Add("X-Error", "Invalid uuid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token := r.URL.Query().Get("token")
		if token == "" {
			w.Header().Add("X-Error", "Empty token")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//@TODO: auth

		// Accept connection
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("Upgrade connection error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		handleAppConnection(apps, conn, aid)
	})
}

func handleAppConnection(apps *hive.Apps, conn *websocket.Conn, aid uuid.UUID) {
	// Register app connection
	apps.AddConnection(aid, conn)

	// Cleanup
	defer func() {
		// Remove app connection
		apps.RemoveConnection(aid, conn)
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
			apps.DispatchMessage(aid, string(message))
		}
	}
}
