package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oldtyt/frigate-telegram/docs"
	"github.com/oldtyt/frigate-telegram/internal/config"
	"github.com/oldtyt/frigate-telegram/internal/log"
	"github.com/oldtyt/frigate-telegram/internal/redis"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var redisErrorText string = "Error setting value, check logs."

type ResponseApi struct {
	IsError      bool   `json:"error"`
	Message      string `json:"message"`
	SendingEvent string `json:"send_event,omitempty"`
	MuteEvent    string `json:"mute_event,omitempty"`
}

func serverDocs(apiPath string) {
	docs.SwaggerInfo.Title = "Frigate Telegram API"
	docs.SwaggerInfo.Description = "This is manage API for frigate telegram service."
	docs.SwaggerInfo.Version = "main"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = apiPath
	// docs.SwaggerInfo.Schemes = []string{"http", "https"}
}

func ReturnResponse(c *gin.Context, data ResponseApi) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Error.Println(err)
		return
	}
	httpResponseStatus := http.StatusOK
	if data.IsError {
		httpResponseStatus = http.StatusBadGateway
	}
	c.Data(httpResponseStatus, "application/json", jsonBytes)
}

// Stop godoc
// @Summary      Ping
// @Description  Pong
// @Tags         status
// @Accept       json
// @Produce      json
// @Success      200
// @Router       /ping [get]
func Ping(c *gin.Context) {
	ReturnResponse(c, ResponseApi{
		IsError: false,
		Message: "pong",
	})
}

// Stop godoc
// @Summary      Stop send event
// @Description  Stop send event
// @Tags         status
// @Accept       json
// @Produce      json
// @Success      200
// @Failure      502
// @Router       /stop [get]
func Stop(c *gin.Context) {
	r := redis.SetStateSendEvent(true)
	text := ""
	if r {
		text = "Stop send message."
		ReturnResponse(c, ResponseApi{
			IsError: false,
			Message: text,
		})
	} else {
		text = redisErrorText
		ReturnResponse(c, ResponseApi{
			IsError: true,
			Message: text,
		})
	}
}

// Resume godoc
// @Summary      Resume send event
// @Description  Resume send event
// @Tags         status
// @Accept       json
// @Produce      json
// @Success      200
// @Failure      502
// @Router       /resume [get]
func Resume(c *gin.Context) {
	r := redis.SetStateSendEvent(false)
	text := ""
	if r {
		text = "Resume send message."
		ReturnResponse(c, ResponseApi{
			IsError: false,
			Message: text,
		})
	} else {
		text = redisErrorText
		ReturnResponse(c, ResponseApi{
			IsError: true,
			Message: text,
		})
	}
}

// Mute godoc
// @Summary      Mute send event
// @Description  Mute send event
// @Tags         status
// @Accept       json
// @Produce      json
// @Success      200
// @Failure      502
// @Router       /mute [get]
func Mute(c *gin.Context) {
	r := redis.SetStateMuteEvent(true)
	text := ""
	if r {
		text = "Mute send message."
		ReturnResponse(c, ResponseApi{
			IsError: false,
			Message: text,
		})
	} else {
		text = redisErrorText
		ReturnResponse(c, ResponseApi{
			IsError: true,
			Message: text,
		})
	}
}

// Unmute godoc
// @Summary      Unmute send event
// @Description  Unmute send event
// @Tags         status
// @Accept       json
// @Produce      json
// @Success      200
// @Failure      502
// @Router       /unmute [get]
func Unmute(c *gin.Context) {
	r := redis.SetStateMuteEvent(false)
	text := ""
	if r {
		text = "Unmute send message."
		ReturnResponse(c, ResponseApi{
			IsError: false,
			Message: text,
		})
	} else {
		text = redisErrorText
		ReturnResponse(c, ResponseApi{
			IsError: true,
			Message: text,
		})
	}
}

// Status godoc
// @Summary      Get status
// @Description  Current status
// @Tags         status
// @Accept       json
// @Produce      json
// @Success      200
// @Failure      502
// @Router       /status [get]
func Status(c *gin.Context) {
	ReturnResponse(c, ResponseApi{
		IsError:      false,
		Message:      "",
		SendingEvent: strconv.FormatBool(redis.GetStateSendEvent()),
		MuteEvent:    strconv.FormatBool(redis.GetStateMuteEvent())})
}

func RunServer(conf *config.Config) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	apiPath := "/api/v1"
	serverDocs(apiPath)

	r.GET(apiPath+"/ping", Ping)
	r.GET(apiPath+"/stop", Stop)
	r.GET(apiPath+"/resume", Resume)
	r.GET(apiPath+"/mute", Mute)
	r.GET(apiPath+"/unmute", Unmute)
	r.GET(apiPath+"/status", Status)

	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Info.Println("Start Rest API on " + conf.RestAPIListenAddr)
	err := r.Run(conf.RestAPIListenAddr)
	if err != nil {
		log.Error.Fatalln("Error starting Rest API: " + err.Error())
	}
}
