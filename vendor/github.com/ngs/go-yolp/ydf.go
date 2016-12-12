package yolp

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// YDF http://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/ydf/
type YDF struct {
	XMLName    xml.Name `xml:"YDF"`
	ResultInfo ResultInfo
}

// ResultInfo ResultInfo
type ResultInfo struct {
	XMLName      xml.Name `xml:"ResultInfo"`
	Status       int
	Count        int
	Total        int
	Start        int
	Latency      float64
	Description  string
	Copyright    string
	CompressType string
}

// Feature Feature
type Feature struct {
	XMLName     xml.Name `xml:"Feature"`
	ID          string   `xml:"Id"`
	Name        string
	Description string
	Geometry    Geometry
}

// GeometryType typed constant
type GeometryType string

const (
	// GeometryTypePoint "point"
	GeometryTypePoint GeometryType = "point"
	// GeometryTypeLinestring "linestring"
	GeometryTypeLinestring GeometryType = "linestring"
	// GeometryTypePolygon "polygon"
	GeometryTypePolygon GeometryType = "polygon"
	// GeometryTypeCircle "circle"
	GeometryTypeCircle GeometryType = "circle"
	// GeometryTypeEllipse "ellipse"
	GeometryTypeEllipse GeometryType = "ellipse"
)

// Geometry Geometry
type Geometry struct {
	XMLName     xml.Name `xml:"Geometry"`
	Type        GeometryType
	Coordinates *Coordinates
	Radius      *Radius
}

// Coordinates Coordinates
type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// UnmarshalXML Unmarshal Coordinates
func (c *Coordinates) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	ar := strings.Split(v, ",")
	if len(ar) != 2 {
		return fmt.Errorf("Invalid format %v", v)
	}
	lat, err := strconv.ParseFloat(ar[0], 64)
	if err != nil {
		return err
	}
	lng, err := strconv.ParseFloat(ar[1], 64)
	if err != nil {
		return err
	}
	*c = Coordinates{lat, lng}
	return nil
}

// Radius Radius
type Radius struct {
	Horizontal float64
	Vertical   float64
}

// UnmarshalXML Unmarshal Radius
func (c *Radius) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	ar := strings.Split(v, ",")
	if len(ar) < 2 {
		return fmt.Errorf("Invalid format %v", v)
	}
	hl, err := strconv.ParseFloat(ar[0], 64)
	if err != nil {
		return err
	}
	vt, err := strconv.ParseFloat(ar[1], 64)
	if err != nil {
		return err
	}
	*c = Radius{hl, vt}
	return nil
}

// Record code and name pair
type Record struct {
	Code string
	Name string
}

// Country Country
type Country struct {
	Record
}

// AddressElement AddressElement
type AddressElement struct {
	Record
	Kana  string
	Level string
}

// Building Building
type Building struct {
	ID    string `xml:"Id"`
	Name  string
	Floor string
	Area  string
}

// Road Road
type Road struct {
	Name        string
	Kana        string
	PopularName string
	PopularKana string
}
