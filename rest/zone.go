package rest

import (
	"encoding/json"
	"github.com/1f349/verbena/internal/database"
	"net/http"
	"strconv"
)

type Zone struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Serial  int64  `json:"serial"`
	Admin   string `json:"admin"`
	Refresh int32  `json:"refresh"`
	Retry   int32  `json:"retry"`
	Expire  int32  `json:"expire"`
	Ttl     int32  `json:"ttl"`
	Active  bool   `json:"active"`
}

func (c *Client) GetZones() ([]database.Zone, error) {
	resp, err := doRequest(c, http.MethodGet, "/zones", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var zones []database.Zone
	err = json.NewDecoder(resp.Body).Decode(&zones)
	if err != nil {
		return nil, err
	}
	return zones, nil
}

func (c *Client) GetZone(zoneId int64) (database.Zone, error) {
	resp, err := doRequest(c, http.MethodGet, "/zones/"+strconv.FormatInt(zoneId, 10), nil)
	if err != nil {
		return database.Zone{}, err
	}
	defer resp.Body.Close()

	var zone database.Zone
	err = json.NewDecoder(resp.Body).Decode(&zone)
	if err != nil {
		return database.Zone{}, err
	}
	return zone, nil
}
