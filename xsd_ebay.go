package main

// type EbValidator interface {
// 	NoCall() bool
// 	AllValuesExcept() (string, bool)
// 	OnlyTheseValues() (string, bool)
// 	MaxLength() (string, bool)
// 	MaxOccurs() (int, bool)
// 	Min() (string, bool)
// 	Max() (string, bool)
// 	Default() (string, bool)
// }

type EbAppInfo struct {
	MaxDepth  int
	MinOccurs int
	//PresentDetails
	EbMaxOccurs
	EbAllValuesExcept
	EbOnlyTheseValues
	EbMaxLength
	EbMin
	EbMax
	EbDefault

	EbListBasedOn
	EbNoCalls

	CallInfo     []ebCallInfo
	CallName     string
	SeeLink      []ebSeeLink
	Summary      string
	RelatedCalls []string

	DeprecationVersion int
	DeprecationDetails string
	EndOfLifeVersion   int
	UseInstead         *string
}

type ebCallInfo struct {
	EbMaxOccurs
	EbAllValuesExcept
	EbOnlyTheseValues
	EbMaxLength
	EbMin
	EbMax
	EbDefault

	EbNoCalls

	AllCalls *string
	CallName []string
	Details  string
	Context  string

	AllCallsExcept string

	RequiredInput string
	Returned      string
	Summary       string
	SeeLink       ebSeeLink
}

type ebSeeLink struct {
	Title string
	URL   string
}

type EbNoCalls struct {
	LNoCalls *string `xml:"noCalls"`
	UNoCalls *string `xml:"NoCalls"`
}

func (n EbNoCalls) NoCall() bool {
	if n.LNoCalls != nil {
		return true
	}
	return n.UNoCalls != nil
}

type EbAllValuesExcept struct {
	LowerAllValuesExcept *string `xml:"allValuesExcept"`
	UpperAllValuesExcept *string `xml:"AllValuesExcept"`
}

func (n EbAllValuesExcept) AllValuesExcept() (string, bool) {
	if n.LowerAllValuesExcept != nil && *n.LowerAllValuesExcept != "" {
		return *n.LowerAllValuesExcept, true
	}
	if n.UpperAllValuesExcept != nil && *n.UpperAllValuesExcept != "" {
		return *n.UpperAllValuesExcept, true
	}
	return "", false
}

type EbOnlyTheseValues struct {
	LOnlyTheseValues *string `xml:"onlyTheseValues"`
	UOnlyTheseValues *string `xml:"OnlyTheseValues"`
}

func (n EbOnlyTheseValues) OnlyTheseValues() (string, bool) {
	if n.LOnlyTheseValues != nil && *n.LOnlyTheseValues != "" {
		return *n.LOnlyTheseValues, true
	}
	if n.UOnlyTheseValues != nil && *n.UOnlyTheseValues != "" {
		return *n.UOnlyTheseValues, true
	}
	return "", false
}

type EbMaxLength struct {
	LMaxLength *string `xml:"maxLength"`
	UMaxLength *string `xml:"MaxLength"`
}

func (n EbMaxLength) MaxLength() (string, bool) {
	if n.LMaxLength != nil && *n.LMaxLength != "" {
		return *n.LMaxLength, true
	}
	if n.UMaxLength != nil && *n.UMaxLength != "" {
		return *n.UMaxLength, true
	}
	return "", false
}

type EbListBasedOn struct {
	LowerCaseListBasedOn *string `xml:"listBasedOn"`
	UpperCaseListBasedOn *string `xml:"ListBasedOn"`
}

func (n EbListBasedOn) ListBasedOn() (string, bool) {
	if n.LowerCaseListBasedOn != nil && *n.LowerCaseListBasedOn != "" {
		return *n.LowerCaseListBasedOn, true
	}
	if n.UpperCaseListBasedOn != nil && *n.UpperCaseListBasedOn != "" {
		return *n.UpperCaseListBasedOn, true
	}
	return "", false
}

type EbMaxOccurs struct {
	LowerCaseMaxOccurs *int `xml:"maxOccurs"`
	UpperCaseMaxOccurs *int `xml:"MaxOccurs"`
}

func (n EbMaxOccurs) MaxOccurs() (int, bool) {
	if n.LowerCaseMaxOccurs != nil {
		return *n.LowerCaseMaxOccurs, true
	}
	if n.UpperCaseMaxOccurs != nil {
		return *n.UpperCaseMaxOccurs, true
	}
	return 0, false
}

type EbMin struct {
	LowerMin *string `xml:"min"`
	UpperMin *string `xml:"Min"`
}

func (n EbMin) Min() (string, bool) {
	if n.LowerMin != nil && *n.LowerMin != "" {
		return *n.LowerMin, true
	}
	if n.UpperMin != nil && *n.UpperMin != "" {
		return *n.UpperMin, true
	}
	return "", false
}

type EbMax struct {
	LowerMax *string `xml:"max"`
	UpperMax *string `xml:"Max"`
}

func (n EbMax) Max() (string, bool) {
	if n.LowerMax != nil && *n.LowerMax != "" {
		return *n.LowerMax, true
	}
	if n.UpperMax != nil && *n.UpperMax != "" {
		return *n.UpperMax, true
	}
	return "", false
}

type EbDefault struct {
	LowerDefault *string `xml:"default"`
	UpperDefault *string `xml:"Default"`
}

func (n EbDefault) Default() (string, bool) {
	if n.LowerDefault != nil && *n.LowerDefault != "" {
		return *n.LowerDefault, true
	}
	if n.UpperDefault != nil && *n.UpperDefault != "" {
		return *n.UpperDefault, true
	}
	return "", false
}
