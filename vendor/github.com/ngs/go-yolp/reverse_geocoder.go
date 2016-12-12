package yolp

import "net/url"

// ReverseGeocoderRequest ReverseGeocoder API request
type ReverseGeocoderRequest struct {
	client     *Client
	Parameters GeocoderParams
}

// Query returns query parameters
func (req ReverseGeocoderRequest) Query() url.Values {
	return req.Parameters.Query()
}

// HTTPMethod GET
func (req ReverseGeocoderRequest) HTTPMethod() string {
	return "GET"
}

// Endpoint http://reverse.search.olp.yahooapis.jp/OpenLocalPlatform/V1/reverseGeoCoder
func (req ReverseGeocoderRequest) Endpoint() string {
	return "http://reverse.search.olp.yahooapis.jp/OpenLocalPlatform/V1/reverseGeoCoder"
}

// ReverseGeocoderFeature ReverseGeocoder Feature
type ReverseGeocoderFeature struct {
	Feature
	Property ReverseGeocoderProperty
}

// ReverseGeocoderProperty ReverseGeocoder Property
type ReverseGeocoderProperty struct {
	Country        Country
	Address        string
	AddressElement []AddressElement
	Building       []Building
	Road           []Road
}

// ReverseGeocoderResponse ReverseGeocoder API response
type ReverseGeocoderResponse struct {
	YDF
	Feature []ReverseGeocoderFeature
}

// ReverseGeocoder

// ReverseGeocoder requests ReverseGeocoder API
// http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/reversegeocoder.html
func (client *Client) ReverseGeocoder(params GeocoderParams) *ReverseGeocoderRequest {
	return &ReverseGeocoderRequest{client, params}
}

// Do sends request
func (req *ReverseGeocoderRequest) Do() (*ReverseGeocoderResponse, error) {
	var res ReverseGeocoderResponse
	if _, err := req.client.DoRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
