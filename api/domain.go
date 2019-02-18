package api

import "encoding/xml"

// RecordType represents the possible types of DNS records
type RecordType string

var (
	// RecordTypeA represents an A-record
	RecordTypeA RecordType = "A"
	// RecordTypeAAAA represents an AAAA-record
	RecordTypeAAAA RecordType = "AAAA"
	// RecordTypeCNAME represents a CNAME-record
	RecordTypeCNAME RecordType = "CNAME"
	// RecordTypeMX represents an MX-record
	RecordTypeMX RecordType = "MX"
	// RecordTypeNS represents an NS-record
	RecordTypeNS RecordType = "NS"
	// RecordTypeTXT represents a TXT-record
	RecordTypeTXT RecordType = "TXT"
	// RecordTypeSRV represents an SRV-record
	RecordTypeSRV RecordType = "SRV"
)

// Record represents an EuroDNS Record object
// See https://agent.api-eurodns.com/doc/record/info
type Record struct {
	XMLName    xml.Name `xml:"record,omitempty"`
	ID         int      `xml:"id,attr"`
	Data       string   `xml:"record data,omitempty"`
	Expire     int      `xml:"record expire,omitempty"`
	Host       string   `xml:"record host,omitempty"`
	Priority   int      `xml:"record priority,omitempty"`
	Refresh    int      `xml:"record refresh,omitempty"`
	RespPerson string   `xml:"record resp_person,omitempty"`
	Retry      int      `xml:"record retry,omitempty"`
	TTL        int      `xml:"record ttl,omitempty"`
	Type       string   `xml:"record type,omitempty"`
}

// Zone represents an EuroDNS Zone object
// See https://agent.api-eurodns.com/doc/zone/info
type Zone struct {
	XMLName xml.Name  `xml:"zone,omitempty"`
	Name    string    `xml:"zone name"`
	Records []*Record `xml:"zone records>record,omitempty"`
}

type domainList struct {
	XMLName xml.Name `xml:"resData,omitempty"`
	Count   int      `xml:"domain numElements,attr"`
	Domains []string `xml:"domain list>name"`
}

type recordInfo struct {
	Record
	XMLName xml.Name `xml:"resData,omitempty"`
}

type zoneList struct {
	XMLName xml.Name `xml:"resData,omitempty"`
	Zones   []string `xml:"zone list>name"`
}

type zoneInfo struct {
	Zone
	XMLName xml.Name `xml:"resData,omitempty"`
}
