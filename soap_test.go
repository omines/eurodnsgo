package eurodnsgo

import (
	"testing"
)

func TestSOAPEnvelope(t *testing.T) {
	var v interface{}
	sr := SoapRequest{
		Entity: "entity",
		Method: "test",
		Result: &v,
	}

	if e := sr.getEnvelope(); len(e) == 0 {
		t.Fatal("envelope of a soap request should return content")
	}
	// TODO extend tests
}
