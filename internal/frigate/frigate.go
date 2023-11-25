package frigate

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/oldtyt/frigate-telegram/internal/config"
	"github.com/oldtyt/frigate-telegram/internal/log"
	"github.com/oldtyt/frigate-telegram/internal/redis"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type EventsStruct []struct {
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

type EventStruct struct {
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

var Events EventsStruct
var Event EventStruct

func ErrorSend(TextError string, bot *tgbotapi.BotAPI) {
	conf := config.New()
	_, err := bot.Send(tgbotapi.NewMessage(conf.TelegramChatID, TextError))
	if err != nil {
		log.Error.Println(err.Error())
	}
	log.Error.Fatalln(TextError)
}

func SaveThumbnail(Thumbnail string, bot *tgbotapi.BotAPI) string {
	// Decode string Thumbnail base64
	dec, err := base64.StdEncoding.DecodeString(Thumbnail)
	if err != nil {
		ErrorSend("Error when base64 string decode: "+err.Error(), bot)
	}

	// Generate uniq filename
	filename := "/tmp/" + strconv.FormatInt(int64(rand.Intn(10000000000)), 10) + ".jpg"
	f, err := os.Create(filename)
	if err != nil {
		ErrorSend("Error when create file: "+err.Error(), bot)
	}
	defer f.Close()
	if _, err := f.Write(dec); err != nil {
		ErrorSend("Error when write file: "+err.Error(), bot)
	}
	if err := f.Sync(); err != nil {
		ErrorSend("Error when sync file: "+err.Error(), bot)
	}
	return filename
}

func GetEvents(FrigateURL string, bot *tgbotapi.BotAPI) EventsStruct {
	conf := config.New()

	FrigateURL = FrigateURL + "?limit=" + strconv.Itoa(conf.FrigateEventLimit)

	log.Debug.Println("Geting events from Frigate via URL: " + FrigateURL)

	// Request to Frigate
	resp, err := http.Get(FrigateURL)
	if err != nil {
		ErrorSend("Error get events from Frigate, error: "+err.Error(), bot)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != 200 {
		ErrorSend("Response status != 200, when getting events from Frigate.\nExit.", bot)
	}

	// Read data from response
	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorSend("Can't read JSON: "+err.Error(), bot)
	}

	// Parse data from JSON to struct
	err1 := json.Unmarshal(byteValue, &Events)
	if err1 != nil {
		ErrorSend("Error unmarshal json: "+err1.Error(), bot)
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
	filename := "/tmp/" + strconv.FormatInt(int64(rand.Intn(10000000000)), 10) + ".mp4"

	// Create clip file
	f, err := os.Create(filename)
	if err != nil {
		ErrorSend("Error when create file: "+err.Error(), bot)
	}
	defer f.Close()

	// Sleep for wait full save videos
	time.Sleep(time.Duration(conf.SleepTime) * time.Second)
	
	// Download clip file
	resp, err := http.Get(ClipURL)
	if err != nil {
		ErrorSend("Error clip download: "+err.Error(), bot)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		ErrorSend("Return bad status: "+resp.Status, bot)
	}

	// Writer the body to file
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		ErrorSend("Error clip write: "+err.Error(), bot)
	}
	return filename
}

func SendMessageEvent(FrigateEvent EventStruct, bot *tgbotapi.BotAPI) {
	// Get config
	conf := config.New()

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
	text = text + fmt.Sprintf("┣*Top score*\n┗ `%f", (FrigateEvent.TopScore*100)) + "%`\n"
	text = text + "┣*Event id*\n┗ `" + FrigateEvent.ID + "`\n"
	text = text + "┣*Event URL*\n┗ " + conf.FrigateExternalURL + "/events?cameras=" + FrigateEvent.Camera + "&labels=" + FrigateEvent.Label

	// Save thumbnail
	FilePathThumbnail := SaveThumbnail(FrigateEvent.Thumbnail, bot)
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
		ErrorSend("Error send media group message: "+err.Error(), bot)
	}

	if messages == nil {
		ErrorSend("No received messages", bot)
	}
	var State string
	State = "InProgress"
	if FrigateEvent.EndTime != 0 {
		State = "Finished"
	}
	go redis.AddNewEvent(FrigateEvent.ID, State)
}

func ParseEvents(FrigateEvents EventsStruct, bot *tgbotapi.BotAPI) {
	// Parse events
	for Event := range FrigateEvents {
		if redis.CheckEvent(FrigateEvents[Event].ID) {
			log.Info.Println("Found new event. ID - ", FrigateEvents[Event].ID)
			SendMessageEvent(FrigateEvents[Event], bot)
		}
	}
}
