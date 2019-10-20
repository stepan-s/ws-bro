package endpoint

import (
	"encoding/json"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"net/http"
)

type Stats struct {
	Users *hive.Stats
	Apps *hive.AppsStats
}

func BindStats(users *hive.Users, apps *hive.Apps, pattern string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		stats := Stats{
			Users: &users.Stats,
			Apps:  &apps.Stats,
		}
		w.Header().Add("Content-Type", "application/json")
		info, err := json.Marshal(stats)
		if err != nil {
			log.Error("Fail bind stats: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write([]byte(info))
		if err != nil {
			log.Error("Fail bind stats: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
