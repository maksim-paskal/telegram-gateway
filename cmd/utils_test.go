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
	"testing"
	"time"
)

func TestFormatTelegramMessage(t *testing.T) {
	t.Parallel()

	ans := formatTelegramMessage("a", "b")

	const right = "\n*a* `b`"

	if ans != right {
		t.Errorf("formatTelegramMessage = %s; want %s", ans, right)
	}
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	d, _ := time.ParseDuration("4h30m")

	const right = "4h30m0s"

	if ans := formatDuration(d); ans != right {
		t.Errorf("formatTelegramMessage = %s; want %s", ans, right)
	}
}

func TestTextTemplate(t *testing.T) {
	t.Parallel()

	data := make(map[string]string)

	data["test"] = "My Test Value"

	ans, err := templateString("test {{ .test | ToLower }}", data)
	if err != nil {
		t.Error(err)
	}

	if right := "test my test value"; ans != right {
		t.Errorf("templateString = %s; want %s", ans, right)
	}
}
