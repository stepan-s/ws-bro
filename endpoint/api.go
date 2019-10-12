package endpoint

import (
	"fmt"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func BindApi(users *hive.Users, pattern string, apiKey string, authKey string) {
	http.HandleFunc(pattern+"/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Auth") != apiKey {
			w.WriteHeader(403)
			return
		}

		uid, err := strconv.ParseInt(r.URL.Query().Get("uid"), 10, 32)
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(400)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(400)
			return
		}

		users.SendMessage(uint32(uid), string(body))
	})

	http.HandleFunc(pattern+"/sign-auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Auth") != apiKey {
			w.WriteHeader(403)
			return
		}

		uid, err := strconv.ParseInt(r.URL.Query().Get("uid"), 10, 32)
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(400)
			return
		}

		now := time.Now().Unix()
		sign := SignAuth(uint32(uid), now, authKey)

		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		_, err2 := w.Write([]byte(fmt.Sprintf("ts=%d&sign=%s", now, sign)))
		if err2 != nil {
			log.Error("Fail sign auth: %v", err)
		}
	})
}