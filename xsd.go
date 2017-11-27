package main

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var TypeMap map[string]string = map[string]string{
	"other":        "string",
	"token":        "string",
	"dateTime":     "string",
	"duration":     "string",
	"time":         "string",
	"anyURI":       "string",
	"base64Binary": "[]byte",
	"string":       "string",
	"boolean":      "bool",
	"float":        "float64", //32
	"double":       "float64",
	"decimal":      "float64",
	"int":          "int64", //32
	"long":         "int64",
}

var SubstituteMap map[string]string = map[string]string{
	"int32":   "NullInt64",
	"int64":   "NullInt64",
	"string":  "NullString",
	"float32": "NullFloat64",
	"float64": "NullFloat64",
	"bool":    "NullBool",
}

var NullableType map[string]bool = map[string]bool{
	"[]byte":      true,
	"string":      true,
	"time.Time":   true,
	"NullInt64":   true,
	"NullString":  true,
	"NullFloat64": true,
	"NullBool":    true,
}

var SliceableType map[string]string = map[string]string{
	"NullString":  "NullStringList",
	"NullInt64":   "NullInt64List",
	"NullFloat64": "NullFloat64List",
}

type Type string

func (e Type) Nullable() bool {
	if v, k := NullableType[e.GoType()]; k {
		return v
	}
	s, ok := FindSimple(e.String())
	if ok {
		if v, k := NullableType[s.Restriction.Base.GoType()]; k {
			return v
		}
	}
	return false
}

func (e Type) IsRequest() bool {
	return strings.HasSuffix(string(e), "RequestType")
}

func (e Type) IsBasic() bool {
	if _, k := TypeMap[e.String()]; k {
		return true
	}
	return false
}

func (e Type) IsNS() bool {
	if strings.Contains(string(e), ":") {
		return strings.SplitAfterN(string(e), ":", 2)[0] != "xs:"
	}
	return e[:3] == "ns:"
}

func (e Type) IsXS() bool {
	if len(string(e)) < 4 {
		panic(string(e))
	}
	return e[:3] == "xs:"
}

func (e Type) IsSimpleType() bool {
	if e.IsXS() {
		return false
	}
	_, ok := FindSimple(e.String())
	return ok
}

func (e Type) IsComplexType() bool {
	if e.IsXS() {
		return false
	}
	_, ok := FindComplex(e.String())
	return ok
}

func (e Type) String() string {
	if strings.Contains(string(e), ":") {
		return strings.SplitAfterN(string(e), ":", 2)[1]
	}
	return string(e)
}

func (s Type) GoType(noSub ...bool) string {
	if !s.IsXS() {
		return s.String()
	}

	t := ""
	k := false
	if t, k = TypeMap[s.String()]; !k {
		log.Fatal("could not find go type for ", s.String())
	}
	if len(noSub) == 0 {
		if tSub, k := SubstituteMap[t]; k {
			t = tSub
		}
	}
	return t
}

type schema struct {
	attributeFormDefault string `xml:"attributeFormDefault,attr"`
	BlockDefault         string `xml:"blockDefault,attr"`
	ElementFormDefault   string `xml:"elementFormDefault,attr"`
	FinalDefault         string `xml:"finalDefault,attr"`
	Id                   string `xml:"id,attr"`
	TargetNamespace      string `xml:"targetNamespace,attr"`
	Xmlns                string `xml:"xmlns,attr"`
	Version              string `xml:"version,attr"`
	XmlLang              string `xml:"lang,attr"`

	//Include    []includeMany `xml:"include"`
	//Import     []importMany  `xml:"import"`
	Annotation annotation `xml:"annotation"`
	//redefine
	Attribute []attribute `xml:"attribute"`
	//attributeGroup
	Element []element `xml:"element"`
	//group
	//notation
	SimpleType  []simpleType  `xml:"simpleType"`
	ComplexType []complexType `xml:"complexType"`
}

type element struct {
	// Abstract          string `xml:"abstract,attr"`
	// Block             string `xml:"block,attr"`
	// Default           string `xml:"default,attr"`
	// SubstitutionGroup string `xml:"substitutionGroup,attr"`
	// Final             string `xml:"final,attr"`
	// Fixed             string `xml:"fixed,attr"`
	// Form              string `xml:"form,attr"`
	// ID                string `xml:"id,attr"`
	MaxOccurs string `xml:"maxOccurs,attr"`
	MinOccurs string `xml:"minOccurs,attr"`
	Name      string `xml:"name,attr"`
	Nillable  bool   `xml:"nillable,attr"`
	// Ref               string `xml:"ref,attr"`
	Type Type `xml:"type,attr"`

	Annotation  *annotation  `xml:"annotation"`
	SimpleType  *simpleType  `xml:"simpleType"`
	ComplexType *complexType `xml:"complexType"`
	//key
	//keyref
	//unique
}

func (e element) SliceLen() (int, bool) {
	if strings.Contains(string(e.Type), "base64Binary") {
		return 0, true
	}
	if e.MaxOccurs == "" {
		return 0, false
	}
	if e.MaxOccurs == "unbounded" {
		return 0, true
	}
	mo, err := strconv.Atoi(e.MaxOccurs)
	if err != nil {
		log.Fatal(err)
	}
	return mo, mo > 1
}

// https://msdn.microsoft.com/en-us/library/ms256067(v=vs.110).aspx
//  Number of occurrences: Unlimited within schema; one time within element.
type complexType struct {
	Abstract bool `xml:"abstract,attr"`
	// Block    string `xml:"block,attr"`
	// Final    string `xml:"final,attr"`
	// ID       string `xml:"id,attr"`
	// Mixed    string `xml:"mixed,attr"`
	Name string `xml:"name,attr"`

	Annotation annotation `xml:"annotation"`

	// The complex type has character data or a simpleType as content and contains no elements, but may contain attributes.
	SimpleContent *simpleContent `xml:"simpleContent"`

	// The complex type contains only elements or no element content (empty).
	ComplexContent *complexContent `xml:"complexContent"`
	//group
	//all
	//choice

	// The complex type contains the elements defined in the specified sequence.
	Sequence  *sequence   `xml:"sequence"`
	Attribute []attribute `xml:"attribute"`
	//attributeGroup
	//anyAttribute
}

// https://msdn.microsoft.com/en-us/library/ms256053(v=vs.110).aspx
// Number of occurrences: One time
// Optional. annotation
// Required. One and only one of the following elements: restriction (complexContent), or extension (complexContent).
type complexContent struct {
	ID    string `xml:"id,attr"`
	Mixed string `xml:"mixed,attr"`

	Annotation annotation `xml:"annotation"`

	Restriction *restrictioncomplexContent `xml:"restriction"` //OR
	Extension   *extensioncomplexContent   `xml:"extension"`   //
}

// https://msdn.microsoft.com/en-us/library/ms256061(v=vs.110).aspx
// Number of occurrences: One time
type restrictioncomplexContent struct {
	Base Type   `xml:"base,attr"`
	Id   string `xml:"id,attr"`

	//group
	//all
	//choice
	Sequence  sequence    `xml:"sequence"`
	Attribute []attribute `xml:"attribute"`
	//attributeGroup
	//anyAttribute
}

// https://msdn.microsoft.com/en-us/library/ms256161(v=vs.110).aspx
// Number of occurrences: One time
type extensioncomplexContent struct {
	Base Type   `xml:"base,attr"`
	ID   string `xml:"id,attr"`

	Annotation annotation  `xml:"annotation"`
	Attribute  []attribute `xml:"attribute"`
	//attributeGroup
	//anyAttribute
	//choice
	//all
	Sequence sequence `xml:"sequence"`
	//group
}

// https://msdn.microsoft.com/en-us/library/ms256106(v=vs.110).aspx
// Number of occurrences: One time
// Optional — annotation
// Required — One and only one of the following elements: restriction (simpleContent), or extension (simpleContent).
type simpleContent struct {
	ID string `xml:"id,attr"`

	Annotation annotation `xml:"annotation"`

	Restriction *restrictionSimpleContent `xml:"restriction"` //OR
	Extension   *extensionSimpleContent   `xml:"extension"`   //
}

// https://msdn.microsoft.com/en-us/library/ms256056(v=vs.110).aspx
// Number of occurrences: One time
type extensionSimpleContent struct {
	Base Type   `xml:"base,attr"`
	ID   string `xml:"id,attr"`

	Annotation annotation  `xml:"annotation"`
	Attribute  []attribute `xml:"attribute"`
	//attributeGroup
	//anyAttribute
}

// https://msdn.microsoft.com/en-us/library/ms256219(v=vs.110).aspx
// Number of occurrences: One time
type restrictionSimpleContent struct {
	Base string `xml:"base,attr"`
	Id   string `xml:"id,attr"`

	Annotation annotation `xml:"annotation"`
	//fractionDigits
	Enumeration []enumeration `xml:"enumeration"`
	//length
	//maxExclusive
	//maxInclusive
	//maxLength
	//minExclusive
	//minInclusive
	//minLength
	//pattern
	SimpleType simpleType `xml:"simpleType"`
	//totalDigits
	//whiteSpace
	Attribute attribute `xml:"attribute"`
	//attributeGroup
	//anyAttribute
}

// https://msdn.microsoft.com/en-us/library/ms256219(v=vs.110).aspx
type enumeration struct {
	Value string `xml:"value,attr"`

	Annotation annotation `xml:"annotation"`
}

// https://msdn.microsoft.com/en-us/library/ms256143(v=vs.110).aspx
// Number of occurrences: Defined one time in the schema element. Referred to multiple times in complex types or attribute groups.
type attribute struct {
	Default string `xml:"default,attr"`
	Fixed   string `xml:"fixed,attr"`
	Form    string `xml:"form,attr"`
	ID      string `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	Ref     string `xml:"ref,attr"`
	Type    Type   `xml:"type,attr"`
	Use     string `xml:"use,attr"`

	Annotation *annotation  `xml:"annotation"`
	SimpleType []simpleType `xml:"simpleType"`
}

type simpleType struct {
	Final string `xml:"final,attr"`
	ID    string `xml:"id,attr"`
	Name  string `xml:"name,attr"`

	Annotation annotation `xml:"annotation"`
	//list,
	Restriction *restrictionSimpleType `xml:"restriction"`
	//union
}

type restrictionSimpleType struct {
	Base Type   `xml:"base,attr"`
	ID   string `xml:"id,attr"`

	Annotation annotation `xml:"annotation"`
	//fractionDigits
	Enumeration []enumeration `xml:"enumeration"` //
	//length
	//maxExclusive
	//maxInclusive
	//maxLength
	//minExclusive
	//minInclusive
	//minLength
	//pattern
	SimpleType simpleType `xml:"simpleType"`
	//totalDigits
	//whiteSpace
}

// type include struct {
// 	Id             string `xml:"Id,attr"`
// 	SchemaLocation string `xml:"schemaLocation,attr"`

// 	Annotation annotation
// }

// type importT struct {
// 	Id             string `xml:"id,attr"`
// 	Namespace      string `xml:"namespace,attr"`
// 	SchemaLocation string `xml:"schemaLocation,attr"`

// 	Annotation annotation
// }

// https://msdn.microsoft.com/en-us/library/ms256102(v=vs.110).aspx
type annotation struct {
	AppInfo       appInfo         `xml:"appinfo"`
	Documentation []documentation `xml:"documentation"`
}

func (e annotation) IncludedIn(call string, request bool) bool {
	if call == "" {
		return true
	}
	for _, a := range e.AppInfo.CallInfo {
		if a.AllCallsExcept != "" {
			excepts := strings.Split(strings.Replace(a.AllCallsExcept, " ", "", -1), ",")
			if contains(excepts, call) {
				return false
			}
			return true
		}
		if a.AllCalls != nil {
			if request {
				return a.RequiredInput != ""
			} else { //response
				return a.Returned != ""
			}
		}
		if request {
			if contains(a.CallName, call) {
				return a.RequiredInput != ""
			}
		} else { //response
			if contains(a.CallName, call) {
				return a.Returned != ""
			}
		}
	}
	return false
}

func (e attribute) NeedsValidation(callName string) (ValidationContainer, bool) {
	list := ValidationContainer{}

	if e.Annotation == nil {
		return list, false
	}
	a := e.Annotation
	if a.RequiredFor(callName) || e.Use == "required" {
		list.New(ValTypRequired, nil)
	}

	if nlist, ok := a.AppInfo.ValidationRules(callName); ok {
		list = append(list, nlist...)
	}

	return list, list.Len() > 0
}

type ValidationRule struct {
	Type  ValidationType
	Value interface{}
}

func (v ValidationRule) ValueInt() (int, error) {
	strValue := v.Value.(string)
	if strValue == "length of longest name in ShippingRegionCodeType and CountryCodeType" {
		codeTypes := []string{"ShippingRegionCodeType", "CountryCodeType"}
		var biggestInt int
		for _, k := range codeTypes {
			if splx, ok := FindSimple(k); ok {
				for _, s := range splx.Restriction.Enumeration {
					if biggestInt < len(s.Value) {
						biggestInt = len(s.Value)
					}
				}
			} else {
				return 0, errors.New("could not find simple type: " + k)
			}
		}
		return biggestInt, nil
	}

	if mxRegx := regexp.MustCompile(`Currently, the maximum length is (\d+) `).FindStringSubmatch(strValue); len(mxRegx) == 2 {
		strValue = mxRegx[1]
		goto parse
	}
	if mxRegx := regexp.MustCompile(`allocates up to (\d+) characters`).FindStringSubmatch(strValue); len(mxRegx) == 2 {
		strValue = mxRegx[1]
	}
parse:
	value := strings.Split(strValue, " ")[0]
	flt, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return int(flt), nil
}

type ValidationContainer []ValidationRule

func (x *ValidationContainer) New(validation ValidationType, value interface{}) {
	*x = append(*x, ValidationRule{Type: validation, Value: value})
}

func (x ValidationContainer) Len() int {
	return len(x)
}

func (x *ValidationContainer) Includes(value ValidationType) (rule *ValidationRule, ok bool) {
	for _, v := range *x {
		if value == v.Type {
			return &v, true
		}
	}
	return nil, false
}

func (x *ValidationContainer) Except(values ...ValidationType) (rules ValidationContainer) {
	for _, v := range *x {
		for _, exc := range values {
			if exc == v.Type {
				goto next
			}
		}
		rules = append(rules, v)
	next:
	}
	return
}

func (a appInfo) ValidationRules(callName string) (ValidationContainer, bool) {
	list := ValidationContainer{}

	if strings.HasSuffix(callName, "Response") {
		return list, false
	}

	if x, ok := a.MaxOccurs(); ok {
		list.New(ValTypMaxOccurs, x)
	}
	if x, ok := a.AllValuesExcept(); ok {
		list.New(ValTypAllValuesExcept, x)
	}
	if x, ok := a.OnlyTheseValues(); ok {
		list.New(ValTypOnlyTheseValues, x)
	}
	if x, ok := a.MaxLength(); ok {
		list.New(ValTypMaxLength, x)
	}
	if x, ok := a.Min(); ok {
		list.New(ValTypMin, x)
	}
	if x, ok := a.Max(); ok {
		list.New(ValTypMax, x)
	}
	for _, a2 := range a.CallInfo {
		if contains(a2.CallName, callName) {
			if x, ok := a2.MaxOccurs(); ok {
				list.New(ValTypMaxOccurs, x)
			}
			if x, ok := a2.AllValuesExcept(); ok {
				list.New(ValTypAllValuesExcept, x)
			}
			if x, ok := a2.OnlyTheseValues(); ok {
				list.New(ValTypOnlyTheseValues, x)
			}
			if x, ok := a2.MaxLength(); ok {
				list.New(ValTypMaxLength, x)
			}
			if x, ok := a2.Min(); ok {
				list.New(ValTypMin, x)
			}
			if x, ok := a2.Max(); ok {
				list.New(ValTypMax, x)
			}
		}
	}
	return list, list.Len() > 0
}

func (e element) NeedsValidation(callName string) (ValidationContainer, bool) {
	list := ValidationContainer{}
	if strings.HasSuffix(callName, "Response") {
		return list, false
	}
	if e.Annotation == nil {
		return list, false
	}
	a := e.Annotation

	if a.RequiredFor(callName) {
		list.New(ValTypRequired, nil)
	} else {
		return list, list.Len() > 0
	}

	if nlist, ok := a.AppInfo.ValidationRules(callName); ok {
		list = append(list, nlist...)
	}

	return list, list.Len() > 0
}

func (a annotation) RequiredFor(callName string) bool {
	for _, ci := range a.AppInfo.CallInfo {
		if ci.AllCallsExcept != "" {
			excepts := strings.Split(strings.Replace(ci.AllCallsExcept, " ", "", -1), ",")
			if contains(excepts, callName) {
				return false
			}
			return true
		}
		if ci.AllCalls != nil {
			return ci.RequiredInput == "Yes"
		}
		for _, c := range ci.CallName {
			if c == callName {
				return ci.RequiredInput == "Yes"
			}
		}
	}
	return false
}

func (a annotation) RequiredInput() bool {
	isRequiredInput := false
	for _, ci := range a.AppInfo.CallInfo {
		if ci.RequiredInput != "" {
			isRequiredInput = true
		}
	}
	return isRequiredInput
}

func (a annotation) Skip() bool {
	if a.AppInfo.NoCall() {
		return true
	}

	if a.AppInfo.CallName != "" && !contains(exportedElements, a.AppInfo.CallName) {
		return true
	}

	for _, p := range a.AppInfo.CallInfo {
		if p.AllCalls != nil {
			return false
		}
		if p.NoCall() {
			return true
		}
		if p.AllCallsExcept != "" {
			excepts := strings.Split(strings.Replace(p.AllCallsExcept, " ", "", -1), ",")
			found := 0
			for _, m := range exportedElements {
				if contains(excepts, m) {
					found++
				}
			}
			if found == len(exportedElements) {
				return true
			}
			return false
		}
		for _, m := range p.CallName {
			if contains(exportedElements, m) {
				return false
			}
		}
	}

	return len(a.AppInfo.CallInfo) != 0
}

// https://msdn.microsoft.com/en-us/library/ms256134(v=vs.110).aspx
type appInfo struct {
	//Source string `xml:"source,attr"`

	EbAppInfo
}

// https://msdn.microsoft.com/en-us/library/ms256112(v=vs.110).aspx
type documentation struct {
	Source  string `xml:"source,attr"`
	XmlLang string `xml:"lang,attr"`

	//Any well-formed XML content.
	Version  string
	Contents string `xml:",chardata"`
}

// https://msdn.microsoft.com/en-us/library/ms256089(v=vs.110).aspx
// One time within group; otherwise, unlimited.
type sequence struct {
	ID        string `xml:"id,attr"`
	MaxOccurs string `xml:"maxOccurs,attr"`
	MinOccurs string `xml:"minOccurs,attr"`

	Annotation annotation `xml:"annotation"`
	//any
	//choice
	Element []element `xml:"element"`
	//group
	//Sequence []sequence `xml:"sequence"`
}
