/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	pprof "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/mux"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//nolint:gochecknoglobals
var (
	gitVersion = "dev"
	domains    map[string]ConfigDomains
)

const (
	serverReadTimeout    = 5 * time.Second
	serverRequestTimeout = 60 * time.Second
	serverWriteTimeout   = 70 * time.Second
)

func main() {
	flag.Parse()

	if *appConfig.showVersion {
		fmt.Println(appConfig.Version) //nolint:forbidigo
		os.Exit(0)
	}

	var err error

	logLevel, err := log.ParseLevel(*appConfig.logLevel)
	if err != nil {
		log.WithError(err).Fatal()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.DeferExitHandler(func() {
		cancel()
		time.Sleep(*appConfig.gracefulStop)
	})

	interruptionSignal := make(chan os.Signal, 1)
	signal.Notify(interruptionSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case <-ctx.Done():
		case <-interruptionSignal:
			log.Info("Got interruption signal")
			cancel()
		}
	}()

	hook, err := logrushooksentry.NewHook(logrushooksentry.Options{
		Release: appConfig.Version,
	})
	if err != nil {
		log.WithError(err).Fatal()
	}

	log.AddHook(hook)

	log.SetLevel(logLevel)
	log.SetReportCaller(true)

	if !*appConfig.logPretty {
		log.SetFormatter(&log.JSONFormatter{})
	}

	if logLevel == log.DebugLevel {
		log.SetReportCaller(true)
	}

	log.Infof("Starting telegram-gateway %s", appConfig.Version)

	// load config file
	yamlFile, err := os.ReadFile(*appConfig.configFileName)
	if err != nil {
		log.WithError(err).Fatalf("error in reading config %s", *appConfig.configFileName)
	}

	config := Config{}
	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		log.WithError(err).Fatal("error in Unmarshal")
	}

	config.fillDefaults()

	if log.GetLevel() >= log.DebugLevel {
		configYaml, err := yaml.Marshal(config)
		if err != nil {
			log.WithError(err).Error()
		}

		log.Debug("using config\n", string(configYaml))
	}

	err = tgbotapi.SetLogger(&BotLogger{})

	if err != nil {
		log.WithError(err).Fatal("error in setting logger")
	}

	domains = make(map[string]ConfigDomains)

	for _, domain := range config.Domains {
		bot, err := tgbotapi.NewBotAPI(domain.Token)
		if err != nil {
			log.WithError(err).Fatalf("[domain=%s] error connecting to bot", domain.Name)
		}

		log.Printf("[domain=%s] Authorized on account %s", domain.Name, bot.Self.UserName)

		domain.bot = bot
		domains[domain.Name] = domain

		if log.GetLevel() >= log.DebugLevel {
			log.Debugf("[domain=%s] add debug to bot", domain.Name)

			bot.Debug = true
		}
	}

	if len(domains[*appConfig.defaultDomain].Name) == 0 {
		log.WithError(err).Fatalf("in configuration has no default domain (%s)", *appConfig.defaultDomain)
	}

	if *appConfig.chatServer {
		startChatServer()
	}

	router := mux.NewRouter()
	router.HandleFunc("/{name}/prom", handleProm)
	router.HandleFunc("/{name}/sentry", handleSentry)
	router.HandleFunc("/{name}/message", handleMessage)
	router.HandleFunc("/{name}/test", handleTest)
	// default routes
	router.HandleFunc("/prom", handleProm)
	router.HandleFunc("/sentry", handleSentry)
	router.HandleFunc("/message", handleMessage)
	router.HandleFunc("/test", handleTest)
	router.HandleFunc("/healthz", handleHealthz)

	// pprof
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))

	log.Printf("Staring server on %s", *appConfig.address)

	timeoutMessage := fmt.Sprintf("Server timeout after %s", serverRequestTimeout)

	server := &http.Server{
		Addr:         *appConfig.address,
		Handler:      http.TimeoutHandler(router, serverRequestTimeout, timeoutMessage),
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), *appConfig.gracefulStop)
		defer shutdownCancel()

		_ = server.Shutdown(shutdownCtx) //nolint:contextcheck
	}()

	err = server.ListenAndServe()
	if err != nil {
		log.WithError(err).Fatal("ListenAndServe")
	}
}

func startChatServer() {
	log.Info("Staring ChatServer")

	domain := domains[*appConfig.defaultDomain]
	u := tgbotapi.NewUpdate(0)

	u.Timeout = 60

	updates := domain.bot.GetUpdatesChan(u)

	log.Debug("range updates")

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("update.Message.Chat.ID=%d", update.Message.Chat.ID))

		msg.ReplyToMessageID = update.Message.MessageID

		_, err := domain.bot.Send(msg)
		if err != nil {
			log.WithError(err).Fatal()
		}
	}
}
