package app

import (
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/ngs/go-amazon-product-advertising-api/amazon"
)

// App main app
type App struct {
	Line   *linebot.Client
	Amazon *amazon.Client
}

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
	app := &App{
		Line:   line,
		Amazon: amazon,
	}
	return app, nil
}

func (app *App) Run() error {
	return nil
}
