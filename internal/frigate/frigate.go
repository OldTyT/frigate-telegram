package frigate

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
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
	// SubLabel           []any       `json:"sub_label"`
	Thumbnail string      `json:"thumbnail"`
	TopScore  interface{} `json:"top_score"`
	Zones     []any       `json:"zones"`
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
	// SubLabel           []any       `json:"sub_label"`
	Thumbnail string      `json:"thumbnail"`
	TopScore  interface{} `json:"top_score"`
	Zones     []any       `json:"zones"`
}

var Events EventsStruct
var Event EventStruct

func NormalizeTagText(text string) string {
	var alphabetCheck = regexp.MustCompile(`^[A-Za-z]+$`)
	var NormalizedText []string
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		wordString := fmt.Sprintf("%c", runes[i])
		if _, err := strconv.Atoi(wordString); err == nil {
			NormalizedText = append(NormalizedText, wordString)
		}
		if alphabetCheck.MatchString(wordString) {
			NormalizedText = append(NormalizedText, wordString)
		}
	}
	return strings.Join(NormalizedText, "")
}

func GetTagList(Tags []any) []string {
	var my_tags []string
	for _, zone := range Tags {
		if zone != nil {
			my_tags = append(my_tags, NormalizeTagText(zone.(string)))
		}
	}
	return my_tags
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

func WarnSend(TextError string, bot *tgbotapi.BotAPI, EventID string) {
	conf := config.New()
	TextError += "\nEventID: " + EventID
	_, err := bot.Send(tgbotapi.NewMessage(conf.TelegramChatID, TextError))
	if err != nil {
		log.Error.Println(err.Error())
	}
	log.Warn.Println(TextError)
}

func SaveThumbnail(EventID string, Thumbnail string, bot *tgbotapi.BotAPI) string {
	log.Debug.Printf("Processing thumbnail for event ID: %s", EventID)

	// Verify that we have a non-empty thumbnail string
	if Thumbnail == "" {
		ErrorSend("Empty thumbnail string received", bot, EventID)
	}

	// Decode string Thumbnail base64
	dec, err := base64.StdEncoding.DecodeString(Thumbnail)
	if err != nil {
		ErrorSend("Error when base64 string decode: "+err.Error(), bot, EventID)
	}

	// Check if we got any data after decoding
	if len(dec) == 0 {
		ErrorSend("Decoded thumbnail is empty", bot, EventID)
	}

	log.Debug.Printf("Decoded thumbnail size: %d bytes", len(dec))

	// Generate uniq filename
	filename := "/tmp/" + EventID + ".jpg"
	f, err := os.Create(filename)
	if err != nil {
		ErrorSend("Error when create file: "+err.Error(), bot, EventID)
	}
	defer f.Close()

	// Write data to file
	bytesWritten, err := f.Write(dec)
	if err != nil {
		ErrorSend("Error when write file: "+err.Error(), bot, EventID)
	}

	// Check if we wrote anything
	if bytesWritten == 0 {
		ErrorSend("No data written to thumbnail file", bot, EventID)
	}

	log.Debug.Printf("Written %d bytes to %s", bytesWritten, filename)

	// Ensure file is properly synced to disk
	err = f.Sync()
	if err != nil {
		ErrorSend("Error when sync file: "+err.Error(), bot, EventID)
	}

	// Verify file exists and has content
	fileInfo, err := os.Stat(filename)
	if err != nil {
		ErrorSend("Error verifying thumbnail file: "+err.Error(), bot, EventID)
	}

	if fileInfo.Size() == 0 {
		ErrorSend("Thumbnail file is empty after write", bot, EventID)
	}

	log.Debug.Printf("Successfully saved thumbnail to %s (size: %d bytes)", filename, fileInfo.Size())
	return filename
}

func DownloadThumbnail(EventID string, bot *tgbotapi.BotAPI) string {
	// Get config
	conf := config.New()

	// Generate thumbnail URL
	ThumbnailURL := conf.FrigateURL + "/api/events/" + EventID + "/thumbnail.jpg"
	log.Debug.Println("Downloading thumbnail from URL: " + ThumbnailURL)

	// Generate uniq filename
	filename := "/tmp/" + EventID + ".jpg"

	// Download thumbnail file
	resp, err := http.Get(ThumbnailURL)
	if err != nil {
		ErrorSend("Error thumbnail download: "+err.Error(), bot, EventID)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		ErrorSend("Return bad status: "+resp.Status, bot, EventID)
	}

	// Read content length if available
	contentLength := resp.ContentLength
	if contentLength == 0 {
		ErrorSend("Received empty thumbnail from server (content length is 0)", bot, EventID)
	}
	log.Debug.Printf("Expected thumbnail content length: %d bytes", contentLength)

	// Create thumbnail file
	f, err := os.Create(filename)
	if err != nil {
		ErrorSend("Error when create file: "+err.Error(), bot, EventID)
	}
	defer f.Close() // Ensure file is closed even if there's an error

	// Write the body to file
	bytesWritten, err := io.Copy(f, resp.Body)
	if err != nil {
		ErrorSend("Error thumbnail write: "+err.Error(), bot, EventID)
	}
	log.Debug.Printf("Written %d bytes to %s", bytesWritten, filename)

	// Check if we wrote anything
	if bytesWritten == 0 {
		ErrorSend("No data written to thumbnail file", bot, EventID)
	}

	// Ensure file is properly synced to disk
	err = f.Sync()
	if err != nil {
		ErrorSend("Error syncing file to disk: "+err.Error(), bot, EventID)
	}

	// Verify file exists and has content
	fileInfo, err := os.Stat(filename)
	if err != nil {
		ErrorSend("Error verifying thumbnail file: "+err.Error(), bot, EventID)
	}

	if fileInfo.Size() == 0 {
		ErrorSend("Thumbnail file is empty after download", bot, EventID)
	}

	log.Debug.Printf("Successfully downloaded thumbnail to %s (size: %d bytes)", filename, fileInfo.Size())
	return filename
}

func GetEvents(FrigateURL string, bot *tgbotapi.BotAPI, SetBefore bool) EventsStruct {
	conf := config.New()

	FrigateURL = FrigateURL + "?limit=" + strconv.Itoa(conf.FrigateEventLimit)

	if SetBefore {
		timestamp := time.Now().UTC().Unix()
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
		WarnSend("Response status != 200, when getting events from Frigate.", bot, "ALL")
		return EventsStruct{}
	}

	// Read data from response
	byteValue, err := io.ReadAll(resp.Body)
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
	log.Debug.Println("Downloading clip from URL: " + ClipURL)

	// Generate uniq filename
	filename := "/tmp/" + EventID + ".mp4"

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

	// Read content length if available
	contentLength := resp.ContentLength
	if contentLength == 0 {
		WarnSend("Received empty clip from server (content length is 0)", bot, EventID)
		return ""
	}
	log.Debug.Printf("Expected content length: %d bytes", contentLength)

	// Create clip file
	f, err := os.Create(filename)
	if err != nil {
		ErrorSend("Error when create file: "+err.Error(), bot, EventID)
	}
	defer f.Close() // Ensure file is closed even if there's an error

	// Writer the body to file
	bytesWritten, err := io.Copy(f, resp.Body)
	if err != nil {
		ErrorSend("Error clip write: "+err.Error(), bot, EventID)
	}
	log.Debug.Printf("Written %d bytes to %s", bytesWritten, filename)

	// Check if we wrote anything
	if bytesWritten == 0 {
		WarnSend("No data written to clip file", bot, EventID)
		return ""
	}

	// Ensure file is properly synced to disk
	err = f.Sync()
	if err != nil {
		ErrorSend("Error syncing file to disk: "+err.Error(), bot, EventID)
	}

	// Close the file
	err = f.Close()
	if err != nil {
		ErrorSend("Error closing clip file: "+err.Error(), bot, EventID)
	}

	// Verify file exists and has content
	fileInfo, err := os.Stat(filename)
	if err != nil {
		ErrorSend("Error verifying clip file: "+err.Error(), bot, EventID)
	}

	if fileInfo.Size() == 0 {
		ErrorSend("Clip file is empty after download", bot, EventID)
	}

	log.Debug.Printf("Successfully downloaded clip to %s (size: %d bytes)", filename, fileInfo.Size())
	return filename
}

func SendMessageEvent(FrigateEvent EventStruct, bot *tgbotapi.BotAPI) {
	// Get config
	conf := config.New()

	redis.AddNewEvent(FrigateEvent.ID, "InWork", time.Duration(60)*time.Second)

	// Prepare text message
	text := ""
	t_start := time.Unix(int64(FrigateEvent.StartTime), 0)
	if conf.ShortEventMessageFormat {
		// Short message format
		text += fmt.Sprintf("#%s detected on #%s at %s",
			NormalizeTagText(FrigateEvent.Label),
			NormalizeTagText(FrigateEvent.Camera),
			t_start)

	} else {
		// Normal message format
		text += "*Event*\n"
		text += "┣*Camera*\n┗ #" + NormalizeTagText(FrigateEvent.Camera) + "\n"
		text += "┣*Label*\n┗ #" + NormalizeTagText(FrigateEvent.Label) + "\n"
		// if FrigateEvent.SubLabel != nil {
		// 	text += "┣*SubLabel*\n┗ #" + strings.Join(GetTagList(FrigateEvent.SubLabel), ", #") + "\n"
		// }
		text += fmt.Sprintf("┣*Start time*\n┗ `%s", t_start) + "`\n"
		if FrigateEvent.EndTime == 0 {
			text += "┣*End time*\n┗ `In progess`" + "\n"
		} else {
			t_end := time.Unix(int64(FrigateEvent.EndTime), 0)
			text += fmt.Sprintf("┣*End time*\n┗ `%s", t_end) + "`\n"
		}
		text += fmt.Sprintf("┣*Top score*\n┗ `%f", (FrigateEvent.Data.TopScore*100)) + "%`\n"
		text += "┣*Event id*\n┗ `" + FrigateEvent.ID + "`\n"
		text += "┣*Zones*\n┗ #" + strings.Join(GetTagList(FrigateEvent.Zones), ", #") + "\n"
		text += "*URLs*\n"
		text += "┣[Events](" + conf.FrigateExternalURL + "/events?cameras=" + FrigateEvent.Camera + "&labels=" + FrigateEvent.Label + "&zones=" + strings.Join(GetTagList(FrigateEvent.Zones), ",") + ")\n"
		text += "┣[General](" + conf.FrigateExternalURL + ")\n"
		text += "┗[Source clip](" + conf.FrigateExternalURL + "/api/events/" + FrigateEvent.ID + "/clip.mp4)\n"
	}

	var medias []interface{}
	var FilePathThumbnail string

	if conf.IncludeThumbnailEvent {
		// Save thumbnail
		if FrigateEvent.Thumbnail != "" {
			// Try to use the base64 thumbnail first
			log.Debug.Println("Using base64 thumbnail from event data")
			FilePathThumbnail = SaveThumbnail(FrigateEvent.ID, FrigateEvent.Thumbnail, bot)

			// Verify thumbnail file has content
			fileInfo, err := os.Stat(FilePathThumbnail)
			if err != nil || fileInfo.Size() == 0 {
				log.Debug.Println("Base64 thumbnail failed, trying direct download")
				// If base64 method failed, try direct download
				if err == nil {
					os.Remove(FilePathThumbnail) // Remove empty file
				}
				FilePathThumbnail = DownloadThumbnail(FrigateEvent.ID, bot)
			}
		} else {
			// No thumbnail in event data, download directly
			log.Debug.Println("No thumbnail in event data, downloading directly")
			FilePathThumbnail = DownloadThumbnail(FrigateEvent.ID, bot)
		}

		// Verify thumbnail file before adding to media group
		thumbnailInfo, err := os.Stat(FilePathThumbnail)
		if err != nil {
			ErrorSend("Error getting thumbnail file info: "+err.Error(), bot, FrigateEvent.ID)
		}

		if thumbnailInfo.Size() == 0 {
			log.Error.Printf("Thumbnail file is empty: %s", FilePathThumbnail)
			ErrorSend("Cannot send empty thumbnail file", bot, FrigateEvent.ID)
		}

		MediaThumbnail := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(FilePathThumbnail))
		MediaThumbnail.Caption = text
		MediaThumbnail.ParseMode = tgbotapi.ModeMarkdown

		medias = append(medias, MediaThumbnail)
	}

	// Define FilePathClip outside the if block to make it available later
	var FilePathClip string
	var hasClip bool

	if FrigateEvent.HasClip && FrigateEvent.EndTime != 0 {
		// Save clip
		FilePathClip = SaveClip(FrigateEvent.ID, bot)
		hasClip = true
		if FilePathClip == "" {
			hasClip = false
		}
		if hasClip {
			videoInfo, err := os.Stat(FilePathClip)
			if err != nil {
				ErrorSend("Error receiving information about the clip file: "+err.Error(), bot, FrigateEvent.ID)
			}

			// Double check file size
			if videoInfo.Size() == 0 {
				log.Error.Printf("Clip file is empty: %s", FilePathClip)
				hasClip = false
			} else if videoInfo.Size() < 52428800 {
				// Telegram don't send large file see for more: https://github.com/OldTyT/frigate-telegram/issues/5
				// Add clip to media group
				log.Debug.Printf("Adding clip to media group: %s (size: %d bytes)", FilePathClip, videoInfo.Size())
				MediaClip := tgbotapi.NewInputMediaVideo(tgbotapi.FilePath(FilePathClip))

				if !conf.IncludeThumbnailEvent {
					MediaClip.Caption = text
				}

				medias = append(medias, MediaClip)
			} else {
				log.Debug.Printf("Clip file size is too large: %d bytes (limit: 52428800)", videoInfo.Size())
			}
		}
	}

	log.Debug.Printf("Sending media group with %d items", len(medias))

	if len(medias) != 0 {
		// Create message
		msg := tgbotapi.MediaGroupConfig{
			ChatID: conf.TelegramChatID,
			Media:  medias,
		}
		msg.DisableNotification = redis.GetStateMuteEvent()

		messages, err := bot.SendMediaGroup(msg)
		if err != nil {
			log.Error.Printf("Failed to send media group: %s", err.Error())
			if strings.Contains(err.Error(), "file must be non-empty") {
				// Try to get more information about the files we're trying to send
				for i, media := range medias {
					switch m := media.(type) {
					case tgbotapi.InputMediaPhoto:
						if filePath, ok := m.Media.(tgbotapi.FilePath); ok {
							fileInfo, statErr := os.Stat(string(filePath))
							if statErr != nil {
								log.Error.Printf("Media item %d: Cannot get file info: %s", i, statErr.Error())
							} else {
								log.Error.Printf("Media item %d: Photo file exists, size: %d bytes", i, fileInfo.Size())
							}
						}
					case tgbotapi.InputMediaVideo:
						if filePath, ok := m.Media.(tgbotapi.FilePath); ok {
							fileInfo, statErr := os.Stat(string(filePath))
							if statErr != nil {
								log.Error.Printf("Media item %d: Cannot get file info: %s", i, statErr.Error())
							} else {
								log.Error.Printf("Media item %d: Video file exists, size: %d bytes", i, fileInfo.Size())
							}
						}
					}
				}
			}
			ErrorSend("Error send media group message: "+err.Error(), bot, FrigateEvent.ID)
		}

		if messages == nil {
			ErrorSend("No received messages", bot, FrigateEvent.ID)
		}
	} else {
		msg := tgbotapi.NewMessage(conf.TelegramChatID, "")
		msg.Text = text
		if _, err := bot.Send(msg); err != nil {
			log.Error.Fatalln("Error sending message: " + err.Error())
		}
	}

	// Now we can safely remove the files after the media group is sent
	if hasClip {
		os.Remove(FilePathClip)
	}

	if conf.IncludeThumbnailEvent {
		os.Remove(FilePathThumbnail)
	}

	var State string
	State = "InProgress"
	if FrigateEvent.EndTime != 0 {
		State = "Finished"
	}
	redis.AddNewEvent(FrigateEvent.ID, State, time.Duration(conf.RedisTTL)*time.Second)
}

func StringsContains(MyStr string, MySlice []string) bool {
	for _, v := range MySlice {
		if v == MyStr {
			return true
		}
	}
	return false
}

func ParseEvents(FrigateEvents EventsStruct, bot *tgbotapi.BotAPI, WatchDog bool) {
	// Parse events
	conf := config.New()
	RedisKeyPrefix := ""
	if WatchDog {
		RedisKeyPrefix = "WatchDog_"
	}
	for Event := range FrigateEvents {
		// Skip by camera
		if !(len(conf.FrigateExcludeCamera) == 1 && conf.FrigateExcludeCamera[0] == "None") {
			if StringsContains(FrigateEvents[Event].Camera, conf.FrigateExcludeCamera) {
				log.Debug.Println("Skiping event from exclude camera: " + FrigateEvents[Event].Camera)
				continue
			}
		}
		if !(len(conf.FrigateIncludeCamera) == 1 && conf.FrigateIncludeCamera[0] == "All") {
			if !(StringsContains(FrigateEvents[Event].Camera, conf.FrigateIncludeCamera)) {
				log.Debug.Println("Skiping event from include camera: " + FrigateEvents[Event].Camera)
				continue
			}
		}
		// Skip by camera

		// Skip by label
		if !(len(conf.FrigateExcludeLabel) == 1 && conf.FrigateExcludeLabel[0] == "None") {
			if StringsContains(FrigateEvents[Event].Label, conf.FrigateExcludeLabel) {
				log.Debug.Println("Skiping event by exclude label: " + FrigateEvents[Event].Label)
				continue
			}
		}
		if !(len(conf.FrigateIncludeLabel) == 1 && conf.FrigateIncludeLabel[0] == "All") {
			if !(StringsContains(FrigateEvents[Event].Label, conf.FrigateIncludeLabel)) {
				log.Debug.Println("Skiping event by include label: " + FrigateEvents[Event].Label)
				continue
			}
		}
		// Skip by label

		// Skip by zone
		zones := GetTagList(FrigateEvents[Event].Zones)
		needSkip := false
		if !(len(conf.FrigateExcludeZone) == 1 && conf.FrigateExcludeZone[0] == "None") {
			if len(zones) != 0 {
				for _, zone := range zones {
					if StringsContains(zone, conf.FrigateExcludeZone) {
						log.Debug.Println("Skiping event by exclude zone: " + zone)
						needSkip = true
					}
				}
			}
		}
		if needSkip {
			continue
		}
		if !(len(conf.FrigateIncludeZone) == 1 && conf.FrigateIncludeZone[0] == "All") {
			if len(zones) == 0 {
				log.Debug.Println("Skipping the event due to zero zones.")
				continue
			}
			for _, zone := range zones {
				if !(StringsContains(zone, conf.FrigateIncludeZone)) {
					log.Debug.Println("Skiping event by include zone: " + zone)
					needSkip = true
				}
			}
		}
		if needSkip {
			continue
		}
		// Skip by zone

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
	text += "┣*Zones*\n┗ `" + strings.Join(GetTagList(FrigateEvent.Zones), ", ") + "`\n"
	text += "┣*Event URL*\n┗ " + conf.FrigateExternalURL + "/events?cameras=" + FrigateEvent.Camera + "&labels=" + FrigateEvent.Label + "&zones=" + strings.Join(GetTagList(FrigateEvent.Zones), ",")
	msg := tgbotapi.NewMessage(conf.TelegramChatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableNotification = redis.GetStateMuteEvent()
	_, err := bot.Send(msg)
	if err != nil {
		log.Error.Println(err.Error())
	}
	redis.AddNewEvent("WatchDog_"+FrigateEvent.ID, "Finished", time.Duration(conf.RedisTTL)*time.Second)
}

func NotifyEvents(bot *tgbotapi.BotAPI, FrigateEventsURL string) {
	conf := config.New()
	for {
		FrigateEvents := GetEvents(FrigateEventsURL, bot, false)
		ParseEvents(FrigateEvents, bot, true)
		time.Sleep(time.Duration(conf.WatchDogSleepTime) * time.Second)
	}
}
