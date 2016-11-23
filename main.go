package main

import (
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
