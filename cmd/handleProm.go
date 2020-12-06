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
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	template "github.com/prometheus/alertmanager/template"
	log "github.com/sirupsen/logrus"
)

func handleProm(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var err error

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)

		return
	}

	if log.GetLevel() == log.DebugLevel {
		log.Debug(string(bodyBytes))
	}

	data := template.Data{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var message strings.Builder

	message.WriteString(fmt.Sprintf("*status*: %s", strings.ToUpper(data.Status)))

	if len(*appConfig.clusterName) > 0 {
		message.WriteString(formatTelegramMessage("Cluster", *appConfig.clusterName))
	}

	if len(data.Alerts) > 0 {
		alert := data.Alerts[0]
		duration := alert.EndsAt.Sub(alert.StartsAt)

		if alert.EndsAt.IsZero() {
			duration = time.Since(alert.StartsAt)
		}

		if duration.Minutes() >= 1 {
			message.WriteString(formatTelegramMessage("Duration", formatDuration(duration)))
		}
	}

	if len(data.Alerts) > 1 {
		message.WriteString(formatTelegramMessage("Alert Count", fmt.Sprintf("%d", len(data.Alerts))))
	}

	for i := range data.CommonAnnotations.Names() {
		name := data.CommonAnnotations.Names()[i]
		value := data.CommonAnnotations.Values()[i]
		message.WriteString(formatTelegramMessage(name, value))
	}

	for i := range data.CommonLabels.Names() {
		name := data.CommonLabels.Names()[i]
		value := data.CommonLabels.Values()[i]
		message.WriteString(formatTelegramMessage(name, value))
	}

	msg := tgbotapi.NewMessage(*appConfig.chatID, message.String())

	msg.ParseMode = ParseModeMarkdown

	if strings.ToUpper(data.Status) != "RESOLVED" {
		var row []tgbotapi.InlineKeyboardButton

		keyboard := tgbotapi.InlineKeyboardMarkup{}
		btn1 := tgbotapi.NewInlineKeyboardButtonURL("Prometheus", *appConfig.prometheusURL)
		row = append(row, btn1)

		btn2 := tgbotapi.NewInlineKeyboardButtonURL("AlertManager", *appConfig.alertManagerURL)
		row = append(row, btn2)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

		msg.ReplyMarkup = keyboard
	}

	_, err = bot.Send(msg)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)

		return
	}

	_, err = w.Write([]byte("OK"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
	}
}
