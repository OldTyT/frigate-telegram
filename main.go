package main

import (
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/oldtyt/frigate-telegram/internal/config"
	"github.com/oldtyt/frigate-telegram/internal/frigate"
	"github.com/oldtyt/frigate-telegram/internal/log"
	"github.com/oldtyt/frigate-telegram/internal/redis"
	"github.com/oldtyt/frigate-telegram/internal/telegram"
)

// FrigateEvents is frigate events struct
var FrigateEvents frigate.EventsStruct

// FrigateEvent is frigate event struct
var FrigateEvent frigate.EventStruct

func main() {
	// Initializing logger
	log.LogFunc()
	// Get config
	conf := config.New()

	// Prepare startup msg
	startupMsg := "Starting frigate-telegram.\n"
	startupMsg += "Frigate URL:  " + conf.FrigateURL + "\n"
	log.Info.Println(startupMsg)

	// Initializing telegram bot
	bot, err := tgbotapi.NewBotAPI(conf.TelegramBotToken)
	if err != nil {
		log.Error.Fatalln("Error initalizing telegram bot: " + err.Error())
	}
	bot.Debug = conf.Debug
	log.Info.Println("Authorized on account " + bot.Self.UserName)

	// Send startup msg.
	_, errmsg := bot.Send(tgbotapi.NewMessage(conf.TelegramChatID, startupMsg))
	if errmsg != nil {
		log.Error.Println(errmsg.Error())
	}

	// Starting ping command handler(healthcheck)
	go telegram.ChatBot(bot, conf)

	FrigateEventsURL := conf.FrigateURL + "/api/events"

	if conf.SendTextEvent {
		go frigate.NotifyEvents(bot, FrigateEventsURL)
	}
	// Starting loop for getting events from Frigate
	for {
		if redis.GetStateSendEvent() {
			FrigateEvents := frigate.GetEvents(FrigateEventsURL, bot, true)
			frigate.ParseEvents(FrigateEvents, bot, false)
		} else {
			log.Debug.Println("Skiping send events.")
		}
		time.Sleep(time.Duration(conf.SleepTime) * time.Second)
		log.Debug.Println("Sleeping for " + strconv.Itoa(conf.SleepTime) + " seconds.")
	}
}
