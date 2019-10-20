package endpoint

import (
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var AuthSignTTL int64 = 1800

func SignAuth(uid uint32, ts int64, authKey string) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%d:%d:%s", uid, ts, authKey)))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Bind http handler
func BindUsers(users *hive.Users, pattern string, allowedOrigins string, authKey string) {

	origins := make(map[string]bool)
	{
		entries := strings.Split(allowedOrigins, ";")
		for _, entry := range entries {
			origins[strings.TrimSpace(entry)] = true
		}
	}

	var upgrader = websocket.Upgrader{
		CheckOrigin:func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			u, err := url.Parse(origin)
			if err != nil {
				log.Error("Fail parse origin: %v", err)
				return false
			}
			// Check domain
			domain := u.Hostname()
			_, exists := origins[domain]
			for !exists {
				// Check base domains
				pos := strings.Index(domain, ".")
				if pos < 0 {
					break
				}
				domain = domain[pos + 1:]
				_, exists = origins[domain]
			}
			if !exists {
				log.Warning("Decline connection, reason: disallowed origin %s", origin)
				return false
			}
			return true
		},
	}

	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Auth
		var uid uint32 = 0
		{
			rUid, err := strconv.ParseInt(r.URL.Query().Get("uid"), 10, 32)
			if err != nil {
				w.Header().Add("X-Error", "Invalid uid")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			uid = uint32(rUid)

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
			if (now - ts) > AuthSignTTL {
				log.Warning("Decline connection, reason: incorrect ts: %d for user: %d", ts, uid)
				w.Header().Add("X-Error", "Expired sign")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			hash := sha256.New()
			hash.Write([]byte(fmt.Sprintf("%d:%d:%s", rUid, ts, authKey)))
			if SignAuth(uid, ts, authKey) != sign {
				log.Warning("Decline connection, reason: incorrect sign for user: %d", uid)
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

		handleUserConnection(users, conn, uid)
	})
}

func handleUserConnection(users *hive.Users, conn *websocket.Conn, uid uint32)  {
	// Register user connection
	users.AddConnection(uid, conn)

	// Cleanup
	defer func() {
		// Remove connection from user
		users.RemoveConnection(uid, conn)
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
			users.DispatchMessage(uid, message)
		}
	}
}
