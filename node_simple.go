package main

import (
	"fmt"
	"strings"
)

var list = [...]string{"a1", "2"}

func (c simpleType) Generate() {
	if _, ok := Enums[c.GetName()]; ok {
		return
	}
	Enums[c.GetName()] = NewBuffer()
	Funcs[c.GetName()] = NewBuffer()
	cleanName := UpperFirstLetter(strings.TrimSuffix(c.GetName(), "CodeType"))
	Enums[c.GetName()].Sprintf("type %s %s\r\n", c.GetName(), c.GetType().GoType(true))
	if c.GetType().GoType(true) == "string" {
		Funcs[c.GetName()].Sprintf("func (x %s) String() string { return string(x) }", UpperFirstLetter(c.GetName()))
	}
	if c.Restriction != nil && len(c.Restriction.Enumeration) > 0 {
		Enums[c.GetName()].Sprintf("const (\r\n")
		Funcs[c.GetName()+"List"] = NewBuffer()
		Funcs[c.GetName()+"List"].Sprintf("var %sList = [...]string{", UpperFirstLetter(c.GetName()))
		for i, e := range c.Restriction.Enumeration {
			if i == 0 {
				Enums[c.GetName()].Sprintf("\t%[1]s_%[3]s %[4]s = \"%[2]s\"\r\n", cleanName, e.Value, UpperFirstLetter(e.Value), UpperFirstLetter(c.GetName()))
				Funcs[c.GetName()+"List"].Sprintf("\"%s\"", e.Value)
				continue
			}
			if e.Value == "CustomCode" {
				continue
			}
			if i > 0 {
				Funcs[c.GetName()+"List"].Sprintf(",")
			}

			Funcs[c.GetName()+"List"].Sprintf("\"%s\"", e.Value)
			Enums[c.GetName()].Sprintf("\t%s_%s = \"%s\"\r\n", cleanName, UpperFirstLetter(e.Value), e.Value)
		}
		Funcs[c.GetName()+"List"].Sprintf("}")

		Funcs[c.GetName()+"Helper"] = NewBuffer()
		Funcs[c.GetName()+"Helper"].Sprintf("func (x *%s) Set(value string) error {\r\n", UpperFirstLetter(c.GetName()))
		Funcs[c.GetName()+"Helper"].Sprintf(`if contains(%[1]sList[:], value) { 
					*x = %[1]s(value)
					return nil
				} else {
					return errors.New("invalid value for %[1]s")
				}
			}`, UpperFirstLetter(c.GetName()))

		Enums[c.GetName()].Sprintf(")\r\n")
	}
}

func (e simpleType) Setter(typeName string) {}

func (e simpleType) DeepValidator(callName, path string) bool {
	if e.Annotation.RequiredFor(callName) {
		return true
	}
	return false
}

func (e simpleType) Validator(callName, path string) {
	if e.Annotation.RequiredFor(callName) {
		Validator[callName].Sprintf("//%s.%s // Simple: %s\r\n", path, e.GetName(), callName)
	}
}

func (c simpleType) GoLine() string {
	return fmt.Sprintf("//8%s %s //simple", UpperFirstLetter(c.GetName()), c.GetType())
}

func (c simpleType) GetName() string {
	return c.Name
}

func (c simpleType) GetRelated() Xyer {
	return nil
}

func (c simpleType) GetType() Type {
	return c.Restriction.Base
}

func (c simpleType) GetElements() (r []Xyer) {
	return nil
}
