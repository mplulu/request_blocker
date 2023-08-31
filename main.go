package main

import (
	"net/http"
	"net/url"

	"github.com/mplulu/rano"
	"github.com/mplulu/renv"
	"github.com/mplulu/request_blocker/env"
	"github.com/mplulu/request_blocker/limit_rate"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	var env *env.ENV
	renv.ParseCmd(&env)

	tlgBot := rano.NewRano(env.TelegramBotToken, []string{
		env.TelegramChatId,
	})
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	e.HideBanner = true
	limitRateCenter := limit_rate.NewCenter(tlgBot)
	go limitRateCenter.Start()

	// ignoreLogPrefix := []string{}
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			// for _, prefix := range ignoreLogPrefix {
			// 	if strings.HasPrefix(c.Request().URL.Path, prefix) {
			// 		return true
			// 	}
			// }
			return false
		},
		CustomTimeFormat: "02/01/06T15:04:05Z",
		Format:           "[${host}]${remote_ip},t=${time_custom},d=${latency_human},s=${status},m=${method},uri=${uri}\n",
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	e.Use(limitRateCenter.MiddlewareLimitRate)
	// Setup proxy
	url, err := url.Parse(env.TargetURl)
	if err != nil {
		e.Logger.Fatal(err)
	}

	targets := []*middleware.ProxyTarget{
		{
			URL: url,
		},
	}
	e.Use(middleware.Proxy(middleware.NewRoundRobinBalancer(targets)))
	e.Logger.Fatal(e.Start(env.Host))
}
