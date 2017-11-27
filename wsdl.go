package main

type definitions struct {
	Name            string `xml:"name,attr"`
	TargetNamespace string `xml:"targetNamespace,attr"`
	XMime           string `xml:"xmine,attr"`
	Soap12          string `xml:"soap12,attr"`
	XS              string `xml:"xs,attr"`
	Soap            string `xml:"soap,attr"`
	TNS             string `xml:"tns,attr"`
	HTTP            string `xml:"http,attr"`
	WSDL            string `xml:"wsdl,attr"`
	Mine            string `xml:"mine,attr"`
	XmlNS           string `xml:"xmlns,attr"`

	Types types `xml:"types"`

	Service service `xml:"service"`
}

type types struct {
	Schema schema `xml:"schema"`
}

type service struct {
	Name          string        `xml:"name,attr"`
	Documentation documentation `xml:"documentation"`
}
