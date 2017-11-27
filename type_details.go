package main

import (
	"fmt"
	"log"
	"strings"
)

type TypeDetails struct {
	Field      string
	Type       Type
	AliasFor   Type
	IsPointer  bool
	IsSlice    bool
	SimpleType bool
	key        string
}

type ValidationType byte

const (
	ValTypMaxOccurs ValidationType = iota
	ValTypAllValuesExcept
	ValTypOnlyTheseValues
	ValTypMaxLength
	ValTypRequired
	ValTypMin
	ValTypMax
	ValTypDefault
)

func (t TypeDetails) Path(path string) string {
	base := path + "." + UpperFirstLetter(t.Field)
	if t.key != "" {
		base = fmt.Sprintf("%s.%s[%s]", path, UpperFirstLetter(t.Field), t.key)
	}
	if t.AliasFor != t.Type && t.AliasFor.GoType() == "string" {
		base += ".String()"
	}

	return base
}

func (t *TypeDetails) Key(key string) *TypeDetails {
	t.key = key
	return t
}

type ValidationParam [2]string

func (p ValidationParam) Condition() string {
	return p[0]
}

func (p ValidationParam) Error() string {
	return p[1]
}

func (p *ValidationParam) New(condition, err string) {
	p[0] = condition
	p[1] = err
}

func (t TypeDetails) ValidationParams(rule ValidationRule, path string) (p ValidationParam) {
	fpath := path + "." + UpperFirstLetter(t.Field)
	setterChecker := func() bool {
		if k := t.IsSet(fpath); k != "" {
			p.New(
				k,
				fmt.Sprintf("field %s must be set", t.Field))
			return true
		}
		return false
	}
	switch rule.Type {
	case ValTypMaxOccurs:
		if t.IsSlice {
			p.New(
				fmt.Sprintf("len(%[1]s) > %[2]v", t.Path(path), rule.Value),
				fmt.Sprintf("() field %s must be between 0 and %v", fpath, rule.Value))
		}
	case ValTypAllValuesExcept:
		parts := strings.Split(rule.Value.(string), ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		p.New(
			fmt.Sprintf("contains([]string{\"%[2]s\"}, string(%[1]s))", t.Path(path), strings.Join(parts, "\", \"")),
			fmt.Sprintf("field %s contains invalid value", fpath))
	case ValTypOnlyTheseValues:
		parts := strings.Split(rule.Value.(string), ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		p.New(
			fmt.Sprintf("!contains([]string{\"%[2]s\"}, string(%[1]s))", t.Path(path), strings.Join(parts, "\", \"")),
			fmt.Sprintf("field %s contains invalid value", fpath))
	case ValTypMaxLength:
		valueInt, err1 := rule.ValueInt()
		if err1 != nil {
			log.Printf("Could not parse MaxLength value for %s, skipping validation line. Error: `%s`\r\nValue: `%v`", fpath, err1, rule.Value)
		}
		p.New(
			fmt.Sprintf(t.CheckMaxLenght(t.Path(path), valueInt)),
			fmt.Sprintf("field %s must be between 1 and %d characters long", fpath, valueInt))
	case ValTypRequired:
		if !setterChecker() {
			log.Fatalf("ValTypRequired: validation handling is not implemented for %s %v %+v", fpath, rule, t)
		}
	case ValTypMin:
		switch t.AliasFor.GoType() {
		case "NullFloat64", "NullInt64":
			valueInt, err1 := rule.ValueInt()
			if err1 != nil {
				log.Printf("Could not parse Min value for %s, skipping validation line. Error: `%s`\r\nValue: `%v`", t.Path(path), err1, rule.Value)
			}
			p.New(
				fmt.Sprintf("%[1]s.Valid && %[1]s.Value() < %[2]d", t.Path(path), valueInt),
				fmt.Sprintf("(max) field %s must be more than %d of length", fpath, valueInt))
		default:
			log.Fatalf("ValTypMin: validation handling is not implemented for %s %v %+v", fpath, rule, t)
		}
	case ValTypMax:
		switch t.AliasFor.GoType() {
		case "NullFloat64", "NullInt64":
			valueInt, err1 := rule.ValueInt()
			if err1 != nil {
				log.Printf("Could not parse Max value for %s, skipping validation line. Error: `%s`\r\nValue: `%v`", t.Path(path), err1, rule.Value)
			}
			p.New(
				fmt.Sprintf("%[1]s.Valid && %[1]s.Value() > %[2]d", t.Path(path), valueInt),
				fmt.Sprintf("ValTypMax: field %s must be between 1 and %d", fpath, valueInt))
		default:
			log.Fatal("ValTypMax: validation handling is not implemented for %s %v %+v", fpath, rule, t)
		}
	default:
		log.Fatal("validation handling is not implemented")
	}
	if len(p) == 0 {
		log.Fatalf("validation did not implemenet any handling for %+v, rule: %+v", t, rule)
	}
	return
}

func (t TypeDetails) ValidationString(rule ValidationRule, path string) string {
	var condition []string
	var err string
	fpath := path + "." + UpperFirstLetter(t.Field)
	setterChecker := func() bool {
		if k := t.IsSet(fpath); k != "" {
			condition = append(condition, k)
			err = fmt.Sprintf("field %s must be set", t.Field)
			return true
		}
		return false
	}
	switch rule.Type {
	case ValTypMaxOccurs:
		if t.IsSlice {
			condition = append(condition, fmt.Sprintf("len(%[1]s) > %[2]v", t.Path(path), rule.Value))
			err = fmt.Sprintf("field %s must be between 0 and %v", fpath, rule.Value)
		} else {
			return ""
		}
	case ValTypAllValuesExcept:
		parts := strings.Split(rule.Value.(string), ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		condition = append(condition, fmt.Sprintf("contains([]string{\"%[2]s\"}, string(%[1]s))", t.Path(path), strings.Join(parts, "\", \"")))
		err = fmt.Sprintf("field %s contains invalid value", fpath)
	case ValTypOnlyTheseValues:
		parts := strings.Split(rule.Value.(string), ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		condition = append(condition, fmt.Sprintf("!contains([]string{\"%[2]s\"}, string(%[1]s))", t.Path(path), strings.Join(parts, "\", \"")))
		err = fmt.Sprintf("field %s contains invalid value", fpath)
	case ValTypMaxLength:
		valueInt, err1 := rule.ValueInt()
		if err1 != nil {
			log.Printf("Could not parse MaxLength value for %s, skipping validation line. Error: `%s`\r\nValue: `%v`", fpath, err1, rule.Value)
			return ""
		}
		if k := t.CheckMaxLenght(t.Path(path), valueInt); k != "" {
			condition = append(condition, k)
			err = fmt.Sprintf("field %s must be between 1 and %d characters long", fpath, valueInt)
		} else {
			return ""
		}
	case ValTypRequired:
		if !setterChecker() {
			log.Fatalf("ValTypRequired: validation handling is not implemented for %s %v %+v", fpath, rule, t)
		}
	case ValTypMin:
		valueInt, err1 := rule.ValueInt()
		if err1 != nil {
			log.Printf("Could not parse Min value for %s, skipping validation line. Error: `%s`\r\nValue: `%v`", t.Path(path), err1, rule.Value)
			return ""
		}
		condition = append(condition, t.Min(t.Path(path), valueInt))
		err = fmt.Sprintf("(max) field %s must be more than %d of length", fpath, valueInt)
	case ValTypMax:
		valueInt, err1 := rule.ValueInt()
		if err1 != nil {
			log.Printf("Could not parse Max value for %s, skipping validation line. Error: `%s`\r\nValue: `%v`", t.Path(path), err1, rule.Value)
			return ""
		}
		condition = append(condition, t.Max(t.Path(path), valueInt))
		err = fmt.Sprintf("ValTypMax: field %s must be between 1 and %d", fpath, valueInt)
	default:
		log.Fatal("validation handling is not implemented")
	}
	if len(condition) == 0 {
		log.Fatalf("validation did not implemenet any handling for %+v, rule: %+v", t, rule)
	}

	return fmt.Sprintf("if %s { return errors.New(\"%s\") }\r\n", strings.Join(condition, " && "), err)
}

func (t TypeDetails) T() string {
	if t.SimpleType {
		return t.AliasFor.GoType(true)
	}

	return t.AliasFor.GoType()
}

func (t TypeDetails) CheckMaxLenght(path string, value int) string {
	switch t.T() {
	case "string":
		return fmt.Sprintf("len(%s) > %v", path, value)
	case "int32", "int64":
		return fmt.Sprintf("%s > %v", path, value)
	case "NullString":
		return fmt.Sprintf("len(%s.NullString.String) > %v", path, value)
	case "AmountType":
		return fmt.Sprintf("%[1]s.Value.Valid && len(%[1]s.Value.String()) > %[2]v", path, value)
	case "NullFloat64":
		return fmt.Sprintf("%[1]s.Valid && len(%[1]s.String()) > %[2]v", path, value)
	case "CategoryType":
		return ""
	}
	log.Fatalf("TypeDetails.CheckMaxLenght: unknown %s : %+v", t.T(), t)
	return ""
}

func (t TypeDetails) IsSet(path string) string {
	if t.IsPointer {
		return fmt.Sprintf("%s == nil", path)
	}
	if t.IsSlice {
		return fmt.Sprintf("len(%s) == 0", path)
	}
	switch t.T() {
	case "string":
		return fmt.Sprintf("%s == \"\"", path)
	case "int32", "int64":
		return fmt.Sprintf("%s == 0", path)
	case "NullString", "NullFloat64", "NullInt64", "NullBool":
		return fmt.Sprintf("!%s.Valid", path)
	}
	log.Fatalf("TypeDetails.IsSet: unknown %s : %+v", t.T(), t)
	return ""
}

func (t TypeDetails) Min(path string, value int) string {
	switch t.T() {
	case "int32", "int64":
		return fmt.Sprintf("%s < %d", path, value)
	case "NullFloat64", "NullInt64":
		return fmt.Sprintf("%s.Value() < %d", path, value)
	case "AmountType":
		return fmt.Sprintf("%s.Value.Value() < %d", path, value)
	}
	log.Fatalf("TypeDetails.Min: unknown %s : %+v", t.T(), t)
	return ""
}

func (t TypeDetails) Max(path string, value int) string {
	switch t.T() {
	case "int32", "int64":
		return fmt.Sprintf("%s > %d", path, value)
	case "NullFloat64", "NullInt64":
		return fmt.Sprintf("%s.Value() > %d", path, value)
	case "AmountType":
		return fmt.Sprintf("%s.Value.Value() > %d", path, value)
	}
	log.Fatalf("TypeDetails.Min: unknown %s : %+v", t.T(), t)
	return ""
}
