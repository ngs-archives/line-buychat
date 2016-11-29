package app

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/line/line-bot-sdk-go/linebot"
)

func (app *App) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Method Not Allowed: %v", r.Method), http.StatusMethodNotAllowed)
	}
	events, err := app.Line.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	log.Printf("Got events %v", events)
	for _, event := range events {
		app.handleEvent(event)
	}
	r.Write(bytes.NewBufferString("OK"))
}

func (app *App) handleEvent(event *linebot.Event) error {
	switch event.Type {
	case linebot.EventTypeMessage:
		switch message := event.Message.(type) {
		case *linebot.TextMessage:
			return app.handleTextMessage(event.ReplyToken, message)
			// case *linebot.LocationMessage:
			//   TODO: search local travel books
		}
	case linebot.EventTypePostback:
		return app.handlePostbackData(event.ReplyToken, event.Postback.Data)
	}
	return nil
}

func (app *App) handleTextMessage(replyToken string, message *linebot.TextMessage) error {
	items := app.searchItems(message.Text)
	var columns []*linebot.CarouselColumn
	for _, item := range items {
		action := linebot.NewPostbackTemplateAction("カートに追加", "add-cart-"+item.ASIN, "")
		column := linebot.NewCarouselColumn(item.MediumImage.URL, item.ItemAttributes.Title, item.ItemAttributes.Manufacturer, action)
		columns = append(columns, column)
	}
	msg := linebot.NewTemplateMessage("Unsupported client", linebot.NewCarouselTemplate(columns...))
	_, err := app.Line.ReplyMessage(replyToken, msg).Do()
	return err
}

func (app *App) handlePostbackData(replyToken string, data string) error {
	return nil
}
