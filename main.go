package main

import (
	"flag"
	"github.com/stepan-s/ws-bro/endpoint"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"net/http"
	"os"
)

func main() {
	var addr = flag.String("addr", "localhost:443", "http service address")
	var allowedOrigins = flag.String("allowed-origins", "", "allowed origins")
	var authKey = flag.String("auth-key", "", "auth key")
	var certFilename = flag.String("cert-file", "", "certificate path")
	var privKeyFilename = flag.String("key-file", "", "private key path")
	var apiKey = flag.String("api-key", "", "api key")
	var devPageTemplate = flag.String("dev-page-template", "devpage.html", "dev page template path")
	flag.Parse()

	log.Init(os.Stdout, log.DEBUG)
	log.Info("Starting")

	users := hive.NewUsers()
	go func() {
		for {
			msg := users.ReceiveMessage()
			log.Info("User:%d say:%s", msg.Uid, msg.Payload)
			users.SendMessage(msg.Uid, "Hi! User")
		}
	}()

	if len(*devPageTemplate) > 0 {
		endpoint.BindDevPage("/dev", *devPageTemplate, *apiKey)
	}
	endpoint.BindStats(users, "/stats")
	endpoint.BindApi(users, "/api", *apiKey, *authKey)
	endpoint.BindUsers(users, "/bro", *allowedOrigins, *authKey)

	err := http.ListenAndServeTLS(*addr, *certFilename, *privKeyFilename, nil)
	if err != nil {
		log.Emergency("Server error: %v", err)
	}
	log.Info("Stopped")
}
