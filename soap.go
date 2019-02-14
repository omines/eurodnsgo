package eurodnsgo

import (
	"bytes"
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

// getSOAPArg returns XML representing given input argument as SOAP parameters
// in combination with getSOAPArgs you can build SOAP body
func getSOAPArg(p Param) (output string) {
	entity := p.Entity()
	name := p.Key()
	input := p.Value()

	ns := entity + ":" + name
	output = fmt.Sprintf(`<%s>`, ns)
	switch input.(type) {
	case string:
		output += input.(string)
	case int, int32, int64:
		output += fmt.Sprintf(`%d`, input)
	case ParamsContainer:
		output += string(getSOAPArgs(input.(ParamsContainer)))
	}
	output += fmt.Sprintf(`</%s>`, ns)

	return
}

// getSOAPArgs returns XML representing given name and argument as SOAP body
func getSOAPArgs(pc ParamsContainer) []byte {
	var buf bytes.Buffer

	for _, p := range pc.Params() {
		buf.WriteString(fmt.Sprintf("<%s:%s>", p.Entity(), p.Key()))
		buf.WriteString(getSOAPArg(p))
		buf.WriteString(fmt.Sprintf("</%s:%s>", p.Entity(), p.Key()))
	}

	return buf.Bytes()
}

type paramsEncoder interface {
	EncodeParams(ParamsContainer)
	EncodeArgs(string) string
}

type Param interface {
	Entity() string
	Key() string
	Value() interface{}
}

func NewParam(s, k string, v interface{}) Param {
	return soapParam{s, k, v}
}

type soapParam struct {
	entity string
	key    string
	value  interface{}
}

func (s soapParam) Entity() string {
	return s.entity
}

func (s soapParam) Key() string {
	return s.key
}

func (s soapParam) Value() interface{} {
	return s.value
}

// ParamsContainer is the interface a type should implement to be able to hold
// SOAP parameters
type ParamsContainer interface {
	Len() int
	AddParam(Param)
	Params() []Param
}

type soapParams struct {
	params   []Param
	children []ParamsContainer
}

// Add adds parameter data to the end of this SoapParams
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

func (s soapParams) Params() []Param {
	return s.params
}

// The SoapRequest contains all needed data to perform a request at the
// eurodnsgo server.
type SoapRequest struct {
	soapParams
	Namespace string
	Method    string
	Result    interface{}
}

func (sr *SoapRequest) Entity() string {
	return sr.Namespace
}
func (sr *SoapRequest) Key() string {
	return sr.Method
}
func (sr *SoapRequest) Value() interface{} {
	return sr.soapParams
}

// NewSoapRequest creates a new SoapRequest instance
func NewSoapRequest(domain, method string, result interface{}) *SoapRequest {
	return &SoapRequest{
		Namespace: domain,
		Method:    method,
		Result:    result,
	}
}

func (sr *SoapRequest) getEnvelope() string {
	t := strings.Replace(soapEnvelopeFixture, "{ENTITY}", sr.Entity(), -1)
	t = strings.Replace(t, "{METHOD}", sr.Method, -1)
	t = fmt.Sprintf(t, getSOAPArg(sr))
	return "xml=" + url.QueryEscape(url.QueryEscape(t))
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

func (s *soapClient) call(req *SoapRequest) ([]byte, error) {
	// get http request for soap request
	httpReq, err := s.httpReqForSoapRequest(*req)
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
func (s soapClient) httpReqForSoapRequest(req SoapRequest) (*http.Request, error) {
	url := fmt.Sprintf("https://%s", s.host)

	b := bytes.NewBuffer([]byte(req.getEnvelope()))
	httpReq, err := http.NewRequest("POST", url, b)
	if err != nil {
		return nil, err
	}

	authStr := base64.StdEncoding.EncodeToString([]byte(s.login + ":" + s.password))
	httpReq.Header.Add("Authorization", "Basic "+authStr)
	httpReq.Header.Add("Connection", "close")
	httpReq.Header.Add("Content-type", "application/x-www-form-urlencoded")
	httpReq.Header.Add("Content-length", string(b.Len()))

	return httpReq, nil
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
