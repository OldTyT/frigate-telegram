[![Go Report Card](https://goreportcard.com/badge/github.com/OldTyT/frigate-telegram)](https://goreportcard.com/report/OldTyT/frigate-telegram)
[![GolangCI](https://golangci.com/badges/github.com/OldTyT/frigate-telegram.svg)](https://golangci.com/r/github.com/OldTyT/frigate-telegram)

# Frigate telegram

Frigate telegram event notifications.

---

## ENV variables

| Variable | Default value | Description |
| ----------- | ----------- | ----------- |
| `TELEGRAM_BOT_TOKEN` | `""`| Token for telegram bot. |
| `FRIGATE_URL` | `http://localhost:5000` | Internal link in frigate. |
| `FRIGATE_EVENT_LIMIT` | `3`| 	Limit the number of events returned. |
| `DEBUG` | `False` | Debug mode. |
| `TELEGRAM_CHAT_ID` | `0` | Telegram chat id. |
| `SLEEP_TIME`| `5` | Sleep time after cycle, in second. |
| `FRIGATE_EXTERNAL_URL` | `http://localhost:5000` | External link in frigate(need for generate link in message). |
