package rest

import (
	"errors"
	"github.com/1f349/verbena/internal/zone"
	"github.com/miekg/dns"
	"net/netip"
)

type RecordValue struct {
	Value    string `json:"value"`
	Priority int32  `json:"priority,omitempty"`
	Weight   int32  `json:"weight,omitempty"`
	Port     uint16 `json:"port,omitempty"`
}

func (v RecordValue) IsValidForType(recordType string) bool {
	ty := zone.RecordTypeFromString(recordType)
	switch ty {
	case zone.NS:
		return v.Priority == 0 && v.Weight == 0 && v.Port == 0 && validDomainName(v.Value)
	case zone.MX:
		return v.Priority > 0 && v.Weight == 0 && v.Port == 0 && validDomainName(v.Value)
	case zone.A:
		if v.Priority != 0 && v.Weight != 0 && v.Port != 0 {
			return false
		}
		v4, err := netip.ParseAddr(v.Value)
		if err != nil {
			return false
		}
		return v4.Is4()
	case zone.AAAA:
		if v.Priority != 0 && v.Weight != 0 && v.Port != 0 {
			return false
		}
		v6, err := netip.ParseAddr(v.Value)
		if err != nil {
			return false
		}
		return v6.Is6()
	case zone.CNAME:
		return v.Priority == 0 && v.Weight == 0 && v.Port == 0 && validDomainName(v.Value)
	case zone.TXT:
		return v.Value != "" && v.Priority == 0 && v.Weight == 0 && v.Port == 0
	default:
		return false
	}
}

func ParseRecordValue(recordType string, value string) (RecordValue, error) {
	ty := zone.RecordTypeFromString(recordType)
	switch ty {
	case zone.NS:
		if !validDomainName(value) {
			return RecordValue{}, errors.New("invalid NS record")
		}
		return RecordValue{Value: value}, nil
	case zone.MX:
		if !validDomainName(value) {
			return RecordValue{}, errors.New("invalid MX record")
		}
		return RecordValue{Value: value}, nil
	case zone.A:
		v4, err := netip.ParseAddr(value)
		if err != nil || !v4.Is4() {
			return RecordValue{}, errors.New("invalid A record")
		}
		return RecordValue{Value: v4.String()}, nil
	case zone.AAAA:
		v6, err := netip.ParseAddr(value)
		if err != nil || !v6.Is6() {
			return RecordValue{}, errors.New("invalid AAAA record")
		}
		return RecordValue{Value: v6.String()}, nil
	case zone.CNAME:
		if !validDomainName(value) {
			return RecordValue{}, errors.New("invalid CNAME record")
		}
		return RecordValue{Value: value}, nil
	case zone.TXT:
		return RecordValue{Value: value}, nil
	default:
		return RecordValue{}, errors.New("invalid record type")
	}
}

func validDomainName(s string) bool {
	_, ok := dns.IsDomainName(s)
	return ok
}
