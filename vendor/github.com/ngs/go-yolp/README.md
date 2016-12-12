go-yolp
=======

[![Build Status](https://travis-ci.org/ngs/go-yolp.svg?branch=master)](https://travis-ci.org/ngs/go-yolp)
[![GoDoc](https://godoc.org/github.com/ngs/go-yolp?status.svg)](https://godoc.org/github.com/ngs/go-yolp)
[![Go Report Card](https://goreportcard.com/badge/github.com/ngs/go-yolp)](https://goreportcard.com/report/github.com/ngs/go-yolp)
[![Coverage Status](https://coveralls.io/repos/github/ngs/go-yolp/badge.svg?branch=master)](https://coveralls.io/github/ngs/go-yolp?branch=master)

Go Client Library for [Yahoo! Open Local Platform API]

How to Use
----------

```sh
go get -u github.com/ngs/go-yolp
```

```go
package main

import (
	"fmt"
	"log"

	yolp "github.com/ngs/go-yolp"
)

func main() {
	client, err := yolp.NewFromEnvionment()
	if err != nil {
		log.Fatal(err)
	}
	req := client.ReverseGeocoder(yolp.GeocoderParams{
		Latitude:  35.62172852580437,
		Longitude: 139.6999476850032,
		Datum:     yolp.WGS,
	})
	res, err := req.Do()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.Feature[0].Property.Address)
}
```

```sh
export YDN_APP_ID=${YDN_APP_ID}
export YDN_SECRET=${YDN_APP_SECRET}

# go run foo.go
```

## Coverage

- [ ] [Yahoo!ローカルサーチAPI](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/localsearch.html)
- [ ] [Yahoo!ジオコーダAPI](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/geocoder.html)
- [x] [Yahoo!リバースジオコーダAPI](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/reversegeocoder.html)
- [ ] [気象情報API](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/weather.html)
- [ ] [郵便番号検索API](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/zipcodesearch.html)
- [ ] [クチコミ検索API](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/reviewsearch.html)
- [ ] [場所情報API](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/placeinfo.html)
- [ ] [住所ディレクトリAPI](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/addressdirectory.html)
- [ ] [経路地図API](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/routemap.html)
- [ ] [施設内検索API](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/buildindSearch.html)
- [ ] [コンテンツジオコーダAPI](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/contentsgeocoder.html)
- [ ] [ルート沿い検索API](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/spatialSearch.html)
- [ ] [2点間距離API](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/distance.html)
- [ ] [業種マスターAPI](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/genreCode.html)
- [ ] [店舗名寄せAPI](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/getgid.html)
- [ ] [測地系変換API](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/datum.html)
- [ ] [標高API](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/altitude.html)
- [ ] [カセットサーチAPI](http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/cassetteSearch.html)

## Author

[Atsushi Nagase]

## License

See [LICENSE]

[Atsushi Nagase]: https://ngs.io
[LICENSE]: LICENSE
[Yahoo! Open Local Platform API]: http://developer.yahoo.co.jp/webapi/map/
