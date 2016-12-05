package app

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/ngs/go-amazon-product-advertising-api/amazon"
	"github.com/stvp/rollbar"
)

const cartKeyPrefix = "buychat:line:"

func cartClearAction() linebot.TemplateAction {
	return linebot.NewPostbackTemplateAction("空にする", `{"Action":"`+string(PostbackActionClearCart)+`"}`, "")
}

// CartSize returns cart size
func (app *App) CartSize(cartKey string) (int, error) {
	app.ReconnectRedisIfNeeeded()
	return redis.Int(app.RedisConn.Do("LLEN", cartKey))
}

// ClearCart clears items
func (app *App) ClearCart(cartKey string) error {
	app.ReconnectRedisIfNeeeded()
	return app.RedisConn.Send("DEL", cartKey)
}

// AddCartItem adds items to cart
func (app *App) AddCartItem(cartKey string, ASIN string) error {
	app.ReconnectRedisIfNeeeded()
	if err := app.RedisConn.Send("LPUSH", cartKey, ASIN); err != nil {
		return err
	}
	return app.RedisConn.Flush()
}

// RemoveCartItem removes items from cart
func (app *App) RemoveCartItem(cartKey string, ASIN string) error {
	app.ReconnectRedisIfNeeeded()
	if err := app.RedisConn.Send("LREM", cartKey, 1, ASIN); err != nil {
		return err
	}
	return app.RedisConn.Flush()
}

func (app *App) getCartItems(cartKey string) ([]string, error) {
	app.ReconnectRedisIfNeeeded()
	return redis.Strings(app.RedisConn.Do("LRANGE", cartKey, 0, -1))
}

// HandleCart handles GET /cart/:cartid
func (app *App) HandleCart(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	cartKey := cartKeyPrefix + params["type"] + ":" + params["id"]
	res, err := app.getCartItems(cartKey)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	app.Log.Printf("Cart %v %v", cartKey, res)
	if len(res) > 0 {
		app.Log.Printf("%v %v", cartKey, res)
		params := amazon.CartCreateParameters{}
		quantities := map[string]int{}
		for _, asin := range res {
			quantities[asin]++
		}
		for asin, quantity := range quantities {
			params.Items.AddASIN(asin, quantity)
		}
		retryCount := 0
		for {
			res, err := app.Amazon().CartCreate(params).Do()
			if err != nil {
				if strings.Contains(err.Error(), requestThrottleError) {
					if retryCount < retryMax {
						retryCount++
						app.Log.Printf("Retrying %d/%d", retryCount, retryMax)
						time.Sleep(time.Second)
						continue
					}
					http.Error(w, "申し訳ありません、すこし待ってから、もう一度開いてください", 400)
					return
				}
				http.Error(w, err.Error(), 500)
				rollbar.Error(rollbar.ERR, err)
				rollbar.Wait()
				return
			}
			http.Redirect(w, r, res.Cart.MobileCartURL, 303)
			return
		}
	} else {
		http.Error(w, "カートにまだ何も追加されていません", 404)
	}
}

// HandleAddCart handles add cart
func (app *App) HandleAddCart(replyToken string, data PostbackData, cartKey string) error {
	size, err := app.CartSize(cartKey)
	cartURL := os.Getenv("HTTP_BASE") + "/cart/" +
		strings.Replace(strings.Replace(cartKey, cartKeyPrefix, "", 1), ":", "/", 1)
	cartURLAction := linebot.NewURITemplateAction("購入する", cartURL)
	cartShowAction := linebot.NewPostbackTemplateAction("カートを見る", `{"Action":"`+string(PostbackActionShowCart)+`"}`, "")
	if err != nil {
		return err
	}
	if size >= 5 {
		_, err = app.Line.ReplyMessage(replyToken, linebot.NewTemplateMessage("カートが一杯です",
			linebot.NewButtonsTemplate("", "カートが一杯です", "Amazon のカートに追加するか、空にしてください",
				cartURLAction,
				cartShowAction,
				cartClearAction(),
			))).Do()
		return err
	}
	if err = app.AddCartItem(cartKey, data.ASIN); err != nil {
		return err
	}
	msg1 := linebot.NewTextMessage(`カートに追加しました`)
	msg2 := linebot.NewTemplateMessage("カートに追加しました: "+data.Title,
		linebot.NewButtonsTemplate(data.ImageURL, data.Title, data.Label, cartShowAction, cartURLAction))
	_, err = app.Line.ReplyMessage(replyToken, msg1, msg2).Do()
	return err
}

// HandleClearCart handles clear cart
func (app *App) HandleClearCart(replyToken string, cartKey string) error {
	if err := app.ClearCart(cartKey); err != nil {
		return err
	}
	msg1 := linebot.NewTextMessage(`カートを空にしました`)
	_, err := app.Line.ReplyMessage(replyToken, msg1).Do()
	return err
}

// HandleShowCart handles show cart
func (app *App) HandleShowCart(replyToken string, cartKey string) error {
	ids, err := app.getCartItems(cartKey)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return app.ReplyText(replyToken, "カートに何もはいっていません")
	}
	items, err := app.lookupItems(ids)
	if err != nil {
		if strings.Contains(err.Error(), requestThrottleError) {
			return app.ReplyText(replyToken, "申し訳ありません、すこし待ってから、もう一度送信してださい")
		}
		return err
	}
	template := getAmazonItemCarousel(items,
		func(item amazon.Item, imgURL string, label string, title string) []linebot.TemplateAction {
			postbackData := &PostbackData{
				Action: PostbackActionRemoveCart,
				ASIN:   item.ASIN,
				Title:  title,
			}
			bytes, _ := json.Marshal(postbackData)
			return []linebot.TemplateAction{
				linebot.NewPostbackTemplateAction("カートから削除", string(bytes), ""),
				linebot.NewURITemplateAction("Amazon で見る", item.DetailPageURL),
			}
		})
	cartURL := os.Getenv("HTTP_BASE") + "/cart/" +
		strings.Replace(strings.Replace(cartKey, cartKeyPrefix, "", 1), ":", "/", 1)
	msg1 := linebot.NewTextMessage("カートに " + strconv.Itoa(len(ids)) + "個の商品が入っています")
	msg2 := linebot.NewTemplateMessage("カートの内容", template)
	msg3 := linebot.NewTemplateMessage("Amazon で購入しますか？",
		linebot.NewConfirmTemplate("Amazon で購入しますか？",
			linebot.NewURITemplateAction("購入する", cartURL),
			cartClearAction(),
		))
	json, _ := msg2.MarshalJSON()
	app.Log.Println(string(json))
	_, err = app.Line.ReplyMessage(replyToken, msg1, msg2, msg3).Do()
	return nil
}

// HandleRemoveCart handles remove cart
func (app *App) HandleRemoveCart(replyToken string, data PostbackData, cartKey string) error {
	if err := app.RemoveCartItem(cartKey, data.ASIN); err != nil {
		return err
	}
	return app.ReplyText(replyToken, "カートから削除しました: "+data.Title)
}
