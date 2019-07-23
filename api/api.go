package api

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/omines/eurodnsgo"
)

func schedule(ctx context.Context, c eurodnsgo.Client, sr *eurodnsgo.SoapRequest) error {
	ch, err := c.Schedule(ctx, sr)
	if err != nil {
		return err
	}
	defer close(ch)

	// Block until result returns
	_ = <-ch

	return nil
}

func parseTag(t reflect.StructTag) (string, string) {
	st := string(t)
	if !strings.Contains(st, "xml:") {
		return "", ""
	}

	// get the part from the first " to the first ,
	s := strings.Index(st, "\"")
	if s == -1 {
		return "", ""
	}
	e := strings.Index(st, ",")
	if e == -1 {
		return "", ""
	}
	substr := st[s+1 : e]

	parts := strings.Split(substr, " ")
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}

func xmlEncode(val interface{}) ([]byte, error) {
	if reflect.ValueOf(val).Kind() != reflect.Struct {
		return nil, fmt.Errorf("unsupported type %s", reflect.ValueOf(val).Kind())
	}

	var res string
	v := reflect.ValueOf(val)
	e := reflect.TypeOf(val)
	for i := 0; i < v.NumField(); i++ {
		ns, name := parseTag(e.Field(i).Tag)
		if ns == "" || name == "" {
			continue
		}
		tag := fmt.Sprintf("%s:%s", ns, name)
		switch v.Field(i).Kind() {
		case reflect.Int:
			res += fmt.Sprintf("<%s>%d</%s>", tag, v.Field(i).Int(), tag)
		case reflect.String:
			res += fmt.Sprintf("<%s>%s</%s>", tag, v.Field(i).String(), tag)
		case reflect.Struct:
			sub, err := xmlEncode(v)
			if err != nil {
				return nil, err
			}
			res += string(sub)
		}
	}
	return []byte(res), nil
}

// MutationType defines update methods to be used
type MutationType string

var (
	// Add to add an entity
	Add MutationType = "add"
	// Change to edit an entity
	Change MutationType = "change"
	// Remove to remove an entity
	Remove MutationType = "remove"
)

// GetDomainList returns an string list with all manageble domain names
func GetDomainList(ctx context.Context, c eurodnsgo.Client) ([]string, error) {
	var v domainList

	sr := eurodnsgo.NewSoapRequest("domain", "list", &v)

	err := schedule(ctx, c, sr)

	return v.Domains, err
}

// GetRecordInfo returns all data inside a specific record
func GetRecordInfo(ctx context.Context, c eurodnsgo.Client, id int) (Record, error) {
	var v recordInfo

	sr := eurodnsgo.NewSoapRequest("record", "info", &v)
	sr.AddParam(eurodnsgo.NewParam("record", "id", id))

	err := schedule(ctx, c, sr)

	return v.Record, err
}

// GetZoneList gets the list of registered zones
func GetZoneList(ctx context.Context, c eurodnsgo.Client) ([]string, error) {
	var v zoneList

	sr := eurodnsgo.NewSoapRequest("zone", "list", &v)

	err := schedule(ctx, c, sr)

	return v.Zones, err
}

// GetZoneInfo gets more information about a zone
func GetZoneInfo(ctx context.Context, c eurodnsgo.Client, domain string) (Zone, error) {
	var v zoneInfo

	sr := eurodnsgo.NewSoapRequest("zone", "info", &v)
	sr.AddParam(eurodnsgo.NewParam("zone", "name", domain))

	err := schedule(ctx, c, sr)

	return v.Zone, err
}

func addRecordRequest(v interface{}, z Zone, r Record) (*eurodnsgo.SoapRequest, error) {
	sr := eurodnsgo.NewSoapRequest("zone", "update", &v)

	rr, err := xmlEncode(r)

	zoneRecord := eurodnsgo.NewParam("zone", "record", rr)
	zoneAdd := eurodnsgo.NewParam("zone", "add", zoneRecord)
	zoneRecords := eurodnsgo.NewParam("zone", "records", zoneAdd)
	zoneName := eurodnsgo.NewParam("zone", "name", z.Name)

	sr.AddParam(zoneName)
	sr.AddParam(zoneRecords)
	return sr, err
}

// ZoneRecordAdd adds a new Record object to a Zone
func ZoneRecordAdd(ctx context.Context, c eurodnsgo.Client, z Zone, r Record) error {
	var v interface{}

	rr, err := addRecordRequest(v, z, r)
	if err != nil {
		return err
	}
	err = schedule(ctx, c, rr)

	return err
}

func changeRecordRequest(v interface{}, z Zone, r Record) (*eurodnsgo.SoapRequest, error) {
	sr := eurodnsgo.NewSoapRequest("zone", "update", &v)

	rr, err := xmlEncode(r)

	zoneRecord := eurodnsgo.NewParam("zone", "record", rr, eurodnsgo.Attr{Key: "id", Value: r.ID})
	zoneChange := eurodnsgo.NewParam("zone", "change", zoneRecord)
	zoneRecords := eurodnsgo.NewParam("zone", "records", zoneChange)
	zoneName := eurodnsgo.NewParam("zone", "name", z.Name)

	sr.AddParam(zoneName)
	sr.AddParam(zoneRecords)
	return sr, err
}

// ZoneRecordChange changes a Record object inside a Zone
func ZoneRecordChange(ctx context.Context, c eurodnsgo.Client, z Zone, r Record) error {
	var v interface{}

	cr, err := changeRecordRequest(v, z, r)
	if err != nil {
		return err
	}
	err = schedule(ctx, c, cr)

	return err
}

func deleteRecordRequest(v interface{}, z Zone, r Record) *eurodnsgo.SoapRequest {
	sr := eurodnsgo.NewSoapRequest("zone", "update", &v)

	zoneRecord := eurodnsgo.NewParam("zone", "record", nil, eurodnsgo.Attr{Key: "id", Value: r.ID})
	zoneRemove := eurodnsgo.NewParam("zone", "remove", zoneRecord)
	zoneRecords := eurodnsgo.NewParam("zone", "records", zoneRemove)
	zoneName := eurodnsgo.NewParam("zone", "name", z.Name)

	sr.AddParam(zoneName)
	sr.AddParam(zoneRecords)
	return sr
}

// ZoneRecordDelete deletes a Record object from a Zone
func ZoneRecordDelete(ctx context.Context, c eurodnsgo.Client, z Zone, r Record) error {
	var v interface{}

	dr := deleteRecordRequest(v, z, r)
	err := schedule(ctx, c, dr)

	return err
}
