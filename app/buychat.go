package app

import (
	"log"
	"net/http"
	"os"

	zbar "github.com/PeterCxy/gozbar"
	"github.com/stvp/rollbar"

	"github.com/gorilla/mux"

	"github.com/garyburd/redigo/redis"
	apachelog "github.com/lestrrat/go-apache-logformat"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/ngs/go-amazon-product-advertising-api/amazon"
)

// App main app
type App struct {
	ZbarScanner   *zbar.Scanner
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
	scanner := zbar.NewScanner()
	scanner.SetConfig(0, zbar.CFG_ENABLE, 1)
	if err != nil {
		return nil, err
	}
	logger := log.New(os.Stderr, "[buychat]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	rollbar.Token = os.Getenv("ROLLBAR_KEY")
	if env := os.Getenv("ROLLBAR_ENV"); env != "" {
		rollbar.Environment = env
	}
	app := &App{
		Line:        line,
		Log:         logger,
		ZbarScanner: scanner,
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
	defer app.ZbarScanner.Destroy()
	return http.ListenAndServe(":"+port, mw)
}
