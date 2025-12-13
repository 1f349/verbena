package zone

import "github.com/gobuffalo/nulls"

type SoaRecord struct {
	Nameserver string
	Admin      string
	Serial     uint32
	Refresh    uint32
	Retry      uint32
	Expire     uint32
	TimeToLive uint32
}

type Record struct {
	Name       string
	TimeToLive nulls.UInt32
	Type       RecordType
	Value      string
}
