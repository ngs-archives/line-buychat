package app

import "github.com/ngs/go-amazon-product-advertising-api/amazon"

func (app *App) searchItems(keyword string) []amazon.Item {
	param := amazon.ItemSearchParameters{
		Keywords:    keyword,
		SearchIndex: amazon.SearchIndexAll,
		ResponseGroups: []amazon.ItemSearchResponseGroup{
			amazon.ItemSearchResponseGroupLarge,
		},
	}
	res, err := app.Amazon.ItemSearch(param).Do()
	if err != nil {
		app.Log.Printf("Got error %v %v", err, param)
		return res.Items.Item
	}
	return res.Items.Item
}
