package endpoint

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"net/http"
)

type Stats struct {
	Users *hive.UsersStats
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
		_, err = w.Write(info)
		if err != nil {
			log.Error("Fail bind stats: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func BindMetrics(users *hive.Users, apps *hive.Apps, pattern string) {
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_users_total_connections_accepted",
			Help: "The total number of accepted user connections",
		}, func() float64 {
			return float64(users.Stats.TotalConnectionsAccepted)
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "wsbro_users_current_connections",
			Help: "The current number of user connections",
		}, func() float64 {
			return float64(users.Stats.CurrentConnections)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_users_total_users_connected",
			Help: "The total number of accepted users",
		}, func() float64 {
			return float64(users.Stats.TotalUsersConnected)
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "wsbro_users_current_users_connected",
			Help: "The current number of users",
		}, func() float64 {
			return float64(users.Stats.CurrentUsersConnected)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_users_messages_received",
			Help: "The total number of received messages from user",
		}, func() float64 {
			return float64(users.Stats.MessagesReceived)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_users_messages_transmitted",
			Help: "The total number of transmitted messages to user",
		}, func() float64 {
			return float64(users.Stats.MessagesTransmitted)
		}))

	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_apps_total_connections_accepted",
			Help: "The total number of accepted app connections",
		}, func() float64 {
			return float64(apps.Stats.TotalConnectionsAccepted)
		}))
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "wsbro_apps_current_connections",
			Help: "The current number of app connections",
		}, func() float64 {
			return float64(apps.Stats.CurrentConnections)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_apps_messages_received",
			Help: "The total number of received messages from app",
		}, func() float64 {
			return float64(apps.Stats.MessagesReceived)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_apps_messages_transmitted",
			Help: "The total number of transmitted messages to app",
		}, func() float64 {
			return float64(apps.Stats.MessagesTransmitted)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_apps_total_disconnects",
			Help: "The total number apps disconnects",
		}, func() float64 {
			return float64(apps.Stats.TotalDisconnects)
		}))
	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "wsbro_apps_total_reconnects",
			Help: "The total number apps reconnects",
		}, func() float64 {
			return float64(apps.Stats.TotalReconnects)
		}))

	http.Handle(pattern, promhttp.Handler())
}
