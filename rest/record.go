package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/nulls"
	"net/http"
	"strconv"
)

type Record struct {
	ID     int64       `json:"id"`
	Name   string      `json:"name"`
	ZoneID int64       `json:"zone_id"`
	Ttl    nulls.Int32 `json:"ttl"`
	Type   string      `json:"type"`
	Value  RecordValue `json:"value"`
	Active bool        `json:"active"`
}

type CreateRecord struct {
	Name   string      `json:"name"`
	Ttl    nulls.Int32 `json:"ttl"`
	Type   string      `json:"type"`
	Value  RecordValue `json:"value"`
	Active bool        `json:"active"`
}

type PutRecord struct {
	Ttl    nulls.Int32 `json:"ttl"`
	Value  RecordValue `json:"value"`
	Active bool        `json:"active"`
}

func (c *Client) GetZoneRecords(zoneId int64) ([]Record, error) {
	resp, err := doRequest(c, http.MethodGet, "/zones/"+strconv.FormatInt(zoneId, 10)+"/records", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var records []Record
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (c *Client) GetZoneRecord(zoneId, recordId int64) (Record, error) {
	resp, err := doRequest(c, http.MethodGet, "/zones/"+strconv.FormatInt(zoneId, 10)+"/records/"+strconv.FormatInt(recordId, 10), nil)
	if err != nil {
		return Record{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Record{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var record Record
	err = json.NewDecoder(resp.Body).Decode(&record)
	if err != nil {
		return Record{}, err
	}
	return record, nil
}

func (c *Client) CreateZoneRecord(zoneId int64, createRecord CreateRecord) error {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(createRecord)
	if err != nil {
		return err
	}

	resp, err := doRequest(c, http.MethodPost, "/zones/"+strconv.FormatInt(zoneId, 10)+"/records", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func (c *Client) UpdateZoneRecord(zoneId, recordId int64, putRecord PutRecord) (Record, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(putRecord)
	if err != nil {
		return Record{}, err
	}

	resp, err := doRequest(c, http.MethodPut, "/zones/"+strconv.FormatInt(zoneId, 10)+"/records/"+strconv.FormatInt(recordId, 10), buf)
	if err != nil {
		return Record{}, err
	}
	defer resp.Body.Close()

	var record Record
	err = json.NewDecoder(resp.Body).Decode(&record)
	if err != nil {
		return Record{}, err
	}
	return record, nil
}

func (c *Client) DeleteZoneRecord(zoneId, recordId int64) error {
	resp, err := doRequest(c, http.MethodDelete, "/zones/"+strconv.FormatInt(zoneId, 10)+"/records/"+strconv.FormatInt(recordId, 10), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}
