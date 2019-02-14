package api

import (
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
	sr.AddArgument("record", "id", id)

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
	sr.AddArgument("zone", "name", domain)

	err := schedule(c, sr)

	return v.Zone, err
}
