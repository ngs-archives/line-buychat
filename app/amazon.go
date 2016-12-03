package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/ngs/go-amazon-product-advertising-api/amazon"
)

var currentClient = 0

// Amazon returns amazon client
func (app *App) Amazon() *amazon.Client {
	client := app.AmazonClients[currentClient]
	currentClient++
	app.Log.Printf("Using Amazon Client %d of %d", currentClient, len(app.AmazonClients))
	if currentClient == len(app.AmazonClients) {
		currentClient = 0
	}
	return client
}

func (app *App) setupAmazonClients() error {
	accessKeyIDs := strings.Split(os.Getenv("AWS_ACCESS_KEY_ID"), ":")
	secretAccessKeys := strings.Split(os.Getenv("AWS_SECRET_ACCESS_KEY"), ":")
	if len(accessKeyIDs) != len(secretAccessKeys) {
		return fmt.Errorf("Specified %d Access Key IDs, but Secret Access Keys was %d",
			len(accessKeyIDs), len(secretAccessKeys))
	}
	clients := []*amazon.Client{}
	associateTag := os.Getenv("AWS_ASSOCIATE_TAG")
	for i, key := range accessKeyIDs {
		secret := secretAccessKeys[i]
		client, err := amazon.New(key, secret, associateTag, amazon.RegionJapan)
		if err != nil {
			return err
		}
		clients = append(clients, client)
	}
	app.AmazonClients = clients
	return nil
}

func (app *App) searchItems(keyword string) ([]amazon.Item, error) {
	param := amazon.ItemSearchParameters{
		Keywords:      keyword,
		SearchIndex:   amazon.SearchIndexBlended,
		OnlyAvailable: true,
		ResponseGroups: []amazon.ItemSearchResponseGroup{
			amazon.ItemSearchResponseGroupLarge,
		},
	}
	res, err := app.Amazon().ItemSearch(param).Do()
	if err != nil {
		app.Log.Printf("Got error %v %v", err, param)
		return []amazon.Item{}, err
	}
	return res.Items.Item, nil
}
