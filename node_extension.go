package main

import (
	"fmt"
)

func (c extensionSimpleContent) Generate() {
	for _, e := range c.GetElements() {
		e.Generate()
	}
	if y := c.GetRelated(); y != nil {
		y.Generate()
	}
}

func (e extensionSimpleContent) Setter(typeName string) {}

func (e extensionSimpleContent) DeepValidator(callName, path string) bool {
	if e.Annotation.RequiredFor(callName) {
		return true
	}
	return false
}

func (e extensionSimpleContent) Validator(callName, path string) {
	if e.Annotation.RequiredFor(callName) {
		Validator[callName].Sprintf("//extensionSimpleContent.Validator %s %s\r\n", callName, path)
	}
}

func (c extensionSimpleContent) GoLine() string {
	return fmt.Sprintf("%s %s `xml:\",chardata\" json:\"%s,omitempty\"`", UpperFirstLetter(c.GetName()), c.GetType().GoType(), ToSnake(c.GetName()))
}

func (c extensionSimpleContent) GetName() string {
	return "Value"
}

func (c extensionSimpleContent) GetRelated() Xyer {
	if c.Base.IsNS() {
		return Find(c.Base.String())
	}
	return nil
}

func (c extensionSimpleContent) GetType() Type {
	return c.Base
}

func (c extensionSimpleContent) GetElements() (r []Xyer) {
	for _, a := range c.Attribute {
		r = append(r, a)
	}
	return
}
