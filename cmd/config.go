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

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ExtraLabels struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type TelegramButton struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type ConfigDomains struct {
	// name of domain
	Name string `yaml:"name"`
	// telegram token
	Token string `yaml:"token"`
	// telegram chatID
	ChatID int64 `yaml:"chatID"` //nolint:tagliatelle
	// add labels to alert
	ExtraLabels []ExtraLabels `yaml:"extraLabels"`
	// add buttons to prometheus alerts
	PrometheusButtons []TelegramButton `yaml:"prometheusButtons"`
	// add buttons to prometheus alerts
	SentryButtons []TelegramButton `yaml:"sentryButtons"`
	// pointer to BotAPI
	bot *tgbotapi.BotAPI
}

type Config struct {
	// defaults values to all domains
	Defaults ConfigDomains `yaml:"defaults"`
	// telegram domains
	Domains []ConfigDomains `yaml:"domains"`
}

// add defaults values to domains.
func (config *Config) fillDefaults() {
	for i, domain := range config.Domains {
		if len(domain.ExtraLabels) == 0 {
			config.Domains[i].ExtraLabels = config.Defaults.ExtraLabels
		}

		if len(domain.Token) == 0 {
			config.Domains[i].Token = config.Defaults.Token
		}

		if len(domain.PrometheusButtons) == 0 {
			config.Domains[i].PrometheusButtons = config.Defaults.PrometheusButtons
		}

		if len(domain.SentryButtons) == 0 {
			config.Domains[i].SentryButtons = config.Defaults.SentryButtons
		}
	}
}

type appConfigType struct {
	Version        string
	showVersion    *bool
	port           *int
	logLevel       *string
	logPretty      *bool
	chatServer     *bool
	defaultDomain  *string
	configFileName *string
}

//nolint:gochecknoglobals
var appConfig = appConfigType{
	Version:        gitVersion,
	showVersion:    flag.Bool("version", false, "show version"),
	port:           flag.Int("server.port", defaultPort, "server port"),
	logLevel:       flag.String("log.level", "INFO", "log level"),
	logPretty:      flag.Bool("log.pretty", false, "log in pretty format"),
	chatServer:     flag.Bool("enableChatServer", false, "enableChatServer"),
	defaultDomain:  flag.String("defaultDomain", "default", "domain for default"),
	configFileName: flag.String("config", "config.yaml", "config yaml path"),
}
