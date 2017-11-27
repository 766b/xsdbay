package main

import (
	"encoding/xml"
	"io/ioutil"
	"testing"
)

func Test_annotation_RequiredInput(t *testing.T) {
	data, err := ioutil.ReadFile("ebaysvc_1035.xsd")
	if err != nil {
		t.Fatal(err)
	}

	err = xml.Unmarshal(data, &xsdSc)
	if err != nil {
		t.Fatal(err)
	}
	loadAllCalls()

	tests := []struct {
		name    string
		complex string
		field   string
		want    bool
	}{
		{
			"TransactionType:ContainingOrder",
			"TransactionType",
			"ContainingOrder",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cplx, ok := FindComplex(tt.complex)
			if !ok {
				t.Errorf("could not find complex: `%s`", tt.complex)
			}
			found := false
			for _, f := range cplx.GetElements() {
				if f.GetName() == tt.field {
					found = true
					if got := f.(element).Annotation.RequiredInput(); got != tt.want {
						t.Errorf("annotation.RequiredInput() = %v, want %v | %+v", got, tt.want, f.(element).Annotation)
					}
				}
			}
			if !found {
				t.Errorf("could not find field `%s` in complex type `%s`", tt.field, tt.complex)
			}
		})
	}
}
