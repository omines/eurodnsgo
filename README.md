[![Build Status](https://travis-ci.com/omines/eurodnsgo.svg?branch=master)](https://travis-ci.com/omines/eurodnsgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/omines/eurodnsgo)](https://goreportcard.com/report/github.com/omines/eurodnsgo)
[![License MIT](https://img.shields.io/badge/License-MIT-brightgreen.svg)](https://github.com/omines/eurodnsgo/blob/master/LICENSE)
[![GoDoc eurodnsgo](https://godoc.org/github.com/omines/eurodnsgo?status.svg)](https://godoc.org/github.com/omines/eurodnsgo)

# EuroDNS API bindings for Go.

Small wrapper with bindings to connect to the EuroDNS API with golang.

## Installation

~~~~
go get github.com/omines/eurodnsgo
~~~~

Or you can manually git clone the repository to
`$(go env GOPATH)/src/github.com/omines/eurodnsgo`.

## Documentation

#### Create a client

The eurodnsgo packages need to be imported

```go
// Example imports
import (
    "github.com/omines/eurodnsgo"
    "github.com/omines/eurodnsgo/api"
)
```

Create a `eurodnsgo.Client` and provide a `context.Context`

```go
// Provide for a context.Context. either create it or derive from it
ctx := context.TODO()
client, err := eurodnsgo.NewClient(eurodnsgo.ClientConfig{
    Host:     "api.eurodns-endpoint.org",
    Username: "username",
    Password: "password",
})
```

#### Get a list of registered domains

```go
// This function will return a slice of strings representing the domain
// names
domainList, err := api.GetDomainList(ctx, client)
```

#### Get a list of available zones

```go
// This function will return a slice of string with all the available
// domain names
zoneList, err := api.GetZoneList(ctx, client)
```

#### Get more information about a zone and load it DNS records

```go
// Will get detailed information from a zone with zone-name 'x'. The
// function will return a fully populated api.Zone instance
zone, err := api.GetZoneInfo(ctx, client, "fqdn.org")
```

#### Retrieve a full list of records from the Zone struct 

```go
// Te api.Zone instance contains a slice of api.Record pointers
// holding the list of api.Record objects available.
// TODO: These should be passed by value
var recordList []*api.Record = zone.Records
```

#### Mutate an existing Record

```go
// Update an instance of api.Record and feed it to ZoneRecordChange
// to have the changes persisted against the API
record := recordList[0]
record.Data = "8.8.8.8"
record.TTL = 3600
err = api.ZoneRecordChange(ctx, client, zone, *record)
```

#### Create a new Record

```go
// To add a new DNS record use ZoneRecordAdd. Provide any properties
// that is allowed
err = api.ZoneRecordAdd(ctx, client, zone, api.Record{
    Host: "@",
    TTL:  3600,
    Type: api.RecordTypeA,
    Data: "8.8.8.8",
})
```

## Legal

This software was developed for internal use at [Omines Full Service Internetbureau](https://www.omines.nl/)
in Eindhoven, the Netherlands. It is shared with the general public under the permissive MIT license, without
any guarantee of fitness for any particular purpose. Refer to the included `LICENSE` file for more details.
