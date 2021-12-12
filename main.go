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
	var userAuthSignTTL = flag.Int64("user-auth-sign-ttl", endpoint.UserAuthSignTTL, "user auth sign ttl in seconds")
	var appAuthSignTTL = flag.Int64("app-auth-sign-ttl", endpoint.AppAuthSignTTL, "app auth sign ttl in seconds")
	var certFilename = flag.String("cert-file", "", "certificate path")
	var privKeyFilename = flag.String("key-file", "", "private key path")
	var apiKey = flag.String("api-key", "", "api key")
	var uidsApiUrl = flag.String("uids-api-url", "", "get uids by aid")
	var devPageTemplate = flag.String("dev-page-template", "", "dev page template path")
	var logLevel = flag.Int64("log-level", log.DEBUG, "log level")
	flag.Parse()

	var logLevelValue = uint8(*logLevel)
	if logLevelValue < log.NONE {
		logLevelValue = log.NONE
	}
	if logLevelValue > log.DEBUG {
		logLevelValue = log.DEBUG
	}
	log.Init(os.Stdout, logLevelValue)
	log.Info("Starting")

	log.Info("Options:")
	log.Info("  addr: %v", *addr)
	log.Info("  allowed-origins: %v", *allowedOrigins)
	if *authKey != "" {
		log.Info("  auth-key: set")
	} else {
		log.Info("  auth-key: not set (random used)")
	}
	log.Info("  user-auth-sign-ttl: %v", *userAuthSignTTL)
	log.Info("  app-auth-sign-ttl: %v", *appAuthSignTTL)
	log.Info("  cert-file: %v", *certFilename)
	log.Info("  key-file: %v", *privKeyFilename)
	if *apiKey != "" {
		log.Info("  api-key: set")
	} else {
		log.Info("  api-key: not set")
	}
	log.Info("  uids-api-url: %v", *uidsApiUrl)
	log.Info("  dev-page-template: %v", *devPageTemplate)
	log.Info("  log-level: %v, used: %v", *logLevel, logLevelValue)

	endpoint.UserAuthSignTTL = *userAuthSignTTL
	endpoint.AppAuthSignTTL = *appAuthSignTTL

	if *authKey == "" {
		// Create auth key id empty
		hash := sha256.New()
		hash.Write([]byte(fmt.Sprintf("%s%d", *apiKey, time.Now().Unix())))
		key := fmt.Sprintf("%x", hash.Sum(nil))
		authKey = &key
	}

	usersStats := hive.NewUsersStats()
	appsStats := hive.NewAppsStats()

	users := hive.NewUsers(usersStats)
	apps := hive.NewApps(*uidsApiUrl, appsStats)
	hive.RouterStart(users, apps)

	if len(*devPageTemplate) > 0 {
		log.Alert("Binding dev page handler - don't use in production - secrets leak!")
		endpoint.BindDevPage("/dev", *devPageTemplate, *apiKey)
	}
	endpoint.BindStats(usersStats, appsStats, "/stats")
	endpoint.BindMetrics(usersStats, appsStats, "/metrics")
	endpoint.BindApi(users, apps, "/api", *apiKey, *authKey)
	endpoint.BindUsers(users, "/bro", *allowedOrigins, *authKey)
	endpoint.BindApps(apps, "/app", *authKey)

	srv := &http.Server{Addr: *addr}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		err := srv.Shutdown(context.Background())
		if err != nil {
			// Error from closing listeners, or context timeout:
			log.Error("Server shutdown: %v", err)
		}
	}()

	err := srv.ListenAndServeTLS(*certFilename, *privKeyFilename)
	if err != http.ErrServerClosed {
		log.Emergency("Server error: %v", err)
		os.Exit(1)
	}
	log.Info("Stopped")
}
