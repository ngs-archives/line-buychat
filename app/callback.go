package app

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
)

func (app *App) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Method Not Allowed: %v", r.Method), http.StatusMethodNotAllowed)
	}
	events, err := app.Line.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			http.Error(w, err.Error(), 400)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}
	log.Printf("Got events %v", events)
	for _, event := range events {
		if err := app.handleEvent(event); err != nil {
			app.Log.Printf("Got error %v %v", err, event)
			http.Error(w, err.Error(), 500)
			return
		}
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
	if len(items) == 0 {
		app.Line.ReplyMessage(replyToken, linebot.NewTextMessage("ごめんなさい、"+message.Text+"に該当する商品はみつかりませんでした"))
		return nil
	}
	var columns []*linebot.CarouselColumn
	for i, item := range items {
		if i == 5 {
			break
		}
		title := []rune(item.ItemAttributes.Title)
		if len(title) > 40 {
			title = title[0:40]
		}
		column := linebot.NewCarouselColumn(
			strings.Replace(item.LargeImage.URL, "http://ecx.images-amazon.com/", "https://images-na.ssl-images-amazon.com/", -1),
			string(title[0:len(title)]),
			item.ItemAttributes.Manufacturer+" ",
			linebot.NewPostbackTemplateAction("カートに追加", "add-cart-"+item.ASIN, ""),
			linebot.NewURITemplateAction("Amazon で見る", item.DetailPageURL),
		)
		columns = append(columns, column)
	}
	msg := linebot.NewTemplateMessage(message.Text+"の検索結果", linebot.NewCarouselTemplate(columns...))
	json, _ := msg.MarshalJSON()
	app.Log.Println(string(json))
	_, err := app.Line.ReplyMessage(replyToken, msg).Do()
	return err
}

func (app *App) handlePostbackData(replyToken string, data string) error {
	return nil
}
