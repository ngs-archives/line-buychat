package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/ngs/go-amazon-product-advertising-api/amazon"
)

const noimgURL = "https://buychat.s3-ap-northeast-1.amazonaws.com/line-carousel-noimg.png"

// HandleCallback handles POST /callback
func (app *App) HandleCallback(w http.ResponseWriter, r *http.Request) {
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
		if err := app.HandleEvent(event); err != nil {
			app.Log.Printf("Got error %v %v", err, event)
			if err = app.ReplyText(event.ReplyToken, "ごめんなさい、検索中にエラーが発生してしまいました"); err != nil {
				app.Log.Printf("Got error again %v %v", err, event)
				http.Error(w, err.Error(), 500)
			}
			return
		}
	}
	r.Write(bytes.NewBufferString("OK"))
}

// HandleEvent handles webhook event
func (app *App) HandleEvent(event *linebot.Event) error {
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
			return app.HandleTextMessage(event.ReplyToken, message.Text)
			// case *linebot.LocationMessage:
			//   TODO: search local travel books
			// case *linebot.ImageMessage:
			//   TODO: search ISBN
		}
	case linebot.EventTypePostback:
		return app.HandlePostbackData(event.ReplyToken, event.Postback.Data, cartKey)
	}
	return nil
}

// ReplyText replies text
func (app *App) ReplyText(replyToken string, text string) error {
	_, err := app.Line.ReplyMessage(replyToken,
		linebot.NewTextMessage(text)).Do()
	return err
}

// HandleTextMessage handles text message
func (app *App) HandleTextMessage(replyToken string, text string) error {
	items, err := app.searchItems(text)
	if err != nil {
		if apiErr, ok := err.(amazon.Error); ok && apiErr.Code == amazon.RequestThrottled {
			return app.ReplyText(replyToken, "申し訳ありません、すこし待ってから、もう一度送信してださい")
		}
		return err
	}
	if len(items) == 0 {
		return app.ReplyText(replyToken, `ごめんなさい、"`+text+`" に該当する商品はみつかりませんでした`)
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
		strTitle := string(title[0:len(title)])
		postbackData := &PostbackData{
			Action:   PostbackActionAddCart,
			ASIN:     item.ASIN,
			ImageURL: imgURL,
			Label:    label,
			Title:    strTitle,
		}
		bytes, _ := json.Marshal(postbackData)
		column := linebot.NewCarouselColumn(
			imgURL,
			strTitle,
			label,
			linebot.NewPostbackTemplateAction("カートに追加", string(bytes), ""),
			linebot.NewURITemplateAction("Amazon で見る", item.DetailPageURL),
		)
		columns = append(columns, column)
	}
	msg := linebot.NewTemplateMessage(`"`+text+`"の検索結果`, linebot.NewCarouselTemplate(columns...))
	json, _ := msg.MarshalJSON()
	app.Log.Println(string(json))
	_, err = app.Line.ReplyMessage(replyToken, msg).Do()
	return err
}

// HandlePostbackData handles postback data
func (app *App) HandlePostbackData(replyToken string, dataString string, cartKey string) error {
	app.Log.Println(dataString, cartKey)
	var data PostbackData
	if err := json.Unmarshal([]byte(dataString), &data); err != nil {
		return err
	}
	switch data.Action {
	case PostbackActionAddCart:
		return app.HandleAddCart(replyToken, data, cartKey)
	case PostbackActionClearCart:
		return app.HandleClearCart(replyToken, cartKey)
	}
	return nil
}
