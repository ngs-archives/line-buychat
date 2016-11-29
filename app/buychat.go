package app

import (
	"log"
	"net/http"
	"os"

	apachelog "github.com/lestrrat/go-apache-logformat"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/ngs/go-amazon-product-advertising-api/amazon"
)

// App main app
type App struct {
	Line   *linebot.Client
	Amazon *amazon.Client
	Log    *log.Logger
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
	amazon, err := amazon.NewFromEnvionment()
	if err != nil {
		return nil, err
	}
	logger := log.New(os.Stderr, "[buychat]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	app := &App{
		Line:   line,
		Amazon: amazon,
		Log:    logger,
	}
	return app, nil
}

// Run runs HTTP server
func (app *App) Run() error {
	s := http.NewServeMux()
	s.HandleFunc("/callback", app.handleCallback)
	mw := apachelog.CombinedLog.Wrap(s, os.Stderr)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return http.ListenAndServe(":"+port, mw)
}
