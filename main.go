package main

import (
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/oldtyt/frigate-telegram/internal/config"
	"github.com/oldtyt/frigate-telegram/internal/frigate"
	"github.com/oldtyt/frigate-telegram/internal/log"
)

var FrigateEvents frigate.EventsStruct
var FrigateEvent frigate.EventStruct

func PongBot(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "help":
			msg.Text = "I understand /ping."
		case "ping":
			msg.Text = "pong"
		case "pong":
			msg.Text = "ping"
		case "status":
			msg.Text = "I'm ok."
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Error.Fatalln("Error sending message: " + err.Error())
		}
	}
}

func main() {
	// Initializing logger
	log.LogFunc()
	// Get config
	conf := config.New()
	log.Info.Println("Starting frigate-telegram.")
	log.Info.Println("Version:      " + config.Version)
	log.Info.Println("Frigate URL:  " + conf.FrigateURL)

	// Initializing telegram bot
	bot, err := tgbotapi.NewBotAPI(conf.TelegramBotToken)
	if err != nil {
		log.Error.Fatalln("Error initalizing telegram bot: " + err.Error())
	}
	bot.Debug = conf.Debug
	log.Info.Println("Authorized on account " + bot.Self.UserName)

	FrigateEventsURL := conf.FrigateURL + "/api/events"

	// Starting ping command handler(healthcheck)
	go PongBot(bot)

	// Starting loop for getting events from Frigate
	for true {
		FrigateEvents := frigate.GetEvents(FrigateEventsURL, bot)
		go frigate.ParseEvents(FrigateEvents, bot)

		time.Sleep(time.Duration(conf.SleepTime) * time.Second)
		log.Debug.Println("Sleeping for " + strconv.Itoa(conf.SleepTime) + " seconds.")
	}
}
