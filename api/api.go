package api

import (
	"fmt"

	"github.com/omines/eurodnsgo"
)

func schedule(c eurodnsgo.Client, sr *eurodnsgo.SoapRequest) error {
	ch, err := c.Schedule(sr)
	if err != nil {
		return err
	}
	defer close(ch)

	// Block until result returns
	_ = <-ch

	return nil
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
func GetDomainList(c eurodnsgo.Client) ([]string, error) {
	var v domainList

	sr := eurodnsgo.NewSoapRequest("domain", "list", &v)

	err := schedule(c, sr)

	return v.Domains, err
}

// GetRecordInfo returns all data inside a specific record
func GetRecordInfo(c eurodnsgo.Client, id int) (Record, error) {
	var v recordInfo

	sr := eurodnsgo.NewSoapRequest("record", "info", &v)
	sr.AddParam(eurodnsgo.NewParam("record", "id", id))

	err := schedule(c, sr)

	return v.Record, err
}

// GetZoneList gets the list of registered zones
func GetZoneList(c eurodnsgo.Client) ([]string, error) {
	var v zoneList

	sr := eurodnsgo.NewSoapRequest("zone", "list", &v)

	err := schedule(c, sr)

	return v.Zones, err
}

// GetZoneInfo gets more information about a zone
func GetZoneInfo(c eurodnsgo.Client, domain string) (Zone, error) {
	var v zoneInfo

	sr := eurodnsgo.NewSoapRequest("zone", "info", &v)
	sr.AddParam(eurodnsgo.NewParam("zone", "name", domain))

	err := schedule(c, sr)

	return v.Zone, err
}

// ZoneRecordAdd adds a new Record object to a Zone
func ZoneRecordAdd(c eurodnsgo.Client, z Zone, r Record) error {
	var v interface{}

	sr := eurodnsgo.NewSoapRequest("zone", "update", &v)

	fmt.Println(sr)

	return nil
}

// ZoneRecordChange changes a Record object inside a Zone
func ZoneRecordChange(c eurodnsgo.Client, z Zone, r Record) error {
	return nil
}

// ZoneRecordDelete deletes a Record object from a Zone
func ZoneRecordDelete(c eurodnsgo.Client, z Zone, r Record) error {
	return nil
}
