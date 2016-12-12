package yolp

import (
	"encoding/xml"
	"strings"
)

// Error error response
type Error struct {
	XMLName xml.Name `xml:"Error"`
	Message string
}

// Error returns error message
func (e Error) Error() string {
	return strings.TrimSpace(e.Message)
}
