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

		users.SendEvent(hive.UserMessageEvent{Uid: uint32(uid), RawMessage: body})
	})

	http.HandleFunc(pattern+"/app/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Auth") != apiKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		aid, err := uuid.Parse(r.URL.Query().Get("aid"))
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

		apps.SendEvent(hive.AppMessageToEvent{Aid: aid, Uid: hive.SYSUID, RawMessage: body})
	})

	http.HandleFunc(pattern+"/user/sign-auth", func(w http.ResponseWriter, r *http.Request) {
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
		sign := SignUserAuth(uint32(uid), now, authKey)

		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		_, err2 := w.Write([]byte(fmt.Sprintf("uid=%d&ts=%d&sign=%s", uid, now, sign)))
		if err2 != nil {
			log.Error("Fail sign auth: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	http.HandleFunc(pattern+"/app/sign-auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Auth") != apiKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		aid, err := uuid.Parse(r.URL.Query().Get("aid"))
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		now := time.Now().Unix()
		sign := SignAppAuth(aid, now, authKey)

		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		_, err2 := w.Write([]byte(fmt.Sprintf("aid=%s&ts=%d&sign=%s", aid.String(), now, sign)))
		if err2 != nil {
			log.Error("Fail sign auth: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	http.HandleFunc(pattern+"/app/attach", func(w http.ResponseWriter, r *http.Request) {
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

		aid, err := uuid.Parse(r.URL.Query().Get("aid"))
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		attachMessage, err := hive.MessageUserAttachedPack(&hive.MessageUserAttached{
			Action: hive.ACTION_ATTACHED,
			List:   []uuid.UUID{aid},
		})
		if err != nil {
			log.Error("Fail pack: %v, aid:%v", err, aid)
		} else {
			users.SendEvent(hive.UserMessageEvent{Uid: uint32(uid), RawMessage: attachMessage});
		}

		apps.UpdateUids(hive.AppUidsEvent{
			Cmd: hive.ADD,
			Aid:  aid,
			Uids: []uint32{uint32(uid)},
		})
	})

	http.HandleFunc(pattern+"/app/detach", func(w http.ResponseWriter, r *http.Request) {
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

		aid, err := uuid.Parse(r.URL.Query().Get("aid"))
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		apps.UpdateUids(hive.AppUidsEvent{
			Cmd: hive.REMOVE,
			Aid:  aid,
			Uids: []uint32{uint32(uid)},
		})

		detachMessage, err := hive.MessageUserAttachedPack(&hive.MessageUserAttached{
			Action: hive.ACTION_DETACHED,
			List:   []uuid.UUID{aid},
		})
		if err != nil {
			log.Error("Fail pack: %v, aid:%v", err, aid)
		} else {
			users.SendEvent(hive.UserMessageEvent{Uid: uint32(uid), RawMessage: detachMessage});
		}
	})
}
