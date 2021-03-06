package sdb

import (
	"github.com/abneptis/GoAWS"
	"net/url"
)

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
)

type Domain struct {
	URL  *url.URL
	conn *aws.Conn
	Name string
}

//   if domain != "" { params["DomainName"] = domain }
//  params["Action"] = action

func (self *Domain) DeleteAttribute(s *aws.Signer, item string, attrs, expected AttributeList) (err error) {
	var resp *http.Response

	vl := attrs.Values(ATTRIBUTE_LIST)

	for k, v := range expected.Values(EXPECTED_LIST) {
		vl[k] = v
	}

	vl.Set("Action", "DeleteAttribute")
	vl.Set("DomainName", self.Name)
	vl.Set("ItemName", item)

	req := aws.NewRequest(self.URL, "GET", nil, vl)
	err = s.SignRequestV2(req, aws.Canonicalize, DEFAULT_API_VERSION, 0)
	if err == nil {
		resp, err = self.conn.Request(req)
	}
	if err == nil {
		defer resp.Body.Close()
		err = aws.CodeToError(resp.StatusCode)
	}
	return
}

func (self *Domain) GetAttribute(s *aws.Signer, item string, attrs AttributeList, consist bool) (a []Attribute, err error) {
	var resp *http.Response

	vl := attrs.Values(ATTRIBUTE_LIST)

	vl.Set("Action", "GetAttributes")
	vl.Set("DomainName", self.Name)
	vl.Set("ItemName", item)

	if consist {
		vl.Set("ConsistentRead", "true")
	}

	req := aws.NewRequest(self.URL, "GET", nil, vl)
	err = s.SignRequestV2(req, aws.Canonicalize, DEFAULT_API_VERSION, 0)
	if err == nil {
		resp, err = self.conn.Request(req)
	}
	if err == nil {
		defer resp.Body.Close()
		err = aws.CodeToError(resp.StatusCode)
	}
	if err == nil {
		var response getattributesresponse
		ob, _ := httputil.DumpResponse(resp, true)
		os.Stdout.Write(ob)
		err = xml.NewDecoder(resp.Body).Decode(&response)
		if err == nil {
			a = response.Attributes
		}
	}
	return
}

func (self *Domain) Select(id *aws.Signer, what, where string, consist bool, items chan<- Item) (err error) {
	var resp *http.Response

	vl := url.Values{}

	vl.Set("Action", "Select")
	if where != "" {
		where = " where " + where
	}
	vl.Set("SelectExpression", fmt.Sprintf("select %s from %s%s", what, self.Name, where))

	if consist {
		vl.Set("ConsistentRead", "true")
	}
	done := false
	nextToken := ""
	for err == nil && !done {
		vl.Del("NextToken")
		if nextToken != "" {
			vl.Set("NextToken", nextToken)
		}
		req := aws.NewRequest(self.URL, "GET", nil, vl)
		err = id.SignRequestV2(req, aws.Canonicalize, DEFAULT_API_VERSION, 0)
		if err == nil {
			resp, err = self.conn.Request(req)
		}
		if err == nil {
			ob, _ := httputil.DumpResponse(resp, true)
			os.Stdout.Write(ob)
			xresp := selectresponse{}
			err = xml.NewDecoder(resp.Body).Decode(&xresp)
			if err == nil {
				fmt.Printf("XML == %+v", xresp)
				for i := range xresp.Items {
					items <- xresp.Items[i]
				}
				nextToken = xresp.NextToken
				done = (nextToken == "")
			}
			resp.Body.Close()
		}
	}
	return
}

// Closes the underlying connection
func (self *Domain) Close() (err error) {
	return self.conn.Close()
}
