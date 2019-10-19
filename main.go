package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/stepan-s/ws-bro/endpoint"
	"github.com/stepan-s/ws-bro/hive"
	"github.com/stepan-s/ws-bro/log"
	"net/http"
	"os"
	"os/signal"
	"time"
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

	if *authKey == "" {
		// Create auth key id empty
		hash := sha256.New()
		hash.Write([]byte(fmt.Sprintf("%s%d", apiKey, time.Now().Unix())))
		key := fmt.Sprintf("%x", hash.Sum(nil))
		authKey = &key
	}

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

	srv := &http.Server{Addr: *addr}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		err := srv.Shutdown(context.Background());
		if err != nil {
			// Error from closing listeners, or context timeout:
			log.Error("Server shutdown: %v", err)
		}
	}()

	err := srv.ListenAndServeTLS(*certFilename, *privKeyFilename)
	if err != http.ErrServerClosed {
		log.Emergency("Server error: %v", err)
	}
	log.Info("Stopped")
}
