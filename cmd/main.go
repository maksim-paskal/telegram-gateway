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
	"fmt"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

//nolint:gochecknoglobals
var (
	buildTime string
	bot       *tgbotapi.BotAPI
)

func main() {
	log.Infof("Starting telegram-gateway %s-%s", appConfig.Version, buildTime)

	kingpin.Version(appConfig.Version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var err error

	logLevel, err := log.ParseLevel(*appConfig.logLevel)
	if err != nil {
		log.Panic(err)
	}

	log.SetLevel(logLevel)

	bot, err = tgbotapi.NewBotAPI(*appConfig.chatToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	if *appConfig.chatServer {
		log.Info("Staring ChatServer")

		u := tgbotapi.NewUpdate(0)

		u.Timeout = 60

		updates, _ := bot.GetUpdatesChan(u)

		for update := range updates {
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("update.Message.Chat.ID=%d", update.Message.Chat.ID))

			msg.ReplyToMessageID = update.Message.MessageID

			_, err := bot.Send(msg)
			if err != nil {
				log.Error(err)
			}
		}
	}

	http.HandleFunc("/prom", handleProm)
	http.HandleFunc("/sentry", handleSentry)
	http.HandleFunc("/message", handleMessage)
	http.HandleFunc("/test", handleTest)
	log.Printf("Staring server on port %d", *appConfig.port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", *appConfig.port), nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
