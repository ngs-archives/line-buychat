package app

import "github.com/ngs/go-amazon-product-advertising-api/amazon"

func (app *App) searchItems(keyword string) []amazon.Item {
	res, _ := app.Amazon.ItemSearch(amazon.ItemSearchParameters{}).Do()
	return res.Items.Item
}
