package endpoint

import (
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var AuthSignTTL int64 = 1800

func SignAuth(uid uint32, ts int64, authKey string) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("$d:$d:%s", uid, ts, authKey)))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Bind http handler
func BindUsers(users *hive.Users, pattern string, allowedOrigins string, authKey string) {
	var upgrader = websocket.Upgrader{} // use default options

	origins := make(map[string]bool)
	{
		entries := strings.Split(allowedOrigins, ";")
		for _, entry := range entries {
			origins[strings.TrimSpace(entry)] = true
		}
	}

	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Auth
		var uid uint32 = 0
		{
			origin := r.Header.Get("Origin")
			_, exists := origins[origin]
			if !exists {
				log.Warning("Decline connection, reason: disallowed origin %s", origin)
				w.WriteHeader(403)
				return
			}

			rUid, err := strconv.ParseInt(r.URL.Query().Get("uid"), 10, 32)
			if err != nil {
				w.Header().Add("X-Error", "Invalid uid")
				w.WriteHeader(400)
				return
			}
			uid = uint32(rUid)

			sign := r.URL.Query().Get("sign")
			if sign == "" {
				w.Header().Add("X-Error", "Empty sign")
				w.WriteHeader(400)
				return
			}

			ts, err := strconv.ParseInt(r.URL.Query().Get("ts"), 10, 64)
			if err != nil {
				w.Header().Add("X-Error", "Invalid ts")
				w.WriteHeader(400)
				return
			}

			now := time.Now().Unix()
			if (now - ts) > AuthSignTTL {
				log.Warning("Decline connection, reason: incorrect ts: %d for user: %d", ts, uid)
				w.Header().Add("X-Error", "Expired sign")
				w.WriteHeader(403)
				return
			}

			hash := sha256.New()
			hash.Write([]byte(fmt.Sprintf("$d:$d:%s", rUid, ts, authKey)))
			if SignAuth(uid, ts, authKey) != sign {
				log.Warning("Decline connection, reason: incorrect sign for user: %d", uid)
				w.Header().Add("X-Error", "Invalid sign")
				w.WriteHeader(403)
				return
			}
		}

		// Accept connection
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("Upgrade connection error: %v", err)
			return
		}

		// Register user connection
		users.AddConnection(uid, c)

		// Cleanup
		defer func() {
			// Remove connection from user
			users.RemoveConnection(uid, c)
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
				if err.Error() != "websocket: close 1005 (no status)" {
					log.Error("Connection read error: %v", err)
				}
				break
			}
			if mt == websocket.TextMessage {
				users.DispatchMessage(uid, string(message))
			}
		}
	})
}
