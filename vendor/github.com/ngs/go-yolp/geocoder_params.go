package yolp

import (
	"fmt"
	"net/url"
)

// Datum typed constants
type Datum string

const (
	// WGS wgs
	WGS Datum = "wgs"
	// Tokyo tky
	Tokyo Datum = "tky"
)

// GeocoderParams generic parameters for geocoder APIs
type GeocoderParams struct {
	Latitude  float64
	Longitude float64
	Datum     Datum
}

// Query returns query parameters
func (params GeocoderParams) Query() url.Values {
	v := url.Values{}
	v["lat"] = []string{fmt.Sprint(params.Latitude)}
	v["lon"] = []string{fmt.Sprint(params.Longitude)}
	strDatum := string(params.Datum)
	if strDatum != "" {
		v["datum"] = []string{strDatum}
	}
	return v
}
