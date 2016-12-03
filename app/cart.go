package app

import (
	"net/http"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

const cartKeyPrefix = "buychat:line:"

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
	} else {
		http.Error(w, "Cart not found", 404)
	}
}
