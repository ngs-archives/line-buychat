package app

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/garyburd/redigo/redis"
	apachelog "github.com/lestrrat/go-apache-logformat"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/ngs/go-amazon-product-advertising-api/amazon"
)

// App main app
type App struct {
	Line          *linebot.Client
	AmazonClients []*amazon.Client
	Log           *log.Logger
	RedisConn     redis.Conn
}

// New returns new app
func New() (*App, error) {
	line, err := linebot.New(
		os.Getenv("LINE_CHANNEL_SECRET"),
		os.Getenv("LINE_CHANNEL_TOKEN"),
	)
	if err != nil {
		return nil, err
	}
	logger := log.New(os.Stderr, "[buychat]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	app := &App{
		Line: line,
		Log:  logger,
	}
	if err := app.setupAmazonClients(); err != nil {
		return nil, err
	}
	if err := app.SetupRedis(); err != nil {
		return nil, err
	}
	return app, nil
}

// Run runs HTTP server
func (app *App) Run() error {
	router := mux.NewRouter()
	router.HandleFunc("/callback", app.HandleCallback).Methods("POST")
	router.HandleFunc("/cart/{type}/{id}", app.HandleCart).Methods("GET")
	mw := apachelog.CombinedLog.Wrap(router, os.Stderr)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return http.ListenAndServe(":"+port, mw)
}
