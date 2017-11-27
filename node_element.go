package main

import (
	"fmt"
	"log"
	"strings"
)

func (c element) Generate() {
	fmt.Printf("%s %s\r\n", c.GetName(), c.GetType())
	if y := c.GetRelated(); y != nil {
		y.Generate()
		for _, e := range y.GetElements() {
			e.Generate()
		}
	}
}

func (e element) Setter(typeName string) {
	if strings.HasSuffix(typeName, "RequestType") {
		if e.TypeDetails().IsSlice {
			funcIdx := fmt.Sprintf("%s_Append%s", typeName, UpperFirstLetter(e.GetName()))
			Funcs[funcIdx] = NewBuffer()

			if _, yes := SliceableType[e.GetType().GoType()]; yes {
				Funcs[funcIdx].Sprintf(`func (x *%[1]s) Append%[2]s(v ...%[3]s) {
					x.%[2]s.Append(v...)
				}
				`, typeName, UpperFirstLetter(e.GetName()), e.GetType().GoType(false))
			} else {
				Funcs[funcIdx].Sprintf(`func (x *%[1]s) Append%[2]s(v ...%[3]s) {
					x.%[2]s = append(x.%[2]s, v...)
				}
				`, typeName, UpperFirstLetter(e.GetName()), e.GetType().GoType())
			}
		}
	}
	if strings.HasSuffix(typeName, "ResponseType") {
		if e.GetType().GoType() == "AckCodeType" && e.GetName() == "Ack" {
			if splx, ok := FindSimple("AckCodeType"); ok {
				funcIdx := fmt.Sprintf("%s_AckCodeType%s", typeName, UpperFirstLetter(e.GetName()))
				Funcs[funcIdx] = NewBuffer()
				funcCount := 0
				for _, e := range splx.Restriction.Enumeration {
					if e.Annotation.Skip() || e.Value == "CustomCode" {
						continue
					}
					Funcs[funcIdx].Sprintf(`func (x %[1]s) %[2]s() bool {
						return x.Ack == Ack_%[2]s
					}
					`, typeName, UpperFirstLetter(e.Value))
					funcCount++
				}
				if funcCount == 0 {
					delete(Funcs, funcIdx)
				}
			}
		}
	}
}

func (e element) DeepValidator(callName, path string) bool {
	path = path + "." + e.Name

	if e.GetType().IsBasic() && e.TypeDetails().IsSlice {
		return false
	}
	if _, yes := e.NeedsValidation(callName); yes {
		return true
	}
	return false
}

func (e element) Validator(callName, path string) {
	rules2, ok := e.NeedsValidation(callName)
	if !ok {
		return
	}
	rules := rules2.Except(ValTypRequired, ValTypMaxOccurs)
	related := e.GetRelated()
	key := ""
	newPath := fmt.Sprintf("%s.%s", path, UpperFirstLetter(e.GetName()))
	hasDeepValidationRequirement := false
	loopBracket, pointerBracket := false, false
	if related != nil {
		if e.Annotation.AppInfo.MaxDepth == 0 {
			hasDeepValidationRequirement = related.DeepValidator(callName, newPath)
		}
	}

	if rule, yes := rules2.Includes(ValTypMaxOccurs); yes {
		Validator[callName].Sprintf("%s", e.TypeDetails().Key(key).ValidationString(*rule, path))
	}
	if rule, yes := rules2.Includes(ValTypRequired); yes {
		Validator[callName].Sprintf("%s", e.TypeDetails().Key(key).ValidationString(*rule, path))
	}

	if e.TypeDetails().IsPointer && hasDeepValidationRequirement {
		pointerBracket = true
		Validator[callName].Sprintf("if %s != nil {\r\n", newPath)
	}

	if e.TypeDetails().IsSlice {
		key = fmt.Sprintf("i%d", hash(path))
		newPath = path + "." + UpperFirstLetter(e.GetName()) + "[" + key + "]"

		if rules.Len() > 0 || hasDeepValidationRequirement {
			loopBracket = true
			Validator[callName].Sprintf("for %s := range %s {\r\n", key, fmt.Sprintf("%s.%s", path, UpperFirstLetter(e.GetName())))
		} else {
			//Validator[callName].Sprintf("// No validation for %s %+v | %v\r\n", newPath, rules, rules2)
		}
	}

	//for _, r := range rules {
	//Validator[callName].Sprintf("//\r\n%s", e.TypeDetails().Key(key).ValidationString(r, path))
	//}

	if related != nil {
		related.Validator(callName, newPath)
	}

	if loopBracket {
		Validator[callName].Sprintf("}\r\n")
	}
	if pointerBracket {
		Validator[callName].Sprintf("}\r\n")
	}
}

func (e element) TypeDetails() *TypeDetails {
	t := &TypeDetails{}
	t.Field = e.GetName()
	t.Type = e.GetType()
	_, t.IsSlice = e.SliceLen()
	if e.Annotation != nil {
		if listBasedOn, ok := e.Annotation.AppInfo.ListBasedOn(); ok {
			if !strings.Contains(listBasedOn, ",") {
				if x, ok := FindSimple(listBasedOn); ok {
					t.SimpleType = true
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

func (e element) TransformType() string {
	var let Type = e.Type

	if e.Annotation != nil {
		if listBasedOn, ok := e.Annotation.AppInfo.ListBasedOn(); ok {
			if !strings.Contains(listBasedOn, ",") {
				if Type(listBasedOn).Nullable() {
					let = Type(listBasedOn)
				}
				if x, ok := FindSimple(listBasedOn); ok {
					x.Generate()
				}
			} else {
				listBasedOn = strings.Replace(listBasedOn, " ", "", -1)
				for _, k := range strings.Split(listBasedOn, ",") {
					if x, ok := FindSimple(k); ok {
						x.Generate()
					}
				}
			}
		}
	}

	if _, yes := e.SliceLen(); yes {
		if s, yes := SliceableType[let.GoType()]; yes {
			return s
		}
		return "[]" + let.GoType()
	}
	if let.Nullable() {
		return let.GoType()
	}

	return "*" + let.GoType()
}

func (c element) GoLine() string {
	return fmt.Sprintf("%s %s `xml:\"%s,omitempty\" json:\"%s,omitempty\"`", UpperFirstLetter(c.GetName()), c.TransformType(), c.GetName(), ToSnake(c.GetName()))
}

func (c element) GetName() string {
	return c.Name
}

func (c element) GetRelated() Xyer {
	if c.Type.IsXS() {
		return nil
	}
	return Find(c.Type.String())
}

func (c element) GetType() Type {
	return c.Type
}

func (c element) GetElements() (r []Xyer) {
	return nil
}
