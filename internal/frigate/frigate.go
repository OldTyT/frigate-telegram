package frigate

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/oldtyt/frigate-telegram/internal/config"
	"github.com/oldtyt/frigate-telegram/internal/log"
	"github.com/oldtyt/frigate-telegram/internal/redis"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type EventsStruct []struct {
	Box    interface{} `json:"box"`
	Camera string      `json:"camera"`
	Data   struct {
		Attributes []interface{} `json:"attributes"`
		Box        []float64     `json:"box"`
		Region     []float64     `json:"region"`
		Score      float64       `json:"score"`
		TopScore   float64       `json:"top_score"`
		Type       string        `json:"type"`
	} `json:"data"`
	EndTime            float64     `json:"end_time"`
	FalsePositive      interface{} `json:"false_positive"`
	HasClip            bool        `json:"has_clip"`
	HasSnapshot        bool        `json:"has_snapshot"`
	ID                 string      `json:"id"`
	Label              string      `json:"label"`
	PlusID             interface{} `json:"plus_id"`
	RetainIndefinitely bool        `json:"retain_indefinitely"`
	StartTime          float64     `json:"start_time"`
	SubLabel           interface{} `json:"sub_label"`
	Thumbnail          string      `json:"thumbnail"`
	TopScore           interface{} `json:"top_score"`
	Zones              []any       `json:"zones"`
}

type EventStruct struct {
	Box    interface{} `json:"box"`
	Camera string      `json:"camera"`
	Data   struct {
		Attributes []interface{} `json:"attributes"`
		Box        []float64     `json:"box"`
		Region     []float64     `json:"region"`
		Score      float64       `json:"score"`
		TopScore   float64       `json:"top_score"`
		Type       string        `json:"type"`
	} `json:"data"`
	EndTime            float64     `json:"end_time"`
	FalsePositive      interface{} `json:"false_positive"`
	HasClip            bool        `json:"has_clip"`
	HasSnapshot        bool        `json:"has_snapshot"`
	ID                 string      `json:"id"`
	Label              string      `json:"label"`
	PlusID             interface{} `json:"plus_id"`
	RetainIndefinitely bool        `json:"retain_indefinitely"`
	StartTime          float64     `json:"start_time"`
	SubLabel           interface{} `json:"sub_label"`
	Thumbnail          string      `json:"thumbnail"`
	TopScore           interface{} `json:"top_score"`
	Zones              []any       `json:"zones"`
}

var Events EventsStruct
var Event EventStruct
var floatType = reflect.TypeOf(float64(0))
var stringType = reflect.TypeOf("")

func GETZones(Zones []any) []string {
	var my_zones []string
	for _, zone := range Zones {
		my_zones = append(my_zones, zone.(string))
	}
	return my_zones
}

func ErrorSend(TextError string, bot *tgbotapi.BotAPI, EventID string) {
	conf := config.New()
	TextError += "\nEventID: " + EventID
	_, err := bot.Send(tgbotapi.NewMessage(conf.TelegramChatID, TextError))
	if err != nil {
		log.Error.Println(err.Error())
	}
	log.Error.Fatalln(TextError)
}

func SaveThumbnail(EventID string, Thumbnail string, bot *tgbotapi.BotAPI) string {
	// Decode string Thumbnail base64
	dec, err := base64.StdEncoding.DecodeString(Thumbnail)
	if err != nil {
		ErrorSend("Error when base64 string decode: "+err.Error(), bot, EventID)
	}

	// Generate uniq filename
	filename := "/tmp/" + EventID + ".jpg"
	f, err := os.Create(filename)
	if err != nil {
		ErrorSend("Error when create file: "+err.Error(), bot, EventID)
	}
	defer f.Close()
	if _, err := f.Write(dec); err != nil {
		ErrorSend("Error when write file: "+err.Error(), bot, EventID)
	}
	if err := f.Sync(); err != nil {
		ErrorSend("Error when sync file: "+err.Error(), bot, EventID)
	}
	return filename
}

func GetEvents(FrigateURL string, bot *tgbotapi.BotAPI, SetBefore bool) EventsStruct {
	conf := config.New()

	FrigateURL = FrigateURL + "?limit=" + strconv.Itoa(conf.FrigateEventLimit)

	if SetBefore {
		timestamp := time.Now().UTC().UnixNano()
		timestamp = timestamp - int64(conf.EventBeforeSeconds)
		FrigateURL = FrigateURL + "&before=" + strconv.FormatInt(timestamp, 10)
	}

	log.Debug.Println("Geting events from Frigate via URL: " + FrigateURL)

	// Request to Frigate
	resp, err := http.Get(FrigateURL)
	if err != nil {
		ErrorSend("Error get events from Frigate, error: "+err.Error(), bot, "ALL")
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != 200 {
		ErrorSend("Response status != 200, when getting events from Frigate.\nExit.", bot, "ALL")
	}

	// Read data from response
	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorSend("Can't read JSON: "+err.Error(), bot, "ALL")
	}

	// Parse data from JSON to struct
	err1 := json.Unmarshal(byteValue, &Events)
	if err1 != nil {
		ErrorSend("Error unmarshal json: "+err1.Error(), bot, "ALL")
		if e, ok := err.(*json.SyntaxError); ok {
			log.Info.Println("syntax error at byte offset " + strconv.Itoa(int(e.Offset)))
		}
		log.Info.Println("Exit.")
	}

	// Return Events
	return Events
}

func SaveClip(EventID string, bot *tgbotapi.BotAPI) string {
	// Get config
	conf := config.New()

	// Generate clip URL
	ClipURL := conf.FrigateURL + "/api/events/" + EventID + "/clip.mp4"

	// Generate uniq filename
	filename := "/tmp/" + EventID + ".mp4"

	// Create clip file
	f, err := os.Create(filename)
	if err != nil {
		ErrorSend("Error when create file: "+err.Error(), bot, EventID)
	}
	defer f.Close()

	// Download clip file
	resp, err := http.Get(ClipURL)
	if err != nil {
		ErrorSend("Error clip download: "+err.Error(), bot, EventID)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		ErrorSend("Return bad status: "+resp.Status, bot, EventID)
	}

	// Writer the body to file
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		ErrorSend("Error clip write: "+err.Error(), bot, EventID)
	}
	return filename
}

func SendMessageEvent(FrigateEvent EventStruct, bot *tgbotapi.BotAPI) {
	// Get config
	conf := config.New()

	redis.AddNewEvent(FrigateEvent.ID, "InWork", time.Duration(60)*time.Second)

	// Prepare text message
	text := "*Event*\n"
	text = text + "┣*Camera*\n┗ `" + FrigateEvent.Camera + "`\n"
	text = text + "┣*Label*\n┗ `" + FrigateEvent.Label + "`\n"
	t_start := time.Unix(int64(FrigateEvent.StartTime), 0)
	text = text + fmt.Sprintf("┣*Start time*\n┗ `%s", t_start) + "`\n"
	if FrigateEvent.EndTime == 0 {
		text = text + "┣*End time*\n┗ `In progess`" + "\n"
	} else {
		t_end := time.Unix(int64(FrigateEvent.EndTime), 0)
		text = text + fmt.Sprintf("┣*End time*\n┗ `%s", t_end) + "`\n"
	}
	text = text + fmt.Sprintf("┣*Top score*\n┗ `%f", (FrigateEvent.Data.TopScore*100)) + "%`\n"
	text = text + "┣*Event id*\n┗ `" + FrigateEvent.ID + "`\n"
	text = text + "┣*Zones*\n┗ `" + strings.Join(GETZones(FrigateEvent.Zones), ", ") + "`\n"
	text = text + "┣*Event URL*\n┗ " + conf.FrigateExternalURL + "/events?cameras=" + FrigateEvent.Camera + "&labels=" + FrigateEvent.Label + "&zones=" + strings.Join(GETZones(FrigateEvent.Zones), ",")

	// Save thumbnail
	FilePathThumbnail := SaveThumbnail(FrigateEvent.ID, FrigateEvent.Thumbnail, bot)
	defer os.Remove(FilePathThumbnail)

	var medias []interface{}
	MediaThumbnail := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(FilePathThumbnail))
	MediaThumbnail.Caption = text
	MediaThumbnail.ParseMode = tgbotapi.ModeMarkdown
	medias = append(medias, MediaThumbnail)

	if FrigateEvent.HasClip && FrigateEvent.EndTime != 0 {
		// Save clip
		FilePathClip := SaveClip(FrigateEvent.ID, bot)
		defer os.Remove(FilePathClip)

		// Add clip to media group
		MediaClip := tgbotapi.NewInputMediaVideo(tgbotapi.FilePath(FilePathClip))
		medias = append(medias, MediaClip)
	}

	// Create message
	msg := tgbotapi.MediaGroupConfig{
		ChatID: conf.TelegramChatID,
		Media:  medias,
	}
	messages, err := bot.SendMediaGroup(msg)
	if err != nil {
		ErrorSend("Error send media group message: "+err.Error(), bot, FrigateEvent.ID)
	}

	if messages == nil {
		ErrorSend("No received messages", bot, FrigateEvent.ID)
	}
	var State string
	State = "InProgress"
	if FrigateEvent.EndTime != 0 {
		State = "Finished"
	}
	redis.AddNewEvent(FrigateEvent.ID, State, time.Duration(conf.RedisTTL)*time.Second)
}

func ParseEvents(FrigateEvents EventsStruct, bot *tgbotapi.BotAPI, WatchDog bool) {
	// Parse events
	RedisKeyPrefix := ""
	if WatchDog {
		RedisKeyPrefix = "WatchDog_"
	}
	for Event := range FrigateEvents {
		if redis.CheckEvent(RedisKeyPrefix + FrigateEvents[Event].ID) {
			if WatchDog {
				SendTextEvent(FrigateEvents[Event], bot)
			} else {
				go SendMessageEvent(FrigateEvents[Event], bot)
			}
		}
	}
}

func SendTextEvent(FrigateEvent EventStruct, bot *tgbotapi.BotAPI) {
	conf := config.New()
	text := "*New event*\n"
	text += "┣*Camera*\n┗ `" + FrigateEvent.Camera + "`\n"
	text += "┣*Label*\n┗ `" + FrigateEvent.Label + "`\n"
	t_start := time.Unix(int64(FrigateEvent.StartTime), 0)
	text += fmt.Sprintf("┣*Start time*\n┗ `%s", t_start) + "`\n"
	text += fmt.Sprintf("┣*Top score*\n┗ `%f", (FrigateEvent.Data.TopScore*100)) + "%`\n"
	text += "┣*Event id*\n┗ `" + FrigateEvent.ID + "`\n"
	text += "┣*Zones*\n┗ `" + strings.Join(GETZones(FrigateEvent.Zones), ", ") + "`\n"
	text += "┣*Event URL*\n┗ " + conf.FrigateExternalURL + "/events?cameras=" + FrigateEvent.Camera + "&labels=" + FrigateEvent.Label + "&zones=" + strings.Join(GETZones(FrigateEvent.Zones), ",")
	msg := tgbotapi.NewMessage(conf.TelegramChatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(msg)
	if err != nil {
		log.Error.Println(err.Error())
	}
	redis.AddNewEvent("WatchDog_"+FrigateEvent.ID, "Finished", time.Duration(conf.RedisTTL)*time.Second)
}

func NotifyEvents(bot *tgbotapi.BotAPI, FrigateEventsURL string) {
	conf := config.New()
	for true {
		FrigateEvents := GetEvents(FrigateEventsURL, bot, true)
		ParseEvents(FrigateEvents, bot, true)
		time.Sleep(time.Duration(conf.WatchDogSleepTime) * time.Second)
	}
}
