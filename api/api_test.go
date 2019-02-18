package api

import (
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
	e := `<?xml version="1.0" encoding="UTF-8"?>
<request xmlns:zone="http://www.eurodns.com/zone">
	<zone:update><zone:name>zone</zone:name><zone:records><zone:add><zone:record><record:data></record:data><record:expire>0</record:expire><record:host></record:host><record:priority>0</record:priority><record:refresh>0</record:refresh><record:resp_person></record:resp_person><record:retry>0</record:retry><record:ttl>0</record:ttl><record:type>A</record:type></zone:record></zone:add></zone:records></zone:update>
</request>`
	var v interface{}
	r := Record{
		ID:   1234,
		Type: string(RecordTypeA),
	}
	sr, _ := addRecordRequest(v, Zone{Name: "zone"}, r)

	testParams(t, sr, e)
}
