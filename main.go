package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/oldtyt/frigate-telegram/internal/config"
	"github.com/oldtyt/frigate-telegram/internal/log"
)

type FrigateEventsStruct []struct {
	Area               any     `json:"area"`
	Box                any     `json:"box"`
	Camera             string  `json:"camera"`
	EndTime            float64 `json:"end_time"`
	FalsePositive      any     `json:"false_positive"`
	HasClip            bool    `json:"has_clip"`
	HasSnapshot        bool    `json:"has_snapshot"`
	ID                 string  `json:"id"`
	Label              string  `json:"label"`
	PlusID             any     `json:"plus_id"`
	Ratio              any     `json:"ratio"`
	Region             any     `json:"region"`
	RetainIndefinitely bool    `json:"retain_indefinitely"`
	StartTime          float64 `json:"start_time"`
	SubLabel           any     `json:"sub_label"`
	Thumbnail          string  `json:"thumbnail"`
	TopScore           float64 `json:"top_score"`
	Zones              []any   `json:"zones"`
}

type FrigateEventStruct struct {
	Area               any     `json:"area"`
	Box                any     `json:"box"`
	Camera             string  `json:"camera"`
	EndTime            float64 `json:"end_time"`
	FalsePositive      any     `json:"false_positive"`
	HasClip            bool    `json:"has_clip"`
	HasSnapshot        bool    `json:"has_snapshot"`
	ID                 string  `json:"id"`
	Label              string  `json:"label"`
	PlusID             any     `json:"plus_id"`
	Ratio              any     `json:"ratio"`
	Region             any     `json:"region"`
	RetainIndefinitely bool    `json:"retain_indefinitely"`
	StartTime          float64 `json:"start_time"`
	SubLabel           any     `json:"sub_label"`
	Thumbnail          string  `json:"thumbnail"`
	TopScore           float64 `json:"top_score"`
	Zones              []any   `json:"zones"`
}

var FrigateEvents FrigateEventsStruct
var FrigateEvent FrigateEventStruct
var FrigateAfter float64 = 1

func GetEvents(FrigateURL string) FrigateEventsStruct {
	conf := config.New()

	FrigateURL = FrigateURL + "?limit=" + strconv.Itoa(conf.FrigateEventLimit)
	if FrigateAfter > 0 {
		FrigateURL = FrigateURL + fmt.Sprintf("&after=%f", FrigateAfter)
	}

	log.Debug.Println("Geting events from Frigate via URL: " + FrigateURL)

	// Request to Frigate
	resp, err := http.Get(FrigateURL)
	if err != nil {
		log.Error.Fatalln("Error get events fron Frigate, error: " + err.Error())

	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != 200 {
		log.Error.Fatalln("Response status != 200, when getting events from Frigate.\nExit.")
	}

	// Read data from response
	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error.Fatalln("Can't read JSON: " + err.Error())
	}

	// Parse data from JSON to struct
	err1 := json.Unmarshal(byteValue, &FrigateEvents)
	if err1 != nil {
		log.Error.Println("Error unmarshal json: " + err1.Error())
		if e, ok := err.(*json.SyntaxError); ok {
			log.Info.Println("syntax error at byte offset " + strconv.Itoa(int(e.Offset)))
		}
		log.Info.Println("Exit.")
	}

	// Return Events
	return FrigateEvents
}

func SaveThumbnail(Thumbnail string) string {
	dec, err := base64.StdEncoding.DecodeString(Thumbnail)
	if err != nil {
		log.Error.Println("Error when base64 string decode: " + err.Error())

	}
	filename := "/tmp/" + strconv.FormatInt(int64(rand.Intn(10000000000)), 10) + ".jpg"
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := f.Write(dec); err != nil {
		panic(err)
	}
	if err := f.Sync(); err != nil {
		panic(err)
	}
	return filename
}

func SendMessage(FrigateEvent FrigateEventStruct, bot *tgbotapi.BotAPI) {
	conf := config.New()

	// Prepare text message
	text := ""
	text = text + "Camera: " + FrigateEvent.Camera + "\n"
	text = text + "Label: " + FrigateEvent.Label + "\n"
	t_start := time.Unix(int64(FrigateEvent.StartTime), 0)
	t_end := time.Unix(int64(FrigateEvent.EndTime), 0)
	text = text + fmt.Sprintf("Start time: %s", t_start) + "\n"
	text = text + fmt.Sprintf("End time: %s", t_end) + "\n"
	text = text + fmt.Sprintf("Top score: %f", (FrigateEvent.TopScore*100)) + "%\n"
	text = text + conf.FrigateExternalURL + "/events?cameras=" + FrigateEvent.Camera + "&labels=" + FrigateEvent.Label

	// Save thumbnail
	filepath := SaveThumbnail(FrigateEvent.Thumbnail)
	defer os.Remove(filepath)

	// Send message
	log.Debug.Println("Sending message to Telegram chat id: ", conf.TelegramChatID)
	msg := tgbotapi.NewPhoto(conf.TelegramChatID, tgbotapi.FilePath(filepath))
	msg.Caption = text
	_, err := bot.Send(msg)

	if err != nil {
		log.Error.Println("Error sending message: " + err.Error())
	}
}

func main() {
	log.LogFunc()
	conf := config.New()
	log.Info.Println("Starting frigate-telegram.")
	log.Info.Println("Version:      " + config.Version)
	log.Info.Println("Frigate URL:  " + conf.FrigateURL)

	// Initializing telegram bot
	bot, err := tgbotapi.NewBotAPI(conf.TelegramBotToken)
	if err != nil {
		log.Error.Println("Error initalizing telegram bot: " + err.Error())
	}
	bot.Debug = conf.Debug
	log.Info.Println("Authorized on account " + bot.Self.UserName)

	FrigateEventsURL := conf.FrigateURL + "/api/events"

	// Starting loop for getting events from Frigate
	for true {
		FrigateEvents := GetEvents(FrigateEventsURL)

		// Parse events
		for Events := range FrigateEvents {
			if FrigateAfter == 0 {
				FrigateAfter = FrigateEvents[Events].EndTime
			}
			if FrigateEvents[Events].EndTime > FrigateAfter {
				FrigateAfter = FrigateEvents[Events].EndTime
			}
			SendMessage(FrigateEvents[Events], bot)
		}
		time.Sleep(time.Duration(conf.SleepTime) * time.Second)
		log.Debug.Println("Sleeping for " + strconv.Itoa(conf.SleepTime) + " seconds.")
	}
}
