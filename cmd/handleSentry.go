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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type sentryStructEvent struct {
	Title    string            `json:"title"`
	Release  string            `json:"release"`
	Tags     [][]string        `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

type sentryStruct struct {
	Project string            `json:"project"`
	URL     string            `json:"url"`
	Event   sentryStructEvent `json:"event"`
}

func handleSentry(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	name := params["name"]
	if len(name) == 0 {
		name = *appConfig.defaultDomain
	}

	log.Debugf("name=%s", name)

	domain := domains[name]

	if len(domain.Name) == 0 {
		err := ErrNameNotFound
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Error()

		return
	}

	var err error

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Error()

		return
	}

	r.Body.Close()

	if log.GetLevel() >= log.DebugLevel {
		fmt.Println(string(bodyBytes)) //nolint:forbidigo
	}

	var data sentryStruct
	err = json.Unmarshal(bodyBytes, &data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Error()

		return
	}

	var message strings.Builder

	message.WriteString(formatTelegramMessage("Sentry.Project", data.Project))

	if len(data.Event.Release) > 0 {
		message.WriteString(formatTelegramMessage("Release", data.Event.Release))
	}

	message.WriteString(formatTelegramMessage("Title", data.Event.Title))

	alertLabels := make(map[string]string)

	for _, tag := range data.Event.Tags {
		message.WriteString(formatTelegramMessage(fmt.Sprintf("tag=\"%s\"", tag[0]), tag[1]))
		alertLabels[tag[0]] = tag[1]
	}

	msg := tgbotapi.NewMessage(domain.ChatID, message.String())
	msg.ParseMode = ParseModeMarkdown

	row := []tgbotapi.InlineKeyboardButton{}

	keyboard := tgbotapi.InlineKeyboardMarkup{}

	// add Open Sentry button
	btn1 := tgbotapi.NewInlineKeyboardButtonURL("Open Sentry", data.URL)
	row = append(row, btn1)

	// add extra buttons
	for _, button := range domain.SentryButtons {
		url, err := templateString(button.Value, alertLabels)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.WithError(err).Error()

			return
		}

		btn := tgbotapi.NewInlineKeyboardButtonURL(button.Name, url)
		row = append(row, btn)
	}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	msg.ReplyMarkup = keyboard

	_, err = domain.bot.Send(msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Error()

		return
	}

	_, err = w.Write([]byte("OK"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Error()
	}
}
