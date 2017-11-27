package main

import (
	"fmt"
	"log"
	"strings"
)

func (c complexType) Generate() {
	if _, ok := Types[c.GetName()]; ok {
		return
	}

	Types[c.GetName()] = NewBuffer()
	Types[c.GetName()].Sprintf("type %s struct {\r\n", c.GetType())
	if strings.HasSuffix(c.GetName(), "RequestType") && !c.Abstract && contains(exportedElements, strings.TrimSuffix(c.GetName(), "RequestType")) {
		Types[c.GetName()].Sprintf("\tXMLName	xml.Name `xml:\"" + strings.TrimSuffix(c.Name, "Type") + "\" json:\"-\"`\r\n")
		Types[c.GetName()].Sprintf("\tXmlnsAttr `xml:\"xmlns,attr\" json:\"-\"`\r\n\r\n")
	}

	for _, e := range c.GetElements() {
		Types[c.GetName()].Sprintf("\t%s\r\n", e.GoLine())
		if r := e.GetRelated(); r != nil {
			defer r.Generate()
		}
	}

	Types[c.GetName()].Sprintf("}\r\n")
	if y := c.GetRelated(); y != nil {
		defer y.Generate()
	}
}

func (e complexType) Setter(typeName string) {
	// if e.GetType().IsRequest() {
	for _, i := range e.GetElements() {
		i.Setter(e.GetName())
	}
	// }
}

func (e complexType) DeepValidator(callName, path string) bool {
	for _, x := range e.GetElements() {
		pathX := path + "." + e.Name
		if x.GetType().String() == e.Name {
			continue
		}
		if x.DeepValidator(callName, pathX) {
			return true
		}
	}
	return false
}

func (e complexType) Validator(callName, path string) {
	if _, ok := Validator[callName]; !ok {
		Validator[callName] = NewBuffer()
	}
	if path == "" {
		path = "x"
	}

	for _, f := range e.GetElements() {
		f.Validator(callName, path)
	}
}

func (c complexType) GoLine() string {
	if strings.HasPrefix(c.GetName(), "Abstract") && c.GetName() == c.GetType().String() {
		return c.GetType().String() + "//complexType"
	}

	return fmt.Sprintf("%s %s //complexType", c.GetName(), c.GetType().String())
}

func (c complexType) GetName() string {
	return c.Name
}

func (c complexType) GetRelated() Xyer {
	return Find(c.GetType().String())
}

func (c complexType) GetType() Type {
	return Type(c.Name)
}

func (c complexType) GetElements() (r []Xyer) {
	var callID string
	var request bool
	if strings.HasSuffix(c.Name, "RequestType") && !c.Abstract {
		callID = strings.TrimSuffix(c.Name, "RequestType")
		request = true
	}
	if strings.HasSuffix(c.Name, "ResponseType") && !c.Abstract {
		callID = strings.TrimSuffix(c.Name, "ResponseType")
	}
	if c.SimpleContent != nil {
		for _, a := range c.SimpleContent.Extension.Attribute {
			if a.Annotation.Skip() || !a.Annotation.IncludedIn(callID, request) {
				continue
			}
			r = append(r, a)
		}
		r = append(r, c.SimpleContent.Extension)
	}
	if c.ComplexContent != nil {
		if c.ComplexContent.Restriction != nil {
			log.Fatal("complexType.ComplexContent.Restriction: not null at ", c.Name)
		}

		if c.ComplexContent.Extension != nil {
			if base, ok := FindComplex(c.ComplexContent.Extension.Base.String()); !ok {
				log.Fatalf("could not find complex type: %s", c.GetType())
			} else {
				for _, e := range base.Sequence.Element {
					if e.Annotation.Skip() || !e.Annotation.IncludedIn(callID, request) {
						continue
					}
					r = append(r, e)
				}
			}

			if len(c.ComplexContent.Extension.Attribute) > 0 {
				log.Fatal("BASE.Attribute: not null at ", c.Name)
			}

			for _, s := range c.ComplexContent.Extension.Sequence.Element {
				if s.Annotation.Skip() || !s.Annotation.IncludedIn(callID, request) {
					continue
				}
				r = append(r, s)
			}
		}
	}
	if c.Sequence != nil {
		for _, e := range c.Sequence.Element {
			if e.Annotation.Skip() || !e.Annotation.IncludedIn(callID, request) {
				continue
			}
			r = append(r, e)
		}
	}
	return
}
