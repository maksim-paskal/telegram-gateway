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
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//nolint:gochecknoglobals
var (
	gitVersion string = "dev"
	buildTime  string
	domains    map[string]ConfigDomains
)

func main() {
	flag.Parse()

	var err error

	logLevel, err := log.ParseLevel(*appConfig.logLevel)
	if err != nil {
		log.Panic(err)
	}

	log.SetLevel(logLevel)

	if logLevel == log.DebugLevel {
		log.SetReportCaller(true)
	}

	log.Infof("Starting telegram-gateway %s", appConfig.Version)

	// load config file
	yamlFile, err := ioutil.ReadFile(*appConfig.configFileName)
	if err != nil {
		log.Fatalf("error in reading config %s, %v", *appConfig.configFileName, err)
	}

	log.Debugf("using config file:\n%s", string(yamlFile))

	config := Config{}
	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		log.Fatal("error in Unmarshal", err)
	}

	domains = make(map[string]ConfigDomains)

	for _, domain := range config.Domains {
		bot, err := tgbotapi.NewBotAPI(domain.Token)
		if err != nil {
			log.Panicf("error connecting to bot %s, %v", domain.Name, err)
		}

		log.Printf("Authorized on account %s", bot.Self.UserName)

		domain.bot = bot
		domains[domain.Name] = domain

		if log.GetLevel() <= log.DebugLevel {
			log.Debug("add debug to bot")

			bot.Debug = true
		}
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

	log.Printf("Staring server on port %d", *appConfig.port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", *appConfig.port), router)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func startChatServer() {
	log.Info("Staring ChatServer")

	domain := domains[DomainDefault]
	u := tgbotapi.NewUpdate(0)

	u.Timeout = 60

	updates, err := domain.bot.GetUpdatesChan(u)
	if err != nil {
		log.Error(err)
	}

	log.Debug("range updates")

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("update.Message.Chat.ID=%d", update.Message.Chat.ID))

		msg.ReplyToMessageID = update.Message.MessageID

		_, err := domain.bot.Send(msg)
		if err != nil {
			log.Error(err)
		}
	}
}
