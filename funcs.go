package main

import "fmt"

func requester(typeName string) string {
	return fmt.Sprintf(`func (x *%[1]sRequestType) Request(eBayAuthToken, siteID string) (response %[1]sResponseType, err error) {
		if x.RequesterCredentials == nil {
			x.RequesterCredentials = &XMLRequesterCredentialsType{}
		}
		x.RequesterCredentials.EBayAuthToken.Set(eBayAuthToken)
		
		if RequestValidation {
			if err = x.Validate(); err != nil {
				return
			}
		}
		
		req := newRequester("%[1]s", siteID, &response)
		if err = xml.NewEncoder(req.body).Encode(x); err != nil {
			return
		}
	
		if err = req.request(); err != nil {
			return
		}
	
		return
	}
	`, typeName)
}

func xmlEncoder(typeName string) string {
	return fmt.Sprintf(`func (x %sRequestType) MarshalXMLEncode(w io.Writer) error {
		if RequestValidation { 
			if err := x.Validate(); err != nil {
				return err
			}
		}
		return xml.NewEncoder(w).Encode(x)
	}
	`, typeName)
}

func xmlMarshaler(typeName string) string {
	return fmt.Sprintf(`func (x %sRequestType) MarshalXML() ([]byte, error) {
		if RequestValidation { 
			if err := x.Validate(); err != nil {
				return nil, err
			}
		}
		return xml.Marshal(x)
	}
	`, typeName)
}

func validator(typeName, body string) string {
	return fmt.Sprintf(`func (x %[1]sRequestType) Validate() error {
		%[2]s
		return nil
	}
	`, typeName, body)
}
