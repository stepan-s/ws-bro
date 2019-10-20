package endpoint

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func BindApi(users *hive.Users, apps *hive.Apps, pattern string, apiKey string, authKey string) {
	http.HandleFunc(pattern+"/user/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Auth") != apiKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		uid, err := strconv.ParseInt(r.URL.Query().Get("uid"), 10, 32)
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		users.SendMessage(uint32(uid), body)
	})

	http.HandleFunc(pattern+"/app/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Auth") != apiKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		aid, err := uuid.Parse(r.URL.Query().Get("uuid"))
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		apps.SendMessage(aid, body)
	})

	http.HandleFunc(pattern+"/sign-auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Auth") != apiKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		uid, err := strconv.ParseInt(r.URL.Query().Get("uid"), 10, 32)
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		now := time.Now().Unix()
		sign := SignAuth(uint32(uid), now, authKey)

		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		_, err2 := w.Write([]byte(fmt.Sprintf("uid=%d&ts=%d&sign=%s", uid, now, sign)))
		if err2 != nil {
			log.Error("Fail sign auth: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
