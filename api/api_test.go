package api

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/omines/eurodnsgo"
)

func trimContent(s string) string {
	s = strings.TrimFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	return s
}

func testParams(t *testing.T, sr *eurodnsgo.SoapRequest, expected string) {
	preparedContent := sr.PrepareContent()

	if trimContent(preparedContent) != trimContent(expected) {
		t.Errorf("Content does not match expected.\n\nExpected: %s\n\nReceived: %s", expected, sr.PrepareContent())
	}
}

func TestDomainListParams(t *testing.T) {
	e := `<?xml version="1.0" encoding="UTF-8"?>
<request xmlns:domain="http://www.eurodns.com/domain">
	<domain:list></domain:list>
</request>`

	var v domainList
	sr := eurodnsgo.NewSoapRequest("domain", "list", &v)

	testParams(t, sr, e)
}

func TestRecordInfoParams(t *testing.T) {
	e := `<?xml version="1.0" encoding="UTF-8"?>
<request xmlns:record="http://www.eurodns.com/record">
	<record:info><record:id>1234</record:id></record:info>
</request>`

	var v recordInfo
	sr := eurodnsgo.NewSoapRequest("record", "info", &v)
	sr.AddParam(eurodnsgo.NewParam("record", "id", 1234))

	testParams(t, sr, e)
}

func TestAddRecordRequest(t *testing.T) {
	var v interface{}
	r := Record{
		ID:   1234,
		Type: string(RecordTypeA),
	}
	sr, _ := addRecordRequest(v, Zone{}, r)

	fmt.Println(sr.PrepareContent()) // TODO finish
}
