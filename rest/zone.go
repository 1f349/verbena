package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/miekg/dns"
)

type Zone struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Serial  uint32 `json:"serial"`
	Admin   string `json:"admin"`
	Refresh int32  `json:"refresh"`
	Retry   int32  `json:"retry"`
	Expire  int32  `json:"expire"`
	Ttl     int32  `json:"ttl"`
	Active  bool   `json:"active"`

	Nameservers []string `json:"nameservers"`
}

func (c *Client) GetZones() ([]Zone, error) {
	resp, err := doRequest(c, http.MethodGet, "/zones", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var zones []Zone
	err = json.NewDecoder(resp.Body).Decode(&zones)
	if err != nil {
		return nil, err
	}
	return zones, nil
}

func (c *Client) GetZone(zoneId int64) (Zone, error) {
	resp, err := doRequest(c, http.MethodGet, "/zones/"+strconv.FormatInt(zoneId, 10), nil)
	if err != nil {
		return Zone{}, err
	}
	defer resp.Body.Close()

	var zone Zone
	err = json.NewDecoder(resp.Body).Decode(&zone)
	if err != nil {
		return Zone{}, err
	}
	return zone, nil
}

func (c *Client) LookupZone(zoneName string) (int64, error) {
	_, validDomain := dns.IsDomainName(zoneName)
	if !validDomain {
		return 0, fmt.Errorf("invalid zone: %s", zoneName)
	}

	resp, err := doRequest(c, http.MethodGet, "/zones/lookup/"+zoneName, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var zoneId struct {
		ID int64 `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&zoneId)
	if err != nil {
		return 0, err
	}
	return zoneId.ID, nil
}
