package telegram

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/oldtyt/frigate-telegram/internal/config"
	"github.com/oldtyt/frigate-telegram/internal/log"
	"github.com/oldtyt/frigate-telegram/internal/redis"
)

var redisErrorText string = "Error setting value, check logs."

// CommandKeyboard returns an inline keyboard with all available commands
func CommandKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Status", "/status"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏸️ Stop", "/stop"),
			tgbotapi.NewInlineKeyboardButtonData("▶️ Resume", "/resume"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔕 Mute", "/mute"),
			tgbotapi.NewInlineKeyboardButtonData("🔔 Unmute", "/unmute"),
		),
	)
}

// ChatBot is needed to check the work of the bot.
func ChatBot(bot *tgbotapi.BotAPI, conf *config.Config) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// Handle callback queries from inline keyboard buttons
		if update.CallbackQuery != nil {
			// Extract the command from callback data
			callbackData := update.CallbackQuery.Data

			// Create a new message to respond to the callback
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "")
			sendMessage := true

			// Process the command from callback data
			switch callbackData {
			case "/status":
				sendMessage, msg = Status(msg, conf)
			case "/stop":
				sendMessage, msg = Stop(msg, conf)
			case "/resume":
				sendMessage, msg = Resume(msg, conf)
			case "/mute":
				sendMessage, msg = Mute(msg, conf)
			case "/unmute":
				sendMessage, msg = Unmute(msg, conf)
			default:
				msg.Text = "I don't know that command"
			}

			// Answer the callback query to remove the loading state
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			if _, err := bot.Request(callback); err != nil {
				log.Error.Println("Error answering callback query: " + err.Error())
			}

			// Send the response message
			if sendMessage {
				if _, err := bot.Send(msg); err != nil {
					log.Error.Println("Error sending message: " + err.Error())
				}
			}
			continue
		}

		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		sendMessage := true

		// Check if it's a command
		if update.Message.IsCommand() {
			// Extract the command from the Message.
			switch update.Message.Command() {
			case "help":
				sendMessage, msg = Help(msg, conf)
				// Add inline keyboard to help message
				msg.ReplyMarkup = CommandKeyboard()
			case "start":
				msg.Text = "Welcome to Frigate Telegram Bot! Here are the available commands:"
				msg.ReplyMarkup = CommandKeyboard()
			case "ping":
				msg.Text = "pong"
			case "king":
				msg.Text = "kong"
			case "pong":
				msg.Text = "ping"
			case "status":
				sendMessage, msg = Status(msg, conf)
			case "stop":
				sendMessage, msg = Stop(msg, conf)
			case "resume":
				sendMessage, msg = Resume(msg, conf)
			case "mute":
				sendMessage, msg = Mute(msg, conf)
			case "unmute":
				sendMessage, msg = Unmute(msg, conf)
			default:
				msg.Text = "I don't know that command"
			}
		}

		if !sendMessage {
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.Text = "I don't know that command"
		}
		if _, err := bot.Send(msg); err != nil {
			log.Error.Fatalln("Error sending message: " + err.Error())
		}
	}
}

func Help(msg tgbotapi.MessageConfig, conf *config.Config) (bool, tgbotapi.MessageConfig) {
	if msg.BaseChat.ChatID == conf.TelegramChatID {
		text := "*Available Commands:*\n\n"
		text += "📊 Current status: /status\n"
		text += "⏸️ Stop send events: /stop\n"
		text += "▶️ Resume send events: /resume\n"
		text += "🔕 Mute send events: /mute\n"
		text += "🔔 Unmute send events: /unmute\n\n"
		text += "You can use the buttons below or type the commands.\n\n"
		text += "Commands work only in chat id: `" + strconv.FormatInt(conf.TelegramChatID, 10) + "` (Current chat)"
		msg.Text = text
		msg.ParseMode = tgbotapi.ModeMarkdown
		return true, msg
	}
	return false, msg
}

func Status(msg tgbotapi.MessageConfig, conf *config.Config) (bool, tgbotapi.MessageConfig) {
	if msg.BaseChat.ChatID == conf.TelegramChatID {
		text := "Send event: `" + strconv.FormatBool(redis.GetStateSendEvent()) + "`\n"
		text += "Mute event: `" + strconv.FormatBool(redis.GetStateMuteEvent()) + "`\n"
		msg.Text = text
		msg.ParseMode = tgbotapi.ModeMarkdown
		return true, msg
	}
	return false, msg
}

func Stop(msg tgbotapi.MessageConfig, conf *config.Config) (bool, tgbotapi.MessageConfig) {
	if msg.BaseChat.ChatID == conf.TelegramChatID {
		r := redis.SetStateSendEvent(true)
		if r {
			msg.Text = "Stop send message."
			return true, msg
		} else {
			msg.Text = redisErrorText
			return true, msg
		}
	}
	return false, msg
}

func Resume(msg tgbotapi.MessageConfig, conf *config.Config) (bool, tgbotapi.MessageConfig) {
	if msg.BaseChat.ChatID == conf.TelegramChatID {
		r := redis.SetStateSendEvent(false)
		if r {
			msg.Text = "Resume send message."
			return true, msg
		} else {
			msg.Text = redisErrorText
			return true, msg
		}
	}
	return false, msg
}

func Mute(msg tgbotapi.MessageConfig, conf *config.Config) (bool, tgbotapi.MessageConfig) {
	if msg.BaseChat.ChatID == conf.TelegramChatID {
		r := redis.SetStateMuteEvent(true)
		if r {
			msg.Text = "Mute send message."
			return true, msg
		} else {
			msg.Text = redisErrorText
			return true, msg
		}
	}
	return false, msg
}

func Unmute(msg tgbotapi.MessageConfig, conf *config.Config) (bool, tgbotapi.MessageConfig) {
	if msg.BaseChat.ChatID == conf.TelegramChatID {
		r := redis.SetStateMuteEvent(false)
		if r {
			msg.Text = "Unmute send message."
			return true, msg
		} else {
			msg.Text = redisErrorText
			return true, msg
		}
	}
	return false, msg
}
