package zone

import "github.com/gobuffalo/nulls"

type Record struct {
	Name       string
	TimeToLive nulls.UInt32
	Type       RecordType
	Value      string
}
