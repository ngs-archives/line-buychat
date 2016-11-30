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
	if event.Source == nil {
		return nil
	}
	cartKey := ""
	switch event.Source.Type {
	case linebot.EventSourceTypeRoom:
		cartKey = fmt.Sprintf("buychat:line:room:%v", event.Source.RoomID)
		break
	case linebot.EventSourceTypeGroup:
		cartKey = fmt.Sprintf("buychat:line:group:%v", event.Source.GroupID)
		break
	case linebot.EventSourceTypeUser:
		cartKey = fmt.Sprintf("buychat:line:user:%v", event.Source.UserID)
		break
	}
	switch event.Type {
	case linebot.EventTypeMessage:
		switch message := event.Message.(type) {
		case *linebot.TextMessage:
			return app.handleTextMessage(event.ReplyToken, message.Text)
			// case *linebot.LocationMessage:
			//   TODO: search local travel books
		}
	case linebot.EventTypePostback:
		return app.handlePostbackData(event.ReplyToken, event.Postback.Data, cartKey)
	}
	return nil
}

func (app *App) handleTextMessage(replyToken string, text string) error {
	items := app.searchItems(text)
	if len(items) == 0 {
		_, err := app.Line.ReplyMessage(replyToken,
			linebot.NewTextMessage(`ごめんなさい、"`+text+`" に該当する商品はみつかりませんでした`)).Do()
		return err
	}
	var columns []*linebot.CarouselColumn
	for _, item := range items {
		if len(columns) == 5 {
			break
		}
		title := []rune(item.ItemAttributes.Title)
		if len(title) == 0 || len(item.DetailPageURL) == 0 {
			continue
		}
		if len(title) > 40 {
			title = title[0:40]
		}
		imgURL := item.LargeImage.URL
		if imgURL == "" {
			imgURL = noimgURL
		} else {
			imgURL = strings.Replace(imgURL, "http://ecx.images-amazon.com/", "https://images-na.ssl-images-amazon.com/", -1)
		}
		label := ""
		if len(item.ItemAttributes.Author) > 0 && len(item.ItemAttributes.Author[0]) > 0 {
			label = item.ItemAttributes.Author[0]
		}
		if label == "" && len(item.ItemAttributes.Artist) > 0 {
			label = item.ItemAttributes.Artist
		}
		if label == "" && len(label) == 0 && len(item.ItemAttributes.Creator.Name) > 0 {
			label = item.ItemAttributes.Creator.Name
		}
		if label == "" {
			label = item.ItemAttributes.Manufacturer
		}
		if label == "" {
			label = "-"
		}
		column := linebot.NewCarouselColumn(
			imgURL,
			string(title[0:len(title)]),
			label,
			linebot.NewPostbackTemplateAction("カートに追加", "add-cart:"+item.ASIN, ""),
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

func (app *App) handlePostbackData(replyToken string, data string, cartKey string) error {
	app.Log.Println(data, cartKey)
	return nil
}
