package rest

import (
	"github.com/gobuffalo/nulls"
)

type Record struct {
	ID     int64       `json:"id"`
	Name   string      `json:"name"`
	ZoneID int64       `json:"zone_id"`
	Ttl    nulls.Int32 `json:"ttl"`
	Type   string      `json:"type"`
	Value  string      `json:"value"`
	Active bool        `json:"active"`
}
