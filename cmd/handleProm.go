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
	"io"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/mux"
	template "github.com/prometheus/alertmanager/template"
	log "github.com/sirupsen/logrus"
)

func handleProm(w http.ResponseWriter, r *http.Request) {
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

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Error()

		return
	}

	r.Body.Close()

	if log.GetLevel() >= log.DebugLevel {
		fmt.Println(string(bodyBytes)) //nolint:forbidigo
	}

	data := template.Data{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var message strings.Builder

	message.WriteString(fmt.Sprintf("*status*: %s", strings.ToUpper(data.Status)))

	for _, extraLabels := range domain.ExtraLabels {
		message.WriteString(formatTelegramMessage(extraLabels.Name, extraLabels.Value))
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

	alertLabels := make(map[string]string)

	for i := range data.CommonAnnotations.Names() {
		name := data.CommonAnnotations.Names()[i]
		value := data.CommonAnnotations.Values()[i]
		message.WriteString(formatTelegramMessage(name, value))
		alertLabels[strings.ToLower(name)] = value
	}

	for i := range data.CommonLabels.Names() {
		name := data.CommonLabels.Names()[i]
		value := data.CommonLabels.Values()[i]
		message.WriteString(formatTelegramMessage(name, value))
		alertLabels[strings.ToLower(name)] = value
	}

	msg := tgbotapi.NewMessage(domain.ChatID, message.String())

	msg.ParseMode = ParseModeMarkdown

	if strings.ToUpper(data.Status) != "RESOLVED" && len(domain.PrometheusButtons) > 0 {
		row := []tgbotapi.InlineKeyboardButton{}

		keyboard := tgbotapi.InlineKeyboardMarkup{}

		for _, button := range domain.PrometheusButtons {
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
	}

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
