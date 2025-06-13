[![Go Report Card](https://goreportcard.com/badge/github.com/OldTyT/frigate-telegram)](https://goreportcard.com/report/OldTyT/frigate-telegram)
[![GolangCI](https://golangci.com/badges/github.com/OldTyT/frigate-telegram.svg)](https://golangci.com/r/github.com/OldTyT/frigate-telegram)

# Frigate telegram

Frigate telegram event notifications.

---

## Example of work

![alt text](https://raw.githubusercontent.com/OldTyT/frigate-telegram/main/resources/img/telegram_msg.png)

## How to start

1. Install docker
2. Download `docker-compose.yml` file:
```bash
https://raw.githubusercontent.com/OldTyT/frigate-telegram/main/docker-compose.yml
```
3. Change environment variables in docker-compose
4. Deploy:
```bash
docker compose up -d
```
5. Profit!

### Environment variables

| Variable | Default value | Description |
| ----------- | ----------- | ----------- |
| `TELEGRAM_BOT_TOKEN` | `""`| Token for telegram bot. |
| `FRIGATE_URL` | `http://localhost:5000` | Internal link in frigate. |
| `FRIGATE_EVENT_LIMIT` | `20`| 	Limit the number of events returned. |
| `DEBUG` | `False` | Debug mode. |
| `TELEGRAM_CHAT_ID` | `0` | Telegram chat id. |
| `SLEEP_TIME`| `5` | Sleep time after cycle, in second. |
| `FRIGATE_EXTERNAL_URL` | `http://localhost:5000` | External link in frigate(need for generate link in message). |
| `TZ` | `""` | Timezone |
| `REDIS_ADDR` | `localhost:6379` | IP and port redis |
| `REDIS_PASSWORD` | `""` | Redis password |
| `REDIS_DB` | `0` | Redis DB |
| `REDIS_PROTOCOL` | `3` | Redis protocol |
| `REDIS_TTL` | `1209600` | Redis TTL for key event(in seconds) |
| `TIME_WAIT_SAVE` | `30` | Wait for fully video event created(in seconds) |
| `WATCH_DOG_SLEEP_TIME` | `3` | Sleep watch dog goroutine seconds |
| `EVENT_BEFORE_SECONDS` | `300` | Send event before seconds |
| `SEND_TEXT_EVENT` | `False` | Send text event without media |
| `FRIGATE_EXCLUDE_CAMERA` | `None` | List exclude frigate camera, separate `,` |
| `FRIGATE_INCLUDE_CAMERA` | `All` | List Include frigate camera, separate `,` |
| `FRIGATE_EXCLUDE_LABEL` | `None` | List exclude frigate event, separate `,` |
| `FRIGATE_INCLUDE_LABEL` | `All` | List Include frigate event, separate `,` |
| `FRIGATE_EXCLUDE_ZONE` | `None` | List exclude frigate zone, separate `,` |
| `FRIGATE_INCLUDE_ZONE` | `All` | List Include frigate zone, separate `,` |
| `REST_API_ENABLE` | `False` | Enabling the http rest API |
| `REST_API_LISTEN_ADDR` | `:8080` | Rest API listen addr |
| `SHORT_EVENT_MESSAGE_FORMAT` | `False` | Short event message format |
| `INCLUDE_THUMBNAIL_EVENT` | `True` | Include thumbnail from event to messsage |


## Features

### Rest API

First the API needs to be enabled in the ENV. The docker-compose.yml has the ENV already but set to "False" per default.
REST_API_ENABLE: True

The Full URL: http://IP-OF-DOCKER-HOST:8080/api/v1/COMMAND

Possible Commands:
- /mute
- /ping
- /resume
- /status
- /stop
- /unmute

For more details Swagger aviaible on: `http://localhost:8080/docs/index.html`

### Mute/unmute events messages

You can enable or disable notifications for event messages (data is stored in Redis, ensuring persistence across restarts).

Commands:
* `/mute`
* `/unmute`

> [!WARNING]
> For security reasons, commands only work in the TelegramChatID chat.

### Stop/resume send events messages

You can pause or resume sending notifications for event messages (data is stored in Redis, ensuring persistence across restarts).

Commands:
* `/stop`
* `/resume`

> [!WARNING]
> For security reasons, commands only work in the TelegramChatID chat.
