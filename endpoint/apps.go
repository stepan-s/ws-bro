package endpoint

import (
	"crypto/sha256"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"net/http"
	"strconv"
	"time"
)

var AppAuthSignTTL int64 = 60

func SignAppAuth(aid uuid.UUID, ts int64, authKey string) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s:%d:%s", aid.String(), ts, authKey)))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Bind http handler
func BindApps(apps *hive.Apps, pattern string, authKey string) {
	var upgrader = websocket.Upgrader{}

	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Auth

		token := r.URL.Query().Get("token")
		if token == "" {
			w.Header().Add("X-Error", "Empty token")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var aid uuid.UUID
		{
			aid, err := uuid.Parse(r.URL.Query().Get("aid"))
			if err != nil {
				w.Header().Add("X-Error", "Invalid aid")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			sign := r.URL.Query().Get("sign")
			if sign == "" {
				w.Header().Add("X-Error", "Empty sign")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			ts, err := strconv.ParseInt(r.URL.Query().Get("ts"), 10, 64)
			if err != nil {
				w.Header().Add("X-Error", "Invalid ts")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			now := time.Now().Unix()
			if (now - ts) > AppAuthSignTTL {
				log.Warning("Decline connection, reason: incorrect ts: %d for app: %s", ts, aid.String())
				w.Header().Add("X-Error", "Expired sign")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if SignAppAuth(aid, ts, authKey) != sign {
				log.Warning("Decline connection, reason: incorrect sign for app: %s", aid.String())
				w.Header().Add("X-Error", "Invalid sign")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// Accept connection
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("Upgrade connection error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		apps.HandleConnection(conn, aid)
	})
}
