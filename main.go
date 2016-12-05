package main

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"github.com/ngs/line-buychat/app"
)

func main() {
	app, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
