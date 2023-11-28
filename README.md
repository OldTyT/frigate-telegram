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
