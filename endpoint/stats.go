package endpoint

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"net/http"
)

type UsersStats interface {
	GetData() hive.UsersStatsData
}

type AppsStats interface {
	GetData() hive.AppsStatsData
}

type Stats struct {
	Users hive.UsersStatsData
	Apps  hive.AppsStatsData
}

func BindStats(users UsersStats, apps AppsStats, pattern string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		stats := Stats{
			Users: users.GetData(),
			Apps:  apps.GetData(),
		}
		w.Header().Add("Content-Type", "application/json")
		info, err := json.Marshal(stats)
		if err != nil {
			log.Error("Fail bind stats: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(info)
		if err != nil {
			log.Error("Fail bind stats: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func BindMetrics(u UsersStats, a AppsStats, pattern string) {
	users := u.GetData()
	apps := a.GetData()
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_users_total_connections_accepted",
			Help: "The total number of accepted user connections",
		}, func() float64 {
			return float64(users.TotalConnectionsAccepted)
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "wsbro_users_current_connections",
			Help: "The current number of user connections",
		}, func() float64 {
			return float64(users.CurrentConnections)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_users_total_users_connected",
			Help: "The total number of accepted users",
		}, func() float64 {
			return float64(users.TotalUsersConnected)
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "wsbro_users_current_users_connected",
			Help: "The current number of users",
		}, func() float64 {
			return float64(users.CurrentUsersConnected)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_users_messages_received",
			Help: "The total number of received messages from user",
		}, func() float64 {
			return float64(users.MessagesReceived)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_users_messages_transmitted",
			Help: "The total number of transmitted messages to user",
		}, func() float64 {
			return float64(users.MessagesTransmitted)
		}))

	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_apps_total_connections_accepted",
			Help: "The total number of accepted app connections",
		}, func() float64 {
			return float64(apps.TotalConnectionsAccepted)
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "wsbro_apps_current_connections",
			Help: "The current number of app connections",
		}, func() float64 {
			return float64(apps.CurrentConnections)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_apps_messages_received",
			Help: "The total number of received messages from app",
		}, func() float64 {
			return float64(apps.MessagesReceived)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_apps_messages_transmitted",
			Help: "The total number of transmitted messages to app",
		}, func() float64 {
			return float64(apps.MessagesTransmitted)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_apps_total_disconnects",
			Help: "The total number apps disconnects",
		}, func() float64 {
			return float64(apps.TotalDisconnects)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_apps_total_reconnects",
			Help: "The total number apps reconnects",
		}, func() float64 {
			return float64(apps.TotalReconnects)
		}))

	http.Handle(pattern, promhttp.Handler())
}
