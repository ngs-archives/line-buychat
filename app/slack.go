package app

import (
	"os"

	"github.com/nlopes/slack"
)

func (app *App) setupSlack() error {
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	app.Slack = api
	return nil
}
