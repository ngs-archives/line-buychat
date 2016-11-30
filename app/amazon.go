package app

import (
	"encoding/xml"

	"github.com/ngs/go-amazon-product-advertising-api/amazon"
)

func (app *App) searchItems(keyword string) []amazon.Item {
	param := amazon.ItemSearchParameters{
		Keywords:      keyword,
		SearchIndex:   amazon.SearchIndexBlended,
		OnlyAvailable: true,
		ResponseGroups: []amazon.ItemSearchResponseGroup{
			amazon.ItemSearchResponseGroupLarge,
		},
	}
	res, err := app.Amazon.ItemSearch(param).Do()
	if err != nil {
		app.Log.Printf("Got error %v %v", err, param)
		return []amazon.Item{}
	}
	xml, _ := xml.Marshal(res.Items)
	app.Log.Println(string(xml))
	return res.Items.Item
}
