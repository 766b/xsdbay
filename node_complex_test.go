package main

import (
	"encoding/xml"
	"io/ioutil"
	"testing"
)

func Test_complexType_DeepValidator(t *testing.T) {
	fileType = extXSD
	data, err := ioutil.ReadFile("ebaysvc.xsd")
	if err != nil {
		t.Fatal(err)
	}

	err = xml.Unmarshal(data, &xsdSc)
	if err != nil {
		t.Fatal(err)
	}

	loadAllCalls()

	type args struct {
		callName string
		path     string
	}
	tests := []struct {
		name string
		e    string
		args args
		want bool
	}{
		{
			name: "CharityType:RelistFixedPriceItem",
			e:    "CharityType",
			args: args{callName: "RelistFixedPriceItem", path: "x"},
			want: false,
		},
		{
			name: "SalesTaxType:RelistFixedPriceItem",
			e:    "SalesTaxType",
			args: args{callName: "RelistFixedPriceItem", path: "x"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cplx, ok := FindComplex(tt.e)
			if !ok {
				t.Errorf("Complex: %s not found", tt.e)
			}
			if got := cplx.DeepValidator(tt.args.callName, tt.args.path); got != tt.want {
				t.Errorf("complexType.DeepValidator() = %v, want %v", got, tt.want)
			}
		})
	}
}
