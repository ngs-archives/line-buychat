package yolp

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Request interface
type Request interface {
	HTTPMethod() string
	Endpoint() string
	Query() url.Values
}

// Client YOLP Client
type Client struct {
	AppID  string
	Secret string
}

// New returns new client
func New(appID string, secret string) (*Client, error) {
	if appID == "" {
		return nil, errors.New("AppID is not specified")
	}
	if secret == "" {
		return nil, errors.New("Secret is not specified")
	}
	return &Client{
		AppID:  appID,
		Secret: secret,
	}, nil
}

// NewFromEnvionment returns new client from environment variables
func NewFromEnvionment() (*Client, error) {
	return New(
		os.Getenv("YDN_APP_ID"),
		os.Getenv("YDN_SECRET"),
	)
}

// URL returns URL for request
func (client *Client) URL(req Request) *url.URL {
	url, _ := url.Parse(req.Endpoint())
	q := req.Query()
	q["output"] = []string{"xml"}
	q["appid"] = []string{client.AppID}
	url.RawQuery = q.Encode()
	return url
}

// DoRequest sends HTTP request
func (client *Client) DoRequest(req Request, responseObject interface{}) (*http.Response, error) {
	url := client.URL(req)
	method := req.HTTPMethod()
	var res *http.Response
	var err error
	switch strings.ToUpper(method) {
	case "GET":
		res, err = http.Get(url.String())
		break
	case "POST":
		formData := url.Query()
		url.RawQuery = ""
		res, err = http.PostForm(url.String(), formData)
		break
	default:
		return nil, fmt.Errorf("Unsupported HTTP method: %v", method)
	}
	if err != nil {
		return nil, err
	}
	if data, _ := ioutil.ReadAll(res.Body); data != nil {
		if err = xml.Unmarshal(data, responseObject); err != nil {
			var e Error
			if err2 := xml.Unmarshal(data, &e); err2 == nil {
				return nil, e
			}
			return nil, err
		}
	}
	return res, err
}
