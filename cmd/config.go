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
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

type appConfigType struct {
	Version         string
	port            *int
	logLevel        *string
	chatServer      *bool
	chatToken       *string
	chatID          *int64
	alertManagerURL *string
	prometheusURL   *string
	clusterName     *string
}

//nolint:gochecknoglobals
var appConfig = appConfigType{
	Version: "1.0.4",
	port: kingpin.Flag(
		"server.port",
		"port",
	).Default("9090").Int(),
	logLevel: kingpin.Flag(
		"log.level",
		"log level",
	).Default("INFO").String(),
	chatServer: kingpin.Flag(
		"enableChatServer",
		"enableChatServer",
	).Default("false").Bool(),
	chatToken: kingpin.Flag(
		"chat.token",
		"chat.token",
	).Default(os.Getenv("CHAT_TOKEN")).String(),
	chatID: kingpin.Flag(
		"chat.id",
		"chat.id",
	).Default(os.Getenv("CHAT_ID")).Int64(),
	alertManagerURL: kingpin.Flag(
		"alertmanager.url",
		"alertmanager.url",
	).Default("https://alertmanager.paskal-dev.com").String(),
	prometheusURL: kingpin.Flag(
		"prometheus.url",
		"prometheus.url",
	).Default("https://prometheus.paskal-dev.com/alerts").String(),
	clusterName: kingpin.Flag(
		"cluster.name",
		"cluster.name",
	).String(),
}
