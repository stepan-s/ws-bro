package endpoint

import (
	"fmt"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"net/http"
)

func BindStats(users *hive.Users, pattern string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		info := fmt.Sprintf(
			`Current users connected: %d
Current connections: %d
Total users connected: %d
Total connections accepted: %d`,
			users.Stats.CurrentUsersConnected,
			users.Stats.CurrentConnections,
			users.Stats.TotalUsersConnected,
			users.Stats.TotalConnectionsAccepted)
		_, err := w.Write([]byte(info))
		if err != nil {
			log.Error("Fail bind stats: %v", err)
		}
	})
}
