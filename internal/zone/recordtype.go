package zone

import "fmt"

type RecordType uint8

const (
	invalidRecordType RecordType = iota
	NS
	MX
	A
	AAAA
	CNAME
	TXT
	SRV
	CAA
)

func (t RecordType) IsValid() bool {
	return t > invalidRecordType && t <= CAA
}

var recordTypeToString = []string{
	"NS",
	"MX",
	"A",
	"AAAA",
	"CNAME",
	"TXT",
	"SRV",
	"CAA",
}

func (t RecordType) String() string {
	if !t.IsValid() {
		return fmt.Sprintf("%%!RecordType(%d)", uint8(t))
	}
	return recordTypeToString[t-1]
}

var stringToRecordType = map[string]RecordType{
	"NS":    NS,
	"MX":    MX,
	"A":     A,
	"AAAA":  AAAA,
	"CNAME": CNAME,
	"TXT":   TXT,
	"SRV":   SRV,
	"CAA":   CAA,
}

func RecordTypeFromString(s string) RecordType {
	return stringToRecordType[s]
}
