package gock

import (
	"net/http"
	"net/url"
	"regexp"
	"sync"
)

// mutex is used interally for locking thread-sensitive functions.
var mutex = &sync.Mutex{}

// config global singleton store.
var config = struct {
	Networking        bool
	NetworkingFilters []FilterRequestFunc
}{}

// New creates and registers a new HTTP mock with
// default settings and returns the Request DSL for HTTP mock
// definition and set up.
func New(uri string) *Request {
	Intercept()

	res := NewResponse()
	req := NewRequest()
	req.URLStruct, res.Error = url.Parse(normalizeURI(uri))

	// Create the new mock expectation
	exp := NewMock(req, res)
	Register(exp)

	return req
}

// Intercepting returns true if gock is currently able to intercept.
func Intercepting() bool {
	return http.DefaultTransport == DefaultTransport
}

// Intercept enables HTTP traffic interception via http.DefaultTransport.
// If you are using a custom HTTP transport, you have to use `gock.Transport()`
func Intercept() {
	if !Intercepting() {
		http.DefaultTransport = DefaultTransport
	}
}

// InterceptClient allows the developer to intercept HTTP traffic using
// a custom http.Client who uses a non default http.Transport/http.RoundTripper implementation.
func InterceptClient(cli *http.Client) {
	trans := NewTransport()
	trans.Transport = cli.Transport
	cli.Transport = trans
}

// RestoreClient allows the developer to disable and restore the
// original transport in the given http.Client.
func RestoreClient(cli *http.Client) {
	trans, ok := cli.Transport.(*Transport)
	if !ok {
		return
	}
	cli.Transport = trans.Transport
}

// Disable disables HTTP traffic interception by gock.
func Disable() {
	mutex.Lock()
	defer mutex.Unlock()
	http.DefaultTransport = NativeTransport
}

// Off disables the default HTTP interceptors and removes
// all the registered mocks, even if they has not been intercepted yet.
func Off() {
	Flush()
	Disable()
}

// EnableNetworking enables real HTTP networking
func EnableNetworking() {
	mutex.Lock()
	defer mutex.Unlock()
	config.Networking = true
}

// DisableNetworking disables real HTTP networking
func DisableNetworking() {
	mutex.Lock()
	defer mutex.Unlock()
	config.Networking = false
}

// NetworkingFilter determines if an http.Request should be triggered or not.
func NetworkingFilter(fn FilterRequestFunc) {
	mutex.Lock()
	defer mutex.Unlock()
	config.NetworkingFilters = append(config.NetworkingFilters, fn)
}

// DisableNetworkingFilters disables registered networking filters.
func DisableNetworkingFilters() {
	mutex.Lock()
	defer mutex.Unlock()
	config.NetworkingFilters = []FilterRequestFunc{}
}

func normalizeURI(uri string) string {
	if ok, _ := regexp.MatchString("^http[s]?", uri); !ok {
		return "http://" + uri
	}
	return uri
}
