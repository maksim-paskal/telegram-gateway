package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	template "github.com/prometheus/alertmanager/template"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	buildTime string
)

func formatTelegramMessage(name string, value string) string {
	return fmt.Sprintf("\n*%s*``` %s ```", name, value)
}
func formatDuration(d time.Duration) string {
	seconds := int64(d.Seconds()) % 60
	minutes := int64(d.Minutes()) % 60
	hours := int64(d.Hours()) % 24
	days := int64(d/(24*time.Hour)) % 365 % 7

	var duration strings.Builder

	if days > 0 {
		duration.WriteString(fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		duration.WriteString(fmt.Sprintf("%dh", hours))
	}
	duration.WriteString(fmt.Sprintf("%dm%ds", minutes, seconds))
	return duration.String()
}

func test(w http.ResponseWriter, r *http.Request) {
	msg := tgbotapi.NewMessage(*appConfig.chatID, "*test*")
	msg.ParseMode = "Markdown"
	_, err := bot.Send(msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
		return
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	log.Info(bodyString)

	w.Write([]byte("OK"))
}
func prom(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var err error

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
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
		message.WriteString(fmt.Sprintf("Cluster: %s", *appConfig.clusterName))
	}

	if len(data.Alerts) > 0 {
		alert := data.Alerts[0]
		duration := alert.EndsAt.Sub(alert.StartsAt)
		if alert.EndsAt.IsZero() {
			duration = time.Now().Sub(alert.StartsAt)
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
	msg.ParseMode = "Markdown"

	if strings.ToUpper(data.Status) != "RESOLVED" {
		keyboard := tgbotapi.InlineKeyboardMarkup{}
		var row []tgbotapi.InlineKeyboardButton
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
	w.Write([]byte("OK"))
}

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

func sentry(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var err error

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	if log.GetLevel() == log.DebugLevel {
		log.Debug(string(bodyBytes))
	}

	var data sentryStruct
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
		return
	}
	var message strings.Builder

	message.WriteString(formatTelegramMessage("Sentry.Project", data.Project))
	if len(data.Event.Release) > 0 {
		message.WriteString(formatTelegramMessage("Release", data.Event.Release))
	}
	message.WriteString(formatTelegramMessage("Title", data.Event.Title))

	for _, tag := range data.Event.Tags {
		message.WriteString(formatTelegramMessage(fmt.Sprintf("tag=\"%s\"", tag[0]), tag[1]))
	}

	msg := tgbotapi.NewMessage(*appConfig.chatID, message.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	var row []tgbotapi.InlineKeyboardButton
	btn1 := tgbotapi.NewInlineKeyboardButtonURL("Open Sentry", data.URL)
	row = append(row, btn1)

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	msg.ReplyMarkup = keyboard

	_, err = bot.Send(msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
		return
	}

	w.Write([]byte("OK"))
}

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

var appConfig = appConfigType{
	Version: "1.0.1",
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

var bot *tgbotapi.BotAPI

func main() {
	log.Infof("Starting telegram-gateway %s-%s", appConfig.Version, buildTime)

	kingpin.Version(appConfig.Version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var err error

	logLevel, err := log.ParseLevel(*appConfig.logLevel)

	if err != nil {
		log.Panic(err)
	}

	log.SetLevel(logLevel)

	bot, err = tgbotapi.NewBotAPI(*appConfig.chatToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	if *appConfig.chatServer {
		log.Info("Staring ChatServer")
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates, _ := bot.GetUpdatesChan(u)

		for update := range updates {
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("update.Message.Chat.ID=%d", update.Message.Chat.ID))
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}

	http.HandleFunc("/prom", prom)
	http.HandleFunc("/sentry", sentry)
	http.HandleFunc("/test", test)
	log.Printf("Staring server on port %d", *appConfig.port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", *appConfig.port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
