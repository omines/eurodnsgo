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
func getSOAPArg(entity, name string, input interface{}) (output string) {
	ns := entity + ":" + name
	switch input.(type) {
	case []string:
		i := input.([]string)
		output = fmt.Sprintf(`<%s>`, ns)
		for _, x := range i {
			output = output + fmt.Sprintf(`<item xsi:type="xsd:string">%s</item>`, x)
		}
		output = output + fmt.Sprintf(`<%s>`, ns)
	case string:
		output = fmt.Sprintf(`<%s>%s</%s>`, ns, input, ns)
	case int, int32, int64:
		output = fmt.Sprintf(`<%s>%d</%s>`, ns, input, ns)
	}

	return
}

// getSOAPArgs returns XML representing given name and argument as SOAP body
func getSOAPArgs(entity string, method string, input ...string) []byte {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("<%s:%s>", entity, method))
	for _, x := range input {
		buf.WriteString(x)
	}
	buf.WriteString(fmt.Sprintf("</%s:%s>", entity, method))

	return buf.Bytes()
}

type paramsEncoder interface {
	EncodeParams(ParamsContainer)
	EncodeArgs(string) string
}

// ParamsContainer is the interface a type should implement to be able to hold
// SOAP parameters
type ParamsContainer interface {
	Len() int
	Add(string, interface{})
}

type soapParams struct {
	keys   []string
	values []interface{}
}

// Add adds parameter data to the end of this SoapParams
func (s *soapParams) Add(k string, v interface{}) {
	if s.keys == nil {
		s.keys = make([]string, 0)
	}

	if s.values == nil {
		s.values = make([]interface{}, 0)
	}

	s.keys = append(s.keys, k)
	s.values = append(s.values, v)
}

// Len returns amount of parameters set in this SoapParams
func (s soapParams) Len() int {
	return len(s.keys)
}

// The SoapRequest contains all needed data to perform a request at the
// eurodnsgo server.
type SoapRequest struct {
	Method string
	Entity string
	Result interface{}

	args   []string
	params *soapParams
}

// AddArgument adds an argument to the SoapRequest; the arguments ared used to
// fill the XML request body as well as to create a valid signature for the
// request
func (sr *SoapRequest) AddArgument(entity, key string, value interface{}) {
	if sr.params == nil {
		sr.params = &soapParams{}
	}

	// check if value implements paramsEncoder
	if pe, ok := value.(paramsEncoder); ok {
		sr.args = append(sr.args, pe.EncodeArgs(key))
		pe.EncodeParams(sr.params)
		return
	}

	switch value.(type) {
	case []string:
		sr.params.Add(fmt.Sprintf("%d", sr.params.Len()), value)
		sr.args = append(sr.args, getSOAPArg(entity, key, value))
	case string:
		sr.params.Add(fmt.Sprintf("%d", sr.params.Len()), value)
		sr.args = append(sr.args, getSOAPArg(entity, key, value.(string)))
	case int, int8, int16, int32, int64:
		sr.params.Add(fmt.Sprintf("%d", sr.params.Len()), value)
		sr.args = append(sr.args, getSOAPArg(entity, key, fmt.Sprintf("%d", value)))
	default:
		// check if value implements the String interface
		if str, ok := value.(fmt.Stringer); ok {
			sr.params.Add(fmt.Sprintf("%d", sr.params.Len()), str.String())
			sr.args = append(sr.args, getSOAPArg(entity, key, str.String()))
		}
	}
}

func (sr *SoapRequest) getEnvelope() string {
	t := strings.Replace(soapEnvelopeFixture, "{ENTITY}", sr.Entity, -1)
	t = strings.Replace(t, "{METHOD}", sr.Method, -1)
	t = fmt.Sprintf(t, getSOAPArgs(sr.Entity, sr.Method, sr.args...))
	return "xml=" + url.QueryEscape(url.QueryEscape(t))
}

// NewSoapRequest creates a new SoapRequest instace
func NewSoapRequest(domain, method string, result interface{}) *SoapRequest {
	return &SoapRequest{
		Entity: domain,
		Method: method,
		Result: result,
	}
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
