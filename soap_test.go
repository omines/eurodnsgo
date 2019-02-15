package eurodnsgo

import (
	"testing"
)

func TestSoapRequest(t *testing.T) {
	sr := SoapRequest{
		Namespace: "entity",
		Method:    "test",
	}
	var v interface{} = &sr
	var ok bool

	_, ok = v.(Param)
	if !ok {
		t.Fatal("SoapRequest should implement Param interface")
	}

	_, ok = v.(ParamsContainer)
	if !ok {
		t.Fatal("SoapRequest should implement ParamContainer interface")
	}
}

func TestSOAPEnvelope(t *testing.T) {
	var v interface{}
	sr := SoapRequest{
		Namespace: "entity",
		Method:    "test",
		Result:    &v,
	}

	if e := sr.getEnvelope(); len(e) == 0 {
		t.Fatal("envelope of a soap request should return content")
	}
	// TODO extend tests
}

var addParamResult = `<?xml version="1.0" encoding="UTF-8"?>
<request xmlns:entity="http://www.eurodns.com/entity">
	<entity:test><test:a id="2"name="name">12</test:a></entity:test>
</request>
`

func TestAddParam(t *testing.T) {
	var v interface{}
	sr := SoapRequest{
		Namespace: "entity",
		Method:    "test",
		Result:    &v,
	}
	sr.AddParam(NewParam("test", "a", 12, Attr{"id", 2}, Attr{"name", "name"}))

	var e string
	if e = sr.prepareContent(); len(e) == 0 {
		t.Fatal("envelope of a soap request should return content")
	}

	if e != addParamResult {
		t.Fatalf("expect prepareContent to match our expected result, received \"%s\"", e)
	}
}
