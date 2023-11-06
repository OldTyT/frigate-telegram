package main

import (
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/oldtyt/frigate-telegram/internal/config"
	"github.com/oldtyt/frigate-telegram/internal/log"
	"github.com/oldtyt/frigate-telegram/internal/frigate"
)


var FrigateEvents frigate.EventsStruct
var FrigateEvent frigate.EventStruct

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

	// Starting loop for getting events from Frigate
	for true {
		FrigateEvents := frigate.GetEvents(FrigateEventsURL)
		go frigate.ParseEvent(FrigateEvents, bot)

		time.Sleep(time.Duration(conf.SleepTime) * time.Second)
		log.Debug.Println("Sleeping for " + strconv.Itoa(conf.SleepTime) + " seconds.")
	}
}
