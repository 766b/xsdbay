xsdbay
===
`xsdbay` generates Go structs and helper methods based on eBay's `ebaysvc.xsd` file.

Installation
---

    go install github.com/766b/xsdbay

Usage
---

    -i (string, optional)
        Input file
    -o (string, optional)
        Output Go file (Default: ebaysvc_####.go)
    -e (string, optional)
        Elements to be exported (comma separated)    
    -latest
        Download latest XSD version
    -cache-xsd
        Cache downloaded XSD file
    -apiver (string, optional)
        API Version
    -download (string, optional)
        XSD link (default "http://developer.ebay.com/webservices/latest/ebaysvc.xsd")    

Examples
---

    xsdbay -latest

Download latest version and generate output for all API calls.

    xsdbay -latest -e "AddItem"

Download latest version and generate output for AddItem call, related elements and code types.

Request Helper Methods
---
    func (*RequestType) Request(eBayAuthToken, siteID string) (response *ResponseType, err error)
    func (*RequestType) MarshalXMLEncode(w io.Writer) error
    func (*RequestType) MarshalXML() ([]byte, error)
    func (*RequestType) Validate() error

Response Helper Methods
---
    func (x *ResponseType) Success() bool
    func (x *ResponseType) Failure() bool
    func (x *ResponseType) Warning() bool
    func (x *ResponseType) PartialFailure() bool

CodeType Helper Methods
---
    var *CodeTypeList = [...]string{...}
    func (*CodeType) Set(value string) error
    func (*CodeType) String() string

eBay Settings
---
    var (
        // API gateway address. Sandbox or production.
        // Default: not set
	    APIGateway string

        // API compatibility.
        // If in sandbox mode, use lowest version possible.
        // Default: Variable is autofilled based on XSD file version.
        APICompatibilityLevel string = "1033"

        // eBay Crededentials needed for authentication calls. Refer to API docs.
        APIDevName string
        APIAppName string
        APICertName string

        // Validate input based on defined rules in XSD file.
        // Validation is not guaranteed to be error free and/or catch incorrectly 
        // enetered data that might cause request to fail.
        // Default: false
        RequestValidation bool
    )