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
	"io/ioutil"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func handleTest(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	name := params["name"]
	if len(name) == 0 {
		name = *appConfig.defaultDomain
	}

	domain := domains[name]

	if len(domain.Name) == 0 {
		err := ErrorNameNotFound
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Error()

		return
	}

	msg := tgbotapi.NewMessage(domain.ChatID, "*test*")

	msg.ParseMode = ParseModeMarkdown

	_, err := domain.bot.Send(msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Error()
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Error()

		return
	}

	r.Body.Close()

	bodyString := string(bodyBytes)
	log.Info(bodyString)

	_, err = w.Write([]byte("OK"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).Error()
	}
}
