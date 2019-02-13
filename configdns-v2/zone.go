package dnsv2

import (
	"sync"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/client-v1"
)


var (
	zoneWriteLock sync.Mutex
)

// Zone represents a DNS zone
/*{
    "zone": "river.com",
    "type": "secondary",
    "masters": [
        "1.2.3.4",
        "1.2.3.5"
    ],
    "comment": "Adding bodies of water"
}

{
    "activationState": "ACTIVE",
    "contractId": "C-1FRYVV3",
    "lastActivationDate": "2018-03-20T06:49:30Z",
    "lastModifiedBy": "vwwuq65mjvsrbvcr",
    "lastModifiedDate": "2019-01-28T12:05:13Z",
    "signAndServe": false,
    "type": "PRIMARY",
    "versionId": "2e9aa959-5e99-405c-b233-360639449fa1",
    "zone": "akamaideveloper.net"
}

*/

type ZoneQueryString struct {
  ContractId         string   `json:"contractid,omitempty"`
  Gid                string   `json:"lastactivationdate,omitempty"`
}

type ZoneCreate struct {
	Zone               string   `json:"zone,omitempty"`
	Type               string   `json:"type,omitempty"`
	Masters            []string `json:"masters,omitempty"`
	Comment            string   `json:"comment,omitempty"`
}


type ZoneResponse struct {
	Zone               string   `json:"zone,omitempty"`
	Type               string   `json:"type,omitempty"`
	Masters            []string `json:"masters,omitempty"`
	Comment            string   `json:"comment,omitempty"`
	ActivationState    string   `json:"activationstate,omitempty"`
	ContractId         string   `json:"contractid,omitempty"`
	LastActivationDate string   `json:"lastactivationdate,omitempty"`
	LastModifiedBy     string   `json:"lastmodifiedby,omitempty"`
	LastModifiedDate   string   `json:"lastmodifieddate,omitempty"`
	SignAndServe       bool   `json:"signandserve,omitempty"`
	VersionId          string   `json:"versionid,omitempty"`
}

// NewZone creates a new Zone
//func NewZone(contractid string, hostname string) *Zone {
func NewZone(params ZoneCreate) *ZoneCreate {
	zone := &ZoneCreate{Zone: params.Zone,Type: params.Type, Masters: params.Masters, Comment: params.Comment}
	return zone
}

func NewZoneResponse(params ZoneCreate) *ZoneResponse {
	zone := &ZoneResponse{Zone: params.Zone}
	return zone
}

func NewZoneQueryString(ContractId string, gid string) *ZoneQueryString {
	zonequerystring := &ZoneQueryString{ContractId: ContractId, Gid: gid }
	return zonequerystring
}

// GetZone retrieves a DNS Zone for a given hostname
//func GetZone(contractid string,hostname string) (*ZoneResponse, error) {
func GetZone(params ZoneCreate) (*ZoneResponse, error) {
	zone := NewZoneResponse(params)
	req, err := client.NewRequest(
		Config,
		"GET",
		"/config-dns/v2/zones/"+zone.Zone,
		nil,
	)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(Config, req)
	if err != nil {
		return nil, err
	}

	if client.IsError(res) && res.StatusCode != 404 {
		return nil, client.NewAPIError(res)
	} else if res.StatusCode == 404 {
		return nil, &ZoneError{zoneName: params.Zone}
	} else {
		err = client.BodyJSON(res, zone)
		if err != nil {
			return nil, err
		}

		return zone, nil
	}
}

// Save updates the Zone
func (zone *ZoneCreate ) Save(zonequerystring ZoneQueryString) error {
	// This lock will restrict the concurrency of API calls
	// to 1 save request at a time. This is needed for the Soa.Serial value which
	// is required to be incremented for every subsequent update to a zone
	// so we have to save just one request at a time to ensure this is always
	// incremented properly
	zoneWriteLock.Lock()
	defer zoneWriteLock.Unlock()

	req, err := client.NewJSONRequest(
		Config,
		"POST",
		"/config-dns/v2/zones/?contractId="+zonequerystring.ContractId+"&gid="+zonequerystring.Gid,
		zone,
	)
	if err != nil {
		return err
	}

	res, err := client.Do(Config, req)

	// Network error
	if err != nil {
		return &ZoneError{
			zoneName:         zone.Zone,
			httpErrorMessage: err.Error(),
			err:              err,
		}
	}

	// API error
	if client.IsError(res) {
		err := client.NewAPIError(res)
		return &ZoneError{zoneName: zone.Zone, apiErrorMessage: err.Detail, err: err}
	}

	/*for {
		updatedZone, err := GetZone(zone.ContractId,zone.Zone)
		if err != nil {
			return err
		}

		if updatedZone.VersionId != zone.VersionId {
			*zone = *updatedZone
			break
		}
		time.Sleep(time.Second)
	}*/

	return nil
}

func (zone *ZoneCreate) Delete(zonequerystring ZoneQueryString) error {
	// remove all the records except for SOA
	// which is required and save the zone
	return zone.Save(zonequerystring)
}
