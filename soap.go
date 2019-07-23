package eurodnsgo

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	// format for SOAP envelopes
	soapEnvelopeFixture string = `<?xml version="1.0" encoding="UTF-8"?>
<request xmlns:{ENTITY}="http://www.eurodns.com/{ENTITY}">
	%s
</request>
`
)

type paramsEncoder interface {
	EncodeParams(ParamsContainer)
	EncodeArgs(string) string
}

// XMLable is a small interface the records can implement to be
// recognizable as application bound.
type XMLable interface {
	ToXML() string
}

// Attr is a simple key/value store
type Attr struct {
	Key   string
	Value interface{}
}

// Param describes an element inside the request structure for
// the API
type Param interface {
	Entity() string
	Key() string

	Value() interface{}
	Attrs() []Attr
}

// NewParam creates a new Param interface to be used as parameter
// for the API.
func NewParam(s, k string, v interface{}, attrs ...Attr) Param {
	return &soapParam{s, k, v, attrs}
}

// SoapParamContainer is an struct to expose some methods to the
// outside. TODO should be better implemented inside the interfaces
// so pointer receivers don't cause trouble anymore.
type SoapParamContainer struct {
	soapParams
}

type soapParam struct {
	entity string
	key    string
	value  interface{}
	attrs  []Attr
}

// Entity is here to provide Param interface
func (s soapParam) Entity() string {
	return s.entity
}

// Key is here to provide Param interface
func (s soapParam) Key() string {
	return s.key
}

// Value is here to provide Param interface
func (s soapParam) Value() interface{} {
	return s.value
}

// Attrs is here to provide Param interface
func (s soapParam) Attrs() []Attr {
	return s.attrs
}

// ParamsContainer is the interface a type should implement to be able to hold
// SOAP parameters
type ParamsContainer interface {
	Len() int
	Params() []Param
}

type soapParams struct {
	params []Param
}

// AddParam adds parameter data to the end of this SoapParams
func (s *soapParams) AddParam(p Param) {
	if s.params == nil {
		s.params = make([]Param, 0)
	}
	s.params = append(s.params, p)
}

// Len returns amount of parameters set in this SoapParams
func (s soapParams) Len() int {
	return len(s.params)
}

// Params returns the internal Param interfaces
func (s soapParams) Params() []Param {
	return s.params
}

// The SoapRequest contains all needed data to perform a request at the
// EuroDNS server.
type SoapRequest struct {
	soapParams
	Namespace string
	Method    string
	Result    interface{}
	IsTest    bool // default false
}

// Entity is here to provide Param interface
func (sr *SoapRequest) Entity() string {
	return sr.Namespace
}

// Key is here to provide Param interface
func (sr *SoapRequest) Key() string {
	return sr.Method
}

// Value is here to provide Param interface
func (sr *SoapRequest) Value() interface{} {
	return sr.soapParams
}

// Attrs is here to provide Param interface
func (sr *SoapRequest) Attrs() []Attr {
	var v = make([]Attr, 0)
	return v
}

// NewSoapRequest creates a new SoapRequest instance
func NewSoapRequest(domain, method string, result interface{}) *SoapRequest {
	return &SoapRequest{
		Namespace: domain,
		Method:    method,
		Result:    result,
	}
}

// PrepareContent returns the unencoded xml payload which will be send based on the
// SoapRequest params. This function can be used to validate generated XML against
// the EuroDNS documentation in API tests.
func (sr *SoapRequest) PrepareContent() string {
	t := strings.Replace(soapEnvelopeFixture, "{ENTITY}", sr.Entity(), -1)
	t = strings.Replace(t, "{METHOD}", sr.Method, -1)
	t = fmt.Sprintf(t, getSOAPArg(sr))
	return t
}

func (sr *SoapRequest) getEnvelope() string {
	return "xml=" + url.QueryEscape(url.QueryEscape(sr.PrepareContent()))
}

type soapClient struct {
	login     string
	password  string
	host      string
	callDelay int
}

type soapResult struct {
	XMLName xml.Name `xml:"result"`
	Message string   `xml:"msg"`
	Code    int      `xml:"code,attr"`
}

type soapData struct {
	XMLName  xml.Name `xml:"resData"`
	Contents []byte   `xml:",innerxml"`
}

type soapEnvelope struct {
	XMLName  xml.Name `xml:"response"`
	Result   soapResult
	Data     soapData
	InnerXML []byte `xml:",innerxml"`
}

func (s *soapClient) call(ctx context.Context, req *SoapRequest) ([]byte, error) {
	// get http request for soap request
	httpReq, err := s.httpReqForSoapRequest(ctx, *req)
	if err != nil {
		return nil, err
	}

	http := &http.Client{}
	res, err := http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request error:\n%s", err.Error())
	}
	defer res.Body.Close()

	// read entire response body
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// parse SOAP response into given result interface
	parsed, err := parseSoapResponse(b, req)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

// httpReqForSoapRequest creates the HTTP request for a specific SoapRequest
// this includes setting the URL, POST body and cookies
func (s soapClient) httpReqForSoapRequest(ctx context.Context, req SoapRequest) (*http.Request, error) {
	url := fmt.Sprintf("https://%s", s.host)

	b := bytes.NewBuffer([]byte(req.getEnvelope()))
	httpReq, err := http.NewRequest("POST", url, b)
	if err != nil {
		return nil, err
	}
	httpReq = httpReq.WithContext(ctx)

	authStr := base64.StdEncoding.EncodeToString([]byte(s.login + ":" + s.password))
	httpReq.Header.Add("Authorization", "Basic "+authStr)
	httpReq.Header.Add("Connection", "close")
	httpReq.Header.Add("Content-type", "application/x-www-form-urlencoded")
	httpReq.Header.Add("Content-length", string(b.Len()))

	return httpReq, nil
}

// getSOAPArg returns XML representing given input argument as SOAP parameters
// in combination with getSOAPArgs you can build SOAP body
func getSOAPArg(p Param) (output string) {
	entity := p.Entity()
	name := p.Key()
	input := p.Value()
	attrs := p.Attrs()
	var attr string

	for _, a := range attrs {
		switch a.Value.(type) {
		case string:
			attr += fmt.Sprintf(`%s="%s"`, a.Key, a.Value)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			attr += fmt.Sprintf(`%s="%d"`, a.Key, a.Value)
		}
	}

	ns := entity + ":" + name
	if attr == "" {
		output = fmt.Sprintf(`<%s>`, ns)
	} else {
		output = fmt.Sprintf(`<%s %s>`, ns, attr)
	}
	switch input.(type) {
	case []byte:
		output += string(input.([]byte))
	case string:
		output += input.(string)
	case int, int32, int64:
		output += fmt.Sprintf(`%d`, input)
	case Param:
		output += string(getSOAPArg(input.(Param)))
	case ParamsContainer:
		output += string(getSOAPArgs(input.(ParamsContainer)))
	case XMLable:
		output += string(input.(XMLable).ToXML())
	}
	output += fmt.Sprintf(`</%s>`, ns)

	return
}

// getSOAPArgs returns XML representing given name and argument as SOAP body
func getSOAPArgs(pc ParamsContainer) []byte {
	var buf bytes.Buffer

	for _, p := range pc.Params() {
		buf.WriteString(getSOAPArg(p))
	}

	return buf.Bytes()
}

func parseSoapResponse(data []byte, sr *SoapRequest) ([]byte, error) {
	var env soapEnvelope
	if err := xml.Unmarshal(data, &env); err != nil {
		return nil, err
	}

	if env.Result.Code != 1000 {
		return nil, errors.New(env.Result.Message)
	}

	// wrap the inner content
	content := fmt.Sprintf("<resData>%s</resData>", string(env.Data.Contents))

	if err := xml.Unmarshal([]byte(content), sr.Result); err != nil {
		return nil, err
	}

	return env.Data.Contents, nil
}
