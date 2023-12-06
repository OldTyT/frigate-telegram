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

	// Prepare startup msg
	startup_msg := "Starting frigate-telegram.\n"
	startup_msg += "Version:      " + config.Version + "\n"
	startup_msg += "Frigate URL:  " + conf.FrigateURL + "\n"
	log.Info.Println(startup_msg)

	// Initializing telegram bot
	bot, err := tgbotapi.NewBotAPI(conf.TelegramBotToken)
	if err != nil {
		log.Error.Fatalln("Error initalizing telegram bot: " + err.Error())
	}
	bot.Debug = conf.Debug
	log.Info.Println("Authorized on account " + bot.Self.UserName)

	// Send startup msg.
	_, errmsg := bot.Send(tgbotapi.NewMessage(conf.TelegramChatID, startup_msg))
	if errmsg != nil {
		log.Error.Println(errmsg.Error())
	}

	// Starting ping command handler(healthcheck)
	go PongBot(bot)

	FrigateEventsURL := conf.FrigateURL + "/api/events"

	go frigate.NotifyEvents(bot, FrigateEventsURL)
	// Starting loop for getting events from Frigate
	for true {
		FrigateEvents := frigate.GetEvents(FrigateEventsURL, bot, true)
		frigate.ParseEvents(FrigateEvents, bot, false)
		time.Sleep(time.Duration(conf.SleepTime) * time.Second)
		log.Debug.Println("Sleeping for " + strconv.Itoa(conf.SleepTime) + " seconds.")
	}
}
