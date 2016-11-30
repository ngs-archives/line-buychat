package app

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
)

const noimgURL = "https://buychat.s3-ap-northeast-1.amazonaws.com/line-carousel-noimg.png"

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
	if event.Source != nil {
		app.Log.Printf("User:%v Group:%v Room:%v Type:%v",
			event.Source.UserID, event.Source.GroupID, event.Source.RoomID, event.Source.Type)
	}
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
	text := message.Text
	items := app.searchItems(text)
	if len(items) == 0 {
		_, err := app.Line.ReplyMessage(replyToken,
			linebot.NewTextMessage(`ごめんなさい、"`+text+`" に該当する商品はみつかりませんでした`)).Do()
		return err
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
		imgURL := item.LargeImage.URL
		if imgURL == "" {
			imgURL = noimgURL
		} else {
			imgURL = strings.Replace(imgURL, "http://ecx.images-amazon.com/", "https://images-na.ssl-images-amazon.com/", -1)
		}
		label := item.ItemAttributes.Manufacturer
		if len(item.ItemAttributes.Author) > 0 && len(item.ItemAttributes.Author[0]) > 0 {
			label = item.ItemAttributes.Author[0]
		}
		if len(label) == 0 && len(item.ItemAttributes.Creator.Name) > 0 {
			label = item.ItemAttributes.Creator.Name
		}
		column := linebot.NewCarouselColumn(
			imgURL,
			string(title[0:len(title)]),
			label+" ",
			linebot.NewPostbackTemplateAction("カートに追加", "add-cart-"+item.ASIN, ""),
			linebot.NewURITemplateAction("Amazon で見る", item.DetailPageURL),
		)
		columns = append(columns, column)
	}
	msg := linebot.NewTemplateMessage(`"`+text+`"の検索結果`, linebot.NewCarouselTemplate(columns...))
	json, _ := msg.MarshalJSON()
	app.Log.Println(string(json))
	_, err := app.Line.ReplyMessage(replyToken, msg).Do()
	return err
}

func (app *App) handlePostbackData(replyToken string, data string) error {
	return nil
}
