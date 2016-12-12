package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"strings"

	zbar "github.com/PeterCxy/gozbar"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/ngs/go-amazon-product-advertising-api/amazon"
	"github.com/stvp/rollbar"
)

const noimgURL = "https://buychat.s3-ap-northeast-1.amazonaws.com/line-carousel-noimg.png"

// HandleCallback handles POST /callback
func (app *App) HandleCallback(w http.ResponseWriter, r *http.Request) {
	events, err := app.Line.ParseRequest(r)
	if err != nil {
		rollbar.Error(rollbar.ERR, err)
		if err == linebot.ErrInvalidSignature {
			http.Error(w, err.Error(), 400)
		} else {
			http.Error(w, err.Error(), 500)
		}
		rollbar.Wait()
		return
	}
	log.Printf("Got events %v", events)
	for _, event := range events {
		if err := app.HandleEvent(event); err != nil {
			rollbar.Error(rollbar.ERR, err)
			app.Log.Printf("Got error %v %v", err, event)
			if err = app.ReplyText(event.ReplyToken, "ごめんなさい、検索中にエラーが発生してしまいました"); err != nil {
				rollbar.Error(rollbar.ERR, err)
				app.Log.Printf("Got error again %v %v", err, event)
				http.Error(w, err.Error(), 500)
			}
			rollbar.Wait()
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
		case *linebot.LocationMessage:
			app.Log.Printf("Location %v %v %v %v\n", message.Title, message.Address, message.Latitude, message.Longitude)
			//   TODO: search local travel books
			return nil
		case *linebot.ImageMessage:
			messageContent, err := app.Line.GetMessageContent(message.ID).Do()
			if err != nil {
				return err
			}
			return app.HandleImage(event.ReplyToken, messageContent.Content)
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
		if strings.Contains(err.Error(), requestThrottleError) {
			return app.ReplyText(replyToken, "申し訳ありません、すこし待ってから、もう一度送信してださい")
		}
		return err
	}
	if len(items) == 0 {
		return app.ReplyText(replyToken, `ごめんなさい、"`+text+`" に該当する商品はみつかりませんでした`)
	}
	return app.replyItemCarousel(replyToken, `"`+text+`"の検索結果`, items)
}

func (app *App) replyItemCarousel(replyToken string, altText string, items []amazon.Item) error {
	template := getAmazonItemCarousel(items,
		func(item amazon.Item, imgURL string, label string, title string) []linebot.TemplateAction {
			postbackData := &PostbackData{
				Action:   PostbackActionAddCart,
				ASIN:     item.ASIN,
				ImageURL: imgURL,
				Label:    label,
				Title:    title,
			}
			bytes, _ := json.Marshal(postbackData)
			return []linebot.TemplateAction{
				linebot.NewPostbackTemplateAction("カートに追加", string(bytes), ""),
				linebot.NewURITemplateAction("Amazon で見る", item.DetailPageURL),
			}
		})
	msg := linebot.NewTemplateMessage(altText, template)
	json, _ := msg.MarshalJSON()
	app.Log.Println(string(json))
	_, err := app.Line.ReplyMessage(replyToken, msg).Do()
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
	case PostbackActionShowCart:
		return app.HandleShowCart(replyToken, cartKey)
	case PostbackActionRemoveCart:
		return app.HandleRemoveCart(replyToken, data, cartKey)
	}
	return nil
}

// HandleImage handles image
func (app *App) HandleImage(replyToken string, content io.ReadCloser) error {
	src, _, err := image.Decode(content)
	if err != nil {
		return app.ReplyText(replyToken, "バーコードを検知できませんでした")
	}
	img := zbar.FromImage(src)
	itemIDs := []string{}
	app.ZbarScanner.Scan(img)
	if img.First() != nil {
		img.First().Each(func(text string) {
			itemIDs = append(itemIDs, text)
		})
	}
	app.Log.Println(itemIDs)
	if len(itemIDs) > 0 {
		items, err := app.searchItems(strings.Join(itemIDs, " "))
		str := strings.Join(itemIDs, ",")
		if err != nil {
			if strings.Contains(err.Error(), requestThrottleError) {
				return app.ReplyText(replyToken, "申し訳ありません、すこし待ってから、もう一度送信してださい")
			}
			return err
		}
		if len(items) > 0 {
			return app.replyItemCarousel(replyToken, `バーコード "`+str+`" の検索結果`, items)
		}
		return app.ReplyText(replyToken, `ごめんなさい、バーコード "`+str+`" に該当する商品はみつかりませんでした`)
	}
	return app.ReplyText(replyToken, "バーコードを検知できませんでした")
}
