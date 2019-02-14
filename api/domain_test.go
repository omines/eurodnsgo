package api

import (
	"encoding/xml"
	"testing"
)

var zoneListXml = `
<resData>
	<zone:list>
		<zone:name>example1.org</zone:name>
		<zone:name>example2.org</zone:name>
		<zone:name>example3.org</zone:name>
	</zone:list>
</resData>
`

func TestUnmarshalZoneList(t *testing.T) {
	var zl zoneList
	if err := xml.Unmarshal([]byte(zoneListXml), &zl); err != nil {
		t.Fatal(err)
	}

	if len(zl.Zones) != 3 {
		t.Fatalf("Example contains 3 zones, parsed %d", len(zl.Zones))
	}
}
