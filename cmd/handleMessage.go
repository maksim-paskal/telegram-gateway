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
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func handleMessage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	name := params["name"]
	if len(name) == 0 {
		name = DomainDefault
	}

	log.Debugf("name=%s", name)

	domain := domains[name]

	if len(domain.Name) == 0 {
		err := ErrorNameNotFound
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)

		return
	}

	var err error

	var message strings.Builder

	for k, v := range r.URL.Query() {
		if k != "url" && k != "url.title" && len(v[0]) > 0 {
			message.WriteString(formatTelegramMessage(k, v[0]))
		}
	}

	msg := tgbotapi.NewMessage(domain.ChatID, message.String())
	msg.ParseMode = ParseModeMarkdown

	if len(r.URL.Query()["url"]) > 0 {
		keyboard := tgbotapi.InlineKeyboardMarkup{}

		var row []tgbotapi.InlineKeyboardButton

		caption := "Open"

		if len(r.URL.Query()["url.title"]) > 0 && len(r.URL.Query()["url.title"][0]) > 0 {
			caption = r.URL.Query()["url.title"][0]
		}

		btn1 := tgbotapi.NewInlineKeyboardButtonURL(caption, r.URL.Query()["url"][0])
		row = append(row, btn1)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
		msg.ReplyMarkup = keyboard
	}

	_, err = domain.bot.Send(msg)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
	}

	_, err = w.Write([]byte("OK"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
	}
}
