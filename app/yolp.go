package app

import yolp "github.com/ngs/go-yolp"

func (app *App) setupYOLPClient() error {
	client, err := yolp.NewFromEnvionment()
	if err != nil {
		return err
	}
	app.YOLP = client
	return nil
}
