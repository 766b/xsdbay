package main

import (
	"fmt"
	"log"
	"strings"
)

func (c attribute) Generate() {
	fmt.Printf("%s\r\n", c.GoLine())
	if y := c.GetRelated(); y != nil {
		y.Generate()
		for _, e := range y.GetElements() {
			e.Generate()
		}
	}
}

func (e attribute) Setter(callName string) {}

func (e attribute) DeepValidator(callName, path string) bool {
	return false
	if e.Annotation.AppInfo.MaxDepth != 0 {
		return false
	}
	if !e.Annotation.RequiredInput() {
		return false
	}
	path = path + "." + e.Name
	if _, yes := e.NeedsValidation(callName); yes {
		return true
	}
	return false
}

func (e attribute) Validator(callName, path string) {
	if rules, ok := e.NeedsValidation(callName); ok {
		details := e.TypeDetails()
		for _, rule := range rules {
			Validator[callName].Sprintf("%s\r\n", details.ValidationString(rule, path))
			if x := e.GetRelated(); x != nil {
				x.Validator(callName, path+"."+details.Field)
			}
		}
	}
}

func (e attribute) TypeDetails() *TypeDetails {
	t := &TypeDetails{}
	t.Field = e.GetName()
	t.Type = e.GetType()

	if e.Annotation != nil {
		if listBasedOn, ok := e.Annotation.AppInfo.ListBasedOn(); ok {
			if !strings.Contains(listBasedOn, ",") {
				if x, ok := FindSimple(listBasedOn); ok {
					t.AliasFor = x.GetType()
				} else {
					log.Fatalf("could not find simple type -`%s`- ", listBasedOn)
				}
			}
		} else {
			if x, ok := FindSimple(e.GetType().String()); ok {
				t.SimpleType = true
				t.AliasFor = x.GetType()
			}
		}
	}
	if t.AliasFor == "" {
		t.AliasFor = e.GetType()
	}
	t.IsPointer = !e.GetType().Nullable() && !t.IsSlice
	return t
}

func (c attribute) GoLine() string {
	goType := c.GetType().GoType(false)
	if c.Use == "optional" {
		goType = c.GetType().GoType()
	}
	return fmt.Sprintf("%[1]s %[2]s `xml:\"%[3]s,attr,omitempty\" json:\"%[4]s,omitempty\"` //attribute", UpperFirstLetter(c.GetName()), goType, c.GetName(), ToSnake(c.GetName()))
}

func (c attribute) GetName() string {
	return c.Name
}

func (c attribute) GetRelated() Xyer {
	if c.Type.IsNS() {
		return Find(c.Type.String())
	}
	return nil
}

func (c attribute) GetType() Type {
	return c.Type
}

func (c attribute) GetElements() (r []Xyer) {
	return
}
