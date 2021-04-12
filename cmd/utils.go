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
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
)

func formatTelegramMessage(name string, value string) string {
	return fmt.Sprintf("\n*%s*``` %s ```", name, value)
}

//nolint:gomnd
func formatDuration(d time.Duration) string {
	seconds := int64(d.Seconds()) % 60
	minutes := int64(d.Minutes()) % 60
	days := int64(d/(24*time.Hour)) % 365 % 7

	var duration strings.Builder

	if days > 0 {
		duration.WriteString(fmt.Sprintf("%dd", days))
	}

	if hours := int64(d.Hours()) % 24; hours > 0 {
		duration.WriteString(fmt.Sprintf("%dh", hours))
	}

	duration.WriteString(fmt.Sprintf("%dm%ds", minutes, seconds))

	return duration.String()
}

func templateString(text string, v interface{}) (string, error) {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"ToLower": strings.ToLower,
	}

	tmpl, err := template.New("buttons").Funcs(funcMap).Parse(text)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer

	err = tmpl.Execute(&tpl, v)
	if err != nil {
		return "", err
	}

	log.Debugf("text=%s,formated=%s", text, tpl.String())

	return tpl.String(), nil
}
