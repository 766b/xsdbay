package main

//ebaysvc
var templateEbaySVC = `package ebaysvc

import (
	"encoding/json"
	"bytes"
	"database/sql"
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"io"
	"strings"
)

var (
	APIGateway string

	// X-EBAY-API-COMPATIBILITY-LEVEL
	// Required: Always.
	// The eBay release version that your application supports. See the eBay Schema Versioning Strategy for information about how the version affects the way eBay processes your request.
	APICompatibilityLevel string = "%[1]s"

	// X-EBAY-API-SITEID
	// Required: Always
	// eBay site to which you want to send the request. See SiteCodeType for a list of valid site ID values. This is usually the eBay site an item is listed on or that a user is
	// registered on, depending on the purpose of the call. See Specifying the Target Site to understand how the site ID may affect validation of the call and how it may affect
	// the data that is returned. For calls like AddItem, the site that you pass in the body of the request must be consistent with this header. Note: In AddItem, you specify
	// the 2-letter site code. In this header, you specify the numeric site ID.
	// APISiteID string

	// X-EBAY-API-DEV-NAME
	// Required: Conditionally
	// Your Developer ID (DevID), as registered with the eBay Developers Program. The developer ID is unique to each licensed developer (or company).
	// This value is only required for calls that set up and retrieve a user's authentication token (these calls are: GetSessionID, FetchToken, GetTokenStatus, and RevokeToken).
	// In all other calls, this value is ignored.. If you lose your keys you can retrieve them using the View Keys link on your My Account page. Here is the direct link to the Keys
	// page (requires signin): http://developer.ebay.com/DevZone/account/keys.asp
	APIDevName string

	// X-EBAY-API-APP-NAME
	// Required: Conditionally
	// Your application ID (AppID), as registered with the eBay Developers Program. This value is only required for calls that set up and retrieve a user's authentication
	// token (e.g., FetchToken). In all other calls, this value is ignored. Do not specify this value in AddItem and other calls that list items. The application ID is unique
	// to each application created by the developer. The application ID and certificate ID are issued in pairs. Multiple application/certificate ID pairs can be issued for a
	// single developer ID.
	APIAppName string

	// X-EBAY-API-CERT-NAME
	// Required: Conditionally
	// Your certificate ID (CertID), as registered with the eBay Developers Program. This value is only required for calls that set up and retrieve a user's authentication token
	// (e.g., FetchToken). In all other calls, this value is ignored. Do not specify this value in AddItem and other calls that list items. The certificate ID is unique to each
	// application created by the developer.
	APICertName string

	ErrAPIAppNameNotSet  error = errors.New("APIAppName is not set")
	ErrAPIDevNameNotSet  error = errors.New("APIDevName is not set")
	ErrAPICertNameNotSet error = errors.New("APICertName is not set")
	ErrAPISiteIDNotSet   error = errors.New("APISiteID is not set")
	ErrAPIGatewayNotSet  error = errors.New("APIGateway is not set")

	RequestValidation bool
)

type xbayRequester struct {
	callName string
	siteID   string
	body     *bytes.Buffer
	response interface{}
}

func newRequester(callname, siteID string, response interface{}) *xbayRequester {
	return &xbayRequester{
		callName: callname,
		siteID:   siteID,
		body:     bytes.NewBufferString(xml.Header),
		response: response,
	}
}

func (x *xbayRequester) request() error {
	if x.siteID == "" {
		return ErrAPISiteIDNotSet
	}
	if APIGateway == "" {
		return ErrAPIGatewayNotSet
	}
	client := &http.Client{}
	request, err := http.NewRequest("POST", APIGateway, x.body)
	if err != nil {
		return err
	}

	switch x.callName {
	case "GetSessionID", "FetchToken", "GetTokenStatus", "RevokeToken":
		if APIDevName == "" {
			return ErrAPIDevNameNotSet
		}
		if APIAppName == "" {
			return ErrAPIAppNameNotSet
		}
		if APICertName == "" {
			return ErrAPICertNameNotSet
		}
		request.Header.Add("X-EBAY-API-DEV-NAME", APIDevName)
		request.Header.Add("X-EBAY-API-APP-NAME", APIAppName)
		request.Header.Add("X-EBAY-API-CERT-NAME", APICertName)
	}
	request.Header.Add("X-EBAY-API-COMPATIBILITY-LEVEL", APICompatibilityLevel)
	request.Header.Add("X-EBAY-API-SITEID", x.siteID)
	request.Header.Add("X-EBAY-API-CALL-NAME", x.callName)

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	return xml.NewDecoder(response.Body).Decode(x.response)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

type XmlnsAttr byte

func (m XmlnsAttr) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{name, "urn:ebay:apis:eBLBaseComponents"}, nil
}
`

var templateNulls = `
type NullInt64 struct {
	sql.NullInt64
}

type NullInt64List []NullInt64

func (l *NullInt64List) Append(value ...int64) *NullInt64List {
	for _, v := range value {
		n := NullInt64{}
		n.Set(v)
		*l = append(*l, n)
	}
	return l
}

func (n *NullInt64) Set(value int64) {
	n.Int64 = value
	n.Valid = true
}

func (n NullInt64) Value() int64 {
	return n.Int64
}

func (n NullInt64) String() string {
	if !n.Valid {
		return ""
	}
	return strconv.FormatInt(n.Int64, 10)
}

func (n NullInt64) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	if !n.Valid {
		return
	}
	return e.EncodeElement(n.Int64, start)
}

func (n NullInt64) MarshalJSON() (value []byte, e error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return strconv.AppendInt(value, n.Int64, 10), nil
}

func (n NullInt64) MarshalText() (value []byte, err error) {
	if !n.Valid {
		return nil, nil
	}

	return strconv.AppendInt(value, n.Int64, 10), nil
}

func (n *NullInt64) UnmarshalText(text []byte) (err error) {
	if text == nil {
		n.Valid = false
		return
	}

	if n.Int64, err = strconv.ParseInt(string(text), 10, 64); err != nil {
		n.Int64 = 0
		n.Valid = false
		return
	}
	n.Valid = true
	return
}

type NullStringList []NullString

func (l *NullStringList) Append(value ...string) *NullStringList {
	for _, v := range value {
		n := NullString{}
		n.Set(v)
		*l = append(*l, n)
	}
	return l
}

type NullString struct {
	sql.NullString
}

func (n *NullString) Set(value string) {
	n.NullString.String = value
	n.Valid = true
}

func (n NullString) Value() string {
	return n.NullString.String
}

func (n NullString) String() string {
	if !n.Valid {
		return ""
	}
	return n.NullString.String
}

func (n NullString) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	if !n.Valid {
		return
	}
	return e.EncodeElement(n.NullString.String, start)
}

func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(n.NullString.String)
}

func (n NullString) MarshalText() ([]byte, error) {
	if !n.Valid {
		return nil, nil
	}
	return []byte(n.NullString.String), nil
}

func (n *NullString) UnmarshalText(text []byte) (err error) {
	if text == nil {
		n.Valid = false
		return
	}
	n.NullString.String = string(text)
	n.Valid = true
	return
}

type NullFloat64 struct {
	sql.NullFloat64
}


type NullFloat64List []NullFloat64

func (l *NullFloat64List) Append(value ...float64) *NullFloat64List {
	for _, v := range value {
		n := NullFloat64{}
		n.Set(v)
		*l = append(*l, n)
	}
	return l
}

func (n *NullFloat64) Set(value float64) {
	n.Float64 = value
	n.Valid = true
}

func (n NullFloat64) Value() float64 {
	return n.Float64
}

func (n NullFloat64) String() string {
	if !n.Valid {
		return ""
	}
	return strconv.FormatFloat(n.Float64, 'f', -1, 64)
}

func (n NullFloat64) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	if !n.Valid {
		return
	}
	return e.EncodeElement(n.Float64, start)
}

func (n NullFloat64) MarshalJSON() (value []byte, e error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return strconv.AppendFloat(value, n.Float64, 'f', -1, 64), nil
}

func (n NullFloat64) MarshalText() (value []byte, err error) {
	if !n.Valid {
		return
	}
	return strconv.AppendFloat(value, n.Float64, 'f', -1, 64), nil
}

func (n *NullFloat64) UnmarshalText(text []byte) (err error) {
	if text == nil {
		n.Valid = false
		return
	}
	if n.Float64, err = strconv.ParseFloat(string(text), 64); err != nil {
		n.Float64 = 0
		n.Valid = false
		return
	}
	n.Valid = true
	return
}

type NullBool struct {
	sql.NullBool
}

func (n *NullBool) Set(value bool) {
	n.Bool = value
	n.Valid = true
}

func (n NullBool) Value() bool {
	return n.Bool
}

func (n NullBool) String() string {
	if !n.Valid {
		return ""
	}
	if n.Bool {
		return "true"
	}
	return "false"
}

func (n NullBool) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	if !n.Valid {
		return
	}
	return e.EncodeElement(n.Bool, start)
}


func (n NullBool) MarshalJSON() (value []byte, e error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return strconv.AppendBool(value, n.Bool), nil
}

func (n NullBool) MarshalText() (value []byte, err error) {
	if !n.Valid {
		return
	}
	return strconv.AppendBool(value, n.Bool), nil
}

func (n *NullBool) UnmarshalText(text []byte) (err error) {
	if text == nil {
		n.Valid = false
		return
	}
	switch strings.ToLower(string(text)) {
	case "false":
		n.Bool = false
	case "true":
		n.Bool = true
	default:
		n.Bool = false
		return
	}
	n.Valid = true
	return
}
`
