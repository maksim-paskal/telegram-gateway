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

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ConfigDomains struct {
	Name            string `yaml:"name"`
	Token           string `yaml:"token"`
	ChatID          int64  `yaml:"chatID"`
	ClusterName     string `yaml:"clusterName"`
	PrometheusURL   string `yaml:"prometheusURL"`
	AlertManagerURL string `yaml:"alertManagerURL"`
	bot             *tgbotapi.BotAPI
}

type Config struct {
	Domains []ConfigDomains `yaml:"domains"`
}

type appConfigType struct {
	Version        string
	showVersion    *bool
	port           *int
	logLevel       *string
	chatServer     *bool
	configFileName *string
}

//nolint:gochecknoglobals
var appConfig = appConfigType{
	Version:        fmt.Sprintf("%s-%s", gitVersion, buildTime),
	showVersion:    flag.Bool("version", false, "show version"),
	port:           flag.Int("server.port", 9090, "server port"),
	logLevel:       flag.String("log.level", "INFO", "log level"),
	chatServer:     flag.Bool("enableChatServer", false, "enableChatServer"),
	configFileName: flag.String("config", "config.yaml", "config yaml path"),
}
