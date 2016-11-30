package app

import (
	"net/http"
	"os"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/ngs/go-amazon-product-advertising-api/amazon"
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

// HandleCart handles GET /cart/:cartid
func (app *App) HandleCart(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	cartKey := cartKeyPrefix + params["type"] + ":" + params["id"]
	app.ReconnectRedisIfNeeeded()
	res, err := redis.Strings(app.RedisConn.Do("LRANGE", cartKey, 0, -1))
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
		res, err := app.Amazon().CartCreate(params).Do()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		http.Redirect(w, r, res.Cart.MobileCartURL, 303)
	} else {
		http.Error(w, "Cart not found", 404)
	}
}

// HandleAddCart handles add cart
func (app *App) HandleAddCart(replyToken string, data PostbackData, cartKey string) error {
	size, err := app.CartSize(cartKey)
	cartURL := os.Getenv("HTTP_BASE") + "/cart/" +
		strings.Replace(strings.Replace(cartKey, cartKeyPrefix, "", 1), ":", "/", 1)
	cartURLAction := linebot.NewURITemplateAction("カートを見る", cartURL)
	if err != nil {
		return err
	}
	if size >= 5 {
		_, err = app.Line.ReplyMessage(replyToken, linebot.NewTemplateMessage("カートが一杯です",
			linebot.NewConfirmTemplate("Amazon のカートに追加するか、空にしてください",
				cartURLAction, cartClearAction()))).Do()
		return err
	}
	if err = app.AddCartItem(cartKey, data.ASIN); err != nil {
		return err
	}
	msg1 := linebot.NewTextMessage(`カートに追加しました`)
	msg2 := linebot.NewTemplateMessage("カートに追加しました: "+data.Title,
		linebot.NewButtonsTemplate(data.ImageURL, data.Title, data.Label, cartURLAction))
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
